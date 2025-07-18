package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set via ldflags during build
	clientVersion = "unknown"
)

// newVersionCmd creates a new version command
func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  "Print the version information for kubectl-mtv",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("kubectl-mtv version: %s\n", clientVersion)
		},
	}

	return cmd
}
