package cmd

import (
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
		Use:          "get",
		Short:        "Get resources",
		Long:         `Get various MTV resources including plans, providers, mappings, and inventory`,
		SilenceUsage: true,
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
		Use:          "plan [NAME]",
		Short:        "Get migration plans",
		Long:         `Get migration plans`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
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

			return plan.List(config.KubeConfigFlags, namespace, watch, outputFormat, planName)
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
		Use:          "plan-vms NAME",
		Short:        "Get VMs in a migration plan",
		Long:         `Get VMs in a migration plan`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get plan name from positional argument
			name := args[0]

			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting VMs from plan", namespace, config.AllNamespaces)

			return plan.ListVMs(config.KubeConfigFlags, name, namespace, watch)
		},
	}

	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch VM status with live updates")

	return cmd
}

func newGetProviderCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:          "provider [NAME]",
		Short:        "Get providers",
		Long:         `Get virtualization providers`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
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
			return provider.List(config.KubeConfigFlags, namespace, baseURL, outputFormat, providerName)
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
		Use:          "mapping [NAME]",
		Short:        "Get mappings",
		Long:         `Get network and storage mappings`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
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

			return mapping.List(config.KubeConfigFlags, mappingType, namespace, outputFormat, mappingName)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&mappingType, "type", "t", "", "Mapping type (network, storage, all)")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "inventory",
		Short:        "Get inventory resources",
		Long:         `Get inventory resources from providers`,
		SilenceUsage: true,
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

	// Add datacenter subcommand with plural alias
	datacenterCmd := newGetInventoryDataCenterCmd()
	datacenterCmd.Aliases = []string{"datacenters"}
	cmd.AddCommand(datacenterCmd)

	// Add cluster subcommand with plural alias
	clusterCmd := newGetInventoryClusterCmd()
	clusterCmd.Aliases = []string{"clusters"}
	cmd.AddCommand(clusterCmd)

	// Add disk subcommand with plural alias
	diskCmd := newGetInventoryDiskCmd()
	diskCmd.Aliases = []string{"disks"}
	cmd.AddCommand(diskCmd)

	// Add disk profile subcommand with plural alias
	diskProfileCmd := newGetInventoryDiskProfileCmd()
	diskProfileCmd.Aliases = []string{"diskprofiles", "disk-profiles"}
	cmd.AddCommand(diskProfileCmd)

	// Add NIC profile subcommand with plural alias
	nicProfileCmd := newGetInventoryNICProfileCmd()
	nicProfileCmd.Aliases = []string{"nicprofiles", "nic-profiles"}
	cmd.AddCommand(nicProfileCmd)

	// Add OpenStack-specific resources
	instanceCmd := newGetInventoryInstanceCmd()
	instanceCmd.Aliases = []string{"instances"}
	cmd.AddCommand(instanceCmd)

	imageCmd := newGetInventoryImageCmd()
	imageCmd.Aliases = []string{"images"}
	cmd.AddCommand(imageCmd)

	flavorCmd := newGetInventoryFlavorCmd()
	flavorCmd.Aliases = []string{"flavors"}
	cmd.AddCommand(flavorCmd)

	projectCmd := newGetInventoryProjectCmd()
	projectCmd.Aliases = []string{"projects"}
	cmd.AddCommand(projectCmd)

	// Add vSphere-specific resources
	datastoreCmd := newGetInventoryDatastoreCmd()
	datastoreCmd.Aliases = []string{"datastores"}
	cmd.AddCommand(datastoreCmd)

	resourcePoolCmd := newGetInventoryResourcePoolCmd()
	resourcePoolCmd.Aliases = []string{"resourcepools", "resource-pools"}
	cmd.AddCommand(resourcePoolCmd)

	folderCmd := newGetInventoryFolderCmd()
	folderCmd.Aliases = []string{"folders"}
	cmd.AddCommand(folderCmd)

	pvcCmd := newGetInventoryPVCCmd()
	pvcCmd.Aliases = []string{"pvcs", "persistentvolumeclaims"}
	cmd.AddCommand(pvcCmd)

	dataVolumeCmd := newGetInventoryDataVolumeCmd()
	dataVolumeCmd.Aliases = []string{"datavolumes", "data-volumes"}
	cmd.AddCommand(dataVolumeCmd)

	return cmd
}

func newGetInventoryHostCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "host PROVIDER",
		Short:        "Get hosts from a provider (ovirt, vsphere)",
		Long:         `Get hosts from a provider (ovirt, vsphere)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
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

			return inventory.ListHosts(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
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
		Use:          "namespace PROVIDER",
		Short:        "Get namespaces from a provider (openshift, openstack)",
		Long:         `Get namespaces from a provider (openshift, openstack)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
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

			return inventory.ListNamespaces(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
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
		Use:          "network PROVIDER",
		Short:        "Get networks from a provider (ovirt, vsphere, openstack, ova, openshift)",
		Long:         `Get networks from a provider (ovirt, vsphere, openstack, ova, openshift)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
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

			return inventory.ListNetworks(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
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
		Use:          "storage PROVIDER",
		Short:        "Get storage from a provider (ovirt, vsphere, ova, openstack, openshift)",
		Long:         `Get storage from a provider (ovirt, vsphere, ova, openstack, openshift)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
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

			return inventory.ListStorage(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
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
		Use:          "vm PROVIDER",
		Short:        "Get VMs from a provider (ovirt, vsphere, openstack, ova, openshift)",
		Long:         `Get VMs from a provider (ovirt, vsphere, openstack, ova, openshift)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
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

			return inventory.ListVMs(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, extendedOutput, query, watch)
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

func newGetInventoryDataCenterCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "datacenter PROVIDER",
		Short:        "Get datacenters from a provider (ovirt, vsphere)",
		Long:         `Get datacenters from a provider (ovirt, vsphere)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting datacenters from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListDataCenters(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryClusterCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "cluster PROVIDER",
		Short:        "Get clusters from a provider (ovirt, vsphere)",
		Long:         `Get clusters from a provider (ovirt, vsphere)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting clusters from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListClusters(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryDiskCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "disk PROVIDER",
		Short:        "Get disks from a provider (ovirt, openstack, ova)",
		Long:         `Get disks from a provider (ovirt, openstack, ova)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting disks from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListDisks(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryDiskProfileCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "disk-profile PROVIDER",
		Short:        "Get disk profiles from a provider (ovirt)",
		Long:         `Get disk profiles from a provider (ovirt)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting disk profiles from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListDiskProfiles(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryNICProfileCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "nic-profile PROVIDER",
		Short:        "Get NIC profiles from a provider (ovirt)",
		Long:         `Get NIC profiles from a provider (ovirt)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting NIC profiles from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListNICProfiles(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryInstanceCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "instance PROVIDER",
		Short:        "Get instances from a provider (openstack)",
		Long:         `Get instances from a provider (openstack)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting instances from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListInstances(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryImageCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "image PROVIDER",
		Short:        "Get images from a provider (openstack)",
		Long:         `Get images from a provider (openstack)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting images from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListImages(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryFlavorCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "flavor PROVIDER",
		Short:        "Get flavors from a provider (openstack)",
		Long:         `Get flavors from a provider (openstack)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting flavors from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListFlavors(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryProjectCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "project PROVIDER",
		Short:        "Get projects from a provider (openstack)",
		Long:         `Get projects from a provider (openstack)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting projects from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListProjects(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryDatastoreCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "datastore PROVIDER",
		Short:        "Get datastores from a provider (vsphere)",
		Long:         `Get datastores from a provider (vsphere)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting datastores from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListDatastores(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryResourcePoolCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "resource-pool PROVIDER",
		Short:        "Get resource pools from a provider (vsphere)",
		Long:         `Get resource pools from a provider (vsphere)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting resource pools from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListResourcePools(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryFolderCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "folder PROVIDER",
		Short:        "Get folders from a provider (vsphere)",
		Long:         `Get folders from a provider (vsphere)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting folders from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListFolders(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryPVCCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "pvc PROVIDER",
		Short:        "Get persistent volume claims from a provider (openshift)",
		Long:         `Get persistent volume claims from a provider (openshift)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting PVCs from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListPersistentVolumeClaims(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}

func newGetInventoryDataVolumeCmd() *cobra.Command {
	var inventoryURL string
	var outputFormat string
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:          "data-volume PROVIDER",
		Short:        "Get data volumes from a provider (openshift)",
		Long:         `Get data volumes from a provider (openshift)`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			config := GetGlobalConfig()
			namespace := client.ResolveNamespaceWithAllFlag(config.KubeConfigFlags, config.AllNamespaces)

			// Log the operation being performed
			logNamespaceOperation("Getting data volumes from provider", namespace, config.AllNamespaces)
			logOutputFormat(outputFormat)

			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(config.KubeConfigFlags, namespace)
			}

			return inventory.ListDataVolumes(config.KubeConfigFlags, provider, namespace, inventoryURL, outputFormat, query, watch)
		},
	}

	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Inventory service URL")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
	addOutputFormatCompletion(cmd, "output")

	return cmd
}
