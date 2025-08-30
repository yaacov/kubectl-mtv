#!/usr/bin/env python3
"""
FastMCP Server for virtctl (KubeVirt Virtual Machine Management)

This server provides comprehensive tools to interact with KubeVirt virtual machines
through virtctl commands. It assumes virtctl is installed and the user is
logged into a Kubernetes cluster with KubeVirt deployed.

## Prerequisites and Setup

BEFORE USING THESE TOOLS:
1. Verify KubeVirt is installed: kubectl get kubevirt -n kubevirt
2. Ensure virtctl is available: virtctl version
3. Check cluster access: kubectl get nodes
4. Verify CDI for DataVolumes: kubectl get cdi -n cdi

## Quick Start Workflow

COMPLETE VM DEPLOYMENT (5 steps):
1. virtctl_cluster_resources("datasources", "all") → Find available OS images
2. virtctl_create_vm_advanced(name="my-vm", volumes={"volume_import": [...]}, infer_instancetype=True) → Create VM
3. virtctl_vm_lifecycle("my-vm", "start") → Start VM
4. virtctl_service_management("expose", "my-vm", expose_config={...}) → Expose as service
5. virtctl_diagnostics("guestosinfo", "my-vm") → Monitor VM health

## Common Use Cases

VM CREATION PATTERNS:
- Development VM: Use u1.small + DataSource inference + SSH keys
- Production VM: Use specific instance types + preferences + access credentials + labels
- Windows VM: Use sysprep volumes + RDP access + Windows preferences
- GPU Workload: Use GPU instance types + special preferences + device passthrough

TROUBLESHOOTING WORKFLOW:
1. virtctl_diagnostics("guestosinfo", "vm") → Check VM state
2. virtctl_image_operations("memory-dump", vm_name="vm") → Debug crashes
3. virtctl_volume_management("vm", "list") → Check storage issues
4. virtctl_cluster_resources("storageclasses") → Verify storage availability

MAINTENANCE WORKFLOW:
1. virtctl_image_operations("memory-dump") → Backup before changes
2. virtctl_vm_lifecycle("vm", "migrate") → Move to maintenance node
3. virtctl_volume_management("vm", "add") → Add temporary storage
4. virtctl_diagnostics → Monitor during maintenance
5. virtctl_volume_management("vm", "remove") → Clean up after maintenance

## Tool Categories (9 tools total)

1. **VM Lifecycle Management** - Start, stop, restart, pause, migrate VMs
2. **VM Creation** - Advanced VM creation with full configuration support
3. **Volume Management** - Hot-plug volume operations on running VMs
4. **Image Operations** - Upload, export, libguestfs, memory dumps
5. **Service Management** - Expose VMs as Kubernetes services
6. **Diagnostics** - Guest OS info, filesystem listing, monitoring
7. **Resource Creation** - Instance types and preferences templates
8. **DataSource Management** - Boot image lifecycle management
9. **Cluster Discovery** - Available resources exploration

## Integration Notes

- **With kubectl-mtv**: VMs migrated via MTV can be managed with these tools
- **Resource Reuse**: DataSources created for MTV are reusable for direct VM creation
- **Cross-tool Workflows**: Use discovery tools → creation tools → lifecycle tools → diagnostics
- **Error Recovery**: Each tool provides detailed error messages and suggested fixes

## Complete Troubleshooting Workflows

**VM Won't Start:**
1. virtctl_cluster_resources("instancetypes") → Check resource availability
2. virtctl_cluster_resources("datasources") → Verify boot image exists
3. virtctl_diagnostics("version") → Check tool compatibility
4. virtctl_vm_lifecycle(vm, "start", dry_run=True) → Test without execution
5. Check kubectl events: kubectl get events --field-selector involvedObject.name={vm}

**VM Network Issues:**
1. virtctl_diagnostics("guestosinfo", vm) → Verify guest agent connectivity
2. virtctl_service_management("expose", vm, expose_config={...}) → Test service creation
3. virtctl_diagnostics("fslist", vm) → Check if VM filesystem is accessible
4. Check pod network: kubectl get pod -l kubevirt.io/vm={vm}

**Storage Problems:**
1. virtctl_volume_management(vm, "list") → Check attached volumes
2. virtctl_cluster_resources("storageclasses") → Verify storage availability
3. virtctl_volume_management(vm, "add", dry_run=True) → Test volume operations
4. Check PVC status: kubectl get pvc -n {namespace}

**Performance Issues:**
1. virtctl_diagnostics("guestosinfo", vm) → Check VM resource usage
2. virtctl_diagnostics("fslist", vm) → Monitor disk usage
3. virtctl_vm_lifecycle(vm, "migrate") → Move to less loaded node
4. virtctl_volume_management → Add more storage if needed

## Cross-Tool Integration Patterns

**Development to Production Pipeline:**
1. virtctl_cluster_resources → Discover available resources
2. virtctl_datasource_management → Prepare OS images
3. virtctl_create_vm_advanced → Create dev VM
4. virtctl_service_management → Expose for testing
5. Scale to production with different instance types

**Maintenance Workflow:**
1. virtctl_image_operations("memory-dump") → Backup VM state
2. virtctl_vm_lifecycle("migrate") → Move to maintenance node
3. virtctl_volume_management("add") → Add maintenance volumes
4. Perform maintenance tasks
5. virtctl_volume_management("remove") → Clean up
6. virtctl_vm_lifecycle("migrate") → Return to original node

**Monitoring and Health Checks:**
- Regular: virtctl_diagnostics("guestosinfo", vm) for health
- Storage: virtctl_diagnostics("fslist", vm) for capacity monitoring
- Network: virtctl_service_management for connectivity testing
- Resources: virtctl_cluster_resources for capacity planning
"""

import os
import subprocess
import json
import shlex
import argparse
from typing import Any, Dict, List
import yaml

from fastmcp import FastMCP

# Get the directory containing this script
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))


def get_current_namespace():
    """Get the current Kubernetes namespace from kubectl context."""
    try:
        result = subprocess.run(
            [
                "kubectl",
                "config",
                "view",
                "--minify",
                "--output",
                "jsonpath={..namespace}",
            ],
            capture_output=True,
            text=True,
            check=True,
        )
        namespace = result.stdout.strip()
        return namespace if namespace else "default"
    except (subprocess.CalledProcessError, FileNotFoundError):
        # Fallback to default if kubectl is not available or context is not set
        return "default"


def get_package_version():
    """Get version from installed package metadata."""
    try:
        from importlib import metadata

        return metadata.version("mtv-mcp")
    except Exception:
        # Fallback if package metadata unavailable
        return "dev-version"


# Initialize the MCP server instance immediately
mcp = FastMCP(name="kubevirt", version=get_package_version())


class VirtctlError(Exception):
    """Custom exception for virtctl command errors."""

    pass


def _format_shell_command(cmd: list[str]) -> str:
    """Format a command list as a properly quoted shell command string.

    Args:
        cmd: List of command arguments

    Returns:
        Properly quoted shell command string
    """
    return " ".join(shlex.quote(arg) for arg in cmd)


def _run_virtctl_command(
    args: list[str],
    capture_output: bool = True,
    check: bool = True,
    input_data: str = None,
    **kwargs,
) -> subprocess.CompletedProcess:
    """Run a virtctl command and return the result.

    Args:
        args: Command arguments (excluding 'virtctl')
        capture_output: Whether to capture stdout/stderr
        check: Whether to raise exception on non-zero exit
        input_data: Optional input data to pipe to command
        **kwargs: Additional subprocess.run arguments

    Returns:
        CompletedProcess result

    Raises:
        VirtctlError: If command fails and check=True
    """
    # Determine virtctl command - check for kubectl virt plugin first
    virtctl_cmd = ["virtctl"] if _has_kubectl_virt_plugin() else ["virtctl"]

    cmd = virtctl_cmd + args

    try:
        kwargs.setdefault("timeout", 120)
        result = subprocess.run(
            cmd,
            capture_output=capture_output,
            text=True,
            input=input_data,
            check=False,  # We'll handle errors ourselves
            **kwargs,
        )

        if check and result.returncode != 0:
            error_msg = f"Command failed: {_format_shell_command(cmd)}"
            if result.stderr:
                error_msg += f"\nError: {result.stderr.strip()}"
            if result.stdout:
                error_msg += f"\nOutput: {result.stdout.strip()}"
            raise VirtctlError(error_msg)

        return result

    except FileNotFoundError:
        cmd_str = _format_shell_command(cmd)
        raise VirtctlError(
            f"Command not found: {cmd_str}. Please ensure virtctl is installed or kubectl virt plugin is available."
        )
    except Exception as e:
        cmd_str = _format_shell_command(cmd)
        raise VirtctlError(f"Error running command {cmd_str}: {str(e)}")


def _has_kubectl_virt_plugin() -> bool:
    """Check if kubectl virt plugin is available."""
    try:
        result = subprocess.run(
            ["virtctl", "version"], capture_output=True, text=True, timeout=120
        )
        return result.returncode == 0
    except Exception:
        return False


