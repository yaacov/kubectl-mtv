package status

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Migration status constants
const (
	StatusRunning   = "Running"
	StatusFailed    = "Failed"
	StatusSucceeded = "Succeeded"
	StatusCanceled  = "Canceled"
	StatusUnknown   = "-"
	StatusExecuting = "Executing"
)

// IsPlanReady checks if a migration plan is ready
func IsPlanReady(plan *unstructured.Unstructured) (bool, error) {
	conditions, exists, err := unstructured.NestedSlice(plan.Object, "status", "conditions")
	if err != nil {
		return false, fmt.Errorf("failed to get plan conditions: %v", err)
	}

	if !exists {
		return false, fmt.Errorf("migration plan status conditions not found")
	}

	for _, c := range conditions {
		condition, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condition, "type")
		condStatus, _, _ := unstructured.NestedString(condition, "status")

		if condType == "Ready" && condStatus == "True" {
			return true, nil
		}
	}

	return false, nil
}

// GetRunningMigration checks if a plan has any running migrations
func GetRunningMigration(
	client dynamic.Interface,
	namespace string,
	plan *unstructured.Unstructured,
	migrationsGVR schema.GroupVersionResource,
) (*unstructured.Unstructured, error) {
	// Get the plan UID
	planUID, found, err := unstructured.NestedString(plan.Object, "metadata", "uid")
	if !found || err != nil {
		return nil, fmt.Errorf("failed to get plan UID: %v", err)
	}

	// Get all migrations in the namespace
	migrationList, err := client.Resource(migrationsGVR).
		Namespace(namespace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list migrations: %v", err)
	}

	// Check each migration
	for _, migration := range migrationList.Items {
		// Check if this migration references our plan
		planRef, found, _ := unstructured.NestedMap(migration.Object, "spec", "plan")
		if !found {
			continue
		}

		refUID, found, _ := unstructured.NestedString(planRef, "uid")
		if !found || refUID != planUID {
			continue
		}

		// Check if the migration is running
		conditions, exists, err := unstructured.NestedSlice(migration.Object, "status", "conditions")
		if err != nil || !exists {
			continue
		}

		for _, c := range conditions {
			condition, ok := c.(map[string]interface{})
			if !ok {
				continue
			}

			condType, _, _ := unstructured.NestedString(condition, "type")
			condStatus, _, _ := unstructured.NestedString(condition, "status")

			if condType == "Running" && condStatus == "True" {
				return &migration, nil
			}
		}
	}

	return nil, nil
}

// GetPlanStatus returns the status of a plan
func GetPlanStatus(plan *unstructured.Unstructured) (string, error) {
	conditions, exists, err := unstructured.NestedSlice(plan.Object, "status", "conditions")
	if err != nil {
		return StatusUnknown, fmt.Errorf("failed to get plan conditions: %v", err)
	}

	if !exists {
		return StatusUnknown, fmt.Errorf("plan status conditions not found")
	}

	// Check all conditions with correct precedence
	for _, c := range conditions {
		condition, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condition, "type")
		condStatus, _, _ := unstructured.NestedString(condition, "status")

		if condStatus != "True" {
			continue
		}

		// Return first condition that matches the current status
		switch condType {
		case StatusFailed:
			return StatusFailed, nil
		case StatusSucceeded:
			return StatusSucceeded, nil
		case StatusCanceled:
			return StatusCanceled, nil
		case StatusRunning:
			return StatusRunning, nil
		case StatusExecuting:
			return StatusExecuting, nil
		}
	}

	return StatusUnknown, nil
}

// VMStats contains statistics about VMs in a migration
type VMStats struct {
	Total     int
	Completed int
	Succeeded int
	Failed    int
	Canceled  int
}

