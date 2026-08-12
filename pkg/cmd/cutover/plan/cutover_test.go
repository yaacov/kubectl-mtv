package plan

import (
	"reflect"
	"sort"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func planWithVMs(vms ...map[string]interface{}) *unstructured.Unstructured {
	items := make([]interface{}, 0, len(vms))
	for _, vm := range vms {
		items = append(items, vm)
	}
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"vms": items,
			},
		},
	}
}

func vm(name, id string) map[string]interface{} {
	return map[string]interface{}{"name": name, "id": id}
}

func TestResolveVMRefs(t *testing.T) {
	planObj := planWithVMs(
		vm("web01", "vm-1"),
		vm("db01", "vm-2"),
		vm("", "vm-3"),
	)

	tests := []struct {
		name      string
		vmRefs    []string
		wantIDs   []string
		wantNames []string
		wantErr   bool
	}{
		{
			name:      "resolve by name",
			vmRefs:    []string{"web01"},
			wantIDs:   []string{"vm-1"},
			wantNames: []string{"web01"},
		},
		{
			name:      "resolve by id fallback",
			vmRefs:    []string{"vm-2"},
			wantIDs:   []string{"vm-2"},
			wantNames: []string{"db01"},
		},
		{
			name:      "resolve multiple mixed name and id",
			vmRefs:    []string{"web01", "vm-2"},
			wantIDs:   []string{"vm-1", "vm-2"},
			wantNames: []string{"web01", "db01"},
		},
		{
			name:      "resolve vm with no name by id",
			vmRefs:    []string{"vm-3"},
			wantIDs:   []string{"vm-3"},
			wantNames: []string{"vm-3"},
		},
		{
			name:    "unknown ref errors",
			vmRefs:  []string{"does-not-exist"},
			wantErr: true,
		},
		{
			name:    "partial match still errors",
			vmRefs:  []string{"web01", "does-not-exist"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, names, err := resolveVMRefs(planObj, tt.vmRefs)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(ids, tt.wantIDs) {
				t.Errorf("ids = %v, want %v", ids, tt.wantIDs)
			}
			if !reflect.DeepEqual(names, tt.wantNames) {
				t.Errorf("names = %v, want %v", names, tt.wantNames)
			}
		})
	}
}

func TestMergeVMCutovers(t *testing.T) {
	existing := []interface{}{
		map[string]interface{}{"id": "vm-1", "cutover": "2026-01-01T00:00:00Z"},
		map[string]interface{}{"id": "vm-2", "cutover": "2026-01-02T00:00:00Z"},
	}

	tests := []struct {
		name     string
		existing []interface{}
		new      []vmCutoverEntry
		want     map[string]string
	}{
		{
			name:     "add new vm without touching existing",
			existing: existing,
			new:      []vmCutoverEntry{{ID: "vm-3", Cutover: "2026-01-03T00:00:00Z"}},
			want: map[string]string{
				"vm-1": "2026-01-01T00:00:00Z",
				"vm-2": "2026-01-02T00:00:00Z",
				"vm-3": "2026-01-03T00:00:00Z",
			},
		},
		{
			name:     "override existing vm",
			existing: existing,
			new:      []vmCutoverEntry{{ID: "vm-1", Cutover: "2026-05-05T00:00:00Z"}},
			want: map[string]string{
				"vm-1": "2026-05-05T00:00:00Z",
				"vm-2": "2026-01-02T00:00:00Z",
			},
		},
		{
			name:     "empty existing",
			existing: nil,
			new:      []vmCutoverEntry{{ID: "vm-1", Cutover: "2026-01-01T00:00:00Z"}},
			want: map[string]string{
				"vm-1": "2026-01-01T00:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := mergeVMCutovers(tt.existing, tt.new)

			got := make(map[string]string, len(merged))
			for _, entryObj := range merged {
				entry, ok := entryObj.(map[string]interface{})
				if !ok {
					t.Fatalf("unexpected entry type: %T", entryObj)
				}
				id, _ := entry["id"].(string)
				cutover, _ := entry["cutover"].(string)
				got[id] = cutover
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("merged = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeVMCutoversStableIDs(t *testing.T) {
	merged := mergeVMCutovers(nil, []vmCutoverEntry{
		{ID: "vm-2", Cutover: "2026-01-02T00:00:00Z"},
		{ID: "vm-1", Cutover: "2026-01-01T00:00:00Z"},
	})

	var ids []string
	for _, entryObj := range merged {
		entry := entryObj.(map[string]interface{})
		ids = append(ids, entry["id"].(string))
	}
	sort.Strings(ids)

	if !reflect.DeepEqual(ids, []string{"vm-1", "vm-2"}) {
		t.Errorf("ids = %v, want [vm-1 vm-2]", ids)
	}
}
