package storage

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
)

// GetDefaultTargetStorage retrieves the default target storage name based on provider inventory.
func GetDefaultTargetStorage(configFlags *genericclioptions.ConfigFlags, targetProvider *unstructured.Unstructured, namespace, inventoryURL string) (string, error) {
	// Use opnshift storage inventory for target provider
	targetStorageSubPath := "storageclasses?detail=4"

	// Fetch target storage inventory
	targetStorageInventory, err := client.FetchProviderInventory(configFlags, inventoryURL, targetProvider, targetStorageSubPath)
	if err != nil {
		return "", fmt.Errorf("failed to fetch target storage inventory: %v", err)
	}

	targetStorageArray, ok := targetStorageInventory.([]interface{})
	if !ok || len(targetStorageArray) == 0 {
		return "", fmt.Errorf("no target storage resources found for provider %s", targetProvider.GetName())
	}

	// Find the default target storage based on annotations
	var defaultTargetStorage map[string]interface{}
	for _, item := range targetStorageArray {
		storage, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		object, ok := storage["object"].(map[string]interface{})
		if !ok {
			continue
		}

		metadata, ok := object["metadata"].(map[string]interface{})
		if !ok {
			continue
		}

		annotations, ok := metadata["annotations"].(map[string]interface{})
		if !ok {
			continue
		}

		// Check for "storageclass.kubevirt.io/is-default-virt-class" annotation
		if isDefaultVirtClass, ok := annotations["storageclass.kubevirt.io/is-default-virt-class"].(string); ok && isDefaultVirtClass == "true" {
			defaultTargetStorage = storage
			break
		}

		// Check for "storageclass.kubernetes.io/is-default-class" annotation
		if isDefaultClass, ok := annotations["storageclass.kubernetes.io/is-default-class"].(string); ok && isDefaultClass == "true" {
			defaultTargetStorage = storage
			break
		}
	}

	// If no default storage is found, use the first storage in the list
	if defaultTargetStorage == nil {
		defaultTargetStorage = targetStorageArray[0].(map[string]interface{})
	}

	defaultTargetStorageName, ok := defaultTargetStorage["name"].(string)
	if !ok {
		return "", fmt.Errorf("failed to get default target storage name")
	}

	return defaultTargetStorageName, nil
}
