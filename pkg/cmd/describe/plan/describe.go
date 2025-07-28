package plan

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/get/plan/status"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/output"
)

// Describe describes a migration plan
func Describe(configFlags *genericclioptions.ConfigFlags, name, namespace string) error {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Get the plan
	plan, err := c.Resource(client.PlansGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get plan: %v", err)
	}

	// Print the plan details
	fmt.Printf("\n%s", output.ColorizedSeparator(105, output.YellowColor))
	fmt.Printf("\n%s\n", output.Cyan("MIGRATION PLAN"))

	// Basic Information
	fmt.Printf("%s %s\n", output.Bold("Name:"), output.Yellow(plan.GetName()))
	fmt.Printf("%s %s\n", output.Bold("Namespace:"), output.Yellow(plan.GetNamespace()))
	fmt.Printf("%s %s\n", output.Bold("Created:"), output.Yellow(plan.GetCreationTimestamp().String()))

	// Get archived status
	archived, exists, _ := unstructured.NestedBool(plan.Object, "spec", "archived")
	if exists {
		fmt.Printf("%s %s\n", output.Bold("Archived:"), output.Yellow(fmt.Sprintf("%t", archived)))
	} else {
		fmt.Printf("%s %s\n", output.Bold("Archived:"), output.Yellow("false"))
	}

	// Plan Details
	planDetails, _ := status.GetPlanDetails(c, namespace, plan, client.MigrationsGVR)
	fmt.Printf("%s %s\n", output.Bold("Ready:"), output.ColorizeBoolean(planDetails.IsReady))
	fmt.Printf("%s %s\n", output.Bold("Status:"), output.ColorizeStatus(planDetails.Status))

	// Display enhanced spec section
	displayPlanSpec(plan)

	// Display enhanced mappings section
	networkMapping, _, _ := unstructured.NestedString(plan.Object, "spec", "map", "network", "name")
	storageMapping, _, _ := unstructured.NestedString(plan.Object, "spec", "map", "storage", "name")
	displayPlanMappings(networkMapping, storageMapping)

	// Running Migration
	if planDetails.RunningMigration != nil {
		fmt.Printf("\n%s\n", output.Cyan("RUNNING MIGRATION"))
		fmt.Printf("%s %s\n", output.Bold("Name:"), output.Yellow(planDetails.RunningMigration.GetName()))
		fmt.Printf("%s  Total:     %s, Completed: %s\n",
			output.Bold("Migration Progress:"),
			output.Blue(fmt.Sprintf("%3d", planDetails.VMStats.Total)),
			output.Blue(fmt.Sprintf("%3d", planDetails.VMStats.Completed)))
		fmt.Printf("%s Succeeded: %s, Failed:    %s, Canceled:  %s\n",
			output.Bold("VM Status:          "),
			output.Green(fmt.Sprintf("%3d", planDetails.VMStats.Succeeded)),
			output.Red(fmt.Sprintf("%3d", planDetails.VMStats.Failed)),
			output.Yellow(fmt.Sprintf("%3d", planDetails.VMStats.Canceled)))
		printDiskProgress(planDetails.DiskProgress)
	}

	// Latest Migration
	if planDetails.LatestMigration != nil {
		fmt.Printf("\n%s\n", output.Cyan("LATEST MIGRATION"))
		fmt.Printf("%s %s\n", output.Bold("Name:"), output.Yellow(planDetails.LatestMigration.GetName()))
		fmt.Printf("%s  Total:     %s, Completed: %s\n",
			output.Bold("Migration Progress:"),
			output.Blue(fmt.Sprintf("%3d", planDetails.VMStats.Total)),
			output.Blue(fmt.Sprintf("%3d", planDetails.VMStats.Completed)))
		fmt.Printf("%s Succeeded: %s, Failed:    %s, Canceled:  %s\n",
			output.Bold("VM Status:          "),
			output.Green(fmt.Sprintf("%3d", planDetails.VMStats.Succeeded)),
			output.Red(fmt.Sprintf("%3d", planDetails.VMStats.Failed)),
			output.Yellow(fmt.Sprintf("%3d", planDetails.VMStats.Canceled)))
		printDiskProgress(planDetails.DiskProgress)
	}

	// Display network mapping
	if networkMapping != "" {
		if err := displayNetworkMapping(c, namespace, networkMapping); err != nil {
			fmt.Printf("Failed to display network mapping: %v\n", err)
		}
	}

	// Display storage mapping
	if storageMapping != "" {
		if err := displayStorageMapping(c, namespace, storageMapping); err != nil {
			fmt.Printf("Failed to display storage mapping: %v\n", err)
		}
	}

	// Display conditions
	conditions, exists, _ := unstructured.NestedSlice(plan.Object, "status", "conditions")
	if exists {
		displayConditions(conditions)
	}

	return nil
}

