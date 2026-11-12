package flags

import "testing"

func TestBuildOffloadPlugin_VSphereNfs(t *testing.T) {
	plugin, err := BuildOffloadPlugin(NfsOffloadPlugin, NfsOffloadVendor, "pure-creds", "", "Filesystem")
	if err != nil {
		t.Fatalf("BuildOffloadPlugin() error = %v", err)
	}
	if plugin.VSphereNfsCopyPluginConfig == nil {
		t.Fatal("VSphereNfsCopyPluginConfig is nil")
	}
	if plugin.VSphereXcopyPluginConfig != nil || plugin.CsiVolumeImport != nil {
		t.Fatal("vsphereNfs set another offload config")
	}
	if plugin.VSphereNfsCopyPluginConfig.SecretRef != "pure-creds" {
		t.Errorf("SecretRef = %q", plugin.VSphereNfsCopyPluginConfig.SecretRef)
	}
	if string(plugin.VSphereNfsCopyPluginConfig.StorageVendorProduct) != NfsOffloadVendor {
		t.Errorf("vendor = %q", plugin.VSphereNfsCopyPluginConfig.StorageVendorProduct)
	}
}

func TestBuildOffloadPlugin_VSphereNfsRejects(t *testing.T) {
	cases := []struct {
		name    string
		vendor  string
		hosts   string
		volMode string
	}{
		{name: "vendor", vendor: "ontap", volMode: "Filesystem"},
		{name: "hosts", vendor: NfsOffloadVendor, hosts: "host-10", volMode: "Filesystem"},
		{name: "block", vendor: NfsOffloadVendor, volMode: "Block"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := BuildOffloadPlugin(NfsOffloadPlugin, tc.vendor, "secret", tc.hosts, tc.volMode)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestBuildOffloadPlugin_VSphereKeepsMigrationHosts(t *testing.T) {
	plugin, err := BuildOffloadPlugin("vsphere", "ontap", "secret", "host-10+ host-11", "")
	if err != nil {
		t.Fatalf("BuildOffloadPlugin() error = %v", err)
	}
	hosts := plugin.VSphereXcopyPluginConfig.DedicatedMigrationHosts
	if len(hosts) != 2 || hosts[0] != "host-10" || hosts[1] != "host-11" {
		t.Fatalf("hosts = %#v", hosts)
	}
}
