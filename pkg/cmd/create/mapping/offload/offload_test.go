package offload

import "testing"

func TestBuildSecret_UsesForkliftStorageKeys(t *testing.T) {
	opts := SecretOptions{
		VSphereUsername: "ignored-user",
		VSpherePassword: "ignored-pass",
		VSphereURL:      "https://ignored-vcenter",
		StorageUsername: "storage-user",
		StoragePassword: "storage-pass",
		StorageEndpoint: "https://storage.example.com",
		CACert:          "test-ca-cert",
		InsecureSkipTLS: true,
	}

	secret, err := BuildSecret("openshift-mtv", "example", opts, true)
	if err != nil {
		t.Fatalf("BuildSecret() error = %v", err)
	}

	if got := string(secret.Data["STORAGE_USERNAME"]); got != "storage-user" {
		t.Fatalf("STORAGE_USERNAME = %q, want %q", got, "storage-user")
	}
	if got := string(secret.Data["STORAGE_PASSWORD"]); got != "storage-pass" {
		t.Fatalf("STORAGE_PASSWORD = %q, want %q", got, "storage-pass")
	}
	if got := string(secret.Data["STORAGE_HOSTNAME"]); got != "https://storage.example.com" {
		t.Fatalf("STORAGE_HOSTNAME = %q, want %q", got, "https://storage.example.com")
	}
	if got := string(secret.Data["STORAGE_SKIP_SSL_VERIFICATION"]); got != "true" {
		t.Fatalf("STORAGE_SKIP_SSL_VERIFICATION = %q, want %q", got, "true")
	}
	if got := string(secret.Data["ca.crt"]); got != "test-ca-cert" {
		t.Fatalf("ca.crt = %q, want %q", got, "test-ca-cert")
	}

	legacyKeys := []string{
		"user",
		"password",
		"url",
		"storageUser",
		"storagePassword",
		"storageEndpoint",
		"insecureSkipVerify",
		"cacert",
	}
	for _, key := range legacyKeys {
		if _, found := secret.Data[key]; found {
			t.Fatalf("unexpected legacy key %q found in generated secret", key)
		}
	}
}

func TestValidateSecretFields_AllowsStorageOnlyInlineCredentials(t *testing.T) {
	opts := SecretOptions{
		StorageUsername: "storage-user",
		StoragePassword: "storage-pass",
		StorageEndpoint: "https://storage.example.com",
	}

	if err := ValidateSecretFields(opts); err != nil {
		t.Fatalf("ValidateSecretFields() error = %v, want nil", err)
	}
}

func TestValidateSecretFields_RequiresStorageFields(t *testing.T) {
	opts := SecretOptions{
		VSphereUsername: "legacy-user",
	}

	err := ValidateSecretFields(opts)
	if err == nil {
		t.Fatal("ValidateSecretFields() error = nil, want missing storage fields")
	}
	if got := err.Error(); got == "" {
		t.Fatal("ValidateSecretFields() returned an empty error")
	}
}
