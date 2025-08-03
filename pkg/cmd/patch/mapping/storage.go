package mapping

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/mapping"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
)

// patchStorageMapping patches an existing storage mapping
func patchStorageMapping(configFlags *genericclioptions.ConfigFlags, name, namespace, addPairs, updatePairs, removePairs, inventoryURL string) error {
	klog.V(2).Infof("Patching storage mapping '%s' in namespace '%s'", name, namespace)

	dynamicClient, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Get the existing mapping
	existingMapping, err := dynamicClient.Resource(client.StorageMapGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get storage mapping '%s': %v", name, err)
	}

	// Extract source provider for pair resolution
	sourceProviderName, sourceProviderNamespace, err := getSourceProviderFromMapping(existingMapping)
	if err != nil {
		return fmt.Errorf("failed to get source provider from mapping: %v", err)
	}

	if sourceProviderNamespace != "" {
		klog.V(2).Infof("Using source provider '%s/%s' for storage pair resolution", sourceProviderNamespace, sourceProviderName)
	} else {
		klog.V(2).Infof("Using source provider '%s' for storage pair resolution", sourceProviderName)
	}

	// Convert existing mapping to typed format
	var existingStorageMap forkliftv1beta1.StorageMap
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(existingMapping.Object, &existingStorageMap)
	if err != nil {
		return fmt.Errorf("failed to convert existing mapping: %v", err)
	}

	// Make a copy of the current pairs to work with
	originalPairs := existingStorageMap.Spec.Map
	workingPairs := make([]forkliftv1beta1.StoragePair, len(originalPairs))
	copy(workingPairs, originalPairs)
	klog.V(3).Infof("Current mapping has %d storage pairs", len(workingPairs))

	// Process removals first
	if removePairs != "" {
		sourcesToRemove := parseSourcesToRemove(removePairs)
		klog.V(2).Infof("Removing %d storage pairs from mapping", len(sourcesToRemove))
		workingPairs = removeSourceFromStoragePairs(workingPairs, sourcesToRemove)
		klog.V(2).Infof("Successfully removed storage pairs from mapping '%s'", name)
	}

	// Process additions
	if addPairs != "" {
		klog.V(2).Infof("Adding storage pairs to mapping: %s", addPairs)
		newPairs, err := mapping.ParseStoragePairs(addPairs, sourceProviderNamespace, configFlags, sourceProviderName, inventoryURL)
		if err != nil {
			return fmt.Errorf("failed to parse add-pairs: %v", err)
		}

		// Check for duplicate sources
		duplicates := checkStorageSourceDuplicates(workingPairs, newPairs)
		if len(duplicates) > 0 {
			klog.V(1).Infof("Warning: Found duplicate sources in add-pairs, skipping: %v", duplicates)
			fmt.Printf("Warning: Skipping duplicate sources: %s\n", strings.Join(duplicates, ", "))
			newPairs = filterOutDuplicateStoragePairs(workingPairs, newPairs)
		}

		if len(newPairs) > 0 {
			workingPairs = append(workingPairs, newPairs...)
			klog.V(2).Infof("Added %d storage pairs to mapping '%s'", len(newPairs), name)
		} else {
			klog.V(2).Infof("No new storage pairs to add after filtering duplicates")
		}
	}

	// Process updates
	if updatePairs != "" {
		klog.V(2).Infof("Updating storage pairs in mapping: %s", updatePairs)
		updatePairsList, err := mapping.ParseStoragePairs(updatePairs, sourceProviderNamespace, configFlags, sourceProviderName, inventoryURL)
		if err != nil {
			return fmt.Errorf("failed to parse update-pairs: %v", err)
		}
		workingPairs = updateStoragePairsBySource(workingPairs, updatePairsList)
		klog.V(2).Infof("Updated %d storage pairs in mapping '%s'", len(updatePairsList), name)
	}

	klog.V(3).Infof("Final working pairs count: %d", len(workingPairs))

	// Patch the spec.map field
	patchData := map[string]interface{}{
		"spec": map[string]interface{}{
			"map": workingPairs,
		},
	}

	patchBytes, err := runtime.Encode(unstructured.UnstructuredJSONScheme, &unstructured.Unstructured{Object: patchData})
	if err != nil {
		return fmt.Errorf("failed to encode patch data: %v", err)
	}

	// Apply the patch
	_, err = dynamicClient.Resource(client.StorageMapGVR).Namespace(namespace).Patch(
		context.TODO(),
		name,
		types.StrategicMergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to patch storage mapping: %v", err)
	}

	fmt.Printf("storagemap/%s patched\n", name)
	return nil
}

