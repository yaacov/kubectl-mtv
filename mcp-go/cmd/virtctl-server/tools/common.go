package tools

import (
	"encoding/json"

	"github.com/yaacov/kubectl-mtv/mcp-go/pkg/mtvmcp"
)

// getResourceType handles resource type queries for cluster resources
func getResourceType(resourceType, scope, namespace, labelSelector string, showLabels bool) (string, error) {
	// Map resource types to kubectl get commands
	var args []string

	switch resourceType {
	case "instancetypes":
		if scope == "namespaced" {
			args = []string{"get", "virtualmachineinstancetype"}
		} else if scope == "cluster" {
			args = []string{"get", "virtualmachineclusterinstancetype"}
		} else {
			// Get both
			clusterResult, _ := mtvmcp.RunKubectlCommand([]string{"get", "virtualmachineclusterinstancetype", "-o", "json"})
			namespResult, _ := mtvmcp.RunKubectlCommand([]string{"get", "virtualmachineinstancetype", "-A", "-o", "json"})

			// Extract stdout from CommandResponse wrappers
			clusterOut := mtvmcp.ExtractStdoutFromResponse(clusterResult)
			namespOut := mtvmcp.ExtractStdoutFromResponse(namespResult)

			combined := map[string]interface{}{
				"cluster":    json.RawMessage(clusterOut),
				"namespaced": json.RawMessage(namespOut),
			}
			result, _ := json.MarshalIndent(combined, "", "  ")
			return string(result), nil
		}

	case "preferences":
		if scope == "namespaced" {
			args = []string{"get", "virtualmachinepreference"}
		} else if scope == "cluster" {
			args = []string{"get", "virtualmachineclusterpreference"}
		} else {
			// Get both
			clusterResult, _ := mtvmcp.RunKubectlCommand([]string{"get", "virtualmachineclusterpreference", "-o", "json"})
			namespResult, _ := mtvmcp.RunKubectlCommand([]string{"get", "virtualmachinepreference", "-A", "-o", "json"})

			// Extract stdout from CommandResponse wrappers
			clusterOut := mtvmcp.ExtractStdoutFromResponse(clusterResult)
			namespOut := mtvmcp.ExtractStdoutFromResponse(namespResult)

			combined := map[string]interface{}{
				"cluster":    json.RawMessage(clusterOut),
				"namespaced": json.RawMessage(namespOut),
			}
			result, _ := json.MarshalIndent(combined, "", "  ")
			return string(result), nil
		}

	case "datasources":
		args = []string{"get", "datasource"}

	case "storageclasses":
		args = []string{"get", "storageclass"}
	}

	// Add namespace if applicable
	if namespace != "" && resourceType != "storageclasses" {
		args = append(args, "-n", namespace)
	} else if resourceType != "storageclasses" && scope != "cluster" {
		args = append(args, "-A")
	}

	// Add label selector
	if labelSelector != "" {
		args = append(args, "-l", labelSelector)
	}

	// Add show-labels
	if showLabels {
		args = append(args, "--show-labels")
	}

	// Add JSON output
	args = append(args, "-o", "json")

	// Execute command
	result, err := mtvmcp.RunKubectlCommand(args)
	if err != nil {
		return "", err
	}

	// Extract stdout from the CommandResponse wrapper
	output := mtvmcp.ExtractStdoutFromResponse(result)
	return output, nil
}
