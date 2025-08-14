# Setting Up kubectl-mtv MCP Server

This guide explains how to set up and use the kubectl-mtv MCP server with various MCP-compatible applications like Cursor, Claude Desktop, and other MCP clients.

## Prerequisites

Before setting up the MCP server, ensure you have:

1. **kubectl-mtv installed** and available in your PATH
2. **Kubernetes cluster access** with MTV (Migration Toolkit for Virtualization) deployed
3. **Python 3.8+** installed
4. **MCP-compatible client** (Cursor, Claude Desktop, etc.)

## Quick Setup

1. **Install the MTV MCP package**:
   ```bash
   pip install mtv-mcp
   ```

2. **Configure your MCP client** (see client-specific instructions below)

## Client Configuration

### Cursor IDE

1. **Open Cursor Settings**:
   - Press `Cmd/Ctrl + ,` to open settings
   - Search for "MCP" or look for Model Context Protocol settings

2. **Add the MCP server configuration**:
   ```json
   {
     "mcpServers": {
       "kubectl-mtv": {
         "command": "kubectl-mtv-mcp",
         "args": [],
         "env": {}
       }
     }
   }
   ```

3. **Restart Cursor** for the changes to take effect.

### Claude Desktop

1. **Locate the configuration file**:
   - **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
   - **Linux**: `~/.config/claude/claude_desktop_config.json`

2. **Add the server configuration**:
   ```json
   {
     "mcpServers": {
       "kubectl-mtv": {
         "command": "kubectl-mtv-mcp",
         "args": [],
         "env": {}
       }
     }
   }
   ```

3. **Restart Claude Desktop**.

### Claude Code (CLI Installation)

For Claude Code users, you can use the built-in `mcp install` command for the easiest setup:

1. **Install the MCP server directly**:
   ```bash
   claude mcp add kubectl-mtv-mcp
   ```

2. **Verify the installation**:
   ```bash
   claude mcp list
   ```
   You should see `kubectl-mtv` in the list of installed servers.

3. **Optional: Install the write server** for destructive operations:
   ```bash
   # Only install if you need write/modify capabilities
   claude mcp add kubectl-mtv-write-mcp
   ```

5. **The server(s) are now available** - Claude Code will automatically load and use the servers.

#### Auto-Execution Settings (Optional)

By default, Claude CLI may prompt for user consent before executing MCP tool operations. To enable automatic execution without prompts:

1. **Create or edit Claude CLI configuration**:
   ```bash
   # Create config directory if it doesn't exist
   mkdir -p ~/.config/claude
   
   # Edit the configuration file
   nano ~/.config/claude/config.json
   ```

2. **Add MCP auto-execution settings**:
   ```json
   {
     "mcp": {
       "servers": {
         "kubectl-mtv": {
           "command": "kubectl-mtv-mcp",
           "args": [],
           "autoExecute": true,
           "confirmBeforeExecution": false,
           "trusted": true
         }
       }
     },
     "execution": {
       "autoExecute": true,
       "trustedServers": ["kubectl-mtv"]
     }
   }
   ```

3. **Alternative: Use environment variables**:
   ```bash
   # Add to your ~/.bashrc or ~/.zshrc
   export CLAUDE_MCP_AUTO_EXECUTE=true
   export CLAUDE_MCP_TRUSTED_SERVERS="kubectl-mtv"
   export CLAUDE_MCP_CONFIRM_EXECUTION=false
   ```

4. **Command-line flags** (temporary override):
   ```bash
   # Run Claude CLI with auto-execution enabled
   claude --mcp-auto-execute --mcp-trusted-server kubectl-mtv
   ```

### Generic MCP Client

For other MCP-compatible applications, use this general configuration format:

```json
{
  "servers": {
    "kubectl-mtv": {
      "command": "kubectl-mtv-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

## HTTP Server Mode (Alternative Setup)

Instead of using STDIO transport (where each client spawns its own server process), you can run the MCP server as a persistent HTTP service. This is useful for:

- **Multiple clients** connecting to the same server instance
- **Remote access** to the MCP server over the network
- **Production deployments** where you want a long-running service
- **Docker/container environments** where HTTP is preferred

### Setting Up HTTP Server

1. **Start the HTTP server**:
   ```bash
   # Default: localhost:8000
   kubectl-mtv-mcp --transport http
   
   # Custom host/port
   kubectl-mtv-mcp --transport http --host 0.0.0.0 --port 9000
   
   # With debug logging
   kubectl-mtv-mcp --transport http --log-level debug
   ```

2. **Configure your MCP client** to use HTTP transport:
   ```json
   {
     "mcpServers": {
       "kubectl-mtv": {
         "transport": "http",
         "url": "http://127.0.0.1:8000/mcp",
         "description": "kubectl-mtv MCP server running as HTTP service"
       }
     }
   }
   ```

3. **Restart your MCP client**.

### HTTP Server Options

Available command-line options for HTTP mode:

- `--transport http` - Enable HTTP transport
- `--host HOST` - Host to bind to (default: 127.0.0.1)
- `--port PORT` - Port to bind to (default: 8000)
- `--path PATH` - Endpoint path (default: /mcp)
- `--log-level LEVEL` - Log level (debug, info, warning, error)

## Security Considerations

- The MCP server executes kubectl-mtv commands with your current Kubernetes permissions
- Ensure your MCP client is running in a secure environment
- Consider using dedicated service accounts for production environments
- Be cautious when sharing MCP server access with others