// checkStorageSourceDuplicates checks if any of the new pairs have sources that already exist in current pairs
func checkStorageSourceDuplicates(currentPairs []forkliftv1beta1.StoragePair, newPairs []forkliftv1beta1.StoragePair) []string {
	var duplicates []string

	// Create a map of existing sources for quick lookup
	existingSourceMap := make(map[string]bool)
	for _, pair := range currentPairs {
		if pair.Source.Name != "" {
			existingSourceMap[pair.Source.Name] = true
		}
		if pair.Source.ID != "" {
			existingSourceMap[pair.Source.ID] = true
		}
	}

	// Check new pairs against existing sources
	for _, newPair := range newPairs {
		sourceName := newPair.Source.Name
		sourceID := newPair.Source.ID

		if sourceName != "" && existingSourceMap[sourceName] {
			duplicates = append(duplicates, sourceName)
		} else if sourceID != "" && existingSourceMap[sourceID] {
			duplicates = append(duplicates, sourceID)
		}
	}

	return duplicates
}

// filterOutDuplicateStoragePairs removes pairs that have duplicate sources, keeping only unique ones
func filterOutDuplicateStoragePairs(currentPairs []forkliftv1beta1.StoragePair, newPairs []forkliftv1beta1.StoragePair) []forkliftv1beta1.StoragePair {
	// Create a map of existing sources for quick lookup
	existingSourceMap := make(map[string]bool)
	for _, pair := range currentPairs {
		if pair.Source.Name != "" {
			existingSourceMap[pair.Source.Name] = true
		}
		if pair.Source.ID != "" {
			existingSourceMap[pair.Source.ID] = true
		}
	}

	// Filter new pairs to exclude duplicates
	var filteredPairs []forkliftv1beta1.StoragePair
	for _, newPair := range newPairs {
		sourceName := newPair.Source.Name
		sourceID := newPair.Source.ID

		isDuplicate := false
		if sourceName != "" && existingSourceMap[sourceName] {
			isDuplicate = true
		} else if sourceID != "" && existingSourceMap[sourceID] {
			isDuplicate = true
		}

		if !isDuplicate {
			filteredPairs = append(filteredPairs, newPair)
		}
	}

	return filteredPairs
}

// removeSourceFromStoragePairs removes pairs with matching source names/IDs from a list
func removeSourceFromStoragePairs(pairs []forkliftv1beta1.StoragePair, sourcesToRemove []string) []forkliftv1beta1.StoragePair {
	var filteredPairs []forkliftv1beta1.StoragePair

	for _, pair := range pairs {
		shouldRemove := false
		for _, sourceToRemove := range sourcesToRemove {
			if pair.Source.Name == sourceToRemove || pair.Source.ID == sourceToRemove {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			filteredPairs = append(filteredPairs, pair)
		}
	}

	return filteredPairs
}

// updateStoragePairsBySource updates or adds pairs based on source name/ID matching
func updateStoragePairsBySource(existingPairs []forkliftv1beta1.StoragePair, newPairs []forkliftv1beta1.StoragePair) []forkliftv1beta1.StoragePair {
	updatedPairs := make([]forkliftv1beta1.StoragePair, len(existingPairs))
	copy(updatedPairs, existingPairs)

	for _, newPair := range newPairs {
		found := false
		for i, existingPair := range updatedPairs {
			if (existingPair.Source.Name != "" && existingPair.Source.Name == newPair.Source.Name) ||
				(existingPair.Source.ID != "" && existingPair.Source.ID == newPair.Source.ID) {
				// Update existing pair
				updatedPairs[i] = newPair
				found = true
				break
			}
		}
		if !found {
			// Add new pair
			updatedPairs = append(updatedPairs, newPair)
		}
	}

	return updatedPairs
}
