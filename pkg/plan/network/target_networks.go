package network

import (
	"fmt"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/inventory"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NetworkInfo holds the name and namespace of a network.
type NetworkInfo struct {
	Name      string
	Namespace string
}

// GetTargetNetworks retrieves the target networks from the target provider.
func GetTargetNetworks(configFlags *genericclioptions.ConfigFlags, targetProviderName, namespace, inventoryURL string) ([]NetworkInfo, error) {
	// Get target provider
	targetProvider, err := inventory.GetProviderByName(configFlags, targetProviderName, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get target provider: %v", err)
	}

	// Fetch target network inventory (OpenShift NetworkAttachmentDefinitions)
	targetNetworkInventory, err := client.FetchProviderInventory(configFlags, inventoryURL, targetProvider, "networkattachmentdefinitions?detail=4")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch target network inventory: %v", err)
	}

	targetNetworksArray, ok := targetNetworkInventory.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format: expected array for target networks inventory")
	}

	// Collect target network names and namespaces
	var targetNetworks []NetworkInfo
	for _, item := range targetNetworksArray {
		network, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		object, ok := network["object"].(map[string]interface{})
		if !ok {
			continue
		}

		metadata, ok := object["metadata"].(map[string]interface{})
		if !ok {
			continue
		}

		name, ok := metadata["name"].(string)
		if !ok {
			continue
		}

		namespace, ok := metadata["namespace"].(string)
		if !ok {
			continue
		}

		targetNetworks = append(targetNetworks, NetworkInfo{Name: name, Namespace: namespace})
	}

	return targetNetworks, nil
}
