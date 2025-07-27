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

	// Determine the target storage to use for all source storages
	var targetStorage forkliftv1beta1.DestinationStorage

	// If a default target storage class is specified, use it for all source storages
	if opts.DefaultTargetStorageClass != "" {
		targetStorage = forkliftv1beta1.DestinationStorage{
			StorageClass: opts.DefaultTargetStorageClass,
		}
		klog.V(4).Infof("DEBUG: Using explicit default storage class for all sources: %s", opts.DefaultTargetStorageClass)
	} else if len(targetStorages) > 0 {
		// Use the first target storage (which should be the best one from FetchTargetStorages)
		targetStorage = targetStorages[0]
		klog.V(4).Infof("DEBUG: Using first target storage for all sources: %s", targetStorage.StorageClass)
	} else {
		// Fall back to empty storage class (system default)
		targetStorage = forkliftv1beta1.DestinationStorage{}
		klog.V(4).Infof("DEBUG: No target storages available, using system default")
	}

	// Map all source storages to the selected target storage
	for _, sourceStorage := range sourceStorages {
		storagePairs = append(storagePairs, forkliftv1beta1.StoragePair{
			Source:      sourceStorage,
			Destination: targetStorage,
		})
		klog.V(4).Infof("DEBUG: Mapped source storage %s -> %s", sourceStorage.Name, targetStorage.StorageClass)
	}

	klog.V(4).Infof("DEBUG: Created %d storage pairs", len(storagePairs))
	return storagePairs, nil
}
