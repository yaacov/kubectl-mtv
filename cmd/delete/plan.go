package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/delete/plan"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
	"github.com/yaacov/kubectl-mtv/pkg/util/flags"
)

// NewPlanCmd creates the plan deletion command
func NewPlanCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	var all bool
	var skipArchive bool
	var cleanAll bool

	cmd := &cobra.Command{
		Use:               "plan [NAME...] [--all] [--skip-archive] [--clean-all]",
		Short:             "Delete one or more migration plans",
		Long:              "Delete one or more migration plans. By default, plans are archived before deletion. Use --skip-archive to skip archiving and delete immediately. Use --clean-all to archive, delete VMs on failed migration, then delete.",
		Args:              flags.ValidateAllFlagArgs(func() bool { return all }, 1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.PlanNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			var planNames []string
			if all {
				// Get all plan names from the namespace
				var err error
				planNames, err = client.GetAllPlanNames(kubeConfigFlags, namespace)
				if err != nil {
					return fmt.Errorf("failed to get all plan names: %v", err)
				}
				if len(planNames) == 0 {
					fmt.Printf("No plans found in namespace %s\n", namespace)
					return nil
				}
			} else {
				planNames = args
			}

			// Loop over each plan name and delete it
			for _, name := range planNames {
				err := plan.Delete(kubeConfigFlags, name, namespace, skipArchive, cleanAll)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Delete all migration plans in the namespace")
	cmd.Flags().BoolVar(&skipArchive, "skip-archive", false, "Skip archiving and delete the plan immediately")
	cmd.Flags().BoolVar(&cleanAll, "clean-all", false, "Archive, delete VMs on failed migration, then delete")

	return cmd
}
