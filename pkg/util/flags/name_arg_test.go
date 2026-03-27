package flags

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestResolveNameArg(t *testing.T) {
	tests := []struct {
		name     string
		flagVal  string
		args     []string
		wantName string
		wantErr  bool
	}{
		{
			name:     "no args and no flag",
			flagVal:  "",
			args:     nil,
			wantName: "",
		},
		{
			name:     "flag only",
			flagVal:  "my-plan",
			args:     nil,
			wantName: "my-plan",
		},
		{
			name:     "positional arg only",
			flagVal:  "",
			args:     []string{"my-plan"},
			wantName: "my-plan",
		},
		{
			name:    "both flag and positional arg",
			flagVal: "flag-name",
			args:    []string{"arg-name"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := tt.flagVal
			err := ResolveNameArg(&val, tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.wantName {
				t.Errorf("got name %q, want %q", val, tt.wantName)
			}
		})
	}
}

func TestResolveNamesArg(t *testing.T) {
	tests := []struct {
		name      string
		flagVals  []string
		args      []string
		wantNames []string
		wantErr   bool
	}{
		{
			name:      "no args and no flag",
			flagVals:  nil,
			args:      nil,
			wantNames: nil,
		},
		{
			name:      "flag only",
			flagVals:  []string{"plan1", "plan2"},
			args:      nil,
			wantNames: []string{"plan1", "plan2"},
		},
		{
			name:      "positional arg only",
			flagVals:  nil,
			args:      []string{"my-plan"},
			wantNames: []string{"my-plan"},
		},
		{
			name:     "both flag and positional arg",
			flagVals: []string{"flag-name"},
			args:     []string{"arg-name"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals := tt.flagVals
			err := ResolveNamesArg(&vals, tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(vals) != len(tt.wantNames) {
				t.Fatalf("got %d names, want %d", len(vals), len(tt.wantNames))
			}
			for i := range vals {
				if vals[i] != tt.wantNames[i] {
					t.Errorf("name[%d] = %q, want %q", i, vals[i], tt.wantNames[i])
				}
			}
		})
	}
}

func TestMarkRequiredForMCP(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "Resource name")

	MarkRequiredForMCP(cmd, "name")

	f := cmd.Flags().Lookup("name")
	if f.Annotations == nil {
		t.Fatal("expected annotations to be set")
	}
	vals, ok := f.Annotations[MCPRequiredFlag]
	if !ok {
		t.Fatal("expected MCPRequiredFlag annotation")
	}
	if len(vals) != 1 || vals[0] != "true" {
		t.Errorf("got annotation value %v, want [\"true\"]", vals)
	}

	// Verify Cobra does NOT enforce this flag as required
	_, ok = f.Annotations[cobra.BashCompOneRequiredFlag]
	if ok {
		t.Error("MCPRequiredFlag must not set BashCompOneRequiredFlag (breaks positional args)")
	}
}
