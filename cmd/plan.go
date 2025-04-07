package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	planv1beta1 "github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1/plan"
	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/plan"
	"gopkg.in/yaml.v3"
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
	cmd.AddCommand(newCancelVMsCmd())

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

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a migration plan",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			opts := plan.CreatePlanOptions{
				Name:                    name,
				Namespace:               namespace,
				SourceProvider:          sourceProvider,
				TargetProvider:          targetProvider,
				NetworkMapping:          networkMapping,
				StorageMapping:          storageMapping,
				VMList:                  vmList,
				Description:             description,
				TargetNamespace:         targetNamespace,
				Warm:                    warm,
				TransferNetwork:         transferNetwork,
				PreserveClusterCPUModel: preserveClusterCPUModel,
				PreserveStaticIPs:       preserveStaticIPs,
				PVCNameTemplate:         pvcNameTemplate,
				VolumeNameTemplate:      volumeNameTemplate,
				NetworkNameTemplate:     networkNameTemplate,
				MigrateSharedDisks:      migrateSharedDisks,
				ConfigFlags:             kubeConfigFlags,
				InventoryURL:            inventoryURL,
			}

			return plan.Create(opts)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Plan name")
	cmd.Flags().StringVar(&sourceProvider, "source", "", "Source provider name")
	cmd.Flags().StringVar(&targetProvider, "target", "", "Target provider name")
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
	cmd.Flags().StringVar(&inventoryURL, "inventory-url", "", "Base URL for the inventory service")

	if err := cmd.MarkFlagRequired("name"); err != nil {
		fmt.Printf("Warning: error marking 'provider' flag as required: %v\n", err)
	}

	return cmd
}

func newListPlanCmd() *cobra.Command {
	var watch bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List migration plans",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			return plan.List(kubeConfigFlags, namespace, watch)
		},
	}

	cmd.Flags().BoolVar(&watch, "watch", false, "Watch migration plans with live updates using tview")

	return cmd
}

func newStartPlanCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a migration plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			return plan.Start(kubeConfigFlags, name, namespace)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Plan name")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		fmt.Printf("Warning: error marking 'provider' flag as required: %v\n", err)
	}

	return cmd
}

func newDescribePlanCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a migration plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)
			return plan.Describe(kubeConfigFlags, name, namespace)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Plan name")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		fmt.Printf("Warning: error marking 'provider' flag as required: %v\n", err)
	}

	return cmd
}

func newCancelVMsCmd() *cobra.Command {
	var planName string
	var vmNamesOrFile string

	cmd := &cobra.Command{
		Use:   "cancel-vms",
		Short: "Cancel specific VMs in a running migration",
		RunE: func(cmd *cobra.Command, args []string) error {
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

	cmd.Flags().StringVar(&planName, "name", "", "Plan name")
	cmd.Flags().StringVar(&vmNamesOrFile, "vms", "", "List of VM names to cancel (comma-separated) or path to file containing VM names (prefix with @)")

	if err := cmd.MarkFlagRequired("name"); err != nil {
		fmt.Printf("Warning: error marking 'name' flag as required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("vms"); err != nil {
		fmt.Printf("Warning: error marking 'vms' flag as required: %v\n", err)
	}

	return cmd
}
