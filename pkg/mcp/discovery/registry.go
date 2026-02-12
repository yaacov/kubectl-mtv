package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// Registry holds discovered kubectl-mtv commands organized by read/write access.
type Registry struct {
	// ReadOnly contains commands that don't modify cluster state
	ReadOnly map[string]*Command

	// ReadOnlyOrder preserves the original registration order of read-only
	// commands from help --machine. This is used for example selection so
	// the first command per group matches the developer's intended ordering.
	ReadOnlyOrder []string

	// ReadWrite contains commands that modify cluster state
	ReadWrite map[string]*Command

	// ReadWriteOrder preserves the original registration order of read-write commands.
	ReadWriteOrder []string

	// Parents contains non-runnable structural parent commands
	// (e.g., "get/inventory") for description lookup during group compaction.
	Parents map[string]*Command

	// GlobalFlags are flags available to all commands
	GlobalFlags []Flag

	// RootDescription is the main kubectl-mtv short description
	RootDescription string

	// LongDescription is the extended CLI description with domain context
	// (e.g., "Migrate virtual machines from VMware vSphere, oVirt...")
	LongDescription string
}

// NewRegistry creates a new registry by calling kubectl-mtv help --machine.
// This single call returns the complete command schema as JSON.
// It uses os.Executable() to call the same binary that is running the MCP server,
// ensuring the help schema always matches the server's code (avoids version mismatch
// when a different kubectl-mtv version is installed in PATH).
func NewRegistry(ctx context.Context) (*Registry, error) {
	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Use the current executable to ensure help matches the running server
	self, err := os.Executable()
	if err != nil {
		// Fall back to PATH lookup if os.Executable fails
		self = "kubectl-mtv"
	}
	cmd := exec.CommandContext(cmdCtx, self, "help", "--machine")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get machine help: %w", err)
	}

	var schema HelpSchema
	if err := json.Unmarshal(output, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse help schema: %w", err)
	}

	registry := &Registry{
		ReadOnly:        make(map[string]*Command),
		ReadWrite:       make(map[string]*Command),
		Parents:         make(map[string]*Command),
		GlobalFlags:     schema.GlobalFlags,
		RootDescription: schema.Description,
		LongDescription: schema.LongDescription,
	}

	// Categorize commands by read/write based on category field.
	// Admin commands (completions, help, mcp-server, version) are skipped
	// entirely — they are irrelevant to LLM tool use.
	// Non-runnable parent commands are stored separately for description lookup.
	//
	// Backward compatibility: older CLI versions may not emit the "runnable" field.
	// We detect this by checking if ANY command has Runnable=true. If none do,
	// the schema predates the field and all commands are leaf commands.
	schemaHasRunnable := false
	for i := range schema.Commands {
		if schema.Commands[i].Runnable {
			schemaHasRunnable = true
			break
		}
	}

	for i := range schema.Commands {
		cmd := &schema.Commands[i]
		pathKey := cmd.PathKey()

		// Determine if this command is runnable:
		// - If the schema has the field: trust cmd.Runnable
		// - If the schema lacks the field: all commands are leaf (runnable)
		isRunnable := cmd.Runnable || !schemaHasRunnable

		// Store non-runnable parents separately for group description lookup
		if !isRunnable {
			registry.Parents[pathKey] = cmd
			continue
		}

		switch cmd.Category {
		case "read":
			registry.ReadOnly[pathKey] = cmd
			registry.ReadOnlyOrder = append(registry.ReadOnlyOrder, pathKey)
		case "admin":
			// Skip admin commands (shell completions, help, version, etc.)
			continue
		default:
			// "write" category
			registry.ReadWrite[pathKey] = cmd
			registry.ReadWriteOrder = append(registry.ReadWriteOrder, pathKey)
		}
	}

	return registry, nil
}

// GetCommand returns a command by its path key (e.g., "get/plan").
func (r *Registry) GetCommand(pathKey string) *Command {
	if cmd, ok := r.ReadOnly[pathKey]; ok {
		return cmd
	}
	if cmd, ok := r.ReadWrite[pathKey]; ok {
		return cmd
	}
	return nil
}

// GetCommandByPath returns a command by its path slice.
func (r *Registry) GetCommandByPath(path []string) *Command {
	key := strings.Join(path, "/")
	return r.GetCommand(key)
}

// ListReadOnlyCommands returns sorted list of read-only command paths.
func (r *Registry) ListReadOnlyCommands() []string {
	var commands []string
	for key := range r.ReadOnly {
		commands = append(commands, key)
	}
	sort.Strings(commands)
	return commands
}

