package plan

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/get/plan/status"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/output"
	"github.com/yaacov/kubectl-mtv/pkg/util/watch"
)

// formatDiskSize formats bytes as human-readable size
func formatDiskSize(bytes int64) string {
	const gb = 1024 * 1024 * 1024
	if bytes <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
}

// ListDisks lists all disk transfers in a migration plan
func ListDisks(ctx context.Context, configFlags *genericclioptions.ConfigFlags, name, namespace string, watchMode bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return listDisksOnce(ctx, configFlags, name, namespace)
		}, watch.DefaultInterval)
	}

	return listDisksOnce(ctx, configFlags, name, namespace)
}

// listDisksOnce lists disk transfers in a migration plan once (helper function for ListDisks)
func listDisksOnce(ctx context.Context, configFlags *genericclioptions.ConfigFlags, name, namespace string) error {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Get the plan
	plan, err := c.Resource(client.PlansGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
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
		fmt.Printf("%s %s\n\n", output.Bold("Disk transfers in migration plan:"), output.Yellow(name))
		fmt.Println("No migration information found. Disk transfer details will be available after the plan starts running.")

		// Print VMs from plan spec
		specVMs, exists, err := unstructured.NestedSlice(plan.Object, "spec", "vms")
		if err == nil && exists && len(specVMs) > 0 {
			fmt.Printf("\n%s\n", output.Bold("Plan VM Specifications:"))
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
				rows = append(rows, []string{output.Yellow(vmName), output.Cyan(vmID)})
			}

			PrintTable(headers, rows, colWidths)
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

	fmt.Print("\n", output.ColorizedSeparator(105, output.YellowColor))
	fmt.Printf("\n%s\n", output.Bold("MIGRATION PLAN - DISK TRANSFERS"))

	fmt.Printf("%s %s\n", output.Bold("Plan:"), output.Yellow(name))
	fmt.Printf("%s %s\n", output.Bold("Migration:"), output.Yellow(migration.GetName()))

	// Print VM and disk information
	for _, v := range vms {
		vm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		vmName, _, _ := unstructured.NestedString(vm, "name")
		vmID, _, _ := unstructured.NestedString(vm, "id")
		vmPhase, _, _ := unstructured.NestedString(vm, "phase")

		vmCompletionStatus := getVMCompletionStatus(vm)

		fmt.Print("\n", output.ColorizedSeparator(105, output.CyanColor))
		fmt.Printf("\n%s %s (%s %s)\n", output.Bold("VM:"), output.Yellow(vmName), output.Bold("vmID="), output.Cyan(vmID))
		fmt.Printf("%s %s  %s %s\n", output.Bold("Phase:"), output.ColorizeStatus(vmPhase), output.Bold("Status:"), output.ColorizeStatus(vmCompletionStatus))

		// Find and display disk transfer information
		pipeline, exists, _ := unstructured.NestedSlice(vm, "pipeline")
		if !exists || len(pipeline) == 0 {
			fmt.Println("No pipeline information available.")
			continue
		}

		// Find DiskTransfer phases and extract tasks
		hasDiskTransfers := false
		for _, p := range pipeline {
			phase, ok := p.(map[string]interface{})
			if !ok {
				continue
			}

			phaseName, _, _ := unstructured.NestedString(phase, "name")
			// Check if this is a disk transfer phase
			if !strings.HasPrefix(phaseName, "DiskTransfer") {
				continue
			}

			hasDiskTransfers = true
			phaseStatus, _, _ := unstructured.NestedString(phase, "phase")
			phaseStarted, _, _ := unstructured.NestedString(phase, "started")
			phaseCompleted, _, _ := unstructured.NestedString(phase, "completed")

			// Get phase-level progress
			progressMap, progressExists, _ := unstructured.NestedMap(phase, "progress")
			phaseProgress := "-"
			if phaseStatus == status.StatusCompleted {
				phaseProgress = output.Green("100.0%")
			} else if progressExists {
				progCompleted, _, _ := unstructured.NestedInt64(progressMap, "completed")
				progTotal, _, _ := unstructured.NestedInt64(progressMap, "total")
				if progTotal > 0 {
					percentage := float64(progCompleted) / float64(progTotal) * 100
					if percentage > 100.0 {
						percentage = 100.0
					}
					phaseProgress = output.Cyan(fmt.Sprintf("%.1f%%", percentage))
				}
			}

			fmt.Printf("\n%s %s  %s %s  %s %s\n",
				output.Bold("Phase:"), output.Yellow(phaseName),
				output.Bold("Status:"), output.ColorizeStatus(phaseStatus),
				output.Bold("Progress:"), phaseProgress)

			if phaseStarted != "" {
				fmt.Printf("%s %s", output.Bold("Started:"), phaseStarted)
				if phaseCompleted != "" {
					fmt.Printf("  %s %s", output.Bold("Completed:"), phaseCompleted)
				}
				fmt.Println()
			}

			// Extract and display individual disk tasks
			tasks, tasksExist, _ := unstructured.NestedSlice(phase, "tasks")
			if tasksExist && len(tasks) > 0 {
				fmt.Printf("\n%s\n", output.Bold("Disk Transfers:"))
				headers := []string{"NAME", "PHASE", "PROGRESS", "SIZE", "STARTED", "COMPLETED"}
				colWidths := []int{35, 12, 12, 12, 22, 22}
				rows := make([][]string, 0, len(tasks))

				for _, t := range tasks {
					task, ok := t.(map[string]interface{})
					if !ok {
						continue
					}

					taskName, _, _ := unstructured.NestedString(task, "name")
					taskPhase, _, _ := unstructured.NestedString(task, "phase")
					taskStarted, _, _ := unstructured.NestedString(task, "started")
					taskCompleted, _, _ := unstructured.NestedString(task, "completed")

					// Truncate task name if too long
					if len(taskName) > colWidths[0] {
						taskName = taskName[:colWidths[0]-3] + "..."
					}

					// Get task progress
					progress := "-"
					size := "-"
					taskProgressMap, taskProgressExists, _ := unstructured.NestedMap(task, "progress")
					if taskProgressExists {
						taskCompleted, _, _ := unstructured.NestedInt64(taskProgressMap, "completed")
						taskTotal, _, _ := unstructured.NestedInt64(taskProgressMap, "total")
						if taskTotal > 0 {
							percentage := float64(taskCompleted) / float64(taskTotal) * 100
							if percentage > 100.0 {
								percentage = 100.0
							}
							progressText := fmt.Sprintf("%.1f%%", percentage)

							// Color based on status
							switch vmCompletionStatus {
							case status.StatusFailed:
								progress = output.Red(progressText)
							case status.StatusCanceled:
								progress = output.Yellow(progressText)
							case status.StatusSucceeded, status.StatusCompleted:
								progress = output.Green(progressText)
							default:
								if percentage >= 100 {
									progress = output.Green(progressText)
								} else {
									progress = output.Cyan(progressText)
								}
							}
							size = formatDiskSize(taskTotal)
						}
					} else if taskPhase == status.StatusCompleted {
						progress = output.Green("100.0%")
					}

					// Format timestamps (show "-" if empty)
					startedStr := taskStarted
					if startedStr == "" {
						startedStr = "-"
					}
					completedStr := taskCompleted
					if completedStr == "" {
						completedStr = "-"
					}

					rows = append(rows, []string{
						taskName,
						output.ColorizeStatus(taskPhase),
						progress,
						size,
						startedStr,
						completedStr,
					})
				}

				PrintTable(headers, rows, colWidths)
			} else {
				fmt.Println("No individual disk transfer tasks found.")
			}
		}

		if !hasDiskTransfers {
			// Show pipeline summary if no disk transfer phases found yet
			fmt.Printf("\n%s\n", output.Bold("Pipeline Summary:"))
			headers := []string{"PHASE", "NAME", "STARTED", "COMPLETED", "PROGRESS"}
			colWidths := []int{12, 25, 22, 22, 12}
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

				progressMap, progressExists, _ := unstructured.NestedMap(phase, "progress")
				if phaseStatus == status.StatusCompleted {
					progress = output.Green("100.0%")
				} else if progressExists {
					progCompleted, _, _ := unstructured.NestedInt64(progressMap, "completed")
					progTotal, _, _ := unstructured.NestedInt64(progressMap, "total")
					if progTotal > 0 {
						percentage := float64(progCompleted) / float64(progTotal) * 100
						if percentage > 100.0 {
							percentage = 100.0
						}
						progress = output.Cyan(fmt.Sprintf("%.1f%%", percentage))
					}
				}

				startedStr := phaseStarted
				if startedStr == "" {
					startedStr = "-"
				}
				completedStr := phaseCompleted
				if completedStr == "" {
					completedStr = "-"
				}

				rows = append(rows, []string{
					output.ColorizeStatus(phaseStatus),
					output.Bold(phaseName),
					startedStr,
					completedStr,
					progress,
				})
			}

			PrintTable(headers, rows, colWidths)
		}
	}

	return nil
}
