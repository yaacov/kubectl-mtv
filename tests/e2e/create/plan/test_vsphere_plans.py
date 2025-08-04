"""
Test cases for kubectl-mtv migration plan creation from vSphere providers.

This test validates the creation of migration plans using vSphere as the source provider
and verifies that plans become ready.
"""

import time

import pytest

from e2e.utils import (
    wait_for_plan_ready,
    generate_provider_name,
    get_or_create_provider,
)
from e2e.test_constants import VSPHERE_TEST_VMS


@pytest.mark.create
@pytest.mark.plan
@pytest.mark.vsphere
@pytest.mark.requires_credentials
class TestVSpherePlanCreation:
    """Test cases for migration plan creation from vSphere providers."""

    @pytest.fixture(scope="class")
    def vsphere_provider(self, test_namespace, provider_credentials):
        """Create a vSphere provider for plan testing."""
        creds = provider_credentials["vsphere"]

        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("vSphere credentials not available in environment")

        # Generate provider name based on type and configuration
        provider_name = generate_provider_name(
            "vsphere", creds["url"], skip_tls=creds.get("insecure", False)
        )

        # Create command
        cmd_parts = [
            "create provider",
            provider_name,
            "--type vsphere",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
        ]

        # Add insecure skip TLS if specified
        if creds.get("insecure"):
            cmd_parts.append("--provider-insecure-skip-tls")

        create_cmd = " ".join(cmd_parts)

        # Create provider if it doesn't already exist
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    @pytest.fixture(scope="class")
    def target_provider(self, test_namespace):
        """Ensure the target OpenShift provider exists for plan testing."""
        # Generate provider name based on type and configuration
        provider_name = generate_provider_name("openshift", "localhost", skip_tls=True)

        # Create command for OpenShift target provider
        create_cmd = f"create provider {provider_name} --type openshift"

        # Create provider if it doesn't already exist
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    def test_create_plan_from_vsphere(
        self, test_namespace, vsphere_provider, target_provider
    ):
        """Test creating a migration plan from vSphere provider."""
        # Use the first available VM as comma-separated string
        selected_vm = VSPHERE_TEST_VMS[0]
        plan_name = f"test-plan-vsphere-{int(time.time())}"

        # Create plan command
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {vsphere_provider}",
            f"--target {target_provider}",
            f"--vms '{selected_vm}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create plan
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("plan", plan_name)

        # Wait for plan to be ready
        wait_for_plan_ready(test_namespace, plan_name)

    def test_create_multi_vm_plan_from_vsphere(
        self, test_namespace, vsphere_provider, target_provider
    ):
        """Test creating a migration plan with multiple VMs from vSphere provider."""
        # Use multiple VMs from the inventory dump data
        if len(VSPHERE_TEST_VMS) < 2:
            pytest.skip("Need at least 2 VMs for multi-VM test")

        # Use all available VMs for comprehensive testing
        selected_vms = ",".join(VSPHERE_TEST_VMS)
        plan_name = f"test-multi-plan-vsphere-{int(time.time())}"

        # Create plan command with multiple VMs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {vsphere_provider}",
            f"--target {target_provider}",
            f"--vms '{selected_vms}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create plan
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("plan", plan_name)

        # Wait for plan to be ready (longer timeout for multi-VM plans)
        wait_for_plan_ready(test_namespace, plan_name)