// ListReadWriteCommands returns sorted list of read-write command paths.
func (r *Registry) ListReadWriteCommands() []string {
	var commands []string
	for key := range r.ReadWrite {
		commands = append(commands, key)
	}
	sort.Strings(commands)
	return commands
}

// IsReadOnly checks if a command path is read-only.
func (r *Registry) IsReadOnly(pathKey string) bool {
	_, ok := r.ReadOnly[pathKey]
	return ok
}

// IsReadWrite checks if a command path is read-write.
func (r *Registry) IsReadWrite(pathKey string) bool {
	_, ok := r.ReadWrite[pathKey]
	return ok
}

// GenerateReadOnlyDescription generates a description for the read-only tool.
func (r *Registry) GenerateReadOnlyDescription() string {
	roots := uniqueRootVerbs(r.ReadOnly)

	var sb strings.Builder
	sb.WriteString("Execute read-only kubectl-mtv commands to query MTV resources.\n\n")
	sb.WriteString(fmt.Sprintf("Commands: %s\n\n", strings.Join(roots, ", ")))
	sb.WriteString("Available commands:\n")

	commands := r.ListReadOnlyCommands()
	for _, key := range commands {
		cmd := r.ReadOnly[key]
		// Format: "get plan [NAME]" -> "get plan [NAME] - Get migration plans"
		usage := formatUsageShort(cmd)
		sb.WriteString(fmt.Sprintf("- %s - %s\n", usage, cmd.Description))
	}

	sb.WriteString(r.formatGlobalFlags())

	// Include extended notes from commands that have substantial LongDescription
	sb.WriteString(r.generateReadOnlyCommandNotes())

	sb.WriteString("\nEnvironment Variable References:\n")
	sb.WriteString("- Use ${ENV_VAR_NAME} syntax to pass environment variable references as flag values\n")

	return sb.String()
}

// GenerateReadWriteDescription generates a description for the read-write tool.
func (r *Registry) GenerateReadWriteDescription() string {
	var sb strings.Builder
	sb.WriteString("Execute kubectl-mtv commands that modify cluster state.\n\n")
	sb.WriteString("WARNING: These commands create, modify, or delete resources.\n\n")
	readRoots := uniqueRootVerbs(r.ReadOnly)
	sb.WriteString(fmt.Sprintf("NOTE: For read-only operations (%s), use the mtv_read tool instead.\n\n", strings.Join(readRoots, ", ")))
	sb.WriteString("Available commands:\n")

	commands := r.ListReadWriteCommands()
	for _, key := range commands {
		cmd := r.ReadWrite[key]
		usage := formatUsageShort(cmd)
		sb.WriteString(fmt.Sprintf("- %s - %s\n", usage, cmd.Description))
	}

	sb.WriteString(r.formatGlobalFlags())

	// Append per-command flag reference for complex write commands
	sb.WriteString(r.generateFlagReference())

	sb.WriteString("\nEnvironment Variable References:\n")
	sb.WriteString("You can pass environment variable references instead of literal values for any flag:\n")
	sb.WriteString("- Use ${ENV_VAR_NAME} syntax with curly braces (e.g., url: \"${GOVC_URL}\", password: \"${VCENTER_PASSWORD}\")\n")
	sb.WriteString("- Env vars can be embedded in strings (e.g., url: \"${GOVC_URL}/sdk\", endpoint: \"https://${HOST}:${PORT}/api\")\n")
	sb.WriteString("- IMPORTANT: Only ${VAR} format is recognized as env var reference. Bare $VAR is treated as literal value.\n")
	sb.WriteString("- The MCP server resolves the env var at execution time\n")
	sb.WriteString("- Sensitive values (passwords, tokens) are masked in command output for security\n")

	return sb.String()
}

