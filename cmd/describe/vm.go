package describe

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/cmd/get"
	plan "github.com/yaacov/kubectl-mtv/pkg/cmd/describe/vm"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
)

// NewVMCmd creates the VM description command
func NewVMCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() get.GlobalConfigGetter) *cobra.Command {
	var watch bool
	var vmName string

	cmd := &cobra.Command{
		Use:               "plan-vm NAME",
		Short:             "Describe VM status in a migration plan",
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.PlanNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get plan name from positional argument
			name := args[0]

			// Get the global configuration
			config := getGlobalConfig()

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(config.GetKubeConfigFlags())
			return plan.DescribeVM(config.GetKubeConfigFlags(), name, namespace, vmName, watch, config.GetUseUTC())
		},
	}

	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch VM status with live updates")
	cmd.Flags().StringVar(&vmName, "vm", "", "VM name to describe")

	err := cmd.MarkFlagRequired("vm")
	if err != nil {
		fmt.Printf("Warning: error marking 'vm' flag as required: %v\n", err)
	}

	return cmd
}
