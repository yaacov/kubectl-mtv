package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/pkg/mcp/discovery"
)

// --- Tool definition tests ---

func TestGetMTVWriteTool(t *testing.T) {
	registry := testRegistry()
	tool := GetMTVWriteTool(registry)

	if tool.Name != "mtv_write" {
		t.Errorf("Name = %q, want %q", tool.Name, "mtv_write")
	}

	if tool.Description == "" {
		t.Error("Description should not be empty")
	}

	// Description should reference write commands
	for _, keyword := range []string{"create provider", "create plan"} {
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

// --- buildWriteArgs tests ---

func TestBuildWriteArgs(t *testing.T) {
	tests := []struct {
		name         string
		cmdPath      string
		flags        map[string]any
		wantContains []string
		wantMissing  []string
	}{
		{
			name:         "simple command",
			cmdPath:      "create/provider",
			wantContains: []string{"create", "provider"},
		},
		{
			name:         "with name in flags",
			cmdPath:      "delete/plan",
			flags:        map[string]any{"name": "my-plan"},
			wantContains: []string{"delete", "plan", "--name", "my-plan"},
		},
		{
			name:         "with namespace in flags",
			cmdPath:      "start/plan",
			flags:        map[string]any{"name": "my-plan", "namespace": "demo"},
			wantContains: []string{"start", "plan", "--name", "my-plan", "--namespace", "demo"},
		},
		{
			name:    "with flags",
			cmdPath: "create/provider",
			flags: map[string]any{
				"name":                       "my-provider",
				"type":                       "vsphere",
				"url":                        "https://vcenter.example.com",
				"provider-insecure-skip-tls": true,
			},
			wantContains: []string{"create", "provider", "--name", "my-provider", "--type", "vsphere", "--url", "--provider-insecure-skip-tls"},
		},
		{
			name:         "does not auto-add output format",
			cmdPath:      "create/plan",
			flags:        map[string]any{"name": "test-plan"},
			wantContains: []string{"create", "plan", "--name", "test-plan"},
			wantMissing:  []string{"--output"},
		},
		{
			name:         "namespace extracted from flags not duplicated",
			cmdPath:      "start/plan",
			flags:        map[string]any{"name": "my-plan", "namespace": "real-ns"},
			wantContains: []string{"--namespace", "real-ns"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildWriteArgs(tt.cmdPath, tt.flags)
			joined := strings.Join(result, " ")

			for _, want := range tt.wantContains {
				if !strings.Contains(joined, want) {
					t.Errorf("buildWriteArgs() = %v, should contain %q", result, want)
				}
			}
			for _, notWant := range tt.wantMissing {
				if strings.Contains(joined, notWant) {
					t.Errorf("buildWriteArgs() = %v, should NOT contain %q", result, notWant)
				}
			}
		})
	}
}

// --- Handler validation error tests ---

func TestHandleMTVWrite_ValidationErrors(t *testing.T) {
	registry := testRegistry()
	handler := HandleMTVWrite(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name      string
		input     MTVWriteInput
		wantError string
	}{
		{
			name:      "unknown command",
			input:     MTVWriteInput{Command: "nonexistent command"},
			wantError: "unknown command",
		},
		{
			name:      "read command rejected",
			input:     MTVWriteInput{Command: "get plan"},
			wantError: "read-only operation",
		},
		{
			name:      "full CLI command rejected",
			input:     MTVWriteInput{Command: "kubectl-mtv create provider --name test"},
			wantError: "subcommand path",
		},
		{
			name:      "embedded tool output rejected",
			input:     MTVWriteInput{Command: `create plan {"output": "something"}`},
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

func TestHandleMTVWrite_DryRun(t *testing.T) {
	registry := &discovery.Registry{
		ReadOnly: map[string]*discovery.Command{
			"get/plan": {Path: []string{"get", "plan"}, PathString: "get plan", Description: "Get plans"},
		},
		ReadWrite: map[string]*discovery.Command{
			"create/provider": {
				Path: []string{"create", "provider"}, PathString: "create provider", Description: "Create provider",
			},
			"delete/plan": {
				Path: []string{"delete", "plan"}, PathString: "delete plan", Description: "Delete plan",
			},
			"start/plan": {
				Path: []string{"start", "plan"}, PathString: "start plan", Description: "Start plan",
			},
		},
	}

	handler := HandleMTVWrite(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name         string
		input        MTVWriteInput
		wantContains []string
	}{
		{
			name: "create provider with flags",
			input: MTVWriteInput{
				Command: "create provider",
				Flags: map[string]any{
					"name": "my-vsphere",
					"type": "vsphere",
					"url":  "https://vcenter.example.com",
				},
				DryRun: true,
			},
			wantContains: []string{"kubectl-mtv", "create", "provider", "--name", "my-vsphere", "--type", "vsphere", "--url"},
		},
		{
			name: "delete plan with name and namespace",
			input: MTVWriteInput{
				Command: "delete plan",
				Flags:   map[string]any{"name": "old-plan", "namespace": "demo"},
				DryRun:  true,
			},
			wantContains: []string{"kubectl-mtv", "delete", "plan", "old-plan", "--namespace", "demo"},
		},
		{
			name: "start plan with named arg",
			input: MTVWriteInput{
				Command: "start plan",
				Flags:   map[string]any{"name": "my-plan"},
				DryRun:  true,
			},
			wantContains: []string{"kubectl-mtv", "start", "plan", "my-plan"},
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

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output = %q, should contain %q", output, want)
				}
			}
		})
	}
}
