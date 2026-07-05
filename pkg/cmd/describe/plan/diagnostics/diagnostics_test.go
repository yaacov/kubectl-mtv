package diagnostics

import (
	"testing"
)

func TestExtractVMErrors_NoErrors(t *testing.T) {
	vmStatus := map[string]interface{}{
		"phase": "Completed",
		"pipeline": []interface{}{
			map[string]interface{}{
				"name":  "Initialize",
				"phase": "Completed",
			},
		},
	}

	errStr, conditions, stepErrors := ExtractVMErrors(vmStatus)
	if errStr != "" {
		t.Errorf("expected empty error, got %q", errStr)
	}
	if len(conditions) != 0 {
		t.Errorf("expected no conditions, got %d", len(conditions))
	}
	if len(stepErrors) != 0 {
		t.Errorf("expected no step errors, got %d", len(stepErrors))
	}
}

func TestExtractVMErrors_WithStepError(t *testing.T) {
	vmStatus := map[string]interface{}{
		"phase": "CopyDisks",
		"pipeline": []interface{}{
			map[string]interface{}{
				"name":  "DiskTransfer",
				"phase": "Error",
				"error": map[string]interface{}{
					"reasons": []interface{}{"disk copy failed", "connection timeout"},
				},
			},
		},
	}

	_, _, stepErrors := ExtractVMErrors(vmStatus)
	if len(stepErrors) != 1 {
		t.Fatalf("expected 1 step error, got %d", len(stepErrors))
	}
	if stepErrors[0].Step != "DiskTransfer" {
		t.Errorf("expected step 'DiskTransfer', got %q", stepErrors[0].Step)
	}
	if stepErrors[0].Reason != "disk copy failed; connection timeout" {
		t.Errorf("unexpected reason: %q", stepErrors[0].Reason)
	}
	if stepErrors[0].Message != "disk copy failed; connection timeout" {
		t.Errorf("unexpected message: %q", stepErrors[0].Message)
	}
}

func TestExtractVMErrors_WithConditions(t *testing.T) {
	vmStatus := map[string]interface{}{
		"phase": "Failed",
		"conditions": []interface{}{
			map[string]interface{}{
				"type":    "Failed",
				"status":  "True",
				"reason":  "ConversionFailed",
				"message": "virt-v2v conversion failed",
			},
		},
	}

	_, conditions, _ := ExtractVMErrors(vmStatus)
	if len(conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(conditions))
	}
	if conditions[0].Type != "Failed" {
		t.Errorf("expected type 'Failed', got %q", conditions[0].Type)
	}
	if conditions[0].Reason != "ConversionFailed" {
		t.Errorf("expected reason 'ConversionFailed', got %q", conditions[0].Reason)
	}
}

func TestExtractVMErrors_ReadyConditionSkipped(t *testing.T) {
	vmStatus := map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{
				"type":   "Ready",
				"status": "True",
			},
		},
	}

	_, conditions, _ := ExtractVMErrors(vmStatus)
	if len(conditions) != 0 {
		t.Errorf("Ready=True condition should be skipped, got %d conditions", len(conditions))
	}
}

func TestExtractVMErrors_TaskLevelError(t *testing.T) {
	vmStatus := map[string]interface{}{
		"pipeline": []interface{}{
			map[string]interface{}{
				"name":  "ImageConversion",
				"phase": "Error",
				"tasks": []interface{}{
					map[string]interface{}{
						"name":  "convert-disks",
						"phase": "Error",
						"error": map[string]interface{}{
							"phase":   "ConvertGuest",
							"reasons": []interface{}{"virt-v2v exited with code 1"},
						},
					},
				},
			},
		},
	}

	_, _, stepErrors := ExtractVMErrors(vmStatus)
	if len(stepErrors) != 1 {
		t.Fatalf("expected 1 step error (task-level), got %d", len(stepErrors))
	}
	if stepErrors[0].Step != "ImageConversion/convert-disks" {
		t.Errorf("expected step 'ImageConversion/convert-disks', got %q", stepErrors[0].Step)
	}
	if stepErrors[0].Message != "virt-v2v exited with code 1" {
		t.Errorf("unexpected message: %q", stepErrors[0].Message)
	}
}

