package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/pkg/mcp/util"
)

// KubectlDebugInput represents the input for the kubectl_debug tool.
type KubectlDebugInput struct {
	Action string `json:"action" jsonschema:"logs | get | describe | events"`

	ResourceType string `json:"resource_type,omitempty" jsonschema:"Resource type (pods, pvc, datavolume, virtualmachine, events, jobs, configmaps, deployments, services, or any k8s resource)"`

	Name string `json:"name,omitempty" jsonschema:"Resource name. Logs: prefer deployments/name for stable names (e.g. deployments/forklift-controller) since pod names have random suffixes. get/describe: optional."`

	Namespace string `json:"namespace,omitempty" jsonschema:"Kubernetes namespace"`

	AllNamespaces bool `json:"all_namespaces,omitempty" jsonschema:"Query all namespaces"`

	Labels string `json:"labels,omitempty" jsonschema:"Label selector (e.g. plan=my-plan,vmID=vm-123)"`

	Container string `json:"container,omitempty" jsonschema:"Container (multi-container pods)"`

	Previous bool `json:"previous,omitempty" jsonschema:"Logs from crashed container"`

	TailLines int `json:"tail_lines,omitempty" jsonschema:"Log lines from end (default 500, -1 for all)"`

	Since string `json:"since,omitempty" jsonschema:"Logs newer than (e.g. 1h, 30m)"`

	Output string `json:"output,omitempty" jsonschema:"Output: json, yaml, wide, name"`

	DryRun bool `json:"dry_run,omitempty" jsonschema:"Preview without executing"`

	FieldSelector string `json:"field_selector,omitempty" jsonschema:"Events field filter (e.g. type=Warning, reason=FailedScheduling)"`

	SortBy string `json:"sort_by,omitempty" jsonschema:"Sort events by JSONPath (e.g. .lastTimestamp)"`

	ForResource string `json:"for_resource,omitempty" jsonschema:"Events for resource (e.g. pod/my-pod, pvc/my-pvc)"`

	Grep string `json:"grep,omitempty" jsonschema:"Filter logs by regex (e.g. error|warning)"`

	IgnoreCase bool `json:"ignore_case,omitempty" jsonschema:"Case-insensitive grep"`

	NoTimestamps bool `json:"no_timestamps,omitempty" jsonschema:"Disable timestamps"`

	LogFormat string `json:"log_format,omitempty" jsonschema:"Log format: json, text, pretty"`

	// filter_*: structured log filtering for forklift-controller JSON logs
	FilterPlan      string `json:"filter_plan,omitempty" jsonschema:"Filter by plan name"`
	FilterProvider  string `json:"filter_provider,omitempty" jsonschema:"Filter by provider name"`
	FilterVM        string `json:"filter_vm,omitempty" jsonschema:"Filter by VM name/ID"`
	FilterMigration string `json:"filter_migration,omitempty" jsonschema:"Filter by migration name"`
	FilterLevel     string `json:"filter_level,omitempty" jsonschema:"Filter by level: info, debug, error, warn"`
	FilterLogger    string `json:"filter_logger,omitempty" jsonschema:"Filter by logger: plan, provider, migration, networkMap, storageMap"`
}

