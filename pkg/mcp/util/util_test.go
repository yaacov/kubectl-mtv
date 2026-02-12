package util

import (
	"context"
	"net/http"
	"os"
	"strings"
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

// --- UnmarshalJSONResponse tests ---

func TestUnmarshalJSONResponse(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantData   bool // expect "data" key in result
		wantOutput bool // expect "output" key in result
		wantErr    bool
		checkData  func(t *testing.T, data interface{})
	}{
		{
			name:     "stdout is JSON object",
			input:    `{"command":"kubectl-mtv get plan","return_value":0,"stdout":"{\"items\":[],\"kind\":\"PlanList\"}","stderr":""}`,
			wantData: true,
			checkData: func(t *testing.T, data interface{}) {
				m, ok := data.(map[string]interface{})
				if !ok {
					t.Fatalf("expected map, got %T", data)
				}
				if _, exists := m["kind"]; !exists {
					t.Error("parsed data should contain 'kind'")
				}
			},
		},
		{
			name:     "stdout is JSON array",
			input:    `{"command":"test","return_value":0,"stdout":"[{\"name\":\"plan1\"},{\"name\":\"plan2\"}]","stderr":""}`,
			wantData: true,
			checkData: func(t *testing.T, data interface{}) {
				arr, ok := data.([]interface{})
				if !ok {
					t.Fatalf("expected array, got %T", data)
				}
				if len(arr) != 2 {
					t.Errorf("expected 2 items, got %d", len(arr))
				}
			},
		},
		{
			name:       "stdout is plain text",
			input:      `{"command":"test","return_value":0,"stdout":"NAME    STATUS\nplan1   Ready","stderr":""}`,
			wantOutput: true,
		},
		{
			name:  "empty stdout",
			input: `{"command":"test","return_value":0,"stdout":"","stderr":""}`,
		},
		{
			name:    "invalid JSON input",
			input:   `not json at all`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := UnmarshalJSONResponse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantData {
				data, exists := result["data"]
				if !exists {
					t.Fatal("expected 'data' key in result")
				}
				if _, hasStdout := result["stdout"]; hasStdout {
					t.Error("'stdout' should be removed when 'data' is set")
				}
				if tt.checkData != nil {
					tt.checkData(t, data)
				}
			}

			if tt.wantOutput {
				output, exists := result["output"]
				if !exists {
					t.Fatal("expected 'output' key in result")
				}
				if _, ok := output.(string); !ok {
					t.Errorf("output should be a string, got %T", output)
				}
				if _, hasStdout := result["stdout"]; hasStdout {
					t.Error("'stdout' should be removed when 'output' is set")
				}
			}
		})
	}
}

// --- formatShellCommand tests ---

func TestFormatShellCommand(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		args        []string
		wantContain string
		wantMissing string
	}{
		{
			name:        "simple command",
			cmd:         "kubectl-mtv",
			args:        []string{"get", "plan", "-n", "demo"},
			wantContain: "kubectl-mtv get plan -n demo",
		},
		{
			name:        "password is sanitized",
			cmd:         "kubectl-mtv",
			args:        []string{"create", "provider", "--password", "s3cret"},
			wantContain: "--password",
			wantMissing: "s3cret",
		},
		{
			name:        "token is sanitized",
			cmd:         "kubectl",
			args:        []string{"--token", "my-secret-token", "get", "pods"},
			wantContain: "--token",
			wantMissing: "my-secret-token",
		},
		{
			name:        "offload password is sanitized",
			cmd:         "kubectl-mtv",
			args:        []string{"create", "mapping", "--offload-vsphere-password", "pass123"},
			wantContain: "--offload-vsphere-password",
			wantMissing: "pass123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatShellCommand(tt.cmd, tt.args)

			if tt.wantContain != "" && !strings.Contains(result, tt.wantContain) {
				t.Errorf("formatShellCommand() = %q, should contain %q", result, tt.wantContain)
			}
			if tt.wantMissing != "" && strings.Contains(result, tt.wantMissing) {
				t.Errorf("formatShellCommand() = %q, should NOT contain %q", result, tt.wantMissing)
			}
		})
	}
}

