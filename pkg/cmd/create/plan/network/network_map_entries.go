package network

import (
	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"
)

// CreateNetworkMapEntries creates NetworkMap entries based on source network IDs and target networks.
func CreateNetworkMapEntries(sourceNetworkIDs map[string]bool, targetNetworks []NetworkInfo) []forkliftv1beta1.NetworkPair {
	var networkPairs []forkliftv1beta1.NetworkPair
	i := 0
	for networkID := range sourceNetworkIDs {
		var destinationNetwork forkliftv1beta1.DestinationNetwork

		if i < len(targetNetworks) {
			destinationNetwork = forkliftv1beta1.DestinationNetwork{
				Name:      targetNetworks[i].Name,
				Namespace: targetNetworks[i].Namespace,
				Type:      "multus",
			}
		} else {
			// If no target network is available, use "Pod" for the first source network and "Ignore" for the rest
			if i == 0 {
				destinationNetwork = forkliftv1beta1.DestinationNetwork{
					Type: "pod",
				}
			} else {
				destinationNetwork = forkliftv1beta1.DestinationNetwork{
					Type: "ignored",
				}
			}
		}

		networkPairs = append(networkPairs, forkliftv1beta1.NetworkPair{
			Source: ref.Ref{
				ID: networkID,
			},
			Destination: destinationNetwork,
		})
		i++
	}

	return networkPairs
}
