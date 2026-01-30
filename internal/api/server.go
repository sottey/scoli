package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	notesDir           string
	logger             *slog.Logger
	emailSchedulerOnce sync.Once
}

var timeNow = time.Now

const inboxNotePath = "Inbox.md"
const scratchNotePath = "scratch.md"
const dailyFolderName = "Daily"
const dailyDateLayout = "2006-01-02"
const sheetsFolderName = "Sheets"
const sheetExtension = ".jsh"

type TemplateContext struct {
	Title  string
	Path   string
	Folder string
}

type TreeNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Type     string     `json:"type"`
	Children []TreeNode `json:"children,omitempty"`
}

type NoteResponse struct {
	Path     string    `json:"path"`
	Content  string    `json:"content"`
	Modified time.Time `json:"modified"`
}

type NotePayload struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type NoteRenamePayload struct {
	Path    string `json:"path"`
	NewPath string `json:"newPath"`
}

type FolderPayload struct {
	Path    string `json:"path"`
	NewPath string `json:"newPath"`
}

type SearchResult struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
	ID   string `json:"id,omitempty"`
}

type TagGroup struct {
	Tag   string         `json:"tag"`
	Notes []SearchResult `json:"notes"`
}

type MentionGroup struct {
	Mention string         `json:"mention"`
	Notes   []SearchResult `json:"notes"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	if err := s.ensureInboxNote(); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to ensure inbox note")
		return
	}
	if err := s.ensureDailyNote(); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to ensure daily note")
		return
	}
	settings, _, err := s.loadSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load settings")
		return
	}
	pathParam := r.URL.Query().Get("path")
	absPath, relPath, err := s.resolvePath(pathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "path not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to read path")
		return
	}
	if !info.IsDir() {
		writeError(w, http.StatusBadRequest, "path must be a folder")
		return
	}

	root := TreeNode{
		Name: "Notes",
		Path: relPath,
		Type: "folder",
	}

	children, err := s.buildTree(absPath, relPath, settings.ShowTemplates, settings.NotesSortBy, settings.NotesSortOrder)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to build tree")
		return
	}
	root.Children = children

	writeJSON(w, http.StatusOK, root)
}

func (s *Server) handleGetNote(w http.ResponseWriter, r *http.Request) {
	pathParam := r.URL.Query().Get("path")
	if strings.TrimSpace(pathParam) == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	absPath, relPath, err := s.resolvePath(pathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			if strings.EqualFold(relPath, inboxNotePath) {
				if err := s.ensureInboxNote(); err != nil {
					writeError(w, http.StatusInternalServerError, "unable to ensure inbox note")
					return
				}
				info, err = os.Stat(absPath)
			}
		}
		if err != nil {
			if os.IsNotExist(err) {
				writeError(w, http.StatusNotFound, "note not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "unable to read note")
			return
		}
	}
	if info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is a folder")
		return
	}
	if !isNoteFile(absPath) {
		writeError(w, http.StatusBadRequest, "not a note file")
		return
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to read note")
		return
	}

	resp := NoteResponse{
		Path:     relPath,
		Content:  string(data),
		Modified: info.ModTime(),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleCreateNote(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[NotePayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(payload.Path) == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	pathParam := strings.TrimSpace(payload.Path)
	if isTemplate(pathParam) {
		pathParam = ensureTemplate(pathParam)
	} else {
		pathParam = ensureMarkdown(pathParam)
	}
	absPath, relPath, err := s.resolvePath(pathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateReservedRootPath(relPath, false); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if isDailyPath(relPath) && !isTemplate(relPath) {
		if _, ok := parseDailyNoteDate(relPath); !ok {
			writeError(w, http.StatusBadRequest, "daily notes must use YYYY-MM-DD")
			return
		}
	}

	if _, err := os.Stat(absPath); err == nil {
		writeError(w, http.StatusConflict, "note already exists")
		return
	} else if !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "unable to check note")
		return
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to create parent folders")
		return
	}

	content := payload.Content
	notice := ""
	templateContent, ok, err := s.folderTemplateContent(filepath.Dir(absPath))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load folder template")
		return
	}
	if ok {
		targetTime := inferTargetTime(relPath, timeNow())
		conditioned, warnings := applyTemplateConditionals(string(templateContent), targetTime)
		content = applyTemplatePlaceholders(conditioned, targetTime, templateContext(relPath))
		if len(warnings) > 0 {
			notice = strings.Join(warnings, " ")
			s.logger.Warn("template condition warnings", "path", relPath, "warning", notice)
		}
	}

	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to create note")
		return
	}

	// Task parsing is done on demand from note contents.

	s.logger.Info("note created", "path", relPath, "bytes", len(content), "template", ok)
	if notice != "" {
		writeJSON(w, http.StatusCreated, map[string]string{"path": relPath, "notice": notice})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"path": relPath})
}

func (s *Server) handleUpdateNote(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[NotePayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(payload.Path) == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	absPath, relPath, err := s.resolvePath(payload.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "note not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to read note")
		return
	}
	if info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is a folder")
		return
	}
	if !isNoteFile(absPath) {
		writeError(w, http.StatusBadRequest, "not a note file")
		return
	}

	if err := os.WriteFile(absPath, []byte(payload.Content), 0o644); err != nil {
		s.logger.Error("unable to update note", "path", relPath, "absPath", absPath, "error", err)
		writeError(w, http.StatusInternalServerError, "unable to update note")
		return
	}

	// Task parsing is done on demand from note contents.

	s.logger.Info("note updated", "path", relPath, "bytes", len(payload.Content))
	writeJSON(w, http.StatusOK, map[string]string{"path": relPath})
}

func (s *Server) handleDeleteNote(w http.ResponseWriter, r *http.Request) {
	pathParam := r.URL.Query().Get("path")
	if strings.TrimSpace(pathParam) == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	absPath, relPath, err := s.resolvePath(pathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "note not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to read note")
		return
	}
	if info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is a folder")
		return
	}
	if !isNoteFile(absPath) {
		writeError(w, http.StatusBadRequest, "not a note file")
		return
	}

	if err := os.Remove(absPath); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to delete note")
		return
	}

	s.logger.Info("note deleted", "path", relPath)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleGetFile(w http.ResponseWriter, r *http.Request) {
	pathParam := r.URL.Query().Get("path")
	if strings.TrimSpace(pathParam) == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	absPath, _, err := s.resolvePath(pathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to read file")
		return
	}
	if info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is a folder")
		return
	}

	http.ServeFile(w, r, absPath)
}

func (s *Server) handleCreateFolder(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[FolderPayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(payload.Path) == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	absPath, relPath, err := s.resolvePath(payload.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateReservedRootPath(relPath, true); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := os.Stat(absPath); err == nil {
		writeError(w, http.StatusConflict, "folder already exists")
		return
	} else if !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "unable to check folder")
		return
	}

	if err := os.MkdirAll(absPath, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to create folder")
		return
	}

	s.logger.Info("folder created", "path", relPath)
	writeJSON(w, http.StatusCreated, map[string]string{"path": relPath})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		writeError(w, http.StatusBadRequest, "query is required")
		return
	}

	lowerQuery := strings.ToLower(query)
	var results []SearchResult

	err := filepath.WalkDir(s.notesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if isIgnoredFile(d.Name()) {
			return nil
		}
		if !isMarkdown(d.Name()) {
			return nil
		}

		rel, err := filepath.Rel(s.notesDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		nameLower := strings.ToLower(d.Name())
		if strings.Contains(nameLower, lowerQuery) {
			results = append(results, SearchResult{
				Path: rel,
				Name: d.Name(),
				Type: "note",
			})
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if strings.Contains(strings.ToLower(string(data)), lowerQuery) {
			results = append(results, SearchResult{
				Path: rel,
				Name: d.Name(),
				Type: "note",
			})
		}

		return nil
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to search notes")
		return
	}

	// Task search is intentionally disabled for now.

	writeJSON(w, http.StatusOK, results)
}

func (s *Server) handleTags(w http.ResponseWriter, r *http.Request) {
	tagPattern := regexp.MustCompile(`(^|\s)#([A-Za-z]+)\b`)
	tagMap := make(map[string]map[string]string)

	err := filepath.WalkDir(s.notesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if isIgnoredFile(d.Name()) {
			return nil
		}
		if !isMarkdown(d.Name()) {
			return nil
		}

		rel, err := filepath.Rel(s.notesDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		cleaned := stripCodeBlocksAndInline(string(data))
		matches := tagPattern.FindAllStringSubmatch(cleaned, -1)
		if len(matches) == 0 {
			return nil
		}

		baseName := filepath.Base(rel)
		for _, match := range matches {
			tag := match[2]
			if tag == "" {
				continue
			}
			if tagMap[tag] == nil {
				tagMap[tag] = make(map[string]string)
			}
			tagMap[tag][rel] = baseName
		}

		return nil
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to list tags")
		return
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}
	sort.Slice(tags, func(i, j int) bool {
		return strings.ToLower(tags[i]) < strings.ToLower(tags[j])
	})

	groups := make([]TagGroup, 0, len(tags))
	for _, tag := range tags {
		notesMap := tagMap[tag]
		notes := make([]SearchResult, 0, len(notesMap))
		for path, name := range notesMap {
			notes = append(notes, SearchResult{Path: path, Name: name})
		}
		sort.Slice(notes, func(i, j int) bool {
			nameA := strings.ToLower(notes[i].Name)
			nameB := strings.ToLower(notes[j].Name)
			if nameA == nameB {
				return notes[i].Path < notes[j].Path
			}
			return nameA < nameB
		})
		groups = append(groups, TagGroup{Tag: tag, Notes: notes})
	}

	writeJSON(w, http.StatusOK, groups)
}

func (s *Server) handleMentions(w http.ResponseWriter, r *http.Request) {
	mentionMap := make(map[string]map[string]string)

	err := filepath.WalkDir(s.notesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if isIgnoredFile(d.Name()) {
			return nil
		}
		if !isMarkdown(d.Name()) {
			return nil
		}

		rel, err := filepath.Rel(s.notesDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		cleaned := stripCodeBlocksAndInline(string(data))
		matches := taskMentionPattern.FindAllStringSubmatch(cleaned, -1)
		if len(matches) == 0 {
			return nil
		}

		baseName := filepath.Base(rel)
		for _, match := range matches {
			if len(match) < 3 {
				continue
			}
			mention := strings.ToLower(match[2])
			if mention == "" {
				continue
			}
			if mentionMap[mention] == nil {
				mentionMap[mention] = make(map[string]string)
			}
			mentionMap[mention][rel] = baseName
		}

		return nil
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to list mentions")
		return
	}

	mentions := make([]string, 0, len(mentionMap))
	for mention := range mentionMap {
		mentions = append(mentions, mention)
	}
	sort.Slice(mentions, func(i, j int) bool {
		return strings.ToLower(mentions[i]) < strings.ToLower(mentions[j])
	})

	groups := make([]MentionGroup, 0, len(mentions))
	for _, mention := range mentions {
		notesMap := mentionMap[mention]
		notes := make([]SearchResult, 0, len(notesMap))
		for path, name := range notesMap {
			notes = append(notes, SearchResult{Path: path, Name: name})
		}
		sort.Slice(notes, func(i, j int) bool {
			nameA := strings.ToLower(notes[i].Name)
			nameB := strings.ToLower(notes[j].Name)
			if nameA == nameB {
				return notes[i].Path < notes[j].Path
			}
			return nameA < nameB
		})
		groups = append(groups, MentionGroup{Mention: mention, Notes: notes})
	}

	writeJSON(w, http.StatusOK, groups)
}

func (s *Server) handleRenameNote(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[NoteRenamePayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(payload.Path) == "" || strings.TrimSpace(payload.NewPath) == "" {
		writeError(w, http.StatusBadRequest, "path and newPath are required")
		return
	}

	absPath, relPath, err := s.resolvePath(payload.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "note not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to read note")
		return
	}
	if info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is a folder")
		return
	}
	if !isNoteFile(absPath) {
		writeError(w, http.StatusBadRequest, "not a note file")
		return
	}

	newPathInput := strings.TrimSpace(payload.NewPath)
	if isTemplate(absPath) {
		newPathInput = ensureTemplate(newPathInput)
	} else {
		newPathInput = ensureMarkdown(newPathInput)
	}
	absNewPath, relNewPath, err := s.resolvePath(newPathInput)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateReservedRootPath(relNewPath, false); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if isDailyPath(relNewPath) && !isTemplate(relNewPath) {
		if _, ok := parseDailyNoteDate(relNewPath); !ok {
			writeError(w, http.StatusBadRequest, "daily notes must use YYYY-MM-DD")
			return
		}
	}

	if _, err := os.Stat(absNewPath); err == nil {
		writeError(w, http.StatusConflict, "destination already exists")
		return
	} else if !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "unable to check destination")
		return
	}

	if err := os.MkdirAll(filepath.Dir(absNewPath), 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to prepare destination")
		return
	}

	if err := os.Rename(absPath, absNewPath); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to rename note")
		return
	}

	s.logger.Info("note renamed", "path", relPath, "newPath", relNewPath)
	writeJSON(w, http.StatusOK, map[string]string{"path": relPath, "newPath": relNewPath})
}

