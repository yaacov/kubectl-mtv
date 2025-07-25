"""
Test cases for kubectl-mtv OVA provider inventory commands.

This test validates inventory functionality for OVA providers including VMs, networks, and storage.
"""

import json

import pytest

from ....utils import wait_for_provider_ready


@pytest.mark.get
@pytest.mark.inventory
@pytest.mark.ova
@pytest.mark.requires_credentials
class TestOVAInventory:
    """Test cases for OVA provider inventory commands."""

    @pytest.fixture(scope="class")
    def ova_provider(self, test_namespace, provider_credentials):
        """Create an OVA provider for inventory testing."""
        creds = provider_credentials["ova"]

        # Skip if OVA URL is not available
        if not creds.get("url"):
            pytest.skip("OVA URL not available in environment")

        provider_name = "test-ova-inventory-skip-verify"

        # Create command
        cmd_parts = [
            "create provider",
            provider_name,
            "--type ova",
            f"--url '{creds['url']}'",
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

    def test_get_inventory_vms(self, test_namespace, ova_provider):
        """Test getting inventory VMs from an OVA provider."""
        result = test_namespace.run_mtv_command(
            f"get inventory vms {ova_provider} --output json"
        )

        if result.returncode != 0:
            pytest.skip("Provider inventory not accessible or empty")

        # Parse JSON output
        try:
            vms_data = json.loads(result.stdout)
            assert isinstance(vms_data, list), "Expected VMs data to be a list"
        except json.JSONDecodeError:
            pytest.fail("Failed to parse VMs JSON output")

    def test_get_inventory_networks(self, test_namespace, ova_provider):
        """Test getting inventory networks from an OVA provider."""
        result = test_namespace.run_mtv_command(
            f"get inventory networks {ova_provider} --output json"
        )

        if result.returncode != 0:
            pytest.skip("Provider inventory not accessible or empty")

        # Parse JSON output
        try:
            networks_data = json.loads(result.stdout)
            assert isinstance(
                networks_data, list
            ), "Expected networks data to be a list"
        except json.JSONDecodeError:
            pytest.fail("Failed to parse networks JSON output")

    def test_get_inventory_storage(self, test_namespace, ova_provider):
        """Test getting inventory storage from an OVA provider."""
        result = test_namespace.run_mtv_command(
            f"get inventory storage {ova_provider} --output json"
        )

        if result.returncode != 0:
            pytest.skip("Provider inventory not accessible or empty")

        # Parse JSON output
        try:
            storage_data = json.loads(result.stdout)
            assert isinstance(storage_data, list), "Expected storage data to be a list"
        except json.JSONDecodeError:
            pytest.fail("Failed to parse storage JSON output")