// --- Context helpers tests ---

func TestWithKubeToken(t *testing.T) {
	ctx := context.Background()

	// No token in fresh context
	token, ok := GetKubeToken(ctx)
	if ok || token != "" {
		t.Error("fresh context should not have a token")
	}

	// Add token
	ctx = WithKubeToken(ctx, "my-token")
	token, ok = GetKubeToken(ctx)
	if !ok || token != "my-token" {
		t.Errorf("expected token 'my-token', got %q (ok=%v)", token, ok)
	}
}

func TestWithKubeServer(t *testing.T) {
	ctx := context.Background()

	// No server in fresh context
	server, ok := GetKubeServer(ctx)
	if ok || server != "" {
		t.Error("fresh context should not have a server")
	}

	// Add server
	ctx = WithKubeServer(ctx, "https://api.example.com:6443")
	server, ok = GetKubeServer(ctx)
	if !ok || server != "https://api.example.com:6443" {
		t.Errorf("expected server URL, got %q (ok=%v)", server, ok)
	}
}

func TestWithDryRun(t *testing.T) {
	ctx := context.Background()

	// Default is false
	if GetDryRun(ctx) {
		t.Error("fresh context should not have dry run enabled")
	}

	// Enable dry run
	ctx = WithDryRun(ctx, true)
	if !GetDryRun(ctx) {
		t.Error("dry run should be enabled after setting true")
	}

	// Disable dry run
	ctx = WithDryRun(ctx, false)
	if GetDryRun(ctx) {
		t.Error("dry run should be disabled after setting false")
	}
}

func TestGetKubeToken_NilContext(t *testing.T) {
	token, ok := GetKubeToken(nil) //nolint:staticcheck // intentionally testing nil context behavior
	if ok || token != "" {
		t.Error("nil context should return empty token")
	}
}

func TestGetKubeServer_NilContext(t *testing.T) {
	server, ok := GetKubeServer(nil) //nolint:staticcheck // intentionally testing nil context behavior
	if ok || server != "" {
		t.Error("nil context should return empty server")
	}
}

func TestGetDryRun_NilContext(t *testing.T) {
	if GetDryRun(nil) { //nolint:staticcheck // intentionally testing nil context behavior
		t.Error("nil context should return false for dry run")
	}
}

// --- WithKubeCredsFromHeaders tests ---

func TestWithKubeCredsFromHeaders(t *testing.T) {
	tests := []struct {
		name       string
		headers    http.Header
		wantToken  string
		wantServer string
	}{
		{
			name:    "nil headers",
			headers: nil,
		},
		{
			name:    "empty headers",
			headers: http.Header{},
		},
		{
			name: "bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer my-k8s-token"},
			},
			wantToken: "my-k8s-token",
		},
		{
			name: "kubernetes server",
			headers: http.Header{
				"X-Kubernetes-Server": []string{"https://api.cluster.local:6443"},
			},
			wantServer: "https://api.cluster.local:6443",
		},
		{
			name: "both token and server",
			headers: http.Header{
				"Authorization":       []string{"Bearer both-token"},
				"X-Kubernetes-Server": []string{"https://api.example.com"},
			},
			wantToken:  "both-token",
			wantServer: "https://api.example.com",
		},
		{
			name: "non-bearer auth header ignored",
			headers: http.Header{
				"Authorization": []string{"Basic dXNlcjpwYXNz"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = WithKubeCredsFromHeaders(ctx, tt.headers)

			token, _ := GetKubeToken(ctx)
			if token != tt.wantToken {
				t.Errorf("token = %q, want %q", token, tt.wantToken)
			}

			server, _ := GetKubeServer(ctx)
			if server != tt.wantServer {
				t.Errorf("server = %q, want %q", server, tt.wantServer)
			}
		})
	}
}