func (s *Server) handleRenameFolder(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[FolderPayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(payload.Path) == "" || strings.TrimSpace(payload.NewPath) == "" {
		writeError(w, http.StatusBadRequest, "path and newPath are required")
		return
	}

	absPath, relPath, err := s.resolvePath(payload.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	absNewPath, relNewPath, err := s.resolvePath(payload.NewPath)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateReservedRootPath(relNewPath, true); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.EqualFold(relPath, dailyFolderName) || strings.EqualFold(relNewPath, dailyFolderName) {
		writeError(w, http.StatusBadRequest, "Daily root cannot be moved or renamed")
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "folder not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to read folder")
		return
	}
	if !info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is not a folder")
		return
	}

	if _, err := os.Stat(absNewPath); err == nil {
		writeError(w, http.StatusConflict, "destination already exists")
		return
	} else if !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "unable to check destination")
		return
	}

	if err := os.MkdirAll(filepath.Dir(absNewPath), 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to prepare destination")
		return
	}

	if err := os.Rename(absPath, absNewPath); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to rename folder")
		return
	}

	s.logger.Info("folder renamed", "path", relPath, "newPath", relNewPath)
	writeJSON(w, http.StatusOK, map[string]string{"path": relPath, "newPath": relNewPath})
}

func (s *Server) handleDeleteFolder(w http.ResponseWriter, r *http.Request) {
	pathParam := r.URL.Query().Get("path")
	if strings.TrimSpace(pathParam) == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	absPath, relPath, err := s.resolvePath(pathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "folder not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to read folder")
		return
	}
	if !info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is not a folder")
		return
	}

	if err := os.RemoveAll(absPath); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to delete folder")
		return
	}

	s.logger.Info("folder deleted", "path", relPath)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

