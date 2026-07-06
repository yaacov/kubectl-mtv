package diagnostics

import (
	"strings"
	"testing"

	"github.com/yaacov/kubectl-mtv/pkg/util/describe"
)

func TestRender_BasicReport(t *testing.T) {
	report := &DiagnosticsReport{
		PlanName:      "test-plan",
		PlanUID:       "uid-123",
		MigrationName: "test-plan-abc",
		MigrationUID:  "mig-456",
		TargetNS:      "target-ns",
		Config: ConfigContext{
			SourceProvider: "vsphere-provider",
			MigrationType:  "warm",
			VDDKImage:      "quay.io/test/vddk:8.0.0",
		},
		VMs: []VMDiagnostics{
			{
				Name:  "test-vm",
				ID:    "vm-100",
				Phase: "CopyDisks",
				Pods: []PodDiagnostics{
					{
						Name:       "importer-test-vm-100-xyz",
						Phase:      "Running",
						Container:  "importer",
						ErrorCount: 0,
						WarnCount:  1,
						LogTail:    []string{"Transferred 50%"},
					},
				},
				Events: []EventEntry{
					{
						Type:    "Warning",
						Reason:  "FailedScheduling",
						Object:  "Pod/importer-test-vm-100-xyz",
						Message: "0/6 nodes available",
						Age:     "5m",
					},
				},
			},
		},
	}

	b := describe.NewBuilder("TEST")
	Render(b, report)

	desc := b.Build()

	// Verify the DIAGNOSTICS section was added
	found := false
	for _, s := range desc.Sections {
		if s.Title == "DIAGNOSTICS" {
			found = true
			// Should have VDDK field at section level
			hasVDDK := false
			for _, f := range s.Fields {
				if f.Label == "VDDK Image" && f.Value == "quay.io/test/vddk:8.0.0" {
					hasVDDK = true
				}
			}
			if !hasVDDK {
				t.Error("expected VDDK Image field in DIAGNOSTICS section")
			}
			// Should have VM subsection
			if len(s.SubSections) < 1 {
				t.Errorf("expected at least 1 subsection (VM), got %d", len(s.SubSections))
			}
			if s.SubSections[0].Title != "VM: test-vm (vm-100)" {
				t.Errorf("expected subsection 'VM: test-vm (vm-100)', got %q", s.SubSections[0].Title)
			}
			// VM subsection should have Text blocks (Pod, Events)
			vmSub := s.SubSections[0]
			if len(vmSub.Texts) < 2 {
				t.Errorf("expected at least 2 text blocks (Pod + Events), got %d", len(vmSub.Texts))
			}
			break
		}
	}
	if !found {
		t.Error("DIAGNOSTICS section not found in output")
	}
}

func TestRender_NoVDDK(t *testing.T) {
	report := &DiagnosticsReport{
		Config: ConfigContext{
			SourceProvider: "vsphere-provider",
			MigrationType:  "cold",
			VDDKImage:      "",
		},
	}

	b := describe.NewBuilder("TEST")
	Render(b, report)

	desc := b.Build()
	for _, s := range desc.Sections {
		if s.Title == "DIAGNOSTICS" {
			for _, f := range s.Fields {
				if f.Label == "VDDK Image" && f.Value != "Not configured" {
					t.Errorf("expected 'Not configured' for empty VDDK, got %q", f.Value)
				}
			}
		}
	}
}

func TestRender_WithConversion(t *testing.T) {
	report := &DiagnosticsReport{
		Config: ConfigContext{
			SourceProvider: "vsphere-provider",
			MigrationType:  "cold",
			VDDKImage:      "quay.io/test/vddk:8.0.0",
		},
		VMs: []VMDiagnostics{
			{
				Name:  "vm-with-conversion",
				ID:    "vm-200",
				Phase: "ImageConversion",
				Conversion: &ConversionInfo{
					Name:    "conv-vm-200",
					Phase:   "Running",
					Message: "Converting guest OS",
					PodName: "virt-v2v-vm-200-xyz",
				},
			},
		},
	}

	b := describe.NewBuilder("TEST")
	Render(b, report)

	desc := b.Build()
	found := false
	for _, s := range desc.Sections {
		if s.Title == "DIAGNOSTICS" {
			for _, sub := range s.SubSections {
				if sub.Title == "VM: vm-with-conversion (vm-200)" {
					for _, txt := range sub.Texts {
						if strings.Contains(txt.Content, "conv-vm-200") {
							found = true
						}
					}
				}
			}
		}
	}
	if !found {
		t.Error("Conversion info not found in VM subsection texts")
	}
}

