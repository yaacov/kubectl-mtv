package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/mcp-go/pkg/mtvmcp"
)

// VirtctlDataSourceManagementInput represents the input for VirtctlDataSourceManagement
type VirtctlDataSourceManagementInput struct {
	Operation    string                 `json:"operation" jsonschema:"DataSource operation (create list clone)"`
	Name         string                 `json:"name,omitempty" jsonschema:"DataSource name (required for create)"`
	Namespace    string                 `json:"namespace,omitempty" jsonschema:"Kubernetes namespace (optional)"`
	SourceConfig map[string]interface{} `json:"source_config,omitempty" jsonschema:"Source configuration for DataSource creation (optional)"`
}

// GetVirtctlDataSourceManagementTool returns the tool definition
func GetVirtctlDataSourceManagementTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "VirtctlDataSourceManagement",
		Description: `List available DataSources for VM boot images.

    DataSources provide ready-to-use OS images that can be used as boot sources
    when creating VMs. They often include metadata annotations that enable automatic
    instance type and preference selection.

    What are DataSources?
    - Pre-imported VM images (from HTTP URLs, container registries, or PVCs)
    - Contain OS images ready for VM creation
    - May include annotations for automatic resource configuration:
      * instancetype.kubevirt.io/default-instancetype: suggested compute resources
      * instancetype.kubevirt.io/default-preference: OS-specific optimizations

    This tool lists existing DataSources to help you discover available boot images.

    Operations:
    - list: Display available DataSources in JSON format

    Args:
        operation: Must be "list" (required)
        namespace: Kubernetes namespace to query (optional, omit for current namespace)

    Returns:
        JSON list of DataSource objects with their specifications, status, and metadata

    Example:
        # List all DataSources in current namespace
        VirtctlDataSourceManagement(operation="list")

        # List DataSources in specific namespace
        VirtctlDataSourceManagement(operation="list", namespace="openshift-virtualization-os-images")`,
	}
}

func HandleVirtctlDataSourceManagement(ctx context.Context, req *mcp.CallToolRequest, input VirtctlDataSourceManagementInput) (*mcp.CallToolResult, any, error) {
	// Validate required parameters
	if err := mtvmcp.ValidateRequiredParams(map[string]string{
		"operation": input.Operation,
	}); err != nil {
		return nil, "", err
	}

	var args []string
	switch input.Operation {
	case "list":
		args = []string{"get", "datasource"}
	case "create":
		return nil, "", fmt.Errorf("create not implemented; generate a CDI DataSource manifest and apply with kubectl")
	case "clone":
		return nil, "", fmt.Errorf("clone not implemented")
	default:
		return nil, "", fmt.Errorf("invalid operation: %s. Valid operations: list", input.Operation)
	}

	namespace := mtvmcp.ResolveNamespace(input.Namespace)
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	// list via kubectl and return JSON
	args = append(args, "-o", "json")
	result, err := mtvmcp.RunKubectlCommand(args)
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
