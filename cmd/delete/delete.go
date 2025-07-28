package delete

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewDeleteCmd creates the delete command with all its subcommands
func NewDeleteCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Delete resources",
		Long:         `Delete resources like mappings, plans, and providers`,
		SilenceUsage: true,
	}

	cmd.AddCommand(NewMappingCmd(kubeConfigFlags))
	cmd.AddCommand(NewPlanCmd(kubeConfigFlags))
	cmd.AddCommand(NewProviderCmd(kubeConfigFlags))
	cmd.AddCommand(NewHostCmd(kubeConfigFlags))

	return cmd
}
