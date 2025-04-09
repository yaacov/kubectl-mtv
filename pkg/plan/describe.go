package plan

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/output"
	"github.com/yaacov/kubectl-mtv/pkg/plan/status"
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
	fmt.Printf("\n%s\n", output.Bold("MIGRATION PLAN"))

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

	// Spec Fields
	source, _, _ := unstructured.NestedString(plan.Object, "spec", "provider", "source", "name")
	target, _, _ := unstructured.NestedString(plan.Object, "spec", "provider", "destination", "name")
	targetNamespace, _, _ := unstructured.NestedString(plan.Object, "spec", "targetNamespace")
	warm, _, _ := unstructured.NestedBool(plan.Object, "spec", "warm")
	transferNetwork, _, _ := unstructured.NestedString(plan.Object, "spec", "transferNetwork", "name")
	description, _, _ := unstructured.NestedString(plan.Object, "spec", "description")
	preserveCPUModel, _, _ := unstructured.NestedBool(plan.Object, "spec", "preserveClusterCPUModel")
	preserveStaticIPs, _, _ := unstructured.NestedBool(plan.Object, "spec", "preserveStaticIPs")

	fmt.Printf("\n%s\n", output.Bold("Spec:"))
	fmt.Printf("  Source Provider:   %s\n", source)
	fmt.Printf("  Target Provider:   %s\n", target)
	fmt.Printf("  Target Namespace:  %s\n", targetNamespace)
	fmt.Printf("  Warm:             %t\n", warm)
	fmt.Printf("  Transfer Network: %s\n", transferNetwork)
	fmt.Printf("  Description:      %s\n", description)
	fmt.Printf("  Preserve CPU:     %t\n", preserveCPUModel)
	fmt.Printf("  Preserve IPs:     %t\n", preserveStaticIPs)

	// Mappings
	networkMapping, _, _ := unstructured.NestedString(plan.Object, "spec", "map", "network", "name")
	storageMapping, _, _ := unstructured.NestedString(plan.Object, "spec", "map", "storage", "name")
	fmt.Printf("\n%s\n", output.Bold("Mappings:"))
	fmt.Printf("  Network:          %s\n", networkMapping)
	fmt.Printf("  Storage:          %s\n", storageMapping)

	// Running Migration
	if planDetails.RunningMigration != nil {
		fmt.Printf("\n%s   %s\n", output.Bold("Running Migration:"), output.Yellow(planDetails.RunningMigration.GetName()))
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
		fmt.Printf("\n%s      %s\n", output.Bold("Migration name:"), output.Yellow(planDetails.LatestMigration.GetName()))
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

// printTable prints formatted table with headers and rows
func printTable(headers []string, rows [][]string, colWidths []int) {
	// Print headers
	for i, header := range headers {
		fmt.Printf(fmt.Sprintf("%%-%ds", colWidths[i]), header)
		if i < len(headers)-1 {
			fmt.Print(" ")
		}
	}
	fmt.Println()

	// Print header separators
	for i, width := range colWidths {
		fmt.Print(strings.Repeat("-", width))
		if i < len(colWidths)-1 {
			fmt.Print(" ")
		}
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			// Calculate padding based on visible length
			visLen := output.VisibleLength(cell)
			padding := colWidths[i] - visLen

			// Print the cell with its formatting intact
			fmt.Print(cell)

			// Add necessary padding after the cell
			if padding > 0 {
				fmt.Print(strings.Repeat(" ", padding))
			}

			if i < len(row)-1 {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
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
		headers := []string{"SOURCE ID", "TYPE", "NAME"}
		colWidths := []int{42, 9, 50}
		rows := make([][]string, 0, len(networkPairs))

		for _, pair := range networkPairs {
			if p, ok := pair.(map[string]interface{}); ok {
				sourceID, _, _ := unstructured.NestedString(p, "source", "id")
				destType, _, _ := unstructured.NestedString(p, "destination", "type")
				destName, _, _ := unstructured.NestedString(p, "destination", "name")
				destNS, _, _ := unstructured.NestedString(p, "destination", "namespace")

				destination := getDestinationString(destName, destNS, destType)
				rows = append(rows, []string{sourceID, destType, destination})
			}
		}

		fmt.Println()
		printTable(headers, rows, colWidths)
	}
	return nil
}

// getDestinationString formats the destination string based on the type and namespace
func getDestinationString(name, namespace, destType string) string {
	switch destType {
	case "pod":
		return "Pod Network"
	case "ignored":
		return "Not Migrated"
	default:
		if namespace != "" {
			return fmt.Sprintf("%s/%s", namespace, name)
		}
		return name
	}
}

// displayStorageMapping prints storage mapping details
func displayStorageMapping(c dynamic.Interface, namespace, storageMapping string) error {
	storageMap, err := c.Resource(client.StorageMapGVR).Namespace(namespace).Get(context.TODO(), storageMapping, metav1.GetOptions{})
	if err != nil {
		return err
	}

	storagePairs, exists, _ := unstructured.NestedSlice(storageMap.Object, "spec", "map")
	if exists && len(storagePairs) > 0 {
		headers := []string{"SOURCE ID", "STORAGE CLASS"}
		colWidths := []int{42, 60}
		rows := make([][]string, 0, len(storagePairs))

		for _, pair := range storagePairs {
			if p, ok := pair.(map[string]interface{}); ok {
				sourceID, _, _ := unstructured.NestedString(p, "source", "id")
				destClass, _, _ := unstructured.NestedString(p, "destination", "storageClass")
				rows = append(rows, []string{sourceID, destClass})
			}
		}

		fmt.Println()
		printTable(headers, rows, colWidths)
	}
	return nil
}

// displayConditions prints conditions information
func displayConditions(conditions []interface{}) {
	if len(conditions) > 0 {
		headers := []string{"TYPE", "STATUS", "REASON", "MESSAGE"}
		colWidths := []int{20, 10, 20, 50}
		rows := make([][]string, 0, len(conditions))

		for _, c := range conditions {
			condition, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			condType, _, _ := unstructured.NestedString(condition, "type")
			status, _, _ := unstructured.NestedString(condition, "status")
			reason, _, _ := unstructured.NestedString(condition, "reason")
			message, _, _ := unstructured.NestedString(condition, "message")

			if len(message) > 47 {
				message = message[:47] + "..."
			}
			rows = append(rows, []string{condType, status, reason, message})
		}

		fmt.Printf("\n%s\n", output.Bold("Conditions:"))
		printTable(headers, rows, colWidths)
	}
}
