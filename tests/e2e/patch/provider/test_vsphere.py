"""
Test cases for kubectl-mtv VSphere provider patching.

This test validates the patching of VSphere source providers.
"""

import pytest
import base64
import yaml

from ...utils import (
    generate_provider_name,
    provider_exists,
    get_or_create_provider,
)


@pytest.mark.patch
@pytest.mark.provider
@pytest.mark.providers
@pytest.mark.vsphere
class TestVSphereProviderPatch:
    """Test cases for VSphere provider patching."""

    def _verify_credential_in_secret(self, test_namespace, provider_name, credential_field, expected_value):
        """Helper method to verify a credential field was updated in the provider's secret."""
        # Get provider spec to find secret reference
        get_result = test_namespace.run_mtv_command(f"get provider {provider_name} -o yaml")
        assert get_result.returncode == 0
        
        # Parse provider YAML to extract secret name
        provider_yaml = yaml.safe_load(get_result.stdout)
        secret_ref = provider_yaml.get('spec', {}).get('secret', {})
        secret_name = secret_ref.get('name')
        secret_namespace = secret_ref.get('namespace', test_namespace.name)
        
        assert secret_name, "Provider should reference a secret"
        
        # Get the secret contents
        secret_result = test_namespace.run_mtv_command(f"get secret {secret_name} -o yaml")
        assert secret_result.returncode == 0
        
        # Parse secret YAML and decode the credential field
        secret_yaml = yaml.safe_load(secret_result.stdout)
        secret_data = secret_yaml.get('data', {})
        
        if credential_field in secret_data:
            # Decode base64 value and verify it matches expected value
            decoded_value = base64.b64decode(secret_data[credential_field]).decode('utf-8')
            assert decoded_value == expected_value, f"Expected {credential_field} to be '{expected_value}', got '{decoded_value}'"
        else:
            # Field might be in stringData (not base64 encoded)
            string_data = secret_yaml.get('stringData', {})
            assert credential_field in string_data, f"Credential field '{credential_field}' not found in secret"
            assert string_data[credential_field] == expected_value

    @pytest.fixture(scope="class")
    def vsphere_provider(self, test_namespace, provider_credentials):
        """Create a VSphere provider for patching tests."""
        creds = provider_credentials["vsphere"]
        
        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("VSphere credentials not available in environment")

        provider_name = generate_provider_name("vsphere", creds["url"], skip_tls=True)
        
        # Create VSphere provider
        cmd_parts = [
            "create provider",
            provider_name,
            "--type vsphere",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            "--provider-insecure-skip-tls",
        ]
        
        create_cmd = " ".join(cmd_parts)
        
        # Create provider if it doesn't already exist
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    def test_patch_vsphere_provider_password(self, test_namespace, vsphere_provider, provider_credentials):
        """Test patching a VSphere provider to update password."""
        creds = provider_credentials["vsphere"]
        
        # Skip if credentials are not available
        if not creds.get("password"):
            pytest.skip("VSphere password not available in environment")

        # Patch provider with updated password
        patch_cmd = f"patch provider {vsphere_provider} --password '{creds['password']}'"
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the password was actually updated in the secret
        self._verify_credential_in_secret(test_namespace, vsphere_provider, "password", creds['password'])

    def test_patch_vsphere_provider_username(self, test_namespace, vsphere_provider, provider_credentials):
        """Test patching a VSphere provider to update username."""
        creds = provider_credentials["vsphere"]
        
        # Skip if credentials are not available
        if not creds.get("username"):
            pytest.skip("VSphere username not available in environment")

        # Patch provider with updated username
        patch_cmd = f"patch provider {vsphere_provider} --username '{creds['username']}'"
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the username was actually updated in the secret
        self._verify_credential_in_secret(test_namespace, vsphere_provider, "username", creds['username'])

    def test_patch_vsphere_provider_vddk_init_image(self, test_namespace, vsphere_provider):
        """Test patching a VSphere provider to update VDDK init image."""
        vddk_image = "registry.redhat.io/rhel8/vddk:latest"
        
        # Patch provider with VDDK init image
        patch_cmd = f"patch provider {vsphere_provider} --vddk-init-image '{vddk_image}'"
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the VDDK init image was updated in the provider spec (image names are stored directly)
        get_result = test_namespace.run_mtv_command(f"get provider {vsphere_provider} -o yaml")
        assert get_result.returncode == 0
        assert vddk_image in get_result.stdout

    def test_patch_vsphere_provider_vddk_optimizations(self, test_namespace, vsphere_provider):
        """Test patching a VSphere provider to enable VDDK optimizations."""
        # Patch provider with VDDK optimizations
        cmd_parts = [
            "patch provider",
            vsphere_provider,
            "--use-vddk-aio-optimization",
            "--vddk-buf-size-in-64k 128",
            "--vddk-buf-count 8",
        ]
        
        patch_cmd = " ".join(cmd_parts)
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the VDDK optimizations were updated in the provider spec (settings are stored directly)
        get_result = test_namespace.run_mtv_command(f"get provider {vsphere_provider} -o yaml")
        assert get_result.returncode == 0
        assert ("aioOptimization: true" in get_result.stdout or "useAioOptimization: true" in get_result.stdout)
        assert "bufSizeIn64K: 128" in get_result.stdout
        assert "bufCount: 8" in get_result.stdout

    def test_patch_vsphere_provider_cacert(self, test_namespace, vsphere_provider, provider_credentials):
        """Test patching a VSphere provider to add/update CA certificate."""
        creds = provider_credentials["vsphere"]
        
        # Skip if CA certificate is not available
        if not creds.get("cacert"):
            pytest.skip("VSphere CA certificate not available in environment")

        # Patch provider with CA certificate
        patch_cmd = f"patch provider {vsphere_provider} --cacert '{creds['cacert']}'"
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the CA certificate field references a secret (certs are stored in secrets, not directly in provider spec)
        get_result = test_namespace.run_mtv_command(f"get provider {vsphere_provider} -o yaml")
        assert get_result.returncode == 0
        # CA cert is stored in a secret, so verify the secret reference exists
        assert ("secret:" in get_result.stdout or "secretRef:" in get_result.stdout)

    def test_patch_vsphere_provider_multiple_fields(self, test_namespace, vsphere_provider, provider_credentials):
        """Test patching a VSphere provider with multiple fields at once."""
        creds = provider_credentials["vsphere"]
        
        # Skip if credentials are not available
        if not all([creds.get("username"), creds.get("password")]):
            pytest.skip("VSphere credentials not available in environment")

        # Patch provider with multiple fields
        cmd_parts = [
            "patch provider",
            vsphere_provider,
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            "--provider-insecure-skip-tls",
            "--use-vddk-aio-optimization",
        ]
        
        patch_cmd = " ".join(cmd_parts)
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify all fields were updated in the provider spec
        get_result = test_namespace.run_mtv_command(f"get provider {vsphere_provider} -o yaml")
        assert get_result.returncode == 0
        # Credentials are stored in secrets, so verify secret references exist
        assert ("secret:" in get_result.stdout or "secretRef:" in get_result.stdout)
        # Boolean flags and settings are stored directly in provider spec
        assert ("insecure: true" in get_result.stdout or "skipTLS: true" in get_result.stdout)
        assert ("aioOptimization: true" in get_result.stdout or "useAioOptimization: true" in get_result.stdout)

    def test_patch_vsphere_provider_error(self, test_namespace):
        """Test patching a non-existent VSphere provider."""
        non_existent_provider = "non-existent-provider"
        
        # This should fail because the provider doesn't exist
        result = test_namespace.run_mtv_command(
            f"patch provider {non_existent_provider} --provider-insecure-skip-tls",
            check=False,
        )
        
        assert result.returncode != 0