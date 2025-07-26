package create

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/mapping"
)

// NewMappingCmd creates the mapping creation command
func NewMappingCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	var mappingType string
	var sourceProvider, targetProvider string
	var fromFile string
	var networkPairs string
	var storagePairs string
	var inventoryURL string

	cmd := &cobra.Command{
		Use:          "mapping NAME",
		Short:        "Create a new mapping",
		Long:         `Create a new network or storage mapping`,
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

			var err error
			switch mappingType {
			case "network":
				err = mapping.CreateNetwork(kubeConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile, networkPairs, inventoryURL)
			case "storage":
				err = mapping.CreateStorage(kubeConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile, storagePairs, inventoryURL)
			default:
				err = fmt.Errorf("unsupported mapping type: %s. Use 'network' or 'storage'", mappingType)
			}

			return err
		},
	}

	cmd.Flags().StringVarP(&mappingType, "type", "t", "", "Mapping type (network, storage)")
	cmd.Flags().StringVarP(&sourceProvider, "source", "S", "", "Source provider name")
	cmd.Flags().StringVarP(&targetProvider, "target", "T", "", "Target provider name")
	cmd.Flags().StringVarP(&fromFile, "from-file", "f", "", "Create mapping from YAML/JSON file")
	cmd.Flags().StringVar(&networkPairs, "network-pairs", "", "Network mapping pairs in format 'source:target-namespace/target-network', 'source:target-network', 'source:pod', or 'source:ignored' (comma-separated)")
	cmd.Flags().StringVar(&storagePairs, "storage-pairs", "", "Storage mapping pairs in format 'source:storage-class' (comma-separated). Note: storage classes are cluster-scoped")
	cmd.Flags().StringVarP(&inventoryURL, "inventory-url", "i", os.Getenv("MTV_INVENTORY_URL"), "Base URL for the inventory service")

	// Add completion for mapping type flag
	if err := cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"network", "storage"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	if err := cmd.MarkFlagRequired("type"); err != nil {
		panic(err)
	}

	return cmd
}
