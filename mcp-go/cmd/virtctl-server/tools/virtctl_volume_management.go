package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/mcp-go/pkg/mtvmcp"
)

// VirtctlVolumeManagementInput represents the input for VirtctlVolumeManagement
type VirtctlVolumeManagementInput struct {
	VMName     string `json:"vm_name" jsonschema:"Name of the virtual machine"`
	Operation  string `json:"operation" jsonschema:"Volume operation (add remove list)"`
	Namespace  string `json:"namespace,omitempty" jsonschema:"Kubernetes namespace containing the VM (optional)"`
	VolumeName string `json:"volume_name,omitempty" jsonschema:"Name of the volume (required for add/remove)"`
	Persist    bool   `json:"persist,omitempty" jsonschema:"Persist volume changes to VM spec (optional)"`
	DryRun     bool   `json:"dry_run,omitempty" jsonschema:"Show what would be done without executing (optional)"`
}

// GetVirtctlVolumeManagementTool returns the tool definition
func GetVirtctlVolumeManagementTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "VirtctlVolumeManagement",
		Description: `Hot-plug storage operations for running VMs.

    IMPORTANT LIMITATION:
    virtctl addvolume/removevolume only support:
    - --volume-name (required): Name for the volume in the VM
    - --persist (optional): Make the volume survive VM restarts
    - --serial (optional): Serial number for the volume

    The underlying PVC/DataVolume must be created separately using kubectl.
    This tool ONLY attaches/detaches existing PVCs to/from running VMs.

    PREREQUISITES:
    - VM must be running: kubectl get vmi {vm_name} -n {namespace}
    - For add operation: Target PVC must exist and be available
    - Hot-plug capability: VM must support virtio or SCSI hot-plug

    OPERATION GUIDE:

    ADD: Hot-plug attach existing PVC to running VM
    - Attaches an existing PVC to a running VM
    - PVC must be pre-created (use kubectl create/apply)
    - VM sees new disk immediately (check with lsblk inside VM)
    - Use persist=True to make volume survive VM restarts
    - volume_name is the name the volume will have inside the VM spec

    REMOVE: Hot-plug detach storage from running VM
    - Removes the volume from the running VM
    - VM immediately loses access to disk
    - Data on PVC persists after removal
    - Always unmount inside VM before removing
    - PVC is not deleted (use kubectl delete pvc if needed)

    LIST: Show all volumes currently attached to VM
    - Shows both hot-plugged and permanent volumes in VMI JSON spec
    - Implementation: Uses kubectl to get VMI (virtctl has no volumes list command)
    - Returns full VMI JSON including .spec.volumes and .spec.domain.devices.disks

    TROUBLESHOOTING:

    Common Issues:
    - "Volume already exists": Choose different volume_name
    - "PVC not found": Create PVC first with kubectl
    - "VM not found": Verify VM is running (not just created)
    - "Permission denied": Check PVC access modes and pod security

    Best Practices:
    - Create PVC/DataVolume first using kubectl before adding
    - Use descriptive volume names: "mysql-data", "nginx-logs"
    - Use persist=True for permanent storage
    - Unmount inside VM before removing volumes
    - PVCs remain after removal; delete manually if needed

    Pre-operation Checks:
    - VM status: kubectl get vmi {vm_name} -n {namespace}
    - PVC availability: kubectl get pvc -n {namespace}
    - Disk space in VM: df -h (inside VM)

    Args:
        vm_name: Name of the virtual machine
        operation: Volume operation (add, remove, list)
        namespace: Kubernetes namespace containing the VM (optional)
        volume_name: Name of the volume in VM spec (required for add/remove)
        persist: Persist volume changes to VM spec (optional)
        dry_run: Show what would be done without executing (optional)

    Common Workflows:
        # Step 1: Create PVC (using kubectl, not this tool)
        # kubectl create -f my-pvc.yaml

        # Step 2: Attach existing PVC to VM
        virtctl_volume_management(vm_name="app-vm", operation="add", namespace="production",
                                volume_name="data-disk", persist=true)

        # Step 3: Use the volume inside the VM
        # Inside VM: lsblk, mount /dev/vdb /mnt/data

        # Step 4: Remove volume when done
        # Inside VM: umount /mnt/data
        virtctl_volume_management(vm_name="app-vm", operation="remove", namespace="production",
                                volume_name="data-disk")

        # Volume inventory check
        virtctl_volume_management(vm_name="database-vm", operation="list", namespace="production")`,
	}
}

func HandleVirtctlVolumeManagement(ctx context.Context, req *mcp.CallToolRequest, input VirtctlVolumeManagementInput) (*mcp.CallToolResult, any, error) {
	// Validate required parameters
	if err := mtvmcp.ValidateRequiredParams(map[string]string{
		"vm_name":   input.VMName,
		"operation": input.Operation,
	}); err != nil {
		return nil, "", err
	}

	// Validate operation
	validOps := map[string]bool{"add": true, "remove": true, "list": true}
	if !validOps[input.Operation] {
		return nil, "", fmt.Errorf("invalid operation: %s. Valid operations: add, remove, list", input.Operation)
	}

	// Build command based on operation
	var args []string
	if input.Operation == "list" {
		// Use kubectl to get VMI spec since virtctl doesn't have a volumes list command
		namespace := mtvmcp.ResolveNamespace(input.Namespace)
		kubectlArgs := []string{"get", "vmi", input.VMName}
		if namespace != "" {
			kubectlArgs = append(kubectlArgs, "-n", namespace)
		}
		kubectlArgs = append(kubectlArgs, "-o", "json")

		result, err := mtvmcp.RunKubectlCommand(kubectlArgs)
		if err != nil {
			return nil, "", err
		}

		// Unmarshal the full CommandResponse to provide complete diagnostic information
		data, err := mtvmcp.UnmarshalJSONResponse(result)
		if err != nil {
			return nil, "", err
		}
		return nil, data, nil
	} else if input.Operation == "add" {
		args = []string{"addvolume", input.VMName}

		// Validate required parameters for add operation
		if input.VolumeName == "" {
			return nil, "", fmt.Errorf("volume_name is required for add operation")
		}
		args = append(args, "--volume-name", input.VolumeName)

		// NOTE: virtctl addvolume supports --volume-name and optional --persist/--serial.
		// Creating/choosing the underlying PVC/DV must be handled outside this command.

		if input.Persist {
			args = append(args, "--persist")
		}
	} else { // remove
		args = []string{"removevolume", input.VMName}

		// Validate required parameters for remove operation
		if input.VolumeName == "" {
			return nil, "", fmt.Errorf("volume_name is required for remove operation")
		}
		args = append(args, "--volume-name", input.VolumeName)
	}

	namespace := mtvmcp.ResolveNamespace(input.Namespace)
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	// Some virtctl versions support per-command --dry-run; verify before adding.
	if input.DryRun {
		args = append(args, "--dry-run")
	}

	output, err := mtvmcp.RunVirtctlCommand(args)
	if err != nil {
		return nil, "", err
	}

	return nil, output, nil
}
