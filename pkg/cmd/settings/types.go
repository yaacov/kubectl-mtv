// Package settings provides types and utilities for managing ForkliftController settings.
package settings

import "sort"

// SettingType represents the data type of a setting.
type SettingType string

const (
	// TypeString represents a string setting.
	TypeString SettingType = "string"
	// TypeBool represents a boolean setting.
	TypeBool SettingType = "bool"
	// TypeInt represents an integer setting.
	TypeInt SettingType = "int"
)

// SettingCategory represents the category of a setting.
type SettingCategory string

const (
	// CategoryImage represents container image settings.
	CategoryImage SettingCategory = "image"
	// CategoryFeature represents feature flag settings.
	CategoryFeature SettingCategory = "feature"
	// CategoryPerformance represents performance tuning settings.
	CategoryPerformance SettingCategory = "performance"
	// CategoryDebug represents debugging settings.
	CategoryDebug SettingCategory = "debug"
	// CategoryVirtV2V represents virt-v2v container settings.
	CategoryVirtV2V SettingCategory = "virt-v2v"
	// CategoryPopulator represents volume populator container settings.
	CategoryPopulator SettingCategory = "populator"
	// CategoryHook represents hook container settings.
	CategoryHook SettingCategory = "hook"
	// CategoryOVA represents OVA provider server container settings.
	CategoryOVA SettingCategory = "ova"
	// CategoryHyperV represents HyperV provider server container settings.
	CategoryHyperV SettingCategory = "hyperv"
)

// SettingDefinition defines metadata for a ForkliftController setting.
type SettingDefinition struct {
	// Name is the setting key in the ForkliftController spec (snake_case).
	Name string
	// Type is the data type of the setting.
	Type SettingType
	// Default is the default value if not set.
	Default interface{}
	// Description is a human-readable description of the setting.
	Description string
	// Category groups related settings together.
	Category SettingCategory
}

// SettingValue represents a setting with its current and default values.
type SettingValue struct {
	// Name is the setting key.
	Name string
	// Value is the current value (nil if not set).
	Value interface{}
	// Default is the default value.
	Default interface{}
	// IsSet indicates whether the value is explicitly set.
	IsSet bool
	// Definition contains the setting metadata.
	Definition SettingDefinition
}

// CategoryOrder defines the display order for categories.
var CategoryOrder = []SettingCategory{
	CategoryImage,
	CategoryFeature,
	CategoryPerformance,
	CategoryDebug,
	CategoryVirtV2V,
	CategoryPopulator,
	CategoryHook,
	CategoryOVA,
	CategoryHyperV,
}

