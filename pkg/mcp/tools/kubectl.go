package tools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/pkg/mcp/util"
)

// KubectlDebugInput represents the input for the kubectl_debug tool.
type KubectlDebugInput struct {
	// Action is the kubectl action to perform: "logs", "get", "describe", or "events"
	Action string `json:"action" jsonschema:"kubectl action: logs for pod logs, get for listing resources, describe for resource details, events for event querying"`

	// ResourceType is the Kubernetes resource type (for get/describe actions)
	ResourceType string `json:"resource_type,omitempty" jsonschema:"Resource type for get/describe (e.g. pods, pvc, datavolume, virtualmachine, events)"`

	// Name is the specific resource name (optional for get, required for logs)
	Name string `json:"name,omitempty" jsonschema:"Resource name (required for logs, optional for get/describe)"`

	// Namespace is the Kubernetes namespace
	Namespace string `json:"namespace,omitempty" jsonschema:"Target Kubernetes namespace"`

	// AllNamespaces queries across all namespaces
	AllNamespaces bool `json:"all_namespaces,omitempty" jsonschema:"Query across all namespaces (for get action)"`

	// Labels is a label selector for filtering resources
	Labels string `json:"labels,omitempty" jsonschema:"Label selector (e.g. plan=my-plan,vmID=vm-123)"`

	// Container specifies which container to get logs from
	Container string `json:"container,omitempty" jsonschema:"Container name for logs (when pod has multiple containers)"`

	// Previous gets logs from the previous container instance
	Previous bool `json:"previous,omitempty" jsonschema:"Get logs from previous container instance (for crashed containers)"`

	// TailLines limits the number of log lines returned
	TailLines int `json:"tail_lines,omitempty" jsonschema:"Number of log lines to return from the end (default: all)"`

	// Since returns logs newer than a relative duration (e.g., "1h", "30m")
	Since string `json:"since,omitempty" jsonschema:"Return logs newer than duration (e.g. 1h, 30m, 5s)"`

	// Output format for get/describe (json, yaml, wide, or name)
	Output string `json:"output,omitempty" jsonschema:"Output format: json, yaml, wide, name (default: json for get)"`

	// DryRun shows the command without executing
	DryRun bool `json:"dry_run,omitempty" jsonschema:"Show command without executing (educational mode)"`

	// FieldSelector filters resources by field (for events action)
	FieldSelector string `json:"field_selector,omitempty" jsonschema:"Field selector for events (e.g. involvedObject.name=my-pod, type=Warning, reason=FailedScheduling)"`

	// SortBy sorts the output by a JSONPath expression (for events action)
	SortBy string `json:"sort_by,omitempty" jsonschema:"Sort events by JSONPath (e.g. .lastTimestamp, .metadata.creationTimestamp)"`

	// ForResource gets events for a specific resource (for events action)
	ForResource string `json:"for_resource,omitempty" jsonschema:"Get events for a specific resource (e.g. pod/my-pod, pvc/my-pvc)"`
}

// GetKubectlDebugTool returns the tool definition for kubectl debugging.
func GetKubectlDebugTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "kubectl_debug",
		Description: `Debug MTV migrations using standard kubectl commands.

This tool provides access to kubectl for debugging migration issues:

Actions:
- logs: Get pod logs (useful for forklift-controller, virt-v2v pods)
- get: List Kubernetes resources (pods, pvc, datavolume, virtualmachine)
- describe: Get detailed resource information
- events: Get Kubernetes events with specialized filtering for debugging

Common use cases:
- Get forklift controller logs: action="logs", name="forklift-controller-xxx", namespace="openshift-mtv"
- List migration pods: action="get", resource_type="pods", labels="plan=my-plan"
- Check PVC status: action="get", resource_type="pvc", labels="migration=xxx"
- Debug failed pod: action="logs", name="virt-v2v-xxx", previous=true

Events examples:
- Get events for a pod: action="events", for_resource="pod/virt-v2v-xxx", namespace="target-ns"
- Get warning events: action="events", field_selector="type=Warning", namespace="target-ns"
- Get events sorted by time: action="events", sort_by=".lastTimestamp", namespace="target-ns"
- Get events for failed scheduling: action="events", field_selector="reason=FailedScheduling"

Tips:
- Use labels to filter resources related to specific migrations
- Use tail_lines to limit log output
- Use previous=true to get logs from crashed containers
- Use since to get recent logs (e.g., "1h" for last hour)
- Use for_resource to get events related to a specific pod or PVC

IMPORTANT: When responding, always start by showing the user the executed command from the 'command' field in the response (e.g., "Executed: kubectl get pods -n openshift-mtv").`,
	}
}

