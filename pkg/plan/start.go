package plan

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	forkliftv1beta1 "github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/plan/status"
)

// Start starts a migration plan
func Start(configFlags *genericclioptions.ConfigFlags, name, namespace string) error {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Get the plan
	plan, err := c.Resource(client.PlansGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get plan: %v", err)
	}

	// Check if the plan is ready
	planReady, err := status.IsPlanReady(plan)
	if err != nil {
		return err
	}
	if !planReady {
		return fmt.Errorf("migration plan '%s' is not ready", name)
	}

	// Check if the plan has running migrations
	hasRunning, err := status.HasRunningMigration(c, namespace, plan, client.MigrationsGVR)
	if err != nil {
		return err
	}
	if hasRunning {
		return fmt.Errorf("migration plan '%s' already has a running migration", name)
	}

	// Check if the plan has already succeeded
	planStatus, err := status.GetPlanStatus(plan)
	if err != nil {
		return err
	}
	if planStatus == status.StatusSucceeded {
		return fmt.Errorf("migration plan '%s' has already succeeded", name)
	}

	// Extract the plan's UID
	planUID := string(plan.GetUID())

	// Create a migration object using structured type
	migration := &forkliftv1beta1.Migration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-migration", name),
			Namespace: namespace,
		},
		Spec: forkliftv1beta1.MigrationSpec{
			Plan: corev1.ObjectReference{
				Name:      name,
				Namespace: namespace,
				UID:       types.UID(planUID),
			},
		},
	}
	migration.Kind = "Migration"
	migration.APIVersion = forkliftv1beta1.SchemeGroupVersion.String()

	// Convert Migration object to Unstructured
	unstructuredMigration, err := runtime.DefaultUnstructuredConverter.ToUnstructured(migration)
	if err != nil {
		return fmt.Errorf("failed to convert Migration to Unstructured: %v", err)
	}
	migrationUnstructured := &unstructured.Unstructured{Object: unstructuredMigration}

	// Create the migration in the specified namespace
	_, err = c.Resource(client.MigrationsGVR).Namespace(namespace).Create(context.TODO(), migrationUnstructured, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create migration: %v", err)
	}

	fmt.Printf("Migration started for plan '%s' in namespace '%s'\n", name, namespace)
	return nil
}