def _parse_json_output(output: str) -> Any:
    """Parse JSON output from virtctl command.

    Args:
        output: JSON string output

    Returns:
        Parsed JSON object

    Raises:
        VirtctlError: If output is not valid JSON
    """
    if not output.strip():
        return {}

    try:
        return json.loads(output)
    except json.JSONDecodeError as e:
        raise VirtctlError(f"Failed to parse JSON output: {str(e)}")


# VM Lifecycle Management Tools
@mcp.tool()
def virtctl_vm_lifecycle(
    vm_name: str,
    operation: str,
    namespace: str = None,
    grace_period: int = 0,
    force: bool = False,
    dry_run: bool = False,
    node_name: str = "",
    timeout: str = "",
) -> str:
    """Unified VM power state management for virtual machines.

    PREREQUISITES:
    - VM must exist: kubectl get vm {vm_name} -n {namespace}
    - For migration: Ensure live migration is enabled and nodes are ready
    - For stop: Check if VM has important processes running

    OPERATIONS GUIDE:

    START: Boot a stopped VM
    - Best for: Cold starts, post-maintenance restarts
    - Prerequisites: VM must be in "Stopped" state
    - Time: Usually 30-120 seconds depending on OS

    STOP: Graceful VM shutdown
    - Best for: Planned maintenance, clean shutdowns
    - Use grace_period for applications that need time to close
    - Use force=True only when VM is unresponsive

    RESTART: Stop then start (maintains VM configuration)
    - Best for: Applying configuration changes, troubleshooting
    - Safer than stop/start sequence as it's atomic

    MIGRATE: Live migration to different node
    - Best for: Node maintenance, load balancing, hardware issues
    - Prerequisites: Both nodes must support live migration
    - Requires shared storage accessible from both nodes

    PAUSE/UNPAUSE: Freeze/resume VM state
    - Best for: Temporary resource reclamation, debugging
    - VM memory stays allocated but CPU stops
    - Much faster than stop/start

    SOFT-REBOOT: Guest OS reboot without stopping VM
    - Best for: OS updates, configuration changes
    - Requires guest agent running in VM
    - Faster than restart operation

    TROUBLESHOOTING:

    Common Issues:
    - "VM not found": Check vm_name and namespace spelling
    - "Migration failed": Verify shared storage and network connectivity
    - "Stop timeout": Increase grace_period or use force=True
    - "Start failed": Check resource availability and image accessibility

    Recommended Grace Periods:
    - Database VMs: 60-300 seconds (data consistency)
    - Web servers: 30-60 seconds (connection draining)
    - Development VMs: 10-30 seconds (minimal data)
    - Windows VMs: 120-300 seconds (slower shutdown)

    Pre-operation Checks:
    - VM status: kubectl get vm {vm_name} -n {namespace}
    - Resource usage: kubectl top pod -n {namespace}
    - Node health: kubectl get nodes
    - Storage: kubectl get pvc -n {namespace}

    Args:
        vm_name: Name of the virtual machine
        operation: Lifecycle operation (start, stop, restart, pause, unpause, migrate, soft-reboot)
        namespace: Kubernetes namespace containing the VM (optional)
        grace_period: Graceful termination period in seconds (optional)
        force: Force the operation (optional)
        dry_run: Show what would be done without executing (optional)
        node_name: Target node name for migrate operation (optional)
        timeout: Operation timeout (optional)

    Returns:
        Command output confirming the lifecycle operation

    Workflow Examples:
        # Production VM restart sequence
        virtctl_vm_lifecycle("prod-db", "stop", "production", grace_period=180)
        # Wait for clean shutdown, then:
        virtctl_vm_lifecycle("prod-db", "start", "production")

        # Node maintenance migration
        virtctl_vm_lifecycle("critical-vm", "migrate", "production",
                           node_name="worker-3", dry_run=True)  # Test first
        virtctl_vm_lifecycle("critical-vm", "migrate", "production",
                           node_name="worker-3")  # Execute

        # Emergency shutdown
        virtctl_vm_lifecycle("stuck-vm", "stop", "production",
                           grace_period=10, force=True)

        # Development cycle
        virtctl_vm_lifecycle("dev-vm", "pause", "development")    # Free resources
        # Later...
        virtctl_vm_lifecycle("dev-vm", "unpause", "development")  # Resume work
    """
    # Build base command
    args = []

    # Map operations to virtctl commands
    operation_map = {
        "start": "start",
        "stop": "stop",
        "restart": "restart",
        "pause": "pause",
        "unpause": "unpause",
        "migrate": "migrate",
        "soft-reboot": "soft-reboot",
    }

    if operation not in operation_map:
        raise VirtctlError(
            f"Invalid operation: {operation}. Valid operations: {list(operation_map.keys())}"
        )

    args.append(operation_map[operation])
    args.append(vm_name)

    # Use current namespace if none provided
    if namespace is None:
        namespace = get_current_namespace()

    # Add namespace
    if namespace:
        args.extend(["-n", namespace])

    # Add operation-specific arguments
    if operation in ["stop", "restart"] and grace_period > 0:
        args.extend(["--grace-period", str(grace_period)])

    if operation in ["stop", "restart"] and force:
        args.append("--force")

    if operation == "migrate" and node_name:
        args.extend(["--node", node_name])

    if dry_run:
        args.append("--dry-run")

    if timeout:
        args.extend(["--timeout", timeout])

    result = _run_virtctl_command(args)

    return json.dumps(
        {
            "status": "success",
            "message": "VM lifecycle operation completed successfully",
            "command": _format_shell_command(["virtctl"] + args),
            "output": result.stdout.strip(),
        },
        indent=2,
    )


