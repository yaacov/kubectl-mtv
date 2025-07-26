package inventory

import (
	"fmt"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/util/output"
	querypkg "github.com/yaacov/kubectl-mtv/pkg/util/query"
	"github.com/yaacov/kubectl-mtv/pkg/util/watch"
)

// ListInstances queries the provider's instance inventory and displays the results
func ListInstances(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, watchMode bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return listInstancesOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query)
		}, 10*time.Second)
	}

	return listInstancesOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query)
}

func listInstancesOnce(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string) error {
	// Get the provider object
	provider, err := GetProviderByName(kubeConfigFlags, providerName, namespace)
	if err != nil {
		return err
	}

	// Create a new provider client
	providerClient := NewProviderClient(kubeConfigFlags, provider, inventoryURL)

	// Get provider type to verify instance support
	providerType, err := providerClient.GetProviderType()
	if err != nil {
		return fmt.Errorf("failed to get provider type: %v", err)
	}

	// Define default headers
	defaultHeaders := []output.Header{
		{DisplayName: "NAME", JSONPath: "name"},
		{DisplayName: "ID", JSONPath: "id"},
		{DisplayName: "STATUS", JSONPath: "status"},
		{DisplayName: "FLAVOR", JSONPath: "flavor.name"},
		{DisplayName: "IMAGE", JSONPath: "image.name"},
		{DisplayName: "PROJECT", JSONPath: "project.name"},
	}

	// Fetch instances inventory from the provider
	var data interface{}
	switch providerType {
	case "openstack":
		data, err = providerClient.GetInstances(4)
	default:
		return fmt.Errorf("provider type '%s' does not support instance inventory", providerType)
	}

	if err != nil {
		return fmt.Errorf("failed to get instances from provider: %v", err)
	}

	// Parse query options for advanced query features
	var queryOpts *querypkg.QueryOptions
	if query != "" {
		queryOpts, err = querypkg.ParseQueryString(query)
		if err != nil {
			return fmt.Errorf("failed to parse query: %v", err)
		}

		// Apply query filter
		data, err = querypkg.ApplyQueryInterface(data, query)
		if err != nil {
			return fmt.Errorf("failed to apply query: %v", err)
		}
	}

	// Format and display the results
	emptyMessage := fmt.Sprintf("No instances found for provider %s", providerName)
	switch outputFormat {
	case "json":
		return output.PrintJSONWithEmpty(data, emptyMessage)
	case "yaml":
		return output.PrintYAMLWithEmpty(data, emptyMessage)
	case "table":
		return output.PrintTableWithQuery(data, defaultHeaders, queryOpts, emptyMessage)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// ListImages queries the provider's image inventory and displays the results
func ListImages(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, watchMode bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return listImagesOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query)
		}, 10*time.Second)
	}

	return listImagesOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query)
}

func listImagesOnce(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string) error {
	// Get the provider object
	provider, err := GetProviderByName(kubeConfigFlags, providerName, namespace)
	if err != nil {
		return err
	}

	// Create a new provider client
	providerClient := NewProviderClient(kubeConfigFlags, provider, inventoryURL)

	// Get provider type to verify image support
	providerType, err := providerClient.GetProviderType()
	if err != nil {
		return fmt.Errorf("failed to get provider type: %v", err)
	}

	// Define default headers
	defaultHeaders := []output.Header{
		{DisplayName: "NAME", JSONPath: "name"},
		{DisplayName: "ID", JSONPath: "id"},
		{DisplayName: "STATUS", JSONPath: "status"},
		{DisplayName: "SIZE", JSONPath: "sizeHuman"},
		{DisplayName: "VISIBILITY", JSONPath: "visibility"},
	}

	// Fetch images inventory from the provider
	var data interface{}
	switch providerType {
	case "openstack":
		data, err = providerClient.GetImages(4)
	default:
		return fmt.Errorf("provider type '%s' does not support image inventory", providerType)
	}

	if err != nil {
		return fmt.Errorf("failed to get images from provider: %v", err)
	}

	// Process data to add human-readable sizes
	data = addHumanReadableImageSizes(data)

	// Parse query options for advanced query features
	var queryOpts *querypkg.QueryOptions
	if query != "" {
		queryOpts, err = querypkg.ParseQueryString(query)
		if err != nil {
			return fmt.Errorf("failed to parse query: %v", err)
		}

		// Apply query filter
		data, err = querypkg.ApplyQueryInterface(data, query)
		if err != nil {
			return fmt.Errorf("failed to apply query: %v", err)
		}
	}

	// Format and display the results
	emptyMessage := fmt.Sprintf("No images found for provider %s", providerName)
	switch outputFormat {
	case "json":
		return output.PrintJSONWithEmpty(data, emptyMessage)
	case "yaml":
		return output.PrintYAMLWithEmpty(data, emptyMessage)
	case "table":
		return output.PrintTableWithQuery(data, defaultHeaders, queryOpts, emptyMessage)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// ListFlavors queries the provider's flavor inventory and displays the results
func ListFlavors(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, watchMode bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return listFlavorsOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query)
		}, 10*time.Second)
	}

	return listFlavorsOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query)
}

