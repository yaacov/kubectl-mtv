package settings

import (
	"errors"
	"strings"
	"testing"
)

// --- FormatValue ---

func TestFormatValue_NotSet(t *testing.T) {
	sv := SettingValue{IsSet: false}
	if got := FormatValue(sv); got != "(not set)" {
		t.Errorf("FormatValue(not set) = %q, want %q", got, "(not set)")
	}
}

func TestFormatValue_NilValue(t *testing.T) {
	sv := SettingValue{IsSet: true, Value: nil}
	if got := FormatValue(sv); got != "(not set)" {
		t.Errorf("FormatValue(nil) = %q, want %q", got, "(not set)")
	}
}

func TestFormatValue_Bool(t *testing.T) {
	tests := []struct {
		value    bool
		expected string
	}{
		{true, "true"},
		{false, "false"},
	}
	for _, tt := range tests {
		sv := SettingValue{IsSet: true, Value: tt.value}
		if got := FormatValue(sv); got != tt.expected {
			t.Errorf("FormatValue(%v) = %q, want %q", tt.value, got, tt.expected)
		}
	}
}

func TestFormatValue_Int(t *testing.T) {
	sv := SettingValue{IsSet: true, Value: 42}
	if got := FormatValue(sv); got != "42" {
		t.Errorf("FormatValue(42) = %q, want %q", got, "42")
	}
}

func TestFormatValue_String(t *testing.T) {
	sv := SettingValue{IsSet: true, Value: "quay.io/test:latest"}
	if got := FormatValue(sv); got != "quay.io/test:latest" {
		t.Errorf("FormatValue(string) = %q", got)
	}
}

func TestFormatValue_EmptyString(t *testing.T) {
	sv := SettingValue{IsSet: true, Value: ""}
	if got := FormatValue(sv); got != "(empty)" {
		t.Errorf("FormatValue(empty string) = %q, want %q", got, "(empty)")
	}
}

// --- FormatDefault ---

func TestFormatDefault_Bool(t *testing.T) {
	def := SettingDefinition{Default: true}
	if got := FormatDefault(def); got != "true" {
		t.Errorf("FormatDefault(true) = %q", got)
	}
	def.Default = false
	if got := FormatDefault(def); got != "false" {
		t.Errorf("FormatDefault(false) = %q", got)
	}
}

func TestFormatDefault_Int(t *testing.T) {
	def := SettingDefinition{Default: 20}
	if got := FormatDefault(def); got != "20" {
		t.Errorf("FormatDefault(20) = %q", got)
	}
}

func TestFormatDefault_String(t *testing.T) {
	def := SettingDefinition{Default: "4000m"}
	if got := FormatDefault(def); got != "4000m" {
		t.Errorf("FormatDefault(string) = %q", got)
	}
}

func TestFormatDefault_EmptyString(t *testing.T) {
	def := SettingDefinition{Default: ""}
	if got := FormatDefault(def); got != "-" {
		t.Errorf("FormatDefault(empty) = %q, want %q", got, "-")
	}
}

// --- wrapClusterError ---

func TestWrapClusterError_ConnectionRefused(t *testing.T) {
	err := wrapClusterError(errors.New("dial tcp: connection refused"), "test")
	if !strings.Contains(err.Error(), "cannot connect to Kubernetes cluster") {
		t.Errorf("expected connection refused message, got: %s", err.Error())
	}
}

func TestWrapClusterError_Unauthorized(t *testing.T) {
	err := wrapClusterError(errors.New("Unauthorized"), "test")
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("expected auth failed message, got: %s", err.Error())
	}

	err = wrapClusterError(errors.New("unauthorized access"), "test")
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("expected auth failed message for lowercase, got: %s", err.Error())
	}
}

func TestWrapClusterError_Forbidden(t *testing.T) {
	err := wrapClusterError(errors.New("forbidden"), "test")
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected permission denied message, got: %s", err.Error())
	}

	err = wrapClusterError(errors.New("Forbidden: user cannot get"), "test")
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected permission denied for capitalized, got: %s", err.Error())
	}
}

func TestWrapClusterError_ResourceNotFound(t *testing.T) {
	err := wrapClusterError(errors.New("the server could not find the requested resource"), "test")
	if !strings.Contains(err.Error(), "MTV") {
		t.Errorf("expected MTV not installed message, got: %s", err.Error())
	}

	err = wrapClusterError(errors.New("no matches for kind \"ForkliftController\""), "test")
	if !strings.Contains(err.Error(), "MTV") {
		t.Errorf("expected MTV not installed for no matches, got: %s", err.Error())
	}
}

func TestWrapClusterError_GenericError(t *testing.T) {
	err := wrapClusterError(errors.New("some random error"), "do something")
	if !strings.HasPrefix(err.Error(), "do something:") {
		t.Errorf("expected operation prefix, got: %s", err.Error())
	}
}

// --- extractSettingValue ---

func TestExtractSettingValue_NilSpec(t *testing.T) {
	def := SettingDefinition{Name: "test", Type: TypeString, Default: "default"}
	sv := extractSettingValue(nil, def)

	if sv.IsSet {
		t.Error("expected IsSet=false for nil spec")
	}
	if sv.Default != "default" {
		t.Errorf("expected Default=%q, got %v", "default", sv.Default)
	}
}

