package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/version"
	"github.com/yaacov/kubectl-mtv/pkg/util/flags"
)

// NewVersionCmd creates the version command
func NewVersionCmd(clientVersion string, kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	outputFormatFlag := flags.NewOutputFormatTypeFlag()

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  "Print the version information for kubectl-mtv and MTV Operator",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get version information
			versionInfo := version.GetVersionInfo(clientVersion, kubeConfigFlags)

			// Format and output the version information
			output, err := versionInfo.FormatOutput(outputFormatFlag.GetValue())
			if err != nil {
				return err
			}

			fmt.Print(output)
			return nil
		},
	}

	cmd.Flags().VarP(outputFormatFlag, "output", "o", "Output format (json, yaml, table)")

	// Add completion for output format flag
	if err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return outputFormatFlag.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	return cmd
}
