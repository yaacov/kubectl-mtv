package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/pkg/mcp/util"
)

// MTVHelpInput represents the input for the mtv_help tool.
type MTVHelpInput struct {
	// Command is the kubectl-mtv command or topic to get help for.
	// Examples: "create plan", "get inventory vm", "tsl", "karl"
	Command string `json:"command" jsonschema:"Command or topic (e.g. create plan, get inventory vm, tsl, karl)"`
}

// GetMTVHelpTool returns the tool definition for on-demand help.
func GetMTVHelpTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "mtv_help",
		Description: `Get help: flags, usage, examples for any command, or syntax refs for topics.

Commands: any from mtv_read/mtv_write (e.g. "create plan", "get inventory vm")
Topics: "tsl" (query language), "karl" (affinity rules)`,
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command":      map[string]any{"type": "string", "description": "Executed command"},
				"return_value": map[string]any{"type": "integer", "description": "Exit code (0=success)"},
				"data": map[string]any{
					"description": "Help output (object or array)",
					"oneOf": []map[string]any{
						{"type": "object"},
						{"type": "array"},
					},
				},
				"output": map[string]any{"type": "string", "description": "Text output"},
				"stderr": map[string]any{"type": "string", "description": "Error output"},
			},
		},
	}
}

// HandleMTVHelp handles the mtv_help tool invocation.
func HandleMTVHelp(ctx context.Context, req *mcp.CallToolRequest, input MTVHelpInput) (*mcp.CallToolResult, any, error) {
	// Extract K8s credentials from HTTP headers (for SSE mode)
	if req.Extra != nil && req.Extra.Header != nil {
		ctx = util.WithKubeCredsFromHeaders(ctx, req.Extra.Header)
	}

	command := strings.TrimSpace(input.Command)
	if command == "" {
		return nil, nil, fmt.Errorf("command is required (e.g. \"create plan\", \"tsl\", \"karl\")")
	}

	// Build args: help --machine [command parts...]
	args := []string{"help", "--machine"}
	parts := strings.Fields(command)
	args = append(args, parts...)

	// Execute kubectl-mtv help --machine [command]
	result, err := util.RunKubectlMTVCommand(ctx, args)
	if err != nil {
		return nil, nil, fmt.Errorf("help command failed: %w", err)
	}

	// Parse and return result
	data, err := util.UnmarshalJSONResponse(result)
	if err != nil {
		return nil, nil, err
	}

	return nil, data, nil
}
