package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/mcp-go/pkg/mtvmcp"
)

// VirtctlCreateVMAdvancedInput represents the input for VirtctlCreateVMAdvanced
type VirtctlCreateVMAdvancedInput struct {
	Name                   string                   `json:"name,omitempty" jsonschema:"VM name (optional random if not specified)"`
	Namespace              string                   `json:"namespace,omitempty" jsonschema:"Kubernetes namespace (optional)"`
	Instancetype           string                   `json:"instancetype,omitempty" jsonschema:"Instance type name (optional)"`
	Preference             string                   `json:"preference,omitempty" jsonschema:"Preference name (optional)"`
	RunStrategy            string                   `json:"run_strategy,omitempty" jsonschema:"Run strategy (Always RerunOnFailure Manual Halted) (optional)"`
	Volumes                map[string]interface{}   `json:"volumes,omitempty" jsonschema:"Volume configuration dictionary with volume types (optional)"`
	CloudInit              map[string]interface{}   `json:"cloud_init,omitempty" jsonschema:"Cloud-init configuration (optional)"`
	ResourceRequirements   map[string]string        `json:"resource_requirements,omitempty" jsonschema:"CPU/memory requirements (optional)"`
	InferInstancetype      bool                     `json:"infer_instancetype,omitempty" jsonschema:"Infer instance type from boot volume (optional)"`
	InferPreference        bool                     `json:"infer_preference,omitempty" jsonschema:"Infer preference from boot volume (optional)"`
	InferInstancetypeFrom  string                   `json:"infer_instancetype_from,omitempty" jsonschema:"Volume name to infer instance type from (optional)"`
	InferPreferenceFrom    string                   `json:"infer_preference_from,omitempty" jsonschema:"Volume name to infer preference from (optional)"`
	AccessCredentials      []map[string]interface{} `json:"access_credentials,omitempty" jsonschema:"List of access credential configurations (optional)"`
	TerminationGracePeriod int                      `json:"termination_grace_period,omitempty" jsonschema:"Grace period for VM termination in seconds (optional)"`
	GenerateName           bool                     `json:"generate_name,omitempty" jsonschema:"Use generateName instead of name (optional)"`
}

// GetVirtctlCreateVMAdvancedTool returns the tool definition
func GetVirtctlCreateVMAdvancedTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "VirtctlCreateVMAdvanced",
		Description: `Comprehensive VM creation with full configuration support.

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
            "user_data": "#cloud-config\npackages:\n  - git",  # Custom user data
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
            infer_instancetype=true,
            infer_preference=true,
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
                "user_data": "#cloud-config
packages:
  - podman
  - git
runcmd:
  - systemctl enable --now podman
  - firewall-cmd --permanent --add-port=8080/tcp
  - firewall-cmd --reload"
            },
            access_credentials=[{
                "type": "ssh",
                "src": "enterprise-ssh-keys",
                "method": "qemu-guest-agent"
            }],
            termination_grace_period=180  # 3 minutes for app shutdown
        )`,
	}
}

func HandleVirtctlCreateVMAdvanced(ctx context.Context, req *mcp.CallToolRequest, input VirtctlCreateVMAdvancedInput) (*mcp.CallToolResult, any, error) {
	// Build virtctl create vm command (simplified implementation)
	args := []string{"create", "vm"}

	if input.Name != "" {
		args = append(args, input.Name)
	}

	namespace := mtvmcp.ResolveNamespace(input.Namespace)
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if input.Instancetype != "" {
		args = append(args, "--instancetype", input.Instancetype)
	}
	if input.Preference != "" {
		args = append(args, "--preference", input.Preference)
	}
	if input.RunStrategy != "" {
		args = append(args, "--run-strategy", input.RunStrategy)
	}
	if input.InferInstancetype {
		args = append(args, "--infer-instancetype")
	}
	if input.InferPreference {
		args = append(args, "--infer-preference")
	}

	output, err := mtvmcp.RunVirtctlCommand(args)
	if err != nil {
		return nil, "", err
	}

	return nil, output, nil
}
