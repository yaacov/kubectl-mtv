package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// Version is set via ldflags during build
	clientVersion = "unknown"
)

// VersionInfo holds all version-related information
type VersionInfo struct {
	ClientVersion     string `json:"clientVersion" yaml:"clientVersion"`
	OperatorVersion   string `json:"operatorVersion" yaml:"operatorVersion"`
	OperatorStatus    string `json:"operatorStatus" yaml:"operatorStatus"`
	OperatorNamespace string `json:"operatorNamespace,omitempty" yaml:"operatorNamespace,omitempty"`
	InventoryURL      string `json:"inventoryURL" yaml:"inventoryURL"`
	InventoryStatus   string `json:"inventoryStatus" yaml:"inventoryStatus"`
}

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
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  "Print the version information for kubectl-mtv and MTV Operator",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get MTV Operator information
			controllerVersion, controllerStatus, controllerNamespace := getMTVControllerInfo()

			// Get inventory information
			inventoryURL, inventoryStatus := getInventoryInfo()

			// Create version info struct
			versionInfo := VersionInfo{
				ClientVersion:     clientVersion,
				OperatorVersion:   controllerVersion,
				OperatorStatus:    controllerStatus,
				OperatorNamespace: controllerNamespace,
				InventoryURL:      inventoryURL,
				InventoryStatus:   inventoryStatus,
			}

			// Output based on format
			switch outputFormat {
			case "json":
				// Convert to map for JSON output
				versionMap := map[string]interface{}{
					"clientVersion":   versionInfo.ClientVersion,
					"operatorVersion": versionInfo.OperatorVersion,
					"operatorStatus":  versionInfo.OperatorStatus,
					"inventoryURL":    versionInfo.InventoryURL,
					"inventoryStatus": versionInfo.InventoryStatus,
				}
				if versionInfo.OperatorNamespace != "" {
					versionMap["operatorNamespace"] = versionInfo.OperatorNamespace
				}

				jsonBytes, err := json.MarshalIndent(versionMap, "", "  ")
				if err != nil {
					return fmt.Errorf("error marshaling JSON: %w", err)
				}
				fmt.Println(string(jsonBytes))
			case "yaml":
				// Convert to map for YAML output
				versionMap := map[string]interface{}{
					"clientVersion":   versionInfo.ClientVersion,
					"operatorVersion": versionInfo.OperatorVersion,
					"operatorStatus":  versionInfo.OperatorStatus,
					"inventoryURL":    versionInfo.InventoryURL,
					"inventoryStatus": versionInfo.InventoryStatus,
				}
				if versionInfo.OperatorNamespace != "" {
					versionMap["operatorNamespace"] = versionInfo.OperatorNamespace
				}

				yamlBytes, err := yaml.Marshal(versionMap)
				if err != nil {
					return fmt.Errorf("error marshaling YAML: %w", err)
				}
				fmt.Print(string(yamlBytes))
			default:
				// Default table/text output
				fmt.Printf("kubectl-mtv version: %s\n", versionInfo.ClientVersion)
				fmt.Printf("Operator version: %s\n", versionInfo.OperatorVersion)
				fmt.Printf("Operator status: %s\n", versionInfo.OperatorStatus)
				if versionInfo.OperatorNamespace != "" {
					fmt.Printf("Operator namespace: %s\n", versionInfo.OperatorNamespace)
				}
				fmt.Printf("Inventory URL: %s\n", versionInfo.InventoryURL)
				fmt.Printf("Inventory status: %s\n", versionInfo.InventoryStatus)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format: json, yaml, or table (default)")

	return cmd
}
