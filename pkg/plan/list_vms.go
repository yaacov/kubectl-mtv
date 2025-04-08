package plan

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/plan/status"
	"github.com/yaacov/kubectl-mtv/pkg/watch"
)

// ListVMs lists all VMs in a migration plan
func ListVMs(configFlags *genericclioptions.ConfigFlags, name, namespace string, watchMode bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return listVMsOnce(configFlags, name, namespace)
		}, 20*time.Second)
	}

	return listVMsOnce(configFlags, name, namespace)
}

// listVMsOnce lists VMs in a migration plan once (helper function for ListVMs)
func listVMsOnce(configFlags *genericclioptions.ConfigFlags, name, namespace string) error {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Get the plan
	plan, err := c.Resource(client.PlansGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get plan: %v", err)
	}

	// Get plan details
	planDetails, _ := status.GetPlanDetails(c, namespace, plan, client.MigrationsGVR)

	// Get migration object to display VM details
	migration := planDetails.RunningMigration
	if migration == nil {
		migration = planDetails.LatestMigration
	}
	if migration == nil {
		fmt.Printf("VMs in migration plan: %s\n\n", name)
		fmt.Println("No migration information found. VM details will be available after the plan starts running.")

		// Print VMs from plan spec
		specVMs, exists, err := unstructured.NestedSlice(plan.Object, "spec", "vms")
		if err == nil && exists && len(specVMs) > 0 {
			fmt.Printf("\nPlan VM Specifications:\n")
			headers := []string{"NAME", "ID"}
			colWidths := []int{40, 20}
			rows := make([][]string, 0, len(specVMs))

			for _, v := range specVMs {
				vm, ok := v.(map[string]interface{})
				if !ok {
					continue
				}

				vmName, _, _ := unstructured.NestedString(vm, "name")
				vmID, _, _ := unstructured.NestedString(vm, "id")
				rows = append(rows, []string{vmName, vmID})
			}

			printTable(headers, rows, colWidths)
		}

		return nil
	}

	// Get VMs list from migration status
	vms, exists, err := unstructured.NestedSlice(migration.Object, "status", "vms")
	if err != nil {
		return fmt.Errorf("failed to get VM list: %v", err)
	}
	if !exists {
		return fmt.Errorf("no VMs found in migration status")
	}

	fmt.Print("\n-------------------------------------------------------------------------------------------------------------\n")
	fmt.Printf("VMs in migration plan: %s\n", name)
	fmt.Printf("Migration: %s\n", migration.GetName())

	// Print VM information
	for _, v := range vms {
		vm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		fmt.Print("\n=============================================================================================================\n")

		vmName, _, _ := unstructured.NestedString(vm, "name")
		vmID, _, _ := unstructured.NestedString(vm, "id")
		vmPhase, _, _ := unstructured.NestedString(vm, "phase")
		vmOS, _, _ := unstructured.NestedString(vm, "operatingSystem")
		started, _, _ := unstructured.NestedString(vm, "started")
		completed, _, _ := unstructured.NestedString(vm, "completed")

		fmt.Printf("VM: %s (ID: %s)\n", vmName, vmID)
		fmt.Printf("Phase: %s\n", vmPhase)
		fmt.Printf("OS: %s\n", vmOS)
		if started != "" {
			fmt.Printf("Started: %s\n", started)
		}
		if completed != "" {
			fmt.Printf("Completed: %s\n", completed)
		}

		// Print pipeline information
		pipeline, exists, _ := unstructured.NestedSlice(vm, "pipeline")
		if exists && len(pipeline) > 0 {
			fmt.Printf("\nPipeline:\n")
			headers := []string{"PHASE", "NAME", "STARTED", "COMPLETED", "PROGRESS"}
			colWidths := []int{15, 25, 25, 25, 15}
			rows := make([][]string, 0, len(pipeline))

			for _, p := range pipeline {
				phase, ok := p.(map[string]interface{})
				if !ok {
					continue
				}

				phaseName, _, _ := unstructured.NestedString(phase, "name")
				phaseStatus, _, _ := unstructured.NestedString(phase, "phase")
				phaseStarted, _, _ := unstructured.NestedString(phase, "started")
				phaseCompleted, _, _ := unstructured.NestedString(phase, "completed")

				progress := "-"
				if vmPhase == "Completed" || phaseStatus == "Completed" {
					progress = fmt.Sprintf("%14.1f%%", 100.0)
				} else {
					progressMap, exists, _ := unstructured.NestedMap(phase, "progress")
					if exists {
						completed, _, _ := unstructured.NestedInt64(progressMap, "completed")
						total, _, _ := unstructured.NestedInt64(progressMap, "total")
						if total > 0 {
							percentage := float64(completed) / float64(total) * 100
							progress = fmt.Sprintf("%14.1f%%", percentage)
						}
					}
				}

				rows = append(rows, []string{
					phaseStatus,
					phaseName,
					formatTime(phaseStarted),
					formatTime(phaseCompleted),
					progress,
				})
			}

			printTable(headers, rows, colWidths)
		}
	}

	return nil
}

func formatTime(timestamp string) string {
	if timestamp == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}
	return t.Format("2006-01-02 15:04:05")
}
