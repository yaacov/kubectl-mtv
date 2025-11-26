package provider

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/provider/providerutil"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/output"
)

// getDynamicInventoryColumns returns appropriate inventory columns based on provider types
func getDynamicInventoryColumns(providerTypes map[string]bool) []output.Header {
	var columns []output.Header
	vmColumnAdded := false
	hostsColumnAdded := false

	// Add vSphere-specific columns
	if providerTypes["vsphere"] {
		if len(providerTypes) == 1 {
			// Only vSphere - show VMS, HOSTS, and more details
			columns = append(columns,
				output.Header{DisplayName: "VMS", JSONPath: "vmCount"},
				output.Header{DisplayName: "HOSTS", JSONPath: "hostCount"},
				output.Header{DisplayName: "DATACENTERS", JSONPath: "datacenterCount"},
				output.Header{DisplayName: "CLUSTERS", JSONPath: "clusterCount"},
				output.Header{DisplayName: "DATASTORES", JSONPath: "datastoreCount"},
			)
			vmColumnAdded = true
			hostsColumnAdded = true
		}
	}

	// Add oVirt-specific columns
	if providerTypes["ovirt"] {
		if len(providerTypes) == 1 {
			// Only oVirt - show VMS, HOSTS, and more details
			columns = append(columns,
				output.Header{DisplayName: "VMS", JSONPath: "vmCount"},
				output.Header{DisplayName: "HOSTS", JSONPath: "hostCount"},
				output.Header{DisplayName: "DATACENTERS", JSONPath: "datacenterCount"},
				output.Header{DisplayName: "CLUSTERS", JSONPath: "clusterCount"},
				output.Header{DisplayName: "STORAGE-DOMAINS", JSONPath: "storageDomainCount"},
			)
			vmColumnAdded = true
			hostsColumnAdded = true
		}
	}

	// Add OVA-specific columns
	if providerTypes["ova"] {
		if len(providerTypes) == 1 {
			// Only OVA - show VMS
			columns = append(columns, output.Header{DisplayName: "VMS", JSONPath: "vmCount"})
			vmColumnAdded = true
		}
	}

	// Add OpenStack-specific columns
	if providerTypes["openstack"] {
		if len(providerTypes) == 1 {
			// Only OpenStack - use INSTANCES and show more details
			columns = append(columns,
				output.Header{DisplayName: "INSTANCES", JSONPath: "vmCount"},
				output.Header{DisplayName: "REGIONS", JSONPath: "regionCount"},
				output.Header{DisplayName: "PROJECTS", JSONPath: "projectCount"},
				output.Header{DisplayName: "VOLUMES", JSONPath: "volumeCount"},
			)
			vmColumnAdded = true
		}
	}

	// Add OpenShift-specific columns
	if providerTypes["openshift"] {
		if len(providerTypes) == 1 {
			// Only OpenShift - show VMS and more details
			columns = append(columns,
				output.Header{DisplayName: "VMS", JSONPath: "vmCount"},
				output.Header{DisplayName: "STORAGE-CLASSES", JSONPath: "storageClassCount"},
			)
			vmColumnAdded = true
		}
	}

	// Add EC2-specific columns
	if providerTypes["ec2"] {
		if len(providerTypes) == 1 {
			// Only EC2 - use INSTANCES
			columns = append(columns, output.Header{DisplayName: "INSTANCES", JSONPath: "vmCount"})
			vmColumnAdded = true
			// EC2 doesn't have additional inventory counts from the backend yet
			// but we're prepared for when it does
		}
	}

	// Handle mixed environments - add default VM and HOSTS columns if not already added
	if !vmColumnAdded {
		// Mixed environment - use generic VMS term
		columns = append(columns, output.Header{DisplayName: "VMS", JSONPath: "vmCount"})
	}

	if !hostsColumnAdded && (providerTypes["vsphere"] || providerTypes["ovirt"]) {
		// Mixed environment with vsphere/ovirt - add HOSTS column
		columns = append(columns, output.Header{DisplayName: "HOSTS", JSONPath: "hostCount"})
	}

	// Add NETWORKS column for all providers (most have it)
	columns = append(columns, output.Header{DisplayName: "NETWORKS", JSONPath: "networkCount"})

	return columns
}

