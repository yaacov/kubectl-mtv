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

### Install from PyPI:
```bash
pip install mtv-mcp
```

3. Configure your MCP client with the generated files

## Setup & Usage

For detailed setup instructions with MCP clients like Cursor and Claude Desktop, see **[MCP_SETUP.md](MCP_SETUP.md)**.

### Claude MCP Install (Easiest)

If you're using Claude Code, you can install directly:

```bash
# Read-only server (recommended for most users)
claude mcp add kubectl-mtv-mcp

# Write server (USE WITH CAUTION - can modify/delete resources)
claude mcp add kubectl-mtv-write-mcp
```

## Development

### Development Installation

For development work, install from source:

```bash
cd mcp
pip install -r requirements.txt
pip install -r requirements-dev.txt
```

### Manual Server Configuration

For development or when you need direct control, you can use the servers directly:

```bash
# Read-only server (recommended for most users)
claude mcp add python /full/path/to/kubectl-mtv/mcp/kubev2v/kubectl_mtv_server.py

# Write server (USE WITH CAUTION - can modify/delete resources)
claude mcp add python /full/path/to/kubectl-mtv/mcp/kubev2v/kubectl_mtv_write_server.py
```

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

# Build and publishing
make build             # Build both source distribution and wheel using python -m build
make build-sdist       # Build source distribution only
make build-wheel       # Build wheel only
make upload-test       # Upload to TestPyPI
make upload            # Upload to PyPI

# Configuration and cleanup
make config            # Generate MCP client configuration files
make clean             # Clean build artifacts
```

### Building and Publishing

To build and upload the package to PyPI:

```bash
# Build the package
python -m build

# Upload to PyPI
twine upload dist/*
```
