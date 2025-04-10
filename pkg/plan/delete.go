package plan

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
)

// Delete removes a plan by name from the cluster
func Delete(configFlags *genericclioptions.ConfigFlags, name, namespace string) error {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Archive the plan
	err = Archive(configFlags, name, namespace, true)
	if err != nil {
		return fmt.Errorf("failed to archive plan: %v", err)
	}

	// Wait for the Archived condition to be true
	fmt.Printf("Waiting for plan '%s' to be archived...\n", name)
	for {
		plan, err := c.Resource(client.PlansGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get plan: %v", err)
		}

		conditions, exists, err := unstructured.NestedSlice(plan.Object, "status", "conditions")
		if err != nil || !exists {
			return fmt.Errorf("failed to get plan conditions: %v", err)
		}

		archived := false
		for _, condition := range conditions {
			cond, ok := condition.(map[string]interface{})
			if !ok {
				continue
			}

			condType, _, _ := unstructured.NestedString(cond, "type")
			condStatus, _, _ := unstructured.NestedString(cond, "status")

			if condType == "Archived" && condStatus == "True" {
				archived = true
				break
			}
		}

		if archived {
			break
		}

		// Wait before checking again
		time.Sleep(2 * time.Second)
	}

	// Delete the plan
	err = c.Resource(client.PlansGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete plan: %v", err)
	}

	fmt.Printf("Plan '%s' deleted from namespace '%s'\n", name, namespace)
	return nil
}
