package inventory

import (
	"context"
	"fmt"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/util/output"
	querypkg "github.com/yaacov/kubectl-mtv/pkg/util/query"
	"github.com/yaacov/kubectl-mtv/pkg/util/watch"
)

// ListEC2InstancesWithInsecure queries the provider's EC2 instance inventory with optional insecure TLS skip verification
func ListEC2InstancesWithInsecure(ctx context.Context, kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, watchMode bool, insecureSkipTLS bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return listEC2InstancesOnce(ctx, kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query, insecureSkipTLS)
		}, 10*time.Second)
	}

	return listEC2InstancesOnce(ctx, kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query, insecureSkipTLS)
}

func listEC2InstancesOnce(ctx context.Context, kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, insecureSkipTLS bool) error {
	// Get the provider object
	provider, err := GetProviderByName(ctx, kubeConfigFlags, providerName, namespace)
	if err != nil {
		return err
	}

	// Create a new provider client
	providerClient := NewProviderClientWithInsecure(kubeConfigFlags, provider, inventoryURL, insecureSkipTLS)

	// Get provider type to verify EC2 support
	providerType, err := providerClient.GetProviderType()
	if err != nil {
		return fmt.Errorf("failed to get provider type: %v", err)
	}

	// Verify this is an EC2 provider
	if providerType != "ec2" {
		return fmt.Errorf("provider type '%s' is not an EC2 provider", providerType)
	}

	// Define default headers for EC2 instances
	// Note: AWS API returns PascalCase field names
	// NAME column shows Name tag if available, otherwise falls back to InstanceId
	defaultHeaders := []output.Header{
		{DisplayName: "NAME", JSONPath: "name"},
		{DisplayName: "TYPE", JSONPath: "InstanceType"},
		{DisplayName: "STATE", JSONPath: "State.Name"},
		{DisplayName: "PLATFORM", JSONPath: "PlatformDetails"},
		{DisplayName: "AZ", JSONPath: "Placement.AvailabilityZone"},
		{DisplayName: "PUBLIC-IP", JSONPath: "PublicIpAddress"},
		{DisplayName: "PRIVATE-IP", JSONPath: "PrivateIpAddress"},
	}

	// Fetch EC2 instances from the provider
	data, err := providerClient.GetVMs(ctx, 4)
	if err != nil {
		return fmt.Errorf("failed to get EC2 instances from provider: %v", err)
	}

	// Process data to extract name from tags
	data = addEC2InstanceNames(data)

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
	emptyMessage := fmt.Sprintf("No EC2 instances found for provider %s", providerName)
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

// ListEC2VolumesWithInsecure queries the provider's EC2 EBS volume inventory with optional insecure TLS skip verification
func ListEC2VolumesWithInsecure(ctx context.Context, kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, watchMode bool, insecureSkipTLS bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return listEC2VolumesOnce(ctx, kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query, insecureSkipTLS)
		}, 10*time.Second)
	}

	return listEC2VolumesOnce(ctx, kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query, insecureSkipTLS)
}

func listEC2VolumesOnce(ctx context.Context, kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, insecureSkipTLS bool) error {
	// Get the provider object
	provider, err := GetProviderByName(ctx, kubeConfigFlags, providerName, namespace)
	if err != nil {
		return err
	}

	// Create a new provider client
	providerClient := NewProviderClientWithInsecure(kubeConfigFlags, provider, inventoryURL, insecureSkipTLS)

	// Get provider type to verify EC2 support
	providerType, err := providerClient.GetProviderType()
	if err != nil {
		return fmt.Errorf("failed to get provider type: %v", err)
	}

	// Verify this is an EC2 provider
	if providerType != "ec2" {
		return fmt.Errorf("provider type '%s' is not an EC2 provider", providerType)
	}

	// Define default headers for EC2 volumes
	// Note: AWS API returns PascalCase field names
	defaultHeaders := []output.Header{
		{DisplayName: "NAME", JSONPath: "name"},
		{DisplayName: "ID", JSONPath: "VolumeId"},
		{DisplayName: "SIZE", JSONPath: "sizeHuman"},
		{DisplayName: "TYPE", JSONPath: "VolumeType"},
		{DisplayName: "STATE", JSONPath: "State"},
		{DisplayName: "IOPS", JSONPath: "Iops"},
		{DisplayName: "THROUGHPUT", JSONPath: "Throughput"},
		{DisplayName: "ATTACHED-TO", JSONPath: "attachedTo"},
	}

	// Fetch EC2 volumes from the provider
	data, err := providerClient.GetVolumes(ctx, 4)
	if err != nil {
		return fmt.Errorf("failed to get EC2 volumes from provider: %v", err)
	}

	// Process data to add human-readable fields
	data = addEC2VolumeFields(data)

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
	emptyMessage := fmt.Sprintf("No EC2 volumes found for provider %s", providerName)
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

// ListEC2VolumeTypesWithInsecure queries the provider's EC2 volume type inventory with optional insecure TLS skip verification
func ListEC2VolumeTypesWithInsecure(ctx context.Context, kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, watchMode bool, insecureSkipTLS bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return listEC2VolumeTypesOnce(ctx, kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query, insecureSkipTLS)
		}, 10*time.Second)
	}

	return listEC2VolumeTypesOnce(ctx, kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query, insecureSkipTLS)
}

