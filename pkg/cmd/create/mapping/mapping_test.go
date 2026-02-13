package mapping

import (
	"strings"
	"testing"
)

// --- parseProviderReference ---

func TestParseProviderReference_NameOnly(t *testing.T) {
	name, ns := parseProviderReference("my-provider", "default-ns")
	if name != "my-provider" {
		t.Errorf("expected name 'my-provider', got %q", name)
	}
	if ns != "default-ns" {
		t.Errorf("expected namespace 'default-ns', got %q", ns)
	}
}

func TestParseProviderReference_NamespaceSlashName(t *testing.T) {
	name, ns := parseProviderReference("other-ns/my-provider", "default-ns")
	if name != "my-provider" {
		t.Errorf("expected name 'my-provider', got %q", name)
	}
	if ns != "other-ns" {
		t.Errorf("expected namespace 'other-ns', got %q", ns)
	}
}

func TestParseProviderReference_TrimsSpaces(t *testing.T) {
	name, ns := parseProviderReference("  ns / name  ", "default")
	if name != "name" {
		t.Errorf("expected name 'name', got %q", name)
	}
	if ns != "ns" {
		t.Errorf("expected namespace 'ns', got %q", ns)
	}
}

func TestParseProviderReference_EmptyDefault(t *testing.T) {
	name, ns := parseProviderReference("provider", "")
	if name != "provider" {
		t.Errorf("expected name 'provider', got %q", name)
	}
	if ns != "" {
		t.Errorf("expected empty namespace, got %q", ns)
	}
}

func TestParseProviderReference_MultipleSlashes(t *testing.T) {
	// SplitN with n=2 means only the first slash is used as separator
	name, ns := parseProviderReference("ns/name/extra", "default")
	if name != "name/extra" {
		t.Errorf("expected name 'name/extra', got %q", name)
	}
	if ns != "ns" {
		t.Errorf("expected namespace 'ns', got %q", ns)
	}
}

// --- validateNetworkPairsTargets ---

func TestValidateNetworkPairsTargets_Empty(t *testing.T) {
	err := validateNetworkPairsTargets("")
	if err != nil {
		t.Errorf("expected nil for empty input, got: %v", err)
	}
}

func TestValidateNetworkPairsTargets_SingleDefault(t *testing.T) {
	err := validateNetworkPairsTargets("source1:default")
	if err != nil {
		t.Errorf("expected nil for single default, got: %v", err)
	}
}

func TestValidateNetworkPairsTargets_DuplicateDefault(t *testing.T) {
	err := validateNetworkPairsTargets("source1:default,source2:default")
	if err == nil {
		t.Error("expected error for duplicate pod network")
	}
	if !strings.Contains(err.Error(), "Pod network") {
		t.Errorf("expected pod network error message, got: %s", err.Error())
	}
}

func TestValidateNetworkPairsTargets_MultipleNADs(t *testing.T) {
	// NAD targets can be reused
	err := validateNetworkPairsTargets("source1:ns/my-nad,source2:ns/my-nad")
	if err != nil {
		t.Errorf("expected nil for duplicate NAD targets, got: %v", err)
	}
}

func TestValidateNetworkPairsTargets_MultipleIgnored(t *testing.T) {
	err := validateNetworkPairsTargets("source1:ignored,source2:ignored")
	if err != nil {
		t.Errorf("expected nil for duplicate ignored targets, got: %v", err)
	}
}

func TestValidateNetworkPairsTargets_DefaultAndNAD(t *testing.T) {
	err := validateNetworkPairsTargets("source1:default,source2:ns/my-nad")
	if err != nil {
		t.Errorf("expected nil for default + NAD, got: %v", err)
	}
}

func TestValidateNetworkPairsTargets_MalformedPairsSkipped(t *testing.T) {
	// Malformed pairs (no colon) are skipped
	err := validateNetworkPairsTargets("nocolon,source1:default")
	if err != nil {
		t.Errorf("expected nil for malformed pairs, got: %v", err)
	}
}

func TestValidateNetworkPairsTargets_EmptyPairsInList(t *testing.T) {
	err := validateNetworkPairsTargets("source1:default,,source2:ns/nad")
	if err != nil {
		t.Errorf("expected nil with empty entries, got: %v", err)
	}
}

// --- validateVolumeMode ---

func TestValidateVolumeMode_Valid(t *testing.T) {
	validModes := []string{"Filesystem", "Block"}
	for _, mode := range validModes {
		if err := validateVolumeMode(mode); err != nil {
			t.Errorf("validateVolumeMode(%q) = error %v, want nil", mode, err)
		}
	}
}

func TestValidateVolumeMode_Invalid(t *testing.T) {
	invalidModes := []string{"filesystem", "block", "ReadWriteOnce", "", "unknown"}
	for _, mode := range invalidModes {
		if err := validateVolumeMode(mode); err == nil {
			t.Errorf("validateVolumeMode(%q) = nil, want error", mode)
		}
	}
}

// --- validateAccessMode ---

func TestValidateAccessMode_Valid(t *testing.T) {
	validModes := []string{"ReadWriteOnce", "ReadWriteMany", "ReadOnlyMany"}
	for _, mode := range validModes {
		if err := validateAccessMode(mode); err != nil {
			t.Errorf("validateAccessMode(%q) = error %v, want nil", mode, err)
		}
	}
}

func TestValidateAccessMode_Invalid(t *testing.T) {
	invalidModes := []string{"readwriteonce", "ReadWriteAll", "", "Block"}
	for _, mode := range invalidModes {
		if err := validateAccessMode(mode); err == nil {
			t.Errorf("validateAccessMode(%q) = nil, want error", mode)
		}
	}
}

// --- validateOffloadPlugin ---

