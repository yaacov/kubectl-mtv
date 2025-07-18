package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/plan"
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start resources",
		Long:  `Start various MTV resources`,
	}

	cmd.AddCommand(newStartPlanCmd())
	return cmd
}

func newStartPlanCmd() *cobra.Command {
	var cutoverTimeStr string

	cmd := &cobra.Command{
		Use:   "plan NAME [NAME...]",
		Short: "Start one or more migration plans",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			var cutoverTime *time.Time
			if cutoverTimeStr != "" {
				// Parse the provided cutover time
				t, err := time.Parse(time.RFC3339, cutoverTimeStr)
				if err != nil {
					return fmt.Errorf("failed to parse cutover time: %v", err)
				}
				cutoverTime = &t
			}

			// Loop over each plan name and start it
			for _, name := range args {
				err := plan.Start(kubeConfigFlags, name, namespace, cutoverTime)
				if err != nil {
					printCommandError(err, "starting plan", namespace)
					// Continue with other plans even if one fails
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&cutoverTimeStr, "cutover", "c", "", "Cutover time in ISO8601 format (e.g., 2023-12-31T15:30:00Z, '$(date -d \"+1 hour\" --iso-8601=sec)' ). If not provided, defaults to 1 hour from now.")

	return cmd
}
