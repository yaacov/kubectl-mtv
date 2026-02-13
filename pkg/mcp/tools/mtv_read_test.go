package tools

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/pkg/mcp/discovery"
	"github.com/yaacov/kubectl-mtv/pkg/mcp/util"
)

// testRegistry builds a minimal registry for testing.
func testRegistry() *discovery.Registry {
	return &discovery.Registry{
		ReadOnly: map[string]*discovery.Command{
			"get/plan": {
				Path: []string{"get", "plan"}, PathString: "get plan", Description: "Get migration plans",
			},
			"get/provider": {
				Path: []string{"get", "provider"}, PathString: "get provider", Description: "Get providers",
			},
			"get/inventory/vm": {
				Path: []string{"get", "inventory", "vm"}, PathString: "get inventory vm", Description: "Get VMs",
			},
			"describe/plan": {
				Path: []string{"describe", "plan"}, PathString: "describe plan", Description: "Describe plan",
			},
			"health": {Path: []string{"health"}, PathString: "health", Description: "Health check"},
		},
		ReadWrite: map[string]*discovery.Command{
			"create/provider": {
				Path: []string{"create", "provider"}, PathString: "create provider", Description: "Create provider",
			},
			"create/plan": {
				Path: []string{"create", "plan"}, PathString: "create plan", Description: "Create plan",
			},
			"delete/plan": {
				Path: []string{"delete", "plan"}, PathString: "delete plan", Description: "Delete plan",
			},
			"start/plan": {
				Path: []string{"start", "plan"}, PathString: "start plan", Description: "Start plan",
			},
		},
	}
}

// --- Tool definition tests ---

