package settings

import (
	"sort"
	"testing"
)

// --- GetAllSettings ---

func TestGetAllSettings_ReturnsAllEntries(t *testing.T) {
	all := GetAllSettings()

	// AllSettings must include all supported settings
	for name := range SupportedSettings {
		if _, ok := all[name]; !ok {
			t.Errorf("GetAllSettings() missing supported setting %q", name)
		}
	}

	// Must have more entries than just supported (extended settings exist)
	if len(all) <= len(SupportedSettings) {
		t.Errorf("GetAllSettings() returned %d entries, expected more than %d supported", len(all), len(SupportedSettings))
	}
}

func TestGetAllSettings_IsAllSettings(t *testing.T) {
	all := GetAllSettings()
	if len(all) != len(AllSettings) {
		t.Errorf("GetAllSettings() size %d != AllSettings size %d", len(all), len(AllSettings))
	}
}

// --- GetAllSettingNames ---

func TestGetAllSettingNames_ContainsAll(t *testing.T) {
	names := GetAllSettingNames()
	all := GetAllSettings()

	if len(names) != len(all) {
		t.Errorf("GetAllSettingNames() returned %d names, expected %d", len(names), len(all))
	}

	nameSet := map[string]bool{}
	for _, n := range names {
		nameSet[n] = true
	}
	for name := range all {
		if !nameSet[name] {
			t.Errorf("GetAllSettingNames() missing %q", name)
		}
	}
}

func TestGetAllSettingNames_NoDuplicates(t *testing.T) {
	names := GetAllSettingNames()
	seen := map[string]bool{}
	for _, n := range names {
		if seen[n] {
			t.Errorf("GetAllSettingNames() has duplicate %q", n)
		}
		seen[n] = true
	}
}

// --- GetSettingNames ---

func TestGetSettingNames_ContainsSupportedOnly(t *testing.T) {
	names := GetSettingNames()

	if len(names) != len(SupportedSettings) {
		t.Errorf("GetSettingNames() returned %d names, expected %d", len(names), len(SupportedSettings))
	}

	for _, n := range names {
		if _, ok := SupportedSettings[n]; !ok {
			t.Errorf("GetSettingNames() returned non-supported setting %q", n)
		}
	}
}

func TestGetSettingNames_SortedWithinCategory(t *testing.T) {
	names := GetSettingNames()

	byCategory := map[SettingCategory][]string{}
	for _, n := range names {
		def := SupportedSettings[n]
		byCategory[def.Category] = append(byCategory[def.Category], n)
	}

	for cat, catNames := range byCategory {
		if !sort.StringsAreSorted(catNames) {
			t.Errorf("GetSettingNames() not sorted within category %q: %v", cat, catNames)
		}
	}
}

// --- GetSettingsByCategory ---

func TestGetSettingsByCategory_GroupsCorrectly(t *testing.T) {
	grouped := GetSettingsByCategory()

	totalCount := 0
	for _, defs := range grouped {
		totalCount += len(defs)
	}
	if totalCount != len(SupportedSettings) {
		t.Errorf("GetSettingsByCategory() total %d, expected %d", totalCount, len(SupportedSettings))
	}

	for cat, defs := range grouped {
		for _, def := range defs {
			if def.Category != cat {
				t.Errorf("setting %q has category %q but grouped under %q", def.Name, def.Category, cat)
			}
		}
	}
}

func TestGetSettingsByCategory_HasExpectedCategories(t *testing.T) {
	grouped := GetSettingsByCategory()

	expected := []SettingCategory{CategoryImage, CategoryFeature, CategoryPerformance}
	for _, cat := range expected {
		if _, ok := grouped[cat]; !ok {
			t.Errorf("GetSettingsByCategory() missing category %q", cat)
		}
	}
}

// --- IsValidSetting ---

func TestIsValidSetting_KnownSettings(t *testing.T) {
	for name := range SupportedSettings {
		if !IsValidSetting(name) {
			t.Errorf("IsValidSetting(%q) = false, want true", name)
		}
	}
}

func TestIsValidSetting_ExtendedOnlyReturnFalse(t *testing.T) {
	for name := range AllSettings {
		if _, ok := SupportedSettings[name]; ok {
			continue
		}
		if IsValidSetting(name) {
			t.Errorf("IsValidSetting(%q) = true for non-supported setting, want false", name)
		}
	}
}

func TestIsValidSetting_InvalidName(t *testing.T) {
	if IsValidSetting("nonexistent_setting_xyzzy") {
		t.Error("IsValidSetting(nonexistent) = true, want false")
	}
}

// --- IsValidAnySetting ---

func TestIsValidAnySetting_SupportedSettings(t *testing.T) {
	for name := range SupportedSettings {
		if !IsValidAnySetting(name) {
			t.Errorf("IsValidAnySetting(%q) = false for supported setting", name)
		}
	}
}

func TestIsValidAnySetting_AllSettings(t *testing.T) {
	for name := range AllSettings {
		if !IsValidAnySetting(name) {
			t.Errorf("IsValidAnySetting(%q) = false for known setting", name)
		}
	}
}

func TestIsValidAnySetting_InvalidName(t *testing.T) {
	if IsValidAnySetting("nonexistent_setting_xyzzy") {
		t.Error("IsValidAnySetting(nonexistent) = true, want false")
	}
}

