"""
Test cases for kubectl-mtv network and storage mapping creation from ESXi providers.

This test validates the creation of network and storage mappings using ESXi as the source provider
and OpenShift as the target provider.
"""

import time

import pytest

from e2e.utils import (
    wait_for_provider_ready,
    wait_for_network_mapping_ready,
    wait_for_storage_mapping_ready,
)


# Hardcoded network names from ESXi inventory data
ESXI_NETWORKS = [
    {"source": "Mgmt Network", "target": "test-nad-1"},
    {"source": "VM Network", "target": "test-nad-2"},
]

# Hardcoded storage names from ESXi inventory data
ESXI_DATASTORES = [
    {
        "source": "mtv-nfs-rhos-v8",
        "target": "ocs-storagecluster-ceph-rbd-virtualization",
    },
    {"source": "nfs-us", "target": "ocs-storagecluster-ceph-rbd"},
    {"source": "mtv-nfs-us-v8", "target": "csi-manila-ceph"},
]


@pytest.mark.create
@pytest.mark.mapping
@pytest.mark.esxi
@pytest.mark.requires_credentials
class TestESXiMappingCreation:
    """Test cases for network and storage mapping creation from ESXi providers."""

    @pytest.fixture(scope="class")
    def esxi_provider(self, test_namespace, provider_credentials):
        """Create an ESXi provider for mapping testing."""
        creds = provider_credentials["esxi"]

        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("ESXi credentials not available in environment")

        provider_name = "test-esxi-map-skip-verify"

        # Create command with insecure skip TLS
        cmd_parts = [
            "create provider",
            provider_name,
            "--type vsphere",  # ESXi uses vsphere type
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

    def test_create_network_mapping_from_esxi(self, test_namespace, esxi_provider):
        """Test creating a network mapping from ESXi provider."""
        mapping_name = f"test-network-map-esxi-{int(time.time())}"

        # Build network pairs string
        network_pairs = ",".join(
            [f"{n['source']}:{n['target']}" for n in ESXI_NETWORKS]
        )

        # Create network mapping command
        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {esxi_provider}",
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

    def test_create_storage_mapping_from_esxi(self, test_namespace, esxi_provider):
        """Test creating a storage mapping from ESXi provider."""
        mapping_name = f"test-storage-map-esxi-{int(time.time())}"

        # Build storage pairs string
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in ESXI_DATASTORES]
        )

        # Create storage mapping command
        cmd_parts = [
            "create mapping storage",
            mapping_name,
            f"--source {esxi_provider}",
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

    def test_create_mapping_with_vddk_option(self, test_namespace, esxi_provider):
        """Test creating mappings that might be used with VDDK acceleration."""
        # First create network mapping
        network_map_name = f"test-vddk-network-map-esxi-{int(time.time())}"

        # Create network mapping command
        cmd_parts = [
            "create mapping network",
            network_map_name,
            f"--source {esxi_provider}",
            "--target test-openshift-target",
            "--network-pairs 'VM Network:default'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create network mapping
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("networkmap", network_map_name)

        # Create storage mapping
        storage_map_name = f"test-vddk-storage-map-esxi-{int(time.time())}"

        cmd_parts = [
            "create mapping storage",
            storage_map_name,
            f"--source {esxi_provider}",
            "--target test-openshift-target",
            "--storage-pairs 'mtv-nfs-rhos-v8:ocs-storagecluster-ceph-rbd-virtualization'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create storage mapping
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("storagemap", storage_map_name)

        # Wait for both mappings to be ready
        wait_for_network_mapping_ready(test_namespace, network_map_name)
        wait_for_storage_mapping_ready(test_namespace, storage_map_name)
