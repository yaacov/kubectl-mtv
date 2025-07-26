package get

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/mapping"
)

// NewMappingCmd creates the get mapping command
func NewMappingCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	var mappingType string
	var outputFormat string

	cmd := &cobra.Command{
		Use:          "mapping [NAME]",
		Short:        "Get mappings",
		Long:         `Get network and storage mappings`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := getGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.GetKubeConfigFlags(), config.GetAllNamespaces())

			// Get optional mapping name from arguments
			var mappingName string
			if len(args) > 0 {
				mappingName = args[0]
			}

			// Log the operation being performed
			if mappingName != "" {
				logNamespaceOperation("Getting mapping", namespace, config.GetAllNamespaces())
			} else {
				logNamespaceOperation("Getting mappings", namespace, config.GetAllNamespaces())
			}
			logOutputFormat(outputFormat)

			return mapping.List(config.GetKubeConfigFlags(), mappingType, namespace, outputFormat, mappingName)
		},
	}

	cmd.Flags().StringVarP(&mappingType, "type", "t", "network", "Mapping type (network, storage)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	addOutputFormatCompletion(cmd, "output")

	// Add completion for mapping type flag
	if err := cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"network", "storage"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	return cmd
}
