package describe

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/cmd/get"
	"github.com/yaacov/kubectl-mtv/pkg/cmd/describe/plan"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
)

// NewPlanCmd creates the plan description command
func NewPlanCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() get.GlobalConfigGetter) *cobra.Command {
	var withVMs bool

	cmd := &cobra.Command{
		Use:               "plan NAME",
		Short:             "Describe a migration plan",
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.PlanNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Get the global configuration
			config := getGlobalConfig()

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(config.GetKubeConfigFlags())
			return plan.Describe(config.GetKubeConfigFlags(), name, namespace, withVMs, config.GetUseUTC())
		},
	}

	cmd.Flags().BoolVar(&withVMs, "with-vms", false, "Include list of VMs in the plan specification")

	return cmd
}