type treeNodeSortEntry struct {
	node    TreeNode
	created time.Time
	updated time.Time
}

func compareStringsInsensitive(a, b string) int {
	return strings.Compare(strings.ToLower(a), strings.ToLower(b))
}

func compareTimes(a, b time.Time) int {
	if a.Before(b) {
		return -1
	}
	if a.After(b) {
		return 1
	}
	return 0
}

func compareTreeEntries(a, b treeNodeSortEntry, sortBy, sortOrder string) bool {
	typeOrder := map[string]int{
		"folder": 0,
		"file":   1,
		"asset":  2,
		"pdf":    3,
		"csv":    4,
	}
	if a.node.Type != b.node.Type {
		return typeOrder[a.node.Type] < typeOrder[b.node.Type]
	}

	cmp := 0
	switch sortBy {
	case notesSortByCreated:
		cmp = compareTimes(a.created, b.created)
	case notesSortByUpdated:
		cmp = compareTimes(a.updated, b.updated)
	default:
		cmp = compareStringsInsensitive(a.node.Name, b.node.Name)
	}
	if cmp == 0 {
		cmp = compareStringsInsensitive(a.node.Name, b.node.Name)
	}
	if sortOrder == notesSortOrderDesc {
		cmp = -cmp
	}
	return cmp < 0
}

