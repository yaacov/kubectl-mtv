package start

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/cmd/get"
)

// NewStartCmd creates the start command with all its subcommands
func NewStartCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() get.GlobalConfigGetter) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "start",
		Short:        "Start resources",
		Long:         `Start various MTV resources`,
		SilenceUsage: true,
	}

	cmd.AddCommand(NewPlanCmd(kubeConfigFlags, getGlobalConfig))
	return cmd
}
