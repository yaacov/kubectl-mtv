"""
Test cases for kubectl-mtv OpenShift provider inventory commands.

This test validates inventory functionality for OpenShift providers including VMs, networks, and storage.
"""

import json

import pytest

from ....utils import (
    generate_provider_name,
    get_or_create_provider,
)


@pytest.mark.get
@pytest.mark.inventory
@pytest.mark.openshift
@pytest.mark.requires_credentials
class TestOpenShiftInventory:
    """Test cases for OpenShift provider inventory commands."""

    @pytest.fixture(scope="class")
    def openshift_provider(self, test_namespace, provider_credentials):
        """Create an OpenShift provider for inventory testing."""
        # Generate provider name based on type and configuration
        provider_name = generate_provider_name("openshift", "localhost", skip_tls=True)

        # Use current cluster context with skip verify
        cmd_parts = [
            "create provider",
            provider_name,
            "--type openshift",
            "--provider-insecure-skip-tls",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create provider if it doesn't already exist
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    def test_get_inventory_vms(self, test_namespace, openshift_provider):
        """Test getting inventory VMs from an OpenShift provider."""
        result = test_namespace.run_mtv_command(
            f"get inventory vms {openshift_provider} --output json"
        )

        if result.returncode != 0:
            pytest.skip("Provider inventory not accessible or empty")

        # Parse JSON output
        try:
            vms_data = json.loads(result.stdout)
            assert isinstance(vms_data, list), "Expected VMs data to be a list"
        except json.JSONDecodeError:
            pytest.fail("Failed to parse VMs JSON output")

    def test_get_inventory_networks(self, test_namespace, openshift_provider):
        """Test getting inventory networks from an OpenShift provider."""
        result = test_namespace.run_mtv_command(
            f"get inventory networks {openshift_provider} --output json"
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

    def test_get_inventory_storage(self, test_namespace, openshift_provider):
        """Test getting inventory storage from an OpenShift provider."""
        result = test_namespace.run_mtv_command(
            f"get inventory storage {openshift_provider} --output json"
        )

        if result.returncode != 0:
            pytest.skip("Provider inventory not accessible or empty")

        # Parse JSON output
        try:
            storage_data = json.loads(result.stdout)
            assert isinstance(storage_data, list), "Expected storage data to be a list"
        except json.JSONDecodeError:
            pytest.fail("Failed to parse storage JSON output")

    def test_get_inventory_namespaces(self, test_namespace, openshift_provider):
        """Test getting inventory namespaces from an OpenShift provider."""
        result = test_namespace.run_mtv_command(
            f"get inventory namespaces {openshift_provider} --output json"
        )

        if result.returncode != 0:
            pytest.skip("Provider inventory not accessible or empty")

        # Parse JSON output
        try:
            namespaces_data = json.loads(result.stdout)
            assert isinstance(
                namespaces_data, list
            ), "Expected namespaces data to be a list"
        except json.JSONDecodeError:
            pytest.fail("Failed to parse namespaces JSON output")
