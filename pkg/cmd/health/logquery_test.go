package health

import (
	"strings"
	"testing"
)

// --- FilterByPattern tests ---

func TestFilterByPattern(t *testing.T) {
	logs := "line 1 error found\nline 2 all good\nline 3 ERROR uppercase\nline 4 warning here"

	tests := []struct {
		name       string
		pattern    string
		ignoreCase bool
		wantErr    bool
		wantLines  int
		wantMatch  string
	}{
		{
			name:      "empty pattern returns all",
			pattern:   "",
			wantLines: 4,
		},
		{
			name:      "case-sensitive match",
			pattern:   "error",
			wantLines: 1,
			wantMatch: "line 1",
		},
		{
			name:       "case-insensitive match",
			pattern:    "error",
			ignoreCase: true,
			wantLines:  2,
		},
		{
			name:      "regex or pattern",
			pattern:   "error|warning",
			wantLines: 2,
		},
		{
			name:    "invalid regex returns error",
			pattern: "[invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterByPattern(logs, tt.pattern, tt.ignoreCase)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			lines := strings.Split(result, "\n")
			nonEmpty := 0
			for _, l := range lines {
				if l != "" {
					nonEmpty++
				}
			}

			if tt.pattern == "" {
				if result != logs {
					t.Error("empty pattern should return original logs")
				}
			} else if nonEmpty != tt.wantLines {
				t.Errorf("got %d non-empty lines, want %d", nonEmpty, tt.wantLines)
			}

			if tt.wantMatch != "" && !strings.Contains(result, tt.wantMatch) {
				t.Errorf("result should contain %q", tt.wantMatch)
			}
		})
	}
}

// --- LooksLikeJSONLogs tests ---

func TestLooksLikeJSONLogs(t *testing.T) {
	tests := []struct {
		name   string
		logs   string
		expect bool
	}{
		{
			name:   "valid JSON with level and msg",
			logs:   `{"level":"info","ts":"2026-02-05","logger":"plan","msg":"Reconcile started."}`,
			expect: true,
		},
		{
			name:   "timestamp-prefixed JSON",
			logs:   `2026-02-05T10:45:52.123Z {"level":"info","ts":"2026-02-05","logger":"plan","msg":"Started."}`,
			expect: true,
		},
		{
			name:   "plain text logs",
			logs:   "Starting virt-v2v conversion\nDisk 1/1 copied\nConversion complete",
			expect: false,
		},
		{
			name:   "empty string",
			logs:   "",
			expect: false,
		},
		{
			name:   "JSON without level field",
			logs:   `{"ts":"2026-02-05","msg":"no level"}`,
			expect: false,
		},
		{
			name:   "JSON without msg field",
			logs:   `{"level":"info","ts":"2026-02-05"}`,
			expect: false,
		},
		{
			name:   "mixed - JSON line among non-JSON",
			logs:   "some text\n{\"level\":\"info\",\"msg\":\"test\"}\nmore text",
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LooksLikeJSONLogs(tt.logs)
			if result != tt.expect {
				t.Errorf("LooksLikeJSONLogs() = %v, want %v", result, tt.expect)
			}
		})
	}
}

// --- MatchesFilters tests ---

