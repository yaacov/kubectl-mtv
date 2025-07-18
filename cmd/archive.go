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
		Use:   "plan NAME",
		Short: "Archive a migration plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			err := plan.Archive(kubeConfigFlags, name, namespace, true)
			if err != nil {
				printCommandError(err, "archiving plan", namespace)
			}
			return nil
		},
	}

	return cmd
}
