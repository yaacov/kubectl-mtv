"""
Test cases for kubectl-mtv VMware ESXi provider creation.

This test validates the creation of VMware ESXi providers using the sdk-endpoint flag.
"""

import pytest

from ..utils import verify_provider_created


@pytest.mark.provider
@pytest.mark.esxi
@pytest.mark.requires_credentials
class TestESXiProvider:
    """Test cases for VMware ESXi provider creation."""

    def test_create_esxi_provider_skip_verify(
        self, test_namespace, provider_credentials
    ):
        """Test creating an ESXi provider with TLS verification skipped."""
        creds = provider_credentials["esxi"]

        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("VMware ESXi credentials not available in environment")

        provider_name = "test-esxi-skip-verify"

        # Create command with insecure skip TLS and sdk-endpoint esxi
        cmd_parts = [
            "create provider",
            provider_name,
            "--type vsphere",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            "--provider-insecure-skip-tls",
            "--sdk-endpoint esxi",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "vsphere")

    def test_create_esxi_provider_with_cacert(
        self, test_namespace, provider_credentials
    ):
        """Test creating an ESXi provider with CA certificate."""
        creds = provider_credentials["esxi"]

        # Skip if credentials or CA cert are not available
        required_fields = ["url", "username", "password", "cacert"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip(
                "VMware ESXi credentials with CA certificate not available in environment"
            )

        provider_name = "test-esxi-cacert"

        # Create command with CA cert and sdk-endpoint esxi
        cmd_parts = [
            "create provider",
            provider_name,
            "--type vsphere",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--cacert '{creds['cacert']}'",
            "--sdk-endpoint esxi",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "vsphere")

    def test_create_esxi_provider_with_vddk(
        self, test_namespace, provider_credentials
    ):
        """Test creating an ESXi provider with VDDK init image."""
        creds = provider_credentials["esxi"]

        # Skip if credentials or VDDK image are not available
        required_fields = ["url", "username", "password", "vddk_init_image"]
        if not all([creds.get(field) for field in required_fields]):
            pytest.skip(
                "VMware ESXi credentials with VDDK init image not available in environment"
            )

        provider_name = "test-esxi-vddk"

        # Create command with VDDK init image and sdk-endpoint esxi
        cmd_parts = [
            "create provider",
            provider_name,
            "--type vsphere",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--vddk-init-image '{creds['vddk_init_image']}'",
            "--provider-insecure-skip-tls",
            "--sdk-endpoint esxi",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Verify provider was created
        verify_provider_created(test_namespace, provider_name, "vsphere")

    def test_create_esxi_provider_error(self, test_namespace):
        """Test creating an ESXi provider with missing required fields."""
        provider_name = "test-esxi-error"

        # This should fail because ESXi (vSphere type) requires URL, username, password
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type vsphere --sdk-endpoint esxi", check=False
        )

        assert result.returncode != 0
