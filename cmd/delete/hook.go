package delete

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/delete/hook"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
)

// NewHookCmd creates the delete hook command
func NewHookCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "hook NAME [NAME...]",
		Short:             "Delete one or more migration hooks",
		Args:              cobra.MinimumNArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.HookResourceNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Loop over each hook name and delete it
			for _, name := range args {
				err := hook.Delete(kubeConfigFlags, name, namespace)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	return cmd
}