# VM Creation Tools
@mcp.tool()
def virtctl_create_vm_advanced(
    name: str = "",
    namespace: str = None,
    instancetype: str = "",
    preference: str = "",
    run_strategy: str = "",
    volumes: Dict[str, Any] = None,
    cloud_init: Dict[str, Any] = None,
    resource_requirements: Dict[str, str] = None,
    infer_instancetype: bool = False,
    infer_preference: bool = False,
    infer_instancetype_from: str = "",
    infer_preference_from: str = "",
    access_credentials: List[Dict[str, Any]] = None,
    termination_grace_period: int = 0,
    generate_name: bool = False,
) -> str:
    """Comprehensive VM creation with full configuration support.

    This advanced tool supports most virtctl create vm capabilities:
    - Instance types and preferences with automatic inference
    - Multiple volume sources (PVCs, DataVolumes, container disks, blank volumes)
    - Complete cloud-init integration with user data, SSH keys, passwords
    - Resource requirements and limits
    - Access credentials for SSH and password injection

    Args:
        name: VM name (optional, random if not specified)
        namespace: Kubernetes namespace (optional)
        instancetype: Instance type name (optional)
        preference: Preference name (optional)
        run_strategy: Run strategy (Always, RerunOnFailure, Manual, Halted) (optional)
        volumes: Volume configuration dictionary with volume types (optional)
        cloud_init: Cloud-init configuration (optional)
        resource_requirements: CPU/memory requirements (optional)
        infer_instancetype: Infer instance type from boot volume (optional)
        infer_preference: Infer preference from boot volume (optional)
        infer_instancetype_from: Volume name to infer instance type from (optional)
        infer_preference_from: Volume name to infer preference from (optional)
        access_credentials: List of access credential configurations (optional)
        termination_grace_period: Grace period for VM termination in seconds (optional)
        generate_name: Use generateName instead of name (optional)

    Returns:
        Raw YAML manifest from virtctl create vm command

    Volume Configuration Format:
        volumes = {
            "volume_import": [{
                "type": "pvc",  # pvc, dv, ds (DataSource), blank
                "src": "fedora-base",  # Source name/path (not needed for blank)
                "name": "rootdisk",  # Volume name
                "size": "20Gi",  # Volume size (required for blank, optional for others)
                "storage_class": "fast-ssd",  # Storage class (optional)
                "namespace": "default",  # Source namespace (optional)
                "bootorder": 1  # Boot order (optional)
            }],
            "volume_containerdisk": [{
                "src": "quay.io/containerdisks/fedora:latest",
                "name": "containerdisk",
                "bootorder": 2
            }],

            "volume_pvc": [{
                "src": "existing-pvc",
                "name": "mounted-disk"
            }],
            "volume_sysprep": [{
                "name": "sysprep-config",
                "src": "windows-config",
                "type": "configmap"  # configmap or secret
            }]
        }

    Cloud-Init Configuration:
        cloud_init = {
            "user": "fedora",  # Default user (will be created if doesn't exist)
            "ssh_key": "ssh-rsa AAAA...",  # SSH public key (from ~/.ssh/id_rsa.pub)
            "password_file": "/path/to/password",  # Password file path
            "ga_manage_ssh": True,  # Enable guest agent SSH key management
            "user_data": "#cloud-config\\npackages:\\n  - git",  # Custom user data
            "user_data_base64": "I2Nsb3VkLWNvbmZpZw==",  # Base64 encoded user data
        }

    ACCESS CONTROL SETUP:

    SSH Key Authentication (Recommended):
    1. Generate SSH key pair: ssh-keygen -t rsa -b 4096 -C "vm-access"
    2. Use public key content in ssh_key field: cat ~/.ssh/id_rsa.pub
    3. Enable guest agent SSH management for dynamic key rotation
    4. Create Kubernetes secret for key storage: kubectl create secret generic vm-keys --from-file=authorized_keys=~/.ssh/id_rsa.pub

    Password Authentication:
    1. Create password file: echo "mypassword" > /tmp/vm-password
    2. Or use Kubernetes secret: kubectl create secret generic vm-passwords --from-literal=password=mypassword
    3. Reference in password_file parameter or use access_credentials

    User Management:
    - Default user gets sudo privileges on most cloud images
    - User will be created if it doesn't exist in the image
    - Consider using standard usernames: ubuntu, fedora, centos, cloud-user, admin

    Access Credentials (Advanced):
        access_credentials = [{
            "type": "ssh",  # ssh or password
            "src": "my-keys",  # Kubernetes Secret name containing keys/passwords
            "user": "myuser",  # Target user (defaults to cloud-init user)
            "method": "qemu-guest-agent"  # qemu-guest-agent (dynamic) or cloud-init (static)
        }]

    Guest Agent SSH Management:
    - Allows dynamic SSH key injection/rotation without VM restart
    - Requires qemu-guest-agent installed in VM image
    - Enables SELinux policy: setsebool -P virt_qemu_ga_manage_ssh on
    - More secure than static cloud-init keys

    Examples:
        # Basic VM with DataSource inference
        virtctl_create_vm_advanced(
            name="fedora-vm",
            volumes={
                "volume_import": [{
                    "type": "ds",
                    "src": "fedora-42-cloud",
                    "name": "rootdisk",
                    "size": "20Gi"
                }]
            },
            infer_instancetype=True,
            infer_preference=True,
            cloud_init={
                "user": "fedora",
                "ssh_key": "ssh-rsa AAAA..."
            }
        )

        # Basic VM with container disk
        virtctl_create_vm_advanced(
            name="fedora-vm-container",
            instancetype="u1.medium",
            volumes={
                "volume_containerdisk": [{
                    "src": "quay.io/containerdisks/fedora:42",
                    "name": "containerdisk"
                }]
            },
            cloud_init={
                "user": "fedora"
            }
        )

        # Complex Windows VM with all options
        virtctl_create_vm_advanced(
            name="windows-server",
            namespace="production",
            instancetype="virtualmachineclusterinstancetype/high-performance",
            preference="virtualmachinepreference/windows-server-2019",
            run_strategy="Always",
            volumes={
                "volume_import": [{
                    "type": "dv",
                    "src": "windows-server-2019",
                    "name": "system-disk",
                    "size": "60Gi",
                    "storage_class": "fast-ssd"
                }],

                "volume_sysprep": [{
                    "name": "sysprep-config",
                    "src": "windows-config",
                    "type": "configmap"
                }]
            },
            resource_requirements={
                "memory": "8Gi",
                "cpu": "4"
            },
            termination_grace_period=300,  # 5 minutes for graceful shutdown
            access_credentials=[{
                "type": "password",
                "src": "windows-admin-password",
                "user": "Administrator",
                "method": "cloud-init"
            }]
        )

        # Production VM with advanced networking and storage
        virtctl_create_vm_advanced(
            name="enterprise-app",
            namespace="production",
            instancetype="u1.xlarge",
            preference="rhel9-server",
            run_strategy="Always",
            volumes={
                "volume_import": [{
                    "type": "ds",
                    "src": "rhel9-datasource",
                    "name": "root",
                    "size": "50Gi",
                    "storage_class": "fast-ssd"
                }],
                "volume_import": [{
                    "type": "blank",
                    "name": "app-data",
                    "size": "500Gi",
                    "storage_class": "standard"
                }, {
                    "type": "blank",
                    "name": "logs",
                    "size": "100Gi",
                    "storage_class": "fast-ssd"
                }]
            },
            cloud_init={
                "user": "rhel",
                "ssh_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC...",
                "ga_manage_ssh": True,
                "user_data": "#cloud-config\npackages:\n  - podman\n  - git\nruncmd:\n  - systemctl enable --now podman\n  - firewall-cmd --permanent --add-port=8080/tcp\n  - firewall-cmd --reload"
            },
            access_credentials=[{
                "type": "ssh",
                "src": "enterprise-ssh-keys",
                "method": "qemu-guest-agent"
            }],
            termination_grace_period=180  # 3 minutes for app shutdown
        )
    """
    # Build virtctl create vm command
    args = ["create", "vm"]

    # Use current namespace if none provided
    if namespace is None:
        namespace = get_current_namespace()

    # Add name or generate name
    if name:
        if generate_name:
            args.extend(["--generate-name", name])
        else:
            args.extend(["--name", name])

    # Add namespace
    if namespace:
        args.extend(["-n", namespace])

    # Add instance type and preference
    if instancetype:
        args.extend(["--instancetype", instancetype])
    if preference:
        args.extend(["--preference", preference])

    # Add run strategy
    if run_strategy:
        args.extend(["--run-strategy", run_strategy])

    # Add inference options
    has_inference_compatible_volumes = False
    if volumes:
        # Check for DataSources, DataVolumes, or PVCs
        if "volume_import" in volumes and any(
            vol.get("type") in ["pvc", "dv", "ds"] for vol in volumes["volume_import"]
        ):
            has_inference_compatible_volumes = True
        elif "volume_pvc" in volumes:
            has_inference_compatible_volumes = True

    # Add inference flags for compatible volumes
    if has_inference_compatible_volumes:
        if infer_instancetype:
            args.append("--infer-instancetype")
        if infer_preference:
            args.append("--infer-preference")
        if infer_instancetype_from:
            args.extend(["--infer-instancetype-from", infer_instancetype_from])
        if infer_preference_from:
            args.extend(["--infer-preference-from", infer_preference_from])

    # Add volume configurations
    if volumes:
        # Handle volume imports (PVCs, DataVolumes, DataSources)
        if "volume_import" in volumes:
            for vol in volumes["volume_import"]:
                vol_spec = f"type:{vol['type']},src:{vol['src']}"
                if "name" in vol:
                    vol_spec += f",name:{vol['name']}"
                if "size" in vol:
                    vol_spec += f",size:{vol['size']}"
                if "storage_class" in vol:
                    vol_spec += f",storageClass:{vol['storage_class']}"
                if "namespace" in vol:
                    vol_spec += f",namespace:{vol['namespace']}"
                if "bootorder" in vol:
                    vol_spec += f",bootorder:{vol['bootorder']}"
                args.extend(["--volume-import", vol_spec])

        # Handle container disks
        if "volume_containerdisk" in volumes:
            for vol in volumes["volume_containerdisk"]:
                vol_spec = f"src:{vol['src']}"
                if "name" in vol:
                    vol_spec += f",name:{vol['name']}"
                if "bootorder" in vol:
                    vol_spec += f",bootorder:{vol['bootorder']}"
                args.extend(["--volume-containerdisk", vol_spec])



        # Handle existing PVCs
        if "volume_pvc" in volumes:
            for vol in volumes["volume_pvc"]:
                vol_spec = f"src:{vol['src']}"
                if "name" in vol:
                    vol_spec += f",name:{vol['name']}"
                args.extend(["--volume-pvc", vol_spec])

        # Handle sysprep volumes
        if "volume_sysprep" in volumes:
            for vol in volumes["volume_sysprep"]:
                vol_spec = f"src:{vol['src']}"
                if "name" in vol:
                    vol_spec += f",name:{vol['name']}"
                if "type" in vol:
                    vol_spec += f",type:{vol['type']}"
                args.extend(["--volume-sysprep", vol_spec])

    # Add cloud-init configuration
    if cloud_init:
        if "user" in cloud_init:
            args.extend(["--user", cloud_init["user"]])
        if "ssh_key" in cloud_init:
            args.extend(["--ssh-key", cloud_init["ssh_key"]])
        if "password_file" in cloud_init:
            args.extend(["--password-file", cloud_init["password_file"]])
        if cloud_init.get("ga_manage_ssh", False):
            args.append("--ga-manage-ssh")
        if "user_data" in cloud_init:
            args.extend(["--user-data", cloud_init["user_data"]])
        if "user_data_base64" in cloud_init:
            args.extend(["--user-data-base64", cloud_init["user_data_base64"]])

    # Add resource requirements
    if resource_requirements:
        if "memory" in resource_requirements:
            args.extend(["--memory", resource_requirements["memory"]])
        if "cpu" in resource_requirements:
            args.extend(["--cpu", resource_requirements["cpu"]])

    # Add access credentials
    if access_credentials:
        for cred in access_credentials:
            cred_spec = f"type:{cred['type']},src:{cred['src']}"
            if "user" in cred:
                cred_spec += f",user:{cred['user']}"
            if "method" in cred:
                cred_spec += f",method:{cred['method']}"
            args.extend(["--access-cred", cred_spec])

    # Add termination grace period
    if termination_grace_period > 0:
        args.extend(["--termination-grace-period", str(termination_grace_period)])

    result = _run_virtctl_command(args)
    return result.stdout


