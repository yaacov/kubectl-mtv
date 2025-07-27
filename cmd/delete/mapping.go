package delete

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/delete/mapping"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
)

// NewMappingCmd creates the mapping deletion command with subcommands
func NewMappingCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "mapping",
		Short:        "Delete mappings",
		Long:         `Delete network and storage mappings`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is specified, show help
			return cmd.Help()
		},
	}

	// Add subcommands for network and storage
	cmd.AddCommand(newDeleteNetworkMappingCmd(kubeConfigFlags))
	cmd.AddCommand(newDeleteStorageMappingCmd(kubeConfigFlags))

	return cmd
}

// newDeleteNetworkMappingCmd creates the delete network mapping subcommand
func newDeleteNetworkMappingCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "network NAME [NAME...]",
		Short:        "Delete one or more network mappings",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completion.MappingNameCompletion(kubeConfigFlags, "network")(cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Loop over each mapping name and delete it
			for _, name := range args {
				err := mapping.Delete(kubeConfigFlags, name, namespace, "network")
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	return cmd
}

// newDeleteStorageMappingCmd creates the delete storage mapping subcommand
func newDeleteStorageMappingCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "storage NAME [NAME...]",
		Short:        "Delete one or more storage mappings",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completion.MappingNameCompletion(kubeConfigFlags, "storage")(cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Loop over each mapping name and delete it
			for _, name := range args {
				err := mapping.Delete(kubeConfigFlags, name, namespace, "storage")
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	return cmd
}
