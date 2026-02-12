package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// Registry holds discovered kubectl-mtv commands organized by read/write access.
type Registry struct {
	// ReadOnly contains commands that don't modify cluster state
	ReadOnly map[string]*Command

	// ReadWrite contains commands that modify cluster state
	ReadWrite map[string]*Command

	// GlobalFlags are flags available to all commands
	GlobalFlags []Flag

	// RootDescription is the main kubectl-mtv description
	RootDescription string
}

// NewRegistry creates a new registry by calling kubectl-mtv help --machine.
// This single call returns the complete command schema as JSON.
func NewRegistry(ctx context.Context) (*Registry, error) {
	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "kubectl-mtv", "help", "--machine")
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
		GlobalFlags:     schema.GlobalFlags,
		RootDescription: schema.Description,
	}

	// Categorize commands by read/write based on category field
	for i := range schema.Commands {
		cmd := &schema.Commands[i]
		pathKey := cmd.PathKey()

		switch cmd.Category {
		case "read":
			registry.ReadOnly[pathKey] = cmd
		default:
			// "write" and "admin" categories go to ReadWrite
			registry.ReadWrite[pathKey] = cmd
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
	var sb strings.Builder
	sb.WriteString("Execute read-only kubectl-mtv commands to query MTV resources.\n\n")
	sb.WriteString("USE THIS TOOL FOR: health checks, getting resources, describing resources, viewing settings.\n")
	sb.WriteString("Commands: health, get, describe, settings get\n\n")
	sb.WriteString("Available commands:\n")

	commands := r.ListReadOnlyCommands()
	for _, key := range commands {
		cmd := r.ReadOnly[key]
		// Format: "get plan [NAME]" -> "get plan [NAME] - Get migration plans"
		usage := formatUsageShort(cmd)
		sb.WriteString(fmt.Sprintf("- %s - %s\n", usage, cmd.Description))
	}

	sb.WriteString("\nCommon flags:\n")
	sb.WriteString("- namespace: Target Kubernetes namespace\n")
	sb.WriteString("- all_namespaces: Query across all namespaces\n")
	sb.WriteString("- output: Output format (table, json, yaml)\n")

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
	sb.WriteString("NOTE: For read-only operations (health, get, describe, settings get), use the mtv_read tool instead.\n\n")
	sb.WriteString("Available commands:\n")

	commands := r.ListReadWriteCommands()
	for _, key := range commands {
		cmd := r.ReadWrite[key]
		usage := formatUsageShort(cmd)
		sb.WriteString(fmt.Sprintf("- %s - %s\n", usage, cmd.Description))
	}

	sb.WriteString("\nCommon flags:\n")
	sb.WriteString("- namespace: Target Kubernetes namespace\n")

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
// It includes only the command list and a hint to use mtv_help for details.
// The input schema (jsonschema tags on MTVReadInput) already describes parameters.
func (r *Registry) GenerateMinimalReadOnlyDescription() string {
	var sb strings.Builder
	sb.WriteString("Query MTV resources (read-only).\n\n")
	sb.WriteString("Commands:\n")

	// Collect inventory resource names separately to compact them
	var inventoryResources []string
	commands := r.ListReadOnlyCommands()
	for _, key := range commands {
		cmd := r.ReadOnly[key]
		if strings.HasPrefix(key, "get/inventory/") {
			// Extract the resource name (last path element) and any positional args
			parts := strings.Split(key, "/")
			resource := parts[len(parts)-1]
			posArgs := cmd.PositionalArgsString()
			if posArgs != "" && posArgs != "PROVIDER" {
				// Special case like "provider [PROVIDER_NAME]"
				resource += " " + posArgs
			}
			inventoryResources = append(inventoryResources, resource)
			continue
		}
		usage := formatUsageShort(cmd)
		sb.WriteString(fmt.Sprintf("- %s - %s\n", usage, cmd.Description))
	}

	// Write compacted inventory block
	if len(inventoryResources) > 0 {
		sb.WriteString("- get inventory RESOURCE PROVIDER - Query provider inventory.\n")
		sb.WriteString(fmt.Sprintf("  Resources: %s\n", strings.Join(inventoryResources, ", ")))
	}

	sb.WriteString("\nDefault output is table. For structured data, use flags: {output: \"json\"} with fields to limit size.\n")
	sb.WriteString("Use mtv_help for flags, query syntax (TSL), and examples.\n")
	sb.WriteString("Flags support ${VAR} env refs (resolved at runtime).\n")

	return sb.String()
}

// GenerateMinimalReadWriteDescription generates a minimal description for the read-write tool.
// It includes only the command list and hints to use mtv_help for details.
// The input schema (jsonschema tags on MTVWriteInput) already describes parameters.
func (r *Registry) GenerateMinimalReadWriteDescription() string {
	// Commands irrelevant to LLM use: shell completions, meta commands, bare parent commands
	skipCommands := map[string]bool{
		"completion/bash": true, "completion/fish": true,
		"completion/powershell": true, "completion/zsh": true,
		"help": true, "mcp-server": true, "version": true,
		"patch": true, // bare parent; real commands are patch/plan, patch/mapping/*
	}

	var sb strings.Builder
	sb.WriteString("Create, modify, or delete MTV resources (write operations).\n\n")
	sb.WriteString("Commands:\n")

	commands := r.ListReadWriteCommands()
	for _, key := range commands {
		if skipCommands[key] {
			continue
		}
		cmd := r.ReadWrite[key]
		usage := formatUsageShort(cmd)
		sb.WriteString(fmt.Sprintf("- %s - %s\n", usage, cmd.Description))
	}

	sb.WriteString("\nCall mtv_help before create/patch to learn required flags, TSL query syntax, and KARL affinity rules.\n")
	sb.WriteString("Flags support ${VAR} env refs (e.g. url: \"${GOVC_URL}/sdk\"). Only ${VAR} format is recognized.\n")

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

// formatUsageShort returns a short usage string for a command.
// Example: "get plan [NAME]" or "create provider NAME"
func formatUsageShort(cmd *Command) string {
	path := cmd.CommandPath()
	positionalArgs := cmd.PositionalArgsString()
	if positionalArgs != "" {
		return path + " " + positionalArgs
	}
	return path
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
