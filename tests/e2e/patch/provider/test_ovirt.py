"""
Test cases for kubectl-mtv OVirt provider patching.

This test validates the patching of OVirt source providers.
"""

import pytest

from ...utils import (
    generate_provider_name,
    get_or_create_provider,
)


@pytest.mark.patch
@pytest.mark.provider
@pytest.mark.providers
@pytest.mark.ovirt
class TestOVirtProviderPatch:
    """Test cases for OVirt provider patching."""

    @pytest.fixture(scope="class")
    def ovirt_provider(self, test_namespace, provider_credentials):
        """Create an OVirt provider for patching tests."""
        creds = provider_credentials["ovirt"]

        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("OVirt credentials not available in environment")

        provider_name = generate_provider_name("ovirt", creds["url"], skip_tls=True)

        # Create OVirt provider
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

    def test_patch_ovirt_provider_password(
        self, test_namespace, ovirt_provider, provider_credentials
    ):
        """Test patching an OVirt provider to update password."""
        creds = provider_credentials["ovirt"]

        # Skip if credentials are not available
        if not creds.get("password"):
            pytest.skip("OVirt password not available in environment")

        # Patch provider with updated password
        patch_cmd = f"patch provider {ovirt_provider} --password '{creds['password']}'"

        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0

        # Verify the password field references a secret (credentials are stored in secrets, not directly in provider spec)
        get_result = test_namespace.run_mtv_command(
            f"get provider {ovirt_provider} -o yaml"
        )
        assert get_result.returncode == 0
        # Password is stored in a secret, so verify the secret reference exists
        assert "secret:" in get_result.stdout or "secretRef:" in get_result.stdout

    def test_patch_ovirt_provider_username(
        self, test_namespace, ovirt_provider, provider_credentials
    ):
        """Test patching an OVirt provider to update username."""
        creds = provider_credentials["ovirt"]

        # Skip if credentials are not available
        if not creds.get("username"):
            pytest.skip("OVirt username not available in environment")

        # Patch provider with updated username
        patch_cmd = f"patch provider {ovirt_provider} --username '{creds['username']}'"

        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0

        # Verify the username field references a secret (credentials are stored in secrets, not directly in provider spec)
        get_result = test_namespace.run_mtv_command(
            f"get provider {ovirt_provider} -o yaml"
        )
        assert get_result.returncode == 0
        # Username is stored in a secret, so verify the secret reference exists
        assert "secret:" in get_result.stdout or "secretRef:" in get_result.stdout

    def test_patch_ovirt_provider_insecure_skip_tls(
        self, test_namespace, ovirt_provider
    ):
        """Test patching an OVirt provider to enable insecure skip TLS."""
        # Patch provider to enable insecure skip TLS
        patch_cmd = f"patch provider {ovirt_provider} --provider-insecure-skip-tls"

        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0

        # Verify the insecure skip TLS setting was updated (boolean flags are stored directly in provider spec)
        get_result = test_namespace.run_mtv_command(
            f"get provider {ovirt_provider} -o yaml"
        )
        assert get_result.returncode == 0
        assert 'insecureSkipVerify: "true"' in get_result.stdout

    def test_patch_ovirt_provider_cacert(
        self, test_namespace, ovirt_provider, provider_credentials
    ):
        """Test patching an OVirt provider to add/update CA certificate."""
        creds = provider_credentials["ovirt"]

        # Skip if CA certificate is not available
        if not creds.get("cacert"):
            pytest.skip("OVirt CA certificate not available in environment")

        # Patch provider with CA certificate
        patch_cmd = f"patch provider {ovirt_provider} --cacert '{creds['cacert']}'"

        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0

        # Verify the CA certificate field references a secret (certs are stored in secrets, not directly in provider spec)
        get_result = test_namespace.run_mtv_command(
            f"get provider {ovirt_provider} -o yaml"
        )
        assert get_result.returncode == 0
        # CA cert is stored in a secret, so verify the secret reference exists
        assert "secret:" in get_result.stdout or "secretRef:" in get_result.stdout

    def test_patch_ovirt_provider_multiple_fields(
        self, test_namespace, ovirt_provider, provider_credentials
    ):
        """Test patching an OVirt provider with multiple fields at once."""
        creds = provider_credentials["ovirt"]

        # Skip if credentials are not available
        if not all([creds.get("username"), creds.get("password")]):
            pytest.skip("OVirt credentials not available in environment")

        # Patch provider with multiple fields
        cmd_parts = [
            "patch provider",
            ovirt_provider,
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            "--provider-insecure-skip-tls",
        ]

        patch_cmd = " ".join(cmd_parts)

        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0

        # Verify all fields were updated in the provider spec
        get_result = test_namespace.run_mtv_command(
            f"get provider {ovirt_provider} -o yaml"
        )
        assert get_result.returncode == 0
        # Credentials are stored in secrets, so verify secret references exist
        assert "secret:" in get_result.stdout or "secretRef:" in get_result.stdout
        # Boolean flags are stored directly in provider spec
        assert 'insecureSkipVerify: "true"' in get_result.stdout

    def test_patch_ovirt_provider_error(self, test_namespace):
        """Test patching a non-existent OVirt provider."""
        non_existent_provider = "non-existent-provider"

        # This should fail because the provider doesn't exist
        result = test_namespace.run_mtv_command(
            f"patch provider {non_existent_provider} --provider-insecure-skip-tls",
            check=False,
        )

        assert result.returncode != 0
