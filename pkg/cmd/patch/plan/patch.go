package plan

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/yaacov/karl-interpreter/pkg/karl"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
)

// PatchPlanOptions contains all the options for patching a plan
type PatchPlanOptions struct {
	ConfigFlags *genericclioptions.ConfigFlags
	Name        string
	Namespace   string

	// Core plan fields
	TransferNetwork      string
	InstallLegacyDrivers string
	MigrationType        string
	TargetLabels         []string
	TargetNodeSelector   []string
	UseCompatibilityMode bool
	TargetAffinity       string
	TargetNamespace      string
	TargetPowerState     string

	// Additional plan fields
	Description                    string
	PreserveClusterCPUModel        bool
	PreserveStaticIPs              bool
	PVCNameTemplate                string
	VolumeNameTemplate             string
	NetworkNameTemplate            string
	MigrateSharedDisks             bool
	Archived                       bool
	PVCNameTemplateUseGenerateName bool
	DeleteGuestConversionPod       bool
	SkipGuestConversion            bool
	Warm                           bool

	// Flag change tracking
	UseCompatibilityModeChanged           bool
	PreserveClusterCPUModelChanged        bool
	PreserveStaticIPsChanged              bool
	MigrateSharedDisksChanged             bool
	ArchivedChanged                       bool
	PVCNameTemplateUseGenerateNameChanged bool
	DeleteGuestConversionPodChanged       bool
	SkipGuestConversionChanged            bool
	WarmChanged                           bool
}

