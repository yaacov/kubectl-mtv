package storage

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/provider"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/inventory"
)

// CreateDefaultStorageMapOptions encapsulates the parameters for creating a default storage map.
type CreateDefaultStorageMapOptions struct {
	Name                      string
	Namespace                 string
	SourceProvider            string
	TargetProvider            string
	ConfigFlags               *genericclioptions.ConfigFlags
	InventoryURL              string
	PlanVMNames               []string
	DefaultTargetStorageClass string
}

// CreateDefaultStorageMap creates a default storage map.
// Returns the name of the created map and any error that occurred.
func CreateDefaultStorageMap(opts CreateDefaultStorageMapOptions) (string, error) {
	c, err := client.GetDynamicClient(opts.ConfigFlags)
	if err != nil {
		return "", fmt.Errorf("failed to get client: %v", err)
	}

	var defaultTargetStorageName string

	// If default target storage class is specified, use it
	if opts.DefaultTargetStorageClass != "" {
		defaultTargetStorageName = opts.DefaultTargetStorageClass
	} else {
		// Get target provider and find default storage
		targetProvider, err := inventory.GetProviderByName(opts.ConfigFlags, opts.TargetProvider, opts.Namespace)
		if err != nil {
			return "", fmt.Errorf("failed to get target provider: %v", err)
		}

		defaultTargetStorageName, err = GetDefaultTargetStorage(opts.ConfigFlags, targetProvider, opts.Namespace, opts.InventoryURL)
		if err != nil {
			return "", err
		}
	}

	// Get source provider
	sourceProvider, err := inventory.GetProviderByName(opts.ConfigFlags, opts.SourceProvider, opts.Namespace)
	if err != nil {
		return "", fmt.Errorf("failed to get source provider: %v", err)
	}

	datastoreIDs, err := GetSourceDatastoreIDs(opts.ConfigFlags, sourceProvider, opts.InventoryURL, opts.PlanVMNames)
	if err != nil {
		return "", err
	}

	// Create StorageMap entries
	var storagePairs []forkliftv1beta1.StoragePair
	for datastoreID := range datastoreIDs {
		storagePairs = append(storagePairs, forkliftv1beta1.StoragePair{
			Source: ref.Ref{
				ID: datastoreID,
			},
			Destination: forkliftv1beta1.DestinationStorage{
				StorageClass: defaultTargetStorageName,
			},
		})
	}

	// If no storage pairs were created, create a dummy entry
	if len(storagePairs) == 0 {
		storagePairs = []forkliftv1beta1.StoragePair{
			{
				Source: ref.Ref{
					Name: defaultTargetStorageName,
				},
				Destination: forkliftv1beta1.DestinationStorage{
					StorageClass: defaultTargetStorageName,
				},
			},
		}
	}

	// Create a new StorageMap object
	storageMap := &forkliftv1beta1.StorageMap{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: opts.Name + "-",
			Namespace:    opts.Namespace,
		},
		Spec: forkliftv1beta1.StorageMapSpec{
			Provider: provider.Pair{
				Source: corev1.ObjectReference{
					Kind:       "Provider",
					APIVersion: forkliftv1beta1.SchemeGroupVersion.String(),
					Name:       opts.SourceProvider,
					Namespace:  opts.Namespace,
				},
				Destination: corev1.ObjectReference{
					Kind:       "Provider",
					APIVersion: forkliftv1beta1.SchemeGroupVersion.String(),
					Name:       opts.TargetProvider,
					Namespace:  opts.Namespace,
				},
			},
			Map: storagePairs,
		},
	}
	storageMap.Kind = "StorageMap"
	storageMap.APIVersion = forkliftv1beta1.SchemeGroupVersion.String()

	// Convert StorageMap object to Unstructured
	unstructuredStorageMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(storageMap)
	if err != nil {
		return "", fmt.Errorf("failed to convert StorageMap to Unstructured: %v", err)
	}
	storageMapUnstructured := &unstructured.Unstructured{Object: unstructuredStorageMap}

	// Create the StorageMap in the specified namespace
	createdMap, err := c.Resource(client.StorageMapGVR).Namespace(opts.Namespace).Create(context.TODO(), storageMapUnstructured, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create storage map: %v", err)
	}

	fmt.Printf("StorageMap '%s' created in namespace '%s'\n", createdMap.GetName(), opts.Namespace)
	return createdMap.GetName(), nil
}
