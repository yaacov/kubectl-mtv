"""
Test cases for kubectl-mtv OpenShift provider creation.

This test validates the creation of OpenShift target providers.
"""

import json

import pytest

from utils import verify_provider_created, generate_unique_resource_name


@pytest.mark.provider
@pytest.mark.openshift
class TestOpenShiftProvider:
    """Test cases for OpenShift provider creation."""
    
    def test_create_openshift_provider(self, test_namespace, provider_credentials):
        """Test creating an OpenShift target provider."""
        creds = provider_credentials["openshift"]
        provider_name = generate_unique_resource_name("test-openshift-provider")
        
        # For OpenShift provider, we can often use the current cluster
        if creds.get("url") and creds.get("token"):
            # Use explicit credentials
            create_cmd = (
                f"create provider {provider_name} --type openshift "
                f"--url '{creds['url']}' --token '{creds['token']}'"
            )
        else:
            # Use current cluster context (most common case)
            create_cmd = f"create provider {provider_name} --type openshift"
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "openshift")
    
    def test_list_providers_after_creation(self, test_namespace):
        """Test listing providers after creating one."""
        provider_name = generate_unique_resource_name("test-list-provider")
        
        # Create a simple OpenShift provider
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type openshift"
        )
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # List providers
        result = test_namespace.run_mtv_command("get providers")
        assert result.returncode == 0
        assert provider_name in result.stdout
    
    def test_get_provider_details(self, test_namespace):
        """Test getting details of a created provider."""
        provider_name = generate_unique_resource_name("test-details-provider")
        
        # Create a simple OpenShift provider
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type openshift"
        )
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Get provider details
        result = test_namespace.run_mtv_command(f"get provider {provider_name}")
        assert result.returncode == 0
        assert provider_name in result.stdout
        
        # Test JSON output
        result = test_namespace.run_mtv_command(f"get provider {provider_name} -o json")
        assert result.returncode == 0
        
        # Should be valid JSON
        provider_list = json.loads(result.stdout)
        assert len(provider_list) == 1, f"Expected 1 provider, got {len(provider_list)}"
        provider_data = provider_list[0]
        assert isinstance(provider_data, dict)
        assert provider_data.get("metadata", {}).get("name") == provider_name
    
    def test_delete_provider(self, test_namespace):
        """Test deleting a provider."""
        provider_name = generate_unique_resource_name("test-delete-provider")
        
        # Create a simple OpenShift provider
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type openshift"
        )
        assert result.returncode == 0
        
        # Verify it exists
        result = test_namespace.run_mtv_command(f"get provider {provider_name}")
        assert result.returncode == 0
        
        # Delete the provider
        result = test_namespace.run_mtv_command(f"delete provider {provider_name}")
        assert result.returncode == 0
        
        # Verify it's gone
        result = test_namespace.run_mtv_command(
            f"get provider {provider_name}",
            check=False
        )
        assert result.returncode != 0
