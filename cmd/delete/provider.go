package delete

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/delete/provider"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
)

// NewProviderCmd creates the provider deletion command
func NewProviderCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "provider NAME [NAME...]",
		Short:             "Delete one or more providers",
		Args:              cobra.MinimumNArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.ProviderNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Loop over each provider name and delete it
			for _, name := range args {
				err := provider.Delete(kubeConfigFlags, name, namespace)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	return cmd
}