// GenerateMinimalReadOnlyDescription generates a minimal description for the read-only tool.
// It puts the command list first (most critical for the LM), followed by examples
// and a hint to use mtv_help. Domain context and conventions are omitted to stay
// within client description limits and avoid noise for smaller models.
func (r *Registry) GenerateMinimalReadOnlyDescription() string {
	var sb strings.Builder

	sb.WriteString("Query MTV resources (read-only).\n")
	sb.WriteString("\nCommands:\n")

	// Detect sibling groups (e.g., get/inventory/*) to compact them
	const minGroupSize = 3
	groups, groupedKeys := detectSiblingGroups(r.ReadOnly, r.Parents, minGroupSize)

	// List non-grouped commands normally
	commands := r.ListReadOnlyCommands()
	for _, key := range commands {
		if groupedKeys[key] {
			continue
		}
		cmd := r.ReadOnly[key]
		usage := formatUsageShort(cmd)
		sb.WriteString(fmt.Sprintf("  %s - %s\n", usage, cmd.Description))
	}

	// Write compacted sibling groups with shared required arg shown as <flag>
	for _, group := range groups {
		parentDisplay := strings.ReplaceAll(group.ParentPath, "/", " ")

		if group.SharedRequiredArg != "" {
			argLower := strings.ToLower(group.SharedRequiredArg)
			sb.WriteString(fmt.Sprintf("  %s RESOURCE <%s> - %s\n", parentDisplay, argLower, group.Description))
		} else {
			sb.WriteString(fmt.Sprintf("  %s RESOURCE - %s\n", parentDisplay, group.Description))
		}

		// Use provider-grouped rendering for inventory groups
		if grouped := r.formatGroupedResources(group.ParentPath); grouped != "" {
			sb.WriteString(grouped)
		} else {
			sb.WriteString(fmt.Sprintf("    Resources: %s\n", strings.Join(group.Children, ", ")))
		}
	}

	// Clarify how positional args map to the flags object (MCP-specific concern)
	sb.WriteString("\nPositional args (shown as <arg> or [arg]) are passed as named entries in flags (e.g. flags: {name: \"my-plan\"}).\n")

	// Representative examples derived from source command data
	examples := r.pickRepresentativeExamples(r.ReadOnly, r.ReadOnlyOrder, 6)
	if len(examples) > 0 {
		sb.WriteString("\nExamples:\n")
		for _, ex := range examples {
			sb.WriteString(fmt.Sprintf("  %s\n", ex))
		}
	}

	sb.WriteString(r.formatGlobalFlags())

	sb.WriteString("\nDefault output is table. For structured data, use flags: {output: \"json\"} and fields: [\"name\", \"status\"] to limit response size.\n")
	sb.WriteString("The fields parameter is a top-level array (not inside flags) that filters JSON output to only the listed keys.\n")
	sb.WriteString("Use mtv_help for flags, TSL (Tree Search Language) query syntax for filtering inventory, and examples.\n")
	sb.WriteString("IMPORTANT: Before writing inventory queries, call mtv_help(\"tsl\") to learn available fields per provider and query syntax.\n")
	return sb.String()
}

// GenerateMinimalReadWriteDescription generates a minimal description for the read-write tool.
// It puts the command list first (most critical for the LM), followed by examples
// and a hint to use mtv_help. Domain context and conventions are omitted to stay
// within client description limits and avoid noise for smaller models.
func (r *Registry) GenerateMinimalReadWriteDescription() string {
	// Detect bare parent commands to skip them from the listing.
	// Admin commands are already filtered out by NewRegistry.
	bareParents := detectBareParents(r.ReadWrite)

	var sb strings.Builder

	sb.WriteString("Create, modify, or delete MTV resources (write operations).\n")
	sb.WriteString("\nCommands:\n")

	commands := r.ListReadWriteCommands()
	for _, key := range commands {
		if bareParents[key] {
			continue
		}
		cmd := r.ReadWrite[key]
		usage := formatUsageShort(cmd)
		sb.WriteString(fmt.Sprintf("  %s - %s\n", usage, cmd.Description))
	}

	// Clarify how positional args map to the flags object (MCP-specific concern)
	sb.WriteString("\nPositional args (shown as <arg> or [arg]) are passed as named entries in flags (e.g. flags: {name: \"my-plan\"}).\n")

	// Representative examples derived from source command data.
	// Use 8 slots to cover the most common verb groups (archive, cancel, create,
	// cutover, delete, patch, settings, start) — 6 was too few and omitted
	// frequently used operations like "start plan".
	examples := r.pickRepresentativeExamples(r.ReadWrite, r.ReadWriteOrder, 8)
	if len(examples) > 0 {
		sb.WriteString("\nExamples:\n")
		for _, ex := range examples {
			sb.WriteString(fmt.Sprintf("  %s\n", ex))
		}
	}

	sb.WriteString(r.formatGlobalFlags())

	sb.WriteString("\nCall mtv_help before create/patch to learn required flags, TSL (Tree Search Language) query syntax, and KARL (affinity/anti-affinity rule) syntax.\n")
	return sb.String()
}

