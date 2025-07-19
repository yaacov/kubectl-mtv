package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/inventory"
	"github.com/yaacov/kubectl-mtv/pkg/mapping"
	"github.com/yaacov/kubectl-mtv/pkg/plan"
	"github.com/yaacov/kubectl-mtv/pkg/provider"
)

// getOutputFormatCompletions returns valid output format options for completion
func getOutputFormatCompletions() []string {
	return []string{"table", "json", "yaml"}
}

// addOutputFormatCompletion adds completion for output format flags
func addOutputFormatCompletion(cmd *cobra.Command, flagName string) {
	if err := cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getOutputFormatCompletions(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}
}

// printCommandError provides consistent error messaging across commands
// It prints helpful error information when an error occurs
func printCommandError(err error, operation string, namespace string) {
	fmt.Printf("Error %s: %v\n", operation, err)
	fmt.Printf("You can use the '--help' flag for more information on usage.\n")
}

// logNamespaceOperation logs namespace-specific operations with consistent formatting
func logNamespaceOperation(operation string, namespace string, allNamespaces bool) {
	if allNamespaces {
		logInfof("%s from all namespaces", operation)
	} else {
		logInfof("%s from namespace: %s", operation, namespace)
	}
}

// logOutputFormat logs the output format being used
func logOutputFormat(format string) {
	logDebugf("Output format: %s", format)
}

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get resources",
		Long:  `Get various MTV resources including plans, providers, mappings, and inventory`,
	}

	// Add plan subcommand with plural alias
	planCmd := newGetPlanCmd()
	planCmd.Aliases = []string{"plans"}
	cmd.AddCommand(planCmd)

	// Add plan-vms subcommand
	cmd.AddCommand(newGetPlanVMsCmd())

	// Add provider subcommand with plural alias
	providerCmd := newGetProviderCmd()
	providerCmd.Aliases = []string{"providers"}
	cmd.AddCommand(providerCmd)

	// Add mapping subcommand with plural alias
	mappingCmd := newGetMappingCmd()
	mappingCmd.Aliases = []string{"mappings"}
	cmd.AddCommand(mappingCmd)

	// Add inventory subcommand
	cmd.AddCommand(newGetInventoryCmd())

	return cmd
}

