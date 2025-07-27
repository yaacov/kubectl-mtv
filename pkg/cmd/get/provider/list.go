package provider

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/provider/providerutil"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/output"
)

// getProviders retrieves all providers from the given namespace
func getProviders(dynamicClient dynamic.Interface, namespace string) (*unstructured.UnstructuredList, error) {
	if namespace != "" {
		return dynamicClient.Resource(client.ProvidersGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	} else {
		return dynamicClient.Resource(client.ProvidersGVR).List(context.TODO(), metav1.ListOptions{})
	}
}

// getSpecificProvider retrieves a specific provider by name
func getSpecificProvider(dynamicClient dynamic.Interface, namespace, providerName string) (*unstructured.UnstructuredList, error) {
	if namespace != "" {
		// If namespace is specified, get the specific resource
		provider, err := dynamicClient.Resource(client.ProvidersGVR).Namespace(namespace).Get(context.TODO(), providerName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// Create a list with just this provider
		return &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{*provider},
		}, nil
	} else {
		// If no namespace specified, list all and filter by name
		providers, err := dynamicClient.Resource(client.ProvidersGVR).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list providers: %v", err)
		}

		var filteredItems []unstructured.Unstructured
		for _, provider := range providers.Items {
			if provider.GetName() == providerName {
				filteredItems = append(filteredItems, provider)
			}
		}

		if len(filteredItems) == 0 {
			return nil, fmt.Errorf("provider '%s' not found", providerName)
		}

		return &unstructured.UnstructuredList{
			Items: filteredItems,
		}, nil
	}
}

// List lists providers
func List(configFlags *genericclioptions.ConfigFlags, namespace string, baseURL string, outputFormat string, providerName string) error {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	var providers *unstructured.UnstructuredList
	if providerName != "" {
		// Get specific provider by name
		providers, err = getSpecificProvider(c, namespace, providerName)
		if err != nil {
			return fmt.Errorf("failed to get provider: %v", err)
		}
	} else {
		// Get all providers
		providers, err = getProviders(c, namespace)
		if err != nil {
			return fmt.Errorf("failed to list providers: %v", err)
		}
	}

	// Format validation
	outputFormat = strings.ToLower(outputFormat)
	if outputFormat != "table" && outputFormat != "json" && outputFormat != "yaml" {
		return fmt.Errorf("unsupported output format: %s. Supported formats: table, json, yaml", outputFormat)
	}

	// If baseURL is empty, try to discover it from an OpenShift Route
	if baseURL == "" {
		route, err := client.GetForkliftInventoryRoute(configFlags, namespace)
		if err == nil && route != nil {
			host, found, _ := unstructured.NestedString(route.Object, "spec", "host")
			if found && host != "" {
				baseURL = fmt.Sprintf("https://%s", host)
			}
		}
	}

	// Create printer items with condition statuses incorporated
	items := []map[string]interface{}{}
	for i := range providers.Items {
		provider := &providers.Items[i]

		// Extract condition statuses
		conditionStatuses := providerutil.ExtractProviderConditionStatuses(provider.Object)

		// Create a new printer item with needed fields
		item := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      provider.GetName(),
				"namespace": provider.GetNamespace(),
			},
			"spec":   provider.Object["spec"],
			"status": provider.Object["status"],
			"conditionStatuses": map[string]interface{}{
				"ConnectionStatus": conditionStatuses.ConnectionStatus,
				"ValidationStatus": conditionStatuses.ValidationStatus,
				"InventoryStatus":  conditionStatuses.InventoryStatus,
				"ReadyStatus":      conditionStatuses.ReadyStatus,
			},
			"object": provider.Object, // Include the original object
		}

		if baseURL != "" {
			inventory, err := client.FetchProviderInventory(configFlags, baseURL, provider, "")
			if err == nil && inventory != nil {
				if inventoryMap, ok := inventory.(map[string]interface{}); ok {
					item["vmCount"] = inventoryMap["vmCount"]
					item["hostCount"] = inventoryMap["hostCount"]
					item["datacenterCount"] = inventoryMap["datacenterCount"]
					item["clusterCount"] = inventoryMap["clusterCount"]
					item["networkCount"] = inventoryMap["networkCount"]
					item["datastoreCount"] = inventoryMap["datastoreCount"]
					item["storageClassCount"] = inventoryMap["storageClassCount"]
					item["product"] = inventoryMap["product"]
				}
			}
		}

		// Add the item to the list
		items = append(items, item)
	}

	// Handle different output formats
	switch outputFormat {
	case "json":
		// Use JSON printer
		jsonPrinter := output.NewJSONPrinter().
			WithPrettyPrint(true).
			AddItems(items)

		if len(providers.Items) == 0 {
			return jsonPrinter.PrintEmpty("No providers found in namespace " + namespace)
		}
		return jsonPrinter.Print()
	case "yaml":
		// Use YAML printer
		yamlPrinter := output.NewYAMLPrinter().
			AddItems(items)

		if len(providers.Items) == 0 {
			return yamlPrinter.PrintEmpty("No providers found in namespace " + namespace)
		}
		return yamlPrinter.Print()
	default:
		// Use Table printer (default)
		var headers []output.Header

		// Add NAME column first
		headers = append(headers, output.Header{DisplayName: "NAME", JSONPath: "metadata.name"})

		// Add NAMESPACE column after NAME when listing across all namespaces
		if namespace == "" {
			headers = append(headers, output.Header{DisplayName: "NAMESPACE", JSONPath: "metadata.namespace"})
		}

		// Add remaining columns
		headers = append(headers,
			output.Header{DisplayName: "TYPE", JSONPath: "spec.type"},
			output.Header{DisplayName: "URL", JSONPath: "spec.url"},
			output.Header{DisplayName: "STATUS", JSONPath: "status.phase"},
			output.Header{DisplayName: "CONNECTED", JSONPath: "conditionStatuses.ConnectionStatus"},
			output.Header{DisplayName: "INVENTORY", JSONPath: "conditionStatuses.InventoryStatus"},
			output.Header{DisplayName: "READY", JSONPath: "conditionStatuses.ReadyStatus"},
			output.Header{DisplayName: "VMS", JSONPath: "vmCount"},
			output.Header{DisplayName: "HOSTS", JSONPath: "hostCount"},
		)

		tablePrinter := output.NewTablePrinter().WithHeaders(headers...).AddItems(items)

		if len(providers.Items) == 0 {
			if err := tablePrinter.PrintEmpty("No providers found in namespace " + namespace); err != nil {
				return fmt.Errorf("error printing empty table: %v", err)
			}
		} else {
			if err := tablePrinter.Print(); err != nil {
				return fmt.Errorf("error printing table: %v", err)
			}
		}
	}

	return nil
}