// generateFlagReference builds a concise per-command flag reference for write commands.
// It includes all flags for commands that have required flags or many flags (complex commands),
// so the LLM can construct valid calls without guessing flag names.
func (r *Registry) generateFlagReference() string {
	var sb strings.Builder

	// Collect commands that need flag documentation:
	// 1. Commands with any required flags (these fail 100% without flag knowledge)
	// 2. Key complex commands (create/patch provider, create plan, create mapping)
	type commandEntry struct {
		pathKey string
		cmd     *Command
	}

	// Get sorted list of write commands
	keys := r.ListReadWriteCommands()

	// First pass: identify commands with required flags or many flags
	var flaggedCommands []commandEntry
	for _, key := range keys {
		cmd := r.ReadWrite[key]
		if cmd == nil || len(cmd.Flags) == 0 {
			continue
		}

		hasRequired := false
		for _, f := range cmd.Flags {
			if f.Required {
				hasRequired = true
				break
			}
		}

		// Include if: has required flags, or is a complex command (>5 flags)
		if hasRequired || len(cmd.Flags) > 5 {
			flaggedCommands = append(flaggedCommands, commandEntry{key, cmd})
		}
	}

	if len(flaggedCommands) == 0 {
		return ""
	}

	sb.WriteString("\nFlag reference for complex commands:\n")

	for _, entry := range flaggedCommands {
		cmd := entry.cmd
		cmdPath := strings.ReplaceAll(entry.pathKey, "/", " ")

		// Include the command's LongDescription when available, as it may contain
		// syntax references (e.g., KARL affinity syntax, TSL query language)
		if cmd.LongDescription != "" {
			sb.WriteString(fmt.Sprintf("\n%s notes:\n%s\n", cmdPath, cmd.LongDescription))
		}

		sb.WriteString(fmt.Sprintf("\n%s flags:\n", cmdPath))

		for _, f := range cmd.Flags {
			if f.Hidden {
				continue
			}

			// Format: "  --name (type) - description [REQUIRED] [enum: a|b|c]"
			line := fmt.Sprintf("  --%s", f.Name)

			// Add type for non-bool flags
			if f.Type != "bool" {
				line += fmt.Sprintf(" (%s)", f.Type)
			}

			line += fmt.Sprintf(" - %s", f.Description)

			if f.Required {
				line += " [REQUIRED]"
			}

			if len(f.Enum) > 0 {
				line += fmt.Sprintf(" [enum: %s]", strings.Join(f.Enum, "|"))
			}

			sb.WriteString(line + "\n")
		}
	}

	return sb.String()
}

// generateReadOnlyCommandNotes includes LongDescription from read-only commands
// that have substantial documentation (e.g., query language syntax references).
// This surfaces documentation that was added to Cobra Long descriptions into the
// MCP tool description, so AI clients can discover syntax without external docs.
func (r *Registry) generateReadOnlyCommandNotes() string {
	var sb strings.Builder

	// Minimum length threshold to avoid including trivial one-liner descriptions
	const minLongDescLength = 200

	commands := r.ListReadOnlyCommands()
	var hasNotes bool

	for _, key := range commands {
		cmd := r.ReadOnly[key]
		if cmd == nil || len(cmd.LongDescription) < minLongDescLength {
			continue
		}

		if !hasNotes {
			sb.WriteString("\nCommand notes:\n")
			hasNotes = true
		}

		cmdPath := strings.ReplaceAll(key, "/", " ")
		sb.WriteString(fmt.Sprintf("\n%s:\n%s\n", cmdPath, cmd.LongDescription))
	}

	return sb.String()
}

