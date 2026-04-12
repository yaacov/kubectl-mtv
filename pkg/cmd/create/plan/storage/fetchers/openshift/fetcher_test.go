package openshift

import (
	"testing"
)

func scItem(name string, annotations map[string]interface{}) interface{} {
	item := map[string]interface{}{"name": name}
	if annotations != nil {
		item["object"] = map[string]interface{}{
			"metadata": map[string]interface{}{
				"annotations": annotations,
			},
		}
	}
	return item
}

func TestBuildTargetStorageList_Priority(t *testing.T) {
	tests := []struct {
		name          string
		storageArray  []interface{}
		expectedFirst string
		expectedCount int
		expectErr     bool
	}{
		{
			name: "Virt annotation wins over everything",
			storageArray: []interface{}{
				scItem("plain-sc", nil),
				scItem("k8s-default", map[string]interface{}{
					"storageclass.kubernetes.io/is-default-class": "true",
				}),
				scItem("virt-default", map[string]interface{}{
					"storageclass.kubevirt.io/is-default-virt-class": "true",
				}),
				scItem("ceph-virtualization", nil),
			},
			expectedFirst: "virt-default",
			expectedCount: 4,
		},
		{
			name: "K8s annotation wins when no virt annotation",
			storageArray: []interface{}{
				scItem("plain-sc", nil),
				scItem("k8s-default", map[string]interface{}{
					"storageclass.kubernetes.io/is-default-class": "true",
				}),
				scItem("ceph-virtualization", nil),
			},
			expectedFirst: "k8s-default",
			expectedCount: 3,
		},
		{
			name: "Virtualization name wins when no annotations",
			storageArray: []interface{}{
				scItem("plain-sc", nil),
				scItem("ocs-virtualization", nil),
				scItem("another-sc", nil),
			},
			expectedFirst: "ocs-virtualization",
			expectedCount: 3,
		},
		{
			name: "First available when nothing else matches",
			storageArray: []interface{}{
				scItem("first-sc", nil),
				scItem("second-sc", nil),
			},
			expectedFirst: "first-sc",
			expectedCount: 2,
		},
		{
			name:          "Empty list returns error",
			storageArray:  []interface{}{},
			expectErr:     true,
			expectedCount: 0,
		},
		{
			name: "Single SC returned as-is",
			storageArray: []interface{}{
				scItem("only-sc", nil),
			},
			expectedFirst: "only-sc",
			expectedCount: 1,
		},
		{
			name: "All SCs returned, not just the default",
			storageArray: []interface{}{
				scItem("sc-a", nil),
				scItem("sc-b", nil),
				scItem("sc-c", map[string]interface{}{
					"storageclass.kubernetes.io/is-default-class": "true",
				}),
			},
			expectedFirst: "sc-c",
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildTargetStorageList(tt.storageArray)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != tt.expectedCount {
				t.Fatalf("expected %d SCs, got %d", tt.expectedCount, len(result))
			}
			if result[0].StorageClass != tt.expectedFirst {
				t.Errorf("expected first SC '%s', got '%s'", tt.expectedFirst, result[0].StorageClass)
			}
		})
	}
}

func TestBuildTargetStorageList_NoDuplicateDefault(t *testing.T) {
	storageArray := []interface{}{
		scItem("default-sc", map[string]interface{}{
			"storageclass.kubevirt.io/is-default-virt-class": "true",
		}),
		scItem("other-sc", nil),
	}

	result, err := buildTargetStorageList(storageArray)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// default-sc should appear exactly once (at index 0)
	count := 0
	for _, s := range result {
		if s.StorageClass == "default-sc" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("default-sc appears %d times, expected 1", count)
	}
}
