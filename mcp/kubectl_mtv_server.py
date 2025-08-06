#!/usr/bin/env python3
"""
FastMCP Server for kubectl-mtv

This server provides tools to interact with Migration Toolkit for Virtualization (MTV)
through kubectl-mtv commands. It assumes kubectl-mtv is installed and the user is 
logged into a Kubernetes cluster with MTV deployed.

MTV Context:
- MTV helps migrate VMs from other KubeVirt clusters, vSphere, oVirt, OpenStack, and OVA files to KubeVirt
- Typical workflow: Provider -> Inventory Discovery -> Mappings -> Plans -> Migration
- These tools provide read-only discovery and monitoring capabilities
- See MTV_CONTEXT.md for detailed guidance on tool usage patterns

Tool Categories:
- Provider Management: list_providers
- Migration Planning: list_plans, list_mappings, list_hosts, list_hooks  
- Inventory Discovery: list_inventory_* tools for exploring source environments
- Debugging: get_controller_logs for troubleshooting MTV controller issues
- Storage Debugging: get_migration_pvcs, get_migration_datavolumes, get_migration_storage for tracking VM migration storage
"""

import os
import subprocess
import json
from typing import Any, Optional

from fastmcp import FastMCP

# Get the directory containing this script and the docs directory
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
DOCS_DIR = os.path.join(os.path.dirname(SCRIPT_DIR), "docs")

# Initialize the FastMCP server
mcp = FastMCP("kubectl-mtv")


class KubectlMTVError(Exception):
    """Custom exception for kubectl-mtv command errors."""
    pass


async def run_kubectl_mtv_command(args: list[str]) -> str:
    """Run a kubectl-mtv command and return the output."""
    try:
        cmd = ["kubectl-mtv"] + args
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout
    except subprocess.CalledProcessError as e:
        error_msg = f"Command failed: {' '.join(cmd)}\nError: {e.stderr}"
        raise KubectlMTVError(error_msg) from e
    except FileNotFoundError:
        raise KubectlMTVError("kubectl-mtv not found in PATH") from None


async def run_kubectl_command(args: list[str]) -> str:
    """Run a kubectl command and return the output."""
    try:
        cmd = ["kubectl"] + args
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout
    except subprocess.CalledProcessError as e:
        error_msg = f"Command failed: {' '.join(cmd)}\nError: {e.stderr}"
        raise KubectlMTVError(error_msg) from e
    except FileNotFoundError:
        raise KubectlMTVError("kubectl not found in PATH") from None


async def get_mtv_operator_namespace() -> str:
    """Get the MTV operator namespace from kubectl-mtv version."""
    try:
        version_output = await run_kubectl_mtv_command(["version", "-o", "json"])
        version_data = json.loads(version_output)
        namespace = version_data.get("operatorNamespace")
        if not namespace:
            raise KubectlMTVError("operatorNamespace not found in kubectl-mtv version output")
        return namespace
    except json.JSONDecodeError as e:
        raise KubectlMTVError(f"Failed to parse kubectl-mtv version JSON: {e}") from e


async def find_controller_pod(namespace: str) -> str:
    """Find the forklift-controller pod in the given namespace."""
    try:
        # Get pods in the namespace and look for forklift-controller
        pods_output = await run_kubectl_command([
            "get", "pods", "-n", namespace, 
            "-o", "json"
        ])
        pods_data = json.loads(pods_output)
        
        controller_pods = []
        for pod in pods_data.get("items", []):
            pod_name = pod.get("metadata", {}).get("name", "")
            if pod_name.startswith("forklift-controller-"):
                # Check if pod is running
                phase = pod.get("status", {}).get("phase", "")
                if phase == "Running":
                    controller_pods.append(pod_name)
        
        if not controller_pods:
            raise KubectlMTVError(f"No running forklift-controller pods found in namespace {namespace}")
        
        # Return the first running controller pod
        return controller_pods[0]
        
    except json.JSONDecodeError as e:
        raise KubectlMTVError(f"Failed to parse kubectl pods JSON: {e}") from e