# Volume Management Tools
@mcp.tool()
def virtctl_volume_management(
    vm_name: str,
    operation: str,
    namespace: str = None,
    volume_name: str = "",
    source: str = "",
    disk_config: Dict[str, Any] = None,
    persist: bool = False,
    dry_run: bool = False,
    size: str = "",
    storage_class: str = "",
) -> str:
    """Comprehensive hot-plug storage operations for running VMs.

    PREREQUISITES:
    - VM must be running: kubectl get vmi {vm_name} -n {namespace}
    - For PVC source: PVC must exist and be available
    - For blank volumes: Storage class must support dynamic provisioning
    - Hot-plug capability: VM must support virtio or SCSI hot-plug

    OPERATION GUIDE:

    ADD: Hot-plug attach new storage to running VM
    - Best for: Adding data storage, backup volumes, temporary space
    - VM sees new disk immediately (check with lsblk inside VM)
    - Use persist=True to survive VM restarts
    - Test with dry_run=True first for validation

    REMOVE: Hot-plug detach storage from running VM
    - Best for: Removing temporary storage, freeing resources
    - VM immediately loses access to disk
    - Data on PVC persists after removal
    - Always unmount inside VM before removing

    LIST: Show all volumes currently attached to VM
    - Best for: Inventory, troubleshooting storage issues
    - Shows both hot-plugged and permanent volumes
    - Use kubectl describe vm for more detailed volume info

    VOLUME SOURCE TYPES:

    pvc:pvc-name - Use existing PersistentVolumeClaim
    - Best for: Shared storage, pre-populated data
    - PVC must be ReadWriteOnce or ReadWriteMany
    - Check PVC status: kubectl get pvc {pvc-name} -n {namespace}

    blank:volume-name - Create new empty volume
    - Best for: Database storage, logs, temporary space
    - Dynamically provisions new PVC
    - Must specify size and optionally storage_class

    registry:image-url - Mount container image as disk
    - Best for: Tools, utilities, read-only data
    - Image is mounted as read-only filesystem
    - Good for diagnostic tools, configuration data

    DISK CONFIGURATION OPTIONS:

    Bus Types:
    - virtio: Best performance (Linux VMs)
    - sata: Good compatibility (older OSes)
    - scsi: Enterprise features (Windows, databases)

    Cache Modes:
    - writeback: Best performance (default, good for most uses)
    - writethrough: Data integrity (slower but safer)
    - none: Direct I/O (best for databases)

    Storage Class Selection:
    - fast-ssd: High IOPS (databases, logs)
    - standard: Balanced performance (general use)
    - slow-hdd: Cheap storage (backups, archives)

    TROUBLESHOOTING:

    Common Issues:
    - "Volume already exists": Choose different volume_name
    - "PVC not found": Verify PVC exists and namespace is correct
    - "Storage class not found": Check available classes with kubectl get sc
    - "VM not found": Verify VM is running (not just created)
    - "Permission denied": Check PVC access modes and pod security

    Best Practices:
    - Always test with dry_run=True first
    - Use descriptive volume names: "mysql-data", "nginx-logs"
    - Match storage class to performance needs
    - Use persist=True for permanent storage
    - Unmount inside VM before removing volumes

    Pre-operation Checks:
    - VM status: kubectl get vmi {vm_name} -n {namespace}
    - PVC availability: kubectl get pvc -n {namespace}
    - Storage classes: kubectl get sc
    - Disk space in VM: df -h (inside VM)

    Args:
        vm_name: Name of the virtual machine
        operation: Volume operation (add, remove, list)
        namespace: Kubernetes namespace containing the VM (optional)
        volume_name: Name of the volume (required for add/remove)
        source: Volume source specification for add operation (optional)
        disk_config: Disk configuration options (optional)
        persist: Persist volume changes to VM spec (optional)
        dry_run: Show what would be done without executing (optional)
        size: Size for new blank volumes (required for blank volumes)
        storage_class: Storage class for new volumes (optional)

    Common Workflows:
        # Database storage expansion
        virtctl_volume_management("postgres-vm", "add", "production",
                                volume_name="additional-data",
                                source="blank:postgres-data-2",
                                size="500Gi",
                                storage_class="fast-ssd",
                                disk_config={"bus": "virtio", "cache": "none"},
                                persist=True, dry_run=True)  # Test first

        # Temporary diagnostic volume
        virtctl_volume_management("debug-vm", "add", "development",
                                volume_name="tools",
                                source="registry:quay.io/debug-tools:latest",
                                disk_config={"readonly": True})

        # Backup volume workflow
        virtctl_volume_management("app-vm", "add", "production",
                                volume_name="backup-mount",
                                source="pvc:backup-pvc")
        # Run backup inside VM, then:
        virtctl_volume_management("app-vm", "remove", "production",
                                volume_name="backup-mount")

        # Volume inventory check
        virtctl_volume_management("database-vm", "list", "production")
    """
    if operation == "list":
        # Use current namespace if none provided
        if namespace is None:
            namespace = get_current_namespace()

        # For list operation, we'll use kubectl to get VM volumes info
        args = ["get", "vm", vm_name]
        if namespace:
            args.extend(["-n", namespace])
        args.extend(
            ["-o", "jsonpath={.spec.template.spec.domain.devices.disks[*].name}"]
        )

        try:
            # Try kubectl first for volume listing
            result = subprocess.run(
                ["kubectl"] + args,
                capture_output=True,
                text=True,
                check=True,
                timeout=120,
            )
            volumes = result.stdout.strip().split()
            return f"Volumes attached to VM '{vm_name}':\n" + "\n".join(
                f"- {vol}" for vol in volumes if vol
            )
        except Exception:
            return f"Could not list volumes for VM '{vm_name}'. VM may not exist or you may lack permissions."

    # Build command for add/remove operations
    args = []

    if operation == "add":
        args.append("addvolume")
        args.append(vm_name)

        if not volume_name:
            raise VirtctlError("volume_name is required for add operation")

        args.extend(["--volume-name", volume_name])

        # Handle different source types
        if source:
            if source.startswith("pvc:"):
                pvc_name = source[4:]  # Remove "pvc:" prefix
                args.extend(["--volume-source", f"pvc:{pvc_name}"])
            elif source.startswith("blank:"):
                vol_name = source[6:]  # Remove "blank:" prefix
                if not size:
                    raise VirtctlError("size is required for blank volumes")
                args.extend(["--volume-source", f"blank:{vol_name}"])
                args.extend(["--size", size])
                if storage_class:
                    args.extend(["--storageclass", storage_class])
            elif source.startswith("registry:"):
                image_url = source[9:]  # Remove "registry:" prefix
                args.extend(["--volume-source", f"registry:{image_url}"])
            else:
                args.extend(["--volume-source", source])

        # Add disk configuration
        if disk_config:
            if "bus" in disk_config:
                args.extend(["--bus", disk_config["bus"]])
            if "cache" in disk_config:
                args.extend(["--cache", disk_config["cache"]])
            if "serial" in disk_config:
                args.extend(["--serial", disk_config["serial"]])
            if disk_config.get("readonly", False):
                args.append("--readonly")

        if persist:
            args.append("--persist")

    elif operation == "remove":
        args.append("removevolume")
        args.append(vm_name)

        if not volume_name:
            raise VirtctlError("volume_name is required for remove operation")

        args.extend(["--volume-name", volume_name])

    else:
        raise VirtctlError(
            f"Invalid operation: {operation}. Valid operations: add, remove, list"
        )

    # Use current namespace if none provided
    if namespace is None:
        namespace = get_current_namespace()

    # Add common arguments
    if namespace:
        args.extend(["-n", namespace])

    if dry_run and operation != "list":
        args.append("--dry-run")

    result = _run_virtctl_command(args)

    return f"Volume {operation} operation completed successfully:\nCommand: {_format_shell_command(['virtctl'] + args)}\nOutput: {result.stdout.strip()}"


