package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/mapping"
	"github.com/yaacov/kubectl-mtv/pkg/plan"
	"github.com/yaacov/kubectl-mtv/pkg/provider"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources",
		Long:  `Delete resources like mappings, plans, and providers`,
	}

	cmd.AddCommand(newDeleteMappingCmd())
	cmd.AddCommand(newDeletePlanCmdMain())
	cmd.AddCommand(newDeleteProviderCmdMain())

	return cmd
}

func newDeleteMappingCmd() *cobra.Command {
	var mappingType string

	cmd := &cobra.Command{
		Use:   "mapping NAME",
		Short: "Delete a mapping",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			err := mapping.Delete(kubeConfigFlags, name, namespace, mappingType)
			if err != nil {
				printCommandError(err, "deleting mapping", namespace)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&mappingType, "type", "network", "Mapping type (network, storage)")

	return cmd
}

func newDeletePlanCmdMain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan NAME",
		Short: "Delete a migration plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			err := plan.Delete(kubeConfigFlags, name, namespace)
			if err != nil {
				printCommandError(err, "deleting plan", namespace)
			}
			return nil
		},
	}

	return cmd
}

func newDeleteProviderCmdMain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider NAME",
		Short: "Delete a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			err := provider.Delete(kubeConfigFlags, name, namespace)
			if err != nil {
				printCommandError(err, "deleting provider", namespace)
			}
			return nil
		},
	}

	return cmd
}
