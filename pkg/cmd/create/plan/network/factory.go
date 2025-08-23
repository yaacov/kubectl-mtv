package network

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/provider"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/fetchers"
	openshiftFetcher "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/fetchers/openshift"
	openstackFetcher "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/fetchers/openstack"
	ovaFetcher "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/fetchers/ova"
	ovirtFetcher "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/fetchers/ovirt"
	vsphereFetcher "github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/fetchers/vsphere"
	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/network/mapper"
	"github.com/yaacov/kubectl-mtv/pkg/cmd/get/inventory"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
)

// NetworkMapperInterface defines the interface for network mapping operations
type NetworkMapperInterface interface {
	// GetSourceNetworks extracts network information from the source provider for the specified VMs
	GetSourceNetworks(configFlags *genericclioptions.ConfigFlags, providerName, namespace, inventoryURL string, planVMNames []string) ([]ref.Ref, error)

	// GetTargetNetworks extracts available network information from the target provider
	GetTargetNetworks(configFlags *genericclioptions.ConfigFlags, providerName, namespace, inventoryURL string) ([]forkliftv1beta1.DestinationNetwork, error)

	// CreateNetworkPairs creates network mapping pairs based on source networks, target networks, and optional default network
	CreateNetworkPairs(sourceNetworks []ref.Ref, targetNetworks []forkliftv1beta1.DestinationNetwork, defaultTargetNetwork string, namespace string) ([]forkliftv1beta1.NetworkPair, error)
}

// NetworkMapperOptions contains common options for network mapping
type NetworkMapperOptions struct {
	Name                    string
	Namespace               string
	SourceProvider          string
	SourceProviderNamespace string
	TargetProvider          string
	TargetProviderNamespace string
	ConfigFlags             *genericclioptions.ConfigFlags
	InventoryURL            string
	PlanVMNames             []string
	DefaultTargetNetwork    string
}

// GetSourceNetworkFetcher returns the appropriate source network fetcher based on provider type
func GetSourceNetworkFetcher(configFlags *genericclioptions.ConfigFlags, providerName, namespace string) (fetchers.SourceNetworkFetcher, error) {
	// Get the provider object to determine its type
	provider, err := inventory.GetProviderByName(configFlags, providerName, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %v", err)
	}

	// Create a provider client to get the provider type
	providerClient := inventory.NewProviderClient(configFlags, provider, "")
	providerType, err := providerClient.GetProviderType()
	if err != nil {
		return nil, fmt.Errorf("failed to get provider type: %v", err)
	}

	klog.V(4).Infof("DEBUG: GetSourceNetworkFetcher - Provider: %s, Type: %s", providerName, providerType)

	// Return the appropriate fetcher based on provider type
	switch providerType {
	case "openstack":
		klog.V(4).Infof("DEBUG: Using OpenStack source network fetcher for %s", providerName)
		return openstackFetcher.NewOpenStackNetworkFetcher(), nil
	case "vsphere":
		klog.V(4).Infof("DEBUG: Using VSphere source network fetcher for %s", providerName)
		return vsphereFetcher.NewVSphereNetworkFetcher(), nil
	case "openshift":
		klog.V(4).Infof("DEBUG: Using OpenShift source network fetcher for %s", providerName)
		return openshiftFetcher.NewOpenShiftNetworkFetcher(), nil
	case "ova":
		klog.V(4).Infof("DEBUG: Using OVA source network fetcher for %s", providerName)
		return ovaFetcher.NewOVANetworkFetcher(), nil
	case "ovirt":
		klog.V(4).Infof("DEBUG: Using oVirt source network fetcher for %s", providerName)
		return ovirtFetcher.NewOvirtNetworkFetcher(), nil
	default:
		return nil, fmt.Errorf("unsupported source provider type: %s", providerType)
	}
}

// GetTargetNetworkFetcher returns the appropriate target network fetcher based on provider type
func GetTargetNetworkFetcher(configFlags *genericclioptions.ConfigFlags, providerName, namespace string) (fetchers.TargetNetworkFetcher, error) {
	// Get the provider object to determine its type
	provider, err := inventory.GetProviderByName(configFlags, providerName, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %v", err)
	}

	// Create a provider client to get the provider type
	providerClient := inventory.NewProviderClient(configFlags, provider, "")
	providerType, err := providerClient.GetProviderType()
	if err != nil {
		return nil, fmt.Errorf("failed to get provider type: %v", err)
	}

	klog.V(4).Infof("DEBUG: GetTargetNetworkFetcher - Provider: %s, Type: %s", providerName, providerType)

	// Return the appropriate fetcher based on provider type
	switch providerType {
	case "openstack":
		klog.V(4).Infof("DEBUG: Using OpenStack target network fetcher for %s", providerName)
		return openstackFetcher.NewOpenStackNetworkFetcher(), nil
	case "vsphere":
		klog.V(4).Infof("DEBUG: Using VSphere target network fetcher for %s", providerName)
		return vsphereFetcher.NewVSphereNetworkFetcher(), nil
	case "openshift":
		klog.V(4).Infof("DEBUG: Using OpenShift target network fetcher for %s", providerName)
		return openshiftFetcher.NewOpenShiftNetworkFetcher(), nil
	case "ova":
		klog.V(4).Infof("DEBUG: Using OVA target network fetcher for %s", providerName)
		return ovaFetcher.NewOVANetworkFetcher(), nil
	case "ovirt":
		klog.V(4).Infof("DEBUG: Using oVirt target network fetcher for %s", providerName)
		return ovirtFetcher.NewOvirtNetworkFetcher(), nil
	default:
		return nil, fmt.Errorf("unsupported target provider type: %s", providerType)
	}
}

