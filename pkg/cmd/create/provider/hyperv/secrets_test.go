package hyperv

import (
	"testing"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/provider/providerutil"
)

func TestBuildSecret_UsesStandardCACertKey(t *testing.T) {
	secret, err := buildSecret("openshift-mtv", "source", providerutil.ProviderOptions{
		Username: "user",
		Password: "pass",
		CACert:   "cert-data",
	})
	if err != nil {
		t.Fatalf("buildSecret() error = %v", err)
	}

	if got := string(secret.Data["ca.crt"]); got != "cert-data" {
		t.Fatalf("ca.crt = %q, want %q", got, "cert-data")
	}
	if _, found := secret.Data["cacert"]; found {
		t.Fatal("unexpected legacy cacert key in Hyper-V secret")
	}
}
