package get

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/get/plan"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
	"github.com/yaacov/kubectl-mtv/pkg/util/flags"
)

// NewPlanCmd creates the get plan command
func NewPlanCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	outputFormatFlag := flags.NewOutputFormatTypeFlag()
	var watch bool

	cmd := &cobra.Command{
		Use:               "plan [NAME]",
		Short:             "Get migration plans",
		Long:              `Get migration plans`,
		Args:              cobra.MaximumNArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.PlanNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := getGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.GetKubeConfigFlags(), config.GetAllNamespaces())

			// Get optional plan name from arguments
			var planName string
			if len(args) > 0 {
				planName = args[0]
			}

			// Log the operation being performed
			if planName != "" {
				logNamespaceOperation("Getting plan", namespace, config.GetAllNamespaces())
			} else {
				logNamespaceOperation("Getting plans", namespace, config.GetAllNamespaces())
			}
			logOutputFormat(outputFormatFlag.GetValue())

			return plan.List(config.GetKubeConfigFlags(), namespace, watch, outputFormatFlag.GetValue(), planName)
		},
	}

	cmd.Flags().VarP(outputFormatFlag, "output", "o", "Output format (table, json, yaml)")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")

	// Add completion for output format flag
	if err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return outputFormatFlag.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	return cmd
}

// NewPlanVMsCmd creates the get plan-vms command
func NewPlanVMsCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	outputFormatFlag := flags.NewOutputFormatTypeFlag()
	var watch bool

	cmd := &cobra.Command{
		Use:               "plan-vms NAME",
		Short:             "Get VMs status in a migration plan",
		Long:              `Get VMs status in a migration plan`,
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.PlanNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			config := getGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.GetKubeConfigFlags(), config.GetAllNamespaces())

			// Log the operation being performed
			logNamespaceOperation("Getting plan VMs", namespace, config.GetAllNamespaces())
			logOutputFormat(outputFormatFlag.GetValue())

			return plan.ListVMs(config.GetKubeConfigFlags(), name, namespace, watch)
		},
	}

	cmd.Flags().VarP(outputFormatFlag, "output", "o", "Output format (table, json, yaml)")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")

	// Add completion for output format flag
	if err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return outputFormatFlag.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	return cmd
}
