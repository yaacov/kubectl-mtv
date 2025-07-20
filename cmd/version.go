package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// Version is set via ldflags during build
	clientVersion = "unknown"
)

// getInventoryInfo returns information about the MTV inventory service
func getInventoryInfo() (string, string) {
	config := GetGlobalConfig()
	namespace := client.ResolveNamespace(config.KubeConfigFlags)

	// Try to discover inventory URL
	inventoryURL := client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
	if inventoryURL != "" {
		return inventoryURL, "available"
	}

	return "not found", "not available"
}

// getMTVControllerInfo returns information about the MTV Operator
func getMTVControllerInfo() (string, string, string) {
	config := GetGlobalConfig()

	// Try to get dynamic client
	dynamicClient, err := client.GetDynamicClient(config.KubeConfigFlags)
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

// newVersionCmd creates a new version command
func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  "Print the version information for kubectl-mtv and MTV Operator",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("kubectl-mtv version: %s\n", clientVersion)

			// Get MTV Operator information
			controllerVersion, controllerStatus, controllerNamespace := getMTVControllerInfo()
			fmt.Printf("Operator version: %s\n", controllerVersion)
			fmt.Printf("Operator status: %s\n", controllerStatus)
			if controllerNamespace != "" {
				fmt.Printf("Operator namespace: %s\n", controllerNamespace)
			}

			// Get inventory information
			inventoryURL, inventoryStatus := getInventoryInfo()
			fmt.Printf("Inventory URL: %s\n", inventoryURL)
			fmt.Printf("Inventory status: %s\n", inventoryStatus)
		},
	}

	return cmd
}
