package azure

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"
	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/storage/mapper"
)

// AzureStorageMapper implements storage mapping for Azure providers
type AzureStorageMapper struct{}

// NewAzureStorageMapper creates a new Azure storage mapper
func NewAzureStorageMapper() mapper.StorageMapper {
	return &AzureStorageMapper{}
}

// getDiskTypeConfig returns the volume mode and access mode for an Azure disk type SKU
func getDiskTypeConfig(diskType string) (corev1.PersistentVolumeMode, corev1.PersistentVolumeAccessMode) {
	// All Azure managed disk types map to block mode for best performance
	switch strings.ToLower(diskType) {
	case "premium_lrs", "premium_zrs", "premiumv2_lrs", "ultrassd_lrs":
		return corev1.PersistentVolumeBlock, corev1.ReadWriteOnce
	case "standard_lrs", "standardssd_lrs", "standardssd_zrs":
		return corev1.PersistentVolumeBlock, corev1.ReadWriteOnce
	default:
		return corev1.PersistentVolumeBlock, corev1.ReadWriteOnce
	}
}

// findMatchingAzureStorageClass tries to find a target SC that matches the given
// Azure disk type by name. Returns "" when no match is found.
func findMatchingAzureStorageClass(diskType string, targetStorages []forkliftv1beta1.DestinationStorage) string {
	if len(targetStorages) == 0 || diskType == "" {
		return ""
	}

	diskTypeLower := strings.ToLower(diskType)

	for _, storage := range targetStorages {
		scName := strings.ToLower(storage.StorageClass)
		if scName == "" {
			continue
		}

		if scName == diskTypeLower {
			klog.V(4).Infof("DEBUG: Azure storage mapper - Found exact match: %s -> %s", diskType, storage.StorageClass)
			return storage.StorageClass
		}

		if strings.Contains(scName, diskTypeLower) {
			klog.V(4).Infof("DEBUG: Azure storage mapper - Found contains match: %s -> %s", diskType, storage.StorageClass)
			return storage.StorageClass
		}

		// Match without underscore (premium-lrs -> premium_lrs)
		diskTypeNormalized := strings.ReplaceAll(diskTypeLower, "_", "-")
		if strings.Contains(scName, diskTypeNormalized) {
			klog.V(4).Infof("DEBUG: Azure storage mapper - Found normalized match: %s -> %s", diskType, storage.StorageClass)
			return storage.StorageClass
		}
	}

	klog.V(4).Infof("DEBUG: Azure storage mapper - No name match for %s", diskType)
	return ""
}

// CreateStoragePairs creates storage mapping pairs for Azure -> OpenShift migrations.
func (m *AzureStorageMapper) CreateStoragePairs(sourceStorages []ref.Ref, targetStorages []forkliftv1beta1.DestinationStorage, opts mapper.StorageMappingOptions) ([]forkliftv1beta1.StoragePair, error) {
	var storagePairs []forkliftv1beta1.StoragePair

	if opts.TargetProviderType != "" && opts.TargetProviderType != "openshift" {
		klog.V(2).Infof("WARNING: Azure storage mapper - Target provider type is '%s', not 'openshift'. Azure->%s migrations may not work as expected.",
			opts.TargetProviderType, opts.TargetProviderType)
	}

	klog.V(4).Infof("DEBUG: Azure storage mapper - Creating storage pairs for %d source disk types", len(sourceStorages))

	if len(sourceStorages) == 0 {
		klog.V(4).Infof("DEBUG: No source storages to map")
		return storagePairs, nil
	}

	// User specified a default SC -- map every source to it
	if opts.DefaultTargetStorageClass != "" {
		klog.V(4).Infof("DEBUG: Azure storage mapper - Using user-defined storage class '%s' for all types", opts.DefaultTargetStorageClass)
		for _, sourceStorage := range sourceStorages {
			volumeMode, accessMode := getDiskTypeConfig(sourceStorage.Name)
			storagePairs = append(storagePairs, forkliftv1beta1.StoragePair{
				Source: sourceStorage,
				Destination: forkliftv1beta1.DestinationStorage{
					StorageClass: opts.DefaultTargetStorageClass,
					VolumeMode:   volumeMode,
					AccessMode:   accessMode,
				},
			})
		}
		klog.V(4).Infof("DEBUG: Azure storage mapper - Created %d storage pairs (user default)", len(storagePairs))

		storagePairs = mapper.ApplyOffloadToPairs(storagePairs, opts)

		return storagePairs, nil
	}

	// Auto-match each Azure disk type against target SC list, with fallback to default
	defaultSC := ""
	if len(targetStorages) > 0 {
		defaultSC = targetStorages[0].StorageClass
	}

	for _, sourceStorage := range sourceStorages {
		diskType := sourceStorage.Name
		volumeMode, accessMode := getDiskTypeConfig(diskType)

		ocpStorageClass := findMatchingAzureStorageClass(diskType, targetStorages)
		if ocpStorageClass != "" {
			klog.V(4).Infof("DEBUG: Azure storage mapper - Matched %s -> %s", diskType, ocpStorageClass)
		} else if defaultSC != "" {
			ocpStorageClass = defaultSC
			klog.V(2).Infof("WARNING: Azure storage mapper - No match for %s, using default SC '%s'", diskType, defaultSC)
		} else {
			klog.V(2).Infof("WARNING: Azure storage mapper - No target storage class available for %s, skipping", diskType)
			continue
		}

		storagePairs = append(storagePairs, forkliftv1beta1.StoragePair{
			Source: sourceStorage,
			Destination: forkliftv1beta1.DestinationStorage{
				StorageClass: ocpStorageClass,
				VolumeMode:   volumeMode,
				AccessMode:   accessMode,
			},
		})
		klog.V(4).Infof("DEBUG: Azure storage mapper - Mapped %s -> %s (mode: %s, access: %s)",
			diskType, ocpStorageClass, volumeMode, accessMode)
	}

	klog.V(4).Infof("DEBUG: Azure storage mapper - Created %d storage pairs", len(storagePairs))

	storagePairs = mapper.ApplyOffloadToPairs(storagePairs, opts)

	return storagePairs, nil
}
