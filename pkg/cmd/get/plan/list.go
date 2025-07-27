package plan

import (
	"context"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/get/plan/status"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/output"
	"github.com/yaacov/kubectl-mtv/pkg/util/watch"
)

// getPlans retrieves all plans from the given namespace
func getPlans(dynamicClient dynamic.Interface, namespace string) (*unstructured.UnstructuredList, error) {
	if namespace != "" {
		return dynamicClient.Resource(client.PlansGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	} else {
		return dynamicClient.Resource(client.PlansGVR).List(context.TODO(), metav1.ListOptions{})
	}
}

// getSpecificPlan retrieves a specific plan by name
func getSpecificPlan(dynamicClient dynamic.Interface, namespace, planName string) (*unstructured.UnstructuredList, error) {
	if namespace != "" {
		// If namespace is specified, get the specific resource
		plan, err := dynamicClient.Resource(client.PlansGVR).Namespace(namespace).Get(context.TODO(), planName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// Create a list with just this plan
		return &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{*plan},
		}, nil
	} else {
		// If no namespace specified, list all and filter by name
		plans, err := dynamicClient.Resource(client.PlansGVR).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list plans: %v", err)
		}

		var filteredItems []unstructured.Unstructured
		for _, plan := range plans.Items {
			if plan.GetName() == planName {
				filteredItems = append(filteredItems, plan)
			}
		}

		if len(filteredItems) == 0 {
			return nil, fmt.Errorf("plan '%s' not found", planName)
		}

		return &unstructured.UnstructuredList{
			Items: filteredItems,
		}, nil
	}
}

// ListPlans lists migration plans without watch functionality
func ListPlans(configFlags *genericclioptions.ConfigFlags, namespace string, outputFormat string, planName string) error {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	var plans *unstructured.UnstructuredList
	if planName != "" {
		// Get specific plan by name
		plans, err = getSpecificPlan(c, namespace, planName)
		if err != nil {
			return fmt.Errorf("failed to get plan: %v", err)
		}
	} else {
		// Get all plans
		plans, err = getPlans(c, namespace)
		if err != nil {
			return fmt.Errorf("failed to list plans: %v", err)
		}
	}

	// Format validation
	outputFormat = strings.ToLower(outputFormat)
	if outputFormat != "table" && outputFormat != "json" && outputFormat != "yaml" {
		return fmt.Errorf("unsupported output format: %s. Supported formats: table, json, yaml", outputFormat)
	}

	// Create printer items
	items := []map[string]interface{}{}
	for _, p := range plans.Items {
		source, _, _ := unstructured.NestedString(p.Object, "spec", "provider", "source", "name")
		target, _, _ := unstructured.NestedString(p.Object, "spec", "provider", "destination", "name")
		vms, _, _ := unstructured.NestedSlice(p.Object, "spec", "vms")
		creationTime := p.GetCreationTimestamp()

		// Get archived status
		archived, exists, _ := unstructured.NestedBool(p.Object, "spec", "archived")
		if !exists {
			archived = false
		}

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
			"archived": fmt.Sprintf("%t", archived),
			"object":   p.Object, // Include the original object
		}

		// Add the item to the list
		items = append(items, item)
	}

	// Handle different output formats
	switch outputFormat {
	case "json":
		// Use JSON printer
		jsonPrinter := output.NewJSONPrinter().
			WithPrettyPrint(true).
			AddItems(items)

		if len(plans.Items) == 0 {
			return jsonPrinter.PrintEmpty("No plans found in namespace " + namespace)
		}
		return jsonPrinter.Print()
	case "yaml":
		// Use YAML printer
		yamlPrinter := output.NewYAMLPrinter().
			AddItems(items)

		if len(plans.Items) == 0 {
			return yamlPrinter.PrintEmpty("No plans found in namespace " + namespace)
		}
		return yamlPrinter.Print()
	}

	// Use Table printer (default)
	var headers []output.Header

	// Add NAME column first
	headers = append(headers, output.Header{DisplayName: "NAME", JSONPath: "metadata.name"})

	// Add NAMESPACE column after NAME when listing across all namespaces
	if namespace == "" {
		headers = append(headers, output.Header{DisplayName: "NAMESPACE", JSONPath: "metadata.namespace"})
	}

	// Add remaining columns
	headers = append(headers,
		output.Header{DisplayName: "SOURCE", JSONPath: "source"},
		output.Header{DisplayName: "TARGET", JSONPath: "target"},
		output.Header{DisplayName: "VMS", JSONPath: "vms"},
		output.Header{DisplayName: "READY", JSONPath: "ready"},
		output.Header{DisplayName: "STATUS", JSONPath: "status"},
		output.Header{DisplayName: "PROGRESS", JSONPath: "progress"},
		output.Header{DisplayName: "CUTOVER", JSONPath: "cutover"},
		output.Header{DisplayName: "ARCHIVED", JSONPath: "archived"},
		output.Header{DisplayName: "CREATED", JSONPath: "created"},
	)

	tablePrinter := output.NewTablePrinter().WithHeaders(headers...).AddItems(items)

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
func List(configFlags *genericclioptions.ConfigFlags, namespace string, watchMode bool, outputFormat string, planName string) error {
	if watchMode {
		if outputFormat != "table" {
			return fmt.Errorf("watch mode only supports table output format")
		}
		return watch.Watch(func() error {
			return ListPlans(configFlags, namespace, outputFormat, planName)
		}, 15*time.Second)
	}

	return ListPlans(configFlags, namespace, outputFormat, planName)
}
