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

	// PlanSpec fields
	var planSpec forkliftv1beta1.PlanSpec
	var transferNetwork string
	var installLegacyDrivers string // "true", "false", or "" for nil
	migrationTypeFlag := flags.NewMigrationTypeFlag()
	var targetLabels []string
	var targetNodeSelector []string
	var useCompatibilityMode bool
	var targetAffinity string

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

			// Handle migration type
			if migrationTypeFlag.GetValue() == "" {
				if planSpec.Warm {
					return fmt.Errorf("setting --warm flag is not supported when migration type is specified")
				}

				planSpec.Type = migrationTypeFlag.GetValue()
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
			}

			err := plan.Create(opts)
			return err
		},
	}

	cmd.Flags().StringVarP(&sourceProvider, "source", "S", "", "Source provider name")
	cmd.Flags().StringVarP(&targetProvider, "target", "t", "", "Target provider name")
	cmd.Flags().StringVar(&networkMapping, "network-mapping", "", "Network mapping name")
	cmd.Flags().StringVar(&storageMapping, "storage-mapping", "", "Storage mapping name")
	cmd.Flags().StringVar(&vmNamesOrFile, "vms", "", "List of VM names (comma-separated) or path to YAML/JSON file containing a list of VM structs")

	// PlanSpec flags
	cmd.Flags().StringVar(&planSpec.Description, "description", "", "Plan description")
	cmd.Flags().StringVar(&planSpec.TargetNamespace, "target-namespace", "", "Target namespace")
	cmd.Flags().StringVar(&transferNetwork, "transfer-network", "", "The network attachment definition that should be used for disk transfer")
	cmd.Flags().BoolVar(&planSpec.PreserveClusterCPUModel, "preserve-cluster-cpu-model", false, "Preserve the CPU model and flags the VM runs with in its oVirt cluster")
	cmd.Flags().BoolVar(&planSpec.PreserveStaticIPs, "preserve-static-ips", false, "Preserve static IPs of VMs in vSphere")
	cmd.Flags().StringVar(&planSpec.PVCNameTemplate, "pvc-name-template", "", "PVCNameTemplate is a template for generating PVC names for VM disks")
	cmd.Flags().StringVar(&planSpec.VolumeNameTemplate, "volume-name-template", "", "VolumeNameTemplate is a template for generating volume interface names in the target virtual machine")
	cmd.Flags().StringVar(&planSpec.NetworkNameTemplate, "network-name-template", "", "NetworkNameTemplate is a template for generating network interface names in the target virtual machine")
	cmd.Flags().BoolVar(&planSpec.MigrateSharedDisks, "migrate-shared-disks", true, "Determines if the plan should migrate shared disks")
	cmd.Flags().StringVarP(&inventoryURL, "inventory-url", "i", os.Getenv("MTV_INVENTORY_URL"), "Base URL for the inventory service")
	cmd.Flags().BoolVar(&planSpec.Archived, "archived", false, "Whether this plan should be archived")
	cmd.Flags().BoolVar(&planSpec.PVCNameTemplateUseGenerateName, "pvc-name-template-use-generate-name", true, "Use generateName instead of name for PVC name template")
	cmd.Flags().BoolVar(&planSpec.DeleteGuestConversionPod, "delete-guest-conversion-pod", false, "Delete guest conversion pod after successful migration")
	cmd.Flags().BoolVar(&planSpec.SkipGuestConversion, "skip-guest-conversion", false, "Skip the guest conversion process")
	cmd.Flags().StringVar(&installLegacyDrivers, "install-legacy-drivers", "", "Install legacy Windows drivers (true/false, leave empty for auto-detection)")
	cmd.Flags().VarP(migrationTypeFlag, "migration-type", "m", "Migration type: cold, warm, or live (supersedes --warm flag)")
	cmd.Flags().StringVarP(&defaultTargetNetwork, "default-target-network", "N", "", "Default target network for network mapping. Use 'pod' for pod networking, or specify a NetworkAttachmentDefinition name")
	cmd.Flags().StringVar(&defaultTargetStorageClass, "default-target-storage-class", "", "Default target storage class for storage mapping")
	cmd.Flags().BoolVar(&useCompatibilityMode, "use-compatibility-mode", true, "Use compatibility devices (SATA bus, E1000E NIC) when skipGuestConversion is true to ensure bootability")
	cmd.Flags().StringSliceVarP(&targetLabels, "target-labels", "L", nil, "Target labels to be added to the VM (e.g., key1=value1,key2=value2)")
	cmd.Flags().StringSliceVar(&targetNodeSelector, "target-node-selector", nil, "Target node selector to constrain VM scheduling (e.g., key1=value1,key2=value2)")
	cmd.Flags().BoolVar(&planSpec.Warm, "warm", false, "Enable warm migration (can also be set with --migration-type=warm)")
	cmd.Flags().StringVar(&targetAffinity, "target-affinity", "", "Target affinity to constrain VM scheduling using KARL syntax (e.g. 'REQUIRE pods(app=database) on node')")

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

	return cmd
}
