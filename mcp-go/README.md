# kubectl-mtv MCP Servers

Model Context Protocol (MCP) servers that provide AI assistants with tools to interact with Migration Toolkit for Virtualization (MTV) and KubeVirt through kubectl-mtv and virtctl commands.

## What it does

These MCP servers enable AI assistants to help with MTV and KubeVirt operations by providing:

- **kubectl-mtv-mcp**: Safe operations for monitoring, troubleshooting, and discovering MTV resources and provider inventories
- **kubectl-mtv-write-mcp**: **USE WITH CAUTION** - Full lifecycle management including creating, modifying, and deleting MTV resources
- **virtctl-mcp**: Comprehensive virtual machine management through virtctl commands including VM lifecycle, diagnostics, and cluster resource discovery

## Key Features

These MCP servers offer:

- **Single binary deployment** - Self-contained executables with no runtime dependencies
- **Smaller footprint** - Optimized binaries (~5MB each)
- **Better performance** - Native compilation and efficient resource usage
- **Easy distribution** - Static binaries for multiple platforms
- **Consistent tooling** - Seamlessly integrates with kubectl-mtv

## Prerequisites

- `kubectl-mtv` binary installed and available in PATH (for MTV operations)
- `virtctl` binary installed and available in PATH (for KubeVirt operations)  
- Access to a Kubernetes cluster with MTV and/or KubeVirt deployed

## Quick Installation

### Option A: Download Pre-built Binaries (Recommended)

```bash
# Download for your platform
# Linux AMD64
curl -LO https://github.com/yaacov/kubectl-mtv/releases/latest/download/mcp-servers-v0.0.1-linux-amd64.tar.gz

# macOS Apple Silicon
curl -LO https://github.com/yaacov/kubectl-mtv/releases/latest/download/mcp-servers-v0.0.1-darwin-arm64.tar.gz

# macOS Intel
curl -LO https://github.com/yaacov/kubectl-mtv/releases/latest/download/mcp-servers-v0.0.1-darwin-amd64.tar.gz

# Extract and install
tar -xzf mcp-servers-*.tar.gz
mkdir -p ~/.local/bin
mv kubectl-mtv-mcp kubectl-mtv-write-mcp virtctl-mcp ~/.local/bin/
```

### Option B: Build from Source

```bash
# Clone the repository
git clone https://github.com/yaacov/kubectl-mtv.git
cd kubectl-mtv/mcp-go

# Build and install
make build
make install
```

## Setup & Usage

For detailed setup instructions with MCP clients like Cursor and Claude Desktop, see **[MCP_SETUP.md](MCP_SETUP.md)**.

### Quick Start with Claude Code

```bash
# kubectl-mtv-mcp: read-only operations (recommended for most users)
claude mcp add kubectl-mtv-read kubectl-mtv-mcp

# kubectl-mtv-write-mcp: USE WITH CAUTION - can modify/delete resources
claude mcp add kubectl-mtv-write kubectl-mtv-write-mcp

# virtctl-mcp: KubeVirt VM management
claude mcp add kubevirt virtctl-mcp
```

## Servers Overview

### kubectl-mtv-mcp (6 tools)

**Safe for all users** - Read-only operations:

| Tool | Description |
|------|-------------|
| `ListResources` | List MTV resources (providers, plans, mappings, hosts, hooks) |
| `ListInventory` | Query provider inventories with SQL-like syntax |
| `GetLogs` | Retrieve logs from MTV controller and importer pods |
| `GetMigrationStorage` | Access migration PVCs and DataVolumes |
| `GetVersion` | Get kubectl-mtv and MTV operator version information |
| `GetPlanVms` | Get VMs and their status from migration plans |

### kubectl-mtv-write-mcp (13 tools)

**USE WITH CAUTION** - Can modify/delete resources:

| Tool | Description |
|------|-------------|
| `ManagePlanLifecycle` | Start, cancel, cutover, archive, unarchive plans |
| `CreateProvider` | Create source virtualization provider connections |
| `ManageMapping` | Create, delete, patch network and storage mappings |
| `CreatePlan` | Create comprehensive migration plans |
| `CreateHost` | Create migration hosts for direct data transfer |
| `CreateHook` | Create migration hooks for custom automation |
| `DeleteProvider` | Delete one or more providers |
| `DeletePlan` | Delete one or more migration plans |
| `DeleteHost` | Delete one or more migration hosts |
| `DeleteHook` | Delete one or more migration hooks |
| `PatchProvider` | Patch/modify existing providers |
| `PatchPlan` | Patch/modify migration plans |
| `PatchPlanVm` | Patch VM-specific fields in plans |

