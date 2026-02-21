---
name: mcp-tool-development
description: Guide for adding or modifying MCP tools and understanding command discovery in kubectl-mtv. Use when working on the MCP server, adding new tools, or changing how commands are exposed to AI assistants.
---

# MCP Tool Development

The MCP server exposes kubectl-mtv commands as structured tools for AI assistants. CLI commands are auto-discovered; MCP tools rarely need changes.

## Architecture

```
cmd/mcpserver/mcpserver.go       -> Server startup, tool registration
pkg/mcp/discovery/registry.go   -> Command discovery from help --machine
pkg/mcp/tools/mtv_read.go       -> Read-only tool (get, describe, health)
pkg/mcp/tools/mtv_write.go      -> Write tool (create, delete, patch, start)
pkg/mcp/tools/mtv_help.go       -> Help/documentation tool
pkg/mcp/util/util.go            -> Command execution, response parsing
```

## Auto-Discovery (Most Common Case)

New CLI commands are automatically exposed via MCP. No MCP code changes needed if:

1. The command is under a known verb (`get`, `describe`, `create`, `delete`, `patch`, `start`, `cancel`, `archive`, `unarchive`, `cutover`, `health`).
2. The verb maps to a category in `pkg/cmd/help/generator.go` `getCategory()`.
3. The command is runnable (leaf node with `RunE`).

The discovery flow: `help --machine` -> JSON schema -> `Registry` -> tool descriptions.

## Tool Input Pattern

Each tool has an input struct with `jsonschema` tags:

```go
type MTVReadInput struct {
    Command string         `json:"command" jsonschema:"Command path (e.g. get plan, get inventory vm)"`
    Flags   map[string]any `json:"flags,omitempty" jsonschema:"All parameters as key-value pairs"`
    DryRun  bool           `json:"dry_run,omitempty" jsonschema:"If true, returns CLI command instead of executing"`
    Fields  []string       `json:"fields,omitempty" jsonschema:"Limit JSON to these top-level keys only"`
}
```

## Tool Definition Pattern

```go
func GetMTVReadTool(registry *discovery.Registry) *mcp.Tool {
    return &mcp.Tool{
        Name:         "mtv_read",
        Description:  registry.GenerateReadOnlyDescription(),
        OutputSchema: mtvOutputSchema,
    }
}
```

## Handler Pattern

```go
func HandleMTVRead(registry *discovery.Registry) func(context.Context, *mcp.CallToolRequest, MTVReadInput) (*mcp.CallToolResult, any, error) {
    return func(ctx context.Context, req *mcp.CallToolRequest, input MTVReadInput) (*mcp.CallToolResult, any, error) {
        ctx = extractKubeCredsFromRequest(ctx, req)
        if err := validateCommandInput(input.Command); err != nil {
            return nil, nil, err
        }
        cmdPath := normalizeCommandPath(input.Command)
        if !registry.IsReadOnly(cmdPath) {
            return nil, nil, fmt.Errorf("unknown or write command '%s'", input.Command)
        }
        if input.DryRun {
            ctx = util.WithDryRun(ctx, true)
        }
        args := buildArgs(cmdPath, input.Flags)
        result, err := util.RunKubectlMTVCommand(ctx, args)
        if err != nil {
            return nil, nil, fmt.Errorf("command failed: %w", err)
        }
        data, err := util.UnmarshalJSONResponse(result)
        if err != nil {
            return nil, nil, err
        }
        if errResult := buildCLIErrorResult(data); errResult != nil {
            return errResult, nil, nil
        }
        return nil, data, nil
    }
}
```

## Tool Registration

In `cmd/mcpserver/mcpserver.go` `createMCPServerWithHeaderCapture()`:

```go
// Read tools (always registered)
tools.AddToolWithCoercion(server, tools.GetMTVReadTool(registry),
    wrapWithHeaders(tools.HandleMTVRead(registry), capturedHeaders))
mcp.AddTool(server, tools.GetMTVHelpTool(),
    wrapWithHeaders(tools.HandleMTVHelp, capturedHeaders))

// Write tool (skipped in read-only mode)
if !readOnlyMode {
    tools.AddToolWithCoercion(server, tools.GetMTVWriteTool(registry),
        wrapWithHeaders(tools.HandleMTVWrite(registry), capturedHeaders))
}
```

- `AddToolWithCoercion` handles boolean coercion (string "true"/"false" -> bool). Use for tools with boolean flags.
- `mcp.AddTool` is standard registration. Use when no coercion needed.
- `wrapWithHeaders` injects HTTP auth headers for SSE mode.

## Adding a New MCP Tool

1. Create `pkg/mcp/tools/mtv_newtool.go` with input struct, `GetTool()`, and `HandleTool()`.
2. Register in `cmd/mcpserver/mcpserver.go` `createMCPServerWithHeaderCapture()`.
3. Add tests in `pkg/mcp/tools/mtv_newtool_test.go`.

## Key Utilities

| Function | Location | Purpose |
|----------|----------|---------|
| `util.RunKubectlMTVCommand()` | `pkg/mcp/util/util.go` | Execute kubectl-mtv subprocess |
| `util.UnmarshalJSONResponse()` | `pkg/mcp/util/util.go` | Parse JSON, strip `command` field, truncate |
| `buildCLIErrorResult()` | `pkg/mcp/tools/mtv_read.go` | Convert non-zero exit to `IsError` response |
| `validateCommandInput()` | `pkg/mcp/tools/mtv_read.go` | Reject garbled LLM input early |
| `normalizeCommandPath()` | `pkg/mcp/tools/mtv_read.go` | `"get plan"` -> `"get/plan"` |
| `buildArgs()` | `pkg/mcp/tools/mtv_read.go` | Flags map -> CLI args |
| `help.MarkMCPHidden()` | `pkg/cmd/help/generator.go` | Hide flags from MCP schema |

## Shared Output Schema

All tools return the same envelope:

```json
{
  "return_value": 0,
  "data": {},
  "output": "text output",
  "stderr": "error output"
}
```

## Registry

The `Registry` (`pkg/mcp/discovery/`) splits commands into `ReadOnly` and `ReadWrite` maps keyed by path (`"get/plan"`, `"create/provider"`). It also holds `GlobalFlags`, `Parents` (non-runnable nodes like `get/inventory`), and generates tool descriptions.