# Image Operations Tools
@mcp.tool()
def virtctl_image_operations(
    operation: str,
    vm_name: str = "",
    namespace: str = None,
    image_path: str = "",
    pvc_name: str = "",
    size: str = "",
    storage_class: str = "",
    upload_config: Dict[str, Any] = None,
    guestfs_config: Dict[str, Any] = None,
    memory_dump_config: Dict[str, Any] = None,
    export_config: Dict[str, Any] = None,
) -> str:
    """Advanced disk and image management.

    This unified tool handles:
    - upload: Image uploads to PVCs/DataVolumes
    - export: VM exports and backups
    - guestfs: Libguestfs operations for disk inspection
    - memory-dump: Memory dumps for debugging

    Args:
        operation: Image operation (upload, export, guestfs, memory-dump)
        vm_name: VM name (required for memory-dump, export)
        namespace: Kubernetes namespace (optional)
        image_path: Local image file path (required for upload)
        pvc_name: PVC name (required for upload, guestfs)
        size: PVC size for new volumes (optional)
        storage_class: Storage class for new PVCs (optional)
        upload_config: Upload configuration options (optional)
        guestfs_config: Libguestfs configuration options (optional)
        memory_dump_config: Memory dump configuration (optional)
        export_config: Export configuration options (optional)
    """
    args = []

    if operation == "upload":
        if not image_path or not pvc_name:
            raise VirtctlError(
                "image_path and pvc_name are required for upload operation"
            )

        args.extend(["image-upload", pvc_name])
        args.extend(["--image-path", image_path])

        if size:
            args.extend(["--size", size])
        if storage_class:
            args.extend(["--storageclass", storage_class])

        if upload_config:
            if upload_config.get("insecure", False):
                args.append("--insecure")
            if upload_config.get("block_volume", False):
                args.append("--block-volume")

    elif operation == "guestfs":
        if not pvc_name:
            raise VirtctlError("pvc_name is required for guestfs operation")

        args.extend(["guestfs", pvc_name])

        if guestfs_config:
            if guestfs_config.get("kvm", False):
                args.append("--kvm")
            if "pull_policy" in guestfs_config:
                args.extend(["--pull-policy", guestfs_config["pull_policy"]])

    elif operation == "memory-dump":
        if not vm_name:
            raise VirtctlError("vm_name is required for memory-dump operation")

        args.extend(["memory-dump", "get", vm_name])

        if memory_dump_config:
            if memory_dump_config.get("create_claim", False):
                args.append("--create-claim")
            if "claim_name" in memory_dump_config:
                args.extend(["--claim-name", memory_dump_config["claim_name"]])

    elif operation == "export":
        if not vm_name:
            raise VirtctlError("vm_name is required for export operation")

        args.extend(["vmexport", "download", vm_name])

        if export_config:
            if "output" in export_config:
                args.extend(["--output", export_config["output"]])
            if export_config.get("port_forward", False):
                args.append("--port-forward")

    else:
        raise VirtctlError(f"Invalid operation: {operation}")

    # Use current namespace if none provided
    if namespace is None:
        namespace = get_current_namespace()

    if namespace:
        args.extend(["-n", namespace])

    result = _run_virtctl_command(args)
    return f"Image {operation} operation completed successfully:\nCommand: {_format_shell_command(['virtctl'] + args)}\nOutput: {result.stdout.strip()}"


# Service Management Tools
@mcp.tool()
def virtctl_service_management(
    operation: str,
    resource_name: str,
    resource_type: str = "vm",
    namespace: str = None,
    expose_config: Dict[str, Any] = None,
    service_name: str = "",
) -> str:
    """Network services and connectivity management.

    PREREQUISITES:
    - VM must be running with network interfaces
    - For expose: VM must have services listening on target ports
    - Network connectivity: Ensure firewall rules allow traffic
    - Service mesh: Check if service mesh policies affect traffic

    OPERATION GUIDE:

    EXPOSE: Create Kubernetes Service for VM external access
    - Best for: Production workloads, web servers, databases, APIs
    - Creates stable endpoint for VM services
    - Supports LoadBalancer, NodePort, ClusterIP service types
    - Use for persistent access to VM services

    UNEXPOSE: Remove Kubernetes Service for VM
    - Best for: Removing temporary access, decommissioning services
    - Cleans up service resources and endpoints
    - Active connections to service will be terminated

    SERVICE TYPE SELECTION:

    LoadBalancer:
    - Best for: Production internet-facing services
    - Requires cloud provider load balancer support
    - Gets external IP address
    - Supports SSL termination, health checks

    NodePort:
    - Best for: Testing, development, internal services
    - Exposes service on all cluster nodes
    - Uses high-numbered port (30000-32767)
    - Accessible via any node IP

    ClusterIP (default):
    - Best for: Internal services, microservices communication
    - Only accessible from within cluster
    - Most secure option for internal communication

    PORT CONFIGURATION:

    Common Service Patterns:
    - Web servers: port=80, target_port=8080
    - HTTPS services: port=443, target_port=8443
    - SSH access: port=22, target_port=22
    - Databases: port=5432, target_port=5432 (PostgreSQL)
    - APIs: port=8080, target_port=8080

    TROUBLESHOOTING:

    Common Issues:
    - "Service not accessible": Check VM is running and listening on port
    - "Connection refused": Verify firewall rules inside VM
    - "LoadBalancer pending": Cloud provider may not support LoadBalancer
    - "Port already in use": Choose different external port

    Debugging Steps:
    1. Check VM status: kubectl get vmi {resource_name} -n {namespace}
    2. Test VM port: virtctl console {resource_name} then netstat -tlnp
    3. Check service: kubectl get svc {service_name} -n {namespace}
    4. Verify endpoints: kubectl get endpoints {service_name} -n {namespace}
    5. Test connectivity: kubectl port-forward svc/{service_name} local:remote

    Best Practices:
    - Use descriptive service names: web-server-svc, db-primary-svc
    - Add resource labels for monitoring and discovery
    - Use ClusterIP for internal service communication
    - Reserve LoadBalancer for internet-facing services
    - Test port accessibility before creating services

    Args:
        operation: Service operation (expose, unexpose)
        resource_name: Name of the VM or resource to manage
        resource_type: Resource type (vm, vmi) (optional, default: vm)
        namespace: Kubernetes namespace (optional)
        expose_config: Configuration for service exposure (optional)
        service_name: Name of the service for unexpose operation (optional)

    Expose Configuration Examples:
        # Basic web service
        expose_config = {
            "port": 80,
            "target_port": 8080,
            "service_type": "LoadBalancer",
            "service_name": "web-service"
        }

        # Internal database service
        expose_config = {
            "port": 5432,
            "target_port": 5432,
            "service_type": "ClusterIP",
            "service_name": "postgres-internal"
        }

        # Development service with NodePort
        expose_config = {
            "port": 3000,
            "target_port": 3000,
            "service_type": "NodePort",
            "service_name": "dev-app"
        }

    Workflow Examples:
        # Expose production web service
        virtctl_service_management("expose", "web-server", "vm", "production",
                                 expose_config={
                                     "port": 80,
                                     "target_port": 8080,
                                     "service_type": "LoadBalancer",
                                     "service_name": "web-service"
                                 })

        # Expose internal API service
        virtctl_service_management("expose", "api-server", "vm",
                                 expose_config={
                                     "port": 8080,
                                     "target_port": 8080,
                                     "service_type": "ClusterIP"
                                 })

        # Remove service
        virtctl_service_management("unexpose", "test-vm", "vm", "development",
                                 service_name="test-service")
    """
    args = []

    if operation == "expose":
        args.extend(["expose", resource_type, resource_name])

        if expose_config:
            if "service_name" in expose_config:
                args.extend(["--name", expose_config["service_name"]])
            if "port" in expose_config:
                args.extend(["--port", str(expose_config["port"])])
            if "target_port" in expose_config:
                args.extend(["--target-port", str(expose_config["target_port"])])
            if "service_type" in expose_config:
                args.extend(["--type", expose_config["service_type"]])

    elif operation == "unexpose":
        if not service_name:
            service_name = f"{resource_name}-service"

        # Use current namespace if none provided
        if namespace is None:
            namespace = get_current_namespace()

        kubectl_args = ["delete", "service", service_name]
        if namespace:
            kubectl_args.extend(["-n", namespace])

        try:
            result = subprocess.run(
                ["kubectl"] + kubectl_args,
                capture_output=True,
                text=True,
                check=True,
                timeout=120,
            )
            return f"Service unexpose completed: {result.stdout.strip()}"
        except subprocess.CalledProcessError as e:
            raise VirtctlError(f"Failed to unexpose service: {e.stderr}")

    # Use current namespace if none provided (for expose operation)
    if namespace is None:
        namespace = get_current_namespace()

    if namespace:
        args.extend(["-n", namespace])

    result = _run_virtctl_command(args)
    return f"Service {operation} completed: {result.stdout.strip()}"


