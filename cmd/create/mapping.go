package create

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/mapping"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
)

// NewMappingCmd creates the mapping creation command with subcommands
func NewMappingCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "mapping",
		Short:        "Create a new mapping",
		Long:         `Create a new network or storage mapping`,
		SilenceUsage: true,
	}

	// Add subcommands for network and storage
	cmd.AddCommand(newNetworkMappingCmd(kubeConfigFlags))
	cmd.AddCommand(newStorageMappingCmd(kubeConfigFlags))

	return cmd
}

// newNetworkMappingCmd creates the network mapping subcommand
func newNetworkMappingCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	var sourceProvider, targetProvider string
	var networkPairs string
	var inventoryURL string

	cmd := &cobra.Command{
		Use:          "network NAME",
		Short:        "Create a new network mapping",
		Long:         `Create a new network mapping between source and target providers`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// If inventoryURL is empty, try to discover it
			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(kubeConfigFlags, namespace)
			}

			return mapping.CreateNetwork(kubeConfigFlags, name, namespace, sourceProvider, targetProvider, networkPairs, inventoryURL)
		},
	}

	cmd.Flags().StringVarP(&sourceProvider, "source", "S", "", "Source provider name")
	cmd.Flags().StringVarP(&targetProvider, "target", "T", "", "Target provider name")
	cmd.Flags().StringVar(&networkPairs, "network-pairs", "", "Network mapping pairs in format 'source:target-namespace/target-network', 'source:target-network', 'source:pod', or 'source:ignored' (comma-separated)")
	cmd.Flags().StringVarP(&inventoryURL, "inventory-url", "i", os.Getenv("MTV_INVENTORY_URL"), "Base URL for the inventory service")

	return cmd
}

// newStorageMappingCmd creates the storage mapping subcommand
func newStorageMappingCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	var sourceProvider, targetProvider string
	var storagePairs string
	var inventoryURL string

	cmd := &cobra.Command{
		Use:          "storage NAME",
		Short:        "Create a new storage mapping",
		Long:         `Create a new storage mapping between source and target providers`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// If inventoryURL is empty, try to discover it
			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(kubeConfigFlags, namespace)
			}

			return mapping.CreateStorage(kubeConfigFlags, name, namespace, sourceProvider, targetProvider, storagePairs, inventoryURL)
		},
	}

	cmd.Flags().StringVarP(&sourceProvider, "source", "S", "", "Source provider name")
	cmd.Flags().StringVarP(&targetProvider, "target", "T", "", "Target provider name")
	cmd.Flags().StringVar(&storagePairs, "storage-pairs", "", "Storage mapping pairs in format 'source:storage-class' (comma-separated). Note: storage classes are cluster-scoped")
	cmd.Flags().StringVarP(&inventoryURL, "inventory-url", "i", os.Getenv("MTV_INVENTORY_URL"), "Base URL for the inventory service")

	return cmd
}
