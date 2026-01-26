package api

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type SheetPayload struct {
	Path string     `json:"path"`
	Data [][]string `json:"data"`
}

type SheetRenamePayload struct {
	Path    string `json:"path"`
	NewPath string `json:"newPath"`
}

type SheetImportPayload struct {
	Path string `json:"path"`
	CSV  string `json:"csv"`
}

type SheetResponse struct {
	Path     string     `json:"path"`
	Data     [][]string `json:"data"`
	Modified time.Time  `json:"modified"`
}

type sheetFile struct {
	Data [][]string `json:"data"`
}

func (s *Server) handleSheetsTree(w http.ResponseWriter, r *http.Request) {
	root := TreeNode{
		Name: "Sheets",
		Path: "",
		Type: "folder",
	}

	sheetsDir := filepath.Join(s.notesDir, sheetsFolderName)
	info, err := os.Stat(sheetsDir)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusOK, root)
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to read sheets folder")
		return
	}
	if !info.IsDir() {
		writeError(w, http.StatusInternalServerError, "sheets path is not a folder")
		return
	}

	children, err := s.buildSheetsTree(sheetsDir, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to build sheets tree")
		return
	}
	root.Children = children
	writeJSON(w, http.StatusOK, root)
}

func (s *Server) handleSheetsGet(w http.ResponseWriter, r *http.Request) {
	pathParam := strings.TrimSpace(r.URL.Query().Get("path"))
	if pathParam == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}
	pathParam = ensureSheetExtension(pathParam)
	absPath, relPath, err := s.resolveSheetPath(pathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "sheet not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to read sheet")
		return
	}
	if info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is a folder")
		return
	}
	if !isSheetFile(absPath) {
		writeError(w, http.StatusBadRequest, "not a sheet file")
		return
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to read sheet")
		return
	}
	sheet, err := decodeSheetFile(data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to parse sheet data")
		return
	}

	resp := SheetResponse{
		Path:     relPath,
		Data:     sheet.Data,
		Modified: info.ModTime(),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleSheetsCreate(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[SheetPayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(payload.Path) == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	pathParam := ensureSheetExtension(strings.TrimSpace(payload.Path))
	absPath, relPath, err := s.resolveSheetPath(pathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := os.Stat(absPath); err == nil {
		writeError(w, http.StatusConflict, "sheet already exists")
		return
	} else if !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "unable to check sheet")
		return
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to create parent folders")
		return
	}

	if err := writeSheetFile(absPath, payload.Data); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to write sheet")
		return
	}

	s.logger.Info("sheet created", "path", relPath)
	writeJSON(w, http.StatusOK, map[string]string{"path": relPath})
}

func (s *Server) handleSheetsUpdate(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[SheetPayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(payload.Path) == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	pathParam := ensureSheetExtension(strings.TrimSpace(payload.Path))
	absPath, relPath, err := s.resolveSheetPath(pathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "sheet not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to read sheet")
		return
	}
	if info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is a folder")
		return
	}
	if !isSheetFile(absPath) {
		writeError(w, http.StatusBadRequest, "not a sheet file")
		return
	}

	if err := writeSheetFile(absPath, payload.Data); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to update sheet")
		return
	}

	s.logger.Info("sheet updated", "path", relPath)
	writeJSON(w, http.StatusOK, map[string]string{"path": relPath})
}

func (s *Server) handleSheetsRename(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[SheetRenamePayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(payload.Path) == "" || strings.TrimSpace(payload.NewPath) == "" {
		writeError(w, http.StatusBadRequest, "path and newPath are required")
		return
	}

	pathParam := ensureSheetExtension(strings.TrimSpace(payload.Path))
	newPathParam := ensureSheetExtension(strings.TrimSpace(payload.NewPath))
	absPath, relPath, err := s.resolveSheetPath(pathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	newAbsPath, newRelPath, err := s.resolveSheetPath(newPathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "sheet not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to read sheet")
		return
	}

	if _, err := os.Stat(newAbsPath); err == nil {
		writeError(w, http.StatusConflict, "sheet already exists")
		return
	} else if !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "unable to check destination")
		return
	}

	if err := os.MkdirAll(filepath.Dir(newAbsPath), 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to create parent folders")
		return
	}

	if err := os.Rename(absPath, newAbsPath); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to rename sheet")
		return
	}

	s.logger.Info("sheet renamed", "path", relPath, "newPath", newRelPath)
	writeJSON(w, http.StatusOK, map[string]string{"path": relPath, "newPath": newRelPath})
}

