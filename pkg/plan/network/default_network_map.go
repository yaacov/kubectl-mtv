package network

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	forkliftv1beta1 "github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1"
	"github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1/provider"
	"github.com/yaacov/kubectl-mtv/pkg/client"
)

// CreateDefaultNetworkMapOptions encapsulates the parameters for creating a default network map.
type CreateDefaultNetworkMapOptions struct {
	Name           string
	Namespace      string
	SourceProvider string
	TargetProvider string
	ConfigFlags    *genericclioptions.ConfigFlags
	InventoryURL   string
	PlanVMNames    []string
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

	// Get target networks
	targetNetworks, err := GetTargetNetworks(opts.ConfigFlags, opts.TargetProvider, opts.Namespace, opts.InventoryURL)
	if err != nil {
		return "", err
	}

	// Create NetworkMap entries
	networkPairs := CreateNetworkMapEntries(sourceNetworkIDs, targetNetworks)

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
