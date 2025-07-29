package patch

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/patch/plan"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	"github.com/yaacov/kubectl-mtv/pkg/util/completion"
	"github.com/yaacov/kubectl-mtv/pkg/util/flags"
)

// NewPlanCmd creates the patch plan command
func NewPlanCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	// Editable PlanSpec fields
	var transferNetwork string
	var installLegacyDrivers string // "true", "false", or "" for nil
	migrationTypeFlag := flags.NewMigrationTypeFlag()
	var targetLabels []string
	var targetNodeSelector []string
	var useCompatibilityMode bool
	var targetAffinity string
	var targetNamespace string

	// Missing flags from create plan
	var description string
	var preserveClusterCPUModel bool
	var preserveStaticIPs bool
	var pvcNameTemplate string
	var volumeNameTemplate string
	var networkNameTemplate string
	var migrateSharedDisks bool
	var archived bool
	var pvcNameTemplateUseGenerateName bool
	var deleteGuestConversionPod bool
	var skipGuestConversion bool
	var warm bool

	// Boolean tracking for flag changes
	var useCompatibilityModeChanged bool
	var preserveClusterCPUModelChanged bool
	var preserveStaticIPsChanged bool
	var migrateSharedDisksChanged bool
	var archivedChanged bool
	var pvcNameTemplateUseGenerateNameChanged bool
	var deleteGuestConversionPodChanged bool
	var skipGuestConversionChanged bool
	var warmChanged bool

	cmd := &cobra.Command{
		Use:               "plan NAME",
		Short:             "Patch an existing migration plan",
		Long:              `Patch an existing migration plan by updating editable fields. Provider references, mappings, and VMs cannot be changed.`,
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: completion.PlanNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Track flag changes
			useCompatibilityModeChanged = cmd.Flag("use-compatibility-mode").Changed
			preserveClusterCPUModelChanged = cmd.Flag("preserve-cluster-cpu-model").Changed
			preserveStaticIPsChanged = cmd.Flag("preserve-static-ips").Changed
			migrateSharedDisksChanged = cmd.Flag("migrate-shared-disks").Changed
			archivedChanged = cmd.Flag("archived").Changed
			pvcNameTemplateUseGenerateNameChanged = cmd.Flag("pvc-name-template-use-generate-name").Changed
			deleteGuestConversionPodChanged = cmd.Flag("delete-guest-conversion-pod").Changed
			skipGuestConversionChanged = cmd.Flag("skip-guest-conversion").Changed
			warmChanged = cmd.Flag("warm").Changed

			return plan.PatchPlan(plan.PatchPlanOptions{
				ConfigFlags: kubeConfigFlags,
				Name:        name,
				Namespace:   namespace,

				// Core plan fields
				TransferNetwork:      transferNetwork,
				InstallLegacyDrivers: installLegacyDrivers,
				MigrationType:        migrationTypeFlag.GetValue(),
				TargetLabels:         targetLabels,
				TargetNodeSelector:   targetNodeSelector,
				UseCompatibilityMode: useCompatibilityMode,
				TargetAffinity:       targetAffinity,
				TargetNamespace:      targetNamespace,

				// Additional plan fields
				Description:                    description,
				PreserveClusterCPUModel:        preserveClusterCPUModel,
				PreserveStaticIPs:              preserveStaticIPs,
				PVCNameTemplate:                pvcNameTemplate,
				VolumeNameTemplate:             volumeNameTemplate,
				NetworkNameTemplate:            networkNameTemplate,
				MigrateSharedDisks:             migrateSharedDisks,
				Archived:                       archived,
				PVCNameTemplateUseGenerateName: pvcNameTemplateUseGenerateName,
				DeleteGuestConversionPod:       deleteGuestConversionPod,
				SkipGuestConversion:            skipGuestConversion,
				Warm:                           warm,

				// Flag change tracking
				UseCompatibilityModeChanged:           useCompatibilityModeChanged,
				PreserveClusterCPUModelChanged:        preserveClusterCPUModelChanged,
				PreserveStaticIPsChanged:              preserveStaticIPsChanged,
				MigrateSharedDisksChanged:             migrateSharedDisksChanged,
				ArchivedChanged:                       archivedChanged,
				PVCNameTemplateUseGenerateNameChanged: pvcNameTemplateUseGenerateNameChanged,
				DeleteGuestConversionPodChanged:       deleteGuestConversionPodChanged,
				SkipGuestConversionChanged:            skipGuestConversionChanged,
				WarmChanged:                           warmChanged,
			})
		},
	}

	// Core editable plan flags
	cmd.Flags().StringVar(&transferNetwork, "transfer-network", "", "Network to use for transferring VM data")
	cmd.Flags().StringVar(&installLegacyDrivers, "install-legacy-drivers", "", "Install legacy drivers (true/false)")
	cmd.Flags().Var(migrationTypeFlag, "migration-type", "Migration type (cold, warm)")
	cmd.Flags().StringSliceVar(&targetLabels, "target-labels", []string{}, "Target VM labels in format key=value (can be specified multiple times)")
	cmd.Flags().StringSliceVar(&targetNodeSelector, "target-node-selector", []string{}, "Target node selector in format key=value (can be specified multiple times)")
	cmd.Flags().BoolVar(&useCompatibilityMode, "use-compatibility-mode", false, "Use compatibility mode for migration")
	cmd.Flags().StringVar(&targetAffinity, "target-affinity", "", "Target affinity using KARL syntax (e.g. 'REQUIRE pods(app=database) on node')")
	cmd.Flags().StringVar(&targetNamespace, "target-namespace", "", "Target namespace for migrated VMs")

	// Plan metadata and configuration flags
	cmd.Flags().StringVar(&description, "description", "", "Plan description")
	cmd.Flags().BoolVar(&preserveClusterCPUModel, "preserve-cluster-cpu-model", false, "Preserve the CPU model and flags the VM runs with in its oVirt cluster")
	cmd.Flags().BoolVar(&preserveStaticIPs, "preserve-static-ips", false, "Preserve static IPs of VMs in vSphere")
	cmd.Flags().StringVar(&pvcNameTemplate, "pvc-name-template", "", "Template for generating PVC names for VM disks")
	cmd.Flags().StringVar(&volumeNameTemplate, "volume-name-template", "", "Template for generating volume interface names in the target VM")
	cmd.Flags().StringVar(&networkNameTemplate, "network-name-template", "", "Template for generating network interface names in the target VM")
	cmd.Flags().BoolVar(&migrateSharedDisks, "migrate-shared-disks", true, "Determines if the plan should migrate shared disks")
	cmd.Flags().BoolVar(&archived, "archived", false, "Whether this plan should be archived")
	cmd.Flags().BoolVar(&pvcNameTemplateUseGenerateName, "pvc-name-template-use-generate-name", true, "Use generateName instead of name for PVC name template")
	cmd.Flags().BoolVar(&deleteGuestConversionPod, "delete-guest-conversion-pod", false, "Delete guest conversion pod after successful migration")
	cmd.Flags().BoolVar(&skipGuestConversion, "skip-guest-conversion", false, "Skip the guest conversion process")
	cmd.Flags().BoolVar(&warm, "warm", false, "Enable warm migration (use --migration-type=warm instead)")

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

