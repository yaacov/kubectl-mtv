"""
Test cases for kubectl-mtv network and storage mapping creation from OVA providers.

This test validates the creation of network and storage mappings using OVA as the source provider.
"""

import time

import pytest

from e2e.utils import (
    wait_for_network_mapping_ready,
    wait_for_storage_mapping_ready,
)
from e2e.test_constants import TARGET_PROVIDER_NAME, OVA_NETWORKS, OVA_STORAGE


@pytest.mark.create
@pytest.mark.mapping
@pytest.mark.ova
@pytest.mark.requires_credentials
class TestOVAMappingCreation:
    """Test cases for network and storage mapping creation from OVA providers."""

    # Provider fixtures are now session-scoped in conftest.py

    def test_create_network_mapping_from_ova(self, test_namespace, ova_provider):
        """Test creating a network mapping from OVA provider."""
        mapping_name = f"test-network-map-ova-{int(time.time())}"

        # Build network pairs string
        network_pairs = ",".join([f"{n['source']}:{n['target']}" for n in OVA_NETWORKS])

        # Create network mapping command
        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {ova_provider}",
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

    def test_create_minimal_network_mapping_from_ova(
        self, test_namespace, ova_provider
    ):
        """Test creating a minimal network mapping from OVA provider with a single network."""
        mapping_name = f"test-minimal-network-map-ova-{int(time.time())}"

        # Create network mapping command with single network
        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {ova_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            "--network-pairs 'VM Network:default'",  # Single network to default is OK
        ]

        create_cmd = " ".join(cmd_parts)

        # Create network mapping
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("networkmap", mapping_name)

        # Wait for network mapping to be ready
        wait_for_network_mapping_ready(test_namespace, mapping_name)

        # Verify VM Network is in the mapping
        result = test_namespace.run_kubectl_command(
            f"get networkmap {mapping_name} -o yaml"
        )
        assert result.returncode == 0
        # Verify network mapping contains expected source network from constants
        from ...test_constants import OVA_NETWORK_PAIRS

        expected_network = OVA_NETWORK_PAIRS[0]["source"]
        assert expected_network in result.stdout

    def test_create_storage_mapping_from_ova(self, test_namespace, ova_provider):
        """Test creating a storage mapping from OVA provider."""
        mapping_name = f"test-storage-map-ova-{int(time.time())}"

        # Build storage pairs string
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in OVA_STORAGE])

        # Create storage mapping command
        cmd_parts = [
            "create mapping storage",
            mapping_name,
            f"--source {ova_provider}",
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
