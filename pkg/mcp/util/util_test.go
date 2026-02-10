package util

import (
	"os"
	"testing"
)

func TestResolveEnvVar(t *testing.T) {
	// Set test environment variables
	os.Setenv("TEST_HOST", "https://10.6.46.250")
	os.Setenv("TEST_PORT", "443")
	os.Setenv("TEST_USER", "admin")
	defer func() {
		os.Unsetenv("TEST_HOST")
		os.Unsetenv("TEST_PORT")
		os.Unsetenv("TEST_USER")
	}()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "whole value env var reference",
			input: "${TEST_HOST}",
			want:  "https://10.6.46.250",
		},
		{
			name:  "embedded env var with path suffix",
			input: "${TEST_HOST}/sdk",
			want:  "https://10.6.46.250/sdk",
		},
		{
			name:  "embedded env var with prefix and suffix",
			input: "https://${TEST_USER}@example.com",
			want:  "https://admin@example.com",
		},
		{
			name:  "multiple env var references",
			input: "${TEST_HOST}:${TEST_PORT}/api",
			want:  "https://10.6.46.250:443/api",
		},
		{
			name:  "literal string no env vars",
			input: "https://example.com/sdk",
			want:  "https://example.com/sdk",
		},
		{
			name:  "literal dollar sign without braces",
			input: "$ecureP@ss",
			want:  "$ecureP@ss",
		},
		{
			name:  "bare $VAR without braces passes through",
			input: "$TEST_HOST",
			want:  "$TEST_HOST",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:    "unset env var returns error",
			input:   "${NONEXISTENT_VAR_12345}",
			wantErr: true,
		},
		{
			name:  "empty env var name passes through (not a valid pattern)",
			input: "${}",
			want:  "${}",
		},
		{
			name:    "unset env var embedded in string returns error",
			input:   "${NONEXISTENT_VAR_12345}/sdk",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveEnvVar(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("resolveEnvVar(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("resolveEnvVar(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("resolveEnvVar(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveEnvVars(t *testing.T) {
	os.Setenv("TEST_URL", "https://vcenter.example.com")
	os.Setenv("TEST_PASS", "s3cret")
	defer func() {
		os.Unsetenv("TEST_URL")
		os.Unsetenv("TEST_PASS")
	}()

	tests := []struct {
		name    string
		args    []string
		want    []string
		wantErr bool
	}{
		{
			name: "flags are skipped, values are resolved",
			args: []string{"--url", "${TEST_URL}/sdk", "--password", "${TEST_PASS}"},
			want: []string{"--url", "https://vcenter.example.com/sdk", "--password", "s3cret"},
		},
		{
			name: "non-env values pass through",
			args: []string{"--type", "vsphere", "--url", "https://literal.example.com"},
			want: []string{"--type", "vsphere", "--url", "https://literal.example.com"},
		},
		{
			name: "positional args with env vars",
			args: []string{"create", "provider", "my-provider", "--url", "${TEST_URL}"},
			want: []string{"create", "provider", "my-provider", "--url", "https://vcenter.example.com"},
		},
		{
			name:    "error on unset env var",
			args:    []string{"--url", "${DOES_NOT_EXIST_XYZ}"},
			wantErr: true,
		},
		{
			name: "short flags are skipped",
			args: []string{"-n", "demo", "-p", "${TEST_PASS}"},
			want: []string{"-n", "demo", "-p", "s3cret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveEnvVars(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveEnvVars(%v) expected error, got nil", tt.args)
				}
				return
			}
			if err != nil {
				t.Errorf("ResolveEnvVars(%v) unexpected error: %v", tt.args, err)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("ResolveEnvVars(%v) = %v (len %d), want %v (len %d)", tt.args, got, len(got), tt.want, len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ResolveEnvVars(%v)[%d] = %q, want %q", tt.args, i, got[i], tt.want[i])
				}
			}
		})
	}
}
