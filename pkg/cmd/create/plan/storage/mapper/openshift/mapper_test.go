package openshift

import (
	"testing"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"

	storagemapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/storage/mapper"
)

func TestOpenShiftStorageMapper_CreateStoragePairs_SameNameMatching(t *testing.T) {
	tests := []struct {
		name                string
		sourceStorages      []ref.Ref
		targetStorages      []forkliftv1beta1.DestinationStorage
		sourceProviderType  string
		targetProviderType  string
		expectedPairs       int
		expectedSameName    bool
		expectedTargetNames []string
	}{
		{
			name: "OCP-to-OCP: All sources match by name",
			sourceStorages: []ref.Ref{
				{Name: "fast-ssd", ID: "src-1"},
				{Name: "slow-hdd", ID: "src-2"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "fast-ssd"},
				{StorageClass: "slow-hdd"},
				{StorageClass: "nvme-storage"},
			},
			sourceProviderType:  "openshift",
			targetProviderType:  "openshift",
			expectedPairs:       2,
			expectedSameName:    true,
			expectedTargetNames: []string{"fast-ssd", "slow-hdd"},
		},
		{
			name: "OCP-to-OCP: Some sources don't match by name - fallback to default",
			sourceStorages: []ref.Ref{
				{Name: "fast-ssd", ID: "src-1"},
				{Name: "unknown-storage", ID: "src-2"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "fast-ssd"},
				{StorageClass: "slow-hdd"},
			},
			sourceProviderType:  "openshift",
			targetProviderType:  "openshift",
			expectedPairs:       2,
			expectedSameName:    false,
			expectedTargetNames: []string{"fast-ssd", "fast-ssd"}, // Both map to first (default)
		},
		{
			name: "OCP-to-non-OCP: Use default behavior",
			sourceStorages: []ref.Ref{
				{Name: "datastore1", ID: "src-1"},
				{Name: "datastore2", ID: "src-2"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "ocs-storagecluster-ceph-rbd"},
			},
			sourceProviderType:  "openshift",
			targetProviderType:  "vsphere",
			expectedPairs:       2,
			expectedSameName:    false,
			expectedTargetNames: []string{"ocs-storagecluster-ceph-rbd", "ocs-storagecluster-ceph-rbd"},
		},
		{
			name:                "OCP-to-OCP: Empty sources",
			sourceStorages:      []ref.Ref{},
			targetStorages:      []forkliftv1beta1.DestinationStorage{{StorageClass: "default"}},
			sourceProviderType:  "openshift",
			targetProviderType:  "openshift",
			expectedPairs:       0,
			expectedSameName:    false,
			expectedTargetNames: []string{},
		},
		{
			name: "OCP-to-OCP: Single source with match",
			sourceStorages: []ref.Ref{
				{Name: "fast-ssd", ID: "src-1"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "fast-ssd"},
			},
			sourceProviderType:  "openshift",
			targetProviderType:  "openshift",
			expectedPairs:       1,
			expectedSameName:    true,
			expectedTargetNames: []string{"fast-ssd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageMapper := NewOpenShiftStorageMapper()

			opts := storagemapper.StorageMappingOptions{
				SourceProviderType: tt.sourceProviderType,
				TargetProviderType: tt.targetProviderType,
			}

			pairs, err := storageMapper.CreateStoragePairs(tt.sourceStorages, tt.targetStorages, opts)
			if err != nil {
				t.Fatalf("CreateStoragePairs() error = %v", err)
			}

			if len(pairs) != tt.expectedPairs {
				t.Errorf("CreateStoragePairs() got %d pairs, want %d", len(pairs), tt.expectedPairs)
			}

			// Verify target storage class names
			for i, pair := range pairs {
				if i < len(tt.expectedTargetNames) {
					if pair.Destination.StorageClass != tt.expectedTargetNames[i] {
						t.Errorf("Pair %d: got target %s, want %s", i, pair.Destination.StorageClass, tt.expectedTargetNames[i])
					}
				}
			}

			// Verify source names are preserved
			for i, pair := range pairs {
				if i < len(tt.sourceStorages) {
					if pair.Source.Name != tt.sourceStorages[i].Name {
						t.Errorf("Pair %d: source name %s != expected %s", i, pair.Source.Name, tt.sourceStorages[i].Name)
					}
				}
			}
		})
	}
}