// GetKubectlDebugTool returns the tool definition for kubectl debugging.
func GetKubectlDebugTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "kubectl_debug",
		Description: `Debug MTV migrations using standard kubectl commands.

This tool provides access to kubectl for debugging migration issues:

Actions:
- logs: Get pod logs (useful for forklift-controller, virt-v2v pods)
- get: List Kubernetes resources (pods, pvc, datavolume, virtualmachine)
- describe: Get detailed resource information
- events: Get Kubernetes events with specialized filtering for debugging

Common use cases:
- Get forklift controller logs: action="logs", name="deployments/forklift-controller", namespace="openshift-mtv"
- List migration pods: action="get", resource_type="pods", labels="plan=my-plan"
- Check PVC status: action="get", resource_type="pvc", labels="migration=xxx"
- Debug failed pod: action="logs", name="virt-v2v-xxx", previous=true

Log name formats:
- Use deployments/name for stable names (e.g. name="deployments/forklift-controller")
- Use pod names directly when targeting specific pods (e.g. name="virt-v2v-cold-xyz")

Events examples:
- Get events for a pod: action="events", for_resource="pod/virt-v2v-xxx", namespace="target-ns"
- Get warning events: action="events", field_selector="type=Warning", namespace="target-ns"
- Get events sorted by time: action="events", sort_by=".lastTimestamp", namespace="target-ns"
- Get events for failed scheduling: action="events", field_selector="reason=FailedScheduling"

Log filtering (for scanning large logs):
- Get error logs: action="logs", name="deployments/forklift-controller", grep="error|ERROR", tail_lines=1000
- Case-insensitive search: action="logs", name="deployments/forklift-controller", grep="warning", ignore_case=true
- Find migration issues: action="logs", name="virt-v2v-xxx", grep="disk|transfer|failed"

JSON log filtering (for forklift controller structured logs):
- Filter by plan: filter_plan="my-plan" to get logs for a specific plan
- Filter by provider: filter_provider="vsphere-provider" for provider logs
- Filter by VM: filter_vm="vm-123" for VM-specific logs
- Filter by migration: filter_migration="migration-abc" for migration logs
- Filter by level: filter_level="error" for only error logs
- Filter by logger: filter_logger="plan" for plan reconciliation logs

Example use cases:
- Debug plan execution: filter_plan="my-plan", filter_level="error"
- Track VM migration: filter_vm="web-server-01"
- Monitor provider: filter_provider="vmware-prod", tail_lines=100

JSON log response format:
The response shape changes depending on whether logs are JSON and which log_format is used:
- log_format="json" (default for JSON logs): response has "logs" array in data (data.logs) instead of "output". Each entry is a parsed object with fields: level, ts, logger, msg, and optional context: plan (object with name, namespace), provider (object with name, namespace), map (object with name, namespace), migration (object with name, namespace), vm, vmName, vmID, reQ.
- log_format="text": response has "output" string in data containing filtered raw JSONL lines.
- log_format="pretty": response has "output" string in data with human-readable formatted lines like "[LEVEL] timestamp logger: message context".
- Non-JSON logs (e.g., virt-v2v pods): response always has "output" string in data with raw text. JSON filters are ignored.
- If JSON filters are specified but logs are not JSON, a "warning" field is added to data.

Example raw forklift controller JSON log line:
{"level":"info","ts":"2026-02-05 10:45:52","logger":"plan|zw4bt","msg":"Reconcile started.","plan":{"name":"my-plan","namespace":"demo"}}

When log_format="json", this is parsed into data.logs as:
{"level":"info","ts":"2026-02-05 10:45:52","logger":"plan|zw4bt","msg":"Reconcile started.","plan":{"name":"my-plan","namespace":"demo"}}

JSON log auto-detection:
- The tool automatically detects if logs are in JSON format by examining the first log line
- JSON parsing is only applied when logs contain structured JSON entries (with "level" and "msg" fields)
- Non-JSON logs (e.g., virt-v2v output) are returned as raw text without parsing
- If JSON filters are specified but logs are not JSON, a warning is returned and filters are ignored

Default behavior:
- log_format=json by default for JSON logs (use log_format="text" for raw JSONL, log_format="pretty" for human-readable)
- Non-JSON logs are always returned as raw text in the "output" field
- timestamps=true by default (use no_timestamps=true to disable)
- tail_lines=500 by default (use tail_lines=-1 for all logs)

Tips:
- Use labels to filter resources related to specific migrations
- Use tail_lines to limit log output (default 500, use -1 for all)
- Use previous=true to get logs from crashed containers
- Use since to get recent logs (e.g., "1h" for last hour)
- Use for_resource to get events related to a specific pod or PVC
- Use grep with tail_lines to efficiently scan large log files
- Combine JSON filters with grep for complex queries`,
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command":      map[string]any{"type": "string", "description": "The executed command"},
				"return_value": map[string]any{"type": "integer", "description": "Exit code (0 = success)"},
				"data": map[string]any{
					"type":        "object",
					"description": "Structured response data. Contains different fields depending on log format.",
					"properties": map[string]any{
						"logs": map[string]any{
							"type":        "array",
							"description": "Array of parsed JSON log entries (present when log_format=json and logs are JSON). Each entry has: level, ts, logger, msg, and optional plan, provider, map, migration (objects with name/namespace), vm, vmName, vmID, reQ fields. Malformed lines appear as {raw: string}.",
						},
						"output": map[string]any{
							"type":        "string",
							"description": "Raw text output (present for non-JSON logs, or when log_format=text or log_format=pretty)",
						},
						"warning": map[string]any{
							"type":        "string",
							"description": "Warning message, e.g. when JSON filters are specified but logs are not in JSON format",
						},
					},
				},
				"output": map[string]any{"type": "string", "description": "Plain text output (when not JSON)"},
				"stderr": map[string]any{"type": "string", "description": "Error output if any"},
			},
		},
	}
}

