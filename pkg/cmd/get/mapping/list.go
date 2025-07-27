package mapping

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"

	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/output"
)

// extractProviderName gets a provider name from the mapping spec
func extractProviderName(mapping unstructured.Unstructured, providerType string) string {
	provider, found, _ := unstructured.NestedMap(mapping.Object, "spec", "provider", providerType)
	if !found || provider == nil {
		return ""
	}

	if name, ok := provider["name"].(string); ok {
		return name
	}
	return ""
}

// createMappingItem creates a standardized mapping item for output
func createMappingItem(mapping unstructured.Unstructured, mappingType string) map[string]interface{} {
	item := map[string]interface{}{
		"name":      mapping.GetName(),
		"namespace": mapping.GetNamespace(),
		"type":      mappingType,
		"source":    extractProviderName(mapping, "source"),
		"target":    extractProviderName(mapping, "destination"),
		"created":   mapping.GetCreationTimestamp().Format("2006-01-02 15:04:05"),
		"object":    mapping.Object, // Include the original object
	}

	// Add owner information if available
	if len(mapping.GetOwnerReferences()) > 0 {
		ownerRef := mapping.GetOwnerReferences()[0]
		item["owner"] = ownerRef.Name
		item["ownerKind"] = ownerRef.Kind
	}

	return item
}

// List lists network and storage mappings
func List(configFlags *genericclioptions.ConfigFlags, mappingType, namespace, outputFormat string, mappingName string) error {
	return listMappings(configFlags, mappingType, namespace, outputFormat, mappingName)
}

