package inventory

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/util/output"
	querypkg "github.com/yaacov/kubectl-mtv/pkg/util/query"
	"github.com/yaacov/kubectl-mtv/pkg/util/watch"
)

// countNetworkHosts calculates the number of hosts connected to a network
func countNetworkHosts(network map[string]interface{}) int {
	hosts, exists := network["host"]
	if !exists {
		return 0
	}

	hostsArray, ok := hosts.([]interface{})
	if !ok {
		return 0
	}

	return len(hostsArray)
}

// ListNetworks queries the provider's network inventory and displays the results
func ListNetworks(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, watchMode bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return listNetworksOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query)
		}, 10*time.Second)
	}

	return listNetworksOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query)
}

func listNetworksOnce(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string) error {
	// Get the provider object
	provider, err := GetProviderByName(kubeConfigFlags, providerName, namespace)
	if err != nil {
		return err
	}

	// Create a new provider client
	providerClient := NewProviderClient(kubeConfigFlags, provider, inventoryURL)

	// Get provider type to determine resource path and headers
	providerType, err := providerClient.GetProviderType()
	if err != nil {
		return fmt.Errorf("failed to get provider type: %v", err)
	}

	// Define default headers based on provider type
	var defaultHeaders []output.Header
	switch providerType {
	case "openshift":
		defaultHeaders = []output.Header{
			{DisplayName: "NAME", JSONPath: "name"},
			{DisplayName: "NAMESPACE", JSONPath: "namespace"},
			{DisplayName: "ID", JSONPath: "id"},
			{DisplayName: "CREATED", JSONPath: "object.metadata.creationTimestamp"},
		}
	default:
		defaultHeaders = []output.Header{
			{DisplayName: "NAME", JSONPath: "name"},
			{DisplayName: "ID", JSONPath: "id"},
			{DisplayName: "VARIANT", JSONPath: "variant"},
			{DisplayName: "HOSTS", JSONPath: "hostCount"},
			{DisplayName: "VLAN", JSONPath: "vlanId"},
			{DisplayName: "REVISION", JSONPath: "revision"},
		}
	}

	// Fetch network inventory from the provider based on provider type
	var data interface{}
	switch providerType {
	case "openshift":
		// For OpenShift, get network attachment definitions
		data, err = providerClient.GetResourceCollection("networkattachmentdefinitions", 4)
	default:
		// For other providers, get networks
		data, err = providerClient.GetNetworks(4)
	}
	if err != nil {
		return fmt.Errorf("failed to fetch network inventory: %v", err)
	}

	// Verify data is an array
	dataArray, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("unexpected data format: expected array for network inventory")
	}

	// Convert to expected format
	networks := make([]map[string]interface{}, 0, len(dataArray))
	for _, item := range dataArray {
		if network, ok := item.(map[string]interface{}); ok {
			// Add provider name to each network
			network["provider"] = providerName

			// Add host count
			network["hostCount"] = countNetworkHosts(network)

			networks = append(networks, network)
		}
	}

	// Parse and apply query options
	queryOpts, err := querypkg.ParseQueryString(query)
	if err != nil {
		return fmt.Errorf("invalid query string: %v", err)
	}

	// Apply query options (sorting, filtering, limiting)
	networks, err = querypkg.ApplyQuery(networks, queryOpts)
	if err != nil {
		return fmt.Errorf("error applying query: %v", err)
	}

	// Format validation
	outputFormat = strings.ToLower(outputFormat)
	if outputFormat != "table" && outputFormat != "json" && outputFormat != "yaml" {
		return fmt.Errorf("unsupported output format: %s. Supported formats: table, json, yaml", outputFormat)
	}

	// Handle different output formats
	emptyMessage := fmt.Sprintf("No networks found for provider %s", providerName)
	switch outputFormat {
	case "json":
		return output.PrintJSONWithEmpty(networks, emptyMessage)
	case "yaml":
		return output.PrintYAMLWithEmpty(networks, emptyMessage)
	default:
		return output.PrintTableWithQuery(networks, defaultHeaders, queryOpts, emptyMessage)
	}
}