// PatchPlan patches an existing migration plan
func PatchPlan(opts PatchPlanOptions) error {
	klog.V(2).Infof("Patching plan '%s' in namespace '%s'", opts.Name, opts.Namespace)

	dynamicClient, err := client.GetDynamicClient(opts.ConfigFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Get the existing plan
	existingPlan, err := dynamicClient.Resource(client.PlansGVR).Namespace(opts.Namespace).Get(context.TODO(), opts.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get plan '%s': %v", opts.Name, err)
	}

	// Convert to typed plan
	var plan forkliftv1beta1.Plan
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(existingPlan.Object, &plan)
	if err != nil {
		return fmt.Errorf("failed to convert plan: %v", err)
	}

	// Track if updates were made
	planUpdated := false

	// Update transfer network if provided
	if opts.TransferNetwork != "" {
		klog.V(2).Infof("Updating transfer network to '%s'", opts.TransferNetwork)

		// Parse network name and namespace (supports "namespace/name" or "name" format)
		var networkName, networkNamespace string
		if strings.Contains(opts.TransferNetwork, "/") {
			parts := strings.SplitN(opts.TransferNetwork, "/", 2)
			networkNamespace = strings.TrimSpace(parts[0])
			networkName = strings.TrimSpace(parts[1])
		} else {
			networkName = strings.TrimSpace(opts.TransferNetwork)
			networkNamespace = opts.Namespace // Use plan namespace as default
		}

		plan.Spec.TransferNetwork = &corev1.ObjectReference{
			Kind:       "NetworkAttachmentDefinition",
			APIVersion: "k8s.cni.cncf.io/v1",
			Name:       networkName,
			Namespace:  networkNamespace,
		}
		planUpdated = true
	}

	// Update install legacy drivers if provided
	if opts.InstallLegacyDrivers != "" {
		switch strings.ToLower(opts.InstallLegacyDrivers) {
		case "true":
			legacyDrivers := true
			plan.Spec.InstallLegacyDrivers = &legacyDrivers
			klog.V(2).Infof("Updated install legacy drivers to true")
			planUpdated = true
		case "false":
			legacyDrivers := false
			plan.Spec.InstallLegacyDrivers = &legacyDrivers
			klog.V(2).Infof("Updated install legacy drivers to false")
			planUpdated = true
		default:
			return fmt.Errorf("invalid value for install-legacy-drivers: %s (must be 'true' or 'false')", opts.InstallLegacyDrivers)
		}
	}

	// Update migration type if provided
	if opts.MigrationType != "" {
		plan.Spec.Type = opts.MigrationType
		klog.V(2).Infof("Updated migration type to '%s'", opts.MigrationType)
		planUpdated = true
	}

	// Update target labels if provided
	if len(opts.TargetLabels) > 0 {
		labelMap, err := parseKeyValuePairs(opts.TargetLabels, "target-labels")
		if err != nil {
			return fmt.Errorf("failed to parse target labels: %v", err)
		}
		plan.Spec.TargetLabels = labelMap
		klog.V(2).Infof("Updated target labels: %v", labelMap)
		planUpdated = true
	}

	// Update target node selector if provided
	if len(opts.TargetNodeSelector) > 0 {
		nodeSelectorMap, err := parseKeyValuePairs(opts.TargetNodeSelector, "target-node-selector")
		if err != nil {
			return fmt.Errorf("failed to parse target node selector: %v", err)
		}
		plan.Spec.TargetNodeSelector = nodeSelectorMap
		klog.V(2).Infof("Updated target node selector: %v", nodeSelectorMap)
		planUpdated = true
	}

	// Update use compatibility mode if flag was changed
	if opts.UseCompatibilityModeChanged {
		plan.Spec.UseCompatibilityMode = opts.UseCompatibilityMode
		klog.V(2).Infof("Updated use compatibility mode to %t", opts.UseCompatibilityMode)
		planUpdated = true
	}

	// Update target affinity if provided (using karl-interpreter)
	if opts.TargetAffinity != "" {
		interpreter := karl.NewKARLInterpreter()
		err := interpreter.Parse(opts.TargetAffinity)
		if err != nil {
			return fmt.Errorf("failed to parse target affinity KARL rule: %v", err)
		}

		affinity, err := interpreter.ToAffinity()
		if err != nil {
			return fmt.Errorf("failed to convert KARL rule to affinity: %v", err)
		}
		plan.Spec.TargetAffinity = affinity
		klog.V(2).Infof("Updated target affinity configuration")
		planUpdated = true
	}

	// Update target namespace if provided
	if opts.TargetNamespace != "" {
		plan.Spec.TargetNamespace = opts.TargetNamespace
		klog.V(2).Infof("Updated target namespace to '%s'", opts.TargetNamespace)
		planUpdated = true
	}

	// Update target power state if provided
	if opts.TargetPowerState != "" {
		plan.Spec.TargetPowerState = opts.TargetPowerState
		klog.V(2).Infof("Updated target power state to '%s'", opts.TargetPowerState)
		planUpdated = true
	}

	// Update description if provided
	if opts.Description != "" {
		plan.Spec.Description = opts.Description
		klog.V(2).Infof("Updated description to '%s'", opts.Description)
		planUpdated = true
	}

	// Update preserve cluster CPU model if flag was changed
	if opts.PreserveClusterCPUModelChanged {
		plan.Spec.PreserveClusterCPUModel = opts.PreserveClusterCPUModel
		klog.V(2).Infof("Updated preserve cluster CPU model to %t", opts.PreserveClusterCPUModel)
		planUpdated = true
	}

	// Update preserve static IPs if flag was changed
	if opts.PreserveStaticIPsChanged {
		plan.Spec.PreserveStaticIPs = opts.PreserveStaticIPs
		klog.V(2).Infof("Updated preserve static IPs to %t", opts.PreserveStaticIPs)
		planUpdated = true
	}

	// Update PVC name template if provided
	if opts.PVCNameTemplate != "" {
		plan.Spec.PVCNameTemplate = opts.PVCNameTemplate
		klog.V(2).Infof("Updated PVC name template to '%s'", opts.PVCNameTemplate)
		planUpdated = true
	}

	// Update volume name template if provided
	if opts.VolumeNameTemplate != "" {
		plan.Spec.VolumeNameTemplate = opts.VolumeNameTemplate
		klog.V(2).Infof("Updated volume name template to '%s'", opts.VolumeNameTemplate)
		planUpdated = true
	}

	// Update network name template if provided
	if opts.NetworkNameTemplate != "" {
		plan.Spec.NetworkNameTemplate = opts.NetworkNameTemplate
		klog.V(2).Infof("Updated network name template to '%s'", opts.NetworkNameTemplate)
		planUpdated = true
	}

	// Update migrate shared disks if flag was changed
	if opts.MigrateSharedDisksChanged {
		plan.Spec.MigrateSharedDisks = opts.MigrateSharedDisks
		klog.V(2).Infof("Updated migrate shared disks to %t", opts.MigrateSharedDisks)
		planUpdated = true
	}

	// Update archived if flag was changed
	if opts.ArchivedChanged {
		plan.Spec.Archived = opts.Archived
		klog.V(2).Infof("Updated archived to %t", opts.Archived)
		planUpdated = true
	}

	// Update PVC name template use generate name if flag was changed
	if opts.PVCNameTemplateUseGenerateNameChanged {
		plan.Spec.PVCNameTemplateUseGenerateName = opts.PVCNameTemplateUseGenerateName
		klog.V(2).Infof("Updated PVC name template use generate name to %t", opts.PVCNameTemplateUseGenerateName)
		planUpdated = true
	}

	// Update delete guest conversion pod if flag was changed
	if opts.DeleteGuestConversionPodChanged {
		plan.Spec.DeleteGuestConversionPod = opts.DeleteGuestConversionPod
		klog.V(2).Infof("Updated delete guest conversion pod to %t", opts.DeleteGuestConversionPod)
		planUpdated = true
	}

	// Update skip guest conversion if flag was changed
	if opts.SkipGuestConversionChanged {
		plan.Spec.SkipGuestConversion = opts.SkipGuestConversion
		klog.V(2).Infof("Updated skip guest conversion to %t", opts.SkipGuestConversion)
		planUpdated = true
	}

	// Update warm migration if flag was changed
	if opts.WarmChanged {
		plan.Spec.Warm = opts.Warm
		klog.V(2).Infof("Updated warm migration to %t", opts.Warm)
		planUpdated = true
	}

	// Update the plan if any changes were made
	if planUpdated {
		err = updatePlan(opts.ConfigFlags, &plan)
		if err != nil {
			return fmt.Errorf("failed to update plan: %v", err)
		}
		fmt.Printf("plan/%s patched\n", opts.Name)
	} else {
		fmt.Printf("plan/%s unchanged (no updates specified)\n", opts.Name)
	}

	return nil
}

// PatchPlanVM patches a specific VM within a plan's VM list
func PatchPlanVM(configFlags *genericclioptions.ConfigFlags, planName, vmName, namespace string,
	targetName, rootDisk, instanceType, pvcNameTemplate, volumeNameTemplate, networkNameTemplate, luksSecret, targetPowerState string,
	addPreHook, addPostHook, removeHook string, clearHooks bool) error {

	klog.V(2).Infof("Patching VM '%s' in plan '%s'", vmName, planName)

	dynamicClient, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Get the existing plan
	existingPlan, err := dynamicClient.Resource(client.PlansGVR).Namespace(namespace).Get(context.TODO(), planName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get plan '%s': %v", planName, err)
	}

	// Get the VMs slice from spec.vms
	specVMs, exists, err := unstructured.NestedSlice(existingPlan.Object, "spec", "vms")
	if err != nil {
		return fmt.Errorf("failed to get VMs from plan spec: %v", err)
	}
	if !exists {
		return fmt.Errorf("no VMs found in plan '%s'", planName)
	}

	// Find the VM in the plan's VMs list
	vmIndex := -1
	for i, v := range specVMs {
		vm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		currentVMName, _, _ := unstructured.NestedString(vm, "name")
		if currentVMName == vmName {
			vmIndex = i
			break
		}
	}

	if vmIndex == -1 {
		return fmt.Errorf("VM '%s' not found in plan '%s'", vmName, planName)
	}

	// Get the VM object to modify
	vm, ok := specVMs[vmIndex].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid VM data structure for VM '%s'", vmName)
	}

	// Track if updates were made
	vmUpdated := false

	// Update target name if provided
	if targetName != "" {
		err = unstructured.SetNestedField(vm, targetName, "targetName")
		if err != nil {
			return fmt.Errorf("failed to set target name: %v", err)
		}
		klog.V(2).Infof("Updated VM target name to '%s'", targetName)
		vmUpdated = true
	}

	// Update root disk if provided
	if rootDisk != "" {
		err = unstructured.SetNestedField(vm, rootDisk, "rootDisk")
		if err != nil {
			return fmt.Errorf("failed to set root disk: %v", err)
		}
		klog.V(2).Infof("Updated VM root disk to '%s'", rootDisk)
		vmUpdated = true
	}

	// Update instance type if provided
	if instanceType != "" {
		err = unstructured.SetNestedField(vm, instanceType, "instanceType")
		if err != nil {
			return fmt.Errorf("failed to set instance type: %v", err)
		}
		klog.V(2).Infof("Updated VM instance type to '%s'", instanceType)
		vmUpdated = true
	}

	// Update PVC name template if provided
	if pvcNameTemplate != "" {
		err = unstructured.SetNestedField(vm, pvcNameTemplate, "pvcNameTemplate")
		if err != nil {
			return fmt.Errorf("failed to set PVC name template: %v", err)
		}
		klog.V(2).Infof("Updated VM PVC name template to '%s'", pvcNameTemplate)
		vmUpdated = true
	}

	// Update volume name template if provided
	if volumeNameTemplate != "" {
		err = unstructured.SetNestedField(vm, volumeNameTemplate, "volumeNameTemplate")
		if err != nil {
			return fmt.Errorf("failed to set volume name template: %v", err)
		}
		klog.V(2).Infof("Updated VM volume name template to '%s'", volumeNameTemplate)
		vmUpdated = true
	}

	// Update network name template if provided
	if networkNameTemplate != "" {
		err = unstructured.SetNestedField(vm, networkNameTemplate, "networkNameTemplate")
		if err != nil {
			return fmt.Errorf("failed to set network name template: %v", err)
		}
		klog.V(2).Infof("Updated VM network name template to '%s'", networkNameTemplate)
		vmUpdated = true
	}

	// Update LUKS secret if provided
	if luksSecret != "" {
		luksRef := map[string]interface{}{
			"kind":      "Secret",
			"name":      luksSecret,
			"namespace": namespace,
		}
		err = unstructured.SetNestedMap(vm, luksRef, "luks")
		if err != nil {
			return fmt.Errorf("failed to set LUKS secret: %v", err)
		}
		klog.V(2).Infof("Updated VM LUKS secret to '%s'", luksSecret)
		vmUpdated = true
	}

	// Update target power state if provided
	if targetPowerState != "" {
		err = unstructured.SetNestedField(vm, targetPowerState, "targetPowerState")
		if err != nil {
			return fmt.Errorf("failed to set target power state: %v", err)
		}
		klog.V(2).Infof("Updated VM target power state to '%s'", targetPowerState)
		vmUpdated = true
	}

	// Handle hook operations
	hooksUpdated, err := updateVMHooksUnstructured(vm, namespace, addPreHook, addPostHook, removeHook, clearHooks)
	if err != nil {
		return fmt.Errorf("failed to update VM hooks: %v", err)
	}
	if hooksUpdated {
		vmUpdated = true
	}

	// Update the plan if any changes were made
	if vmUpdated {
		// Set the modified VMs slice back to the plan
		err = unstructured.SetNestedSlice(existingPlan.Object, specVMs, "spec", "vms")
		if err != nil {
			return fmt.Errorf("failed to update VMs in plan: %v", err)
		}

		// Update the plan directly
		_, err = dynamicClient.Resource(client.PlansGVR).Namespace(namespace).Update(context.TODO(), existingPlan, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update plan: %v", err)
		}
		fmt.Printf("plan/%s vm/%s patched\n", planName, vmName)
	} else {
		fmt.Printf("plan/%s vm/%s unchanged (no updates specified)\n", planName, vmName)
	}

	return nil
}

