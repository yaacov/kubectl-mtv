package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/pkg/mcp/util"
)

// --- CoerceBooleans tests ---

// testInput is a struct used for testing CoerceBooleans.
type testInput struct {
	Command       string         `json:"command" jsonschema:"test command"`
	AllNamespaces bool           `json:"all_namespaces,omitempty" jsonschema:"Query all namespaces"`
	DryRun        bool           `json:"dry_run,omitempty" jsonschema:"Dry run mode"`
	Previous      bool           `json:"previous,omitempty" jsonschema:"Previous container"`
	Name          string         `json:"name,omitempty" jsonschema:"Resource name"`
	Flags         map[string]any `json:"flags,omitempty" jsonschema:"Additional flags"`
}

func TestCoerceBooleans_ProperBooleans(t *testing.T) {
	// Proper JSON booleans should pass through unchanged
	data := json.RawMessage(`{"command":"get plan","all_namespaces":true,"dry_run":false}`)
	result := CoerceBooleans[testInput](data)

	var m map[string]any
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}
	if m["all_namespaces"] != true {
		t.Errorf("all_namespaces = %v, want true", m["all_namespaces"])
	}
	if m["dry_run"] != false {
		t.Errorf("dry_run = %v, want false", m["dry_run"])
	}
}

func TestCoerceBooleans_StringTrue(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"lowercase true", `"true"`},
		{"capitalized True", `"True"`},
		{"uppercase TRUE", `"TRUE"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := json.RawMessage(`{"command":"get plan","all_namespaces":` + tt.value + `}`)
			result := CoerceBooleans[testInput](data)

			var m map[string]any
			if err := json.Unmarshal(result, &m); err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}
			if m["all_namespaces"] != true {
				t.Errorf("all_namespaces = %v (%T), want true (bool)", m["all_namespaces"], m["all_namespaces"])
			}
		})
	}
}

func TestCoerceBooleans_StringFalse(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"lowercase false", `"false"`},
		{"capitalized False", `"False"`},
		{"uppercase FALSE", `"FALSE"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := json.RawMessage(`{"command":"get plan","dry_run":` + tt.value + `}`)
			result := CoerceBooleans[testInput](data)

			var m map[string]any
			if err := json.Unmarshal(result, &m); err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}
			if m["dry_run"] != false {
				t.Errorf("dry_run = %v (%T), want false (bool)", m["dry_run"], m["dry_run"])
			}
		})
	}
}

func TestCoerceBooleans_MixedCorrectAndString(t *testing.T) {
	// Mix of proper booleans and string booleans
	data := json.RawMessage(`{"command":"get plan","all_namespaces":true,"dry_run":"True","previous":"false"}`)
	result := CoerceBooleans[testInput](data)

	var m map[string]any
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}
	if m["all_namespaces"] != true {
		t.Errorf("all_namespaces = %v, want true", m["all_namespaces"])
	}
	if m["dry_run"] != true {
		t.Errorf("dry_run = %v, want true (coerced from \"True\")", m["dry_run"])
	}
	if m["previous"] != false {
		t.Errorf("previous = %v, want false (coerced from \"false\")", m["previous"])
	}
}

func TestCoerceBooleans_NoBooleanFields(t *testing.T) {
	// Struct with no boolean fields should return data unchanged
	type noBools struct {
		Name string `json:"name"`
	}
	data := json.RawMessage(`{"name":"test"}`)
	result := CoerceBooleans[noBools](data)

	if string(result) != string(data) {
		t.Errorf("CoerceBooleans changed data when no bool fields exist: got %s, want %s", result, data)
	}
}

func TestCoerceBooleans_NonBoolStringNotCoerced(t *testing.T) {
	// String values that aren't "true"/"false" should not be coerced
	data := json.RawMessage(`{"command":"get plan","all_namespaces":"yes","name":"test"}`)
	result := CoerceBooleans[testInput](data)

	var m map[string]any
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}
	// "yes" is not a recognized boolean string, so it should remain a string
	if _, ok := m["all_namespaces"].(string); !ok {
		t.Errorf("all_namespaces should remain a string for unrecognized value 'yes', got %T", m["all_namespaces"])
	}
}

func TestCoerceBooleans_EmptyData(t *testing.T) {
	result := CoerceBooleans[testInput](nil)
	if result != nil {
		t.Errorf("Expected nil for nil input, got %s", result)
	}

	result = CoerceBooleans[testInput](json.RawMessage{})
	if len(result) != 0 {
		t.Errorf("Expected empty for empty input, got %s", result)
	}
}

