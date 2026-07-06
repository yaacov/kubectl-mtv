package diagnostics

import "testing"

func TestDetectV2VStage_FullFlow(t *testing.T) {
	lines := []string{
		"Building command: virt-v2v [-v -x -o kubevirt]",
		"virt-v2v monitoring: Prometheus progress counter registered.",
		"[   0.0] Setting up the source: -i libvirt",
		"libguestfs: trace: set_verbose true",
		"[  12.5] Opening the source: -i libvirt",
		"[  30.2] Inspecting the source: -i libvirt",
		"[  45.1] Mapping filesystem data",
		"[  50.0] Creating an overlay to protect the source",
		"[  55.3] Setting up the destination: -o kubevirt",
		"[ 147.3] Copying disk 1/1",
		"virt-v2v monitoring: Progress update, completed 50 %",
		"nbdkit: debug: read block",
		"virt-v2v monitoring: Progress update, completed 75 %",
		"[ 358.4] Creating output metadata",
		"[ 360.1] Finishing off",
	}

	stage, progress := detectV2VStage(lines)
	if stage != "finish" {
		t.Errorf("expected stage 'finish', got %q", stage)
	}
	if progress != "75" {
		t.Errorf("expected progress '75', got %q", progress)
	}
}

func TestDetectV2VStage_StuckInDiskCopy(t *testing.T) {
	lines := []string{
		"[   0.0] Setting up the source: -i libvirt",
		"[  12.5] Opening the source: -i libvirt",
		"[  30.2] Inspecting the source: -i libvirt",
		"[ 147.3] Copying disk 1/1",
		"virt-v2v monitoring: Progress update, completed 42 %",
		"nbdkit: debug: read block offset=1024 count=512",
	}

	stage, progress := detectV2VStage(lines)
	if stage != "disk-copy" {
		t.Errorf("expected stage 'disk-copy', got %q", stage)
	}
	if progress != "42" {
		t.Errorf("expected progress '42', got %q", progress)
	}
}

func TestDetectV2VStage_StuckInSourceSetup(t *testing.T) {
	lines := []string{
		"Building command: virt-v2v [-v -x -o kubevirt]",
		"[   0.0] Setting up the source: -i libvirt",
		"libguestfs: trace: set_verbose true",
	}

	stage, progress := detectV2VStage(lines)
	if stage != "source-setup" {
		t.Errorf("expected stage 'source-setup', got %q", stage)
	}
	if progress != "" {
		t.Errorf("expected empty progress, got %q", progress)
	}
}

func TestDetectV2VStage_EmptyLog(t *testing.T) {
	stage, progress := detectV2VStage(nil)
	if stage != "" {
		t.Errorf("expected empty stage, got %q", stage)
	}
	if progress != "" {
		t.Errorf("expected empty progress, got %q", progress)
	}
}

func TestDetectV2VStage_EarlyBootNoPhaseMarkers(t *testing.T) {
	// Kernel/dracut lines during appliance boot have 6 decimal digits and are NOT phase markers.
	// detectV2VStage returns empty; the caller (buildPodDiagnostics) defaults to "init".
	lines := []string{
		"*** Including module: drm ***",
		"[   36.063171] dracut[808] *** Including module: drm ***",
		"Omitting driver radeon",
		"[   36.100000] kernel: some boot message",
	}

	stage, progress := detectV2VStage(lines)
	if stage != "" {
		t.Errorf("expected empty stage (caller defaults to 'init'), got %q", stage)
	}
	if progress != "" {
		t.Errorf("expected empty progress, got %q", progress)
	}
}

func TestDetectV2VStage_NonV2VLog(t *testing.T) {
	lines := []string{
		"time=\"2026-01-01T00:00:00Z\" level=info msg=\"Starting importer\"",
		"time=\"2026-01-01T00:01:00Z\" level=info msg=\"Import complete\"",
	}

	stage, progress := detectV2VStage(lines)
	if stage != "" {
		t.Errorf("expected empty stage for non-v2v log, got %q", stage)
	}
	if progress != "" {
		t.Errorf("expected empty progress for non-v2v log, got %q", progress)
	}
}

func TestDetectV2VStage_ProgressResetsOnSourceSetup(t *testing.T) {
	lines := []string{
		"[ 147.3] Copying disk 1/1",
		"virt-v2v monitoring: Progress update, completed 100 %",
		"[   0.0] Setting up the source: -i libvirt",
	}

	stage, progress := detectV2VStage(lines)
	if stage != "source-setup" {
		t.Errorf("expected stage 'source-setup', got %q", stage)
	}
	if progress != "" {
		t.Errorf("expected progress to be reset, got %q", progress)
	}
}

func TestDetectV2VStage_AllStages(t *testing.T) {
	tests := []struct {
		line  string
		stage string
	}{
		{"[   0.0] Setting up the source: -i libvirt", "source-setup"},
		{"[  12.0] Opening the source: -i libvirt", "source-open"},
		{"[  30.0] Inspecting the source: -i libvirt", "inspect"},
		{"[  40.0] Mapping filesystem data", "map-fs"},
		{"[  50.0] Creating an overlay to protect the source", "overlay"},
		{"[  55.0] Setting up the destination: -o kubevirt", "dest-setup"},
		{"[ 147.0] Copying disk 1/1", "disk-copy"},
		{"[ 358.0] Creating output metadata", "metadata"},
		{"[ 360.0] Finishing off", "finish"},
	}

	for _, tc := range tests {
		stage, _ := detectV2VStage([]string{tc.line})
		if stage != tc.stage {
			t.Errorf("line %q: expected stage %q, got %q", tc.line, tc.stage, stage)
		}
	}
}

func TestMatchV2VStage(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"Setting up the source: -i libvirt", "source-setup"},
		{"Opening the source: -i libvirt", "source-open"},
		{"Inspecting the source: -i libvirt", "inspect"},
		{"Mapping filesystem data", "map-fs"},
		{"Creating an overlay to protect the source", "overlay"},
		{"Setting up the destination: -o kubevirt", "dest-setup"},
		{"Copying disk 1/1", "disk-copy"},
		{"Creating output metadata", "metadata"},
		{"Finishing off", "finish"},
		{"Some random message", ""},
	}

	for _, tc := range tests {
		got := matchV2VStage(tc.message)
		if got != tc.expected {
			t.Errorf("matchV2VStage(%q) = %q, want %q", tc.message, got, tc.expected)
		}
	}
}
