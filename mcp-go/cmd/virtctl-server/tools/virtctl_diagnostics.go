package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/mcp-go/pkg/mtvmcp"
)

// VirtctlDiagnosticsInput represents the input for VirtctlDiagnostics
type VirtctlDiagnosticsInput struct {
	DiagnosticType string `json:"diagnostic_type" jsonschema:"Diagnostic operation (guestosinfo fslist userlist version)"`
	VMName         string `json:"vm_name,omitempty" jsonschema:"Name of the virtual machine (optional not required for version operation)"`
	Namespace      string `json:"namespace,omitempty" jsonschema:"Kubernetes namespace containing the VM (optional)"`
}

// GetVirtctlDiagnosticsTool returns the tool definition
func GetVirtctlDiagnosticsTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "VirtctlDiagnostics",
		Description: `Comprehensive VM diagnostics and monitoring capabilities.

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
    1. virtctl_diagnostics(diagnostic_type="version") → Verify tool compatibility
    2. virtctl_diagnostics(diagnostic_type="guestosinfo", vm_name=vm) → Confirm guest agent connectivity
    3. virtctl_diagnostics(diagnostic_type="fslist", vm_name=vm) → Check storage health
    4. virtctl_diagnostics(diagnostic_type="userlist", vm_name=vm) → Verify access control

    Troubleshooting Sequence:
    1. Check VM status: kubectl get vmi {vm_name} -n {namespace}
    2. Check guest agent: virtctl_diagnostics(diagnostic_type="guestosinfo", vm_name=vm)
    3. If agent fails: Access console and check agent service
    4. Storage issues: virtctl_diagnostics(diagnostic_type="fslist", vm_name=vm)
    5. Access issues: virtctl_diagnostics(diagnostic_type="userlist", vm_name=vm)

    Args:
        diagnostic_type: Diagnostic operation (guestosinfo, fslist, userlist, version)
        vm_name: Name of the virtual machine (optional, not required for version operation)
        namespace: Kubernetes namespace containing the VM (optional)

    Returns:
        Diagnostic information about the VM or system

    Troubleshooting Examples:
        # Quick health check
        virtctl_diagnostics(diagnostic_type="guestosinfo", vm_name="prod-vm", namespace="production")

        # Storage investigation
        virtctl_diagnostics(diagnostic_type="fslist", vm_name="db-vm", namespace="production")

        # Security audit
        virtctl_diagnostics(diagnostic_type="userlist", vm_name="web-vm", namespace="production")

        # Tool compatibility check
        virtctl_diagnostics(diagnostic_type="version")`,
	}
}

func HandleVirtctlDiagnostics(ctx context.Context, req *mcp.CallToolRequest, input VirtctlDiagnosticsInput) (*mcp.CallToolResult, any, error) {
	// Validate diagnostic type
	validTypes := map[string]bool{
		"guestosinfo": true, "fslist": true, "userlist": true, "version": true,
	}

	if !validTypes[input.DiagnosticType] {
		return nil, "", fmt.Errorf("invalid diagnostic_type: %s. Valid types: guestosinfo, fslist, userlist, version", input.DiagnosticType)
	}

	// Build virtctl command
	var args []string

	if input.DiagnosticType == "version" {
		args = []string{"version"}
	} else {
		args = []string{input.DiagnosticType, input.VMName}

		// Resolve namespace
		namespace := mtvmcp.ResolveNamespace(input.Namespace)
		if namespace != "" {
			args = append(args, "-n", namespace)
		}
	}

	// Execute command
	output, err := mtvmcp.RunVirtctlCommand(args)
	if err != nil {
		return nil, "", err
	}

	// Try to parse as JSON if possible
	formatted, _ := mtvmcp.ParseJSONOutput(output)
	return nil, formatted, nil
}
