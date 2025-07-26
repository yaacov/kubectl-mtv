package storage

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/util/client"
)

// GetSourceDatastoreIDs retrieves the unique datastore IDs from the source VMs' disks.
func GetSourceDatastoreIDs(configFlags *genericclioptions.ConfigFlags, sourceProvider *unstructured.Unstructured, inventoryURL string, planVMNames []string) (map[string]bool, error) {
	// Fetch source VMs inventory
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

	// Collect unique datastore IDs from VMs' disks
	datastoreIDs := make(map[string]bool)
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

		disks, exists := vm["disks"]
		if !exists {
			continue
		}

		disksArray, ok := disks.([]interface{})
		if !ok {
			continue
		}

		for _, d := range disksArray {
			disk, ok := d.(map[string]interface{})
			if !ok {
				continue
			}

			datastore, exists := disk["datastore"].(map[string]interface{})
			if !exists {
				continue
			}

			datastoreID, ok := datastore["id"].(string)
			if !ok {
				continue
			}

			datastoreIDs[datastoreID] = true
		}
	}

	return datastoreIDs, nil
}
