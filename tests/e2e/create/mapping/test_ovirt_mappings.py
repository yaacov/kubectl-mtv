"""
Test cases for kubectl-mtv network and storage mapping creation from oVirt providers.

This test validates the creation of network and storage mappings using oVirt as the source provider.
"""

import time

import pytest

from e2e.utils import (
    wait_for_network_mapping_ready,
    wait_for_storage_mapping_ready,
    generate_provider_name,
    get_or_create_provider,
)
from e2e.test_constants import TARGET_PROVIDER_NAME, OVIRT_NETWORKS, OVIRT_DATASTORES


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

        # Generate provider name based on type and configuration
        provider_name = generate_provider_name("ovirt", creds["url"], skip_tls=True)

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

        # Create provider if it doesn't already exist
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    def test_create_network_mapping_from_ovirt(self, test_namespace, ovirt_provider):
        """Test creating a network mapping from oVirt provider."""
        mapping_name = f"test-network-map-ovirt-{int(time.time())}"

        # Use first two networks for basic test
        network_pairs = ",".join(
            [f"{n['source']}:{n['target']}" for n in OVIRT_NETWORKS[:2]]
        )

        # Create network mapping command
        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {ovirt_provider}",
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

    def test_create_storage_mapping_from_ovirt(self, test_namespace, ovirt_provider):
        """Test creating a storage mapping from oVirt provider."""
        mapping_name = f"test-storage-map-ovirt-{int(time.time())}"

        # Use first three storage domains for basic test
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OVIRT_DATASTORES[:3]]
        )

        # Create storage mapping command
        cmd_parts = [
            "create mapping storage",
            mapping_name,
            f"--source {ovirt_provider}",
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

    def test_create_complex_network_mapping_from_ovirt(
        self, test_namespace, ovirt_provider
    ):
        """Test creating a complex network mapping with multiple networks from oVirt provider."""
        mapping_name = f"test-complex-network-map-ovirt-{int(time.time())}"

        # Use all networks for complex test
        network_pairs = ",".join(
            [f"{n['source']}:{n['target']}" for n in OVIRT_NETWORKS]
        )

        # Create network mapping command
        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {ovirt_provider}",
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

        # Verify all source networks are in the mapping
        result = test_namespace.run_kubectl_command(
            f"get networkmap {mapping_name} -o yaml"
        )
        assert result.returncode == 0
        for network in OVIRT_NETWORKS:
            assert network["source"] in result.stdout
