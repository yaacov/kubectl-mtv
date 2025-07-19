package cmd

import (
	"flag"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
)

// GlobalConfig holds global configuration flags that are passed to all subcommands
type GlobalConfig struct {
	Verbosity       int
	AllNamespaces   bool
	KubeConfigFlags *genericclioptions.ConfigFlags
}

var (
	kubeConfigFlags *genericclioptions.ConfigFlags
	rootCmd         *cobra.Command
	globalConfig    *GlobalConfig
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

// GetGlobalConfig returns the global configuration instance
func GetGlobalConfig() *GlobalConfig {
	return globalConfig
}

func init() {
	kubeConfigFlags = genericclioptions.NewConfigFlags(true)

	// Initialize global configuration
	globalConfig = &GlobalConfig{
		KubeConfigFlags: kubeConfigFlags,
	}

	rootCmd = &cobra.Command{
		Use:   "kubectl-mtv",
		Short: "Migration Toolkit for Virtualization CLI",
		Long: `Migration Toolkit for Virtualization (MTV) CLI
A kubectl plugin for migrating VMs from oVirt, VMware, OpenStack, and OVA files to KubeVirt.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Initialize klog with the verbosity level
			klog.InitFlags(nil)
			if err := flag.Set("v", fmt.Sprintf("%d", globalConfig.Verbosity)); err != nil {
				klog.Warningf("Failed to set klog verbosity: %v", err)
			}

			// Log global configuration if verbosity is enabled
			klog.V(2).Infof("Global configuration - Verbosity: %d, All Namespaces: %t",
				globalConfig.Verbosity, globalConfig.AllNamespaces)
		},
	}

	kubeConfigFlags.AddFlags(rootCmd.PersistentFlags())

	// Add global flags
	rootCmd.PersistentFlags().IntVarP(&globalConfig.Verbosity, "verbose", "v", 0, "verbose output level (0=silent, 1=info, 2=debug, 3=trace)")
	rootCmd.PersistentFlags().BoolVarP(&globalConfig.AllNamespaces, "all-namespaces", "A", false, "list resources across all namespaces")

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
