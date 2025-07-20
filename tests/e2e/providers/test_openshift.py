"""
Test cases for kubectl-mtv OpenShift provider creation.

This test validates the creation of OpenShift target providers.
"""

import pytest

from ..utils import verify_provider_created


@pytest.mark.provider
@pytest.mark.openshift
class TestOpenShiftProvider:
    """Test cases for OpenShift provider creation."""
    
    def test_create_openshift_provider_skip_verify(self, test_namespace, provider_credentials):
        """Test creating an OpenShift provider with TLS verification skipped."""
        creds = provider_credentials["openshift"]
        provider_name = "test-openshift-skip-verify"
        
        # For OpenShift provider, we can often use the current cluster
        if creds.get("url") and creds.get("token"):
            # Use explicit credentials with skip verify
            create_cmd = (
                f"create provider {provider_name} --type openshift "
                f"--url '{creds['url']}' --token '{creds['token']}' "
                "--provider-insecure-skip-tls"
            )
        else:
            # Use current cluster context with skip verify (this may not apply the flag effectively)
            create_cmd = f"create provider {provider_name} --type openshift --provider-insecure-skip-tls"
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "openshift")

    def test_create_openshift_provider_with_cacert(self, test_namespace, provider_credentials):
        """Test creating an OpenShift provider with CA certificate."""
        creds = provider_credentials["openshift"]
        
        # Skip if CA cert is not available
        if not creds.get("cacert"):
            pytest.skip("OpenShift CA certificate not available in environment")
        
        provider_name = "test-openshift-cacert"
        
        if creds.get("url") and creds.get("token"):
            # Use explicit credentials with CA cert
            create_cmd = (
                f"create provider {provider_name} --type openshift "
                f"--url '{creds['url']}' --token '{creds['token']}' "
                f"--cacert '{creds['cacert']}'"
            )
        else:
            # Use current cluster context with CA cert (may not be applicable)
            create_cmd = f"create provider {provider_name} --type openshift --cacert '{creds['cacert']}'"
        
        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "openshift")

    def test_create_openshift_provider_error(self, test_namespace):
        """Test creating an OpenShift provider with invalid configuration."""
        provider_name = "test-openshift-error"
        
        # This should fail because providing a token without URL is invalid
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type openshift --token 'invalid-token-without-url'",
            check=False
        )
        
        assert result.returncode != 0
