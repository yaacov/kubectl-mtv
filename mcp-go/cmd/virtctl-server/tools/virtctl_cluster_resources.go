package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// VirtctlClusterResourcesInput represents the input for VirtctlClusterResources
type VirtctlClusterResourcesInput struct {
	ResourceType  string `json:"resource_type" jsonschema:"Type of resource (instancetypes preferences datasources storageclasses all)"`
	Scope         string `json:"scope,omitempty" jsonschema:"Resource scope (all cluster namespaced) (optional default: all)"`
	Namespace     string `json:"namespace,omitempty" jsonschema:"Kubernetes namespace for namespaced resources (optional)"`
	LabelSelector string `json:"label_selector,omitempty" jsonschema:"Label selector for filtering (optional)"`
	ShowLabels    bool   `json:"show_labels,omitempty" jsonschema:"Show resource labels (optional)"`
}

// GetVirtctlClusterResourcesTool returns the tool definition
func GetVirtctlClusterResourcesTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "VirtctlClusterResources",
		Description: `Discover available cluster resources for VM configuration.

    This essential tool helps discover what resources are available in the
    cluster for VM creation. Understanding available resources is crucial
    for creating properly configured VMs.

    RESOURCE DISCOVERY TIPS:

    Instance Types (CPU/Memory configs):
    - Common types: u1.nano (1CPU,1Gi), u1.micro (1CPU,2Gi), u1.small (1CPU,4Gi),
      u1.medium (2CPU,8Gi), u1.large (4CPU,16Gi), u1.xlarge (8CPU,32Gi)
    - Check both cluster-scoped and namespaced variants
    - Look for annotations like "instancetype.kubevirt.io/common-instancetypes-version"
    - Use kubectl describe to see detailed CPU/memory specifications

    Preferences (VM features/settings):
    - Common OS preferences: fedora, ubuntu, centos, rhel8, rhel9, windows10, windows11
    - Preferences define machine type (q35, pc), CPU features, firmware settings
    - Look for "instancetype.kubevirt.io/default-preference" annotations on DataSources
    - Check preference specs for CPU topology, machine features, firmware types

    DataSources (Boot images):
    - Look for labels: os=linux|windows, version=X.X, arch=amd64|arm64
    - Check annotations for recommended instance types and preferences
    - Common sources: quay.io/containerdisks, registry.redhat.io, public cloud images
    - Use kubectl describe to see source URLs and storage requirements

    Storage Classes:
    - Prioritize classes with "storageclass.kubevirt.io/is-default-virt-class=true"
    - Look for "virtualization" in storage class names (e.g., "ocs-virtualization-rbd")
    - Check provisioner types: ceph-rbd, csi-driver, local-storage, etc.
    - Consider performance characteristics: fast-ssd, standard, slow-hdd

    Args:
        resource_type: Type of resource (instancetypes, preferences, datasources, storageclasses, all)
        scope: Resource scope (all, cluster, namespaced) (optional)
        namespace: Kubernetes namespace for namespaced resources (optional)
        label_selector: Label selector for filtering (e.g., "os=linux,version=22.04") (optional)
        show_labels: Show resource labels (optional)

    Returns:
        Available cluster resources information with discovery hints

    Examples:
        # Discover all available instance types
        virtctl_cluster_resources(resource_type="instancetypes", scope="all")

        # Find Linux DataSources only
        virtctl_cluster_resources(resource_type="datasources", scope="cluster", label_selector="os=linux", show_labels=true)

        # Get detailed preference information
        virtctl_cluster_resources(resource_type="preferences", scope="all")

        # Find virtualization-optimized storage classes
        virtctl_cluster_resources(resource_type="storageclasses", show_labels=true)`,
	}
}

func HandleVirtctlClusterResources(ctx context.Context, req *mcp.CallToolRequest, input VirtctlClusterResourcesInput) (*mcp.CallToolResult, any, error) {
	// Validate resource type
	validTypes := map[string]bool{
		"instancetypes": true, "preferences": true, "datasources": true, "storageclasses": true, "all": true,
	}

	if !validTypes[input.ResourceType] {
		return nil, "", fmt.Errorf("invalid resource_type: %s. Valid types: instancetypes, preferences, datasources, storageclasses, all", input.ResourceType)
	}

	// For "all", get all resource types
	if input.ResourceType == "all" {
		results := make(map[string]interface{})
		for resType := range validTypes {
			if resType == "all" {
				continue
			}
			output, err := getResourceType(resType, input.Scope, input.Namespace, input.LabelSelector, input.ShowLabels)
			if err == nil {
				results[resType] = json.RawMessage(output)
			}
		}
		formatted, _ := json.MarshalIndent(results, "", "  ")
		return nil, string(formatted), nil
	}

	// Get single resource type
	output, err := getResourceType(input.ResourceType, input.Scope, input.Namespace, input.LabelSelector, input.ShowLabels)
	if err != nil {
		return nil, "", err
	}

	return nil, output, nil
}
