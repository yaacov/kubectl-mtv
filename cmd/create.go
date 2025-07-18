package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	planv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/plan"
	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/flags"
	"github.com/yaacov/kubectl-mtv/pkg/mapping"
	"github.com/yaacov/kubectl-mtv/pkg/plan"
	"github.com/yaacov/kubectl-mtv/pkg/provider"
	"github.com/yaacov/kubectl-mtv/pkg/vddk"
	"gopkg.in/yaml.v3"
	cnv "kubevirt.io/api/core/v1"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create resources",
		Long:  `Create various MTV resources like providers, plans, mappings, and VDDK images`,
	}

	cmd.AddCommand(newCreateProviderCmd())
	cmd.AddCommand(newCreatePlanCmd())
	cmd.AddCommand(newCreateMappingCmd())
	cmd.AddCommand(newCreateVddkCmd())

	return cmd
}

func newCreateProviderCmd() *cobra.Command {
	var secret string
	providerType := flags.NewProviderTypeFlag()

	// Add Provider credential flags
	var url, username, password, cacert, token string
	var insecureSkipTLS bool
	var vddkInitImage string

	// Check if MTV_VDDK_INIT_IMAGE environment variable is set
	if envVddkInitImage := os.Getenv("MTV_VDDK_INIT_IMAGE"); envVddkInitImage != "" {
		vddkInitImage = envVddkInitImage
	}

	cmd := &cobra.Command{
		Use:   "provider NAME",
		Short: "Create a new provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Check if cacert starts with @ and load from file if so
			if strings.HasPrefix(cacert, "@") {
				filePath := cacert[1:]
				fileContent, err := os.ReadFile(filePath)
				if err != nil {
					return err
				}
				cacert = string(fileContent)
			}

			err := provider.Create(kubeConfigFlags, providerType.GetValue(), name, namespace, secret,
				url, username, password, cacert, insecureSkipTLS, vddkInitImage, token)
			if err != nil {
				printCommandError(err, "creating provider", namespace)
			}
			return nil
		},
	}

	cmd.Flags().Var(providerType, "type", "Provider type (openshift, vsphere, ovirt, openstack, ova)")
	cmd.Flags().StringVar(&secret, "secret", "", "Secret containing provider credentials")

	// Add completion for provider type flag
	if err := cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return providerType.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	// Provider credential flags
	cmd.Flags().StringVarP(&url, "url", "U", "", "Provider URL")
	cmd.Flags().StringVarP(&username, "username", "u", "", "Provider credentials username")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Provider credentials password")
	cmd.Flags().StringVarP(&token, "token", "T", "", "Provider authentication token (used for openshift provider)")
	cmd.Flags().StringVar(&cacert, "cacert", "", "Provider CA certificate (use @filename to load from file)")
	cmd.Flags().BoolVar(&insecureSkipTLS, "provider-insecure-skip-tls", false, "Skip TLS verification when connecting to the provider")
	cmd.Flags().StringVar(&vddkInitImage, "vddk-init-image", vddkInitImage, "Virtual Disk Development Kit (VDDK) container init image path")

	if err := cmd.MarkFlagRequired("type"); err != nil {
		panic(err)
	}

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
		Use:   "plan NAME",
		Short: "Create a migration plan",
		Args:  cobra.ExactArgs(1),
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

			err := plan.Create(opts)
			if err != nil {
				printCommandError(err, "creating plan", namespace)
			}
			return nil
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

func newCreateMappingCmd() *cobra.Command {
	var mappingType string
	var sourceProvider, targetProvider string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "mapping NAME",
		Short: "Create a new mapping",
		Long:  `Create a new network or storage mapping`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			var err error
			switch mappingType {
			case "network":
				err = mapping.CreateNetwork(kubeConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile)
			case "storage":
				err = mapping.CreateStorage(kubeConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile)
			default:
				return fmt.Errorf("unsupported mapping type: %s. Use 'network' or 'storage'", mappingType)
			}

			if err != nil {
				printCommandError(err, "creating mapping", namespace)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&mappingType, "type", "t", "", "Mapping type (network, storage)")
	cmd.Flags().StringVarP(&sourceProvider, "source", "S", "", "Source provider name")
	cmd.Flags().StringVarP(&targetProvider, "target", "T", "", "Target provider name")
	cmd.Flags().StringVarP(&fromFile, "from-file", "f", "", "Create mapping from YAML/JSON file")

	// Add completion for mapping type flag
	if err := cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"network", "storage"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	if err := cmd.MarkFlagRequired("type"); err != nil {
		panic(err)
	}

	return cmd
}

func newCreateVddkCmd() *cobra.Command {
	var vddkTarGz, vddkTag, vddkBuildDir string
	var vddkPush bool

	cmd := &cobra.Command{
		Use:   "vddk-image",
		Short: "Create a VDDK image for MTV",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := vddk.BuildImage(vddkTarGz, vddkTag, vddkBuildDir, vddkPush)
			if err != nil {
				fmt.Printf("Error building VDDK image: %v\n", err)
				fmt.Printf("Please ensure you have the correct permissions and Docker/Podman is available.\n")
				fmt.Printf("You can use the '--help' flag for more information on usage.\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&vddkTarGz, "tar", "", "Path to VMware VDDK tar.gz file (required), e.g. VMware-vix-disklib.tar.gz")
	cmd.Flags().StringVar(&vddkTag, "tag", "", "Container image tag (required), e.g. quay.io/example/vddk:8.0.1")
	cmd.Flags().StringVar(&vddkBuildDir, "build-dir", "", "Build directory (optional, uses tmp dir if not set)")
	cmd.Flags().BoolVar(&vddkPush, "push", false, "Push image after build (optional)")

	if err := cmd.MarkFlagRequired("tar"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("tag"); err != nil {
		panic(err)
	}

	return cmd
}