func TestCoerceBooleans_MalformedJSON(t *testing.T) {
	data := json.RawMessage(`{invalid json}`)
	result := CoerceBooleans[testInput](data)

	// Malformed JSON should be returned unchanged
	if string(result) != string(data) {
		t.Errorf("CoerceBooleans changed malformed JSON: got %s, want %s", result, data)
	}
}

func TestCoerceBooleans_NonStructField(t *testing.T) {
	// String fields should not be affected even if their values look like booleans
	data := json.RawMessage(`{"command":"true","name":"false"}`)
	result := CoerceBooleans[testInput](data)

	var m map[string]any
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}
	if m["command"] != "true" {
		t.Errorf("command = %v, should remain string \"true\"", m["command"])
	}
	if m["name"] != "false" {
		t.Errorf("name = %v, should remain string \"false\"", m["name"])
	}
}

func TestCoerceBooleans_UnmarshalAfterCoerce(t *testing.T) {
	// Verify that coerced data can be successfully unmarshaled into the typed struct
	data := json.RawMessage(`{"command":"get plan","all_namespaces":"True","dry_run":"false"}`)
	result := CoerceBooleans[testInput](data)

	var input testInput
	if err := json.Unmarshal(result, &input); err != nil {
		t.Fatalf("Failed to unmarshal coerced data into struct: %v", err)
	}
	if input.Command != "get plan" {
		t.Errorf("Command = %q, want %q", input.Command, "get plan")
	}
	if !input.AllNamespaces {
		t.Error("AllNamespaces should be true after coercion from \"True\"")
	}
	if input.DryRun {
		t.Error("DryRun should be false after coercion from \"false\"")
	}
}

// --- CoerceBooleans with real tool input types ---

func TestCoerceBooleans_MTVReadInput(t *testing.T) {
	data := json.RawMessage(`{"command":"get plan","all_namespaces":"True","dry_run":"FALSE"}`)
	result := CoerceBooleans[MTVReadInput](data)

	var input MTVReadInput
	if err := json.Unmarshal(result, &input); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if !input.AllNamespaces {
		t.Error("AllNamespaces should be true")
	}
	if input.DryRun {
		t.Error("DryRun should be false")
	}
}

func TestCoerceBooleans_MTVWriteInput(t *testing.T) {
	data := json.RawMessage(`{"command":"create plan","dry_run":"True"}`)
	result := CoerceBooleans[MTVWriteInput](data)

	var input MTVWriteInput
	if err := json.Unmarshal(result, &input); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if !input.DryRun {
		t.Error("DryRun should be true")
	}
}

func TestCoerceBooleans_KubectlDebugInput(t *testing.T) {
	data := json.RawMessage(`{"action":"logs","all_namespaces":"True","previous":"true","dry_run":"False","ignore_case":"TRUE","no_timestamps":"false"}`)
	result := CoerceBooleans[KubectlDebugInput](data)

	var input KubectlDebugInput
	if err := json.Unmarshal(result, &input); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if !input.AllNamespaces {
		t.Error("AllNamespaces should be true")
	}
	if !input.Previous {
		t.Error("Previous should be true")
	}
	if input.DryRun {
		t.Error("DryRun should be false")
	}
	if !input.IgnoreCase {
		t.Error("IgnoreCase should be true")
	}
	if input.NoTimestamps {
		t.Error("NoTimestamps should be false")
	}
}

// --- parseBoolValue tests ---

func TestParseBoolValue(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"string true lowercase", "true", true},
		{"string True capitalized", "True", true},
		{"string TRUE uppercase", "TRUE", true},
		{"string false lowercase", "false", false},
		{"string False capitalized", "False", false},
		{"string FALSE uppercase", "FALSE", false},
		{"string 1", "1", true},
		{"string 0", "0", false},
		{"string other", "yes", false},
		{"string empty", "", false},
		{"float64 1", float64(1), true},
		{"float64 0", float64(0), false},
		{"float64 nonzero", float64(42), true},
		{"nil", nil, false},
		{"int (unexpected type)", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBoolValue(tt.value)
			if result != tt.expected {
				t.Errorf("parseBoolValue(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// --- AddToolWithCoercion tests ---

// testToolInput is a simple input struct for testing AddToolWithCoercion.
type testToolInput struct {
	Name    string `json:"name" jsonschema:"Resource name"`
	Verbose bool   `json:"verbose,omitempty" jsonschema:"Enable verbose output"`
}

func TestAddToolWithCoercion_SchemaSet(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "1.0",
	}, nil)

	handler := func(_ context.Context, _ *mcp.CallToolRequest, _ testToolInput) (*mcp.CallToolResult, any, error) {
		return nil, map[string]any{"status": "ok"}, nil
	}

	tool := &mcp.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		OutputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}

	// Register with coercion wrapper
	AddToolWithCoercion(server, tool, handler)

	// Verify that the tool's InputSchema was set
	if tool.InputSchema == nil {
		t.Fatal("Tool InputSchema should be set after AddToolWithCoercion")
	}
}

