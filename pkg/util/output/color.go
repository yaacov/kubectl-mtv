package output

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ANSI color codes
const (
	Reset       = "\033[0m"
	BoldText    = "\033[1m"
	RedColor    = "\033[31m"
	GreenColor  = "\033[32m"
	YellowColor = "\033[33m"
	BlueColor   = "\033[34m"
	PurpleColor = "\033[35m"
	CyanColor   = "\033[36m"
	White       = "\033[37m"
	BoldRed     = "\033[1;31m"
	BoldGreen   = "\033[1;32m"
	BoldYellow  = "\033[1;33m"
	BoldBlue    = "\033[1;34m"
)

// ansiRegex is a regular expression that matches ANSI color escape codes
var ansiRegex = regexp.MustCompile("\033\\[[0-9;]*m")

// Bold returns a bold-formatted string
func Bold(text string) string {
	return ColorizedString(text, BoldText)
}

// ColorizedString returns a string with the specified color applied
func ColorizedString(text string, color string) string {
	return color + text + Reset
}

// Yellow returns a yellow-colored string
func Yellow(text string) string {
	return ColorizedString(text, YellowColor)
}

// Green returns a green-colored string
func Green(text string) string {
	return ColorizedString(text, GreenColor)
}

// Red returns a red-colored string
func Red(text string) string {
	return ColorizedString(text, RedColor)
}

// Blue returns a blue-colored string
func Blue(text string) string {
	return ColorizedString(text, BlueColor)
}

// Cyan returns a cyan-colored string
func Cyan(text string) string {
	return ColorizedString(text, CyanColor)
}

// StripANSI removes ANSI color codes from a string
func StripANSI(text string) string {
	return ansiRegex.ReplaceAllString(text, "")
}

// VisibleLength returns the visible rune count of a string, excluding ANSI color codes
func VisibleLength(text string) int {
	return utf8.RuneCountInString(StripANSI(text))
}

// ColorizeStatus returns a colored string based on status value
func ColorizeStatus(status string) string {
	status = strings.TrimSpace(status)
	switch strings.ToLower(status) {
	case "running":
		return Blue(status)
	case "executing":
		return Blue(status)
	case "completed", "succeeded":
		return Green(status)
	case "pending":
		return Yellow(status)
	case "failed":
		return Red(status)
	case "canceled":
		return Cyan(status)
	default:
		return status
	}
}

// ColorizePowerState returns a colored string based on VM power state
func ColorizePowerState(state string) string {
	state = strings.TrimSpace(state)
	switch strings.ToLower(state) {
	case "running":
		return Green(state)
	case "stopped":
		return Yellow(state)
	case "not found":
		return Red(state)
	default:
		return state
	}
}

// ColorizeNumber returns a blue-colored number for migration progress
func ColorizeNumber(number interface{}) string {
	return Blue(fmt.Sprintf("%v", number))
}

// ColorizeBoolean returns a colored string based on boolean value
func ColorizeBoolean(b bool) string {
	if b {
		return Green(fmt.Sprintf("%t", b))
	}
	return fmt.Sprintf("%t", b)
}

// TruncateANSI truncates text to maxWidth visible characters while preserving
// ANSI color codes. Appends "..." and a Reset code when truncation occurs.
func TruncateANSI(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if VisibleLength(text) <= maxWidth {
		return text
	}

	truncWidth := maxWidth - 3
	if truncWidth < 0 {
		truncWidth = 0
	}

	var result strings.Builder
	visCount := 0
	runes := []rune(text)
	i := 0

	for i < len(runes) && visCount < truncWidth {
		if runes[i] == '\033' {
			for i < len(runes) {
				result.WriteRune(runes[i])
				if runes[i] == 'm' {
					i++
					break
				}
				i++
			}
			continue
		}
		result.WriteRune(runes[i])
		visCount++
		i++
	}

	result.WriteString(Reset)
	if truncWidth < maxWidth {
		result.WriteString("...")
	}
	return result.String()
}

// ColorizedSeparator returns a separator line with the specified color
func ColorizedSeparator(length int, color string) string {
	return ColorizedString(strings.Repeat("=", length), color)
}
