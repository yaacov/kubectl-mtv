package help

import (
	"testing"

	"github.com/spf13/cobra"
)

// --- helpers ---

// testEnumFlag implements EnumValuer for testing enum detection.
type testEnumFlag struct {
	val         string
	validValues []string
}

func (f *testEnumFlag) String() string     { return f.val }
func (f *testEnumFlag) Set(v string) error { f.val = v; return nil }
func (f *testEnumFlag) Type() string       { return "string" }
func (f *testEnumFlag) GetValidValues() []string {
	return f.validValues
}

// buildTree creates a tiny command tree for testing:
//
//	root
//	├── get
//	│   ├── plan       (runnable, read)
//	│   └── provider   (runnable, read)
//	├── create
//	│   └── plan       (runnable, write)
//	├── delete
//	│   └── plan       (runnable, write)
//	└── health         (runnable, read)
func buildTree() *cobra.Command {
	root := &cobra.Command{
		Use:   "kubectl-mtv",
		Short: "Migrate virtual machines",
		Long:  "Long description of kubectl-mtv",
	}
	root.PersistentFlags().IntP("verbose", "v", 0, "Verbosity level")
	root.PersistentFlags().BoolP("all-namespaces", "A", false, "All namespaces")

	// --- get ---
	getCmd := &cobra.Command{Use: "get", Short: "Get resources"}

	getPlanCmd := &cobra.Command{
		Use:     "plan",
		Short:   "Get migration plans",
		Long:    "Get migration plans with optional filters.",
		Aliases: []string{"plans"},
		Example: `  # List all plans
  kubectl-mtv get plan

  # Get a specific plan
  kubectl-mtv get plan --name my-plan`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	getPlanCmd.Flags().StringP("name", "", "", "Plan name")
	getPlanCmd.Flags().StringP("output", "o", "table", "Output format")

	getProviderCmd := &cobra.Command{
		Use:   "provider",
		Short: "Get providers",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
	getProviderCmd.Flags().StringP("name", "", "", "Provider name")

	getCmd.AddCommand(getPlanCmd, getProviderCmd)

	// --- create ---
	createCmd := &cobra.Command{Use: "create", Short: "Create resources"}

	createPlanCmd := &cobra.Command{
		Use:   "plan",
		Short: "Create a migration plan",
		Args:  cobra.NoArgs,
		Example: `  # Create a plan
  kubectl-mtv create plan --name my-plan \
    --source vsphere-prod \
    --target host`,
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	createPlanCmd.Flags().String("name", "", "Plan name")
	_ = createPlanCmd.MarkFlagRequired("name")
	createPlanCmd.Flags().String("source", "", "Source provider")
	createPlanCmd.Flags().String("target", "", "Target")

	createCmd.AddCommand(createPlanCmd)

	// --- delete ---
	deleteCmd := &cobra.Command{Use: "delete", Short: "Delete resources"}
	deletePlanCmd := &cobra.Command{
		Use:   "plan",
		Short: "Delete migration plans",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
	deletePlanCmd.Flags().StringSlice("name", nil, "Plan names (comma-separated)")
	deletePlanCmd.Flags().Bool("all", false, "Delete all plans")
	deleteCmd.AddCommand(deletePlanCmd)

	// --- health ---
	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "Check MTV health",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}

	root.AddCommand(getCmd, createCmd, deleteCmd, healthCmd)
	return root
}

// --- Generate tests ---

func TestGenerate_BasicSchema(t *testing.T) {
	root := buildTree()
	schema := Generate(root, "v0.1.0", DefaultOptions())

	if schema.Version != SchemaVersion {
		t.Errorf("expected version %s, got %s", SchemaVersion, schema.Version)
	}
	if schema.CLIVersion != "v0.1.0" {
		t.Errorf("expected cli version v0.1.0, got %s", schema.CLIVersion)
	}
	if schema.Name != "kubectl-mtv" {
		t.Errorf("expected name kubectl-mtv, got %s", schema.Name)
	}
	if schema.Description != "Migrate virtual machines" {
		t.Errorf("expected short description, got %s", schema.Description)
	}
	if schema.LongDescription != "Long description of kubectl-mtv" {
		t.Errorf("expected long description, got %s", schema.LongDescription)
	}
}

func TestGenerate_CommandCount(t *testing.T) {
	root := buildTree()
	schema := Generate(root, "v0.1.0", DefaultOptions())

	// We expect: get plan, get provider, create plan, delete plan, health = 5
	if len(schema.Commands) != 5 {
		t.Errorf("expected 5 commands, got %d", len(schema.Commands))
		for _, c := range schema.Commands {
			t.Logf("  command: %s", c.PathString)
		}
	}
}

func TestGenerate_ReadOnlyFilter(t *testing.T) {
	root := buildTree()
	opts := DefaultOptions()
	opts.ReadOnly = true
	schema := Generate(root, "v0.1.0", opts)

	// Read-only: get plan, get provider, health = 3
	if len(schema.Commands) != 3 {
		t.Errorf("expected 3 read-only commands, got %d", len(schema.Commands))
		for _, c := range schema.Commands {
			t.Logf("  command: %s (category=%s)", c.PathString, c.Category)
		}
	}
	for _, c := range schema.Commands {
		if c.Category != "read" {
			t.Errorf("expected category read, got %s for %s", c.Category, c.PathString)
		}
	}
}

func TestGenerate_WriteFilter(t *testing.T) {
	root := buildTree()
	opts := DefaultOptions()
	opts.Write = true
	schema := Generate(root, "v0.1.0", opts)

	// Write: create plan, delete plan = 2
	if len(schema.Commands) != 2 {
		t.Errorf("expected 2 write commands, got %d", len(schema.Commands))
		for _, c := range schema.Commands {
			t.Logf("  command: %s (category=%s)", c.PathString, c.Category)
		}
	}
	for _, c := range schema.Commands {
		if c.Category != "write" {
			t.Errorf("expected category write, got %s for %s", c.Category, c.PathString)
		}
	}
}

func TestGenerate_ShortMode(t *testing.T) {
	root := buildTree()
	opts := DefaultOptions()
	opts.Short = true
	schema := Generate(root, "v0.1.0", opts)

	for _, c := range schema.Commands {
		if c.LongDescription != "" {
			t.Errorf("expected empty long description in short mode for %s, got %q", c.PathString, c.LongDescription)
		}
		if len(c.Examples) > 0 {
			t.Errorf("expected no examples in short mode for %s, got %d", c.PathString, len(c.Examples))
		}
	}
}

func TestGenerate_GlobalFlags(t *testing.T) {
	root := buildTree()
	opts := DefaultOptions()
	opts.IncludeGlobalFlags = true
	schema := Generate(root, "v0.1.0", opts)

	if len(schema.GlobalFlags) == 0 {
		t.Error("expected global flags to be populated")
	}

	flagNames := map[string]bool{}
	for _, f := range schema.GlobalFlags {
		flagNames[f.Name] = true
	}
	if !flagNames["verbose"] {
		t.Error("expected global flag 'verbose'")
	}
	if !flagNames["all-namespaces"] {
		t.Error("expected global flag 'all-namespaces'")
	}
}

func TestGenerate_NoGlobalFlags(t *testing.T) {
	root := buildTree()
	opts := DefaultOptions()
	opts.IncludeGlobalFlags = false
	schema := Generate(root, "v0.1.0", opts)

	if len(schema.GlobalFlags) != 0 {
		t.Errorf("expected no global flags, got %d", len(schema.GlobalFlags))
	}
}

func TestGenerate_HiddenCommandsExcluded(t *testing.T) {
	root := buildTree()
	hiddenCmd := &cobra.Command{
		Use:    "secret",
		Short:  "Hidden command",
		Hidden: true,
		RunE:   func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(hiddenCmd)

	schema := Generate(root, "v0.1.0", DefaultOptions())
	for _, c := range schema.Commands {
		if c.Name == "secret" {
			t.Error("hidden command should not appear without IncludeHidden")
		}
	}
}

func TestGenerate_HiddenCommandsIncluded(t *testing.T) {
	root := buildTree()
	hiddenCmd := &cobra.Command{
		Use:    "secret",
		Short:  "Hidden command",
		Hidden: true,
		RunE:   func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(hiddenCmd)

	opts := DefaultOptions()
	opts.IncludeHidden = true
	schema := Generate(root, "v0.1.0", opts)

	found := false
	for _, c := range schema.Commands {
		if c.Name == "secret" {
			found = true
			break
		}
	}
	if !found {
		t.Error("hidden command should appear with IncludeHidden")
	}
}

// --- commandToSchema tests ---

func TestCommandToSchema_BasicFields(t *testing.T) {
	root := buildTree()
	// Find "get plan"
	getCmd, _, _ := root.Find([]string{"get", "plan"})
	path := []string{"get", "plan"}
	c := commandToSchema(getCmd, path, DefaultOptions())

	if c.Name != "plan" {
		t.Errorf("expected name 'plan', got %q", c.Name)
	}
	if c.PathString != "get plan" {
		t.Errorf("expected path_string 'get plan', got %q", c.PathString)
	}
	if c.Category != "read" {
		t.Errorf("expected category read, got %q", c.Category)
	}
	if c.Description != "Get migration plans" {
		t.Errorf("expected description, got %q", c.Description)
	}
	if c.LongDescription != "Get migration plans with optional filters." {
		t.Errorf("expected long description, got %q", c.LongDescription)
	}
	if len(c.Aliases) != 1 || c.Aliases[0] != "plans" {
		t.Errorf("expected aliases [plans], got %v", c.Aliases)
	}
}

func TestCommandToSchema_Flags(t *testing.T) {
	root := buildTree()
	getCmd, _, _ := root.Find([]string{"get", "plan"})
	path := []string{"get", "plan"}
	c := commandToSchema(getCmd, path, DefaultOptions())

	flagNames := map[string]bool{}
	for _, f := range c.Flags {
		flagNames[f.Name] = true
	}
	if !flagNames["name"] {
		t.Error("expected flag 'name'")
	}
	if !flagNames["output"] {
		t.Error("expected flag 'output'")
	}
}

func TestCommandToSchema_RequiredFlag(t *testing.T) {
	root := buildTree()
	createPlan, _, _ := root.Find([]string{"create", "plan"})
	path := []string{"create", "plan"}
	c := commandToSchema(createPlan, path, DefaultOptions())

	for _, f := range c.Flags {
		if f.Name == "name" && !f.Required {
			t.Error("expected 'name' flag to be marked as required")
		}
	}
}

func TestCommandToSchema_Examples(t *testing.T) {
	root := buildTree()
	getCmd, _, _ := root.Find([]string{"get", "plan"})
	path := []string{"get", "plan"}
	c := commandToSchema(getCmd, path, DefaultOptions())

	if len(c.Examples) != 2 {
		t.Fatalf("expected 2 examples, got %d", len(c.Examples))
	}
	if c.Examples[0].Description != "List all plans" {
		t.Errorf("expected first example description, got %q", c.Examples[0].Description)
	}
	if c.Examples[0].Command != "kubectl-mtv get plan" {
		t.Errorf("expected first example command, got %q", c.Examples[0].Command)
	}
}

// --- flagToSchema tests ---

func TestFlagToSchema_String(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("name", "n", "default-val", "The name")
	f := cmd.Flags().Lookup("name")

	schema := flagToSchema(f)
	if schema.Name != "name" {
		t.Errorf("expected name 'name', got %q", schema.Name)
	}
	if schema.Shorthand != "n" {
		t.Errorf("expected shorthand 'n', got %q", schema.Shorthand)
	}
	if schema.Type != "string" {
		t.Errorf("expected type 'string', got %q", schema.Type)
	}
	if schema.Default != "default-val" {
		t.Errorf("expected default 'default-val', got %v", schema.Default)
	}
	if schema.Description != "The name" {
		t.Errorf("expected description, got %q", schema.Description)
	}
}

func TestFlagToSchema_Bool(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("verbose", false, "Verbose output")
	f := cmd.Flags().Lookup("verbose")

	schema := flagToSchema(f)
	if schema.Type != "bool" {
		t.Errorf("expected type 'bool', got %q", schema.Type)
	}
	if def, ok := schema.Default.(bool); !ok || def != false {
		t.Errorf("expected default false, got %v", schema.Default)
	}
}

func TestFlagToSchema_BoolTrue(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("enabled", true, "Enable feature")
	f := cmd.Flags().Lookup("enabled")

	schema := flagToSchema(f)
	if def, ok := schema.Default.(bool); !ok || def != true {
		t.Errorf("expected default true, got %v", schema.Default)
	}
}

func TestFlagToSchema_Int(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Int("count", 42, "Number of items")
	f := cmd.Flags().Lookup("count")

	schema := flagToSchema(f)
	if schema.Type != "int" {
		t.Errorf("expected type 'int', got %q", schema.Type)
	}
	if def, ok := schema.Default.(int64); !ok || def != 42 {
		t.Errorf("expected default 42, got %v (type %T)", schema.Default, schema.Default)
	}
}

func TestFlagToSchema_Float(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Float64("ratio", 3.14, "Ratio")
	f := cmd.Flags().Lookup("ratio")

	schema := flagToSchema(f)
	if schema.Type != "float64" {
		t.Errorf("expected type 'float64', got %q", schema.Type)
	}
	if def, ok := schema.Default.(float64); !ok || def != 3.14 {
		t.Errorf("expected default 3.14, got %v", schema.Default)
	}
}

func TestFlagToSchema_StringSlice(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringSlice("names", nil, "Names")
	f := cmd.Flags().Lookup("names")

	schema := flagToSchema(f)
	if schema.Type != "stringSlice" {
		t.Errorf("expected type 'stringSlice', got %q", schema.Type)
	}
	// Default for nil StringSlice is "[]"
	if def, ok := schema.Default.([]string); !ok || len(def) != 0 {
		t.Errorf("expected default empty slice, got %v (type %T)", schema.Default, schema.Default)
	}
}

func TestFlagToSchema_Hidden(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("secret", "", "Secret flag")
	_ = cmd.Flags().MarkHidden("secret")
	f := cmd.Flags().Lookup("secret")

	schema := flagToSchema(f)
	if !schema.Hidden {
		t.Error("expected flag to be marked hidden")
	}
}

func TestFlagToSchema_Enum(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	ef := &testEnumFlag{val: "json", validValues: []string{"json", "yaml", "table"}}
	cmd.Flags().Var(ef, "output", "Output format")
	f := cmd.Flags().Lookup("output")

	schema := flagToSchema(f)
	if len(schema.Enum) != 3 {
		t.Fatalf("expected 3 enum values, got %d", len(schema.Enum))
	}
	if schema.Enum[0] != "json" || schema.Enum[1] != "yaml" || schema.Enum[2] != "table" {
		t.Errorf("unexpected enum values: %v", schema.Enum)
	}
}

// --- getCategory tests ---

func TestGetCategory(t *testing.T) {
	tests := []struct {
		path     []string
		expected string
	}{
		{[]string{}, "admin"},
		{[]string{"get"}, "read"},
		{[]string{"get", "plan"}, "read"},
		{[]string{"describe"}, "read"},
		{[]string{"describe", "plan"}, "read"},
		{[]string{"health"}, "read"},
		{[]string{"create"}, "write"},
		{[]string{"create", "plan"}, "write"},
		{[]string{"delete"}, "write"},
		{[]string{"delete", "plan"}, "write"},
		{[]string{"patch"}, "write"},
		{[]string{"patch", "plan"}, "write"},
		{[]string{"start"}, "write"},
		{[]string{"start", "plan"}, "write"},
		{[]string{"cancel"}, "write"},
		{[]string{"archive"}, "write"},
		{[]string{"unarchive"}, "write"},
		{[]string{"cutover"}, "write"},
		{[]string{"settings"}, "read"},
		{[]string{"settings", "get"}, "read"},
		{[]string{"settings", "set"}, "write"},
		{[]string{"settings", "unset"}, "write"},
		{[]string{"unknown"}, "admin"},
		{[]string{"help"}, "admin"},
	}

	for _, tt := range tests {
		t.Run(joinPath(tt.path), func(t *testing.T) {
			result := getCategory(tt.path)
			if result != tt.expected {
				t.Errorf("getCategory(%v) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func joinPath(path []string) string {
	if len(path) == 0 {
		return "<empty>"
	}
	result := path[0]
	for _, p := range path[1:] {
		result += "/" + p
	}
	return result
}

// --- parseExamples tests ---

func TestParseExamples_Empty(t *testing.T) {
	examples := parseExamples("")
	if examples != nil {
		t.Errorf("expected nil for empty input, got %v", examples)
	}
}

func TestParseExamples_SingleLine(t *testing.T) {
	input := `  # List plans
  kubectl-mtv get plan`
	examples := parseExamples(input)

	if len(examples) != 1 {
		t.Fatalf("expected 1 example, got %d", len(examples))
	}
	if examples[0].Description != "List plans" {
		t.Errorf("expected description 'List plans', got %q", examples[0].Description)
	}
	if examples[0].Command != "kubectl-mtv get plan" {
		t.Errorf("expected command 'kubectl-mtv get plan', got %q", examples[0].Command)
	}
}

func TestParseExamples_MultipleExamples(t *testing.T) {
	input := `  # List all plans
  kubectl-mtv get plan

  # Get a specific plan
  kubectl-mtv get plan --name my-plan`
	examples := parseExamples(input)

	if len(examples) != 2 {
		t.Fatalf("expected 2 examples, got %d", len(examples))
	}
	if examples[0].Description != "List all plans" {
		t.Errorf("example[0] desc = %q", examples[0].Description)
	}
	if examples[1].Description != "Get a specific plan" {
		t.Errorf("example[1] desc = %q", examples[1].Description)
	}
	if examples[1].Command != "kubectl-mtv get plan --name my-plan" {
		t.Errorf("example[1] cmd = %q", examples[1].Command)
	}
}

func TestParseExamples_BackslashContinuation(t *testing.T) {
	input := `  # Create a plan with multiple flags
  kubectl-mtv create plan --name my-plan \
    --source vsphere-prod \
    --target host`
	examples := parseExamples(input)

	if len(examples) != 1 {
		t.Fatalf("expected 1 example, got %d", len(examples))
	}
	expected := "kubectl-mtv create plan --name my-plan --source vsphere-prod --target host"
	if examples[0].Command != expected {
		t.Errorf("expected command %q, got %q", expected, examples[0].Command)
	}
}

func TestParseExamples_NoDescription(t *testing.T) {
	input := `  kubectl-mtv get plan`
	examples := parseExamples(input)

	if len(examples) != 1 {
		t.Fatalf("expected 1 example, got %d", len(examples))
	}
	if examples[0].Description != "" {
		t.Errorf("expected empty description, got %q", examples[0].Description)
	}
}

func TestParseExamples_TrailingBackslash(t *testing.T) {
	// Edge case: example ends with backslash but no continuation
	input := `  # Trailing backslash edge case
  kubectl-mtv get plan \`
	examples := parseExamples(input)

	if len(examples) != 1 {
		t.Fatalf("expected 1 example, got %d", len(examples))
	}
	if examples[0].Command != "kubectl-mtv get plan" {
		t.Errorf("expected 'kubectl-mtv get plan', got %q", examples[0].Command)
	}
}

func TestParseExamples_ConsecutiveDescriptions(t *testing.T) {
	// When two # comments appear in a row, the second one overwrites the first
	input := `  # First comment
  # Actual description
  kubectl-mtv get plan`
	examples := parseExamples(input)

	if len(examples) != 1 {
		t.Fatalf("expected 1 example, got %d", len(examples))
	}
	if examples[0].Description != "Actual description" {
		t.Errorf("expected 'Actual description', got %q", examples[0].Description)
	}
}

func TestParseExamples_DescriptionThenNewExample(t *testing.T) {
	// A pending command should be flushed when a new description appears
	input := `  # First example
  kubectl-mtv get plan \
    --name foo
  # Second example
  kubectl-mtv delete plan --name bar`
	examples := parseExamples(input)

	if len(examples) != 2 {
		t.Fatalf("expected 2 examples, got %d", len(examples))
	}
	if examples[0].Command != "kubectl-mtv get plan --name foo" {
		t.Errorf("example[0] cmd = %q", examples[0].Command)
	}
	if examples[1].Command != "kubectl-mtv delete plan --name bar" {
		t.Errorf("example[1] cmd = %q", examples[1].Command)
	}
}

// --- FilterByPath tests ---

func TestFilterByPath_NoFilter(t *testing.T) {
	root := buildTree()
	schema := Generate(root, "v0.1.0", DefaultOptions())
	total := len(schema.Commands)

	n := FilterByPath(schema, nil)
	if n != total {
		t.Errorf("expected %d commands with nil prefix, got %d", total, n)
	}
}

func TestFilterByPath_GetCommands(t *testing.T) {
	root := buildTree()
	schema := Generate(root, "v0.1.0", DefaultOptions())

	n := FilterByPath(schema, []string{"get"})
	if n != 2 {
		t.Errorf("expected 2 'get' commands, got %d", n)
		for _, c := range schema.Commands {
			t.Logf("  command: %s", c.PathString)
		}
	}
}

func TestFilterByPath_SpecificCommand(t *testing.T) {
	root := buildTree()
	schema := Generate(root, "v0.1.0", DefaultOptions())

	n := FilterByPath(schema, []string{"get", "plan"})
	if n != 1 {
		t.Errorf("expected 1 'get plan' command, got %d", n)
	}
	if len(schema.Commands) != 1 {
		t.Fatalf("expected exactly 1 command after filter")
	}
	if schema.Commands[0].PathString != "get plan" {
		t.Errorf("expected 'get plan', got %q", schema.Commands[0].PathString)
	}
}

func TestFilterByPath_NoMatch(t *testing.T) {
	root := buildTree()
	schema := Generate(root, "v0.1.0", DefaultOptions())

	n := FilterByPath(schema, []string{"nonexistent"})
	if n != 0 {
		t.Errorf("expected 0 commands for nonexistent prefix, got %d", n)
	}
}

// --- walkCommands tests ---

func TestWalkCommands_CountsAllNodes(t *testing.T) {
	root := buildTree()
	var visited []string
	walkCommands(root, []string{}, func(cmd *cobra.Command, path []string) {
		if len(path) > 0 {
			visited = append(visited, joinPath(path))
		}
	})

	// Expect: get, get/plan, get/provider, create, create/plan, delete, delete/plan, health
	// Plus cobra auto-added commands (completion, help) might appear too
	if len(visited) < 8 {
		t.Errorf("expected at least 8 visited nodes, got %d: %v", len(visited), visited)
	}
}

// --- IsRequiredFlag tests ---

func TestIsRequiredFlag_Required(t *testing.T) {
	cmd := &cobra.Command{Use: "test", RunE: func(cmd *cobra.Command, args []string) error { return nil }}
	cmd.Flags().String("name", "", "Name")
	_ = cmd.MarkFlagRequired("name")

	if !IsRequiredFlag(cmd, "name") {
		t.Error("expected 'name' to be required")
	}
}

func TestIsRequiredFlag_NotRequired(t *testing.T) {
	cmd := &cobra.Command{Use: "test", RunE: func(cmd *cobra.Command, args []string) error { return nil }}
	cmd.Flags().String("name", "", "Name")

	if IsRequiredFlag(cmd, "name") {
		t.Error("expected 'name' to not be required")
	}
}

func TestIsRequiredFlag_NonExistentFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test", RunE: func(cmd *cobra.Command, args []string) error { return nil }}

	if IsRequiredFlag(cmd, "nonexistent") {
		t.Error("expected false for non-existent flag")
	}
}

// --- DefaultOptions tests ---

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.ReadOnly {
		t.Error("expected ReadOnly to be false")
	}
	if opts.Write {
		t.Error("expected Write to be false")
	}
	if !opts.IncludeGlobalFlags {
		t.Error("expected IncludeGlobalFlags to be true")
	}
	if opts.IncludeHidden {
		t.Error("expected IncludeHidden to be false")
	}
	if opts.Short {
		t.Error("expected Short to be false")
	}
}

// --- Non-runnable command tests ---

func TestGenerate_NonRunnableParentExcluded(t *testing.T) {
	root := buildTree()
	schema := Generate(root, "v0.1.0", DefaultOptions())

	for _, c := range schema.Commands {
		// "get", "create", "delete" are non-runnable parents with < 3 runnable children each
		// They should NOT appear in the schema
		if c.PathString == "get" || c.PathString == "create" || c.PathString == "delete" {
			t.Errorf("non-runnable parent %q should not appear in schema", c.PathString)
		}
	}
}

func TestGenerate_NonRunnableParentIncludedWith3PlusChildren(t *testing.T) {
	// Create a command tree where a depth-2 non-runnable parent has 3+ runnable children
	root := &cobra.Command{Use: "kubectl-mtv", Short: "CLI"}
	getCmd := &cobra.Command{Use: "get", Short: "Get resources"}
	inventoryCmd := &cobra.Command{Use: "inventory", Short: "Get inventory resources"}

	for _, name := range []string{"vm", "network", "storage"} {
		child := &cobra.Command{
			Use:   name,
			Short: "Get " + name,
			RunE:  func(cmd *cobra.Command, args []string) error { return nil },
		}
		inventoryCmd.AddCommand(child)
	}
	getCmd.AddCommand(inventoryCmd)
	root.AddCommand(getCmd)

	schema := Generate(root, "v0.1.0", DefaultOptions())

	found := false
	for _, c := range schema.Commands {
		if c.PathString == "get inventory" {
			found = true
			if c.Runnable {
				t.Error("expected 'get inventory' to be non-runnable")
			}
			break
		}
	}
	if !found {
		t.Error("expected 'get inventory' to appear (non-runnable parent with 3+ children)")
	}
}

// --- Shorthand detection via schema ---

func TestFlagToSchema_Shorthand(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("provider", "p", "", "Provider name")
	f := cmd.Flags().Lookup("provider")

	schema := flagToSchema(f)
	if schema.Shorthand != "p" {
		t.Errorf("expected shorthand 'p', got %q", schema.Shorthand)
	}
}

func TestFlagToSchema_NoShorthand(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "Name")
	f := cmd.Flags().Lookup("name")

	schema := flagToSchema(f)
	if schema.Shorthand != "" {
		t.Errorf("expected empty shorthand, got %q", schema.Shorthand)
	}
}
