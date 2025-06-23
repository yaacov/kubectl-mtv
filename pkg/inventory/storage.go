package inventory

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/output"
	querypkg "github.com/yaacov/kubectl-mtv/pkg/query"
)

// ListStorage queries the provider's storage inventory and displays the results
func ListStorage(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string) error {
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
			{DisplayName: "ID", JSONPath: "id"},
			{DisplayName: "DEFAULT", JSONPath: "object.metadata.annotations[storageclass.kubernetes.io/is-default-class]"},
			{DisplayName: "VIRT-DEFAULT", JSONPath: "object.metadata.annotations[storageclass.kubevirt.io/is-default-virt-class]"},
		}
	default:
		defaultHeaders = []output.Header{
			{DisplayName: "NAME", JSONPath: "name"},
			{DisplayName: "ID", JSONPath: "id"},
			{DisplayName: "TYPE", JSONPath: "type"},
			{DisplayName: "CAPACITY", JSONPath: "capacityHuman"},
			{DisplayName: "FREE", JSONPath: "freeHuman"},
			{DisplayName: "MAINTENANCE", JSONPath: "maintenance"},
		}
	}

	// Select appropriate subPath based on provider type
	var subPath string
	switch providerType {
	case "openshift":
		subPath = "storageclasses?detail=4"
	default:
		subPath = "datastores?detail=4"
	}

	// Fetch storage inventory from the provider
	data, err := client.FetchProviderInventory(kubeConfigFlags, inventoryURL, provider, subPath)
	if err != nil {
		return fmt.Errorf("failed to fetch storage inventory: %v", err)
	}

	// Verify data is an array
	dataArray, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("unexpected data format: expected array for storage inventory")
	}

	// Convert to expected format
	storages := make([]map[string]interface{}, 0, len(dataArray))
	for _, item := range dataArray {
		if storage, ok := item.(map[string]interface{}); ok {
			// Add provider name to each storage
			storage["provider"] = providerName

			// Humanize capacity and free space
			if capacity, exists := storage["capacity"]; exists {
				if capacityFloat, ok := capacity.(float64); ok {
					storage["capacityHuman"] = humanizeBytes(capacityFloat)
				} else if capacityNum, ok := capacity.(int64); ok {
					storage["capacityHuman"] = humanizeBytes(float64(capacityNum))
				}
			}

			if free, exists := storage["free"]; exists {
				if freeFloat, ok := free.(float64); ok {
					storage["freeHuman"] = humanizeBytes(freeFloat)
				} else if freeNum, ok := free.(int64); ok {
					storage["freeHuman"] = humanizeBytes(float64(freeNum))
				}
			}

			storages = append(storages, storage)
		}
	}

	// Parse and apply query options
	queryOpts, err := querypkg.ParseQueryString(query)
	if err != nil {
		return fmt.Errorf("invalid query string: %v", err)
	}

	// Apply query options (sorting, filtering, limiting)
	storages, err = querypkg.ApplyQuery(storages, queryOpts)
	if err != nil {
		return fmt.Errorf("error applying query: %v", err)
	}

	// Format validation
	outputFormat = strings.ToLower(outputFormat)
	if outputFormat != "table" && outputFormat != "json" && outputFormat != "yaml" {
		return fmt.Errorf("unsupported output format: %s. Supported formats: table, json, yaml", outputFormat)
	}

	// Handle different output formats
	if outputFormat == "json" {
		// Use JSON printer
		jsonPrinter := output.NewJSONPrinter().
			WithPrettyPrint(true).
			AddItems(storages)

		if len(storages) == 0 {
			return jsonPrinter.PrintEmpty(fmt.Sprintf("No storages found for provider %s", providerName))
		}
		return jsonPrinter.Print()
	} else if outputFormat == "yaml" {
		// Use YAML printer
		yamlPrinter := output.NewYAMLPrinter().
			AddItems(storages)

		if len(storages) == 0 {
			return yamlPrinter.PrintEmpty(fmt.Sprintf("No storages found for provider %s", providerName))
		}
		return yamlPrinter.Print()
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

		tablePrinter.AddItems(storages)

		if len(storages) == 0 {
			return tablePrinter.PrintEmpty(fmt.Sprintf("No storage resources found for provider %s", providerName))
		}
		return tablePrinter.Print()
	}
}
