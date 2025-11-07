package vsphere

import (
	"testing"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"

	storagemapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/storage/mapper"
)

func TestVSphereStorageMapper_CreateStoragePairs_DefaultBehavior(t *testing.T) {
	tests := []struct {
		name                      string
		sourceStorages            []ref.Ref
		targetStorages            []forkliftv1beta1.DestinationStorage
		defaultTargetStorageClass string
		expectedPairs             int
		expectedAllToSameTarget   bool
		expectedTargetClass       string
	}{
		{
			name: "All sources map to first target storage",
			sourceStorages: []ref.Ref{
				{Name: "datastore1", ID: "ds-1"},
				{Name: "datastore2", ID: "ds-2"},
				{Name: "datastore3", ID: "ds-3"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "ocs-storagecluster-ceph-rbd"},
				{StorageClass: "ocs-storagecluster-ceph-rbd-virtualization"},
			},
			expectedPairs:           3,
			expectedAllToSameTarget: true,
			expectedTargetClass:     "ocs-storagecluster-ceph-rbd",
		},
		{
			name: "User-defined default takes priority",
			sourceStorages: []ref.Ref{
				{Name: "datastore1", ID: "ds-1"},
				{Name: "datastore2", ID: "ds-2"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "auto-selected"},
			},
			defaultTargetStorageClass: "user-defined",
			expectedPairs:             2,
			expectedAllToSameTarget:   true,
			expectedTargetClass:       "user-defined",
		},
		{
			name: "Empty target storages with user default",
			sourceStorages: []ref.Ref{
				{Name: "datastore1", ID: "ds-1"},
			},
			targetStorages:            []forkliftv1beta1.DestinationStorage{},
			defaultTargetStorageClass: "user-defined",
			expectedPairs:             1,
			expectedAllToSameTarget:   true,
			expectedTargetClass:       "user-defined",
		},
		{
			name: "Empty target storages without user default",
			sourceStorages: []ref.Ref{
				{Name: "datastore1", ID: "ds-1"},
			},
			targetStorages:          []forkliftv1beta1.DestinationStorage{},
			expectedPairs:           1,
			expectedAllToSameTarget: true,
			expectedTargetClass:     "", // System default
		},
		{
			name:                    "Empty sources",
			sourceStorages:          []ref.Ref{},
			targetStorages:          []forkliftv1beta1.DestinationStorage{{StorageClass: "any"}},
			expectedPairs:           0,
			expectedAllToSameTarget: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageMapper := NewVSphereStorageMapper()

			opts := storagemapper.StorageMappingOptions{
				DefaultTargetStorageClass: tt.defaultTargetStorageClass,
			}

			pairs, err := storageMapper.CreateStoragePairs(tt.sourceStorages, tt.targetStorages, opts)
			if err != nil {
				t.Fatalf("CreateStoragePairs() error = %v", err)
			}

			if len(pairs) != tt.expectedPairs {
				t.Errorf("CreateStoragePairs() got %d pairs, want %d", len(pairs), tt.expectedPairs)
			}

			// Verify all pairs map to the same target (generic behavior)
			if tt.expectedAllToSameTarget && len(pairs) > 1 {
				firstTarget := pairs[0].Destination.StorageClass
				for i, pair := range pairs[1:] {
					if pair.Destination.StorageClass != firstTarget {
						t.Errorf("Pair %d: target %s != first target %s (expected all same)",
							i+1, pair.Destination.StorageClass, firstTarget)
					}
				}
			}

			// Verify expected target class
			if len(pairs) > 0 && tt.expectedTargetClass != "" {
				if pairs[0].Destination.StorageClass != tt.expectedTargetClass {
					t.Errorf("Target class: got %s, want %s",
						pairs[0].Destination.StorageClass, tt.expectedTargetClass)
				}
			}

			// Verify source names are preserved
			for i, pair := range pairs {
				if i < len(tt.sourceStorages) {
					if pair.Source.Name != tt.sourceStorages[i].Name {
						t.Errorf("Pair %d: source name %s != expected %s",
							i, pair.Source.Name, tt.sourceStorages[i].Name)
					}
				}
			}
		})
	}
}

func TestVSphereStorageMapper_NoSameNameMatching(t *testing.T) {
	// This test verifies that vSphere mapper does NOT use same-name matching
	// even when source and target names match
	storageMapper := NewVSphereStorageMapper()

	sourceStorages := []ref.Ref{
		{Name: "identical-name", ID: "src-1"},
		{Name: "another-name", ID: "src-2"},
	}

	targetStorages := []forkliftv1beta1.DestinationStorage{
		{StorageClass: "identical-name"}, // Same name as source
		{StorageClass: "different-name"},
	}

	opts := storagemapper.StorageMappingOptions{
		SourceProviderType: "vsphere",
		TargetProviderType: "openshift",
	}

	pairs, err := storageMapper.CreateStoragePairs(sourceStorages, targetStorages, opts)
	if err != nil {
		t.Fatalf("CreateStoragePairs() error = %v", err)
	}

	// Both sources should map to the same target (first one) - NOT same-name matching
	if len(pairs) != 2 {
		t.Fatalf("Expected 2 pairs, got %d", len(pairs))
	}

	expectedTarget := "identical-name" // First target
	for i, pair := range pairs {
		if pair.Destination.StorageClass != expectedTarget {
			t.Errorf("Pair %d: got target %s, want %s (all should map to first target)",
				i, pair.Destination.StorageClass, expectedTarget)
		}
	}
}

func TestVSphereStorageMapper_ImplementsInterface(t *testing.T) {
	var _ storagemapper.StorageMapper = &VSphereStorageMapper{}
}