// updatePlan updates the plan resource
func updatePlan(configFlags *genericclioptions.ConfigFlags, plan *forkliftv1beta1.Plan) error {
	dynamicClient, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Convert to unstructured
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(plan)
	if err != nil {
		return fmt.Errorf("failed to convert to unstructured: %v", err)
	}

	unstructuredPlan := &unstructured.Unstructured{Object: unstructuredObj}

	// Update the plan
	_, err = dynamicClient.Resource(client.PlansGVR).Namespace(plan.Namespace).Update(context.TODO(), unstructuredPlan, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update plan: %v", err)
	}

	return nil
}

// parseKeyValuePairs parses key=value pairs from string slice
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
				return nil, fmt.Errorf("invalid %s: %s (expected key=value format)", fieldName, pair)
			}
		}
	}
	return result, nil
}

// updateVMHooksUnstructured handles hook operations for a VM
func updateVMHooksUnstructured(vm map[string]interface{}, namespace, addPreHook, addPostHook, removeHook string, clearHooks bool) (bool, error) {
	updated := false

	// Get existing hooks or create empty slice
	hooks, _, _ := unstructured.NestedSlice(vm, "hooks")
	if hooks == nil {
		hooks = []interface{}{}
	}

	// Clear all hooks if requested
	if clearHooks {
		if len(hooks) > 0 {
			err := unstructured.SetNestedSlice(vm, []interface{}{}, "hooks")
			if err != nil {
				return false, fmt.Errorf("failed to clear hooks: %v", err)
			}
			klog.V(2).Infof("Cleared all hooks from VM")
			updated = true
		}
		return updated, nil
	}

	// Remove specific hook if requested
	if removeHook != "" {
		originalLen := len(hooks)
		var filteredHooks []interface{}
		for _, h := range hooks {
			hook, ok := h.(map[string]interface{})
			if !ok {
				filteredHooks = append(filteredHooks, h)
				continue
			}
			hookName, _, _ := unstructured.NestedString(hook, "hook", "name")
			if hookName != strings.TrimSpace(removeHook) {
				filteredHooks = append(filteredHooks, h)
			}
		}
		if len(filteredHooks) < originalLen {
			err := unstructured.SetNestedSlice(vm, filteredHooks, "hooks")
			if err != nil {
				return false, fmt.Errorf("failed to remove hook: %v", err)
			}
			klog.V(2).Infof("Removed hook '%s' from VM", removeHook)
			updated = true
			hooks = filteredHooks
		}
	}

	// Add pre-hook if requested
	if addPreHook != "" {
		hookName := strings.TrimSpace(addPreHook)

		// Check if this pre-hook already exists
		hookExists := false
		for _, h := range hooks {
			hook, ok := h.(map[string]interface{})
			if !ok {
				continue
			}
			existingHookName, _, _ := unstructured.NestedString(hook, "hook", "name")
			step, _, _ := unstructured.NestedString(hook, "step")
			if existingHookName == hookName && step == "PreHook" {
				hookExists = true
				break
			}
		}

		if !hookExists {
			preHookRef := map[string]interface{}{
				"step": "PreHook",
				"hook": map[string]interface{}{
					"kind":       "Hook",
					"apiVersion": "forklift.konveyor.io/v1beta1",
					"name":       hookName,
					"namespace":  namespace,
				},
			}
			hooks = append(hooks, preHookRef)
			err := unstructured.SetNestedSlice(vm, hooks, "hooks")
			if err != nil {
				return false, fmt.Errorf("failed to add pre-hook: %v", err)
			}
			klog.V(2).Infof("Added pre-hook '%s' to VM", hookName)
			updated = true
		} else {
			klog.V(1).Infof("Pre-hook '%s' already exists for VM, skipping", hookName)
		}
	}

	// Add post-hook if requested
	if addPostHook != "" {
		hookName := strings.TrimSpace(addPostHook)

		// Check if this post-hook already exists
		hookExists := false
		for _, h := range hooks {
			hook, ok := h.(map[string]interface{})
			if !ok {
				continue
			}
			existingHookName, _, _ := unstructured.NestedString(hook, "hook", "name")
			step, _, _ := unstructured.NestedString(hook, "step")
			if existingHookName == hookName && step == "PostHook" {
				hookExists = true
				break
			}
		}

		if !hookExists {
			postHookRef := map[string]interface{}{
				"step": "PostHook",
				"hook": map[string]interface{}{
					"kind":       "Hook",
					"apiVersion": "forklift.konveyor.io/v1beta1",
					"name":       hookName,
					"namespace":  namespace,
				},
			}
			hooks = append(hooks, postHookRef)
			err := unstructured.SetNestedSlice(vm, hooks, "hooks")
			if err != nil {
				return false, fmt.Errorf("failed to add post-hook: %v", err)
			}
			klog.V(2).Infof("Added post-hook '%s' to VM", hookName)
			updated = true
		} else {
			klog.V(1).Infof("Post-hook '%s' already exists for VM, skipping", hookName)
		}
	}

	return updated, nil
}
