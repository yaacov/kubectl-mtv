package unarchive

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewUnArchiveCmd creates the unarchive command with all its subcommands
func NewUnArchiveCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "unarchive",
		Short:        "Un-archive resources",
		Long:         `Un-archive various MTV resources`,
		SilenceUsage: true,
	}

	cmd.AddCommand(NewPlanCmd(kubeConfigFlags))
	return cmd
}
