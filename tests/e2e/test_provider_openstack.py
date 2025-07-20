"""
Test cases for kubectl-mtv OpenStack provider creation.

This test validates the creation of OpenStack providers.
"""

import json
import time

import pytest

from utils import verify_provider_created, generate_unique_resource_name


@pytest.mark.provider
@pytest.mark.openstack
@pytest.mark.requires_credentials
class TestOpenStackProvider:
    """Test cases for OpenStack provider creation."""
    
    def test_create_openstack_provider(self, test_namespace, provider_credentials):
        """Test creating an OpenStack provider."""
        creds = provider_credentials["openstack"]
        
        # Skip if credentials are not available
        required_fields = ["url", "username", "password", "domain_name", "project_name"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip("OpenStack credentials not available in environment")
        
        provider_name = generate_unique_resource_name("test-openstack-provider")
        
        # Build create command
        cmd_parts = [
            "create provider", provider_name,
            "--type openstack",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--provider-domain-name '{creds['domain_name']}'",
            f"--provider-project-name '{creds['project_name']}'"
        ]
        
        if creds.get("region_name"):
            cmd_parts.append(f"--provider-region-name '{creds['region_name']}'")
        
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
        verify_provider_created(test_namespace, provider_name, "openstack")
    
    def test_create_openstack_provider_with_region(self, test_namespace, provider_credentials):
        """Test creating an OpenStack provider with specific region."""
        creds = provider_credentials["openstack"]
        
        # Skip if credentials are not available
        required_fields = ["url", "username", "password", "domain_name", "project_name", "region_name"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip("OpenStack credentials with region not available in environment")
        
        provider_name = generate_unique_resource_name("test-openstack-region-provider")
        
        # Build create command with specific region
        cmd_parts = [
            f"create provider {provider_name} --type openstack",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--provider-domain-name '{creds['domain_name']}'",
            f"--provider-project-name '{creds['project_name']}'",
            f"--provider-region-name '{creds['region_name']}'"
        ]
        
        # Add insecure flag if specified in credentials
        if creds.get("insecure"):
            cmd_parts.append("--provider-insecure-skip-tls")
        
        # Add CA cert if specified in credentials
        if creds.get("cacert"):
            cmd_parts.append(f"--cacert '{creds['cacert']}'")
        
        create_cmd = " ".join(cmd_parts)
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "openstack")
    
    def test_create_openstack_provider_with_insecure_tls(self, test_namespace, provider_credentials):
        """Test creating an OpenStack provider with insecure TLS skip."""
        creds = provider_credentials["openstack"]
        
        # Skip if credentials are not available
        required_fields = ["url", "username", "password", "domain_name", "project_name"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip("OpenStack credentials not available in environment")
        
        provider_name = generate_unique_resource_name("test-openstack-insecure-provider")
        
        # Build create command with insecure TLS
        create_cmd = (
            f"create provider {provider_name} --type openstack "
            f"--url '{creds['url']}' "
            f"--username '{creds['username']}' "
            f"--password '{creds['password']}' "
            f"--provider-domain-name '{creds['domain_name']}' "
            f"--provider-project-name '{creds['project_name']}' "
            f"--provider-region-name '{creds['region_name']}' "
            "--provider-insecure-skip-tls"
        )
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "openstack")
