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
		Parents:         make(map[string]*Command),
		GlobalFlags:     schema.GlobalFlags,
		RootDescription: schema.Description,
		LongDescription: schema.LongDescription,
	}

	for i := range schema.Commands {
		cmd := &schema.Commands[i]
		pathKey := cmd.PathKey()

		// Store non-runnable parents separately, matching NewRegistry behavior
		if !cmd.Runnable {
			registry.Parents[pathKey] = cmd
			continue
		}

		switch cmd.Category {
		case "read":
			registry.ReadOnly[pathKey] = cmd
			registry.ReadOnlyOrder = append(registry.ReadOnlyOrder, pathKey)
		case "admin":
			// Skip admin commands, matching NewRegistry behavior
			continue
		default:
			registry.ReadWrite[pathKey] = cmd
			registry.ReadWriteOrder = append(registry.ReadWriteOrder, pathKey)
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
		name          string
		cmdPath       string
		flags         map[string]string
		namespace     string
		allNamespaces bool
		expected      []string
	}{
		{
			name:     "simple command",
			cmdPath:  "get/plan",
			expected: []string{"get", "plan"},
		},
		{
			name:     "with flag",
			cmdPath:  "get/plan",
			flags:    map[string]string{"name": "my-plan"},
			expected: []string{"get", "plan", "--name", "my-plan"},
		},
		{
			name:      "with namespace",
			cmdPath:   "get/plan",
			namespace: "test-ns",
			expected:  []string{"get", "plan", "--namespace", "test-ns"},
		},
		{
			name:          "with all namespaces",
			cmdPath:       "get/plan",
			allNamespaces: true,
			expected:      []string{"get", "plan", "--all-namespaces"},
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
			expected:  []string{"get", "plan", "--namespace", "test-ns"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildCommandArgs(tt.cmdPath, tt.flags, tt.namespace, tt.allNamespaces)

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
			name: "command path",
			cmd: &Command{
				PathString: "get plan",
			},
			expected: "get plan",
		},
		{
			name: "multi-segment path",
			cmd: &Command{
				PathString: "create provider",
			},
			expected: "create provider",
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
			},
		},
		ReadWrite: map[string]*Command{},
		GlobalFlags: []Flag{
			{Name: "namespace", Description: "Target namespace"},
		},
	}

	result := registry.GenerateReadOnlyDescription()

	if !strings.Contains(result, "read-only") {
		t.Error("Description should mention read-only")
	}
	if !strings.Contains(result, "get plan") {
		t.Error("Description should include command")
	}
	if !strings.Contains(result, "Get migration plans") {
		t.Error("Description should include command description")
	}
	// Root verbs should be derived from the command map
	if !strings.Contains(result, "Commands: get") {
		t.Error("Description should contain data-derived root verbs")
	}
	// Global flags should be derived from GlobalFlags data (important flags)
	if !strings.Contains(result, "namespace: Target namespace") {
		t.Error("Description should contain global flag descriptions from data")
	}
}

