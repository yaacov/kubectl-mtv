package get

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/get/inventory"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
	"github.com/yaacov/kubectl-mtv/pkg/util/flags"
)

// NewInventoryNetworkCmd creates the get inventory network command
func NewInventoryNetworkCmd(kubeConfigFlags *genericclioptions.ConfigFlags, globalConfig GlobalConfigGetter) *cobra.Command {
	outputFormatFlag := flags.NewOutputFormatTypeFlag()
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:   "network PROVIDER",
		Short: "Get networks from a provider",
		Long: `Get networks from a provider's inventory.

Queries the MTV inventory service to list networks available in the source provider.
Use --query (-q) to filter results using TSL query syntax.`,
		Example: `  # List all networks from a vSphere provider
  kubectl-mtv get inventory network vsphere-prod

  # Filter networks by name
  kubectl-mtv get inventory network vsphere-prod -q "where name ~= 'VM Network*'"

  # Output as JSON
  kubectl-mtv get inventory network vsphere-prod -o json`,
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.ProviderNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if !watch {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 280*time.Second)
				defer cancel()
			}

			provider := args[0]

			namespace := client.ResolveNamespaceWithAllFlag(globalConfig.GetKubeConfigFlags(), globalConfig.GetAllNamespaces())

			logNamespaceOperation("Getting networks from provider", namespace, globalConfig.GetAllNamespaces())
			logOutputFormat(outputFormatFlag.GetValue())

			// Get inventory URL and insecure skip TLS from global config (auto-discovers if needed)
			inventoryURL := globalConfig.GetInventoryURL()
			inventoryInsecureSkipTLS := globalConfig.GetInventoryInsecureSkipTLS()

			return inventory.ListNetworksWithInsecure(ctx, globalConfig.GetKubeConfigFlags(), provider, namespace, inventoryURL, outputFormatFlag.GetValue(), query, watch, inventoryInsecureSkipTLS)
		},
	}

	cmd.Flags().VarP(outputFormatFlag, "output", "o", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter using TSL syntax (e.g. \"where name ~= 'web-*' and cpuCount > 4\")")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")

	// Add completion for output format flag
	if err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return outputFormatFlag.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	return cmd
}

// NewInventoryStorageCmd creates the get inventory storage command
func NewInventoryStorageCmd(kubeConfigFlags *genericclioptions.ConfigFlags, globalConfig GlobalConfigGetter) *cobra.Command {
	outputFormatFlag := flags.NewOutputFormatTypeFlag()
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:   "storage PROVIDER",
		Short: "Get storage from a provider",
		Long: `Get storage resources from a provider's inventory.

Queries the MTV inventory service to list storage domains (oVirt), datastores (vSphere),
or storage classes (OpenShift) available in the source provider.`,
		Example: `  # List all storage from a vSphere provider
  kubectl-mtv get inventory storage vsphere-prod

  # Filter storage by name pattern
  kubectl-mtv get inventory storage ovirt-prod -q "where name ~= 'data*'"

  # Output as YAML
  kubectl-mtv get inventory storage vsphere-prod -o yaml`,
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.ProviderNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if !watch {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 280*time.Second)
				defer cancel()
			}

			provider := args[0]

			namespace := client.ResolveNamespaceWithAllFlag(globalConfig.GetKubeConfigFlags(), globalConfig.GetAllNamespaces())

			logNamespaceOperation("Getting storage from provider", namespace, globalConfig.GetAllNamespaces())
			logOutputFormat(outputFormatFlag.GetValue())

			// Get inventory URL and insecure skip TLS from global config (auto-discovers if needed)
			inventoryURL := globalConfig.GetInventoryURL()
			inventoryInsecureSkipTLS := globalConfig.GetInventoryInsecureSkipTLS()

			return inventory.ListStorageWithInsecure(ctx, globalConfig.GetKubeConfigFlags(), provider, namespace, inventoryURL, outputFormatFlag.GetValue(), query, watch, inventoryInsecureSkipTLS)
		},
	}

	cmd.Flags().VarP(outputFormatFlag, "output", "o", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter using TSL syntax (e.g. \"where name ~= 'web-*' and cpuCount > 4\")")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")

	// Add completion for output format flag
	if err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return outputFormatFlag.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	return cmd
}

