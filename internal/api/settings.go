package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const settingsFileName = "settings.json"

const (
	notesSortByName    = "name"
	notesSortByCreated = "created"
	notesSortByUpdated = "updated"

	notesSortOrderAsc  = "asc"
	notesSortOrderDesc = "desc"
)

func isValidNotesSortBy(value string) bool {
	switch value {
	case notesSortByName, notesSortByCreated, notesSortByUpdated:
		return true
	default:
		return false
	}
}

func isValidNotesSortOrder(value string) bool {
	switch value {
	case notesSortOrderAsc, notesSortOrderDesc:
		return true
	default:
		return false
	}
}

type Settings struct {
	Version              int               `json:"version"`
	DarkMode             bool              `json:"darkMode"`
	DefaultView          string            `json:"defaultView"`
	SidebarWidth         int               `json:"sidebarWidth"`
	DefaultFolder        string            `json:"defaultFolder"`
	ShowTemplates        bool              `json:"showTemplates"`
	ShowAiNode           bool              `json:"showAiNode"`
	NotesSortBy          string            `json:"notesSortBy"`
	NotesSortOrder       string            `json:"notesSortOrder"`
	ExternalCommandsPath string            `json:"externalCommandsPath"`
	RootIcons            map[string]string `json:"rootIcons,omitempty"`
}

type SettingsResponse struct {
	Settings Settings  `json:"settings"`
	Build    BuildInfo `json:"build"`
	Notice   string    `json:"notice,omitempty"`
}

type SettingsPayload struct {
	DarkMode             *bool   `json:"darkMode,omitempty"`
	DefaultView          *string `json:"defaultView,omitempty"`
	SidebarWidth         *int    `json:"sidebarWidth,omitempty"`
	DefaultFolder        *string `json:"defaultFolder,omitempty"`
	ShowTemplates        *bool   `json:"showTemplates,omitempty"`
	ShowAiNode           *bool   `json:"showAiNode,omitempty"`
	NotesSortBy          *string `json:"notesSortBy,omitempty"`
	NotesSortOrder       *string `json:"notesSortOrder,omitempty"`
	ExternalCommandsPath *string `json:"externalCommandsPath,omitempty"`
}

func (s *Server) handleSettingsGet(w http.ResponseWriter, r *http.Request) {
	settings, notice, err := s.loadSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load settings")
		return
	}

	resp := SettingsResponse{Settings: settings}
	resp.Build = BuildInfo{
		GitTag:    BuildGitTag,
		DockerTag: BuildDockerTag,
		CommitSHA: BuildCommitSHA,
	}
	if notice != "" {
		resp.Notice = notice
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeJSON[SettingsPayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateSettingsPayload(payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	settings, _, err := s.loadSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load settings")
		return
	}

	changed := make([]string, 0, 8)
	if payload.DarkMode != nil {
		settings.DarkMode = *payload.DarkMode
		changed = append(changed, "darkMode")
	}
	if payload.DefaultView != nil {
		settings.DefaultView = *payload.DefaultView
		changed = append(changed, "defaultView")
	}
	if payload.SidebarWidth != nil {
		settings.SidebarWidth = *payload.SidebarWidth
		changed = append(changed, "sidebarWidth")
	}
	if payload.DefaultFolder != nil {
		settings.DefaultFolder = *payload.DefaultFolder
		changed = append(changed, "defaultFolder")
	}
	if payload.ShowTemplates != nil {
		settings.ShowTemplates = *payload.ShowTemplates
		changed = append(changed, "showTemplates")
	}
	if payload.ShowAiNode != nil {
		settings.ShowAiNode = *payload.ShowAiNode
		changed = append(changed, "showAiNode")
	}
	if payload.NotesSortBy != nil {
		settings.NotesSortBy = *payload.NotesSortBy
		changed = append(changed, "notesSortBy")
	}
	if payload.NotesSortOrder != nil {
		settings.NotesSortOrder = *payload.NotesSortOrder
		changed = append(changed, "notesSortOrder")
	}
	if payload.ExternalCommandsPath != nil {
		settings.ExternalCommandsPath = *payload.ExternalCommandsPath
		changed = append(changed, "externalCommandsPath")
	}
	if err := s.saveSettings(settings); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save settings")
		return
	}

	s.logger.Info("settings updated", "fields", strings.Join(changed, ","))
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) settingsFilePath() string {
	return filepath.Join(s.notesDir, settingsFileName)
}

