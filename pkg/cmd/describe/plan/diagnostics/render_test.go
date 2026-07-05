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