// NewInventoryVMCmd creates the get inventory vm command
func NewInventoryVMCmd(kubeConfigFlags *genericclioptions.ConfigFlags, globalConfig GlobalConfigGetter) *cobra.Command {
	outputFormatFlag := flags.NewVMInventoryOutputTypeFlag()
	var extendedOutput bool
	var query string
	var watch bool

	cmd := &cobra.Command{
		Use:   "vm PROVIDER",
		Short: "Get VMs from a provider",
		Long: `Get virtual machines from a provider's inventory.

Queries the MTV inventory service to list VMs available for migration.
Use --query (-q) to filter results using TSL query syntax. The --extended
flag shows additional VM details.

Output format 'planvms' generates YAML suitable for use with 'create plan --vms @file'.

Query Language (TSL) Syntax:
  Used by -q "where ..." to filter inventory results.

  Structure: [SELECT fields] WHERE condition [ORDER BY field [ASC|DESC]] [LIMIT n]

  Comparison:     =  !=  <>  <  <=  >  >=
  String match:   like (% wildcard), ilike (case-insensitive), ~= (regex), !~ (not regex)
  Logical:        and, or, not
  Set/range:      in ('val1','val2'), between X and Y
  Null checks:    is null, is not null
  Array funcs:    len(field), sum(field.sub), any field.sub > N, all field.sub >= N
  Field access:   dot notation (e.g. parent.id, disks.datastore.id, guest.distribution)

  Discovering available fields:
    Inventory items are dynamic JSON from the provider. Field names vary by provider
    type. To see all available fields, run:
      kubectl-mtv get inventory vm <provider> -o json
    Any field visible in the JSON output can be used in queries with dot notation.

  Common VM fields (vSphere):
    name, id, powerState, cpuCount, memoryMB, guestId, guestName, isTemplate
    firmware, storageUsed, ipAddress, hostName, host, path, uuid, connectionState
    changeTrackingEnabled, coresPerSocket, secureBoot, tpmEnabled
    parent.id, parent.kind (folder reference)
    disks (array): disks.capacity, disks.datastore.id, disks.file, disks.shared
    nics (array): nics.mac, nics.network.id
    networks (array): networks.id, networks.kind
    concerns (array): concerns.category, concerns.assessment, concerns.label

  Common VM fields (oVirt):
    name, id, status, memory, cpuSockets, cpuCores, cpuThreads, osType, guestName
    cluster, host, path, haEnabled, stateless, placementPolicyAffinity, display
    guest.distribution, guest.fullVersion
    diskAttachments (array): diskAttachments.disk, diskAttachments.interface
    nics (array): nics.name, nics.mac, nics.interface, nics.ipAddress, nics.profile
    concerns (array): concerns.category, concerns.assessment, concerns.label

  Common VM fields (OpenStack):
    name, id, status, flavor.name, image.name, project.name

  Common VM fields (EC2, PascalCase):
    name, InstanceType, State.Name, PlatformDetails
    Placement.AvailabilityZone, PublicIpAddress, PrivateIpAddress, VpcId, SubnetId

  Examples:
    where name ~= 'prod-*'
    where powerState = 'poweredOn' and memoryMB > 4096       (vSphere)
    where status = 'up' and memory > 4294967296              (oVirt, memory in bytes)
    where isTemplate = false and firmware = 'efi'            (vSphere)
    where guestId ~= 'rhel.*' and storageUsed > 53687091200
    where len(disks) > 1 and cpuCount <= 8
    where name like '%web%' order by memoryMB desc limit 10
    where path ~= '/Production/.*'`,
		Example: `  # List all VMs from a vSphere provider
  kubectl-mtv get inventory vm vsphere-prod

  # Filter VMs by name pattern
  kubectl-mtv get inventory vm vsphere-prod -q "where name ~= 'web-*'"

  # Get VMs with more than 4 CPUs and 8GB memory
  kubectl-mtv get inventory vm vsphere-prod -q "where cpuCount > 4 and memoryMB > 8192"

  # Show extended VM details
  kubectl-mtv get inventory vm vsphere-prod --extended

  # Export VMs for plan creation
  kubectl-mtv get inventory vm vsphere-prod -q "where name ~= 'prod-*'" -o planvms > vms.yaml
  kubectl-mtv create plan my-migration --vms @vms.yaml`,
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.ProviderNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if !watch {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 280*time.Second)
				defer cancel()
			}

			provider := args[0]

			namespace := client.ResolveNamespaceWithAllFlag(globalConfig.GetKubeConfigFlags(), globalConfig.GetAllNamespaces())

			logNamespaceOperation("Getting VMs from provider", namespace, globalConfig.GetAllNamespaces())
			logOutputFormat(outputFormatFlag.GetValue())

			// Get inventory URL and insecure skip TLS from global config (auto-discovers if needed)
			inventoryURL := globalConfig.GetInventoryURL()
			inventoryInsecureSkipTLS := globalConfig.GetInventoryInsecureSkipTLS()

			return inventory.ListVMsWithInsecure(ctx, globalConfig.GetKubeConfigFlags(), provider, namespace, inventoryURL, outputFormatFlag.GetValue(), extendedOutput, query, watch, inventoryInsecureSkipTLS)
		},
	}

	cmd.Flags().VarP(outputFormatFlag, "output", "o", "Output format (table, json, yaml, planvms)")
	cmd.Flags().BoolVar(&extendedOutput, "extended", false, "Show extended output")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Query filter using TSL syntax (e.g. \"where name ~= 'web-*' and cpuCount > 4\")")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")

	// Custom completion for inventory VM output format that includes planvms
	if err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return outputFormatFlag.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	return cmd
}
