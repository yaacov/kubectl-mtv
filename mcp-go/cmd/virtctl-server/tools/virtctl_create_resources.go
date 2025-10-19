package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/mcp-go/pkg/mtvmcp"
)

// VirtctlCreateResourcesInput represents the input for VirtctlCreateResources
type VirtctlCreateResourcesInput struct {
	ResourceType string `json:"resource_type" jsonschema:"Type of resource (instancetype preference)"`
	Name         string `json:"name" jsonschema:"Name of the resource"`
	Namespaced   bool   `json:"namespaced,omitempty" jsonschema:"Create namespaced resource (default: cluster-scoped)"`
	Namespace    string `json:"namespace,omitempty" jsonschema:"Kubernetes namespace (required if namespaced=True)"`
}

// GetVirtctlCreateResourcesTool returns the tool definition
func GetVirtctlCreateResourcesTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "VirtctlCreateResources",
		Description: `Create basic instance type or preference resource templates by name.

    This tool creates empty/default instance types or preferences that can be
    further configured using kubectl or other tools. The resources are created
    with minimal default settings.

    Resource Types:
    - instancetype: Defines CPU and memory allocations for VMs
    - preference: Defines VM behavior, features, and hardware characteristics

    Scope:
    - Cluster-scoped (default): Available to all namespaces
    - Namespaced: Specific to one namespace when namespaced=True

    Note: This tool creates resources by name only. Advanced configuration
    (CPU topology, memory settings, GPU resources, firmware options, etc.)
    must be applied after creation using kubectl patch or edit.

    Args:
        resource_type: Type of resource - 'instancetype' or 'preference' (required)
        name: Name of the resource (required)
        namespaced: Create namespaced resource instead of cluster-scoped (optional, default False)
        namespace: Kubernetes namespace (required if namespaced=True)

    Examples:
        # Create cluster-scoped instance type
        VirtctlCreateResources(resource_type="instancetype", name="u1.small")

        # Create cluster-scoped preference
        VirtctlCreateResources(resource_type="preference", name="fedora-server")

        # Create namespaced instance type
        VirtctlCreateResources(resource_type="instancetype", name="dev-small",
                              namespaced=True, namespace="development")

        # Create namespaced preference
        VirtctlCreateResources(resource_type="preference", name="windows-desktop",
                              namespaced=True, namespace="testing")`,
	}
}

func HandleVirtctlCreateResources(ctx context.Context, req *mcp.CallToolRequest, input VirtctlCreateResourcesInput) (*mcp.CallToolResult, any, error) {
	// Validate resource type
	validTypes := map[string]bool{"instancetype": true, "preference": true}
	if !validTypes[input.ResourceType] {
		return nil, "", fmt.Errorf("invalid resource_type: %s. Valid types: instancetype, preference", input.ResourceType)
	}

	// Build command (simplified - would need more detail in production)
	args := []string{"create", input.ResourceType, input.Name}

	if input.Namespaced {
		namespace := mtvmcp.ResolveNamespace(input.Namespace)
		if namespace != "" {
			args = append(args, "-n", namespace)
		}
	}

	output, err := mtvmcp.RunVirtctlCommand(args)
	if err != nil {
		return nil, "", err
	}

	return nil, output, nil
}
