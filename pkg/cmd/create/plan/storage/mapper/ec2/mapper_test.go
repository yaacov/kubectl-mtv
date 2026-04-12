package ec2

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"

	storagemapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/storage/mapper"
)

func TestEC2StorageMapper_DefaultSCBranch(t *testing.T) {
	m := NewEC2StorageMapper()

	sources := []ref.Ref{
		{Name: "gp3"},
		{Name: "io2"},
		{Name: "st1"},
	}
	targets := []forkliftv1beta1.DestinationStorage{
		{StorageClass: "ocs-storagecluster-ceph-rbd"},
	}

	opts := storagemapper.StorageMappingOptions{
		DefaultTargetStorageClass: "my-custom-sc",
		TargetProviderType:        "openshift",
	}

	pairs, err := m.CreateStoragePairs(sources, targets, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
	}
	for i, pair := range pairs {
		if pair.Destination.StorageClass != "my-custom-sc" {
			t.Errorf("pair %d: expected target 'my-custom-sc', got '%s'", i, pair.Destination.StorageClass)
		}
		if pair.Source.Name != sources[i].Name {
			t.Errorf("pair %d: expected source '%s', got '%s'", i, sources[i].Name, pair.Source.Name)
		}
	}
}

func TestEC2StorageMapper_AutoMatch(t *testing.T) {
	m := NewEC2StorageMapper()

	sources := []ref.Ref{
		{Name: "gp3"},
		{Name: "io2"},
	}
	targets := []forkliftv1beta1.DestinationStorage{
		{StorageClass: "ocs-storagecluster-ceph-rbd"},
		{StorageClass: "gp3-csi"},
		{StorageClass: "io2"},
	}

	opts := storagemapper.StorageMappingOptions{
		TargetProviderType: "openshift",
	}

	pairs, err := m.CreateStoragePairs(sources, targets, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// gp3 should match gp3-csi (prefix match)
	if pairs[0].Destination.StorageClass != "gp3-csi" {
		t.Errorf("gp3: expected target 'gp3-csi', got '%s'", pairs[0].Destination.StorageClass)
	}
	// io2 should match io2 (exact match)
	if pairs[1].Destination.StorageClass != "io2" {
		t.Errorf("io2: expected target 'io2', got '%s'", pairs[1].Destination.StorageClass)
	}
}

func TestEC2StorageMapper_GapFill(t *testing.T) {
	m := NewEC2StorageMapper()

	sources := []ref.Ref{
		{Name: "gp3"},
		{Name: "st1"},
	}
	// Only gp3-csi matches; st1 has no EBS-pattern match.
	// Default SC (index 0) is "ocs-storagecluster-ceph-rbd".
	targets := []forkliftv1beta1.DestinationStorage{
		{StorageClass: "ocs-storagecluster-ceph-rbd"},
		{StorageClass: "gp3-csi"},
	}

	opts := storagemapper.StorageMappingOptions{
		TargetProviderType: "openshift",
	}

	pairs, err := m.CreateStoragePairs(sources, targets, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// gp3 matches gp3-csi
	if pairs[0].Destination.StorageClass != "gp3-csi" {
		t.Errorf("gp3: expected target 'gp3-csi', got '%s'", pairs[0].Destination.StorageClass)
	}
	// st1 has no match -> gap-filled with default SC
	if pairs[1].Destination.StorageClass != "ocs-storagecluster-ceph-rbd" {
		t.Errorf("st1: expected gap-fill target 'ocs-storagecluster-ceph-rbd', got '%s'", pairs[1].Destination.StorageClass)
	}
}

func TestEC2StorageMapper_EmptySources(t *testing.T) {
	m := NewEC2StorageMapper()

	pairs, err := m.CreateStoragePairs(
		[]ref.Ref{},
		[]forkliftv1beta1.DestinationStorage{{StorageClass: "any"}},
		storagemapper.StorageMappingOptions{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 0 {
		t.Errorf("expected 0 pairs, got %d", len(pairs))
	}
}

func TestEC2StorageMapper_VolumeModes(t *testing.T) {
	m := NewEC2StorageMapper()

	sources := []ref.Ref{
		{Name: "gp3"},
		{Name: "st1"},
	}
	targets := []forkliftv1beta1.DestinationStorage{
		{StorageClass: "default-sc"},
	}

	opts := storagemapper.StorageMappingOptions{
		DefaultTargetStorageClass: "default-sc",
	}

	pairs, err := m.CreateStoragePairs(sources, targets, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// gp3 (SSD) -> Block mode
	if pairs[0].Destination.VolumeMode != corev1.PersistentVolumeBlock {
		t.Errorf("gp3: expected Block mode, got %s", pairs[0].Destination.VolumeMode)
	}
	// st1 (HDD) -> Filesystem mode
	if pairs[1].Destination.VolumeMode != corev1.PersistentVolumeFilesystem {
		t.Errorf("st1: expected Filesystem mode, got %s", pairs[1].Destination.VolumeMode)
	}
}

func TestEC2StorageMapper_ImplementsInterface(t *testing.T) {
	var _ storagemapper.StorageMapper = &EC2StorageMapper{}
}