// printDiskProgress prints disk transfer progress information
func printDiskProgress(progress status.ProgressStats) {
	if progress.Total > 0 {
		percentage := float64(progress.Completed) / float64(progress.Total) * 100
		progressText := fmt.Sprintf("%.1f%% (%d/%d GB)",
			percentage,
			progress.Completed/(1024),
			progress.Total/(1024))

		if percentage >= 100 {
			fmt.Printf("%s       %s\n", output.Bold("Disk Transfer:"), output.Green(progressText))
		} else {
			fmt.Printf("%s       %s\n", output.Bold("Disk Transfer:"), output.Yellow(progressText))
		}
	}
}

// displayNetworkMapping prints network mapping details
func displayNetworkMapping(c dynamic.Interface, namespace, networkMapping string) error {
	networkMap, err := c.Resource(client.NetworkMapGVR).Namespace(namespace).Get(context.TODO(), networkMapping, metav1.GetOptions{})
	if err != nil {
		return err
	}

	networkPairs, exists, _ := unstructured.NestedSlice(networkMap.Object, "spec", "map")
	if exists && len(networkPairs) > 0 {
		fmt.Printf("\n%s\n", output.Cyan("NETWORK MAPPING DETAILS"))
		return output.PrintMappingTable(networkPairs, formatPlanMappingEntry)
	}
	return nil
}

// displayStorageMapping prints storage mapping details
func displayStorageMapping(c dynamic.Interface, namespace, storageMapping string) error {
	storageMap, err := c.Resource(client.StorageMapGVR).Namespace(namespace).Get(context.TODO(), storageMapping, metav1.GetOptions{})
	if err != nil {
		return err
	}

	storagePairs, exists, _ := unstructured.NestedSlice(storageMap.Object, "spec", "map")
	if exists && len(storagePairs) > 0 {
		fmt.Printf("\n%s\n", output.Cyan("STORAGE MAPPING DETAILS"))
		return output.PrintMappingTable(storagePairs, formatPlanMappingEntry)
	}
	return nil
}

// displayConditions prints conditions information using shared formatting
func displayConditions(conditions []interface{}) {
	if len(conditions) > 0 {
		fmt.Printf("\n%s\n", output.Cyan("STATUS"))
		output.PrintConditions(conditions)
	}
}

