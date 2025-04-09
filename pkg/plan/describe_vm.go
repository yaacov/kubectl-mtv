package plan

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/plan/status"
	"github.com/yaacov/kubectl-mtv/pkg/watch"
)

// DescribeVM describes a specific VM in a migration plan
func DescribeVM(configFlags *genericclioptions.ConfigFlags, name, namespace, vmName string, watchMode bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return describeVMOnce(configFlags, name, namespace, vmName)
		}, 20*time.Second)
	}

	return describeVMOnce(configFlags, name, namespace, vmName)
}

// Helper function to truncate strings to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	// Reserve 3 characters for the ellipsis
	return s[:maxLen-3] + "..."
}

func describeVMOnce(configFlags *genericclioptions.ConfigFlags, name, namespace, vmName string) error {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Get the plan
	plan, err := c.Resource(client.PlansGVR).Namespace(namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get plan: %v", err)
	}

	// First check if VM exists in plan spec
	specVMs, exists, err := unstructured.NestedSlice(plan.Object, "spec", "vms")
	if err != nil || !exists {
		fmt.Printf("No VMs found in plan '%s' specification\n", name)
		return nil
	}

	// Find VM ID from spec
	var vmID string
	for _, v := range specVMs {
		vm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		currentVMName, _, _ := unstructured.NestedString(vm, "name")
		if currentVMName == vmName {
			vmID, _, _ = unstructured.NestedString(vm, "id")
			break
		}
	}

	if vmID == "" {
		fmt.Printf("VM '%s' is not part of plan '%s'\n", vmName, name)
		return nil
	}

	// Get plan details
	planDetails, _ := status.GetPlanDetails(c, namespace, plan, client.MigrationsGVR)

	// Get migration object to display VM details
	migration := planDetails.RunningMigration
	if migration == nil {
		migration = planDetails.LatestMigration
	}
	if migration == nil {
		fmt.Printf("No migration found for plan '%s'. VM details will be available after the plan starts running.\n", name)
		return nil
	}

	// Get VMs list from migration status
	vms, exists, err := unstructured.NestedSlice(migration.Object, "status", "vms")
	if err != nil {
		return fmt.Errorf("failed to get VM list: %v", err)
	}
	if !exists {
		fmt.Printf("No VM status information found in migration. Please wait for the migration to start.\n")
		return nil
	}

	// Find the specified VM using vmID
	var targetVM map[string]interface{}
	for _, v := range vms {
		vm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		currentVMID, _, _ := unstructured.NestedString(vm, "id")
		if currentVMID == vmID {
			targetVM = vm
			break
		}
	}

	if targetVM == nil {
		fmt.Printf("VM '%s' (ID: %s) status not yet available in migration\n", vmName, vmID)
		return nil
	}

	// Print VM details
	fmt.Printf("VM Details for: %s\n", vmName)
	fmt.Printf("Migration Plan: %s\n", name)
	fmt.Printf("Migration: %s\n", migration.GetName())
	fmt.Println("\n----------------------------------------")

	// Print basic VM information
	vmID, _, _ = unstructured.NestedString(targetVM, "id")
	vmPhase, _, _ := unstructured.NestedString(targetVM, "phase")
	vmOS, _, _ := unstructured.NestedString(targetVM, "operatingSystem")
	started, _, _ := unstructured.NestedString(targetVM, "started")
	completed, _, _ := unstructured.NestedString(targetVM, "completed")
	newName, _, _ := unstructured.NestedString(targetVM, "newName")

	fmt.Printf("ID: %s\n", vmID)
	fmt.Printf("Phase: %s\n", vmPhase)
	fmt.Printf("OS: %s\n", vmOS)
	if newName != "" {
		fmt.Printf("New Name: %s\n", newName)
	}
	if started != "" {
		fmt.Printf("Started: %s\n", formatTime(started))
	}
	if completed != "" {
		fmt.Printf("Completed: %s\n", formatTime(completed))
	}

	// Print conditions
	conditions, exists, _ := unstructured.NestedSlice(targetVM, "conditions")
	if exists && len(conditions) > 0 {

		fmt.Print("\n=============================================================================================================")

		fmt.Printf("\nConditions:\n")
		headers := []string{"TYPE", "STATUS", "CATEGORY", "MESSAGE"}
		colWidths := []int{15, 10, 15, 50}
		rows := make([][]string, 0, len(conditions))

		for _, c := range conditions {
			condition, ok := c.(map[string]interface{})
			if !ok {
				continue
			}

			condType, _, _ := unstructured.NestedString(condition, "type")
			status, _, _ := unstructured.NestedString(condition, "status")
			category, _, _ := unstructured.NestedString(condition, "category")
			message, _, _ := unstructured.NestedString(condition, "message")

			rows = append(rows, []string{condType, status, category, message})
		}

		if len(rows) > 0 {
			printTable(headers, rows, colWidths)
		}
	}

	// Print pipeline information
	pipeline, exists, _ := unstructured.NestedSlice(targetVM, "pipeline")
	if exists {
		fmt.Print("\n=============================================================================================================")

		fmt.Printf("\nPipeline:\n")
		for _, p := range pipeline {
			phase, ok := p.(map[string]interface{})
			if !ok {
				continue
			}

			phaseName, _, _ := unstructured.NestedString(phase, "name")
			phaseDesc, _, _ := unstructured.NestedString(phase, "description")
			phaseStatus, _, _ := unstructured.NestedString(phase, "phase")
			phaseStarted, _, _ := unstructured.NestedString(phase, "started")
			phaseCompleted, _, _ := unstructured.NestedString(phase, "completed")

			fmt.Printf("\n[%s] %s\n", phaseName, phaseDesc)
			fmt.Printf("Status: %s\n", phaseStatus)
			fmt.Printf("Started: %s\n", formatTime(phaseStarted))
			if phaseCompleted != "" {
				fmt.Printf("Completed: %s\n", formatTime(phaseCompleted))
			}

			// Print progress
			progressMap, exists, _ := unstructured.NestedMap(phase, "progress")
			if exists {
				completed, _, _ := unstructured.NestedInt64(progressMap, "completed")
				total, _, _ := unstructured.NestedInt64(progressMap, "total")
				if total > 0 {
					percentage := float64(completed) / float64(total) * 100
					fmt.Printf("Progress: %.1f%% (%d/%d)\n", percentage, completed, total)
				}
			}

			// Print tasks if they exist
			tasks, exists, _ := unstructured.NestedSlice(phase, "tasks")
			if exists && len(tasks) > 0 {
				fmt.Printf("\nTasks:\n")
				headers := []string{"NAME", "PHASE", "PROGRESS", "STARTED", "COMPLETED"}
				colWidths := []int{40, 10, 15, 20, 20}
				rows := make([][]string, 0, len(tasks))

				for _, t := range tasks {
					task, ok := t.(map[string]interface{})
					if !ok {
						continue
					}

					taskName, _, _ := unstructured.NestedString(task, "name")
					// Truncate task name if longer than column width
					taskName = truncateString(taskName, colWidths[0])

					taskPhase, _, _ := unstructured.NestedString(task, "phase")
					taskStarted, _, _ := unstructured.NestedString(task, "started")
					taskCompleted, _, _ := unstructured.NestedString(task, "completed")

					progress := "-"
					progressMap, exists, _ := unstructured.NestedMap(task, "progress")
					if exists {
						completed, _, _ := unstructured.NestedInt64(progressMap, "completed")
						total, _, _ := unstructured.NestedInt64(progressMap, "total")
						if total > 0 {
							percentage := float64(completed) / float64(total) * 100
							progress = fmt.Sprintf("%.1f%%", percentage)
						}
					}

					rows = append(rows, []string{
						taskName,
						taskPhase,
						progress,
						formatTime(taskStarted),
						formatTime(taskCompleted),
					})
				}

				printTable(headers, rows, colWidths)
			}
		}
	}

	return nil
}
