package describe

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/describe/host"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
)

// NewHostCmd creates the host description command
func NewHostCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "host NAME",
		Short:             "Describe a migration host",
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.HostResourceNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			return host.Describe(kubeConfigFlags, name, namespace)
		},
	}

	return cmd
}
