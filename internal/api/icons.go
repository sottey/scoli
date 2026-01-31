package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxRootIconSize = 1 << 20 // 1MB

var allowedRootIconKeys = map[string]struct{}{
	"notes":    {},
	"daily":    {},
	"tasks":    {},
	"tags":     {},
	"mentions": {},
	"journal":  {},
	"inbox":    {},
	"sheets":   {},
	"ai":       {},
}

func (s *Server) handleRootIconUpload(w http.ResponseWriter, r *http.Request) {
	rootKey := strings.TrimSpace(r.URL.Query().Get("root"))
	if _, ok := allowedRootIconKeys[rootKey]; !ok {
		writeError(w, http.StatusBadRequest, "invalid root icon target")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRootIconSize+1024)
	if err := r.ParseMultipartForm(maxRootIconSize + 1024); err != nil {
		writeError(w, http.StatusBadRequest, "invalid icon upload")
		return
	}

	file, header, err := r.FormFile("icon")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing icon file")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !isAllowedIconExt(ext) {
		writeError(w, http.StatusBadRequest, "icon must be png, svg, or ico")
		return
	}

	filename, err := generateIconFilename(rootKey, ext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to create icon filename")
		return
	}

	iconsDir := filepath.Join("internal", "ui", "web", "icons")
	if err := os.MkdirAll(iconsDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save icon")
		return
	}

	targetPath := filepath.Join(iconsDir, filename)
	if err := writeIconFile(targetPath, file); err != nil {
		if errors.Is(err, errIconTooLarge) {
			writeError(w, http.StatusBadRequest, "icon must be 1MB or smaller")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to save icon")
		return
	}

	settings, _, err := s.loadSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to update settings")
		return
	}
	if settings.RootIcons == nil {
		settings.RootIcons = map[string]string{}
	}
	settings.RootIcons[rootKey] = "/icons/" + filename
	if err := s.saveSettings(settings); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to update settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"root": rootKey,
		"path": "/icons/" + filename,
	})
}

func (s *Server) handleRootIconReset(w http.ResponseWriter, r *http.Request) {
	rootKey := strings.TrimSpace(r.URL.Query().Get("root"))
	if _, ok := allowedRootIconKeys[rootKey]; !ok {
		writeError(w, http.StatusBadRequest, "invalid root icon target")
		return
	}

	settings, _, err := s.loadSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to update settings")
		return
	}
	if settings.RootIcons == nil {
		settings.RootIcons = map[string]string{}
	}
	iconPath := settings.RootIcons[rootKey]
	delete(settings.RootIcons, rootKey)
	if err := s.saveSettings(settings); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to update settings")
		return
	}

	if iconPath != "" {
		removeIconFile(iconPath)
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"root": rootKey,
	})
}

var errIconTooLarge = errors.New("icon too large")

func writeIconFile(path string, src io.Reader) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	limited := &io.LimitedReader{R: src, N: maxRootIconSize + 1}
	written, err := io.Copy(out, limited)
	if err != nil {
		return err
	}
	if written > maxRootIconSize {
		return errIconTooLarge
	}
	return nil
}

func isAllowedIconExt(ext string) bool {
	switch ext {
	case ".png", ".svg", ".ico":
		return true
	default:
		return false
	}
}

func generateIconFilename(rootKey, ext string) (string, error) {
	nonce := make([]byte, 8)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	timestamp := time.Now().UTC().Format("20060102-150405")
	return rootKey + "-" + timestamp + "-" + hex.EncodeToString(nonce) + ext, nil
}

func removeIconFile(iconPath string) {
	if !strings.HasPrefix(iconPath, "/icons/") {
		return
	}
	filename := strings.TrimPrefix(iconPath, "/icons/")
	if filename == "" {
		return
	}
	iconsDir := filepath.Join("internal", "ui", "web", "icons")
	target := filepath.Join(iconsDir, filename)
	cleanDir := filepath.Clean(iconsDir) + string(filepath.Separator)
	if !strings.HasPrefix(filepath.Clean(target), cleanDir) {
		return
	}
	_ = os.Remove(target)
}
