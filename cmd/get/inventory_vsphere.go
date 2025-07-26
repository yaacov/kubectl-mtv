package get

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/inventory"
)

// NewInventoryDatastoreCmd creates the get inventory datastore command
func NewInventoryDatastoreCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "datastore PROVIDER",
		Short:        "Get datastores from a provider (vsphere)",
		Long:         `Get datastores from a provider (vsphere)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := getGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.GetKubeConfigFlags(), config.GetAllNamespaces())

			logNamespaceOperation("Getting datastores from provider", namespace, config.GetAllNamespaces())
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.GetKubeConfigFlags(), namespace)
			}

			return inventory.ListDatastores(config.GetKubeConfigFlags(), provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

// NewInventoryResourcePoolCmd creates the get inventory resource-pool command
func NewInventoryResourcePoolCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "resource-pool PROVIDER",
		Short:        "Get resource pools from a provider (vsphere)",
		Long:         `Get resource pools from a provider (vsphere)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := getGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.GetKubeConfigFlags(), config.GetAllNamespaces())

			logNamespaceOperation("Getting resource pools from provider", namespace, config.GetAllNamespaces())
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.GetKubeConfigFlags(), namespace)
			}

			return inventory.ListResourcePools(config.GetKubeConfigFlags(), provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

// NewInventoryFolderCmd creates the get inventory folder command
func NewInventoryFolderCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "folder PROVIDER",
		Short:        "Get folders from a provider (vsphere)",
		Long:         `Get folders from a provider (vsphere)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := getGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.GetKubeConfigFlags(), config.GetAllNamespaces())

			logNamespaceOperation("Getting folders from provider", namespace, config.GetAllNamespaces())
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.GetKubeConfigFlags(), namespace)
			}

			return inventory.ListFolders(config.GetKubeConfigFlags(), provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}
