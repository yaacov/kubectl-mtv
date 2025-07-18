package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/plan"
)

func newDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe resources",
		Long:  `Describe migration plans and VMs in migration plans`,
	}

	cmd.AddCommand(newDescribePlanSubCmd())
	cmd.AddCommand(newDescribeVMSubCmd())

	return cmd
}

func newDescribePlanSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan NAME",
		Short: "Describe a migration plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			err := plan.Describe(kubeConfigFlags, name, namespace)
			if err != nil {
				printCommandError(err, "describing plan", namespace)
			}
			return nil
		},
	}

	return cmd
}

func newDescribeVMSubCmd() *cobra.Command {
	var watch bool
	var vmName string

	cmd := &cobra.Command{
		Use:   "plan-vm NAME",
		Short: "Describe VM status in a migration plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get plan name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			err := plan.DescribeVM(kubeConfigFlags, name, namespace, vmName, watch)
			if err != nil {
				printCommandError(err, "describing VM in plan", namespace)
			}
			return nil
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