// displayPlanSpec displays the plan specification in a beautified format
func displayPlanSpec(plan *unstructured.Unstructured) {
	source, _, _ := unstructured.NestedString(plan.Object, "spec", "provider", "source", "name")
	target, _, _ := unstructured.NestedString(plan.Object, "spec", "provider", "destination", "name")
	targetNamespace, _, _ := unstructured.NestedString(plan.Object, "spec", "targetNamespace")
	warm, _, _ := unstructured.NestedBool(plan.Object, "spec", "warm")
	transferNetwork, _, _ := unstructured.NestedString(plan.Object, "spec", "transferNetwork", "name")
	description, _, _ := unstructured.NestedString(plan.Object, "spec", "description")
	preserveCPUModel, _, _ := unstructured.NestedBool(plan.Object, "spec", "preserveClusterCPUModel")
	preserveStaticIPs, _, _ := unstructured.NestedBool(plan.Object, "spec", "preserveStaticIPs")

	fmt.Printf("\n%s\n", output.Cyan("SPECIFICATION"))

	// Provider section
	fmt.Printf("%s\n", output.Bold("Providers:"))
	fmt.Printf("  %s %s\n", output.Bold("Source:"), output.Yellow(source))
	fmt.Printf("  %s %s\n", output.Bold("Target:"), output.Yellow(target))

	// Migration settings
	fmt.Printf("\n%s\n", output.Bold("Migration Settings:"))
	fmt.Printf("  %s %s\n", output.Bold("Target Namespace:"), output.Yellow(targetNamespace))
	fmt.Printf("  %s %s\n", output.Bold("Warm Migration:"), output.ColorizeBoolean(warm))
	if transferNetwork != "" {
		fmt.Printf("  %s %s\n", output.Bold("Transfer Network:"), output.Yellow(transferNetwork))
	}

	// Advanced settings
	fmt.Printf("\n%s\n", output.Bold("Advanced Settings:"))
	fmt.Printf("  %s %s\n", output.Bold("Preserve CPU Model:"), output.ColorizeBoolean(preserveCPUModel))
	fmt.Printf("  %s %s\n", output.Bold("Preserve Static IPs:"), output.ColorizeBoolean(preserveStaticIPs))

	// Description
	if description != "" {
		fmt.Printf("\n%s\n", output.Bold("Description:"))
		fmt.Printf("  %s\n", description)
	}
}

// displayPlanMappings displays the mapping references in a beautified format
func displayPlanMappings(networkMapping, storageMapping string) {
	fmt.Printf("\n%s\n", output.Cyan("MAPPINGS"))

	if networkMapping != "" {
		fmt.Printf("%s %s\n", output.Bold("Network Mapping:"), output.Yellow(networkMapping))
	} else {
		fmt.Printf("%s %s\n", output.Bold("Network Mapping:"), output.Red("Not specified"))
	}

	if storageMapping != "" {
		fmt.Printf("%s %s\n", output.Bold("Storage Mapping:"), output.Yellow(storageMapping))
	} else {
		fmt.Printf("%s %s\n", output.Bold("Storage Mapping:"), output.Red("Not specified"))
	}
}

// formatPlanMappingEntry formats a single mapping entry (source or destination) as a string
func formatPlanMappingEntry(entryMap map[string]interface{}, entryType string) string {
	entry, found, _ := unstructured.NestedMap(entryMap, entryType)
	if !found {
		return ""
	}

	var parts []string

	// Common fields that might be present
	if id, ok := entry["id"].(string); ok && id != "" {
		parts = append(parts, fmt.Sprintf("ID: %s", id))
	}

	if name, ok := entry["name"].(string); ok && name != "" {
		parts = append(parts, fmt.Sprintf("Name: %s", name))
	}

	if path, ok := entry["path"].(string); ok && path != "" {
		parts = append(parts, fmt.Sprintf("Path: %s", path))
	}

	// For storage mappings
	if storageClass, ok := entry["storageClass"].(string); ok && storageClass != "" {
		parts = append(parts, fmt.Sprintf("Storage Class: %s", storageClass))
	}

	if accessMode, ok := entry["accessMode"].(string); ok && accessMode != "" {
		parts = append(parts, fmt.Sprintf("Access Mode: %s", accessMode))
	}

	// For network mappings
	if vlan, ok := entry["vlan"].(string); ok && vlan != "" {
		parts = append(parts, fmt.Sprintf("VLAN: %s", vlan))
	}

	if destType, ok := entry["type"].(string); ok && destType != "" {
		parts = append(parts, fmt.Sprintf("Type: %s", destType))
	}

	if namespace, ok := entry["namespace"].(string); ok && namespace != "" {
		parts = append(parts, fmt.Sprintf("Namespace: %s", namespace))
	}

	if multus, found, _ := unstructured.NestedMap(entry, "multus"); found {
		if networkName, ok := multus["networkName"].(string); ok && networkName != "" {
			parts = append(parts, fmt.Sprintf("Multus Network: %s", networkName))
		}
	}

	// Join all parts with newlines for multi-line cell display
	return strings.Join(parts, "\n")
}
