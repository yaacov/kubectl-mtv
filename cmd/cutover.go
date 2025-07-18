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
		Use:   "plan NAME",
		Short: "Set the cutover time for a warm migration plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get plan name from positional argument
			planName := args[0]

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

			err := plan.Cutover(kubeConfigFlags, planName, namespace, cutoverTime)
			if err != nil {
				printCommandError(err, "setting cutover time for plan", namespace)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&cutoverTimeStr, "cutover", "c", "", "Cutover time in ISO8601 format (e.g., 2023-12-31T15:30:00Z, '$(date --iso-8601=sec)'). If not specified, defaults to current time.")

	return cmd
}
