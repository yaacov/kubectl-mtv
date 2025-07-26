package output

// PrintTable prints the given data as a table using TablePrinter and headers
import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/yaacov/kubectl-mtv/pkg/util/query"
)

// PrintTableWithQuery prints the given data as a table using TablePrinter,
// supporting dynamic headers from query options and empty message handling.
func PrintTableWithQuery(data interface{}, defaultHeaders []Header, queryOpts *query.QueryOptions, emptyMessage string) error {
	items, ok := data.([]map[string]interface{})
	if !ok {
		if item, ok := data.(map[string]interface{}); ok {
			// Handle single item map
			items = []map[string]interface{}{item}
		} else if slice, ok := data.([]interface{}); ok {
			// Handle []interface{} from JSON unmarshaling
			items = make([]map[string]interface{}, len(slice))
			for i, item := range slice {
				if mapItem, ok := item.(map[string]interface{}); ok {
					items[i] = mapItem
				} else {
					return fmt.Errorf("unsupported data type for table output: slice contains non-map elements")
				}
			}
		} else {
			return fmt.Errorf("unsupported data type for table output")
		}
	}

	var printer *TablePrinter

	// Check if we should use custom headers from SELECT clause
	if queryOpts != nil && queryOpts.HasSelect {
		headers := make([]Header, 0, len(queryOpts.Select))
		for _, sel := range queryOpts.Select {
			headers = append(headers, Header{
				DisplayName: sel.Alias,
				JSONPath:    sel.Alias,
			})
		}
		printer = NewTablePrinter().
			WithHeaders(headers...).
			WithSelectOptions(queryOpts.Select)
	} else {
		// Use the provided default headers
		printer = NewTablePrinter().WithHeaders(defaultHeaders...)
	}

	if len(items) == 0 && emptyMessage != "" {
		return printer.PrintEmpty(emptyMessage)
	}

	printer.AddItems(items)
	return printer.Print()
}

// Header represents a table column header with display text and a JSON path
type Header struct {
	DisplayName string
	JSONPath    string
}

// TablePrinter prints tabular data with dynamically sized columns
type TablePrinter struct {
	headers       []Header
	items         []map[string]interface{}
	padding       int
	minWidth      int
	writer        io.Writer
	maxColWidth   int
	expandedData  map[int]string       // Stores expanded data for each row by index
	selectOptions []query.SelectOption // Optional: select options for advanced extraction
}

// NewTablePrinter creates a new TablePrinter
func NewTablePrinter() *TablePrinter {
	return &TablePrinter{
		headers:      []Header{},
		items:        []map[string]interface{}{},
		padding:      2,
		minWidth:     10,
		writer:       os.Stdout,
		maxColWidth:  50, // Prevent very wide columns
		expandedData: make(map[int]string),
	}
}

// WithHeaders sets the table headers with display names and JSON paths
func (t *TablePrinter) WithHeaders(headers ...Header) *TablePrinter {
	t.headers = headers
	return t
}

// WithPadding sets the padding between columns
func (t *TablePrinter) WithPadding(padding int) *TablePrinter {
	t.padding = padding
	return t
}

// WithMinWidth sets the minimum column width
func (t *TablePrinter) WithMinWidth(minWidth int) *TablePrinter {
	t.minWidth = minWidth
	return t
}

// WithMaxWidth sets the maximum column width
func (t *TablePrinter) WithMaxWidth(maxWidth int) *TablePrinter {
	t.maxColWidth = maxWidth
	return t
}

// WithWriter sets the output writer
func (t *TablePrinter) WithWriter(writer io.Writer) *TablePrinter {
	t.writer = writer
	return t
}

// WithExpandedData sets expanded data for a specific row index
func (t *TablePrinter) WithExpandedData(index int, data string) *TablePrinter {
	t.expandedData[index] = data
	return t
}

// WithSelectOptions sets the select options for the table printer
func (t *TablePrinter) WithSelectOptions(selectOptions []query.SelectOption) *TablePrinter {
	t.selectOptions = selectOptions
	return t
}

