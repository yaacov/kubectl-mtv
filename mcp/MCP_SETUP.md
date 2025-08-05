# Setting Up kubectl-mtv MCP Server

This guide explains how to set up and use the kubectl-mtv MCP server with various MCP-compatible applications like Cursor, Claude Desktop, and other MCP clients.

## Prerequisites

Before setting up the MCP server, ensure you have:

1. **kubectl-mtv installed** and available in your PATH
2. **Kubernetes cluster access** with MTV (Migration Toolkit for Virtualization) deployed
3. **Python 3.8+** installed
4. **MCP-compatible client** (Cursor, Claude Desktop, etc.)

## Quick Setup

1. **Install dependencies** (if not already installed):
   ```bash
   pip install -r requirements.txt
   ```

2. **Generate configuration files**:
   ```bash
   python generate_config.py
   ```
   This will create configuration files with the correct paths for your installation.

3. **Configure your MCP client** (see client-specific instructions below)

## Client Configuration

### Cursor IDE

1. **Open Cursor Settings**:
   - Press `Cmd/Ctrl + ,` to open settings
   - Search for "MCP" or look for Model Context Protocol settings

2. **Add the MCP server configuration**:
   - **Easy way**: Run `python generate_config.py` and use `cursor-mcp-config-generated.json`
   - **Manual way**: Use the provided `cursor-mcp-config.json` as a template and update the path:

   ```json
   {
     "mcpServers": {
       "kubectl-mtv": {
         "command": "python",
         "args": ["/full/path/to/your/kubectl-mtv/mcp/kubectl_mtv_server.py"],
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
   - **Easy way**: Copy content from `claude-desktop-config-generated.json` (after running `python generate_config.py`)
   - **Manual way**: Add this configuration:
   ```json
   {
     "mcpServers": {
       "kubectl-mtv": {
         "command": "python",
         "args": ["/full/path/to/your/kubectl-mtv/mcp/kubectl_mtv_server.py"],
         "env": {}
       }
     }
   }
   ```

3. **Restart Claude Desktop**.

### Generic MCP Client

For other MCP-compatible applications, use this general configuration format:

```json
{
  "servers": {
    "kubectl-mtv": {
      "command": "python",
      "args": ["/full/path/to/your/kubectl-mtv/mcp/kubectl_mtv_server.py"],
      "env": {}
    }
  }
}
```

## Security Considerations

- The MCP server executes kubectl-mtv commands with your current Kubernetes permissions
- Ensure your MCP client is running in a secure environment
- Consider using dedicated service accounts for production environments
- Be cautious when sharing MCP server access with others
