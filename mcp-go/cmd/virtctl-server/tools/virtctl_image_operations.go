package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/mcp-go/pkg/mtvmcp"
)

// VirtctlImageOperationsInput represents the input for VirtctlImageOperations
type VirtctlImageOperationsInput struct {
	Operation        string                 `json:"operation" jsonschema:"Image operation (upload export guestfs memory-dump)"`
	VMName           string                 `json:"vm_name,omitempty" jsonschema:"VM name (required for memory-dump export)"`
	Namespace        string                 `json:"namespace,omitempty" jsonschema:"Kubernetes namespace (optional)"`
	ImagePath        string                 `json:"image_path,omitempty" jsonschema:"Local image file path (required for upload)"`
	PVCName          string                 `json:"pvc_name,omitempty" jsonschema:"PVC name (required for upload guestfs)"`
	Size             string                 `json:"size,omitempty" jsonschema:"PVC size for new volumes (optional)"`
	StorageClass     string                 `json:"storage_class,omitempty" jsonschema:"Storage class for new PVCs (optional)"`
	UploadConfig     map[string]interface{} `json:"upload_config,omitempty" jsonschema:"Upload configuration options (optional)"`
	GuestfsConfig    map[string]interface{} `json:"guestfs_config,omitempty" jsonschema:"Libguestfs configuration options (optional)"`
	MemoryDumpConfig map[string]interface{} `json:"memory_dump_config,omitempty" jsonschema:"Memory dump configuration (optional)"`
	ExportConfig     map[string]interface{} `json:"export_config,omitempty" jsonschema:"Export configuration options (optional)"`
}

// GetVirtctlImageOperationsTool returns the tool definition
func GetVirtctlImageOperationsTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "VirtctlImageOperations",
		Description: `Advanced disk and image management.

    This unified tool handles:
    - upload: Image uploads to PVCs/DataVolumes
    - export: VM exports and backups
    - guestfs: Libguestfs operations for disk inspection
    - memory-dump: Memory dumps for debugging

    Operations:

    Upload (virtctl image-upload):
    Uploads disk images to PVCs or DataVolumes for use with VMs.

    Required Parameters:
    - pvc_name: Target PVC name
    - image_path: Local path to image file to upload

    Optional Parameters:
    - size: PVC size (e.g., "10Gi") - creates new PVC if doesn't exist
    - storage_class: Storage class for new PVC
    - upload_config: Additional options (map):
        * access_mode: PVC access mode (ReadWriteOnce, ReadWriteMany)
        * volume_mode: Volume mode (Filesystem, Block)
        * insecure: Skip TLS verification (bool)
        * force_bind: Force bind CDI upload pod to target node (bool)
        * no_create: Don't create PVC, only upload to existing (bool)
        * block_volume: Upload to block volume (bool)
        * uploadproxy_url: Custom upload proxy URL (string)

    Export (virtctl vmexport):
    Creates and manages VM exports for backup and migration.

    Required Parameters:
    - vm_name: VM name to export

    Optional Parameters:
    - export_config: Export options (map):
        * action: Export action (create, delete, download, port-forward) [default: create]
        * output: Output file path for download action
        * manifest: Include export manifest (bool)
        * pvc: Export to PVC instead of temp storage
        * ttl: Time-to-live for export (e.g., "1h", "30m")
        * port: Local port for port-forward (int)

    Guestfs (virtctl guestfs):
    Launches a libguestfs pod for direct disk inspection and modification.

    Required Parameters:
    - pvc_name: PVC containing the disk to inspect

    Optional Parameters:
    - guestfs_config: Guestfs options (map):
        * kvm: Enable KVM acceleration (bool)
        * pull_method: Image pull method (node, pod)
        * root_disk_size: Root disk size for guestfs pod (e.g., "10Gi")

    Memory-dump (virtctl memory-dump):
    Creates memory dumps from running VMs for debugging.

    Required Parameters:
    - vm_name: VM name to dump memory from

    Optional Parameters:
    - memory_dump_config: Memory dump options (map):
        * claim_name: Target PVC name for dump
        * create_claim: Create new PVC for dump (bool)
        * volume_mode: Volume mode for new PVC (Filesystem, Block)
        * access_mode: Access mode for new PVC (ReadWriteOnce, ReadWriteMany)
        * storage_class: Storage class for new PVC

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

    Returns:
        Command output from virtctl operation

    Examples:
        # Upload image to PVC
        VirtctlImageOperations(operation="upload", pvc_name="my-disk", 
            image_path="/path/to/image.qcow2", size="10Gi")

        # Upload with custom configuration
        VirtctlImageOperations(operation="upload", pvc_name="my-disk",
            image_path="/path/to/image.qcow2",
            upload_config={"access_mode": "ReadWriteOnce", "volume_mode": "Block", "insecure": true})

        # Export VM
        VirtctlImageOperations(operation="export", vm_name="my-vm",
            export_config={"action": "create", "ttl": "1h"})

        # Launch guestfs for disk inspection
        VirtctlImageOperations(operation="guestfs", pvc_name="my-disk",
            guestfs_config={"kvm": true})

        # Create memory dump
        VirtctlImageOperations(operation="memory-dump", vm_name="my-vm",
            memory_dump_config={"claim_name": "vm-memory-dump", "create_claim": true})`,
	}
}

