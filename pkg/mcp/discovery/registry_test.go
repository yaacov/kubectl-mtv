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

func TestRegistry_ListReadOnlyCommands(t *testing.T) {
	registry := &Registry{
		ReadOnly: map[string]*Command{
			"get/plan":         {Path: []string{"get", "plan"}},
			"describe/vm":      {Path: []string{"describe", "vm"}},
			"get/inventory/vm": {Path: []string{"get", "inventory", "vm"}},
		},
		ReadOnlyOrder: []string{"get/plan", "describe/vm", "get/inventory/vm"},
		ReadWrite: map[string]*Command{
			"create/plan": {Path: []string{"create", "plan"}},
		},
	}

	result := registry.ListReadOnlyCommands()

	// Should preserve Cobra registration order
	expected := []string{"get/plan", "describe/vm", "get/inventory/vm"}
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
		ReadWriteOrder: []string{"start/plan", "create/plan", "delete/plan"},
	}

	result := registry.ListReadWriteCommands()

	// Should preserve Cobra registration order
	expected := []string{"start/plan", "create/plan", "delete/plan"}
	if len(result) != len(expected) {
		t.Fatalf("ListReadWriteCommands() returned %d items, want %d", len(result), len(expected))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("ListReadWriteCommands()[%d] = %q, want %q", i, result[i], v)
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

func TestRegistry_RealHelpMachine_ReadOnlyGroupsInventory(t *testing.T) {
	registry := loadRealRegistry(t)

	result := registry.GenerateReadOnlyDescription()

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
	if !strings.Contains(result, "WORKFLOW") {
		t.Error("Description should include WORKFLOW instruction")
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
	// MTV context preamble should be in server instructions, not duplicated in tool description
	if strings.Contains(result, "Migration Toolkit for Virtualization") {
		t.Error("Tool description should NOT duplicate MTV preamble (now in server instructions)")
	}
	// Should NOT include orphaned convention notes (removed)
	if strings.Contains(result, "Args: <required>, [optional]") {
		t.Error("Minimal description should NOT include the orphaned args convention note")
	}
}

func TestRegistry_RealHelpMachine_ServerInstructions(t *testing.T) {
	registry := loadRealRegistry(t)

	result := registry.GenerateServerInstructions()

	// Should explain what MTV/Forklift is
	if !strings.Contains(result, "Migration Toolkit for Virtualization") {
		t.Error("Server instructions should explain what MTV is")
	}
	if !strings.Contains(result, "Forklift") {
		t.Error("Server instructions should mention the Forklift name")
	}
	// Should list the three tools
	for _, tool := range []string{"mtv_read", "mtv_write", "mtv_help"} {
		if !strings.Contains(result, tool) {
			t.Errorf("Server instructions should mention %q tool", tool)
		}
	}
	// Should establish the workflow pattern
	if !strings.Contains(result, "mtv_help") {
		t.Error("Server instructions should reference mtv_help in workflow")
	}
	if !strings.Contains(result, "Workflow") {
		t.Error("Server instructions should include numbered workflow steps")
	}
}

func TestRegistry_RealHelpMachine_ReadWriteNoBareParents(t *testing.T) {
	registry := loadRealRegistry(t)

	result := registry.GenerateReadWriteDescription()

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
	// Should include WORKFLOW instruction and mtv_help reference
	if !strings.Contains(result, "WORKFLOW") {
		t.Error("Minimal write description should include WORKFLOW instruction")
	}
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

func TestFormatCommandHelp_Nil(t *testing.T) {
	result := FormatCommandHelp(nil)
	if result != "" {
		t.Errorf("nil command should return empty string, got: %q", result)
	}
}

func TestFormatCommandHelp_RequiredFirst(t *testing.T) {
	cmd := &Command{
		PathString: "create provider",
		Flags: []Flag{
			{Name: "url", Type: "string", Description: "Provider URL"},
			{Name: "type", Type: "string", Required: true, Description: "Provider type", Enum: []string{"vsphere", "ovirt"}},
			{Name: "name", Type: "string", Required: true, Description: "Name of the provider"},
			{Name: "internal-id", Type: "string", Description: "Internal ID", Hidden: true},
		},
		Examples: []Example{
			{Description: "Create vSphere", Command: "kubectl-mtv create provider --name prod --type vsphere --url https://vcenter/sdk"},
		},
	}

	result := FormatCommandHelp(cmd)

	if !strings.Contains(result, `--- Help for "create provider" ---`) {
		t.Error("should contain command header")
	}

	// Required flags should appear before optional flags
	typeIdx := strings.Index(result, "--type")
	nameIdx := strings.Index(result, "--name")
	urlIdx := strings.Index(result, "--url")
	if typeIdx < 0 || nameIdx < 0 || urlIdx < 0 {
		t.Fatalf("missing expected flags in output:\n%s", result)
	}
	if urlIdx < typeIdx || urlIdx < nameIdx {
		t.Errorf("optional --url should appear after required flags; type@%d name@%d url@%d", typeIdx, nameIdx, urlIdx)
	}

	// Required markers
	if !strings.Contains(result, "(REQUIRED)") {
		t.Error("required flags should be marked (REQUIRED)")
	}

	// Enum values
	if !strings.Contains(result, "[vsphere, ovirt]") {
		t.Error("enum values should be shown in brackets")
	}

	// Hidden flags excluded
	if strings.Contains(result, "internal_id") || strings.Contains(result, "internal-id") {
		t.Error("hidden flags should not appear")
	}

	// Example in MCP format
	if !strings.Contains(result, "Example:") {
		t.Error("should contain example")
	}
	if !strings.Contains(result, `command: "create provider"`) {
		t.Error("example should be in MCP format")
	}
}

func TestFormatCommandHelp_MultipleExamples(t *testing.T) {
	cmd := &Command{
		PathString: "get plan",
		Flags: []Flag{
			{Name: "namespace", Type: "string", Description: "Target namespace"},
		},
		Examples: []Example{
			{Description: "List all plans", Command: "kubectl-mtv get plan --namespace ns1"},
			{Description: "Get specific plan", Command: "kubectl-mtv get plan --namespace ns1 --name my-plan"},
		},
	}

	result := FormatCommandHelp(cmd)

	// Multiple examples should use "Examples:" header (plural)
	if !strings.Contains(result, "Examples:") {
		t.Error("multiple examples should use plural header")
	}
	if strings.Contains(result, "Example:") && !strings.Contains(result, "Examples:") {
		t.Error("should use 'Examples:' not 'Example:' for multiple examples")
	}
}

func TestFormatCommandHelp_NoFlagsNoExamples(t *testing.T) {
	cmd := &Command{
		PathString: "get plan",
	}

	result := FormatCommandHelp(cmd)

	if !strings.Contains(result, `--- Help for "get plan" ---`) {
		t.Error("should contain header even with no flags/examples")
	}
	if strings.Contains(result, "Flags:") {
		t.Error("should not have Flags section when no flags")
	}
	if strings.Contains(result, "Example") {
		t.Error("should not have Example section when no examples")
	}
}

func TestFormatCommandHelp_RealCreateProvider(t *testing.T) {
	registry := loadRealRegistry(t)

	cmd, ok := registry.ReadWrite["create/provider"]
	if !ok {
		t.Skip("create/provider not found in registry")
	}

	result := FormatCommandHelp(cmd)

	if !strings.Contains(result, `--- Help for "create provider" ---`) {
		t.Error("should contain command header")
	}
	if !strings.Contains(result, "(REQUIRED)") {
		t.Error("should show required flags")
	}
	if !strings.Contains(result, "--type") {
		t.Error("should list --type flag")
	}

	// All examples should be present (no cap)
	allExamples := convertCLIToMCPExamples(cmd, len(cmd.Examples))
	for _, ex := range allExamples {
		if !strings.Contains(result, ex) {
			t.Errorf("missing example in help text: %s", ex)
		}
	}

	t.Logf("FormatCommandHelp output (%d chars):\n%s", len(result), result)
}
