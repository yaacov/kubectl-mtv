package openshift

import "testing"

func TestBuildSecret_UsesStandardCACertKey(t *testing.T) {
	secret := buildSecret("openshift-mtv", "target", "https://api.cluster", "token", "cert-data", false)

	if got := string(secret.Data["ca.crt"]); got != "cert-data" {
		t.Fatalf("ca.crt = %q, want %q", got, "cert-data")
	}
	if _, found := secret.Data["cacert"]; found {
		t.Fatal("unexpected legacy cacert key in OpenShift secret")
	}
}
