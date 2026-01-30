package api

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	emailScheduleDigest = "digest"
	emailScheduleDue    = "due"
)

func (s *Server) startEmailSchedulers() {
	s.emailSchedulerOnce.Do(func() {
		go s.runEmailScheduler(emailScheduleDigest)
		go s.runEmailScheduler(emailScheduleDue)
	})
}

func (s *Server) runEmailScheduler(kind string) {
	for {
		if !s.emailSettingsExists() {
			time.Sleep(1 * time.Minute)
			continue
		}

		settings, _, err := s.loadEmailSettings()
		if err != nil {
			s.logger.Error("email settings load failed", "error", err)
			time.Sleep(2 * time.Minute)
			continue
		}

		enabled, timeStr := emailScheduleConfig(settings, kind)
		if !settings.Enabled || !enabled {
			time.Sleep(5 * time.Minute)
			continue
		}

		nextRun, err := nextScheduledTime(timeStr, timeNow())
		if err != nil {
			s.logger.Error("email schedule invalid time", "kind", kind, "value", timeStr, "error", err)
			time.Sleep(5 * time.Minute)
			continue
		}

		wait := time.Until(nextRun)
		if wait > 0 {
			time.Sleep(wait)
		}

		settings, _, err = s.loadEmailSettings()
		if err != nil {
			s.logger.Error("email settings reload failed", "error", err)
			continue
		}
		enabled, _ = emailScheduleConfig(settings, kind)
		if !settings.Enabled || !enabled {
			continue
		}

		switch kind {
		case emailScheduleDigest:
			if err := s.sendDigestEmail(settings); err != nil {
				s.logger.Error("email digest send failed", "error", err)
			}
		case emailScheduleDue:
			if err := s.sendDueEmail(settings); err != nil {
				s.logger.Error("email due send failed", "error", err)
			}
		}
	}
}

func emailScheduleConfig(settings EmailSettings, kind string) (bool, string) {
	switch kind {
	case emailScheduleDigest:
		return settings.Digest.Enabled, settings.Digest.Time
	case emailScheduleDue:
		return settings.Due.Enabled, settings.Due.Time
	default:
		return false, ""
	}
}

func nextScheduledTime(value string, now time.Time) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, errors.New("time is required")
	}
	parsed, err := time.ParseInLocation("15:04", value, now.Location())
	if err != nil {
		return time.Time{}, err
	}
	next := time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next, nil
}

func (s *Server) emailSettingsExists() bool {
	_, err := os.Stat(s.emailSettingsFilePath())
	return err == nil
}

func (s *Server) sendDigestEmail(settings EmailSettings) error {
	subject := fmt.Sprintf("Scoli Daily Digest - %s", timeNow().Format(dailyDateLayout))
	body, err := s.buildDigestEmail(settings)
	if err != nil {
		return err
	}
	return sendEmail(settings.SMTP, subject, body)
}

func (s *Server) sendDueEmail(settings EmailSettings) error {
	subject := fmt.Sprintf("Scoli Tasks Due - %s", timeNow().Format(dailyDateLayout))
	body, err := s.buildDueEmail(settings)
	if err != nil {
		return err
	}
	return sendEmail(settings.SMTP, subject, body)
}

func (s *Server) buildDigestEmail(settings EmailSettings) (string, error) {
	template := s.loadEmailTemplate(settings.Templates.Digest, defaultDigestTemplate)
	tokens, err := s.buildEmailTokens(settings)
	if err != nil {
		return "", err
	}
	return replaceEmailTokens(template, tokens), nil
}

func (s *Server) buildDueEmail(settings EmailSettings) (string, error) {
	template := s.loadEmailTemplate(settings.Templates.Due, defaultDueTemplate)
	tokens, err := s.buildEmailTokens(settings)
	if err != nil {
		return "", err
	}
	return replaceEmailTokens(template, tokens), nil
}

func (s *Server) loadEmailTemplate(path, fallback string) string {
	cleaned, err := cleanRelPath(path)
	if err != nil || cleaned == "" {
		return fallback
	}
	absPath := filepath.Join(s.notesDir, filepath.FromSlash(cleaned))
	data, err := os.ReadFile(absPath)
	if err != nil {
		return fallback
	}
	return string(data)
}

func replaceEmailTokens(template string, tokens map[string]string) string {
	result := template
	for key, value := range tokens {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}
	return strings.TrimSpace(result) + "\n"
}