func TestIsRelevantEvent(t *testing.T) {
	tests := []struct {
		eventType string
		reason    string
		expected  bool
	}{
		{"Warning", "FailedScheduling", true},
		{"Warning", "BackOff", true},
		{"Normal", "Started", true},
		{"Normal", "Evicted", true},
		{"Normal", "ProvisioningFailed", true},
		{"Normal", "Pulling", false},
		{"Normal", "Pulled", false},
		{"Normal", "Created", false},
	}

	for _, tt := range tests {
		t.Run(tt.eventType+"/"+tt.reason, func(t *testing.T) {
			got := isRelevantEvent(tt.eventType, tt.reason)
			if got != tt.expected {
				t.Errorf("isRelevantEvent(%q, %q) = %v, want %v", tt.eventType, tt.reason, got, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"hello world this is long", 10, "hello w..."},
		{"exactly10!", 10, "exactly10!"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.expected)
			}
		})
	}
}

func TestIsErrorLine(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"I0705 07:11:44.732710 some info line", false},
		{"E0705 07:11:44.732710 error occurred", true},
		{"something failed to connect", true},
		{"fatal: cannot proceed", true},
		{"no error here", true},    // isErrorLine matches \berror\b; filtering is done by isIgnoredLine
		{"error count is 0", true}, // same: raw match; isIgnoredLine filters at collect time
		{"all is well no issues", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := isErrorLine(tt.line)
			if got != tt.expected {
				t.Errorf("isErrorLine(%q) = %v, want %v", tt.line, got, tt.expected)
			}
		})
	}
}

func TestIsIgnoredLine(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"no error here", true},
		{"error count is 0", true},
		{"error count: 0", true},
		{"error count is 10", false},
		{"error count is 20", false},
		{"nbdkit: debug: something happened", true},
		{"Cannot open file \"/etc/vmware/config\"", true},
		{"Failed to determine unit we run in", true},
		{"Failed to connect to bus: Host is down", true},
		{"you can ignore this message", true},
		{"virt-v2v: error: something bad", false},
		{"Error executing v2v command", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := isIgnoredLine(tt.line)
			if got != tt.expected {
				t.Errorf("isIgnoredLine(%q) = %v, want %v", tt.line, got, tt.expected)
			}
		})
	}
}

func TestIsRootCauseLine(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"virt-v2v: error: libguestfs error: inspect_os failed", true},
		{"guestfsd: error: mount exited with status 32", true},
		{"[   38.038333] sd 0:0:0:0: [sdb] tag#133 FAILED Result: hostbyte=DID_OK", true},
		{"I/O error, dev sdb, sector 0", true},
		{"unknown filesystem type 'btrfs'", true},
		{"Error executing v2v command: exit status 1", false},
		{"all is fine", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := isRootCauseLine(tt.line)
			if got != tt.expected {
				t.Errorf("isRootCauseLine(%q) = %v, want %v", tt.line, got, tt.expected)
			}
		})
	}
}

func TestIsWarnLine(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"this is a warning about something", true},
		{"WARN: disk space low", true},
		{"all is well", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := isWarnLine(tt.line)
			if got != tt.expected {
				t.Errorf("isWarnLine(%q) = %v, want %v", tt.line, got, tt.expected)
			}
		})
	}
}

func TestAppendIfNew(t *testing.T) {
	existing := []PodDiagnostics{
		{Name: "pod-1"},
		{Name: "pod-2"},
	}

	// Should not duplicate
	result := appendIfNew(existing, PodDiagnostics{Name: "pod-1"})
	if len(result) != 2 {
		t.Errorf("expected 2 pods (no duplicate), got %d", len(result))
	}

	// Should add new
	result = appendIfNew(existing, PodDiagnostics{Name: "pod-3"})
	if len(result) != 3 {
		t.Errorf("expected 3 pods (new added), got %d", len(result))
	}
}