// pickRepresentativeExamples selects up to maxExamples from the command set,
// one per verb group. Commands are iterated in their original registration
// order (from help --machine). Within each group, the command with the most
// CLI examples is selected — more examples means the developer documented it
// more thoroughly, making it the most representative command for that group.
//
// Commands are grouped by root verb (e.g. "get", "create"), except inventory
// commands which form their own "get/inventory" group.
// Inventory gets 2 examples (to show query patterns like len() and any()).
func (r *Registry) pickRepresentativeExamples(commands map[string]*Command, orderedKeys []string, maxExamples int) []string {
	// Collect the command with the most examples per verb group.
	// Inventory commands (depth >= 3 with "inventory") get their own group
	// so they aren't overshadowed by simpler same-verb commands like get/plan.
	bestPerGroup := make(map[string]*Command)
	var groupOrder []string // preserve encounter order of groups

	for _, key := range orderedKeys {
		cmd := commands[key]
		if cmd == nil || len(cmd.Examples) == 0 {
			continue
		}

		// Group key: root verb for most commands, "get/inventory" for inventory
		parts := strings.Split(key, "/")
		groupKey := parts[0]
		if len(parts) >= 3 && parts[1] == "inventory" {
			groupKey = parts[0] + "/" + parts[1]
		}

		existing := bestPerGroup[groupKey]
		if existing == nil {
			bestPerGroup[groupKey] = cmd
			groupOrder = append(groupOrder, groupKey)
		} else if len(cmd.Examples) > len(existing.Examples) {
			// More examples = more thoroughly documented = better representative
			bestPerGroup[groupKey] = cmd
		}
	}

	var examples []string
	for _, group := range groupOrder {
		if len(examples) >= maxExamples {
			break
		}
		cmd := bestPerGroup[group]

		// Inventory gets 2 examples (to show query patterns); others get 1.
		perCmd := 1
		if strings.Contains(group, "inventory") {
			perCmd = 2
		}

		mcpExamples := convertCLIToMCPExamples(cmd, perCmd)
		for _, ex := range mcpExamples {
			if len(examples) >= maxExamples {
				break
			}
			examples = append(examples, ex)
		}
	}
	return examples
}

// sensitiveFlags lists flag names whose values should not appear in MCP examples.
// These are replaced with a placeholder to avoid leaking credentials.
var sensitiveFlags = map[string]bool{
	"password": true, "token": true, "secret": true, "secret-name": true,
}

// convertCLIToMCPExamples converts up to n CLI examples of a command into
// MCP-style call format strings. CLI commands should define their most
// instructive examples first — this is the source of truth for MCP descriptions.
// Duplicate MCP call strings (e.g., examples that differ only in description)
// are deduplicated.
func convertCLIToMCPExamples(cmd *Command, n int) []string {
	if len(cmd.Examples) == 0 || n <= 0 {
		return nil
	}

	pathString := cmd.CommandPath()
	seen := make(map[string]bool)
	var results []string

	for _, ex := range cmd.Examples {
		if len(results) >= n {
			break
		}
		mcpCall := formatCLIAsMCP(ex.Command, pathString, cmd.PositionalArgs)
		if mcpCall != "" && !seen[mcpCall] {
			results = append(results, mcpCall)
			seen[mcpCall] = true
		}
	}
	return results
}

// formatCLIAsMCP parses a single CLI command string and converts it to
// MCP-style call format: {command: "...", flags: {...}}.
// It strips the CLI prefix, extracts positional args (mapped to flag names),
// and collects --flag value pairs.
//
// Examples:
//
//	"kubectl-mtv get plan" → {command: "get plan"}
//	"kubectl-mtv get inventory vm vsphere-prod -q \"where len(nics) >= 2\"" →
//	  {command: "get inventory vm", flags: {provider: "vsphere-prod", query: "where len(nics) >= 2"}}
func formatCLIAsMCP(cliCmd string, pathString string, posArgs []Arg) string {
	cliCmd = strings.TrimSpace(cliCmd)
	if cliCmd == "" {
		return ""
	}
	cliCmd = strings.TrimPrefix(cliCmd, "kubectl-mtv ")
	cliCmd = strings.TrimPrefix(cliCmd, "kubectl mtv ")

	// Strip the command path from the front to get args + flags
	rest := cliCmd
	pathParts := strings.Fields(pathString)
	for _, part := range pathParts {
		rest = strings.TrimSpace(rest)
		if strings.HasPrefix(rest, part+" ") {
			rest = rest[len(part)+1:]
		} else if rest == part {
			rest = ""
		}
	}
	rest = strings.TrimSpace(rest)

	if rest == "" {
		return fmt.Sprintf("{command: \"%s\"}", pathString)
	}

	tokens := shellTokenize(rest)
	flagMap := make(map[string]string)

	// Phase 1: extract positional args (tokens before any flag)
	posIdx := 0
	i := 0
	for ; i < len(tokens); i++ {
		if strings.HasPrefix(tokens[i], "-") {
			break // reached flags
		}
		if posIdx < len(posArgs) {
			argName := strings.ToLower(strings.TrimSuffix(posArgs[posIdx].Name, "..."))
			flagMap[argName] = tokens[i]
			posIdx++
		}
		// Extra positional tokens beyond declared args are ignored
	}

	// Phase 2: extract --flag value pairs
	for ; i < len(tokens); i++ {
		tok := tokens[i]
		if !strings.HasPrefix(tok, "-") {
			continue // skip stray tokens
		}
		// Strip leading dashes and handle --flag=value syntax
		flagName := strings.TrimLeft(tok, "-")
		var flagValue string
		if eqIdx := strings.Index(flagName, "="); eqIdx >= 0 {
			flagValue = flagName[eqIdx+1:]
			flagName = flagName[:eqIdx]
		} else if i+1 < len(tokens) && !strings.HasPrefix(tokens[i+1], "-") {
			i++
			flagValue = tokens[i]
		} else {
			// Boolean flag (no value)
			flagValue = "true"
		}

		// Skip sensitive flags to avoid credentials in descriptions
		if sensitiveFlags[flagName] {
			continue
		}

		// Strip surrounding quotes from values
		flagValue = strings.Trim(flagValue, "'\"")

		// Use underscores for MCP JSON convention
		displayName := strings.ReplaceAll(flagName, "-", "_")
		flagMap[displayName] = flagValue
	}

	if len(flagMap) == 0 {
		return fmt.Sprintf("{command: \"%s\"}", pathString)
	}

	var flagParts []string
	for k, v := range flagMap {
		if v == "true" {
			flagParts = append(flagParts, fmt.Sprintf("%s: true", k))
		} else {
			flagParts = append(flagParts, fmt.Sprintf("%s: \"%s\"", k, v))
		}
	}
	// Sort for deterministic output
	sort.Strings(flagParts)
	return fmt.Sprintf("{command: \"%s\", flags: {%s}}", pathString, strings.Join(flagParts, ", "))
}

