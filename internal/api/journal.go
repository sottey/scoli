package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const journalFolderName = "journal"
const journalFileName = "journal.json"

type JournalEntry struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type JournalListResponse struct {
	Entries []JournalEntry `json:"entries"`
}

type JournalCreatePayload struct {
	Content string `json:"content"`
}

type JournalUpdatePayload struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type JournalArchivePayload struct {
	ID string `json:"id"`
}

type JournalArchiveResponse struct {
	Status      string `json:"status"`
	ArchivePath string `json:"archivePath"`
}

func (s *Server) journalDirPath() string {
	return filepath.Join(s.notesDir, journalFolderName)
}

func (s *Server) journalFilePath() string {
	return filepath.Join(s.journalDirPath(), journalFileName)
}

func (s *Server) handleJournalList(w http.ResponseWriter, r *http.Request) {
	entries, err := s.loadJournalEntries()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load journal")
		return
	}
	writeJSON(w, http.StatusOK, JournalListResponse{Entries: entries})
}

func (s *Server) handleJournalCreate(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[JournalCreatePayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(payload.Content) == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	entries, err := s.loadJournalEntries()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load journal")
		return
	}

	now := time.Now()
	entry := JournalEntry{
		ID:        fmt.Sprintf("%d", now.UnixNano()),
		Content:   payload.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	entries = append(entries, entry)
	sortJournalEntries(entries)

	if err := s.saveJournalEntries(s.journalFilePath(), entries); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save journal")
		return
	}

	writeJSON(w, http.StatusCreated, entry)
}

func (s *Server) handleJournalUpdate(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[JournalUpdatePayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	entries, err := s.loadJournalEntries()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load journal")
		return
	}

	found := false
	for i := range entries {
		if entries[i].ID == payload.ID {
			entries[i].Content = payload.Content
			entries[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}
	if !found {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}
	sortJournalEntries(entries)

	if err := s.saveJournalEntries(s.journalFilePath(), entries); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save journal")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) handleJournalDelete(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	entries, err := s.loadJournalEntries()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load journal")
		return
	}

	next := entries[:0]
	removed := false
	for _, entry := range entries {
		if entry.ID == id {
			removed = true
			continue
		}
		next = append(next, entry)
	}
	if !removed {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}

	if err := s.saveJournalEntries(s.journalFilePath(), next); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save journal")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleJournalArchive(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[JournalArchivePayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(payload.ID) == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	entries, err := s.loadJournalEntries()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load journal")
		return
	}

	var archived *JournalEntry
	next := entries[:0]
	for _, entry := range entries {
		if entry.ID == payload.ID {
			copyEntry := entry
			archived = &copyEntry
			continue
		}
		next = append(next, entry)
	}
	if archived == nil {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}

	archiveDate := time.Now().Format("2006-01-02")
	archiveName := fmt.Sprintf("journal-archive-%s.json", archiveDate)
	archivePath := filepath.Join(s.journalDirPath(), archiveName)

	archiveEntries, err := loadJournalEntriesFromPath(archivePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load archive")
		return
	}
	archiveEntries = append(archiveEntries, *archived)
	sortJournalEntries(archiveEntries)

	if err := s.saveJournalEntries(archivePath, archiveEntries); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save archive")
		return
	}

	if err := s.saveJournalEntries(s.journalFilePath(), next); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save journal")
		return
	}

	writeJSON(w, http.StatusOK, JournalArchiveResponse{
		Status:      "archived",
		ArchivePath: filepath.ToSlash(filepath.Join(journalFolderName, archiveName)),
	})
}

func (s *Server) handleJournalArchiveAll(w http.ResponseWriter, r *http.Request) {
	entries, err := s.loadJournalEntries()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load journal")
		return
	}
	if len(entries) == 0 {
		writeJSON(w, http.StatusOK, map[string]string{"status": "empty"})
		return
	}

	archiveDate := time.Now().Format("2006-01-02")
	archiveName := fmt.Sprintf("journal-archive-%s.json", archiveDate)
	archivePath := filepath.Join(s.journalDirPath(), archiveName)

	archiveEntries, err := loadJournalEntriesFromPath(archivePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load archive")
		return
	}
	archiveEntries = append(archiveEntries, entries...)
	sortJournalEntries(archiveEntries)

	if err := s.saveJournalEntries(archivePath, archiveEntries); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save archive")
		return
	}

	if err := s.saveJournalEntries(s.journalFilePath(), []JournalEntry{}); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save journal")
		return
	}

	writeJSON(w, http.StatusOK, JournalArchiveResponse{
		Status:      "archived",
		ArchivePath: filepath.ToSlash(filepath.Join(journalFolderName, archiveName)),
	})
}

func (s *Server) handleJournalArchiveList(w http.ResponseWriter, r *http.Request) {
	if err := os.MkdirAll(s.journalDirPath(), 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to prepare journal folder")
		return
	}
	entries, err := os.ReadDir(s.journalDirPath())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to read journal folder")
		return
	}

	type archiveItem struct {
		Date  string `json:"date"`
		Path  string `json:"path"`
		Count int    `json:"count"`
	}

	var archives []archiveItem
	archiveCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "journal-archive-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		date := strings.TrimSuffix(strings.TrimPrefix(name, "journal-archive-"), ".json")
		if date == "" {
			continue
		}
		archivePath := filepath.Join(s.journalDirPath(), name)
		archiveEntries, err := loadJournalEntriesFromPath(archivePath)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "unable to load archive")
			return
		}
		entryCount := len(archiveEntries)
		archiveCount += entryCount
		archives = append(archives, archiveItem{
			Date:  date,
			Path:  filepath.ToSlash(filepath.Join(journalFolderName, name)),
			Count: entryCount,
		})
	}

	sort.Slice(archives, func(i, j int) bool {
		return archives[i].Date > archives[j].Date
	})

	activeEntries, err := s.loadJournalEntries()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load journal")
		return
	}
	activeCount := len(activeEntries)

	writeJSON(w, http.StatusOK, map[string]any{
		"archives":     archives,
		"activeCount":  activeCount,
		"archiveCount": archiveCount,
		"totalCount":   activeCount + archiveCount,
	})
}

func (s *Server) handleJournalArchiveGet(w http.ResponseWriter, r *http.Request) {
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		writeError(w, http.StatusBadRequest, "date is required")
		return
	}
	name := fmt.Sprintf("journal-archive-%s.json", date)
	path := filepath.Join(s.journalDirPath(), name)
	entries, err := loadJournalEntriesFromPath(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load archive")
		return
	}
	writeJSON(w, http.StatusOK, JournalListResponse{Entries: entries})
}

func (s *Server) loadJournalEntries() ([]JournalEntry, error) {
	return loadJournalEntriesFromPath(s.journalFilePath())
}

func loadJournalEntriesFromPath(path string) ([]JournalEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []JournalEntry{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []JournalEntry{}, nil
	}
	var entries []JournalEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *Server) saveJournalEntries(path string, entries []JournalEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	sortJournalEntries(entries)
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func sortJournalEntries(entries []JournalEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CreatedAt.After(entries[j].CreatedAt)
	})
}
