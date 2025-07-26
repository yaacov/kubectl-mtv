package mapper

import (
	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"
	"k8s.io/klog/v2"
)

// StorageMappingOptions contains options for storage mapping
type StorageMappingOptions struct {
	DefaultTargetStorageClass string
}

// CreateStoragePairs creates storage mapping pairs using generic logic
func CreateStoragePairs(sourceStorages []ref.Ref, targetStorages []forkliftv1beta1.DestinationStorage, opts StorageMappingOptions) ([]forkliftv1beta1.StoragePair, error) {
	var storagePairs []forkliftv1beta1.StoragePair

	klog.V(4).Infof("DEBUG: Creating storage pairs - %d source storages, %d target storages", len(sourceStorages), len(targetStorages))
	klog.V(4).Infof("DEBUG: Default target storage class: '%s'", opts.DefaultTargetStorageClass)

	// If a default target storage class is specified, use it for all source storages
	if opts.DefaultTargetStorageClass != "" {
		defaultDestination := forkliftv1beta1.DestinationStorage{
			StorageClass: opts.DefaultTargetStorageClass,
		}
		klog.V(4).Infof("DEBUG: Using explicit default storage class for all sources: %s", opts.DefaultTargetStorageClass)

		for _, sourceStorage := range sourceStorages {
			storagePairs = append(storagePairs, forkliftv1beta1.StoragePair{
				Source:      sourceStorage,
				Destination: defaultDestination,
			})
		}
		return storagePairs, nil
	}

	// Auto-mapping: try to match source storages with target storages intelligently
	klog.V(4).Infof("DEBUG: Available target storages for auto-mapping:")
	for i, targetStorage := range targetStorages {
		klog.V(4).Infof("  [%d] %s", i, targetStorage.StorageClass)
	}

	// Create name-based lookup for target storages
	targetStorageMap := make(map[string]forkliftv1beta1.DestinationStorage)
	for _, targetStorage := range targetStorages {
		// Use storage class name as key for matching
		targetStorageMap[targetStorage.StorageClass] = targetStorage
	}

	// Create pairs for each source storage
	for _, sourceStorage := range sourceStorages {
		destination := findBestTargetStorage(sourceStorage, targetStorages, targetStorageMap)

		storagePairs = append(storagePairs, forkliftv1beta1.StoragePair{
			Source:      sourceStorage,
			Destination: destination,
		})
	}

	return storagePairs, nil
}

// findBestTargetStorage finds the best target storage for a given source storage
func findBestTargetStorage(sourceStorage ref.Ref, targetStorages []forkliftv1beta1.DestinationStorage, targetStorageMap map[string]forkliftv1beta1.DestinationStorage) forkliftv1beta1.DestinationStorage {
	// Strategy 1: Try exact name match
	if targetStorage, exists := targetStorageMap[sourceStorage.Name]; exists {
		klog.V(4).Infof("DEBUG: Found exact name match for %s -> %s",
			sourceStorage.Name, targetStorage.StorageClass)
		return targetStorage
	}

	// Strategy 2: Use first available storage class
	if len(targetStorages) > 0 {
		target := targetStorages[0]
		klog.V(4).Infof("DEBUG: No exact match for %s, using first available storage: %s",
			sourceStorage.Name, target.StorageClass)
		return target
	}

	// Strategy 3: Fall back to empty storage class (system default)
	klog.V(4).Infof("DEBUG: No target storages available for %s, using system default", sourceStorage.Name)
	return forkliftv1beta1.DestinationStorage{} // Empty means system default
}
