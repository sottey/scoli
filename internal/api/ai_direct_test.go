package api

import (
	"strings"
	"testing"
	"time"
)

func TestIsCompletedTasksThisWeekQuestion(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: "can you tell me what tasks I have completed this week so far?", want: true},
		{input: "what tasks have I worked on this week?", want: true},
		{input: "what tasks are due this week?", want: false},
		{input: "what did I complete today?", want: false},
		{input: "", want: false},
	}
	for _, tc := range tests {
		if got := isCompletedTasksThisWeekQuestion(tc.input); got != tc.want {
			t.Fatalf("isCompletedTasksThisWeekQuestion(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestIsWeeklyTaskStatusQuestion(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: "tell me about incomplete, completed and archived tasks this week", want: true},
		{input: "show open tasks this week", want: true},
		{input: "what tasks this week", want: false},
		{input: "completed tasks today", want: false},
	}
	for _, tc := range tests {
		if got := isWeeklyTaskStatusQuestion(tc.input); got != tc.want {
			t.Fatalf("isWeeklyTaskStatusQuestion(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestStartOfWeekMonday(t *testing.T) {
	monday := time.Date(2026, time.February, 2, 15, 0, 0, 0, time.Local)
	if got := startOfWeekMonday(monday); !sameDay(got, dateOnly(monday)) {
		t.Fatalf("expected Monday to remain unchanged, got %s", got.Format("2006-01-02"))
	}

	wednesday := time.Date(2026, time.February, 4, 15, 0, 0, 0, time.Local)
	got := startOfWeekMonday(wednesday)
	want := time.Date(2026, time.February, 2, 0, 0, 0, 0, time.Local)
	if !sameDay(got, want) {
		t.Fatalf("expected %s, got %s", want.Format("2006-01-02"), got.Format("2006-01-02"))
	}

	sunday := time.Date(2026, time.February, 8, 15, 0, 0, 0, time.Local)
	got = startOfWeekMonday(sunday)
	want = time.Date(2026, time.February, 2, 0, 0, 0, 0, time.Local)
	if !sameDay(got, want) {
		t.Fatalf("expected %s, got %s", want.Format("2006-01-02"), got.Format("2006-01-02"))
	}
}

func TestFormatRecentChatHistory(t *testing.T) {
	history := []AIChatMessage{
		{Role: "user", Content: "First question"},
		{Role: "assistant", Content: "First answer"},
		{Role: "user", Content: "Second question"},
	}
	result := formatRecentChatHistory(history, 2)
	if !strings.Contains(result, "assistant: First answer") {
		t.Fatalf("expected assistant line in history, got %q", result)
	}
	if !strings.Contains(result, "user: Second question") {
		t.Fatalf("expected last user line in history, got %q", result)
	}
	if strings.Contains(result, "First question") {
		t.Fatalf("did not expect truncated-out first question in history, got %q", result)
	}
}
