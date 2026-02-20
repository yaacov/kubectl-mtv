package tui

import (
	"strings"
)

// ANSI escape codes for search highlighting
const (
	highlightOn  = "\033[7m"      // reverse video for regular matches
	highlightOff = "\033[27m"     // reset reverse video
	focusOn      = "\033[7;33;1m" // reverse video + bold yellow for the focused match
	focusOff     = "\033[0m"      // full reset after focus highlight
)

// matchInfo stores the position of a search match.
type matchInfo struct {
	Line   int
	Column int
}

// findMatches returns all match positions for a case-insensitive search in content.
func findMatches(content, term string) []matchInfo {
	if term == "" {
		return nil
	}

	lowerTerm := strings.ToLower(term)
	lines := strings.Split(content, "\n")
	var matches []matchInfo

	for lineIdx, line := range lines {
		lowerLine := strings.ToLower(stripAnsiCodes(line))
		start := 0
		for {
			idx := strings.Index(lowerLine[start:], lowerTerm)
			if idx < 0 {
				break
			}
			matches = append(matches, matchInfo{Line: lineIdx, Column: start + idx})
			start += idx + len(lowerTerm)
		}
	}

	return matches
}

// highlightContent wraps all case-insensitive matches of term with ANSI reverse-video codes.
// It works on the visible text while preserving existing ANSI escape sequences.
func highlightContent(content, term string) string {
	return highlightContentWithFocus(content, term, nil, -1)
}

// highlightContentWithFocus highlights all matches and applies a distinct style to the
// focused match (identified by focusIndex into the global matches slice).
// When focusIndex < 0, all matches use the same regular highlight.
func highlightContentWithFocus(content, term string, matches []matchInfo, focusIndex int) string {
	if term == "" {
		return content
	}

	lowerTerm := strings.ToLower(term)
	lines := strings.Split(content, "\n")
	result := make([]string, len(lines))

	// Build a set of (line, column) for the focused match so highlightLine can identify it
	focusLine, focusCol := -1, -1
	if focusIndex >= 0 && focusIndex < len(matches) {
		focusLine = matches[focusIndex].Line
		focusCol = matches[focusIndex].Column
	}

	for i, line := range lines {
		result[i] = highlightLine(line, lowerTerm, i, focusLine, focusCol)
	}

	return strings.Join(result, "\n")
}

// highlightLine highlights all occurrences of lowerTerm in a single line,
// accounting for ANSI escape sequences in the original text.
// focusLine/focusCol identify the single match that should use the focus style.
func highlightLine(line, lowerTerm string, lineIdx, focusLine, focusCol int) string {
	stripped := stripAnsiCodes(line)
	lowerStripped := strings.ToLower(stripped)

	if !strings.Contains(lowerStripped, lowerTerm) {
		return line
	}

	// Build a mapping from stripped-text index to original-text index
	strippedToOrig := buildIndexMapping(line)

	// Find match positions in stripped text
	type span struct {
		start, end int
		focused    bool
	}
	var spans []span
	start := 0
	for {
		idx := strings.Index(lowerStripped[start:], lowerTerm)
		if idx < 0 {
			break
		}
		s := start + idx
		isFocused := lineIdx == focusLine && s == focusCol
		spans = append(spans, span{s, s + len(lowerTerm), isFocused})
		start = s + len(lowerTerm)
	}

	if len(spans) == 0 {
		return line
	}

	// Insert highlight codes at the original positions
	var b strings.Builder
	origIdx := 0
	for _, sp := range spans {
		origStart := strippedToOrig[sp.start]
		origEnd := strippedToOrig[sp.end]
		b.WriteString(line[origIdx:origStart])
		if sp.focused {
			b.WriteString(focusOn)
			b.WriteString(line[origStart:origEnd])
			b.WriteString(focusOff)
		} else {
			b.WriteString(highlightOn)
			b.WriteString(line[origStart:origEnd])
			b.WriteString(highlightOff)
		}
		origIdx = origEnd
	}
	b.WriteString(line[origIdx:])

	return b.String()
}

// buildIndexMapping creates a mapping from stripped-text byte index to original-text byte index.
// For a stripped string of length N, the returned slice has N+1 entries (including end position).
func buildIndexMapping(line string) []int {
	mapping := make([]int, 0, len(line)+1)
	inEscape := false

	for i := 0; i < len(line); i++ {
		if line[i] == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if line[i] == 'm' {
				inEscape = false
			}
			continue
		}
		mapping = append(mapping, i)
	}
	// Sentinel for end-of-string
	mapping = append(mapping, len(line))

	return mapping
}

// lineForMatch returns the line number for a given match index.
func lineForMatch(matches []matchInfo, index int) int {
	if len(matches) == 0 || index < 0 || index >= len(matches) {
		return 0
	}
	return matches[index].Line
}

// stripAnsiCodes removes ANSI escape sequences for plain-text operations.
func stripAnsiCodes(s string) string {
	var b strings.Builder
	inEscape := false

	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		b.WriteRune(r)
	}

	return b.String()
}