func TestValidateOffloadPlugin_Valid(t *testing.T) {
	if err := validateOffloadPlugin("vsphere"); err != nil {
		t.Errorf("validateOffloadPlugin(vsphere) = error %v", err)
	}
}

func TestValidateOffloadPlugin_Invalid(t *testing.T) {
	invalidPlugins := []string{"ovirt", "openstack", "", "VSphere"}
	for _, plugin := range invalidPlugins {
		if err := validateOffloadPlugin(plugin); err == nil {
			t.Errorf("validateOffloadPlugin(%q) = nil, want error", plugin)
		}
	}
}

// --- validateOffloadVendor ---

func TestValidateOffloadVendor_Valid(t *testing.T) {
	validVendors := []string{
		"flashsystem", "vantara", "ontap", "primera3par",
		"pureFlashArray", "powerflex", "powermax", "powerstore", "infinibox",
	}
	for _, vendor := range validVendors {
		if err := validateOffloadVendor(vendor); err != nil {
			t.Errorf("validateOffloadVendor(%q) = error %v", vendor, err)
		}
	}
}

func TestValidateOffloadVendor_Invalid(t *testing.T) {
	invalidVendors := []string{"netapp", "", "FlashSystem", "unknown"}
	for _, vendor := range invalidVendors {
		if err := validateOffloadVendor(vendor); err == nil {
			t.Errorf("validateOffloadVendor(%q) = nil, want error", vendor)
		}
	}
}

// --- validateOffloadSecretFields ---

func TestValidateOffloadSecretFields_NoFieldsSet(t *testing.T) {
	opts := StorageCreateOptions{}
	if err := validateOffloadSecretFields(opts); err != nil {
		t.Errorf("expected nil for no fields, got: %v", err)
	}
}

func TestValidateOffloadSecretFields_AllRequired(t *testing.T) {
	opts := StorageCreateOptions{
		OffloadVSphereUsername: "user",
		OffloadVSpherePassword: "pass",
		OffloadVSphereURL:      "https://vcenter",
		OffloadStorageUsername: "storuser",
		OffloadStoragePassword: "storpass",
		OffloadStorageEndpoint: "https://storage",
	}
	if err := validateOffloadSecretFields(opts); err != nil {
		t.Errorf("expected nil for all required fields, got: %v", err)
	}
}

func TestValidateOffloadSecretFields_PartialFields(t *testing.T) {
	opts := StorageCreateOptions{
		OffloadVSphereUsername: "user",
		// Missing all other required fields
	}
	err := validateOffloadSecretFields(opts)
	if err == nil {
		t.Error("expected error for partial fields")
	}
	if !strings.Contains(err.Error(), "--offload-vsphere-password") {
		t.Errorf("expected missing password in error, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "--offload-storage-username") {
		t.Errorf("expected missing storage username in error, got: %s", err.Error())
	}
}

func TestValidateOffloadSecretFields_OnlyInsecureTLS(t *testing.T) {
	opts := StorageCreateOptions{
		OffloadInsecureSkipTLS: true,
	}
	err := validateOffloadSecretFields(opts)
	if err == nil {
		t.Error("expected error when only insecure TLS is set")
	}
}

func TestValidateOffloadSecretFields_OnlyCACert(t *testing.T) {
	opts := StorageCreateOptions{
		OffloadCACert: "/path/to/ca.crt",
	}
	err := validateOffloadSecretFields(opts)
	if err == nil {
		t.Error("expected error when only CA cert is set")
	}
}

func TestValidateOffloadSecretFields_AllRequiredPlusOptional(t *testing.T) {
	opts := StorageCreateOptions{
		OffloadVSphereUsername: "user",
		OffloadVSpherePassword: "pass",
		OffloadVSphereURL:      "https://vcenter",
		OffloadStorageUsername: "storuser",
		OffloadStoragePassword: "storpass",
		OffloadStorageEndpoint: "https://storage",
		OffloadCACert:          "/path/to/ca.crt",
		OffloadInsecureSkipTLS: true,
	}
	if err := validateOffloadSecretFields(opts); err != nil {
		t.Errorf("expected nil with all required + optional, got: %v", err)
	}
}

// --- needsOffloadSecret ---

func TestNeedsOffloadSecret_NoFieldsSet(t *testing.T) {
	opts := StorageCreateOptions{}
	if needsOffloadSecret(opts) {
		t.Error("expected false when no fields set")
	}
}

func TestNeedsOffloadSecret_ExistingSecretProvided(t *testing.T) {
	opts := StorageCreateOptions{
		DefaultOffloadSecret:   "existing-secret",
		OffloadVSphereUsername: "user",
	}
	if needsOffloadSecret(opts) {
		t.Error("expected false when existing secret is provided")
	}
}

func TestNeedsOffloadSecret_FieldsWithoutExistingSecret(t *testing.T) {
	tests := []struct {
		name string
		opts StorageCreateOptions
	}{
		{"username", StorageCreateOptions{OffloadVSphereUsername: "u"}},
		{"password", StorageCreateOptions{OffloadVSpherePassword: "p"}},
		{"url", StorageCreateOptions{OffloadVSphereURL: "https://vc"}},
		{"storage username", StorageCreateOptions{OffloadStorageUsername: "su"}},
		{"storage password", StorageCreateOptions{OffloadStoragePassword: "sp"}},
		{"storage endpoint", StorageCreateOptions{OffloadStorageEndpoint: "https://s"}},
		{"ca cert", StorageCreateOptions{OffloadCACert: "/path"}},
		{"insecure tls", StorageCreateOptions{OffloadInsecureSkipTLS: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !needsOffloadSecret(tt.opts) {
				t.Errorf("expected true when %s is set", tt.name)
			}
		})
	}
}
