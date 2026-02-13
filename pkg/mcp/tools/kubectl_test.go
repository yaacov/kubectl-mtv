package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/pkg/mcp/util"
)

// --- Tool definition tests ---

func TestGetMinimalKubectlLogsTool(t *testing.T) {
	tool := GetMinimalKubectlLogsTool()

	if tool.Name != "kubectl_logs" {
		t.Errorf("Name = %q, want %q", tool.Name, "kubectl_logs")
	}

	if tool.Description == "" {
		t.Error("Description should not be empty")
	}

	// Should contain key log-specific keywords
	for _, keyword := range []string{
		"forklift-controller",
		"filter_plan",
		"log_format",
		"filter_level",
		"deployments/forklift-controller",
		"tail_lines",
	} {
		if !strings.Contains(tool.Description, keyword) {
			t.Errorf("Description should contain %q", keyword)
		}
	}

	// OutputSchema should have expected properties
	schema, ok := tool.OutputSchema.(map[string]any)
	if !ok {
		t.Fatal("OutputSchema should be a map")
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("OutputSchema should have properties")
	}
	for _, key := range []string{"return_value", "data", "output", "stderr"} {
		if _, exists := props[key]; !exists {
			t.Errorf("OutputSchema.properties should contain %q", key)
		}
	}
	// "command" should NOT be in the output schema (stripped to prevent CLI mimicry)
	if _, exists := props["command"]; exists {
		t.Error("OutputSchema.properties should NOT contain 'command' (stripped to help small LLMs)")
	}
}

func TestGetMinimalKubectlTool(t *testing.T) {
	tool := GetMinimalKubectlTool()

	if tool.Name != "kubectl" {
		t.Errorf("Name = %q, want %q", tool.Name, "kubectl")
	}

	if tool.Description == "" {
		t.Error("Description should not be empty")
	}

	// Should mention the three actions
	for _, action := range []string{"get", "describe", "events"} {
		if !strings.Contains(tool.Description, action) {
			t.Errorf("Description should contain action %q", action)
		}
	}

	// Should contain key resource-inspection keywords
	for _, keyword := range []string{
		"resource_type",
		"namespace",
		"mtv_read",
	} {
		if !strings.Contains(tool.Description, keyword) {
			t.Errorf("Description should contain %q", keyword)
		}
	}

	// OutputSchema should have expected properties
	schema, ok := tool.OutputSchema.(map[string]any)
	if !ok {
		t.Fatal("OutputSchema should be a map")
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("OutputSchema should have properties")
	}
	for _, key := range []string{"return_value", "data", "output", "stderr"} {
		if _, exists := props[key]; !exists {
			t.Errorf("OutputSchema.properties should contain %q", key)
		}
	}
	// "command" should NOT be in the output schema (stripped to prevent CLI mimicry)
	if _, exists := props["command"]; exists {
		t.Error("OutputSchema.properties should NOT contain 'command' (stripped to help small LLMs)")
	}
}

// --- buildLogsArgs tests ---

func TestBuildLogsArgs(t *testing.T) {
	tests := []struct {
		name         string
		params       kubectlDebugParams
		wantContains []string
		wantMissing  []string
	}{
		{
			name:         "basic logs with default tail",
			params:       kubectlDebugParams{Name: "deployments/forklift-controller", Namespace: "openshift-mtv"},
			wantContains: []string{"logs", "deployments/forklift-controller", "-n", "openshift-mtv", "--tail", "500", "--timestamps"},
		},
		{
			name:         "custom tail lines",
			params:       kubectlDebugParams{Name: "my-pod", TailLines: 100},
			wantContains: []string{"--tail", "100"},
			wantMissing:  []string{"--tail 500"},
		},
		{
			name:        "tail -1 gets all logs",
			params:      kubectlDebugParams{Name: "my-pod", TailLines: -1},
			wantMissing: []string{"--tail"},
		},
		{
			name:         "with previous",
			params:       kubectlDebugParams{Name: "crashed-pod", Previous: true},
			wantContains: []string{"--previous"},
		},
		{
			name:         "with container",
			params:       kubectlDebugParams{Name: "multi-container-pod", Container: "main"},
			wantContains: []string{"-c", "main"},
		},
		{
			name:         "with since",
			params:       kubectlDebugParams{Name: "my-pod", Since: "1h"},
			wantContains: []string{"--since", "1h"},
		},
		{
			name:        "no timestamps when disabled",
			params:      kubectlDebugParams{Name: "my-pod", NoTimestamps: true},
			wantMissing: []string{"--timestamps"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildLogsArgs(tt.params)
			joined := strings.Join(result, " ")

			for _, want := range tt.wantContains {
				if !strings.Contains(joined, want) {
					t.Errorf("buildLogsArgs() = %v, should contain %q", result, want)
				}
			}
			for _, notWant := range tt.wantMissing {
				if strings.Contains(joined, notWant) {
					t.Errorf("buildLogsArgs() = %v, should NOT contain %q", result, notWant)
				}
			}
		})
	}
}

