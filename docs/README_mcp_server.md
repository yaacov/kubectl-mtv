# MCP Server Guide

kubectl-mtv includes a built-in MCP (Model Context Protocol) server that provides AI assistants with comprehensive access to Forklift migration resources.

## Overview

The MCP server enables AI assistants (like Claude, Cursor IDE, and other MCP-compatible tools) to:

- **Read operations**: List resources, query inventory, get logs, monitor migrations
- **Write operations**: Create/delete/patch providers, plans, mappings, and more

## Installation

The MCP server is built into kubectl-mtv. No separate installation is required.

```bash
# Verify installation
kubectl mtv mcp-server --help
```

## Server Modes

### Stdio Mode (Default)

Stdio mode is designed for AI assistants that communicate via standard input/output (stdin/stdout). This is the recommended mode for most AI integrations.

```bash
kubectl mtv mcp-server
```

### SSE Mode (HTTP Server)

SSE (Server-Sent Events) mode runs an HTTP server that provides MCP access over HTTP. This is useful for web-based integrations or remote access.

```bash
kubectl mtv mcp-server --sse --host 127.0.0.1 --port 8080
```

## Command Line Options

```bash
kubectl mtv mcp-server [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--sse` | boolean | false | Run in SSE (Server-Sent Events) mode over HTTP |
| `--host` | string | 127.0.0.1 | Host address to bind to for SSE mode |
| `--port` | string | 8080 | Port to listen on for SSE mode |

## Integration with AI Assistants

### Claude Desktop

**Option 1: Using Claude CLI (Recommended)**

If you have the Claude CLI installed, you can add the MCP server interactively:

```bash
claude mcp add kubectl-mtv kubectl mtv mcp-server
```

When prompted, provide:
- **Command**: `kubectl`
- **Args**: `mtv mcp-server`

**Option 2: Manual Configuration**

1. Open Claude Desktop configuration:
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
   - Linux: `~/.config/Claude/claude_desktop_config.json`

2. Add kubectl-mtv MCP server:

```json
{
  "mcpServers": {
    "kubectl-mtv": {
      "command": "kubectl",
      "args": ["mtv", "mcp-server"]
    }
  }
}
```

3. Restart Claude Desktop

**Verification:**

After adding the server, you can verify it's working by asking Claude:
```
List all MTV providers
```

If configured correctly, Claude will be able to access your kubectl-mtv resources.

### Cursor IDE

1. Open Cursor Settings
2. Navigate to MCP settings
3. Add a new MCP server:
   - **Name**: kubectl-mtv
   - **Command**: `kubectl`
   - **Args**: `mtv mcp-server`

4. Restart Cursor

### Custom Integration

For custom integrations, connect to the MCP server using the MCP SDK:

**Stdio mode:**
```python
from mcp import StdioClient

client = StdioClient(
    command="kubectl",
    args=["mtv", "mcp-server"]
)
```

**SSE mode:**
```python
from mcp import SSEClient

client = SSEClient(
    url="http://127.0.0.1:8080/sse"
)
```
