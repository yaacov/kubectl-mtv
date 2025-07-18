package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	kubeConfigFlags *genericclioptions.ConfigFlags
	rootCmd         *cobra.Command
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	kubeConfigFlags = genericclioptions.NewConfigFlags(true)

	rootCmd = &cobra.Command{
		Use:   "kubectl-mtv",
		Short: "Migration Toolkit for Virtualization CLI",
		Long: `Migration Toolkit for Virtualization (MTV) CLI
A kubectl plugin for migrating VMs from oVirt, VMware, OpenStack, and OVA files to KubeVirt.`,
	}

	kubeConfigFlags.AddFlags(rootCmd.PersistentFlags())

	// Add standard commands for various resources
	rootCmd.AddCommand(newGetCmd())
	rootCmd.AddCommand(newDeleteCmd())
	rootCmd.AddCommand(newCreateCmd())
	rootCmd.AddCommand(newDescribeCmd())

	// Plan commands
	rootCmd.AddCommand(newStartCmd())
	rootCmd.AddCommand(newCancelCmd())
	rootCmd.AddCommand(newCutoverCmd())
	rootCmd.AddCommand(newArchiveCmd())
	rootCmd.AddCommand(newUnArchiveCmd())

	// Version command
	rootCmd.AddCommand(newVersionCmd())
}
