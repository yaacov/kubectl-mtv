package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/pkg/mcp/discovery"
	"github.com/yaacov/kubectl-mtv/pkg/mcp/util"
)

// MTVWriteInput represents the input for the mtv_write tool.
type MTVWriteInput struct {
	Command string `json:"command" jsonschema:"Command path (e.g. create provider, delete plan, patch mapping)"`

	Args []string `json:"args,omitempty" jsonschema:"Positional args (resource name)"`

	Flags map[string]any `json:"flags,omitempty" jsonschema:"Flags as key-value pairs (e.g. type: vsphere, url: https://vcenter.example.com)"`

	Namespace string `json:"namespace,omitempty" jsonschema:"Kubernetes namespace"`

	DryRun bool `json:"dry_run,omitempty" jsonschema:"Preview without executing"`
}

// GetMTVWriteTool returns the tool definition for read-write MTV commands.
func GetMTVWriteTool(registry *discovery.Registry) *mcp.Tool {
	description := registry.GenerateReadWriteDescription()

	return &mcp.Tool{
		Name:        "mtv_write",
		Description: description,
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command":      map[string]any{"type": "string", "description": "Executed command"},
				"return_value": map[string]any{"type": "integer", "description": "Exit code (0=success)"},
				"data":         map[string]any{"type": "object", "description": "Response data"},
				"output":       map[string]any{"type": "string", "description": "Text output"},
				"stderr":       map[string]any{"type": "string", "description": "Error output"},
			},
		},
	}
}

// GetMinimalMTVWriteTool returns a minimal tool definition for read-write MTV commands.
// The input schema (jsonschema tags on MTVWriteInput) already describes parameters.
// The description only lists available commands and hints to use mtv_help.
func GetMinimalMTVWriteTool(registry *discovery.Registry) *mcp.Tool {
	description := registry.GenerateMinimalReadWriteDescription()

	return &mcp.Tool{
		Name:        "mtv_write",
		Description: description,
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command":      map[string]any{"type": "string", "description": "Executed command"},
				"return_value": map[string]any{"type": "integer", "description": "Exit code (0=success)"},
				"data":         map[string]any{"type": "object", "description": "Response data"},
				"output":       map[string]any{"type": "string", "description": "Text output"},
				"stderr":       map[string]any{"type": "string", "description": "Error output"},
			},
		},
	}
}

// HandleMTVWrite returns a handler function for the mtv_write tool.
func HandleMTVWrite(registry *discovery.Registry) func(context.Context, *mcp.CallToolRequest, MTVWriteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input MTVWriteInput) (*mcp.CallToolResult, any, error) {
		// Extract K8s credentials from HTTP headers (for SSE mode)
		if req.Extra != nil && req.Extra.Header != nil {
			ctx = util.WithKubeCredsFromHeaders(ctx, req.Extra.Header)
		}

		// Normalize command path
		cmdPath := normalizeCommandPath(input.Command)

		// Validate command exists and is read-write
		if !registry.IsReadWrite(cmdPath) {
			if registry.IsReadOnly(cmdPath) {
				return nil, nil, fmt.Errorf("command '%s' is a read-only operation, use mtv_read tool instead", input.Command)
			}
			// List available commands in error, converting path keys to user-friendly format
			available := registry.ListReadWriteCommands()
			for i, cmd := range available {
				available[i] = strings.ReplaceAll(cmd, "/", " ")
			}
			return nil, nil, fmt.Errorf("unknown command '%s'. Available write commands: %s", input.Command, strings.Join(available, ", "))
		}

		// Enable dry run mode if requested
		if input.DryRun {
			ctx = util.WithDryRun(ctx, true)
		}

		// Build command arguments
		args := buildWriteArgs(cmdPath, input.Args, input.Flags, input.Namespace)

		// Execute command
		result, err := util.RunKubectlMTVCommand(ctx, args)
		if err != nil {
			return nil, nil, fmt.Errorf("command failed: %w", err)
		}

		// Parse and return result
		data, err := util.UnmarshalJSONResponse(result)
		if err != nil {
			return nil, nil, err
		}

		return nil, data, nil
	}
}

// buildWriteArgs builds the command-line arguments for kubectl-mtv write commands.
func buildWriteArgs(cmdPath string, positionalArgs []string, flags map[string]any, namespace string) []string {
	var args []string

	// Add command path parts
	parts := strings.Split(cmdPath, "/")
	args = append(args, parts...)

	// Add positional arguments
	args = append(args, positionalArgs...)

	// Add namespace flag
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	// Note: Write commands typically don't support -o json output format
	// so we don't add it automatically like we do for read commands

	// Skip set for already handled flags
	skipFlags := map[string]bool{
		"namespace": true, "n": true,
	}

	// Add other flags using the normalizer
	args = appendNormalizedFlags(args, flags, skipFlags)

	return args
}