// --- buildGetArgs tests ---

func TestBuildGetArgs(t *testing.T) {
	origFormat := util.GetOutputFormat()
	defer util.SetOutputFormat(origFormat)
	util.SetOutputFormat("json")

	tests := []struct {
		name         string
		params       kubectlDebugParams
		wantContains []string
		wantMissing  []string
	}{
		{
			name:         "basic get pods",
			params:       kubectlDebugParams{ResourceType: "pods", Namespace: "openshift-mtv"},
			wantContains: []string{"get", "pods", "-n", "openshift-mtv", "-o", "json"},
		},
		{
			name:         "get with labels",
			params:       kubectlDebugParams{ResourceType: "pods", Labels: "plan=my-plan"},
			wantContains: []string{"get", "pods", "-l", "plan=my-plan"},
		},
		{
			name:         "get with name",
			params:       kubectlDebugParams{ResourceType: "pvc", Name: "my-pvc"},
			wantContains: []string{"get", "pvc", "my-pvc"},
		},
		{
			name:         "get all namespaces",
			params:       kubectlDebugParams{ResourceType: "pods", AllNamespaces: true},
			wantContains: []string{"get", "pods", "-A"},
			wantMissing:  []string{"-n"},
		},
		{
			name:         "get with custom output",
			params:       kubectlDebugParams{ResourceType: "pods", Output: "wide"},
			wantContains: []string{"-o", "wide"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildGetArgs(tt.params)
			joined := strings.Join(result, " ")

			for _, want := range tt.wantContains {
				if !strings.Contains(joined, want) {
					t.Errorf("buildGetArgs() = %v, should contain %q", result, want)
				}
			}
			for _, notWant := range tt.wantMissing {
				if strings.Contains(joined, notWant) {
					t.Errorf("buildGetArgs() = %v, should NOT contain %q", result, notWant)
				}
			}
		})
	}
}

// --- buildDescribeArgs tests ---

func TestBuildDescribeArgs(t *testing.T) {
	tests := []struct {
		name         string
		params       kubectlDebugParams
		wantContains []string
	}{
		{
			name:         "describe pod",
			params:       kubectlDebugParams{ResourceType: "pods", Name: "my-pod", Namespace: "demo"},
			wantContains: []string{"describe", "pods", "my-pod", "-n", "demo"},
		},
		{
			name:         "describe with labels",
			params:       kubectlDebugParams{ResourceType: "pods", Labels: "app=test"},
			wantContains: []string{"describe", "pods", "-l", "app=test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDescribeArgs(tt.params)
			joined := strings.Join(result, " ")

			for _, want := range tt.wantContains {
				if !strings.Contains(joined, want) {
					t.Errorf("buildDescribeArgs() = %v, should contain %q", result, want)
				}
			}
		})
	}
}

// --- buildEventsArgs tests ---