func compareTreeEntriesDaily(a, b treeNodeSortEntry) bool {
	aIsFolder := a.node.Type == "folder"
	bIsFolder := b.node.Type == "folder"
	if aIsFolder != bIsFolder {
		return !aIsFolder
	}

	cmp := compareStringsInsensitive(a.node.Name, b.node.Name)
	if cmp == 0 {
		cmp = compareStringsInsensitive(a.node.Path, b.node.Path)
	}
	return cmp > 0
}

func (s *Server) buildTree(absPath, relPath string, showTemplates bool, sortBy, sortOrder string) ([]TreeNode, error) {
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	var nodes []treeNodeSortEntry
	for _, entry := range entries {
		name := entry.Name()
		if relPath == "" && entry.IsDir() && strings.EqualFold(name, sheetsFolderName) {
			continue
		}
		if relPath == "" && entry.IsDir() && strings.EqualFold(name, emailTemplatesDir) {
			continue
		}
		childRel := filepath.Join(relPath, name)
		childAbs := filepath.Join(absPath, name)
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		createdAt := getBestEffortCreatedTime(info)
		updatedAt := info.ModTime()

		if entry.IsDir() {
			children, err := s.buildTree(childAbs, childRel, showTemplates, sortBy, sortOrder)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, treeNodeSortEntry{
				node: TreeNode{
					Name:     name,
					Path:     filepath.ToSlash(childRel),
					Type:     "folder",
					Children: children,
				},
				created: createdAt,
				updated: updatedAt,
			})
			continue
		}

		if isIgnoredFile(name) {
			continue
		}

		if !isMarkdown(name) {
			if showTemplates && isTemplate(name) {
				nodes = append(nodes, treeNodeSortEntry{
					node: TreeNode{
						Name: name,
						Path: filepath.ToSlash(childRel),
						Type: "file",
					},
					created: createdAt,
					updated: updatedAt,
				})
				continue
			}
			if isImage(name) {
				nodes = append(nodes, treeNodeSortEntry{
					node: TreeNode{
						Name: name,
						Path: filepath.ToSlash(childRel),
						Type: "asset",
					},
					created: createdAt,
					updated: updatedAt,
				})
				continue
			}
			if isPDF(name) {
				nodes = append(nodes, treeNodeSortEntry{
					node: TreeNode{
						Name: name,
						Path: filepath.ToSlash(childRel),
						Type: "pdf",
					},
					created: createdAt,
					updated: updatedAt,
				})
				continue
			}
			if isCSV(name) {
				nodes = append(nodes, treeNodeSortEntry{
					node: TreeNode{
						Name: name,
						Path: filepath.ToSlash(childRel),
						Type: "csv",
					},
					created: createdAt,
					updated: updatedAt,
				})
				continue
			}
			if isSheetFile(name) {
				continue
			}
			continue
		}

		nodes = append(nodes, treeNodeSortEntry{
			node: TreeNode{
				Name: name,
				Path: filepath.ToSlash(childRel),
				Type: "file",
			},
			created: createdAt,
			updated: updatedAt,
		})
	}

	isDailyFolder := isDailyPath(relPath)
	sort.Slice(nodes, func(i, j int) bool {
		if isDailyFolder {
			return compareTreeEntriesDaily(nodes[i], nodes[j])
		}
		return compareTreeEntries(nodes[i], nodes[j], sortBy, sortOrder)
	})

	result := make([]TreeNode, 0, len(nodes))
	for _, entry := range nodes {
		result = append(result, entry.node)
	}

	return result, nil
}

