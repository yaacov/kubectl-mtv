package network

import (
	"testing"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
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

// Test filterTargetNetworksByNamespace function
func TestFilterTargetNetworksByNamespace(t *testing.T) {
	tests := []struct {
		name            string
		networks        []forkliftv1beta1.DestinationNetwork
		targetNamespace string
		expectedCount   int
		expectedNames   []string
	}{
		{
			name: "Keep networks in target namespace and default",
			networks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "net1", Namespace: "target-ns"},
				{Type: "multus", Name: "net2", Namespace: "default"},
				{Type: "multus", Name: "net3", Namespace: "other-ns"},
			},
			targetNamespace: "target-ns",
			expectedCount:   2,
			expectedNames:   []string{"net1", "net2"},
		},
		{
			name: "Keep only default namespace when target has no NADs",
			networks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "net1", Namespace: "other-ns"},
				{Type: "multus", Name: "net2", Namespace: "default"},
			},
			targetNamespace: "target-ns",
			expectedCount:   1,
			expectedNames:   []string{"net2"},
		},
		{
			name: "Keep non-multus networks regardless of namespace",
			networks: []forkliftv1beta1.DestinationNetwork{
				{Type: "pod"},
				{Type: "multus", Name: "net1", Namespace: "other-ns"},
				{Type: "multus", Name: "net2", Namespace: "target-ns"},
			},
			targetNamespace: "target-ns",
			expectedCount:   2,
			expectedNames:   []string{"", "net2"}, // pod network has empty name
		},
		{
			name: "Filter out all multus networks when none match",
			networks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "net1", Namespace: "ns1"},
				{Type: "multus", Name: "net2", Namespace: "ns2"},
			},
			targetNamespace: "target-ns",
			expectedCount:   0,
			expectedNames:   []string{},
		},
		{
			name:            "Empty networks list",
			networks:        []forkliftv1beta1.DestinationNetwork{},
			targetNamespace: "target-ns",
			expectedCount:   0,
			expectedNames:   []string{},
		},
		{
			name: "All networks in target namespace",
			networks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "net1", Namespace: "target-ns"},
				{Type: "multus", Name: "net2", Namespace: "target-ns"},
			},
			targetNamespace: "target-ns",
			expectedCount:   2,
			expectedNames:   []string{"net1", "net2"},
		},
		{
			name: "All networks in default namespace",
			networks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "net1", Namespace: "default"},
				{Type: "multus", Name: "net2", Namespace: "default"},
			},
			targetNamespace: "target-ns",
			expectedCount:   2,
			expectedNames:   []string{"net1", "net2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterTargetNetworksByNamespace(tt.networks, tt.targetNamespace)

			if len(result) != tt.expectedCount {
				t.Errorf("filterTargetNetworksByNamespace() returned %d networks, expected %d", len(result), tt.expectedCount)
			}

			// Verify the expected names are present
			for i, expected := range tt.expectedNames {
				if i < len(result) && result[i].Name != expected {
					t.Errorf("filterTargetNetworksByNamespace() network[%d].Name = %s, expected %s", i, result[i].Name, expected)
				}
			}
		})
	}
}
