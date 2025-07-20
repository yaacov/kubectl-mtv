"""
Test cases for kubectl-mtv oVirt provider creation.

This test validates the creation of oVirt providers.
"""

import json

import pytest

from utils import verify_provider_created


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
        result = test_namespace.run_mtv_command(f"get provider {provider_name} -o json", check=False)
        
        # If command failed, provider doesn't exist
        if result.returncode != 0:
            pytest.fail(f"Provider {provider_name} not found. Command output: {result.stderr}")
        
        # Parse provider data
        try:
            provider_list = json.loads(result.stdout)
        except json.JSONDecodeError as e:
            pytest.fail(f"Failed to parse provider JSON output: {e}. Output: {result.stdout}")
        
        if len(provider_list) != 1:
            pytest.fail(f"Expected 1 provider, got {len(provider_list)}")
        
        provider_data = provider_list[0]
        
        # Verify basic provider properties
        metadata = provider_data.get("metadata", {})
        spec = provider_data.get("spec", {})
        
        if metadata.get("name") != provider_name:
            pytest.fail(f"Provider name mismatch: expected {provider_name}, got {metadata.get('name')}")
        
        if spec.get("type") != provider_type:
            pytest.fail(f"Provider type mismatch: expected {provider_type}, got {spec.get('type')}")
        
        # Check provider status for errors
        status = provider_data.get("status", {})
        print(f"Provider {provider_name} status: {status}")
        
        # Check for error conditions
        conditions = status.get("conditions", [])
        for condition in conditions:
            condition_type = condition.get("type", "")
            condition_status = condition.get("status", "")
            condition_reason = condition.get("reason", "")
            condition_message = condition.get("message", "")
            
            # Fail if there are explicit error conditions
            if condition_type in ["Ready", "ConnectionTestSucceeded"] and condition_status == "False":
                if condition_reason in ["Error", "Failed", "ValidationFailed", "ConnectionFailed"]:
                    pytest.fail(
                        f"Provider {provider_name} is in error state. "
                        f"Condition: {condition_type}={condition_status}, "
                        f"Reason: {condition_reason}, Message: {condition_message}"
                    )
        
        # Additional validation: check if provider has been marked as invalid
        if status.get("phase") == "Failed":
            pytest.fail(f"Provider {provider_name} is in Failed phase: {status}")
        
        print(f"Provider {provider_name} verified successfully")
