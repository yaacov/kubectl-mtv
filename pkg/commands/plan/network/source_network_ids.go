package network

import (
	"fmt"

	"github.com/yaacov/kubectl-mtv/pkg/commands/inventory"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// GetSourceNetworkIDs retrieves the unique network IDs from the source VMs.
func GetSourceNetworkIDs(configFlags *genericclioptions.ConfigFlags, sourceProviderName, namespace, inventoryURL string, planVMNames []string) (map[string]bool, error) {
	// Get source provider
	sourceProvider, err := inventory.GetProviderByName(configFlags, sourceProviderName, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get source provider: %v", err)
	}

	// Fetch source VMs inventory to get network IDs
	sourceVMsInventory, err := client.FetchProviderInventory(configFlags, inventoryURL, sourceProvider, "vms?detail=4")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch source VMs inventory: %v", err)
	}

	sourceVMsArray, ok := sourceVMsInventory.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format: expected array for source VMs inventory")
	}

	// Create a map of plan VM names for faster lookup
	planVMs := make(map[string]bool)
	for _, vmName := range planVMNames {
		planVMs[vmName] = true
	}

	// Collect unique network IDs from VMs
	sourceNetworkIDs := make(map[string]bool)
	for _, item := range sourceVMsArray {
		vm, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if this VM is in the plan
		vmName, ok := vm["name"].(string)
		if !ok || !planVMs[vmName] {
			continue
		}

		networks, exists := vm["networks"]
		if !exists {
			continue
		}

		networksArray, ok := networks.([]interface{})
		if !ok {
			continue
		}

		for _, n := range networksArray {
			network, ok := n.(map[string]interface{})
			if !ok {
				continue
			}

			networkID, ok := network["id"].(string)
			if !ok {
				continue
			}

			sourceNetworkIDs[networkID] = true
		}
	}

	return sourceNetworkIDs, nil
}
