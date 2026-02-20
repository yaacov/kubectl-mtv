package plan

import (
	"testing"
)

func TestExtractPowerStatus(t *testing.T) {
	tests := []struct {
		name string
		vm   map[string]interface{}
		want string
	}{
		{
			name: "vsphere poweredOn",
			vm:   map[string]interface{}{"powerState": "poweredOn"},
			want: "Running",
		},
		{
			name: "vsphere poweredOff",
			vm:   map[string]interface{}{"powerState": "poweredOff"},
			want: "Stopped",
		},
		{
			name: "ovirt up",
			vm:   map[string]interface{}{"powerState": "up"},
			want: "Running",
		},
		{
			name: "ovirt down",
			vm:   map[string]interface{}{"powerState": "down"},
			want: "Stopped",
		},
		{
			name: "status running",
			vm:   map[string]interface{}{"status": "Running"},
			want: "Running",
		},
		{
			name: "status stopped",
			vm:   map[string]interface{}{"status": "Stopped"},
			want: "Stopped",
		},
		{
			name: "KubeVirt printableStatus Stopped",
			vm: map[string]interface{}{
				"object": map[string]interface{}{
					"status": map[string]interface{}{
						"printableStatus": "Stopped",
					},
				},
			},
			want: "Stopped",
		},
		{
			name: "KubeVirt printableStatus Running",
			vm: map[string]interface{}{
				"object": map[string]interface{}{
					"status": map[string]interface{}{
						"printableStatus": "Running",
					},
				},
			},
			want: "Running",
		},
		{
			name: "nested object.status.phase running",
			vm: map[string]interface{}{
				"object": map[string]interface{}{
					"status": map[string]interface{}{
						"phase": "Running",
					},
				},
			},
			want: "Running",
		},
		{
			name: "EC2 State.Name running",
			vm: map[string]interface{}{
				"State": map[string]interface{}{
					"Name": "running",
				},
			},
			want: "Running",
		},
		{
			name: "EC2 State.Name stopped",
			vm: map[string]interface{}{
				"State": map[string]interface{}{
					"Name": "stopped",
				},
			},
			want: "Stopped",
		},
		{
			name: "empty vm",
			vm:   map[string]interface{}{},
			want: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPowerStatus(tt.vm)
			if got != tt.want {
				t.Errorf("extractPowerStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		name string
		vm   map[string]interface{}
		want string
	}{
		{
			name: "direct ipAddress",
			vm:   map[string]interface{}{"ipAddress": "10.0.0.1"},
			want: "10.0.0.1",
		},
		{
			name: "EC2 PublicIpAddress",
			vm:   map[string]interface{}{"PublicIpAddress": "54.1.2.3"},
			want: "54.1.2.3",
		},
		{
			name: "EC2 PrivateIpAddress only",
			vm:   map[string]interface{}{"PrivateIpAddress": "172.16.0.5"},
			want: "172.16.0.5",
		},
		{
			name: "nics array with ipAddress",
			vm: map[string]interface{}{
				"nics": []interface{}{
					map[string]interface{}{"ipAddress": "192.168.1.10"},
				},
			},
			want: "192.168.1.10",
		},
		{
			name: "nics array with ipAddresses list",
			vm: map[string]interface{}{
				"nics": []interface{}{
					map[string]interface{}{
						"ipAddresses": []interface{}{"10.10.10.1", "10.10.10.2"},
					},
				},
			},
			want: "10.10.10.1",
		},
		{
			name: "openshift workload interfaces",
			vm: map[string]interface{}{
				"object": map[string]interface{}{
					"status": map[string]interface{}{
						"interfaces": []interface{}{
							map[string]interface{}{"ipAddress": "10.244.0.5"},
						},
					},
				},
			},
			want: "10.244.0.5",
		},
		{
			name: "empty vm",
			vm:   map[string]interface{}{},
			want: "-",
		},
		{
			name: "prefers direct ipAddress over nics",
			vm: map[string]interface{}{
				"ipAddress": "10.0.0.1",
				"nics": []interface{}{
					map[string]interface{}{"ipAddress": "10.0.0.2"},
				},
			},
			want: "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractIP(tt.vm)
			if got != tt.want {
				t.Errorf("extractIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveTargetName(t *testing.T) {
	tests := []struct {
		name           string
		specTargetName string
		migVM          map[string]interface{}
		sourceName     string
		want           string
	}{
		{
			name:           "spec target name takes priority",
			specTargetName: "custom-target",
			migVM:          map[string]interface{}{"newName": "mig-target"},
			sourceName:     "source-vm",
			want:           "custom-target",
		},
		{
			name:           "falls back to migration newName",
			specTargetName: "",
			migVM:          map[string]interface{}{"newName": "mig-target"},
			sourceName:     "source-vm",
			want:           "mig-target",
		},
		{
			name:           "falls back to source name",
			specTargetName: "",
			migVM:          nil,
			sourceName:     "source-vm",
			want:           "source-vm",
		},
		{
			name:           "migration VM with empty newName falls back to source",
			specTargetName: "",
			migVM:          map[string]interface{}{},
			sourceName:     "source-vm",
			want:           "source-vm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveTargetName(tt.specTargetName, tt.migVM, tt.sourceName)
			if got != tt.want {
				t.Errorf("resolveTargetName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildProgressString(t *testing.T) {
	tests := []struct {
		name  string
		migVM map[string]interface{}
		want  string
	}{
		{
			name:  "nil migVM",
			migVM: nil,
			want:  "-",
		},
		{
			name:  "empty phase",
			migVM: map[string]interface{}{},
			want:  "-",
		},
		{
			name: "succeeded VM",
			migVM: map[string]interface{}{
				"phase": "Completed",
				"conditions": []interface{}{
					map[string]interface{}{"type": "Succeeded", "status": "True"},
				},
			},
			want: "Completed",
		},
		{
			name: "failed VM",
			migVM: map[string]interface{}{
				"phase": "Completed",
				"conditions": []interface{}{
					map[string]interface{}{"type": "Failed", "status": "True"},
				},
			},
			want: "Failed",
		},
		{
			name: "canceled VM",
			migVM: map[string]interface{}{
				"phase": "Completed",
				"conditions": []interface{}{
					map[string]interface{}{"type": "Canceled", "status": "True"},
				},
			},
			want: "Canceled",
		},
		{
			name: "running VM with pipeline progress",
			migVM: map[string]interface{}{
				"phase": "CopyDisks",
				"pipeline": []interface{}{
					map[string]interface{}{
						"name": "DiskTransferV2v",
						"progress": map[string]interface{}{
							"completed": int64(50),
							"total":     int64(100),
						},
					},
					map[string]interface{}{
						"name": "ImageConversion",
						"progress": map[string]interface{}{
							"completed": int64(0),
							"total":     int64(100),
						},
					},
				},
			},
			want: "CopyDisks (25%)",
		},
		{
			name: "running VM with no pipeline",
			migVM: map[string]interface{}{
				"phase": "CopyDisks",
			},
			want: "CopyDisks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildProgressString(tt.migVM)
			if got != tt.want {
				t.Errorf("buildProgressString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLookupSourceVM(t *testing.T) {
	sourceVMs := map[string]map[string]interface{}{
		"vm-123": {
			"id":         "vm-123",
			"name":       "web-server",
			"powerState": "poweredOn",
			"ipAddress":  "10.0.0.1",
		},
		"name:web-server": {
			"id":         "vm-123",
			"name":       "web-server",
			"powerState": "poweredOn",
			"ipAddress":  "10.0.0.1",
		},
	}

	tests := []struct {
		name       string
		sourceVMs  map[string]map[string]interface{}
		vmID       string
		vmName     string
		wantStatus string
		wantIP     string
	}{
		{
			name:       "found by ID",
			sourceVMs:  sourceVMs,
			vmID:       "vm-123",
			vmName:     "web-server",
			wantStatus: "Running",
			wantIP:     "10.0.0.1",
		},
		{
			name:       "found by name fallback",
			sourceVMs:  sourceVMs,
			vmID:       "unknown-id",
			vmName:     "web-server",
			wantStatus: "Running",
			wantIP:     "10.0.0.1",
		},
		{
			name:       "not found",
			sourceVMs:  sourceVMs,
			vmID:       "unknown-id",
			vmName:     "unknown-name",
			wantStatus: "Not Found",
			wantIP:     "-",
		},
		{
			name:       "nil source map",
			sourceVMs:  nil,
			vmID:       "vm-123",
			vmName:     "web-server",
			wantStatus: "-",
			wantIP:     "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotIP := lookupSourceVM(tt.sourceVMs, tt.vmID, tt.vmName)
			if gotStatus != tt.wantStatus {
				t.Errorf("lookupSourceVM() status = %q, want %q", gotStatus, tt.wantStatus)
			}
			if gotIP != tt.wantIP {
				t.Errorf("lookupSourceVM() ip = %q, want %q", gotIP, tt.wantIP)
			}
		})
	}
}

func TestLookupTargetWorkload(t *testing.T) {
	targetWorkloads := map[string]map[string]interface{}{
		"web-server": {
			"name": "web-server",
			"object": map[string]interface{}{
				"status": map[string]interface{}{
					"phase": "Running",
					"interfaces": []interface{}{
						map[string]interface{}{"ipAddress": "10.244.0.5"},
					},
				},
			},
		},
	}

	tests := []struct {
		name       string
		workloads  map[string]map[string]interface{}
		targetName string
		wantStatus string
		wantIP     string
	}{
		{
			name:       "found",
			workloads:  targetWorkloads,
			targetName: "web-server",
			wantStatus: "Running",
			wantIP:     "10.244.0.5",
		},
		{
			name:       "not found",
			workloads:  targetWorkloads,
			targetName: "unknown",
			wantStatus: "Not Found",
			wantIP:     "-",
		},
		{
			name:       "nil workloads map",
			workloads:  nil,
			targetName: "web-server",
			wantStatus: "-",
			wantIP:     "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotIP := lookupTargetWorkload(tt.workloads, tt.targetName)
			if gotStatus != tt.wantStatus {
				t.Errorf("lookupTargetWorkload() status = %q, want %q", gotStatus, tt.wantStatus)
			}
			if gotIP != tt.wantIP {
				t.Errorf("lookupTargetWorkload() ip = %q, want %q", gotIP, tt.wantIP)
			}
		})
	}
}