func listFlavorsOnce(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string) error {
	// Get the provider object
	provider, err := GetProviderByName(kubeConfigFlags, providerName, namespace)
	if err != nil {
		return err
	}

	// Create a new provider client
	providerClient := NewProviderClient(kubeConfigFlags, provider, inventoryURL)

	// Get provider type to verify flavor support
	providerType, err := providerClient.GetProviderType()
	if err != nil {
		return fmt.Errorf("failed to get provider type: %v", err)
	}

	// Define default headers
	defaultHeaders := []output.Header{
		{DisplayName: "NAME", JSONPath: "name"},
		{DisplayName: "ID", JSONPath: "id"},
		{DisplayName: "VCPUS", JSONPath: "vcpus"},
		{DisplayName: "RAM", JSONPath: "ramHuman"},
		{DisplayName: "DISK", JSONPath: "diskHuman"},
		{DisplayName: "EPHEMERAL", JSONPath: "ephemeralHuman"},
	}

	// Fetch flavors inventory from the provider
	var data interface{}
	switch providerType {
	case "openstack":
		data, err = providerClient.GetFlavors(4)
	default:
		return fmt.Errorf("provider type '%s' does not support flavor inventory", providerType)
	}

	if err != nil {
		return fmt.Errorf("failed to get flavors from provider: %v", err)
	}

	// Process data to add human-readable sizes
	data = addHumanReadableFlavorSizes(data)

	// Parse query options for advanced query features
	var queryOpts *querypkg.QueryOptions
	if query != "" {
		queryOpts, err = querypkg.ParseQueryString(query)
		if err != nil {
			return fmt.Errorf("failed to parse query: %v", err)
		}

		// Apply query filter
		data, err = querypkg.ApplyQueryInterface(data, query)
		if err != nil {
			return fmt.Errorf("failed to apply query: %v", err)
		}
	}

	// Format and display the results
	emptyMessage := fmt.Sprintf("No flavors found for provider %s", providerName)
	switch outputFormat {
	case "json":
		return output.PrintJSONWithEmpty(data, emptyMessage)
	case "yaml":
		return output.PrintYAMLWithEmpty(data, emptyMessage)
	case "table":
		return output.PrintTableWithQuery(data, defaultHeaders, queryOpts, emptyMessage)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// ListProjects queries the provider's project inventory and displays the results
func ListProjects(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, watchMode bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return listProjectsOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query)
		}, 10*time.Second)
	}

	return listProjectsOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query)
}