async def _build_base_args(arguments: dict[str, Any]) -> list[str]:
    """Build base arguments for kubectl-mtv commands."""
    args = []
    
    if arguments.get("all_namespaces", False):
        args.extend(["-A"])
    elif "namespace" in arguments and arguments["namespace"]:
        args.extend(["-n", arguments["namespace"]])
    
    output_format = arguments.get("output_format", "json")
    if output_format != "table":
        args.extend(["-o", output_format])
    
    return args


async def _list_inventory(arguments: dict[str, Any], resource_type: str) -> str:
    """List inventory resources from a provider."""
    provider_name = arguments["provider_name"]
    args = ["get", "inventory", resource_type, provider_name]
    
    if "namespace" in arguments and arguments["namespace"]:
        args.extend(["-n", arguments["namespace"]])
    
    if "query" in arguments and arguments["query"]:
        args.extend(["-q", arguments["query"]])
    
    output_format = arguments.get("output_format", "json")
    if output_format != "table":
        args.extend(["-o", output_format])
    
    return await run_kubectl_mtv_command(args)


async def _list_inventory_generic(arguments: dict[str, Any]) -> str:
    """List any inventory resource type from a provider."""
    resource_type = arguments["resource_type"]
    provider_name = arguments["provider_name"]
    args = ["get", "inventory", resource_type, provider_name]
    
    if "namespace" in arguments and arguments["namespace"]:
        args.extend(["-n", arguments["namespace"]])
    
    if "query" in arguments and arguments["query"]:
        args.extend(["-q", arguments["query"]])
    
    output_format = arguments.get("output_format", "json")
    if output_format != "table":
        args.extend(["-o", output_format])
    
    return await run_kubectl_mtv_command(args)


# Resources for documentation
@mcp.resource("kubectl-mtv://usage")
async def usage_guide() -> str:
    """Comprehensive kubectl-mtv usage documentation covering all commands, providers, plans, mappings, and workflows."""
    try:
        file_path = os.path.join(DOCS_DIR, "README-usage.md")
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    except FileNotFoundError:
        return "kubectl-mtv usage documentation file not found"
    except Exception as e:
        return f"Error reading kubectl-mtv usage documentation: {str(e)}"


@mcp.resource("kubectl-mtv://inventory")
async def inventory_guide() -> str:
    """Complete guide to inventory discovery, query language, and filtering resources from providers."""
    try:
        file_path = os.path.join(DOCS_DIR, "README_inventory.md")
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    except FileNotFoundError:
        return "kubectl-mtv inventory documentation file not found"
    except Exception as e:
        return f"Error reading kubectl-mtv inventory documentation: {str(e)}"


@mcp.resource("kubectl-mtv://hooks")
async def hooks_guide() -> str:
    """Guide to migration hooks for customizing and automating migration workflows."""
    try:
        file_path = os.path.join(DOCS_DIR, "README_hooks.md")
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    except FileNotFoundError:
        return "kubectl-mtv hooks documentation file not found"
    except Exception as e:
        return f"Error reading kubectl-mtv hooks documentation: {str(e)}"


@mcp.resource("kubectl-mtv://mapping-pairs")
async def mapping_pairs_guide() -> str:
    """Guide to network and storage mapping configuration using pairs syntax."""
    try:
        file_path = os.path.join(DOCS_DIR, "README_mapping_pairs.md")
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    except FileNotFoundError:
        return "kubectl-mtv mapping pairs documentation file not found"
    except Exception as e:
        return f"Error reading kubectl-mtv mapping pairs documentation: {str(e)}"


@mcp.resource("kubectl-mtv://demo")
async def demo_guide() -> str:
    """Step-by-step tutorial demonstrating a complete migration workflow."""
    try:
        file_path = os.path.join(DOCS_DIR, "README_demo.md")
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    except FileNotFoundError:
        return "kubectl-mtv demo documentation file not found"
    except Exception as e:
        return f"Error reading kubectl-mtv demo documentation: {str(e)}"


