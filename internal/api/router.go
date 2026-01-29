package api

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
)

func NewRouter(notesDir string, logger ...*slog.Logger) chi.Router {
	var baseLogger *slog.Logger
	if len(logger) > 0 && logger[0] != nil {
		baseLogger = logger[0]
	} else {
		baseLogger = slog.Default()
	}
	s := &Server{
		notesDir: notesDir,
		logger:   baseLogger.With("component", "api"),
	}

	r := chi.NewRouter()
	r.Get("/health", s.handleHealth)
	r.Get("/tree", s.handleTree)
	r.Get("/notes", s.handleGetNote)
	r.Post("/notes", s.handleCreateNote)
	r.Patch("/notes", s.handleUpdateNote)
	r.Patch("/notes/rename", s.handleRenameNote)
	r.Delete("/notes", s.handleDeleteNote)
	r.Get("/files", s.handleGetFile)
	r.Get("/search", s.handleSearch)
	r.Get("/journal", s.handleJournalList)
	r.Post("/journal", s.handleJournalCreate)
	r.Patch("/journal", s.handleJournalUpdate)
	r.Delete("/journal", s.handleJournalDelete)
	r.Post("/journal/archive", s.handleJournalArchive)
	r.Post("/journal/archive-all", s.handleJournalArchiveAll)
	r.Get("/journal/archives", s.handleJournalArchiveList)
	r.Get("/journal/archive", s.handleJournalArchiveGet)
	r.Get("/tags", s.handleTags)
	r.Get("/mentions", s.handleMentions)
	r.Get("/settings", s.handleSettingsGet)
	r.Patch("/settings", s.handleSettingsUpdate)
	r.Post("/icons/root", s.handleRootIconUpload)
	r.Delete("/icons/root", s.handleRootIconReset)
	r.Post("/folders", s.handleCreateFolder)
	r.Patch("/folders", s.handleRenameFolder)
	r.Delete("/folders", s.handleDeleteFolder)
	r.Get("/tasks", s.handleTasksList)
	r.Get("/tasks/for-note", s.handleTasksForNote)
	r.Get("/tasks/filters", s.handleTaskFiltersGet)
	r.Put("/tasks/filters", s.handleTaskFiltersUpdate)
	r.Patch("/tasks/toggle", s.handleTasksToggle)
	r.Patch("/tasks/archive", s.handleTasksArchive)
	r.Get("/sheets/tree", s.handleSheetsTree)
	r.Get("/sheets", s.handleSheetsGet)
	r.Post("/sheets", s.handleSheetsCreate)
	r.Patch("/sheets", s.handleSheetsUpdate)
	r.Patch("/sheets/rename", s.handleSheetsRename)
	r.Delete("/sheets", s.handleSheetsDelete)
	r.Post("/sheets/import", s.handleSheetsImport)
	r.Get("/sheets/export", s.handleSheetsExport)

	return r
}
