package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/plan"
)

func newUnArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unarchive",
		Short: "Un-archive resources",
		Long:  `Un-archive various MTV resources`,
	}

	cmd.AddCommand(newUnArchivePlanCmd())
	return cmd
}

func newUnArchivePlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan NAME [NAME...]",
		Short: "Un-archive one or more migration plans",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Loop over each plan name and unarchive it
			for _, name := range args {
				err := plan.Archive(kubeConfigFlags, name, namespace, false) // Set archived to false for unarchiving
				if err != nil {
					printCommandError(err, "unarchiving plan", namespace)
					// Continue with other plans even if one fails
				}
			}
			return nil
		},
	}
	return cmd
}
