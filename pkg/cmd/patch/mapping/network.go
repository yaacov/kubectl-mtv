package mapping

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/mapping"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
)

// patchNetworkMapping patches an existing network mapping
func patchNetworkMapping(configFlags *genericclioptions.ConfigFlags, name, namespace, addPairs, updatePairs, removePairs, inventoryURL string) error {
	klog.V(2).Infof("Patching network mapping '%s' in namespace '%s'", name, namespace)

	dynamicClient, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Get the existing mapping
	existingMapping, err := dynamicClient.Resource(client.NetworkMapGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get network mapping '%s': %v", name, err)
	}

	// Extract source provider for pair resolution
	sourceProvider, err := getSourceProviderFromMapping(existingMapping)
	if err != nil {
		return fmt.Errorf("failed to get source provider from mapping: %v", err)
	}

	klog.V(2).Infof("Using source provider '%s' for network pair resolution", sourceProvider)

	// Convert existing mapping to typed format
	var existingNetworkMap forkliftv1beta1.NetworkMap
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(existingMapping.Object, &existingNetworkMap)
	if err != nil {
		return fmt.Errorf("failed to convert existing mapping: %v", err)
	}

	// Get current pairs
	currentPairs := existingNetworkMap.Spec.Map
	klog.V(3).Infof("Current mapping has %d network pairs", len(currentPairs))

	// Process removals first
	if removePairs != "" {
		sourcesToRemove := parseSourcesToRemove(removePairs)
		klog.V(2).Infof("Removing %d network pairs from mapping", len(sourcesToRemove))
		currentPairs = removeSourceFromNetworkPairs(currentPairs, sourcesToRemove)
		klog.V(2).Infof("Successfully removed network pairs from mapping '%s'", name)
	}

	// Process additions
	if addPairs != "" {
		klog.V(2).Infof("Adding network pairs to mapping: %s", addPairs)
		newPairs, err := mapping.ParseNetworkPairs(addPairs, namespace, configFlags, sourceProvider, inventoryURL)
		if err != nil {
			return fmt.Errorf("failed to parse add-pairs: %v", err)
		}

		// Check for duplicate sources
		duplicates := checkNetworkSourceDuplicates(currentPairs, newPairs)
		if len(duplicates) > 0 {
			klog.V(1).Infof("Warning: Found duplicate sources in add-pairs, skipping: %v", duplicates)
			fmt.Printf("Warning: Skipping duplicate sources: %s\n", strings.Join(duplicates, ", "))
			newPairs = filterOutDuplicateNetworkPairs(currentPairs, newPairs)
		}

		if len(newPairs) > 0 {
			currentPairs = append(currentPairs, newPairs...)
			klog.V(2).Infof("Added %d network pairs to mapping '%s'", len(newPairs), name)
		} else {
			klog.V(2).Infof("No new network pairs to add after filtering duplicates")
		}
	}

	// Process updates
	if updatePairs != "" {
		klog.V(2).Infof("Updating network pairs in mapping: %s", updatePairs)
		updatePairsList, err := mapping.ParseNetworkPairs(updatePairs, namespace, configFlags, sourceProvider, inventoryURL)
		if err != nil {
			return fmt.Errorf("failed to parse update-pairs: %v", err)
		}
		currentPairs = updateNetworkPairsBySource(currentPairs, updatePairsList)
		klog.V(2).Infof("Updated %d network pairs in mapping '%s'", len(updatePairsList), name)
	}

	// Update the mapping
	existingNetworkMap.Spec.Map = currentPairs
	klog.V(3).Infof("Final mapping has %d network pairs", len(currentPairs))

	// Convert back to unstructured
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&existingNetworkMap)
	if err != nil {
		return fmt.Errorf("failed to convert to unstructured: %v", err)
	}

	patchedMapping := &unstructured.Unstructured{Object: unstructuredObj}
	patchedMapping.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   client.Group,
		Version: client.Version,
		Kind:    "NetworkMap",
	})

	// Update the resource
	_, err = dynamicClient.Resource(client.NetworkMapGVR).Namespace(namespace).Update(context.TODO(), patchedMapping, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update network mapping: %v", err)
	}

	fmt.Printf("networkmap/%s patched\n", name)
	return nil
}
