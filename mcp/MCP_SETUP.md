# Setting Up kubectl-mtv MCP Server

This guide explains how to set up and use the kubectl-mtv MCP server with various MCP-compatible applications like Cursor, Claude Desktop, and other MCP clients.

## What is MCP?

MCP (Model Context Protocol) is an open standard that enables AI assistants to interact directly with tools and services in your local environment. For kubectl-mtv, MCP creates a bridge between AI coding assistants and your Kubernetes clusters, allowing these assistants to execute kubectl-mtv commands, retrieve VM migration information, and manage KubeVirt VMs without requiring you to copy/paste output. When configured, your AI assistant can directly query cluster resources, check migration status, and even help troubleshoot issues by having real-time access to your environment. MCP maintains security by running entirely on your local machine, with the AI only receiving the command results rather than direct cluster access.

## Prerequisites

Before setting up the MCP server, ensure you have:

1. **kubectl-mtv installed** and available in your PATH (for MTV operations)
2. **virtctl installed** and available in your PATH (for KubeVirt operations)
3. **Kubernetes cluster access** with MTV and/or KubeVirt deployed
4. **Python 3.8+** installed
5. **MCP-compatible client** (Cursor, Claude Desktop, etc.)

## Quick Setup
   
   **Option A: Using pip**
   ```bash
   pip install mtv-mcp
   ```
   
   **Option B: Download from GitHub releases**
   ```bash
   # Download the server executables
   curl -LO https://github.com/yaacov/kubectl-mtv/releases/latest/download/kubectl-mtv-mcp-servers-linux-amd64.tar.gz
   
   # Download the checksum file
   curl -LO https://github.com/yaacov/kubectl-mtv/releases/latest/download/kubectl-mtv-mcp-servers-linux-amd64.tar.gz.sha256sum
   
   # Verify the download (compare output with the content of the .sha256sum file)
   sha256sum kubectl-mtv-mcp-servers-linux-amd64.tar.gz
   
   # Extract the executables
   tar -xzf kubectl-mtv-mcp-servers-linux-amd64.tar.gz
   
   # Move the executables to a directory in your PATH
   # (assuming ~/.local/bin is in your PATH)
   mkdir -p ~/.local/bin
   mv kubectl-mtv-mcp kubectl-mtv-write-mcp kubevirt-mcp ~/.local/bin/
   
   # Clean up downloaded files
   rm kubectl-mtv-mcp-servers-linux-amd64.tar.gz kubectl-mtv-mcp-servers-linux-amd64.tar.gz.sha256sum
   ```
   
   Note: If ~/.local/bin is not in your PATH, either add it or use another directory that is (like /usr/local/bin, which may require sudo).

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
       },
       "kubectl-mtv-write": {
         "command": "kubectl-mtv-write-mcp",
         "args": [],
         "env": {}
       },
       "kubevirt": {
         "command": "kubevirt-mcp",
         "args": [],
         "env": {}
       }
     }
   }
   ```

3. **Restart Cursor** for the changes to take effect.

### Claude Desktop

1. **Locate the configuration file**:
   - **Linux**: `~/.config/claude/claude_desktop_config.json`

2. **Add the server configuration**:
   ```json
   {
     "mcpServers": {
       "kubectl-mtv": {
         "command": "kubectl-mtv-mcp",
         "args": [],
         "env": {}
       },
       "kubectl-mtv-write": {
         "command": "kubectl-mtv-write-mcp",
         "args": [],
         "env": {}
       },
       "kubevirt": {
         "command": "kubevirt-mcp",
         "args": [],
         "env": {}
       }
     }
   }
   ```

3. **Restart Claude Desktop**.

### Claude Code (CLI Installation)

For Claude Code users, you can use the built-in `mcp install` command for the easiest setup:

1. **Install the MCP servers directly**:
   ```bash
   # MTV read-only server (recommended for most users)
   claude mcp add kubectl-mtv-mcp
   
   # MTV write server (USE WITH CAUTION - can modify/delete resources)  
   claude mcp add kubectl-mtv-write-mcp
   
   # KubeVirt server (VM management through virtctl commands)
   claude mcp add kubevirt-mcp
   ```

2. **Verify the installation**:
   ```bash
   claude mcp list
   ```
   You should see `kubectl-mtv`, `kubectl-mtv-write`, and `kubevirt` in the list of installed servers.

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
         },
         "kubectl-mtv-write": {
           "command": "kubectl-mtv-write-mcp",
           "args": [],
           "autoExecute": true,
           "confirmBeforeExecution": false,
           "trusted": true
         },
         "kubevirt": {
           "command": "kubevirt-mcp",
           "args": [],
           "autoExecute": true,
           "confirmBeforeExecution": false,
           "trusted": true
         }
       }
     },
     "execution": {
       "autoExecute": true,
       "trustedServers": ["kubectl-mtv", "kubectl-mtv-write", "kubevirt"]
     }
   }
   ```

3. **Alternative: Use environment variables**:
   ```bash
   # Add to your ~/.bashrc or ~/.zshrc
   export CLAUDE_MCP_AUTO_EXECUTE=true
   export CLAUDE_MCP_TRUSTED_SERVERS="kubectl-mtv,kubectl-mtv-write,kubevirt"
   export CLAUDE_MCP_CONFIRM_EXECUTION=false
   ```

4. **Command-line flags** (temporary override):
   ```bash
   # Run Claude CLI with auto-execution enabled for all servers
   claude --mcp-auto-execute --mcp-trusted-server kubectl-mtv --mcp-trusted-server kubectl-mtv-write --mcp-trusted-server kubevirt
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
    },
    "kubectl-mtv-write": {
      "command": "kubectl-mtv-write-mcp",
      "args": [],
      "env": {}
    },
    "kubevirt": {
      "command": "kubevirt-mcp",
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
