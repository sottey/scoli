package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type AIChatList struct {
	Version int          `json:"version"`
	Chats   []AIChatMeta `json:"chats"`
}

type AIChatMeta struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
	MessageCount int    `json:"messageCount"`
	Archived     bool   `json:"archived"`
}

type AIChat struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
	Archived  bool            `json:"archived"`
	Messages  []AIChatMessage `json:"messages"`
}

type AIChatMessage struct {
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	CreatedAt string         `json:"createdAt"`
	Sources   []AIChatSource `json:"sources,omitempty"`
}

type AIChatSource struct {
	Path    string `json:"path"`
	Heading string `json:"heading,omitempty"`
	Snippet string `json:"snippet,omitempty"`
}

type AIChatCreatePayload struct {
	Title string `json:"title"`
}

type AIChatMessagePayload struct {
	Content string `json:"content"`
}

type AIChatMessageResponse struct {
	Chat AIChat `json:"chat"`
}

func (s *Server) handleAIChatsList(w http.ResponseWriter, r *http.Request) {
	list, err := s.loadAIChatList()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load ai chats")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleAIChatCreate(w http.ResponseWriter, r *http.Request) {
	payload, _ := decodeJSON[AIChatCreatePayload](r.Body)
	title := strings.TrimSpace(payload.Title)
	if title == "" {
		title = "New Chat"
	}
	id, err := generateChatID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to create chat")
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	chat := AIChat{
		ID:        id,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  []AIChatMessage{},
	}

	s.aiMu.Lock()
	defer s.aiMu.Unlock()

	if err := s.saveAIChat(chat); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save chat")
		return
	}
	list, err := s.loadAIChatList()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load ai chats")
		return
	}
	meta := AIChatMeta{
		ID:           chat.ID,
		Title:        chat.Title,
		CreatedAt:    chat.CreatedAt,
		UpdatedAt:    chat.UpdatedAt,
		MessageCount: 0,
		Archived:     chat.Archived,
	}
	list.Chats = append([]AIChatMeta{meta}, list.Chats...)
	if err := s.saveAIChatList(list); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save ai chats")
		return
	}
	writeJSON(w, http.StatusCreated, meta)
}

func (s *Server) handleAIChatGet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing chat id")
		return
	}
	chat, err := s.loadAIChat(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, "chat not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to load chat")
		return
	}
	writeJSON(w, http.StatusOK, chat)
}

func (s *Server) handleAIChatMessage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing chat id")
		return
	}
	payload, err := decodeJSON[AIChatMessagePayload](r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	content := strings.TrimSpace(payload.Content)
	if content == "" {
		writeError(w, http.StatusBadRequest, "message content is required")
		return
	}

	settings, _, err := s.loadAISettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load ai settings")
		return
	}
	if settings.APIKey == "" {
		writeError(w, http.StatusBadRequest, "missing OpenAI API key; set apiKey in Notes/.ai/ai-settings.json")
		return
	}

	idx, err := s.getAIIndex()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to open ai index")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	matches, err := idx.query(ctx, settings, s.notesDir, content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ai query failed: "+err.Error())
		return
	}

	maxChunks := settings.MaxContextChunks
	if maxChunks <= 0 {
		maxChunks = settings.TopK
	}
	if maxChunks <= 0 {
		maxChunks = 6
	}
	if len(matches) > maxChunks {
		matches = matches[:maxChunks]
	}

	prompt := buildAIPrompt(content, matches)
	answer, err := openAIRespond(ctx, settings, prompt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get ai response")
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	userMessage := AIChatMessage{
		Role:      "user",
		Content:   content,
		CreatedAt: now,
	}
	sources := make([]AIChatSource, 0, len(matches))
	for _, match := range matches {
		snippet := strings.TrimSpace(match.Content)
		if len(snippet) > 240 {
			snippet = snippet[:240] + "..."
		}
		sources = append(sources, AIChatSource{
			Path:    match.NotePath,
			Heading: match.Heading,
			Snippet: snippet,
		})
	}
	assistantMessage := AIChatMessage{
		Role:      "assistant",
		Content:   strings.TrimSpace(answer),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Sources:   sources,
	}

	s.aiMu.Lock()
	defer s.aiMu.Unlock()

	chat, err := s.loadAIChat(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, "chat not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to load chat")
		return
	}
	if chat.Archived {
		writeError(w, http.StatusConflict, "chat is archived")
		return
	}
	chat.Messages = append(chat.Messages, userMessage, assistantMessage)
	if chat.Title == "New Chat" && len(chat.Messages) > 0 {
		chat.Title = truncateTitle(content)
	}
	chat.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := s.saveAIChat(chat); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save chat")
		return
	}
	if err := s.updateAIChatMeta(chat); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to update chat list")
		return
	}
	writeJSON(w, http.StatusOK, AIChatMessageResponse{Chat: chat})
}

func buildAIPrompt(question string, matches []AIChunkMatch) string {
	var builder strings.Builder
	builder.WriteString("Question:\n")
	builder.WriteString(question)
	builder.WriteString("\n\nSnippets:\n")
	for i, match := range matches {
		builder.WriteString("[")
		builder.WriteString(strconv.Itoa(i + 1))
		builder.WriteString("] ")
		builder.WriteString(match.NotePath)
		if match.Heading != "" {
			builder.WriteString(" — ")
			builder.WriteString(match.Heading)
		}
		builder.WriteString("\n")
		builder.WriteString(match.Content)
		builder.WriteString("\n\n")
	}
	return strings.TrimSpace(builder.String())
}

