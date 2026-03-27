package ovirt

import (
	"sort"
	"testing"
)

func TestCollectStorageDomains(t *testing.T) {
	tests := []struct {
		name                     string
		disks                    []interface{}
		diskIDSet                map[string]bool
		expectedStorageDomainIDs []string
	}{
		{
			name: "image disks are collected",
			disks: []interface{}{
				map[string]interface{}{
					"id":            "disk-1",
					"storageType":   "image",
					"storageDomain": "sd-nfs-1",
				},
				map[string]interface{}{
					"id":            "disk-2",
					"storageType":   "image",
					"storageDomain": "sd-iscsi-1",
				},
			},
			diskIDSet:                map[string]bool{"disk-1": true, "disk-2": true},
			expectedStorageDomainIDs: []string{"sd-iscsi-1", "sd-nfs-1"},
		},
		{
			name: "LUN disks are skipped",
			disks: []interface{}{
				map[string]interface{}{
					"id":          "disk-lun-1",
					"storageType": "lun",
				},
				map[string]interface{}{
					"id":          "disk-lun-2",
					"storageType": "lun",
				},
			},
			diskIDSet:                map[string]bool{"disk-lun-1": true, "disk-lun-2": true},
			expectedStorageDomainIDs: []string{},
		},
		{
			name: "mixed LUN and image disks only map image storage domains",
			disks: []interface{}{
				map[string]interface{}{
					"id":            "disk-img",
					"storageType":   "image",
					"storageDomain": "sd-nfs-1",
				},
				map[string]interface{}{
					"id":          "disk-lun",
					"storageType": "lun",
				},
			},
			diskIDSet:                map[string]bool{"disk-img": true, "disk-lun": true},
			expectedStorageDomainIDs: []string{"sd-nfs-1"},
		},
		{
			name: "disks not in diskIDSet are ignored",
			disks: []interface{}{
				map[string]interface{}{
					"id":            "disk-other",
					"storageType":   "image",
					"storageDomain": "sd-other",
				},
				map[string]interface{}{
					"id":            "disk-mine",
					"storageType":   "image",
					"storageDomain": "sd-nfs-1",
				},
			},
			diskIDSet:                map[string]bool{"disk-mine": true},
			expectedStorageDomainIDs: []string{"sd-nfs-1"},
		},
		{
			name: "disk without storageType is still collected",
			disks: []interface{}{
				map[string]interface{}{
					"id":            "disk-no-type",
					"storageDomain": "sd-nfs-1",
				},
			},
			diskIDSet:                map[string]bool{"disk-no-type": true},
			expectedStorageDomainIDs: []string{"sd-nfs-1"},
		},
		{
			name:                     "empty disk list returns no storage domains",
			disks:                    []interface{}{},
			diskIDSet:                map[string]bool{"disk-1": true},
			expectedStorageDomainIDs: []string{},
		},
		{
			name: "deduplicates storage domains from multiple image disks",
			disks: []interface{}{
				map[string]interface{}{
					"id":            "disk-1",
					"storageType":   "image",
					"storageDomain": "sd-nfs-1",
				},
				map[string]interface{}{
					"id":            "disk-2",
					"storageType":   "image",
					"storageDomain": "sd-nfs-1",
				},
			},
			diskIDSet:                map[string]bool{"disk-1": true, "disk-2": true},
			expectedStorageDomainIDs: []string{"sd-nfs-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageDomainIDSet := make(map[string]bool)
			collectStorageDomains(tt.disks, tt.diskIDSet, storageDomainIDSet)

			got := make([]string, 0, len(storageDomainIDSet))
			for id := range storageDomainIDSet {
				got = append(got, id)
			}
			sort.Strings(got)

			want := tt.expectedStorageDomainIDs
			if len(want) == 0 {
				want = []string{}
			}
			sort.Strings(want)

			if len(got) != len(want) {
				t.Fatalf("storage domain count: got %d (%v), want %d (%v)", len(got), got, len(want), want)
			}
			for i := range got {
				if got[i] != want[i] {
					t.Errorf("storage domain [%d]: got %q, want %q", i, got[i], want[i])
				}
			}
		})
	}
}
