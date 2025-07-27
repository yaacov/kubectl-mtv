"""
Test cases for kubectl-mtv network and storage mapping creation from vSphere providers.

This test validates the creation of network and storage mappings using vSphere as the source provider
and OpenShift as the target provider.
"""

import time

import pytest

from e2e.utils import wait_for_provider_ready, wait_for_network_mapping_ready, wait_for_storage_mapping_ready


# Hardcoded network names from vSphere inventory data
VSPHERE_NETWORKS = [
    {"source": "Mgmt Network", "target": "test-nad-1"},
    {"source": "VM Network", "target": "test-nad-2"}
]

# Hardcoded storage names from vSphere inventory data  
VSPHERE_DATASTORES = [
    {"source": "nfs-us-mtv-v8", "target": "ocs-storagecluster-ceph-rbd-virtualization"},
    {"source": "nfs-us-virt", "target": "ocs-storagecluster-ceph-rbd"},
    {"source": "datastore1", "target": "csi-manila-ceph"}
]


@pytest.mark.create
@pytest.mark.mapping
@pytest.mark.vsphere
@pytest.mark.requires_credentials
class TestVSphereMappingCreation:
    """Test cases for network and storage mapping creation from vSphere providers."""

    @pytest.fixture(scope="class")
    def vsphere_provider(self, test_namespace, provider_credentials):
        """Create a vSphere provider for mapping testing."""
        creds = provider_credentials["vsphere"]

        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("VMware vSphere credentials not available in environment")

        provider_name = "test-vsphere-map-skip-verify"

        # Create command with insecure skip TLS
        cmd_parts = [
            "create provider",
            provider_name,
            "--type vsphere",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            "--provider-insecure-skip-tls",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Wait for provider to be ready
        wait_for_provider_ready(test_namespace, provider_name)

        return provider_name

    def test_create_network_mapping_from_vsphere(self, test_namespace, vsphere_provider):
        """Test creating a network mapping from vSphere provider."""
        mapping_name = f"test-network-map-vsphere-{int(time.time())}"
        
        # Build network pairs string
        network_pairs = ",".join([f"{n['source']}:{n['target']}" for n in VSPHERE_NETWORKS])
        
        # Create network mapping command
        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {vsphere_provider}",
            "--target test-openshift-target",
            f"--network-pairs '{network_pairs}'",
        ]
        
        create_cmd = " ".join(cmd_parts)
        
        # Create network mapping
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("networkmap", mapping_name)
        
        # Wait for network mapping to be ready
        wait_for_network_mapping_ready(test_namespace, mapping_name)

    def test_create_storage_mapping_from_vsphere(self, test_namespace, vsphere_provider):
        """Test creating a storage mapping from vSphere provider."""
        mapping_name = f"test-storage-map-vsphere-{int(time.time())}"
        
        # Build storage pairs string
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in VSPHERE_DATASTORES])
        
        # Create storage mapping command
        cmd_parts = [
            "create mapping storage",
            mapping_name,
            f"--source {vsphere_provider}",
            "--target test-openshift-target",
            f"--storage-pairs '{storage_pairs}'",
        ]
        
        create_cmd = " ".join(cmd_parts)
        
        # Create storage mapping
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("storagemap", mapping_name)
        
        # Wait for storage mapping to be ready
        wait_for_storage_mapping_ready(test_namespace, mapping_name)

 