# Tool functions using FastMCP
@mcp.tool()
async def list_providers(
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json"
) -> str:
    """List all MTV providers in the cluster.
    
    Providers connect MTV to source virtualization platforms (vSphere, oVirt, OpenStack, OVA, other KubeVirt clusters). 
    This is typically the first step in understanding your MTV setup.
    
    Args:
        namespace: Kubernetes namespace to query (optional, defaults to current namespace)
        all_namespaces: List providers across all namespaces
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted provider information or table output
    """
    args = ["get", "provider"] + await _build_base_args({
        "namespace": namespace, 
        "all_namespaces": all_namespaces, 
        "output_format": output_format
    })
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def list_plans(
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json",
    watch: bool = False
) -> str:
    """List all MTV migration plans in the cluster.
    
    Plans define which VMs to migrate and track migration progress. 
    Use this to monitor active migrations or see migration history.
    
    Args:
        namespace: Kubernetes namespace to query (optional)
        all_namespaces: List plans across all namespaces
        output_format: Output format (json, yaml, or table)
        watch: Watch for changes (not recommended for MCP)
        
    Returns:
        JSON/YAML formatted plan information or table output
    """
    args = ["get", "plan"] + await _build_base_args({
        "namespace": namespace,
        "all_namespaces": all_namespaces,
        "output_format": output_format
    })
    
    if watch:
        args.append("--watch")
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def list_mappings(
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json"
) -> str:
    """List all MTV mappings (network and storage) in the cluster.
    
    Args:
        namespace: Kubernetes namespace to query (optional)
        all_namespaces: List mappings across all namespaces
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted mapping information or table output
    """
    args = ["get", "mapping"] + await _build_base_args({
        "namespace": namespace,
        "all_namespaces": all_namespaces,
        "output_format": output_format
    })
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def list_hosts(
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json"
) -> str:
    """List all MTV migration hosts in the cluster.
    
    Args:
        namespace: Kubernetes namespace to query (optional)
        all_namespaces: List hosts across all namespaces
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted host information or table output
    """
    args = ["get", "host"] + await _build_base_args({
        "namespace": namespace,
        "all_namespaces": all_namespaces,
        "output_format": output_format
    })
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def list_hooks(
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json"
) -> str:
    """List all MTV migration hooks in the cluster.
    
    Args:
        namespace: Kubernetes namespace to query (optional)
        all_namespaces: List hooks across all namespaces  
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted hook information or table output
    """
    args = ["get", "hook"] + await _build_base_args({
        "namespace": namespace,
        "all_namespaces": all_namespaces,
        "output_format": output_format
    })
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def list_inventory_vms(
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List VMs from a specific provider's inventory.
    
    This shows all VMs available for migration from a source platform. 
    Use queries to filter by memory, CPU, power state, or name. Essential for migration planning.
    
    QUERY EXAMPLES:
    vSphere:
    - Basic filtering: "WHERE powerState = 'poweredOn'"
    - Memory filtering: "WHERE memoryMB > 4096"
    - Complex: "WHERE powerState = 'poweredOn' AND memoryMB > 4096 ORDER BY cpuCount DESC"
    - Field selection: "SELECT name, powerStateHuman AS state, memoryGB, cpuCount WHERE memoryMB > 2048"
    - Find concerns: "WHERE warningConcerns > 0 OR criticalConcerns > 0"
    - Guest OS: "WHERE guestName LIKE '%Windows%'"
    - Shared disks: "WHERE any(disks[*].shared = true)"
    
    oVirt:
    - Active VMs: "WHERE status = 'up'"
    - Memory filtering: "WHERE memory > 4294967296"  (bytes)
    - CPU filtering: "WHERE cpuCores > 2 AND cpuSockets >= 1"
    - Network info: "SELECT name, nics[*].mac, nics[*].ipAddress WHERE len(nics) > 0"
    
    OpenStack:
    - Active instances: "WHERE status = 'ACTIVE'"
    - By flavor: "WHERE flavorID = 'specific-flavor-id'"
    - Network addresses: "SELECT name, addresses WHERE len(addresses) > 0"
    - Recent VMs: "WHERE created > '2023-01-01' ORDER BY created DESC"
    
    COMMON VM FIELDS (varies by provider):
    vSphere: name, powerState, powerStateHuman, cpuCount, memoryMB, memoryGB, ipAddress, 
             guestName, host, uuid, disks[*].capacity, disks[*].shared, networks[*], 
             warningConcerns, criticalConcerns, concernsHuman, storageUsed, storageUsedGB
    oVirt: name, status, memory, cpuCores, cpuSockets, cpuThreads, guestName, host, 
           diskAttachments[*], nics[*].mac, nics[*].ipAddress, placementPolicyAffinity
    OpenStack: name, status, flavorID, imageID, hostID, addresses, created, updated, 
               attachedVolumes[*], securityGroups[*], metadata, keyName, tenantID
    
    Args:
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query using SQL-like syntax with WHERE/SELECT/ORDER BY/LIMIT
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted VM inventory or table output
    """
    return await _list_inventory({
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    }, "vm")


@mcp.tool()
async def list_inventory_networks(
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List networks from a specific provider's inventory.
    
    QUERY EXAMPLES:
    vSphere:
    - Name filtering: "WHERE name LIKE '%VM Network%'"
    - Network type: "WHERE variant = 'Standard'"
    - VLAN networks: "WHERE vlanId != ''"
    - Field selection: "SELECT name, variant, vlanId, path WHERE variant = 'Standard'"
    
    oVirt:
    - Management networks: "WHERE name LIKE '%mgmt%' OR description LIKE '%Management%'"
    - VM networks: "WHERE 'vm' IN usages"
    - VLAN networks: "WHERE vlan != ''"
    - NIC profiles: "SELECT name, nicProfiles, vlan WHERE len(nicProfiles) > 0"
    
    OpenStack:
    - Active networks: "WHERE status = 'ACTIVE'"
    - Shared networks: "WHERE shared = true"
    - By project: "WHERE projectID = 'specific-project-id'"
    - Recent networks: "WHERE createdAt > '2023-01-01' ORDER BY createdAt DESC"
    
    COMMON NETWORK FIELDS (varies by provider):
    vSphere: name, id, variant (Standard, DvSwitch), vlanId, hostCount, path, parent.id
    oVirt: name, id, dataCenter, vlan, usages[*], nicProfiles[*], description, hostCount
    OpenStack: name, id, status, shared, adminStateUp, projectID, tenantID, subnets[*], createdAt, updatedAt
    
    Args:
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query using SQL-like syntax with WHERE/SELECT/ORDER BY/LIMIT
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted network inventory or table output
    """
    return await _list_inventory({
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    }, "network")


@mcp.tool()
async def list_inventory_storage(
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List storage from a specific provider's inventory.
    
    QUERY EXAMPLES:
    vSphere:
    - Available space: "WHERE free > 500Gi ORDER BY free DESC"
    - Storage type: "WHERE type = 'NFS'"
    - Maintenance check: "WHERE maintenance != 'normal'"
    - Utilization: "SELECT name, capacityHuman, freeHuman, type WHERE (capacity - free) / capacity > 0.7"
    - Large datastores: "WHERE capacity > 1Ti ORDER BY capacity DESC"
    - SAN storage: "WHERE len(backingDevicesNames) > 0"
    
    oVirt:
    - Data domains: "WHERE type = 'data'"
    - Available storage: "WHERE free > 107374182400 ORDER BY free DESC"  # >100GB
    - Storage technology: "WHERE storage.type = 'nfs'"
    - By datacenter: "WHERE dataCenter = 'specific-datacenter-id'"
    - Export domains: "WHERE type = 'export'"
    
    OpenStack:
    - Available storage: "WHERE free > 107374182400 ORDER BY free DESC"
    - By type: "WHERE type = 'specific-storage-type'"
    
    COMMON STORAGE FIELDS (varies by provider):
    vSphere: name, type (NFS, VMFS), capacity, capacityHuman, free, freeHuman, maintenance, path, backingDevicesNames[*]
    oVirt: name, type (data, export, image), capacity, capacityHuman, free, freeHuman, dataCenter, storage.type (nfs, fcp, glance), path
    OpenStack: name, type, capacity, capacityHuman, free, freeHuman
    
    Args:
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query using SQL-like syntax with WHERE/SELECT/ORDER BY/LIMIT
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted storage inventory or table output
    """
    return await _list_inventory({
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    }, "storage")


@mcp.tool()
async def list_inventory_hosts(
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List hosts from a specific provider's inventory.
    
    QUERY EXAMPLES:
    vSphere:
    - Active hosts: "WHERE inMaintenance = false"
    - High-capacity hosts: "WHERE cpuCores > 16 AND cpuSockets >= 2"
    - Network speed: "WHERE any(networkAdapters[*].linkSpeed >= 10000)"
    - Storage connectivity: "SELECT name, cpuCores, len(datastores) AS datastoreCount WHERE len(datastores) > 5"
    - Physical NICs: "SELECT name, networking.pNICs[*].linkSpeed WHERE len(networking.pNICs) > 4"
    
    oVirt:
    - By cluster: "WHERE cluster = 'specific-cluster-id'"
    - Resource analysis: "SELECT name, cpuCores, cpuSockets, memory WHERE cpuCores > 8"
    - Status check: "WHERE status != 'up'"
    
    COMMON HOST FIELDS:
    - name, inMaintenance, cpuCores, cpuSockets, managementServerIp, cluster
    - networkAdapters[*].ipAddress, networkAdapters[*].linkSpeed, networkAdapters[*].mtu
    - datastores[*].id, networking.pNICs[*].linkSpeed, hostScsiDisks[*]
    
    Args:
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query using SQL-like syntax with WHERE/SELECT/ORDER BY/LIMIT
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted host inventory or table output
    """
    return await _list_inventory({
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    }, "host")


@mcp.tool()
async def list_inventory_clusters(
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List clusters from a specific provider's inventory (oVirt, vSphere).
    
    QUERY EXAMPLES:
    vSphere:
    - High availability clusters: "WHERE haEnabled = true"
    - DRS enabled clusters: "WHERE drsEnabled = true"
    - Resource pools: "SELECT name, totalCpuMhz, totalMemoryMB WHERE totalCpuMhz > 50000"
    - By datacenter: "WHERE dataCenter = 'specific-datacenter-id'"
    
    oVirt:
    - Active clusters: "WHERE status = 'up'"
    - CPU architecture: "WHERE architecture = 'x86_64'"
    - Cluster features: "SELECT name, cpuArchitecture, version, haReservation WHERE haReservation = true"
    - By datacenter: "WHERE dataCenter = 'specific-datacenter-id'"
    
    COMMON CLUSTER FIELDS (varies by provider):
    vSphere: name, id, dataCenter, haEnabled, drsEnabled, totalCpuMhz, totalMemoryMB, numHosts
    oVirt: name, id, dataCenter, status, architecture, version, haReservation, cpuArchitecture
    
    Args:
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query using SQL-like syntax with WHERE/SELECT/ORDER BY/LIMIT
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted cluster inventory or table output
    """
    return await _list_inventory({
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    }, "cluster")


@mcp.tool()
async def list_inventory_datacenters(
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List datacenters from a specific provider's inventory (oVirt, vSphere).
    
    QUERY EXAMPLES:
    vSphere:
    - All datacenters: "SELECT name, id, path ORDER BY name"
    - By location: "WHERE name LIKE '%US%' OR name LIKE '%East%'"
    - Resource summary: "SELECT name, totalClusters, totalHosts WHERE totalHosts > 10"
    
    oVirt:
    - Active datacenters: "WHERE status = 'up'"
    - By version: "WHERE version LIKE '4.%'"
    - Storage analysis: "SELECT name, storageFormat, quotaMode WHERE quotaMode = 'enabled'"
    - MAC pools: "SELECT name, macPoolRanges WHERE len(macPoolRanges) > 0"
    
    COMMON DATACENTER FIELDS (varies by provider):
    vSphere: name, id, path, totalClusters, totalHosts, totalVMs
    oVirt: name, id, status, version, storageFormat, quotaMode, macPoolRanges[*]
    
    Args:
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query using SQL-like syntax with WHERE/SELECT/ORDER BY/LIMIT
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted datacenter inventory or table output
    """
    return await _list_inventory({
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    }, "datacenter")


@mcp.tool()
async def list_inventory_generic(
    resource_type: str,
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List any inventory resource type from a provider (advanced users).
    
    This is the most flexible inventory tool - use it to access any resource type not covered 
    by the specific tools above. All the same query syntax applies.
    
    AVAILABLE RESOURCE TYPES BY PROVIDER:
    vSphere: vm, network, storage, host, cluster, datacenter, datastore, folder, resource-pool
    oVirt: vm, network, storage, host, cluster, datacenter, disk, disk-profile, nic-profile
    OpenStack: vm, network, storage, flavor, image, instance, project
    OpenShift: vm, network, storage, namespace, pvc, data-volume
    
    QUERY SYNTAX:
    All inventory tools support SQL-like queries with:
    - SELECT field1, field2 AS alias, function(field3) AS name
    - WHERE condition (using Tree Search Language - TSL)
    - ORDER BY field1 [ASC|DESC], field2
    - LIMIT n
    
    OPERATORS: =, !=, <, <=, >, >=, LIKE, ILIKE, ~= (regex), IN, BETWEEN, AND, OR, NOT
    FUNCTIONS: sum(), len(), any(), all()
    LITERALS: strings ('text'), numbers (1024, 2.5Gi), dates ('2023-01-01'), booleans (true/false)
    
    EXAMPLE QUERIES FOR SPECIALIZED RESOURCES:
    Folders (vSphere): "WHERE name LIKE '%VM%' AND type = 'vm'"
    Flavors (OpenStack): "WHERE vcpus >= 4 AND ram >= 8192 ORDER BY vcpus DESC"
    Disk Profiles (oVirt): "WHERE storageDomain = 'specific-domain-id'"
    NIC Profiles (oVirt): "WHERE networkFilter != ''"
    Projects (OpenStack): "WHERE enabled = true ORDER BY name"
    Resource Pools (vSphere): "WHERE cpuAllocation.limit > 0"
    
    Args:
        resource_type: Type of inventory resource to list
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query using SQL-like syntax with WHERE/SELECT/ORDER BY/LIMIT
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted inventory or table output
    """
    return await _list_inventory_generic({
        "resource_type": resource_type,
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    })


@mcp.tool()
async def get_controller_logs(
    container: str = "main",
    lines: int = 100,
    follow: bool = False,
    namespace: str = ""
) -> str:
    """Get logs from the MTV controller pod for debugging.
    
    This tool automatically finds the MTV controller namespace using kubectl-mtv version,
    locates the forklift-controller pod, and retrieves logs from the specified container.
    
    Args:
        container: Container name to get logs from (main, inventory). Defaults to "main"
        lines: Number of recent log lines to retrieve. Defaults to 100
        follow: Follow log output (stream logs). Not recommended for MCP usage
        namespace: Override the MTV operator namespace (optional, auto-detected if not provided)
        
    Returns:
        Controller pod logs from the specified container
    """
    try:
        # Get the MTV operator namespace if not provided
        if not namespace:
            namespace = await get_mtv_operator_namespace()
        
        # Find the controller pod
        pod_name = await find_controller_pod(namespace)
        
        # Build kubectl logs command
        logs_args = ["logs", "-n", namespace, pod_name, "-c", container, "--tail", str(lines)]
        
        if follow:
            logs_args.append("-f")
        
        # Get the logs
        return await run_kubectl_command(logs_args)
        
    except Exception as e:
        return f"Error retrieving controller logs: {str(e)}"


@mcp.tool()
async def get_migration_pvcs(
    migration_id: str = "",
    plan_id: str = "",
    vm_id: str = "",
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json"
) -> str:
    """Get PersistentVolumeClaims related to VM migrations.
    
    Find PVCs that are part of VM migrations by searching for specific labels:
    - migration: Migration UUID
    - plan: Plan UUID  
    - vmID: VM identifier (e.g., vm-45)
    
    Args:
        migration_id: Migration UUID to filter by (optional)
        plan_id: Plan UUID to filter by (optional)
        vm_id: VM ID to filter by (optional)
        namespace: Kubernetes namespace to search in (optional)
        all_namespaces: Search across all namespaces
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted PVC information or table output
    """
    try:
        # Build label selector
        label_selectors = []
        if migration_id:
            label_selectors.append(f"migration={migration_id}")
        if plan_id:
            label_selectors.append(f"plan={plan_id}")
        if vm_id:
            label_selectors.append(f"vmID={vm_id}")
        
        # Build kubectl command
        cmd_args = ["get", "pvc"]
        
        if all_namespaces:
            cmd_args.append("-A")
        elif namespace:
            cmd_args.extend(["-n", namespace])
        
        if label_selectors:
            cmd_args.extend(["-l", ",".join(label_selectors)])
        
        if output_format != "table":
            cmd_args.extend(["-o", output_format])
        
        return await run_kubectl_command(cmd_args)
        
    except Exception as e:
        return f"Error retrieving migration PVCs: {str(e)}"


@mcp.tool()
async def get_migration_datavolumes(
    migration_id: str = "",
    plan_id: str = "",
    vm_id: str = "",
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json"
) -> str:
    """Get DataVolumes related to VM migrations.
    
    Find DataVolumes that are part of VM migrations by searching for specific labels:
    - migration: Migration UUID
    - plan: Plan UUID  
    - vmID: VM identifier (e.g., vm-45)
    
    Args:
        migration_id: Migration UUID to filter by (optional)
        plan_id: Plan UUID to filter by (optional)
        vm_id: VM ID to filter by (optional)
        namespace: Kubernetes namespace to search in (optional)
        all_namespaces: Search across all namespaces
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted DataVolume information or table output
    """
    try:
        # Build label selector
        label_selectors = []
        if migration_id:
            label_selectors.append(f"migration={migration_id}")
        if plan_id:
            label_selectors.append(f"plan={plan_id}")
        if vm_id:
            label_selectors.append(f"vmID={vm_id}")
        
        # Build kubectl command
        cmd_args = ["get", "datavolumes"]
        
        if all_namespaces:
            cmd_args.append("-A")
        elif namespace:
            cmd_args.extend(["-n", namespace])
        
        if label_selectors:
            cmd_args.extend(["-l", ",".join(label_selectors)])
        
        if output_format != "table":
            cmd_args.extend(["-o", output_format])
        
        return await run_kubectl_command(cmd_args)
        
    except Exception as e:
        return f"Error retrieving migration DataVolumes: {str(e)}"


@mcp.tool()
async def get_migration_storage(
    migration_id: str = "",
    plan_id: str = "",
    vm_id: str = "",
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json"
) -> str:
    """Get all storage resources (PVCs and DataVolumes) related to VM migrations.
    
    Find both PVCs and DataVolumes that are part of VM migrations by searching for specific labels:
    - migration: Migration UUID
    - plan: Plan UUID  
    - vmID: VM identifier (e.g., vm-45)
    
    This is a convenience tool that combines results from both PVCs and DataVolumes.
    
    Args:
        migration_id: Migration UUID to filter by (optional)
        plan_id: Plan UUID to filter by (optional)
        vm_id: VM ID to filter by (optional)
        namespace: Kubernetes namespace to search in (optional)
        all_namespaces: Search across all namespaces
        output_format: Output format (json, yaml, or table)
        
    Returns:
        Combined JSON/YAML formatted storage information or table output
    """
    try:
        # Get both PVCs and DataVolumes
        pvcs_result = await get_migration_pvcs(
            migration_id=migration_id,
            plan_id=plan_id,
            vm_id=vm_id,
            namespace=namespace,
            all_namespaces=all_namespaces,
            output_format=output_format
        )
        
        dvs_result = await get_migration_datavolumes(
            migration_id=migration_id,
            plan_id=plan_id,
            vm_id=vm_id,
            namespace=namespace,
            all_namespaces=all_namespaces,
            output_format=output_format
        )
        
        if output_format == "table":
            return f"=== PersistentVolumeClaims ===\n{pvcs_result}\n\n=== DataVolumes ===\n{dvs_result}"
        elif output_format == "json":
            # Try to combine JSON results
            try:
                import json
                pvcs_data = json.loads(pvcs_result) if pvcs_result.strip() else {"items": []}
                dvs_data = json.loads(dvs_result) if dvs_result.strip() else {"items": []}
                
                combined = {
                    "pvcs": pvcs_data,
                    "datavolumes": dvs_data
                }
                return json.dumps(combined, indent=2)
            except json.JSONDecodeError:
                return f"PVCs:\n{pvcs_result}\n\nDataVolumes:\n{dvs_result}"
        else:
            return f"PVCs:\n{pvcs_result}\n\nDataVolumes:\n{dvs_result}"
        
    except Exception as e:
        return f"Error retrieving migration storage resources: {str(e)}"


@mcp.tool()
async def get_version(
    output_format: str = "json"
) -> str:
    """Get kubectl-mtv and MTV operator version information.
    
    This tool provides comprehensive version information including:
    - kubectl-mtv client version
    - MTV operator version and status
    - MTV operator namespace
    - MTV inventory service URL and availability
    
    This is essential for troubleshooting MTV setup and understanding the deployment.
    
    Args:
        output_format: Output format (json, yaml, or table)
        
    Returns:
        Version information in the specified format
    """
    args = ["version"]
    if output_format != "table":
        args.extend(["-o", output_format])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def get_plan_vms(
    plan_name: str,
    namespace: str = "",
    output_format: str = "json",
    watch: bool = False
) -> str:
    """Get VMs and their status from a specific migration plan.
    
    This shows all VMs included in a migration plan along with their current migration status,
    progress, and any issues. Essential for monitoring migration progress and troubleshooting
    specific VM migration problems.
    
    Args:
        plan_name: Name of the migration plan to query
        namespace: Kubernetes namespace containing the plan (optional)
        output_format: Output format (json, yaml, or table)
        watch: Watch for changes (not recommended for MCP usage)
        
    Returns:
        JSON/YAML formatted VM status information or table output
    """
    args = ["get", "plan-vms", plan_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    if output_format != "table":
        args.extend(["-o", output_format])
        
    if watch:
        args.append("--watch")
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def describe_plan(
    plan_name: str,
    namespace: str = "",
    with_vms: bool = False
) -> str:
    """Get detailed information about a migration plan.
    
    This provides comprehensive details about a migration plan including its configuration,
    status, progress, and optionally the list of VMs. Essential for understanding plan
    setup and troubleshooting migration issues.
    
    Args:
        plan_name: Name of the migration plan to describe
        namespace: Kubernetes namespace containing the plan (optional)
        with_vms: Include list of VMs in the plan specification
        
    Returns:
        Detailed text description of the migration plan
    """
    args = ["describe", "plan", plan_name]
    
    if namespace:
        args.extend(["-n", namespace])
        
    if with_vms:
        args.append("--with-vms")
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def describe_vm(
    plan_name: str,
    vm_name: str,
    namespace: str = "",
    watch: bool = False
) -> str:
    """Get detailed status of a specific VM in a migration plan.
    
    This provides detailed information about a VM's migration progress, including current
    phase, errors, warnings, and migration steps. Essential for troubleshooting specific
    VM migration issues.
    
    Args:
        plan_name: Name of the migration plan containing the VM
        vm_name: Name of the VM to describe
        namespace: Kubernetes namespace containing the plan (optional)
        watch: Watch VM status with live updates (not recommended for MCP usage)
        
    Returns:
        Detailed text description of the VM's migration status
    """
    args = ["describe", "plan-vm", plan_name, "--vm", vm_name]
    
    if namespace:
        args.extend(["-n", namespace])
        
    if watch:
        args.extend(["--watch"])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def describe_host(
    host_name: str,
    namespace: str = ""
) -> str:
    """Get detailed information about a migration host.
    
    This provides comprehensive details about a migration host including its configuration,
    status, and capabilities. Useful for troubleshooting host connectivity and setup issues.
    
    Args:
        host_name: Name of the migration host to describe
        namespace: Kubernetes namespace containing the host (optional)
        
    Returns:
        Detailed text description of the migration host
    """
    args = ["describe", "host", host_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def describe_mapping(
    mapping_type: str,
    mapping_name: str,
    namespace: str = ""
) -> str:
    """Get detailed information about a network or storage mapping.
    
    This provides comprehensive details about network or storage mappings including
    source and destination mappings, rules, and configuration. Essential for
    troubleshooting mapping setup and connectivity issues.
    
    Args:
        mapping_type: Type of mapping ('network' or 'storage')
        mapping_name: Name of the mapping to describe
        namespace: Kubernetes namespace containing the mapping (optional)
        
    Returns:
        Detailed text description of the mapping configuration
    """
    if mapping_type not in ["network", "storage"]:
        return "Error: mapping_type must be 'network' or 'storage'"
    
    args = ["describe", "mapping", mapping_type, mapping_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    return await run_kubectl_mtv_command(args)


if __name__ == "__main__":
    mcp.run()