func (s *Server) loadSettings() (Settings, string, error) {
	path := s.settingsFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			settings := Settings{
				Version:              7,
				DarkMode:             false,
				DefaultView:          "split",
				SidebarWidth:         300,
				DefaultFolder:        "",
				ShowTemplates:        true,
				ShowAiNode:           true,
				NotesSortBy:          notesSortByName,
				NotesSortOrder:       notesSortOrderAsc,
				ExternalCommandsPath: "",
				RootIcons:            map[string]string{},
			}
			if err := os.MkdirAll(s.notesDir, 0o755); err != nil {
				return settings, "", err
			}
			if err := s.saveSettings(settings); err != nil {
				return settings, "", err
			}
			s.logger.Info("settings created", "path", path)
			return settings, "Created settings.json", nil
		}
		return Settings{}, "", err
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return Settings{}, "", err
	}
	if settings.Version == 0 {
		settings.Version = 2
	}
	if settings.DefaultView == "" {
		settings.DefaultView = "split"
	}
	if settings.NotesSortBy == "" || !isValidNotesSortBy(settings.NotesSortBy) {
		settings.NotesSortBy = notesSortByName
	}
	if settings.NotesSortOrder == "" || !isValidNotesSortOrder(settings.NotesSortOrder) {
		settings.NotesSortOrder = notesSortOrderAsc
	}
	if settings.SidebarWidth == 0 {
		settings.SidebarWidth = 300
	}
	if settings.DefaultFolder == "." {
		settings.DefaultFolder = ""
	}
	if settings.Version < 2 {
		settings.ShowTemplates = true
		settings.Version = 2
	}
	if settings.Version < 3 {
		settings.NotesSortBy = notesSortByName
		settings.NotesSortOrder = notesSortOrderAsc
		settings.Version = 3
	}
	if settings.Version < 4 {
		if settings.ExternalCommandsPath == "" {
			settings.ExternalCommandsPath = ""
		}
		settings.Version = 4
	}
	if settings.Version < 5 {
		if settings.RootIcons == nil {
			settings.RootIcons = map[string]string{}
		}
		settings.Version = 5
	}
	if settings.Version < 6 {
		settings.Version = 6
	}
	if settings.Version < 7 {
		settings.ShowAiNode = true
		settings.Version = 7
	}
	if settings.RootIcons == nil {
		settings.RootIcons = map[string]string{}
	}

	return settings, "", nil
}

func (s *Server) saveSettings(settings Settings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.settingsFilePath(), data, 0o644)
}

func validateSettingsPayload(payload SettingsPayload) error {
	if payload.DefaultView != nil {
		switch *payload.DefaultView {
		case "edit", "preview", "split":
			// ok
		default:
			return errors.New("defaultView must be edit, preview, or split")
		}
	}
	if payload.SidebarWidth != nil {
		if *payload.SidebarWidth < 220 || *payload.SidebarWidth > 600 {
			return errors.New("sidebarWidth must be between 220 and 600")
		}
	}
	if payload.NotesSortBy != nil {
		if !isValidNotesSortBy(*payload.NotesSortBy) {
			return errors.New("notesSortBy must be name, created, or updated")
		}
	}
	if payload.NotesSortOrder != nil {
		if !isValidNotesSortOrder(*payload.NotesSortOrder) {
			return errors.New("notesSortOrder must be asc or desc")
		}
	}
	if payload.DefaultFolder != nil {
		cleaned, err := cleanRelPath(*payload.DefaultFolder)
		if err != nil {
			return err
		}
		*payload.DefaultFolder = cleaned
	}
	if payload.ExternalCommandsPath != nil {
		cleaned, err := cleanRelPath(*payload.ExternalCommandsPath)
		if err != nil {
			return err
		}
		*payload.ExternalCommandsPath = cleaned
	}
	return nil
}
