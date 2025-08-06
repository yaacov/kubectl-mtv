#!/usr/bin/env python3
"""
FastMCP Server for kubectl-mtv (Write Operations)

This server provides WRITE/DESTRUCTIVE tools to interact with Migration Toolkit for Virtualization (MTV)
through kubectl-mtv commands. It assumes kubectl-mtv is installed and the user is 
logged into a Kubernetes cluster with MTV deployed.

SECURITY WARNING: This server provides destructive operations that can modify, create, or delete MTV resources.
Only enable this server if you need to make changes to your MTV environment.
For safe read-only operations, use the main kubectl_mtv_server.py

MTV Context:
- MTV helps migrate VMs from other KubeVirt clusters, vSphere, oVirt, OpenStack, and OVA files to KubeVirt
- Typical workflow: Provider -> Inventory Discovery -> Mappings -> Plans -> Migration
- These tools allow you to create, modify, and manage MTV resources and migration lifecycle

Tool Categories:
- Resource Creation: create_provider, create_mapping, create_plan, create_host, create_hook, create_vddk
- Resource Deletion: delete_provider, delete_mapping, delete_plan, delete_host, delete_hook  
- Resource Modification: patch_provider, patch_mapping, patch_plan
- Plan Lifecycle: start_plan, cancel_plan, cutover_plan, archive_plan, unarchive_plan
"""

import os
import subprocess
import json
from typing import Any, Optional

from fastmcp import FastMCP

# Get the directory containing this script
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))

# Initialize the FastMCP server
mcp = FastMCP("kubectl-mtv-write")


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


