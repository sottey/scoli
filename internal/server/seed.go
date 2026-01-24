package server

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func seedNotesIfEmpty(notesDir, seedDir string, logger *slog.Logger) error {
	info, err := os.Stat(seedDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("seed dir missing; skipping seeding", "seedDir", seedDir)
			return nil
		}
		return fmt.Errorf("stat seed dir: %w", err)
	}
	if !info.IsDir() {
		logger.Info("seed dir is not a directory; skipping seeding", "seedDir", seedDir)
		return nil
	}

	entries, err := os.ReadDir(notesDir)
	if err != nil {
		return fmt.Errorf("read notes dir: %w", err)
	}
	hasContent := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		hasContent = true
		break
	}
	if hasContent {
		logger.Info("notes dir not empty; skipping seeding", "notesDir", notesDir)
		return nil
	}

	if err := filepath.WalkDir(seedDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == seedDir {
			return nil
		}

		relPath, err := filepath.Rel(seedDir, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(notesDir, relPath)
		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}
		return copyFile(path, destPath, fileInfo.Mode())
	}); err != nil {
		return fmt.Errorf("copy seed notes: %w", err)
	}

	logger.Info("seeded notes directory", "notesDir", notesDir, "seedDir", seedDir)
	return nil
}

func copyFile(src, dest string, mode fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
