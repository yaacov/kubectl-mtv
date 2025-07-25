"""
Test cases for kubectl-mtv OpenStack provider creation.

This test validates the creation of OpenStack providers.
"""

import pytest

from ...utils import wait_for_provider_ready


@pytest.mark.create
@pytest.mark.provider
@pytest.mark.providers
@pytest.mark.openstack
@pytest.mark.requires_credentials
class TestOpenStackProvider:
    """Test cases for OpenStack provider creation."""

    def test_create_openstack_provider_skip_verify(
        self, test_namespace, provider_credentials
    ):
        """Test creating an OpenStack provider with TLS verification skipped."""
        creds = provider_credentials["openstack"]

        # Skip if credentials are not available
        required_fields = ["url", "username", "password", "domain_name", "project_name"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip("OpenStack credentials not available in environment")

        provider_name = "test-openstack-skip-verify"

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
            "--provider-insecure-skip-tls",
        ]

        if creds.get("region_name"):
            cmd_parts.append(f"--provider-region-name '{creds['region_name']}'")

        create_cmd = " ".join(cmd_parts)

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Wait for provider to be ready
        wait_for_provider_ready(test_namespace, provider_name)

    def test_create_openstack_provider_with_cacert(
        self, test_namespace, provider_credentials
    ):
        """Test creating an OpenStack provider with CA certificate."""
        creds = provider_credentials["openstack"]

        # Skip if credentials or CA cert are not available
        required_fields = [
            "url",
            "username",
            "password",
            "domain_name",
            "project_name",
            "cacert",
        ]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip(
                "OpenStack credentials with CA certificate not available in environment"
            )

        provider_name = "test-openstack-cacert"

        # Create command with CA cert
        cmd_parts = [
            "create provider",
            provider_name,
            "--type openstack",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--provider-domain-name '{creds['domain_name']}'",
            f"--provider-project-name '{creds['project_name']}'",
            f"--cacert '{creds['cacert']}'",
        ]

        if creds.get("region_name"):
            cmd_parts.append(f"--provider-region-name '{creds['region_name']}'")

        create_cmd = " ".join(cmd_parts)

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Wait for provider to be ready
        wait_for_provider_ready(test_namespace, provider_name)

    def test_create_openstack_provider_error(self, test_namespace):
        """Test creating an OpenStack provider with missing required fields."""
        provider_name = "test-openstack-error"

        # This should fail because OpenStack requires URL, username, password, domain, project
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type openstack", check=False
        )

        assert result.returncode != 0
