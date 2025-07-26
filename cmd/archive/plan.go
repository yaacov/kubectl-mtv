package archive

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/archive/plan"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
)

// NewPlanCmd creates the plan archiving command
func NewPlanCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "plan NAME [NAME...]",
		Short:             "Archive one or more migration plans",
		Args:              cobra.MinimumNArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.PlanNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Loop over each plan name and archive it
			for _, name := range args {
				err := plan.Archive(kubeConfigFlags, name, namespace, true)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	return cmd
}