func (s *Server) loadAIChatList() (AIChatList, error) {
	path := s.aiChatsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			list := AIChatList{Version: 1, Chats: []AIChatMeta{}}
			if err := s.saveAIChatList(list); err != nil {
				return list, err
			}
			return list, nil
		}
		return AIChatList{}, err
	}
	var list AIChatList
	if err := json.Unmarshal(data, &list); err != nil {
		return AIChatList{}, err
	}
	if list.Version == 0 {
		list.Version = 1
	}
	if list.Chats == nil {
		list.Chats = []AIChatMeta{}
	}
	return list, nil
}

func (s *Server) saveAIChatList(list AIChatList) error {
	if err := os.MkdirAll(s.aiDirPath(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.aiChatsPath(), data, 0o644)
}

func (s *Server) chatFilePath(id string) string {
	return filepath.Join(s.aiChatsDirPath(), id+".json")
}

func (s *Server) loadAIChat(id string) (AIChat, error) {
	data, err := os.ReadFile(s.chatFilePath(id))
	if err != nil {
		return AIChat{}, err
	}
	var chat AIChat
	if err := json.Unmarshal(data, &chat); err != nil {
		return AIChat{}, err
	}
	return chat, nil
}

func (s *Server) saveAIChat(chat AIChat) error {
	if err := os.MkdirAll(s.aiChatsDirPath(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(chat, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.chatFilePath(chat.ID), data, 0o644)
}

func (s *Server) updateAIChatMeta(chat AIChat) error {
	list, err := s.loadAIChatList()
	if err != nil {
		return err
	}
	messageCount := 0
	for _, msg := range chat.Messages {
		if msg.Role == "user" || msg.Role == "assistant" {
			messageCount++
		}
	}
	updated := false
	for i, item := range list.Chats {
		if item.ID == chat.ID {
			list.Chats[i].Title = chat.Title
			list.Chats[i].UpdatedAt = chat.UpdatedAt
			list.Chats[i].MessageCount = messageCount
			list.Chats[i].Archived = chat.Archived
			updated = true
			break
		}
	}
	if !updated {
		list.Chats = append([]AIChatMeta{{
			ID:           chat.ID,
			Title:        chat.Title,
			CreatedAt:    chat.CreatedAt,
			UpdatedAt:    chat.UpdatedAt,
			MessageCount: messageCount,
			Archived:     chat.Archived,
		}}, list.Chats...)
	}
	return s.saveAIChatList(list)
}

func (s *Server) handleAIChatArchive(w http.ResponseWriter, r *http.Request) {
	s.handleAIChatArchiveToggle(w, r, true)
}

func (s *Server) handleAIChatUnarchive(w http.ResponseWriter, r *http.Request) {
	s.handleAIChatArchiveToggle(w, r, false)
}

func (s *Server) handleAIChatArchiveToggle(w http.ResponseWriter, r *http.Request, archived bool) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing chat id")
		return
	}

	s.aiMu.Lock()
	defer s.aiMu.Unlock()

	chat, err := s.loadAIChat(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, "chat not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to load chat")
		return
	}
	if chat.Archived == archived {
		messageCount := 0
		for _, msg := range chat.Messages {
			if msg.Role == "user" || msg.Role == "assistant" {
				messageCount++
			}
		}
		writeJSON(w, http.StatusOK, AIChatMeta{
			ID:           chat.ID,
			Title:        chat.Title,
			CreatedAt:    chat.CreatedAt,
			UpdatedAt:    chat.UpdatedAt,
			MessageCount: messageCount,
			Archived:     chat.Archived,
		})
		return
	}
	chat.Archived = archived
	chat.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := s.saveAIChat(chat); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save chat")
		return
	}
	if err := s.updateAIChatMeta(chat); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to update chat list")
		return
	}

	messageCount := 0
	for _, msg := range chat.Messages {
		if msg.Role == "user" || msg.Role == "assistant" {
			messageCount++
		}
	}
	writeJSON(w, http.StatusOK, AIChatMeta{
		ID:           chat.ID,
		Title:        chat.Title,
		CreatedAt:    chat.CreatedAt,
		UpdatedAt:    chat.UpdatedAt,
		MessageCount: messageCount,
		Archived:     chat.Archived,
	})
}

func (s *Server) handleAIChatDelete(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing chat id")
		return
	}

	s.aiMu.Lock()
	defer s.aiMu.Unlock()

	list, err := s.loadAIChatList()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load ai chats")
		return
	}

	found := false
	filtered := make([]AIChatMeta, 0, len(list.Chats))
	for _, chat := range list.Chats {
		if chat.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, chat)
	}
	if !found {
		writeError(w, http.StatusNotFound, "chat not found")
		return
	}

	if err := os.Remove(s.chatFilePath(id)); err != nil && !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "unable to delete chat")
		return
	}

	list.Chats = filtered
	if err := s.saveAIChatList(list); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to save ai chats")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func generateChatID() (string, error) {
	nonce := make([]byte, 8)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	timestamp := time.Now().UTC().Format("20060102T150405")
	return timestamp + "-" + hex.EncodeToString(nonce), nil
}

func truncateTitle(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "New Chat"
	}
	if len(trimmed) > 48 {
		return strings.TrimSpace(trimmed[:48]) + "..."
	}
	return trimmed
}