func TestBuildEventsArgs(t *testing.T) {
	origFormat := util.GetOutputFormat()
	defer util.SetOutputFormat(origFormat)
	util.SetOutputFormat("json")

	tests := []struct {
		name         string
		params       kubectlDebugParams
		wantContains []string
	}{
		{
			name:         "events with namespace",
			params:       kubectlDebugParams{Namespace: "demo"},
			wantContains: []string{"get", "events", "-n", "demo"},
		},
		{
			name:         "events with for_resource",
			params:       kubectlDebugParams{ForResource: "pod/my-pod", Namespace: "demo"},
			wantContains: []string{"--for", "pod/my-pod"},
		},
		{
			name:         "events with field selector",
			params:       kubectlDebugParams{FieldSelector: "type=Warning", Namespace: "demo"},
			wantContains: []string{"--field-selector", "type=Warning"},
		},
		{
			name:         "events with sort by",
			params:       kubectlDebugParams{SortBy: ".lastTimestamp", Namespace: "demo"},
			wantContains: []string{"--sort-by", ".lastTimestamp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildEventsArgs(tt.params)
			joined := strings.Join(result, " ")

			for _, want := range tt.wantContains {
				if !strings.Contains(joined, want) {
					t.Errorf("buildEventsArgs() = %v, should contain %q", result, want)
				}
			}
		})
	}
}

// --- Handler validation error tests ---

func TestHandleKubectlLogs_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name      string
		input     KubectlLogsInput
		wantError string
	}{
		{
			name:      "logs without name",
			input:     KubectlLogsInput{},
			wantError: "'name' is required",
		},
		{
			name: "logs with name in flags works",
			input: KubectlLogsInput{
				Flags:  map[string]any{"name": "my-pod"},
				DryRun: true,
			},
			wantError: "", // no error expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := HandleKubectlLogs(ctx, req, tt.input)
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("error = %q, should contain %q", err.Error(), tt.wantError)
			}
		})
	}
}

func TestHandleKubectl_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name      string
		input     KubectlInput
		wantError string
	}{
		{
			name:      "unknown action",
			input:     KubectlInput{Action: "invalid"},
			wantError: "unknown action",
		},
		{
			name:      "get without resource_type",
			input:     KubectlInput{Action: "get"},
			wantError: "requires 'resource_type'",
		},
		{
			name:      "describe without resource_type",
			input:     KubectlInput{Action: "describe"},
			wantError: "requires 'resource_type'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := HandleKubectl(ctx, req, tt.input)
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("error = %q, should contain %q", err.Error(), tt.wantError)
			}
		})
	}
}

// --- Handler DryRun tests ---

func TestHandleKubectlLogs_DryRun(t *testing.T) {
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name         string
		input        KubectlLogsInput
		wantContains []string
	}{
		{
			name: "basic logs",
			input: KubectlLogsInput{
				Flags:  map[string]any{"name": "deployments/forklift-controller", "namespace": "openshift-mtv"},
				DryRun: true,
			},
			wantContains: []string{"kubectl", "logs", "deployments/forklift-controller", "-n", "openshift-mtv", "--tail", "500", "--timestamps"},
		},
		{
			name: "logs with previous and container",
			input: KubectlLogsInput{
				Flags:  map[string]any{"name": "crashed-pod", "container": "main", "previous": true, "tail_lines": 200},
				DryRun: true,
			},
			wantContains: []string{"kubectl", "logs", "crashed-pod", "-c", "main", "--previous", "--tail", "200"},
		},
		{
			name: "logs with since",
			input: KubectlLogsInput{
				Flags:  map[string]any{"name": "my-pod", "since": "30m"},
				DryRun: true,
			},
			wantContains: []string{"--since", "30m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, data, err := HandleKubectlLogs(ctx, req, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			dataMap, ok := data.(map[string]interface{})
			if !ok {
				t.Fatalf("expected map[string]interface{}, got %T", data)
			}

			// In dry-run mode, the CLI command is in "output" (command field is stripped)
			output, ok := dataMap["output"].(string)
			if !ok {
				t.Fatal("response should have 'output' string field in dry-run mode")
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output = %q, should contain %q", output, want)
				}
			}
		})
	}
}

