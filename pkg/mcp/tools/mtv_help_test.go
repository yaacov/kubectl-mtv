package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// --- Tool definition tests ---

func TestGetMTVHelpTool(t *testing.T) {
	tool := GetMTVHelpTool()

	if tool.Name != "mtv_help" {
		t.Errorf("Name = %q, want %q", tool.Name, "mtv_help")
	}

	if tool.Description == "" {
		t.Error("Description should not be empty")
	}

	// Description should reference key capabilities
	for _, keyword := range []string{"help", "flags", "tsl", "karl"} {
		if !strings.Contains(strings.ToLower(tool.Description), keyword) {
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

// --- Handler validation error tests ---

func TestHandleMTVHelp_EmptyCommand(t *testing.T) {
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, _, err := HandleMTVHelp(ctx, req, MTVHelpInput{Command: ""})
	if err == nil {
		t.Fatal("expected error for empty command, got nil")
	}
	if !strings.Contains(err.Error(), "command is required") {
		t.Errorf("error = %q, should contain 'command is required'", err.Error())
	}
}

func TestHandleMTVHelp_WhitespaceCommand(t *testing.T) {
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, _, err := HandleMTVHelp(ctx, req, MTVHelpInput{Command: "   "})
	if err == nil {
		t.Fatal("expected error for whitespace command, got nil")
	}
	if !strings.Contains(err.Error(), "command is required") {
		t.Errorf("error = %q, should contain 'command is required'", err.Error())
	}
}