// --- Default kube credentials tests ---

func TestDefaultKubeServer(t *testing.T) {
	// Save and restore
	orig := GetDefaultKubeServer()
	defer SetDefaultKubeServer(orig)

	// Default is empty
	SetDefaultKubeServer("")
	if got := GetDefaultKubeServer(); got != "" {
		t.Errorf("GetDefaultKubeServer() should be empty, got %q", got)
	}

	// Set a server
	SetDefaultKubeServer("https://api.example.com:6443")
	if got := GetDefaultKubeServer(); got != "https://api.example.com:6443" {
		t.Errorf("GetDefaultKubeServer() = %q, want %q", got, "https://api.example.com:6443")
	}

	// Clear it
	SetDefaultKubeServer("")
	if got := GetDefaultKubeServer(); got != "" {
		t.Errorf("GetDefaultKubeServer() should be empty after clearing, got %q", got)
	}
}

func TestDefaultKubeToken(t *testing.T) {
	// Save and restore
	orig := GetDefaultKubeToken()
	defer SetDefaultKubeToken(orig)

	// Default is empty
	SetDefaultKubeToken("")
	if got := GetDefaultKubeToken(); got != "" {
		t.Errorf("GetDefaultKubeToken() should be empty, got %q", got)
	}

	// Set a token
	SetDefaultKubeToken("my-cli-token")
	if got := GetDefaultKubeToken(); got != "my-cli-token" {
		t.Errorf("GetDefaultKubeToken() = %q, want %q", got, "my-cli-token")
	}

	// Clear it
	SetDefaultKubeToken("")
	if got := GetDefaultKubeToken(); got != "" {
		t.Errorf("GetDefaultKubeToken() should be empty after clearing, got %q", got)
	}
}

