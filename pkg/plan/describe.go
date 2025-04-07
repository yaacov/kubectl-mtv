package plan

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
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

	fmt.Printf("Name:              %s\n", plan.GetName())
	fmt.Printf("Namespace:         %s\n", plan.GetNamespace())
	fmt.Printf("Created:           %s\n", plan.GetCreationTimestamp())

	source, _, _ := unstructured.NestedString(plan.Object, "spec", "provider", "source", "name")
	target, _, _ := unstructured.NestedString(plan.Object, "spec", "provider", "destination", "name")
	networkMapping, _, _ := unstructured.NestedString(plan.Object, "spec", "map", "network", "name")
	storageMapping, _, _ := unstructured.NestedString(plan.Object, "spec", "map", "storage", "name")

	fmt.Printf("Source Provider:   %s\n", source)
	fmt.Printf("Target Provider:   %s\n", target)
	fmt.Printf("Network Mapping:   %s\n", networkMapping)
	fmt.Printf("Storage Mapping:   %s\n", storageMapping)

	vms, _, _ := unstructured.NestedSlice(plan.Object, "spec", "vms")
	fmt.Printf("VMs:               %d\n", len(vms))

	fmt.Println("\nVM Details:")
	fmt.Printf("%-36s %-30s %-15s\n", "ID", "NAME", "STATUS")
	for _, vmObj := range vms {
		vm := vmObj.(map[string]interface{})
		id, _ := vm["id"].(string)
		name, ok := vm["name"].(string)
		if !ok {
			name = "-"
		}
		status, ok := vm["status"].(string)
		if !ok {
			status = "Pending"
		}
		fmt.Printf("%-36s %-30s %-15s\n", id, name, status)
	}

	return nil
}
