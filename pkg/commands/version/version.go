package version

import (
	"context"
	"strings"

	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Info holds all version-related information
type Info struct {
	ClientVersion     string `json:"clientVersion" yaml:"clientVersion"`
	OperatorVersion   string `json:"operatorVersion" yaml:"operatorVersion"`
	OperatorStatus    string `json:"operatorStatus" yaml:"operatorStatus"`
	OperatorNamespace string `json:"operatorNamespace,omitempty" yaml:"operatorNamespace,omitempty"`
	InventoryURL      string `json:"inventoryURL" yaml:"inventoryURL"`
	InventoryStatus   string `json:"inventoryStatus" yaml:"inventoryStatus"`
}

// GetInventoryInfo returns information about the MTV inventory service
func GetInventoryInfo(kubeConfigFlags *genericclioptions.ConfigFlags) (string, string) {
	namespace := client.ResolveNamespace(kubeConfigFlags)

	// Try to discover inventory URL
	inventoryURL := client.DiscoverInventoryURL(kubeConfigFlags, namespace)
	if inventoryURL != "" {
		return inventoryURL, "available"
	}

	return "not found", "not available"
}

// GetMTVControllerInfo returns information about the MTV Operator
func GetMTVControllerInfo(kubeConfigFlags *genericclioptions.ConfigFlags) (string, string, string) {
	// Try to get dynamic client
	dynamicClient, err := client.GetDynamicClient(kubeConfigFlags)
	if err != nil {
		return "unknown", "error connecting to cluster", ""
	}

	// Check if MTV is installed by looking for the providers CRD
	crdGVR := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}

	crd, err := dynamicClient.Resource(crdGVR).Get(context.TODO(), "providers.forklift.konveyor.io", metav1.GetOptions{})
	if err != nil {
		return "not found", "not available", ""
	}

	// Extract version and namespace from annotations
	version := "unknown"
	namespace := "unknown"
	status := "installed"

	if crd != nil {
		// Try to get version from operator annotation
		if annotations, found, _ := unstructured.NestedStringMap(crd.Object, "metadata", "annotations"); found {
			// Look for operatorframework.io/installed-alongside annotation
			for key, value := range annotations {
				if strings.HasPrefix(key, "operatorframework.io/installed-alongside-") {
					// Format: namespace/operator-name.version
					parts := strings.Split(value, "/")
					if len(parts) == 2 {
						namespace = parts[0]
						version = parts[1]
					}
					break
				}
			}
		}
	}

	return version, status, namespace
}

// GetVersionInfo gathers all version information
func GetVersionInfo(clientVersion string, kubeConfigFlags *genericclioptions.ConfigFlags) Info {
	// Get MTV Operator information
	controllerVersion, controllerStatus, controllerNamespace := GetMTVControllerInfo(kubeConfigFlags)

	// Get inventory information
	inventoryURL, inventoryStatus := GetInventoryInfo(kubeConfigFlags)

	return Info{
		ClientVersion:     clientVersion,
		OperatorVersion:   controllerVersion,
		OperatorStatus:    controllerStatus,
		OperatorNamespace: controllerNamespace,
		InventoryURL:      inventoryURL,
		InventoryStatus:   inventoryStatus,
	}
}
