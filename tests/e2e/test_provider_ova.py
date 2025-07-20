"""
Test cases for kubectl-mtv OVA provider creation.

This test validates the creation of OVA providers.
"""

import json

import pytest

from utils import verify_provider_created


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
        
        provider_name = "test-ova-provider"
        
        # Build create command
        create_cmd = f"create provider {provider_name} --type ova --url '{creds['url']}'"
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "ova")
    
    def test_create_ova_provider_missing_url(self, test_namespace):
        """Test creating an OVA provider with missing URL."""
        provider_name = "test-ova-no-url-provider"
        
        # This should fail because OVA requires URL
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type ova",
            check=False
        )
        
        assert result.returncode != 0
