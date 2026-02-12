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
				PositionalArgs: []discovery.Arg{{Name: "NAME", Required: false}},
			},
			"get/provider": {
				Path: []string{"get", "provider"}, PathString: "get provider", Description: "Get providers",
				PositionalArgs: []discovery.Arg{{Name: "NAME", Required: false}},
			},
			"get/inventory/vm": {
				Path: []string{"get", "inventory", "vm"}, PathString: "get inventory vm", Description: "Get VMs",
				PositionalArgs: []discovery.Arg{{Name: "PROVIDER", Required: true}},
			},
			"describe/plan": {
				Path: []string{"describe", "plan"}, PathString: "describe plan", Description: "Describe plan",
				PositionalArgs: []discovery.Arg{{Name: "NAME", Required: true}},
			},
			"health": {Path: []string{"health"}, PathString: "health", Description: "Health check"},
		},
		ReadWrite: map[string]*discovery.Command{
			"create/provider": {
				Path: []string{"create", "provider"}, PathString: "create provider", Description: "Create provider",
				PositionalArgs: []discovery.Arg{{Name: "NAME", Required: true}},
			},
			"create/plan": {
				Path: []string{"create", "plan"}, PathString: "create plan", Description: "Create plan",
				PositionalArgs: []discovery.Arg{{Name: "NAME", Required: true}},
			},
			"delete/plan": {
				Path: []string{"delete", "plan"}, PathString: "delete plan", Description: "Delete plan",
				PositionalArgs: []discovery.Arg{{Name: "NAME", Required: false, Variadic: true}},
			},
			"start/plan": {
				Path: []string{"start", "plan"}, PathString: "start plan", Description: "Start plan",
				PositionalArgs: []discovery.Arg{{Name: "NAME", Required: false, Variadic: true}},
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
	for _, key := range []string{"command", "return_value", "data", "output", "stderr"} {
		if _, exists := props[key]; !exists {
			t.Errorf("OutputSchema.properties should contain %q", key)
		}
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

// --- extractPositionalArgs tests ---

func TestExtractPositionalArgs_FromFlags(t *testing.T) {
	cmd := &discovery.Command{
		PositionalArgs: []discovery.Arg{
			{Name: "PROVIDER", Required: true},
		},
	}

	flags := map[string]any{"provider": "my-vsphere", "output": "json"}
	result := extractPositionalArgs(cmd, flags)

	if len(result) != 1 || result[0] != "my-vsphere" {
		t.Errorf("Expected [my-vsphere], got %v", result)
	}
	// "provider" should be removed from flags to avoid double-passing
	if _, ok := flags["provider"]; ok {
		t.Error("Expected 'provider' to be removed from flags after extraction")
	}
	// "output" should remain
	if _, ok := flags["output"]; !ok {
		t.Error("Expected 'output' to remain in flags")
	}
}

func TestExtractPositionalArgs_NilCmd(t *testing.T) {
	result := extractPositionalArgs(nil, map[string]any{"name": "test"})
	if result != nil {
		t.Errorf("Expected nil for nil command, got %v", result)
	}
}

func TestExtractPositionalArgs_NoPositionalArgs(t *testing.T) {
	cmd := &discovery.Command{}
	result := extractPositionalArgs(cmd, map[string]any{"name": "test"})
	if result != nil {
		t.Errorf("Expected nil for command with no positional args, got %v", result)
	}
}

func TestExtractPositionalArgs_MultipleArgs(t *testing.T) {
	cmd := &discovery.Command{
		PositionalArgs: []discovery.Arg{
			{Name: "PLAN_NAME", Required: true},
			{Name: "VM_NAME", Required: true},
		},
	}

	flags := map[string]any{"plan_name": "my-plan", "vm_name": "my-vm", "output": "json"}
	result := extractPositionalArgs(cmd, flags)

	if len(result) != 2 {
		t.Fatalf("Expected 2 args, got %v", result)
	}
	if result[0] != "my-plan" || result[1] != "my-vm" {
		t.Errorf("Expected [my-plan, my-vm], got %v", result)
	}
	if _, ok := flags["plan_name"]; ok {
		t.Error("plan_name should be removed from flags")
	}
	if _, ok := flags["vm_name"]; ok {
		t.Error("vm_name should be removed from flags")
	}
}

func TestExtractPositionalArgs_HyphenVariant(t *testing.T) {
	cmd := &discovery.Command{
		PositionalArgs: []discovery.Arg{
			{Name: "PLAN_NAME", Required: true},
		},
	}

	// LLM might use "plan-name" instead of "plan_name"
	flags := map[string]any{"plan-name": "my-plan"}
	result := extractPositionalArgs(cmd, flags)

	if len(result) != 1 || result[0] != "my-plan" {
		t.Errorf("Expected [my-plan] from hyphen variant, got %v", result)
	}
}

func TestExtractPositionalArgs_OptionalArg(t *testing.T) {
	cmd := &discovery.Command{
		PositionalArgs: []discovery.Arg{
			{Name: "NAME", Required: false},
		},
	}

	// When the optional arg is provided, extract it
	flags := map[string]any{"name": "my-plan"}
	result := extractPositionalArgs(cmd, flags)

	if len(result) != 1 || result[0] != "my-plan" {
		t.Errorf("Expected [my-plan], got %v", result)
	}

	// When not provided, result should be empty
	flags2 := map[string]any{"output": "json"}
	result2 := extractPositionalArgs(cmd, flags2)
	if len(result2) != 0 {
		t.Errorf("Expected empty for missing optional arg, got %v", result2)
	}
}

func TestExtractPositionalArgs_VariadicArray(t *testing.T) {
	cmd := &discovery.Command{
		PositionalArgs: []discovery.Arg{
			{Name: "NAME", Required: false, Variadic: true},
		},
	}

	// JSON array should expand to multiple args
	flags := map[string]any{"name": []interface{}{"plan-a", "plan-b", "plan-c"}}
	result := extractPositionalArgs(cmd, flags)

	if len(result) != 3 {
		t.Fatalf("Expected 3 args, got %d: %v", len(result), result)
	}
	if result[0] != "plan-a" || result[1] != "plan-b" || result[2] != "plan-c" {
		t.Errorf("Expected [plan-a, plan-b, plan-c], got %v", result)
	}
	// Flag should be removed
	if _, ok := flags["name"]; ok {
		t.Error("Expected 'name' to be removed from flags after extraction")
	}
}

func TestExtractPositionalArgs_VariadicSpaceSeparated(t *testing.T) {
	cmd := &discovery.Command{
		PositionalArgs: []discovery.Arg{
			{Name: "NAME", Required: false, Variadic: true},
		},
	}

	// Space-separated string should be split into multiple args
	// (variadic args are K8s resource names, never contain spaces)
	flags := map[string]any{"name": "plan-a plan-b plan-c"}
	result := extractPositionalArgs(cmd, flags)

	if len(result) != 3 {
		t.Fatalf("Expected 3 args, got %d: %v", len(result), result)
	}
	if result[0] != "plan-a" || result[1] != "plan-b" || result[2] != "plan-c" {
		t.Errorf("Expected [plan-a, plan-b, plan-c], got %v", result)
	}
}

func TestExtractPositionalArgs_NameWithEllipsisSuffix(t *testing.T) {
	// The help JSON encodes variadic args as "NAME..." (with trailing dots)
	// and may not emit a separate "variadic" boolean field.
	// extractPositionalArgs must strip the "..." suffix to match the LLM's
	// flag key "name" and detect variadic from the suffix.
	cmd := &discovery.Command{
		PositionalArgs: []discovery.Arg{
			{Name: "NAME...", Required: false}, // Variadic: false (not emitted in JSON)
		},
	}

	// Single value — should be extracted as positional arg, not passed as --name
	flags := map[string]any{"name": "vcenter-migration", "namespace": "demo"}
	result := extractPositionalArgs(cmd, flags)

	if len(result) != 1 || result[0] != "vcenter-migration" {
		t.Errorf("Expected [vcenter-migration], got %v", result)
	}
	// "name" should be removed from flags; "namespace" should remain
	if _, ok := flags["name"]; ok {
		t.Error("Expected 'name' to be removed from flags after extraction")
	}
	if _, ok := flags["namespace"]; !ok {
		t.Error("Expected 'namespace' to remain in flags")
	}
}

func TestExtractPositionalArgs_EllipsisSuffixVariadicArray(t *testing.T) {
	// Variadic detection from "..." suffix should handle JSON arrays
	cmd := &discovery.Command{
		PositionalArgs: []discovery.Arg{
			{Name: "NAME...", Required: false}, // Variadic detected from suffix
		},
	}

	flags := map[string]any{"name": []interface{}{"plan-a", "plan-b"}}
	result := extractPositionalArgs(cmd, flags)

	if len(result) != 2 {
		t.Fatalf("Expected 2 args, got %d: %v", len(result), result)
	}
	if result[0] != "plan-a" || result[1] != "plan-b" {
		t.Errorf("Expected [plan-a, plan-b], got %v", result)
	}
}

func TestExtractPositionalArgs_EllipsisSuffixVariadicSpaceSeparated(t *testing.T) {
	// Variadic detection from "..." suffix should handle space-separated strings
	cmd := &discovery.Command{
		PositionalArgs: []discovery.Arg{
			{Name: "NAME...", Required: false}, // Variadic detected from suffix
		},
	}

	flags := map[string]any{"name": "plan-a plan-b plan-c"}
	result := extractPositionalArgs(cmd, flags)

	if len(result) != 3 {
		t.Fatalf("Expected 3 args, got %d: %v", len(result), result)
	}
	if result[0] != "plan-a" || result[1] != "plan-b" || result[2] != "plan-c" {
		t.Errorf("Expected [plan-a, plan-b, plan-c], got %v", result)
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
		args         []string
		flags        map[string]any
		outputFormat string // configured default output format
		wantContains []string
		wantMissing  []string
	}{
		{
			name:         "simple command with default json output",
			cmdPath:      "get/plan",
			outputFormat: "json",
			wantContains: []string{"get", "plan", "-o", "json"},
		},
		{
			name:         "with namespace in flags",
			cmdPath:      "get/plan",
			flags:        map[string]any{"namespace": "demo"},
			outputFormat: "json",
			wantContains: []string{"get", "plan", "-n", "demo"},
		},
		{
			name:         "with all_namespaces in flags",
			cmdPath:      "get/plan",
			flags:        map[string]any{"all_namespaces": true},
			outputFormat: "json",
			wantContains: []string{"get", "plan", "-A"},
		},
		{
			name:         "with positional args",
			cmdPath:      "get/inventory/vm",
			args:         []string{"my-provider"},
			outputFormat: "json",
			wantContains: []string{"get", "inventory", "vm", "my-provider"},
		},
		{
			name:         "with inventory_url in flags",
			cmdPath:      "get/inventory/vm",
			args:         []string{"my-provider"},
			flags:        map[string]any{"inventory_url": "http://localhost:9090"},
			outputFormat: "json",
			wantContains: []string{"--inventory-url", "http://localhost:9090"},
		},
		{
			name:         "user output overrides default",
			cmdPath:      "get/plan",
			flags:        map[string]any{"output": "yaml"},
			outputFormat: "json",
			wantContains: []string{"-o", "yaml"},
			wantMissing:  []string{"-o json"},
		},
		{
			name:         "text output format omits -o flag",
			cmdPath:      "get/plan",
			outputFormat: "text",
			wantContains: []string{"get", "plan"},
			wantMissing:  []string{"-o"},
		},
		{
			name:         "custom flags are passed through",
			cmdPath:      "get/inventory/vm",
			args:         []string{"my-provider"},
			flags:        map[string]any{"query": "where name ~= 'prod-.*'", "extended": true},
			outputFormat: "json",
			wantContains: []string{"--query", "--extended"},
		},
		{
			name:         "namespace extracted from flags not duplicated",
			cmdPath:      "get/plan",
			flags:        map[string]any{"namespace": "real-ns"},
			outputFormat: "json",
			wantContains: []string{"-n", "real-ns"},
			wantMissing:  []string{"--namespace"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util.SetOutputFormat(tt.outputFormat)
			result := buildArgs(tt.cmdPath, tt.args, tt.flags)
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
			wantContains: []string{"kubectl-mtv", "get", "plan", "-n", "demo", "-o", "json"},
		},
		{
			name: "get plan all namespaces",
			input: MTVReadInput{
				Command: "get plan",
				Flags:   map[string]any{"all_namespaces": true},
				DryRun:  true,
			},
			wantContains: []string{"kubectl-mtv", "get", "plan", "-A"},
		},
		{
			name: "get inventory vm with provider via named flag",
			input: MTVReadInput{
				Command: "get inventory vm",
				Flags:   map[string]any{"provider": "my-vsphere"},
				DryRun:  true,
			},
			wantContains: []string{"kubectl-mtv", "get", "inventory", "vm", "my-vsphere"},
		},
		{
			name: "describe plan with name via named flag",
			input: MTVReadInput{
				Command: "describe plan",
				Flags:   map[string]any{"name": "my-plan", "namespace": "test-ns"},
				DryRun:  true,
			},
			wantContains: []string{"kubectl-mtv", "describe", "plan", "my-plan", "-n", "test-ns"},
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

			command, ok := dataMap["command"].(string)
			if !ok {
				t.Fatal("response should have 'command' string field")
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(command, want) {
					t.Errorf("command = %q, should contain %q", command, want)
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
		"command":      "kubectl-mtv get inventory vm vsphere-provider -n demo -o json",
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
	if result["command"] != "kubectl-mtv get inventory vm vsphere-provider -n demo -o json" {
		t.Error("command envelope field should be preserved")
	}
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
		"command":      "kubectl-mtv health -o json",
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
		"command":      "test",
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
		"command":      "test",
		"return_value": float64(0),
		"output":       "some plain text output",
	}

	result := filterResponseFields(data, []string{"name", "id"})

	// When there's no "data" key, the response should be unchanged
	if result["output"] != "some plain text output" {
		t.Error("output should be preserved when no data key exists")
	}
	if result["command"] != "test" {
		t.Error("command should be preserved")
	}
}

func TestFilterResponseFields_NonObjectItems(t *testing.T) {
	data := map[string]interface{}{
		"command":      "test",
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
		"command":      "kubectl-mtv get plan -o json",
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
	data := map[string]interface{}{
		"command": "kubectl-mtv health",
	}

	result := buildCLIErrorResult(data)
	if result != nil {
		t.Error("missing return_value should not produce an error result")
	}
}

func TestBuildCLIErrorResult_NonZeroExit(t *testing.T) {
	data := map[string]interface{}{
		"command":      "kubectl-mtv get inventory vm vsphere-provider -o json",
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

	// Should contain the command, exit code, and stderr message
	if !strings.Contains(text.Text, "exit 1") {
		t.Errorf("error text should contain exit code, got: %s", text.Text)
	}
	if !strings.Contains(text.Text, "vsphere-provider") {
		t.Errorf("error text should contain stderr message, got: %s", text.Text)
	}
	if !strings.Contains(text.Text, "kubectl-mtv get inventory vm") {
		t.Errorf("error text should contain command, got: %s", text.Text)
	}
}

func TestBuildCLIErrorResult_NoStderr(t *testing.T) {
	data := map[string]interface{}{
		"command":      "kubectl-mtv get plan -o json",
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

func TestBuildCLIErrorResult_NoCommand(t *testing.T) {
	data := map[string]interface{}{
		"return_value": float64(1),
		"stderr":       "some error",
	}

	result := buildCLIErrorResult(data)
	if result == nil {
		t.Fatal("non-zero return_value should produce an error result")
	}

	text := result.Content[0].(*mcp.TextContent)
	// Should still work without command
	if !strings.Contains(text.Text, "some error") {
		t.Errorf("error text should contain stderr, got: %s", text.Text)
	}
	// Should not have the "[command]" prefix
	if strings.HasPrefix(text.Text, "[") {
		t.Errorf("error text should not have command prefix when command is empty, got: %s", text.Text)
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
