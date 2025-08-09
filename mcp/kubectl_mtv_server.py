#!/usr/bin/env python3
"""
FastMCP Server for kubectl-mtv (Read-Only Operations)

This server provides READ-ONLY tools to interact with Migration Toolkit for Virtualization (MTV)
through kubectl-mtv commands. It assumes kubectl-mtv is installed and the user is 
logged into a Kubernetes cluster with MTV deployed.

MTV Context:
- MTV helps migrate VMs from other KubeVirt clusters, vSphere, oVirt, OpenStack, and OVA files to KubeVirt
- These tools provide comprehensive read-only discovery, monitoring and troubleshooting capabilities

Typical Workflow (using write tools):
1. Create Providers: connect to source/target environments
2. Inventory Discovery: explore available VMs, networks, storage
3. Create Mappings: define network and storage mappings
4. Create Plans: select VMs using list_inventory_vms(output_format="planvms") and configure migration settings

Let the user edit the plan if needed.
5. Start Migration: execute the migration
6. Monitor Progress: track status and troubleshoot issues

Tool Categories:
- Provider Management: list_providers
- Migration Planning: list_plans, list_mappings, list_hosts, list_hooks, get_plan_vms
- Inventory Discovery: list_inventory_* tools for exploring source environments
- Version Information: get_version for deployment details
- Debugging: get_logs for troubleshooting MTV controller and importer pod issues
- Storage Debugging: get_migration_pvcs, get_migration_datavolumes, get_migration_storage for tracking VM migration storage with detailed diagnostics

Integration with Write Tools:
Use these read tools to discover and prepare data, then use write tools to create/modify resources:
- list_inventory_vms(output_format="planvms") output can be used directly with create_plan(vms="@file.yaml")
- get_plan_vms() shows migration status for troubleshooting with cancel_plan()
- list_inventory_networks/storage() help identify mappings for create_*_mapping() tools
- Use output_format="planvms" specifically for plan VM selection (minimal VM structures)
"""

import os
import subprocess
import json
from typing import Any, Optional

from fastmcp import FastMCP

# Get the directory containing this script
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))

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



# Sub-action methods for basic resources
async def _list_providers(namespace: str, all_namespaces: bool, output_format: str, watch: bool) -> str:
    """List providers implementation."""
    args = ["get", "provider"] + await _build_base_args({
        "namespace": namespace, 
        "all_namespaces": all_namespaces, 
        "output_format": output_format
    })
    if watch:
        args.append("--watch")
    return await run_kubectl_mtv_command(args)

async def _list_plans(namespace: str, all_namespaces: bool, output_format: str, watch: bool) -> str:
    """List plans implementation."""
    args = ["get", "plan"] + await _build_base_args({
        "namespace": namespace,
        "all_namespaces": all_namespaces,
        "output_format": output_format
    })
    if watch:
        args.append("--watch")
    return await run_kubectl_mtv_command(args)

async def _list_mappings(namespace: str, all_namespaces: bool, output_format: str, watch: bool) -> str:
    """List mappings implementation."""
    args = ["get", "mapping"] + await _build_base_args({
        "namespace": namespace,
        "all_namespaces": all_namespaces,
        "output_format": output_format
    })
    if watch:
        args.append("--watch")
    return await run_kubectl_mtv_command(args)

async def _list_hosts(namespace: str, all_namespaces: bool, output_format: str, watch: bool) -> str:
    """List hosts implementation."""
    args = ["get", "host"] + await _build_base_args({
        "namespace": namespace,
        "all_namespaces": all_namespaces,
        "output_format": output_format
    })
    if watch:
        args.append("--watch")
    return await run_kubectl_mtv_command(args)

async def _list_hooks(namespace: str, all_namespaces: bool, output_format: str, watch: bool) -> str:
    """List hooks implementation."""
    args = ["get", "hook"] + await _build_base_args({
        "namespace": namespace,
        "all_namespaces": all_namespaces,
        "output_format": output_format
    })
    if watch:
        args.append("--watch")
    return await run_kubectl_mtv_command(args)