// AddItem adds an item to the table
func (t *TablePrinter) AddItem(item map[string]interface{}) *TablePrinter {
	t.items = append(t.items, item)
	return t
}

// AddItemWithExpanded adds an item to the table with expanded data
func (t *TablePrinter) AddItemWithExpanded(item map[string]interface{}, expanded string) *TablePrinter {
	index := len(t.items)
	t.items = append(t.items, item)
	t.expandedData[index] = expanded
	return t
}

// AddItems adds multiple items to the table
func (t *TablePrinter) AddItems(items []map[string]interface{}) *TablePrinter {
	t.items = append(t.items, items...)
	return t
}

// extractValue extracts a value from an item using a JSON path
func (t *TablePrinter) extractValue(item map[string]interface{}, path string) string {
	if path == "" {
		// No path provided, return empty string
		return ""
	}

	// Use query.GetValue if selectOptions are set, otherwise fallback to GetValueByPathString
	if len(t.selectOptions) > 0 {
		val, err := query.GetValue(item, path, t.selectOptions)
		if err != nil {
			return ""
		}
		return valueToString(val)
	}

	value, err := query.GetValueByPathString(item, path)
	if err != nil {
		return ""
	}

	return valueToString(value)
}

// valueToString converts a value of any supported type to a string
func valueToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case float32:
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		// For other types, use default string conversion
		return fmt.Sprintf("%v", v)
	}
}

// calculateColumnWidths determines the optimal width for each column
func (t *TablePrinter) calculateColumnWidths() []int {
	numCols := len(t.headers)
	if numCols == 0 {
		return []int{}
	}

	// Initialize widths with minimum values
	widths := make([]int, numCols)
	for i := range widths {
		widths[i] = t.minWidth
	}

	// Check header widths
	for i, header := range t.headers {
		headerWidth := utf8.RuneCountInString(header.DisplayName)
		if headerWidth > widths[i] {
			widths[i] = min(headerWidth, t.maxColWidth)
		}
	}

	// Calculate row data for width determination
	for _, item := range t.items {
		for i, header := range t.headers {
			value := t.extractValue(item, header.JSONPath)
			cellWidth := utf8.RuneCountInString(value)
			if cellWidth > widths[i] {
				widths[i] = min(cellWidth, t.maxColWidth)
			}
		}
	}

	return widths
}

// Print prints the table with dynamic column widths
func (t *TablePrinter) Print() error {
	widths := t.calculateColumnWidths()
	if len(widths) == 0 {
		return nil
	}

	// Print headers
	headerRow := make([]string, len(t.headers))
	for i, header := range t.headers {
		headerRow[i] = header.DisplayName
	}
	t.printRow(headerRow, widths)

	// Print item rows and expanded data if available
	for i, item := range t.items {
		row := make([]string, len(t.headers))
		for j, header := range t.headers {
			row[j] = t.extractValue(item, header.JSONPath)
		}
		t.printRow(row, widths)

		// Print expanded data if it exists for this row
		if expanded, exists := t.expandedData[i]; exists && expanded != "" {
			// Split expanded data into lines and add prefix
			lines := strings.Split(expanded, "\n")
			for _, line := range lines {
				fmt.Fprintf(t.writer, "  │ %s\n", line)
			}
		}
	}

	return nil
}

// PrintEmpty prints a message when there are no items to display
func (t *TablePrinter) PrintEmpty(message string) error {
	fmt.Fprintln(t.writer, message)
	return nil
}

// printRow prints a single row with the specified column widths
func (t *TablePrinter) printRow(row []string, widths []int) {
	var sb strings.Builder

	for i, cell := range row {
		if i >= len(widths) {
			break
		}

		// Truncate if the cell is too long
		displayCell := cell
		if utf8.RuneCountInString(cell) > t.maxColWidth {
			displayCell = cell[:t.maxColWidth-3] + "..."
		}

		// Format with proper padding
		format := fmt.Sprintf("%%-%ds", widths[i]+t.padding)
		sb.WriteString(fmt.Sprintf(format, displayCell))
	}

	fmt.Fprintln(t.writer, strings.TrimRight(sb.String(), " "))
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