func TestHandleKubectl_DryRun(t *testing.T) {
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	origFormat := util.GetOutputFormat()
	defer util.SetOutputFormat(origFormat)
	util.SetOutputFormat("json")

	tests := []struct {
		name         string
		input        KubectlInput
		wantContains []string
	}{
		{
			name: "get action",
			input: KubectlInput{
				Action: "get",
				Flags:  map[string]any{"resource_type": "pods", "labels": "plan=my-plan", "namespace": "demo"},
				DryRun: true,
			},
			wantContains: []string{"kubectl", "get", "pods", "-n", "demo", "-l", "plan=my-plan", "-o", "json"},
		},
		{
			name: "describe action",
			input: KubectlInput{
				Action: "describe",
				Flags:  map[string]any{"resource_type": "pods", "name": "virt-v2v-cold-123", "namespace": "demo"},
				DryRun: true,
			},
			wantContains: []string{"kubectl", "describe", "pods", "virt-v2v-cold-123", "-n", "demo"},
		},
		{
			name: "events action",
			input: KubectlInput{
				Action: "events",
				Flags:  map[string]any{"field_selector": "type=Warning", "namespace": "demo"},
				DryRun: true,
			},
			wantContains: []string{"kubectl", "get", "events", "-n", "demo", "--field-selector", "type=Warning"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, data, err := HandleKubectl(ctx, req, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			dataMap, ok := data.(map[string]interface{})
			if !ok {
				t.Fatalf("expected map[string]interface{}, got %T", data)
			}

			// In dry-run mode, the CLI command is in "output" (command field is stripped)
			output, ok := dataMap["output"].(string)
			if !ok {
				t.Fatal("response should have 'output' string field in dry-run mode")
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output = %q, should contain %q", output, want)
				}
			}
		})
	}
}

// --- filterLogsByPattern tests ---

func TestFilterLogsByPattern(t *testing.T) {
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
			result, err := filterLogsByPattern(logs, tt.pattern, tt.ignoreCase)
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
			// Count non-empty lines for the empty pattern case
			nonEmpty := 0
			for _, l := range lines {
				if l != "" {
					nonEmpty++
				}
			}

			if tt.pattern == "" {
				// Empty pattern returns original string unchanged
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

// --- looksLikeJSONLogs tests ---

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
			result := looksLikeJSONLogs(tt.logs)
			if result != tt.expect {
				t.Errorf("looksLikeJSONLogs() = %v, want %v", result, tt.expect)
			}
		})
	}
}

// --- matchesParamFilters tests ---

