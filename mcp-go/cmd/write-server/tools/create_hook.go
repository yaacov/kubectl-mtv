package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/mcp-go/pkg/mtvmcp"
)

// CreateHookInput represents the input for CreateHook
type CreateHookInput struct {
	HookName       string `json:"hook_name" jsonschema:"required"`
	Image          string `json:"image" jsonschema:"required"`
	Namespace      string `json:"namespace,omitempty"`
	ServiceAccount string `json:"service_account,omitempty"`
	Playbook       string `json:"playbook,omitempty"`
	Deadline       int    `json:"deadline,omitempty"`
}

// GetCreateHookTool returns the tool definition
func GetCreateHookTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "CreateHook",
		Description: `Create a migration hook for custom automation during migrations.

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
                   playbook="- name: Check connectivity\n  ping:\n    data: test")

        # Create hook with playbook from file
        create_hook("post-migration-cleanup", "my-registry/ansible:latest",
                   playbook="@/path/to/cleanup-playbook.yaml",
                   service_account="migration-hooks", deadline=300)

        # Create simple validation hook
        create_hook("validate-target", "my-registry/validator:latest",
                   service_account="migration-validator", deadline=600)`,
	}
}

func HandleCreateHook(ctx context.Context, req *mcp.CallToolRequest, input CreateHookInput) (*mcp.CallToolResult, any, error) {
	// Validate required parameters
	if err := mtvmcp.ValidateRequiredParams(map[string]string{
		"hook_name": input.HookName,
		"image":     input.Image,
	}); err != nil {
		return nil, "", err
	}

	args := []string{"create", "hook", input.HookName}

	if input.Namespace != "" {
		args = append(args, "-n", input.Namespace)
	}

	args = append(args, "--image", input.Image)

	if input.ServiceAccount != "" {
		args = append(args, "--service-account", input.ServiceAccount)
	}
	if input.Playbook != "" {
		args = append(args, "--playbook", input.Playbook)
	}
	if input.Deadline > 0 {
		args = append(args, "--deadline", fmt.Sprintf("%d", input.Deadline))
	}

	result, err := mtvmcp.RunKubectlMTVCommand(args)
	if err != nil {
		return nil, "", err
	}

	// Unmarshal the full CommandResponse to provide complete diagnostic information
	data, err := mtvmcp.UnmarshalJSONResponse(result)
	if err != nil {
		return nil, "", err
	}
	return nil, data, nil
}
