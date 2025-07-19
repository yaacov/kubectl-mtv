package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/plan"
)

func newCutoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cutover",
		Short: "Set cutover time for resources",
		Long:  `Set cutover time for various MTV resources`,
	}

	cmd.AddCommand(newCutoverPlanCmd())
	return cmd
}

func newCutoverPlanCmd() *cobra.Command {
	var cutoverTimeStr string

	cmd := &cobra.Command{
		Use:   "plan NAME [NAME...]",
		Short: "Set the cutover time for one or more warm migration plans",
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

			// Loop over each plan name and set cutover time
			for _, planName := range args {
				err := plan.Cutover(kubeConfigFlags, planName, namespace, cutoverTime)
				if err != nil {
					printCommandError(err, "setting cutover time for plan", namespace)
					// Continue with other plans even if one fails
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&cutoverTimeStr, "cutover", "c", "", "Cutover time in ISO8601 format (e.g., 2023-12-31T15:30:00Z, '$(date --iso-8601=sec)'). If not specified, defaults to current time.")

	return cmd
}
