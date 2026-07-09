package openstack

import "testing"

func TestBuildSecret_UsesStandardCACertKey(t *testing.T) {
	secret := buildSecret("openshift-mtv", "source", "user", "pass", "https://keystone", "cert-data", "", false, "default", "demo", "regionOne")

	if got := string(secret.Data["ca.crt"]); got != "cert-data" {
		t.Fatalf("ca.crt = %q, want %q", got, "cert-data")
	}
	if _, found := secret.Data["cacert"]; found {
		t.Fatal("unexpected legacy cacert key in OpenStack secret")
	}
}
