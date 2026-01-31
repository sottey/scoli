package api

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"io/fs"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type AIIndex struct {
	db     *sql.DB
	logger *slog.Logger
	mu     sync.Mutex
}

type AIChunk struct {
	NotePath   string
	Heading    string
	ChunkIndex int
	Content    string
	Embedding  []float32
}

type AIChunkMatch struct {
	NotePath string
	Heading  string
	Content  string
	Score    float64
}

func (s *Server) getAIIndex() (*AIIndex, error) {
	var initErr error
	s.aiIndexOnce.Do(func() {
		db, err := sql.Open("sqlite", s.aiIndexPath())
		if err != nil {
			initErr = err
			return
		}
		db.SetMaxOpenConns(1)
		idx := &AIIndex{
			db:     db,
			logger: s.logger.With("component", "ai-index"),
		}
		if err := idx.ensureSchema(); err != nil {
			initErr = err
			return
		}
		s.aiIndexStore = idx
	})
	if initErr != nil {
		return nil, initErr
	}
	if s.aiIndexStore == nil {
		return nil, errors.New("ai index unavailable")
	}
	return s.aiIndexStore, nil
}

func (idx *AIIndex) ensureSchema() error {
	const notesTable = `
CREATE TABLE IF NOT EXISTS notes (
	note_path TEXT PRIMARY KEY,
	note_modified INTEGER NOT NULL
);`
	const chunksTable = `
CREATE TABLE IF NOT EXISTS chunks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	note_path TEXT NOT NULL,
	heading TEXT,
	chunk_index INTEGER NOT NULL,
	content TEXT NOT NULL,
	embedding BLOB NOT NULL,
	FOREIGN KEY(note_path) REFERENCES notes(note_path) ON DELETE CASCADE
);`
	const chunksIndex = `CREATE INDEX IF NOT EXISTS idx_chunks_note_path ON chunks(note_path);`
	_, err := idx.db.Exec(notesTable)
	if err != nil {
		return err
	}
	if _, err := idx.db.Exec(chunksTable); err != nil {
		return err
	}
	if _, err := idx.db.Exec(chunksIndex); err != nil {
		return err
	}
	return nil
}

func (idx *AIIndex) ensureIndex(ctx context.Context, settings AISettings, notesDir string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	notes, err := listMarkdownNotes(notesDir)
	if err != nil {
		return err
	}
	existing, err := idx.fetchIndexedNotes()
	if err != nil {
		return err
	}

	for _, note := range notes {
		mod := note.Modified.Unix()
		if existingMod, ok := existing[note.Path]; ok && existingMod == mod {
			delete(existing, note.Path)
			continue
		}
		if err := idx.upsertNote(ctx, settings, notesDir, note); err != nil {
			return err
		}
		delete(existing, note.Path)
	}

	if len(existing) > 0 {
		for notePath := range existing {
			if err := idx.deleteNote(notePath); err != nil {
				return err
			}
		}
	}

	return nil
}

type noteInfo struct {
	Path     string
	Modified time.Time
}

