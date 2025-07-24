package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/version"
)

var (
	// Version is set via ldflags during build
	clientVersion = "unknown"
)

// newVersionCmd creates a new version command
func newVersionCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  "Print the version information for kubectl-mtv and MTV Operator",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := GetGlobalConfig()

			// Get version information
			versionInfo := version.GetVersionInfo(clientVersion, config.KubeConfigFlags)

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