func (s *Server) buildEmailTokens(settings EmailSettings) (map[string]string, error) {
	tasks, _, err := s.listTasks()
	if err != nil {
		return nil, err
	}
	now := timeNow()
	today := dateOnly(now)
	upcomingDays := settings.Due.WindowDays
	if upcomingDays < 0 {
		upcomingDays = 0
	}
	upcomingLimit := today.AddDate(0, 0, upcomingDays)

	var overdue []TaskItem
	var dueToday []TaskItem
	var upcoming []TaskItem
	openTasks := make([]TaskItem, 0, len(tasks))

	for _, task := range tasks {
		if task.Completed {
			continue
		}
		openTasks = append(openTasks, task)
		if task.DueDateISO == "" {
			continue
		}
		dueDate, err := time.ParseInLocation(dailyDateLayout, task.DueDateISO, now.Location())
		if err != nil {
			continue
		}
		dueDate = dateOnly(dueDate)
		switch {
		case dueDate.Before(today):
			if settings.Due.IncludeOverdue {
				overdue = append(overdue, task)
			}
		case sameDay(dueDate, today):
			dueToday = append(dueToday, task)
		case upcomingDays > 0 && dueDate.After(today) && !dueDate.After(upcomingLimit):
			upcoming = append(upcoming, task)
		}
	}

	sortTasksByDue(overdue)
	sortTasksByDue(dueToday)
	sortTasksByDue(upcoming)

	tasksByProject := formatTasksByProject(openTasks)
	yesterdaySummary, completedYesterday := s.buildYesterdaySummary()

	return map[string]string{
		"date":                today.Format(dailyDateLayout),
		"tasks_overdue":       formatTaskList(overdue),
		"tasks_today":         formatTaskList(dueToday),
		"tasks_upcoming":      formatTaskList(upcoming),
		"tasks_by_project":    tasksByProject,
		"notes_summary":       yesterdaySummary,
		"completed_yesterday": completedYesterday,
	}, nil
}

func sortTasksByDue(tasks []TaskItem) {
	sort.Slice(tasks, func(i, j int) bool {
		left := tasks[i]
		right := tasks[j]
		if left.DueDateISO != right.DueDateISO {
			return left.DueDateISO < right.DueDateISO
		}
		if left.Priority != right.Priority {
			return left.Priority > right.Priority
		}
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		return left.LineNumber < right.LineNumber
	})
}

func formatTaskList(tasks []TaskItem) string {
	if len(tasks) == 0 {
		return "None"
	}
	lines := make([]string, 0, len(tasks))
	for _, task := range tasks {
		lines = append(lines, formatTaskLine(task))
	}
	return strings.Join(lines, "\n")
}

func formatTasksByProject(tasks []TaskItem) string {
	if len(tasks) == 0 {
		return "No open tasks."
	}
	byProject := make(map[string][]TaskItem)
	for _, task := range tasks {
		project := strings.TrimSpace(task.Project)
		if project == "" {
			project = "No Project"
		}
		byProject[project] = append(byProject[project], task)
	}
	projects := make([]string, 0, len(byProject))
	for project := range byProject {
		projects = append(projects, project)
	}
	sort.Slice(projects, func(i, j int) bool {
		a := projects[i]
		b := projects[j]
		if a == "No Project" && b != "No Project" {
			return false
		}
		if b == "No Project" && a != "No Project" {
			return true
		}
		return strings.ToLower(a) < strings.ToLower(b)
	})

	var lines []string
	for _, project := range projects {
		lines = append(lines, project+":")
		items := byProject[project]
		sortTasksByDue(items)
		for _, task := range items {
			lines = append(lines, "  "+formatTaskLine(task))
		}
	}
	return strings.Join(lines, "\n")
}

func formatTaskLine(task TaskItem) string {
	parts := []string{task.Text}
	if task.Project != "" {
		parts = append(parts, "+"+task.Project)
	}
	if task.DueDateISO != "" {
		parts = append(parts, "due "+task.DueDateISO)
	}
	if task.Path != "" {
		parts = append(parts, task.Path)
	}
	return "- " + strings.Join(parts, " · ")
}

func (s *Server) buildYesterdaySummary() (string, string) {
	yesterday := dateOnly(timeNow()).AddDate(0, 0, -1)
	relPath := dailyNotePathForDate(yesterday)
	absPath := filepath.Join(s.notesDir, filepath.FromSlash(relPath))
	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Sprintf("No daily note found for %s.", yesterday.Format(dailyDateLayout)), "None"
	}
	content := string(data)
	noteSummary := summarizeNoteContent(content, 5)
	completed := summarizeCompletedTasks(content)
	return noteSummary, completed
}

func dailyNotePathForDate(value time.Time) string {
	return filepath.ToSlash(filepath.Join(dailyFolderName, value.Format(dailyDateLayout)+".md"))
}

func summarizeNoteContent(content string, maxLines int) string {
	lines := strings.Split(content, "\n")
	summary := make([]string, 0, maxLines)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		summary = append(summary, "- "+trimmed)
		if len(summary) >= maxLines {
			break
		}
	}
	if len(summary) == 0 {
		return "No summary lines found."
	}
	return strings.Join(summary, "\n")
}

func summarizeCompletedTasks(content string) string {
	parsed := parseTodoLines(content)
	var lines []string
	for _, todo := range parsed {
		if todo.Completed {
			if todo.Text != "" {
				lines = append(lines, "- "+todo.Text)
			}
		}
	}
	if len(lines) == 0 {
		return "None"
	}
	return strings.Join(lines, "\n")
}
