package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Group is the API group for MTV CRDs
const Group = "forklift.konveyor.io"

// Version is the API version for MTV CRDs
const Version = "v1beta1"

// Resource GVRs for MTV CRDs
var (
	ProvidersGVR = schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: "providers",
	}

	NetworkMapGVR = schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: "networkmaps",
	}

	StorageMapGVR = schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: "storagemaps",
	}

	PlansGVR = schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: "plans",
	}

	MigrationsGVR = schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: "migrations",
	}

	HostsGVR = schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: "hosts",
	}

	HooksGVR = schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: "hooks",
	}

	ForkliftControllersGVR = schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: "forkliftcontrollers",
	}

	// SecretGVR is used to access secrets in the cluster
	SecretsGVR = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	}

	// ConfigMapsGVR is used to access configmaps in the cluster
	ConfigMapsGVR = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "configmaps",
	}

	// RouteGVR is used to access routes in an Openshift cluster
	RouteGVR = schema.GroupVersionResource{
		Group:    "route.openshift.io",
		Version:  "v1",
		Resource: "routes",
	}
)

// GetDynamicClient returns a dynamic client for interacting with MTV CRDs
func GetDynamicClient(configFlags *genericclioptions.ConfigFlags) (dynamic.Interface, error) {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %v", err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	return client, nil
}

// GetKubernetesClientset returns a kubernetes clientset for interacting with the Kubernetes API
func GetKubernetesClientset(configFlags *genericclioptions.ConfigFlags) (*kubernetes.Clientset, error) {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %v", err)
	}

	return clientset, nil
}

// GetAuthenticatedTransport returns an HTTP transport configured with Kubernetes authentication
func GetAuthenticatedTransport(configFlags *genericclioptions.ConfigFlags) (http.RoundTripper, error) {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %v", err)
	}

	// Create a transport wrapper that adds authentication
	transport, err := rest.TransportFor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticated transport: %v", err)
	}

	return transport, nil
}

// GetAuthenticatedHTTPClient returns an HTTP client configured with Kubernetes authentication
func GetAuthenticatedHTTPClient(configFlags *genericclioptions.ConfigFlags, baseURL string) (*HTTPClient, error) {
	transport, err := GetAuthenticatedTransport(configFlags)
	if err != nil {
		return nil, err
	}

	return NewHTTPClient(baseURL, transport), nil
}

// HTTPClient represents a client for making HTTP requests with authentication
type HTTPClient struct {
	BaseURL    string
	httpClient *http.Client
}

// NewHTTPClient creates a new HTTP client with the given base URL and authentication
func NewHTTPClient(baseURL string, transport http.RoundTripper) *HTTPClient {
	if transport == nil {
		transport = http.DefaultTransport
	}

	client := &http.Client{
		Transport: transport,
	}

	return &HTTPClient{
		BaseURL:    baseURL,
		httpClient: client,
	}
}

// Get performs an HTTP GET request to the specified path with credentials
func (c *HTTPClient) Get(path string) ([]byte, error) {
	// Split the path into path part and query part
	parts := strings.SplitN(path, "?", 2)
	pathPart := parts[0]

	// Construct the base URL from baseURL and path part
	fullURL, err := url.JoinPath(c.BaseURL, pathPart)
	if err != nil {
		return nil, fmt.Errorf("failed to construct URL: %v", err)
	}

	// Append query string if it exists
	if len(parts) > 1 {
		fullURL = fullURL + "?" + parts[1]
	}

	// Create the request
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check for non-success status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}