func (s *Server) resolvePath(input string) (string, string, error) {
	clean, err := cleanRelPath(input)
	if err != nil {
		return "", "", err
	}

	absPath := filepath.Join(s.notesDir, clean)
	relCheck, err := filepath.Rel(s.notesDir, absPath)
	if err != nil {
		return "", "", err
	}
	if relCheck == ".." || strings.HasPrefix(relCheck, ".."+string(os.PathSeparator)) {
		return "", "", errors.New("path escapes notes directory")
	}

	return absPath, filepath.ToSlash(clean), nil
}

func dailyNoteExists(rootDir, noteName string) (bool, error) {
	if noteName == "" {
		return false, nil
	}
	sentinel := errors.New("daily note found")
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != noteName {
			return nil
		}
		if !isMarkdown(d.Name()) {
			return nil
		}
		return sentinel
	})
	if err != nil {
		if errors.Is(err, sentinel) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (s *Server) ensureDailyNote() error {
	dailyDir := filepath.Join(s.notesDir, dailyFolderName)
	info, err := os.Stat(dailyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}

	today := timeNow().Format(dailyDateLayout)
	noteName := today + ".md"
	exists, err := dailyNoteExists(dailyDir, noteName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	notePath := filepath.Join(dailyDir, noteName)

	content, ok, err := s.folderTemplateContent(dailyDir)
	if err != nil {
		return err
	}
	if !ok {
		content = nil
	}
	finalContent := string(content)
	if ok {
		relPath := filepath.ToSlash(filepath.Join(dailyFolderName, today+".md"))
		targetTime := inferTargetTime(relPath, timeNow())
		conditioned, warnings := applyTemplateConditionals(finalContent, targetTime)
		finalContent = applyTemplatePlaceholders(conditioned, targetTime, templateContext(relPath))
		for _, warning := range warnings {
			s.logger.Warn("template condition warning", "path", relPath, "warning", warning)
		}
	}
	return os.WriteFile(notePath, []byte(finalContent), 0o644)
}

func (s *Server) ensureInboxNote() error {
	absPath, _, err := s.resolvePath(inboxNotePath)
	if err != nil {
		return err
	}
	info, err := os.Stat(absPath)
	if err == nil {
		if info.IsDir() {
			return nil
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(absPath, []byte(""), 0o644)
}

func (s *Server) folderTemplateContent(dir string) ([]byte, bool, error) {
	templatePath := filepath.Join(dir, "default.template")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return content, true, nil
}

func applyTemplatePlaceholders(input string, now time.Time, ctx TemplateContext) string {
	const tokenPrefix = "{{"
	const tokenSuffix = "}}"
	start := strings.Index(input, tokenPrefix)
	if start == -1 {
		return input
	}

	var out strings.Builder
	out.Grow(len(input))
	remaining := input
	for {
		start = strings.Index(remaining, tokenPrefix)
		if start == -1 {
			out.WriteString(remaining)
			break
		}
		out.WriteString(remaining[:start])
		remaining = remaining[start+len(tokenPrefix):]
		end := strings.Index(remaining, tokenSuffix)
		if end == -1 {
			out.WriteString(tokenPrefix)
			out.WriteString(remaining)
			break
		}
		token := remaining[:end]
		out.WriteString(resolveTemplateToken(token, now, ctx))
		remaining = remaining[end+len(tokenSuffix):]
	}
	return out.String()
}

func resolveTemplateToken(token string, now time.Time, ctx TemplateContext) string {
	if token == "title" {
		return ctx.Title
	}
	if token == "path" {
		return ctx.Path
	}
	if token == "folder" {
		return ctx.Folder
	}
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return "{{" + token + "}}"
	}
	key := parts[0]
	format := parts[1]
	layout := layoutFromTemplate(format)
	switch key {
	case "date", "time", "datetime", "day", "year", "month":
		return now.Format(layout)
	default:
		return "{{" + token + "}}"
	}
}

func layoutFromTemplate(format string) string {
	replacer := strings.NewReplacer(
		"YYYY", "2006",
		"MM", "01",
		"DD", "02",
		"HH", "15",
		"mm", "04",
		"ss", "05",
		"dddd", "Monday",
		"ddd", "Mon",
	)
	return replacer.Replace(format)
}

func templateContext(relPath string) TemplateContext {
	path := filepath.ToSlash(relPath)
	folder := filepath.ToSlash(filepath.Dir(relPath))
	if folder == "." {
		folder = ""
	}
	return TemplateContext{
		Title:  strings.TrimSuffix(filepath.Base(relPath), filepath.Ext(relPath)),
		Path:   path,
		Folder: folder,
	}
}

func inferTargetTime(relPath string, fallback time.Time) time.Time {
	base := strings.TrimSuffix(filepath.Base(relPath), filepath.Ext(relPath))
	parsed, err := time.ParseInLocation("2006-01-02", base, time.Local)
	if err != nil {
		return fallback
	}
	return time.Date(
		parsed.Year(),
		parsed.Month(),
		parsed.Day(),
		fallback.Hour(),
		fallback.Minute(),
		fallback.Second(),
		fallback.Nanosecond(),
		fallback.Location(),
	)
}

func applyTemplateConditionals(input string, target time.Time) (string, []string) {
	if input == "" {
		return input, nil
	}
	lines := strings.Split(input, "\n")
	output := make([]string, 0, len(lines))
	warnings := make([]string, 0)

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		lineEnding := ""
		lineContent := line
		if strings.HasSuffix(lineContent, "\r") {
			lineEnding = "\r"
			lineContent = strings.TrimSuffix(lineContent, "\r")
		}
		trimmedLeft := strings.TrimLeft(lineContent, " \t")
		if strings.HasPrefix(trimmedLeft, "{{if:") {
			end := strings.Index(trimmedLeft, "}}")
			if end != -1 {
				expr := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmedLeft[:end+2], "{{if:"), "}}"))
				if strings.TrimSpace(trimmedLeft[end+2:]) == "" {
					// Block form.
					conditionMet, warn := evaluateTemplateCondition(expr, target)
					if warn != "" {
						warnings = append(warnings, fmt.Sprintf("Invalid template condition %q: %s", expr, warn))
						conditionMet = false
					}
					foundEnd := false
					for j := i + 1; j < len(lines); j++ {
						blockLine := strings.TrimSpace(strings.TrimSuffix(lines[j], "\r"))
						if blockLine == "{{endif}}" {
							foundEnd = true
							i = j
							break
						}
						if conditionMet {
							output = append(output, lines[j])
						}
					}
					if !foundEnd {
						warnings = append(warnings, "Invalid template condition block: missing {{endif}}.")
						break
					}
					continue
				}

				// Inline form.
				conditionMet, warn := evaluateTemplateCondition(expr, target)
				if warn != "" {
					warnings = append(warnings, fmt.Sprintf("Invalid template condition %q: %s", expr, warn))
					continue
				}
				if conditionMet {
					leading := lineContent[:len(lineContent)-len(strings.TrimLeft(lineContent, " \t"))]
					rest := strings.TrimPrefix(strings.TrimLeft(lineContent, " \t"), trimmedLeft[:end+2])
					output = append(output, leading+rest+lineEnding)
				}
				continue
			}
		}

		output = append(output, line)
	}

	result := strings.Join(output, "\n")
	return result, warnings
}

