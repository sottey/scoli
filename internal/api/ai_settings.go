package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
)

const aiSettingsFileName = "ai-settings.json"
const aiChatsFileName = "chats.json"
const aiChatsDirName = "chats"
const aiIndexFileName = "index.sqlite"

type AISettings struct {
	Version          int     `json:"version"`
	APIKey           string  `json:"apiKey"`
	ChatModel        string  `json:"chatModel"`
	EmbedModel       string  `json:"embedModel"`
	TopK             int     `json:"topK"`
	MaxContextChunks int     `json:"maxContextChunks"`
	Temperature      float64 `json:"temperature"`
	MaxOutputTokens  int     `json:"maxOutputTokens"`
	ChunkCharLimit   int     `json:"chunkCharLimit"`
	SectionCharLimit int     `json:"sectionCharLimit"`
}

type AISettingsResponse struct {
	Settings   AISettings `json:"settings"`
	Configured bool       `json:"configured"`
	Notice     string     `json:"notice,omitempty"`
}

func defaultAISettings() AISettings {
	return AISettings{
		Version:          1,
		APIKey:           "",
		ChatModel:        "gpt-4o-mini",
		EmbedModel:       "text-embedding-3-small",
		TopK:             6,
		MaxContextChunks: 6,
		Temperature:      0.2,
		MaxOutputTokens:  500,
		ChunkCharLimit:   1600,
		SectionCharLimit: 5000,
	}
}

func applyAISettingsDefaults(settings *AISettings) {
	if settings.Version == 0 {
		settings.Version = 1
	}
	if settings.ChatModel == "" {
		settings.ChatModel = "gpt-4o-mini"
	}
	if settings.EmbedModel == "" {
		settings.EmbedModel = "text-embedding-3-small"
	}
	if settings.TopK <= 0 {
		settings.TopK = 6
	}
	if settings.MaxContextChunks <= 0 {
		settings.MaxContextChunks = settings.TopK
	}
	if settings.Temperature < 0 {
		settings.Temperature = 0
	}
	if settings.Temperature > 2 {
		settings.Temperature = 2
	}
	if settings.MaxOutputTokens <= 0 {
		settings.MaxOutputTokens = 500
	}
	if settings.ChunkCharLimit <= 0 {
		settings.ChunkCharLimit = 1600
	}
	if settings.SectionCharLimit <= 0 {
		settings.SectionCharLimit = 5000
	}
}

func (s *Server) aiDirPath() string {
	return filepath.Join(s.notesDir, aiFolderName)
}

func (s *Server) aiSettingsPath() string {
	return filepath.Join(s.aiDirPath(), aiSettingsFileName)
}

func (s *Server) aiChatsPath() string {
	return filepath.Join(s.aiDirPath(), aiChatsFileName)
}

func (s *Server) aiChatsDirPath() string {
	return filepath.Join(s.aiDirPath(), aiChatsDirName)
}

func (s *Server) aiIndexPath() string {
	return filepath.Join(s.aiDirPath(), aiIndexFileName)
}

func (s *Server) ensureAIStorage() error {
	if err := os.MkdirAll(s.aiDirPath(), 0o755); err != nil {
		return err
	}
	_, _, err := s.loadAISettings()
	return err
}

func (s *Server) loadAISettings() (AISettings, string, error) {
	path := s.aiSettingsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			settings := defaultAISettings()
			if err := os.MkdirAll(s.aiDirPath(), 0o755); err != nil {
				return settings, "", err
			}
			if err := s.saveAISettings(settings); err != nil {
				return settings, "", err
			}
			s.logger.Info("ai settings created", "path", path)
			return settings, "Created ai-settings.json", nil
		}
		return AISettings{}, "", err
	}

	var settings AISettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return AISettings{}, "", err
	}
	applyAISettingsDefaults(&settings)
	return settings, "", nil
}

func (s *Server) saveAISettings(settings AISettings) error {
	if settings.Version == 0 {
		return errors.New("ai settings version is required")
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.aiSettingsPath(), data, 0o644)
}

func (s *Server) handleAISettingsGet(w http.ResponseWriter, r *http.Request) {
	settings, notice, err := s.loadAISettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load ai settings")
		return
	}
	sanitized := settings
	sanitized.APIKey = ""
	resp := AISettingsResponse{
		Settings:   sanitized,
		Configured: settings.APIKey != "",
	}
	if notice != "" {
		resp.Notice = notice
	}
	writeJSON(w, http.StatusOK, resp)
}
