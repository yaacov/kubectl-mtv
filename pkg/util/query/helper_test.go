package query

import (
	"testing"
)

func TestApplyQueryInterface(t *testing.T) {
	tests := []struct {
		name      string
		data      interface{}
		query     string
		wantLen   int
		wantErr   bool
		errSubstr string
	}{
		{
			name: "[]map[string]interface{} filters correctly",
			data: []map[string]interface{}{
				{"name": "a", "status": "running"},
				{"name": "b", "status": "stopped"},
			},
			query:   "where status = 'running'",
			wantLen: 1,
		},
		{
			name: "[]interface{} with valid maps filters correctly",
			data: []interface{}{
				map[string]interface{}{"name": "a", "status": "running"},
				map[string]interface{}{"name": "b", "status": "stopped"},
				map[string]interface{}{"name": "c", "status": "running"},
			},
			query:   "where status = 'running'",
			wantLen: 2,
		},
		{
			name: "[]interface{} with non-map element returns error with index",
			data: []interface{}{
				map[string]interface{}{"name": "a"},
				"not-a-map",
			},
			query:     "where name = 'a'",
			wantErr:   true,
			errSubstr: "index 1",
		},
		{
			name: "[]interface{} with non-map at index 0 returns error",
			data: []interface{}{
				42,
				map[string]interface{}{"name": "a"},
			},
			query:     "where name = 'a'",
			wantErr:   true,
			errSubstr: "index 0",
		},
		{
			name:    "single map[string]interface{} filters correctly",
			data:    map[string]interface{}{"name": "a", "status": "running"},
			query:   "where status = 'running'",
			wantLen: 1,
		},
		{
			name:    "single map[string]interface{} filters out non-match",
			data:    map[string]interface{}{"name": "a", "status": "stopped"},
			query:   "where status = 'running'",
			wantLen: 0,
		},
		{
			name:    "empty []interface{} returns empty result",
			data:    []interface{}{},
			query:   "where name = 'a'",
			wantLen: 0,
		},
		{
			name:    "unsupported type returns data as-is",
			data:    "a string",
			query:   "where name = 'a'",
			wantLen: -1, // skip length check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ApplyQueryInterface(tt.data, tt.query)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errSubstr != "" {
					if !containsSubstring(err.Error(), tt.errSubstr) {
						t.Fatalf("error %q does not contain %q", err.Error(), tt.errSubstr)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantLen < 0 {
				return // skip length check for unsupported types
			}

			items, ok := result.([]map[string]interface{})
			if !ok {
				t.Fatalf("expected []map[string]interface{}, got %T", result)
			}
			if len(items) != tt.wantLen {
				t.Fatalf("expected %d items, got %d", tt.wantLen, len(items))
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
