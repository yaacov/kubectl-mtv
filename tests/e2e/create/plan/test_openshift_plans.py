"""
Test cases for kubectl-mtv migration plan creation from OpenShift providers.

This test validates the creation of migration plans using OpenShift as the source provider
and verifies that plans become ready.
"""

import time

import pytest

from e2e.utils import wait_for_provider_ready, wait_for_plan_ready


# VM names that exist in the current cluster (created during namespace prep)
OPENSHIFT_TEST_VMS = ["test-vm-1", "test-vm-2"]


@pytest.mark.create
@pytest.mark.plan
@pytest.mark.openshift
@pytest.mark.skip(reason="Skipping OpenShift to OpenShift migration plan tests")
class TestOpenShiftPlanCreation:
    """Test cases for migration plan creation from OpenShift providers."""

    @pytest.fixture(scope="class")
    def openshift_provider(self, test_namespace):
        """Create an OpenShift provider for plan testing."""
        provider_name = "test-openshift-plan-source"

        # For OpenShift-to-OpenShift testing, use current cluster context for source
        # This ensures the VMs we created in namespace prep are available
        create_cmd = f"create provider {provider_name} --type openshift"

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Wait for provider to be ready
        wait_for_provider_ready(test_namespace, provider_name)

        return provider_name

    def test_create_plan_from_openshift(self, test_namespace, openshift_provider):
        """Test creating a migration plan from OpenShift provider."""
        # Use the first available VM as comma-separated string
        selected_vm = OPENSHIFT_TEST_VMS[0]
        plan_name = f"test-plan-openshift-{int(time.time())}"

        # Create plan command
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {openshift_provider}",
            "--target test-openshift-target",
            f"--vms '{selected_vm}'",
            "--target-namespace default",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create plan
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("plan", plan_name)

        # Wait for plan to be ready
        wait_for_plan_ready(test_namespace, plan_name)

    def test_create_multi_vm_plan_from_openshift(
        self, test_namespace, openshift_provider
    ):
        """Test creating a migration plan with multiple VMs from OpenShift provider."""
        # Use first 2 VMs for multi-VM test (OpenShift may have fewer VMs) as comma-separated string
        if len(OPENSHIFT_TEST_VMS) < 2:
            pytest.skip("Need at least 2 VMs for multi-VM test")

        selected_vms = ",".join(OPENSHIFT_TEST_VMS[:2])
        plan_name = f"test-multi-plan-openshift-{int(time.time())}"
        # Create plan command with multiple VMs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {openshift_provider}",
            "--target test-openshift-target",
            f"--vms '{selected_vms}'",
            "--target-namespace default",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create plan
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("plan", plan_name)

        # Wait for plan to be ready (longer timeout for multi-VM plans)
        wait_for_plan_ready(test_namespace, plan_name)
