package delete

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/delete/host"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
)

// NewHostCmd creates the delete host command
func NewHostCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "host NAME [NAME...]",
		Short:             "Delete one or more migration hosts",
		Args:              cobra.MinimumNArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.HostResourceNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Loop over each host name and delete it
			for _, name := range args {
				err := host.Delete(kubeConfigFlags, name, namespace)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	return cmd
}