func evaluateTemplateCondition(expr string, target time.Time) (bool, string) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false, "empty expression"
	}
	if strings.Contains(expr, "&") && !strings.Contains(expr, "=") {
		return false, "invalid format; use field=value and '|' for multiple values"
	}
	terms := strings.Split(expr, "&")
	seenFields := make(map[string]struct{})
	parsed := make([]templateConditionTerm, 0, len(terms))

	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" {
			return false, "empty term"
		}
		if !strings.Contains(term, "=") {
			return false, "invalid format; use field=value and '|' for multiple values"
		}
		parts := strings.SplitN(term, "=", 2)
		if len(parts) != 2 {
			return false, "missing '='"
		}
		field := strings.ToLower(strings.TrimSpace(parts[0]))
		rawValues := strings.TrimSpace(parts[1])
		if field == "" || rawValues == "" {
			return false, "missing field or value"
		}
		if strings.Contains(rawValues, "&") {
			return false, "invalid format; use '|' for multiple values within a field"
		}
		if _, ok := seenFields[field]; ok {
			return false, "duplicate field; use '|' for multiple values"
		}
		seenFields[field] = struct{}{}

		values := splitAndTrim(rawValues, "|")
		if len(values) == 0 {
			return false, "missing value"
		}

		if err := validateTemplateConditionField(field, values); err != nil {
			return false, err.Error()
		}

		parsed = append(parsed, templateConditionTerm{
			field:  field,
			values: values,
		})
	}

	for _, term := range parsed {
		matched, err := evaluateTemplateConditionField(term.field, term.values, target)
		if err != nil {
			return false, err.Error()
		}
		if !matched {
			return false, ""
		}
	}

	return true, ""
}

