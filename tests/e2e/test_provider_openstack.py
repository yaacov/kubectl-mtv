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
