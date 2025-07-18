package mapping

import (
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// CreateNetwork creates a new network mapping
func CreateNetwork(configFlags *genericclioptions.ConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile string) error {
	return createNetworkMapping(configFlags, name, namespace, sourceProvider, targetProvider, fromFile)
}

// CreateStorage creates a new storage mapping
func CreateStorage(configFlags *genericclioptions.ConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile string) error {
	return createStorageMapping(configFlags, name, namespace, sourceProvider, targetProvider, fromFile)
}

// List lists network and storage mappings
func List(configFlags *genericclioptions.ConfigFlags, mappingType, namespace, outputFormat string) error {
	return listMappings(configFlags, mappingType, namespace, outputFormat)
}

// Delete deletes a network or storage mapping
func Delete(configFlags *genericclioptions.ConfigFlags, name, namespace, mappingType string) error {
	switch mappingType {
	case "network":
		return DeleteNetwork(configFlags, name, namespace)
	case "storage":
		return DeleteStorage(configFlags, name, namespace)
	default:
		return fmt.Errorf("unsupported mapping type: %s. Use 'network' or 'storage'", mappingType)
	}
}
