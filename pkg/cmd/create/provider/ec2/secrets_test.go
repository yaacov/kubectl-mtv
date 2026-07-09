package ec2

import "testing"

func TestBuildSecret_UsesStandardCACertKey(t *testing.T) {
	secret := buildSecret("openshift-mtv", "source", "ak", "sk", "https://ec2", "cert-data", "us-east-1", false, "", "")

	if got := string(secret.Data["ca.crt"]); got != "cert-data" {
		t.Fatalf("ca.crt = %q, want %q", got, "cert-data")
	}
	if _, found := secret.Data["cacert"]; found {
		t.Fatal("unexpected legacy cacert key in EC2 secret")
	}
}
