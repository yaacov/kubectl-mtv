package get

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/inventory"
)

// NewInventoryDiskProfileCmd creates the get inventory disk-profile command
func NewInventoryDiskProfileCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "disk-profile PROVIDER",
		Short:        "Get disk profiles from a provider (ovirt)",
		Long:         `Get disk profiles from a provider (ovirt)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := getGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.GetKubeConfigFlags(), config.GetAllNamespaces())

			logNamespaceOperation("Getting disk profiles from provider", namespace, config.GetAllNamespaces())
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.GetKubeConfigFlags(), namespace)
			}

			return inventory.ListDiskProfiles(config.GetKubeConfigFlags(), provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

// NewInventoryNICProfileCmd creates the get inventory nic-profile command
func NewInventoryNICProfileCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "nic-profile PROVIDER",
		Short:        "Get NIC profiles from a provider (ovirt)",
		Long:         `Get NIC profiles from a provider (ovirt)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := getGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.GetKubeConfigFlags(), config.GetAllNamespaces())

			logNamespaceOperation("Getting NIC profiles from provider", namespace, config.GetAllNamespaces())
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.GetKubeConfigFlags(), namespace)
			}

			return inventory.ListNICProfiles(config.GetKubeConfigFlags(), provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}
