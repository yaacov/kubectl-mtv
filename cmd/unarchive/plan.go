package unarchive

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/plan"
)

// NewPlanCmd creates the plan unarchive command
func NewPlanCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "plan NAME [NAME...]",
		Short:        "Un-archive one or more migration plans",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Loop over each plan name and unarchive it
			for _, name := range args {
				err := plan.Archive(kubeConfigFlags, name, namespace, false) // Set archived to false for unarchiving
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	return cmd
}