func (s *Server) handleSheetsDelete(w http.ResponseWriter, r *http.Request) {
	pathParam := strings.TrimSpace(r.URL.Query().Get("path"))
	if pathParam == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	pathParam = ensureSheetExtension(pathParam)
	absPath, relPath, err := s.resolveSheetPath(pathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "sheet not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to read sheet")
		return
	}
	if info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is a folder")
		return
	}
	if !isSheetFile(absPath) {
		writeError(w, http.StatusBadRequest, "not a sheet file")
		return
	}

	if err := os.Remove(absPath); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to delete sheet")
		return
	}

	s.logger.Info("sheet deleted", "path", relPath)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleSheetsImport(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[SheetImportPayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(payload.Path) == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	reader := csv.NewReader(strings.NewReader(payload.CSV))
	records, err := reader.ReadAll()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid CSV data")
		return
	}

	pathParam := ensureSheetExtension(strings.TrimSpace(payload.Path))
	absPath, relPath, err := s.resolveSheetPath(pathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := os.Stat(absPath); err == nil {
		writeError(w, http.StatusConflict, "sheet already exists")
		return
	} else if !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "unable to check sheet")
		return
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to create parent folders")
		return
	}

	if err := writeSheetFile(absPath, records); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to write sheet")
		return
	}

	s.logger.Info("sheet imported", "path", relPath)
	writeJSON(w, http.StatusOK, map[string]string{"path": relPath})
}

func (s *Server) handleSheetsExport(w http.ResponseWriter, r *http.Request) {
	pathParam := strings.TrimSpace(r.URL.Query().Get("path"))
	if pathParam == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}
	pathParam = ensureSheetExtension(pathParam)
	absPath, relPath, err := s.resolveSheetPath(pathParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "sheet not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to read sheet")
		return
	}
	if info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is a folder")
		return
	}
	if !isSheetFile(absPath) {
		writeError(w, http.StatusBadRequest, "not a sheet file")
		return
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to read sheet")
		return
	}
	sheet, err := decodeSheetFile(data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to parse sheet data")
		return
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.WriteAll(sheet.Data); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to export sheet")
		return
	}

	filename := strings.TrimSuffix(filepath.Base(relPath), sheetExtension) + ".csv"
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	_, _ = w.Write(buf.Bytes())
}

func (s *Server) buildSheetsTree(absPath, relPath string) ([]TreeNode, error) {
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	type entry struct {
		node TreeNode
	}
	nodes := make([]entry, 0, len(entries))
	for _, item := range entries {
		name := item.Name()
		if isIgnoredFile(name) {
			continue
		}
		childRel := filepath.Join(relPath, name)
		childAbs := filepath.Join(absPath, name)
		if item.IsDir() {
			children, err := s.buildSheetsTree(childAbs, childRel)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, entry{
				node: TreeNode{
					Name:     name,
					Path:     filepath.ToSlash(childRel),
					Type:     "folder",
					Children: children,
				},
			})
			continue
		}
		if isSheetFile(name) {
			nodes = append(nodes, entry{
				node: TreeNode{
					Name: name,
					Path: filepath.ToSlash(childRel),
					Type: "sheet",
				},
			})
		}
	}

	sort.Slice(nodes, func(i, j int) bool {
		typeOrder := map[string]int{
			"folder": 0,
			"sheet":  1,
		}
		a := nodes[i].node
		b := nodes[j].node
		if a.Type != b.Type {
			return typeOrder[a.Type] < typeOrder[b.Type]
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})

	out := make([]TreeNode, 0, len(nodes))
	for _, entry := range nodes {
		out = append(out, entry.node)
	}
	return out, nil
}

func (s *Server) resolveSheetPath(input string) (string, string, error) {
	clean, err := cleanRelPath(input)
	if err != nil {
		return "", "", err
	}

	lower := strings.ToLower(clean)
	prefix := strings.ToLower(sheetsFolderName) + string(os.PathSeparator)
	if lower == strings.ToLower(sheetsFolderName) {
		clean = ""
	} else if strings.HasPrefix(lower, prefix) {
		clean = clean[len(prefix):]
	}

	baseDir := filepath.Join(s.notesDir, sheetsFolderName)
	absPath := filepath.Join(baseDir, clean)
	relCheck, err := filepath.Rel(baseDir, absPath)
	if err != nil {
		return "", "", err
	}
	if relCheck == ".." || strings.HasPrefix(relCheck, ".."+string(os.PathSeparator)) {
		return "", "", errors.New("path escapes sheets directory")
	}

	return absPath, filepath.ToSlash(clean), nil
}

func writeSheetFile(path string, data [][]string) error {
	payload := sheetFile{Data: normalizeSheetData(data)}
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	return os.WriteFile(path, encoded, 0o644)
}

func decodeSheetFile(data []byte) (sheetFile, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return sheetFile{Data: [][]string{}}, nil
	}
	var parsed sheetFile
	if err := json.Unmarshal(data, &parsed); err != nil {
		return sheetFile{}, err
	}
	if parsed.Data == nil {
		parsed.Data = [][]string{}
	}
	return sheetFile{Data: normalizeSheetData(parsed.Data)}, nil
}

func normalizeSheetData(data [][]string) [][]string {
	if data == nil {
		return [][]string{}
	}
	maxCols := 0
	for _, row := range data {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	if maxCols == 0 {
		return data
	}
	out := make([][]string, len(data))
	for i, row := range data {
		copied := make([]string, maxCols)
		copy(copied, row)
		out[i] = copied
	}
	return out
}
