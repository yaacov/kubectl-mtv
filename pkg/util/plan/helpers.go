package plan

import (
	"fmt"
	"strings"
	"time"
)

// FormatTime formats a timestamp string
func FormatTime(timestamp string) string {
	if timestamp == "" {
		return "N/A"
	}

	// Parse the timestamp
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}

	// Format as "2006-01-02 15:04:05"
	return t.Format("2006-01-02 15:04:05")
}

// PrintTable prints a table with headers and rows
func PrintTable(headers []string, rows [][]string, colWidths []int) {
	// Print headers
	for i, header := range headers {
		fmt.Printf("%-*s", colWidths[i], header)
		if i < len(headers)-1 {
			fmt.Print("  ")
		}
	}
	fmt.Println()

	// Print separator
	for i, width := range colWidths {
		fmt.Print(strings.Repeat("-", width))
		if i < len(colWidths)-1 {
			fmt.Print("  ")
		}
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				fmt.Printf("%-*s", colWidths[i], cell)
				if i < len(row)-1 {
					fmt.Print("  ")
				}
			}
		}
		fmt.Println()
	}
}
