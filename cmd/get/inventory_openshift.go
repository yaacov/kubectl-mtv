package get

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/inventory"
)

// NewInventoryNamespaceCmd creates the get inventory namespace command
func NewInventoryNamespaceCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "namespace PROVIDER",
		Short:        "Get namespaces from a provider (openshift, openstack)",
		Long:         `Get namespaces from a provider (openshift, openstack)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := getGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.GetKubeConfigFlags(), config.GetAllNamespaces())

			logNamespaceOperation("Getting namespaces from provider", namespace, config.GetAllNamespaces())
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.GetKubeConfigFlags(), namespace)
			}

			return inventory.ListNamespaces(config.GetKubeConfigFlags(), provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

// NewInventoryPVCCmd creates the get inventory pvc command
func NewInventoryPVCCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "pvc PROVIDER",
		Short:        "Get PVCs from a provider (openshift)",
		Long:         `Get PVCs from a provider (openshift)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := getGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.GetKubeConfigFlags(), config.GetAllNamespaces())

			logNamespaceOperation("Getting PVCs from provider", namespace, config.GetAllNamespaces())
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.GetKubeConfigFlags(), namespace)
			}

			return inventory.ListPersistentVolumeClaims(config.GetKubeConfigFlags(), provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

// NewInventoryDataVolumeCmd creates the get inventory datavolume command
func NewInventoryDataVolumeCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "datavolume PROVIDER",
		Short:        "Get data volumes from a provider (openshift)",
		Long:         `Get data volumes from a provider (openshift)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := getGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.GetKubeConfigFlags(), config.GetAllNamespaces())

			logNamespaceOperation("Getting data volumes from provider", namespace, config.GetAllNamespaces())
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.GetKubeConfigFlags(), namespace)
			}

			return inventory.ListDataVolumes(config.GetKubeConfigFlags(), provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}
