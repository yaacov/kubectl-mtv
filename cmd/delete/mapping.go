package delete

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/delete/mapping"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
)

// NewMappingCmd creates the mapping deletion command
func NewMappingCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	var mappingType string

	cmd := &cobra.Command{
		Use:          "mapping NAME [NAME...]",
		Short:        "Delete one or more mappings",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Get the mapping type from the flag
			mappingType, _ := cmd.Flags().GetString("type")
			return completion.MappingNameCompletion(kubeConfigFlags, mappingType)(cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Loop over each mapping name and delete it
			for _, name := range args {
				err := mapping.Delete(kubeConfigFlags, name, namespace, mappingType)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&mappingType, "type", "network", "Mapping type (network, storage)")

	return cmd
}
