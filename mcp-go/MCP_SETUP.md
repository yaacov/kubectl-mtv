# Setting Up kubectl-mtv MCP Servers

This guide explains how to set up and use the kubectl-mtv MCP servers with various MCP-compatible applications like Cursor, Claude Desktop, and other MCP clients.

## What is MCP?

MCP (Model Context Protocol) is an open standard that enables AI assistants to interact directly with tools and services in your local environment. For kubectl-mtv, MCP creates a bridge between AI coding assistants and your Kubernetes clusters, allowing these assistants to execute kubectl-mtv commands, retrieve VM migration information, and manage KubeVirt VMs without requiring you to copy/paste output. When configured, your AI assistant can directly query cluster resources, check migration status, and even help troubleshoot issues by having real-time access to your environment. MCP maintains security by running entirely on your local machine, with the AI only receiving the command results rather than direct cluster access.

## Prerequisites

Before setting up the MCP servers, ensure you have:

1. **kubectl-mtv installed** and available in your PATH (for MTV operations)
2. **virtctl installed** and available in your PATH (for KubeVirt operations)
3. **Kubernetes cluster access** with MTV and/or KubeVirt deployed
4. **MCP-compatible client** (Cursor, Claude Desktop, etc.)

## Quick Setup

### Option A: Download Pre-built Binaries (Recommended)

```bash
# Download the latest release for your platform
# Linux AMD64
curl -LO https://github.com/yaacov/kubectl-mtv/releases/latest/download/mcp-servers-v0.0.1-linux-amd64.tar.gz

# macOS Apple Silicon (M1/M2)
curl -LO https://github.com/yaacov/kubectl-mtv/releases/latest/download/mcp-servers-v0.0.1-darwin-arm64.tar.gz

# macOS Intel
curl -LO https://github.com/yaacov/kubectl-mtv/releases/latest/download/mcp-servers-v0.0.1-darwin-amd64.tar.gz

# Verify the download (optional but recommended)
curl -LO https://github.com/yaacov/kubectl-mtv/releases/latest/download/mcp-servers-v0.0.1-darwin-amd64.tar.gz.sha256sum
sha256sum -c mcp-servers-v0.0.1-darwin-amd64.tar.gz.sha256sum

# Extract the binaries
tar -xzf mcp-servers-*.tar.gz

# Install to your PATH (choose one location)
# Option 1: User-local installation
mkdir -p ~/.local/bin
mv kubectl-mtv-mcp kubectl-mtv-write-mcp virtctl-mcp ~/.local/bin/

# Option 2: System-wide installation (requires sudo)
sudo mv kubectl-mtv-mcp kubectl-mtv-write-mcp virtctl-mcp /usr/local/bin/

# Verify installation
kubectl-mtv-mcp --version
kubectl-mtv-write-mcp --version
virtctl-mcp --version
```

**Note**: Make sure `~/.local/bin` is in your PATH. Add this to your `~/.bashrc` or `~/.zshrc` if needed:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Option B: Build from Source

```bash
# Clone the repository
git clone https://github.com/yaacov/kubectl-mtv.git
cd kubectl-mtv/mcp-go

# Build all servers
make build

# Install to your PATH
make install

# Or build for all platforms
make build-all
```

## Server Overview

The MCP servers provide three independent servers:

### 1. Read Server (`kubectl-mtv-mcp`)
**Safe for all users** - Read-only operations for monitoring and troubleshooting:
- List MTV resources (providers, plans, mappings, hosts, hooks)
- Query provider inventories with SQL-like syntax
- Retrieve logs from MTV controller and importer pods
- Access migration storage (PVCs, DataVolumes)
- Get version information
- View VM status in migration plans

### 2. Write Server (`kubectl-mtv-write-mcp`)
**USE WITH CAUTION** - Full lifecycle management:
- Create, modify, and delete providers
- Manage network and storage mappings
- Create and manage migration plans
- Create migration hosts and hooks
- Start, cancel, cutover, archive plans
- Patch existing resources

