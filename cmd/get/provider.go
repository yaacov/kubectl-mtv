package get

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/provider"
)

// NewProviderCmd creates the get provider command
func NewProviderCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	var outputFormat string
	var inventoryURL string

	cmd := &cobra.Command{
		Use:          "provider [NAME]",
		Short:        "Get providers",
		Long:         `Get providers`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := getGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.GetKubeConfigFlags(), config.GetAllNamespaces())

			// Get optional provider name from arguments
			var providerName string
			if len(args) > 0 {
				providerName = args[0]
			}

			// If inventoryURL is empty, try to discover it
			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.GetKubeConfigFlags(), namespace)
			}

			// Log the operation being performed
			if providerName != "" {
				logNamespaceOperation("Getting provider", namespace, config.GetAllNamespaces())
			} else {
				logNamespaceOperation("Getting providers", namespace, config.GetAllNamespaces())
			}
			logOutputFormat(outputFormat)

			return provider.List(config.GetKubeConfigFlags(), namespace, inventoryURL, outputFormat, providerName)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&inventoryURL, "inventory-url", "i", os.Getenv("MTV_INVENTORY_URL"), "Base URL for the inventory service")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}