// CreateNetworkMap creates a network map using the new fetcher-based architecture
func CreateNetworkMap(opts NetworkMapperOptions) (string, error) {
	klog.V(4).Infof("DEBUG: Creating network map - Source: %s, Target: %s, DefaultTargetNetwork: '%s'",
		opts.SourceProvider, opts.TargetProvider, opts.DefaultTargetNetwork)

	// Get source network fetcher using the provider's namespace
	sourceProviderNamespace := opts.SourceProviderNamespace
	if sourceProviderNamespace == "" {
		sourceProviderNamespace = opts.Namespace
	}
	sourceFetcher, err := GetSourceNetworkFetcher(opts.ConfigFlags, opts.SourceProvider, sourceProviderNamespace)
	if err != nil {
		return "", fmt.Errorf("failed to get source network fetcher: %v", err)
	}
	klog.V(4).Infof("DEBUG: Source fetcher created for provider: %s", opts.SourceProvider)

	// Get target network fetcher using the provider's namespace
	targetProviderNamespace := opts.TargetProviderNamespace
	if targetProviderNamespace == "" {
		targetProviderNamespace = opts.Namespace
	}
	targetFetcher, err := GetTargetNetworkFetcher(opts.ConfigFlags, opts.TargetProvider, targetProviderNamespace)
	if err != nil {
		return "", fmt.Errorf("failed to get target network fetcher: %v", err)
	}
	klog.V(4).Infof("DEBUG: Target fetcher created for provider: %s", opts.TargetProvider)

	// Fetch source networks
	sourceNetworks, err := sourceFetcher.FetchSourceNetworks(opts.ConfigFlags, opts.SourceProvider, sourceProviderNamespace, opts.InventoryURL, opts.PlanVMNames)
	if err != nil {
		return "", fmt.Errorf("failed to fetch source networks: %v", err)
	}
	klog.V(4).Infof("DEBUG: Fetched %d source networks", len(sourceNetworks))

	// Fetch target networks
	var targetNetworks []forkliftv1beta1.DestinationNetwork
	if opts.DefaultTargetNetwork == "" || (opts.DefaultTargetNetwork != "default" && opts.DefaultTargetNetwork != "") {
		klog.V(4).Infof("DEBUG: Fetching target networks from target provider: %s", opts.TargetProvider)
		targetNetworks, err = targetFetcher.FetchTargetNetworks(opts.ConfigFlags, opts.TargetProvider, targetProviderNamespace, opts.InventoryURL)
		if err != nil {
			return "", fmt.Errorf("failed to fetch target networks: %v", err)
		}
		klog.V(4).Infof("DEBUG: Fetched %d target networks", len(targetNetworks))
	} else {
		klog.V(4).Infof("DEBUG: Skipping target network fetch due to DefaultTargetNetwork='%s'", opts.DefaultTargetNetwork)
	}

	// Create network pairs using generic mapping logic
	mappingOpts := mapper.NetworkMappingOptions{
		DefaultTargetNetwork: opts.DefaultTargetNetwork,
		Namespace:            opts.Namespace,
	}
	networkPairs, err := mapper.CreateNetworkPairs(sourceNetworks, targetNetworks, mappingOpts)
	if err != nil {
		return "", fmt.Errorf("failed to create network pairs: %v", err)
	}

	// Create the network map using the existing infrastructure
	return createNetworkMap(opts, networkPairs)
}

// createNetworkMap helper function to create the actual network map resource
func createNetworkMap(opts NetworkMapperOptions, networkPairs []forkliftv1beta1.NetworkPair) (string, error) {
	// If no network pairs, create a dummy pair
	if len(networkPairs) == 0 {
		klog.V(4).Infof("DEBUG: No network pairs found, creating dummy pair")
		networkPairs = []forkliftv1beta1.NetworkPair{
			{
				Source: ref.Ref{
					Type: "pod", // Use "pod" type for dummy entry
				},
				Destination: forkliftv1beta1.DestinationNetwork{
					Type: "pod", // Use pod networking as default
				},
			},
		}
	}

	// Create the network map name
	networkMapName := opts.Name + "-network-map"

	// Create NetworkMap object
	networkMap := &forkliftv1beta1.NetworkMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      networkMapName,
			Namespace: opts.Namespace,
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

	// Convert to Unstructured
	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(networkMap)
	if err != nil {
		return "", fmt.Errorf("failed to convert NetworkMap to Unstructured: %v", err)
	}

	networkMapUnstructured := &unstructured.Unstructured{Object: unstructuredMap}

	// Create the network map
	c, err := client.GetDynamicClient(opts.ConfigFlags)
	if err != nil {
		return "", fmt.Errorf("failed to get client: %v", err)
	}

	_, err = c.Resource(client.NetworkMapGVR).Namespace(opts.Namespace).Create(context.TODO(), networkMapUnstructured, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create network map: %v", err)
	}

	klog.V(4).Infof("DEBUG: Created network map '%s' with %d network pairs", networkMapName, len(networkPairs))
	return networkMapName, nil
}
