package api

import (
	"strings"
)

type codeBlockTracker struct {
	inFence      bool
	fenceChar    byte
	fenceLen     int
	prevBlank    bool
	prevIndented bool
}

func (t *codeBlockTracker) isCodeLine(line string) bool {
	line = strings.TrimSuffix(line, "\r")
	if t.inFence {
		if isFenceDelimiter(line, t.fenceChar, t.fenceLen) {
			t.inFence = false
		}
		return true
	}

	if char, count, ok := fenceStart(line); ok {
		t.inFence = true
		t.fenceChar = char
		t.fenceLen = count
		return true
	}

	indented := isIndentedCodeLine(line)
	if indented && (t.prevBlank || t.prevIndented) {
		t.prevBlank = strings.TrimSpace(line) == ""
		t.prevIndented = true
		return true
	}

	t.prevBlank = strings.TrimSpace(line) == ""
	t.prevIndented = false
	return false
}

func isIndentedCodeLine(line string) bool {
	if line == "" {
		return false
	}
	if strings.HasPrefix(line, "\t") {
		return true
	}
	spaceCount := 0
	for i := 0; i < len(line); i++ {
		if line[i] == ' ' {
			spaceCount++
			if spaceCount >= 4 {
				return true
			}
			continue
		}
		break
	}
	return false
}

func fenceStart(line string) (byte, int, bool) {
	trimmed := strings.TrimLeft(line, " \t")
	if len(trimmed) < 3 {
		return 0, 0, false
	}
	char := trimmed[0]
	if char != '`' && char != '~' {
		return 0, 0, false
	}
	count := 0
	for i := 0; i < len(trimmed) && trimmed[i] == char; i++ {
		count++
	}
	if count < 3 {
		return 0, 0, false
	}
	return char, count, true
}

func isFenceDelimiter(line string, char byte, minLen int) bool {
	trimmed := strings.TrimLeft(line, " \t")
	if len(trimmed) < minLen {
		return false
	}
	for i := 0; i < len(trimmed) && trimmed[i] == char; i++ {
		if i+1 >= minLen {
			return true
		}
	}
	return false
}

func stripCodeBlocksAndInline(content string) string {
	lines := strings.Split(content, "\n")
	var b strings.Builder
	tracker := &codeBlockTracker{}
	for i, line := range lines {
		if tracker.isCodeLine(line) {
			continue
		}
		clean := stripInlineCode(line)
		if i > 0 && b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(clean)
	}
	return b.String()
}

func stripInlineCode(text string) string {
	var b strings.Builder
	lastWasSpace := false
	for i := 0; i < len(text); {
		if text[i] != '`' {
			b.WriteByte(text[i])
			lastWasSpace = text[i] == ' '
			i++
			continue
		}
		run := countBackticks(text, i)
		if run == 0 {
			b.WriteByte(text[i])
			i++
			continue
		}
		end := findMatchingBackticks(text, i+run, run)
		if end == -1 {
			b.WriteByte(text[i])
			lastWasSpace = text[i] == ' '
			i++
			continue
		}
		if b.Len() > 0 && !lastWasSpace {
			b.WriteByte(' ')
			lastWasSpace = true
		}
		i = end + run
	}
	return b.String()
}

func maskInlineCode(text string) (string, []string) {
	replacements := make([]string, 0)
	var b strings.Builder
	for i := 0; i < len(text); {
		if text[i] != '`' {
			b.WriteByte(text[i])
			i++
			continue
		}
		run := countBackticks(text, i)
		if run == 0 {
			b.WriteByte(text[i])
			i++
			continue
		}
		end := findMatchingBackticks(text, i+run, run)
		if end == -1 {
			b.WriteByte(text[i])
			i++
			continue
		}
		placeholder := "__CODEBLOCK" + itoa(len(replacements)) + "__"
		replacements = append(replacements, text[i:end+run])
		b.WriteString(placeholder)
		i = end + run
	}
	return b.String(), replacements
}

func restoreInlineCode(text string, replacements []string) string {
	restored := text
	for i, value := range replacements {
		placeholder := "__CODEBLOCK" + itoa(i) + "__"
		restored = strings.ReplaceAll(restored, placeholder, value)
	}
	return restored
}

func countBackticks(text string, start int) int {
	count := 0
	for i := start; i < len(text) && text[i] == '`'; i++ {
		count++
	}
	return count
}

func findMatchingBackticks(text string, start int, run int) int {
	for i := start; i < len(text); {
		if text[i] != '`' {
			i++
			continue
		}
		if countBackticks(text, i) == run {
			return i
		}
		i++
	}
	return -1
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[i:])
}
