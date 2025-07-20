"""
Test cases for kubectl-mtv OVA provider creation.

This test validates the creation of OVA (Open Virtualization Archive) providers.
"""

import json

import pytest

from ..utils import verify_provider_created, generate_unique_resource_name


@pytest.mark.provider
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
        
        # Generate unique provider name (important for shared namespace)
        provider_name = generate_unique_resource_name("test-ova-provider")
        
        # Build create command
        cmd_parts = [
            "create provider", provider_name,
            "--type ova",
            f"--url '{creds['url']}'"
        ]
                
        create_cmd = " ".join(cmd_parts)
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "ova")
         
    def test_create_ova_provider_missing_url(self, test_namespace):
        """Test creating an OVA provider with missing required URL."""
        provider_name = generate_unique_resource_name("test-incomplete-ova-provider")
        
        # This should fail because OVA requires URL
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type ova",
            check=False
        )
        
        assert result.returncode != 0