func TestRender_ControllerLogsWithErrors(t *testing.T) {
	report := &DiagnosticsReport{
		Config: ConfigContext{
			SourceProvider: "vsphere-provider",
			MigrationType:  "cold",
			VDDKImage:      "quay.io/test/vddk:8.0.0",
		},
		ControllerLogs: &ControllerLogAnalysis{
			ErrorCount: 2,
			WarnCount:  1,
			ErrorLines: []string{
				`{"level":"error","msg":"failed to reconcile plan test-plan"}`,
				`{"level":"error","msg":"plan test-plan: connection refused"}`,
			},
			LogTail: []string{
				`{"level":"info","msg":"reconcile plan test-plan"}`,
				`{"level":"error","msg":"failed to reconcile plan test-plan"}`,
				`{"level":"info","msg":"retrying plan test-plan"}`,
			},
		},
	}

	b := describe.NewBuilder("TEST")
	Render(b, report)

	desc := b.Build()
	content := findControllerLogsContent(t, desc)
	if content == "" {
		return
	}
	if !strings.Contains(content, "2 errors, 1 warnings") {
		t.Error("expected log analysis summary '2 errors, 1 warnings'")
	}
	if !strings.Contains(content, "error lines:") {
		t.Error("expected 'error lines:' section")
	}
	if !strings.Contains(content, "failed to reconcile plan test-plan") {
		t.Error("expected error line about reconcile failure")
	}
	if !strings.Contains(content, "log lines:") {
		t.Error("expected 'log lines:' section")
	}
}

func TestRender_ControllerLogsNoErrors(t *testing.T) {
	report := &DiagnosticsReport{
		Config: ConfigContext{
			VDDKImage: "quay.io/test/vddk:8.0.0",
		},
		ControllerLogs: &ControllerLogAnalysis{
			ErrorCount: 0,
			WarnCount:  0,
			LogTail: []string{
				`{"level":"info","msg":"reconcile plan test-plan"}`,
			},
		},
	}

	b := describe.NewBuilder("TEST")
	Render(b, report)

	desc := b.Build()
	content := findControllerLogsContent(t, desc)
	if content == "" {
		return
	}
	if !strings.Contains(content, "0 errors, 0 warnings") {
		t.Error("expected log analysis summary '0 errors, 0 warnings'")
	}
	if strings.Contains(content, "error lines:") {
		t.Error("should not have error lines section when no errors")
	}
	if !strings.Contains(content, "log lines:") {
		t.Error("expected 'log lines:' section")
	}
}

func findControllerLogsContent(t *testing.T, desc *describe.Description) string {
	t.Helper()
	for _, s := range desc.Sections {
		if s.Title == "DIAGNOSTICS" {
			for _, sub := range s.SubSections {
				if sub.Title == "Controller Logs" {
					if len(sub.Texts) == 0 {
						t.Error("Controller Logs subsection has no text blocks")
						return ""
					}
					return sub.Texts[0].Content
				}
			}
			t.Error("Controller Logs subsection not found")
			return ""
		}
	}
	t.Error("DIAGNOSTICS section not found")
	return ""
}

func TestRender_ErrorLinesShown(t *testing.T) {
	report := &DiagnosticsReport{
		Config: ConfigContext{
			SourceProvider: "vsphere-provider",
			MigrationType:  "cold",
			VDDKImage:      "quay.io/test/vddk:8.0.0",
		},
		VMs: []VMDiagnostics{
			{
				Name:  "error-vm",
				ID:    "vm-300",
				Phase: "Failed",
				Error: "conversion failed",
				Pods: []PodDiagnostics{
					{
						Name:       "virt-v2v-pod",
						Phase:      "Failed",
						Container:  "virt-v2v",
						ErrorCount: 2,
						WarnCount:  0,
						ErrorLines: []string{
							"error: disk read failed",
							"fatal: cannot continue",
						},
						LogTail: []string{
							"starting conversion",
							"error: disk read failed",
							"processing block 5",
							"fatal: cannot continue",
						},
					},
				},
			},
		},
	}

	b := describe.NewBuilder("TEST")
	Render(b, report)

	desc := b.Build()
	for _, s := range desc.Sections {
		if s.Title == "DIAGNOSTICS" {
			for _, sub := range s.SubSections {
				for _, txt := range sub.Texts {
					if strings.Contains(txt.Content, "error lines:") {
						if !strings.Contains(txt.Content, "disk read failed") {
							t.Error("expected error line 'disk read failed' in output")
						}
						if !strings.Contains(txt.Content, "cannot continue") {
							t.Error("expected error line 'cannot continue' in output")
						}
						return
					}
				}
			}
		}
	}
	t.Error("Error lines section not found")
}

func TestRender_V2VStageShown(t *testing.T) {
	report := &DiagnosticsReport{
		Config: ConfigContext{
			VDDKImage: "quay.io/test/vddk:8.0.0",
		},
		VMs: []VMDiagnostics{
			{
				Name:  "converting-vm",
				ID:    "vm-400",
				Phase: "CopyDisks",
				Pods: []PodDiagnostics{
					{
						Name:        "virt-v2v-vm-400-xyz",
						Phase:       "Running",
						Container:   "virt-v2v",
						V2VStage:    "disk-copy",
						ProgressPct: "50",
					},
				},
			},
		},
	}

	b := describe.NewBuilder("TEST")
	Render(b, report)

	desc := b.Build()
	podContent := findPodContent(t, desc, "VM: converting-vm (vm-400)")
	if podContent == "" {
		return
	}
	if !strings.Contains(podContent, "V2V Stage:") {
		t.Error("expected 'V2V Stage:' in pod output")
	}
	if !strings.Contains(podContent, "disk-copy") {
		t.Error("expected 'disk-copy' stage in pod output")
	}
	if !strings.Contains(podContent, "50%") {
		t.Error("expected '50%' progress in pod output")
	}
}

