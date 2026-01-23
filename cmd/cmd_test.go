package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestAllCommandsHelp verifies that all commands can render help without panicking.
// This catches flag shorthand conflicts that cause panics when cobra merges flag sets.
// For example, if a local flag uses "-i" shorthand but a global flag already uses "-i",
// cobra will panic with "unable to redefine shorthand".
func TestAllCommandsHelp(t *testing.T) {
	// Get the root command (this triggers all subcommand registration)
	root := rootCmd

	// Recursively test all commands
	var testCommand func(cmd *cobra.Command)
	testCommand = func(cmd *cobra.Command) {
		t.Run(cmd.CommandPath(), func(t *testing.T) {
			// This will panic if there are flag conflicts
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("command %q panicked: %v", cmd.CommandPath(), r)
				}
			}()

			// Trigger flag merging by getting the flag set
			cmd.Flags()
			cmd.InheritedFlags()

			// Try to generate help (this also triggers flag registration)
			_ = cmd.UsageString()
		})

		// Recursively test subcommands
		for _, sub := range cmd.Commands() {
			testCommand(sub)
		}
	}

	testCommand(root)
}
