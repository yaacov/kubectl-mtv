"""
Test cases for kubectl-mtv oVirt provider creation.

This test validates the creation of oVirt providers.
"""

import json
import time

import pytest


@pytest.mark.provider
@pytest.mark.ovirt
@pytest.mark.requires_credentials
class TestOVirtProvider:
    """Test cases for oVirt provider creation."""
    
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
            cmd_parts.append("--provider-insecure-skip-tls")
        
        create_cmd = " ".join(cmd_parts)
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        self._verify_provider_created(test_namespace, provider_name, "ovirt")
    
    def test_create_ovirt_provider_with_insecure_tls(self, test_namespace, provider_credentials):
        """Test creating an oVirt provider with insecure TLS skip."""
        creds = provider_credentials["ovirt"]
        
        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("oVirt credentials not available in environment")
        
        provider_name = "test-ovirt-insecure-provider"
        
        # Build create command with insecure TLS
        create_cmd = (
            f"create provider {provider_name} --type ovirt "
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
        self._verify_provider_created(test_namespace, provider_name, "ovirt")
    
    def test_create_ovirt_provider_with_ca_cert(self, test_namespace, provider_credentials, temp_file):
        """Test creating an oVirt provider with CA certificate."""
        creds = provider_credentials["ovirt"]
        
        # Skip if credentials or CA cert are not available
        required_fields = ["url", "username", "password", "cacert"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip("oVirt credentials with CA certificate not available in environment")
        
        provider_name = "test-ovirt-cacert-provider"
        
        # Write CA cert to temp file
        with open(temp_file, 'w') as f:
            f.write(creds['cacert'])
        
        # Build create command with CA cert file
        create_cmd = (
            f"create provider {provider_name} --type ovirt "
            f"--url '{creds['url']}' "
            f"--username '{creds['username']}' "
            f"--password '{creds['password']}' "
            f"--cacert @{temp_file}"
        )
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        self._verify_provider_created(test_namespace, provider_name, "ovirt")
    
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
