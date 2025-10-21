package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/mcp-go/pkg/mtvmcp"
)

// DeletePlanInput represents the input for DeletePlan
type DeletePlanInput struct {
	PlanName    string `json:"plan_name,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	AllPlans    bool   `json:"all_plans,omitempty"`
	SkipArchive bool   `json:"skip_archive,omitempty"`
	CleanAll    bool   `json:"clean_all,omitempty"`
}

// GetDeletePlanTool returns the tool definition
func GetDeletePlanTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "DeletePlan",
		Description: `Delete one or more migration plans.

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
        DeletePlan(plan_name="my-plan")

        # Delete plan without archiving
        DeletePlan(plan_name="my-plan", skip_archive=true)

        # Delete plan with VM cleanup on failure
        DeletePlan(plan_name="my-plan", clean_all=true)

        # Delete all plans in namespace
        DeletePlan(all_plans=true, namespace="demo")`,
	}
}

func HandleDeletePlan(ctx context.Context, req *mcp.CallToolRequest, input DeletePlanInput) (*mcp.CallToolResult, any, error) {
	args := []string{"delete", "plan"}

	if input.AllPlans {
		args = append(args, "--all")
	} else {
		if input.PlanName == "" {
			return nil, "", fmt.Errorf("plan_name is required when all_plans=false")
		}
		args = append(args, input.PlanName)
	}

	if input.Namespace != "" {
		args = append(args, "-n", input.Namespace)
	}

	if input.SkipArchive {
		args = append(args, "--skip-archive")
	}

	if input.CleanAll {
		args = append(args, "--clean-all")
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
