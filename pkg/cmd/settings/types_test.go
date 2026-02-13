package settings

import (
	"sort"
	"testing"
)

// --- GetAllSettings ---

func TestGetAllSettings_MergesMaps(t *testing.T) {
	all := GetAllSettings()

	// Must include entries from both SupportedSettings and ExtendedSettings
	for name := range SupportedSettings {
		if _, ok := all[name]; !ok {
			t.Errorf("GetAllSettings() missing supported setting %q", name)
		}
	}
	for name := range ExtendedSettings {
		if _, ok := all[name]; !ok {
			t.Errorf("GetAllSettings() missing extended setting %q", name)
		}
	}

	expected := len(SupportedSettings) + len(ExtendedSettings)
	if len(all) != expected {
		t.Errorf("GetAllSettings() returned %d entries, expected %d", len(all), expected)
	}
}

func TestGetAllSettings_ReturnsCopy(t *testing.T) {
	all1 := GetAllSettings()
	all2 := GetAllSettings()

	// Mutating one should not affect the other
	all1["__test_key__"] = SettingDefinition{Name: "test"}
	if _, ok := all2["__test_key__"]; ok {
		t.Error("GetAllSettings() should return a new map each time")
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

	// Group names by category and check sorting within each group
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

	// Verify each setting is in the correct category
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

	// SupportedSettings should have at least image, feature, performance, debug categories
	expected := []SettingCategory{CategoryImage, CategoryFeature, CategoryPerformance, CategoryDebug}
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

func TestIsValidSetting_ExtendedSettingsReturnFalse(t *testing.T) {
	for name := range ExtendedSettings {
		// Extended settings that are NOT in SupportedSettings should return false
		if _, ok := SupportedSettings[name]; ok {
			continue // Skip overlapping entries
		}
		if IsValidSetting(name) {
			t.Errorf("IsValidSetting(%q) = true for extended-only setting, want false", name)
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

func TestIsValidAnySetting_ExtendedSettings(t *testing.T) {
	for name := range ExtendedSettings {
		if !IsValidAnySetting(name) {
			t.Errorf("IsValidAnySetting(%q) = false for extended setting", name)
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

func TestGetSettingDefinition_NotFoundExtended(t *testing.T) {
	// "feature_ui_plugin" is only in ExtendedSettings
	def := GetSettingDefinition("feature_ui_plugin")
	if def != nil {
		t.Error("GetSettingDefinition(extended-only) should return nil")
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
	}
	if def.Name != "vddk_image" {
		t.Errorf("def.Name = %q, want %q", def.Name, "vddk_image")
	}
}

func TestGetAnySettingDefinition_Extended(t *testing.T) {
	def := GetAnySettingDefinition("feature_ui_plugin")
	if def == nil {
		t.Fatal("GetAnySettingDefinition(extended) = nil")
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
		CategoryOVAProxy:    true,
		CategoryConfigMaps:  true,
		CategoryAdvanced:    true,
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

func TestExtendedSettings_NameFieldConsistency(t *testing.T) {
	for key, def := range ExtendedSettings {
		if key != def.Name {
			t.Errorf("ExtendedSettings key %q != def.Name %q", key, def.Name)
		}
		if def.Description == "" {
			t.Errorf("ExtendedSettings[%q] has empty Description", key)
		}
		if def.Type != TypeString && def.Type != TypeBool && def.Type != TypeInt {
			t.Errorf("ExtendedSettings[%q] has invalid Type %q", key, def.Type)
		}
	}
}

func TestNoOverlapBetweenSupportedAndExtended(t *testing.T) {
	for name := range SupportedSettings {
		if _, ok := ExtendedSettings[name]; ok {
			t.Errorf("setting %q exists in both SupportedSettings and ExtendedSettings", name)
		}
	}
}
