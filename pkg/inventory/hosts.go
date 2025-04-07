package inventory

import (
	"fmt"
	"strings"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/output"
	querypkg "github.com/yaacov/kubectl-mtv/pkg/query"
)

// ListHosts queries the provider's host inventory and displays the results
func ListHosts(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string) error {
	// Get the provider object
	provider, err := GetProviderByName(kubeConfigFlags, providerName, namespace)
	if err != nil {
		return err
	}

	// Fetch host inventory from the provider
	data, err := client.FetchProviderInventory(kubeConfigFlags, inventoryURL, provider, "hosts?detail=4")
	if err != nil {
		return fmt.Errorf("failed to fetch host inventory: %v", err)
	}

	// Verify data is an array
	dataArray, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("unexpected data format: expected array for host inventory")
	}

	// Convert to expected format
	hosts := make([]map[string]interface{}, 0, len(dataArray))
	for _, item := range dataArray {
		if host, ok := item.(map[string]interface{}); ok {
			// Add provider name to each host
			host["provider"] = providerName
			hosts = append(hosts, host)
		}
	}

	// Parse and apply query options
	queryOpts, err := querypkg.ParseQueryString(query)
	if err != nil {
		return fmt.Errorf("invalid query string: %v", err)
	}

	// Apply query options (sorting, filtering, limiting)
	hosts, err = querypkg.ApplyQuery(hosts, queryOpts)
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
			AddItems(hosts)

		if len(hosts) == 0 {
			return jsonPrinter.PrintEmpty(fmt.Sprintf("No hosts found for provider %s", providerName))
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
					JSONPath:    sel.Field,
				})
			}
			tablePrinter = output.NewTablePrinter().WithHeaders(headers...)
		} else {
			// Use default Table printer
			tablePrinter = output.NewTablePrinter().WithHeaders(
				output.Header{DisplayName: "NAME", JSONPath: "name"},
				output.Header{DisplayName: "ID", JSONPath: "id"},
				output.Header{DisplayName: "STATUS", JSONPath: "status"},
				output.Header{DisplayName: "VERSION", JSONPath: "productVersion"},
				output.Header{DisplayName: "MGMT IP", JSONPath: "managementServerIp"},
				output.Header{DisplayName: "CORES", JSONPath: "cpuCores"},
				output.Header{DisplayName: "SOCKETS", JSONPath: "cpuSockets"},
				output.Header{DisplayName: "MAINTENANCE", JSONPath: "inMaintenance"},
			)
		}

		tablePrinter.AddItems(hosts)

		if len(hosts) == 0 {
			return tablePrinter.PrintEmpty(fmt.Sprintf("No hosts found for provider %s", providerName))
		}
		return tablePrinter.Print()
	}
}
