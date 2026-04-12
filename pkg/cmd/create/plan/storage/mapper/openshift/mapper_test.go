package openshift

import (
	"testing"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"

	storagemapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/storage/mapper"
)

func TestOpenShiftStorageMapper_DefaultSCBranch(t *testing.T) {
	tests := []struct {
		name                      string
		targetStorages            []forkliftv1beta1.DestinationStorage
		defaultTargetStorageClass string
		expectedTargetClass       string
	}{
		{
			name: "User-defined default takes priority over target list",
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "auto-selected"},
				{StorageClass: "other-class"},
			},
			defaultTargetStorageClass: "user-defined",
			expectedTargetClass:       "user-defined",
		},
		{
			name:                      "User-defined default with empty target list",
			targetStorages:            []forkliftv1beta1.DestinationStorage{},
			defaultTargetStorageClass: "user-defined",
			expectedTargetClass:       "user-defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewOpenShiftStorageMapper()

			sources := []ref.Ref{
				{Name: "sc-a", ID: "src-1"},
				{Name: "sc-b", ID: "src-2"},
			}
			opts := storagemapper.StorageMappingOptions{
				DefaultTargetStorageClass: tt.defaultTargetStorageClass,
				SourceProviderType:        "openshift",
				TargetProviderType:        "openshift",
			}

			pairs, err := m.CreateStoragePairs(sources, tt.targetStorages, opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(pairs) != 2 {
				t.Fatalf("expected 2 pairs, got %d", len(pairs))
			}
			for i, pair := range pairs {
				if pair.Destination.StorageClass != tt.expectedTargetClass {
					t.Errorf("pair %d: expected target '%s', got '%s'", i, tt.expectedTargetClass, pair.Destination.StorageClass)
				}
			}
		})
	}
}

func TestOpenShiftStorageMapper_SameNameMatching(t *testing.T) {
	tests := []struct {
		name                string
		sourceStorages      []ref.Ref
		targetStorages      []forkliftv1beta1.DestinationStorage
		targetProviderType  string
		expectedPairs       int
		expectedTargetNames []string
	}{
		{
			name: "OCP-to-OCP: All sources match by name",
			sourceStorages: []ref.Ref{
				{Name: "fast-ssd", ID: "src-1"},
				{Name: "slow-hdd", ID: "src-2"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "default-sc"},
				{StorageClass: "fast-ssd"},
				{StorageClass: "slow-hdd"},
			},
			targetProviderType:  "openshift",
			expectedPairs:       2,
			expectedTargetNames: []string{"fast-ssd", "slow-hdd"},
		},
		{
			name: "OCP-to-OCP: Partial match — matched sources get same-name, unmatched get default",
			sourceStorages: []ref.Ref{
				{Name: "fast-ssd", ID: "src-1"},
				{Name: "unknown-storage", ID: "src-2"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "default-sc"},
				{StorageClass: "fast-ssd"},
				{StorageClass: "slow-hdd"},
			},
			targetProviderType:  "openshift",
			expectedPairs:       2,
			expectedTargetNames: []string{"fast-ssd", "default-sc"},
		},
		{
			name: "OCP-to-OCP: No name matches — all get default",
			sourceStorages: []ref.Ref{
				{Name: "src-only-a", ID: "src-1"},
				{Name: "src-only-b", ID: "src-2"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "default-sc"},
				{StorageClass: "tgt-only"},
			},
			targetProviderType:  "openshift",
			expectedPairs:       2,
			expectedTargetNames: []string{"default-sc", "default-sc"},
		},
		{
			name: "OCP-to-non-OCP: All sources go to default (no same-name attempt)",
			sourceStorages: []ref.Ref{
				{Name: "fast-ssd", ID: "src-1"},
				{Name: "slow-hdd", ID: "src-2"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "ocs-storagecluster-ceph-rbd"},
				{StorageClass: "fast-ssd"},
			},
			targetProviderType:  "vsphere",
			expectedPairs:       2,
			expectedTargetNames: []string{"ocs-storagecluster-ceph-rbd", "ocs-storagecluster-ceph-rbd"},
		},
		{
			name:                "Empty sources",
			sourceStorages:      []ref.Ref{},
			targetStorages:      []forkliftv1beta1.DestinationStorage{{StorageClass: "default"}},
			targetProviderType:  "openshift",
			expectedPairs:       0,
			expectedTargetNames: []string{},
		},
		{
			name: "OCP-to-OCP: Single source with match",
			sourceStorages: []ref.Ref{
				{Name: "fast-ssd", ID: "src-1"},
			},
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "default-sc"},
				{StorageClass: "fast-ssd"},
			},
			targetProviderType:  "openshift",
			expectedPairs:       1,
			expectedTargetNames: []string{"fast-ssd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewOpenShiftStorageMapper()

			opts := storagemapper.StorageMappingOptions{
				SourceProviderType: "openshift",
				TargetProviderType: tt.targetProviderType,
			}

			pairs, err := m.CreateStoragePairs(tt.sourceStorages, tt.targetStorages, opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(pairs) != tt.expectedPairs {
				t.Fatalf("expected %d pairs, got %d", tt.expectedPairs, len(pairs))
			}

			for i, pair := range pairs {
				if i < len(tt.expectedTargetNames) {
					if pair.Destination.StorageClass != tt.expectedTargetNames[i] {
						t.Errorf("pair %d: expected target '%s', got '%s'",
							i, tt.expectedTargetNames[i], pair.Destination.StorageClass)
					}
				}
				if i < len(tt.sourceStorages) {
					if pair.Source.Name != tt.sourceStorages[i].Name {
						t.Errorf("pair %d: source name '%s' != expected '%s'",
							i, pair.Source.Name, tt.sourceStorages[i].Name)
					}
				}
			}
		})
	}
}

func TestOpenShiftStorageMapper_AutoSelectDefault(t *testing.T) {
	tests := []struct {
		name                 string
		targetStorages       []forkliftv1beta1.DestinationStorage
		expectedDefaultClass string
	}{
		{
			name: "Auto-selected uses first target (best from fetcher)",
			targetStorages: []forkliftv1beta1.DestinationStorage{
				{StorageClass: "best-sc"},
				{StorageClass: "second-sc"},
			},
			expectedDefaultClass: "best-sc",
		},
		{
			name:                 "Empty target list falls back to system default",
			targetStorages:       []forkliftv1beta1.DestinationStorage{},
			expectedDefaultClass: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewOpenShiftStorageMapper()

			sources := []ref.Ref{{Name: "test-source", ID: "src-1"}}
			opts := storagemapper.StorageMappingOptions{
				SourceProviderType: "openshift",
				TargetProviderType: "vsphere",
			}

			pairs, err := m.CreateStoragePairs(sources, tt.targetStorages, opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(pairs) != 1 {
				t.Fatalf("expected 1 pair, got %d", len(pairs))
			}
			if pairs[0].Destination.StorageClass != tt.expectedDefaultClass {
				t.Errorf("expected default class '%s', got '%s'",
					tt.expectedDefaultClass, pairs[0].Destination.StorageClass)
			}
		})
	}
}

func TestOpenShiftStorageMapper_ImplementsInterface(t *testing.T) {
	var _ storagemapper.StorageMapper = &OpenShiftStorageMapper{}
}
