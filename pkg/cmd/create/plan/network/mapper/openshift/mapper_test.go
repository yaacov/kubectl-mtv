package openshift

import (
	"testing"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"

	networkmapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/mapper"
)

func TestOpenShiftNetworkMapper_CreateNetworkPairs_SameNameMatching(t *testing.T) {
	tests := []struct {
		name                string
		sourceNetworks      []ref.Ref
		targetNetworks      []forkliftv1beta1.DestinationNetwork
		sourceProviderType  string
		targetProviderType  string
		expectedPairs       int
		expectedSameName    bool
		expectedTargetNames []string
		expectedTargetTypes []string
	}{
		{
			name: "OCP-to-OCP: All sources match by name",
			sourceNetworks: []ref.Ref{
				{Name: "management-net", ID: "net-1"},
				{Name: "storage-net", ID: "net-2"},
			},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "management-net", Namespace: "default"},
				{Type: "multus", Name: "storage-net", Namespace: "default"},
				{Type: "multus", Name: "backup-net", Namespace: "default"},
			},
			sourceProviderType:  "openshift",
			targetProviderType:  "openshift",
			expectedPairs:       2,
			expectedSameName:    true,
			expectedTargetNames: []string{"management-net", "storage-net"},
			expectedTargetTypes: []string{"multus", "multus"},
		},
		{
			name: "OCP-to-OCP: Some sources don't match - fallback to default",
			sourceNetworks: []ref.Ref{
				{Name: "management-net", ID: "net-1"},
				{Name: "unknown-net", ID: "net-2"},
			},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "management-net", Namespace: "default"},
				{Type: "multus", Name: "storage-net", Namespace: "default"},
			},
			sourceProviderType:  "openshift",
			targetProviderType:  "openshift",
			expectedPairs:       2,
			expectedSameName:    false,
			expectedTargetNames: []string{"management-net", ""},
			expectedTargetTypes: []string{"multus", "ignored"},
		},
		{
			name: "OCP-to-OCP: More sources than targets - fallback",
			sourceNetworks: []ref.Ref{
				{Name: "net1", ID: "net-1"},
				{Name: "net2", ID: "net-2"},
				{Name: "net3", ID: "net-3"},
			},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "net1", Namespace: "default"},
				{Type: "multus", Name: "net2", Namespace: "default"},
			},
			sourceProviderType:  "openshift",
			targetProviderType:  "openshift",
			expectedPairs:       3,
			expectedSameName:    false,
			expectedTargetNames: []string{"net1", "", ""},
			expectedTargetTypes: []string{"multus", "ignored", "ignored"},
		},
		{
			name: "OCP-to-non-OCP: Use default behavior",
			sourceNetworks: []ref.Ref{
				{Name: "VM Network", ID: "net-1"},
				{Name: "Management Network", ID: "net-2"},
			},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "multus-net", Namespace: "default"},
			},
			sourceProviderType:  "openshift",
			targetProviderType:  "vsphere",
			expectedPairs:       2,
			expectedSameName:    false,
			expectedTargetNames: []string{"multus-net", ""},
			expectedTargetTypes: []string{"multus", "ignored"},
		},
		{
			name: "OCP-to-OCP: No multus targets - fallback to pod",
			sourceNetworks: []ref.Ref{
				{Name: "management-net", ID: "net-1"},
			},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "pod"},
			},
			sourceProviderType:  "openshift",
			targetProviderType:  "openshift",
			expectedPairs:       1,
			expectedSameName:    false,
			expectedTargetNames: []string{""},
			expectedTargetTypes: []string{"pod"},
		},
		{
			name:                "OCP-to-OCP: Empty sources",
			sourceNetworks:      []ref.Ref{},
			targetNetworks:      []forkliftv1beta1.DestinationNetwork{{Type: "multus", Name: "default"}},
			sourceProviderType:  "openshift",
			targetProviderType:  "openshift",
			expectedPairs:       0,
			expectedSameName:    false,
			expectedTargetNames: []string{},
			expectedTargetTypes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			networkMapper := NewOpenShiftNetworkMapper()

			opts := networkmapper.NetworkMappingOptions{
				SourceProviderType: tt.sourceProviderType,
				TargetProviderType: tt.targetProviderType,
				Namespace:          "default",
			}

			pairs, err := networkMapper.CreateNetworkPairs(tt.sourceNetworks, tt.targetNetworks, opts)
			if err != nil {
				t.Fatalf("CreateNetworkPairs() error = %v", err)
			}

			if len(pairs) != tt.expectedPairs {
				t.Errorf("CreateNetworkPairs() got %d pairs, want %d", len(pairs), tt.expectedPairs)
			}

			// Verify target network names and types
			for i, pair := range pairs {
				if i < len(tt.expectedTargetNames) {
					if pair.Destination.Name != tt.expectedTargetNames[i] {
						t.Errorf("Pair %d: got target name %s, want %s", i, pair.Destination.Name, tt.expectedTargetNames[i])
					}
				}
				if i < len(tt.expectedTargetTypes) {
					if pair.Destination.Type != tt.expectedTargetTypes[i] {
						t.Errorf("Pair %d: got target type %s, want %s", i, pair.Destination.Type, tt.expectedTargetTypes[i])
					}
				}
			}

			// Verify source names are preserved
			for i, pair := range pairs {
				if i < len(tt.sourceNetworks) {
					if pair.Source.Name != tt.sourceNetworks[i].Name {
						t.Errorf("Pair %d: source name %s != expected %s", i, pair.Source.Name, tt.sourceNetworks[i].Name)
					}
				}
			}
		})
	}
}

