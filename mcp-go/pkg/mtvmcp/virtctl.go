package mtvmcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// VirtctlCommand is the command name to use for virtctl operations.
// Defaults to "virtctl" but can be overridden via VIRTCTL_COMMAND env var or SetVirtctlCommand().
var VirtctlCommand = getVirtctlCommand()

// getVirtctlCommand returns the virtctl command name from env var or default
func getVirtctlCommand() string {
	if cmd := os.Getenv("VIRTCTL_COMMAND"); cmd != "" {
		return cmd
	}
	return "virtctl"
}

// SetVirtctlCommand allows programmatic override of the virtctl command name
func SetVirtctlCommand(command string) {
	VirtctlCommand = command
}

// RunVirtctlCommand executes a virtctl/kubectl-virt command and returns the result
func RunVirtctlCommand(args []string) (string, error) {
	cmd := exec.Command(VirtctlCommand, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set timeout of 120 seconds
	timer := time.AfterFunc(120*time.Second, func() {
		_ = cmd.Process.Kill()
	})
	defer timer.Stop()

	err := cmd.Run()

	// For virtctl/kubectl-virt, we often want to return YAML output directly
	// Check if command succeeded and return stdout
	if err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("%s error: %s", VirtctlCommand, stderr.String())
		}
		return "", fmt.Errorf("%s command failed: %w", VirtctlCommand, err)
	}

	return stdout.String(), nil
}

// GetCurrentNamespace gets the current Kubernetes namespace from kubectl context
func GetCurrentNamespace() string {
	cmd := exec.Command("kubectl", "config", "view", "--minify", "--output", "jsonpath={..namespace}")
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return "default"
	}
	namespace := strings.TrimSpace(string(output))
	if namespace == "" {
		return "default"
	}
	return namespace
}

// ResolveNamespace resolves the namespace to use (provided or current)
func ResolveNamespace(namespace string) string {
	if namespace != "" {
		return namespace
	}
	return GetCurrentNamespace()
}

// ParseJSONOutput attempts to parse JSON output and return it formatted
func ParseJSONOutput(output string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		// If not JSON, return as-is
		return output, nil
	}
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return output, nil
	}
	return string(formatted), nil
}
