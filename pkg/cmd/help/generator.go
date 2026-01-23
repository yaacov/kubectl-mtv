package help

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// SchemaVersion is the current version of the help schema format.
const SchemaVersion = "1.0"

// EnumValuer is an interface for flags that provide valid values.
// Custom flag types can implement this to expose their allowed values.
type EnumValuer interface {
	GetValidValues() []string
}

// Generate creates a HelpSchema from a Cobra command tree.
func Generate(rootCmd *cobra.Command, cliVersion string, opts Options) *HelpSchema {
	schema := &HelpSchema{
		Version:     SchemaVersion,
		CLIVersion:  cliVersion,
		Name:        rootCmd.Name(),
		Description: rootCmd.Short,
		Commands:    []Command{},
		GlobalFlags: []Flag{},
	}

	// Walk command tree - automatically discovers all commands
	walkCommands(rootCmd, []string{}, func(cmd *cobra.Command, path []string) {
		// Skip non-runnable commands (groups without RunE)
		if !cmd.Runnable() {
			return
		}

		// Skip hidden commands unless requested
		if cmd.Hidden && !opts.IncludeHidden {
			return
		}

		// Apply category filter
		category := getCategory(path)
		if opts.ReadOnly && category != "read" {
			return
		}
		if opts.Write && category != "write" {
			return
		}

		schema.Commands = append(schema.Commands, commandToSchema(cmd, path, opts))
	})

	// Extract global flags from persistent flags
	if opts.IncludeGlobalFlags {
		rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			if f.Hidden && !opts.IncludeHidden {
				return
			}
			schema.GlobalFlags = append(schema.GlobalFlags, flagToSchema(f))
		})
	}

	return schema
}

// walkCommands recursively visits all commands in the tree.
func walkCommands(cmd *cobra.Command, path []string, visitor func(*cobra.Command, []string)) {
	visitor(cmd, path)
	for _, child := range cmd.Commands() {
		walkCommands(child, append(append([]string{}, path...), child.Name()), visitor)
	}
}

// commandToSchema converts a Cobra command to our schema format.
func commandToSchema(cmd *cobra.Command, path []string, opts Options) Command {
	c := Command{
		Name:            cmd.Name(),
		Path:            path,
		PathString:      strings.Join(path, " "),
		Description:     cmd.Short,
		LongDescription: cmd.Long,
		Usage:           cmd.UseLine(),
		Category:        getCategory(path),
		Flags:           []Flag{},
		PositionalArgs:  parsePositionalArgs(cmd.Use),
		Examples:        parseExamples(cmd.Example),
	}

	// Copy aliases
	if len(cmd.Aliases) > 0 {
		c.Aliases = append([]string{}, cmd.Aliases...)
	}

	// Extract local flags (not inherited)
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden && !opts.IncludeHidden {
			return
		}
		schema := flagToSchema(f)

		// Check if flag is required (MarkFlagRequired annotates the flag, not the command)
		if ann := f.Annotations; ann != nil {
			if _, ok := ann[cobra.BashCompOneRequiredFlag]; ok {
				schema.Required = true
			}
		}

		// Try to get enum values from the flag value
		if enumValuer, ok := f.Value.(EnumValuer); ok {
			schema.Enum = enumValuer.GetValidValues()
		}

		c.Flags = append(c.Flags, schema)
	})

	return c
}