func listProjectsOnce(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string) error {
	// Get the provider object
	provider, err := GetProviderByName(kubeConfigFlags, providerName, namespace)
	if err != nil {
		return err
	}

	// Create a new provider client
	providerClient := NewProviderClient(kubeConfigFlags, provider, inventoryURL)

	// Get provider type to verify project support
	providerType, err := providerClient.GetProviderType()
	if err != nil {
		return fmt.Errorf("failed to get provider type: %v", err)
	}

	// Define default headers
	defaultHeaders := []output.Header{
		{DisplayName: "NAME", JSONPath: "name"},
		{DisplayName: "ID", JSONPath: "id"},
		{DisplayName: "DESCRIPTION", JSONPath: "description"},
		{DisplayName: "ENABLED", JSONPath: "enabled"},
	}

	// Fetch projects inventory from the provider
	var data interface{}
	switch providerType {
	case "openstack":
		data, err = providerClient.GetProjects(4)
	default:
		return fmt.Errorf("provider type '%s' does not support project inventory", providerType)
	}

	if err != nil {
		return fmt.Errorf("failed to get projects from provider: %v", err)
	}

	// Parse query options for advanced query features
	var queryOpts *querypkg.QueryOptions
	if query != "" {
		queryOpts, err = querypkg.ParseQueryString(query)
		if err != nil {
			return fmt.Errorf("failed to parse query: %v", err)
		}

		// Apply query filter
		data, err = querypkg.ApplyQueryInterface(data, query)
		if err != nil {
			return fmt.Errorf("failed to apply query: %v", err)
		}
	}

	// Format and display the results
	emptyMessage := fmt.Sprintf("No projects found for provider %s", providerName)
	switch outputFormat {
	case "json":
		return output.PrintJSONWithEmpty(data, emptyMessage)
	case "yaml":
		return output.PrintYAMLWithEmpty(data, emptyMessage)
	case "table":
		return output.PrintTableWithQuery(data, defaultHeaders, queryOpts, emptyMessage)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// Helper functions for adding human-readable sizes

// addHumanReadableImageSizes adds human-readable size fields to image data
func addHumanReadableImageSizes(data interface{}) interface{} {
	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			if image, ok := item.(map[string]interface{}); ok {
				if size, exists := image["size"]; exists {
					if sizeVal, ok := size.(float64); ok {
						image["sizeHuman"] = humanizeBytes(sizeVal)
					}
				}
			}
		}
	case map[string]interface{}:
		if size, exists := v["size"]; exists {
			if sizeVal, ok := size.(float64); ok {
				v["sizeHuman"] = humanizeBytes(sizeVal)
			}
		}
	}
	return data
}

// addHumanReadableFlavorSizes adds human-readable size fields to flavor data
func addHumanReadableFlavorSizes(data interface{}) interface{} {
	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			if flavor, ok := item.(map[string]interface{}); ok {
				if ram, exists := flavor["ram"]; exists {
					if ramVal, ok := ram.(float64); ok {
						flavor["ramHuman"] = humanizeBytes(ramVal * 1024 * 1024) // RAM is in MB
					}
				}
				if disk, exists := flavor["disk"]; exists {
					if diskVal, ok := disk.(float64); ok {
						flavor["diskHuman"] = humanizeBytes(diskVal * 1024 * 1024 * 1024) // Disk is in GB
					}
				}
				if ephemeral, exists := flavor["ephemeral"]; exists {
					if ephemeralVal, ok := ephemeral.(float64); ok {
						flavor["ephemeralHuman"] = humanizeBytes(ephemeralVal * 1024 * 1024 * 1024) // Ephemeral is in GB
					}
				}
			}
		}
	case map[string]interface{}:
		if ram, exists := v["ram"]; exists {
			if ramVal, ok := ram.(float64); ok {
				v["ramHuman"] = humanizeBytes(ramVal * 1024 * 1024) // RAM is in MB
			}
		}
		if disk, exists := v["disk"]; exists {
			if diskVal, ok := disk.(float64); ok {
				v["diskHuman"] = humanizeBytes(diskVal * 1024 * 1024 * 1024) // Disk is in GB
			}
		}
		if ephemeral, exists := v["ephemeral"]; exists {
			if ephemeralVal, ok := ephemeral.(float64); ok {
				v["ephemeralHuman"] = humanizeBytes(ephemeralVal * 1024 * 1024 * 1024) // Ephemeral is in GB
			}
		}
	}
	return data
}
