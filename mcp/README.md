# kubectl-mtv MCP Servers

Model Context Protocol (MCP) servers that provide tools to interact with Migration Toolkit for Virtualization (MTV) through kubectl-mtv commands.

For AI assistants and chat applications: These MCP servers provide access to the official kubectl-mtv documentation as MCP resources, giving full context about MTV workflows and tool usage.

## MCP servers

### Read-Only Server (`kubectl_mtv_server.py`)
**For most use cases** - Provides safe, read-only operations for monitoring and troubleshooting.

### Write Operations Server (`kubectl_mtv_write_server.py`) 
**USE WITH CAUTION** - Provides destructive operations that can modify, create, or delete MTV resources.

## Prerequisites

1. **kubectl-mtv installed**: The `kubectl-mtv` binary must be available in your PATH
2. **Kubernetes cluster access**: You must be logged into a Kubernetes cluster with MTV deployed
3. **Python 3.8+**: Required to run the MCP servers

## Features

### Read-Only Server (kubectl_mtv_server.py)

**MTV Resource Discovery & Monitoring:**
- **list_providers**: List all MTV providers in the cluster
- **list_plans**: List all MTV migration plans
- **list_mappings**: List all MTV mappings (network and storage)
- **list_hosts**: List all MTV migration hosts
- **list_hooks**: List all MTV migration hooks
- **get_plan_vms**: Get VMs and their status from a specific migration plan

**Provider Inventory Resources:**
- **list_inventory_vms**: List VMs from a provider's inventory
- **list_inventory_networks**: List networks from a provider's inventory
- **list_inventory_storage**: List storage from a provider's inventory
- **list_inventory_hosts**: List hosts from a provider's inventory
- **list_inventory_clusters**: List clusters from a provider's inventory (oVirt, vSphere)
- **list_inventory_datacenters**: List datacenters from a provider's inventory (oVirt, vSphere)
- **list_inventory_generic**: List any inventory resource type (advanced users)

**Detailed Resource Information:**
- **describe_plan**: Get detailed information about a migration plan
- **describe_vm**: Get detailed status of a specific VM in a migration plan
- **describe_host**: Get detailed information about a migration host
- **describe_mapping**: Get detailed information about a network or storage mapping
- **describe_hook**: Get detailed information about a migration hook

**System Information & Debugging:**
- **get_version**: Get kubectl-mtv and MTV operator version information
- **get_controller_logs**: Get logs from the MTV controller pod for debugging
- **get_migration_pvcs**: Get PersistentVolumeClaims related to VM migrations
- **get_migration_datavolumes**: Get DataVolumes related to VM migrations
- **get_migration_storage**: Get all storage resources related to VM migrations

### Write Operations Server (kubectl_mtv_write_server.py)

**WARNING: These operations can modify or delete MTV resources**

**Plan Lifecycle Management:**
- **start_plan**: Start a migration plan to begin migrating VMs
- **cancel_plan**: Cancel a running migration plan
- **cutover_plan**: Perform cutover for a migration plan

**Resource Creation:**
- **create_provider**: Create a new provider for connecting to source platforms
- **create_mapping**: Create a new network or storage mapping
- **create_plan**: Create a new migration plan
- **create_host**: Create a new migration host
- **create_hook**: Create a new migration hook

**Resource Deletion:**
- **delete_provider**: Delete a provider
- **delete_mapping**: Delete a network or storage mapping
- **delete_plan**: Delete a migration plan
- **delete_host**: Delete a migration host
- **delete_hook**: Delete a migration hook

**Resource Modification:**
- **patch_provider**: Patch/modify an existing provider
- **patch_mapping**: Patch/modify an existing mapping
- **patch_plan**: Patch/modify an existing migration plan (includes archiving/unarchiving)

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

### Running the Servers

**Read-Only Server:**
```bash
python kubectl_mtv_server.py
```

**Write Operations Server:**
```bash
python kubectl_mtv_write_server.py
```

### Configuration for MCP Clients

You can configure either or both servers depending on your needs:

#### Read-Only Server Only
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

#### Both Servers (Full Functionality)
```json
{
  "servers": {
    "kubectl-mtv-read": {
      "command": "python", 
      "args": ["/path/to/mcp/kubectl_mtv_server.py"],
      "env": {}
    },
    "kubectl-mtv-write": {
      "command": "python",
      "args": ["/path/to/mcp/kubectl_mtv_write_server.py"], 
      "env": {}
    }
  }
}
```
