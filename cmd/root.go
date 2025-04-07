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
		Use:   "mtv",
		Short: "Migration Toolkit for Virtualization CLI",
		Long: `Migration Toolkit for Virtualization (MTV) CLI
A kubectl plugin for migrating VMs from oVirt, VMware, OpenStack, and OVA files to KubeVirt.`,
	}

	kubeConfigFlags.AddFlags(rootCmd.PersistentFlags())

	rootCmd.AddCommand(newProviderCmd())
	rootCmd.AddCommand(newMappingCmd())
	rootCmd.AddCommand(newPlanCmd())
	rootCmd.AddCommand(newInventoryCmd())
}