// GetMinimalKubectlDebugTool returns a minimal tool definition for kubectl debugging.
// The input schema (jsonschema tags on KubectlDebugInput) already describes each parameter.
// The description only needs to list valid action values since action is a free-form string.
func GetMinimalKubectlDebugTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "kubectl_debug",
		Description: `Debug MTV migrations via kubectl.
Actions: logs, get, describe, events.
Logs: name (required), container, previous, tail_lines, since, grep, ignore_case, no_timestamps, log_format, filter_*.
Get/describe: resource_type (required), name, labels, output.
Events: for_resource, field_selector, sort_by.
filter_* params apply to forklift-controller JSON logs only.
Logs tip: use deployments/name (e.g. deployments/forklift-controller) for stable names; pod names have random suffixes.
Get/describe tip: resource_type accepts any k8s resource (pods, pvc, datavolume, virtualmachine, jobs, configmaps, deployments, services, etc).`,
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command":      map[string]any{"type": "string", "description": "Executed command"},
				"return_value": map[string]any{"type": "integer", "description": "Exit code (0=success)"},
				"data":         map[string]any{"type": "object", "description": "Response data"},
				"output":       map[string]any{"type": "string", "description": "Text output"},
				"stderr":       map[string]any{"type": "string", "description": "Error output"},
			},
		},
	}
}

