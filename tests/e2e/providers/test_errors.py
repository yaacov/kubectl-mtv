"""
Test cases for kubectl-mtv provider error conditions and edge cases.

This test validates error handling and validation for provider operations.
"""

import pytest

from ..utils import generate_unique_resource_name


@pytest.mark.provider
@pytest.mark.error_cases
class TestProviderErrors:
    """Test cases for provider error conditions."""
    
    def test_create_provider_invalid_type(self, test_namespace):
        """Test creating a provider with invalid type."""
        provider_name = generate_unique_resource_name("test-invalid-type-provider")
        
        # This should fail because "invalid" is not a valid provider type
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type invalid",
            check=False
        )
        
        assert result.returncode != 0
        assert "invalid" in result.stderr.lower() or "unknown" in result.stderr.lower()
    
    def test_create_provider_missing_type(self, test_namespace):
        """Test creating a provider without specifying type."""
        provider_name = generate_unique_resource_name("test-missing-type-provider")
        
        # This should fail because --type is required
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name}",
            check=False
        )
        
        assert result.returncode != 0
    
    def test_create_provider_duplicate_name(self, test_namespace):
        """Test creating a provider with duplicate name."""
        provider_name = generate_unique_resource_name("test-duplicate-provider")
        
        # Create first provider (should succeed)
        result1 = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type openshift"
        )
        assert result1.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Try to create second provider with same name (should fail)
        result2 = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type openshift",
            check=False
        )
        
        assert result2.returncode != 0
        assert "already exists" in result2.stderr.lower() or "conflict" in result2.stderr.lower()
    
    def test_get_nonexistent_provider(self, test_namespace):
        """Test getting a provider that doesn't exist."""
        nonexistent_name = "nonexistent-provider-12345"
        
        result = test_namespace.run_mtv_command(
            f"get provider {nonexistent_name}",
            check=False
        )
        
        assert result.returncode != 0
        assert "not found" in result.stderr.lower() or "notfound" in result.stderr.lower()
    
    def test_delete_nonexistent_provider(self, test_namespace):
        """Test deleting a provider that doesn't exist."""
        nonexistent_name = "nonexistent-provider-12345"
        
        result = test_namespace.run_mtv_command(
            f"delete provider {nonexistent_name}",
            check=False
        )
        
        assert result.returncode != 0
        assert "not found" in result.stderr.lower() or "notfound" in result.stderr.lower()
    
    def test_create_provider_invalid_url_format(self, test_namespace):
        """Test creating providers with invalid URL formats."""
        provider_name = generate_unique_resource_name("test-invalid-url-provider")
        
        invalid_urls = [
            "not-a-url",
            "ftp://invalid-protocol.com",
            "http://",
            "https://",
            "://missing-scheme.com"
        ]
        
        for invalid_url in invalid_urls:
            result = test_namespace.run_mtv_command(
                f"create provider {provider_name}-{hash(invalid_url) % 1000} --type vsphere "
                f"--url '{invalid_url}' --username test --password test",
                check=False
            )
            
            # Should fail with invalid URL
            assert result.returncode != 0
    
    def test_create_provider_missing_required_fields(self, test_namespace):
        """Test creating providers with missing required fields."""
        base_name = generate_unique_resource_name("test-missing-fields")
        
        # vSphere provider missing username
        result = test_namespace.run_mtv_command(
            f"create provider {base_name}-1 --type vsphere "
            f"--url 'https://vcenter.example.com' --password test",
            check=False
        )
        assert result.returncode != 0
        
        # vSphere provider missing password
        result = test_namespace.run_mtv_command(
            f"create provider {base_name}-2 --type vsphere "
            f"--url 'https://vcenter.example.com' --username test",
            check=False
        )
        assert result.returncode != 0
        
        # vSphere provider missing URL
        result = test_namespace.run_mtv_command(
            f"create provider {base_name}-3 --type vsphere "
            f"--username test --password test",
            check=False
        )
        assert result.returncode != 0
    
    def test_create_provider_empty_values(self, test_namespace):
        """Test creating providers with empty values."""
        provider_name = generate_unique_resource_name("test-empty-values-provider")
        
        # Empty URL
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type vsphere "
            f"--url '' --username test --password test",
            check=False
        )
        assert result.returncode != 0
        
        # Empty username
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type vsphere "
            f"--url 'https://vcenter.example.com' --username '' --password test",
            check=False
        )
        assert result.returncode != 0


@pytest.mark.provider
@pytest.mark.error_cases
@pytest.mark.network
class TestProviderNetworkErrors:
    """Test cases for provider network-related errors."""
    
    def test_create_provider_unreachable_host(self, test_namespace):
        """Test creating a provider with unreachable host."""
        provider_name = generate_unique_resource_name("test-unreachable-provider")
        
        # Use a non-routable IP address (RFC 5737)
        unreachable_url = "https://192.0.2.1/sdk"
        
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type vsphere "
            f"--url '{unreachable_url}' --username test --password test",
            check=False
        )
        
        # This might succeed in creation but fail during validation
        # The exact behavior depends on how kubectl-mtv handles provider validation
        # We mainly want to ensure the command doesn't crash
        assert result.returncode in [0, 1]  # Allow both success and failure
    
    def test_create_provider_invalid_port(self, test_namespace):
        """Test creating a provider with invalid port."""
        provider_name = generate_unique_resource_name("test-invalid-port-provider")
        
        # Use an invalid port number
        invalid_url = "https://vcenter.example.com:99999/sdk"
        
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type vsphere "
            f"--url '{invalid_url}' --username test --password test",
            check=False
        )
        
        # Should fail due to invalid port
        assert result.returncode != 0