# Unified Tool functions using FastMCP
@mcp.tool()
async def ListResources(
    resource_type: str,
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json",
    watch: bool = False
) -> str:
    """List MTV resources in the cluster.
    
    Unified tool to list various MTV resource types including providers, plans, mappings, hosts, and hooks.
    This consolidates multiple list operations into a single efficient tool.
    
    Args:
        resource_type: Type of resource to list - 'provider', 'plan', 'mapping', 'host', or 'hook'
        namespace: Kubernetes namespace to query (optional, defaults to current namespace)
        all_namespaces: List resources across all namespaces
        output_format: Output format (json, yaml, or table)
        watch: Watch for changes (not recommended for MCP, only supported for plans)
        
    Returns:
        JSON/YAML formatted resource information or table output
        
    Examples:
        # List all providers
        ListResources("provider")
        
        # List plans across all namespaces with table output
        ListResources("plan", all_namespaces=True, output_format="table")
        
        # Watch plan changes (specific namespace)
        ListResources("plan", namespace="demo", watch=True)
    """
    # Validate resource type
    valid_types = ["provider", "plan", "mapping", "host", "hook"]
    if resource_type not in valid_types:
        raise KubectlMTVError(f"Invalid resource_type '{resource_type}'. Valid types: {', '.join(valid_types)}")
    
    # Route to appropriate sub-action method
    if resource_type == "provider":
        return await _list_providers(namespace, all_namespaces, output_format, watch)
    elif resource_type == "plan":
        return await _list_plans(namespace, all_namespaces, output_format, watch)
    elif resource_type == "mapping":
        return await _list_mappings(namespace, all_namespaces, output_format, watch)
    elif resource_type == "host":
        return await _list_hosts(namespace, all_namespaces, output_format, watch)
    elif resource_type == "hook":
        return await _list_hooks(namespace, all_namespaces, output_format, watch)