// HandleKubectlDebug handles the kubectl_debug tool invocation.
func HandleKubectlDebug(ctx context.Context, req *mcp.CallToolRequest, input KubectlDebugInput) (*mcp.CallToolResult, any, error) {
	// Extract K8s credentials from HTTP headers (for SSE mode)
	if req.Extra != nil && req.Extra.Header != nil {
		ctx = util.WithKubeCredsFromHeaders(ctx, req.Extra.Header)
	}

	// Enable dry run mode if requested
	if input.DryRun {
		ctx = util.WithDryRun(ctx, true)
	}

	var args []string

	switch input.Action {
	case "logs":
		// Logs action requires a resource name (pod name or resource/name like deployments/forklift-controller)
		if input.Name == "" {
			return nil, nil, fmt.Errorf("logs action requires 'name' field (e.g. 'deployments/forklift-controller' or a pod name)")
		}
		args = buildLogsArgs(input)
	case "get":
		// Get action requires a resource type
		if input.ResourceType == "" {
			return nil, nil, fmt.Errorf("get action requires 'resource_type' field (e.g., pods, pvc, events)")
		}
		args = buildGetArgs(input)
	case "describe":
		// Describe action requires a resource type
		if input.ResourceType == "" {
			return nil, nil, fmt.Errorf("describe action requires 'resource_type' field (e.g., pods, pvc, events)")
		}
		args = buildDescribeArgs(input)
	case "events":
		// Events action - specialized event querying
		args = buildEventsArgs(input)
	default:
		return nil, nil, fmt.Errorf("unknown action '%s'. Valid actions: logs, get, describe, events", input.Action)
	}

	// Execute kubectl command
	result, err := util.RunKubectlCommand(ctx, args)
	if err != nil {
		return nil, nil, fmt.Errorf("kubectl command failed: %w", err)
	}

	// Parse and return result
	data, err := util.UnmarshalJSONResponse(result)
	if err != nil {
		return nil, nil, err
	}

	// Process logs action with filtering and formatting
	if input.Action == "logs" {
		if output, ok := data["output"].(string); ok {
			// Apply grep filter first (regex pattern matching)
			if input.Grep != "" {
				filtered, err := filterLogsByPattern(output, input.Grep, input.IgnoreCase)
				if err != nil {
					return nil, nil, err
				}
				output = filtered
			}

			// Check if logs appear to be JSON formatted by inspecting the first line
			isJSONLogs := looksLikeJSONLogs(output)
			hasFilters := hasJSONFilters(input)

			// Warn if JSON filters are requested but logs don't appear to be JSON
			if hasFilters && !isJSONLogs {
				data["warning"] = "JSON filters were specified but logs do not appear to be in JSON format. Filters will be ignored."
			}

			if isJSONLogs {
				// Normalize LogFormat to a valid value before processing
				// Valid formats: "json", "text", "pretty"
				format := input.LogFormat
				switch format {
				case "json", "text", "pretty":
					// Valid format, use as-is
				case "":
					format = "json"
				default:
					// Invalid format specified, default to "json" and warn
					data["warning"] = fmt.Sprintf("Invalid log_format '%s' specified, defaulting to 'json'. Valid formats: json, text, pretty", format)
					format = "json"
				}

				// Update input with normalized format for filterAndFormatJSONLogs
				normalizedInput := input
				normalizedInput.LogFormat = format

				// Apply JSON filtering and formatting for JSON-formatted logs
				formatted, err := filterAndFormatJSONLogs(output, normalizedInput)
				if err != nil {
					return nil, nil, err
				}

				// Set the appropriate output field based on format
				if format == "json" {
					// For JSON format, put parsed entries in "logs" field
					delete(data, "output")
					data["logs"] = formatted
				} else {
					// For text/pretty formats, keep as "output" string
					// Both text and pretty formats return strings from filterAndFormatJSONLogs
					if str, ok := formatted.(string); ok {
						data["output"] = str
					} else {
						// Fallback: convert to JSON string if somehow not a string
						jsonBytes, _ := json.Marshal(formatted)
						data["output"] = string(jsonBytes)
					}
				}
			} else {
				// Non-JSON logs: return as raw text, skip JSON parsing entirely
				data["output"] = output
			}
		}
	}

	return nil, data, nil
}

// buildLogsArgs builds arguments for kubectl logs command.
func buildLogsArgs(input KubectlDebugInput) []string {
	args := []string{"logs"}

	// Resource name is required for logs (supports pod names or resource/name like deployments/forklift-controller)
	if input.Name != "" {
		args = append(args, input.Name)
	}

	// Namespace
	if input.Namespace != "" {
		args = append(args, "-n", input.Namespace)
	}

	// Container
	if input.Container != "" {
		args = append(args, "-c", input.Container)
	}

	// Previous container logs
	if input.Previous {
		args = append(args, "--previous")
	}

	// Tail lines - default to 500 if not specified
	// Use -1 to get all logs (no limit)
	if input.TailLines == 0 {
		// Default to 500 lines to prevent overwhelming output
		args = append(args, "--tail", "500")
	} else if input.TailLines > 0 {
		args = append(args, "--tail", strconv.Itoa(input.TailLines))
	}
	// If TailLines < 0 (e.g., -1), don't add --tail flag to get all logs

	// Since duration
	if input.Since != "" {
		args = append(args, "--since", input.Since)
	}

	// Timestamps enabled by default; use no_timestamps=true to disable
	if !input.NoTimestamps {
		args = append(args, "--timestamps")
	}

	return args
}