func listEC2VolumeTypesOnce(ctx context.Context, kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, insecureSkipTLS bool) error {
	// Get the provider object
	provider, err := GetProviderByName(ctx, kubeConfigFlags, providerName, namespace)
	if err != nil {
		return err
	}

	// Create a new provider client
	providerClient := NewProviderClientWithInsecure(kubeConfigFlags, provider, inventoryURL, insecureSkipTLS)

	// Get provider type to verify EC2 support
	providerType, err := providerClient.GetProviderType()
	if err != nil {
		return fmt.Errorf("failed to get provider type: %v", err)
	}

	// Verify this is an EC2 provider
	if providerType != "ec2" {
		return fmt.Errorf("provider type '%s' is not an EC2 provider", providerType)
	}

	// Define default headers for EC2 volume types
	defaultHeaders := []output.Header{
		{DisplayName: "TYPE", JSONPath: "type"},
		{DisplayName: "DESCRIPTION", JSONPath: "description"},
		{DisplayName: "MAX-IOPS", JSONPath: "maxIOPS"},
		{DisplayName: "MAX-THROUGHPUT", JSONPath: "maxThroughput"},
	}

	// Fetch EC2 volume types (storage classes) from the provider
	data, err := providerClient.GetResourceCollection(ctx, "storages", 4)
	if err != nil {
		return fmt.Errorf("failed to get EC2 volume types from provider: %v", err)
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
	emptyMessage := fmt.Sprintf("No EC2 volume types found for provider %s", providerName)
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

// ListEC2NetworksWithInsecure queries the provider's EC2 network inventory (VPCs and Subnets) with optional insecure TLS skip verification
func ListEC2NetworksWithInsecure(ctx context.Context, kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, watchMode bool, insecureSkipTLS bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return listEC2NetworksOnce(ctx, kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query, insecureSkipTLS)
		}, 10*time.Second)
	}

	return listEC2NetworksOnce(ctx, kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, query, insecureSkipTLS)
}

func listEC2NetworksOnce(ctx context.Context, kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, query string, insecureSkipTLS bool) error {
	// Get the provider object
	provider, err := GetProviderByName(ctx, kubeConfigFlags, providerName, namespace)
	if err != nil {
		return err
	}

	// Create a new provider client
	providerClient := NewProviderClientWithInsecure(kubeConfigFlags, provider, inventoryURL, insecureSkipTLS)

	// Get provider type to verify EC2 support
	providerType, err := providerClient.GetProviderType()
	if err != nil {
		return fmt.Errorf("failed to get provider type: %v", err)
	}

	// Verify this is an EC2 provider
	if providerType != "ec2" {
		return fmt.Errorf("provider type '%s' is not an EC2 provider", providerType)
	}

	// Define default headers for EC2 networks
	// Note: AWS API returns PascalCase field names
	defaultHeaders := []output.Header{
		{DisplayName: "NAME", JSONPath: "name"},
		{DisplayName: "ID", JSONPath: "id"},
		{DisplayName: "TYPE", JSONPath: "networkType"},
		{DisplayName: "CIDR", JSONPath: "CidrBlock"},
		{DisplayName: "STATE", JSONPath: "State"},
		{DisplayName: "DEFAULT", JSONPath: "IsDefault"},
	}

	// Fetch EC2 networks from the provider
	data, err := providerClient.GetNetworks(ctx, 4)
	if err != nil {
		return fmt.Errorf("failed to get EC2 networks from provider: %v", err)
	}

	// Process data to extract names and normalize fields
	data = addEC2NetworkFields(data)

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
	emptyMessage := fmt.Sprintf("No EC2 networks found for provider %s", providerName)
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

// Helper functions for EC2 data processing

// extractNameFromTags extracts the Name tag value from AWS Tags array
func extractNameFromTags(tags interface{}) (string, bool) {
	tagsArray, ok := tags.([]interface{})
	if !ok {
		return "", false
	}
	for _, tagInterface := range tagsArray {
		if tag, ok := tagInterface.(map[string]interface{}); ok {
			if key, keyOk := tag["Key"].(string); keyOk && key == "Name" {
				if value, valueOk := tag["Value"].(string); valueOk {
					return value, true
				}
			}
		}
	}
	return "", false
}

// addEC2InstanceNames extracts instance names from tags
// AWS API returns PascalCase: Tags, Key, Value, InstanceId
func addEC2InstanceNames(data interface{}) interface{} {
	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			if instance, ok := item.(map[string]interface{}); ok {
				// Extract name from tags using shared helper
				if tags, exists := instance["Tags"]; exists {
					if name, found := extractNameFromTags(tags); found {
						instance["name"] = name
					}
				}
				// If no name tag, use instance ID
				if _, hasName := instance["name"]; !hasName {
					if instanceID, ok := instance["InstanceId"].(string); ok {
						instance["name"] = instanceID
					}
				}
			}
		}
	case map[string]interface{}:
		// Single instance
		if tags, exists := v["Tags"]; exists {
			if name, found := extractNameFromTags(tags); found {
				v["name"] = name
			}
		}
		if _, hasName := v["name"]; !hasName {
			if instanceID, ok := v["InstanceId"].(string); ok {
				v["name"] = instanceID
			}
		}
	}
	return data
}

