package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/plan"
	"gopkg.in/yaml.v3"
)

func newCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "cancel",
		Short:        "Cancel resources",
		Long:         `Cancel various MTV resources`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newCancelVMsCmd())
	return cmd
}

func newCancelVMsCmd() *cobra.Command {
	var vmNamesOrFile string

	cmd := &cobra.Command{
		Use:          "plan NAME",
		Short:        "Cancel specific VMs in a running migration plan",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get plan name from positional argument
			planName := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			var vmNames []string

			if strings.HasPrefix(vmNamesOrFile, "@") {
				// It's a file
				filePath := vmNamesOrFile[1:]
				content, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read file %s: %v", filePath, err)
				}

				// Try to unmarshal as JSON or YAML array of strings
				var namesArray []string
				if err := json.Unmarshal(content, &namesArray); err != nil {
					if err := yaml.Unmarshal(content, &namesArray); err != nil {
						return fmt.Errorf("failed to parse VM names from file: %v", err)
					}
				}
				vmNames = namesArray
			} else {
				// It's a comma-separated list
				vmNameSlice := strings.Split(vmNamesOrFile, ",")
				for _, vmName := range vmNameSlice {
					vmNames = append(vmNames, strings.TrimSpace(vmName))
				}
			}

			if len(vmNames) == 0 {
				return fmt.Errorf("no VM names specified to cancel")
			}

			return plan.Cancel(kubeConfigFlags, planName, namespace, vmNames)
		},
	}

	cmd.Flags().StringVar(&vmNamesOrFile, "vms", "", "List of VM names to cancel (comma-separated) or path to file containing VM names (prefix with @)")

	if err := cmd.MarkFlagRequired("vms"); err != nil {
		fmt.Printf("Warning: error marking 'vms' flag as required: %v\n", err)
	}

	return cmd
}
