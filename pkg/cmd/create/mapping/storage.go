package mapping

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/provider"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
)

// parseStoragePairs parses storage pairs in format "source1:target-storage-class,source2:target-storage-class"
// The format "source:namespace/storage-class" is also supported for consistency with network mappings,
// but the namespace part is ignored since storage classes are cluster-scoped resources
func parseStoragePairs(pairStr, defaultNamespace string, configFlags *genericclioptions.ConfigFlags, sourceProvider, inventoryURL string) ([]forkliftv1beta1.StoragePair, error) {
	if pairStr == "" {
		return nil, nil
	}

	var pairs []forkliftv1beta1.StoragePair
	pairList := strings.Split(pairStr, ",")

	for _, pairStr := range pairList {
		pairStr = strings.TrimSpace(pairStr)
		if pairStr == "" {
			continue
		}

		parts := strings.SplitN(pairStr, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid storage pair format '%s': expected 'source:storage-class' or 'source:namespace/storage-class'", pairStr)
		}

		sourceName := strings.TrimSpace(parts[0])
		targetPart := strings.TrimSpace(parts[1])

		// Resolve source storage name to ID
		sourceStorageRef, err := resolveStorageNameToID(configFlags, sourceProvider, defaultNamespace, inventoryURL, sourceName)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve source storage '%s': %v", sourceName, err)
		}

		// Parse target part which can be namespace/storage-class or just storage-class
		// Note: namespace is ignored since storage classes are cluster-scoped
		var targetStorageClass string
		if strings.Contains(targetPart, "/") {
			targetParts := strings.SplitN(targetPart, "/", 2)
			// Ignore the namespace part for storage classes since they are cluster-scoped
			targetStorageClass = strings.TrimSpace(targetParts[1])
		} else {
			// Use the target part as storage class
			targetStorageClass = targetPart
		}

		if targetStorageClass == "" {
			return nil, fmt.Errorf("invalid target format '%s': storage class must be specified", targetPart)
		}

		pair := forkliftv1beta1.StoragePair{
			Source: sourceStorageRef,
			Destination: forkliftv1beta1.DestinationStorage{
				StorageClass: targetStorageClass,
			},
		}

		pairs = append(pairs, pair)
	}

	return pairs, nil
}

// createStorageMapping creates a new storage mapping
func createStorageMapping(configFlags *genericclioptions.ConfigFlags, name, namespace, sourceProvider, targetProvider, storagePairs, inventoryURL string) error {
	dynamicClient, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Parse storage pairs if provided
	var mappingPairs []forkliftv1beta1.StoragePair
	if storagePairs != "" {
		mappingPairs, err = parseStoragePairs(storagePairs, namespace, configFlags, sourceProvider, inventoryURL)
		if err != nil {
			return fmt.Errorf("failed to parse storage pairs: %v", err)
		}
	}

	// Create a typed StorageMap
	storageMap := &forkliftv1beta1.StorageMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: forkliftv1beta1.StorageMapSpec{
			Provider: provider.Pair{
				Source: corev1.ObjectReference{
					Name:      sourceProvider,
					Namespace: namespace,
				},
				Destination: corev1.ObjectReference{
					Name:      targetProvider,
					Namespace: namespace,
				},
			},
			Map: mappingPairs,
		},
	}

	// Convert to unstructured
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(storageMap)
	if err != nil {
		return fmt.Errorf("failed to convert to unstructured: %v", err)
	}

	mapping := &unstructured.Unstructured{Object: unstructuredObj}
	mapping.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   client.Group,
		Version: client.Version,
		Kind:    "StorageMap",
	})

	_, err = dynamicClient.Resource(client.StorageMapGVR).Namespace(namespace).Create(context.TODO(), mapping, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create storage mapping: %v", err)
	}

	fmt.Printf("Storage mapping '%s' created in namespace '%s'\n", name, namespace)
	return nil
}
