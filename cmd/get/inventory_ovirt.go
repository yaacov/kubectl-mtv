package get

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/commands/inventory"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/flags"
)

// NewInventoryDiskProfileCmd creates the get inventory disk-profile command
func NewInventoryDiskProfileCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	var inventoryURL string
	outputFormatFlag := flags.NewOutputFormatTypeFlag()
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
			logOutputFormat(outputFormatFlag.GetValue())

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.GetKubeConfigFlags(), namespace)
			}

			return inventory.ListDiskProfiles(config.GetKubeConfigFlags(), provider, namespace, inventoryURL, outputFormatFlag.GetValue(), query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().VarP(outputFormatFlag, "output", "o", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")

	// Add completion for output format flag
	if err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return outputFormatFlag.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	return cmd
}

// NewInventoryNICProfileCmd creates the get inventory nic-profile command
func NewInventoryNICProfileCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	var inventoryURL string
	outputFormatFlag := flags.NewOutputFormatTypeFlag()
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
			logOutputFormat(outputFormatFlag.GetValue())

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.GetKubeConfigFlags(), namespace)
			}

			return inventory.ListNICProfiles(config.GetKubeConfigFlags(), provider, namespace, inventoryURL, outputFormatFlag.GetValue(), query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().VarP(outputFormatFlag, "output", "o", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")

	// Add completion for output format flag
	if err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return outputFormatFlag.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	return cmd
}
