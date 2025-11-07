package vsphere

import (
	"testing"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"

	networkmapper "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/mapper"
)

func TestVSphereNetworkMapper_CreateNetworkPairs_DefaultBehavior(t *testing.T) {
	tests := []struct {
		name                 string
		sourceNetworks       []ref.Ref
		targetNetworks       []forkliftv1beta1.DestinationNetwork
		defaultTargetNetwork string
		namespace            string
		expectedPairs        int
		expectedFirstTarget  string
		expectedFirstType    string
		expectedOthersType   string
	}{
		{
			name: "First source maps to first multus, others to ignored",
			sourceNetworks: []ref.Ref{
				{Name: "VM Network", ID: "net-1"},
				{Name: "Management Network", ID: "net-2"},
				{Name: "Storage Network", ID: "net-3"},
			},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "multus-net", Namespace: "default"},
				{Type: "multus", Name: "another-net", Namespace: "default"},
			},
			namespace:           "default",
			expectedPairs:       3,
			expectedFirstTarget: "multus-net",
			expectedFirstType:   "multus",
			expectedOthersType:  "ignored",
		},
		{
			name: "User-defined default network",
			sourceNetworks: []ref.Ref{
				{Name: "VM Network", ID: "net-1"},
				{Name: "Management Network", ID: "net-2"},
			},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "multus", Name: "auto-selected", Namespace: "default"},
			},
			defaultTargetNetwork: "custom-ns/custom-net",
			namespace:            "default",
			expectedPairs:        2,
			expectedFirstTarget:  "custom-net",
			expectedFirstType:    "multus",
			expectedOthersType:   "ignored",
		},
		{
			name: "Fallback to pod networking",
			sourceNetworks: []ref.Ref{
				{Name: "VM Network", ID: "net-1"},
			},
			targetNetworks: []forkliftv1beta1.DestinationNetwork{
				{Type: "pod"},
			},
			namespace:           "default",
			expectedPairs:       1,
			expectedFirstTarget: "",
			expectedFirstType:   "pod",
		},
		{
			name:                "Empty sources",
			sourceNetworks:      []ref.Ref{},
			targetNetworks:      []forkliftv1beta1.DestinationNetwork{{Type: "multus", Name: "any"}},
			namespace:           "default",
			expectedPairs:       0,
			expectedFirstTarget: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			networkMapper := NewVSphereNetworkMapper()

			opts := networkmapper.NetworkMappingOptions{
				DefaultTargetNetwork: tt.defaultTargetNetwork,
				Namespace:            tt.namespace,
			}

			pairs, err := networkMapper.CreateNetworkPairs(tt.sourceNetworks, tt.targetNetworks, opts)
			if err != nil {
				t.Fatalf("CreateNetworkPairs() error = %v", err)
			}

			if len(pairs) != tt.expectedPairs {
				t.Errorf("CreateNetworkPairs() got %d pairs, want %d", len(pairs), tt.expectedPairs)
			}

			// Verify first source mapping
			if len(pairs) > 0 {
				firstPair := pairs[0]
				if firstPair.Destination.Name != tt.expectedFirstTarget {
					t.Errorf("First pair target name: got %s, want %s",
						firstPair.Destination.Name, tt.expectedFirstTarget)
				}
				if firstPair.Destination.Type != tt.expectedFirstType {
					t.Errorf("First pair target type: got %s, want %s",
						firstPair.Destination.Type, tt.expectedFirstType)
				}
			}

			// Verify all other sources map to "ignored" (generic behavior)
			if tt.expectedOthersType != "" {
				for i := 1; i < len(pairs); i++ {
					if pairs[i].Destination.Type != tt.expectedOthersType {
						t.Errorf("Pair %d type: got %s, want %s",
							i, pairs[i].Destination.Type, tt.expectedOthersType)
					}
				}
			}

			// Verify source names are preserved
			for i, pair := range pairs {
				if i < len(tt.sourceNetworks) {
					if pair.Source.Name != tt.sourceNetworks[i].Name {
						t.Errorf("Pair %d: source name %s != expected %s",
							i, pair.Source.Name, tt.sourceNetworks[i].Name)
					}
				}
			}
		})
	}
}

func TestVSphereNetworkMapper_NoSameNameMatching(t *testing.T) {
	// This test verifies that vSphere mapper does NOT use same-name matching
	// even when source and target names match
	networkMapper := NewVSphereNetworkMapper()

	sourceNetworks := []ref.Ref{
		{Name: "identical-name", ID: "net-1"},
		{Name: "another-name", ID: "net-2"},
	}

	targetNetworks := []forkliftv1beta1.DestinationNetwork{
		{Type: "multus", Name: "identical-name", Namespace: "default"}, // Same name as source
		{Type: "multus", Name: "different-name", Namespace: "default"},
	}

	opts := networkmapper.NetworkMappingOptions{
		Namespace:          "default",
		SourceProviderType: "vsphere",
		TargetProviderType: "openshift",
	}

	pairs, err := networkMapper.CreateNetworkPairs(sourceNetworks, targetNetworks, opts)
	if err != nil {
		t.Fatalf("CreateNetworkPairs() error = %v", err)
	}

	// First source should map to first target, second source to "ignored" - NOT same-name matching
	if len(pairs) != 2 {
		t.Fatalf("Expected 2 pairs, got %d", len(pairs))
	}

	// First source should map to first target (not same-name matching)
	if pairs[0].Destination.Name != "identical-name" {
		t.Errorf("First pair: got target %s, want identical-name (first available)",
			pairs[0].Destination.Name)
	}
	if pairs[0].Destination.Type != "multus" {
		t.Errorf("First pair: got type %s, want multus", pairs[0].Destination.Type)
	}

	// Second source should map to "ignored"
	if pairs[1].Destination.Type != "ignored" {
		t.Errorf("Second pair: got type %s, want ignored", pairs[1].Destination.Type)
	}
}

func TestVSphereNetworkMapper_ImplementsInterface(t *testing.T) {
	var _ networkmapper.NetworkMapper = &VSphereNetworkMapper{}
}
