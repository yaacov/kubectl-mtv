package cmd

import (
	"fmt"

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
	var name, sourceProvider, targetProvider string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create-network",
		Short: "Create a network map",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			return mapping.CreateNetwork(kubeConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Mapping name")
	cmd.Flags().StringVar(&sourceProvider, "source", "", "Source provider name")
	cmd.Flags().StringVar(&targetProvider, "target", "", "Target provider name")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Create from YAML file")

	if err := cmd.MarkFlagRequired("name"); err != nil {
		fmt.Printf("Warning: error marking 'provider' flag as required: %v\n", err)
	}

	return cmd
}

func newCreateStorageMappingCmd() *cobra.Command {
	var name, sourceProvider, targetProvider string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create-storage",
		Short: "Create a storage map",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			return mapping.CreateStorage(kubeConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Mapping name")
	cmd.Flags().StringVar(&sourceProvider, "source", "", "Source provider name")
	cmd.Flags().StringVar(&targetProvider, "target", "", "Target provider name")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Create from YAML file")

	if err := cmd.MarkFlagRequired("name"); err != nil {
		fmt.Printf("Warning: error marking 'provider' flag as required: %v\n", err)
	}

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
