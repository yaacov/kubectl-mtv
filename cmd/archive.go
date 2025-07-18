package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/plan"
)

func newArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive resources",
		Long:  `Archive various MTV resources`,
	}

	cmd.AddCommand(newArchivePlanCmd())
	return cmd
}

func newArchivePlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan NAME [NAME...]",
		Short: "Archive one or more migration plans",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Loop over each plan name and archive it
			for _, name := range args {
				err := plan.Archive(kubeConfigFlags, name, namespace, true)
				if err != nil {
					printCommandError(err, "archiving plan", namespace)
					// Continue with other plans even if one fails
				}
			}
			return nil
		},
	}

	return cmd
}