// flagToSchema converts a pflag.Flag to our schema format.
func flagToSchema(f *pflag.Flag) Flag {
	flag := Flag{
		Name:        f.Name,
		Shorthand:   f.Shorthand,
		Type:        f.Value.Type(),
		Description: f.Usage,
		Hidden:      f.Hidden,
	}

	// Set default value with proper typing based on flag type
	if f.DefValue != "" {
		flagType := f.Value.Type()
		switch {
		case flagType == "bool":
			// Convert boolean strings to actual booleans
			if f.DefValue == "true" {
				flag.Default = true
			} else if f.DefValue == "false" {
				flag.Default = false
			}
		case flagType == "int" || flagType == "int8" || flagType == "int16" || flagType == "int32" || flagType == "int64":
			// Convert integer strings to numbers
			if v, err := strconv.ParseInt(f.DefValue, 10, 64); err == nil {
				flag.Default = v
			} else {
				flag.Default = f.DefValue
			}
		case flagType == "uint" || flagType == "uint8" || flagType == "uint16" || flagType == "uint32" || flagType == "uint64":
			// Convert unsigned integer strings to numbers
			if v, err := strconv.ParseUint(f.DefValue, 10, 64); err == nil {
				flag.Default = v
			} else {
				flag.Default = f.DefValue
			}
		case flagType == "float32" || flagType == "float64":
			// Convert float strings to numbers
			if v, err := strconv.ParseFloat(f.DefValue, 64); err == nil {
				flag.Default = v
			} else {
				flag.Default = f.DefValue
			}
		case strings.HasSuffix(flagType, "Slice") || strings.HasSuffix(flagType, "Array"):
			// Convert slice/array defaults to empty array or preserve value
			if f.DefValue == "[]" {
				flag.Default = []string{}
			} else {
				flag.Default = f.DefValue
			}
		default:
			// For all other types, preserve as string
			flag.Default = f.DefValue
		}
	}

	// Try to get enum values from the flag value
	if enumValuer, ok := f.Value.(EnumValuer); ok {
		flag.Enum = enumValuer.GetValidValues()
	}

	return flag
}

// getCategory determines the command category based on its path.
func getCategory(path []string) string {
	if len(path) == 0 {
		return "admin"
	}
	switch path[0] {
	case "get", "describe":
		return "read"
	case "create", "delete", "patch", "start", "cancel", "archive", "unarchive", "cutover":
		return "write"
	default:
		return "admin"
	}
}

// parsePositionalArgs extracts positional arguments from the usage string.
// Example: "plan [NAME]" -> [{Name: "NAME", Required: false}]
// Example: "provider NAME" -> [{Name: "NAME", Required: true}]
func parsePositionalArgs(usage string) []PositionalArg {
	var args []PositionalArg

	// Remove the command name (first word)
	parts := strings.Fields(usage)
	if len(parts) <= 1 {
		return args
	}

	for _, part := range parts[1:] {
		// Skip [flags] marker
		if part == "[flags]" {
			continue
		}

		// Check for variadic args (NAME...)
		variadic := strings.HasSuffix(part, "...")
		if variadic {
			part = strings.TrimSuffix(part, "...")
		}

		// Check for optional args [NAME]
		optional := strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]")
		if optional {
			part = strings.TrimPrefix(strings.TrimSuffix(part, "]"), "[")
		}

		// Skip if not an argument pattern (UPPERCASE or contains uppercase)
		if part != strings.ToUpper(part) {
			continue
		}

		args = append(args, PositionalArg{
			Name:     part,
			Required: !optional,
			Variadic: variadic,
		})
	}

	return args
}

// parseExamples parses Cobra-style examples into our format.
// Cobra examples are typically formatted as:
//
//	# Comment describing the example
//	command args
func parseExamples(exampleText string) []Example {
	if exampleText == "" {
		return nil
	}

	var examples []Example
	lines := strings.Split(exampleText, "\n")

	var currentDesc string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			// This is a description comment
			currentDesc = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		} else {
			// This is a command
			examples = append(examples, Example{
				Description: currentDesc,
				Command:     line,
			})
			currentDesc = ""
		}
	}

	return examples
}

// RequiredFlagAnnotation is the annotation key Cobra uses to mark required flags.
// This is used when checking if a flag is required via MarkFlagRequired.
var requiredFlagRegex = regexp.MustCompile(`required`)

// IsRequiredFlag checks if a flag is marked as required on a command.
func IsRequiredFlag(cmd *cobra.Command, flagName string) bool {
	f := cmd.Flag(flagName)
	if f == nil {
		return false
	}
	if f.Annotations == nil {
		return false
	}
	for key := range f.Annotations {
		if requiredFlagRegex.MatchString(key) {
			return true
		}
	}
	return false
}