func TestRender_V2VStageFinishGreen(t *testing.T) {
	report := &DiagnosticsReport{
		Config: ConfigContext{VDDKImage: "quay.io/test/vddk:8.0.0"},
		VMs: []VMDiagnostics{
			{
				Name: "done-vm", ID: "vm-500", Phase: "Completed",
				Pods: []PodDiagnostics{
					{Name: "virt-v2v-done", Phase: "Succeeded", Container: "virt-v2v", V2VStage: "finish"},
				},
			},
		},
	}

	b := describe.NewBuilder("TEST")
	Render(b, report)

	desc := b.Build()
	podContent := findPodContent(t, desc, "VM: done-vm (vm-500)")
	if podContent == "" {
		return
	}
	if !strings.Contains(podContent, "V2V Stage:") {
		t.Error("expected V2V Stage line for finish stage")
	}
	if !strings.Contains(podContent, "finish") {
		t.Error("expected 'finish' in stage output")
	}
}

func TestRender_V2VStageFailedRed(t *testing.T) {
	report := &DiagnosticsReport{
		Config: ConfigContext{VDDKImage: "quay.io/test/vddk:8.0.0"},
		VMs: []VMDiagnostics{
			{
				Name: "failed-vm", ID: "vm-600", Phase: "Failed",
				Pods: []PodDiagnostics{
					{Name: "virt-v2v-failed", Phase: "Failed", Container: "virt-v2v", V2VStage: "source-setup"},
				},
			},
		},
	}

	b := describe.NewBuilder("TEST")
	Render(b, report)

	desc := b.Build()
	podContent := findPodContent(t, desc, "VM: failed-vm (vm-600)")
	if podContent == "" {
		return
	}
	if !strings.Contains(podContent, "V2V Stage:") {
		t.Error("expected V2V Stage line for failed pod")
	}
	if !strings.Contains(podContent, "source-setup") {
		t.Error("expected 'source-setup' stage in failed pod output")
	}
}

func TestRender_V2VStageNotShownForNonV2V(t *testing.T) {
	report := &DiagnosticsReport{
		Config: ConfigContext{VDDKImage: "quay.io/test/vddk:8.0.0"},
		VMs: []VMDiagnostics{
			{
				Name: "import-vm", ID: "vm-700", Phase: "CopyDisks",
				Pods: []PodDiagnostics{
					{Name: "importer-vm-700", Phase: "Running", Container: "importer"},
				},
			},
		},
	}

	b := describe.NewBuilder("TEST")
	Render(b, report)

	desc := b.Build()
	podContent := findPodContent(t, desc, "VM: import-vm (vm-700)")
	if podContent == "" {
		return
	}
	if strings.Contains(podContent, "V2V Stage:") {
		t.Error("V2V Stage should not appear for non-v2v pods")
	}
}

func TestFormatV2VStage(t *testing.T) {
	tests := []struct {
		name        string
		stage       string
		progressPct string
		podPhase    string
		wantContain string
	}{
		{"disk-copy with progress", "disk-copy", "50", "Running", "disk-copy (50%)"},
		{"finish no progress", "finish", "", "Succeeded", "finish"},
		{"failed pod shows stage", "source-setup", "", "Failed", "source-setup"},
		{"evicted pod shows stage", "disk-copy", "75", "Evicted", "disk-copy (75%)"},
		{"default stage passthrough", "inspect", "", "Running", "inspect"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatV2VStage(tc.stage, tc.progressPct, tc.podPhase)
			if !strings.Contains(got, tc.wantContain) {
				t.Errorf("formatV2VStage(%q, %q, %q) = %q, want to contain %q",
					tc.stage, tc.progressPct, tc.podPhase, got, tc.wantContain)
			}
		})
	}
}

func findPodContent(t *testing.T, desc *describe.Description, vmTitle string) string {
	t.Helper()
	for _, s := range desc.Sections {
		if s.Title == "DIAGNOSTICS" {
			for _, sub := range s.SubSections {
				if sub.Title == vmTitle {
					for _, txt := range sub.Texts {
						if strings.Contains(txt.Content, "Name:") {
							return txt.Content
						}
					}
					t.Errorf("Pod text block not found in %q", vmTitle)
					return ""
				}
			}
			t.Errorf("VM subsection %q not found", vmTitle)
			return ""
		}
	}
	t.Error("DIAGNOSTICS section not found")
	return ""
}