// buildGetArgs builds arguments for kubectl get command.
func buildGetArgs(input KubectlDebugInput) []string {
	args := []string{"get"}

	// Resource type
	if input.ResourceType != "" {
		args = append(args, input.ResourceType)
	}

	// Resource name (optional)
	if input.Name != "" {
		args = append(args, input.Name)
	}

	// Namespace
	if input.AllNamespaces {
		args = append(args, "-A")
	} else if input.Namespace != "" {
		args = append(args, "-n", input.Namespace)
	}

	// Label selector
	if input.Labels != "" {
		args = append(args, "-l", input.Labels)
	}

	// Output format - use configured default from MCP server
	output := input.Output
	if output == "" {
		output = util.GetOutputFormat()
	}
	// For "text" format, don't add -o flag to use default output
	if output != "text" {
		args = append(args, "-o", output)
	}

	return args
}

// buildDescribeArgs builds arguments for kubectl describe command.
func buildDescribeArgs(input KubectlDebugInput) []string {
	args := []string{"describe"}

	// Resource type
	if input.ResourceType != "" {
		args = append(args, input.ResourceType)
	}

	// Resource name (optional)
	if input.Name != "" {
		args = append(args, input.Name)
	}

	// Namespace
	if input.AllNamespaces {
		args = append(args, "-A")
	} else if input.Namespace != "" {
		args = append(args, "-n", input.Namespace)
	}

	// Label selector
	if input.Labels != "" {
		args = append(args, "-l", input.Labels)
	}

	return args
}

// buildEventsArgs builds arguments for kubectl get events command with specialized filtering.
func buildEventsArgs(input KubectlDebugInput) []string {
	args := []string{"get", "events"}

	// Namespace
	if input.AllNamespaces {
		args = append(args, "-A")
	} else if input.Namespace != "" {
		args = append(args, "-n", input.Namespace)
	}

	// For a specific resource (e.g., --for pod/my-pod)
	if input.ForResource != "" {
		args = append(args, "--for", input.ForResource)
	}

	// Field selector (e.g., involvedObject.name=my-pod, type=Warning)
	if input.FieldSelector != "" {
		args = append(args, "--field-selector", input.FieldSelector)
	}

	// Sort by (e.g., .lastTimestamp)
	if input.SortBy != "" {
		args = append(args, "--sort-by", input.SortBy)
	}

	// Output format - use configured default from MCP server
	output := input.Output
	if output == "" {
		output = util.GetOutputFormat()
	}
	// For "text" format, don't add -o flag to use default output
	if output != "text" {
		args = append(args, "-o", output)
	}

	return args
}

// filterLogsByPattern filters log lines by a regex pattern.
// If pattern is empty, returns the original logs unchanged.
// If ignoreCase is true, the pattern matching is case-insensitive.
func filterLogsByPattern(logs string, pattern string, ignoreCase bool) (string, error) {
	if pattern == "" {
		return logs, nil
	}

	flags := ""
	if ignoreCase {
		flags = "(?i)"
	}

	re, err := regexp.Compile(flags + pattern)
	if err != nil {
		return "", fmt.Errorf("invalid grep pattern: %w", err)
	}

	var filtered []string
	lines := strings.Split(logs, "\n")
	for _, line := range lines {
		if re.MatchString(line) {
			filtered = append(filtered, line)
		}
	}

	return strings.Join(filtered, "\n"), nil
}

// JSONLogEntry represents a structured log entry from forklift controller.
// The forklift controller outputs JSON logs with fields like:
// {"level":"info","ts":"2026-02-05 10:45:52","logger":"plan|zw4bt","msg":"Reconcile started.","plan":{"name":"my-plan","namespace":"demo"}}
type JSONLogEntry struct {
	Level     string            `json:"level"`
	Ts        string            `json:"ts"`
	Logger    string            `json:"logger"`
	Msg       string            `json:"msg"`
	Plan      map[string]string `json:"plan,omitempty"`
	Provider  map[string]string `json:"provider,omitempty"`
	Map       map[string]string `json:"map,omitempty"`
	Migration map[string]string `json:"migration,omitempty"`
	VM        string            `json:"vm,omitempty"`
	VMName    string            `json:"vmName,omitempty"`
	VMID      string            `json:"vmID,omitempty"`
	ReQ       int               `json:"reQ,omitempty"`
}

