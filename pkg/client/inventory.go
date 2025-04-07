package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// FetchProviderInventory fetches inventory for a specific provider
func FetchProviderInventory(configFlags *genericclioptions.ConfigFlags, baseURL string, provider *unstructured.Unstructured, subPath string) (interface{}, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider is nil")
	}

	httpClient, err := GetAuthenticatedHTTPClient(configFlags, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticated HTTP client: %v", err)
	}

	providerType, found, err := unstructured.NestedString(provider.Object, "spec", "type")
	if err != nil || !found {
		return nil, fmt.Errorf("provider type not found or error retrieving it: %v", err)
	}

	providerUID, found, err := unstructured.NestedString(provider.Object, "metadata", "uid")
	if err != nil || !found {
		return nil, fmt.Errorf("provider UID not found or error retrieving it: %v", err)
	}

	// Construct the path for provider inventory: /providers/<spec.type>/<metadata.uid>
	path := fmt.Sprintf("/providers/%s/%s", url.PathEscape(providerType), url.PathEscape(providerUID))

	// Add subPath if provided
	if subPath != "" {
		path = fmt.Sprintf("%s/%s", path, strings.TrimPrefix(subPath, "/"))
	}

	// Fetch the provider inventory
	responseBytes, err := httpClient.Get(path)
	if err != nil {
		return nil, err
	}

	// Parse the response as JSON
	var result interface{}
	if err := json.Unmarshal(responseBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse provider inventory response: %v", err)
	}

	return result, nil
}
