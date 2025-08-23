package mapping

import (
	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// StorageCreateOptions holds options for creating storage mappings
type StorageCreateOptions struct {
	ConfigFlags          *genericclioptions.ConfigFlags
	Name                 string
	Namespace            string
	SourceProvider       string
	TargetProvider       string
	StoragePairs         string
	InventoryURL         string
	DefaultVolumeMode    string
	DefaultAccessMode    string
	DefaultOffloadPlugin string
	DefaultOffloadSecret string
	DefaultOffloadVendor string
}

// StorageParseOptions holds options for parsing storage pairs
type StorageParseOptions struct {
	PairStr              string
	DefaultNamespace     string
	ConfigFlags          *genericclioptions.ConfigFlags
	SourceProvider       string
	InventoryURL         string
	DefaultVolumeMode    string
	DefaultAccessMode    string
	DefaultOffloadPlugin string
	DefaultOffloadSecret string
	DefaultOffloadVendor string
}

// CreateNetwork creates a new network mapping
func CreateNetwork(configFlags *genericclioptions.ConfigFlags, name, namespace, sourceProvider, targetProvider, networkPairs, inventoryURL string) error {
	return createNetworkMapping(configFlags, name, namespace, sourceProvider, targetProvider, networkPairs, inventoryURL)
}

// CreateStorageWithOptions creates a new storage mapping with additional options for VolumeMode, AccessMode, and OffloadPlugin
func CreateStorageWithOptions(opts StorageCreateOptions) error {
	return createStorageMappingWithOptions(opts.ConfigFlags, opts.Name, opts.Namespace, opts.SourceProvider, opts.TargetProvider, opts.StoragePairs, opts.InventoryURL, opts.DefaultVolumeMode, opts.DefaultAccessMode, opts.DefaultOffloadPlugin, opts.DefaultOffloadSecret, opts.DefaultOffloadVendor)
}

// ParseNetworkPairs parses network pairs and returns the parsed pairs (exported for patch functionality)
func ParseNetworkPairs(pairStr, defaultNamespace string, configFlags *genericclioptions.ConfigFlags, sourceProvider, inventoryURL string) ([]forkliftv1beta1.NetworkPair, error) {
	return parseNetworkPairs(pairStr, defaultNamespace, configFlags, sourceProvider, inventoryURL)
}

// ParseStoragePairsWithOptions parses storage pairs with additional options for VolumeMode, AccessMode, and OffloadPlugin (exported for patch functionality)
func ParseStoragePairsWithOptions(opts StorageParseOptions) ([]forkliftv1beta1.StoragePair, error) {
	return parseStoragePairsWithOptions(opts.PairStr, opts.DefaultNamespace, opts.ConfigFlags, opts.SourceProvider, opts.InventoryURL, opts.DefaultVolumeMode, opts.DefaultAccessMode, opts.DefaultOffloadPlugin, opts.DefaultOffloadSecret, opts.DefaultOffloadVendor)
}
