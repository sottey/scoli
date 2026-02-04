package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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

	s.aiMu.Lock()
	chatSnapshot, err := s.loadAIChat(id)
	s.aiMu.Unlock()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, "chat not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to load chat")
		return
	}
	if chatSnapshot.Archived {
		writeError(w, http.StatusConflict, "chat is archived")
		return
	}

	assistantText, sources, handled, err := s.answerWeeklyTaskStatusQuestion(content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to answer task query")
		return
	}
	if !handled {
		assistantText, sources, handled, err = s.answerDirectTaskQuestion(content)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to answer task query")
		return
	}
	if !handled {
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

		now := time.Now()
		prompt := buildAIPrompt(
			content,
			matches,
			chatSnapshot.Messages,
			s.buildAIStructuredContext(now),
			now,
		)
		answer, err := openAIRespond(ctx, settings, prompt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "unable to get ai response")
			return
		}
		assistantText = strings.TrimSpace(answer)
		sources = make([]AIChatSource, 0, len(matches))
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
	}

	now := time.Now().UTC().Format(time.RFC3339)
	userMessage := AIChatMessage{
		Role:      "user",
		Content:   content,
		CreatedAt: now,
	}
	assistantMessage := AIChatMessage{
		Role:      "assistant",
		Content:   assistantText,
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

func (s *Server) answerWeeklyTaskStatusQuestion(question string) (string, []AIChatSource, bool, error) {
	if !isWeeklyTaskStatusQuestion(question) {
		return "", nil, false, nil
	}

	now := dateOnly(time.Now())
	start := startOfWeekMonday(now)
	end := start.AddDate(0, 0, 6)

	tasks, notice, err := s.listTasks()
	if err != nil {
		return "", nil, true, err
	}

	type datedTask struct {
		task TaskItem
		date time.Time
	}
	incomplete := make([]datedTask, 0)
	completed := make([]datedTask, 0)

	modDateCache := make(map[string]time.Time)
	modDateKnown := make(map[string]bool)
	for _, task := range tasks {
		var activityDate time.Time
		if dailyDate, ok := parseDailyNoteDate(task.Path); ok {
			activityDate = dailyDate
		} else if task.DueDateISO != "" {
			if parsed, parseErr := time.ParseInLocation("2006-01-02", task.DueDateISO, time.Local); parseErr == nil {
				activityDate = dateOnly(parsed)
			}
		}
		if activityDate.IsZero() {
			modDate, ok := modDateCache[task.Path]
			if !ok && !modDateKnown[task.Path] {
				absPath, _, resolveErr := s.resolvePath(task.Path)
				if resolveErr == nil {
					if stat, statErr := os.Stat(absPath); statErr == nil {
						modDate = dateOnly(stat.ModTime())
						modDateCache[task.Path] = modDate
					}
				}
				modDateKnown[task.Path] = true
			}
			activityDate = modDate
		}
		if activityDate.IsZero() || activityDate.Before(start) || activityDate.After(end) {
			continue
		}
		entry := datedTask{task: task, date: activityDate}
		if task.Completed {
			completed = append(completed, entry)
		} else {
			incomplete = append(incomplete, entry)
		}
	}

	archived, err := s.listArchivedTasksForWeek(start, end)
	if err != nil {
		return "", nil, true, err
	}

	sort.Slice(incomplete, func(i, j int) bool {
		if !sameDay(incomplete[i].date, incomplete[j].date) {
			return incomplete[i].date.After(incomplete[j].date)
		}
		if incomplete[i].task.Path != incomplete[j].task.Path {
			return incomplete[i].task.Path < incomplete[j].task.Path
		}
		return incomplete[i].task.LineNumber < incomplete[j].task.LineNumber
	})
	sort.Slice(completed, func(i, j int) bool {
		if !sameDay(completed[i].date, completed[j].date) {
			return completed[i].date.After(completed[j].date)
		}
		if completed[i].task.Path != completed[j].task.Path {
			return completed[i].task.Path < completed[j].task.Path
		}
		return completed[i].task.LineNumber < completed[j].task.LineNumber
	})

	rangeLabel := fmt.Sprintf("%s to %s", start.Format("2006-01-02"), now.Format("2006-01-02"))
	var builder strings.Builder
	builder.WriteString("Task summary this week (")
	builder.WriteString(rangeLabel)
	builder.WriteString("):\n")
	builder.WriteString("- Incomplete: ")
	builder.WriteString(strconv.Itoa(len(incomplete)))
	builder.WriteString("\n- Completed: ")
	builder.WriteString(strconv.Itoa(len(completed)))
	builder.WriteString("\n- Archived: ")
	builder.WriteString(strconv.Itoa(len(archived)))

	appendExamples := func(label string, items []datedTask) {
		max := 5
		if len(items) < max {
			max = len(items)
		}
		if max == 0 {
			return
		}
		builder.WriteString("\n\n")
		builder.WriteString(label)
		builder.WriteString(" examples:\n")
		for i := 0; i < max; i++ {
			item := items[i]
			builder.WriteString("- [")
			builder.WriteString(item.date.Format("2006-01-02"))
			builder.WriteString("] ")
			builder.WriteString(item.task.Text)
			builder.WriteString(" (")
			builder.WriteString(item.task.Path)
			builder.WriteString(":")
			builder.WriteString(strconv.Itoa(item.task.LineNumber))
			builder.WriteString(")\n")
		}
	}
	appendExamples("Incomplete", incomplete)
	appendExamples("Completed", completed)
	if len(archived) > 0 {
		max := 5
		if len(archived) < max {
			max = len(archived)
		}
		builder.WriteString("\n\nArchived examples:\n")
		for i := 0; i < max; i++ {
			item := archived[i]
			builder.WriteString("- [")
			builder.WriteString(item.Date.Format("2006-01-02"))
			builder.WriteString("] ")
			builder.WriteString(item.Text)
			builder.WriteString(" (")
			builder.WriteString(item.Path)
			builder.WriteString(":")
			builder.WriteString(strconv.Itoa(item.LineNumber))
			builder.WriteString(")\n")
		}
	}
	if notice != "" {
		builder.WriteString("\nNote: ")
		builder.WriteString(notice)
	}

	maxSources := 25
	sources := make([]AIChatSource, 0, maxSources)
	for _, item := range incomplete {
		if len(sources) >= maxSources {
			break
		}
		sources = append(sources, AIChatSource{
			Path:    item.task.Path,
			Heading: "Incomplete task",
			Snippet: fmt.Sprintf("[%s] %s (line %d)", item.date.Format("2006-01-02"), item.task.Text, item.task.LineNumber),
		})
	}
	for _, item := range completed {
		if len(sources) >= maxSources {
			break
		}
		sources = append(sources, AIChatSource{
			Path:    item.task.Path,
			Heading: "Completed task",
			Snippet: fmt.Sprintf("[%s] %s (line %d)", item.date.Format("2006-01-02"), item.task.Text, item.task.LineNumber),
		})
	}
	for _, item := range archived {
		if len(sources) >= maxSources {
			break
		}
		sources = append(sources, AIChatSource{
			Path:    item.Path,
			Heading: "Archived task",
			Snippet: fmt.Sprintf("[%s] %s (line %d)", item.Date.Format("2006-01-02"), item.Text, item.LineNumber),
		})
	}
	return strings.TrimSpace(builder.String()), sources, true, nil
}

func (s *Server) answerDirectTaskQuestion(question string) (string, []AIChatSource, bool, error) {
	if !isCompletedTasksThisWeekQuestion(question) {
		return "", nil, false, nil
	}

	now := dateOnly(time.Now())
	start := startOfWeekMonday(now)
	tasks, notice, err := s.listTasks()
	if err != nil {
		return "", nil, true, err
	}

	type completedTaskResult struct {
		task         TaskItem
		activityDate time.Time
	}

	results := make([]completedTaskResult, 0, len(tasks))
	modDateCache := make(map[string]time.Time)
	modDateKnown := make(map[string]bool)
	for _, task := range tasks {
		if !task.Completed {
			continue
		}
		if dailyDate, ok := parseDailyNoteDate(task.Path); ok {
			if !dailyDate.Before(start) && !dailyDate.After(now) {
				results = append(results, completedTaskResult{
					task:         task,
					activityDate: dailyDate,
				})
			}
			continue
		}

		modDate, ok := modDateCache[task.Path]
		if !ok && !modDateKnown[task.Path] {
			absPath, _, resolveErr := s.resolvePath(task.Path)
			if resolveErr == nil {
				if stat, statErr := os.Stat(absPath); statErr == nil {
					modDate = dateOnly(stat.ModTime())
					modDateCache[task.Path] = modDate
				}
			}
			modDateKnown[task.Path] = true
		}
		if !modDate.IsZero() && !modDate.Before(start) && !modDate.After(now) {
			results = append(results, completedTaskResult{
				task:         task,
				activityDate: modDate,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if !sameDay(results[i].activityDate, results[j].activityDate) {
			return results[i].activityDate.After(results[j].activityDate)
		}
		if results[i].task.Path != results[j].task.Path {
			return results[i].task.Path < results[j].task.Path
		}
		return results[i].task.LineNumber < results[j].task.LineNumber
	})

	rangeLabel := fmt.Sprintf("%s to %s", start.Format("2006-01-02"), now.Format("2006-01-02"))
	if len(results) == 0 {
		answer := "I could not find any completed tasks for this week (" + rangeLabel + ")."
		if notice != "" {
			answer += "\n\nNote: " + notice
		}
		return answer, []AIChatSource{}, true, nil
	}

	maxList := 25
	if len(results) < maxList {
		maxList = len(results)
	}
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Completed tasks this week (%s):\n", rangeLabel))
	for i := 0; i < maxList; i++ {
		item := results[i]
		builder.WriteString(fmt.Sprintf("- [%s] %s (%s:%d)\n",
			item.activityDate.Format("2006-01-02"),
			item.task.Text,
			item.task.Path,
			item.task.LineNumber,
		))
	}
	if len(results) > maxList {
		builder.WriteString(fmt.Sprintf("- ...and %d more.\n", len(results)-maxList))
	}
	if notice != "" {
		builder.WriteString("\nNote: ")
		builder.WriteString(notice)
	}

	maxSources := 25
	if len(results) < maxSources {
		maxSources = len(results)
	}
	sources := make([]AIChatSource, 0, maxSources)
	for i := 0; i < maxSources; i++ {
		item := results[i]
		sources = append(sources, AIChatSource{
			Path:    item.task.Path,
			Heading: "Completed task",
			Snippet: fmt.Sprintf("[%s] %s (line %d)", item.activityDate.Format("2006-01-02"), item.task.Text, item.task.LineNumber),
		})
	}

	return strings.TrimSpace(builder.String()), sources, true, nil
}

type archivedTaskResult struct {
	Path       string
	LineNumber int
	Text       string
	Date       time.Time
}

func (s *Server) listArchivedTasksForWeek(start, end time.Time) ([]archivedTaskResult, error) {
	results := make([]archivedTaskResult, 0)
	modDateCache := make(map[string]time.Time)
	pattern := regexp.MustCompile(`^\s*~\s*-\s+\[( |x|X|✓)\]\s+(.+)$`)

	err := filepath.WalkDir(s.notesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if isAiDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if isIgnoredFile(d.Name()) || !isMarkdown(d.Name()) {
			return nil
		}
		rel, err := filepath.Rel(s.notesDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		date := time.Time{}
		if dailyDate, ok := parseDailyNoteDate(rel); ok {
			date = dailyDate
		} else {
			if cached, ok := modDateCache[path]; ok {
				date = cached
			} else if stat, statErr := os.Stat(path); statErr == nil {
				date = dateOnly(stat.ModTime())
				modDateCache[path] = date
			}
		}
		if date.IsZero() || date.Before(start) || date.After(end) {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		lines := strings.Split(string(data), "\n")
		tracker := &codeBlockTracker{}
		for i, line := range lines {
			raw := strings.TrimSuffix(line, "\r")
			if tracker.isCodeLine(raw) {
				continue
			}
			match := pattern.FindStringSubmatch(raw)
			if len(match) < 3 {
				continue
			}
			text := cleanTaskText(match[2])
			if text == "" {
				text = strings.TrimSpace(match[2])
			}
			results = append(results, archivedTaskResult{
				Path:       rel,
				LineNumber: i + 1,
				Text:       text,
				Date:       date,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		if !sameDay(results[i].Date, results[j].Date) {
			return results[i].Date.After(results[j].Date)
		}
		if results[i].Path != results[j].Path {
			return results[i].Path < results[j].Path
		}
		return results[i].LineNumber < results[j].LineNumber
	})
	return results, nil
}

func isWeeklyTaskStatusQuestion(question string) bool {
	lower := strings.ToLower(strings.TrimSpace(question))
	if lower == "" {
		return false
	}
	if !strings.Contains(lower, "task") {
		return false
	}
	weekLike := strings.Contains(lower, "this week") ||
		strings.Contains(lower, "week so far") ||
		strings.Contains(lower, "so far this week")
	if !weekLike {
		return false
	}
	statusLike := strings.Contains(lower, "incomplete") ||
		strings.Contains(lower, "open") ||
		strings.Contains(lower, "pending") ||
		strings.Contains(lower, "completed") ||
		strings.Contains(lower, "finished") ||
		strings.Contains(lower, "done") ||
		strings.Contains(lower, "archived")
	return statusLike
}

func isCompletedTasksThisWeekQuestion(question string) bool {
	lower := strings.ToLower(strings.TrimSpace(question))
	if lower == "" {
		return false
	}
	if strings.Contains(lower, "incomplete") || strings.Contains(lower, "open") || strings.Contains(lower, "archived") {
		return false
	}
	taskLike := strings.Contains(lower, "task")
	completedLike := strings.Contains(lower, "completed") ||
		strings.Contains(lower, "finished") ||
		strings.Contains(lower, "done") ||
		strings.Contains(lower, "worked on")
	weekLike := strings.Contains(lower, "this week") ||
		strings.Contains(lower, "week so far") ||
		strings.Contains(lower, "so far this week")
	return taskLike && completedLike && weekLike
}

func startOfWeekMonday(value time.Time) time.Time {
	weekday := value.Weekday()
	delta := int(weekday) - int(time.Monday)
	if delta < 0 {
		delta += 7
	}
	return dateOnly(value.AddDate(0, 0, -delta))
}

func (s *Server) buildAIStructuredContext(now time.Time) string {
	now = now.In(time.Local)
	today := dateOnly(now)
	weekStart := startOfWeekMonday(today)
	weekEnd := weekStart.AddDate(0, 0, 6)

	tasks, notice, err := s.listTasks()
	if err != nil {
		return fmt.Sprintf(
			"Current date context:\n- Today: %s\n- This week (Monday-Sunday): %s to %s\n\nTask summary unavailable.",
			today.Format("2006-01-02"),
			weekStart.Format("2006-01-02"),
			weekEnd.Format("2006-01-02"),
		)
	}

	type completedRecord struct {
		task TaskItem
		date time.Time
	}
	completedThisWeek := make([]completedRecord, 0)
	overdueOpen := 0
	dueThisWeekOpen := 0
	for _, task := range tasks {
		dueDate := time.Time{}
		if task.DueDateISO != "" {
			if parsed, parseErr := time.ParseInLocation("2006-01-02", task.DueDateISO, time.Local); parseErr == nil {
				dueDate = dateOnly(parsed)
			}
		}

		if task.Completed {
			if dailyDate, ok := parseDailyNoteDate(task.Path); ok {
				if !dailyDate.Before(weekStart) && !dailyDate.After(today) {
					completedThisWeek = append(completedThisWeek, completedRecord{task: task, date: dailyDate})
				}
			}
			continue
		}

		if dueDate.IsZero() {
			continue
		}
		if dueDate.Before(today) {
			overdueOpen++
		}
		if !dueDate.Before(weekStart) && !dueDate.After(weekEnd) {
			dueThisWeekOpen++
		}
	}

	sort.Slice(completedThisWeek, func(i, j int) bool {
		if !sameDay(completedThisWeek[i].date, completedThisWeek[j].date) {
			return completedThisWeek[i].date.After(completedThisWeek[j].date)
		}
		if completedThisWeek[i].task.Path != completedThisWeek[j].task.Path {
			return completedThisWeek[i].task.Path < completedThisWeek[j].task.Path
		}
		return completedThisWeek[i].task.LineNumber < completedThisWeek[j].task.LineNumber
	})

	maxExamples := 8
	if len(completedThisWeek) < maxExamples {
		maxExamples = len(completedThisWeek)
	}

	var builder strings.Builder
	builder.WriteString("Current date context:\n")
	builder.WriteString("- Today: ")
	builder.WriteString(today.Format("2006-01-02"))
	builder.WriteString("\n- This week (Monday-Sunday): ")
	builder.WriteString(weekStart.Format("2006-01-02"))
	builder.WriteString(" to ")
	builder.WriteString(weekEnd.Format("2006-01-02"))
	builder.WriteString("\n\nTask summary:\n")
	builder.WriteString("- Completed tasks this week: ")
	builder.WriteString(strconv.Itoa(len(completedThisWeek)))
	builder.WriteString("\n- Open overdue tasks: ")
	builder.WriteString(strconv.Itoa(overdueOpen))
	builder.WriteString("\n- Open tasks due this week: ")
	builder.WriteString(strconv.Itoa(dueThisWeekOpen))
	if maxExamples > 0 {
		builder.WriteString("\n\nCompleted this week examples:\n")
		for i := 0; i < maxExamples; i++ {
			entry := completedThisWeek[i]
			builder.WriteString("- [")
			builder.WriteString(entry.date.Format("2006-01-02"))
			builder.WriteString("] ")
			builder.WriteString(entry.task.Text)
			builder.WriteString(" (")
			builder.WriteString(entry.task.Path)
			builder.WriteString(":")
			builder.WriteString(strconv.Itoa(entry.task.LineNumber))
			builder.WriteString(")\n")
		}
	}
	if notice != "" {
		builder.WriteString("\nTask parser notice:\n- ")
		builder.WriteString(notice)
	}
	return strings.TrimSpace(builder.String())
}

func formatRecentChatHistory(messages []AIChatMessage, maxMessages int) string {
	if len(messages) == 0 || maxMessages <= 0 {
		return "None"
	}
	start := 0
	if len(messages) > maxMessages {
		start = len(messages) - maxMessages
	}

	var builder strings.Builder
	for i := start; i < len(messages); i++ {
		message := messages[i]
		role := strings.TrimSpace(strings.ToLower(message.Role))
		if role != "user" && role != "assistant" {
			continue
		}
		text := strings.TrimSpace(message.Content)
		if text == "" {
			continue
		}
		if len(text) > 500 {
			text = text[:500] + "..."
		}
		builder.WriteString("- ")
		builder.WriteString(role)
		builder.WriteString(": ")
		builder.WriteString(text)
		builder.WriteString("\n")
	}
	result := strings.TrimSpace(builder.String())
	if result == "" {
		return "None"
	}
	return result
}

func buildAIPrompt(question string, matches []AIChunkMatch, history []AIChatMessage, structuredContext string, now time.Time) string {
	var builder strings.Builder
	builder.WriteString("Context:\n")
	builder.WriteString("- Request timestamp: ")
	builder.WriteString(now.In(time.Local).Format(time.RFC3339))
	builder.WriteString("\n- Relative dates (today, this week, etc.) must be interpreted using local server time.\n\n")
	if strings.TrimSpace(structuredContext) != "" {
		builder.WriteString(structuredContext)
		builder.WriteString("\n\n")
	}
	builder.WriteString("Recent conversation:\n")
	builder.WriteString(formatRecentChatHistory(history, 8))
	builder.WriteString("\n\nQuestion:\n")
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