// getNetworkMappings retrieves all network mappings from the given namespace
func getNetworkMappings(dynamicClient dynamic.Interface, namespace string) ([]map[string]interface{}, error) {
	var networks *unstructured.UnstructuredList
	var err error

	if namespace != "" {
		networks, err = dynamicClient.Resource(client.NetworkMapGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	} else {
		networks, err = dynamicClient.Resource(client.NetworkMapGVR).List(context.TODO(), metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list network mappings: %v", err)
	}

	var items []map[string]interface{}
	for _, mappingItem := range networks.Items {
		item := createMappingItem(mappingItem, "NetworkMap")
		items = append(items, item)
	}

	return items, nil
}

// getStorageMappings retrieves all storage mappings from the given namespace
func getStorageMappings(dynamicClient dynamic.Interface, namespace string) ([]map[string]interface{}, error) {
	var storage *unstructured.UnstructuredList
	var err error

	if namespace != "" {
		storage, err = dynamicClient.Resource(client.StorageMapGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	} else {
		storage, err = dynamicClient.Resource(client.StorageMapGVR).List(context.TODO(), metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list storage mappings: %v", err)
	}

	var items []map[string]interface{}
	for _, mappingItem := range storage.Items {
		item := createMappingItem(mappingItem, "StorageMap")
		items = append(items, item)
	}

	return items, nil
}

// getSpecificNetworkMapping retrieves a specific network mapping by name
func getSpecificNetworkMapping(dynamicClient dynamic.Interface, namespace, mappingName string) ([]map[string]interface{}, error) {
	if namespace != "" {
		// If namespace is specified, get the specific resource
		networkMap, err := dynamicClient.Resource(client.NetworkMapGVR).Namespace(namespace).Get(context.TODO(), mappingName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		item := createMappingItem(*networkMap, "NetworkMap")
		return []map[string]interface{}{item}, nil
	} else {
		// If no namespace specified, list all and filter by name
		networks, err := dynamicClient.Resource(client.NetworkMapGVR).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list network mappings: %v", err)
		}

		var items []map[string]interface{}
		for _, mapping := range networks.Items {
			if mapping.GetName() == mappingName {
				item := createMappingItem(mapping, "NetworkMap")
				items = append(items, item)
			}
		}

		if len(items) == 0 {
			return nil, fmt.Errorf("network mapping '%s' not found", mappingName)
		}

		return items, nil
	}
}

// getSpecificStorageMapping retrieves a specific storage mapping by name
func getSpecificStorageMapping(dynamicClient dynamic.Interface, namespace, mappingName string) ([]map[string]interface{}, error) {
	if namespace != "" {
		// If namespace is specified, get the specific resource
		storageMap, err := dynamicClient.Resource(client.StorageMapGVR).Namespace(namespace).Get(context.TODO(), mappingName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		item := createMappingItem(*storageMap, "StorageMap")
		return []map[string]interface{}{item}, nil
	} else {
		// If no namespace specified, list all and filter by name
		storage, err := dynamicClient.Resource(client.StorageMapGVR).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list storage mappings: %v", err)
		}

		var items []map[string]interface{}
		for _, mapping := range storage.Items {
			if mapping.GetName() == mappingName {
				item := createMappingItem(mapping, "StorageMap")
				items = append(items, item)
			}
		}

		if len(items) == 0 {
			return nil, fmt.Errorf("storage mapping '%s' not found", mappingName)
		}

		return items, nil
	}
}

// getSpecificAllMappings retrieves a specific mapping by name from both network and storage mappings
func getSpecificAllMappings(dynamicClient dynamic.Interface, namespace, mappingName string) ([]map[string]interface{}, error) {
	var allItems []map[string]interface{}

	// Try both types if no specific type is specified
	// First try network mapping
	networkItems, err := getSpecificNetworkMapping(dynamicClient, namespace, mappingName)
	if err == nil && len(networkItems) > 0 {
		allItems = append(allItems, networkItems...)
	}

	// Then try storage mapping
	storageItems, err := getSpecificStorageMapping(dynamicClient, namespace, mappingName)
	if err == nil && len(storageItems) > 0 {
		allItems = append(allItems, storageItems...)
	}

	// If no mappings found, return error
	if len(allItems) == 0 {
		return nil, fmt.Errorf("mapping '%s' not found", mappingName)
	}

	return allItems, nil
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
func listMappings(configFlags *genericclioptions.ConfigFlags, mappingType, namespace, outputFormat string, mappingName string) error {
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

	// If mappingName is specified, get that specific mapping
	if mappingName != "" {
		// Get specific mapping based on type
		switch mappingType {
		case "", "all":
			allItems, err = getSpecificAllMappings(dynamicClient, namespace, mappingName)
		case "network":
			allItems, err = getSpecificNetworkMapping(dynamicClient, namespace, mappingName)
		case "storage":
			allItems, err = getSpecificStorageMapping(dynamicClient, namespace, mappingName)
		default:
			return fmt.Errorf("unsupported mapping type: %s. Supported types: network, storage, all", mappingType)
		}
	} else {
		// Get mappings based on the requested type
		switch mappingType {
		case "network":
			allItems, err = getNetworkMappings(dynamicClient, namespace)
		case "storage":
			allItems, err = getStorageMappings(dynamicClient, namespace)
		case "", "all":
			allItems, err = getAllMappings(dynamicClient, namespace)
		default:
			return fmt.Errorf("unsupported mapping type: %s. Supported types: network, storage, all", mappingType)
		}
	}

	// Handle error if no items found
	if err != nil {
		return err
	}

	// Handle output based on format
	switch outputFormat {
	case "json":
		jsonPrinter := output.NewJSONPrinter().
			WithPrettyPrint(true).
			AddItems(allItems)

		if len(allItems) == 0 {
			return jsonPrinter.PrintEmpty("No mappings found in namespace " + namespace)
		}
		return jsonPrinter.Print()
	case "yaml":
		yamlPrinter := output.NewYAMLPrinter().
			AddItems(allItems)

		if len(allItems) == 0 {
			return yamlPrinter.PrintEmpty("No mappings found in namespace " + namespace)
		}
		return yamlPrinter.Print()
	default:
		// Table output (default)
		var headers []output.Header

		// Add NAME column first
		headers = append(headers, output.Header{DisplayName: "NAME", JSONPath: "name"})

		// Add NAMESPACE column after NAME when listing across all namespaces
		if namespace == "" {
			headers = append(headers, output.Header{DisplayName: "NAMESPACE", JSONPath: "namespace"})
		}

		// Add remaining columns
		headers = append(headers,
			output.Header{DisplayName: "TYPE", JSONPath: "type"},
			output.Header{DisplayName: "SOURCE", JSONPath: "source"},
			output.Header{DisplayName: "TARGET", JSONPath: "target"},
			output.Header{DisplayName: "OWNER", JSONPath: "owner"},
			output.Header{DisplayName: "CREATED", JSONPath: "created"},
		)

		tablePrinter := output.NewTablePrinter().WithHeaders(headers...).AddItems(allItems)

		if len(allItems) == 0 {
			return tablePrinter.PrintEmpty("No mappings found in namespace " + namespace)
		}
		return tablePrinter.Print()
	}
}
