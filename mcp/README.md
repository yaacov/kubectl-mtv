# kubectl-mtv MCP Servers

Model Context Protocol (MCP) servers that provide AI assistants with tools to interact with Migration Toolkit for Virtualization (MTV) through kubectl-mtv commands.

## What it does

These MCP servers enable AI assistants to help with MTV operations by providing:

- **Read-Only Server** (`kubectl_mtv_server.py`): Safe operations for monitoring, troubleshooting, and discovering MTV resources and provider inventories
- **Write Server** (`kubectl_mtv_write_server.py`): **USE WITH CAUTION** - Full lifecycle management including creating, modifying, and deleting MTV resources

## Prerequisites

- `kubectl-mtv` binary installed and available in PATH
- Access to a Kubernetes cluster with MTV deployed
- Python 3.8+

## Quick Installation

1. Install dependencies:
```bash
cd mcp
pip install -r requirements.txt
```

2. Generate MCP client configuration:
```bash
python generate_config.py
```

3. Configure your MCP client with the generated files

## Setup & Usage

For detailed setup instructions with MCP clients like Cursor and Claude Desktop, see **[MCP_SETUP.md](MCP_SETUP.md)**.

### Claude MCP Install (Easiest)

If you're using Claude Code, you can install directly:

```bash
# Read-only server (recommended for most users)
claude mcp add python /full/path/to/kubectl-mtv/mcp/kubev2v/kubectl_mtv_server.py

# Write server (USE WITH CAUTION - can modify/delete resources)
claude mcp add python /full/path/to/kubectl-mtv/mcp/kubev2v/kubectl_mtv_write_server.py
```

## Development

### Important Make Targets

```bash
# Install dependencies and setup
make install           # Install for production use
make install-dev       # Install for development (includes dev tools)

# Run the MCP servers
make run-read          # Start read-only server
make run-write         # Start write server (USE WITH CAUTION)

# Code quality and formatting
make lint              # Run flake8 linting
make format            # Format code with black
make format-check      # Check if code is formatted correctly

# Configuration and cleanup
make config            # Generate MCP client configuration files
make clean             # Clean build artifacts
make clean-all         # Clean everything including generated configs
```