// NewPlanVmsCmd creates the patch plan-vms command
func NewPlanVmsCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	// VM-specific fields that can be patched
	var targetName string
	var rootDisk string
	var instanceType string
	var pvcNameTemplate string
	var volumeNameTemplate string
	var networkNameTemplate string
	var luksSecret string

	// Hook-related flags
	var addPreHook string
	var addPostHook string
	var removeHook string
	var clearHooks bool

	cmd := &cobra.Command{
		Use:               "plan-vms PLAN_NAME VM_NAME",
		Short:             "Patch a specific VM within a migration plan",
		Long:              `Patch VM-specific fields for a VM within a migration plan's VM list.`,
		Args:              cobra.ExactArgs(2),
		SilenceUsage:      true,
		ValidArgsFunction: completion.PlanNameCompletion(kubeConfigFlags),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get arguments
			planName := args[0]
			vmName := args[1]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			return plan.PatchPlanVM(kubeConfigFlags, planName, vmName, namespace,
				targetName, rootDisk, instanceType, pvcNameTemplate, volumeNameTemplate, networkNameTemplate, luksSecret,
				addPreHook, addPostHook, removeHook, clearHooks)
		},
	}

	// VM-specific flags
	cmd.Flags().StringVar(&targetName, "target-name", "", "Custom name for the VM in the target cluster")
	cmd.Flags().StringVar(&rootDisk, "root-disk", "", "The primary disk to boot from")
	cmd.Flags().StringVar(&instanceType, "instance-type", "", "Override the VM's instance type in the target")
	cmd.Flags().StringVar(&pvcNameTemplate, "pvc-name-template", "", "Go template for naming PVCs for this VM's disks")
	cmd.Flags().StringVar(&volumeNameTemplate, "volume-name-template", "", "Go template for naming volume interfaces")
	cmd.Flags().StringVar(&networkNameTemplate, "network-name-template", "", "Go template for naming network interfaces")
	cmd.Flags().StringVar(&luksSecret, "luks-secret", "", "Secret name for disk decryption keys")

	// Hook-related flags
	cmd.Flags().StringVar(&addPreHook, "add-pre-hook", "", "Add a pre-migration hook to this VM")
	cmd.Flags().StringVar(&addPostHook, "add-post-hook", "", "Add a post-migration hook to this VM")
	cmd.Flags().StringVar(&removeHook, "remove-hook", "", "Remove a hook from this VM by hook name")
	cmd.Flags().BoolVar(&clearHooks, "clear-hooks", false, "Remove all hooks from this VM")

	// Add completion for hook flags
	if err := cmd.RegisterFlagCompletionFunc("add-pre-hook", completion.HookResourceNameCompletion(kubeConfigFlags)); err != nil {
		panic(err)
	}

	if err := cmd.RegisterFlagCompletionFunc("add-post-hook", completion.HookResourceNameCompletion(kubeConfigFlags)); err != nil {
		panic(err)
	}

	if err := cmd.RegisterFlagCompletionFunc("remove-hook", completion.HookResourceNameCompletion(kubeConfigFlags)); err != nil {
		panic(err)
	}

	return cmd
}
