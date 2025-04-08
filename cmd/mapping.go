package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/mapping"
)

func newMappingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mapping",
		Short: "Manage network and storage mappings",
		Long:  `Create and manage network and storage mappings for VM migration`,
	}

	cmd.AddCommand(newCreateNetworkMappingCmd())
	cmd.AddCommand(newCreateStorageMappingCmd())
	cmd.AddCommand(newListMappingCmd())

	return cmd
}

func newCreateNetworkMappingCmd() *cobra.Command {
	var sourceProvider, targetProvider string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create-network NAME",
		Short: "Create a network map",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			return mapping.CreateNetwork(kubeConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile)
		},
	}

	cmd.Flags().StringVarP(&sourceProvider, "source", "s", "", "Source provider name")
	cmd.Flags().StringVarP(&targetProvider, "target", "t", "", "Target provider name")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Create from YAML file")

	return cmd
}

func newCreateStorageMappingCmd() *cobra.Command {
	var sourceProvider, targetProvider string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create-storage NAME",
		Short: "Create a storage map",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			return mapping.CreateStorage(kubeConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile)
		},
	}

	cmd.Flags().StringVarP(&sourceProvider, "source", "s", "", "Source provider name")
	cmd.Flags().StringVarP(&targetProvider, "target", "t", "", "Target provider name")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Create from YAML file")

	return cmd
}

func newListMappingCmd() *cobra.Command {
	var mappingType string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mappings",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			return mapping.List(kubeConfigFlags, mappingType, namespace, outputFormat)
		},
	}

	cmd.Flags().StringVar(&mappingType, "type", "all", "Mapping type (network, storage, all)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format. One of: table, json")

	return cmd
}
