package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"

	"github.com/sottey/scoli/internal/api"
	"github.com/sottey/scoli/internal/ui"
)

type Config struct {
	NotesDir string
	SeedDir  string
	Port     int
	LogLevel string
}

func Run(cfg Config) error {
	if cfg.Port <= 0 {
		return fmt.Errorf("port must be positive")
	}

	notesDir, err := filepath.Abs(cfg.NotesDir)
	if err != nil {
		return fmt.Errorf("resolve notes dir: %w", err)
	}

	seedDir := ""
	if cfg.SeedDir != "" {
		seedDir, err = filepath.Abs(cfg.SeedDir)
		if err != nil {
			return fmt.Errorf("resolve seed dir: %w", err)
		}
	}

	if err := os.MkdirAll(notesDir, 0o755); err != nil {
		return fmt.Errorf("ensure notes dir: %w", err)
	}

	level, ok := parseLogLevel(cfg.LogLevel)
	levelVar := new(slog.LevelVar)
	levelVar.Set(level)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: levelVar}))
	slog.SetDefault(logger)
	if !ok && cfg.LogLevel != "" {
		logger.Warn("unknown log level, defaulting to info", "level", cfg.LogLevel)
	}

	if seedDir != "" {
		if err := seedNotesIfEmpty(notesDir, seedDir, logger); err != nil {
			return err
		}
	}

	logger.Info("server starting", "notesDir", notesDir, "port", cfg.Port)

	r := chi.NewRouter()
	r.Use(requestLogger)
	r.Mount("/api/v1", api.NewRouter(notesDir))
	r.Mount("/", ui.NewRouter())

	addr := fmt.Sprintf(":%d", cfg.Port)
	return listenAndServe(addr, r)
}

var listenAndServe = http.ListenAndServe
