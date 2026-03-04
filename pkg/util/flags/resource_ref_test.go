package flags

import (
	"testing"
)

func TestParseResourceRef(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		defaultNamespace string
		wantNamespace    string
		wantName         string
		wantErr          bool
	}{
		{
			name:             "plain name uses default namespace",
			input:            "my-configmap",
			defaultNamespace: "default",
			wantNamespace:    "default",
			wantName:         "my-configmap",
		},
		{
			name:             "namespace/name form",
			input:            "custom-ns/my-configmap",
			defaultNamespace: "default",
			wantNamespace:    "custom-ns",
			wantName:         "my-configmap",
		},
		{
			name:             "whitespace is trimmed",
			input:            " custom-ns / my-configmap ",
			defaultNamespace: "default",
			wantNamespace:    "custom-ns",
			wantName:         "my-configmap",
		},
		{
			name:             "empty namespace rejected",
			input:            "/my-configmap",
			defaultNamespace: "default",
			wantErr:          true,
		},
		{
			name:             "empty name rejected with slash",
			input:            "custom-ns/",
			defaultNamespace: "default",
			wantErr:          true,
		},
		{
			name:             "whitespace-only namespace rejected",
			input:            " /my-configmap",
			defaultNamespace: "default",
			wantErr:          true,
		},
		{
			name:             "whitespace-only name rejected",
			input:            "custom-ns/ ",
			defaultNamespace: "default",
			wantErr:          true,
		},
		{
			name:             "empty input rejected",
			input:            "",
			defaultNamespace: "default",
			wantErr:          true,
		},
		{
			name:             "whitespace-only input rejected",
			input:            "   ",
			defaultNamespace: "default",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns, name, err := ParseResourceRef(tt.input, tt.defaultNamespace)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got namespace=%q name=%q", tt.input, ns, name)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tt.input, err)
			}
			if ns != tt.wantNamespace {
				t.Errorf("namespace: got %q, want %q", ns, tt.wantNamespace)
			}
			if name != tt.wantName {
				t.Errorf("name: got %q, want %q", name, tt.wantName)
			}
		})
	}
}
