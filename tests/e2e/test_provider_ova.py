"""
Test cases for kubectl-mtv OVA provider creation.

This test validates the creation of OVA providers.
"""

import json
import time

import pytest


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
        self._verify_provider_created(test_namespace, provider_name, "ova")
    
    def test_create_ova_provider_missing_url(self, test_namespace):
        """Test creating an OVA provider with missing URL."""
        provider_name = "test-ova-no-url-provider"
        
        # This should fail because OVA requires URL
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type ova",
            check=False
        )
        
        assert result.returncode != 0
    
    def _verify_provider_created(self, test_namespace, provider_name: str, provider_type: str):
        """Verify that a provider was created successfully."""
        # Wait a moment for provider to be created
        time.sleep(2)
        
        # Check if provider exists
        result = test_namespace.run_mtv_command(f"get provider {provider_name} -o json")
        assert result.returncode == 0
        
        # Parse provider data
        provider_list = json.loads(result.stdout)
        assert len(provider_list) == 1, f"Expected 1 provider, got {len(provider_list)}"
        provider_data = provider_list[0]
        assert provider_data.get("metadata", {}).get("name") == provider_name
        assert provider_data.get("spec", {}).get("type") == provider_type
        
        # Check provider status (might take a moment to be ready)
        # This is optional as provider validation might take time
        status = provider_data.get("status") or {}
        print(f"Provider {provider_name} status: {status}")
        
        # Provider should at least be created without immediate errors
        conditions = status.get("conditions", [])
        if conditions:
            # Look for any error conditions
            error_conditions = [
                c for c in conditions 
                if c.get("type") == "Ready" and c.get("status") == "False"
            ]
            if error_conditions:
                print(f"Warning: Provider {provider_name} has error conditions: {error_conditions}")
