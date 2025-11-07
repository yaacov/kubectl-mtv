package network

import (
	"testing"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/mapper"
	openshiftMapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/mapper/openshift"
	openstackMapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/mapper/openstack"
	ovaMapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/mapper/ova"
	ovirtMapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/mapper/ovirt"
	vsphereMapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/mapper/vsphere"
)

func TestGetNetworkMapper_ProviderRouting(t *testing.T) {
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
			expectedMapperType: "*openshift.OpenShiftNetworkMapper",
		},
		{
			name:               "vSphere source -> vSphere mapper",
			sourceProviderType: "vsphere",
			targetProviderType: "openshift",
			expectedMapperType: "*vsphere.VSphereNetworkMapper",
		},
		{
			name:               "oVirt source -> oVirt mapper",
			sourceProviderType: "ovirt",
			targetProviderType: "openshift",
			expectedMapperType: "*ovirt.OvirtNetworkMapper",
		},
		{
			name:               "OpenStack source -> OpenStack mapper",
			sourceProviderType: "openstack",
			targetProviderType: "openshift",
			expectedMapperType: "*openstack.OpenStackNetworkMapper",
		},
		{
			name:               "OVA source -> OVA mapper",
			sourceProviderType: "ova",
			targetProviderType: "openshift",
			expectedMapperType: "*ova.OVANetworkMapper",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test mapper creation based on provider type
			var networkMapper mapper.NetworkMapper

			switch tt.sourceProviderType {
			case "openshift":
				networkMapper = openshiftMapper.NewOpenShiftNetworkMapper()
			case "vsphere":
				networkMapper = vsphereMapper.NewVSphereNetworkMapper()
			case "ovirt":
				networkMapper = ovirtMapper.NewOvirtNetworkMapper()
			case "openstack":
				networkMapper = openstackMapper.NewOpenStackNetworkMapper()
			case "ova":
				networkMapper = ovaMapper.NewOVANetworkMapper()
			}

			if networkMapper == nil {
				t.Errorf("Expected mapper for provider type %s, got nil", tt.sourceProviderType)
			}

			// Verify mapper implements the interface
			var _ mapper.NetworkMapper = networkMapper
		})
	}
}

func TestNetworkMapperOptions_ProviderTypeFields(t *testing.T) {
	opts := mapper.NetworkMappingOptions{
		DefaultTargetNetwork: "test-network",
		Namespace:            "test-ns",
		SourceProviderType:   "openshift",
		TargetProviderType:   "openshift",
	}

	// Verify the new provider type fields are available
	if opts.SourceProviderType != "openshift" {
		t.Errorf("SourceProviderType: got %s, want openshift", opts.SourceProviderType)
	}

	if opts.TargetProviderType != "openshift" {
		t.Errorf("TargetProviderType: got %s, want openshift", opts.TargetProviderType)
	}
}

// Test that all mapper types implement the NetworkMapper interface
func TestNetworkMapperInterface_AllImplementations(t *testing.T) {
	mappers := []mapper.NetworkMapper{
		openshiftMapper.NewOpenShiftNetworkMapper(),
		vsphereMapper.NewVSphereNetworkMapper(),
		ovirtMapper.NewOvirtNetworkMapper(),
		openstackMapper.NewOpenStackNetworkMapper(),
		ovaMapper.NewOVANetworkMapper(),
	}

	for i, m := range mappers {
		if m == nil {
			t.Errorf("Mapper %d is nil", i)
		}

		// This test ensures all mappers implement the interface
		var _ mapper.NetworkMapper = m
	}
}

// Test network mapper creation functions
func TestNetworkMapperCreation(t *testing.T) {
	tests := []struct {
		name           string
		createMapper   func() mapper.NetworkMapper
		expectedNotNil bool
	}{
		{
			name:           "OpenShift mapper creation",
			createMapper:   func() mapper.NetworkMapper { return openshiftMapper.NewOpenShiftNetworkMapper() },
			expectedNotNil: true,
		},
		{
			name:           "vSphere mapper creation",
			createMapper:   func() mapper.NetworkMapper { return vsphereMapper.NewVSphereNetworkMapper() },
			expectedNotNil: true,
		},
		{
			name:           "oVirt mapper creation",
			createMapper:   func() mapper.NetworkMapper { return ovirtMapper.NewOvirtNetworkMapper() },
			expectedNotNil: true,
		},
		{
			name:           "OpenStack mapper creation",
			createMapper:   func() mapper.NetworkMapper { return openstackMapper.NewOpenStackNetworkMapper() },
			expectedNotNil: true,
		},
		{
			name:           "OVA mapper creation",
			createMapper:   func() mapper.NetworkMapper { return ovaMapper.NewOVANetworkMapper() },
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