func HandleVirtctlImageOperations(ctx context.Context, req *mcp.CallToolRequest, input VirtctlImageOperationsInput) (*mcp.CallToolResult, any, error) {
	// Validate operation
	validOps := map[string]bool{"upload": true, "export": true, "guestfs": true, "memory-dump": true}
	if !validOps[input.Operation] {
		return nil, "", fmt.Errorf("invalid operation: %s. Valid operations: upload, export, guestfs, memory-dump", input.Operation)
	}

	// Validate required parameters and build command based on operation
	var args []string
	var err error

	switch input.Operation {
	case "upload":
		args, err = buildUploadCommand(input)
	case "export":
		args, err = buildExportCommand(input)
	case "guestfs":
		args, err = buildGuestfsCommand(input)
	case "memory-dump":
		args, err = buildMemoryDumpCommand(input)
	}

	if err != nil {
		return nil, "", err
	}

	// Add namespace if specified
	namespace := mtvmcp.ResolveNamespace(input.Namespace)
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	output, err := mtvmcp.RunVirtctlCommand(args)
	if err != nil {
		return nil, "", err
	}

	return nil, output, nil
}

// buildUploadCommand builds the command for image upload operation
func buildUploadCommand(input VirtctlImageOperationsInput) ([]string, error) {
	// Validate required parameters
	if err := mtvmcp.ValidateRequiredParams(map[string]string{
		"pvc_name":   input.PVCName,
		"image_path": input.ImagePath,
	}); err != nil {
		return nil, err
	}

	args := []string{"image-upload", input.PVCName}

	// Add image path
	args = append(args, "--image-path", input.ImagePath)

	// Add optional size
	if input.Size != "" {
		args = append(args, "--size", input.Size)
	}

	// Add optional storage class
	if input.StorageClass != "" {
		args = append(args, "--storage-class", input.StorageClass)
	}

	// Process upload_config options
	if input.UploadConfig != nil {
		if accessMode, ok := input.UploadConfig["access_mode"].(string); ok && accessMode != "" {
			args = append(args, "--access-mode", accessMode)
		}

		if volumeMode, ok := input.UploadConfig["volume_mode"].(string); ok && volumeMode != "" {
			args = append(args, "--volume-mode", volumeMode)
		}

		if insecure, ok := input.UploadConfig["insecure"].(bool); ok && insecure {
			args = append(args, "--insecure")
		}

		if forceBind, ok := input.UploadConfig["force_bind"].(bool); ok && forceBind {
			args = append(args, "--force-bind")
		}

		if noCreate, ok := input.UploadConfig["no_create"].(bool); ok && noCreate {
			args = append(args, "--no-create")
		}

		if blockVolume, ok := input.UploadConfig["block_volume"].(bool); ok && blockVolume {
			args = append(args, "--block-volume")
		}

		if uploadproxyURL, ok := input.UploadConfig["uploadproxy_url"].(string); ok && uploadproxyURL != "" {
			args = append(args, "--uploadproxy-url", uploadproxyURL)
		}
	}

	return args, nil
}

