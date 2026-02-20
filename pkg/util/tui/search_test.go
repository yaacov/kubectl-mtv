package tui

import (
	"strings"
	"testing"
)

func TestFindMatches_Empty(t *testing.T) {
	matches := findMatches("hello world", "")
	if len(matches) != 0 {
		t.Errorf("expected no matches for empty term, got %d", len(matches))
	}
}

func TestFindMatches_SingleLine(t *testing.T) {
	matches := findMatches("Hello World Hello", "hello")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].Line != 0 || matches[0].Column != 0 {
		t.Errorf("first match at wrong position: %+v", matches[0])
	}
	if matches[1].Line != 0 || matches[1].Column != 12 {
		t.Errorf("second match at wrong position: %+v", matches[1])
	}
}

func TestFindMatches_MultiLine(t *testing.T) {
	content := "line one\nline two with match\nline three match again"
	matches := findMatches(content, "match")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].Line != 1 {
		t.Errorf("expected match on line 1, got %d", matches[0].Line)
	}
	if matches[1].Line != 2 {
		t.Errorf("expected match on line 2, got %d", matches[1].Line)
	}
}

func TestFindMatches_CaseInsensitive(t *testing.T) {
	matches := findMatches("FoO bAr FOO", "foo")
	if len(matches) != 2 {
		t.Fatalf("expected 2 case-insensitive matches, got %d", len(matches))
	}
}

func TestFindMatches_NoMatch(t *testing.T) {
	matches := findMatches("hello world", "xyz")
	if len(matches) != 0 {
		t.Errorf("expected no matches, got %d", len(matches))
	}
}

func TestFindMatches_WithAnsiCodes(t *testing.T) {
	content := "\033[31mError\033[0m: something failed"
	matches := findMatches(content, "error")
	if len(matches) != 1 {
		t.Fatalf("expected 1 match with ANSI codes, got %d", len(matches))
	}
	if matches[0].Line != 0 {
		t.Errorf("expected match on line 0, got %d", matches[0].Line)
	}
}

func TestHighlightContent_Empty(t *testing.T) {
	result := highlightContent("hello", "")
	if result != "hello" {
		t.Errorf("expected unchanged content for empty term, got %q", result)
	}
}

func TestHighlightContent_Simple(t *testing.T) {
	result := highlightContent("hello world", "world")
	if !strings.Contains(result, highlightOn+"world"+highlightOff) {
		t.Errorf("expected highlighted 'world', got %q", result)
	}
	if !strings.HasPrefix(result, "hello ") {
		t.Errorf("expected prefix preserved, got %q", result)
	}
}

func TestHighlightContent_CaseInsensitive(t *testing.T) {
	result := highlightContent("Hello HELLO", "hello")
	// Both occurrences should be highlighted (preserving original case)
	if count := strings.Count(result, highlightOn); count != 2 {
		t.Errorf("expected 2 highlights, got %d in %q", count, result)
	}
}

func TestHighlightContent_MultiLine(t *testing.T) {
	content := "foo bar\nbaz foo\nqux"
	result := highlightContent(content, "foo")
	lines := strings.Split(result, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], highlightOn) {
		t.Error("expected highlight on line 0")
	}
	if !strings.Contains(lines[1], highlightOn) {
		t.Error("expected highlight on line 1")
	}
	if strings.Contains(lines[2], highlightOn) {
		t.Error("unexpected highlight on line 2")
	}
}

func TestHighlightContent_WithAnsi(t *testing.T) {
	content := "\033[32mSuccess\033[0m: test passed"
	result := highlightContent(content, "success")
	if !strings.Contains(result, highlightOn) {
		t.Error("expected highlight in ANSI-containing text")
	}
	// The ANSI codes should still be present
	if !strings.Contains(result, "\033[32m") {
		t.Error("ANSI codes should be preserved")
	}
}

func TestHighlightContent_NoMatch(t *testing.T) {
	content := "hello world"
	result := highlightContent(content, "xyz")
	if result != content {
		t.Errorf("expected unchanged content when no match, got %q", result)
	}
}

func TestLineForMatch_Empty(t *testing.T) {
	line := lineForMatch(nil, 0)
	if line != 0 {
		t.Errorf("expected 0 for empty matches, got %d", line)
	}
}

func TestLineForMatch_OutOfBounds(t *testing.T) {
	matches := []matchInfo{{Line: 5, Column: 0}}
	if lineForMatch(matches, -1) != 0 {
		t.Error("expected 0 for negative index")
	}
	if lineForMatch(matches, 1) != 0 {
		t.Error("expected 0 for out-of-bounds index")
	}
}

func TestLineForMatch_Valid(t *testing.T) {
	matches := []matchInfo{
		{Line: 2, Column: 0},
		{Line: 5, Column: 3},
		{Line: 10, Column: 1},
	}
	if got := lineForMatch(matches, 0); got != 2 {
		t.Errorf("expected line 2, got %d", got)
	}
	if got := lineForMatch(matches, 1); got != 5 {
		t.Errorf("expected line 5, got %d", got)
	}
	if got := lineForMatch(matches, 2); got != 10 {
		t.Errorf("expected line 10, got %d", got)
	}
}

func TestHighlightContentWithFocus_FocusedMatchGetsDistinctStyle(t *testing.T) {
	content := "foo bar foo baz foo"
	matches := findMatches(content, "foo")
	if len(matches) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(matches))
	}

	// Focus on the second match (index 1)
	result := highlightContentWithFocus(content, "foo", matches, 1)

	// First and third match should use regular highlight
	if strings.Count(result, highlightOn) != 2 {
		t.Errorf("expected 2 regular highlights, got %d in %q", strings.Count(result, highlightOn), result)
	}
	// Second match should use focus highlight
	if !strings.Contains(result, focusOn) {
		t.Errorf("expected focus highlight, got %q", result)
	}
	if strings.Count(result, focusOn) != 1 {
		t.Errorf("expected exactly 1 focus highlight, got %d", strings.Count(result, focusOn))
	}
}

func TestHighlightContentWithFocus_NoFocus(t *testing.T) {
	content := "foo bar foo"
	matches := findMatches(content, "foo")

	// focusIndex = -1 means no focus — same as regular highlightContent
	result := highlightContentWithFocus(content, "foo", matches, -1)
	regular := highlightContent(content, "foo")
	if result != regular {
		t.Errorf("expected same result as highlightContent when focusIndex is -1")
	}
}

func TestHighlightContentWithFocus_MultiLine(t *testing.T) {
	content := "alpha\nbeta alpha\ngamma"
	matches := findMatches(content, "alpha")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}

	// Focus the second match (on line 1)
	result := highlightContentWithFocus(content, "alpha", matches, 1)
	lines := strings.Split(result, "\n")

	// Line 0 should have regular highlight
	if !strings.Contains(lines[0], highlightOn) {
		t.Error("expected regular highlight on line 0")
	}
	if strings.Contains(lines[0], focusOn) {
		t.Error("line 0 should not have focus highlight")
	}

	// Line 1 should have focus highlight
	if !strings.Contains(lines[1], focusOn) {
		t.Error("expected focus highlight on line 1")
	}
}

func TestStripAnsiCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain text", "hello", "hello"},
		{"single code", "\033[31mred\033[0m", "red"},
		{"multiple codes", "\033[1m\033[32mbold green\033[0m", "bold green"},
		{"no content", "\033[0m", ""},
		{"mixed", "before\033[31m middle \033[0mafter", "before middle after"},
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