// SupportedSettings contains all supported ForkliftController settings.
// This is a curated subset of settings that users commonly need to configure.
var SupportedSettings = map[string]SettingDefinition{
	// Container Images
	"vddk_image": {
		Name:        "vddk_image",
		Type:        TypeString,
		Default:     "",
		Description: "VDDK container image for vSphere migrations",
		Category:    CategoryImage,
	},
	"virt_v2v_image_fqin": {
		Name:        "virt_v2v_image_fqin",
		Type:        TypeString,
		Default:     "",
		Description: "Custom virt-v2v container image",
		Category:    CategoryImage,
	},

	// Feature Flags
	"controller_vsphere_incremental_backup": {
		Name:        "controller_vsphere_incremental_backup",
		Type:        TypeBool,
		Default:     true,
		Description: "Enable CBT-based warm migration for vSphere",
		Category:    CategoryFeature,
	},
	"controller_ovirt_warm_migration": {
		Name:        "controller_ovirt_warm_migration",
		Type:        TypeBool,
		Default:     true,
		Description: "Enable warm migration from oVirt",
		Category:    CategoryFeature,
	},
	"feature_copy_offload": {
		Name:        "feature_copy_offload",
		Type:        TypeBool,
		Default:     false,
		Description: "Enable storage array offload (XCOPY)",
		Category:    CategoryFeature,
	},
	"feature_ocp_live_migration": {
		Name:        "feature_ocp_live_migration",
		Type:        TypeBool,
		Default:     false,
		Description: "Enable cross-cluster OpenShift live migration",
		Category:    CategoryFeature,
	},
	"feature_vmware_system_serial_number": {
		Name:        "feature_vmware_system_serial_number",
		Type:        TypeBool,
		Default:     true,
		Description: "Use VMware system serial number for migrated VMs",
		Category:    CategoryFeature,
	},
	"controller_static_udn_ip_addresses": {
		Name:        "controller_static_udn_ip_addresses",
		Type:        TypeBool,
		Default:     true,
		Description: "Enable static IP addresses with User Defined Networks",
		Category:    CategoryFeature,
	},
	"controller_retain_precopy_importer_pods": {
		Name:        "controller_retain_precopy_importer_pods",
		Type:        TypeBool,
		Default:     false,
		Description: "Retain importer pods during warm migration (debugging)",
		Category:    CategoryFeature,
	},
	"feature_ova_appliance_management": {
		Name:        "feature_ova_appliance_management",
		Type:        TypeBool,
		Default:     false,
		Description: "Enable appliance management for OVF-based providers",
		Category:    CategoryFeature,
	},

	// Performance Tuning
	"controller_max_vm_inflight": {
		Name:        "controller_max_vm_inflight",
		Type:        TypeInt,
		Default:     20,
		Description: "Maximum concurrent VM migrations",
		Category:    CategoryPerformance,
	},
	"controller_precopy_interval": {
		Name:        "controller_precopy_interval",
		Type:        TypeInt,
		Default:     60,
		Description: "Minutes between warm migration precopies",
		Category:    CategoryPerformance,
	},
	"controller_max_concurrent_reconciles": {
		Name:        "controller_max_concurrent_reconciles",
		Type:        TypeInt,
		Default:     10,
		Description: "Maximum concurrent controller reconciles",
		Category:    CategoryPerformance,
	},
	"controller_snapshot_removal_timeout_minuts": {
		Name:        "controller_snapshot_removal_timeout_minuts",
		Type:        TypeInt,
		Default:     120,
		Description: "Timeout for snapshot removal (minutes)",
		Category:    CategoryPerformance,
	},
	"controller_vddk_job_active_deadline_sec": {
		Name:        "controller_vddk_job_active_deadline_sec",
		Type:        TypeInt,
		Default:     300,
		Description: "VDDK validation job deadline (seconds)",
		Category:    CategoryPerformance,
	},
	"controller_filesystem_overhead": {
		Name:        "controller_filesystem_overhead",
		Type:        TypeInt,
		Default:     10,
		Description: "Filesystem overhead percentage",
		Category:    CategoryPerformance,
	},
	"controller_block_overhead": {
		Name:        "controller_block_overhead",
		Type:        TypeInt,
		Default:     0,
		Description: "Block storage fixed overhead (bytes)",
		Category:    CategoryPerformance,
	},
	"controller_cleanup_retries": {
		Name:        "controller_cleanup_retries",
		Type:        TypeInt,
		Default:     10,
		Description: "Maximum cleanup retry attempts",
		Category:    CategoryPerformance,
	},
	"controller_snapshot_removal_check_retries": {
		Name:        "controller_snapshot_removal_check_retries",
		Type:        TypeInt,
		Default:     20,
		Description: "Maximum snapshot removal check retries",
		Category:    CategoryPerformance,
	},
	"controller_host_lease_namespace": {
		Name:        "controller_host_lease_namespace",
		Type:        TypeString,
		Default:     "openshift-mtv",
		Description: "Namespace for host lease objects (copy offload)",
		Category:    CategoryPerformance,
	},
	"controller_host_lease_duration_seconds": {
		Name:        "controller_host_lease_duration_seconds",
		Type:        TypeInt,
		Default:     10,
		Description: "Host lease duration in seconds (copy offload)",
		Category:    CategoryPerformance,
	},

	// Debugging
	"controller_log_level": {
		Name:        "controller_log_level",
		Type:        TypeInt,
		Default:     3,
		Description: "Controller log verbosity (0-9)",
		Category:    CategoryDebug,
	},

	// virt-v2v Container Settings
	"virt_v2v_extra_args": {
		Name:        "virt_v2v_extra_args",
		Type:        TypeString,
		Default:     "",
		Description: "Additional virt-v2v command-line arguments",
		Category:    CategoryVirtV2V,
	},
	"virt_v2v_dont_request_kvm": {
		Name:        "virt_v2v_dont_request_kvm",
		Type:        TypeBool,
		Default:     false,
		Description: "Don't request KVM device (use for nested virtualization)",
		Category:    CategoryVirtV2V,
	},
	"virt_v2v_extra_conf_config_map": {
		Name:        "virt_v2v_extra_conf_config_map",
		Type:        TypeString,
		Default:     "",
		Description: "ConfigMap with extra virt-v2v configuration files",
		Category:    CategoryVirtV2V,
	},
	"virt_v2v_container_limits_cpu": {
		Name:        "virt_v2v_container_limits_cpu",
		Type:        TypeString,
		Default:     "4000m",
		Description: "virt-v2v container CPU limit",
		Category:    CategoryVirtV2V,
	},
	"virt_v2v_container_limits_memory": {
		Name:        "virt_v2v_container_limits_memory",
		Type:        TypeString,
		Default:     "8Gi",
		Description: "virt-v2v container memory limit",
		Category:    CategoryVirtV2V,
	},
	"virt_v2v_container_requests_cpu": {
		Name:        "virt_v2v_container_requests_cpu",
		Type:        TypeString,
		Default:     "1000m",
		Description: "virt-v2v container CPU request",
		Category:    CategoryVirtV2V,
	},
	"virt_v2v_container_requests_memory": {
		Name:        "virt_v2v_container_requests_memory",
		Type:        TypeString,
		Default:     "1Gi",
		Description: "virt-v2v container memory request",
		Category:    CategoryVirtV2V,
	},

	// Volume Populator Container Settings
	"populator_container_limits_cpu": {
		Name:        "populator_container_limits_cpu",
		Type:        TypeString,
		Default:     "1000m",
		Description: "Volume populator container CPU limit",
		Category:    CategoryPopulator,
	},
	"populator_container_limits_memory": {
		Name:        "populator_container_limits_memory",
		Type:        TypeString,
		Default:     "1Gi",
		Description: "Volume populator container memory limit",
		Category:    CategoryPopulator,
	},
	"populator_container_requests_cpu": {
		Name:        "populator_container_requests_cpu",
		Type:        TypeString,
		Default:     "100m",
		Description: "Volume populator container CPU request",
		Category:    CategoryPopulator,
	},
	"populator_container_requests_memory": {
		Name:        "populator_container_requests_memory",
		Type:        TypeString,
		Default:     "512Mi",
		Description: "Volume populator container memory request",
		Category:    CategoryPopulator,
	},

	// Hook Container Settings
	"hooks_container_limits_cpu": {
		Name:        "hooks_container_limits_cpu",
		Type:        TypeString,
		Default:     "1000m",
		Description: "Hook container CPU limit",
		Category:    CategoryHook,
	},
	"hooks_container_limits_memory": {
		Name:        "hooks_container_limits_memory",
		Type:        TypeString,
		Default:     "1Gi",
		Description: "Hook container memory limit",
		Category:    CategoryHook,
	},
	"hooks_container_requests_cpu": {
		Name:        "hooks_container_requests_cpu",
		Type:        TypeString,
		Default:     "100m",
		Description: "Hook container CPU request",
		Category:    CategoryHook,
	},
	"hooks_container_requests_memory": {
		Name:        "hooks_container_requests_memory",
		Type:        TypeString,
		Default:     "150Mi",
		Description: "Hook container memory request",
		Category:    CategoryHook,
	},

	// OVA Provider Server Container Settings
	"ova_container_limits_cpu": {
		Name:        "ova_container_limits_cpu",
		Type:        TypeString,
		Default:     "1000m",
		Description: "OVA provider server CPU limit",
		Category:    CategoryOVA,
	},
	"ova_container_limits_memory": {
		Name:        "ova_container_limits_memory",
		Type:        TypeString,
		Default:     "1Gi",
		Description: "OVA provider server memory limit",
		Category:    CategoryOVA,
	},
	"ova_container_requests_cpu": {
		Name:        "ova_container_requests_cpu",
		Type:        TypeString,
		Default:     "100m",
		Description: "OVA provider server CPU request",
		Category:    CategoryOVA,
	},
	"ova_container_requests_memory": {
		Name:        "ova_container_requests_memory",
		Type:        TypeString,
		Default:     "512Mi",
		Description: "OVA provider server memory request",
		Category:    CategoryOVA,
	},

	// HyperV Provider Server Container Settings
	"hyperv_container_limits_cpu": {
		Name:        "hyperv_container_limits_cpu",
		Type:        TypeString,
		Default:     "1000m",
		Description: "HyperV provider server CPU limit",
		Category:    CategoryHyperV,
	},
	"hyperv_container_limits_memory": {
		Name:        "hyperv_container_limits_memory",
		Type:        TypeString,
		Default:     "1Gi",
		Description: "HyperV provider server memory limit",
		Category:    CategoryHyperV,
	},
	"hyperv_container_requests_cpu": {
		Name:        "hyperv_container_requests_cpu",
		Type:        TypeString,
		Default:     "100m",
		Description: "HyperV provider server CPU request",
		Category:    CategoryHyperV,
	},
	"hyperv_container_requests_memory": {
		Name:        "hyperv_container_requests_memory",
		Type:        TypeString,
		Default:     "512Mi",
		Description: "HyperV provider server memory request",
		Category:    CategoryHyperV,
	},
}

// GetSettingNames returns all supported setting names in a consistent order.
func GetSettingNames() []string {
	var names []string
	for _, category := range CategoryOrder {
		// Collect names for this category
		var categoryNames []string
		for name, def := range SupportedSettings {
			if def.Category == category {
				categoryNames = append(categoryNames, name)
			}
		}
		// Sort names within category for deterministic ordering
		sort.Strings(categoryNames)
		names = append(names, categoryNames...)
	}
	return names
}

// GetSettingsByCategory returns settings grouped by category.
func GetSettingsByCategory() map[SettingCategory][]SettingDefinition {
	result := make(map[SettingCategory][]SettingDefinition)
	for _, def := range SupportedSettings {
		result[def.Category] = append(result[def.Category], def)
	}
	return result
}

// IsValidSetting checks if a setting name is supported.
func IsValidSetting(name string) bool {
	_, ok := SupportedSettings[name]
	return ok
}

// GetSettingDefinition returns the definition for a setting, or nil if not found.
func GetSettingDefinition(name string) *SettingDefinition {
	if def, ok := SupportedSettings[name]; ok {
		return &def
	}
	return nil
}
