package inventory

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/output"
	querypkg "github.com/yaacov/kubectl-mtv/pkg/query"
	"github.com/yaacov/kubectl-mtv/pkg/watch"
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

	// Check provider type
	providerType, _, err := unstructured.NestedString(provider.Object, "spec", "type")
	if err != nil {
		return err
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

	// Select appropriate subPath based on provider type
	var subPath string
	switch providerType {
	case "openshift":
		subPath = "networkattachmentdefinitions?detail=4"
	default:
		subPath = "networks?detail=4"
	}

	// Fetch network inventory from the provider
	data, err := client.FetchProviderInventory(kubeConfigFlags, inventoryURL, provider, subPath)
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
	if outputFormat != "table" && outputFormat != "json" {
		return fmt.Errorf("unsupported output format: %s. Supported formats: table, json", outputFormat)
	}

	// Handle different output formats
	if outputFormat == "json" {
		// Use JSON printer
		jsonPrinter := output.NewJSONPrinter().
			WithPrettyPrint(true).
			AddItems(networks)

		if len(networks) == 0 {
			return jsonPrinter.PrintEmpty(fmt.Sprintf("No networks found for provider %s", providerName))
		}
		return jsonPrinter.Print()
	} else {
		var tablePrinter *output.TablePrinter

		// Check if we should use custom headers from SELECT clause
		if queryOpts.HasSelect {
			headers := make([]output.Header, 0, len(queryOpts.Select))
			for _, sel := range queryOpts.Select {
				headers = append(headers, output.Header{
					DisplayName: sel.Alias,
					JSONPath:    sel.Alias,
				})
			}
			tablePrinter = output.NewTablePrinter().
				WithHeaders(headers...).
				WithSelectOptions(queryOpts.Select)
		} else {
			// Use the predefined default headers based on provider type
			tablePrinter = output.NewTablePrinter().WithHeaders(defaultHeaders...)
		}

		tablePrinter.AddItems(networks)

		if len(networks) == 0 {
			return tablePrinter.PrintEmpty(fmt.Sprintf("No networks found for provider %s", providerName))
		}
		return tablePrinter.Print()
	}
}
