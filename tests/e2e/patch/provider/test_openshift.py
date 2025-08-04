"""
Test cases for kubectl-mtv OpenShift provider patching.

This test validates the patching of OpenShift target providers.
"""

import pytest

from ...utils import (
    generate_provider_name,
    provider_exists,
    get_or_create_provider,
)


@pytest.mark.patch
@pytest.mark.provider
@pytest.mark.providers
@pytest.mark.openshift
class TestOpenShiftProviderPatch:
    """Test cases for OpenShift provider patching."""

    @pytest.fixture(scope="class")
    def openshift_provider(self, test_namespace):
        """Create an OpenShift provider for patching tests."""
        provider_name = generate_provider_name("openshift", "localhost", skip_tls=True)
        
        # Create a simple OpenShift provider without URL or token (uses current cluster)
        create_cmd = f"create provider {provider_name} --type openshift"
        
        # Create provider if it doesn't already exist
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    def test_patch_openshift_provider_url(self, test_namespace, openshift_provider, provider_credentials):
        """Test patching an OpenShift provider to add/update URL."""
        creds = provider_credentials["openshift"]
        
        # Skip if credentials are not available
        if not creds.get("url"):
            pytest.skip("OpenShift URL not available in environment")

        # Patch provider with URL
        patch_cmd = f"patch provider {openshift_provider} --url '{creds['url']}'"
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the URL was updated in the provider spec (URLs are stored directly, not in secrets)
        get_result = test_namespace.run_mtv_command(f"get provider {openshift_provider} -o yaml")
        assert get_result.returncode == 0
        assert creds['url'] in get_result.stdout

    def test_patch_openshift_provider_token(self, test_namespace, openshift_provider, provider_credentials):
        """Test patching an OpenShift provider to add/update token."""
        creds = provider_credentials["openshift"]
        
        # Skip if credentials are not available
        if not creds.get("token"):
            pytest.skip("OpenShift token not available in environment")

        # Patch provider with token
        patch_cmd = f"patch provider {openshift_provider} --token '{creds['token']}'"
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the token field references a secret (credentials are stored in secrets, not directly in provider spec)
        get_result = test_namespace.run_mtv_command(f"get provider {openshift_provider} -o yaml")
        assert get_result.returncode == 0
        # Token is stored in a secret, so we verify the secret reference exists
        assert ("secret:" in get_result.stdout or "secretRef:" in get_result.stdout)

    def test_patch_openshift_provider_insecure_skip_tls(self, test_namespace, openshift_provider):
        """Test patching an OpenShift provider to enable insecure skip TLS."""
        # Patch provider to enable insecure skip TLS
        patch_cmd = f"patch provider {openshift_provider} --provider-insecure-skip-tls"
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the insecure skip TLS setting was updated
        get_result = test_namespace.run_mtv_command(f"get provider {openshift_provider} -o yaml")
        assert get_result.returncode == 0
        assert "insecure: true" in get_result.stdout or "skipTLS: true" in get_result.stdout

    def test_patch_openshift_provider_cacert(self, test_namespace, openshift_provider, provider_credentials):
        """Test patching an OpenShift provider to add/update CA certificate."""
        creds = provider_credentials["openshift"]
        
        # Skip if CA certificate is not available
        if not creds.get("cacert"):
            pytest.skip("OpenShift CA certificate not available in environment")

        # Patch provider with CA certificate
        patch_cmd = f"patch provider {openshift_provider} --cacert '{creds['cacert']}'"
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the CA certificate field references a secret (certs are stored in secrets, not directly in provider spec)
        get_result = test_namespace.run_mtv_command(f"get provider {openshift_provider} -o yaml")
        assert get_result.returncode == 0
        # CA cert is stored in a secret, so verify the secret reference exists
        assert ("secret:" in get_result.stdout or "secretRef:" in get_result.stdout)

    def test_patch_openshift_provider_multiple_fields(self, test_namespace, openshift_provider, provider_credentials):
        """Test patching an OpenShift provider with multiple fields at once."""
        creds = provider_credentials["openshift"]
        
        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("token")]):
            pytest.skip("OpenShift credentials not available in environment")

        # Patch provider with multiple fields
        cmd_parts = [
            "patch provider",
            openshift_provider,
            f"--url '{creds['url']}'",
            f"--token '{creds['token']}'",
            "--provider-insecure-skip-tls",
        ]
        
        patch_cmd = " ".join(cmd_parts)
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify all fields were updated in the provider spec
        get_result = test_namespace.run_mtv_command(f"get provider {openshift_provider} -o yaml")
        assert get_result.returncode == 0
        # URL is stored directly in provider spec
        assert creds['url'] in get_result.stdout
        # Credentials are stored in secrets, so verify secret references exist
        assert ("secret:" in get_result.stdout or "secretRef:" in get_result.stdout)
        # Boolean flags are stored directly in provider spec
        assert ("insecure: true" in get_result.stdout or "skipTLS: true" in get_result.stdout)

    def test_patch_openshift_provider_error(self, test_namespace):
        """Test patching a non-existent OpenShift provider."""
        non_existent_provider = "non-existent-provider"
        
        # This should fail because the provider doesn't exist
        result = test_namespace.run_mtv_command(
            f"patch provider {non_existent_provider} --provider-insecure-skip-tls",
            check=False,
        )
        
        assert result.returncode != 0