type templateConditionTerm struct {
	field  string
	values []string
}

func splitAndTrim(input, sep string) []string {
	parts := strings.Split(input, sep)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func validateTemplateConditionField(field string, values []string) error {
	switch field {
	case "day":
		for _, value := range values {
			if isWeekdayLabel(value) || isWeekendLabel(value) {
				continue
			}
			if _, ok := parseWeekday(value); !ok {
				return fmt.Errorf("invalid day %q", value)
			}
		}
		return nil
	case "dom":
		for _, value := range values {
			day, err := strconv.Atoi(value)
			if err != nil || day < 1 || day > 31 {
				return fmt.Errorf("invalid day of month %q", value)
			}
		}
		return nil
	case "date":
		for _, value := range values {
			if _, err := time.ParseInLocation("2006-01-02", value, time.Local); err != nil {
				return fmt.Errorf("invalid date %q", value)
			}
		}
		return nil
	case "month":
		for _, value := range values {
			if _, ok := parseMonth(value); !ok {
				return fmt.Errorf("invalid month %q", value)
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown field %q", field)
	}
}

func evaluateTemplateConditionField(field string, values []string, target time.Time) (bool, error) {
	switch field {
	case "day":
		for _, value := range values {
			if isWeekdayLabel(value) {
				if isWeekdayValue(value, target.Weekday()) {
					return true, nil
				}
				continue
			}
			if isWeekendLabel(value) {
				if isWeekendValue(value, target.Weekday()) {
					return true, nil
				}
				continue
			}
			weekday, ok := parseWeekday(value)
			if !ok {
				return false, fmt.Errorf("invalid day %q", value)
			}
			if target.Weekday() == weekday {
				return true, nil
			}
		}
		return false, nil
	case "dom":
		for _, value := range values {
			day, err := strconv.Atoi(value)
			if err != nil || day < 1 || day > 31 {
				return false, fmt.Errorf("invalid day of month %q", value)
			}
			if target.Day() == day {
				return true, nil
			}
		}
		return false, nil
	case "date":
		for _, value := range values {
			parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
			if err != nil {
				return false, fmt.Errorf("invalid date %q", value)
			}
			if sameDay(target, parsed) {
				return true, nil
			}
		}
		return false, nil
	case "month":
		for _, value := range values {
			month, ok := parseMonth(value)
			if !ok {
				return false, fmt.Errorf("invalid month %q", value)
			}
			if target.Month() == month {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("unknown field %q", field)
	}
}

func parseWeekday(value string) (time.Weekday, bool) {
	key := strings.ToLower(strings.TrimSpace(value))
	switch key {
	case "mon", "monday":
		return time.Monday, true
	case "tue", "tues", "tuesday":
		return time.Tuesday, true
	case "wed", "wednesday":
		return time.Wednesday, true
	case "thu", "thur", "thurs", "thursday":
		return time.Thursday, true
	case "fri", "friday":
		return time.Friday, true
	case "sat", "saturday":
		return time.Saturday, true
	case "sun", "sunday":
		return time.Sunday, true
	default:
		return time.Sunday, false
	}
}

func isWeekdayLabel(value string) bool {
	key := strings.ToLower(strings.TrimSpace(value))
	return key == "weekday" || key == "weekdays"
}

func isWeekendLabel(value string) bool {
	key := strings.ToLower(strings.TrimSpace(value))
	return key == "weekend" || key == "weekends"
}

func isWeekdayValue(value string, weekday time.Weekday) bool {
	if !isWeekdayLabel(value) {
		return false
	}
	return weekday != time.Saturday && weekday != time.Sunday
}

func isWeekendValue(value string, weekday time.Weekday) bool {
	if !isWeekendLabel(value) {
		return false
	}
	return weekday == time.Saturday || weekday == time.Sunday
}

func parseMonth(value string) (time.Month, bool) {
	key := strings.ToLower(strings.TrimSpace(value))
	if num, err := strconv.Atoi(key); err == nil {
		if num >= 1 && num <= 12 {
			return time.Month(num), true
		}
		return time.January, false
	}
	switch key {
	case "jan", "january":
		return time.January, true
	case "feb", "february":
		return time.February, true
	case "mar", "march":
		return time.March, true
	case "apr", "april":
		return time.April, true
	case "may":
		return time.May, true
	case "jun", "june":
		return time.June, true
	case "jul", "july":
		return time.July, true
	case "aug", "august":
		return time.August, true
	case "sep", "sept", "september":
		return time.September, true
	case "oct", "october":
		return time.October, true
	case "nov", "november":
		return time.November, true
	case "dec", "december":
		return time.December, true
	default:
		return time.January, false
	}
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func cleanRelPath(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", nil
	}
	clean := filepath.Clean(filepath.FromSlash(trimmed))
	if clean == "." {
		return "", nil
	}
	if filepath.IsAbs(clean) {
		return "", errors.New("absolute paths are not allowed")
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", errors.New("path escapes notes directory")
	}

	return clean, nil
}

func validateReservedRootPath(relPath string, isFolder bool) error {
	trimmed := strings.TrimSpace(filepath.ToSlash(relPath))
	if trimmed == "" {
		return nil
	}
	if strings.Contains(trimmed, "/") {
		return nil
	}
	name := strings.ToLower(trimmed)
	if isFolder {
		if isReservedRootFolder(name) {
			return errors.New("folder name is reserved")
		}
		return nil
	}
	if isReservedRootFile(name) {
		return errors.New("file name is reserved")
	}
	return nil
}

func isReservedRootFolder(name string) bool {
	switch name {
	case strings.ToLower(dailyFolderName),
		strings.ToLower(sheetsFolderName),
		strings.ToLower(journalFolderName),
		strings.ToLower(emailTemplatesDir):
		return true
	default:
		return false
	}
}

func isReservedRootFile(name string) bool {
	switch name {
	case strings.ToLower(settingsFileName),
		strings.ToLower(emailSettingsFileName),
		strings.ToLower(taskFiltersFileName),
		strings.ToLower(inboxNotePath),
		strings.ToLower(scratchNotePath):
		return true
	default:
		return false
	}
}

func isDailyPath(relPath string) bool {
	trimmed := strings.TrimSpace(relPath)
	if trimmed == "" {
		return false
	}
	lower := strings.ToLower(trimmed)
	dailyLower := strings.ToLower(dailyFolderName)
	return lower == dailyLower || strings.HasPrefix(lower, dailyLower+"/")
}

func ensureMarkdown(path string) string {
	if strings.HasSuffix(strings.ToLower(path), ".md") {
		return path
	}
	return path + ".md"
}

func ensureTemplate(path string) string {
	if isTemplate(path) {
		return path
	}
	return path + ".template"
}

func isMarkdown(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".md")
}

func isTemplate(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".template")
}

func isNoteFile(name string) bool {
	return isMarkdown(name) || isTemplate(name)
}

func isIgnoredFile(name string) bool {
	return strings.HasPrefix(name, "._")
}

func isImage(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".bmp", ".tif", ".tiff", ".avif", ".heic":
		return true
	default:
		return false
	}
}

func isPDF(name string) bool {
	return strings.EqualFold(filepath.Ext(name), ".pdf")
}

func isCSV(name string) bool {
	return strings.EqualFold(filepath.Ext(name), ".csv")
}

func isSheetFile(name string) bool {
	return strings.EqualFold(filepath.Ext(name), sheetExtension)
}

func ensureSheetExtension(path string) string {
	if strings.HasSuffix(strings.ToLower(path), sheetExtension) {
		return path
	}
	return path + sheetExtension
}

func tagsContain(tags []string, query string) bool {
	if query == "" {
		return false
	}
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

func decodeJSON[T any](reader io.Reader) (T, error) {
	var payload T
	dec := json.NewDecoder(reader)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&payload); err != nil {
		return payload, err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return payload, errors.New("unexpected extra data in request body")
	}
	return payload, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	logger := slog.Default().With("component", "api", "status", status)
	switch {
	case status >= http.StatusInternalServerError:
		logger.Error("request error", "message", message)
	case status >= http.StatusBadRequest:
		logger.Warn("request error", "message", message)
	default:
		logger.Info("request error", "message", message)
	}
	writeJSON(w, status, map[string]string{"error": message})
}
