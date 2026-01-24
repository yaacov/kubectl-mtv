package discovery

import (
	"testing"
)

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

func TestRegistry_GenerateReadOnlyDescription(t *testing.T) {
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

	// Check that it contains expected content
	if !contains(result, "read-only") {
		t.Error("Description should mention read-only")
	}
	if !contains(result, "get plan [NAME]") {
		t.Error("Description should include command with args")
	}
	if !contains(result, "Get migration plans") {
		t.Error("Description should include command description")
	}
}

func TestRegistry_GenerateReadWriteDescription(t *testing.T) {
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

	// Check that it contains expected content
	if !contains(result, "WARNING") {
		t.Error("Description should include WARNING")
	}
	if !contains(result, "create plan NAME") {
		t.Error("Description should include command with args")
	}
	if !contains(result, "Create a migration plan") {
		t.Error("Description should include command description")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
