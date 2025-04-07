package client

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// CanAccessRoutesInNamespace checks if the user has permissions to list routes in the given namespace
func CanAccessRoutesInNamespace(configFlags *genericclioptions.ConfigFlags, namespace string) bool {
	return CanAccessResource(configFlags, namespace, RouteGVR, "list")
}

// GetForkliftInventoryRoute attempts to find a route with the forklift inventory service labels
func GetForkliftInventoryRoute(configFlags *genericclioptions.ConfigFlags, namespace string) (*unstructured.Unstructured, error) {
	// Get dynamic client
	c, err := GetDynamicClient(configFlags)
	if err != nil {
		return nil, err
	}

	// Create label selector for forklift inventory route
	labelSelector := "app=forklift,service=forklift-inventory"

	// Check if we have access to the openshift-mtv namespace
	if CanAccessRoutesInNamespace(configFlags, OpenShiftMTVNamespace) {
		// Try to find the route in the openshift-mtv namespace
		routes, err := c.Resource(RouteGVR).Namespace(OpenShiftMTVNamespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})

		// If we find a route in openshift-mtv namespace, return it
		if err == nil && len(routes.Items) > 0 {
			return &routes.Items[0], nil
		}
	}

	// If we couldn't find the route in openshift-mtv or didn't have permissions,
	// try the provided namespace
	routes, err := c.Resource(RouteGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	// Return the first matching route if found
	if len(routes.Items) > 0 {
		return &routes.Items[0], nil
	}

	return nil, fmt.Errorf("no matching route found")
}