# Diagnostics Tools
@mcp.tool()
def virtctl_diagnostics(
    diagnostic_type: str,
    vm_name: str = "",
    namespace: str = None,
) -> str:
    """Comprehensive VM diagnostics and monitoring capabilities.

    PREREQUISITES:
    - VM must be running: kubectl get vmi {vm_name} -n {namespace}
    - Guest agent required: qemu-guest-agent must be installed and running in VM
    - Network connectivity: Guest agent needs communication with host

    DIAGNOSTIC OPERATIONS:

    GUESTOSINFO: Detailed guest operating system information
    - Best for: OS version detection, architecture verification, troubleshooting compatibility
    - Returns: OS name, version, kernel, architecture, timezone
    - Requirements: Guest agent active and responsive
    - Use case: Verify VM OS for automation, compliance auditing

    FSLIST: Guest filesystem inventory and mount points
    - Best for: Storage troubleshooting, capacity planning, mount verification
    - Returns: Filesystem types, mount points, used space, available space
    - Requirements: Guest agent with filesystem access permissions
    - Use case: Debug storage issues, check disk usage, validate mounts

    USERLIST: Active and system users in the guest OS
    - Best for: Security auditing, access control verification, user management
    - Returns: Username, UID, login status, home directory
    - Requirements: Guest agent with user enumeration permissions
    - Use case: Security compliance, troubleshoot login issues

    VERSION: Tool and cluster version information
    - Best for: Compatibility verification, troubleshooting tool issues
    - Returns: virtctl version, KubeVirt version, cluster details
    - Requirements: None (local command)
    - Use case: Version compatibility checks, support troubleshooting

    GUEST AGENT TROUBLESHOOTING:

    Common Issues:
    - "Guest agent not responding": Check agent service in VM
      Linux: systemctl status qemu-guest-agent
      Windows: Check QEMU Guest Agent service

    - "Permission denied": Agent lacks required permissions
      Solution: Run agent with appropriate user privileges

    - "Timeout": Network issues or slow VM response
      Solution: Check VM load, network connectivity

    - "Command not supported": Old agent version
      Solution: Update guest agent to latest version

    Guest Agent Installation:

    Linux (RHEL/CentOS/Fedora):
    - dnf install qemu-guest-agent
    - systemctl enable --now qemu-guest-agent

    Linux (Ubuntu/Debian):
    - apt install qemu-guest-agent
    - systemctl enable --now qemu-guest-agent

    Windows:
    - Install from VirtIO drivers ISO
    - Or download from QEMU project
    - Start "QEMU Guest Agent" service

    MONITORING WORKFLOWS:

    Health Check Sequence:
    1. virtctl_diagnostics("version") → Verify tool compatibility
    2. virtctl_diagnostics("guestosinfo", vm) → Confirm guest agent connectivity
    3. virtctl_diagnostics("fslist", vm) → Check storage health
    4. virtctl_diagnostics("userlist", vm) → Verify access control

    Troubleshooting Sequence:
    1. Check VM status: kubectl get vmi {vm_name} -n {namespace}
    2. Check guest agent: virtctl_diagnostics("guestosinfo", vm)
    3. If agent fails: Access console and check agent service
    4. Storage issues: virtctl_diagnostics("fslist", vm)
    5. Access issues: virtctl_diagnostics("userlist", vm)

    Args:
        diagnostic_type: Diagnostic operation (guestosinfo, fslist, userlist, version)
        vm_name: Name of the virtual machine (optional, not required for version operation)
        namespace: Kubernetes namespace containing the VM (optional)
        output_format: Output format (json only) (optional)

    Returns:
        Diagnostic information about the VM or system

    Troubleshooting Examples:
        # Quick health check
        virtctl_diagnostics("guestosinfo", "prod-vm", "production")

        # Storage investigation
        virtctl_diagnostics("fslist", "db-vm", "production")

        # Security audit
        virtctl_diagnostics("userlist", "web-vm", "production")

        # Tool compatibility check
        virtctl_diagnostics("version")
    """
    args = []

    if diagnostic_type == "guestosinfo":
        args.extend(["guestosinfo", vm_name])
    elif diagnostic_type == "fslist":
        args.extend(["fslist", vm_name])
    elif diagnostic_type == "userlist":
        args.extend(["userlist", vm_name])
    elif diagnostic_type == "version":
        args.extend(["version"])
    else:
        raise VirtctlError(f"Invalid diagnostic type: {diagnostic_type}")

    # Use current namespace if none provided
    if namespace is None:
        namespace = get_current_namespace()

    if namespace:
        args.extend(["-n", namespace])

    result = _run_virtctl_command(args)
    return result.stdout


# Resource Creation Tools
@mcp.tool()
def virtctl_create_resources(
    resource_type: str,
    name: str,
    namespaced: bool = False,
    namespace: str = None,
    instancetype_config: Dict[str, Any] = None,
    preference_config: Dict[str, Any] = None,
) -> str:
    """Advanced resource template creation for instance types and preferences.

    This tool creates reusable VM configuration templates that standardize
    resource allocation and VM behavior across your cluster.

    INSTANCE TYPE CONFIGURATION:

    Instance types define CPU and memory allocations for VMs. They provide
    standardized sizing that can be referenced by name when creating VMs.

    Common Instance Type Patterns:
    - Nano: 1 CPU, 1Gi RAM (minimal workloads, testing)
    - Micro: 1 CPU, 2Gi RAM (light services, development)
    - Small: 1 CPU, 4Gi RAM (small applications)
    - Medium: 2 CPU, 8Gi RAM (standard workloads)
    - Large: 4 CPU, 16Gi RAM (demanding applications)
    - XLarge: 8 CPU, 32Gi RAM (high-performance workloads)

    CPU Configuration:
    - cores: Number of CPU cores
    - threads: Threads per core (for hyperthreading)
    - sockets: CPU sockets (for NUMA topology)
    - dedicatedCPUPlacement: Pin CPUs for performance
    - isolateEmulatorThread: Isolate QEMU threads

    Memory Configuration:
    - guest: Memory available to VM
    - hugepages: Use hugepages for performance
    - overcommitPercent: Allow memory overcommit

    GPU Resources:
    - nvidia.com/gpu: NVIDIA GPU passthrough
    - amd.com/gpu: AMD GPU passthrough
    - intel.com/gpu: Intel GPU passthrough

    PREFERENCE CONFIGURATION:

    Preferences define VM behavior, features, and hardware characteristics.
    They complement instance types with non-resource settings.

    Machine Types:
    - q35: Modern machine type (recommended for new VMs)
    - pc: Legacy machine type (for compatibility)
    - microvm: Minimal machine type (for containers)

    CPU Features:
    - CPU topology: sockets, cores, threads configuration
    - CPU model: host-model, host-passthrough, or specific model
    - CPU features: enable/disable specific CPU features

    Firmware Settings:
    - BIOS: Traditional BIOS boot
    - UEFI: Modern UEFI boot (supports SecureBoot)
    - EFI: EFI boot mode

    Device Settings:
    - autoattachPodInterface: Auto-attach to pod network
    - autoattachSerialConsole: Enable serial console
    - autoattachGraphicsDevice: Enable graphics
    - rng: Random number generator settings

    Clock Settings:
    - timezone: VM timezone
    - utc/localtime: Clock configuration

    USAGE TIPS:

    Naming Conventions:
    - Use descriptive names: "web-server-small", "database-large"
    - Include sizing info: "u1.medium", "c5.xlarge"
    - Consider environment prefixes: "prod-", "dev-", "test-"

    Cluster vs Namespaced:
    - Cluster-scoped: Available to all namespaces (recommended for common configs)
    - Namespaced: Specific to one namespace (for specialized configs)

    Best Practices:
    - Create standard sizing tiers for consistency
    - Document resource templates and their intended use
    - Use preferences for OS-specific optimizations
    - Test configurations before production deployment
    - Version your templates (v1, v2, etc.)

    Args:
        resource_type: Type of resource (instancetype, preference)
        name: Name of the resource
        namespaced: Create namespaced resource (default: cluster-scoped)
        namespace: Kubernetes namespace (required if namespaced=True)
        instancetype_config: Instance type configuration (optional)
        preference_config: Preference configuration (optional)
        output_format: Output format (json only) (optional)

    Instance Type Configuration Examples:
        # Basic CPU/Memory
        instancetype_config = {
            "cpu": 4,  # Number of CPUs
            "memory": "8Gi",  # Memory amount
        }

        # Advanced CPU topology
        instancetype_config = {
            "cpu": 8,
            "memory": "16Gi",
            "cpu_topology": {
                "sockets": 2,
                "cores": 2,
                "threads": 2
            },
            "dedicatedCPUPlacement": True,  # Pin CPUs
            "isolateEmulatorThread": True   # Isolate QEMU threads
        }

        # GPU workload
        instancetype_config = {
            "cpu": 8,
            "memory": "32Gi",
            "gpus": ["nvidia.com/gpu"],  # GPU resources
            "hugepages": "2Mi"  # Use hugepages
        }

        # I/O intensive
        instancetype_config = {
            "cpu": 16,
            "memory": "64Gi",
            "iothreads_policy": "auto",  # Enable I/O threads
            "ioThreadsPolicy": "auto"
        }

    Preference Configuration Examples:
        # Linux server preference
        preference_config = {
            "machine_type": "q35",  # Modern machine type
            "cpu_model": "host-model",  # Use host CPU features
            "firmware": "uefi",  # UEFI boot
            "features": {
                "acpi": {"enabled": True},
                "apic": {"enabled": True},
                "hyperv": {
                    "relaxed": {"enabled": True},
                    "vapic": {"enabled": True},
                    "spinlocks": {"enabled": True, "spinlocks": 8191}
                }
            }
        }

        # Windows desktop preference
        preference_config = {
            "machine_type": "q35",
            "cpu_topology": {
                "sockets": 1,
                "cores": 4,
                "threads": 2
            },
            "clock": {"timezone": "America/New_York"},
            "devices": {
                "autoattachGraphicsDevice": True,
                "autoattachSerialConsole": True,
                "rng": {}  # Enable random number generator
            }
        }

        # Performance-optimized preference
        preference_config = {
            "machine_type": "q35",
            "cpu_model": "host-passthrough",  # Best performance
            "features": {
                "kvm": {"hidden": True},  # Hide virtualization
                "hyperv": {
                    "relaxed": {"enabled": True},
                    "vapic": {"enabled": True},
                    "time": {"enabled": True}
                }
            }
        }

    Examples:
        # Create standard instance type sizes
        virtctl_create_resources(
            "instancetype", "u1.small", False,
            instancetype_config={"cpu": 1, "memory": "4Gi"}
        )

        virtctl_create_resources(
            "instancetype", "u1.medium", False,
            instancetype_config={"cpu": 2, "memory": "8Gi"}
        )

        # Create OS-specific preference
        virtctl_create_resources(
            "preference", "fedora-server", False,
            preference_config={
                "machine_type": "q35",
                "features": {"acpi": True, "apic": True}
            }
        )

        # Create GPU instance type
        virtctl_create_resources(
            "instancetype", "gpu-workload", False,
            instancetype_config={
                "cpu": 8,
                "memory": "32Gi",
                "gpus": ["nvidia.com/gpu"]
            }
        )

        # Create development instance type (namespaced)
        virtctl_create_resources(
            "instancetype", "dev-small", True, "development",
            instancetype_config={"cpu": 1, "memory": "2Gi"}
        )
    """
    args = ["create"]

    if resource_type == "instancetype":
        if namespaced:
            args.append("instancetype")
        else:
            args.append("cluster-instancetype")
    elif resource_type == "preference":
        if namespaced:
            args.append("preference")
        else:
            args.append("cluster-preference")
    else:
        raise VirtctlError(f"Invalid resource type: {resource_type}")

    args.extend(["--name", name])

    if namespaced:
        if not namespace:
            raise VirtctlError("namespace is required for namespaced resources")
        args.extend(["-n", namespace])

    # Add configuration
    if instancetype_config and resource_type == "instancetype":
        if "cpu" in instancetype_config:
            args.extend(["--cpu", str(instancetype_config["cpu"])])
        if "memory" in instancetype_config:
            args.extend(["--memory", instancetype_config["memory"]])

    if preference_config and resource_type == "preference":
        if "machine_type" in preference_config:
            args.extend(["--machine-type", preference_config["machine_type"]])

    result = _run_virtctl_command(args)
    return result.stdout


