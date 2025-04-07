package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/inventory"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func newInventoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inventory",
		Short: "Query provider inventories",
		Long:  `Query provider inventories for VMs, networks, and storage`,
	}

	cmd.AddCommand(newListVMsCmd())
	cmd.AddCommand(newListNetworksCmd())
	cmd.AddCommand(newListStorageCmd())
	cmd.AddCommand(newListHostsCmd())

	return cmd
}

func newListVMsCmd() *cobra.Command {
	var provider string
	var inventoryURL string
	var outputFormat string
	var extendedOutput bool
	var query string

	cmd := &cobra.Command{
		Use:   "vms",
		Short: "List VMs from a provider",
		Long: `List VMs from a provider
		
Query syntax allows:
- SELECT field1, field2 AS alias, field3  (select specific fields with optional aliases)
- WHERE condition                         (filter using tree-search-language conditions)
- ORDER BY field1 [ASC|DESC], field2      (sort results on multiple fields)
- LIMIT n                                 (limit number of results)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// If inventoryURL is empty, try to discover it
			if inventoryURL == "" {
				inventoryURL = discoverInventoryURL(kubeConfigFlags, namespace)
			}

			return inventory.ListVMs(kubeConfigFlags, provider, namespace, inventoryURL, outputFormat, extendedOutput, query)
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "Provider name")
	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Base URL for the inventory service")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format. One of: table, json")
	cmd.Flags().BoolVar(&extendedOutput, "extended", false, "Show extended information in table output")
	cmd.Flags().StringVar(&query, "query", "", "Query string with 'where', 'order by', and 'limit' clauses")

	if err := cmd.MarkFlagRequired("provider"); err != nil {
		fmt.Printf("Warning: error marking 'provider' flag as required: %v\n", err)
	}

	return cmd
}

func newListNetworksCmd() *cobra.Command {
	var provider string
	var inventoryURL string
	var outputFormat string
	var extendedOutput bool
	var query string

	cmd := &cobra.Command{
		Use:   "networks",
		Short: "List networks from a provider",
		Long: `List networks from a provider
		
Query syntax allows:
- SELECT field1, field2 AS alias, field3  (select specific fields with optional aliases)
- WHERE condition                         (filter using tree-search-language conditions)
- ORDER BY field1 [ASC|DESC], field2      (sort results on multiple fields)
- LIMIT n                                 (limit number of results)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// If inventoryURL is empty, try to discover it
			if inventoryURL == "" {
				inventoryURL = discoverInventoryURL(kubeConfigFlags, namespace)
			}

			return inventory.ListNetworks(kubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query)
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "Provider name")
	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Base URL for the inventory service")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format. One of: table, json")
	cmd.Flags().BoolVar(&extendedOutput, "extended", false, "Show extended information in table output")
	cmd.Flags().StringVar(&query, "query", "", "Query string with 'where', 'order by', and 'limit' clauses")

	if err := cmd.MarkFlagRequired("provider"); err != nil {
		fmt.Printf("Warning: error marking 'provider' flag as required: %v\n", err)
	}

	return cmd
}

func newListStorageCmd() *cobra.Command {
	var provider string
	var inventoryURL string
	var outputFormat string
	var extendedOutput bool
	var query string

	cmd := &cobra.Command{
		Use:   "storage",
		Short: "List storage from a provider",
		Long: `List storage from a provider
		
Query syntax allows:
- SELECT field1, field2 AS alias, field3  (select specific fields with optional aliases)
- WHERE condition                         (filter using tree-search-language conditions)
- ORDER BY field1 [ASC|DESC], field2      (sort results on multiple fields)
- LIMIT n                                 (limit number of results)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// If inventoryURL is empty, try to discover it
			if inventoryURL == "" {
				inventoryURL = discoverInventoryURL(kubeConfigFlags, namespace)
			}

			return inventory.ListStorage(kubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query)
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "Provider name")
	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Base URL for the inventory service")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format. One of: table, json")
	cmd.Flags().BoolVar(&extendedOutput, "extended", false, "Show extended information in table output")
	cmd.Flags().StringVar(&query, "query", "", "Query string with 'where', 'order by', and 'limit' clauses")

	if err := cmd.MarkFlagRequired("provider"); err != nil {
		fmt.Printf("Warning: error marking 'provider' flag as required: %v\n", err)
	}

	return cmd
}

func newListHostsCmd() *cobra.Command {
	var provider string
	var inventoryURL string
	var outputFormat string
	var extendedOutput bool
	var query string

	cmd := &cobra.Command{
		Use:   "hosts",
		Short: "List hosts from a provider",
		Long: `List hosts from a provider
		
Query syntax allows:
- SELECT field1, field2 AS alias, field3  (select specific fields with optional aliases)
- WHERE condition                         (filter using tree-search-language conditions)
- ORDER BY field1 [ASC|DESC], field2      (sort results on multiple fields)
- LIMIT n                                 (limit number of results)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// If inventoryURL is empty, try to discover it
			if inventoryURL == "" {
				inventoryURL = discoverInventoryURL(kubeConfigFlags, namespace)
			}

			return inventory.ListHosts(kubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query)
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "Provider name")
	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Base URL for the inventory service")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format. One of: table, json")
	cmd.Flags().BoolVar(&extendedOutput, "extended", false, "Show extended information in table output")
	cmd.Flags().StringVar(&query, "query", "", "Query string with 'where', 'order by', and 'limit' clauses")

	if err := cmd.MarkFlagRequired("provider"); err != nil {
		fmt.Printf("Warning: error marking 'provider' flag as required: %v\n", err)
	}

	return cmd
}

// discoverInventoryURL tries to discover the inventory URL from an OpenShift Route
func discoverInventoryURL(configFlags *genericclioptions.ConfigFlags, namespace string) string {
	route, err := client.GetForkliftInventoryRoute(configFlags, namespace)
	if err == nil && route != nil {
		host, found, _ := unstructured.NestedString(route.Object, "spec", "host")
		if found && host != "" {
			return fmt.Sprintf("https://%s", host)
		}
	}
	return ""
}
