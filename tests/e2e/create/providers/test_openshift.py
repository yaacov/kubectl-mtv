"""
Test cases for kubectl-mtv OpenShift provider creation.

This test validates the creation of OpenShift target providers.
"""

import pytest

from ...utils import (
    wait_for_provider_ready,
    generate_provider_name,
    provider_exists,
)


@pytest.mark.create
@pytest.mark.provider
@pytest.mark.providers
@pytest.mark.openshift
class TestOpenShiftProvider:
    """Test cases for OpenShift provider creation."""

    def test_create_openshift_provider_localhost(self, test_namespace):
        """Test creating a namespaced localhost OpenShift provider using current cluster context."""
        provider_name = generate_provider_name("openshift", "localhost", skip_tls=True)

        # Skip if provider already exists
        if provider_exists(test_namespace, provider_name):
            pytest.skip(f"Provider {provider_name} already exists")

        # Create a simple OpenShift provider without URL or token (uses current cluster)
        create_cmd = f"create provider {provider_name} --type openshift"

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Wait for provider to be ready
        wait_for_provider_ready(test_namespace, provider_name)

    def test_create_openshift_provider_skip_verify(
        self, test_namespace, provider_credentials
    ):
        """Test creating an OpenShift provider with external URL and TLS verification skipped."""
        creds = provider_credentials["openshift"]

        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("token")]):
            pytest.skip("OpenShift credentials not available in environment")

        provider_name = generate_provider_name("openshift", creds["url"], skip_tls=True)

        # Skip if provider already exists
        if provider_exists(test_namespace, provider_name):
            pytest.skip(f"Provider {provider_name} already exists")

        # Create command with insecure skip TLS
        cmd_parts = [
            "create provider",
            provider_name,
            "--type openshift",
            f"--url '{creds['url']}'",
            f"--token '{creds['token']}'",
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

    def test_create_openshift_provider_with_cacert(
        self, test_namespace, provider_credentials
    ):
        """Test creating an OpenShift provider with CA certificate."""
        creds = provider_credentials["openshift"]

        # Skip if credentials or CA cert are not available
        required_fields = ["url", "token", "cacert"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip(
                "OpenShift credentials with CA certificate not available in environment"
            )

        provider_name = generate_provider_name(
            "openshift", creds["url"], skip_tls=False
        )

        # Skip if provider already exists
        if provider_exists(test_namespace, provider_name):
            pytest.skip(f"Provider {provider_name} already exists")

        # Create command with CA cert
        cmd_parts = [
            "create provider",
            provider_name,
            "--type openshift",
            f"--url '{creds['url']}'",
            f"--token '{creds['token']}'",
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

    def test_create_openshift_provider_error(self, test_namespace):
        """Test creating an OpenShift provider with invalid configuration."""
        provider_name = "test-openshift-error"

        # This should fail because providing a token without URL is invalid
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type openshift --token 'invalid-token-without-url'",
            check=False,
        )

        assert result.returncode != 0