func TestOpenShiftNetworkMapper_CreateNetworkPairs_DefaultNetworkSelection(t *testing.T) {
	tests := []struct {
		name                 string
		targetNetworks       []forkliftv1beta1.DestinationNetwork
		defaultTargetNetwork string
		expectedType         string
		expectedName         string
		expectedNamespace    string
	}{
		{
			name: "User-defined default network",
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "auto-selected", Namespace: "default"},
			},
			defaultTargetNetwork: "custom-namespace/custom-net",
			expectedType:         "multus",
			expectedName:         "custom-net",
			expectedNamespace:    "custom-namespace",
		},
		{
			name: "User-defined pod networking",
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "some-net", Namespace: "default"},
			},
			defaultTargetNetwork: "default",
			expectedType:         "pod",
			expectedName:         "",
			expectedNamespace:    "",
		},
		{
			name: "Auto-select first multus network",
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "first-multus", Namespace: "default"},
				{Type: "multus", Name: "second-multus", Namespace: "default"},
			},
			defaultTargetNetwork: "",
			expectedType:         "multus",
			expectedName:         "first-multus",
			expectedNamespace:    "default",
		},
		{
			name: "Fallback to pod networking",
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "pod"},
			},
			defaultTargetNetwork: "",
			expectedType:         "pod",
			expectedName:         "",
			expectedNamespace:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			networkMapper := NewOpenShiftNetworkMapper()

			sourceNetworks := []ref.Ref{{Name: "test-source", ID: "net-1"}}
			opts := networkmapper.NetworkMappingOptions{
				DefaultTargetNetwork: tt.defaultTargetNetwork,
				Namespace:            "default",
				SourceProviderType:   "openshift",
				TargetProviderType:   "vsphere", // Force default behavior
			}

			pairs, err := networkMapper.CreateNetworkPairs(sourceNetworks, tt.targetNetworks, opts)
			if err != nil {
				t.Fatalf("CreateNetworkPairs() error = %v", err)
			}

			if len(pairs) > 0 {
				dest := pairs[0].Destination
				if dest.Type != tt.expectedType {
					t.Errorf("Default network type: got %s, want %s", dest.Type, tt.expectedType)
				}
				if dest.Name != tt.expectedName {
					t.Errorf("Default network name: got %s, want %s", dest.Name, tt.expectedName)
				}
				if dest.Namespace != tt.expectedNamespace {
					t.Errorf("Default network namespace: got %s, want %s", dest.Namespace, tt.expectedNamespace)
				}
			}
		})
	}
}

