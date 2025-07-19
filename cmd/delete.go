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
		Use:   "mapping NAME [NAME...]",
		Short: "Delete one or more mappings",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Loop over each mapping name and delete it
			for _, name := range args {
				err := mapping.Delete(kubeConfigFlags, name, namespace, mappingType)
				if err != nil {
					printCommandError(err, "deleting mapping", namespace)
					// Continue with other mappings even if one fails
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&mappingType, "type", "network", "Mapping type (network, storage)")

	return cmd
}

func newDeletePlanCmdMain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan NAME [NAME...]",
		Short: "Delete one or more migration plans",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Loop over each plan name and delete it
			for _, name := range args {
				err := plan.Delete(kubeConfigFlags, name, namespace)
				if err != nil {
					printCommandError(err, "deleting plan", namespace)
					// Continue with other plans even if one fails
				}
			}
			return nil
		},
	}

	return cmd
}

func newDeleteProviderCmdMain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider NAME [NAME...]",
		Short: "Delete one or more providers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Loop over each provider name and delete it
			for _, name := range args {
				err := provider.Delete(kubeConfigFlags, name, namespace)
				if err != nil {
					printCommandError(err, "deleting provider", namespace)
					// Continue with other providers even if one fails
				}
			}
			return nil
		},
	}

	return cmd
}