func newGetPlanCmd() *cobra.Command {
	var outputFormat string
	var watch bool

	cmd := &cobra.Command{
		Use:   "plan [NAME]",
		Short: "Get migration plans",
		Long:  `Get migration plans`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Get optional plan name from arguments
			var planName string
			if len(args) > 0 {
				planName = args[0]
			}

			// Log the operation being performed
			if planName != "" {
				logNamespaceOperation("Getting plan", namespace, config.AllNamespaces)
			} else {
				logNamespaceOperation("Getting plans", namespace, config.AllNamespaces)
			}
			logOutputFormat(outputFormat)

			err := plan.List(config.KubeConfigFlags, namespace, watch, outputFormat, planName)
			if err != nil {
				printCommandError(err, "getting plans", namespace)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetPlanVMsCmd() *cobra.Command {
	var watch bool

	cmd := &cobra.Command{
		Use:   "plan-vms NAME",
		Short: "Get VMs in a migration plan",
		Long:  `Get VMs in a migration plan`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get plan name from positional argument
			name := args[0]

			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting VMs from plan", namespace, config.AllNamespaces)

			err := plan.ListVMs(config.KubeConfigFlags, name, namespace, watch)
			if err != nil {
				printCommandError(err, "getting VMs from plan", namespace)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch VM status with live updates")

	return cmd
}

func newGetProviderCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "provider [NAME]",
		Short: "Get providers",
		Long:  `Get virtualization providers`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Get optional provider name from arguments
			var providerName string
			if len(args) > 0 {
				providerName = args[0]
			}

			// Log the operation being performed
			if providerName != "" {
				logNamespaceOperation("Getting provider", namespace, config.AllNamespaces)
			} else {
				logNamespaceOperation("Getting providers", namespace, config.AllNamespaces)
			}
			logOutputFormat(outputFormat)

			baseURL := ""
			err := provider.List(config.KubeConfigFlags, namespace, baseURL, outputFormat, providerName)
			if err != nil {
				printCommandError(err, "getting providers", namespace)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetMappingCmd() *cobra.Command {
	var outputFormat string
	var mappingType string

	cmd := &cobra.Command{
		Use:   "mapping [NAME]",
		Short: "Get mappings",
		Long:  `Get network and storage mappings`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Get optional mapping name from arguments
			var mappingName string
			if len(args) > 0 {
				mappingName = args[0]
			}

			// Log the operation being performed
			if mappingName != "" {
				logNamespaceOperation("Getting mapping", namespace, config.AllNamespaces)
			} else {
				logNamespaceOperation("Getting mappings", namespace, config.AllNamespaces)
			}
			logOutputFormat(outputFormat)

			err := mapping.List(config.KubeConfigFlags, mappingType, namespace, outputFormat, mappingName)
			if err != nil {
				printCommandError(err, "getting mappings", namespace)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&mappingType, "type", "t", "", "Mapping type (network, storage, all)")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inventory",
		Short: "Get inventory resources",
		Long:  `Get inventory resources from providers`,
	}

	// Add host subcommand with plural alias
	hostCmd := newGetInventoryHostCmd()
	hostCmd.Aliases = []string{"hosts"}
	cmd.AddCommand(hostCmd)

	// Add namespace subcommand with plural alias
	namespaceCmd := newGetInventoryNamespaceCmd()
	namespaceCmd.Aliases = []string{"namespaces"}
	cmd.AddCommand(namespaceCmd)

	// Add network subcommand with plural alias
	networkCmd := newGetInventoryNetworkCmd()
	networkCmd.Aliases = []string{"networks"}
	cmd.AddCommand(networkCmd)

	// Add storage subcommand with plural alias
	storageCmd := newGetInventoryStorageCmd()
	storageCmd.Aliases = []string{"storages"}
	cmd.AddCommand(storageCmd)

	// Add vm subcommand with plural alias
	vmCmd := newGetInventoryVMCmd()
	vmCmd.Aliases = []string{"vms"}
	cmd.AddCommand(vmCmd)

	return cmd
}

func newGetInventoryHostCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:   "host PROVIDER",
		Short: "Get hosts from a provider",
		Long:  `Get hosts from a provider`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting hosts from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			err := inventory.ListHosts(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
			if err != nil {
				printCommandError(err, "getting hosts from provider", namespace)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryNamespaceCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:   "namespace PROVIDER",
		Short: "Get namespaces from a provider",
		Long:  `Get namespaces from a provider`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting namespaces from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			err := inventory.ListNamespaces(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
			if err != nil {
				printCommandError(err, "getting namespaces from provider", namespace)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryNetworkCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:   "network PROVIDER",
		Short: "Get networks from a provider",
		Long:  `Get networks from a provider`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting networks from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			err := inventory.ListNetworks(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
			if err != nil {
				printCommandError(err, "getting networks from provider", namespace)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryStorageCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:   "storage PROVIDER",
		Short: "Get storage from a provider",
		Long:  `Get storage from a provider`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting storage from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			err := inventory.ListStorage(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
			if err != nil {
				printCommandError(err, "getting storage from provider", namespace)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryVMCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var extendedOutput bool
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:   "vm PROVIDER",
		Short: "Get VMs from a provider",
		Long:  `Get VMs from a provider`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting VMs from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			err := inventory.ListVMs(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, extendedOutput, query, watch)
			if err != nil {
				printCommandError(err, "getting VMs from provider", namespace)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml, planvms)")
	cmd.Flags().BoolVar(&extendedOutput, "extended", false, "Show extended output")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")

	// Custom completion for inventory VM output format that includes planvms
	if err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json", "yaml", "planvms"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	return cmd
}
