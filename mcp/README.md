# kubectl-mtv MCP Server

A Model Context Protocol (MCP) server that provides tools to interact with Migration Toolkit for Virtualization (MTV) through kubectl-mtv commands.

For AI assistants and chat applications: This MCP server provides access to the official kubectl-mtv documentation as MCP resources, giving full context about MTV workflows and tool usage.

## Prerequisites

1. **kubectl-mtv installed**: The `kubectl-mtv` binary must be available in your PATH
2. **Kubernetes cluster access**: You must be logged into a Kubernetes cluster with MTV deployed
3. **Python 3.8+**: Required to run the MCP server

## Features

This MCP server provides the following capabilities:

### MTV Resource Management
- **list_providers**: List all MTV providers in the cluster
- **list_plans**: List all MTV migration plans
- **list_mappings**: List all MTV mappings (network and storage)
- **list_hosts**: List all MTV migration hosts
- **list_hooks**: List all MTV migration hooks

### Provider Inventory Resources
- **list_inventory_vms**: List VMs from a provider's inventory
- **list_inventory_networks**: List networks from a provider's inventory
- **list_inventory_storage**: List storage from a provider's inventory
- **list_inventory_hosts**: List hosts from a provider's inventory
- **list_inventory_clusters**: List clusters from a provider's inventory (oVirt, vSphere)
- **list_inventory_datacenters**: List datacenters from a provider's inventory (oVirt, vSphere)
- **list_inventory_generic**: List any inventory resource type (advanced users)

## Quick Start

For detailed setup instructions with MCP clients like Cursor and Claude Desktop, see **[MCP_SETUP.md](MCP_SETUP.md)**.

### Manual Installation

1. Install Python dependencies:
```bash
cd mcp
pip install -r requirements.txt
```

2. Generate configuration files:
```bash
python generate_config.py
```

3. Configure your MCP client using the generated configuration files.

## Usage

### Running the Server

```bash
python kubectl_mtv_server.py
```

### Configuration for MCP Clients

Add the following to your MCP client configuration:

```json
{
  "servers": {
    "kubectl-mtv": {
      "command": "python",
      "args": ["/path/to/mcp/kubectl_mtv_server.py"],
      "env": {}
    }
  }
}
```
