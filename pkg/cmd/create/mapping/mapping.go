package mapping

import (
	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// CreateNetwork creates a new network mapping
func CreateNetwork(configFlags *genericclioptions.ConfigFlags, name, namespace, sourceProvider, targetProvider, networkPairs, inventoryURL string) error {
	return createNetworkMapping(configFlags, name, namespace, sourceProvider, targetProvider, networkPairs, inventoryURL)
}

// CreateStorageWithOptions creates a new storage mapping with additional options for VolumeMode, AccessMode, and OffloadPlugin
func CreateStorageWithOptions(configFlags *genericclioptions.ConfigFlags, name, namespace, sourceProvider, targetProvider, storagePairs, inventoryURL string, defaultVolumeMode, defaultAccessMode, defaultOffloadPlugin, defaultOffloadSecret, defaultOffloadVendor string) error {
	return createStorageMappingWithOptions(configFlags, name, namespace, sourceProvider, targetProvider, storagePairs, inventoryURL, defaultVolumeMode, defaultAccessMode, defaultOffloadPlugin, defaultOffloadSecret, defaultOffloadVendor)
}

// ParseNetworkPairs parses network pairs and returns the parsed pairs (exported for patch functionality)
func ParseNetworkPairs(pairStr, defaultNamespace string, configFlags *genericclioptions.ConfigFlags, sourceProvider, inventoryURL string) ([]forkliftv1beta1.NetworkPair, error) {
	return parseNetworkPairs(pairStr, defaultNamespace, configFlags, sourceProvider, inventoryURL)
}

// ParseStoragePairsWithOptions parses storage pairs with additional options for VolumeMode, AccessMode, and OffloadPlugin (exported for patch functionality)
func ParseStoragePairsWithOptions(pairStr, defaultNamespace string, configFlags *genericclioptions.ConfigFlags, sourceProvider, inventoryURL string, defaultVolumeMode, defaultAccessMode, defaultOffloadPlugin, defaultOffloadSecret, defaultOffloadVendor string) ([]forkliftv1beta1.StoragePair, error) {
	return parseStoragePairsWithOptions(pairStr, defaultNamespace, configFlags, sourceProvider, inventoryURL, defaultVolumeMode, defaultAccessMode, defaultOffloadPlugin, defaultOffloadSecret, defaultOffloadVendor)
}