### virtctl-mcp (3 tools)

Virtual machine management:

| Tool | Description |
|------|-------------|
| `virtctl_vm_lifecycle` | VM power state management (start, stop, restart, pause, migrate) |
| `virtctl_diagnostics` | VM diagnostics (guest OS info, filesystems, users) |
| `virtctl_cluster_resources` | Discover available resources (instancetypes, preferences, datasources) |

## Building

### Quick Build

```bash
# Build all servers for current platform
make build

# Build individual servers
make build-kubectl-mtv-mcp
make build-kubectl-mtv-write-mcp
make build-virtctl-mcp
```

### Multi-Architecture Builds

```bash
# Build for specific platforms
make build-linux-amd64      # Linux AMD64
make build-linux-arm64      # Linux ARM64
make build-darwin-amd64     # macOS Intel
make build-darwin-arm64     # macOS Apple Silicon
make build-windows-amd64    # Windows AMD64

# Build for all platforms
make build-all
```

### Code Quality

```bash
# Install development tools (first time)
make install-tools

# Format code
make fmt

# Run linters
make lint

# Run all checks
make verify
```

### Distribution

```bash
# Create release archives with checksums
make dist-all

# Create local distribution
make dist
```

## Development

### Running Locally

```bash
# Run individual servers
make run-kubectl-mtv-mcp          # Start kubectl-mtv-mcp (read-server)
make run-kubectl-mtv-write-mcp    # Start kubectl-mtv-write-mcp (write-server)
make run-virtctl-mcp              # Start virtctl-mcp (virtctl-server)
```

### Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage
```

### Build Optimizations

The Makefile includes production-ready optimizations:

- **Binary size**: 5.2MB (33% smaller than unoptimized)
- **Stripped symbols**: Debug info removed for smaller binaries
- **Version embedding**: Git tags automatically embedded
- **Reproducible builds**: Consistent flags across all platforms
- **Silent output**: Minimal output following Unix philosophy

### Installation

```bash
# Install to GOPATH/bin
make install

# Clean build artifacts
make clean

# Deep clean (includes Go cache)
make clean-all
```

## MCP Client Configuration

### Cursor IDE

Add to your Cursor settings:

```json
{
  "mcpServers": {
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

### Claude Desktop

Edit `~/.config/claude/claude_desktop_config.json` (Linux) or `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS):

```json
{
  "mcpServers": {
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

See [MCP_SETUP.md](MCP_SETUP.md) for complete configuration instructions.

## Security Considerations

- The MCP servers execute kubectl-mtv and virtctl commands with your current Kubernetes permissions
- **kubectl-mtv-mcp**: Safe to use - only performs read operations
- **kubectl-mtv-write-mcp**: **USE WITH CAUTION** - can create, modify, and delete resources
- Ensure your MCP client is running in a secure environment
- Consider using dedicated service accounts for production environments
- By default, servers use stdio transport (no network exposure).
- When started with --sse, an HTTP server is exposed on the configured host:port.
  Prefer binding to 127.0.0.1, and restrict access via firewall when necessary.

## Available Platforms

Pre-built binaries available for:

- Linux AMD64
- Linux ARM64
- macOS Intel (AMD64)
- macOS Apple Silicon (ARM64)
- Windows AMD64

## Troubleshooting

### Server Not Found

```bash
# If binaries are in PATH, find their locations
which kubectl-mtv-mcp kubectl-mtv-write-mcp virtctl-mcp

# Check common installation locations
ls -l ~/.local/bin/kubectl-mtv-* ~/.local/bin/virtctl-mcp
ls -l /usr/local/bin/kubectl-mtv-* /usr/local/bin/virtctl-mcp

# Search your system for the binaries
find ~ -name "kubectl-mtv-mcp" -type f 2>/dev/null

# Add to PATH if needed
export PATH="$HOME/.local/bin:$PATH"

# Or use absolute paths in configuration
```

### Permission Issues

```bash
# Make binaries executable
chmod +x ~/.local/bin/kubectl-mtv-mcp
chmod +x ~/.local/bin/kubectl-mtv-write-mcp
chmod +x ~/.local/bin/virtctl-mcp
```

### Testing Server Connectivity

```bash
# Each server should start and wait for input
kubectl-mtv-mcp
# Press Ctrl+C to exit
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run `make verify` to check code quality
5. Submit a pull request

## License

Apache License 2.0 - See the main kubectl-mtv project for details.