func TestRunKubectlMTVCommand_DefaultCredsFallback(t *testing.T) {
	// Save and restore defaults
	origServer := GetDefaultKubeServer()
	origToken := GetDefaultKubeToken()
	defer func() {
		SetDefaultKubeServer(origServer)
		SetDefaultKubeToken(origToken)
	}()

	// Set CLI defaults
	SetDefaultKubeServer("https://cli-default.example.com:6443")
	SetDefaultKubeToken("cli-default-token")

	// Use dry run mode to capture the command without executing
	ctx := WithDryRun(context.Background(), true)

	result, err := RunKubectlMTVCommand(ctx, []string{"get", "plan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify CLI defaults are used (token is sanitized in output)
	if !strings.Contains(result, "--server") {
		t.Errorf("expected --server flag in command, got: %s", result)
	}
	if !strings.Contains(result, "cli-default.example.com") {
		t.Errorf("expected CLI default server URL in command, got: %s", result)
	}
	if !strings.Contains(result, "--token") {
		t.Errorf("expected --token flag in command, got: %s", result)
	}
}

func TestRunKubectlMTVCommand_ContextOverridesDefaults(t *testing.T) {
	// Save and restore defaults
	origServer := GetDefaultKubeServer()
	origToken := GetDefaultKubeToken()
	defer func() {
		SetDefaultKubeServer(origServer)
		SetDefaultKubeToken(origToken)
	}()

	// Set CLI defaults
	SetDefaultKubeServer("https://cli-default.example.com:6443")
	SetDefaultKubeToken("cli-default-token")

	// Set context values (simulating HTTP headers) that should override CLI defaults
	ctx := context.Background()
	ctx = WithKubeServer(ctx, "https://header-override.example.com:6443")
	ctx = WithKubeToken(ctx, "header-override-token")
	ctx = WithDryRun(ctx, true)

	result, err := RunKubectlMTVCommand(ctx, []string{"get", "plan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Context values should be used, not CLI defaults
	if !strings.Contains(result, "header-override.example.com") {
		t.Errorf("expected context server URL to override CLI default, got: %s", result)
	}
	if strings.Contains(result, "cli-default.example.com") {
		t.Errorf("CLI default server should NOT appear when context has server, got: %s", result)
	}
}

func TestRunKubectlMTVCommand_NoCredsWhenNoneSet(t *testing.T) {
	// Save and restore defaults
	origServer := GetDefaultKubeServer()
	origToken := GetDefaultKubeToken()
	defer func() {
		SetDefaultKubeServer(origServer)
		SetDefaultKubeToken(origToken)
	}()

	// Clear CLI defaults
	SetDefaultKubeServer("")
	SetDefaultKubeToken("")

	// Use dry run mode with no context credentials
	ctx := WithDryRun(context.Background(), true)

	result, err := RunKubectlMTVCommand(ctx, []string{"get", "plan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No --server or --token should appear
	if strings.Contains(result, "--server") {
		t.Errorf("--server flag should NOT appear when no creds set, got: %s", result)
	}
	if strings.Contains(result, "--token") {
		t.Errorf("--token flag should NOT appear when no creds set, got: %s", result)
	}
}

func TestRunKubectlCommand_DefaultCredsFallback(t *testing.T) {
	// Save and restore defaults
	origServer := GetDefaultKubeServer()
	origToken := GetDefaultKubeToken()
	defer func() {
		SetDefaultKubeServer(origServer)
		SetDefaultKubeToken(origToken)
	}()

	// Set CLI defaults
	SetDefaultKubeServer("https://cli-default.example.com:6443")
	SetDefaultKubeToken("cli-default-token")

	// Use dry run mode to capture the command without executing
	ctx := WithDryRun(context.Background(), true)

	result, err := RunKubectlCommand(ctx, []string{"get", "pods"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify CLI defaults are used
	if !strings.Contains(result, "--server") {
		t.Errorf("expected --server flag in command, got: %s", result)
	}
	if !strings.Contains(result, "cli-default.example.com") {
		t.Errorf("expected CLI default server URL in command, got: %s", result)
	}
	if !strings.Contains(result, "--token") {
		t.Errorf("expected --token flag in command, got: %s", result)
	}
}

func TestRunKubectlCommand_ContextOverridesDefaults(t *testing.T) {
	// Save and restore defaults
	origServer := GetDefaultKubeServer()
	origToken := GetDefaultKubeToken()
	defer func() {
		SetDefaultKubeServer(origServer)
		SetDefaultKubeToken(origToken)
	}()

	// Set CLI defaults
	SetDefaultKubeServer("https://cli-default.example.com:6443")
	SetDefaultKubeToken("cli-default-token")

	// Set context values (simulating HTTP headers) that should override CLI defaults
	ctx := context.Background()
	ctx = WithKubeServer(ctx, "https://header-override.example.com:6443")
	ctx = WithKubeToken(ctx, "header-override-token")
	ctx = WithDryRun(ctx, true)

	result, err := RunKubectlCommand(ctx, []string{"get", "pods"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Context values should be used, not CLI defaults
	if !strings.Contains(result, "header-override.example.com") {
		t.Errorf("expected context server URL to override CLI default, got: %s", result)
	}
	if strings.Contains(result, "cli-default.example.com") {
		t.Errorf("CLI default server should NOT appear when context has server, got: %s", result)
	}
}

// --- Output format tests ---

func TestOutputFormat(t *testing.T) {
	// Save and restore
	orig := GetOutputFormat()
	defer SetOutputFormat(orig)

	// Default is "text"
	SetOutputFormat("")
	if got := GetOutputFormat(); got != "text" {
		t.Errorf("SetOutputFormat(\"\") should default to \"text\", got %q", got)
	}

	// Set to json
	SetOutputFormat("json")
	if got := GetOutputFormat(); got != "json" {
		t.Errorf("SetOutputFormat(\"json\") = %q, want \"json\"", got)
	}

	// Set back to text
	SetOutputFormat("text")
	if got := GetOutputFormat(); got != "text" {
		t.Errorf("SetOutputFormat(\"text\") = %q, want \"text\"", got)
	}
}
