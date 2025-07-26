package mapping

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"

	deletemapping "github.com/yaacov/kubectl-mtv/pkg/cmd/delete/mapping"
	"github.com/yaacov/kubectl-mtv/pkg/cmd/get/mapping"
)

// CreateNetwork creates a new network mapping
func CreateNetwork(configFlags *genericclioptions.ConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile, networkPairs, inventoryURL string) error {
	return createNetworkMapping(configFlags, name, namespace, sourceProvider, targetProvider, fromFile, networkPairs, inventoryURL)
}

// CreateStorage creates a new storage mapping
func CreateStorage(configFlags *genericclioptions.ConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile, storagePairs, inventoryURL string) error {
	return createStorageMapping(configFlags, name, namespace, sourceProvider, targetProvider, fromFile, storagePairs, inventoryURL)
}

// List lists network and storage mappings
func List(configFlags *genericclioptions.ConfigFlags, mappingType, namespace, outputFormat string, mappingName string) error {
	return mapping.List(configFlags, mappingType, namespace, outputFormat, mappingName)
}

// Delete deletes a network or storage mapping
func Delete(configFlags *genericclioptions.ConfigFlags, name, namespace, mappingType string) error {
	return deletemapping.Delete(configFlags, name, namespace, mappingType)
}
