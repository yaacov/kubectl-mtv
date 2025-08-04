"""
Test cases for kubectl-mtv network and storage mapping creation from ESXi providers.

This test validates the creation of network and storage mappings using ESXi as the source provider.
"""

import time

import pytest

from e2e.utils import (
    wait_for_network_mapping_ready,
    wait_for_storage_mapping_ready,
    generate_provider_name,
    get_or_create_provider,
)
from e2e.test_constants import TARGET_PROVIDER_NAME, ESXI_NETWORKS, ESXI_DATASTORES


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

        # Generate provider name based on type and configuration (ESXi uses vsphere type without sdk-endpoint flag in mapping tests)
        provider_name = generate_provider_name("vsphere", creds["url"], skip_tls=True)

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

        # Create provider if it doesn't already exist
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

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
            f"--target {TARGET_PROVIDER_NAME}",
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
            f"--target {TARGET_PROVIDER_NAME}",
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
            f"--target {TARGET_PROVIDER_NAME}",
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
            f"--target {TARGET_PROVIDER_NAME}",
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
