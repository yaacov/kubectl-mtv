"""
Test cases for kubectl-mtv OVA provider creation.

This test validates the creation of OVA (Open Virtualization Archive) providers.
"""

import pytest

from ...utils import wait_for_provider_ready


@pytest.mark.create
@pytest.mark.provider
@pytest.mark.providers
@pytest.mark.ova
@pytest.mark.requires_credentials
class TestOVAProvider:
    """Test cases for OVA provider creation."""

    def test_create_ova_provider(self, test_namespace, provider_credentials):
        """Test creating an OVA provider."""
        creds = provider_credentials["ova"]

        # Skip if OVA URL is not available
        if not creds.get("url"):
            pytest.skip("OVA URL not available in environment")

        provider_name = "test-ova-skip-verify"

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

    def test_create_ova_provider_error(self, test_namespace):
        """Test creating an OVA provider with missing required fields."""
        provider_name = "test-ova-error"

        # This should fail because OVA requires URL
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type ova", check=False
        )

        assert result.returncode != 0