func TestAddToolWithCoercion_EndToEnd(t *testing.T) {
	// Verify the full coercion path: string booleans are coerced before
	// reaching the handler, so the handler receives proper bool values.
	// This tests the same CoerceBooleans + unmarshal + handler call that
	// AddToolWithCoercion's raw handler performs internally.
	var receivedVerbose bool
	var receivedName string
	handler := func(_ context.Context, _ *mcp.CallToolRequest, input testToolInput) (*mcp.CallToolResult, any, error) {
		receivedVerbose = input.Verbose
		receivedName = input.Name
		return nil, map[string]any{"name": input.Name, "verbose": input.Verbose}, nil
	}

	// Simulate a tool call with string boolean "True" (what a broken model sends)
	rawArgs := json.RawMessage(`{"name":"test","verbose":"True"}`)

	// This is the same code path that AddToolWithCoercion's wrapper executes:
	// 1. Coerce string booleans
	coerced := CoerceBooleans[testToolInput](rawArgs)

	// 2. Unmarshal into typed input
	var input testToolInput
	if err := json.Unmarshal(coerced, &input); err != nil {
		t.Fatalf("Failed to unmarshal coerced data: %v", err)
	}

	// 3. Call handler with the coerced input
	ctx := context.Background()
	_, _, err := handler(ctx, nil, input)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if !receivedVerbose {
		t.Error("Handler should have received Verbose=true after coercion from \"True\"")
	}
	if receivedName != "test" {
		t.Errorf("Handler received Name=%q, want %q", receivedName, "test")
	}
}

// --- buildArgs flags fallback tests ---

func TestBuildArgs_FlagsAllNamespaces(t *testing.T) {
	tests := []struct {
		name          string
		flags         map[string]any
		allNamespaces bool
		wantA         bool // whether -A should be in the result
	}{
		{
			name:          "top-level true takes priority",
			flags:         nil,
			allNamespaces: true,
			wantA:         true,
		},
		{
			name:          "flags all_namespaces bool true",
			flags:         map[string]any{"all_namespaces": true},
			allNamespaces: false,
			wantA:         true,
		},
		{
			name:          "flags all_namespaces string True",
			flags:         map[string]any{"all_namespaces": "True"},
			allNamespaces: false,
			wantA:         true,
		},
		{
			name:          "flags all_namespaces string true",
			flags:         map[string]any{"all_namespaces": "true"},
			allNamespaces: false,
			wantA:         true,
		},
		{
			name:          "flags A bool true",
			flags:         map[string]any{"A": true},
			allNamespaces: false,
			wantA:         true,
		},
		{
			name:          "flags A string true",
			flags:         map[string]any{"A": "true"},
			allNamespaces: false,
			wantA:         true,
		},
		{
			name:          "flags all_namespaces false",
			flags:         map[string]any{"all_namespaces": false},
			allNamespaces: false,
			wantA:         false,
		},
		{
			name:          "flags all_namespaces string false",
			flags:         map[string]any{"all_namespaces": "false"},
			allNamespaces: false,
			wantA:         false,
		},
		{
			name:          "no flags no all_namespaces",
			flags:         nil,
			allNamespaces: false,
			wantA:         false,
		},
		{
			name:          "top-level true overrides flags false",
			flags:         map[string]any{"all_namespaces": false},
			allNamespaces: true,
			wantA:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use text format to avoid -o json in the result (simplifies assertions)
			origFormat := util.GetOutputFormat()
			util.SetOutputFormat("text")
			defer util.SetOutputFormat(origFormat)

			result := buildArgs("get/plan", nil, tt.flags, "", tt.allNamespaces, "")

			// Use exact element match to avoid false positives from substrings
			hasA := false
			for _, arg := range result {
				if arg == "-A" {
					hasA = true
					break
				}
			}
			if hasA != tt.wantA {
				t.Errorf("buildArgs() = %v, contains -A = %v, want %v", result, hasA, tt.wantA)
			}
		})
	}
}
