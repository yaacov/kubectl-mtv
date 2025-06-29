package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	planv1beta1 "github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1/plan"
	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/plan"
	"gopkg.in/yaml.v3"
	cnv "kubevirt.io/api/core/v1"
)

func newPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Manage migration plans",
		Long:  `Create and manage VM migration plans`,
	}

	cmd.AddCommand(newCreatePlanCmd())
	cmd.AddCommand(newListPlanCmd())
	cmd.AddCommand(newStartPlanCmd())
	cmd.AddCommand(newDescribePlanCmd())
	cmd.AddCommand(newDescribeVMCmd())
	cmd.AddCommand(newCancelVMsCmd())
	cmd.AddCommand(newCutoverCmd())
	cmd.AddCommand(newDeletePlanCmd())
	cmd.AddCommand(newVMsCmd())
	cmd.AddCommand(newArchivePlanCmd())
	return cmd
}

func newCreatePlanCmd() *cobra.Command {
	var name, sourceProvider, targetProvider string
	var networkMapping, storageMapping string
	var vmNamesOrFile string
	var description, targetNamespace string
	var warm, preserveClusterCPUModel, preserveStaticIPs, migrateSharedDisks bool
	var transferNetwork, pvcNameTemplate, volumeNameTemplate, networkNameTemplate string
	var inventoryURL string
	var archived, pvcNameTemplateUseGenerateName, deleteGuestConversionPod bool
	var diskBusStr string
	var defaultTargetNetwork, defaultTargetStorageClass string

	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a migration plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name = args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// If inventoryURL is empty, try to discover it
			if inventoryURL == "" {
				inventoryURL = discoverInventoryURL(kubeConfigFlags, namespace)
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

			// Parse disk bus if provided
			var diskBus cnv.DiskBus
			if diskBusStr != "" {
				diskBus = cnv.DiskBus(diskBusStr)
			}

			opts := plan.CreatePlanOptions{
				Name:                           name,
				Namespace:                      namespace,
				SourceProvider:                 sourceProvider,
				TargetProvider:                 targetProvider,
				NetworkMapping:                 networkMapping,
				StorageMapping:                 storageMapping,
				VMList:                         vmList,
				Description:                    description,
				TargetNamespace:                targetNamespace,
				Warm:                           warm,
				TransferNetwork:                transferNetwork,
				PreserveClusterCPUModel:        preserveClusterCPUModel,
				PreserveStaticIPs:              preserveStaticIPs,
				PVCNameTemplate:                pvcNameTemplate,
				VolumeNameTemplate:             volumeNameTemplate,
				NetworkNameTemplate:            networkNameTemplate,
				MigrateSharedDisks:             migrateSharedDisks,
				ConfigFlags:                    kubeConfigFlags,
				InventoryURL:                   inventoryURL,
				Archived:                       archived,
				DiskBus:                        diskBus,
				PVCNameTemplateUseGenerateName: pvcNameTemplateUseGenerateName,
				DeleteGuestConversionPod:       deleteGuestConversionPod,
				DefaultTargetNetwork:           defaultTargetNetwork,
				DefaultTargetStorageClass:      defaultTargetStorageClass,
			}

			return plan.Create(opts)
		},
	}

	cmd.Flags().StringVarP(&sourceProvider, "source", "S", "", "Source provider name")
	cmd.Flags().StringVarP(&targetProvider, "target", "t", "", "Target provider name")
	cmd.Flags().StringVar(&networkMapping, "network-mapping", "", "Network mapping name")
	cmd.Flags().StringVar(&storageMapping, "storage-mapping", "", "Storage mapping name")
	cmd.Flags().StringVar(&vmNamesOrFile, "vms", "", "List of VM names (comma-separated) or path to YAML/JSON file containing a list of VM structs")

	cmd.Flags().StringVar(&description, "description", "", "Plan description")
	cmd.Flags().StringVar(&targetNamespace, "target-namespace", "", "Target namespace")
	cmd.Flags().BoolVar(&warm, "warm", false, "Whether this is a warm migration")
	cmd.Flags().StringVar(&transferNetwork, "transfer-network", "", "The network attachment definition that should be used for disk transfer")
	cmd.Flags().BoolVar(&preserveClusterCPUModel, "preserve-cluster-cpu-model", false, "Preserve the CPU model and flags the VM runs with in its oVirt cluster")
	cmd.Flags().BoolVar(&preserveStaticIPs, "preserve-static-ips", false, "Preserve static IPs of VMs in vSphere")
	cmd.Flags().StringVar(&pvcNameTemplate, "pvc-name-template", "", "PVCNameTemplate is a template for generating PVC names for VM disks")
	cmd.Flags().StringVar(&volumeNameTemplate, "volume-name-template", "", "VolumeNameTemplate is a template for generating volume interface names in the target virtual machine")
	cmd.Flags().StringVar(&networkNameTemplate, "network-name-template", "", "NetworkNameTemplate is a template for generating network interface names in the target virtual machine")
	cmd.Flags().BoolVar(&migrateSharedDisks, "migrate-shared-disks", true, "Determines if the plan should migrate shared disks")
	cmd.Flags().StringVarP(&inventoryURL, "inventory-url", "i", os.Getenv("MTV_INVENTORY_URL"), "Base URL for the inventory service")
	cmd.Flags().BoolVar(&archived, "archived", false, "Whether this plan should be archived")
	cmd.Flags().StringVar(&diskBusStr, "disk-bus", "", "Disk bus type (deprecated: will be deprecated in 2.8)")
	cmd.Flags().BoolVar(&pvcNameTemplateUseGenerateName, "pvc-name-template-use-generate-name", true, "Use generateName instead of name for PVC name template")
	cmd.Flags().BoolVar(&deleteGuestConversionPod, "delete-guest-conversion-pod", false, "Delete guest conversion pod after successful migration")
	cmd.Flags().StringVarP(&defaultTargetNetwork, "default-target-network", "N", "", "Default target network for network mapping. Use 'pod' for pod networking, or specify a NetworkAttachmentDefinition name")
	cmd.Flags().StringVar(&defaultTargetStorageClass, "default-target-storage-class", "", "Default target storage class for storage mapping")

	return cmd
}