// --- GetSettingDefinition ---

func TestGetSettingDefinition_Found(t *testing.T) {
	def := GetSettingDefinition("vddk_image")
	if def == nil {
		t.Fatal("GetSettingDefinition(vddk_image) = nil, want non-nil")
		return
	}
	if def.Name != "vddk_image" {
		t.Errorf("def.Name = %q, want %q", def.Name, "vddk_image")
	}
	if def.Type != TypeString {
		t.Errorf("def.Type = %q, want %q", def.Type, TypeString)
	}
	if def.Category != CategoryImage {
		t.Errorf("def.Category = %q, want %q", def.Category, CategoryImage)
	}
}

func TestGetSettingDefinition_NotFoundNonSupported(t *testing.T) {
	// "feature_ui_plugin" is in AllSettings but NOT in SupportedSettingNames
	def := GetSettingDefinition("feature_ui_plugin")
	if def != nil {
		t.Error("GetSettingDefinition(non-supported) should return nil")
	}
}

func TestGetSettingDefinition_NotFoundInvalid(t *testing.T) {
	def := GetSettingDefinition("nonexistent_setting")
	if def != nil {
		t.Error("GetSettingDefinition(nonexistent) should return nil")
	}
}

// --- GetAnySettingDefinition ---

func TestGetAnySettingDefinition_Supported(t *testing.T) {
	def := GetAnySettingDefinition("vddk_image")
	if def == nil {
		t.Fatal("GetAnySettingDefinition(supported) = nil")
		return
	}
	if def.Name != "vddk_image" {
		t.Errorf("def.Name = %q, want %q", def.Name, "vddk_image")
	}
}

func TestGetAnySettingDefinition_NonSupported(t *testing.T) {
	def := GetAnySettingDefinition("feature_ui_plugin")
	if def == nil {
		t.Fatal("GetAnySettingDefinition(non-supported) = nil")
		return
	}
	if def.Name != "feature_ui_plugin" {
		t.Errorf("def.Name = %q, want %q", def.Name, "feature_ui_plugin")
	}
}

func TestGetAnySettingDefinition_NotFound(t *testing.T) {
	def := GetAnySettingDefinition("nonexistent_setting")
	if def != nil {
		t.Error("GetAnySettingDefinition(nonexistent) should return nil")
	}
}

func TestCopyOffloadSettingExists(t *testing.T) {
	def := GetSettingDefinition("populator_vsphere_copy_offload_image_fqin")
	if def == nil {
		t.Fatal("populator_vsphere_copy_offload_image_fqin must exist in SupportedSettings")
		return
	}
	if def.Category != CategoryImage {
		t.Errorf("def.Category = %q, want %q", def.Category, CategoryImage)
	}
}

// --- CategoryOrder ---

func TestCategoryOrder_ContainsExpectedCategories(t *testing.T) {
	expected := map[SettingCategory]bool{
		CategoryImage:       true,
		CategoryFeature:     true,
		CategoryPerformance: true,
		CategoryDebug:       true,
		CategoryVirtV2V:     true,
		CategoryPopulator:   true,
		CategoryHook:        true,
		CategoryOVA:         true,
		CategoryHyperV:      true,
		CategoryController:  true,
		CategoryInventory:   true,
		CategoryAPI:         true,
		CategoryUIPlugin:    true,
		CategoryValidation:  true,
		CategoryCLIDownload: true,
		CategoryMCP:         true,
		CategoryOVAProxy:    true,
		CategoryConfigMaps:  true,
		CategoryAdvanced:    true,
		CategoryAAP:         true,
	}

	for _, cat := range CategoryOrder {
		if !expected[cat] {
			t.Errorf("unexpected category %q in CategoryOrder", cat)
		}
		delete(expected, cat)
	}
	for cat := range expected {
		t.Errorf("missing category %q in CategoryOrder", cat)
	}
}

// --- Setting definitions consistency ---

func TestSupportedSettings_NameFieldConsistency(t *testing.T) {
	for key, def := range SupportedSettings {
		if key != def.Name {
			t.Errorf("SupportedSettings key %q != def.Name %q", key, def.Name)
		}
		if def.Description == "" {
			t.Errorf("SupportedSettings[%q] has empty Description", key)
		}
		if def.Type != TypeString && def.Type != TypeBool && def.Type != TypeInt {
			t.Errorf("SupportedSettings[%q] has invalid Type %q", key, def.Type)
		}
	}
}

func TestAllSettings_NameFieldConsistency(t *testing.T) {
	for key, def := range AllSettings {
		if key != def.Name {
			t.Errorf("AllSettings key %q != def.Name %q", key, def.Name)
		}
		if def.Description == "" {
			t.Errorf("AllSettings[%q] has empty Description", key)
		}
		if def.Type != TypeString && def.Type != TypeBool && def.Type != TypeInt {
			t.Errorf("AllSettings[%q] has invalid Type %q", key, def.Type)
		}
	}
}

func TestSupportedSettingNames_AllExistInAllSettings(t *testing.T) {
	for _, name := range SupportedSettingNames {
		if _, ok := AllSettings[name]; !ok {
			t.Errorf("SupportedSettingNames entry %q not found in AllSettings", name)
		}
	}
}
