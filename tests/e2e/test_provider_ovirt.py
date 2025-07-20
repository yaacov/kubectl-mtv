"""
Test cases for kubectl-mtv oVirt provider creation.

This test validates the creation of oVirt providers.
"""

import json

import pytest

from utils import verify_provider_created


@pytest.mark.provider
@pytest.mark.ovirt
@pytest.mark.requires_credentials
class TestOVirtProvider:
    """Test cases for oVirt provider creation."""
    
    def test_create_ovirt_provider(self, test_namespace, provider_credentials):
        """Test creating an oVirt provider."""
        creds = provider_credentials["ovirt"]
        
        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("oVirt credentials not available in environment")
        
        provider_name = "test-ovirt-provider"
        
        # Build create command
        cmd_parts = [
            "create provider", provider_name,
            "--type ovirt",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'"
        ]
        
        if creds.get("cacert"):
            cmd_parts.append(f"--cacert '{creds['cacert']}'")
        
        if creds.get("insecure"):
            cmd_parts.append("--provider-insecure-skip-tls")
        
        create_cmd = " ".join(cmd_parts)
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "ovirt")
    
    def test_create_ovirt_provider_with_insecure_tls(self, test_namespace, provider_credentials):
        """Test creating an oVirt provider with insecure TLS skip."""
        creds = provider_credentials["ovirt"]
        
        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("oVirt credentials not available in environment")
        
        provider_name = "test-ovirt-insecure-provider"
        
        # Build create command with insecure TLS
        create_cmd = (
            f"create provider {provider_name} --type ovirt "
            f"--url '{creds['url']}' "
            f"--username '{creds['username']}' "
            f"--password '{creds['password']}' "
            "--provider-insecure-skip-tls"
        )
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "ovirt")
    
    def test_create_ovirt_provider_with_ca_cert(self, test_namespace, provider_credentials, temp_file):
        """Test creating an oVirt provider with CA certificate."""
        creds = provider_credentials["ovirt"]
        
        # Skip if credentials or CA cert are not available
        required_fields = ["url", "username", "password", "cacert"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip("oVirt credentials with CA certificate not available in environment")
        
        provider_name = "test-ovirt-cacert-provider"
        
        # Write CA cert to temp file
        with open(temp_file, 'w') as f:
            f.write(creds['cacert'])
        
        # Build create command with CA cert file
        create_cmd = (
            f"create provider {provider_name} --type ovirt "
            f"--url '{creds['url']}' "
            f"--username '{creds['username']}' "
            f"--password '{creds['password']}' "
            f"--cacert @{temp_file}"
        )
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "ovirt")
