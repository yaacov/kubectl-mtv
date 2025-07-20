"""
Test cases for kubectl-mtv VMware vSphere provider creation.

This test validates the creation of VMware vSphere providers.
"""

import pytest

from ..utils import verify_provider_created


@pytest.mark.provider
@pytest.mark.vsphere
@pytest.mark.requires_credentials
class TestVSphereProvider:
    """Test cases for VMware vSphere provider creation."""
    
    def test_create_vsphere_provider_skip_verify(self, test_namespace, provider_credentials):
        """Test creating a vSphere provider with TLS verification skipped."""
        creds = provider_credentials["vsphere"]
        
        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("VMware vSphere credentials not available in environment")
        
        provider_name = "test-vsphere-skip-verify"
        
        # Create command with insecure skip TLS
        cmd_parts = [
            "create provider", provider_name,
            "--type vsphere",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            "--provider-insecure-skip-tls"
        ]
        
        create_cmd = " ".join(cmd_parts)
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "vsphere")

    def test_create_vsphere_provider_with_cacert(self, test_namespace, provider_credentials):
        """Test creating a vSphere provider with CA certificate."""
        creds = provider_credentials["vsphere"]
        
        # Skip if credentials or CA cert are not available
        required_fields = ["url", "username", "password", "cacert"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip("VMware vSphere credentials with CA certificate not available in environment")
        
        provider_name = "test-vsphere-cacert"
        
        # Create command with CA cert
        cmd_parts = [
            "create provider", provider_name,
            "--type vsphere",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--cacert '{creds['cacert']}'"
        ]
        
        create_cmd = " ".join(cmd_parts)
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "vsphere")

    def test_create_vsphere_provider_error(self, test_namespace):
        """Test creating a vSphere provider with missing required fields."""
        provider_name = "test-vsphere-error"
        
        # This should fail because vSphere requires URL, username, password
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type vsphere",
            check=False
        )
        
        assert result.returncode != 0