// buildExportCommand builds the command for VM export operation
func buildExportCommand(input VirtctlImageOperationsInput) ([]string, error) {
	// Validate required parameters
	if err := mtvmcp.ValidateRequiredParams(map[string]string{
		"vm_name": input.VMName,
	}); err != nil {
		return nil, err
	}

	// Default action is "create"
	action := "create"
	if input.ExportConfig != nil {
		if a, ok := input.ExportConfig["action"].(string); ok && a != "" {
			action = a
		}
	}

	args := []string{"vmexport", action, input.VMName}

	// Process export_config options
	if input.ExportConfig != nil {
		if output, ok := input.ExportConfig["output"].(string); ok && output != "" {
			args = append(args, "--output", output)
		}

		if manifest, ok := input.ExportConfig["manifest"].(bool); ok && manifest {
			args = append(args, "--manifest")
		}

		if pvc, ok := input.ExportConfig["pvc"].(string); ok && pvc != "" {
			args = append(args, "--pvc", pvc)
		}

		if ttl, ok := input.ExportConfig["ttl"].(string); ok && ttl != "" {
			args = append(args, "--ttl", ttl)
		}

		if port, ok := input.ExportConfig["port"]; ok {
			var portStr string
			switch v := port.(type) {
			case int:
				portStr = fmt.Sprintf("%d", v)
			case float64:
				portStr = fmt.Sprintf("%d", int(v))
			case string:
				portStr = v
			}
			if portStr != "" {
				args = append(args, "--port", portStr)
			}
		}
	}

	return args, nil
}

// buildGuestfsCommand builds the command for guestfs operation
func buildGuestfsCommand(input VirtctlImageOperationsInput) ([]string, error) {
	// Validate required parameters
	if err := mtvmcp.ValidateRequiredParams(map[string]string{
		"pvc_name": input.PVCName,
	}); err != nil {
		return nil, err
	}

	args := []string{"guestfs", input.PVCName}

	// Process guestfs_config options
	if input.GuestfsConfig != nil {
		if kvm, ok := input.GuestfsConfig["kvm"].(bool); ok && kvm {
			args = append(args, "--kvm")
		}

		if pullMethod, ok := input.GuestfsConfig["pull_method"].(string); ok && pullMethod != "" {
			args = append(args, "--pull-method", pullMethod)
		}

		if rootDiskSize, ok := input.GuestfsConfig["root_disk_size"].(string); ok && rootDiskSize != "" {
			args = append(args, "--root-disk-size", rootDiskSize)
		}
	}

	return args, nil
}

// buildMemoryDumpCommand builds the command for memory dump operation
func buildMemoryDumpCommand(input VirtctlImageOperationsInput) ([]string, error) {
	// Validate required parameters
	if err := mtvmcp.ValidateRequiredParams(map[string]string{
		"vm_name": input.VMName,
	}); err != nil {
		return nil, err
	}

	args := []string{"memory-dump", "get", input.VMName}

	// Process memory_dump_config options
	if input.MemoryDumpConfig != nil {
		if claimName, ok := input.MemoryDumpConfig["claim_name"].(string); ok && claimName != "" {
			args = append(args, "--claim-name", claimName)
		}

		if createClaim, ok := input.MemoryDumpConfig["create_claim"].(bool); ok && createClaim {
			args = append(args, "--create-claim")
		}

		if volumeMode, ok := input.MemoryDumpConfig["volume_mode"].(string); ok && volumeMode != "" {
			args = append(args, "--volume-mode", volumeMode)
		}

		if accessMode, ok := input.MemoryDumpConfig["access_mode"].(string); ok && accessMode != "" {
			args = append(args, "--access-mode", accessMode)
		}

		if storageClass, ok := input.MemoryDumpConfig["storage_class"].(string); ok && storageClass != "" {
			args = append(args, "--storage-class", storageClass)
		}
	}

	return args, nil
}