func TestGetMTVReadTool(t *testing.T) {
	registry := testRegistry()
	tool := GetMTVReadTool(registry)

	if tool.Name != "mtv_read" {
		t.Errorf("Name = %q, want %q", tool.Name, "mtv_read")
	}

	if tool.Description == "" {
		t.Error("Description should not be empty")
	}

	// Description should reference read-only commands
	for _, keyword := range []string{"get plan", "health"} {
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

// --- normalizeCommandPath tests ---

func TestNormalizeCommandPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple two words", "get plan", "get/plan"},
		{"three words", "get inventory vm", "get/inventory/vm"},
		{"single word", "health", "health"},
		{"extra whitespace", "  get   plan  ", "get/plan"},
		{"tabs and spaces", "\tget\tplan\t", "get/plan"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeCommandPath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeCommandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// --- buildArgs tests ---

func TestBuildArgs(t *testing.T) {
	// Save and restore the output format
	origFormat := util.GetOutputFormat()
	defer util.SetOutputFormat(origFormat)

	tests := []struct {
		name         string
		cmdPath      string
		flags        map[string]any
		outputFormat string // configured default output format
		wantContains []string
		wantMissing  []string
	}{
		{
			name:         "simple command with default json output",
			cmdPath:      "get/plan",
			outputFormat: "json",
			wantContains: []string{"get", "plan", "--output", "json"},
		},
		{
			name:         "with namespace in flags",
			cmdPath:      "get/plan",
			flags:        map[string]any{"namespace": "demo"},
			outputFormat: "json",
			wantContains: []string{"get", "plan", "--namespace", "demo"},
		},
		{
			name:         "with all_namespaces in flags",
			cmdPath:      "get/plan",
			flags:        map[string]any{"all_namespaces": true},
			outputFormat: "json",
			wantContains: []string{"get", "plan", "--all-namespaces"},
		},
		{
			name:         "with provider in flags",
			cmdPath:      "get/inventory/vm",
			flags:        map[string]any{"provider": "my-provider"},
			outputFormat: "json",
			wantContains: []string{"get", "inventory", "vm", "--provider", "my-provider"},
		},
		{
			name:         "with inventory_url in flags",
			cmdPath:      "get/inventory/vm",
			flags:        map[string]any{"provider": "my-provider", "inventory_url": "http://localhost:9090"},
			outputFormat: "json",
			wantContains: []string{"--inventory-url", "http://localhost:9090"},
		},
		{
			name:         "user output overrides default",
			cmdPath:      "get/plan",
			flags:        map[string]any{"output": "yaml"},
			outputFormat: "json",
			wantContains: []string{"--output", "yaml"},
			wantMissing:  []string{"--output json"},
		},
		{
			name:         "text output format omits --output flag",
			cmdPath:      "get/plan",
			outputFormat: "text",
			wantContains: []string{"get", "plan"},
			wantMissing:  []string{"--output"},
		},
		{
			name:         "custom flags are passed through",
			cmdPath:      "get/inventory/vm",
			flags:        map[string]any{"provider": "my-provider", "query": "where name ~= 'prod-.*'", "extended": true},
			outputFormat: "json",
			wantContains: []string{"--query", "--extended"},
		},
		{
			name:         "namespace extracted from flags not duplicated",
			cmdPath:      "get/plan",
			flags:        map[string]any{"namespace": "real-ns"},
			outputFormat: "json",
			wantContains: []string{"--namespace", "real-ns"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util.SetOutputFormat(tt.outputFormat)
			result := buildArgs(tt.cmdPath, tt.flags)
			joined := strings.Join(result, " ")

			for _, want := range tt.wantContains {
				if !strings.Contains(joined, want) {
					t.Errorf("buildArgs() = %v, should contain %q", result, want)
				}
			}
			for _, notWant := range tt.wantMissing {
				if strings.Contains(joined, notWant) {
					t.Errorf("buildArgs() = %v, should NOT contain %q", result, notWant)
				}
			}
		})
	}
}

// --- appendNormalizedFlags tests ---

func TestAppendNormalizedFlags(t *testing.T) {
	tests := []struct {
		name         string
		flags        map[string]any
		skipFlags    map[string]bool
		wantContains []string
		wantMissing  []string
	}{
		{
			name:         "bool true becomes presence flag",
			flags:        map[string]any{"extended": true},
			wantContains: []string{"--extended"},
		},
		{
			name:         "bool false becomes explicit false",
			flags:        map[string]any{"migrate-shared-disks": false},
			wantContains: []string{"--migrate-shared-disks=false"},
		},
		{
			name:         "string true treated as bool",
			flags:        map[string]any{"watch": "true"},
			wantContains: []string{"--watch"},
		},
		{
			name:         "string false treated as bool",
			flags:        map[string]any{"watch": "false"},
			wantContains: []string{"--watch=false"},
		},
		{
			name:         "string value",
			flags:        map[string]any{"query": "where name = 'test'"},
			wantContains: []string{"--query", "where name = 'test'"},
		},
		{
			name:        "empty string is skipped",
			flags:       map[string]any{"query": ""},
			wantMissing: []string{"--query"},
		},
		{
			name:         "float64 whole number no decimals",
			flags:        map[string]any{"tail-lines": float64(500)},
			wantContains: []string{"--tail-lines", "500"},
			wantMissing:  []string{"500."},
		},
		{
			name:         "float64 fractional",
			flags:        map[string]any{"ratio": float64(0.75)},
			wantContains: []string{"--ratio", "0.75"},
		},
		{
			name:         "single char key uses single dash",
			flags:        map[string]any{"n": "demo"},
			wantContains: []string{"-n", "demo"},
			wantMissing:  []string{"--n"},
		},
		{
			name:        "nil value is skipped",
			flags:       map[string]any{"something": nil},
			wantMissing: []string{"--something"},
		},
		{
			name:         "skip set respected",
			flags:        map[string]any{"namespace": "should-skip", "query": "keep-me"},
			skipFlags:    map[string]bool{"namespace": true},
			wantContains: []string{"--query", "keep-me"},
			wantMissing:  []string{"--namespace"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendNormalizedFlags(nil, tt.flags, tt.skipFlags)
			joined := strings.Join(result, " ")

			for _, want := range tt.wantContains {
				if !strings.Contains(joined, want) {
					t.Errorf("appendNormalizedFlags() = %v, should contain %q", result, want)
				}
			}
			for _, notWant := range tt.wantMissing {
				if strings.Contains(joined, notWant) {
					t.Errorf("appendNormalizedFlags() = %v, should NOT contain %q", result, notWant)
				}
			}
		})
	}
}

// --- validateCommandInput tests ---

func TestValidateCommandInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError string
	}{
		{
			name:  "valid simple command",
			input: "get plan",
		},
		{
			name:  "valid inventory command",
			input: "get inventory vm",
		},
		{
			name:  "valid single word",
			input: "health",
		},
		{
			name:      "full CLI command with kubectl-mtv prefix",
			input:     "kubectl-mtv get plan --namespace demo",
			wantError: "subcommand path",
		},
		{
			name:      "full CLI command with kubectl prefix",
			input:     "kubectl get pods",
			wantError: "subcommand path",
		},
		{
			name:      "embedded tool output with return_value",
			input:     `get plan {"return_value": 0, "output": "..."}`,
			wantError: "previous tool response",
		},
		{
			name:      "embedded tool output with stdout",
			input:     `{"stdout": "some output"}`,
			wantError: "previous tool response",
		},
		{
			name:      "embedded TOOL_CALLS marker",
			input:     `get plan[TOOL_CALLS]more stuff`,
			wantError: "previous tool response",
		},
		{
			name:      "overly long command string",
			input:     strings.Repeat("a", 201),
			wantError: "too long",
		},
		{
			name:  "command at exactly 200 chars is ok",
			input: strings.Repeat("a", 200),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommandInput(tt.input)
			if tt.wantError == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %q, should contain %q", err.Error(), tt.wantError)
				}
			}
		})
	}
}