# Plan Lifecycle Operations
@mcp.tool()
async def start_plan(
    plan_name: str,
    namespace: str = "",
    cutover: str = ""
) -> str:
    """Start one or more migration plans to begin migrating VMs.
    
    This initiates the migration process for all VMs in the specified plan(s).
    Plans must be in a ready state with all prerequisites met before starting.
    
    Cutover Scheduling:
    - If cutover time is specified, plan will start but wait until cutover time for final sync
    - Useful for warm migrations to minimize downtime during business hours
    - Time format: ISO8601 (e.g., '2023-12-31T15:30:00Z' or '$(date -d "+1 hour" --iso-8601=sec)')
    - If not specified, defaults to 1 hour from start time
    
    Prerequisites for Plan Start:
    - Provider connectivity must be validated
    - Network and storage mappings must be configured  
    - VM inventory must be current and accessible
    - Target namespace must exist (if specified)
    - Required secrets and permissions must be in place
    
    Args:
        plan_name: Name of the migration plan to start (required, supports multiple space-separated names)
        namespace: Kubernetes namespace containing the plan (optional)
        cutover: Cutover time in ISO8601 format for warm migrations (optional)
        
    Returns:
        Command output confirming plan start
        
    Examples:
        # Start plan immediately
        start_plan("production-migration")
        
        # Start plan with scheduled cutover
        start_plan("production-migration", cutover="2023-12-25T02:00:00Z")
        
        # Start multiple plans with cutover
        start_plan("plan1 plan2 plan3", cutover="2023-12-25T02:00:00Z")
    """
    # Handle multiple plan names by splitting them
    plan_names = plan_name.split() if " " in plan_name else [plan_name]
    args = ["start", "plan"] + plan_names
    
    if namespace:
        args.extend(["-n", namespace])
    
    if cutover:
        args.extend(["--cutover", cutover])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def cancel_plan(
    plan_name: str,
    vms: str,
    namespace: str = ""
) -> str:
    """Cancel specific VMs in a running migration plan.
    
    This stops the migration process for specific VMs within a plan that is currently executing.
    Only the specified VMs will be cancelled - other VMs in the plan continue migrating.
    VMs that have already completed migration cannot be cancelled.
    
    VM Selection Format:
    - Comma-separated list: "vm1,vm2,vm3"
    - File reference: "@filename" to load VM names from JSON/YAML file
    
    Cancellation Effects:
    - Stops in-progress disk transfers for the specified VMs
    - Releases resources allocated for cancelled VM migrations
    - VMs remain in their original location (source environment)
    - Plan continues with remaining VMs not specified for cancellation
    
    Args:
        plan_name: Name of the migration plan containing VMs to cancel (required)
        vms: VM names to cancel - comma-separated list or @filename (required)
        namespace: Kubernetes namespace containing the plan (optional)
        
    Returns:
        Command output confirming VM cancellation
        
    Examples:
        # Cancel specific VMs by name
        cancel_plan("production-migration", "webserver-01,database-02,cache-03")
        
        # Cancel VMs listed in file
        cancel_plan("production-migration", "@/path/to/vms-to-cancel.yaml")
        
        # Cancel single VM
        cancel_plan("production-migration", "problematic-vm")
    """
    args = ["cancel", "plan", plan_name, "--vms", vms]
    
    if namespace:
        args.extend(["-n", namespace])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def cutover_plan(
    plan_name: str,
    namespace: str = ""
) -> str:
    """Perform cutover for a migration plan.
    
    This performs the final cutover phase of migration, typically switching
    from the source VMs to the migrated VMs in the destination.
    
    Args:
        plan_name: Name of the migration plan to cutover
        namespace: Kubernetes namespace containing the plan (optional)
        
    Returns:
        Command output confirming plan cutover
    """
    args = ["cutover", "plan", plan_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def archive_plan(
    plan_name: str,
    namespace: str = ""
) -> str:
    """Archive a completed migration plan.
    
    This archives a migration plan that has completed successfully,
    cleaning up temporary resources while preserving the plan configuration.
    
    Args:
        plan_name: Name of the migration plan to archive
        namespace: Kubernetes namespace containing the plan (optional)
        
    Returns:
        Command output confirming plan archival
    """
    args = ["archive", "plan", plan_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def unarchive_plan(
    plan_name: str,
    namespace: str = ""
) -> str:
    """Unarchive a migration plan.
    
    This restores an archived migration plan, making it available
    for potential re-execution or modification.
    
    Args:
        plan_name: Name of the migration plan to unarchive
        namespace: Kubernetes namespace containing the plan (optional)
        
    Returns:
        Command output confirming plan unarchival
    """
    args = ["unarchive", "plan", plan_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    return await run_kubectl_mtv_command(args)


# Resource Creation Operations
@mcp.tool()
async def create_provider(
    provider_name: str,
    provider_type: str,
    namespace: str = "",
    secret: str = "",
    url: str = "",
    username: str = "",
    password: str = "",
    cacert: str = "",
    insecure_skip_tls: bool = False,
    token: str = "",
    vddk_init_image: str = "",
    sdk_endpoint: str = "",
    use_vddk_aio_optimization: bool = False,
    vddk_buf_size_in_64k: int = 0,
    vddk_buf_count: int = 0,
    provider_domain_name: str = "",
    provider_project_name: str = "",
    provider_region_name: str = ""
) -> str:
    """Create a new provider for connecting to source virtualization platforms.
    
    Providers connect MTV to source virtualization platforms (vSphere, oVirt, OpenStack, OpenShift, OVA).
    Each provider type requires different authentication and connection parameters.
    
    Provider Types and Required Parameters:
    - vSphere: url, username/password OR token, optional: cacert, vddk_init_image, sdk_endpoint
    - oVirt: url, username, password, optional: cacert
    - OpenStack: url, username, password, provider_domain_name, provider_project_name, provider_region_name
    - OpenShift: url, optional: token, optional: cacert
    - OVA: url
    
    Security Notes:
    - Use cacert parameter with certificate content or prefix with @ to load from file
    - Set insecure_skip_tls=True to skip TLS verification (not recommended for production)
    - For existing secrets, use the secret parameter instead of credentials
    
    vSphere-Specific Options:
    - sdk_endpoint: Set to 'esxi' for direct ESXi connection, 'vcenter' for vCenter (default)
    - vddk_init_image: Custom VDDK container image for disk transfers
    - use_vddk_aio_optimization: Enable VDDK AIO optimization for better performance
    - vddk_buf_size_in_64k: VDDK buffer size in 64K units
    - vddk_buf_count: VDDK buffer count for parallel operations
    
    Args:
        provider_name: Name for the new provider (required)
        provider_type: Type of provider - 'vsphere', 'ovirt', 'openstack', 'openshift', or 'ova' (required)
        namespace: Kubernetes namespace to create the provider in (optional)
        secret: Name of existing secret containing provider credentials (optional, alternative to individual credentials)
        url: Provider URL/endpoint (required for most provider types)
        username: Provider credentials username (required unless using secret or token)
        password: Provider credentials password (required unless using secret or token)
        cacert: Provider CA certificate content or @filename to load from file (optional)
        insecure_skip_tls: Skip TLS verification when connecting to the provider (optional, default False)
        token: Provider authentication token (used for OpenShift provider) (optional)
        vddk_init_image: Virtual Disk Development Kit (VDDK) container init image path (vSphere only)
        sdk_endpoint: SDK endpoint type for vSphere provider - 'vcenter' or 'esxi' (optional)
        use_vddk_aio_optimization: Enable VDDK AIO optimization for vSphere provider (optional)
        vddk_buf_size_in_64k: VDDK buffer size in 64K units (vSphere only, optional)
        vddk_buf_count: VDDK buffer count (vSphere only, optional)
        provider_domain_name: OpenStack domain name (OpenStack only)
        provider_project_name: OpenStack project name (OpenStack only)  
        provider_region_name: OpenStack region name (OpenStack only)
        
    Returns:
        Command output confirming provider creation
        
    Examples:
        # Create vSphere provider with credentials
        create_provider("my-vsphere", "vsphere", url="https://vcenter.example.com", 
                       username="admin", password="password123")
        
        # Create OpenStack provider
        create_provider("my-openstack", "openstack", url="https://keystone.example.com:5000/v3",
                       username="admin", password="password123", provider_domain_name="Default",
                       provider_project_name="admin", provider_region_name="RegionOne")
        
        # Create OpenShift provider with token
        create_provider("my-openshift", "openshift", url="https://api.ocp.example.com:6443",
                       token="sha256~abcdef...")
    """
    args = ["create", "provider", "--type", provider_type, provider_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    # Add authentication parameters
    if secret:
        args.extend(["--secret", secret])
    if url:
        args.extend(["--url", url])
    if username:
        args.extend(["--username", username])
    if password:
        args.extend(["--password", password])
    if cacert:
        args.extend(["--cacert", cacert])
    if insecure_skip_tls:
        args.append("--provider-insecure-skip-tls")
    if token:
        args.extend(["--token", token])
        
    # vSphere-specific parameters
    if vddk_init_image:
        args.extend(["--vddk-init-image", vddk_init_image])
    if sdk_endpoint:
        args.extend(["--sdk-endpoint", sdk_endpoint])
    if use_vddk_aio_optimization:
        args.append("--use-vddk-aio-optimization")
    if vddk_buf_size_in_64k > 0:
        args.extend(["--vddk-buf-size-in-64k", str(vddk_buf_size_in_64k)])
    if vddk_buf_count > 0:
        args.extend(["--vddk-buf-count", str(vddk_buf_count)])
        
    # OpenStack-specific parameters
    if provider_domain_name:
        args.extend(["--provider-domain-name", provider_domain_name])
    if provider_project_name:
        args.extend(["--provider-project-name", provider_project_name])  
    if provider_region_name:
        args.extend(["--provider-region-name", provider_region_name])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def create_network_mapping(
    mapping_name: str,
    source_provider: str,
    target_provider: str,
    namespace: str = "",
    network_pairs: str = "",
    inventory_url: str = ""
) -> str:
    """Create a new network mapping between source and target providers.
    
    Network mappings define how source networks map to target networks during VM migration.
    They ensure that migrated VMs are connected to the appropriate networks in the target environment.
    
    Network Pairs Format:
    The network_pairs parameter supports flexible mapping syntax:
    - 'source:target-namespace/target-network' - Map to network in specific namespace  
    - 'source:target-network' - Map to network in default namespace
    - 'source:default' - Map to pod/default networking
    - 'source:ignored' - Don't migrate this network connection
    
    Multiple pairs can be comma-separated: 'net1:target1,net2:target2,net3:ignored'
    
    Args:
        mapping_name: Name for the new network mapping (required)
        source_provider: Name of the source provider (required)
        target_provider: Name of the target provider (required)
        namespace: Kubernetes namespace to create the mapping in (optional)
        network_pairs: Network mapping pairs in format described above (optional)
        inventory_url: Base URL for the inventory service (optional, auto-discovered if not provided)
        
    Returns:
        Command output confirming network mapping creation
        
    Examples:
        # Create basic network mapping
        create_network_mapping("my-net-mapping", "vsphere-provider", "openshift-provider")
        
        # Create with specific network pairs
        create_network_mapping("my-net-mapping", "vsphere-provider", "openshift-provider",
                              network_pairs="VM Network:default,Management:mgmt/management-net,DMZ:ignored")
    """
    args = ["create", "mapping", "network", mapping_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    if source_provider:
        args.extend(["--source", source_provider])
    if target_provider:
        args.extend(["--target", target_provider])
    if network_pairs:
        args.extend(["--network-pairs", network_pairs])
    if inventory_url:
        args.extend(["--inventory-url", inventory_url])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def create_storage_mapping(
    mapping_name: str,
    source_provider: str,
    target_provider: str,
    namespace: str = "",
    storage_pairs: str = "",
    inventory_url: str = ""
) -> str:
    """Create a new storage mapping between source and target providers.
    
    Storage mappings define how source storage/datastores map to target storage classes during VM migration.
    They ensure that migrated VM disks are provisioned on the appropriate storage in the target environment.
    
    Storage Pairs Format:
    The storage_pairs parameter uses the format: 'source:storage-class'
    - Storage classes are cluster-scoped resources in Kubernetes/OpenShift
    - Multiple pairs can be comma-separated: 'datastore1:fast-ssd,datastore2:slow-hdd'
    
    Args:
        mapping_name: Name for the new storage mapping (required)
        source_provider: Name of the source provider (required)
        target_provider: Name of the target provider (required)
        namespace: Kubernetes namespace to create the mapping in (optional)
        storage_pairs: Storage mapping pairs in format 'source:storage-class' (optional)
        inventory_url: Base URL for the inventory service (optional, auto-discovered if not provided)
        
    Returns:
        Command output confirming storage mapping creation
        
    Examples:
        # Create basic storage mapping
        create_storage_mapping("my-storage-mapping", "vsphere-provider", "openshift-provider")
        
        # Create with specific storage pairs
        create_storage_mapping("my-storage-mapping", "vsphere-provider", "openshift-provider",
                              storage_pairs="fast-datastore:ocs-storagecluster-ceph-rbd,slow-datastore:standard")
    """
    args = ["create", "mapping", "storage", mapping_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    if source_provider:
        args.extend(["--source", source_provider])
    if target_provider:
        args.extend(["--target", target_provider])
    if storage_pairs:
        args.extend(["--storage-pairs", storage_pairs])
    if inventory_url:
        args.extend(["--inventory-url", inventory_url])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def create_plan(
    plan_name: str,
    source_provider: str,
    namespace: str = "",
    target_provider: str = "",
    network_mapping: str = "",
    storage_mapping: str = "",
    network_pairs: str = "",
    storage_pairs: str = "",
    vms: str = "",
    pre_hook: str = "",
    post_hook: str = "",
    description: str = "",
    target_namespace: str = "",
    transfer_network: str = "",
    preserve_cluster_cpu_model: bool = False,
    preserve_static_ips: bool = False,
    pvc_name_template: str = "",
    volume_name_template: str = "",
    network_name_template: str = "",
    migrate_shared_disks: bool = True,
    archived: bool = False,
    pvc_name_template_use_generate_name: bool = True,
    delete_guest_conversion_pod: bool = False,
    skip_guest_conversion: bool = False,
    install_legacy_drivers: str = "",
    migration_type: str = "",
    default_target_network: str = "",
    default_target_storage_class: str = "",
    use_compatibility_mode: bool = True,
    target_labels: str = "",
    target_node_selector: str = "",
    warm: bool = False,
    target_affinity: str = "",
    target_power_state: str = "",
    inventory_url: str = ""
) -> str:
    """Create a new migration plan with comprehensive configuration options.
    
    Migration plans define which VMs to migrate and all the configuration for how they should be migrated.
    Plans coordinate providers, mappings, VM selection, and migration behavior.
    
    VM Selection Options:
    - vms: Comma-separated VM names or @filename for YAML/JSON file with VM list
    
    Migration Types:
    - cold: VMs are shut down during migration (default, most reliable)
    - warm: Initial copy while VM runs, brief downtime for final sync
    - live: Minimal downtime migration (advanced, limited compatibility)
    
    Target Power State Options:
    - on: Start VMs after migration
    - off: Leave VMs stopped after migration  
    - auto: Match source VM power state (default)
    
    Template Variables:
    Templates support Go template syntax with VM context:
    - {{.VM.Name}} - VM name
    - {{.VM.ID}} - VM identifier
    - Use templates for consistent naming across migrated resources
    
    Args:
        plan_name: Name for the new migration plan (required)
        source_provider: Name of the source provider to migrate from (required)
        namespace: Kubernetes namespace to create the plan in (optional)
        target_provider: Name of the target provider to migrate to (optional, defaults to destination cluster)
        network_mapping: Name of existing network mapping to use (optional)
        storage_mapping: Name of existing storage mapping to use (optional)
        network_pairs: Network mapping pairs - 'source:target-namespace/target-network' format (optional)
        storage_pairs: Storage mapping pairs - 'source:storage-class' format (optional)
        vms: VM names (comma-separated) or @filename for VM list file (optional)
        pre_hook: Pre-migration hook to add to all VMs (optional)
        post_hook: Post-migration hook to add to all VMs (optional)
        description: Plan description (optional)
        target_namespace: Target namespace for migrated VMs (optional)
        transfer_network: Network attachment definition for VM data transfer (optional)
        preserve_cluster_cpu_model: Preserve CPU model and flags from oVirt cluster (optional, default False)
        preserve_static_ips: Preserve static IPs of vSphere VMs (optional, default False)
        pvc_name_template: Template for generating PVC names for VM disks (optional)
        volume_name_template: Template for generating volume interface names (optional)
        network_name_template: Template for generating network interface names (optional)
        migrate_shared_disks: Whether to migrate shared disks (optional, default True)
        archived: Whether plan should be archived (optional, default False)
        pvc_name_template_use_generate_name: Use generateName for PVC template (optional, default True)
        delete_guest_conversion_pod: Delete conversion pod after migration (optional, default False)
        skip_guest_conversion: Skip guest conversion process (optional, default False)
        install_legacy_drivers: Install legacy Windows drivers - 'true'/'false' (optional)
        migration_type: Migration type - 'cold', 'warm', or 'live' (optional)
        default_target_network: Default target network - 'default' for pod networking (optional)
        default_target_storage_class: Default target storage class (optional)
        use_compatibility_mode: Use compatibility devices when skipping conversion (optional, default True)
        target_labels: Target VM labels - 'key1=value1,key2=value2' format (optional)
        target_node_selector: Target node selector - 'key1=value1,key2=value2' format (optional)
        warm: Enable warm migration - prefer migration_type parameter (optional, default False)
        target_affinity: Target affinity using KARL syntax (optional)
            KARL (Kubernetes Affinity Rule Language) provides human-readable syntax for pod scheduling rules.
            
            KARL Rule Syntax: [RULE_TYPE] pods([SELECTORS]) on [TOPOLOGY]
            - Rule Types: REQUIRE, PREFER, AVOID, REPEL
            - Target: Only pods() supported (no node affinity)
            - Topology: node, zone, region, rack
            - No AND/OR logic support (single rule only)
            
            KARL Examples:
            - 'REQUIRE pods(app=database) on node' - Co-locate with database pods
            - 'PREFER pods(tier=web) on zone' - Prefer same zone as web pods
            - 'AVOID pods(app=cache) on node' - Separate from cache pods on same node
            - 'REPEL pods(workload=heavy) on zone weight=80' - Soft avoid in same zone
        target_power_state: Target power state - 'on', 'off', or 'auto' (optional)
        inventory_url: Base URL for inventory service (optional, auto-discovered if not provided)
        
    Returns:
        Command output confirming plan creation
        
    Examples:
        # Create basic plan
        create_plan("my-plan", "vsphere-provider")
        
        # Create comprehensive plan
        create_plan("my-plan", "vsphere-provider", 
                   target_namespace="migrated-vms",
                   vms="vm1,vm2,vm3",
                   migration_type="warm",
                   network_pairs="VM Network:default,Management:mgmt/mgmt-net",
                   storage_pairs="fast-datastore:ocs-storagecluster-ceph-rbd",
                   target_power_state="on",
                   description="Production VM migration")
                   
        # Create plan with KARL affinity rules
        create_plan("db-plan", "vsphere-provider",
                   vms="database-vm",
                   target_affinity="REQUIRE pods(app=database) on node",
                   description="Co-locate with existing database pods")
                   
        # KARL affinity with zone preference (soft constraint)
        create_plan("web-plan", "vsphere-provider",
                   vms="web-server-01,web-server-02", 
                   target_affinity="PREFER pods(tier=web) on zone",
                   description="Web servers prefer same zone as existing web pods")
    """
    args = ["create", "plan", plan_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    args.extend(["--source", source_provider])
    if target_provider:
        args.extend(["--target", target_provider])
    if network_mapping:
        args.extend(["--network-mapping", network_mapping])
    if storage_mapping:
        args.extend(["--storage-mapping", storage_mapping])
    if network_pairs:
        args.extend(["--network-pairs", network_pairs])
    if storage_pairs:
        args.extend(["--storage-pairs", storage_pairs])
    if vms:
        args.extend(["--vms", vms])
    if pre_hook:
        args.extend(["--pre-hook", pre_hook])
    if post_hook:
        args.extend(["--post-hook", post_hook])
        
    # Plan configuration
    if description:
        args.extend(["--description", description])
    if target_namespace:
        args.extend(["--target-namespace", target_namespace])
    if transfer_network:
        args.extend(["--transfer-network", transfer_network])
    if preserve_cluster_cpu_model:
        args.append("--preserve-cluster-cpu-model")
    if preserve_static_ips:
        args.append("--preserve-static-ips")
    if pvc_name_template:
        args.extend(["--pvc-name-template", pvc_name_template])
    if volume_name_template:
        args.extend(["--volume-name-template", volume_name_template])
    if network_name_template:
        args.extend(["--network-name-template", network_name_template])
    if not migrate_shared_disks:
        args.append("--migrate-shared-disks=false")
    if archived:
        args.append("--archived")
    if not pvc_name_template_use_generate_name:
        args.append("--pvc-name-template-use-generate-name=false")
    if delete_guest_conversion_pod:
        args.append("--delete-guest-conversion-pod")
    if skip_guest_conversion:
        args.append("--skip-guest-conversion")
    if install_legacy_drivers:
        args.extend(["--install-legacy-drivers", install_legacy_drivers])
    if migration_type:
        args.extend(["--migration-type", migration_type])
    if default_target_network:
        args.extend(["--default-target-network", default_target_network])
    if default_target_storage_class:
        args.extend(["--default-target-storage-class", default_target_storage_class])
    if not use_compatibility_mode:
        args.append("--use-compatibility-mode=false")
    if target_labels:
        args.extend(["--target-labels", target_labels])
    if target_node_selector:
        args.extend(["--target-node-selector", target_node_selector])
    if warm:
        args.append("--warm")
    if target_affinity:
        args.extend(["--target-affinity", target_affinity])
    if target_power_state:
        args.extend(["--target-power-state", target_power_state])
    if inventory_url:
        args.extend(["--inventory-url", inventory_url])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def create_host(
    host_name: str,
    provider: str,
    namespace: str = "",
    username: str = "",
    password: str = "",
    existing_secret: str = "",
    ip_address: str = "",
    network_adapter: str = "",
    host_insecure_skip_tls: bool = False,
    cacert: str = "",
    inventory_url: str = ""
) -> str:
    """Create migration hosts for vSphere providers to enable direct data transfer.
    
    Migration hosts enable direct data transfer from ESXi hosts, bypassing vCenter for improved 
    performance. They allow Forklift to utilize ESXi host interfaces directly for network transfer 
    to OpenShift, provided network connectivity exists between OpenShift worker nodes and ESXi hosts.
    
    Host creation is only supported for vSphere providers and requires the host to exist in the 
    provider's inventory. ESXi endpoint providers can automatically use provider credentials.
    
    IP Address Resolution:
    - ip_address: Use specific IP address for direct connection
    - network_adapter: Use IP from named network adapter in inventory (e.g., "Management Network")
    
    Authentication Options:
    - existing_secret: Use existing Kubernetes secret with credentials
    - username/password: Create new credentials (will create secret automatically)
    - ESXi providers: Can automatically inherit provider credentials
    
    Args:
        host_name: Name of the host in provider inventory (required)
        provider: Name of vSphere provider (required)
        namespace: Kubernetes namespace to create the host in (optional)
        username: Username for host authentication (required unless using existing_secret or ESXi provider)
        password: Password for host authentication (required unless using existing_secret or ESXi provider)
        existing_secret: Name of existing secret for host authentication (optional)
        ip_address: IP address for disk transfer (mutually exclusive with network_adapter)
        network_adapter: Network adapter name to get IP from inventory (mutually exclusive with ip_address)
        host_insecure_skip_tls: Skip TLS verification for host connection (optional, default False)
        cacert: CA certificate content or @filename to load from file (optional)
        inventory_url: Base URL for inventory service (optional, auto-discovered if not provided)
        
    Returns:
        Command output confirming host creation
        
    Examples:
        # Create host with direct IP using existing secret
        create_host("esxi-host-01", "my-vsphere-provider", 
                   existing_secret="esxi-credentials", ip_address="192.168.1.10")
        
        # Create host with network adapter lookup
        create_host("esxi-host-01", "my-vsphere-provider",
                   username="root", password="password123", 
                   network_adapter="Management Network")
        
        # Create host for ESXi endpoint provider (inherits credentials)
        create_host("esxi-host-01", "my-esxi-provider", ip_address="192.168.1.10")
    """
    args = ["create", "host", host_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    if provider:
        args.extend(["--provider", provider])
    if username:
        args.extend(["--username", username])
    if password:
        args.extend(["--password", password])
    if existing_secret:
        args.extend(["--existing-secret", existing_secret])
    if ip_address:
        args.extend(["--ip-address", ip_address])
    if network_adapter:
        args.extend(["--network-adapter", network_adapter])
    if host_insecure_skip_tls:
        args.append("--host-insecure-skip-tls")
    if cacert:
        args.extend(["--cacert", cacert])
    if inventory_url:
        args.extend(["--inventory-url", inventory_url])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def create_hook(
    hook_name: str,
    image: str,
    namespace: str = "",
    service_account: str = "",
    playbook: str = "",
    deadline: int = 0
) -> str:
    """Create a migration hook for custom automation during migrations.
    
    Migration hooks allow you to execute custom logic at various points during the migration 
    process by running container images with Ansible playbooks. Hooks can be used for 
    pre-migration validation, post-migration cleanup, or any custom automation needs.
    
    Playbook Loading:
    - Direct content: Pass playbook YAML content as string
    - File loading: Use @filename syntax to load playbook from file
    
    Hook Execution:
    - Hooks run as Kubernetes Jobs during migration
    - Service account provides RBAC permissions for hook operations
    - Deadline sets timeout for hook execution (0 = no timeout)
    
    Args:
        hook_name: Name for the new migration hook (required)
        image: Container image URL to run (required)
        namespace: Kubernetes namespace to create the hook in (optional)
        service_account: Service account to use for the hook (optional)
        playbook: Ansible playbook content or @filename to load from file (optional)
        deadline: Hook deadline in seconds, 0 for no timeout (optional, default 0)
        
    Returns:
        Command output confirming hook creation
        
    Examples:
        # Create basic hook with inline playbook
        create_hook("pre-migration-check", "my-registry/ansible:latest", 
                   playbook="- name: Check connectivity\\n  ping:\\n    data: test")
        
        # Create hook with playbook from file
        create_hook("post-migration-cleanup", "my-registry/ansible:latest",
                   playbook="@/path/to/cleanup-playbook.yaml",
                   service_account="migration-hooks", deadline=300)
        
        # Create simple validation hook
        create_hook("validate-target", "my-registry/validator:latest",
                   service_account="migration-validator", deadline=600)
    """
    args = ["create", "hook", hook_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    args.extend(["--image", image])
    
    if service_account:
        args.extend(["--service-account", service_account])
    if playbook:
        args.extend(["--playbook", playbook])
    if deadline > 0:
        args.extend(["--deadline", str(deadline)])
    
    return await run_kubectl_mtv_command(args)


# Resource Deletion Operations  
@mcp.tool()
async def delete_provider(
    provider_name: str,
    namespace: str = ""
) -> str:
    """Delete a provider.
    
    WARNING: This will remove the provider and may affect associated plans and mappings.
    
    Args:
        provider_name: Name of the provider to delete
        namespace: Kubernetes namespace containing the provider (optional)
        
    Returns:
        Command output confirming provider deletion
    """
    args = ["delete", "provider", provider_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def delete_network_mapping(
    mapping_name: str,
    namespace: str = ""
) -> str:
    """Delete a network mapping.
    
    WARNING: This will remove the network mapping and may affect associated migration plans.
    Plans using this mapping may need to be updated or recreated.
    
    Args:
        mapping_name: Name of the network mapping to delete (required)
        namespace: Kubernetes namespace containing the mapping (optional)
        
    Returns:
        Command output confirming network mapping deletion
    """
    args = ["delete", "mapping", "network", mapping_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def delete_storage_mapping(
    mapping_name: str,
    namespace: str = ""
) -> str:
    """Delete a storage mapping.
    
    WARNING: This will remove the storage mapping and may affect associated migration plans.
    Plans using this mapping may need to be updated or recreated.
    
    Args:
        mapping_name: Name of the storage mapping to delete (required)
        namespace: Kubernetes namespace containing the mapping (optional)
        
    Returns:
        Command output confirming storage mapping deletion
    """
    args = ["delete", "mapping", "storage", mapping_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def delete_plan(
    plan_name: str,
    namespace: str = ""
) -> str:
    """Delete a migration plan.
    
    WARNING: This will remove the migration plan and all associated migration data.
    
    Args:
        plan_name: Name of the plan to delete
        namespace: Kubernetes namespace containing the plan (optional)
        
    Returns:
        Command output confirming plan deletion
    """
    args = ["delete", "plan", plan_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def delete_host(
    host_name: str,
    namespace: str = ""
) -> str:
    """Delete a migration host.
    
    WARNING: This will remove the migration host.
    
    Args:
        host_name: Name of the host to delete
        namespace: Kubernetes namespace containing the host (optional)
        
    Returns:
        Command output confirming host deletion
    """
    args = ["delete", "host", host_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def delete_hook(
    hook_name: str,
    namespace: str = ""
) -> str:
    """Delete a migration hook.
    
    WARNING: This will remove the migration hook.
    
    Args:
        hook_name: Name of the hook to delete
        namespace: Kubernetes namespace containing the hook (optional)
        
    Returns:
        Command output confirming hook deletion
    """
    args = ["delete", "hook", hook_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    return await run_kubectl_mtv_command(args)


# Patch Operations
@mcp.tool()
async def patch_provider(
    provider_name: str,
    namespace: str = "",
    url: str = "",
    username: str = "",
    password: str = "",
    cacert: str = "",
    insecure_skip_tls: bool = None,
    token: str = "",
    vddk_init_image: str = "",
    use_vddk_aio_optimization: bool = None,
    vddk_buf_size_in_64k: int = 0,
    vddk_buf_count: int = 0,
    provider_domain_name: str = "",
    provider_project_name: str = "",
    provider_region_name: str = ""
) -> str:
    """Patch/modify an existing provider by updating URL, credentials, or VDDK settings.
    
    This allows updating provider configuration without recreating it. Provider type and 
    SDK endpoint cannot be changed through patching.
    
    Editable Provider Settings:
    - Authentication: URL, username, password, token, CA certificate
    - Security: TLS verification settings  
    - vSphere VDDK: Init image, AIO optimization, buffer settings
    - OpenStack: Domain, project, and region names
    
    Certificate Loading:
    - Direct content: Pass certificate content as string
    - File loading: Use @filename syntax to load certificate from file
    
    Boolean Parameters:
    - None (default): Don't change the current value
    - True: Enable the setting
    - False: Disable the setting
    
    Args:
        provider_name: Name of the provider to patch (required)
        namespace: Kubernetes namespace containing the provider (optional)
        url: Provider URL/endpoint (optional)
        username: Provider credentials username (optional)
        password: Provider credentials password (optional)
        cacert: Provider CA certificate content or @filename (optional)
        insecure_skip_tls: Skip TLS verification when connecting (optional)
        token: Provider authentication token (for OpenShift) (optional)
        vddk_init_image: VDDK container init image path (vSphere only) (optional)
        use_vddk_aio_optimization: Enable VDDK AIO optimization (vSphere only) (optional)
        vddk_buf_size_in_64k: VDDK buffer size in 64K units (vSphere only) (optional)
        vddk_buf_count: VDDK buffer count (vSphere only) (optional)
        provider_domain_name: OpenStack domain name (OpenStack only) (optional)
        provider_project_name: OpenStack project name (OpenStack only) (optional)  
        provider_region_name: OpenStack region name (OpenStack only) (optional)
        
    Returns:
        Command output confirming provider patch
        
    Examples:
        # Update vSphere provider credentials and VDDK settings
        patch_provider("my-vsphere", url="https://new-vcenter.example.com",
                      username="newuser", vddk_init_image="my-registry/vddk:latest")
        
        # Update OpenStack provider region
        patch_provider("my-openstack", provider_region_name="RegionTwo")
        
        # Enable VDDK optimization and increase buffer settings
        patch_provider("my-vsphere", use_vddk_aio_optimization=True, 
                      vddk_buf_count=32, vddk_buf_size_in_64k=128)
    """
    args = ["patch", "provider", provider_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    # Add authentication parameters
    if url:
        args.extend(["--url", url])
    if username:
        args.extend(["--username", username])
    if password:
        args.extend(["--password", password])
    if cacert:
        args.extend(["--cacert", cacert])
    if token:
        args.extend(["--token", token])
    
    # Add security parameters (only if explicitly set)
    if insecure_skip_tls is not None:
        if insecure_skip_tls:
            args.append("--provider-insecure-skip-tls")
        # Note: There's no --provider-insecure-skip-tls=false flag, omitting sets to false
        
    # vSphere-specific parameters
    if vddk_init_image:
        args.extend(["--vddk-init-image", vddk_init_image])
    if use_vddk_aio_optimization is not None:
        if use_vddk_aio_optimization:
            args.append("--use-vddk-aio-optimization")
        # Note: There's no --use-vddk-aio-optimization=false flag, omitting sets to false
    if vddk_buf_size_in_64k > 0:
        args.extend(["--vddk-buf-size-in-64k", str(vddk_buf_size_in_64k)])
    if vddk_buf_count > 0:
        args.extend(["--vddk-buf-count", str(vddk_buf_count)])
        
    # OpenStack-specific parameters
    if provider_domain_name:
        args.extend(["--provider-domain-name", provider_domain_name])
    if provider_project_name:
        args.extend(["--provider-project-name", provider_project_name])  
    if provider_region_name:
        args.extend(["--provider-region-name", provider_region_name])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def patch_network_mapping(
    mapping_name: str,
    namespace: str = "",
    add_pairs: str = "",
    update_pairs: str = "",
    remove_pairs: str = "",
    inventory_url: str = ""
) -> str:
    """Patch a network mapping by adding, updating, or removing network pairs.
    
    This allows modifying network mapping configuration without recreating the entire mapping.
    You can add new network pairs, update existing ones, or remove unwanted pairs.
    
    Network Pairs Format:
    The pairs parameters use the same format as create operations:
    - 'source:target-namespace/target-network' - Map to network in specific namespace  
    - 'source:target-network' - Map to network in default namespace
    - 'source:default' - Map to pod/default networking
    - 'source:ignored' - Don't migrate this network connection
    
    Multiple pairs are comma-separated: 'net1:target1,net2:target2,net3:ignored'
    
    Operation Types:
    - add-pairs: Add new network mappings (fails if source already exists)
    - update-pairs: Update existing network mappings (fails if source doesn't exist)
    - remove-pairs: Remove network mappings by source name
    
    Args:
        mapping_name: Name of the network mapping to patch (required)
        namespace: Kubernetes namespace containing the mapping (optional)
        add_pairs: Network pairs to add (optional)
        update_pairs: Network pairs to update (optional)
        remove_pairs: Source network names to remove, comma-separated (optional)
        inventory_url: Base URL for inventory service (optional, auto-discovered if not provided)
        
    Returns:
        Command output confirming network mapping patch
        
    Examples:
        # Add new network mappings
        patch_network_mapping("my-mapping", add_pairs="DMZ:dmz-namespace/dmz-net,Test:ignored")
        
        # Update existing mappings
        patch_network_mapping("my-mapping", update_pairs="Management:mgmt-namespace/new-mgmt-net")
        
        # Remove network mappings
        patch_network_mapping("my-mapping", remove_pairs="OldNetwork,UnusedNetwork")
        
        # Combine operations
        patch_network_mapping("my-mapping", 
                             add_pairs="NewNet:default", 
                             update_pairs="ExistingNet:updated-target",
                             remove_pairs="ObsoleteNet")
    """
    args = ["patch", "mapping", "network", mapping_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    if add_pairs:
        args.extend(["--add-pairs", add_pairs])
    if update_pairs:
        args.extend(["--update-pairs", update_pairs])
    if remove_pairs:
        args.extend(["--remove-pairs", remove_pairs])
    if inventory_url:
        args.extend(["--inventory-url", inventory_url])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def patch_storage_mapping(
    mapping_name: str,
    namespace: str = "",
    add_pairs: str = "",
    update_pairs: str = "",
    remove_pairs: str = "",
    inventory_url: str = ""
) -> str:
    """Patch a storage mapping by adding, updating, or removing storage pairs.
    
    This allows modifying storage mapping configuration without recreating the entire mapping.
    You can add new storage pairs, update existing ones, or remove unwanted pairs.
    
    Storage Pairs Format:
    The pairs parameters use the format: 'source:storage-class'
    - Storage classes are cluster-scoped resources in Kubernetes/OpenShift
    - Multiple pairs are comma-separated: 'datastore1:fast-ssd,datastore2:slow-hdd'
    
    Operation Types:
    - add-pairs: Add new storage mappings (fails if source already exists)
    - update-pairs: Update existing storage mappings (fails if source doesn't exist)  
    - remove-pairs: Remove storage mappings by source name
    
    Args:
        mapping_name: Name of the storage mapping to patch (required)
        namespace: Kubernetes namespace containing the mapping (optional)
        add_pairs: Storage pairs to add in format 'source:storage-class' (optional)
        update_pairs: Storage pairs to update in format 'source:storage-class' (optional)
        remove_pairs: Source storage names to remove, comma-separated (optional)
        inventory_url: Base URL for inventory service (optional, auto-discovered if not provided)
        
    Returns:
        Command output confirming storage mapping patch
        
    Examples:
        # Add new storage mappings
        patch_storage_mapping("my-mapping", add_pairs="ssd-datastore:ocs-storagecluster-ceph-rbd")
        
        # Update existing mappings to different storage classes
        patch_storage_mapping("my-mapping", update_pairs="slow-datastore:standard,fast-datastore:premium")
        
        # Remove storage mappings
        patch_storage_mapping("my-mapping", remove_pairs="unused-datastore,old-datastore")
        
        # Combine operations
        patch_storage_mapping("my-mapping",
                             add_pairs="new-datastore:fast-ssd",
                             update_pairs="existing-datastore:updated-class",
                             remove_pairs="obsolete-datastore")
    """
    args = ["patch", "mapping", "storage", mapping_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    if add_pairs:
        args.extend(["--add-pairs", add_pairs])
    if update_pairs:
        args.extend(["--update-pairs", update_pairs])
    if remove_pairs:
        args.extend(["--remove-pairs", remove_pairs])
    if inventory_url:
        args.extend(["--inventory-url", inventory_url])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def patch_plan(
    plan_name: str,
    namespace: str = "",
    transfer_network: str = "",
    install_legacy_drivers: str = "",
    migration_type: str = "",
    target_labels: str = "",
    target_node_selector: str = "",
    use_compatibility_mode: bool = None,
    target_affinity: str = "",
    target_namespace: str = "",
    target_power_state: str = "",
    description: str = "",
    preserve_cluster_cpu_model: bool = None,
    preserve_static_ips: bool = None,
    pvc_name_template: str = "",
    volume_name_template: str = "",
    network_name_template: str = "",
    migrate_shared_disks: bool = None,
    archived: bool = None,
    pvc_name_template_use_generate_name: bool = None,
    delete_guest_conversion_pod: bool = None,
    skip_guest_conversion: bool = None,
    warm: bool = None
) -> str:
    """Patch/modify various fields of an existing migration plan without modifying its VMs.
    
    This allows updating plan configuration without recreating the entire plan. You can modify
    individual plan properties while preserving the VM list and other unchanged settings.
    
    Boolean Parameters:
    - None (default): Don't change the current value
    - True: Set to true 
    - False: Set to false
    
    Migration Types:
    - cold: Traditional migration with VM shutdown
    - warm: Warm migration with reduced downtime
    
    Target Power State:
    - on: Start VMs after migration
    - off: Leave VMs stopped after migration
    - auto: Match source VM power state
    
    Legacy Drivers:
    - true: Install legacy Windows drivers
    - false: Don't install legacy drivers  
    - (empty): Auto-detect based on guest OS
    
    Args:
        plan_name: Name of the migration plan to patch (required)
        namespace: Kubernetes namespace containing the plan (optional)
        transfer_network: Network to use for transferring VM data (optional)
        install_legacy_drivers: Install legacy drivers - 'true', 'false', or empty for auto (optional)
        migration_type: Migration type - 'cold' or 'warm' (optional)
        target_labels: Target VM labels - 'key=value,key2=value2' format (optional)
        target_node_selector: Target node selector - 'key=value,key2=value2' format (optional)
        use_compatibility_mode: Use compatibility mode for migration (optional)
        target_affinity: Target affinity using KARL syntax (optional)
            KARL (Kubernetes Affinity Rule Language) provides human-readable syntax for pod scheduling rules.
            
            KARL Rule Syntax: [RULE_TYPE] pods([SELECTORS]) on [TOPOLOGY]
            - Rule Types: REQUIRE, PREFER, AVOID, REPEL
            - Target: Only pods() supported (no node affinity)
            - Topology: node, zone, region, rack
            - No AND/OR logic support (single rule only)
            
            KARL Examples:
            - 'REQUIRE pods(app=database) on node' - Co-locate with database pods
            - 'PREFER pods(tier=web) on zone' - Prefer same zone as web pods
            - 'AVOID pods(app=cache) on node' - Separate from cache pods on same node
            - 'REPEL pods(workload=heavy) on zone weight=80' - Soft avoid in same zone
        target_namespace: Target namespace for migrated VMs (optional)
        target_power_state: Target power state - 'on', 'off', or 'auto' (optional)
        description: Plan description (optional)
        preserve_cluster_cpu_model: Preserve CPU model from oVirt cluster (optional)
        preserve_static_ips: Preserve static IPs of vSphere VMs (optional)
        pvc_name_template: Template for generating PVC names (optional)
        volume_name_template: Template for generating volume interface names (optional)
        network_name_template: Template for generating network interface names (optional)
        migrate_shared_disks: Whether to migrate shared disks (optional)
        archived: Whether plan should be archived (optional)
        pvc_name_template_use_generate_name: Use generateName for PVC template (optional)
        delete_guest_conversion_pod: Delete conversion pod after migration (optional)
        skip_guest_conversion: Skip guest conversion process (optional)
        warm: Enable warm migration (optional, prefer migration_type parameter)
        
    Returns:
        Command output confirming plan patch
        
    Examples:
        # Update migration type and target namespace
        patch_plan("my-plan", migration_type="warm", target_namespace="migrated-vms")
        
        # Enable compatibility mode and set target power state
        patch_plan("my-plan", use_compatibility_mode=True, target_power_state="on")
        
        # Add KARL affinity rules for co-location with database
        patch_plan("my-plan", target_affinity="REQUIRE pods(app=database) on node")
        
        # Pod anti-affinity to spread VMs across different nodes
        patch_plan("distributed-app", target_affinity="AVOID pods(app=web) on node")
                 
        # Archive plan and add description
        patch_plan("my-plan", archived=True, description="Completed production migration")
    """
    args = ["patch", "plan", plan_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    # Add string parameters
    if transfer_network:
        args.extend(["--transfer-network", transfer_network])
    if install_legacy_drivers:
        args.extend(["--install-legacy-drivers", install_legacy_drivers])
    if migration_type:
        args.extend(["--migration-type", migration_type])
    if target_labels:
        args.extend(["--target-labels", target_labels])
    if target_node_selector:
        args.extend(["--target-node-selector", target_node_selector])
    if target_affinity:
        args.extend(["--target-affinity", target_affinity])
    if target_namespace:
        args.extend(["--target-namespace", target_namespace])
    if target_power_state:
        args.extend(["--target-power-state", target_power_state])
    if description:
        args.extend(["--description", description])
    if pvc_name_template:
        args.extend(["--pvc-name-template", pvc_name_template])
    if volume_name_template:
        args.extend(["--volume-name-template", volume_name_template])
    if network_name_template:
        args.extend(["--network-name-template", network_name_template])
    
    # Add boolean parameters only if explicitly set (not None)
    if use_compatibility_mode is not None:
        args.extend(["--use-compatibility-mode", str(use_compatibility_mode).lower()])
    if preserve_cluster_cpu_model is not None:
        args.extend(["--preserve-cluster-cpu-model", str(preserve_cluster_cpu_model).lower()])
    if preserve_static_ips is not None:
        args.extend(["--preserve-static-ips", str(preserve_static_ips).lower()])
    if migrate_shared_disks is not None:
        args.extend(["--migrate-shared-disks", str(migrate_shared_disks).lower()])
    if archived is not None:
        args.extend(["--archived", str(archived).lower()])
    if pvc_name_template_use_generate_name is not None:
        args.extend(["--pvc-name-template-use-generate-name", str(pvc_name_template_use_generate_name).lower()])
    if delete_guest_conversion_pod is not None:
        args.extend(["--delete-guest-conversion-pod", str(delete_guest_conversion_pod).lower()])
    if skip_guest_conversion is not None:
        args.extend(["--skip-guest-conversion", str(skip_guest_conversion).lower()])
    if warm is not None:
        args.extend(["--warm", str(warm).lower()])
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def patch_plan_vm(
    plan_name: str,
    vm_name: str,
    namespace: str = "",
    target_name: str = "",
    root_disk: str = "",
    instance_type: str = "",
    pvc_name_template: str = "",
    volume_name_template: str = "",
    network_name_template: str = "",
    luks_secret: str = "",
    target_power_state: str = "",
    add_pre_hook: str = "",
    add_post_hook: str = "",
    remove_hook: str = "",
    clear_hooks: bool = False
) -> str:
    """Patch VM-specific fields for a VM within a migration plan's VM list.
    
    This allows you to customize individual VM settings within a plan without affecting other VMs.
    Useful for setting VM-specific configurations like custom names, storage templates, or hooks.
    
    Template Variables:
    Templates support Go template syntax with VM context:
    - {{.VM.Name}} - Original VM name
    - {{.VM.ID}} - VM identifier
    - Use templates for consistent naming patterns
    
    Hook Management:
    - add_pre_hook: Add a pre-migration hook to this VM
    - add_post_hook: Add a post-migration hook to this VM  
    - remove_hook: Remove a specific hook by name
    - clear_hooks: Remove all hooks from this VM
    
    Args:
        plan_name: Name of the migration plan containing the VM (required)
        vm_name: Name of the VM to patch within the plan (required)
        namespace: Kubernetes namespace containing the plan (optional)
        target_name: Custom name for the VM in the target cluster (optional)
        root_disk: The primary disk to boot from (optional)
        instance_type: Override VM's instance type in target (optional)
        pvc_name_template: Go template for naming PVCs for this VM's disks (optional)
        volume_name_template: Go template for naming volume interfaces (optional)
        network_name_template: Go template for naming network interfaces (optional)
        luks_secret: Secret name for disk decryption keys (optional)
        target_power_state: Target power state for this VM - 'on', 'off', or 'auto' (optional)
        add_pre_hook: Add a pre-migration hook to this VM (optional)
        add_post_hook: Add a post-migration hook to this VM (optional)
        remove_hook: Remove a hook from this VM by hook name (optional)
        clear_hooks: Remove all hooks from this VM (optional, default False)
        
    Returns:
        Command output confirming VM patch
        
    Examples:
        # Customize VM name and power state
        patch_plan_vm("my-plan", "source-vm", target_name="migrated-vm", target_power_state="on")
        
        # Add hooks to specific VM
        patch_plan_vm("my-plan", "database-vm", add_pre_hook="db-backup", add_post_hook="db-validate")
        
        # Use custom PVC naming template
        patch_plan_vm("my-plan", "storage-vm", pvc_name_template="{{.VM.Name}}-disk-{{.DiskIndex}}")
    """
    args = ["patch", "plan-vms", plan_name, vm_name]
    
    if namespace:
        args.extend(["-n", namespace])
    
    # Add VM-specific parameters
    if target_name:
        args.extend(["--target-name", target_name])
    if root_disk:
        args.extend(["--root-disk", root_disk])
    if instance_type:
        args.extend(["--instance-type", instance_type])
    if pvc_name_template:
        args.extend(["--pvc-name-template", pvc_name_template])
    if volume_name_template:
        args.extend(["--volume-name-template", volume_name_template])
    if network_name_template:
        args.extend(["--network-name-template", network_name_template])
    if luks_secret:
        args.extend(["--luks-secret", luks_secret])
    if target_power_state:
        args.extend(["--target-power-state", target_power_state])
    
    # Hook management
    if add_pre_hook:
        args.extend(["--add-pre-hook", add_pre_hook])
    if add_post_hook:
        args.extend(["--add-post-hook", add_post_hook])
    if remove_hook:
        args.extend(["--remove-hook", remove_hook])
    if clear_hooks:
        args.append("--clear-hooks")
    
    return await run_kubectl_mtv_command(args)


if __name__ == "__main__":
    mcp.run()