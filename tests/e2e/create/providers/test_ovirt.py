"""
Test cases for kubectl-mtv oVirt provider creation.

This test validates the creation of oVirt/Red Hat Virtualization providers.
"""

import pytest

from ...utils import wait_for_provider_ready, generate_provider_name, provider_exists


@pytest.mark.create
@pytest.mark.provider
@pytest.mark.providers
@pytest.mark.ovirt
@pytest.mark.requires_credentials
class TestOVirtProvider:
    """Test cases for oVirt provider creation."""

    def test_create_ovirt_provider_skip_verify(
        self, test_namespace, provider_credentials
    ):
        """Test creating an oVirt provider with TLS verification skipped."""
        creds = provider_credentials["ovirt"]

        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("oVirt credentials not available in environment")

        provider_name = generate_provider_name("ovirt", creds["url"], skip_tls=True)

        # Skip if provider already exists
        if provider_exists(test_namespace, provider_name):
            pytest.skip(f"Provider {provider_name} already exists")

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

    def test_create_ovirt_provider_with_cacert(
        self, test_namespace, provider_credentials
    ):
        """Test creating an oVirt provider with CA certificate."""
        creds = provider_credentials["ovirt"]

        # Skip if credentials or CA cert are not available
        required_fields = ["url", "username", "password", "cacert"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip(
                "oVirt credentials with CA certificate not available in environment"
            )

        provider_name = generate_provider_name("ovirt", creds["url"], skip_tls=False)

        # Skip if provider already exists
        if provider_exists(test_namespace, provider_name):
            pytest.skip(f"Provider {provider_name} already exists")

        # Create command with CA cert
        cmd_parts = [
            "create provider",
            provider_name,
            "--type ovirt",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--cacert '{creds['cacert']}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Wait for provider to be ready
        wait_for_provider_ready(test_namespace, provider_name)

    def test_create_ovirt_provider_error(self, test_namespace):
        """Test creating an oVirt provider with missing required fields."""
        provider_name = "test-ovirt-error"

        # This should fail because oVirt requires URL, username, password
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type ovirt", check=False
        )

        assert result.returncode != 0