// GetVMStats extracts VM migration statistics from a migration object
func GetVMStats(migration *unstructured.Unstructured) (VMStats, error) {
	stats := VMStats{}

	// Get the list of VMs from the migration status
	vms, exists, err := unstructured.NestedSlice(migration.Object, "status", "vms")
	if err != nil {
		return stats, fmt.Errorf("failed to get VM list: %v", err)
	}

	if !exists || len(vms) == 0 {
		return stats, nil
	}

	// Update total VM count
	stats.Total = len(vms)

	// Check each VM's status
	for _, v := range vms {
		vm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if the VM migration phase is completed
		phase, _, _ := unstructured.NestedString(vm, "phase")
		if phase == "Completed" {
			stats.Completed++

			// Check conditions to determine success/failure/canceled
			conditions, exists, _ := unstructured.NestedSlice(vm, "conditions")
			if !exists {
				continue
			}

			for _, c := range conditions {
				condition, ok := c.(map[string]interface{})
				if !ok {
					continue
				}

				condType, _, _ := unstructured.NestedString(condition, "type")
				condStatus, _, _ := unstructured.NestedString(condition, "status")

				if condStatus == "True" {
					switch condType {
					case StatusSucceeded:
						stats.Succeeded++
					case StatusFailed:
						stats.Failed++
					case StatusCanceled:
						stats.Canceled++
					}
				}
			}
		}
	}

	return stats, nil
}

// ProgressStats contains progress information for disk transfers
type ProgressStats struct {
	Completed int64
	Total     int64
}

// GetDiskTransferProgress extracts disk transfer progress from a migration object
func GetDiskTransferProgress(migration *unstructured.Unstructured) (ProgressStats, error) {
	stats := ProgressStats{}

	// Get the list of VMs from the migration status
	vms, exists, err := unstructured.NestedSlice(migration.Object, "status", "vms")
	if err != nil {
		return stats, fmt.Errorf("failed to get VM list: %v", err)
	}

	if !exists || len(vms) == 0 {
		return stats, nil
	}

	// Check each VM's pipeline for disk transfer progress
	for _, v := range vms {
		vm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		// Get the pipeline
		pipeline, exists, _ := unstructured.NestedSlice(vm, "pipeline")
		if !exists {
			continue
		}

		// Look for disk transfer phases
		for _, p := range pipeline {
			phase, ok := p.(map[string]interface{})
			if !ok {
				continue
			}

			// Check if this is a disk transfer phase
			name, _, _ := unstructured.NestedString(phase, "name")
			if name == "" || !strings.HasPrefix(name, "DiskTransfer") {
				continue
			}

			// Extract progress information
			completed, found, _ := unstructured.NestedInt64(phase, "progress", "completed")
			if found {
				stats.Completed += completed
			}

			total, found, _ := unstructured.NestedInt64(phase, "progress", "total")
			if found {
				stats.Total += total
			}
		}
	}

	return stats, nil
}

// PlanDetails contains all relevant status information for a plan
type PlanDetails struct {
	IsReady             bool
	HasRunningMigration bool
	Status              string
	VMStats             VMStats
	DiskProgress        ProgressStats
}

// GetPlanDetails returns all details about a plan's status in a single call
func GetPlanDetails(
	client dynamic.Interface,
	namespace string,
	plan *unstructured.Unstructured,
	migrationsGVR schema.GroupVersionResource,
) (PlanDetails, error) {
	details := PlanDetails{}

	// Get if plan is ready
	ready, err := IsPlanReady(plan)
	if err != nil {
		return details, err
	}
	details.IsReady = ready

	// Get if plan has running migration
	runningMigration, err := GetRunningMigration(client, namespace, plan, migrationsGVR)
	if err != nil {
		return details, err
	}
	details.HasRunningMigration = runningMigration != nil

	// Get plan status
	status, err := GetPlanStatus(plan)
	if err != nil {
		return details, err
	}
	details.Status = status

	// If there's a running migration, get VM stats
	if runningMigration != nil {
		// Get the plan UID
		planUID, found, err := unstructured.NestedString(plan.Object, "metadata", "uid")
		if !found || err != nil {
			return details, nil // Continue without VM stats
		}

		// Get all migrations in the namespace
		migrationList, err := client.Resource(migrationsGVR).
			Namespace(namespace).
			List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return details, nil // Continue without VM stats
		}

		// Find the active migration for this plan
		for _, migration := range migrationList.Items {
			planRef, found, _ := unstructured.NestedMap(migration.Object, "spec", "plan")
			if !found {
				continue
			}

			refUID, found, _ := unstructured.NestedString(planRef, "uid")
			if !found || refUID != planUID {
				continue
			}

			// Found the migration for this plan, get VM stats
			vmStats, _ := GetVMStats(&migration)
			details.VMStats = vmStats

			// Get disk transfer progress
			diskProgress, _ := GetDiskTransferProgress(&migration)
			details.DiskProgress = diskProgress

			break
		}
	}

	return details, nil
}
