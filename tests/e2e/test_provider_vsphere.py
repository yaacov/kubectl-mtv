"""
Test cases for kubectl-mtv VMware vSphere provider creation.

This test validates the creation of VMware vSphere providers.
"""

import json
import time

import pytest


@pytest.mark.provider
@pytest.mark.vsphere
@pytest.mark.requires_credentials
class TestVSphereProvider:
    """Test cases for VMware vSphere provider creation."""
    
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
            cmd_parts.append("--provider-insecure-skip-tls")
        
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
    
    def test_create_vsphere_provider_with_insecure_tls(self, test_namespace, provider_credentials):
        """Test creating a vSphere provider with insecure TLS skip."""
        creds = provider_credentials["vsphere"]
        
        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("VMware vSphere credentials not available in environment")
        
        provider_name = "test-vsphere-insecure-provider"
        
        # Build create command with insecure TLS
        create_cmd = (
            f"create provider {provider_name} --type vsphere "
            f"--url '{creds['url']}' "
            f"--username '{creds['username']}' "
            f"--password '{creds['password']}' "
            "--provider-insecure-skip-tls"
        )
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        self._verify_provider_created(test_namespace, provider_name, "vsphere")
    
    def test_create_vsphere_provider_with_vddk(self, test_namespace, provider_credentials):
        """Test creating a vSphere provider with VDDK image."""
        creds = provider_credentials["vsphere"]
        
        # Skip if credentials or VDDK image are not available
        required_fields = ["url", "username", "password", "vddk_init_image"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip("VMware vSphere credentials with VDDK image not available in environment")
        
        provider_name = "test-vsphere-vddk-provider"
        
        # Build create command with VDDK
        cmd_parts = [
            f"create provider {provider_name} --type vsphere",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--vddk-init-image '{creds['vddk_init_image']}'"
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
        self._verify_provider_created(test_namespace, provider_name, "vsphere")
    
    def test_create_vsphere_provider_missing_credentials(self, test_namespace):
        """Test creating a vSphere provider with missing required fields."""
        provider_name = "test-incomplete-vsphere-provider"
        
        # This should fail because vsphere requires URL, username, password
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type vsphere",
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