# DataSource Management Tools
@mcp.tool()
def virtctl_datasource_management(
    operation: str,
    name: str = "",
    namespace: str = None,
    source_config: Dict[str, Any] = None,
    storage_config: Dict[str, Any] = None,
    metadata: Dict[str, Any] = None,
    clone_config: Dict[str, Any] = None,
    output_format: str = "json",
) -> str:
    """Comprehensive DataSource management for VM boot images.

    DataSources are essential for creating VMs as they provide ready-to-use OS images
    that can automatically infer the correct instance types and preferences.

    DATASOURCE BEST PRACTICES:

    Public Image Sources:
    - Fedora Cloud: https://download.fedoraproject.org/pub/fedora/linux/releases/
    - CentOS Stream: https://cloud.centos.org/centos/
    - Ubuntu Cloud Images: https://cloud-images.ubuntu.com/
    - RHEL (subscription required): https://access.redhat.com/downloads/
    - Container Registry Images: quay.io/containerdisks/, registry.redhat.io/

    Popular Container Images:
    - quay.io/containerdisks/fedora:latest, :42, :41
    - quay.io/containerdisks/ubuntu:22.04, :20.04
    - quay.io/containerdisks/centos-stream:9, :8
    - registry.redhat.io/rhel8/rhel-guest-image (requires auth)

    Image Selection Tips:
    - Use cloud images (pre-configured for cloud-init)
    - Choose qcow2 format for better compression and features
    - Consider minimal/cloud variants for smaller size
    - Check for official/maintained sources
    - Verify image checksums when possible

    Metadata Annotations (Important):
    - instancetype.kubevirt.io/default-instancetype: "u1.medium"
    - instancetype.kubevirt.io/default-preference: "fedora"
    - These enable automatic resource selection when creating VMs

    Storage Considerations:
    - Size based on compressed image size + expansion room
    - Use fast storage classes for better VM performance
    - Consider read-only images with separate data volumes

    Operations:
    - list: Display available DataSources
    - create: Generate a DataSource manifest (returns YAML/JSON; does not apply)

    Args:
        operation: DataSource operation (create, list)
        name: DataSource name (required for create)
        namespace: Kubernetes namespace (optional)
        source_config: Source configuration for create operation
        storage_config: Storage configuration for create operation
        metadata: Metadata configuration for create operation
        clone_config: Clone configuration (not used)
        output_format: Output format (json only) (optional)

    Source Configuration Examples:
        # Fedora 42 Cloud Image
        source_config = {
            "source_type": "http",
            "url": "https://download.fedoraproject.org/pub/fedora/linux/releases/42/Cloud/x86_64/images/Fedora-Cloud-Base-42-1.6.x86_64.qcow2"
        }

        # CentOS Stream 9
        source_config = {
            "source_type": "http",
            "url": "https://cloud.centos.org/centos/9-stream/x86_64/images/CentOS-Stream-GenericCloud-9-latest.x86_64.qcow2"
        }

        # Ubuntu 22.04 LTS
        source_config = {
            "source_type": "http",
            "url": "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img"
        }

        # Container Registry
        source_config = {
            "source_type": "registry",
            "registry_config": {
                "url": "docker://quay.io/containerdisks/fedora:42",
                "secret_ref": "registry-credentials"  # optional for private registries
            }
        }

    Storage Configuration:
        storage_config = {
            "size": "20Gi",  # Based on image size + growth
            "storage_class": "fast-ssd",  # Performance matters for VMs
            "access_modes": ["ReadWriteOnce"],  # Standard for VM disks
            "volume_mode": "Filesystem"  # or Block for raw access
        }

    Metadata Configuration (Essential for auto-inference):
        metadata = {
            "labels": {"os": "fedora", "version": "42", "arch": "amd64"},
            "annotations": {
                "instancetype.kubevirt.io/default-instancetype": "u1.small",
                "instancetype.kubevirt.io/default-preference": "fedora",
                "description": "Fedora 42 Cloud Base Image"
            },
            "os_info": {
                "os": "fedora",
                "version": "42",
                "architecture": "amd64"
            },
            "recommended_resources": {
                "instancetype": "u1.small",
                "preference": "fedora"
            }
        }

    Clone Configuration:
        clone_config = {
            "source_name": "fedora-42-base",
            "source_namespace": "default",
            "new_name": "fedora-42-dev",
            "new_namespace": "development"
        }

    Examples:
        # Create Fedora DataSource
        virtctl_datasource_management(
            "create", "fedora-42-cloud",
            source_config={
                "source_type": "http",
                "url": "https://download.fedoraproject.org/pub/fedora/linux/releases/42/Cloud/x86_64/images/Fedora-Cloud-Base-42-1.6.x86_64.qcow2"
            },
            storage_config={"size": "8Gi", "storage_class": "fast-ssd"},
            metadata={
                "labels": {"os": "fedora", "version": "42"},
                "annotations": {
                    "instancetype.kubevirt.io/default-instancetype": "u1.small",
                    "instancetype.kubevirt.io/default-preference": "fedora"
                }
            }
        )

        # Create Ubuntu from container registry
        virtctl_datasource_management(
            "create", "ubuntu-22-04",
            source_config={
                "source_type": "registry",
                "registry_config": {"url": "docker://quay.io/containerdisks/ubuntu:22.04"}
            },
            metadata={
                "labels": {"os": "ubuntu", "version": "22.04"},
                "recommended_resources": {"instancetype": "u1.medium", "preference": "ubuntu"}
            }
        )

        # Clone DataSource for customization
        virtctl_datasource_management(
            "clone",
            clone_config={
                "source_name": "fedora-42-cloud",
                "source_namespace": "default",
                "new_name": "fedora-42-dev",
                "new_namespace": "development"
            }
        )
    """
    if operation == "list":
        # Use current namespace if none provided
        if namespace is None:
            namespace = get_current_namespace()

        args = ["get", "datasources"]
        if namespace:
            args.extend(["-n", namespace])
        else:
            args.append("--all-namespaces")

        args.extend(["-o", "json"])

        try:
            result = subprocess.run(
                ["kubectl"] + args,
                capture_output=True,
                text=True,
                check=True,
                timeout=120,
            )
            return result.stdout
        except subprocess.CalledProcessError as e:
            if "No resources found" in e.stderr:
                return "No DataSources found"
            raise VirtctlError(f"Failed to list DataSources: {e.stderr}")

    elif operation == "create":
        if not name or not source_config:
            raise VirtctlError(
                "name and source_config are required for create operation"
            )

        # Build DataSource manifest
        # Use current namespace if none provided
        if namespace is None:
            namespace = get_current_namespace()

        datasource = {
            "apiVersion": "cdi.kubevirt.io/v1beta1",
            "kind": "DataSource",
            "metadata": {"name": name, "namespace": namespace},
            "spec": {"source": {}},
        }

        # Add source configuration
        if source_config["source_type"] == "http":
            datasource["spec"]["source"]["http"] = {"url": source_config["url"]}
        elif source_config["source_type"] == "registry":
            reg_config = source_config["registry_config"]
            datasource["spec"]["source"]["registry"] = {"url": reg_config["url"]}

        # Add metadata
        if metadata:
            if "labels" in metadata:
                datasource["metadata"]["labels"] = metadata["labels"]
            if "annotations" in metadata:
                datasource["metadata"]["annotations"] = metadata["annotations"]

        return (
            yaml.dump(datasource, default_flow_style=False)
            if output_format == "yaml"
            else json.dumps(datasource, indent=2)
        )

    else:
        raise VirtctlError(f"Invalid operation: {operation}")


