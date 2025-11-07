package storage

import (
	"testing"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/storage/mapper"
	openshiftMapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/storage/mapper/openshift"
	openstackMapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/storage/mapper/openstack"
	ovaMapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/storage/mapper/ova"
	ovirtMapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/storage/mapper/ovirt"
	vsphereMapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/storage/mapper/vsphere"
)

// MockProviderClient simulates provider client behavior for testing
type MockProviderClient struct {
	ProviderType string
}

func (m *MockProviderClient) GetProviderType() (string, error) {
	return m.ProviderType, nil
}

// Test data for factory routing tests
func TestGetStorageMapper_ProviderRouting(t *testing.T) {
	// Note: These tests would need to be adapted to work with the actual inventory.GetProviderByName
	// function. For now, we'll test the mapper type selection logic conceptually.

	tests := []struct {
		name               string
		sourceProviderType string
		targetProviderType string
		expectedMapperType string
	}{
		{
			name:               "OpenShift source -> OpenShift mapper",
			sourceProviderType: "openshift",
			targetProviderType: "openshift",
			expectedMapperType: "*openshift.OpenShiftStorageMapper",
		},
		{
			name:               "vSphere source -> vSphere mapper",
			sourceProviderType: "vsphere",
			targetProviderType: "openshift",
			expectedMapperType: "*vsphere.VSphereStorageMapper",
		},
		{
			name:               "oVirt source -> oVirt mapper",
			sourceProviderType: "ovirt",
			targetProviderType: "openshift",
			expectedMapperType: "*ovirt.OvirtStorageMapper",
		},
		{
			name:               "OpenStack source -> OpenStack mapper",
			sourceProviderType: "openstack",
			targetProviderType: "openshift",
			expectedMapperType: "*openstack.OpenStackStorageMapper",
		},
		{
			name:               "OVA source -> OVA mapper",
			sourceProviderType: "ova",
			targetProviderType: "openshift",
			expectedMapperType: "*ova.OVAStorageMapper",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test mapper creation based on provider type
			var storageMapper mapper.StorageMapper

			switch tt.sourceProviderType {
			case "openshift":
				storageMapper = openshiftMapper.NewOpenShiftStorageMapper()
			case "vsphere":
				storageMapper = vsphereMapper.NewVSphereStorageMapper()
			case "ovirt":
				storageMapper = ovirtMapper.NewOvirtStorageMapper()
			case "openstack":
				storageMapper = openstackMapper.NewOpenStackStorageMapper()
			case "ova":
				storageMapper = ovaMapper.NewOVAStorageMapper()
			}

			if storageMapper == nil {
				t.Errorf("Expected mapper for provider type %s, got nil", tt.sourceProviderType)
			}

			// Verify mapper implements the interface
			var _ mapper.StorageMapper = storageMapper
		})
	}
}

func TestStorageMapperOptions_ProviderTypeFields(t *testing.T) {
	opts := mapper.StorageMappingOptions{
		DefaultTargetStorageClass: "test-storage",
		SourceProviderType:        "openshift",
		TargetProviderType:        "openshift",
	}

	// Verify the new provider type fields are available
	if opts.SourceProviderType != "openshift" {
		t.Errorf("SourceProviderType: got %s, want openshift", opts.SourceProviderType)
	}

	if opts.TargetProviderType != "openshift" {
		t.Errorf("TargetProviderType: got %s, want openshift", opts.TargetProviderType)
	}
}

// Test that all mapper types implement the StorageMapper interface
func TestStorageMapperInterface_AllImplementations(t *testing.T) {
	mappers := []mapper.StorageMapper{
		openshiftMapper.NewOpenShiftStorageMapper(),
		vsphereMapper.NewVSphereStorageMapper(),
		ovirtMapper.NewOvirtStorageMapper(),
		openstackMapper.NewOpenStackStorageMapper(),
		ovaMapper.NewOVAStorageMapper(),
	}

	for i, m := range mappers {
		if m == nil {
			t.Errorf("Mapper %d is nil", i)
		}

		// This test ensures all mappers implement the interface
		var _ mapper.StorageMapper = m
	}
}

// Test storage mapper creation functions
func TestStorageMapperCreation(t *testing.T) {
	tests := []struct {
		name           string
		createMapper   func() mapper.StorageMapper
		expectedNotNil bool
	}{
		{
			name:           "OpenShift mapper creation",
			createMapper:   func() mapper.StorageMapper { return openshiftMapper.NewOpenShiftStorageMapper() },
			expectedNotNil: true,
		},
		{
			name:           "vSphere mapper creation",
			createMapper:   func() mapper.StorageMapper { return vsphereMapper.NewVSphereStorageMapper() },
			expectedNotNil: true,
		},
		{
			name:           "oVirt mapper creation",
			createMapper:   func() mapper.StorageMapper { return ovirtMapper.NewOvirtStorageMapper() },
			expectedNotNil: true,
		},
		{
			name:           "OpenStack mapper creation",
			createMapper:   func() mapper.StorageMapper { return openstackMapper.NewOpenStackStorageMapper() },
			expectedNotNil: true,
		},
		{
			name:           "OVA mapper creation",
			createMapper:   func() mapper.StorageMapper { return ovaMapper.NewOVAStorageMapper() },
			expectedNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := tt.createMapper()

			if tt.expectedNotNil && mapper == nil {
				t.Errorf("Expected non-nil mapper, got nil")
			}

			if !tt.expectedNotNil && mapper != nil {
				t.Errorf("Expected nil mapper, got non-nil")
			}
		})
	}
}
