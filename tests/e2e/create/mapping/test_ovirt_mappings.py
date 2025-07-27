"""
Test cases for kubectl-mtv network and storage mapping creation from oVirt providers.

This test validates the creation of network and storage mappings using oVirt as the source provider
and OpenShift as the target provider.
"""

import time

import pytest

from e2e.utils import wait_for_provider_ready, wait_for_network_mapping_ready, wait_for_storage_mapping_ready


# Hardcoded network names from oVirt inventory data
OVIRT_NETWORKS = [
    {"source": "ovirtmgmt", "target": "test-nad-1"},
    {"source": "vm", "target": "test-nad-2"},
    {"source": "internal", "target": "test-nad-1"},
    {"source": "vlan10", "target": "test-nad-2"}
]

# Hardcoded storage names from oVirt inventory data  
OVIRT_STORAGE_DOMAINS = [
    {"source": "hosted_storage", "target": "ocs-storagecluster-ceph-rbd-virtualization"},
    {"source": "L0_Group_4_LUN1", "target": "ocs-storagecluster-ceph-rbd"},
    {"source": "L0_Group_4_LUN2", "target": "csi-manila-ceph"},
    {"source": "L0_Group_4_LUN3", "target": "csi-manila-netapp"},
    {"source": "export2", "target": "ocs-storagecluster-ceph-rbd"}
]


@pytest.mark.create
@pytest.mark.mapping
@pytest.mark.ovirt
@pytest.mark.requires_credentials
class TestOvirtMappingCreation:
    """Test cases for network and storage mapping creation from oVirt providers."""

    @pytest.fixture(scope="class")
    def ovirt_provider(self, test_namespace, provider_credentials):
        """Create an oVirt provider for mapping testing."""
        creds = provider_credentials["ovirt"]

        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("oVirt credentials not available in environment")

        provider_name = "test-ovirt-map-skip-verify"

        # Create command with insecure skip TLS
        cmd_parts = [
            "create provider",
            provider_name,
            "--type ovirt",
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

    def test_create_network_mapping_from_ovirt(self, test_namespace, ovirt_provider):
        """Test creating a network mapping from oVirt provider."""
        mapping_name = f"test-network-map-ovirt-{int(time.time())}"
        
        # Use first two networks for basic test
        network_pairs = ",".join([f"{n['source']}:{n['target']}" for n in OVIRT_NETWORKS[:2]])
        
        # Create network mapping command
        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {ovirt_provider}",
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

    def test_create_storage_mapping_from_ovirt(self, test_namespace, ovirt_provider):
        """Test creating a storage mapping from oVirt provider."""
        mapping_name = f"test-storage-map-ovirt-{int(time.time())}"
        
        # Use first three storage domains for basic test
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in OVIRT_STORAGE_DOMAINS[:3]])
        
        # Create storage mapping command
        cmd_parts = [
            "create mapping storage",
            mapping_name,
            f"--source {ovirt_provider}",
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

    def test_create_complex_network_mapping_from_ovirt(self, test_namespace, ovirt_provider):
        """Test creating a complex network mapping with multiple networks from oVirt provider."""
        mapping_name = f"test-complex-network-map-ovirt-{int(time.time())}"
        
        # Use all networks for complex test
        network_pairs = ",".join([f"{n['source']}:{n['target']}" for n in OVIRT_NETWORKS])
        
        # Create network mapping command
        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {ovirt_provider}",
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
        
        # Verify all source networks are in the mapping
        result = test_namespace.run_kubectl_command(f"get networkmap {mapping_name} -o yaml")
        assert result.returncode == 0
        for network in OVIRT_NETWORKS:
            assert network["source"] in result.stdout

 