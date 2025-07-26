package get

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/flags"
	"github.com/yaacov/kubectl-mtv/pkg/mapping"
)

// NewMappingCmd creates the get mapping command
func NewMappingCmd(kubeConfigFlags *genericclioptions.ConfigFlags, getGlobalConfig func() GlobalConfigGetter) *cobra.Command {
	mappingTypeFlag := flags.NewMappingTypeFlag()
	outputFormatFlag := flags.NewOutputFormatTypeFlag()

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
			logOutputFormat(outputFormatFlag.GetValue())

			return mapping.List(config.GetKubeConfigFlags(), mappingTypeFlag.GetValue(), namespace, outputFormatFlag.GetValue(), mappingName)
		},
	}

	cmd.Flags().VarP(mappingTypeFlag, "type", "t", "Mapping type (network, storage)")
	cmd.Flags().VarP(outputFormatFlag, "output", "o", "Output format (table, json, yaml)")

	// Add completion for mapping type flag
	if err := cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return mappingTypeFlag.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	// Add completion for output format flag
	if err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return outputFormatFlag.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	return cmd
}
