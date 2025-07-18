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
		Use:   "plan NAME",
		Short: "Un-archive a migration plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			err := plan.Archive(kubeConfigFlags, name, namespace, false) // Set archived to false for unarchiving
			if err != nil {
				printCommandError(err, "unarchiving plan", namespace)
			}
			return nil
		},
	}
	return cmd
}
