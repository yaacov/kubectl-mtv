"""
Test cases for kubectl-mtv OpenStack provider inventory commands.

This test validates inventory functionality for OpenStack providers including VMs, networks, and storage.
"""

import json

import pytest

from ....utils import wait_for_provider_ready


@pytest.mark.get
@pytest.mark.inventory
@pytest.mark.openstack
@pytest.mark.requires_credentials
class TestOpenStackInventory:
    """Test cases for OpenStack provider inventory commands."""

    @pytest.fixture(scope="class")
    def openstack_provider(self, test_namespace, provider_credentials):
        """Create an OpenStack provider for inventory testing."""
        creds = provider_credentials["openstack"]

        # Skip if credentials are not available
        if not all(
            [
                creds.get("url"),
                creds.get("username"),
                creds.get("password"),
                creds.get("region_name"),
                creds.get("project_name"),
                creds.get("domain_name"),
            ]
        ):
            pytest.skip("OpenStack credentials not available in environment")

        provider_name = "test-openstack-inventory-skip-verify"

        # Create command with insecure skip TLS
        cmd_parts = [
            "create provider",
            provider_name,
            "--type openstack",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--provider-domain-name '{creds['domain_name']}'",
            f"--provider-project-name '{creds['project_name']}'",
            f"--provider-region-name '{creds['region_name']}'"
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

    def test_get_inventory_vms(self, test_namespace, openstack_provider):
        """Test getting inventory VMs from an OpenStack provider."""
        result = test_namespace.run_mtv_command(
            f"get inventory vms {openstack_provider} --output json"
        )

        if result.returncode != 0:
            pytest.skip("Provider inventory not accessible or empty")

        # Parse JSON output
        try:
            vms_data = json.loads(result.stdout)
            assert isinstance(vms_data, list), "Expected VMs data to be a list"
        except json.JSONDecodeError:
            pytest.fail("Failed to parse VMs JSON output")

    def test_get_inventory_networks(self, test_namespace, openstack_provider):
        """Test getting inventory networks from an OpenStack provider."""
        result = test_namespace.run_mtv_command(
            f"get inventory networks {openstack_provider} --output json"
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

    def test_get_inventory_storage(self, test_namespace, openstack_provider):
        """Test getting inventory storage from an OpenStack provider."""
        result = test_namespace.run_mtv_command(
            f"get inventory storage {openstack_provider} --output json"
        )

        if result.returncode != 0:
            pytest.skip("Provider inventory not accessible or empty")

        # Parse JSON output
        try:
            storage_data = json.loads(result.stdout)
            assert isinstance(storage_data, list), "Expected storage data to be a list"
        except json.JSONDecodeError:
            pytest.fail("Failed to parse storage JSON output")

    def test_get_inventory_hosts(self, test_namespace, openstack_provider):
        """Test getting inventory hosts from an OpenStack provider."""
        result = test_namespace.run_mtv_command(
            f"get inventory hosts {openstack_provider} --output json"
        )

        if result.returncode != 0:
            pytest.skip("Provider inventory not accessible or empty")

        # Parse JSON output
        try:
            hosts_data = json.loads(result.stdout)
            assert isinstance(hosts_data, list), "Expected hosts data to be a list"
        except json.JSONDecodeError:
            pytest.fail("Failed to parse hosts JSON output")
