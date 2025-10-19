package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/mcp-go/pkg/mtvmcp"
)

// VirtctlVMLifecycleInput represents the input for VirtctlVMLifecycle
type VirtctlVMLifecycleInput struct {
	VMName      string `json:"vm_name" jsonschema:"Name of the virtual machine"`
	Operation   string `json:"operation" jsonschema:"Lifecycle operation (start stop restart pause unpause migrate soft-reboot)"`
	Namespace   string `json:"namespace,omitempty" jsonschema:"Kubernetes namespace containing the VM (optional)"`
	GracePeriod int    `json:"grace_period,omitempty" jsonschema:"Graceful termination period in seconds (optional)"`
	Force       bool   `json:"force,omitempty" jsonschema:"Force the operation (optional)"`
	DryRun      bool   `json:"dry_run,omitempty" jsonschema:"Show what would be done without executing (optional)"`
	NodeName    string `json:"node_name,omitempty" jsonschema:"Target node name for migrate operation (optional)"`
	Timeout     string `json:"timeout,omitempty" jsonschema:"Operation timeout (optional)"`
}

// GetVirtctlVMLifecycleTool returns the tool definition
func GetVirtctlVMLifecycleTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "VirtctlVMLifecycle",
		Description: `Unified VM power state management for virtual machines.

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
        virtctl_vm_lifecycle(vm_name="prod-db", operation="stop", namespace="production", grace_period=180)
        # Wait for clean shutdown, then:
        virtctl_vm_lifecycle(vm_name="prod-db", operation="start", namespace="production")

        # Node maintenance migration
        virtctl_vm_lifecycle(vm_name="critical-vm", operation="migrate", namespace="production",
                           node_name="worker-3", dry_run=true)  # Test first
        virtctl_vm_lifecycle(vm_name="critical-vm", operation="migrate", namespace="production",
                           node_name="worker-3")  # Execute

        # Emergency shutdown
        virtctl_vm_lifecycle(vm_name="stuck-vm", operation="stop", namespace="production",
                           grace_period=10, force=true)

        # Development cycle
        virtctl_vm_lifecycle(vm_name="dev-vm", operation="pause", namespace="development")    # Free resources
        # Later...
        virtctl_vm_lifecycle(vm_name="dev-vm", operation="unpause", namespace="development")  # Resume work`,
	}
}

func HandleVirtctlVMLifecycle(ctx context.Context, req *mcp.CallToolRequest, input VirtctlVMLifecycleInput) (*mcp.CallToolResult, any, error) {
	// Validate operation
	validOps := map[string]bool{
		"start": true, "stop": true, "restart": true,
		"pause": true, "unpause": true, "migrate": true, "soft-reboot": true,
	}

	if !validOps[input.Operation] {
		return nil, "", fmt.Errorf("invalid operation: %s. Valid operations: start, stop, restart, pause, unpause, migrate, soft-reboot", input.Operation)
	}

	// Build virtctl command
	args := []string{input.Operation, input.VMName}

	// Resolve namespace
	namespace := mtvmcp.ResolveNamespace(input.Namespace)
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	// Add operation-specific arguments
	if (input.Operation == "stop" || input.Operation == "restart") && input.GracePeriod > 0 {
		args = append(args, "--grace-period", fmt.Sprintf("%d", input.GracePeriod))
	}

	if (input.Operation == "stop" || input.Operation == "restart") && input.Force {
		args = append(args, "--force")
	}

	if input.Operation == "migrate" && input.NodeName != "" {
		args = append(args, "--node", input.NodeName)
	}

	if input.DryRun {
		args = append(args, "--dry-run")
	}

	if input.Timeout != "" {
		args = append(args, "--timeout", input.Timeout)
	}

	// Execute command
	output, err := mtvmcp.RunVirtctlCommand(args)
	if err != nil {
		return nil, "", err
	}

	// Format response
	response := map[string]interface{}{
		"status":  "success",
		"message": "VM lifecycle operation completed successfully",
		"command": "virtctl " + strings.Join(args, " "),
		"output":  strings.TrimSpace(output),
	}

	result, _ := json.MarshalIndent(response, "", "  ")
	return nil, string(result), nil
}
