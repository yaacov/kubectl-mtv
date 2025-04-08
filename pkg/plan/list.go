package plan

import (
	"context"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/output"
	"github.com/yaacov/kubectl-mtv/pkg/plan/status"
	"github.com/yaacov/kubectl-mtv/pkg/watch"
)

// ListPlans lists migration plans without watch functionality
func ListPlans(configFlags *genericclioptions.ConfigFlags, namespace string, outputFormat string) error {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	plans, err := c.Resource(client.PlansGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list plans: %v", err)
	}

	// Format validation
	outputFormat = strings.ToLower(outputFormat)
	if outputFormat != "table" && outputFormat != "json" {
		return fmt.Errorf("unsupported output format: %s. Supported formats: table, json", outputFormat)
	}

	// Create printer items
	items := []map[string]interface{}{}
	for _, p := range plans.Items {
		source, _, _ := unstructured.NestedString(p.Object, "spec", "provider", "source", "name")
		target, _, _ := unstructured.NestedString(p.Object, "spec", "provider", "destination", "name")
		vms, _, _ := unstructured.NestedSlice(p.Object, "spec", "vms")
		creationTime := p.GetCreationTimestamp()

		// Get plan details (ready, running migration, status)
		planDetails, _ := status.GetPlanDetails(c, namespace, &p, client.MigrationsGVR)

		// Format the VM migration status
		var vmStatus string
		if planDetails.RunningMigration != nil && planDetails.VMStats.Total > 0 {
			vmStatus = fmt.Sprintf("%d/%d (S:%d/F:%d/C:%d)",
				planDetails.VMStats.Completed,
				planDetails.VMStats.Total,
				planDetails.VMStats.Succeeded,
				planDetails.VMStats.Failed,
				planDetails.VMStats.Canceled)
		} else {
			vmStatus = fmt.Sprintf("%d", len(vms))
		}

		// Format the disk transfer progress
		progressStatus := "-"
		if planDetails.RunningMigration != nil && planDetails.DiskProgress.Total > 0 {
			percentage := float64(planDetails.DiskProgress.Completed) / float64(planDetails.DiskProgress.Total) * 100
			progressStatus = fmt.Sprintf("%.1f%% (%d/%d GB)",
				percentage,
				planDetails.DiskProgress.Completed/(1024), // Convert to GB
				planDetails.DiskProgress.Total/(1024))     // Convert to GB
		}

		// Determine cutover information
		cutoverInfo := "cold" // Default for cold migration
		warm, exists, _ := unstructured.NestedBool(p.Object, "spec", "warm")
		if exists && warm {
			cutoverInfo = "-" // Default for warm migration without running migration
			if planDetails.RunningMigration != nil {
				// Extract cutover time from running migration
				cutoverTimeStr, exists, _ := unstructured.NestedString(planDetails.RunningMigration.Object, "spec", "cutover")
				if exists && cutoverTimeStr != "" {
					// Parse the cutover time string
					cutoverTime, err := time.Parse(time.RFC3339, cutoverTimeStr)
					if err == nil {
						cutoverInfo = cutoverTime.Format("2006-01-02 15:04:05")
					}
				}
			}
		}

		// Create a new printer item
		item := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      p.GetName(),
				"namespace": p.GetNamespace(),
			},
			"source":   source,
			"target":   target,
			"created":  creationTime.Format("2006-01-02 15:04:05"),
			"vms":      vmStatus,
			"ready":    fmt.Sprintf("%t", planDetails.IsReady),
			"running":  fmt.Sprintf("%t", planDetails.RunningMigration != nil),
			"status":   planDetails.Status,
			"progress": progressStatus,
			"cutover":  cutoverInfo,
		}

		// Add the item to the list
		items = append(items, item)
	}

	// Handle different output formats
	if outputFormat == "json" {
		// Use JSON printer
		jsonPrinter := output.NewJSONPrinter().
			WithPrettyPrint(true).
			AddItems(items)

		if len(plans.Items) == 0 {
			return jsonPrinter.PrintEmpty("No plans found in namespace " + namespace)
		}
		return jsonPrinter.Print()
	}

	// Use Table printer (default)
	tablePrinter := output.NewTablePrinter().WithHeaders(
		output.Header{DisplayName: "NAME", JSONPath: "metadata.name"},
		output.Header{DisplayName: "SOURCE", JSONPath: "source"},
		output.Header{DisplayName: "TARGET", JSONPath: "target"},
		output.Header{DisplayName: "VMS", JSONPath: "vms"},
		output.Header{DisplayName: "READY", JSONPath: "ready"},
		output.Header{DisplayName: "STATUS", JSONPath: "status"},
		output.Header{DisplayName: "PROGRESS", JSONPath: "progress"},
		output.Header{DisplayName: "CUTOVER", JSONPath: "cutover"},
		output.Header{DisplayName: "CREATED", JSONPath: "created"},
	).AddItems(items)

	if len(plans.Items) == 0 {
		if err := tablePrinter.PrintEmpty("No plans found in namespace " + namespace); err != nil {
			return fmt.Errorf("error printing empty table: %v", err)
		}
	} else {
		if err := tablePrinter.Print(); err != nil {
			return fmt.Errorf("error printing table: %v", err)
		}
	}

	return nil
}

// List lists migration plans with optional watch mode
func List(configFlags *genericclioptions.ConfigFlags, namespace string, watchMode bool, outputFormat string) error {
	if watchMode {
		if outputFormat != "table" {
			return fmt.Errorf("watch mode only supports table output format")
		}
		return watch.Watch(func() error {
			return ListPlans(configFlags, namespace, outputFormat)
		}, 15*time.Second)
	}

	return ListPlans(configFlags, namespace, outputFormat)
}