### 3. Virtctl Server (`virtctl-mcp`)
Comprehensive virtual machine management:
- VM lifecycle: start, stop, restart, pause, migrate
- Diagnostics: guest OS info, filesystem listing, monitoring
- Cluster resources: discover instancetypes, preferences, datasources

## Client Configuration

### Cursor IDE

1. **Open Cursor Settings**:
   - Press `Cmd/Ctrl + ,` to open settings
   - Search for "MCP" or look for Model Context Protocol settings

2. **Add the MCP server configuration**:
   ```json
   {
     "mcpServers": {
       "kubectl-mtv-read": {
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
         "command": "virtctl-mcp",
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
   - **Linux**: `~/.config/claude/claude_desktop_config.json`
   - **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

2. **Add the server configuration**:
   ```json
   {
     "mcpServers": {
       "kubectl-mtv-read": {
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
         "command": "virtctl-mcp",
         "args": [],
         "env": {}
       }
     }
   }
   ```

3. **Restart Claude Desktop**.

### Claude Code (CLI)

1. **Add the servers manually**:
   ```bash
   # MTV read-only server (recommended for most users)
   claude mcp add kubectl-mtv-read kubectl-mtv-mcp
   
   # MTV write server (USE WITH CAUTION - can modify/delete resources)  
   claude mcp add kubectl-mtv-write kubectl-mtv-write-mcp
   
   # KubeVirt server (VM management)
   claude mcp add kubevirt virtctl-mcp
   ```

2. **Verify the installation**:
   ```bash
   claude mcp list
   ```
   You should see `kubectl-mtv-read`, `kubectl-mtv-write`, and `kubevirt` in the list.

3. **The servers are now available** - Claude Code will automatically load them.

#### Using Absolute Paths

If the binaries are not in your PATH, use absolute paths:

```bash
# Add using full paths (adjust paths based on your installation)
claude mcp add kubectl-mtv-read ~/.local/bin/kubectl-mtv-mcp
claude mcp add kubectl-mtv-write ~/.local/bin/kubectl-mtv-write-mcp
claude mcp add kubevirt ~/.local/bin/virtctl-mcp
```

#### Auto-Execution Settings (Optional)

To enable automatic execution without prompts:

1. **Edit Claude CLI configuration**:
   ```bash
   mkdir -p ~/.config/claude
   nano ~/.config/claude/config.json
   ```

2. **Add auto-execution settings**:
   ```json
   {
     "mcp": {
       "servers": {
         "kubectl-mtv-read": {
           "command": "kubectl-mtv-mcp",
           "autoExecute": true,
           "confirmBeforeExecution": false,
           "trusted": true
         },
        "kubectl-mtv-write": {
          "command": "kubectl-mtv-write-mcp",
          "autoExecute": false,
          "confirmBeforeExecution": true,
          "trusted": false
        },
         "kubevirt": {
           "command": "virtctl-mcp",
           "autoExecute": false,
           "confirmBeforeExecution": true,
           "trusted": false
         }
       }
     }
   }
   ```

### Generic MCP Client

For other MCP-compatible applications, use this general configuration format:

```json
{
  "servers": {
    "kubectl-mtv-read": {
      "command": "kubectl-mtv-mcp",
      "args": []
    },
    "kubectl-mtv-write": {
      "command": "kubectl-mtv-write-mcp",
      "args": []
    },
    "kubevirt": {
      "command": "virtctl-mcp",
      "args": []
    }
  }
}
```

## Advanced Configuration

### Custom Binary Locations

If you installed the binaries in a custom location:

```json
{
  "mcpServers": {
    "kubectl-mtv-read": {
      "command": "/custom/path/to/kubectl-mtv-mcp",
      "args": []
    }
  }
}
```

### Environment Variables

Pass custom environment variables to the servers:

```json
{
  "mcpServers": {
    "kubectl-mtv-read": {
      "command": "kubectl-mtv-mcp",
      "args": [],
      "env": {
        "KUBECONFIG": "/path/to/custom/kubeconfig",
        "LOG_LEVEL": "debug"
      }
    }
  }
}
```

### Multiple Kubernetes Contexts

To use different servers for different clusters:

```json
{
  "mcpServers": {
    "mtv-prod": {
      "command": "kubectl-mtv-mcp",
      "env": {
        "KUBECONFIG": "/path/to/prod-kubeconfig"
      }
    },
    "mtv-dev": {
      "command": "kubectl-mtv-mcp",
      "env": {
        "KUBECONFIG": "/path/to/dev-kubeconfig"
      }
    }
  }
}
```

## Troubleshooting

### Servers Not Found

If you get "command not found" errors:

1. **Check if binaries are installed and locate them**:
   ```bash
   # If in PATH:
   which kubectl-mtv-mcp kubectl-mtv-write-mcp virtctl-mcp
   
   # Check common installation locations:
   ls -l ~/.local/bin/kubectl-mtv-* ~/.local/bin/virtctl-mcp
   ls -l /usr/local/bin/kubectl-mtv-* /usr/local/bin/virtctl-mcp
   
   # Search your system:
   find ~ -name "kubectl-mtv-mcp" -type f 2>/dev/null
   find /usr -name "kubectl-mtv-mcp" -type f 2>/dev/null
   ```

2. **Verify PATH** (and add binary location if needed):
   ```bash
   echo $PATH
   
   # Add to PATH if binaries are in ~/.local/bin
   export PATH="$HOME/.local/bin:$PATH"
   # Add to ~/.bashrc or ~/.zshrc to make permanent
   ```

3. **Use absolute paths** in your MCP client configuration with the paths you found

### Permission Denied

If you get permission errors:

```bash
# Make binaries executable
chmod +x ~/.local/bin/kubectl-mtv-mcp
chmod +x ~/.local/bin/kubectl-mtv-write-mcp
chmod +x ~/.local/bin/virtctl-mcp
```

### Connection Issues

If the MCP client can't connect:

1. **Test the servers manually**:
   ```bash
   # Each server should start and wait for input
   kubectl-mtv-mcp
   # Press Ctrl+C to exit
   ```

2. **Check MCP client logs** for error messages

3. **Restart the MCP client** after configuration changes

### Version Mismatches

Ensure all components are up to date:

```bash
# Check server versions
kubectl-mtv-mcp --version
kubectl-mtv-write-mcp --version
virtctl-mcp --version