// RawLogLine represents a log line that could not be parsed as JSON.
// Used to preserve malformed or non-JSON log lines in the output.
type RawLogLine struct {
	Raw string `json:"raw"`
}

// hasJSONFilters returns true if any JSON-specific filters are set.
func hasJSONFilters(input KubectlDebugInput) bool {
	return input.FilterPlan != "" ||
		input.FilterProvider != "" ||
		input.FilterVM != "" ||
		input.FilterMigration != "" ||
		input.FilterLevel != "" ||
		input.FilterLogger != ""
}

// looksLikeJSONLogs checks if the logs appear to be in JSON format by examining up to 5 non-empty lines.
// It handles the kubectl --timestamps prefix (e.g., "2026-02-05T10:45:52.123Z {\"level\":...}")
// Returns true as soon as any scanned line contains valid JSON with expected log fields (level, msg).
// Returns false if none of the scanned lines yield a valid JSON entry.
func looksLikeJSONLogs(logs string) bool {
	lines := strings.Split(logs, "\n")

	// Check up to 5 non-empty lines for JSON format
	const maxLinesToCheck = 5
	checkedLines := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		checkedLines++
		if checkedLines > maxLinesToCheck {
			break
		}

		// Extract JSON part (skip timestamp prefix if present)
		idx := strings.Index(trimmed, "{")
		if idx < 0 {
			// No JSON object found in this line, try next
			continue
		}
		jsonPart := trimmed[idx:]

		// Try to parse as JSON
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(jsonPart), &entry); err != nil {
			// Not valid JSON, try next line
			continue
		}

		// Check for expected JSON log fields (forklift controller format)
		// A valid JSON log should have at least "level" and "msg" fields
		_, hasLevel := entry["level"]
		_, hasMsg := entry["msg"]

		if hasLevel && hasMsg {
			return true
		}
	}

	return false
}

// matchesJSONFilters checks if a log entry matches all specified filters.
func matchesJSONFilters(entry JSONLogEntry, input KubectlDebugInput) bool {
	// Filter by level
	if input.FilterLevel != "" && !strings.EqualFold(entry.Level, input.FilterLevel) {
		return false
	}

	// Filter by logger type (e.g., "plan" matches "plan|zw4bt")
	if input.FilterLogger != "" {
		loggerType := strings.Split(entry.Logger, "|")[0]
		if !strings.EqualFold(loggerType, input.FilterLogger) {
			return false
		}
	}

	// Filter by plan name
	if input.FilterPlan != "" {
		planName := ""
		if entry.Plan != nil {
			planName = entry.Plan["name"]
		}
		if !strings.EqualFold(planName, input.FilterPlan) {
			return false
		}
	}

	// Filter by provider name
	if input.FilterProvider != "" {
		providerName := ""
		if entry.Provider != nil {
			providerName = entry.Provider["name"]
		}
		if !strings.EqualFold(providerName, input.FilterProvider) {
			return false
		}
	}

	// Filter by VM name/ID
	if input.FilterVM != "" {
		vmMatch := strings.EqualFold(entry.VM, input.FilterVM) ||
			strings.EqualFold(entry.VMName, input.FilterVM) ||
			strings.EqualFold(entry.VMID, input.FilterVM)
		if !vmMatch {
			return false
		}
	}

	// Filter by migration name
	// First checks logger type is "migration", then compares migration name
	if input.FilterMigration != "" {
		loggerParts := strings.Split(entry.Logger, "|")
		loggerType := loggerParts[0]

		// Must be a migration logger
		if loggerType != "migration" {
			return false
		}

		// Try to match migration name from entry.Migration["name"] field first
		migrationName := ""
		if entry.Migration != nil {
			migrationName = entry.Migration["name"]
		}

		// If no Migration field, try extracting from logger ID (e.g., "migration|my-migration-name")
		if migrationName == "" && len(loggerParts) > 1 {
			migrationName = loggerParts[1]
		}

		if !strings.EqualFold(migrationName, input.FilterMigration) {
			return false
		}
	}

	return true
}

