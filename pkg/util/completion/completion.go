package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/util/client"
)

// getResourceNames fetches resource names for completion
func getResourceNames(configFlags *genericclioptions.ConfigFlags, gvr schema.GroupVersionResource, namespace string) ([]string, error) {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return nil, err
	}

	var list []string

	// List resources
	if namespace != "" {
		resources, err := c.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, resource := range resources.Items {
			list = append(list, resource.GetName())
		}
	} else {
		resources, err := c.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, resource := range resources.Items {
			list = append(list, resource.GetName())
		}
	}

	return list, nil
}

// PlanNameCompletion provides completion for plan names
func PlanNameCompletion(configFlags *genericclioptions.ConfigFlags) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		namespace := client.ResolveNamespace(configFlags)

		names, err := getResourceNames(configFlags, client.PlansGVR, namespace)
		if err != nil {
			return []string{fmt.Sprintf("Error fetching plans: %v", err)}, cobra.ShellCompDirectiveError
		}

		if len(names) == 0 {
			namespaceMsg := "current namespace"
			if namespace != "" {
				namespaceMsg = fmt.Sprintf("namespace '%s'", namespace)
			}
			return []string{fmt.Sprintf("No migration plans found in %s", namespaceMsg)}, cobra.ShellCompDirectiveError
		}

		// Filter results based on what's already typed
		var filtered []string
		for _, name := range names {
			if strings.HasPrefix(name, toComplete) {
				filtered = append(filtered, name)
			}
		}

		if len(filtered) == 0 && toComplete != "" {
			return []string{fmt.Sprintf("No migration plans matching '%s'", toComplete)}, cobra.ShellCompDirectiveError
		}

		return filtered, cobra.ShellCompDirectiveNoFileComp
	}
}

// ProviderNameCompletion provides completion for provider names
func ProviderNameCompletion(configFlags *genericclioptions.ConfigFlags) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		namespace := client.ResolveNamespace(configFlags)

		names, err := getResourceNames(configFlags, client.ProvidersGVR, namespace)
		if err != nil {
			return []string{fmt.Sprintf("Error fetching providers: %v", err)}, cobra.ShellCompDirectiveError
		}

		if len(names) == 0 {
			namespaceMsg := "current namespace"
			if namespace != "" {
				namespaceMsg = fmt.Sprintf("namespace '%s'", namespace)
			}
			return []string{fmt.Sprintf("No providers found in %s", namespaceMsg)}, cobra.ShellCompDirectiveError
		}

		// Filter results based on what's already typed
		var filtered []string
		for _, name := range names {
			if strings.HasPrefix(name, toComplete) {
				filtered = append(filtered, name)
			}
		}

		if len(filtered) == 0 && toComplete != "" {
			return []string{fmt.Sprintf("No providers matching '%s'", toComplete)}, cobra.ShellCompDirectiveError
		}

		return filtered, cobra.ShellCompDirectiveNoFileComp
	}
}

// MappingNameCompletion provides completion for mapping names
// mappingType should be "network" or "storage"
func MappingNameCompletion(configFlags *genericclioptions.ConfigFlags, mappingType string) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		namespace := client.ResolveNamespace(configFlags)

		var gvr schema.GroupVersionResource
		var resourceType string
		if mappingType == "storage" {
			gvr = client.StorageMapGVR
			resourceType = "storage mappings"
		} else {
			gvr = client.NetworkMapGVR
			resourceType = "network mappings"
		}

		names, err := getResourceNames(configFlags, gvr, namespace)
		if err != nil {
			return []string{fmt.Sprintf("Error fetching %s: %v", resourceType, err)}, cobra.ShellCompDirectiveError
		}

		if len(names) == 0 {
			namespaceMsg := "current namespace"
			if namespace != "" {
				namespaceMsg = fmt.Sprintf("namespace '%s'", namespace)
			}
			return []string{fmt.Sprintf("No %s found in %s", resourceType, namespaceMsg)}, cobra.ShellCompDirectiveError
		}

		// Filter results based on what's already typed
		var filtered []string
		for _, name := range names {
			if strings.HasPrefix(name, toComplete) {
				filtered = append(filtered, name)
			}
		}

		if len(filtered) == 0 && toComplete != "" {
			return []string{fmt.Sprintf("No %s matching '%s'", resourceType, toComplete)}, cobra.ShellCompDirectiveError
		}

		return filtered, cobra.ShellCompDirectiveNoFileComp
	}
}

// MigrationNameCompletion provides completion for migration names
func MigrationNameCompletion(configFlags *genericclioptions.ConfigFlags) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		namespace := client.ResolveNamespace(configFlags)

		names, err := getResourceNames(configFlags, client.MigrationsGVR, namespace)
		if err != nil {
			return []string{fmt.Sprintf("Error fetching migrations: %v", err)}, cobra.ShellCompDirectiveError
		}

		if len(names) == 0 {
			namespaceMsg := "current namespace"
			if namespace != "" {
				namespaceMsg = fmt.Sprintf("namespace '%s'", namespace)
			}
			return []string{fmt.Sprintf("No migrations found in %s", namespaceMsg)}, cobra.ShellCompDirectiveError
		}

		// Filter results based on what's already typed
		var filtered []string
		for _, name := range names {
			if strings.HasPrefix(name, toComplete) {
				filtered = append(filtered, name)
			}
		}

		if len(filtered) == 0 && toComplete != "" {
			return []string{fmt.Sprintf("No migrations matching '%s'", toComplete)}, cobra.ShellCompDirectiveError
		}

		return filtered, cobra.ShellCompDirectiveNoFileComp
	}
}
