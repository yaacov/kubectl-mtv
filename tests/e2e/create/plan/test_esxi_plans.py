"""
Test cases for kubectl-mtv migration plan creation from ESXi providers.

This test validates the creation of migration plans using ESXi as the source provider
and verifies that plans become ready.
ESXi is essentially vSphere with sdk-endpoint set to esxi.
"""

import time

import pytest

from e2e.utils import (
    wait_for_plan_ready,
)
from e2e.test_constants import ESXI_TEST_VMS


@pytest.mark.create
@pytest.mark.plan
@pytest.mark.esxi
@pytest.mark.requires_credentials
class TestESXiPlanCreation:
    """Test cases for migration plan creation from ESXi providers."""

    # Provider fixtures are now session-scoped in conftest.py

    def test_create_plan_from_esxi(
        self, test_namespace, esxi_provider, openshift_provider
    ):
        """Test creating a migration plan from ESXi provider."""
        # Use the first available VM as comma-separated string
        selected_vm = ESXI_TEST_VMS[0]
        plan_name = f"test-plan-esxi-{int(time.time())}"

        # Create plan command
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {esxi_provider}",
            f"--target {openshift_provider}",
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

    def test_create_multi_vm_plan_from_esxi(
        self, test_namespace, esxi_provider, openshift_provider
    ):
        """Test creating a migration plan with multiple VMs from ESXi provider."""
        # Use first 3 VMs for multi-VM test as comma-separated string
        selected_vms = ",".join(ESXI_TEST_VMS[:3])
        plan_name = f"test-multi-plan-esxi-{int(time.time())}"
        # Create plan command with multiple VMs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {esxi_provider}",
            f"--target {openshift_provider}",
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
