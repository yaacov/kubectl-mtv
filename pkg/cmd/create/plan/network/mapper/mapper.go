package mapper

import (
	"strings"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"
	"k8s.io/klog/v2"
)

// NetworkMappingOptions contains options for network mapping
type NetworkMappingOptions struct {
	DefaultTargetNetwork string
	Namespace            string
}

// CreateNetworkPairs creates network mapping pairs using generic logic
func CreateNetworkPairs(sourceNetworks []ref.Ref, targetNetworks []forkliftv1beta1.DestinationNetwork, opts NetworkMappingOptions) ([]forkliftv1beta1.NetworkPair, error) {
	var networkPairs []forkliftv1beta1.NetworkPair

	klog.V(4).Infof("DEBUG: Creating network pairs - %d source networks, %d target networks", len(sourceNetworks), len(targetNetworks))
	klog.V(4).Infof("DEBUG: Default target network: '%s'", opts.DefaultTargetNetwork)

	// If a default target network is specified, use it for all source networks
	if opts.DefaultTargetNetwork != "" {
		defaultDestination := parseDefaultNetwork(opts.DefaultTargetNetwork, opts.Namespace)
		klog.V(4).Infof("DEBUG: Using explicit default network for all sources: %s/%s (%s)",
			defaultDestination.Namespace, defaultDestination.Name, defaultDestination.Type)

		for _, sourceNetwork := range sourceNetworks {
			networkPairs = append(networkPairs, forkliftv1beta1.NetworkPair{
				Source:      sourceNetwork,
				Destination: defaultDestination,
			})
		}
		return networkPairs, nil
	}

	// Auto-mapping: try to match source networks with target networks intelligently
	klog.V(4).Infof("DEBUG: Available target networks for auto-mapping:")
	for i, targetNetwork := range targetNetworks {
		klog.V(4).Infof("  [%d] %s/%s (type: %s)", i, targetNetwork.Namespace, targetNetwork.Name, targetNetwork.Type)
	}

	// Create name-based lookup for target networks
	targetNetworkMap := make(map[string]forkliftv1beta1.DestinationNetwork)
	for _, targetNetwork := range targetNetworks {
		// Use network name as key for matching
		targetNetworkMap[targetNetwork.Name] = targetNetwork
	}

	// Create pairs for each source network
	for _, sourceNetwork := range sourceNetworks {
		destination := findBestTargetNetwork(sourceNetwork, targetNetworks, targetNetworkMap)

		networkPairs = append(networkPairs, forkliftv1beta1.NetworkPair{
			Source:      sourceNetwork,
			Destination: destination,
		})
	}

	return networkPairs, nil
}

// parseDefaultNetwork parses a default network specification
func parseDefaultNetwork(defaultTargetNetwork, namespace string) forkliftv1beta1.DestinationNetwork {
	if defaultTargetNetwork == "default" {
		return forkliftv1beta1.DestinationNetwork{Type: "pod"}
	}

	// Handle "namespace/name" format for multus networks
	if parts := strings.Split(defaultTargetNetwork, "/"); len(parts) == 2 {
		return forkliftv1beta1.DestinationNetwork{
			Type:      "multus",
			Name:      parts[1],
			Namespace: parts[0],
		}
	}

	// Just a name, use the provided namespace
	return forkliftv1beta1.DestinationNetwork{
		Type:      "multus",
		Name:      defaultTargetNetwork,
		Namespace: namespace,
	}
}

// findBestTargetNetwork finds the best target network for a given source network
func findBestTargetNetwork(sourceNetwork ref.Ref, targetNetworks []forkliftv1beta1.DestinationNetwork, targetNetworkMap map[string]forkliftv1beta1.DestinationNetwork) forkliftv1beta1.DestinationNetwork {
	// Strategy 1: Try exact name match
	if targetNet, exists := targetNetworkMap[sourceNetwork.Name]; exists {
		klog.V(4).Infof("DEBUG: Found exact name match for %s -> %s/%s (%s)",
			sourceNetwork.Name, targetNet.Namespace, targetNet.Name, targetNet.Type)
		return targetNet
	}

	// Strategy 2: Use first available multus network (prefer multus over pod)
	for _, targetNetwork := range targetNetworks {
		if targetNetwork.Type == "multus" {
			klog.V(4).Infof("DEBUG: No exact match for %s, using first multus network: %s/%s",
				sourceNetwork.Name, targetNetwork.Namespace, targetNetwork.Name)
			return targetNetwork
		}
	}

	// Strategy 3: Use first available target network (any type)
	if len(targetNetworks) > 0 {
		target := targetNetworks[0]
		klog.V(4).Infof("DEBUG: No multus networks available for %s, using first target: %s/%s (%s)",
			sourceNetwork.Name, target.Namespace, target.Name, target.Type)
		return target
	}

	// Strategy 4: Fall back to default networking only if no targets available
	klog.V(4).Infof("DEBUG: No target networks available for %s, falling back to default networking", sourceNetwork.Name)
	return forkliftv1beta1.DestinationNetwork{Type: "pod"}
}
