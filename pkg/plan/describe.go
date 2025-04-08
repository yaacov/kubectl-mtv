package plan

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
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

	// Basic Information
	fmt.Printf("Name:              %s\n", plan.GetName())
	fmt.Printf("Namespace:         %s\n", plan.GetNamespace())
	fmt.Printf("Created:           %s\n", plan.GetCreationTimestamp())

	// Plan Details
	planDetails, _ := status.GetPlanDetails(c, namespace, plan, client.MigrationsGVR)
	fmt.Printf("Ready:             %t\n", planDetails.IsReady)
	fmt.Printf("Status:            %s\n", planDetails.Status)

	// Spec Fields
	source, _, _ := unstructured.NestedString(plan.Object, "spec", "provider", "source", "name")
	target, _, _ := unstructured.NestedString(plan.Object, "spec", "provider", "destination", "name")
	targetNamespace, _, _ := unstructured.NestedString(plan.Object, "spec", "targetNamespace")
	warm, _, _ := unstructured.NestedBool(plan.Object, "spec", "warm")
	transferNetwork, _, _ := unstructured.NestedString(plan.Object, "spec", "transferNetwork", "name")
	description, _, _ := unstructured.NestedString(plan.Object, "spec", "description")
	preserveCPUModel, _, _ := unstructured.NestedBool(plan.Object, "spec", "preserveClusterCPUModel")
	preserveStaticIPs, _, _ := unstructured.NestedBool(plan.Object, "spec", "preserveStaticIPs")

	fmt.Printf("\nSpec:\n")
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
	fmt.Printf("\nMappings:\n")
	fmt.Printf("  Network:          %s\n", networkMapping)
	fmt.Printf("  Storage:          %s\n", storageMapping)

	// Running Migration
	if planDetails.RunningMigration != nil {
		fmt.Printf("\nRunning Migration:   %s\n", planDetails.RunningMigration.GetName())
		fmt.Printf("Migration Progress:  Total:     %3d, Completed: %3d\n", planDetails.VMStats.Completed, planDetails.VMStats.Total)
		fmt.Printf("VM Status:           Succeeded: %3d, Failed:    %3d, Canceled: %3d\n",
			planDetails.VMStats.Succeeded, planDetails.VMStats.Failed, planDetails.VMStats.Canceled)
		if planDetails.DiskProgress.Total > 0 {
			percentage := float64(planDetails.DiskProgress.Completed) / float64(planDetails.DiskProgress.Total) * 100
			fmt.Printf("Disk Transfer:       %.1f%% (%d/%d GB)\n",
				percentage,
				planDetails.DiskProgress.Completed/(1024),
				planDetails.DiskProgress.Total/(1024))
		}
	}

	if planDetails.LatestMigration != nil {
		fmt.Printf("\nMigration name:      %s\n", planDetails.LatestMigration.GetName())
		fmt.Printf("Migration Progress:  Total:     %3d, Completed: %3d\n", planDetails.VMStats.Completed, planDetails.VMStats.Total)
		fmt.Printf("VM Status:           Succeeded: %3d, Failed:    %3d, Canceled: %3d\n",
			planDetails.VMStats.Succeeded, planDetails.VMStats.Failed, planDetails.VMStats.Canceled)
		if planDetails.DiskProgress.Total > 0 {
			percentage := float64(planDetails.DiskProgress.Completed) / float64(planDetails.DiskProgress.Total) * 100
			fmt.Printf("Disk Transfer:       %.1f%% (%d/%d GB)\n",
				percentage,
				planDetails.DiskProgress.Completed/(1024),
				planDetails.DiskProgress.Total/(1024))
		}
	}

	// Conditions
	conditions, exists, _ := unstructured.NestedSlice(plan.Object, "status", "conditions")
	if exists && len(conditions) > 0 {
		fmt.Printf("\nConditions:\n")
		fmt.Printf("%-20s %-10s %-20s %s\n", "TYPE", "STATUS", "REASON", "MESSAGE")
		fmt.Printf("%-20s %-10s %-20s %s\n", "----", "------", "------", "-------")
		for _, c := range conditions {
			condition, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			condType, _, _ := unstructured.NestedString(condition, "type")
			status, _, _ := unstructured.NestedString(condition, "status")
			reason, _, _ := unstructured.NestedString(condition, "reason")
			message, _, _ := unstructured.NestedString(condition, "message")

			// Truncate message if too long
			if len(message) > 50 {
				message = message[:47] + "..."
			}

			fmt.Printf("%-20s %-10s %-20s %s\n", condType, status, reason, message)
		}
	}

	return nil
}