func TestMatchesFilters(t *testing.T) {
	entry := JSONLogEntry{
		Level:     "info",
		Logger:    "plan|abc123",
		Msg:       "Reconcile started",
		Plan:      map[string]string{"name": "my-plan", "namespace": "demo"},
		Provider:  map[string]string{"name": "my-provider", "namespace": "demo"},
		VM:        "vm-001",
		VMName:    "web-server",
		VMID:      "id-123",
		Migration: map[string]string{"name": "migration-xyz"},
	}

	tests := []struct {
		name   string
		params LogFilterParams
		expect bool
	}{
		{
			name:   "no filters matches all",
			params: LogFilterParams{},
			expect: true,
		},
		{
			name:   "filter level matches",
			params: LogFilterParams{FilterLevel: "info"},
			expect: true,
		},
		{
			name:   "filter level mismatch",
			params: LogFilterParams{FilterLevel: "error"},
			expect: false,
		},
		{
			name:   "filter logger matches prefix",
			params: LogFilterParams{FilterLogger: "plan"},
			expect: true,
		},
		{
			name:   "filter logger mismatch",
			params: LogFilterParams{FilterLogger: "provider"},
			expect: false,
		},
		{
			name:   "filter plan matches",
			params: LogFilterParams{FilterPlan: "my-plan"},
			expect: true,
		},
		{
			name:   "filter plan mismatch",
			params: LogFilterParams{FilterPlan: "other-plan"},
			expect: false,
		},
		{
			name:   "filter provider matches",
			params: LogFilterParams{FilterProvider: "my-provider"},
			expect: true,
		},
		{
			name:   "filter provider mismatch",
			params: LogFilterParams{FilterProvider: "other"},
			expect: false,
		},
		{
			name:   "filter VM by VM field",
			params: LogFilterParams{FilterVM: "vm-001"},
			expect: true,
		},
		{
			name:   "filter VM by VMName field",
			params: LogFilterParams{FilterVM: "web-server"},
			expect: true,
		},
		{
			name:   "filter VM by VMID field",
			params: LogFilterParams{FilterVM: "id-123"},
			expect: true,
		},
		{
			name:   "filter VM mismatch",
			params: LogFilterParams{FilterVM: "nonexistent"},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchesFilters(entry, tt.params)
			if result != tt.expect {
				t.Errorf("MatchesFilters() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestMatchesFilters_Migration(t *testing.T) {
	entry := JSONLogEntry{
		Level:     "info",
		Logger:    "migration|migration-xyz",
		Msg:       "Migration started",
		Migration: map[string]string{"name": "migration-xyz"},
	}

	tests := []struct {
		name   string
		params LogFilterParams
		expect bool
	}{
		{
			name:   "migration filter matches from field",
			params: LogFilterParams{FilterMigration: "migration-xyz"},
			expect: true,
		},
		{
			name:   "migration filter mismatch",
			params: LogFilterParams{FilterMigration: "other-migration"},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchesFilters(entry, tt.params)
			if result != tt.expect {
				t.Errorf("MatchesFilters() = %v, want %v", result, tt.expect)
			}
		})
	}

	nonMigrationEntry := JSONLogEntry{
		Level:  "info",
		Logger: "plan|abc",
		Msg:    "test",
	}
	if MatchesFilters(nonMigrationEntry, LogFilterParams{FilterMigration: "any"}) {
		t.Error("non-migration logger should not match migration filter")
	}
}

// --- FilterAndFormatJSONLogs tests ---

func TestFilterAndFormatJSONLogs(t *testing.T) {
	jsonLogs := `{"level":"info","ts":"2026-02-05 10:00:00","logger":"plan|abc","msg":"Started","plan":{"name":"my-plan","namespace":"demo"}}
{"level":"error","ts":"2026-02-05 10:01:00","logger":"plan|abc","msg":"Failed","plan":{"name":"my-plan","namespace":"demo"}}
{"level":"info","ts":"2026-02-05 10:02:00","logger":"provider|xyz","msg":"Refreshed","provider":{"name":"my-provider","namespace":"demo"}}`

	t.Run("json format returns array", func(t *testing.T) {
		result, err := FilterAndFormatJSONLogs(jsonLogs, LogFilterParams{LogFormat: "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		entries, ok := result.([]interface{})
		if !ok {
			t.Fatalf("expected []interface{}, got %T", result)
		}
		if len(entries) != 3 {
			t.Errorf("got %d entries, want 3", len(entries))
		}
	})

	t.Run("text format returns string", func(t *testing.T) {
		result, err := FilterAndFormatJSONLogs(jsonLogs, LogFilterParams{LogFormat: "text"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		str, ok := result.(string)
		if !ok {
			t.Fatalf("expected string, got %T", result)
		}
		if !strings.Contains(str, "Started") {
			t.Error("text output should contain log messages")
		}
	})

	t.Run("pretty format returns formatted string", func(t *testing.T) {
		result, err := FilterAndFormatJSONLogs(jsonLogs, LogFilterParams{LogFormat: "pretty"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		str, ok := result.(string)
		if !ok {
			t.Fatalf("expected string, got %T", result)
		}
		if !strings.Contains(str, "[INFO]") {
			t.Error("pretty output should contain [INFO] prefix")
		}
		if !strings.Contains(str, "[ERROR]") {
			t.Error("pretty output should contain [ERROR] prefix")
		}
	})

	t.Run("filter by level", func(t *testing.T) {
		result, err := FilterAndFormatJSONLogs(jsonLogs, LogFilterParams{
			LogFormat:   "json",
			FilterLevel: "error",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		entries, ok := result.([]interface{})
		if !ok {
			t.Fatalf("expected []interface{}, got %T", result)
		}
		if len(entries) != 1 {
			t.Errorf("got %d entries, want 1 (only error)", len(entries))
		}
	})

	t.Run("filter by plan name", func(t *testing.T) {
		result, err := FilterAndFormatJSONLogs(jsonLogs, LogFilterParams{
			LogFormat:  "json",
			FilterPlan: "my-plan",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		entries, ok := result.([]interface{})
		if !ok {
			t.Fatalf("expected []interface{}, got %T", result)
		}
		if len(entries) != 2 {
			t.Errorf("got %d entries, want 2 (plan entries only)", len(entries))
		}
	})

	t.Run("malformed line preserved as raw", func(t *testing.T) {
		mixedLogs := `{"level":"info","ts":"2026-02-05","logger":"plan","msg":"OK"}
not a json line
{"level":"error","ts":"2026-02-05","logger":"plan","msg":"fail"}`

		result, err := FilterAndFormatJSONLogs(mixedLogs, LogFilterParams{LogFormat: "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		entries, ok := result.([]interface{})
		if !ok {
			t.Fatalf("expected []interface{}, got %T", result)
		}
		if len(entries) != 3 {
			t.Errorf("got %d entries, want 3 (2 JSON + 1 raw)", len(entries))
		}
		if raw, ok := entries[1].(RawLogLine); ok {
			if !strings.Contains(raw.Raw, "not a json line") {
				t.Error("raw line should preserve original text")
			}
		} else {
			t.Errorf("entry[1] should be RawLogLine, got %T", entries[1])
		}
	})

	t.Run("empty logs returns empty array", func(t *testing.T) {
		result, err := FilterAndFormatJSONLogs("", LogFilterParams{LogFormat: "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		entries, ok := result.([]interface{})
		if !ok {
			t.Fatalf("expected []interface{}, got %T", result)
		}
		if len(entries) != 0 {
			t.Errorf("got %d entries, want 0", len(entries))
		}
	})

	t.Run("default format is json", func(t *testing.T) {
		result, err := FilterAndFormatJSONLogs(jsonLogs, LogFilterParams{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := result.([]interface{}); !ok {
			t.Fatalf("default format should return []interface{}, got %T", result)
		}
	})
}

// --- FormatPrettyLogs tests ---

func TestFormatPrettyLogs(t *testing.T) {
	logLines := []interface{}{
		JSONLogEntry{
			Level:  "info",
			Ts:     "2026-02-05 10:00:00",
			Logger: "plan|abc",
			Msg:    "Reconcile started",
			Plan:   map[string]string{"name": "my-plan", "namespace": "demo"},
		},
		JSONLogEntry{
			Level:    "error",
			Ts:       "2026-02-05 10:01:00",
			Logger:   "provider|xyz",
			Msg:      "Connection failed",
			Provider: map[string]string{"name": "my-vsphere", "namespace": "demo"},
		},
		JSONLogEntry{
			Level:  "info",
			Ts:     "2026-02-05 10:02:00",
			Logger: "plan|abc",
			Msg:    "VM migrating",
			VMName: "web-server-01",
		},
		RawLogLine{Raw: "unparseable line"},
	}

	result := FormatPrettyLogs(logLines)

	if !strings.Contains(result, "[INFO]") {
		t.Error("should contain [INFO] prefix")
	}
	if !strings.Contains(result, "[ERROR]") {
		t.Error("should contain [ERROR] prefix")
	}
	if !strings.Contains(result, "plan=my-plan") {
		t.Error("should contain plan context")
	}
	if !strings.Contains(result, "provider=my-vsphere") {
		t.Error("should contain provider context")
	}
	if !strings.Contains(result, "vm=web-server-01") {
		t.Error("should contain VM context")
	}
	if !strings.Contains(result, "unparseable line") {
		t.Error("should preserve raw lines")
	}
}

// --- ProcessLogs tests ---

func TestProcessLogs(t *testing.T) {
	jsonLogs := `{"level":"info","ts":"2026-02-05 10:00:00","logger":"plan|abc","msg":"Started","plan":{"name":"my-plan","namespace":"demo"}}
{"level":"error","ts":"2026-02-05 10:01:00","logger":"plan|abc","msg":"Failed","plan":{"name":"my-plan","namespace":"demo"}}`

	t.Run("JSON logs with pretty default", func(t *testing.T) {
		result, format, err := ProcessLogs(jsonLogs, LogFilterParams{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if format != "pretty" {
			t.Errorf("format = %q, want %q", format, "pretty")
		}
		str, ok := result.(string)
		if !ok {
			t.Fatalf("expected string for pretty format, got %T", result)
		}
		if !strings.Contains(str, "[INFO]") {
			t.Error("pretty output should contain [INFO]")
		}
	})

	t.Run("JSON logs with json format", func(t *testing.T) {
		result, format, err := ProcessLogs(jsonLogs, LogFilterParams{LogFormat: "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if format != "json" {
			t.Errorf("format = %q, want %q", format, "json")
		}
		entries, ok := result.([]interface{})
		if !ok {
			t.Fatalf("expected []interface{}, got %T", result)
		}
		if len(entries) != 2 {
			t.Errorf("got %d entries, want 2", len(entries))
		}
	})

	t.Run("non-JSON logs returned as text", func(t *testing.T) {
		plainLogs := "Starting conversion\nDisk copied\nDone"
		result, format, err := ProcessLogs(plainLogs, LogFilterParams{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if format != "text" {
			t.Errorf("format = %q, want %q", format, "text")
		}
		str, ok := result.(string)
		if !ok {
			t.Fatalf("expected string, got %T", result)
		}
		if str != plainLogs {
			t.Error("non-JSON logs should be returned unchanged")
		}
	})

	t.Run("grep filter applied before JSON parsing", func(t *testing.T) {
		result, _, err := ProcessLogs(jsonLogs, LogFilterParams{
			Grep:      "Failed",
			LogFormat: "json",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		entries, ok := result.([]interface{})
		if !ok {
			t.Fatalf("expected []interface{}, got %T", result)
		}
		if len(entries) != 1 {
			t.Errorf("got %d entries, want 1", len(entries))
		}
	})

	t.Run("invalid format returns error", func(t *testing.T) {
		_, _, err := ProcessLogs(jsonLogs, LogFilterParams{LogFormat: "invalid"})
		if err == nil {
			t.Fatal("expected error for invalid format")
		}
	})
}

// --- HasFilters tests ---

func TestLogFilterParams_HasFilters(t *testing.T) {
	tests := []struct {
		name   string
		params LogFilterParams
		expect bool
	}{
		{
			name:   "no filters",
			params: LogFilterParams{},
			expect: false,
		},
		{
			name:   "filter plan set",
			params: LogFilterParams{FilterPlan: "my-plan"},
			expect: true,
		},
		{
			name:   "filter level set",
			params: LogFilterParams{FilterLevel: "error"},
			expect: true,
		},
		{
			name:   "grep only is not a JSON filter",
			params: LogFilterParams{Grep: "pattern"},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.HasFilters()
			if result != tt.expect {
				t.Errorf("HasFilters() = %v, want %v", result, tt.expect)
			}
		})
	}
}