@mcp.tool()
async def ListInventory(
    resource_type: str,
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List inventory resources from a provider.
    
    Unified tool to query various resource types from provider inventories.
    Supports all resource types with powerful SQL-like query capabilities.
    
    AVAILABLE RESOURCE TYPES BY PROVIDER:
    vSphere: vm, network, storage, host, cluster, datacenter, datastore, folder, resource-pool
    oVirt: vm, network, storage, host, cluster, datacenter, disk, disk-profile, nic-profile
    OpenStack: vm, network, storage, flavor, image, instance, project, subnet, port, volumetype
    OpenShift: vm, network, storage, namespace, pvc, data-volume
    OVA: vm, network, storage
    
    QUERY SYNTAX:
    All inventory tools support SQL-like queries with Tree Search Language (TSL):
    - SELECT field1, field2 AS alias, function(field3) AS name
    - WHERE condition (using TSL operators and functions)
    - ORDER BY field1 [ASC|DESC], field2
    - LIMIT n
    
    TSL OPERATORS: =, !=, <, <=, >, >=, LIKE, ILIKE, ~= (regex), IN, BETWEEN, AND, OR, NOT
    TSL FUNCTIONS: sum(), len(), any(), all()
    TSL LITERALS: strings ('text'), numbers (1024, 2.5Gi), dates ('2023-01-01'), booleans (true/false)
    TSL ARRAY ACCESS: Use [*] for array elements, dot notation for nested fields (e.g., disks[*].capacity, parent.name)
    
    EXAMPLE QUERIES FOR SPECIALIZED RESOURCES:
    
    Folders (vSphere): 
    - "WHERE name LIKE '%VM%' AND type = 'vm'"
    - "SELECT name, path, parent.name, childrenCount ORDER BY name"
    
    Flavors (OpenStack): 
    - "WHERE vcpus >= 4 AND ram >= 8192 ORDER BY vcpus DESC"
    - "SELECT name, vcpus, ram, disk, isPublic WHERE isPublic = true"
    
    Disk Profiles (oVirt): 
    - "WHERE storageDomain = 'specific-domain-id'"
    - "SELECT name, description, storageDomain, qosId ORDER BY name"
    
    NIC Profiles (oVirt): 
    - "WHERE networkFilter != ''"
    - "SELECT name, description, network, portMirroring, customProperties ORDER BY name"
    
    Projects (OpenStack): 
    - "WHERE enabled = true ORDER BY name"
    - "SELECT name, description, enabled, isDomain ORDER BY name"
    
    Resource Pools (vSphere): 
    - "WHERE cpuAllocation.limit > 0"
    - "SELECT name, cpuLimit, memoryLimit, parent.name ORDER BY cpuLimit DESC"
    
    Subnets (OpenStack):
    - "WHERE enableDhcp = true AND ipVersion = 4"
    - "SELECT name, cidr, gatewayIp, networkId, enableDhcp ORDER BY name"
    
    Images (OpenStack):
    - "WHERE status = 'active' AND visibility = 'public'"
    - "SELECT name, status, visibility, diskFormat, containerFormat, size ORDER BY name"
    
    DataVolumes (OpenShift):
    - "WHERE object.status.phase = 'Succeeded'"
    - "SELECT name, namespace, object.spec.source, object.status.phase ORDER BY name"
    
    PVCs (OpenShift):
    - "WHERE object.status.phase = 'Bound'"
    - "SELECT name, namespace, object.spec.storageClassName, object.status.capacity.storage ORDER BY name"
    
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
async def GetLogs(
    pod_type: str = "controller",
    container: str = "main", 
    lines: int = 100,
    follow: bool = False,
    namespace: str = "",
    plan_id: str = "",
    migration_id: str = "",
    vm_id: str = ""
) -> str:
    """Get logs from MTV-related pods for debugging.
    
    This tool can retrieve logs from:
    1. MTV controller pod (main, inventory containers) - auto-detects namespace and pod
    2. Importer pods - finds pod using migration labels and prime PVC annotations
    
    Pod Types:
    - controller: MTV forklift-controller pod (default)
    - importer: CDI importer pod for VM disk migration
    
    For controller pods:
    - Automatically finds MTV operator namespace and running controller pod
    - Supports 'main' and 'inventory' containers
    
    For importer pods:
    - Uses plan_id, migration_id, vm_id to find migration PVCs
    - Locates prime PVC with cdi.kubevirt.io/storage.import.importPodName annotation
    - Retrieves logs from the importer pod
    
    Args:
        pod_type: Type of pod to get logs from ('controller' or 'importer'). Defaults to 'controller'
        container: Container name for controller pods (main, inventory). Defaults to 'main'
        lines: Number of recent log lines to retrieve. Defaults to 100
        follow: Follow log output (stream logs). Not recommended for MCP usage
        namespace: Override namespace (optional, auto-detected for controller)
        plan_id: Plan UUID for finding importer pods (required for importer type)
        migration_id: Migration UUID for finding importer pods (required for importer type) 
        vm_id: VM ID for finding importer pods (required for importer type)
        
    Returns:
        JSON structure containing pod information and logs:
        {
            "pod": { ... pod JSON with status, conditions, etc ... },
            "logs": "pod logs content"
        }
        
    Examples:
        # Get controller main container logs
        get_logs("controller", "main", 200)
        
        # Get controller inventory container logs  
        get_logs("controller", "inventory", 100)
        
        # Get importer pod logs for specific VM migration
        get_logs("importer", "", 100, False, "demo", "plan-uuid", "migration-uuid", "vm-47")
    """
    try:
        if pod_type == "controller":
            return await _get_controller_logs(container, lines, follow, namespace)
        elif pod_type == "importer":
            return await _get_importer_logs(lines, follow, namespace, plan_id, migration_id, vm_id)
        else:
            return f"Error: Unknown pod_type '{pod_type}'. Supported types: 'controller', 'importer'"
            
    except Exception as e:
        return f"Error retrieving {pod_type} logs: {str(e)}"


async def _get_controller_logs(container: str, lines: int, follow: bool, namespace: str) -> str:
    """Get logs and pod information from the MTV controller pod."""
    # Get the MTV operator namespace if not provided
    if not namespace:
        namespace = await get_mtv_operator_namespace()
    
    # Find the controller pod
    pod_name = await find_controller_pod(namespace)
    
    # Get pod information
    pod_info_output = await run_kubectl_command([
        "get", "pod", "-n", namespace, pod_name, "-o", "json"
    ])
    pod_info = json.loads(pod_info_output)
    
    # Build kubectl logs command
    logs_args = ["logs", "-n", namespace, pod_name, "-c", container, "--tail", str(lines)]
    
    if follow:
        logs_args.append("-f")
    
    # Get the logs
    logs = await run_kubectl_command(logs_args)
    
    # Return structured response
    result = {
        "pod": pod_info,
        "logs": logs
    }
    return json.dumps(result, indent=2)


async def _get_importer_logs(lines: int, follow: bool, namespace: str, plan_id: str, migration_id: str, vm_id: str) -> str:
    """Get logs and pod information from importer pod by finding it via migration labels and prime PVC annotations."""
    if not plan_id or not migration_id or not vm_id:
        raise KubectlMTVError("plan_id, migration_id, and vm_id are required for importer pod logs")
    
    if not namespace:
        raise KubectlMTVError("namespace is required for importer pod logs")
    
    # Step 1: Find PVCs with migration labels
    label_selector = f"plan={plan_id},migration={migration_id},vmID={vm_id}"
    pvcs_output = await run_kubectl_command([
        "get", "pvc", "-n", namespace,
        "-l", label_selector,
        "-o", "json"
    ])
    
    pvcs_data = json.loads(pvcs_output)
    pvcs = pvcs_data.get("items", [])
    
    if not pvcs:
        raise KubectlMTVError(f"No PVCs found with labels plan={plan_id}, migration={migration_id}, vmID={vm_id}")
    
    # Step 2: Find prime PVCs that are owned by the migration PVCs
    migration_pvc_uid = None
    for pvc in pvcs:
        migration_pvc_uid = pvc.get("metadata", {}).get("uid")
        if migration_pvc_uid:
            break
    
    if not migration_pvc_uid:
        raise KubectlMTVError("Could not find migration PVC UID")
    
    # Find prime PVC owned by the migration PVC
    all_pvcs_output = await run_kubectl_command([
        "get", "pvc", "-n", namespace,
        "-o", "json"
    ])
    
    all_pvcs_data = json.loads(all_pvcs_output)
    importer_pod_name = None
    
    for pvc in all_pvcs_data.get("items", []):
        # Check if this PVC is owned by our migration PVC
        owner_refs = pvc.get("metadata", {}).get("ownerReferences", [])
        for owner_ref in owner_refs:
            if owner_ref.get("uid") == migration_pvc_uid:
                # This is a prime PVC owned by our migration PVC
                annotations = pvc.get("metadata", {}).get("annotations", {})
                importer_pod_name = annotations.get("cdi.kubevirt.io/storage.import.importPodName")
                if importer_pod_name:
                    break
        if importer_pod_name:
            break
    
    if not importer_pod_name:
        raise KubectlMTVError(f"Could not find importer pod name in prime PVC annotations for migration PVC UID {migration_pvc_uid}")
    
    # Step 3: Get pod information
    pod_info_output = await run_kubectl_command([
        "get", "pod", "-n", namespace, importer_pod_name, "-o", "json"
    ])
    pod_info = json.loads(pod_info_output)
    
    # Step 4: Get logs from the importer pod
    logs_args = ["logs", "-n", namespace, importer_pod_name, "--tail", str(lines)]
    
    if follow:
        logs_args.append("-f")
    
    logs = await run_kubectl_command(logs_args)
    
    # Return structured response
    result = {
        "pod": pod_info,
        "logs": logs
    }
    return json.dumps(result, indent=2)


# Sub-action methods for migration storage
async def _get_migration_pvcs(
    migration_id: str, plan_id: str, vm_id: str, 
    namespace: str, all_namespaces: bool, output_format: str
) -> str:
    """Get PVCs implementation."""
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
        
        result = await run_kubectl_command(cmd_args)
        
        # For JSON output, enhance with describe information
        if output_format == "json" and result:
            try:
                import json
                pvc_data = json.loads(result)
                
                # Add describe information for each PVC
                for pvc in pvc_data.get("items", []):
                    pvc_name = pvc["metadata"]["name"]
                    pvc_namespace = pvc["metadata"]["namespace"]
                    
                    try:
                        describe_cmd = ["describe", "pvc", pvc_name, "-n", pvc_namespace]
                        describe_result = await run_kubectl_command(describe_cmd)
                        pvc["describe"] = describe_result
                    except Exception as e:
                        pvc["describe"] = f"Could not get describe output: {str(e)}"
                
                return json.dumps(pvc_data, indent=2)
            except json.JSONDecodeError:
                pass
        
        return result
        
    except Exception as e:
        return f"Error retrieving migration PVCs: {str(e)}"

async def _get_migration_datavolumes(
    migration_id: str, plan_id: str, vm_id: str,
    namespace: str, all_namespaces: bool, output_format: str
) -> str:
    """Get DataVolumes implementation."""
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
        
        result = await run_kubectl_command(cmd_args)
        
        # For JSON output, enhance with describe information
        if output_format == "json" and result:
            try:
                import json
                dv_data = json.loads(result)
                
                # Add describe information for each DataVolume
                for dv in dv_data.get("items", []):
                    dv_name = dv["metadata"]["name"]
                    dv_namespace = dv["metadata"]["namespace"]
                    
                    try:
                        describe_cmd = ["describe", "datavolume", dv_name, "-n", dv_namespace]
                        describe_result = await run_kubectl_command(describe_cmd)
                        dv["describe"] = describe_result
                    except Exception as e:
                        dv["describe"] = f"Could not get describe output: {str(e)}"
                
                return json.dumps(dv_data, indent=2)
            except json.JSONDecodeError:
                pass
        
        return result
        
    except Exception as e:
        return f"Error retrieving migration DataVolumes: {str(e)}"


@mcp.tool()
async def GetMigrationStorage(
    resource_type: str = "all",
    migration_id: str = "",
    plan_id: str = "",
    vm_id: str = "",
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json"
) -> str:
    """Get storage resources (PVCs and DataVolumes) related to VM migrations.
    
    Unified tool to access migration storage resources with granular control over resource types.
    Supports filtering by migration labels to find specific storage resources.
    
    Resource Types:
    - 'all': Get both PVCs and DataVolumes (default)
    - 'pvc': Get only PersistentVolumeClaims 
    - 'datavolume': Get only DataVolumes
    
    Find storage resources that are part of VM migrations by searching for specific labels:
    - migration: Migration UUID (NOT the migration name)
    - plan: Plan UUID (NOT the plan name)  
    - vmID: VM identifier (e.g., vm-47)
    
    IMPORTANT: Use UUIDs, not names! 
    - CORRECT: migration_id="4399056b-4f08-497d-a559-3dd530de3459" (UUID from plan status)
    - WRONG: migration_id="migrate-small-vm-mmpj4" (migration name - won't work)
    - CORRECT: plan_id="3943f9a2-d4a4-4326-b25c-57d06ff53c21" (UUID from plan metadata)  
    - WRONG: plan_id="migrate-small-vm" (plan name - won't work)
    
    How to get the correct UUIDs:
    1. Use GetPlanVms() to get migration UUIDs from plan status  
    2. Use ListResources("plan") with json output to get plan UUIDs from metadata.uid
    3. Check kubectl labels: kubectl get pvc,dv -n <namespace> --show-labels
    
    Args:
        resource_type: Type of storage resource - 'all', 'pvc', or 'datavolume' (default 'all')
        migration_id: Migration UUID to filter by (optional) - get from plan VM status
        plan_id: Plan UUID to filter by (optional) - get from plan metadata.uid
        vm_id: VM ID to filter by (optional) - e.g., vm-47, vm-73
        namespace: Kubernetes namespace to search in (optional)
        all_namespaces: Search across all namespaces
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted storage information or table output
        
        Enhanced JSON Output:
        When using JSON output format, resources include:
        - "describe" field with kubectl describe output
        - Complete diagnostic information and events
        
    Examples:
        # Get all storage for specific migration
        GetMigrationStorage("all", "4399056b-4f08-497d-a559-3dd530de3459", 
                           "3943f9a2-d4a4-4326-b25c-57d06ff53c21", "vm-47", "demo")
        
        # Get only PVCs in namespace (table format)
        GetMigrationStorage("pvc", namespace="demo", output_format="table")
        
        # Get only DataVolumes for specific plan
        GetMigrationStorage("datavolume", plan_id="3943f9a2-d4a4-4326-b25c-57d06ff53c21")
    """
    try:
        # Validate resource type
        valid_types = ["all", "pvc", "datavolume"]
        if resource_type not in valid_types:
            raise KubectlMTVError(f"Invalid resource_type '{resource_type}'. Valid types: {', '.join(valid_types)}")
        
        # Route to appropriate sub-action method(s)
        if resource_type == "pvc":
            return await _get_migration_pvcs(migration_id, plan_id, vm_id, namespace, all_namespaces, output_format)
        elif resource_type == "datavolume":
            return await _get_migration_datavolumes(migration_id, plan_id, vm_id, namespace, all_namespaces, output_format)
        else:  # resource_type == "all"
            # Get both PVCs and DataVolumes
            pvcs_result = await _get_migration_pvcs(migration_id, plan_id, vm_id, namespace, all_namespaces, output_format)
            dvs_result = await _get_migration_datavolumes(migration_id, plan_id, vm_id, namespace, all_namespaces, output_format)
            
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
async def GetVersion(
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
async def GetPlanVms(
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
        
    Integration with Write Tools:
        Use this tool to monitor migration progress and troubleshoot:
        1. Monitor progress: get_plan_vms("my-plan") 
        2. Cancel problematic VMs: cancel_plan("my-plan", "failed-vm1,stuck-vm2")
        3. Get detailed logs: get_logs("importer", plan_id="...", migration_id="...", vm_id="...")
    """
    args = ["get", "plan-vms", plan_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    if output_format != "table":
        args.extend(["-o", output_format])
        
    if watch:
        args.append("--watch")
    
    return await run_kubectl_mtv_command(args)





if __name__ == "__main__":
    mcp.run()
    mcp.run()