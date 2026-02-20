package tui

import (
	"strings"
	"testing"
)

func TestStripAnsiCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"no ansi", "hello world", "hello world"},
		{"single code", "\033[31mred\033[0m", "red"},
		{"multiple codes", "\033[1;33mfoo\033[0m bar \033[32mbaz\033[0m", "foo bar baz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAnsiCodes(tt.input)
			if got != tt.expected {
				t.Errorf("stripAnsiCodes(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFindMatches_Empty(t *testing.T) {
	matches := findMatches("some content", "")
	if len(matches) != 0 {
		t.Errorf("expected no matches for empty term, got %d", len(matches))
	}
}

func TestFindMatches_NoMatch(t *testing.T) {
	matches := findMatches("hello world", "xyz")
	if len(matches) != 0 {
		t.Errorf("expected no matches, got %d", len(matches))
	}
}

func TestFindMatches_SingleLine(t *testing.T) {
	matches := findMatches("hello world hello", "hello")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].Line != 0 || matches[0].Column != 0 {
		t.Errorf("first match: got line=%d col=%d, want line=0 col=0", matches[0].Line, matches[0].Column)
	}
	if matches[1].Line != 0 || matches[1].Column != 12 {
		t.Errorf("second match: got line=%d col=%d, want line=0 col=12", matches[1].Line, matches[1].Column)
	}
}

func TestFindMatches_MultiLine(t *testing.T) {
	content := "foo bar\nbaz foo\nqux"
	matches := findMatches(content, "foo")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].Line != 0 || matches[0].Column != 0 {
		t.Errorf("first match: got line=%d col=%d, want line=0 col=0", matches[0].Line, matches[0].Column)
	}
	if matches[1].Line != 1 || matches[1].Column != 4 {
		t.Errorf("second match: got line=%d col=%d, want line=1 col=4", matches[1].Line, matches[1].Column)
	}
}

func TestFindMatches_CaseInsensitive(t *testing.T) {
	matches := findMatches("Hello HELLO hElLo", "hello")
	if len(matches) != 3 {
		t.Errorf("expected 3 case-insensitive matches, got %d", len(matches))
	}
}

func TestFindMatches_WithAnsiCodes(t *testing.T) {
	content := "\033[31mhello\033[0m world \033[32mhello\033[0m"
	matches := findMatches(content, "hello")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches ignoring ANSI, got %d", len(matches))
	}
	if matches[0].Column != 0 {
		t.Errorf("first match col: got %d, want 0", matches[0].Column)
	}
	if matches[1].Column != 12 {
		t.Errorf("second match col: got %d, want 12", matches[1].Column)
	}
}

func TestHighlightContent_Empty(t *testing.T) {
	result := highlightContent("some text", "")
	if result != "some text" {
		t.Errorf("expected unchanged text for empty term")
	}
}

func TestHighlightContent_NoMatch(t *testing.T) {
	result := highlightContent("hello world", "xyz")
	if result != "hello world" {
		t.Errorf("expected unchanged text for no match")
	}
}

func TestHighlightContent_Simple(t *testing.T) {
	result := highlightContent("hello world", "world")
	if !strings.Contains(result, highlightOn+"world"+highlightOff) {
		t.Errorf("expected highlight around 'world', got: %q", result)
	}
}

func TestHighlightContent_CaseInsensitive(t *testing.T) {
	result := highlightContent("Hello World", "hello")
	if !strings.Contains(result, highlightOn) {
		t.Errorf("expected case-insensitive highlight, got: %q", result)
	}
}

func TestHighlightContent_MultiLine(t *testing.T) {
	content := "foo bar\nbaz foo"
	result := highlightContent(content, "foo")
	lines := strings.Split(result, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	for i, line := range lines {
		if !strings.Contains(line, highlightOn) {
			t.Errorf("line %d should contain highlight: %q", i, line)
		}
	}
}

func TestHighlightContent_WithAnsi(t *testing.T) {
	content := "\033[31mhello\033[0m world"
	result := highlightContent(content, "hello")
	stripped := stripAnsiCodes(result)
	if stripped != "hello world" {
		t.Errorf("stripped result should be 'hello world', got %q", stripped)
	}
	if !strings.Contains(result, highlightOn) {
		t.Errorf("result should contain highlight code")
	}
}

func TestHighlightContentWithFocus_FocusedMatch(t *testing.T) {
	content := "hello world hello"
	matches := findMatches(content, "hello")
	result := highlightContentWithFocus(content, "hello", matches, 1)

	if !strings.Contains(result, highlightOn) {
		t.Errorf("should contain regular highlight")
	}
	if !strings.Contains(result, focusOn) {
		t.Errorf("should contain focused highlight")
	}
}

func TestHighlightContentWithFocus_NoFocus(t *testing.T) {
	content := "hello world"
	matches := findMatches(content, "hello")
	result := highlightContentWithFocus(content, "hello", matches, -1)

	if strings.Contains(result, focusOn) {
		t.Errorf("should not contain focus highlight when focusIndex is -1")
	}
	if !strings.Contains(result, highlightOn) {
		t.Errorf("should contain regular highlight")
	}
}

func TestLineForMatch_Empty(t *testing.T) {
	line := lineForMatch(nil, 0)
	if line != 0 {
		t.Errorf("expected 0 for nil matches, got %d", line)
	}
}

func TestLineForMatch_OutOfBounds(t *testing.T) {
	matches := []matchInfo{{Line: 5, Column: 0}}
	line := lineForMatch(matches, 5)
	if line != 0 {
		t.Errorf("expected 0 for out-of-bounds index, got %d", line)
	}
}

func TestLineForMatch_Valid(t *testing.T) {
	matches := []matchInfo{
		{Line: 0, Column: 0},
		{Line: 3, Column: 5},
		{Line: 7, Column: 2},
	}
	if got := lineForMatch(matches, 1); got != 3 {
		t.Errorf("lineForMatch(_, 1) = %d, want 3", got)
	}
	if got := lineForMatch(matches, 2); got != 7 {
		t.Errorf("lineForMatch(_, 2) = %d, want 7", got)
	}
}

func TestBuildIndexMapping(t *testing.T) {
	line := "\033[31mhi\033[0m"
	indexMap := buildIndexMapping(line)
	stripped := stripAnsiCodes(line)

	if len(indexMap) != len(stripped)+1 {
		t.Fatalf("indexMap length %d, expected %d", len(indexMap), len(stripped)+1)
	}
	if indexMap[0] != 5 {
		t.Errorf("indexMap[0] = %d, want 5", indexMap[0])
	}
}