// getProviders retrieves all providers from the given namespace
func getProviders(ctx context.Context, dynamicClient dynamic.Interface, namespace string) (*unstructured.UnstructuredList, error) {
	if namespace != "" {
		return dynamicClient.Resource(client.ProvidersGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	} else {
		return dynamicClient.Resource(client.ProvidersGVR).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	}
}

// getSpecificProvider retrieves a specific provider by name
func getSpecificProvider(ctx context.Context, dynamicClient dynamic.Interface, namespace, providerName string) (*unstructured.UnstructuredList, error) {
	if namespace != "" {
		// If namespace is specified, get the specific resource
		provider, err := dynamicClient.Resource(client.ProvidersGVR).Namespace(namespace).Get(ctx, providerName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// Create a list with just this provider
		return &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{*provider},
		}, nil
	} else {
		// If no namespace specified, list all and filter by name
		providers, err := dynamicClient.Resource(client.ProvidersGVR).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
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
func List(ctx context.Context, configFlags *genericclioptions.ConfigFlags, namespace string, baseURL string, outputFormat string, providerName string, insecureSkipTLS bool) error {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	var providers *unstructured.UnstructuredList
	if providerName != "" {
		// Get specific provider by name
		providers, err = getSpecificProvider(ctx, c, namespace, providerName)
		if err != nil {
			return fmt.Errorf("failed to get provider: %v", err)
		}
	} else {
		// Get all providers
		providers, err = getProviders(ctx, c, namespace)
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
		route, err := client.GetForkliftInventoryRoute(ctx, configFlags, namespace)
		if err == nil && route != nil {
			host, found, _ := unstructured.NestedString(route.Object, "spec", "host")
			if found && host != "" {
				baseURL = fmt.Sprintf("https://%s", host)
			}
		}
	}

	// Fetch bulk provider inventory data
	var bulkProviderData map[string][]map[string]interface{}
	if baseURL != "" {
		// Detail level 1: includes basic inventory counts (vmCount, hostCount, etc.) without full resource lists
		if bulk, err := client.FetchProvidersWithDetailAndInsecure(ctx, configFlags, baseURL, 1, insecureSkipTLS); err == nil && bulk != nil {
			if bulkMap, ok := bulk.(map[string]interface{}); ok {
				bulkProviderData = make(map[string][]map[string]interface{})
				// Parse bulk data for each provider type
				for providerType, providerList := range bulkMap {
					if providerListSlice, ok := providerList.([]interface{}); ok {
						var typedProviders []map[string]interface{}
						for _, p := range providerListSlice {
							if providerMap, ok := p.(map[string]interface{}); ok {
								typedProviders = append(typedProviders, providerMap)
							}
						}
						bulkProviderData[providerType] = typedProviders
					}
				}
			} else {
				klog.V(4).Infof("Failed to parse bulk provider response: expected map, got %T", bulk)
			}
		} else {
			klog.V(4).Infof("Failed to fetch bulk provider data: %v", err)
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

		// Extract inventory counts from bulk data
		if bulkProviderData != nil {
			providerType, found, _ := unstructured.NestedString(provider.Object, "spec", "type")
			providerUID := string(provider.GetUID())
			providerName := provider.GetName()

			if found && providerType != "" {
				if providersList, exists := bulkProviderData[providerType]; exists {
					// Find matching provider by UID in bulk data
					providerFound := false
					for _, bulkProvider := range providersList {
						if bulkUID, ok := bulkProvider["uid"].(string); ok && bulkUID == providerUID {
							providerFound = true

							// Extract inventory counts from bulk data
							if vmCount, ok := bulkProvider["vmCount"]; ok {
								item["vmCount"] = vmCount
							}
							if hostCount, ok := bulkProvider["hostCount"]; ok {
								item["hostCount"] = hostCount
							}
							if datacenterCount, ok := bulkProvider["datacenterCount"]; ok {
								item["datacenterCount"] = datacenterCount
							}
							if clusterCount, ok := bulkProvider["clusterCount"]; ok {
								item["clusterCount"] = clusterCount
							}
							if networkCount, ok := bulkProvider["networkCount"]; ok {
								item["networkCount"] = networkCount
							}
							if datastoreCount, ok := bulkProvider["datastoreCount"]; ok {
								item["datastoreCount"] = datastoreCount
							}
							if storageClassCount, ok := bulkProvider["storageClassCount"]; ok {
								item["storageClassCount"] = storageClassCount
							}
							if product, ok := bulkProvider["product"]; ok {
								item["product"] = product
							}
							// Add other provider-type specific counts
							if regionCount, ok := bulkProvider["regionCount"]; ok {
								item["regionCount"] = regionCount
							}
							if projectCount, ok := bulkProvider["projectCount"]; ok {
								item["projectCount"] = projectCount
							}
							if imageCount, ok := bulkProvider["imageCount"]; ok {
								item["imageCount"] = imageCount
							}
							if volumeCount, ok := bulkProvider["volumeCount"]; ok {
								item["volumeCount"] = volumeCount
							}
							if volumeTypeCount, ok := bulkProvider["volumeTypeCount"]; ok {
								item["volumeTypeCount"] = volumeTypeCount
							}
							if diskCount, ok := bulkProvider["diskCount"]; ok {
								item["diskCount"] = diskCount
							}
							if storageCount, ok := bulkProvider["storageCount"]; ok {
								item["storageCount"] = storageCount
							}
							if storageDomainCount, ok := bulkProvider["storageDomainCount"]; ok {
								item["storageDomainCount"] = storageDomainCount
							}
							break
						}
					}

					if !providerFound {
						klog.V(4).Infof("Provider %s (%s) not found in bulk data", providerName, providerUID)
					}
				} else {
					klog.V(4).Infof("No bulk data available for provider type %s", providerType)
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

		// Add common columns
		headers = append(headers,
			output.Header{DisplayName: "TYPE", JSONPath: "spec.type"},
			output.Header{DisplayName: "URL", JSONPath: "spec.url"},
			output.Header{DisplayName: "STATUS", JSONPath: "status.phase"},
			output.Header{DisplayName: "CONNECTED", JSONPath: "conditionStatuses.ConnectionStatus"},
			output.Header{DisplayName: "INVENTORY", JSONPath: "conditionStatuses.InventoryStatus"},
			output.Header{DisplayName: "READY", JSONPath: "conditionStatuses.ReadyStatus"},
		)

		// Determine which provider types are present
		providerTypes := make(map[string]bool)
		for _, item := range items {
			if spec, ok := item["spec"].(map[string]interface{}); ok {
				if providerType, ok := spec["type"].(string); ok {
					providerTypes[providerType] = true
				}
			}
		}

		// Add dynamic columns based on provider types present
		headers = append(headers, getDynamicInventoryColumns(providerTypes)...)

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
