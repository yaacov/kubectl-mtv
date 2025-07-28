package describe

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewDescribeCmd creates the describe command with all its subcommands
func NewDescribeCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "describe",
		Short:        "Describe resources",
		Long:         `Describe migration plans and VMs in migration plans`,
		SilenceUsage: true,
	}

	cmd.AddCommand(NewPlanCmd(kubeConfigFlags))
	cmd.AddCommand(NewVMCmd(kubeConfigFlags))
	cmd.AddCommand(NewHostCmd(kubeConfigFlags))
	cmd.AddCommand(NewMappingCmd(kubeConfigFlags))

	return cmd
}
