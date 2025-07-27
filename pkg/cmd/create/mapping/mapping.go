package mapping

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// CreateNetwork creates a new network mapping
func CreateNetwork(configFlags *genericclioptions.ConfigFlags, name, namespace, sourceProvider, targetProvider, networkPairs, inventoryURL string) error {
	return createNetworkMapping(configFlags, name, namespace, sourceProvider, targetProvider, networkPairs, inventoryURL)
}

// CreateStorage creates a new storage mapping
func CreateStorage(configFlags *genericclioptions.ConfigFlags, name, namespace, sourceProvider, targetProvider, storagePairs, inventoryURL string) error {
	return createStorageMapping(configFlags, name, namespace, sourceProvider, targetProvider, storagePairs, inventoryURL)
}