func newListPlanCmd() *cobra.Command {
	var watch bool
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List migration plans",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			return plan.List(kubeConfigFlags, namespace, watch, outputFormat)
		},
	}

	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch migration plans with live updates using tview")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format. One of: table, json, yaml")

	return cmd
}

func newStartPlanCmd() *cobra.Command {
	var cutoverTimeStr string

	cmd := &cobra.Command{
		Use:   "start NAME",
		Short: "Start a migration plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

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

			return plan.Start(kubeConfigFlags, name, namespace, cutoverTime)
		},
	}

	cmd.Flags().StringVarP(&cutoverTimeStr, "cutover", "c", "", "Cutover time in ISO8601 format (e.g., 2023-12-31T15:30:00Z, '$(date -d \"+1 hour\" --iso-8601=sec)' ). If not provided, defaults to 1 hour from now.")

	return cmd
}

func newDescribePlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe a migration plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			return plan.Describe(kubeConfigFlags, name, namespace)
		},
	}

	return cmd
}

func newDescribeVMCmd() *cobra.Command {
	var watch bool
	var vmName string

	cmd := &cobra.Command{
		Use:   "vm NAME",
		Short: "Describe VM status in a migration plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get plan name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			return plan.DescribeVM(kubeConfigFlags, name, namespace, vmName, watch)
		},
	}

	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch VM status with live updates")
	cmd.Flags().StringVar(&vmName, "vm", "", "VM name to describe")

	err := cmd.MarkFlagRequired("vm")
	if err != nil {
		fmt.Printf("Warning: error marking 'vm' flag as required: %v\n", err)
	}

	return cmd
}

func newCancelVMsCmd() *cobra.Command {
	var vmNamesOrFile string

	cmd := &cobra.Command{
		Use:   "cancel-vms NAME",
		Short: "Cancel specific VMs in a running migration",
		Args:  cobra.ExactArgs(1),
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

func newCutoverCmd() *cobra.Command {
	var cutoverTimeStr string

	cmd := &cobra.Command{
		Use:   "cutover NAME",
		Short: "Set the cutover time for a warm migration",
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

			return plan.Cutover(kubeConfigFlags, planName, namespace, cutoverTime)
		},
	}

	cmd.Flags().StringVarP(&cutoverTimeStr, "cutover", "c", "", "Cutover time in ISO8601 format (e.g., 2023-12-31T15:30:00Z, '$(date --iso-8601=sec)'). If not specified, defaults to current time.")

	return cmd
}

func newDeletePlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a migration plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			return plan.Delete(kubeConfigFlags, name, namespace)
		},
	}

	return cmd
}

func newVMsCmd() *cobra.Command {
	var watch bool

	cmd := &cobra.Command{
		Use:   "vms NAME",
		Short: "List VMs in a migration plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get plan name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			return plan.ListVMs(kubeConfigFlags, name, namespace, watch)
		},
	}

	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch VM status with live updates")

	return cmd
}

func newArchivePlanCmd() *cobra.Command {
	var unarchive bool

	cmd := &cobra.Command{
		Use:   "archive NAME",
		Short: "Archive or unarchive a migration plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// If unarchive flag is specified, set archived to false
			archived := !unarchive

			return plan.Archive(kubeConfigFlags, name, namespace, archived)
		},
	}

	cmd.Flags().BoolVar(&unarchive, "unarchive", false, "Unarchive the plan instead of archiving it")

	return cmd
}