// filterAndFormatJSONLogs parses JSON logs, applies filters, and formats output.
// It returns the processed logs based on the specified format:
// - "json": Array of mixed JSONLogEntry and RawLogLine (for malformed lines)
// - "text": Original raw JSONL lines (filtered)
// - "pretty": Human-readable formatted output
func filterAndFormatJSONLogs(logs string, input KubectlDebugInput) (interface{}, error) {
	lines := strings.Split(strings.TrimSpace(logs), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return []interface{}{}, nil
	}

	var logLines []interface{}
	var filteredLines []string
	hasFilters := hasJSONFilters(input)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle timestamp prefix from kubectl --timestamps flag
		// Format: "2026-02-05T10:45:52.123456789Z {"level":"info",...}"
		// Split on the first '{' to avoid mis-parsing JSON that contains " {" inside message text
		jsonPart := line
		timestampPrefix := ""
		if idx := strings.Index(line, "{"); idx > 0 {
			timestampPrefix = line[:idx]
			jsonPart = line[idx:]
		}

		var entry JSONLogEntry
		if err := json.Unmarshal([]byte(jsonPart), &entry); err != nil {
			// Malformed line - preserve as RawLogLine
			if !hasFilters {
				logLines = append(logLines, RawLogLine{Raw: line})
				filteredLines = append(filteredLines, line)
			}
			continue
		}

		// Apply filters
		if hasFilters && !matchesJSONFilters(entry, input) {
			continue
		}

		logLines = append(logLines, entry)
		filteredLines = append(filteredLines, timestampPrefix+jsonPart)
	}

	// Determine output format (default to "json" for LLM consumption)
	format := input.LogFormat
	if format == "" {
		format = "json"
	}

	switch format {
	case "json":
		return logLines, nil
	case "text":
		return strings.Join(filteredLines, "\n"), nil
	case "pretty":
		return formatPrettyLogs(logLines), nil
	default:
		return logLines, nil
	}
}

// formatPrettyLogs formats log entries in a human-readable format.
// It handles both JSONLogEntry (parsed) and RawLogLine (malformed) types.
func formatPrettyLogs(logLines []interface{}) string {
	var lines []string
	for _, item := range logLines {
		switch v := item.(type) {
		case RawLogLine:
			// Include raw malformed lines as-is
			lines = append(lines, v.Raw)
		case JSONLogEntry:
			// Format: [LEVEL] timestamp logger: message (context)
			levelUpper := strings.ToUpper(v.Level)
			context := ""

			// Add context info
			if v.Plan != nil && v.Plan["name"] != "" {
				context = fmt.Sprintf(" plan=%s", v.Plan["name"])
				if ns := v.Plan["namespace"]; ns != "" {
					context += fmt.Sprintf("/%s", ns)
				}
			} else if v.Provider != nil && v.Provider["name"] != "" {
				context = fmt.Sprintf(" provider=%s", v.Provider["name"])
				if ns := v.Provider["namespace"]; ns != "" {
					context += fmt.Sprintf("/%s", ns)
				}
			} else if v.Map != nil && v.Map["name"] != "" {
				context = fmt.Sprintf(" map=%s", v.Map["name"])
				if ns := v.Map["namespace"]; ns != "" {
					context += fmt.Sprintf("/%s", ns)
				}
			}

			if v.VM != "" {
				context += fmt.Sprintf(" vm=%s", v.VM)
			} else if v.VMName != "" {
				context += fmt.Sprintf(" vm=%s", v.VMName)
			}

			line := fmt.Sprintf("[%s] %s %s: %s%s", levelUpper, v.Ts, v.Logger, v.Msg, context)
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