// addEC2VolumeFields adds human-readable fields to volume data
func addEC2VolumeFields(data interface{}) interface{} {
	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			if volume, ok := item.(map[string]interface{}); ok {
				processEC2Volume(volume)
			}
		}
	case map[string]interface{}:
		processEC2Volume(v)
	}
	return data
}

func processEC2Volume(volume map[string]interface{}) {
	// Extract name from tags using shared helper
	if tags, exists := volume["Tags"]; exists {
		if name, found := extractNameFromTags(tags); found {
			volume["name"] = name
		}
	}
	// If no name tag, use volume ID
	if _, hasName := volume["name"]; !hasName {
		if volumeID, ok := volume["VolumeId"].(string); ok {
			volume["name"] = volumeID
		}
	}

	// Add human-readable size (Size is in GiB)
	if size, exists := volume["Size"]; exists {
		if sizeVal, ok := size.(float64); ok {
			volume["sizeHuman"] = humanizeBytes(sizeVal * 1024 * 1024 * 1024) // Size is in GiB
		}
	}

	// Extract attached instance ID (Attachments, InstanceId)
	if attachments, exists := volume["Attachments"]; exists {
		if attachmentsArray, ok := attachments.([]interface{}); ok && len(attachmentsArray) > 0 {
			if attachment, ok := attachmentsArray[0].(map[string]interface{}); ok {
				if instanceID, ok := attachment["InstanceId"].(string); ok {
					volume["attachedTo"] = instanceID
				}
			}
		}
	}
	if _, hasAttached := volume["attachedTo"]; !hasAttached {
		volume["attachedTo"] = "-"
	}
}

// addEC2NetworkFields adds normalized fields to network data
func addEC2NetworkFields(data interface{}) interface{} {
	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			if network, ok := item.(map[string]interface{}); ok {
				processEC2Network(network)
			}
		}
	case map[string]interface{}:
		processEC2Network(v)
	}
	return data
}

func processEC2Network(network map[string]interface{}) {
	// Extract name from tags using shared helper
	if _, hasName := network["name"]; !hasName {
		if tags, exists := network["Tags"]; exists {
			if name, found := extractNameFromTags(tags); found {
				network["name"] = name
			}
		}
	}

	// Normalize ID field and determine network type based on which ID is present
	// Subnet takes precedence over VPC
	if _, hasID := network["id"]; !hasID {
		if subnetID, ok := network["SubnetId"].(string); ok {
			// It's a Subnet
			network["id"] = subnetID
			network["networkType"] = "subnet"
		} else if vpcID, ok := network["VpcId"].(string); ok {
			// It's a VPC
			network["id"] = vpcID
			network["networkType"] = "vpc"
		}
	}

	// Fallback if networkType wasn't set (malformed data)
	if _, hasType := network["networkType"]; !hasType {
		network["networkType"] = "unknown"
	}

	// If still no name, use ID
	if _, hasName := network["name"]; !hasName {
		if id, ok := network["id"].(string); ok {
			network["name"] = id
		}
	}
}