func TestRegistry_GenerateReadWriteDescription_Synthetic(t *testing.T) {
	registry := &Registry{
		ReadOnly: map[string]*Command{
			"get/plan": {PathString: "get plan", Description: "Get plans"},
		},
		ReadWrite: map[string]*Command{
			"create/plan": {
				PathString:  "create plan",
				Description: "Create a migration plan",
			},
		},
		GlobalFlags: []Flag{
			{Name: "namespace", Description: "Target namespace"},
		},
	}

	result := registry.GenerateReadWriteDescription()

	if !strings.Contains(result, "WARNING") {
		t.Error("Description should include WARNING")
	}
	if !strings.Contains(result, "create plan") {
		t.Error("Description should include command")
	}
	if !strings.Contains(result, "Create a migration plan") {
		t.Error("Description should include command description")
	}
	// Should derive read-only root verbs for the NOTE line
	if !strings.Contains(result, "For read-only operations (get)") {
		t.Errorf("Description should derive read-only root verbs: %s", result)
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

	// Sanity check: the real data should have a reasonable number of commands.
	// Admin commands are filtered out by NewRegistry, so write count is lower.
	if readCount < 30 {
		t.Errorf("Expected at least 30 read-only commands, got %d", readCount)
	}
	if writeCount < 15 {
		t.Errorf("Expected at least 15 write commands (admin excluded), got %d", writeCount)
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
		"--output json",
		"cpuCount",
		"help tsl",
	}

	for _, keyword := range readTSLKeywords {
		if !strings.Contains(readResult, keyword) {
			t.Errorf("Read-only description should contain TSL keyword %q", keyword)
		}
	}
}

// --- Tests for generic helper functions ---

func TestUniqueRootVerbs(t *testing.T) {
	commands := map[string]*Command{
		"get/plan":         {},
		"get/provider":     {},
		"get/inventory/vm": {},
		"describe/plan":    {},
		"health":           {},
		"settings/get":     {},
	}

	roots := uniqueRootVerbs(commands)
	expected := []string{"describe", "get", "health", "settings"}

	if len(roots) != len(expected) {
		t.Fatalf("uniqueRootVerbs() = %v, want %v", roots, expected)
	}
	for i, v := range expected {
		if roots[i] != v {
			t.Errorf("uniqueRootVerbs()[%d] = %q, want %q", i, roots[i], v)
		}
	}
}

func TestDetectBareParents(t *testing.T) {
	commands := map[string]*Command{
		// "patch" is a bare parent: no flags, prefix of patch/plan
		"patch": {Path: []string{"patch"}},
		"patch/plan": {
			Path:  []string{"patch", "plan"},
			Flags: []Flag{{Name: "query", Type: "string"}},
		},
		"patch/mapping": {
			// Also a bare parent: prefix of patch/mapping/network, no flags
			Path: []string{"patch", "mapping"},
		},
		"patch/mapping/network": {
			Path: []string{"patch", "mapping", "network"},
		},
		// "delete/mapping" has flags, so NOT a bare parent even though it's a prefix
		"delete/mapping": {
			Path:  []string{"delete", "mapping"},
			Flags: []Flag{{Name: "all", Type: "bool"}},
		},
		"delete/mapping/network": {
			Path: []string{"delete", "mapping", "network"},
		},
		// "create/plan" is NOT a prefix of anything, so not bare
		"create/plan": {
			Path: []string{"create", "plan"},
		},
	}

	bareParents := detectBareParents(commands)

	if !bareParents["patch"] {
		t.Error("Expected 'patch' to be detected as bare parent")
	}
	if !bareParents["patch/mapping"] {
		t.Error("Expected 'patch/mapping' to be detected as bare parent")
	}
	if bareParents["delete/mapping"] {
		t.Error("'delete/mapping' has flags, should NOT be detected as bare parent")
	}
	if bareParents["create/plan"] {
		t.Error("'create/plan' is not a prefix of anything, should NOT be bare parent")
	}
	if bareParents["patch/plan"] {
		t.Error("'patch/plan' has flags, should NOT be bare parent")
	}
}

func TestDetectDeepSiblingGroups(t *testing.T) {
	commands := map[string]*Command{
		"get/inventory/vm": {
			Name:        "vm",
			Path:        []string{"get", "inventory", "vm"},
			Description: "Get VMs from provider",
		},
		"get/inventory/network": {
			Name:        "network",
			Path:        []string{"get", "inventory", "network"},
			Description: "Get networks from provider",
		},
		"get/inventory/cluster": {
			Name:        "cluster",
			Path:        []string{"get", "inventory", "cluster"},
			Description: "Get clusters from provider",
		},
		"get/plan": {
			Name:        "plan",
			Path:        []string{"get", "plan"},
			Description: "Get plans",
		},
	}

	groups, groupedKeys := detectDeepSiblingGroups(commands, nil)

	if len(groups) != 1 {
		t.Fatalf("Expected 1 sibling group, got %d", len(groups))
	}

	group := groups[0]
	if group.parentPath != "get/inventory" {
		t.Errorf("Expected parent path 'get/inventory', got %q", group.parentPath)
	}
	if len(group.children) != 3 {
		t.Errorf("Expected 3 children, got %d: %v", len(group.children), group.children)
	}

	for _, key := range []string{"get/inventory/vm", "get/inventory/network", "get/inventory/cluster"} {
		if !groupedKeys[key] {
			t.Errorf("Expected %q to be in grouped keys", key)
		}
	}
	if groupedKeys["get/plan"] {
		t.Error("get/plan should not be in a sibling group")
	}
}

func TestDetectDeepSiblingGroups_BelowThreshold(t *testing.T) {
	// Only 2 siblings — below the threshold of 3
	commands := map[string]*Command{
		"get/inventory/vm":      {Name: "vm", Path: []string{"get", "inventory", "vm"}, Description: "VMs"},
		"get/inventory/network": {Name: "network", Path: []string{"get", "inventory", "network"}, Description: "Networks"},
	}

	groups, groupedKeys := detectDeepSiblingGroups(commands, nil)
	if len(groups) != 0 {
		t.Errorf("Expected 0 groups below threshold, got %d", len(groups))
	}
	if len(groupedKeys) != 0 {
		t.Errorf("Expected no grouped keys below threshold, got %d", len(groupedKeys))
	}
}

func TestDetectDeepSiblingGroups_TopLevelNotCompacted(t *testing.T) {
	// Top-level groups (depth 2) should NOT be compacted — they are heterogeneous
	commands := map[string]*Command{
		"get/plan":     {Name: "plan", Path: []string{"get", "plan"}, Description: "Get plans"},
		"get/provider": {Name: "provider", Path: []string{"get", "provider"}, Description: "Get providers"},
		"get/hook":     {Name: "hook", Path: []string{"get", "hook"}, Description: "Get hooks"},
		"get/host":     {Name: "host", Path: []string{"get", "host"}, Description: "Get hosts"},
	}

	groups, groupedKeys := detectDeepSiblingGroups(commands, nil)
	if len(groups) != 0 {
		t.Errorf("Top-level groups (depth 2) should not be compacted, got %d groups", len(groups))
	}
	if len(groupedKeys) != 0 {
		t.Errorf("No keys should be grouped for top-level commands, got %d", len(groupedKeys))
	}
}

func TestFormatGlobalFlags(t *testing.T) {
	registry := &Registry{
		GlobalFlags: []Flag{
			{Name: "namespace", Description: "Target Kubernetes namespace"},
			{Name: "all-namespaces", Description: "Query all namespaces"},
			{Name: "output", Description: "Output format"},
			{Name: "kubeconfig", Description: "Path to kubeconfig"},
			{Name: "verbose", Description: "Enable verbose output"},
		},
	}

	result := registry.formatGlobalFlags()

	// Should include the important global flags
	if !strings.Contains(result, "namespace: Target Kubernetes namespace") {
		t.Error("Should include namespace flag from data")
	}
	if !strings.Contains(result, "all_namespaces: Query all namespaces") {
		t.Error("Should include all_namespaces flag from data (hyphens converted to underscores for MCP JSON)")
	}
	if !strings.Contains(result, "verbose: Enable verbose output") {
		t.Error("Should include verbose flag from data")
	}

	// Should NOT include kubeconfig or output (not in importantGlobalFlags)
	if strings.Contains(result, "kubeconfig") {
		t.Error("Should NOT include kubeconfig flag (not LLM-relevant)")
	}
}

func TestFormatGlobalFlags_Empty(t *testing.T) {
	registry := &Registry{
		GlobalFlags: []Flag{
			{Name: "kubeconfig", Description: "Path to kubeconfig"},
		},
	}

	// No relevant flags → should return empty string
	result := registry.formatGlobalFlags()
	if result != "" {
		t.Errorf("Expected empty string when no relevant flags, got %q", result)
	}
}

func TestRegistry_RealHelpMachine_NoAdminCommands(t *testing.T) {
	registry := loadRealRegistry(t)

	// Admin commands should NOT be present in either map
	adminCommands := []string{
		"completion/bash", "completion/fish", "completion/powershell", "completion/zsh",
		"help", "mcp-server", "version",
	}
	for _, cmd := range adminCommands {
		if registry.ReadOnly[cmd] != nil {
			t.Errorf("Admin command %q should not be in ReadOnly", cmd)
		}
		if registry.ReadWrite[cmd] != nil {
			t.Errorf("Admin command %q should not be in ReadWrite", cmd)
		}
	}
}

func TestRegistry_RealHelpMachine_MinimalReadOnlyGroupsInventory(t *testing.T) {
	registry := loadRealRegistry(t)

	result := registry.GenerateMinimalReadOnlyDescription()

	// Should have a compacted inventory line with "RESOURCE"
	if !strings.Contains(result, "get inventory RESOURCE") {
		t.Errorf("Minimal read-only description should compact inventory commands into 'get inventory RESOURCE', got:\n%s", result)
	}
	// Should list resource names
	if !strings.Contains(result, "vm") {
		t.Error("Compacted inventory should list 'vm' as a resource")
	}
	if !strings.Contains(result, "network") {
		t.Error("Compacted inventory should list 'network' as a resource")
	}
	// Should have examples section
	if !strings.Contains(result, "Examples:") {
		t.Error("Description should include examples section")
	}
	if !strings.Contains(result, "get inventory") {
		t.Error("Examples should include an inventory command")
	}
	if !strings.Contains(result, "query") {
		t.Error("Inventory example should include a query flag to demonstrate TSL filtering")
	}
	if !strings.Contains(result, "TSL") {
		t.Error("Description should include TSL syntax hint")
	}
	if !strings.Contains(result, "mtv_help") {
		t.Error("Description should reference mtv_help for detailed flags")
	}
	// Should NOT list individual inventory commands separately (they should be compacted)
	if strings.Contains(result, "  get inventory vm -") {
		t.Error("Individual inventory commands should be compacted, not listed separately")
	}
	// Should list non-grouped commands like get plan
	if !strings.Contains(result, "get plan") {
		t.Error("Description should list non-grouped commands like 'get plan'")
	}
	// Should include short MTV context preamble
	if !strings.Contains(result, "Migration Toolkit for Virtualization") {
		t.Error("Minimal description should include MTV context preamble")
	}
	// Should NOT include orphaned convention notes (removed)
	if strings.Contains(result, "Args: <required>, [optional]") {
		t.Error("Minimal description should NOT include the orphaned args convention note")
	}
}

func TestRegistry_RealHelpMachine_MinimalReadWriteNoBareParents(t *testing.T) {
	registry := loadRealRegistry(t)

	result := registry.GenerateMinimalReadWriteDescription()

	// Real write commands should be present
	if !strings.Contains(result, "create plan") {
		t.Error("Minimal write description should contain 'create plan'")
	}
	if !strings.Contains(result, "patch plan") {
		t.Error("Minimal write description should contain 'patch plan'")
	}
	// Should have examples derived from source
	if !strings.Contains(result, "Examples:") {
		t.Error("Minimal write description should contain examples section")
	}
	// Should include KARL hint
	if !strings.Contains(result, "KARL") {
		t.Error("Minimal write description should include KARL hint")
	}
	// Should include mtv_help reference
	if !strings.Contains(result, "mtv_help") {
		t.Error("Minimal write description should reference mtv_help")
	}
	// Should NOT include LongDescription (removed to save space)
	if strings.Contains(result, "Migration Toolkit for Virtualization") {
		t.Error("Minimal description should NOT include LongDescription domain context")
	}

	// Bare parent "patch" (if it exists as a structural node) should be excluded.
	// We check by ensuring "patch -" does not appear as a standalone command line.
	for _, line := range strings.Split(result, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "patch - Patch mappings" || trimmed == "patch -" {
			t.Errorf("Bare parent 'patch' should not appear as standalone command: %q", line)
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