# Cluster Resource Discovery Tools
@mcp.tool()
def virtctl_cluster_resources(
    resource_type: str,
    scope: str = "all",
    namespace: str = None,
    label_selector: str = "",
    show_labels: bool = False,
) -> str:
    """Discover available cluster resources for VM configuration.

    This essential tool helps discover what resources are available in the
    cluster for VM creation. Understanding available resources is crucial
    for creating properly configured VMs.

    RESOURCE DISCOVERY TIPS:

    Instance Types (CPU/Memory configs):
    - Common types: u1.nano (1CPU,1Gi), u1.micro (1CPU,2Gi), u1.small (1CPU,4Gi),
      u1.medium (2CPU,8Gi), u1.large (4CPU,16Gi), u1.xlarge (8CPU,32Gi)
    - Check both cluster-scoped and namespaced variants
    - Look for annotations like "instancetype.kubevirt.io/common-instancetypes-version"
    - Use kubectl describe to see detailed CPU/memory specifications

    Preferences (VM features/settings):
    - Common OS preferences: fedora, ubuntu, centos, rhel8, rhel9, windows10, windows11
    - Preferences define machine type (q35, pc), CPU features, firmware settings
    - Look for "instancetype.kubevirt.io/default-preference" annotations on DataSources
    - Check preference specs for CPU topology, machine features, firmware types

    DataSources (Boot images):
    - Look for labels: os=linux|windows, version=X.X, arch=amd64|arm64
    - Check annotations for recommended instance types and preferences
    - Common sources: quay.io/containerdisks, registry.redhat.io, public cloud images
    - Use kubectl describe to see source URLs and storage requirements

    Storage Classes:
    - Prioritize classes with "storageclass.kubevirt.io/is-default-virt-class=true"
    - Look for "virtualization" in storage class names (e.g., "ocs-virtualization-rbd")
    - Check provisioner types: ceph-rbd, csi-driver, local-storage, etc.
    - Consider performance characteristics: fast-ssd, standard, slow-hdd

    Args:
        resource_type: Type of resource (instancetypes, preferences, datasources, storageclasses, all)
        scope: Resource scope (all, cluster, namespaced) (optional)
        namespace: Kubernetes namespace for namespaced resources (optional)
        label_selector: Label selector for filtering (e.g., "os=linux,version=22.04") (optional)
        show_labels: Show resource labels (optional)

    Returns:
        Available cluster resources information with discovery hints

    Examples:
        # Discover all available instance types
        virtctl_cluster_resources("instancetypes", "all")

        # Find Linux DataSources only
        virtctl_cluster_resources("datasources", "cluster", label_selector="os=linux", show_labels=True)

        # Get detailed preference information
        virtctl_cluster_resources("preferences", "all")

        # Find virtualization-optimized storage classes
        virtctl_cluster_resources("storageclasses", show_labels=True)
    """
    results = []

    def get_resources(resource_name: str, namespaced: bool = True):
        args = ["get", resource_name]

        # Use current namespace if none provided
        nonlocal namespace
        if namespace is None:
            namespace = get_current_namespace()

        if namespaced:
            if namespace:
                args.extend(["-n", namespace])
            else:
                args.append("--all-namespaces")

        if label_selector:
            args.extend(["-l", label_selector])

        if show_labels:
            args.append("--show-labels")

        args.extend(["-o", "json"])

        try:
            result = subprocess.run(
                ["kubectl"] + args,
                capture_output=True,
                text=True,
                check=True,
                timeout=120,
            )
            return result.stdout
        except subprocess.CalledProcessError as e:
            if "No resources found" in e.stderr:
                return f"No {resource_name} found"
            return f"Error listing {resource_name}: {e.stderr.strip()}"

    if resource_type in ["instancetypes", "all"]:
        if scope in ["all", "cluster"]:
            cluster_instancetypes = get_resources(
                "virtualmachineclusterinstancetypes", namespaced=False
            )
            results.append(f"=== Cluster Instance Types ===\n{cluster_instancetypes}")

        if scope in ["all", "namespaced"]:
            namespaced_instancetypes = get_resources(
                "virtualmachineinstancetypes", namespaced=True
            )
            results.append(
                f"=== Namespaced Instance Types ===\n{namespaced_instancetypes}"
            )

    if resource_type in ["preferences", "all"]:
        if scope in ["all", "cluster"]:
            cluster_preferences = get_resources(
                "virtualmachineclusterpreferences", namespaced=False
            )
            results.append(f"=== Cluster Preferences ===\n{cluster_preferences}")

        if scope in ["all", "namespaced"]:
            namespaced_preferences = get_resources(
                "virtualmachinepreferences", namespaced=True
            )
            results.append(f"=== Namespaced Preferences ===\n{namespaced_preferences}")

    if resource_type in ["datasources", "all"]:
        datasources = get_resources("datasources", namespaced=True)
        results.append(f"=== DataSources ===\n{datasources}")

    if resource_type in ["storageclasses", "all"]:
        storageclasses = get_resources("storageclasses", namespaced=False)
        results.append(f"=== Storage Classes ===\n{storageclasses}")

    if not results:
        raise VirtctlError(f"Invalid resource type: {resource_type}")

    return "\n\n".join(results)


def print_startup_banner():
    """Print a colorful startup banner with server information."""
    version = get_package_version()

    # ANSI color codes
    BLUE = "\033[94m"
    GREEN = "\033[92m"
    CYAN = "\033[96m"
    LIGHT_GRAY = "\033[37m"
    BOLD = "\033[1m"
    RESET = "\033[0m"

    print(
        f"""
{BLUE}{'='*60}{RESET}
{BOLD}{GREEN}🚀 KubeVirt MCP Server{RESET}
{BLUE}{'='*60}{RESET}

{CYAN}📦 Server:{RESET}      {LIGHT_GRAY}kubevirt (VM management through virtctl){RESET}
{CYAN}🏷️  Version:{RESET}     {LIGHT_GRAY}{version}{RESET}
{CYAN}🌐 Homepage:{RESET}    {LIGHT_GRAY}https://github.com/yaacov/kubectl-mtv{RESET}
{CYAN}📋 Description:{RESET} {LIGHT_GRAY}MCP Server for KubeVirt Virtual Machine Management{RESET}
"""
    )


def parse_args():
    """Parse command-line arguments for MCP server configuration."""
    parser = argparse.ArgumentParser(
        description="KubeVirt MCP Server (VM Management Operations)",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s                          # Run with default stdio transport
  %(prog)s --transport sse           # Run with SSE transport on default port 8080
  %(prog)s --transport sse --port 9000  # Run with SSE transport on custom port
  %(prog)s --listen 0.0.0.0 --port 3000 # Listen on all interfaces, port 3000
        """,
    )

    parser.add_argument(
        "--transport",
        type=str,
        choices=["stdio", "sse"],
        default="stdio",
        help="MCP transport type (default: stdio)",
    )

    parser.add_argument(
        "--port",
        type=int,
        default=8080,
        help="Port to listen on (default: 8080)",
    )

    parser.add_argument(
        "--listen",
        "--host",
        type=str,
        default="localhost",
        help="Host/address to listen on (default: localhost)",
    )

    parser.add_argument(
        "--no-banner",
        action="store_false",
        dest="show_banner",
        default=True,
        help="Hide startup banner (default: show banner)",
    )

    # Keep legacy args for backward compatibility
    parser.add_argument(
        "--name", default="kubevirt", help="Server name (legacy, ignored)"
    )
    parser.add_argument("--version", help="Server version (legacy, ignored)")

    return parser.parse_args()


def main():
    """Main entry point for the virtctl MCP server."""
    args = parse_args()

    # Show banner if requested
    if args.show_banner:
        print_startup_banner()

    # Prepare MCP run configuration
    run_config = {}

    if args.transport == "sse":
        run_config["transport"] = "sse"
        run_config["port"] = args.port
        run_config["host"] = args.listen
    # stdio is the default, no additional config needed

    # Use the global mcp instance that has the tools registered
    # (Don't create a new server instance as that would be empty)
    mcp.run(**run_config)


if __name__ == "__main__":
    main()
