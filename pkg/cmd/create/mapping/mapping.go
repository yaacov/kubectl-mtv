package mapping

import (
	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
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

// ParseNetworkPairs parses network pairs and returns the parsed pairs (exported for patch functionality)
func ParseNetworkPairs(pairStr, defaultNamespace string, configFlags *genericclioptions.ConfigFlags, sourceProvider, inventoryURL string) ([]forkliftv1beta1.NetworkPair, error) {
	return parseNetworkPairs(pairStr, defaultNamespace, configFlags, sourceProvider, inventoryURL)
}

// ParseStoragePairs parses storage pairs and returns the parsed pairs (exported for patch functionality)
func ParseStoragePairs(pairStr, defaultNamespace string, configFlags *genericclioptions.ConfigFlags, sourceProvider, inventoryURL string) ([]forkliftv1beta1.StoragePair, error) {
	return parseStoragePairs(pairStr, defaultNamespace, configFlags, sourceProvider, inventoryURL)
}