# Check kubectl-mtv version
kubectl-mtv version

# Check virtctl version
virtctl version
```

## Security Considerations

- The MCP servers execute kubectl-mtv and virtctl commands with your current Kubernetes permissions
- **Read server**: Safe to use - only performs read operations
- **Write server**: **USE WITH CAUTION** - can create, modify, and delete resources
- Ensure your MCP client is running in a secure environment
- Consider using dedicated service accounts for production environments
- By default, servers use stdio transport (no network exposure).
- When started with --sse, an HTTP server is exposed on the configured host:port.
  Prefer binding to 127.0.0.1, and restrict access via firewall when necessary.

## Verification

Test that everything is working:

1. **Open your MCP client** (Cursor, Claude Desktop, etc.)

2. **Ask the AI to check MTV resources**:
   - "List all MTV providers"
   - "Show migration plans"
   - "Get kubectl-mtv version"

3. **The AI should be able to execute these commands** and show you the results.

## Next Steps

- Read the [README.md](README.md) for tool descriptions and capabilities
- Check the [main kubectl-mtv documentation](../README.md) for kubectl-mtv usage
- See the [MCP documentation](https://modelcontextprotocol.io/) for more about MCP

## Getting Help

If you encounter issues:

1. Check the troubleshooting section above
2. Open an issue on [GitHub](https://github.com/yaacov/kubectl-mtv/issues)
3. Include:
   - Your operating system
   - MCP client (Cursor, Claude Desktop, etc.)
   - Error messages from the client logs
   - Output of `kubectl-mtv-mcp --version`, `kubectl-mtv version`, `virtctl version`

