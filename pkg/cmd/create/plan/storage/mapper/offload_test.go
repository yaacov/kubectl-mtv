package mapper

import (
	"strings"
	"testing"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"github.com/yaacov/kubectl-mtv/pkg/util/flags"
)

func TestApplyOffloadToPairs_VSphereNfs(t *testing.T) {
	pairs := []forkliftv1beta1.StoragePair{{
		Destination: forkliftv1beta1.DestinationStorage{StorageClass: "pure-nfs"},
	}}
	got, err := ApplyOffloadToPairs(pairs, StorageMappingOptions{
		DefaultOffloadPlugin: flags.NfsOffloadPlugin,
		DefaultOffloadVendor: flags.NfsOffloadVendor,
		DefaultOffloadSecret: "pure-creds",
		DefaultVolumeMode:    "Filesystem",
	})
	if err != nil {
		t.Fatalf("ApplyOffloadToPairs() error = %v", err)
	}
	cfg := got[0].OffloadPlugin.VSphereNfsCopyPluginConfig
	if cfg == nil || cfg.SecretRef != "pure-creds" {
		t.Fatalf("nfs config = %#v", got[0].OffloadPlugin)
	}
	if got[0].Destination.VolumeMode != corev1.PersistentVolumeFilesystem {
		t.Fatalf("volume mode = %q", got[0].Destination.VolumeMode)
	}
}

func TestApplyOffloadToPairs_VSphereNfsRejectsBlock(t *testing.T) {
	pairs := []forkliftv1beta1.StoragePair{{
		Destination: forkliftv1beta1.DestinationStorage{StorageClass: "pure-nfs"},
	}}
	_, err := ApplyOffloadToPairs(pairs, StorageMappingOptions{
		DefaultOffloadPlugin: flags.NfsOffloadPlugin,
		DefaultOffloadVendor: flags.NfsOffloadVendor,
		DefaultVolumeMode:    "Block",
	})
	if err == nil || !strings.Contains(err.Error(), "Filesystem") {
		t.Fatalf("error = %v, want Filesystem requirement", err)
	}
}
