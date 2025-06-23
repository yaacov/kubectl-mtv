package inventory

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/output"
	querypkg "github.com/yaacov/kubectl-mtv/pkg/query"
	"github.com/yaacov/kubectl-mtv/pkg/watch"
)

// ListNamespaces queries the provider's namespace inventory and displays the results
func ListNamespaces(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, watchMode bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return listNamespacesOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query)
		}, 10*time.Second)
	}

	return listNamespacesOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query)
}

func listNamespacesOnce(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string) error {
	// Get the provider object
	provider, err := GetProviderByName(kubeConfigFlags, providerName, namespace)
	if err != nil {
		return err
	}

	// Fetch namespace inventory from the provider
	data, err := client.FetchProviderInventory(kubeConfigFlags, inventoryURL, provider, "namespaces?detail=4")
	if err != nil {
		return fmt.Errorf("failed to fetch namespace inventory: %v", err)
	}

	// Verify data is an array
	dataArray, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("unexpected data format: expected array for namespace inventory")
	}

	// Convert to expected format
	namespaces := make([]map[string]interface{}, 0, len(dataArray))
	for _, item := range dataArray {
		if ns, ok := item.(map[string]interface{}); ok {
			// Add provider name to each namespace
			ns["provider"] = providerName
			namespaces = append(namespaces, ns)
		}
	}

	// Parse and apply query options
	queryOpts, err := querypkg.ParseQueryString(query)
	if err != nil {
		return fmt.Errorf("invalid query string: %v", err)
	}

	// Apply query options (sorting, filtering, limiting)
	namespaces, err = querypkg.ApplyQuery(namespaces, queryOpts)
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
		jsonPrinter := output.NewJSONPrinter().
			WithPrettyPrint(true).
			AddItems(namespaces)

		if len(namespaces) == 0 {
			return jsonPrinter.PrintEmpty(fmt.Sprintf("No namespaces found for provider %s", providerName))
		}
		return jsonPrinter.Print()
	} else if outputFormat == "yaml" {
		yamlPrinter := output.NewYAMLPrinter().
			AddItems(namespaces)

		if len(namespaces) == 0 {
			return yamlPrinter.PrintEmpty(fmt.Sprintf("No namespaces found for provider %s", providerName))
		}
		return yamlPrinter.Print()
	} else {
		var tablePrinter *output.TablePrinter

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
			tablePrinter = output.NewTablePrinter().WithHeaders(
				output.Header{DisplayName: "NAME", JSONPath: "name"},
				output.Header{DisplayName: "ID", JSONPath: "id"},
				output.Header{DisplayName: "PROVIDER", JSONPath: "provider"},
			)
		}

		tablePrinter.AddItems(namespaces)

		if len(namespaces) == 0 {
			return tablePrinter.PrintEmpty(fmt.Sprintf("No namespaces found for provider %s", providerName))
		}
		return tablePrinter.Print()
	}
}
