package network

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	forkliftv1beta1 "github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1"
	"github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1/provider"
	"github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1/ref"
	"github.com/yaacov/kubectl-mtv/pkg/client"
)

// CreateDefaultNetworkMapOptions encapsulates the parameters for creating a default network map.
type CreateDefaultNetworkMapOptions struct {
	Name                 string
	Namespace            string
	SourceProvider       string
	TargetProvider       string
	ConfigFlags          *genericclioptions.ConfigFlags
	InventoryURL         string
	PlanVMNames          []string
	DefaultTargetNetwork string
}

// CreateDefaultNetworkMap creates a default network map.
// Returns the name of the created map and any error that occurred.
func CreateDefaultNetworkMap(opts CreateDefaultNetworkMapOptions) (string, error) {
	c, err := client.GetDynamicClient(opts.ConfigFlags)
	if err != nil {
		return "", fmt.Errorf("failed to get client: %v", err)
	}

	// Get source network IDs
	sourceNetworkIDs, err := GetSourceNetworkIDs(opts.ConfigFlags, opts.SourceProvider, opts.Namespace, opts.InventoryURL, opts.PlanVMNames)
	if err != nil {
		return "", err
	}

	var targetNetworks []NetworkInfo

	// If default target network is specified, use it
	if opts.DefaultTargetNetwork != "" {
		// Check if it's "pod" (case insensitive)
		if strings.ToLower(opts.DefaultTargetNetwork) == "pod" {
			// Use pod networking - no target networks needed, will be handled by CreateNetworkMapEntries
			targetNetworks = []NetworkInfo{}
		} else {
			// Use the specified network as the default target network
			targetNetworks = []NetworkInfo{
				{
					Name:      opts.DefaultTargetNetwork,
					Namespace: opts.Namespace, // Use the plan namespace as default
				},
			}
		}
	} else {
		// Get target networks from the target provider inventory
		targetNetworks, err = GetTargetNetworks(opts.ConfigFlags, opts.TargetProvider, opts.Namespace, opts.InventoryURL)
		if err != nil {
			return "", err
		}
	}

	// Create NetworkMap entries
	networkPairs := CreateNetworkMapEntries(sourceNetworkIDs, targetNetworks)

	// If no network pairs were created, create a dummy entry
	if len(networkPairs) == 0 {
		networkPairs = []forkliftv1beta1.NetworkPair{
			{
				Source: ref.Ref{
					Type: "pod", // Use "pod" type for dummy entry
				},
				Destination: forkliftv1beta1.DestinationNetwork{
					Type: "pod", // Use "pod" type for dummy entry
				},
			},
		}
	}

	// Create a new NetworkMap object
	networkMap := &forkliftv1beta1.NetworkMap{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: opts.Name + "-",
			Namespace:    opts.Namespace,
		},
		Spec: forkliftv1beta1.NetworkMapSpec{
			Provider: provider.Pair{
				Source: corev1.ObjectReference{
					Kind:       "Provider",
					APIVersion: forkliftv1beta1.SchemeGroupVersion.String(),
					Name:       opts.SourceProvider,
					Namespace:  opts.Namespace,
				},
				Destination: corev1.ObjectReference{
					Kind:       "Provider",
					APIVersion: forkliftv1beta1.SchemeGroupVersion.String(),
					Name:       opts.TargetProvider,
					Namespace:  opts.Namespace,
				},
			},
			Map: networkPairs,
		},
	}
	networkMap.Kind = "NetworkMap"
	networkMap.APIVersion = forkliftv1beta1.SchemeGroupVersion.String()

	// Convert NetworkMap object to Unstructured
	unstructuredNetworkMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(networkMap)
	if err != nil {
		return "", fmt.Errorf("failed to convert NetworkMap to Unstructured: %v", err)
	}
	networkMapUnstructured := &unstructured.Unstructured{Object: unstructuredNetworkMap}

	// Create the NetworkMap in the specified namespace
	createdMap, err := c.Resource(client.NetworkMapGVR).Namespace(opts.Namespace).Create(context.TODO(), networkMapUnstructured, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create network map: %v", err)
	}

	fmt.Printf("NetworkMap '%s' created in namespace '%s'\n", createdMap.GetName(), opts.Namespace)
	return createdMap.GetName(), nil
}
