package api

import (
	"strings"
	"testing"
)

func TestParseTodoLinesIgnoresCodeBlocksAndInlineMetadata(t *testing.T) {
	content := strings.Join([]string{
		"```",
		"- [ ] Code task #code @dev +proj >2026-02-01 ^3",
		"```",
		"",
		"- [ ] Real task `@inline` #real +Proj >2026-02-01 ^2",
		"",
		"    - [ ] Indented code task #nope @nope",
		"- [ ] Task with `code #inside` and @outside #outside",
	}, "\n")

	todos := parseTodoLines(content)
	if len(todos) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(todos))
	}

	first := todos[0]
	if first.LineNumber != 5 {
		t.Fatalf("expected first task on line 5, got %d", first.LineNumber)
	}
	if first.Project != "proj" {
		t.Fatalf("expected project proj, got %q", first.Project)
	}
	if len(first.Tags) != 1 || first.Tags[0] != "real" {
		t.Fatalf("expected tags [real], got %#v", first.Tags)
	}
	if len(first.Mentions) != 0 {
		t.Fatalf("expected no mentions, got %#v", first.Mentions)
	}
	if first.DueDateRaw != "2026-02-01" {
		t.Fatalf("expected due date raw 2026-02-01, got %q", first.DueDateRaw)
	}
	if first.Priority != 2 {
		t.Fatalf("expected priority 2, got %d", first.Priority)
	}
	if !strings.Contains(first.Text, "`@inline`") {
		t.Fatalf("expected text to preserve inline code, got %q", first.Text)
	}
	if strings.Contains(first.Text, "#real") {
		t.Fatalf("expected metadata to be stripped from text, got %q", first.Text)
	}

	second := todos[1]
	if second.LineNumber != 8 {
		t.Fatalf("expected second task on line 8, got %d", second.LineNumber)
	}
	if len(second.Tags) != 1 || second.Tags[0] != "outside" {
		t.Fatalf("expected tags [outside], got %#v", second.Tags)
	}
	if len(second.Mentions) != 1 || second.Mentions[0] != "outside" {
		t.Fatalf("expected mentions [outside], got %#v", second.Mentions)
	}
	if !strings.Contains(second.Text, "`code #inside`") {
		t.Fatalf("expected text to preserve inline code, got %q", second.Text)
	}
	if strings.Contains(second.Text, "#outside") || strings.Contains(second.Text, "@outside") {
		t.Fatalf("expected metadata removed from text, got %q", second.Text)
	}
}
