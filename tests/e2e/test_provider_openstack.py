"""
Test cases for kubectl-mtv OpenStack provider creation.

This test validates the creation of OpenStack providers.
"""

import json
import time

import pytest


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
        
        provider_name = "test-openstack-provider"
        
        # Build create command
        cmd_parts = [
            "create provider", provider_name,
            "--type openstack",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--domain-name '{creds['domain_name']}'",
            f"--project-name '{creds['project_name']}'"
        ]
        
        if creds.get("region_name"):
            cmd_parts.append(f"--region-name '{creds['region_name']}'")
        
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
        self._verify_provider_created(test_namespace, provider_name, "openstack")
    
    def test_create_openstack_provider_with_region(self, test_namespace, provider_credentials):
        """Test creating an OpenStack provider with specific region."""
        creds = provider_credentials["openstack"]
        
        # Skip if credentials are not available
        required_fields = ["url", "username", "password", "domain_name", "project_name", "region_name"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip("OpenStack credentials with region not available in environment")
        
        provider_name = "test-openstack-region-provider"
        
        # Build create command with specific region
        cmd_parts = [
            f"create provider {provider_name} --type openstack",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--domain-name '{creds['domain_name']}'",
            f"--project-name '{creds['project_name']}'",
            f"--region-name '{creds['region_name']}'"
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
        self._verify_provider_created(test_namespace, provider_name, "openstack")
    
    def test_create_openstack_provider_with_insecure_tls(self, test_namespace, provider_credentials):
        """Test creating an OpenStack provider with insecure TLS skip."""
        creds = provider_credentials["openstack"]
        
        # Skip if credentials are not available
        required_fields = ["url", "username", "password", "domain_name", "project_name"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip("OpenStack credentials not available in environment")
        
        provider_name = "test-openstack-insecure-provider"
        
        # Build create command with insecure TLS
        create_cmd = (
            f"create provider {provider_name} --type openstack "
            f"--url '{creds['url']}' "
            f"--username '{creds['username']}' "
            f"--password '{creds['password']}' "
            f"--domain-name '{creds['domain_name']}' "
            f"--project-name '{creds['project_name']}' "
            "--provider-insecure-skip-tls"
        )
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        self._verify_provider_created(test_namespace, provider_name, "openstack")
    
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
