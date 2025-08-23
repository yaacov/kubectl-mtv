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

Integration with Read Tools:
Use read tools to discover and analyze data before making changes:
- Use ListInventory("vm", "provider", output_format="planvms") to get VM structures for create_plan(vms="@file.yaml")
- Use list_inventory_networks/storage() to identify available resources for mappings
- Use get_plan_vms() to monitor progress and identify VMs for cancel_plan()
- Use list_providers() to verify provider status before creating plans
- Use get_logs() for troubleshooting failed operations

Network Mapping Important Notes:
- All source networks must be mapped (no unmapped source networks allowed)
- Pod networking and specific multus targets can only be mapped once (no duplicate targets)
- Use 'source:ignored' for networks that don't have suitable targets or are not needed

Storage Mapping Important Notes:
- All source storages must be mapped (no unmapped source storages allowed)
- Storage class selection priority: user-defined > virt annotation > k8s annotation > name match > first available
- Virt annotation: storageclass.kubevirt.io/is-default-virt-class=true
- K8s annotation: storageclass.kubernetes.io/is-default-class=true
- Name matching: case-insensitive search for "virtualization" in storage class name
"""

import os
import subprocess
import json

from fastmcp import FastMCP

# Get the directory containing this script
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))

# Initialize the FastMCP server
mcp = FastMCP("kubectl-mtv-write")


class KubectlMTVError(Exception):
    """Custom exception for kubectl-mtv command errors."""

    pass


async def run_kubectl_mtv_command(args: list[str]) -> str:
    """Run a kubectl-mtv command and return structured JSON with command info."""
    cmd = ["kubectl-mtv"] + args
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        response = {
            "command": " ".join(cmd),
            "return_value": 0,
            "stdout": result.stdout,
            "stderr": result.stderr,
        }
        return json.dumps(response, indent=2)
    except subprocess.CalledProcessError as e:
        response = {
            "command": " ".join(cmd),
            "return_value": e.returncode,
            "stdout": e.stdout if e.stdout else "",
            "stderr": e.stderr if e.stderr else "",
        }
        return json.dumps(response, indent=2)
    except FileNotFoundError:
        response = {
            "command": " ".join(cmd),
            "return_value": -1,
            "stdout": "",
            "stderr": "kubectl-mtv not found in PATH",
        }
        return json.dumps(response, indent=2)


# Sub-action methods for plan lifecycle
async def _start_plan(plan_name: str, namespace: str, cutover: str) -> str:
    """Start plan implementation."""
    # Handle multiple plan names by splitting them
    plan_names = plan_name.split() if " " in plan_name else [plan_name]
    args = ["start", "plan"] + plan_names

    if namespace:
        args.extend(["-n", namespace])

    if cutover:
        args.extend(["--cutover", cutover])

    return await run_kubectl_mtv_command(args)


async def _cancel_plan(plan_name: str, namespace: str, vms: str) -> str:
    """Cancel plan implementation."""
    args = ["cancel", "plan", plan_name, "--vms", vms]

    if namespace:
        args.extend(["-n", namespace])

    return await run_kubectl_mtv_command(args)


async def _cutover_plan(plan_name: str, namespace: str) -> str:
    """Cutover plan implementation."""
    args = ["cutover", "plan", plan_name]

    if namespace:
        args.extend(["-n", namespace])

    return await run_kubectl_mtv_command(args)


async def _archive_plan(plan_name: str, namespace: str) -> str:
    """Archive plan implementation."""
    args = ["archive", "plan", plan_name]

    if namespace:
        args.extend(["-n", namespace])

    return await run_kubectl_mtv_command(args)


async def _unarchive_plan(plan_name: str, namespace: str) -> str:
    """Unarchive plan implementation."""
    args = ["unarchive", "plan", plan_name]

    if namespace:
        args.extend(["-n", namespace])

    return await run_kubectl_mtv_command(args)


# Plan Lifecycle Operations
@mcp.tool()
async def ManagePlanLifecycle(
    action: str, plan_name: str, namespace: str = "", cutover: str = "", vms: str = ""
) -> str:
    """Manage migration plan lifecycle operations.

    Unified tool for all plan lifecycle actions including start, cancel, cutover, archive, and unarchive.
    Each action has specific prerequisites and effects on the migration process.

    Actions:
    - 'start': Begin migrating VMs in the plan
    - 'cancel': Cancel specific VMs in a running migration
    - 'cutover': Perform final cutover phase
    - 'archive': Archive completed migration plan
    - 'unarchive': Restore archived migration plan

    Action-Specific Parameters:
    - start: cutover (optional ISO8601 timestamp for warm migrations)
    - cancel: vms (required - VM names to cancel)
    - cutover, archive, unarchive: no additional parameters

    Prerequisites by Action:
    Start:
    - Provider connectivity must be validated
    - Network and storage mappings must be configured
    - VM inventory must be current and accessible
    - Target namespace must exist (if specified)

    Cancel:
    - Plan must be actively running
    - VMs must not have completed migration

    Cutover:
    - Plan must be in warm migration state
    - Initial data sync must be complete

    Archive/Unarchive:
    - Plan must be in completed/archived state respectively

    Args:
        action: Lifecycle action - 'start', 'cancel', 'cutover', 'archive', 'unarchive'
        plan_name: Name of the migration plan (supports space-separated names for start action)
        namespace: Kubernetes namespace containing the plan (optional)
        cutover: Cutover time in ISO8601 format for start action (optional)
        vms: VM names for cancel action - comma-separated or @filename (required for cancel)

    Returns:
        Command output confirming the lifecycle action

    Examples:
        # Start plan immediately
        ManagePlanLifecycle("start", "production-migration")

        # Start plan with scheduled cutover
        ManagePlanLifecycle("start", "production-migration", cutover="2023-12-25T02:00:00Z")

        # Start multiple plans
        ManagePlanLifecycle("start", "plan1 plan2 plan3")

        # Cancel specific VMs
        ManagePlanLifecycle("cancel", "production-migration", vms="webserver-01,database-02")

        # Cancel VMs from file
        ManagePlanLifecycle("cancel", "production-migration", vms="@vms-to-cancel.json")

        # Perform cutover
        ManagePlanLifecycle("cutover", "production-migration")

        # Note: To archive/unarchive plans, use PatchPlan:
        # PatchPlan("production-migration", archived=True)   # Archive
        # PatchPlan("production-migration", archived=False)  # Unarchive
    """
    # Validate action
    valid_actions = ["start", "cancel", "cutover"]
    if action not in valid_actions:
        raise KubectlMTVError(
            f"Invalid action '{action}'. Valid actions: {', '.join(valid_actions)}"
        )

    # Validate action-specific required parameters
    if action == "cancel" and not vms:
        raise KubectlMTVError("The 'vms' parameter is required for cancel action")

    # Route to appropriate sub-action method
    if action == "start":
        return await _start_plan(plan_name, namespace, cutover)
    elif action == "cancel":
        return await _cancel_plan(plan_name, namespace, vms)
    elif action == "cutover":
        return await _cutover_plan(plan_name, namespace)


# Resource Creation Operations
@mcp.tool()
async def CreateProvider(
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
    provider_region_name: str = "",
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

    Certificate Loading Examples:
    - Direct content: cacert="-----BEGIN CERTIFICATE-----\n..."
    - From file: cacert="@/path/to/ca-cert.pem"

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


# Sub-action methods for mapping operations
async def _create_network_mapping(
    mapping_name: str,
    source_provider: str,
    target_provider: str,
    namespace: str,
    pairs: str,
    inventory_url: str,
) -> str:
    """Create network mapping implementation."""
    args = ["create", "mapping", "network", mapping_name]

    if namespace:
        args.extend(["-n", namespace])

    if source_provider:
        args.extend(["--source", source_provider])
    if target_provider:
        args.extend(["--target", target_provider])
    if pairs:
        args.extend(["--network-pairs", pairs])
    if inventory_url:
        args.extend(["--inventory-url", inventory_url])

    return await run_kubectl_mtv_command(args)


async def _create_storage_mapping(
    mapping_name: str,
    source_provider: str,
    target_provider: str,
    namespace: str,
    pairs: str,
    inventory_url: str,
    default_volume_mode: str = "",
    default_access_mode: str = "",
    default_offload_plugin: str = "",
    default_offload_secret: str = "",
    default_offload_vendor: str = "",
) -> str:
    """Create storage mapping implementation with enhanced options."""
    args = ["create", "mapping", "storage", mapping_name]

    if namespace:
        args.extend(["-n", namespace])

    if source_provider:
        args.extend(["--source", source_provider])
    if target_provider:
        args.extend(["--target", target_provider])
    if pairs:
        args.extend(["--storage-pairs", pairs])
    if default_volume_mode:
        args.extend(["--default-volume-mode", default_volume_mode])
    if default_access_mode:
        args.extend(["--default-access-mode", default_access_mode])
    if default_offload_plugin:
        args.extend(["--default-offload-plugin", default_offload_plugin])
    if default_offload_secret:
        args.extend(["--default-offload-secret", default_offload_secret])
    if default_offload_vendor:
        args.extend(["--default-offload-vendor", default_offload_vendor])
    if inventory_url:
        args.extend(["--inventory-url", inventory_url])

    return await run_kubectl_mtv_command(args)


async def _delete_network_mapping(mapping_name: str, namespace: str) -> str:
    """Delete network mapping implementation."""
    args = ["delete", "mapping", "network", mapping_name]

    if namespace:
        args.extend(["-n", namespace])

    return await run_kubectl_mtv_command(args)


async def _delete_storage_mapping(mapping_name: str, namespace: str) -> str:
    """Delete storage mapping implementation."""
    args = ["delete", "mapping", "storage", mapping_name]

    if namespace:
        args.extend(["-n", namespace])

    return await run_kubectl_mtv_command(args)


async def _patch_network_mapping(
    mapping_name: str,
    namespace: str,
    add_pairs: str,
    update_pairs: str,
    remove_pairs: str,
    inventory_url: str,
) -> str:
    """Patch network mapping implementation."""
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


async def _patch_storage_mapping(
    mapping_name: str,
    namespace: str,
    add_pairs: str,
    update_pairs: str,
    remove_pairs: str,
    inventory_url: str,
    default_volume_mode: str = "",
    default_access_mode: str = "",
    default_offload_plugin: str = "",
    default_offload_secret: str = "",
    default_offload_vendor: str = "",
) -> str:
    """Patch storage mapping implementation with enhanced options."""
    args = ["patch", "mapping", "storage", mapping_name]

    if namespace:
        args.extend(["-n", namespace])

    if add_pairs:
        args.extend(["--add-pairs", add_pairs])
    if update_pairs:
        args.extend(["--update-pairs", update_pairs])
    if remove_pairs:
        args.extend(["--remove-pairs", remove_pairs])
    if default_volume_mode:
        args.extend(["--default-volume-mode", default_volume_mode])
    if default_access_mode:
        args.extend(["--default-access-mode", default_access_mode])
    if default_offload_plugin:
        args.extend(["--default-offload-plugin", default_offload_plugin])
    if default_offload_secret:
        args.extend(["--default-offload-secret", default_offload_secret])
    if default_offload_vendor:
        args.extend(["--default-offload-vendor", default_offload_vendor])
    if inventory_url:
        args.extend(["--inventory-url", inventory_url])

    return await run_kubectl_mtv_command(args)


# Unified Mapping Operations
@mcp.tool()
async def ManageMapping(
    action: str,
    mapping_type: str,
    mapping_name: str,
    namespace: str = "",
    source_provider: str = "",
    target_provider: str = "",
    pairs: str = "",
    add_pairs: str = "",
    update_pairs: str = "",
    remove_pairs: str = "",
    inventory_url: str = "",
    default_volume_mode: str = "",
    default_access_mode: str = "",
    default_offload_plugin: str = "",
    default_offload_secret: str = "",
    default_offload_vendor: str = "",
) -> str:
    """Manage network and storage mappings with unified operations.

    Unified tool for creating, deleting, and patching both network and storage mappings.
    Mappings define how source resources map to target resources during VM migration.

    Actions:
    - 'create': Create a new mapping
    - 'delete': Delete an existing mapping
    - 'patch': Modify an existing mapping

    Mapping Types:
    - 'network': Network mappings for VM network interfaces
    - 'storage': Storage mappings for VM disk placement

    Action-Specific Parameters:
    Create:
    - source_provider, target_provider (required)
    - pairs (optional, initial mappings)

    Delete:
    - No additional parameters

    Patch:
    - add_pairs, update_pairs, remove_pairs (at least one required)

    Mapping Pairs Format:
    Network pairs: 'source:target-namespace/target-network' or 'source:target-network'
    Storage pairs: 'source:storage-class[;volumeMode=Block|Filesystem][;accessMode=ReadWriteOnce|ReadWriteMany|ReadOnlyMany][;offloadPlugin=vsphere][;offloadSecret=secret-name][;offloadVendor=vantara|ontap|primera3par|pureFlashArray|powerflex|powermax]' (comma-separated pairs, semicolon-separated parameters)
    Special values: 'source:default' (pod networking), 'source:ignored' (skip network)
    Multiple pairs: comma-separated 'pair1,pair2,pair3'

    Network Mapping Constraints:
    - All source networks must be mapped (no source networks can be left unmapped)
    - Pod networking and specific multus targets can only be mapped once (no duplicate targets)
    - Use 'source:ignored' to map networks that don't have suitable targets or are not needed

    Storage Mapping Constraints:
    - All source storages must be mapped (no source storages can be left unmapped)
    - Preferred storage classes have virt annotation, k8s annotation, or "virtualization" in name
    - Virt annotation: storageclass.kubevirt.io/is-default-virt-class=true (highest priority)
    - K8s annotation: storageclass.kubernetes.io/is-default-class=true (fallback if no virt annotation)
    - Name matching: case-insensitive search for "virtualization" in storage class name
    - Selection priority: user-defined > virt annotation > k8s annotation > name match > first available

    Args:
        action: Action to perform - 'create', 'delete', or 'patch'
        mapping_type: Type of mapping - 'network' or 'storage'
        mapping_name: Name of the mapping
        namespace: Kubernetes namespace (optional)
        source_provider: Source provider name (required for create)
        target_provider: Target provider name (required for create)
        pairs: Initial mapping pairs for create (optional)
        add_pairs: Pairs to add during patch (optional)
        update_pairs: Pairs to update during patch (optional)
        remove_pairs: Source names to remove during patch (optional)
        inventory_url: Inventory service URL (optional)
        default_volume_mode: Default volume mode for storage pairs (Filesystem|Block) (optional)
        default_access_mode: Default access mode for storage pairs (ReadWriteOnce|ReadWriteMany|ReadOnlyMany) (optional)
        default_offload_plugin: Default offload plugin type for storage pairs (optional)
            • Supported plugins: vsphere
        default_offload_secret: Default offload plugin secret name for storage pairs (optional)
        default_offload_vendor: Default offload plugin vendor for storage pairs (optional)
            • Supported vendors: vantara, ontap, primera3par, pureFlashArray, powerflex, powermax

    Returns:
        Command output confirming the mapping operation

    Examples:
        # Create network mapping
        ManageMapping("create", "network", "my-net-mapping",
                     source_provider="vsphere-provider", target_provider="openshift-provider",
                     pairs="VM Network:default,Management:mgmt/mgmt-net")

        # Create storage mapping with enhanced features
        ManageMapping("create", "storage", "my-storage-mapping",
                     source_provider="vsphere-provider", target_provider="openshift-provider",
                     pairs="fast-datastore:ocs-storagecluster-ceph-rbd;volumeMode=Block;accessMode=ReadWriteOnce;offloadPlugin=vsphere;offloadVendor=vantara",
                     default_volume_mode="Block")

        # Auto-selection will prioritize storage classes with these annotations:
        # 1. storageclass.kubevirt.io/is-default-virt-class=true (preferred for virtualization)
        # 2. storageclass.kubernetes.io/is-default-class=true (Kubernetes default)
        # 3. Storage classes with "virtualization" in name (e.g., "ocs-virtualization-rbd")
        # 4. First available storage class if none of the above are found

        # Delete mapping
        ManageMapping("delete", "network", "old-mapping")

        # Patch network mapping - add and remove pairs
        ManageMapping("patch", "network", "my-net-mapping",
                     add_pairs="DMZ:dmz-namespace/dmz-net",
                     remove_pairs="OldNetwork,UnusedNetwork")

        # Patch storage mapping with enhanced options
        ManageMapping("patch", "storage", "my-storage-mapping",
                     update_pairs="slow-datastore:standard;volumeMode=Filesystem,fast-datastore:premium;volumeMode=Block;accessMode=ReadWriteOnce",
                     default_offload_plugin="vsphere", default_offload_vendor="ontap")
    """
    # Validate action and mapping type
    valid_actions = ["create", "delete", "patch"]
    valid_types = ["network", "storage"]

    if action not in valid_actions:
        raise KubectlMTVError(
            f"Invalid action '{action}'. Valid actions: {', '.join(valid_actions)}"
        )

    if mapping_type not in valid_types:
        raise KubectlMTVError(
            f"Invalid mapping_type '{mapping_type}'. Valid types: {', '.join(valid_types)}"
        )

    # Validate action-specific required parameters
    if action == "create":
        if not source_provider or not target_provider:
            raise KubectlMTVError(
                "source_provider and target_provider are required for create action"
            )
    elif action == "patch":
        if not (add_pairs or update_pairs or remove_pairs):
            raise KubectlMTVError(
                "At least one of add_pairs, update_pairs, or remove_pairs is required for patch action"
            )

    # Route to appropriate sub-action method
    if action == "create" and mapping_type == "network":
        return await _create_network_mapping(
            mapping_name,
            source_provider,
            target_provider,
            namespace,
            pairs,
            inventory_url,
        )
    elif action == "create" and mapping_type == "storage":
        return await _create_storage_mapping(
            mapping_name,
            source_provider,
            target_provider,
            namespace,
            pairs,
            inventory_url,
            default_volume_mode,
            default_access_mode,
            default_offload_plugin,
            default_offload_secret,
            default_offload_vendor,
        )
    elif action == "delete" and mapping_type == "network":
        return await _delete_network_mapping(mapping_name, namespace)
    elif action == "delete" and mapping_type == "storage":
        return await _delete_storage_mapping(mapping_name, namespace)
    elif action == "patch" and mapping_type == "network":
        return await _patch_network_mapping(
            mapping_name,
            namespace,
            add_pairs,
            update_pairs,
            remove_pairs,
            inventory_url,
        )
    elif action == "patch" and mapping_type == "storage":
        return await _patch_storage_mapping(
            mapping_name,
            namespace,
            add_pairs,
            update_pairs,
            remove_pairs,
            inventory_url,
            default_volume_mode,
            default_access_mode,
            default_offload_plugin,
            default_offload_secret,
            default_offload_vendor,
        )


@mcp.tool()
async def CreatePlan(
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
    delete_vm_on_fail_migration: bool = False,
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
    inventory_url: str = "",
    default_volume_mode: str = "",
    default_access_mode: str = "",
    default_offload_plugin: str = "",
    default_offload_secret: str = "",
    default_offload_vendor: str = "",
) -> str:
    """Create a new migration plan with comprehensive configuration options.

    Migration plans define which VMs to migrate and all the configuration for how they should be migrated.
    Plans coordinate providers, mappings, VM selection, and migration behavior.

    Automatic Behaviors:
    - Target provider: If not specified, uses first available OpenShift provider automatically
    - Target namespace: If not specified, uses the plan's namespace
    - Network/Storage mappings: Auto-created if not provided or specified as pairs (except for conversion-only migrations which skip storage mapping creation)
    - VM validation: All VMs are validated against provider inventory before plan creation
    - Missing VM handling: VMs not found in provider are automatically removed with warnings

    VM Selection Options:
    - vms: Comma-separated VM names OR @filename for YAML/JSON file with VM structures
    - Both approaches support automatic ID resolution from provider inventory

    VM Selection Examples:
    - Comma-separated: "web-server-01,database-02,cache-03"
    - File-based: "@vm-list.yaml" or "@vm-list.json"

    File Format (@filename):
    Files can contain VM structures in YAML or JSON format. VM IDs are optional and will be
    auto-resolved from inventory if not provided:

    YAML format (minimal - names only):
    - name: vm1
    - name: vm2
    - name: vm3

    YAML format (with IDs - from planvms output):
    - name: vm1
      id: vm-123
    - name: vm2
      id: vm-456

    JSON format (equivalent):
    [
      {"name": "vm1"},
      {"name": "vm2", "id": "vm-456"}
    ]

    Integration with read tools:
    1. Use ListInventory("vm", "provider", output_format="planvms") to get complete VM structures
    2. Save the YAML output to a file: vm-list.yaml
    3. Edit file to select desired VMs (optional)
    4. Use file: vms="@vm-list.yaml"

    Alternative: Create minimal files with just VM names, IDs will be auto-resolved

    Migration Types:
    - cold: VMs are shut down during migration (default, most reliable)
    - warm: Initial copy while VM runs, brief downtime for final sync
    - live: Minimal downtime migration (advanced, limited compatibility)
    - conversion: Only perform guest OS conversion without disk transfer (storage mappings not allowed)

    Note: Both migration_type and warm parameters are supported. If both are specified,
    migration_type takes precedence over the warm flag.

    Conversion-Only Migration Constraints:
    - Cannot use storage_mapping or storage_pairs parameters
    - Storage mapping will be empty in the resulting plan
    - Only network mapping is created/used for VM networking configuration

    Target Power State Options:
    - on: Start VMs after migration
    - off: Leave VMs stopped after migration
    - auto: Match source VM power state (default)

    Template Variables:
    Templates support Go template syntax with different variables for each template type:

    PVC Name Template Variables:
    - {{.VmName}} - VM name
    - {{.PlanName}} - Migration plan name
    - {{.DiskIndex}} - Initial volume index of the disk
    - {{.WinDriveLetter}} - Windows drive letter (lowercase, requires guest agent)
    - {{.RootDiskIndex}} - Index of the root disk
    - {{.Shared}} - True if volume is shared by multiple VMs
    - {{.FileName}} - Source file name (vSphere only, requires guest agent)

    Volume Name Template Variables:
    - {{.PVCName}} - Name of the PVC mounted to the VM
    - {{.VolumeIndex}} - Sequential index of volume interface (0-based)

    Network Name Template Variables:
    - {{.NetworkName}} - Multus network attachment definition name (if applicable)
    - {{.NetworkNamespace}} - Namespace of network attachment definition (if applicable)
    - {{.NetworkType}} - Network type ("Multus" or "Pod")
    - {{.NetworkIndex}} - Sequential index of network interface (0-based)

    Template Examples:
    - PVC: "{{.VmName}}-disk-{{.DiskIndex}}" → "web-server-01-disk-0"
    - PVC: "{{if eq .DiskIndex .RootDiskIndex}}root{{else}}data{{end}}-{{.DiskIndex}}" → "root-0"
    - PVC: "{{if .Shared}}shared-{{end}}{{.VmName}}-{{.DiskIndex}}" → "shared-web-server-01-0"
    - Volume: "disk-{{.VolumeIndex}}" → "disk-0"
    - Volume: "pvc-{{.PVCName}}" → "pvc-web-server-01-disk-0"
    - Network: "net-{{.NetworkIndex}}" → "net-0"
    - Network: "{{if eq .NetworkType \"Pod\"}}pod{{else}}multus-{{.NetworkIndex}}{{end}}" → "pod"

    Available Template Functions:
    Templates support Go text template syntax including the following built-in functions:

    String Functions:
    - lower: Converts string to lowercase → {{ lower "TEXT" }} → text
    - upper: Converts string to uppercase → {{ upper "text" }} → TEXT
    - contains: Checks if string contains substring → {{ contains "hello" "lo" }} → true
    - replace: Replaces occurrences in a string → {{"I Am Henry VIII" | replace " " "-"}} → I-Am-Henry-VIII
    - trim: Removes whitespace from both ends → {{ trim "  text  " }} → text
    - trimAll: Removes specified characters from both ends → {{ trimAll "$" "$5.00$" }} → 5.00
    - trimSuffix: Removes suffix if present → {{ trimSuffix ".go" "file.go" }} → file
    - trimPrefix: Removes prefix if present → {{ trimPrefix "go." "go.file" }} → file
    - title: Converts to title case → {{ title "hello world" }} → Hello World
    - untitle: Converts to lowercase → {{ untitle "Hello World" }} → hello world
    - repeat: Repeats string n times → {{ repeat 3 "abc" }} → abcabcabc
    - substr: Extracts substring from start to end → {{ substr 1 4 "abcdef" }} → bcd
    - nospace: Removes all whitespace → {{ nospace "a b  c" }} → abc
    - trunc: Truncates string to specified length → {{ trunc 3 "abcdef" }} → abc
    - initials: Extracts first letter of each word → {{ initials "John Doe" }} → JD
    - hasPrefix: Checks if string starts with prefix → {{ hasPrefix "go" "golang" }} → true
    - hasSuffix: Checks if string ends with suffix → {{ hasSuffix "ing" "coding" }} → true
    - mustRegexReplaceAll: Replaces matches using regex → {{ mustRegexReplaceAll "a(x*)b" "-ab-axxb-" "${1}W" }} → -W-xxW-

    Math Functions:
    - add: Sum numbers → {{ add 1 2 3 }} → 6
    - add1: Increment by 1 → {{ add1 5 }} → 6
    - sub: Subtract second number from first → {{ sub 5 3 }} → 2
    - div: Integer division → {{ div 10 3 }} → 3
    - mod: Modulo operation → {{ mod 10 3 }} → 1
    - mul: Multiply numbers → {{ mul 2 3 4 }} → 24
    - max: Return largest integer → {{ max 1 5 3 }} → 5
    - min: Return smallest integer → {{ min 1 5 3 }} → 1
    - floor: Round down to nearest integer → {{ floor 3.75 }} → 3.0
    - ceil: Round up to nearest integer → {{ ceil 3.25 }} → 4.0
    - round: Round to specified decimal places → {{ round 3.75159 2 }} → 3.75

    Template Function Examples:
    - PVC with filename processing: "{{.FileName | trimSuffix \".vmdk\" | replace \"_\" \"-\" | lower}}"
    - PVC with conditional formatting: "{{if .Shared}}shared-{{else}}{{.VmName | lower}}-{{end}}disk-{{.DiskIndex}}"
    - Volume with uppercase naming: "{{.VmName | upper}}-VOL-{{.VolumeIndex}}"

    Args:
        plan_name: Name for the new migration plan (required)
        source_provider: Name of the source provider to migrate from (required). Supports namespace/name pattern (e.g., 'other-namespace/my-provider') to reference providers in different namespaces, defaults to plan namespace if not specified.
        namespace: Kubernetes namespace to create the plan in (optional)
        target_provider: Name of the target provider to migrate to (optional, auto-detects first OpenShift provider if not specified). Supports namespace/name pattern (e.g., 'other-namespace/my-provider') to reference providers in different namespaces, defaults to plan namespace if not specified.
        network_mapping: Name of existing network mapping to use (optional, auto-created if not provided)
        storage_mapping: Name of existing storage mapping to use (optional, auto-created if not provided)
        network_pairs: Network mapping pairs (optional, creates mapping if provided) - supports multiple formats:
            • 'source:target-namespace/target-network' - explicit namespace/name format
            • 'source:target-network' - uses plan namespace if no namespace specified
            • 'source:default' - maps to pod networking
            • 'source:ignored' - ignores the source network
            Note: All source networks must be mapped, pod/multus targets can only be used once
        storage_pairs: Storage mapping pairs (optional, creates mapping if provided) - enhanced format with optional parameters:
            • Basic: 'source:storage-class' - simple storage class mapping
            • Enhanced: 'source:storage-class;volumeMode=Block;accessMode=ReadWriteOnce;offloadPlugin=vsphere;offloadSecret=secret;offloadVendor=vantara'
            • All semicolon-separated parameters are optional: volumeMode, accessMode, offloadPlugin, offloadSecret, offloadVendor
            Note: All source storages must be mapped, auto-selection uses
            storageclass.kubevirt.io/is-default-virt-class > storageclass.kubernetes.io/is-default-class >
            name with "virtualization"
        vms: VM names (comma-separated) or @filename for YAML/JSON file with VM structures (optional)
        pre_hook: Pre-migration hook to add to all VMs (optional)
        post_hook: Post-migration hook to add to all VMs (optional)
        description: Plan description (optional)
        target_namespace: Target namespace for migrated VMs (optional, defaults to plan namespace)
        transfer_network: Network attachment definition for VM data transfer - supports 'namespace/network-name' or just 'network-name' (uses plan namespace) (optional)
        preserve_cluster_cpu_model: Preserve CPU model and flags from oVirt cluster (optional, default False)
        preserve_static_ips: Preserve static IPs of vSphere VMs (optional, default False)
        pvc_name_template: Template for generating PVC names for VM disks (optional)
        volume_name_template: Template for generating volume interface names (optional)
        network_name_template: Template for generating network interface names (optional)
        migrate_shared_disks: Whether to migrate shared disks (optional, default True, auto-patched if False)
        archived: Whether plan should be archived (optional, default False)
        pvc_name_template_use_generate_name: Use generateName for PVC template (optional, default True, auto-patched if False)
        delete_guest_conversion_pod: Delete conversion pod after migration (optional, default False)
        skip_guest_conversion: Skip guest conversion process (optional, default False)
        use_compatibility_mode: Use compatibility devices when skipping conversion (optional, default True, auto-patched if False)
        install_legacy_drivers: Install legacy Windows drivers - 'true'/'false' (optional)
        migration_type: Migration type - 'cold', 'warm', 'live', or 'conversion' (optional). Note: 'conversion' type cannot be used with storage_mapping or storage_pairs
        default_target_network: Default target network - 'default' for pod networking, 'namespace/network-name', or just 'network-name' (uses plan namespace) (optional)
        default_target_storage_class: Default target storage class (optional)
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
        default_volume_mode: Default volume mode for storage pairs (Filesystem|Block) (optional)
        default_access_mode: Default access mode for storage pairs (ReadWriteOnce|ReadWriteMany|ReadOnlyMany) (optional)
        default_offload_plugin: Default offload plugin type for storage pairs (optional)
            • Supported plugins: vsphere
        default_offload_secret: Default offload plugin secret name for storage pairs (optional)
        default_offload_vendor: Default offload plugin vendor for storage pairs (optional)
            • Supported vendors: vantara, ontap, primera3par, pureFlashArray, powerflex, powermax

    Returns:
        Command output confirming plan creation

    Examples:
        # Create basic plan (auto-detects target provider, creates mappings)
        create_plan("my-plan", "vsphere-provider")

        # Create plan with providers from different namespaces
        create_plan("my-plan", "source-ns/vsphere-provider",
                   target_provider="target-ns/openshift-provider",
                   namespace="demo")

        # Create comprehensive plan showing optional parameters and namespace/name syntax
        create_plan("my-plan", "vsphere-provider",
                   target_provider="openshift-target",
                   target_namespace="migrated-vms",
                   vms="vm1,vm2,vm3",
                   migration_type="warm",
                   # Network pairs: shows different formats (namespace/name, name-only, default, ignored)
                   network_pairs="VM Network:default,Management:mgmt-ns/mgmt-net,Production:prod-net,DMZ:ignored",
                   # Storage pairs: shows basic and enhanced format with optional parameters
                   storage_pairs="fast-datastore:premium-ssd,slow-datastore:standard-hdd;volumeMode=Block;accessMode=ReadWriteOnce;offloadPlugin=vsphere;offloadVendor=vantara",
                   default_volume_mode="Block",
                   default_target_network="openshift-sriov-network/high-perf-net",
                   transfer_network="sriov-namespace/transfer-network",
                   target_power_state="on",
                   description="Production VM migration showing optional parameters and formats")

        # Plan with comma-separated VM names
        create_plan("my-plan", "vsphere-provider",
                   vms="vm1,vm2,vm3")  # VMs identified by name, IDs auto-resolved

        # Plan with file-based VM selection (from planvms output)
        create_plan("my-plan", "vsphere-provider",
                   vms="@vm-list.yaml",  # from ListInventory("vm", "provider", output_format="planvms")
                   network_mapping="existing-net-map",
                   storage_mapping="existing-storage-map")

        # Plan with minimal file (names only, IDs auto-resolved)
        create_plan("my-plan", "vsphere-provider",
                   vms="@vm-names.yaml")  # YAML with just VM names

        # Create plan with KARL affinity rules
        create_plan("db-plan", "vsphere-provider",
                   vms="database-vm",
                   target_affinity="REQUIRE pods(app=database) on node",
                   description="Co-locate with existing database pods")
    """
    # Validate that conversion-only migrations don't use storage mappings
    if migration_type == "conversion":
        if storage_mapping:
            raise KubectlMTVError(
                "Cannot use storage_mapping with migration_type 'conversion'. "
                "Conversion-only migrations require empty storage mapping."
            )
        if storage_pairs:
            raise KubectlMTVError(
                "Cannot use storage_pairs with migration_type 'conversion'. "
                "Conversion-only migrations require empty storage mapping."
            )

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
    if default_volume_mode:
        args.extend(["--default-volume-mode", default_volume_mode])
    if default_access_mode:
        args.extend(["--default-access-mode", default_access_mode])
    if default_offload_plugin:
        args.extend(["--default-offload-plugin", default_offload_plugin])
    if default_offload_secret:
        args.extend(["--default-offload-secret", default_offload_secret])
    if default_offload_vendor:
        args.extend(["--default-offload-vendor", default_offload_vendor])
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
    if delete_vm_on_fail_migration:
        args.append("--delete-vm-on-fail-migration")
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
async def CreateHost(
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
    inventory_url: str = "",
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
async def CreateHook(
    hook_name: str,
    image: str,
    namespace: str = "",
    service_account: str = "",
    playbook: str = "",
    deadline: int = 0,
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
async def DeleteProvider(
    provider_name: str = "", namespace: str = "", all_providers: bool = False
) -> str:
    """Delete one or more providers.

    WARNING: This will remove providers and may affect associated plans and mappings.

    Args:
        provider_name: Name of the provider to delete (required unless all_providers=True)
        namespace: Kubernetes namespace containing the provider (optional)
        all_providers: Delete all providers in the namespace (optional)

    Returns:
        Command output confirming provider deletion

    Examples:
        # Delete specific provider
        DeleteProvider("my-provider")

        # Delete all providers in namespace
        DeleteProvider(all_providers=True, namespace="demo")
    """
    args = ["delete", "provider"]

    if all_providers:
        args.append("--all")
    else:
        if not provider_name:
            raise ValueError("provider_name is required when all_providers=False")
        args.append(provider_name)

    if namespace:
        args.extend(["-n", namespace])

    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def DeletePlan(
    plan_name: str = "",
    namespace: str = "",
    all_plans: bool = False,
    skip_archive: bool = False,
    clean_all: bool = False,
) -> str:
    """Delete one or more migration plans.

    WARNING: This will remove migration plans and all associated migration data.

    By default, plans are archived before deletion to ensure a clean shutdown. Use skip_archive
    to delete immediately without archiving. Use clean_all to archive, enable VM deletion on
    failed migration, then delete.

    Args:
        plan_name: Name of the plan to delete (required unless all_plans=True)
        namespace: Kubernetes namespace containing the plan (optional)
        all_plans: Delete all plans in the namespace (optional)
        skip_archive: Skip archiving and delete immediately (optional)
        clean_all: Archive, delete VMs on failed migration, then delete (optional)

    Returns:
        Command output confirming plan deletion

    Examples:
        # Delete specific plan with default archiving
        DeletePlan("my-plan")

        # Delete plan without archiving
        DeletePlan("my-plan", skip_archive=True)

        # Delete plan with VM cleanup on failure
        DeletePlan("my-plan", clean_all=True)

        # Delete all plans in namespace
        DeletePlan(all_plans=True, namespace="demo")
    """
    args = ["delete", "plan"]

    if all_plans:
        args.append("--all")
    else:
        if not plan_name:
            raise ValueError("plan_name is required when all_plans=False")
        args.append(plan_name)

    if namespace:
        args.extend(["-n", namespace])

    if skip_archive:
        args.append("--skip-archive")

    if clean_all:
        args.append("--clean-all")

    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def DeleteHost(
    host_name: str = "", namespace: str = "", all_hosts: bool = False
) -> str:
    """Delete one or more migration hosts.

    WARNING: This will remove migration hosts.

    Args:
        host_name: Name of the host to delete (required unless all_hosts=True)
        namespace: Kubernetes namespace containing the host (optional)
        all_hosts: Delete all hosts in the namespace (optional)

    Returns:
        Command output confirming host deletion

    Examples:
        # Delete specific host
        DeleteHost("esxi-host-01")

        # Delete all hosts in namespace
        DeleteHost(all_hosts=True, namespace="demo")
    """
    args = ["delete", "host"]

    if all_hosts:
        args.append("--all")
    else:
        if not host_name:
            raise ValueError("host_name is required when all_hosts=False")
        args.append(host_name)

    if namespace:
        args.extend(["-n", namespace])

    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def DeleteHook(
    hook_name: str = "", namespace: str = "", all_hooks: bool = False
) -> str:
    """Delete one or more migration hooks.

    WARNING: This will remove migration hooks.

    Args:
        hook_name: Name of the hook to delete (required unless all_hooks=True)
        namespace: Kubernetes namespace containing the hook (optional)
        all_hooks: Delete all hooks in the namespace (optional)

    Returns:
        Command output confirming hook deletion

    Examples:
        # Delete specific hook
        DeleteHook("pre-migration-check")

        # Delete all hooks in namespace
        DeleteHook(all_hooks=True, namespace="demo")
    """
    args = ["delete", "hook"]

    if all_hooks:
        args.append("--all")
    else:
        if not hook_name:
            raise ValueError("hook_name is required when all_hooks=False")
        args.append(hook_name)

    if namespace:
        args.extend(["-n", namespace])

    return await run_kubectl_mtv_command(args)


# Patch Operations
@mcp.tool()
async def PatchProvider(
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
    provider_region_name: str = "",
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
async def PatchPlan(
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
    delete_vm_on_fail_migration: bool = None,
    skip_guest_conversion: bool = None,
    warm: bool = None,
) -> str:
    """Patch/modify various fields of an existing migration plan without modifying its VMs.

    This allows updating plan configuration without recreating the entire plan. You can modify
    individual plan properties while preserving the VM list and other unchanged settings.

    Boolean Parameters:
    - None (default): Don't change the current value
    - True: Set to true
    - False: Set to false

    Migration Types:
    - cold: Traditional migration with VM shutdown (most reliable)
    - warm: Warm migration with reduced downtime (initial copy while VM runs)
    - live: Minimal downtime migration (advanced, limited compatibility)
    - conversion: Only perform guest OS conversion without disk transfer

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
        transfer_network: Network to use for transferring VM data - supports 'namespace/network-name' or just 'network-name' (uses plan namespace) (optional)
        install_legacy_drivers: Install legacy drivers - 'true', 'false', or empty for auto (optional)
        migration_type: Migration type - 'cold', 'warm', 'live', or 'conversion' (optional)
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
        delete_vm_on_fail_migration: Delete target VM when migration fails (optional)
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
        args.extend(
            ["--preserve-cluster-cpu-model", str(preserve_cluster_cpu_model).lower()]
        )
    if preserve_static_ips is not None:
        args.extend(["--preserve-static-ips", str(preserve_static_ips).lower()])
    if migrate_shared_disks is not None:
        args.extend(["--migrate-shared-disks", str(migrate_shared_disks).lower()])
    if archived is not None:
        args.extend(["--archived", str(archived).lower()])
    if pvc_name_template_use_generate_name is not None:
        args.extend(
            [
                "--pvc-name-template-use-generate-name",
                str(pvc_name_template_use_generate_name).lower(),
            ]
        )
    if delete_guest_conversion_pod is not None:
        args.extend(
            ["--delete-guest-conversion-pod", str(delete_guest_conversion_pod).lower()]
        )
    if delete_vm_on_fail_migration is not None:
        args.extend(
            ["--delete-vm-on-fail-migration", str(delete_vm_on_fail_migration).lower()]
        )
    if skip_guest_conversion is not None:
        args.extend(["--skip-guest-conversion", str(skip_guest_conversion).lower()])
    if warm is not None:
        args.extend(["--warm", str(warm).lower()])

    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def PatchPlanVm(
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
    clear_hooks: bool = False,
    delete_vm_on_fail_migration: bool = None,
) -> str:
    """Patch VM-specific fields for a VM within a migration plan's VM list.

    This allows you to customize individual VM settings within a plan without affecting other VMs.
    Useful for setting VM-specific configurations like custom names, storage templates, hooks, or LUKS decryption.

    Template Variables:
    Templates support Go template syntax with different variables for each template type:

    PVC Name Template Variables:
    - {{.VmName}} - VM name
    - {{.PlanName}} - Migration plan name
    - {{.DiskIndex}} - Initial volume index of the disk
    - {{.WinDriveLetter}} - Windows drive letter (lowercase, requires guest agent)
    - {{.RootDiskIndex}} - Index of the root disk
    - {{.Shared}} - True if volume is shared by multiple VMs
    - {{.FileName}} - Source file name (vSphere only, requires guest agent)

    Volume Name Template Variables:
    - {{.PVCName}} - Name of the PVC mounted to the VM
    - {{.VolumeIndex}} - Sequential index of volume interface (0-based)

    Network Name Template Variables:
    - {{.NetworkName}} - Multus network attachment definition name (if applicable)
    - {{.NetworkNamespace}} - Namespace of network attachment definition (if applicable)
    - {{.NetworkType}} - Network type ("Multus" or "Pod")
    - {{.NetworkIndex}} - Sequential index of network interface (0-based)

    Available Template Functions:
    Templates support Go text template syntax including the following built-in functions:

    String Functions:
    - lower: Converts string to lowercase → {{ lower "TEXT" }} → text
    - upper: Converts string to uppercase → {{ upper "text" }} → TEXT
    - contains: Checks if string contains substring → {{ contains "hello" "lo" }} → true
    - replace: Replaces occurrences in a string → {{"I Am Henry VIII" | replace " " "-"}} → I-Am-Henry-VIII
    - trim: Removes whitespace from both ends → {{ trim "  text  " }} → text
    - trimAll: Removes specified characters from both ends → {{ trimAll "$" "$5.00$" }} → 5.00
    - trimSuffix: Removes suffix if present → {{ trimSuffix ".go" "file.go" }} → file
    - trimPrefix: Removes prefix if present → {{ trimPrefix "go." "go.file" }} → file
    - title: Converts to title case → {{ title "hello world" }} → Hello World
    - untitle: Converts to lowercase → {{ untitle "Hello World" }} → hello world
    - repeat: Repeats string n times → {{ repeat 3 "abc" }} → abcabcabc
    - substr: Extracts substring from start to end → {{ substr 1 4 "abcdef" }} → bcd
    - nospace: Removes all whitespace → {{ nospace "a b  c" }} → abc
    - trunc: Truncates string to specified length → {{ trunc 3 "abcdef" }} → abc
    - initials: Extracts first letter of each word → {{ initials "John Doe" }} → JD
    - hasPrefix: Checks if string starts with prefix → {{ hasPrefix "go" "golang" }} → true
    - hasSuffix: Checks if string ends with suffix → {{ hasSuffix "ing" "coding" }} → true
    - mustRegexReplaceAll: Replaces matches using regex → {{ mustRegexReplaceAll "a(x*)b" "-ab-axxb-" "${1}W" }} → -W-xxW-

    Math Functions:
    - add: Sum numbers → {{ add 1 2 3 }} → 6
    - add1: Increment by 1 → {{ add1 5 }} → 6
    - sub: Subtract second number from first → {{ sub 5 3 }} → 2
    - div: Integer division → {{ div 10 3 }} → 3
    - mod: Modulo operation → {{ mod 10 3 }} → 1
    - mul: Multiply numbers → {{ mul 2 3 4 }} → 24
    - max: Return largest integer → {{ max 1 5 3 }} → 5
    - min: Return smallest integer → {{ min 1 5 3 }} → 1
    - floor: Round down to nearest integer → {{ floor 3.75 }} → 3.0
    - ceil: Round up to nearest integer → {{ ceil 3.25 }} → 4.0
    - round: Round to specified decimal places → {{ round 3.75159 2 }} → 3.75

    LUKS Secret Usage:
    The luks_secret parameter should reference a Kubernetes Secret containing the actual
    LUKS decryption keys. MTV will use this Secret to decrypt encrypted VM disks during migration.
    The Secret must exist in the same namespace as the migration plan.

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
        luks_secret: Kubernetes Secret name containing LUKS disk decryption keys (optional)
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
        patch_plan_vm("my-plan", "storage-vm", pvc_name_template="{{.VmName}}-disk-{{.DiskIndex}}")
    """
    args = ["patch", "planvm", plan_name, vm_name]

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

    # VM-level options
    if delete_vm_on_fail_migration is not None:
        args.extend(
            ["--delete-vm-on-fail-migration", str(delete_vm_on_fail_migration).lower()]
        )

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