func TestMatchesParamFilters(t *testing.T) {
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
		params kubectlDebugParams
		expect bool
	}{
		{
			name:   "no filters matches all",
			params: kubectlDebugParams{},
			expect: true,
		},
		{
			name:   "filter level matches",
			params: kubectlDebugParams{FilterLevel: "info"},
			expect: true,
		},
		{
			name:   "filter level mismatch",
			params: kubectlDebugParams{FilterLevel: "error"},
			expect: false,
		},
		{
			name:   "filter logger matches prefix",
			params: kubectlDebugParams{FilterLogger: "plan"},
			expect: true,
		},
		{
			name:   "filter logger mismatch",
			params: kubectlDebugParams{FilterLogger: "provider"},
			expect: false,
		},
		{
			name:   "filter plan matches",
			params: kubectlDebugParams{FilterPlan: "my-plan"},
			expect: true,
		},
		{
			name:   "filter plan mismatch",
			params: kubectlDebugParams{FilterPlan: "other-plan"},
			expect: false,
		},
		{
			name:   "filter provider matches",
			params: kubectlDebugParams{FilterProvider: "my-provider"},
			expect: true,
		},
		{
			name:   "filter provider mismatch",
			params: kubectlDebugParams{FilterProvider: "other"},
			expect: false,
		},
		{
			name:   "filter VM by VM field",
			params: kubectlDebugParams{FilterVM: "vm-001"},
			expect: true,
		},
		{
			name:   "filter VM by VMName field",
			params: kubectlDebugParams{FilterVM: "web-server"},
			expect: true,
		},
		{
			name:   "filter VM by VMID field",
			params: kubectlDebugParams{FilterVM: "id-123"},
			expect: true,
		},
		{
			name:   "filter VM mismatch",
			params: kubectlDebugParams{FilterVM: "nonexistent"},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesParamFilters(entry, tt.params)
			if result != tt.expect {
				t.Errorf("matchesParamFilters() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestMatchesParamFilters_Migration(t *testing.T) {
	entry := JSONLogEntry{
		Level:     "info",
		Logger:    "migration|migration-xyz",
		Msg:       "Migration started",
		Migration: map[string]string{"name": "migration-xyz"},
	}

	tests := []struct {
		name   string
		params kubectlDebugParams
		expect bool
	}{
		{
			name:   "migration filter matches from field",
			params: kubectlDebugParams{FilterMigration: "migration-xyz"},
			expect: true,
		},
		{
			name:   "migration filter mismatch",
			params: kubectlDebugParams{FilterMigration: "other-migration"},
			expect: false,
		},
		{
			name:   "migration filter requires migration logger",
			params: kubectlDebugParams{FilterMigration: "migration-xyz"},
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesParamFilters(entry, tt.params)
			if result != tt.expect {
				t.Errorf("matchesParamFilters() = %v, want %v", result, tt.expect)
			}
		})
	}

	// Non-migration logger should not match migration filter
	nonMigrationEntry := JSONLogEntry{
		Level:  "info",
		Logger: "plan|abc",
		Msg:    "test",
	}
	if matchesParamFilters(nonMigrationEntry, kubectlDebugParams{FilterMigration: "any"}) {
		t.Error("non-migration logger should not match migration filter")
	}
}

// --- filterAndFormatJSONLogs tests ---

func TestFilterAndFormatJSONLogs(t *testing.T) {
	jsonLogs := `{"level":"info","ts":"2026-02-05 10:00:00","logger":"plan|abc","msg":"Started","plan":{"name":"my-plan","namespace":"demo"}}
{"level":"error","ts":"2026-02-05 10:01:00","logger":"plan|abc","msg":"Failed","plan":{"name":"my-plan","namespace":"demo"}}
{"level":"info","ts":"2026-02-05 10:02:00","logger":"provider|xyz","msg":"Refreshed","provider":{"name":"my-provider","namespace":"demo"}}`

	t.Run("json format returns array", func(t *testing.T) {
		result, err := filterAndFormatJSONLogs(jsonLogs, kubectlDebugParams{LogFormat: "json"})
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
		result, err := filterAndFormatJSONLogs(jsonLogs, kubectlDebugParams{LogFormat: "text"})
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
		result, err := filterAndFormatJSONLogs(jsonLogs, kubectlDebugParams{LogFormat: "pretty"})
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
		result, err := filterAndFormatJSONLogs(jsonLogs, kubectlDebugParams{
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
		result, err := filterAndFormatJSONLogs(jsonLogs, kubectlDebugParams{
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

		result, err := filterAndFormatJSONLogs(mixedLogs, kubectlDebugParams{LogFormat: "json"})
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
		result, err := filterAndFormatJSONLogs("", kubectlDebugParams{LogFormat: "json"})
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
		result, err := filterAndFormatJSONLogs(jsonLogs, kubectlDebugParams{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := result.([]interface{}); !ok {
			t.Fatalf("default format should return []interface{}, got %T", result)
		}
	})
}

// --- formatPrettyLogs tests ---

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

	result := formatPrettyLogs(logLines)

	// Check level prefixes
	if !strings.Contains(result, "[INFO]") {
		t.Error("should contain [INFO] prefix")
	}
	if !strings.Contains(result, "[ERROR]") {
		t.Error("should contain [ERROR] prefix")
	}

	// Check context annotations
	if !strings.Contains(result, "plan=my-plan") {
		t.Error("should contain plan context")
	}
	if !strings.Contains(result, "provider=my-vsphere") {
		t.Error("should contain provider context")
	}
	if !strings.Contains(result, "vm=web-server-01") {
		t.Error("should contain VM context")
	}

	// Check raw line preserved
	if !strings.Contains(result, "unparseable line") {
		t.Error("should preserve raw lines")
	}
}