func listMarkdownNotes(notesDir string) ([]noteInfo, error) {
	notes := make([]noteInfo, 0)
	err := filepath.WalkDir(notesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if isAiDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if isIgnoredFile(d.Name()) || !isMarkdown(d.Name()) {
			return nil
		}
		rel, err := filepath.Rel(notesDir, path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		notes = append(notes, noteInfo{
			Path:     filepath.ToSlash(rel),
			Modified: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return notes, nil
}

func (idx *AIIndex) fetchIndexedNotes() (map[string]int64, error) {
	rows, err := idx.db.Query("SELECT note_path, note_modified FROM notes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]int64)
	for rows.Next() {
		var path string
		var mod int64
		if err := rows.Scan(&path, &mod); err != nil {
			return nil, err
		}
		result[path] = mod
	}
	return result, rows.Err()
}

func (idx *AIIndex) upsertNote(ctx context.Context, settings AISettings, notesDir string, note noteInfo) error {
	absPath := filepath.Join(notesDir, filepath.FromSlash(note.Path))
	data, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}
	chunks := chunkNoteContent(note.Path, string(data), settings)
	if len(chunks) == 0 {
		chunks = []AIChunk{
			{
				NotePath:   note.Path,
				Heading:    "",
				ChunkIndex: 0,
				Content:    strings.TrimSpace(string(data)),
			},
		}
	}

	filteredChunks := make([]AIChunk, 0, len(chunks))
	inputs := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		trimmed := strings.TrimSpace(chunk.Content)
		if trimmed == "" {
			continue
		}
		chunk.Content = trimmed
		filteredChunks = append(filteredChunks, chunk)
		inputs = append(inputs, trimmed)
	}

	tx, err := idx.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM chunks WHERE note_path = ?", note.Path); err != nil {
		return err
	}
	if _, err := tx.Exec("INSERT OR REPLACE INTO notes (note_path, note_modified) VALUES (?, ?)", note.Path, note.Modified.Unix()); err != nil {
		return err
	}

	if len(filteredChunks) > 0 {
		embeddings, err := openAIEmbeddings(ctx, settings, inputs)
		if err != nil {
			return err
		}
		if len(embeddings) != len(filteredChunks) {
			return errors.New("embedding count mismatch")
		}

		stmt, err := tx.Prepare("INSERT INTO chunks (note_path, heading, chunk_index, content, embedding) VALUES (?, ?, ?, ?, ?)")
		if err != nil {
			return err
		}
		defer stmt.Close()
		for i, chunk := range filteredChunks {
			embedding := encodeEmbedding(embeddings[i])
			if _, err := stmt.Exec(note.Path, chunk.Heading, chunk.ChunkIndex, chunk.Content, embedding); err != nil {
				return err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (idx *AIIndex) deleteNote(notePath string) error {
	tx, err := idx.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("DELETE FROM chunks WHERE note_path = ?", notePath); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM notes WHERE note_path = ?", notePath); err != nil {
		return err
	}
	return tx.Commit()
}

func (idx *AIIndex) query(ctx context.Context, settings AISettings, notesDir, queryText string) ([]AIChunkMatch, error) {
	if err := idx.ensureIndex(ctx, settings, notesDir); err != nil {
		return nil, err
	}
	queryEmbedding, err := openAIEmbeddings(ctx, settings, []string{queryText})
	if err != nil {
		return nil, err
	}
	if len(queryEmbedding) == 0 {
		return nil, errors.New("empty embedding response")
	}
	queryVec := queryEmbedding[0]

	rows, err := idx.db.Query("SELECT note_path, heading, content, embedding FROM chunks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := make([]AIChunkMatch, 0)
	for rows.Next() {
		var path, heading, content string
		var embeddingBytes []byte
		if err := rows.Scan(&path, &heading, &content, &embeddingBytes); err != nil {
			return nil, err
		}
		embedding, err := decodeEmbedding(embeddingBytes)
		if err != nil {
			return nil, err
		}
		score := cosineSimilarity(queryVec, embedding)
		matches = append(matches, AIChunkMatch{
			NotePath: path,
			Heading:  heading,
			Content:  content,
			Score:    score,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	if settings.TopK > 0 && len(matches) > settings.TopK {
		matches = matches[:settings.TopK]
	}
	return matches, nil
}

func encodeEmbedding(vec []float32) []byte {
	data := make([]byte, len(vec)*4)
	for i, v := range vec {
		binary.LittleEndian.PutUint32(data[i*4:], math.Float32bits(v))
	}
	return data
}

func decodeEmbedding(data []byte) ([]float32, error) {
	if len(data)%4 != 0 {
		return nil, errors.New("invalid embedding data")
	}
	vec := make([]float32, len(data)/4)
	for i := 0; i < len(vec); i++ {
		vec[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:]))
	}
	return vec, nil
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := 0; i < len(a); i++ {
		av := float64(a[i])
		bv := float64(b[i])
		dot += av * bv
		normA += av * av
		normB += bv * bv
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

var headingPattern = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

func chunkNoteContent(path, content string, settings AISettings) []AIChunk {
	lines := strings.Split(content, "\n")
	type section struct {
		heading    string
		paragraphs []string
	}
	sections := make([]section, 0, 8)
	currentHeading := ""
	currentParagraphs := make([]string, 0, 4)
	var para []string
	var headingStack []string
	inCodeBlock := false

	flushParagraph := func() {
		if len(para) == 0 {
			return
		}
		text := strings.TrimSpace(strings.Join(para, "\n"))
		if text != "" {
			currentParagraphs = append(currentParagraphs, text)
		}
		para = nil
	}

	flushSection := func() {
		flushParagraph()
		if len(currentParagraphs) == 0 {
			return
		}
		sections = append(sections, section{
			heading:    currentHeading,
			paragraphs: currentParagraphs,
		})
		currentParagraphs = make([]string, 0, 4)
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			para = append(para, line)
			continue
		}
		if !inCodeBlock {
			if match := headingPattern.FindStringSubmatch(trimmed); match != nil {
				flushParagraph()
				flushSection()
				level := len(match[1])
				title := strings.TrimSpace(match[2])
				if title == "" {
					title = "Untitled"
				}
				if level <= len(headingStack) {
					headingStack = headingStack[:level-1]
				}
				headingStack = append(headingStack, title)
				currentHeading = strings.Join(headingStack, " / ")
				continue
			}
		}
		if trimmed == "" {
			flushParagraph()
			continue
		}
		para = append(para, line)
	}
	flushSection()

	chunks := make([]AIChunk, 0)
	chunkIndex := 0
	for _, sec := range sections {
		var buffer string
		flushChunk := func() {
			if strings.TrimSpace(buffer) == "" {
				buffer = ""
				return
			}
			chunks = append(chunks, AIChunk{
				NotePath:   path,
				Heading:    sec.heading,
				ChunkIndex: chunkIndex,
				Content:    strings.TrimSpace(buffer),
			})
			chunkIndex++
			buffer = ""
		}

		for _, paragraph := range sec.paragraphs {
			if paragraph == "" {
				continue
			}
			parts := splitTextByLimit(paragraph, settings.ChunkCharLimit)
			for _, part := range parts {
				if buffer == "" {
					buffer = part
					continue
				}
				if len(buffer)+2+len(part) <= settings.ChunkCharLimit {
					buffer = buffer + "\n\n" + part
					continue
				}
				flushChunk()
				buffer = part
			}
		}
		flushChunk()
	}
	return chunks
}

func splitTextByLimit(text string, limit int) []string {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return nil
	}
	if limit <= 0 || len(clean) <= limit {
		return []string{clean}
	}
	parts := make([]string, 0)
	for len(clean) > limit {
		parts = append(parts, strings.TrimSpace(clean[:limit]))
		clean = strings.TrimSpace(clean[limit:])
	}
	if clean != "" {
		parts = append(parts, clean)
	}
	return parts
}