func TestExtractSettingValue_MissingKey(t *testing.T) {
	spec := map[string]interface{}{}
	def := SettingDefinition{Name: "missing_key", Type: TypeString, Default: "def"}
	sv := extractSettingValue(spec, def)

	if sv.IsSet {
		t.Error("expected IsSet=false for missing key")
	}
}

func TestExtractSettingValue_String(t *testing.T) {
	spec := map[string]interface{}{"vddk_image": "quay.io/test:v1"}
	def := SettingDefinition{Name: "vddk_image", Type: TypeString, Default: ""}
	sv := extractSettingValue(spec, def)

	if !sv.IsSet {
		t.Error("expected IsSet=true")
	}
	if sv.Value != "quay.io/test:v1" {
		t.Errorf("expected Value=%q, got %v", "quay.io/test:v1", sv.Value)
	}
}

func TestExtractSettingValue_BoolNative(t *testing.T) {
	spec := map[string]interface{}{"feature": true}
	def := SettingDefinition{Name: "feature", Type: TypeBool, Default: false}
	sv := extractSettingValue(spec, def)

	if !sv.IsSet {
		t.Error("expected IsSet=true")
	}
	if sv.Value != true {
		t.Errorf("expected Value=true, got %v", sv.Value)
	}
}

func TestExtractSettingValue_BoolFromString(t *testing.T) {
	spec := map[string]interface{}{"feature": "true"}
	def := SettingDefinition{Name: "feature", Type: TypeBool, Default: false}
	sv := extractSettingValue(spec, def)

	if !sv.IsSet {
		t.Error("expected IsSet=true")
	}
	if sv.Value != true {
		t.Errorf("expected Value=true from string, got %v", sv.Value)
	}
}

func TestExtractSettingValue_IntFromFloat64(t *testing.T) {
	// JSON unmarshalling produces float64 for numbers
	spec := map[string]interface{}{"max_vm": float64(30)}
	def := SettingDefinition{Name: "max_vm", Type: TypeInt, Default: 20}
	sv := extractSettingValue(spec, def)

	if !sv.IsSet {
		t.Error("expected IsSet=true")
	}
	if sv.Value != 30 {
		t.Errorf("expected Value=30 from float64, got %v", sv.Value)
	}
}

func TestExtractSettingValue_IntFromInt64(t *testing.T) {
	spec := map[string]interface{}{"max_vm": int64(25)}
	def := SettingDefinition{Name: "max_vm", Type: TypeInt, Default: 20}
	sv := extractSettingValue(spec, def)

	if sv.Value != 25 {
		t.Errorf("expected Value=25 from int64, got %v", sv.Value)
	}
}

func TestExtractSettingValue_IntFromString(t *testing.T) {
	spec := map[string]interface{}{"max_vm": "42"}
	def := SettingDefinition{Name: "max_vm", Type: TypeInt, Default: 20}
	sv := extractSettingValue(spec, def)

	if sv.Value != 42 {
		t.Errorf("expected Value=42 from string, got %v", sv.Value)
	}
}

// --- validateAndConvertValue ---

func TestValidateAndConvertValue_BoolTrue(t *testing.T) {
	def := SettingDefinition{Type: TypeBool}
	trueValues := []string{"true", "True", "TRUE", "1", "yes", "Yes", "YES"}
	for _, v := range trueValues {
		result, err := validateAndConvertValue(v, def)
		if err != nil {
			t.Errorf("validateAndConvertValue(%q) error: %v", v, err)
		}
		if result != true {
			t.Errorf("validateAndConvertValue(%q) = %v, want true", v, result)
		}
	}
}

func TestValidateAndConvertValue_BoolFalse(t *testing.T) {
	def := SettingDefinition{Type: TypeBool}
	falseValues := []string{"false", "False", "FALSE", "0", "no", "No", "NO"}
	for _, v := range falseValues {
		result, err := validateAndConvertValue(v, def)
		if err != nil {
			t.Errorf("validateAndConvertValue(%q) error: %v", v, err)
		}
		if result != false {
			t.Errorf("validateAndConvertValue(%q) = %v, want false", v, result)
		}
	}
}

func TestValidateAndConvertValue_BoolInvalid(t *testing.T) {
	def := SettingDefinition{Type: TypeBool}
	_, err := validateAndConvertValue("maybe", def)
	if err == nil {
		t.Error("expected error for invalid bool value")
	}
	if !strings.Contains(err.Error(), "expected boolean") {
		t.Errorf("expected bool error message, got: %s", err.Error())
	}
}

func TestValidateAndConvertValue_Int(t *testing.T) {
	def := SettingDefinition{Type: TypeInt}
	result, err := validateAndConvertValue("42", def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestValidateAndConvertValue_IntInvalid(t *testing.T) {
	def := SettingDefinition{Type: TypeInt}
	_, err := validateAndConvertValue("not-a-number", def)
	if err == nil {
		t.Error("expected error for invalid int value")
	}
	if !strings.Contains(err.Error(), "expected integer") {
		t.Errorf("expected int error message, got: %s", err.Error())
	}
}

func TestValidateAndConvertValue_String(t *testing.T) {
	def := SettingDefinition{Type: TypeString}
	result, err := validateAndConvertValue("any-value", def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "any-value" {
		t.Errorf("expected 'any-value', got %v", result)
	}
}

func TestValidateAndConvertValue_EmptyString(t *testing.T) {
	def := SettingDefinition{Type: TypeString}
	result, err := validateAndConvertValue("", def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %v", result)
	}
}
