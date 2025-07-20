"""
Test cases for kubectl-mtv OVA provider creation.

This test validates the creation of OVA (Open Virtualization Archive) providers.
"""

import pytest

from ..utils import verify_provider_created


@pytest.mark.provider
@pytest.mark.ova
@pytest.mark.requires_credentials
class TestOVAProvider:
    """Test cases for OVA provider creation."""

    def test_create_ova_provider_skip_verify(
        self, test_namespace, provider_credentials
    ):
        """Test creating an OVA provider with TLS verification skipped."""
        creds = provider_credentials["ova"]

        # Skip if OVA URL is not available
        if not creds.get("url"):
            pytest.skip("OVA URL not available in environment")

        provider_name = "test-ova-skip-verify"

        # Create command with insecure skip TLS (for HTTPS URLs)
        cmd_parts = [
            "create provider",
            provider_name,
            "--type ova",
            f"--url '{creds['url']}'",
            "--provider-insecure-skip-tls",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "ova")

    def test_create_ova_provider_with_cacert(
        self, test_namespace, provider_credentials
    ):
        """Test creating an OVA provider with CA certificate."""
        creds = provider_credentials["ova"]

        # Skip if OVA URL or CA cert are not available
        if not all([creds.get("url"), creds.get("cacert")]):
            pytest.skip("OVA URL with CA certificate not available in environment")

        provider_name = "test-ova-cacert"

        # Create command with CA cert
        cmd_parts = [
            "create provider",
            provider_name,
            "--type ova",
            f"--url '{creds['url']}'",
            f"--cacert '{creds['cacert']}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "ova")

    def test_create_ova_provider_error(self, test_namespace):
        """Test creating an OVA provider with missing required fields."""
        provider_name = "test-ova-error"

        # This should fail because OVA requires URL
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type ova", check=False
        )

        assert result.returncode != 0
