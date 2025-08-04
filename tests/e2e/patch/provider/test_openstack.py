"""
Test cases for kubectl-mtv OpenStack provider patching.

This test validates the patching of OpenStack source providers.
"""

import pytest
import re
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
@pytest.mark.openstack
class TestOpenStackProviderPatch:
    """Test cases for OpenStack provider patching."""

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
    def openstack_provider(self, test_namespace, provider_credentials):
        """Create an OpenStack provider for patching tests."""
        creds = provider_credentials["openstack"]
        
        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("OpenStack credentials not available in environment")

        provider_name = generate_provider_name("openstack", creds["url"], skip_tls=True)
        
        # Create OpenStack provider
        cmd_parts = [
            "create provider",
            provider_name,
            "--type openstack",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            "--provider-insecure-skip-tls",
        ]
        
        create_cmd = " ".join(cmd_parts)
        
        # Create provider if it doesn't already exist
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    def test_patch_openstack_provider_password(self, test_namespace, openstack_provider, provider_credentials):
        """Test patching an OpenStack provider to update password."""
        creds = provider_credentials["openstack"]
        
        # Skip if credentials are not available
        if not creds.get("password"):
            pytest.skip("OpenStack password not available in environment")

        # Patch provider with updated password
        patch_cmd = f"patch provider {openstack_provider} --password '{creds['password']}'"
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the password was actually updated in the secret
        self._verify_credential_in_secret(test_namespace, openstack_provider, "password", creds['password'])

    def test_patch_openstack_provider_username(self, test_namespace, openstack_provider, provider_credentials):
        """Test patching an OpenStack provider to update username."""
        creds = provider_credentials["openstack"]
        
        # Skip if credentials are not available
        if not creds.get("username"):
            pytest.skip("OpenStack username not available in environment")

        # Patch provider with updated username
        patch_cmd = f"patch provider {openstack_provider} --username '{creds['username']}'"
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the username was actually updated in the secret
        self._verify_credential_in_secret(test_namespace, openstack_provider, "username", creds['username'])

    def test_patch_openstack_provider_domain_name(self, test_namespace, openstack_provider, provider_credentials):
        """Test patching an OpenStack provider to update domain name."""
        creds = provider_credentials["openstack"]
        
        # Skip if domain name is not available
        if not creds.get("domain_name"):
            pytest.skip("OpenStack domain name not available in environment")

        # Patch provider with domain name
        patch_cmd = f"patch provider {openstack_provider} --domain-name '{creds['domain_name']}'"
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the domain name was actually updated in the secret
        self._verify_credential_in_secret(test_namespace, openstack_provider, "domainName", creds['domain_name'])

    def test_patch_openstack_provider_project_name(self, test_namespace, openstack_provider, provider_credentials):
        """Test patching an OpenStack provider to update project name."""
        creds = provider_credentials["openstack"]
        
        # Skip if project name is not available
        if not creds.get("project_name"):
            pytest.skip("OpenStack project name not available in environment")

        # Patch provider with project name
        patch_cmd = f"patch provider {openstack_provider} --project-name '{creds['project_name']}'"
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the project name was actually updated in the secret
        self._verify_credential_in_secret(test_namespace, openstack_provider, "projectName", creds['project_name'])

    def test_patch_openstack_provider_region_name(self, test_namespace, openstack_provider, provider_credentials):
        """Test patching an OpenStack provider to update region name."""
        creds = provider_credentials["openstack"]
        
        # Skip if region name is not available
        if not creds.get("region_name"):
            pytest.skip("OpenStack region name not available in environment")

        # Patch provider with region name
        patch_cmd = f"patch provider {openstack_provider} --region-name '{creds['region_name']}'"
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify the region name was actually updated in the secret
        self._verify_credential_in_secret(test_namespace, openstack_provider, "regionName", creds['region_name'])

    def test_patch_openstack_provider_multiple_fields(self, test_namespace, openstack_provider, provider_credentials):
        """Test patching an OpenStack provider with multiple fields at once."""
        creds = provider_credentials["openstack"]
        
        # Skip if credentials are not available
        if not all([creds.get("username"), creds.get("password")]):
            pytest.skip("OpenStack credentials not available in environment")

        # Patch provider with multiple fields
        cmd_parts = [
            "patch provider",
            openstack_provider,
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            "--provider-insecure-skip-tls",
        ]
        
        # Add optional fields if available
        if creds.get("domain_name"):
            cmd_parts.append(f"--domain-name '{creds['domain_name']}'")
        if creds.get("project_name"):
            cmd_parts.append(f"--project-name '{creds['project_name']}'")
        
        patch_cmd = " ".join(cmd_parts)
        
        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0
        
        # Verify all credential fields were updated in the secret
        self._verify_credential_in_secret(test_namespace, openstack_provider, "username", creds['username'])
        self._verify_credential_in_secret(test_namespace, openstack_provider, "password", creds['password'])
        
        # Verify optional fields if they were provided
        if creds.get("domain_name"):
            self._verify_credential_in_secret(test_namespace, openstack_provider, "domainName", creds['domain_name'])
        if creds.get("project_name"):
            self._verify_credential_in_secret(test_namespace, openstack_provider, "projectName", creds['project_name'])
        
        # Verify boolean flags are stored directly in provider spec
        get_result = test_namespace.run_mtv_command(f"get provider {openstack_provider} -o yaml")
        assert get_result.returncode == 0
        assert ("insecure: true" in get_result.stdout or "skipTLS: true" in get_result.stdout)

    def test_patch_openstack_provider_error(self, test_namespace):
        """Test patching a non-existent OpenStack provider."""
        non_existent_provider = "non-existent-provider"
        
        # This should fail because the provider doesn't exist
        result = test_namespace.run_mtv_command(
            f"patch provider {non_existent_provider} --provider-insecure-skip-tls",
            check=False,
        )
        
        assert result.returncode != 0