// uniqueRootVerbs extracts the unique first path element from a set of commands
// and returns them sorted. For example, commands "get/plan", "get/provider",
// shellTokenize splits a string into tokens, respecting double-quoted and
// single-quoted substrings so that values like "VM Network:default" remain
// a single token. Quotes are stripped from the returned tokens.
func shellTokenize(s string) []string {
	var tokens []string
	var current strings.Builder
	inDouble := false
	inSingle := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch == '"' && !inSingle:
			inDouble = !inDouble
		case ch == '\'' && !inDouble:
			inSingle = !inSingle
		case ch == ' ' && !inDouble && !inSingle:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

// "describe/plan", "health" produce ["describe", "get", "health"].
func uniqueRootVerbs(commands map[string]*Command) []string {
	seen := make(map[string]bool)
	for key := range commands {
		parts := strings.SplitN(key, "/", 2)
		seen[parts[0]] = true
	}

	roots := make([]string, 0, len(seen))
	for root := range seen {
		roots = append(roots, root)
	}
	sort.Strings(roots)
	return roots
}

// formatGlobalFlags returns a formatted string of common flags derived from the
// registry's GlobalFlags data. Only flags with LLMRelevant=true (set via the
// "llm-relevant" pflag annotation at the CLI source) are included.
// Extra flag names can be passed to include additional flags beyond the annotated ones.
func (r *Registry) formatGlobalFlags(extraNames ...string) string {
	// Build a set of extra names for fast lookup
	extras := make(map[string]bool, len(extraNames))
	for _, name := range extraNames {
		extras[name] = true
	}

	var sb strings.Builder
	sb.WriteString("\nCommon flags:\n")

	found := false
	for _, f := range r.GlobalFlags {
		if !f.LLMRelevant && !extras[f.Name] {
			continue
		}
		found = true
		// Display flag names with underscores (MCP JSON convention) instead of
		// hyphens (CLI convention). The MCP tools accept both forms at runtime,
		// but showing underscores matches the jsonschema tags and avoids confusion
		// for LLMs constructing JSON flag objects.
		displayName := strings.ReplaceAll(f.Name, "-", "_")
		sb.WriteString(fmt.Sprintf("- %s: %s\n", displayName, f.Description))
	}

	if !found {
		return ""
	}
	return sb.String()
}

// providerDisplayOrder defines the order in which provider categories are shown.
// "common" is a pseudo-provider for resources available across all providers.
var providerDisplayOrder = []string{"common", "vsphere", "ovirt", "openstack", "openshift", "ec2"}

// providerDisplayNames maps provider keys to human-readable labels.
var providerDisplayNames = map[string]string{
	"common":    "Resources",
	"vsphere":   "vSphere",
	"ovirt":     "oVirt",
	"openstack": "OpenStack",
	"openshift": "OpenShift",
	"ec2":       "AWS",
}

// formatGroupedResources renders inventory-style sibling groups with resources
// categorized by provider type. Returns empty string if the group's commands
// don't have provider metadata (falls back to flat list).
func (r *Registry) formatGroupedResources(parentPath string) string {
	// Collect commands under this parent
	prefix := parentPath + "/"
	var children []*Command
	for key, cmd := range r.ReadOnly {
		if strings.HasPrefix(key, prefix) {
			children = append(children, cmd)
		}
	}
	if len(children) == 0 {
		return ""
	}

	// Categorize resources by their exclusive provider.
	// A resource is "common" if it has no provider restriction (empty Providers list).
	// A resource is exclusive to a provider if it has exactly one provider.
	// Resources shared by 2+ providers go into the first matching provider in display order.
	categories := make(map[string][]string)
	hasAnyProviders := false

	// Sort children by name for deterministic output
	sort.Slice(children, func(i, j int) bool {
		return children[i].Name < children[j].Name
	})

	for _, cmd := range children {
		if len(cmd.Providers) == 0 {
			categories["common"] = append(categories["common"], cmd.Name)
		} else {
			hasAnyProviders = true
			if len(cmd.Providers) == 1 {
				categories[cmd.Providers[0]] = append(categories[cmd.Providers[0]], cmd.Name)
			} else {
				// Multi-provider: place in the first matching provider in display order
				placed := false
				for _, prov := range providerDisplayOrder {
					if prov == "common" {
						continue
					}
					for _, cp := range cmd.Providers {
						if cp == prov {
							categories[prov] = append(categories[prov], cmd.Name)
							placed = true
							break
						}
					}
					if placed {
						break
					}
				}
				if !placed {
					categories["common"] = append(categories["common"], cmd.Name)
				}
			}
		}
	}

	// If no commands have provider metadata, fall back to flat rendering
	if !hasAnyProviders {
		return ""
	}

	var sb strings.Builder
	for _, prov := range providerDisplayOrder {
		names, ok := categories[prov]
		if !ok || len(names) == 0 {
			continue
		}
		label := providerDisplayNames[prov]
		if label == "" {
			label = prov
		}
		sb.WriteString(fmt.Sprintf("    %s: %s\n", label, strings.Join(names, ", ")))
	}
	return sb.String()
}

// siblingGroup represents a group of commands that share a common parent path.
// Used to compact many sibling commands (e.g., get/inventory/*) into a single
// summary line in tool descriptions.
type siblingGroup struct {
	// ParentPath is the common parent path (e.g., "get/inventory")
	ParentPath string
	// Children are the child command names (e.g., ["vm", "network", "cluster", ...])
	Children []string
	// SharedRequiredArg is the positional arg shared by all children, if any (e.g., "PROVIDER")
	SharedRequiredArg string
	// Description is taken from the first child command
	Description string
}

// detectSiblingGroups finds groups of commands that share a common parent path
// with at least minGroupSize siblings. When a non-runnable parent command exists
// in the parents map, its description is used for the group. Otherwise, the first
// child's description is used as a fallback.
// Returns the groups and a set of command keys that belong to a group.
func detectSiblingGroups(commands map[string]*Command, parents map[string]*Command, minGroupSize int) ([]siblingGroup, map[string]bool) {
	// Count children per parent path
	parentChildren := make(map[string][]*Command)
	parentChildKeys := make(map[string][]string)

	for key, cmd := range commands {
		parts := strings.Split(key, "/")
		if len(parts) < 3 {
			// Only group commands at depth ≥ 3 (parent has ≥ 2 segments).
			// This prevents compacting heterogeneous top-level groups like get/*
			// or describe/*, while still compacting deeper homogeneous groups
			// like get/inventory/* where all children are structurally similar.
			continue
		}
		parentPath := strings.Join(parts[:len(parts)-1], "/")
		parentChildren[parentPath] = append(parentChildren[parentPath], cmd)
		parentChildKeys[parentPath] = append(parentChildKeys[parentPath], key)
	}

	var groups []siblingGroup
	groupedKeys := make(map[string]bool)

	// Sort parent paths for deterministic output
	var parentPaths []string
	for p := range parentChildren {
		parentPaths = append(parentPaths, p)
	}
	sort.Strings(parentPaths)

	for _, parentPath := range parentPaths {
		children := parentChildren[parentPath]
		if len(children) < minGroupSize {
			continue
		}

		// Determine the shared required positional arg by majority vote.
		// If ≥ 2/3 of children share the same required first positional arg,
		// treat it as the shared arg. This handles outliers like
		// "get inventory provider [PROVIDER_NAME]" among many PROVIDER-required siblings.
		argCounts := make(map[string]int)
		for _, child := range children {
			if len(child.PositionalArgs) > 0 && child.PositionalArgs[0].Required {
				argCounts[child.PositionalArgs[0].Name]++
			}
		}
		sharedArg := ""
		for argName, count := range argCounts {
			if count*3 >= len(children)*2 { // ≥ 2/3 supermajority
				sharedArg = argName
				break
			}
		}

		// Collect child names (last path element), excluding the shared arg
		sort.Slice(children, func(i, j int) bool {
			return children[i].Name < children[j].Name
		})
		var childNames []string
		for _, child := range children {
			name := child.Name
			// Build positional args string excluding the shared arg
			var extraArgs []string
			for _, arg := range child.PositionalArgs {
				if sharedArg != "" && arg.Name == sharedArg && arg.Required {
					continue // Skip the shared arg
				}
				// Show arg names in lowercase to match the flag key the LLM should use
				argLower := strings.ToLower(arg.Name)
				if arg.Required {
					extraArgs = append(extraArgs, argLower)
				} else {
					extraArgs = append(extraArgs, "["+argLower+"]")
				}
			}
			if len(extraArgs) > 0 {
				name += " " + strings.Join(extraArgs, " ")
			}
			childNames = append(childNames, name)
		}

		// Use the parent command's description if available (e.g., "Get inventory resources"),
		// otherwise fall back to the first child's description.
		desc := children[0].Description
		if parents != nil {
			if parent, ok := parents[parentPath]; ok && parent.Description != "" {
				desc = parent.Description
			}
		}
		groups = append(groups, siblingGroup{
			ParentPath:        parentPath,
			Children:          childNames,
			SharedRequiredArg: sharedArg,
			Description:       desc,
		})

		// Mark all child keys as grouped
		for _, key := range parentChildKeys[parentPath] {
			groupedKeys[key] = true
		}
	}

	return groups, groupedKeys
}

// detectBareParents finds commands that are structural grouping nodes rather than
// real commands. A bare parent is a command whose path is a proper prefix of another
// command's path AND that has no positional args AND no command-specific flags.
// Examples: "patch" (parent of patch/plan, patch/mapping/*),
//
//	"patch/mapping" (parent of patch/mapping/network, patch/mapping/storage).
func detectBareParents(commands map[string]*Command) map[string]bool {
	bareParents := make(map[string]bool)

	// Collect all path keys
	keys := make([]string, 0, len(commands))
	for k := range commands {
		keys = append(keys, k)
	}

	for _, key := range keys {
		cmd := commands[key]
		// Must have no positional args and no command-specific flags
		if len(cmd.PositionalArgs) > 0 || len(cmd.Flags) > 0 {
			continue
		}

		// Check if this path is a proper prefix of any other command's path
		prefix := key + "/"
		for _, otherKey := range keys {
			if strings.HasPrefix(otherKey, prefix) {
				bareParents[key] = true
				break
			}
		}
	}

	return bareParents
}

// formatUsageShort returns a short usage string for a command.
// Positional args are shown in lowercase to match the flag key the LLM should use.
// Required args use <angle brackets>, optional use [square brackets].
// Example: "get plan [name]" or "create provider <name>"
func formatUsageShort(cmd *Command) string {
	path := cmd.CommandPath()
	if len(cmd.PositionalArgs) == 0 {
		return path
	}

	var args []string
	for _, arg := range cmd.PositionalArgs {
		name := strings.ToLower(arg.Name)
		if arg.Variadic {
			name += "..."
		}
		if arg.Required {
			name = "<" + name + ">"
		} else {
			name = "[" + name + "]"
		}
		args = append(args, name)
	}
	return path + " " + strings.Join(args, " ")
}

// BuildCommandArgs builds command-line arguments from command path, args, and flags.
func BuildCommandArgs(cmdPath string, positionalArgs []string, flags map[string]string, namespace string, allNamespaces bool) []string {
	var args []string

	// Add command path
	parts := strings.Split(cmdPath, "/")
	args = append(args, parts...)

	// Add positional arguments
	args = append(args, positionalArgs...)

	// Add namespace flags
	if allNamespaces {
		args = append(args, "-A")
	} else if namespace != "" {
		args = append(args, "-n", namespace)
	}

	// Add other flags
	for name, value := range flags {
		if name == "namespace" || name == "all_namespaces" {
			continue // Already handled
		}
		if value == "true" {
			// Boolean flag
			args = append(args, "--"+name)
		} else if value == "false" {
			// Skip false boolean flags
			continue
		} else if value != "" {
			// String/int flag with value
			args = append(args, "--"+name, value)
		}
	}

	return args
}
