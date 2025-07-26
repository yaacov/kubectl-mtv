package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/version"
)

// NewVersionCmd creates the version command
func NewVersionCmd(clientVersion string, kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  "Print the version information for kubectl-mtv and MTV Operator",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get version information
			versionInfo := version.GetVersionInfo(clientVersion, kubeConfigFlags)

			// Format and output the version information
			output, err := versionInfo.FormatOutput(outputFormat)
			if err != nil {
				return err
			}

			fmt.Print(output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format: json, yaml, or table (default)")

	return cmd
}
