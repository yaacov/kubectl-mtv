"""
Test cases for kubectl-mtv provider creation.

This test validates the creation of different provider types.
"""

import json
import time
from typing import Dict, Any

import pytest


@pytest.mark.provider
class TestProviderCreation:
    """Test cases for provider creation commands."""
    
    @pytest.mark.requires_credentials
    def test_create_vsphere_provider(self, test_namespace, provider_credentials):
        """Test creating a VMware vSphere provider."""
        creds = provider_credentials["vsphere"]
        
        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("VMware vSphere credentials not available in environment")
        
        provider_name = "test-vsphere-provider"
        
        # Build create command
        cmd_parts = [
            "create provider", provider_name,
            "--type vsphere",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'"
        ]
        
        if creds.get("cacert"):
            cmd_parts.append(f"--cacert '{creds['cacert']}'")
        
        if creds.get("insecure"):
            cmd_parts.append("--insecure-skip-tls")
        
        if creds.get("vddk_init_image"):
            cmd_parts.append(f"--vddk-init-image '{creds['vddk_init_image']}'")
        
        create_cmd = " ".join(cmd_parts)
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        self._verify_provider_created(test_namespace, provider_name, "vsphere")
    
    @pytest.mark.requires_credentials
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
            cmd_parts.append("--insecure-skip-tls")
        
        create_cmd = " ".join(cmd_parts)
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        self._verify_provider_created(test_namespace, provider_name, "ovirt")
    
    @pytest.mark.requires_credentials
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
            cmd_parts.append("--insecure-skip-tls")
        
        create_cmd = " ".join(cmd_parts)
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        self._verify_provider_created(test_namespace, provider_name, "openstack")
    
    @pytest.mark.requires_credentials
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
    
    def test_create_openshift_provider(self, test_namespace, provider_credentials):
        """Test creating an OpenShift target provider."""
        creds = provider_credentials["openshift"]
        provider_name = "test-openshift-provider"
        
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
        self._verify_provider_created(test_namespace, provider_name, "openshift")
    
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
    
    def test_create_provider_missing_required_fields(self, test_namespace):
        """Test creating a provider with missing required fields."""
        provider_name = "test-incomplete-provider"
        
        # This should fail because vsphere requires URL, username, password
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type vsphere",
            check=False
        )
        
        assert result.returncode != 0
    
    def test_list_providers_after_creation(self, test_namespace):
        """Test listing providers after creating one."""
        provider_name = "test-list-provider"
        
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
        provider_name = "test-details-provider"
        
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
        provider_data = json.loads(result.stdout)
        assert isinstance(provider_data, dict)
        assert provider_data.get("metadata", {}).get("name") == provider_name
    
    def _verify_provider_created(self, test_namespace, provider_name: str, provider_type: str):
        """Verify that a provider was created successfully."""
        # Wait a moment for provider to be created
        time.sleep(2)
        
        # Check if provider exists
        result = test_namespace.run_mtv_command(f"get provider {provider_name} -o json")
        assert result.returncode == 0
        
        # Parse provider data
        provider_data = json.loads(result.stdout)
        assert provider_data.get("metadata", {}).get("name") == provider_name
        assert provider_data.get("spec", {}).get("type") == provider_type
        
        # Check provider status (might take a moment to be ready)
        # This is optional as provider validation might take time
        status = provider_data.get("status", {})
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
    
    def test_delete_provider(self, test_namespace):
        """Test deleting a provider."""
        provider_name = "test-delete-provider"
        
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
