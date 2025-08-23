package create

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	planv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/plan"
	"github.com/spf13/cobra"
	"github.com/yaacov/karl-interpreter/pkg/karl"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
	"github.com/yaacov/kubectl-mtv/pkg/util/flags"
)

// parseKeyValuePairs parses a slice of strings containing comma-separated key=value pairs
// and returns a map[string]string with trimmed keys and values
func parseKeyValuePairs(pairs []string, fieldName string) (map[string]string, error) {
	result := make(map[string]string)
	for _, pairGroup := range pairs {
		// Split by comma to handle multiple pairs in one flag value
		keyValuePairs := strings.Split(pairGroup, ",")
		for _, pair := range keyValuePairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				result[key] = value
			} else {
				return nil, fmt.Errorf("invalid %s: %s", fieldName, pair)
			}
		}
	}
	return result, nil
}

// NewPlanCmd creates the plan creation command
func NewPlanCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	var name, sourceProvider, targetProvider string
	var networkMapping, storageMapping string
	var vmNamesOrFile string
	var inventoryURL string
	var defaultTargetNetwork, defaultTargetStorageClass string
	var networkPairs, storagePairs string
	var preHook, postHook string

	// Storage mapping enhancement options
	var defaultVolumeMode, defaultAccessMode string
	var defaultOffloadPlugin, defaultOffloadSecret, defaultOffloadVendor string

	// PlanSpec fields
	var planSpec forkliftv1beta1.PlanSpec
	var transferNetwork string
	var installLegacyDrivers string // "true", "false", or "" for nil
	migrationTypeFlag := flags.NewMigrationTypeFlag()
	var targetLabels []string
	var targetNodeSelector []string
	var useCompatibilityMode bool
	var targetAffinity string
	var targetPowerState string

	cmd := &cobra.Command{
		Use:          "plan NAME",
		Short:        "Create a migration plan",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name = args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// If inventoryURL is empty, try to discover it
			if inventoryURL == "" {
				inventoryURL = client.DiscoverInventoryURL(kubeConfigFlags, namespace)
			}

			// Validate that existing mapping flags and mapping pair flags are not used together
			if networkMapping != "" && networkPairs != "" {
				return fmt.Errorf("cannot use both --network-mapping and --network-pairs flags")
			}
			if storageMapping != "" && storagePairs != "" {
				return fmt.Errorf("cannot use both --storage-mapping and --storage-pairs flags")
			}

			// Validate that conversion-only migrations don't use storage mappings
			if migrationTypeFlag.GetValue() == "conversion" {
				if storageMapping != "" {
					return fmt.Errorf("cannot use --storage-mapping with migration type 'conversion'. Conversion-only migrations require empty storage mapping")
				}
				if storagePairs != "" {
					return fmt.Errorf("cannot use --storage-pairs with migration type 'conversion'. Conversion-only migrations require empty storage mapping")
				}
			}

			var vmList []planv1beta1.VM

			if strings.HasPrefix(vmNamesOrFile, "@") {
				// It's a file
				filePath := vmNamesOrFile[1:]
				content, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read file %s: %v", filePath, err)
				}

				// Attempt to unmarshal as YAML first, then try JSON
				err = yaml.Unmarshal(content, &vmList)
				if err != nil {
					err = json.Unmarshal(content, &vmList)
					if err != nil {
						return fmt.Errorf("failed to unmarshal file %s as YAML or JSON: %v", filePath, err)
					}
				}
			} else {
				// It's a comma-separated list
				vmNameSlice := strings.Split(vmNamesOrFile, ",")
				for _, vmName := range vmNameSlice {
					newVM := planv1beta1.VM{}
					newVM.Name = strings.TrimSpace(vmName)
					vmList = append(vmList, newVM)
				}
			}

			// Add hooks to all VMs if specified
			if preHook != "" || postHook != "" {
				for i := range vmList {
					var hooks []planv1beta1.HookRef

					// Add pre-hook if specified
					if preHook != "" {
						preHookRef := planv1beta1.HookRef{
							Step: "PreHook",
							Hook: corev1.ObjectReference{
								Kind:       "Hook",
								APIVersion: "forklift.konveyor.io/v1beta1",
								Name:       strings.TrimSpace(preHook),
								Namespace:  namespace,
							},
						}
						hooks = append(hooks, preHookRef)
					}

					// Add post-hook if specified
					if postHook != "" {
						postHookRef := planv1beta1.HookRef{
							Step: "PostHook",
							Hook: corev1.ObjectReference{
								Kind:       "Hook",
								APIVersion: "forklift.konveyor.io/v1beta1",
								Name:       strings.TrimSpace(postHook),
								Namespace:  namespace,
							},
						}
						hooks = append(hooks, postHookRef)
					}

					// Add hooks to the VM (append to existing hooks if any)
					vmList[i].Hooks = append(vmList[i].Hooks, hooks...)
				}
			}

			// Create transfer network reference if provided
			if transferNetwork != "" {
				transferNetworkName := strings.TrimSpace(transferNetwork)
				transferNetworkNamespace := namespace

				// If tansferNetwork has "/", the first part is the namespace
				if strings.Contains(transferNetwork, "/") {
					parts := strings.SplitN(transferNetwork, "/", 2)
					transferNetworkName = strings.TrimSpace(parts[1])
					transferNetworkNamespace = strings.TrimSpace(parts[0])
				}

				planSpec.TransferNetwork = &corev1.ObjectReference{
					Kind:       "NetworkAttachmentDefinition",
					APIVersion: "k8s.cni.cncf.io/v1",
					Name:       transferNetworkName,
					Namespace:  transferNetworkNamespace,
				}
			}

			// Handle InstallLegacyDrivers flag
			if installLegacyDrivers != "" {
				switch installLegacyDrivers {
				case "true":
					val := true
					planSpec.InstallLegacyDrivers = &val
				case "false":
					val := false
					planSpec.InstallLegacyDrivers = &val
				}
			}

			// Handle migration type flag
			if migrationTypeFlag.GetValue() != "" {
				if planSpec.Warm {
					return fmt.Errorf("setting --warm flag is not supported when migration type is specified")
				}

				planSpec.Type = migrationTypeFlag.GetValue()
			}

			// Handle target labels (convert from key=value slice to map)
			if len(targetLabels) > 0 {
				labels, err := parseKeyValuePairs(targetLabels, "target label")
				if err != nil {
					return err
				}
				planSpec.TargetLabels = labels
			}

			// Handle target node selector (convert from key=value slice to map)
			if len(targetNodeSelector) > 0 {
				nodeSelector, err := parseKeyValuePairs(targetNodeSelector, "target node selector")
				if err != nil {
					return err
				}
				planSpec.TargetNodeSelector = nodeSelector
			}

			// Handle target affinity (parse KARL rule)
			if targetAffinity != "" {
				interpreter := karl.NewKARLInterpreter()
				err := interpreter.Parse(targetAffinity)
				if err != nil {
					return fmt.Errorf("failed to parse target affinity KARL rule: %v", err)
				}

				affinity, err := interpreter.ToAffinity()
				if err != nil {
					return fmt.Errorf("failed to convert KARL rule to affinity: %v", err)
				}
				planSpec.TargetAffinity = affinity
			}

			// Handle target power state
			if targetPowerState != "" {
				planSpec.TargetPowerState = planv1beta1.TargetPowerState(targetPowerState)
			}

			// Handle use compatibility mode
			planSpec.UseCompatibilityMode = useCompatibilityMode

			// Set VMs in the PlanSpec
			planSpec.VMs = vmList

			opts := plan.CreatePlanOptions{
				Name:                      name,
				Namespace:                 namespace,
				SourceProvider:            sourceProvider,
				TargetProvider:            targetProvider,
				NetworkMapping:            networkMapping,
				StorageMapping:            storageMapping,
				ConfigFlags:               kubeConfigFlags,
				InventoryURL:              inventoryURL,
				DefaultTargetNetwork:      defaultTargetNetwork,
				DefaultTargetStorageClass: defaultTargetStorageClass,
				PlanSpec:                  planSpec,
				NetworkPairs:              networkPairs,
				StoragePairs:              storagePairs,
				DefaultVolumeMode:         defaultVolumeMode,
				DefaultAccessMode:         defaultAccessMode,
				DefaultOffloadPlugin:      defaultOffloadPlugin,
				DefaultOffloadSecret:      defaultOffloadSecret,
				DefaultOffloadVendor:      defaultOffloadVendor,
			}

			err := plan.Create(opts)
			return err
		},
	}

	cmd.Flags().StringVarP(&sourceProvider, "source", "S", "", "Source provider name (supports namespace/name pattern, defaults to plan namespace)")
	cmd.Flags().StringVarP(&targetProvider, "target", "t", "", "Target provider name (supports namespace/name pattern, defaults to plan namespace)")
	cmd.Flags().StringVar(&networkMapping, "network-mapping", "", "Network mapping name")
	cmd.Flags().StringVar(&storageMapping, "storage-mapping", "", "Storage mapping name")
	cmd.Flags().StringVar(&networkPairs, "network-pairs", "", "Network mapping pairs in format 'source:target-namespace/target-network', 'source:target-network', 'source:default', or 'source:ignored' (comma-separated)")
	cmd.Flags().StringVar(&storagePairs, "storage-pairs", "", "Storage mapping pairs in format 'source:storage-class[;volumeMode=Block|Filesystem][;accessMode=ReadWriteOnce|ReadWriteMany|ReadOnlyMany][;offloadPlugin=vsphere][;offloadSecret=secret-name][;offloadVendor=vantara|ontap|...]' (comma-separated pairs, semicolon-separated parameters)")

	// Storage enhancement flags
	cmd.Flags().StringVar(&defaultVolumeMode, "default-volume-mode", "", "Default volume mode for storage pairs (Filesystem|Block)")
	cmd.Flags().StringVar(&defaultAccessMode, "default-access-mode", "", "Default access mode for storage pairs (ReadWriteOnce|ReadWriteMany|ReadOnlyMany)")
	cmd.Flags().StringVar(&defaultOffloadPlugin, "default-offload-plugin", "", "Default offload plugin type for storage pairs (vsphere)")
	cmd.Flags().StringVar(&defaultOffloadSecret, "default-offload-secret", "", "Default offload plugin secret name for storage pairs")
	cmd.Flags().StringVar(&defaultOffloadVendor, "default-offload-vendor", "", "Default offload plugin vendor for storage pairs (vantara|ontap|primera3par|pureFlashArray|powerflex|powermax)")

	cmd.Flags().StringVar(&vmNamesOrFile, "vms", "", "List of VM names (comma-separated) or path to YAML/JSON file containing a list of VM structs")
	cmd.Flags().StringVar(&preHook, "pre-hook", "", "Pre-migration hook to add to all VMs in the plan")
	cmd.Flags().StringVar(&postHook, "post-hook", "", "Post-migration hook to add to all VMs in the plan")

	// PlanSpec flags
	cmd.Flags().StringVar(&planSpec.Description, "description", "", "Plan description")
	cmd.Flags().StringVar(&planSpec.TargetNamespace, "target-namespace", "", "Target namespace")
	cmd.Flags().StringVar(&transferNetwork, "transfer-network", "", "The network attachment definition that should be used for disk transfer")
	cmd.Flags().BoolVar(&planSpec.PreserveClusterCPUModel, "preserve-cluster-cpu-model", false, "Preserve the CPU model and flags the VM runs with in its oVirt cluster")
	cmd.Flags().BoolVar(&planSpec.PreserveStaticIPs, "preserve-static-ips", false, "Preserve static IPs of VMs in vSphere")
	cmd.Flags().StringVar(&planSpec.PVCNameTemplate, "pvc-name-template", "", "PVCNameTemplate is a template for generating PVC names for VM disks. Variables: {{.VmName}}, {{.PlanName}}, {{.DiskIndex}}, {{.WinDriveLetter}}, {{.RootDiskIndex}}, {{.Shared}}, {{.FileName}}")
	cmd.Flags().StringVar(&planSpec.VolumeNameTemplate, "volume-name-template", "", "VolumeNameTemplate is a template for generating volume interface names in the target virtual machine. Variables: {{.PVCName}}, {{.VolumeIndex}}")
	cmd.Flags().StringVar(&planSpec.NetworkNameTemplate, "network-name-template", "", "NetworkNameTemplate is a template for generating network interface names in the target virtual machine. Variables: {{.NetworkName}}, {{.NetworkNamespace}}, {{.NetworkType}}, {{.NetworkIndex}}")
	cmd.Flags().BoolVar(&planSpec.MigrateSharedDisks, "migrate-shared-disks", true, "Determines if the plan should migrate shared disks")
	cmd.Flags().StringVarP(&inventoryURL, "inventory-url", "i", os.Getenv("MTV_INVENTORY_URL"), "Base URL for the inventory service")
	cmd.Flags().BoolVar(&planSpec.Archived, "archived", false, "Whether this plan should be archived")
	cmd.Flags().BoolVar(&planSpec.PVCNameTemplateUseGenerateName, "pvc-name-template-use-generate-name", true, "Use generateName instead of name for PVC name template")
	cmd.Flags().BoolVar(&planSpec.DeleteGuestConversionPod, "delete-guest-conversion-pod", false, "Delete guest conversion pod after successful migration")
	cmd.Flags().BoolVar(&planSpec.DeleteVmOnFailMigration, "delete-vm-on-fail-migration", false, "Delete target VM when migration fails")
	cmd.Flags().BoolVar(&planSpec.SkipGuestConversion, "skip-guest-conversion", false, "Skip the guest conversion process")
	cmd.Flags().StringVar(&installLegacyDrivers, "install-legacy-drivers", "", "Install legacy Windows drivers (true/false, leave empty for auto-detection)")
	cmd.Flags().VarP(migrationTypeFlag, "migration-type", "m", "Migration type: cold, warm, live, or conversion (supersedes --warm flag)")
	cmd.Flags().StringVarP(&defaultTargetNetwork, "default-target-network", "N", "", "Default target network for network mapping. Use 'default' for pod networking, or specify a NetworkAttachmentDefinition name")
	cmd.Flags().StringVar(&defaultTargetStorageClass, "default-target-storage-class", "", "Default target storage class for storage mapping")
	cmd.Flags().BoolVar(&useCompatibilityMode, "use-compatibility-mode", true, "Use compatibility devices (SATA bus, E1000E NIC) when skipGuestConversion is true to ensure bootability")
	cmd.Flags().StringSliceVarP(&targetLabels, "target-labels", "L", nil, "Target labels to be added to the VM (e.g., key1=value1,key2=value2)")
	cmd.Flags().StringSliceVar(&targetNodeSelector, "target-node-selector", nil, "Target node selector to constrain VM scheduling (e.g., key1=value1,key2=value2)")
	cmd.Flags().BoolVar(&planSpec.Warm, "warm", false, "Enable warm migration (can also be set with --migration-type=warm)")
	cmd.Flags().StringVar(&targetAffinity, "target-affinity", "", "Target affinity to constrain VM scheduling using KARL syntax (e.g. 'REQUIRE pods(app=database) on node')")
	cmd.Flags().StringVar(&targetPowerState, "target-power-state", "", "Target power state for VMs after migration: 'on', 'off', or 'auto' (default: match source VM power state)")

	// Add completion for storage enhancement flags
	if err := cmd.RegisterFlagCompletionFunc("default-volume-mode", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"Filesystem", "Block"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	if err := cmd.RegisterFlagCompletionFunc("default-access-mode", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"ReadWriteOnce", "ReadWriteMany", "ReadOnlyMany"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	if err := cmd.RegisterFlagCompletionFunc("default-offload-plugin", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"vsphere"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	if err := cmd.RegisterFlagCompletionFunc("default-offload-vendor", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"vantara", "ontap", "primera3par", "pureFlashArray", "powerflex", "powermax"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	// Add completion for migration type flag
	if err := cmd.RegisterFlagCompletionFunc("migration-type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return migrationTypeFlag.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	// Add completion for install legacy drivers flag
	if err := cmd.RegisterFlagCompletionFunc("install-legacy-drivers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	// Add completion for target power state flag
	if err := cmd.RegisterFlagCompletionFunc("target-power-state", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"on", "off", "auto"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	// Add completion for pre-hook flag
	if err := cmd.RegisterFlagCompletionFunc("pre-hook", completion.HookResourceNameCompletion(kubeConfigFlags)); err != nil {
		panic(err)
	}

	// Add completion for post-hook flag
	if err := cmd.RegisterFlagCompletionFunc("post-hook", completion.HookResourceNameCompletion(kubeConfigFlags)); err != nil {
		panic(err)
	}

	return cmd
}