func TestOpenShiftStorageMapper_CreateStoragePairs_DefaultStorageClassSelection(t *testing.T) {
	tests := []struct {
		name                      string
		targetStorages            []forkliftv1beta1.DestinationStorage
		defaultTargetStorageClass string
		expectedDefaultClass      string
	}{
		{
			name: "User-defined default takes priority",
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "auto-selected"},
				{StorageClass: "other-class"},
			},
			defaultTargetStorageClass: "user-defined",
			expectedDefaultClass:      "user-defined",
		},
		{
			name: "Auto-selected when no user default",
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "first-available"},
				{StorageClass: "second-class"},
			},
			defaultTargetStorageClass: "",
			expectedDefaultClass:      "first-available",
		},
		{
			name:                      "Empty when no targets and no user default",
			targetStorages:            []forkliftv1beta1.DestinationStorage{},
			defaultTargetStorageClass: "",
			expectedDefaultClass:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageMapper := NewOpenShiftStorageMapper()

			sourceStorages := []ref.Ref{{Name: "test-source", ID: "src-1"}}
			opts := storagemapper.StorageMappingOptions{
				DefaultTargetStorageClass: tt.defaultTargetStorageClass,
				SourceProviderType:        "openshift",
				TargetProviderType:        "vsphere", // Force default behavior
			}

			pairs, err := storageMapper.CreateStoragePairs(sourceStorages, tt.targetStorages, opts)
			if err != nil {
				t.Fatalf("CreateStoragePairs() error = %v", err)
			}

			if len(pairs) > 0 {
				got := pairs[0].Destination.StorageClass
				if got != tt.expectedDefaultClass {
					t.Errorf("Default storage class: got %s, want %s", got, tt.expectedDefaultClass)
				}
			}
		})
	}
}

func TestCanMatchAllStoragesByName(t *testing.T) {
	tests := []struct {
		name           string
		sourceStorages []ref.Ref
		targetStorages []forkliftv1beta1.DestinationStorage
		expected       bool
	}{
		{
			name: "All sources match",
			sourceStorages: []ref.Ref{
				{Name: "fast-ssd"},
				{Name: "slow-hdd"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "fast-ssd"},
				{StorageClass: "slow-hdd"},
				{StorageClass: "nvme-storage"},
			},
			expected: true,
		},
		{
			name: "Some sources don't match",
			sourceStorages: []ref.Ref{
				{Name: "fast-ssd"},
				{Name: "unknown-storage"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "fast-ssd"},
				{StorageClass: "slow-hdd"},
			},
			expected: false,
		},
		{
			name:           "Empty sources",
			sourceStorages: []ref.Ref{},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "any-storage"},
			},
			expected: true,
		},
		{
			name: "Empty targets",
			sourceStorages: []ref.Ref{
				{Name: "some-storage"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{},
			expected:       false,
		},
		{
			name: "Target with empty storage class name",
			sourceStorages: []ref.Ref{
				{Name: "fast-ssd"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: ""},
				{StorageClass: "fast-ssd"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canMatchAllStoragesByName(tt.sourceStorages, tt.targetStorages)
			if result != tt.expected {
				t.Errorf("canMatchAllStoragesByName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCreateSameNameStoragePairs(t *testing.T) {
	sourceStorages := []ref.Ref{
		{Name: "fast-ssd", ID: "src-1"},
		{Name: "slow-hdd", ID: "src-2"},
	}

	targetStorages := []forkliftv1beta1.DestinationStorage{
		{StorageClass: "fast-ssd"},
		{StorageClass: "slow-hdd"},
		{StorageClass: "extra-storage"},
	}

	pairs, err := createSameNameStoragePairs(sourceStorages, targetStorages)
	if err != nil {
		t.Fatalf("createSameNameStoragePairs() error = %v", err)
	}

	if len(pairs) != 2 {
		t.Errorf("Expected 2 pairs, got %d", len(pairs))
	}

	// Verify mappings
	expectedMappings := map[string]string{
		"fast-ssd": "fast-ssd",
		"slow-hdd": "slow-hdd",
	}

	for _, pair := range pairs {
		expectedTarget, exists := expectedMappings[pair.Source.Name]
		if !exists {
			t.Errorf("Unexpected source storage: %s", pair.Source.Name)
			continue
		}

		if pair.Destination.StorageClass != expectedTarget {
			t.Errorf("Source %s mapped to %s, expected %s",
				pair.Source.Name, pair.Destination.StorageClass, expectedTarget)
		}
	}
}

func TestOpenShiftStorageMapper_ImplementsInterface(t *testing.T) {
	var _ storagemapper.StorageMapper = &OpenShiftStorageMapper{}
}
