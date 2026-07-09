package host

import "testing"

func TestBuildHostSecretObject_UsesStandardCACertKey(t *testing.T) {
	secret := buildHostSecretObject("openshift-mtv", "esxi-host", "user", "pass", false, "cert-data", true)

	if got := string(secret.Data["ca.crt"]); got != "cert-data" {
		t.Fatalf("ca.crt = %q, want %q", got, "cert-data")
	}
	if _, found := secret.Data["cacert"]; found {
		t.Fatal("unexpected legacy cacert key in host secret")
	}
}