func TestCanMatchAllNetworksByName(t *testing.T) {
	tests := []struct {
		name           string
		sourceNetworks []ref.Ref
		targetNetworks []forkliftv1beta1.DestinationNetwork
		expected       bool
	}{
		{
			name: "All sources match with unique targets",
			sourceNetworks: []ref.Ref{
				{Name: "net1"},
				{Name: "net2"},
			},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "net1"},
				{Type: "multus", Name: "net2"},
				{Type: "multus", Name: "net3"},
			},
			expected: true,
		},
		{
			name: "Some sources don't match",
			sourceNetworks: []ref.Ref{
				{Name: "net1"},
				{Name: "unknown-net"},
			},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "net1"},
				{Type: "multus", Name: "net2"},
			},
			expected: false,
		},
		{
			name: "More sources than targets",
			sourceNetworks: []ref.Ref{
				{Name: "net1"},
				{Name: "net2"},
				{Name: "net3"},
			},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "net1"},
				{Type: "multus", Name: "net2"},
			},
			expected: false,
		},
		{
			name: "Non-multus targets ignored",
			sourceNetworks: []ref.Ref{
				{Name: "net1"},
			},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "pod"},
				{Type: "multus", Name: "net1"},
			},
			expected: true,
		},
		{
			name:           "Empty sources",
			sourceNetworks: []ref.Ref{},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "any-net"},
			},
			expected: true,
		},
		{
			name: "Empty targets",
			sourceNetworks: []ref.Ref{
				{Name: "some-net"},
			},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canMatchAllNetworksByName(tt.sourceNetworks, tt.targetNetworks)
			if result != tt.expected {
				t.Errorf("canMatchAllNetworksByName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCreateSameNameNetworkPairs(t *testing.T) {
	sourceNetworks := []ref.Ref{
		{Name: "management-net", ID: "net-1"},
		{Name: "storage-net", ID: "net-2"},
	}

	targetNetworks := []forkliftv1beta1.DestinationNetwork{
		{Type: "multus", Name: "management-net", Namespace: "default"},
		{Type: "multus", Name: "storage-net", Namespace: "default"},
		{Type: "multus", Name: "extra-net", Namespace: "default"},
		{Type: "pod"}, // Should be ignored
	}

	pairs, err := createSameNameNetworkPairs(sourceNetworks, targetNetworks)
	if err != nil {
		t.Fatalf("createSameNameNetworkPairs() error = %v", err)
	}

	if len(pairs) != 2 {
		t.Errorf("Expected 2 pairs, got %d", len(pairs))
	}

	// Verify mappings
	expectedMappings := map[string]string{
		"management-net": "management-net",
		"storage-net":    "storage-net",
	}

	for _, pair := range pairs {
		expectedTarget, exists := expectedMappings[pair.Source.Name]
		if !exists {
			t.Errorf("Unexpected source network: %s", pair.Source.Name)
			continue
		}

		if pair.Destination.Name != expectedTarget {
			t.Errorf("Source %s mapped to %s, expected %s",
				pair.Source.Name, pair.Destination.Name, expectedTarget)
		}

		if pair.Destination.Type != "multus" {
			t.Errorf("Expected multus type for network %s, got %s",
				pair.Source.Name, pair.Destination.Type)
		}
	}
}

func TestParseDefaultNetwork(t *testing.T) {
	tests := []struct {
		name                 string
		defaultTargetNetwork string
		namespace            string
		expectedType         string
		expectedName         string
		expectedNamespace    string
	}{
		{
			name:                 "Default pod networking",
			defaultTargetNetwork: "default",
			namespace:            "test-ns",
			expectedType:         "pod",
			expectedName:         "",
			expectedNamespace:    "",
		},
		{
			name:                 "Ignored network",
			defaultTargetNetwork: "ignored",
			namespace:            "test-ns",
			expectedType:         "ignored",
			expectedName:         "",
			expectedNamespace:    "",
		},
		{
			name:                 "Namespace/name format",
			defaultTargetNetwork: "custom-ns/custom-net",
			namespace:            "test-ns",
			expectedType:         "multus",
			expectedName:         "custom-net",
			expectedNamespace:    "custom-ns",
		},
		{
			name:                 "Empty namespace in format",
			defaultTargetNetwork: "/custom-net",
			namespace:            "test-ns",
			expectedType:         "multus",
			expectedName:         "custom-net",
			expectedNamespace:    "test-ns",
		},
		{
			name:                 "Just name",
			defaultTargetNetwork: "custom-net",
			namespace:            "test-ns",
			expectedType:         "multus",
			expectedName:         "custom-net",
			expectedNamespace:    "test-ns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDefaultNetwork(tt.defaultTargetNetwork, tt.namespace)

			if result.Type != tt.expectedType {
				t.Errorf("Type: got %s, want %s", result.Type, tt.expectedType)
			}
			if result.Name != tt.expectedName {
				t.Errorf("Name: got %s, want %s", result.Name, tt.expectedName)
			}
			if result.Namespace != tt.expectedNamespace {
				t.Errorf("Namespace: got %s, want %s", result.Namespace, tt.expectedNamespace)
			}
		})
	}
}

func TestOpenShiftNetworkMapper_ImplementsInterface(t *testing.T) {
	var _ networkmapper.NetworkMapper = &OpenShiftNetworkMapper{}
}
