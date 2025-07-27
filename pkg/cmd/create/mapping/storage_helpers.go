package mapping

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"
	"github.com/yaacov/kubectl-mtv/pkg/cmd/get/inventory"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
)

// resolveVSphereStorageNameToID resolves storage name for VMware vSphere provider
func resolveVSphereStorageNameToID(configFlags *genericclioptions.ConfigFlags, inventoryURL string, provider *unstructured.Unstructured, storageName string) (ref.Ref, error) {
	// Fetch datastores from VMware vSphere
	storageInventory, err := client.FetchProviderInventory(configFlags, inventoryURL, provider, "datastores?detail=4")
	if err != nil {
		return ref.Ref{}, fmt.Errorf("failed to fetch storage inventory: %v", err)
	}

	storageArray, ok := storageInventory.([]interface{})
	if !ok {
		return ref.Ref{}, fmt.Errorf("unexpected data format: expected array for storage inventory")
	}

	// Search for the storage by name
	for _, item := range storageArray {
		storage, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := storage["name"].(string)
		id, _ := storage["id"].(string)

		if name == storageName {
			return ref.Ref{
				Name: name,
				ID:   id,
			}, nil
		}
	}

	return ref.Ref{}, fmt.Errorf("datastore '%s' not found in vSphere provider inventory", storageName)
}

// resolveOvirtStorageNameToID resolves storage name for oVirt provider
func resolveOvirtStorageNameToID(configFlags *genericclioptions.ConfigFlags, inventoryURL string, provider *unstructured.Unstructured, storageName string) (ref.Ref, error) {
	// Fetch storage domains from oVirt
	storageInventory, err := client.FetchProviderInventory(configFlags, inventoryURL, provider, "storagedomains?detail=4")
	if err != nil {
		return ref.Ref{}, fmt.Errorf("failed to fetch storage inventory: %v", err)
	}

	storageArray, ok := storageInventory.([]interface{})
	if !ok {
		return ref.Ref{}, fmt.Errorf("unexpected data format: expected array for storage inventory")
	}

	// Search for the storage by name
	for _, item := range storageArray {
		storage, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := storage["name"].(string)
		id, _ := storage["id"].(string)

		if name == storageName {
			return ref.Ref{
				Name: name,
				ID:   id,
			}, nil
		}
	}

	return ref.Ref{}, fmt.Errorf("storage domain '%s' not found in oVirt provider inventory", storageName)
}

// resolveOpenStackStorageNameToID resolves storage name for OpenStack provider
func resolveOpenStackStorageNameToID(configFlags *genericclioptions.ConfigFlags, inventoryURL string, provider *unstructured.Unstructured, storageName string) (ref.Ref, error) {
	// Fetch storage types from OpenStack
	storageInventory, err := client.FetchProviderInventory(configFlags, inventoryURL, provider, "volumetypes?detail=4")
	if err != nil {
		return ref.Ref{}, fmt.Errorf("failed to fetch storage inventory: %v", err)
	}

	storageArray, ok := storageInventory.([]interface{})
	if !ok {
		return ref.Ref{}, fmt.Errorf("unexpected data format: expected array for storage inventory")
	}

	// Search for the storage by name
	for _, item := range storageArray {
		storage, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := storage["name"].(string)
		id, _ := storage["id"].(string)

		if name == storageName {
			return ref.Ref{
				Name: name,
				ID:   id,
			}, nil
		}
	}

	return ref.Ref{}, fmt.Errorf("storage type '%s' not found in OpenStack provider inventory", storageName)
}

// resolveStorageNameToID resolves a storage name to its ref.Ref by querying the provider inventory
func resolveStorageNameToID(configFlags *genericclioptions.ConfigFlags, providerName, namespace, inventoryURL, storageName string) (ref.Ref, error) {
	// Get source provider
	provider, err := inventory.GetProviderByName(configFlags, providerName, namespace)
	if err != nil {
		return ref.Ref{}, fmt.Errorf("failed to get provider '%s': %v", providerName, err)
	}

	// Check provider type to determine which helper to use
	providerType, _, err := unstructured.NestedString(provider.Object, "spec", "type")
	if err != nil {
		return ref.Ref{}, fmt.Errorf("failed to get provider type: %v", err)
	}

	switch providerType {
	case "openshift":
		// For OpenShift source providers, only include the name in the source reference
		// Storage classes are cluster-scoped resources, so we don't need to resolve the ID
		return ref.Ref{
			Name: storageName,
		}, nil
	case "vsphere":
		return resolveVSphereStorageNameToID(configFlags, inventoryURL, provider, storageName)
	case "ovirt":
		return resolveOvirtStorageNameToID(configFlags, inventoryURL, provider, storageName)
	case "openstack":
		return resolveOpenStackStorageNameToID(configFlags, inventoryURL, provider, storageName)
	case "ova":
		// OVA providers don't have traditional storage inventory
		// The storage name is typically the VMDK file name
		// Just return the storage name as-is without ID resolution
		return ref.Ref{
			Name: storageName,
		}, nil
	default:
		// Default to generic storage endpoint for unknown providers
		storageInventory, err := client.FetchProviderInventory(configFlags, inventoryURL, provider, "storage?detail=4")
		if err != nil {
			return ref.Ref{}, fmt.Errorf("failed to fetch storage inventory: %v", err)
		}

		storageArray, ok := storageInventory.([]interface{})
		if !ok {
			return ref.Ref{}, fmt.Errorf("unexpected data format: expected array for storage inventory")
		}

		// Search for the storage by name
		for _, item := range storageArray {
			storage, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			name, _ := storage["name"].(string)
			id, _ := storage["id"].(string)

			if name == storageName {
				return ref.Ref{
					ID: id,
				}, nil
			}
		}

		return ref.Ref{}, fmt.Errorf("storage '%s' not found in provider '%s' inventory", storageName, providerName)
	}
}
