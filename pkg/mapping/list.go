package mapping

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/output"
)

// getNetworkMappings retrieves all network mappings from the given namespace
func getNetworkMappings(dynamicClient dynamic.Interface, namespace string) ([]map[string]interface{}, error) {
	networks, err := dynamicClient.Resource(client.NetworkMapGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list network mappings: %v", err)
	}

	var items []map[string]interface{}
	for _, mapping := range networks.Items {
		item := createMappingItem(mapping, "NetworkMap")
		items = append(items, item)
	}

	return items, nil
}

// getStorageMappings retrieves all storage mappings from the given namespace
func getStorageMappings(dynamicClient dynamic.Interface, namespace string) ([]map[string]interface{}, error) {
	storage, err := dynamicClient.Resource(client.StorageMapGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list storage mappings: %v", err)
	}

	var items []map[string]interface{}
	for _, mapping := range storage.Items {
		item := createMappingItem(mapping, "StorageMap")
		items = append(items, item)
	}

	return items, nil
}

// getAllMappings retrieves all mappings (network and storage) from the given namespace
func getAllMappings(dynamicClient dynamic.Interface, namespace string) ([]map[string]interface{}, error) {
	var allItems []map[string]interface{}

	networkItems, err := getNetworkMappings(dynamicClient, namespace)
	if err != nil {
		return nil, err
	}
	allItems = append(allItems, networkItems...)

	storageItems, err := getStorageMappings(dynamicClient, namespace)
	if err != nil {
		return nil, err
	}
	allItems = append(allItems, storageItems...)

	return allItems, nil
}

// listMappings lists network and storage mappings
func listMappings(configFlags *genericclioptions.ConfigFlags, mappingType, namespace, outputFormat string) error {
	dynamicClient, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Format validation
	outputFormat = strings.ToLower(outputFormat)
	if outputFormat != "table" && outputFormat != "json" && outputFormat != "yaml" {
		return fmt.Errorf("unsupported output format: %s. Supported formats: table, json, yaml", outputFormat)
	}

	var allItems []map[string]interface{}

	// Get mappings based on the requested type
	switch mappingType {
	case "network":
		allItems, err = getNetworkMappings(dynamicClient, namespace)
	case "storage":
		allItems, err = getStorageMappings(dynamicClient, namespace)
	case "all":
		allItems, err = getAllMappings(dynamicClient, namespace)
	default:
		return fmt.Errorf("unsupported mapping type: %s. Supported types: network, storage, all", mappingType)
	}

	if err != nil {
		return err
	}

	// Handle output based on format
	if outputFormat == "json" {
		jsonPrinter := output.NewJSONPrinter().
			WithPrettyPrint(true).
			AddItems(allItems)

		if len(allItems) == 0 {
			return jsonPrinter.PrintEmpty("No mappings found in namespace " + namespace)
		}
		return jsonPrinter.Print()
	} else if outputFormat == "yaml" {
		yamlPrinter := output.NewYAMLPrinter().
			AddItems(allItems)

		if len(allItems) == 0 {
			return yamlPrinter.PrintEmpty("No mappings found in namespace " + namespace)
		}
		return yamlPrinter.Print()
	} else {
		// Table output (default)
		tablePrinter := output.NewTablePrinter().WithHeaders(
			output.Header{DisplayName: "NAME", JSONPath: "name"},
			output.Header{DisplayName: "TYPE", JSONPath: "type"},
			output.Header{DisplayName: "SOURCE", JSONPath: "source"},
			output.Header{DisplayName: "TARGET", JSONPath: "target"},
			output.Header{DisplayName: "OWNER", JSONPath: "owner"},
			output.Header{DisplayName: "CREATED", JSONPath: "created"},
		).AddItems(allItems)

		if len(allItems) == 0 {
			return tablePrinter.PrintEmpty("No mappings found in namespace " + namespace)
		}
		return tablePrinter.Print()
	}
}
