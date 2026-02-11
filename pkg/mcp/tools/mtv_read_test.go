package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/pkg/mcp/discovery"
	"github.com/yaacov/kubectl-mtv/pkg/mcp/util"
)

// testRegistry builds a minimal registry for testing.
func testRegistry() *discovery.Registry {
	return &discovery.Registry{
		ReadOnly: map[string]*discovery.Command{
			"get/plan":         {Path: []string{"get", "plan"}, PathString: "get plan", Description: "Get migration plans"},
			"get/provider":     {Path: []string{"get", "provider"}, PathString: "get provider", Description: "Get providers"},
			"get/inventory/vm": {Path: []string{"get", "inventory", "vm"}, PathString: "get inventory vm", Description: "Get VMs"},
			"describe/plan":    {Path: []string{"describe", "plan"}, PathString: "describe plan", Description: "Describe plan"},
			"health":           {Path: []string{"health"}, PathString: "health", Description: "Health check"},
		},
		ReadWrite: map[string]*discovery.Command{
			"create/provider": {Path: []string{"create", "provider"}, PathString: "create provider", Description: "Create provider"},
			"create/plan":     {Path: []string{"create", "plan"}, PathString: "create plan", Description: "Create plan"},
			"delete/plan":     {Path: []string{"delete", "plan"}, PathString: "delete plan", Description: "Delete plan"},
			"start/plan":      {Path: []string{"start", "plan"}, PathString: "start plan", Description: "Start plan"},
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

// --- buildArgs tests ---

func TestBuildArgs(t *testing.T) {
	// Save and restore the output format
	origFormat := util.GetOutputFormat()
	defer util.SetOutputFormat(origFormat)

	tests := []struct {
		name          string
		cmdPath       string
		args          []string
		flags         map[string]any
		namespace     string
		allNamespaces bool
		inventoryURL  string
		outputFormat  string // configured default output format
		wantContains  []string
		wantMissing   []string
	}{
		{
			name:         "simple command with default json output",
			cmdPath:      "get/plan",
			outputFormat: "json",
			wantContains: []string{"get", "plan", "-o", "json"},
		},
		{
			name:         "with namespace",
			cmdPath:      "get/plan",
			namespace:    "demo",
			outputFormat: "json",
			wantContains: []string{"get", "plan", "-n", "demo"},
		},
		{
			name:          "with all namespaces",
			cmdPath:       "get/plan",
			allNamespaces: true,
			outputFormat:  "json",
			wantContains:  []string{"get", "plan", "-A"},
		},
		{
			name:         "with positional args",
			cmdPath:      "get/inventory/vm",
			args:         []string{"my-provider"},
			outputFormat: "json",
			wantContains: []string{"get", "inventory", "vm", "my-provider"},
		},
		{
			name:         "with inventory URL",
			cmdPath:      "get/inventory/vm",
			args:         []string{"my-provider"},
			inventoryURL: "http://localhost:9090",
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
			name:         "namespace in flags is skipped (handled by dedicated field)",
			cmdPath:      "get/plan",
			namespace:    "real-ns",
			flags:        map[string]any{"namespace": "ignored-ns"},
			outputFormat: "json",
			wantContains: []string{"-n", "real-ns"},
			wantMissing:  []string{"ignored-ns"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util.SetOutputFormat(tt.outputFormat)
			result := buildArgs(tt.cmdPath, tt.args, tt.flags, tt.namespace, tt.allNamespaces, tt.inventoryURL)
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
				Command:   "get plan",
				Namespace: "demo",
				DryRun:    true,
			},
			wantContains: []string{"kubectl-mtv", "get", "plan", "-n", "demo", "-o", "json"},
		},
		{
			name: "get plan all namespaces",
			input: MTVReadInput{
				Command:       "get plan",
				AllNamespaces: true,
				DryRun:        true,
			},
			wantContains: []string{"kubectl-mtv", "get", "plan", "-A"},
		},
		{
			name: "get inventory vm with provider",
			input: MTVReadInput{
				Command: "get inventory vm",
				Args:    []string{"my-vsphere"},
				DryRun:  true,
			},
			wantContains: []string{"kubectl-mtv", "get", "inventory", "vm", "my-vsphere"},
		},
		{
			name: "describe plan with name",
			input: MTVReadInput{
				Command:   "describe plan",
				Args:      []string{"my-plan"},
				Namespace: "test-ns",
				DryRun:    true,
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
				Command:      "get inventory vm",
				Args:         []string{"my-provider"},
				InventoryURL: "http://localhost:9090",
				DryRun:       true,
			},
			wantContains: []string{"--inventory-url", "http://localhost:9090"},
		},
		{
			name: "with custom flags",
			input: MTVReadInput{
				Command: "get inventory vm",
				Args:    []string{"my-provider"},
				Flags:   map[string]any{"extended": true},
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
