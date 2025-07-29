package mapping

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
)

// PatchNetwork patches a network mapping
func PatchNetwork(configFlags *genericclioptions.ConfigFlags, name, namespace, addPairs, updatePairs, removePairs, inventoryURL string) error {
	return patchNetworkMapping(configFlags, name, namespace, addPairs, updatePairs, removePairs, inventoryURL)
}

// PatchStorage patches a storage mapping
func PatchStorage(configFlags *genericclioptions.ConfigFlags, name, namespace, addPairs, updatePairs, removePairs, inventoryURL string) error {
	return patchStorageMapping(configFlags, name, namespace, addPairs, updatePairs, removePairs, inventoryURL)
}

// checkNetworkSourceDuplicates checks if any of the new pairs have sources that already exist in current pairs
func checkNetworkSourceDuplicates(currentPairs []forkliftv1beta1.NetworkPair, newPairs []forkliftv1beta1.NetworkPair) []string {
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

// filterOutDuplicateNetworkPairs removes pairs that have duplicate sources, keeping only unique ones
func filterOutDuplicateNetworkPairs(currentPairs []forkliftv1beta1.NetworkPair, newPairs []forkliftv1beta1.NetworkPair) []forkliftv1beta1.NetworkPair {
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
	var filteredPairs []forkliftv1beta1.NetworkPair
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

// removeSourceFromList removes pairs with matching source names/IDs from a list
func removeSourceFromNetworkPairs(pairs []forkliftv1beta1.NetworkPair, sourcesToRemove []string) []forkliftv1beta1.NetworkPair {
	var filteredPairs []forkliftv1beta1.NetworkPair

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

// updateNetworkPairsBySource updates or adds pairs based on source name/ID matching
func updateNetworkPairsBySource(existingPairs []forkliftv1beta1.NetworkPair, newPairs []forkliftv1beta1.NetworkPair) []forkliftv1beta1.NetworkPair {
	updatedPairs := make([]forkliftv1beta1.NetworkPair, len(existingPairs))
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

// getSourceProviderFromMapping extracts the source provider name from a mapping
func getSourceProviderFromMapping(mapping *unstructured.Unstructured) (string, error) {
	provider, found, err := unstructured.NestedMap(mapping.Object, "spec", "provider", "source")
	if err != nil {
		return "", fmt.Errorf("failed to get source provider: %v", err)
	}
	if !found || provider == nil {
		return "", fmt.Errorf("source provider not found in mapping")
	}

	if name, ok := provider["name"].(string); ok {
		return name, nil
	}
	return "", fmt.Errorf("source provider name not found")
}

// parseSourcesToRemove parses a comma-separated list of source names to remove
func parseSourcesToRemove(removeStr string) []string {
	if removeStr == "" {
		return nil
	}

	var sources []string
	sourceList := strings.Split(removeStr, ",")

	for _, source := range sourceList {
		source = strings.TrimSpace(source)
		if source != "" {
			sources = append(sources, source)
		}
	}

	return sources
}
