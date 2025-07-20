"""
Test cases for kubectl-mtv provider error conditions and edge cases.

This test validates error handling and edge cases for provider creation.
"""

import pytest


@pytest.mark.provider
@pytest.mark.error_cases
class TestProviderErrorCases:
    """Test cases for provider error conditions and validation."""
    
    def test_create_provider_with_invalid_type(self, test_namespace):
        """Test creating a provider with invalid type."""
        provider_name = "test-invalid-provider"
        
        # This should fail
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type invalid-type",
            check=False
        )
        
        assert result.returncode != 0
        assert "invalid provider type" in result.stderr.lower() or "invalid" in result.stderr.lower()
    
    def test_create_provider_with_empty_name(self, test_namespace):
        """Test creating a provider with empty name."""
        # This should fail
        result = test_namespace.run_mtv_command(
            "create provider --type openshift",
            check=False
        )
        
        assert result.returncode != 0
    
    def test_create_provider_with_duplicate_name(self, test_namespace):
        """Test creating a provider with duplicate name."""
        provider_name = "test-duplicate-provider"
        
        # Create first provider
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type openshift"
        )
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Try to create second provider with same name - should fail
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type openshift",
            check=False
        )
        
        assert result.returncode != 0
    
    def test_create_vsphere_provider_missing_credentials(self, test_namespace):
        """Test creating a vSphere provider with missing required fields."""
        provider_name = "test-incomplete-vsphere-provider"
        
        # This should fail because vsphere requires URL, username, password
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type vsphere",
            check=False
        )
        
        assert result.returncode != 0
    
    def test_create_vsphere_provider_missing_url(self, test_namespace):
        """Test creating a vSphere provider with missing URL."""
        provider_name = "test-vsphere-no-url"
        
        # This should fail because URL is required
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type vsphere --username admin --password secret",
            check=False
        )
        
        assert result.returncode != 0
    
    def test_create_openstack_provider_missing_credentials(self, test_namespace):
        """Test creating an OpenStack provider with missing required fields."""
        provider_name = "test-incomplete-openstack-provider"
        
        # This should fail because OpenStack requires multiple fields
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type openstack --url https://example.com",
            check=False
        )
        
        assert result.returncode != 0
    
    def test_create_ova_provider_missing_url(self, test_namespace):
        """Test creating an OVA provider with missing URL."""
        provider_name = "test-ova-no-url"
        
        # This should fail because OVA requires URL
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type ova",
            check=False
        )
        
        assert result.returncode != 0
    
    def test_get_nonexistent_provider(self, test_namespace):
        """Test getting details of a non-existent provider."""
        # This should fail
        result = test_namespace.run_mtv_command(
            "get provider nonexistent-provider",
            check=False
        )
        
        assert result.returncode != 0
    
    def test_delete_nonexistent_provider(self, test_namespace):
        """Test deleting a non-existent provider."""
        # This should fail
        result = test_namespace.run_mtv_command(
            "delete provider nonexistent-provider",
            check=False
        )
        
        assert result.returncode != 0
