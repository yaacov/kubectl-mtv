package output

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/yaacov/kubectl-mtv/pkg/util/query"
)

func TestPrintMarkdown_Basic(t *testing.T) {
	var buf bytes.Buffer

	printer := NewTablePrinter().
		WithWriter(&buf).
		WithColumns(
			Column{Title: "NAME", Key: "name"},
			Column{Title: "TYPE", Key: "type"},
			Column{Title: "STATUS", Key: "status"},
		).
		AddItems([]map[string]interface{}{
			{"name": "provider1", "type": "vsphere", "status": "Ready"},
			{"name": "provider2", "type": "ovirt", "status": "Not Ready"},
		})

	if err := printer.PrintMarkdown(); err != nil {
		t.Fatalf("PrintMarkdown returned error: %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")

	if len(lines) != 4 {
		t.Fatalf("expected 4 lines (header + sep + 2 rows), got %d:\n%s", len(lines), got)
	}

	if lines[0] != "| NAME | TYPE | STATUS |" {
		t.Errorf("header row = %q, want %q", lines[0], "| NAME | TYPE | STATUS |")
	}
	if lines[1] != "| --- | --- | --- |" {
		t.Errorf("separator row = %q, want %q", lines[1], "| --- | --- | --- |")
	}
	if lines[2] != "| provider1 | vsphere | Ready |" {
		t.Errorf("data row 1 = %q, want %q", lines[2], "| provider1 | vsphere | Ready |")
	}
	if lines[3] != "| provider2 | ovirt | Not Ready |" {
		t.Errorf("data row 2 = %q, want %q", lines[3], "| provider2 | ovirt | Not Ready |")
	}
}

func TestPrintMarkdown_PipeEscaping(t *testing.T) {
	var buf bytes.Buffer

	printer := NewTablePrinter().
		WithWriter(&buf).
		WithColumns(
			Column{Title: "NAME", Key: "name"},
			Column{Title: "DESC", Key: "desc"},
		).
		AddItems([]map[string]interface{}{
			{"name": "test", "desc": "a|b|c"},
		})

	if err := printer.PrintMarkdown(); err != nil {
		t.Fatalf("PrintMarkdown returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, `a\|b\|c`) {
		t.Errorf("pipe characters not escaped in output:\n%s", got)
	}
}

func TestPrintMarkdown_NoANSI(t *testing.T) {
	var buf bytes.Buffer

	printer := NewTablePrinter().
		WithWriter(&buf).
		WithColumns(
			Column{Title: "STATUS", Key: "status"},
		).
		AddItems([]map[string]interface{}{
			{"status": "\033[32mReady\033[0m"},
		})

	if err := printer.PrintMarkdown(); err != nil {
		t.Fatalf("PrintMarkdown returned error: %v", err)
	}

	got := buf.String()
	if strings.Contains(got, "\033[") {
		t.Errorf("ANSI codes found in markdown output:\n%s", got)
	}
	if !strings.Contains(got, "Ready") {
		t.Errorf("value 'Ready' not found in output:\n%s", got)
	}
}

func TestPrintMarkdown_EmptyItems(t *testing.T) {
	var buf bytes.Buffer

	printer := NewTablePrinter().
		WithWriter(&buf).
		WithColumns(
			Column{Title: "NAME", Key: "name"},
		)

	if err := printer.PrintMarkdown(); err != nil {
		t.Fatalf("PrintMarkdown returned error: %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + sep), got %d:\n%s", len(lines), got)
	}
}

func TestPrintMarkdown_NoColumns(t *testing.T) {
	var buf bytes.Buffer

	printer := NewTablePrinter().WithWriter(&buf)

	if err := printer.PrintMarkdown(); err != nil {
		t.Fatalf("PrintMarkdown returned error: %v", err)
	}

	got := buf.String()
	if got != "" {
		t.Errorf("expected empty output for no columns, got:\n%s", got)
	}
}

func TestPrintMarkdownWithQuery_SelectColumns(t *testing.T) {
	var buf bytes.Buffer

	data := []map[string]interface{}{
		{"name": "vm1", "cpu": 4},
		{"name": "vm2", "cpu": 8},
	}

	queryOpts := &query.QueryOptions{
		HasSelect: true,
		Select: []query.SelectOption{
			{Alias: "name", Field: "name"},
			{Alias: "cpu", Field: "cpu"},
		},
	}

	defaultCols := []Column{
		{Title: "NAME", Key: "name"},
		{Title: "CPU", Key: "cpu"},
	}

	printer := NewTablePrinter().
		WithWriter(&buf).
		WithColumns(
			Column{Title: "name", Key: "name"},
			Column{Title: "cpu", Key: "cpu"},
		).
		WithSelectOptions(queryOpts.Select).
		AddItems(data)

	if err := printer.PrintMarkdown(); err != nil {
		t.Fatalf("PrintMarkdown returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "| name | cpu |") {
		t.Errorf("expected SELECT-derived columns, got:\n%s", got)
	}
	if !strings.Contains(got, "| vm1 | 4 |") {
		t.Errorf("expected data row for vm1, got:\n%s", got)
	}

	// Test the top-level PrintMarkdownWithQuery wrapper via stdout capture
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStdout := os.Stdout
	defer func() { os.Stdout = origStdout }()
	defer r.Close()
	os.Stdout = w

	if err := PrintMarkdownWithQuery(data, defaultCols, queryOpts, ""); err != nil {
		t.Fatalf("PrintMarkdownWithQuery returned error: %v", err)
	}
	w.Close()
	os.Stdout = origStdout

	var captured bytes.Buffer
	if _, err := captured.ReadFrom(r); err != nil {
		t.Fatalf("reading pipe: %v", err)
	}

	wrapperGot := captured.String()
	if !strings.Contains(wrapperGot, "| name | cpu |") {
		t.Errorf("wrapper: expected SELECT-derived columns, got:\n%s", wrapperGot)
	}
	if !strings.Contains(wrapperGot, "| vm1 | 4 |") {
		t.Errorf("wrapper: expected data row for vm1, got:\n%s", wrapperGot)
	}
}