// --- Handler validation error tests ---

func TestHandleMTVRead_ValidationErrors(t *testing.T) {
	registry := testRegistry()
	handler := HandleMTVRead(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name      string
		input     MTVReadInput
		wantError string
	}{
		{
			name:      "unknown command",
			input:     MTVReadInput{Command: "nonexistent command"},
			wantError: "unknown command",
		},
		{
			name:      "write command rejected",
			input:     MTVReadInput{Command: "create plan"},
			wantError: "write operation",
		},
		{
			name:      "full CLI command rejected",
			input:     MTVReadInput{Command: "kubectl-mtv get plan --namespace demo"},
			wantError: "subcommand path",
		},
		{
			name:      "embedded tool output rejected",
			input:     MTVReadInput{Command: `get plan {"return_value": 0}`},
			wantError: "previous tool response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := handler(ctx, req, tt.input)
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

func TestHandleMTVRead_DryRun(t *testing.T) {
	registry := testRegistry()
	handler := HandleMTVRead(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Save and restore the output format
	origFormat := util.GetOutputFormat()
	defer util.SetOutputFormat(origFormat)
	util.SetOutputFormat("json")

	tests := []struct {
		name         string
		input        MTVReadInput
		wantContains []string
	}{
		{
			name: "get plan with namespace",
			input: MTVReadInput{
				Command: "get plan",
				Flags:   map[string]any{"namespace": "demo"},
				DryRun:  true,
			},
			wantContains: []string{"kubectl-mtv", "get", "plan", "--namespace", "demo", "--output", "json"},
		},
		{
			name: "get plan all namespaces",
			input: MTVReadInput{
				Command: "get plan",
				Flags:   map[string]any{"all_namespaces": true},
				DryRun:  true,
			},
			wantContains: []string{"kubectl-mtv", "get", "plan", "--all-namespaces"},
		},
		{
			name: "get inventory vm with provider via flag",
			input: MTVReadInput{
				Command: "get inventory vm",
				Flags:   map[string]any{"provider": "my-vsphere"},
				DryRun:  true,
			},
			wantContains: []string{"kubectl-mtv", "get", "inventory", "vm", "--provider", "my-vsphere"},
		},
		{
			name: "describe plan with name via flag",
			input: MTVReadInput{
				Command: "describe plan",
				Flags:   map[string]any{"name": "my-plan", "namespace": "test-ns"},
				DryRun:  true,
			},
			wantContains: []string{"kubectl-mtv", "describe", "plan", "--name", "my-plan", "--namespace", "test-ns"},
		},
		{
			name: "health check",
			input: MTVReadInput{
				Command: "health",
				DryRun:  true,
			},
			wantContains: []string{"kubectl-mtv", "health"},
		},
		{
			name: "with inventory URL",
			input: MTVReadInput{
				Command: "get inventory vm",
				Flags:   map[string]any{"provider": "my-provider", "inventory_url": "http://localhost:9090"},
				DryRun:  true,
			},
			wantContains: []string{"--inventory-url", "http://localhost:9090"},
		},
		{
			name: "with custom flags",
			input: MTVReadInput{
				Command: "get inventory vm",
				Flags:   map[string]any{"provider": "my-provider", "extended": true},
				DryRun:  true,
			},
			wantContains: []string{"--extended"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, data, err := handler(ctx, req, tt.input)
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

			// "command" field should NOT be present (stripped to prevent CLI mimicry)
			if _, exists := dataMap["command"]; exists {
				t.Error("response should NOT contain 'command' field (stripped to help small LLMs)")
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output = %q, should contain %q", output, want)
				}
			}
		})
	}
}

// buildTestBinary builds the kubectl-mtv binary from source into a temp directory
// and returns its path. The binary is cached for the duration of the test.
func buildTestBinary(t *testing.T) string {
	t.Helper()

	binary := t.TempDir() + "/kubectl-mtv"
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	cmd.Dir = findRepoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build kubectl-mtv from source: %v\n%s", err, out)
	}
	return binary
}

// findRepoRoot walks up from the test file to find the repo root (where go.mod is).
func findRepoRoot(t *testing.T) string {
	t.Helper()
	// We know the test is at pkg/mcp/tools/ so repo root is 4 levels up
	// Use go list instead for reliability
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to find repo root: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// newRegistryFromBinary runs the given kubectl-mtv binary with help --machine
// plus optional extra args and builds a Registry from the output.
func newRegistryFromBinary(t *testing.T, binary string, extraArgs ...string) *discovery.Registry {
	t.Helper()

	args := append([]string{"help", "--machine"}, extraArgs...)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, args...)
	output, err := cmd.Output()
	if err != nil {
		t.Skipf("Skipping: binary %s not available: %v", binary, err)
	}

	var schema discovery.HelpSchema
	if err := json.Unmarshal(output, &schema); err != nil {
		t.Fatalf("Failed to parse help schema: %v", err)
	}

	registry := &discovery.Registry{
		ReadOnly:        make(map[string]*discovery.Command),
		ReadWrite:       make(map[string]*discovery.Command),
		GlobalFlags:     schema.GlobalFlags,
		RootDescription: schema.Description,
	}

	for i := range schema.Commands {
		cmd := &schema.Commands[i]
		pathKey := cmd.PathKey()
		switch cmd.Category {
		case "read":
			registry.ReadOnly[pathKey] = cmd
		default:
			registry.ReadWrite[pathKey] = cmd
		}
	}

	return registry
}

// toolSizes holds the size metrics for a single tool description.
type toolSizes struct {
	name  string
	chars int
	lines int
}

// TestLiveToolContextSize builds the kubectl-mtv binary from source, discovers
// commands, and compares the OLD (full) vs NEW (minimal + mtv_help) tool
// descriptions that would be sent to an AI agent.
//
// Run with: go test -v -run TestLiveToolContextSize ./pkg/mcp/tools/
func TestLiveToolContextSize(t *testing.T) {
	// Build kubectl-mtv from source to ensure latest code is tested
	binary := buildTestBinary(t)
	registry := newRegistryFromBinary(t, binary)

	// --- OLD mode: full descriptions (2 tools) ---
	oldRead := registry.GenerateReadOnlyDescription()
	oldWrite := registry.GenerateReadWriteDescription()

	oldTools := []toolSizes{
		{name: "mtv_read", chars: len(oldRead), lines: strings.Count(oldRead, "\n") + 1},
		{name: "mtv_write", chars: len(oldWrite), lines: strings.Count(oldWrite, "\n") + 1},
	}

	// --- NEW mode: minimal descriptions (5 tools) ---
	newRead := registry.GenerateMinimalReadOnlyDescription()
	newWrite := registry.GenerateMinimalReadWriteDescription()
	newLogs := GetMinimalKubectlLogsTool().Description
	newKubectl := GetMinimalKubectlTool().Description
	newHelp := GetMTVHelpTool().Description

	newTools := []toolSizes{
		{name: "mtv_read", chars: len(newRead), lines: strings.Count(newRead, "\n") + 1},
		{name: "mtv_write", chars: len(newWrite), lines: strings.Count(newWrite, "\n") + 1},
		{name: "kubectl_logs", chars: len(newLogs), lines: strings.Count(newLogs, "\n") + 1},
		{name: "kubectl", chars: len(newKubectl), lines: strings.Count(newKubectl, "\n") + 1},
		{name: "mtv_help", chars: len(newHelp), lines: strings.Count(newHelp, "\n") + 1},
	}

	// --- Report ---
	t.Logf("=== MCP Tool Context Size Report (live kubectl-mtv) ===")
	t.Logf("")

	oldTotal := 0
	t.Logf("--- OLD mode (full descriptions, 2 tools) ---")
	for _, ts := range oldTools {
		t.Logf("  %-20s %6d chars  %4d lines  ~%d tokens", ts.name, ts.chars, ts.lines, ts.chars/4)
		oldTotal += ts.chars
	}
	t.Logf("  %-20s %6d chars            ~%d tokens", "TOTAL", oldTotal, oldTotal/4)

	t.Logf("")

	newTotal := 0
	t.Logf("--- NEW mode (minimal + mtv_help, 5 tools) ---")
	for _, ts := range newTools {
		t.Logf("  %-20s %6d chars  %4d lines  ~%d tokens", ts.name, ts.chars, ts.lines, ts.chars/4)
		newTotal += ts.chars
	}
	t.Logf("  %-20s %6d chars            ~%d tokens", "TOTAL", newTotal, newTotal/4)

	t.Logf("")

	saved := oldTotal - newTotal
	pct := float64(saved) / float64(oldTotal) * 100
	t.Logf("--- SAVINGS ---")
	t.Logf("  Saved: %d chars / ~%d tokens (%.1f%% reduction)", saved, saved/4, pct)
	t.Logf("  Read commands:  %d", len(registry.ReadOnly))
	t.Logf("  Write commands: %d", len(registry.ReadWrite))
	t.Logf("====================================================")

	// Dump new minimal descriptions for inspection
	t.Logf("")
	t.Logf("--- [NEW] mtv_read description ---")
	t.Logf("\n%s", newRead)
	t.Logf("--- [NEW] mtv_write description ---")
	t.Logf("\n%s", newWrite)
	t.Logf("--- [NEW] kubectl_logs description ---")
	t.Logf("\n%s", newLogs)
	t.Logf("--- [NEW] kubectl description ---")
	t.Logf("\n%s", newKubectl)
	t.Logf("--- [NEW] mtv_help description ---")
	t.Logf("\n%s", newHelp)
}

// --- filterResponseFields tests ---

func TestFilterResponseFields_ArrayData(t *testing.T) {
	data := map[string]interface{}{
		"return_value": float64(0),
		"data": []interface{}{
			map[string]interface{}{
				"name":       "vm-1",
				"id":         "vm-101",
				"powerState": "poweredOn",
				"concerns":   []interface{}{"cbt-disabled"},
				"memoryMB":   2048,
				"disks":      []interface{}{"disk-1", "disk-2"},
			},
			map[string]interface{}{
				"name":       "vm-2",
				"id":         "vm-102",
				"powerState": "poweredOff",
				"concerns":   []interface{}{},
				"memoryMB":   4096,
				"disks":      []interface{}{"disk-3"},
			},
		},
	}

	result := filterResponseFields(data, []string{"name", "id", "concerns"})

	// Envelope fields should be preserved
	if result["return_value"] != float64(0) {
		t.Error("return_value envelope field should be preserved")
	}

	// Data should be filtered
	items, ok := result["data"].([]interface{})
	if !ok {
		t.Fatal("data should be []interface{}")
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// First item: only name, id, concerns
	item0 := items[0].(map[string]interface{})
	if item0["name"] != "vm-1" {
		t.Errorf("item[0].name = %v, want vm-1", item0["name"])
	}
	if item0["id"] != "vm-101" {
		t.Errorf("item[0].id = %v, want vm-101", item0["id"])
	}
	if _, exists := item0["concerns"]; !exists {
		t.Error("item[0].concerns should exist")
	}
	// Excluded fields should be absent
	if _, exists := item0["powerState"]; exists {
		t.Error("item[0].powerState should be filtered out")
	}
	if _, exists := item0["memoryMB"]; exists {
		t.Error("item[0].memoryMB should be filtered out")
	}
	if _, exists := item0["disks"]; exists {
		t.Error("item[0].disks should be filtered out")
	}
}

func TestFilterResponseFields_ObjectData(t *testing.T) {
	data := map[string]interface{}{
		"return_value": float64(0),
		"data": map[string]interface{}{
			"overallStatus": "Healthy",
			"pods":          []interface{}{"pod-1", "pod-2"},
			"providers":     []interface{}{"prov-1"},
			"operator":      map[string]interface{}{"version": "2.10.4"},
		},
	}

	result := filterResponseFields(data, []string{"overallStatus", "providers"})

	obj, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("data should be map[string]interface{}")
	}

	if obj["overallStatus"] != "Healthy" {
		t.Errorf("overallStatus = %v, want Healthy", obj["overallStatus"])
	}
	if _, exists := obj["providers"]; !exists {
		t.Error("providers should exist")
	}
	if _, exists := obj["pods"]; exists {
		t.Error("pods should be filtered out")
	}
	if _, exists := obj["operator"]; exists {
		t.Error("operator should be filtered out")
	}
}

func TestFilterResponseFields_EmptyFields(t *testing.T) {
	data := map[string]interface{}{
		"return_value": float64(0),
		"data": []interface{}{
			map[string]interface{}{"name": "vm-1", "id": "vm-101"},
		},
	}

	result := filterResponseFields(data, []string{})

	// With empty fields, data should be unchanged
	items := result["data"].([]interface{})
	item0 := items[0].(map[string]interface{})
	if item0["name"] != "vm-1" {
		t.Error("data should not be modified when fields is empty")
	}
	if item0["id"] != "vm-101" {
		t.Error("data should not be modified when fields is empty")
	}
}

func TestFilterResponseFields_NoDataKey(t *testing.T) {
	data := map[string]interface{}{
		"return_value": float64(0),
		"output":       "some plain text output",
	}

	result := filterResponseFields(data, []string{"name", "id"})

	// When there's no "data" key, the response should be unchanged
	if result["output"] != "some plain text output" {
		t.Error("output should be preserved when no data key exists")
	}
}

func TestFilterResponseFields_NonObjectItems(t *testing.T) {
	data := map[string]interface{}{
		"return_value": float64(0),
		"data": []interface{}{
			"string-item",
			float64(42),
			map[string]interface{}{"name": "vm-1", "id": "vm-101", "extra": "value"},
		},
	}

	result := filterResponseFields(data, []string{"name"})

	items := result["data"].([]interface{})
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	// Non-object items kept as-is
	if items[0] != "string-item" {
		t.Errorf("string item should be preserved, got %v", items[0])
	}
	if items[1] != float64(42) {
		t.Errorf("number item should be preserved, got %v", items[1])
	}
	// Object item filtered
	obj := items[2].(map[string]interface{})
	if obj["name"] != "vm-1" {
		t.Error("name should be kept")
	}
	if _, exists := obj["id"]; exists {
		t.Error("id should be filtered out")
	}
	if _, exists := obj["extra"]; exists {
		t.Error("extra should be filtered out")
	}
}

// --- buildCLIErrorResult tests ---

func TestBuildCLIErrorResult_Success(t *testing.T) {
	// return_value == 0 should return nil (no error)
	data := map[string]interface{}{
		"return_value": float64(0),
		"data":         []interface{}{},
	}

	result := buildCLIErrorResult(data)
	if result != nil {
		t.Error("return_value 0 should not produce an error result")
	}
}

func TestBuildCLIErrorResult_MissingReturnValue(t *testing.T) {
	// No return_value key should return nil (no error)
	data := map[string]interface{}{}

	result := buildCLIErrorResult(data)
	if result != nil {
		t.Error("missing return_value should not produce an error result")
	}
}

func TestBuildCLIErrorResult_NonZeroExit(t *testing.T) {
	data := map[string]interface{}{
		"return_value": float64(1),
		"stderr":       "Error: failed to get provider 'vsphere-provider': providers.forklift.konveyor.io \"vsphere-provider\" not found\n",
		"stdout":       "",
	}

	result := buildCLIErrorResult(data)
	if result == nil {
		t.Fatal("non-zero return_value should produce an error result")
	}
	if !result.IsError {
		t.Error("result.IsError should be true")
	}
	if len(result.Content) == 0 {
		t.Fatal("result should have content")
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content should be TextContent, got %T", result.Content[0])
	}

	// Should contain exit code and stderr message
	if !strings.Contains(text.Text, "exit 1") {
		t.Errorf("error text should contain exit code, got: %s", text.Text)
	}
	if !strings.Contains(text.Text, "vsphere-provider") {
		t.Errorf("error text should contain stderr message, got: %s", text.Text)
	}
}

func TestBuildCLIErrorResult_NoStderr(t *testing.T) {
	data := map[string]interface{}{
		"return_value": float64(2),
		"stderr":       "",
	}

	result := buildCLIErrorResult(data)
	if result == nil {
		t.Fatal("non-zero return_value should produce an error result")
	}
	if !result.IsError {
		t.Error("result.IsError should be true")
	}

	text := result.Content[0].(*mcp.TextContent)
	if !strings.Contains(text.Text, "exit 2") {
		t.Errorf("error text should contain exit code, got: %s", text.Text)
	}
}

func TestBuildCLIErrorResult_StderrOnly(t *testing.T) {
	data := map[string]interface{}{
		"return_value": float64(1),
		"stderr":       "some error",
	}

	result := buildCLIErrorResult(data)
	if result == nil {
		t.Fatal("non-zero return_value should produce an error result")
	}

	text := result.Content[0].(*mcp.TextContent)
	if !strings.Contains(text.Text, "some error") {
		t.Errorf("error text should contain stderr, got: %s", text.Text)
	}
	if !strings.Contains(text.Text, "exit 1") {
		t.Errorf("error text should contain exit code, got: %s", text.Text)
	}
}

func TestFilterMapFields(t *testing.T) {
	m := map[string]interface{}{
		"name":       "test-vm",
		"id":         "vm-123",
		"powerState": "poweredOn",
		"memoryMB":   2048,
		"concerns":   []interface{}{"issue1"},
	}

	allowed := map[string]bool{"name": true, "concerns": true}
	result := filterMapFields(m, allowed)

	if len(result) != 2 {
		t.Errorf("expected 2 fields, got %d", len(result))
	}
	if result["name"] != "test-vm" {
		t.Errorf("name = %v, want test-vm", result["name"])
	}
	if _, exists := result["concerns"]; !exists {
		t.Error("concerns should exist")
	}
	if _, exists := result["powerState"]; exists {
		t.Error("powerState should be filtered out")
	}
}