// HandleKubectlDebug handles the kubectl_debug tool invocation.
func HandleKubectlDebug(ctx context.Context, req *mcp.CallToolRequest, input KubectlDebugInput) (*mcp.CallToolResult, any, error) {
	// Extract K8s credentials from HTTP headers (for SSE mode)
	if req.Extra != nil && req.Extra.Header != nil {
		ctx = util.WithKubeCredsFromHeaders(ctx, req.Extra.Header)
	}

	// Enable dry run mode if requested
	if input.DryRun {
		ctx = util.WithDryRun(ctx, true)
	}

	var args []string

	switch input.Action {
	case "logs":
		// Logs action requires a pod name
		if input.Name == "" {
			return nil, nil, fmt.Errorf("logs action requires 'name' field (pod name)")
		}
		args = buildLogsArgs(input)
	case "get":
		// Get action requires a resource type
		if input.ResourceType == "" {
			return nil, nil, fmt.Errorf("get action requires 'resource_type' field (e.g., pods, pvc, events)")
		}
		args = buildGetArgs(input)
	case "describe":
		// Describe action requires a resource type
		if input.ResourceType == "" {
			return nil, nil, fmt.Errorf("describe action requires 'resource_type' field (e.g., pods, pvc, events)")
		}
		args = buildDescribeArgs(input)
	case "events":
		// Events action - specialized event querying
		args = buildEventsArgs(input)
	default:
		return nil, nil, fmt.Errorf("unknown action '%s'. Valid actions: logs, get, describe, events", input.Action)
	}

	// Execute kubectl command
	result, err := util.RunKubectlCommand(ctx, args)
	if err != nil {
		return nil, nil, fmt.Errorf("kubectl command failed: %w", err)
	}

	// Parse and return result
	data, err := util.UnmarshalJSONResponse(result)
	if err != nil {
		return nil, nil, err
	}

	return nil, data, nil
}

// buildLogsArgs builds arguments for kubectl logs command.
func buildLogsArgs(input KubectlDebugInput) []string {
	args := []string{"logs"}

	// Pod name is required for logs
	if input.Name != "" {
		args = append(args, input.Name)
	}

	// Namespace
	if input.Namespace != "" {
		args = append(args, "-n", input.Namespace)
	}

	// Container
	if input.Container != "" {
		args = append(args, "-c", input.Container)
	}

	// Previous container logs
	if input.Previous {
		args = append(args, "--previous")
	}

	// Tail lines
	if input.TailLines > 0 {
		args = append(args, "--tail", strconv.Itoa(input.TailLines))
	}

	// Since duration
	if input.Since != "" {
		args = append(args, "--since", input.Since)
	}

	return args
}

// buildGetArgs builds arguments for kubectl get command.
func buildGetArgs(input KubectlDebugInput) []string {
	args := []string{"get"}

	// Resource type
	if input.ResourceType != "" {
		args = append(args, input.ResourceType)
	}

	// Resource name (optional)
	if input.Name != "" {
		args = append(args, input.Name)
	}

	// Namespace
	if input.AllNamespaces {
		args = append(args, "-A")
	} else if input.Namespace != "" {
		args = append(args, "-n", input.Namespace)
	}

	// Label selector
	if input.Labels != "" {
		args = append(args, "-l", input.Labels)
	}

	// Output format - default to json
	output := input.Output
	if output == "" {
		output = "json"
	}
	args = append(args, "-o", output)

	return args
}

// buildDescribeArgs builds arguments for kubectl describe command.
func buildDescribeArgs(input KubectlDebugInput) []string {
	args := []string{"describe"}

	// Resource type
	if input.ResourceType != "" {
		args = append(args, input.ResourceType)
	}

	// Resource name (optional)
	if input.Name != "" {
		args = append(args, input.Name)
	}

	// Namespace
	if input.AllNamespaces {
		args = append(args, "-A")
	} else if input.Namespace != "" {
		args = append(args, "-n", input.Namespace)
	}

	// Label selector
	if input.Labels != "" {
		args = append(args, "-l", input.Labels)
	}

	return args
}

// buildEventsArgs builds arguments for kubectl get events command with specialized filtering.
func buildEventsArgs(input KubectlDebugInput) []string {
	args := []string{"get", "events"}

	// Namespace
	if input.AllNamespaces {
		args = append(args, "-A")
	} else if input.Namespace != "" {
		args = append(args, "-n", input.Namespace)
	}

	// For a specific resource (e.g., --for pod/my-pod)
	if input.ForResource != "" {
		args = append(args, "--for", input.ForResource)
	}

	// Field selector (e.g., involvedObject.name=my-pod, type=Warning)
	if input.FieldSelector != "" {
		args = append(args, "--field-selector", input.FieldSelector)
	}

	// Sort by (e.g., .lastTimestamp)
	if input.SortBy != "" {
		args = append(args, "--sort-by", input.SortBy)
	}

	// Output format - default to json
	output := input.Output
	if output == "" {
		output = "json"
	}
	args = append(args, "-o", output)

	return args
}
