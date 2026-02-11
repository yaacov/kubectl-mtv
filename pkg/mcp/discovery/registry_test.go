package discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// testdataPath returns the path to the help_machine_output.json fixture
// located in the testdata/ directory alongside this test file.
func testdataPath() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "testdata", "help_machine_output.json")
}

// loadRealRegistry loads a Registry from the real help_machine_output.json file.
// It skips the test if the file is not available.
func loadRealRegistry(t *testing.T) *Registry {
	t.Helper()

	data, err := os.ReadFile(testdataPath())
	if err != nil {
		t.Skipf("Skipping: could not read help_machine_output.json: %v", err)
	}

	var schema HelpSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("Failed to parse help_machine_output.json: %v", err)
	}

	registry := &Registry{
		ReadOnly:        make(map[string]*Command),
		ReadWrite:       make(map[string]*Command),
		GlobalFlags:     schema.GlobalFlags,
		RootDescription: schema.Description,
	}

	for i := range schema.Commands {
		cmd := &schema.Commands[i]
		pathKey := cmd.PathKey()
		switch cmd.Category {
		case "read":
			registry.ReadOnly[pathKey] = cmd
		default:
			registry.ReadWrite[pathKey] = cmd
		}
	}

	return registry
}

func TestCommand_PathKey(t *testing.T) {
	tests := []struct {
		name     string
		cmd      Command
		expected string
	}{
		{
			name: "single level path",
			cmd: Command{
				Path: []string{"get"},
			},
			expected: "get",
		},
		{
			name: "two level path",
			cmd: Command{
				Path: []string{"get", "plan"},
			},
			expected: "get/plan",
		},
		{
			name: "three level path",
			cmd: Command{
				Path: []string{"get", "inventory", "vm"},
			},
			expected: "get/inventory/vm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cmd.PathKey()
			if result != tt.expected {
				t.Errorf("PathKey() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCommand_CommandPath(t *testing.T) {
	tests := []struct {
		name     string
		cmd      Command
		expected string
	}{
		{
			name: "uses PathString if available",
			cmd: Command{
				Path:       []string{"get", "plan"},
				PathString: "get plan",
			},
			expected: "get plan",
		},
		{
			name: "falls back to joining Path",
			cmd: Command{
				Path: []string{"get", "inventory", "vm"},
			},
			expected: "get inventory vm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cmd.CommandPath()
			if result != tt.expected {
				t.Errorf("CommandPath() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCommand_PositionalArgsString(t *testing.T) {
	tests := []struct {
		name     string
		cmd      Command
		expected string
	}{
		{
			name:     "no positional args",
			cmd:      Command{},
			expected: "",
		},
		{
			name: "required positional arg",
			cmd: Command{
				PositionalArgs: []Arg{
					{Name: "NAME", Required: true},
				},
			},
			expected: "NAME",
		},
		{
			name: "optional positional arg",
			cmd: Command{
				PositionalArgs: []Arg{
					{Name: "NAME", Required: false},
				},
			},
			expected: "[NAME]",
		},
		{
			name: "mixed positional args",
			cmd: Command{
				PositionalArgs: []Arg{
					{Name: "PROVIDER", Required: true},
					{Name: "NAME", Required: false},
				},
			},
			expected: "PROVIDER [NAME]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cmd.PositionalArgsString()
			if result != tt.expected {
				t.Errorf("PositionalArgsString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRegistry_IsReadOnly(t *testing.T) {
	registry := &Registry{
		ReadOnly: map[string]*Command{
			"get/plan":    {Path: []string{"get", "plan"}},
			"describe/vm": {Path: []string{"describe", "vm"}},
		},
		ReadWrite: map[string]*Command{
			"create/plan": {Path: []string{"create", "plan"}},
			"delete/plan": {Path: []string{"delete", "plan"}},
		},
	}

	tests := []struct {
		pathKey  string
		expected bool
	}{
		{"get/plan", true},
		{"describe/vm", true},
		{"create/plan", false},
		{"delete/plan", false},
		{"unknown/cmd", false},
	}

	for _, tt := range tests {
		t.Run(tt.pathKey, func(t *testing.T) {
			result := registry.IsReadOnly(tt.pathKey)
			if result != tt.expected {
				t.Errorf("IsReadOnly(%q) = %v, want %v", tt.pathKey, result, tt.expected)
			}
		})
	}
}

func TestRegistry_IsReadWrite(t *testing.T) {
	registry := &Registry{
		ReadOnly: map[string]*Command{
			"get/plan": {Path: []string{"get", "plan"}},
		},
		ReadWrite: map[string]*Command{
			"create/plan": {Path: []string{"create", "plan"}},
			"delete/plan": {Path: []string{"delete", "plan"}},
		},
	}

	tests := []struct {
		pathKey  string
		expected bool
	}{
		{"get/plan", false},
		{"create/plan", true},
		{"delete/plan", true},
		{"unknown/cmd", false},
	}

	for _, tt := range tests {
		t.Run(tt.pathKey, func(t *testing.T) {
			result := registry.IsReadWrite(tt.pathKey)
			if result != tt.expected {
				t.Errorf("IsReadWrite(%q) = %v, want %v", tt.pathKey, result, tt.expected)
			}
		})
	}
}

func TestRegistry_GetCommand(t *testing.T) {
	readCmd := &Command{Path: []string{"get", "plan"}, Description: "Get plans"}
	writeCmd := &Command{Path: []string{"create", "plan"}, Description: "Create plan"}

	registry := &Registry{
		ReadOnly: map[string]*Command{
			"get/plan": readCmd,
		},
		ReadWrite: map[string]*Command{
			"create/plan": writeCmd,
		},
	}

	tests := []struct {
		pathKey  string
		expected *Command
	}{
		{"get/plan", readCmd},
		{"create/plan", writeCmd},
		{"unknown/cmd", nil},
	}

	for _, tt := range tests {
		t.Run(tt.pathKey, func(t *testing.T) {
			result := registry.GetCommand(tt.pathKey)
			if result != tt.expected {
				t.Errorf("GetCommand(%q) = %v, want %v", tt.pathKey, result, tt.expected)
			}
		})
	}
}

func TestRegistry_ListReadOnlyCommands(t *testing.T) {
	registry := &Registry{
		ReadOnly: map[string]*Command{
			"get/plan":         {Path: []string{"get", "plan"}},
			"describe/vm":      {Path: []string{"describe", "vm"}},
			"get/inventory/vm": {Path: []string{"get", "inventory", "vm"}},
		},
		ReadWrite: map[string]*Command{
			"create/plan": {Path: []string{"create", "plan"}},
		},
	}

	result := registry.ListReadOnlyCommands()

	// Should be sorted
	expected := []string{"describe/vm", "get/inventory/vm", "get/plan"}
	if len(result) != len(expected) {
		t.Fatalf("ListReadOnlyCommands() returned %d items, want %d", len(result), len(expected))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("ListReadOnlyCommands()[%d] = %q, want %q", i, result[i], v)
		}
	}
}

func TestRegistry_ListReadWriteCommands(t *testing.T) {
	registry := &Registry{
		ReadOnly: map[string]*Command{
			"get/plan": {Path: []string{"get", "plan"}},
		},
		ReadWrite: map[string]*Command{
			"create/plan": {Path: []string{"create", "plan"}},
			"delete/plan": {Path: []string{"delete", "plan"}},
			"start/plan":  {Path: []string{"start", "plan"}},
		},
	}

	result := registry.ListReadWriteCommands()

	// Should be sorted
	expected := []string{"create/plan", "delete/plan", "start/plan"}
	if len(result) != len(expected) {
		t.Fatalf("ListReadWriteCommands() returned %d items, want %d", len(result), len(expected))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("ListReadWriteCommands()[%d] = %q, want %q", i, result[i], v)
		}
	}
}

func TestBuildCommandArgs(t *testing.T) {
	tests := []struct {
		name           string
		cmdPath        string
		positionalArgs []string
		flags          map[string]string
		namespace      string
		allNamespaces  bool
		expected       []string
	}{
		{
			name:     "simple command",
			cmdPath:  "get/plan",
			expected: []string{"get", "plan"},
		},
		{
			name:           "with positional args",
			cmdPath:        "get/plan",
			positionalArgs: []string{"my-plan"},
			expected:       []string{"get", "plan", "my-plan"},
		},
		{
			name:      "with namespace",
			cmdPath:   "get/plan",
			namespace: "test-ns",
			expected:  []string{"get", "plan", "-n", "test-ns"},
		},
		{
			name:          "with all namespaces",
			cmdPath:       "get/plan",
			allNamespaces: true,
			expected:      []string{"get", "plan", "-A"},
		},
		{
			name:     "with string flag",
			cmdPath:  "get/plan",
			flags:    map[string]string{"output": "json"},
			expected: []string{"get", "plan", "--output", "json"},
		},
		{
			name:     "with boolean true flag",
			cmdPath:  "get/plan",
			flags:    map[string]string{"watch": "true"},
			expected: []string{"get", "plan", "--watch"},
		},
		{
			name:     "with boolean false flag - skipped",
			cmdPath:  "get/plan",
			flags:    map[string]string{"watch": "false"},
			expected: []string{"get", "plan"},
		},
		{
			name:      "namespace flag in map is ignored",
			cmdPath:   "get/plan",
			namespace: "test-ns",
			flags:     map[string]string{"namespace": "other-ns"},
			expected:  []string{"get", "plan", "-n", "test-ns"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildCommandArgs(tt.cmdPath, tt.positionalArgs, tt.flags, tt.namespace, tt.allNamespaces)

			if len(result) != len(tt.expected) {
				t.Fatalf("BuildCommandArgs() returned %v, want %v", result, tt.expected)
			}

			for i, v := range tt.expected {
				if result[i] != v {
					t.Errorf("BuildCommandArgs()[%d] = %q, want %q", i, result[i], v)
				}
			}
		})
	}
}

func TestFormatUsageShort(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *Command
		expected string
	}{
		{
			name: "command without args",
			cmd: &Command{
				PathString: "get plan",
			},
			expected: "get plan",
		},
		{
			name: "command with required arg",
			cmd: &Command{
				PathString: "create provider",
				PositionalArgs: []Arg{
					{Name: "NAME", Required: true},
				},
			},
			expected: "create provider NAME",
		},
		{
			name: "command with optional arg",
			cmd: &Command{
				PathString: "get plan",
				PositionalArgs: []Arg{
					{Name: "NAME", Required: false},
				},
			},
			expected: "get plan [NAME]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUsageShort(tt.cmd)
			if result != tt.expected {
				t.Errorf("formatUsageShort() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRegistry_GenerateReadOnlyDescription_Synthetic(t *testing.T) {
	registry := &Registry{
		ReadOnly: map[string]*Command{
			"get/plan": {
				PathString:  "get plan",
				Description: "Get migration plans",
				PositionalArgs: []Arg{
					{Name: "NAME", Required: false},
				},
			},
		},
		ReadWrite: map[string]*Command{},
	}

	result := registry.GenerateReadOnlyDescription()

	if !strings.Contains(result, "read-only") {
		t.Error("Description should mention read-only")
	}
	if !strings.Contains(result, "get plan [NAME]") {
		t.Error("Description should include command with args")
	}
	if !strings.Contains(result, "Get migration plans") {
		t.Error("Description should include command description")
	}
}

func TestRegistry_GenerateReadWriteDescription_Synthetic(t *testing.T) {
	registry := &Registry{
		ReadOnly: map[string]*Command{},
		ReadWrite: map[string]*Command{
			"create/plan": {
				PathString:  "create plan",
				Description: "Create a migration plan",
				PositionalArgs: []Arg{
					{Name: "NAME", Required: true},
				},
			},
		},
	}

	result := registry.GenerateReadWriteDescription()

	if !strings.Contains(result, "WARNING") {
		t.Error("Description should include WARNING")
	}
	if !strings.Contains(result, "create plan NAME") {
		t.Error("Description should include command with args")
	}
	if !strings.Contains(result, "Create a migration plan") {
		t.Error("Description should include command description")
	}
}

// --- Tests using real help --machine output ---

func TestRegistry_RealHelpMachine_ReadOnlyDescription(t *testing.T) {
	registry := loadRealRegistry(t)

	result := registry.GenerateReadOnlyDescription()

	// Should contain key read-only commands
	for _, expected := range []string{
		"health",
		"get plan",
		"get provider",
		"get inventory vm",
		"describe plan",
		"settings get",
	} {
		if !strings.Contains(result, expected) {
			t.Errorf("Read-only description should contain %q", expected)
		}
	}

	// Should NOT contain write commands
	for _, notExpected := range []string{
		"create provider",
		"delete plan",
		"start plan",
	} {
		if strings.Contains(result, notExpected) {
			t.Errorf("Read-only description should NOT contain write command %q", notExpected)
		}
	}
}

func TestRegistry_RealHelpMachine_ReadWriteDescription(t *testing.T) {
	registry := loadRealRegistry(t)

	result := registry.GenerateReadWriteDescription()

	// Should contain key write commands
	for _, expected := range []string{
		"create provider",
		"create plan",
		"delete plan",
		"start plan",
		"patch provider",
		"cancel plan",
	} {
		if !strings.Contains(result, expected) {
			t.Errorf("Read-write description should contain %q", expected)
		}
	}

	// Should contain env var documentation with embedded support
	if !strings.Contains(result, "${ENV_VAR_NAME}") {
		t.Error("Description should document env var syntax")
	}
	if !strings.Contains(result, "${GOVC_URL}/sdk") {
		t.Error("Description should document embedded env var references")
	}
}

func TestRegistry_RealHelpMachine_FlagReference(t *testing.T) {
	registry := loadRealRegistry(t)

	result := registry.GenerateReadWriteDescription()

	// Should contain the flag reference section
	if !strings.Contains(result, "Flag reference for complex commands:") {
		t.Fatal("Description should contain flag reference section")
	}

	// Should surface required flags that previously caused 100% failure
	requiredFlags := []struct {
		command string
		flag    string
	}{
		{"create provider", "--type"},
		{"create provider", "[REQUIRED]"},
		{"create provider", "[enum:"},
		{"cancel plan", "--vms"},
		{"cancel plan", "[REQUIRED]"},
		{"create host", "--provider"},
		{"create host", "[REQUIRED]"},
		{"create vddk-image", "--tag"},
		{"create vddk-image", "--tar"},
	}

	for _, rf := range requiredFlags {
		if !strings.Contains(result, rf.flag) {
			t.Errorf("Flag reference should contain %q (for %s)", rf.flag, rf.command)
		}
	}

	// Should surface key flags that the LLM needs for common operations
	keyFlags := []string{
		"--provider-insecure-skip-tls",
		"--url",
		"--username",
		"--password",
		"--sdk-endpoint",
		"--vddk-init-image",
		"--source",
		"--target",
		"--migration-type",
	}

	for _, flag := range keyFlags {
		if !strings.Contains(result, flag) {
			t.Errorf("Flag reference should contain key flag %q", flag)
		}
	}

	// Should surface enum values for constrained flags
	enumValues := []string{
		"openshift",
		"vsphere",
		"ovirt",
		"openstack",
		"cold",
		"warm",
	}

	for _, val := range enumValues {
		if !strings.Contains(result, val) {
			t.Errorf("Flag reference should contain enum value %q", val)
		}
	}
}

func TestRegistry_RealHelpMachine_CommandCounts(t *testing.T) {
	registry := loadRealRegistry(t)

	readCount := len(registry.ReadOnly)
	writeCount := len(registry.ReadWrite)

	// Sanity check: the real data should have a reasonable number of commands
	if readCount < 30 {
		t.Errorf("Expected at least 30 read-only commands, got %d", readCount)
	}
	if writeCount < 20 {
		t.Errorf("Expected at least 20 read-write commands, got %d", writeCount)
	}

	// Verify key commands exist
	keyReadCommands := []string{"health", "get/plan", "get/provider", "get/inventory/vm", "describe/plan"}
	for _, cmd := range keyReadCommands {
		if registry.ReadOnly[cmd] == nil {
			t.Errorf("Expected read-only command %q to exist", cmd)
		}
	}

	keyWriteCommands := []string{"create/provider", "create/plan", "delete/plan", "start/plan", "patch/provider", "cancel/plan"}
	for _, cmd := range keyWriteCommands {
		if registry.ReadWrite[cmd] == nil {
			t.Errorf("Expected read-write command %q to exist", cmd)
		}
	}
}

func TestRegistry_RealHelpMachine_KARLReference(t *testing.T) {
	registry := loadRealRegistry(t)

	result := registry.GenerateReadWriteDescription()

	// KARL syntax reference should be surfaced via LongDescription of create plan and patch plan.
	// Detailed KARL syntax is available via 'help karl'; command descriptions contain a summary.
	karlKeywords := []string{
		"Affinity Syntax (KARL)",
		"REQUIRE",
		"PREFER",
		"AVOID",
		"REPEL",
		"pods(",
		"weight=",
		"Topology:",
		"help karl",
	}

	for _, keyword := range karlKeywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("Read-write description should contain KARL keyword %q", keyword)
		}
	}
}

func TestRegistry_RealHelpMachine_QueryLanguageReference(t *testing.T) {
	registry := loadRealRegistry(t)

	// Query language reference should appear in the write description (via create plan LongDescription).
	// Detailed TSL syntax and field lists are available via 'help tsl'; command descriptions contain a summary.
	writeResult := registry.GenerateReadWriteDescription()
	tslKeywords := []string{
		"Query Language (TSL)",
		"where",
		"~=",
		"cpuCount",
		"memoryMB",
		"powerState",
		"len(disks)",
		"help tsl",
	}

	for _, keyword := range tslKeywords {
		if !strings.Contains(writeResult, keyword) {
			t.Errorf("Read-write description should contain TSL keyword %q", keyword)
		}
	}

	// Query language reference should also appear in the read-only description
	// (via get inventory vm LongDescription)
	readResult := registry.GenerateReadOnlyDescription()
	readTSLKeywords := []string{
		"Query Language (TSL)",
		"where",
		"like",
		"~=",
		"ORDER BY",
		"-o json",
		"cpuCount",
		"help tsl",
	}

	for _, keyword := range readTSLKeywords {
		if !strings.Contains(readResult, keyword) {
			t.Errorf("Read-only description should contain TSL keyword %q", keyword)
		}
	}
}

func TestRegistry_RealHelpMachine_CreateProviderFlags(t *testing.T) {
	registry := loadRealRegistry(t)

	cmd := registry.ReadWrite["create/provider"]
	if cmd == nil {
		t.Fatal("create/provider command not found in registry")
	}

	// Should have a reasonable number of flags
	if len(cmd.Flags) < 10 {
		t.Errorf("create provider should have at least 10 flags, got %d", len(cmd.Flags))
	}

	// --type should be required with enum values
	var typeFlag *Flag
	for i := range cmd.Flags {
		if cmd.Flags[i].Name == "type" {
			typeFlag = &cmd.Flags[i]
			break
		}
	}

	if typeFlag == nil {
		t.Fatal("create provider should have a --type flag")
	}
	if !typeFlag.Required {
		t.Error("--type flag should be marked as required")
	}
	if len(typeFlag.Enum) == 0 {
		t.Error("--type flag should have enum values")
	}

	// --provider-insecure-skip-tls should exist
	found := false
	for _, f := range cmd.Flags {
		if f.Name == "provider-insecure-skip-tls" {
			found = true
			break
		}
	}
	if !found {
		t.Error("create provider should have --provider-insecure-skip-tls flag")
	}
}
