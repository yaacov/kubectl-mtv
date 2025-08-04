"""
Test cases for kubectl-mtv migration plan creation from OpenStack providers.

This test validates the creation of migration plans using OpenStack as the source provider
and verifies that plans become ready.
"""

import time

import pytest

from e2e.utils import (
    wait_for_plan_ready,
    generate_provider_name,
    get_or_create_provider,
)
from e2e.test_constants import TARGET_PROVIDER_NAME


OPENSTACK_TEST_VMS = ["infra-mtv-node-207", "infra-mtv-node-18"]


@pytest.mark.create
@pytest.mark.plan
@pytest.mark.openstack
@pytest.mark.requires_credentials
class TestOpenStackPlanCreation:
    """Test cases for migration plan creation from OpenStack providers."""

    @pytest.fixture(scope="class")
    def openstack_provider(self, test_namespace, provider_credentials):
        """Create an OpenStack provider for plan testing."""
        creds = provider_credentials["openstack"]

        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("OpenStack credentials not available in environment")

        # Generate provider name based on type and configuration
        provider_name = generate_provider_name("openstack", creds["url"], skip_tls=True)

        # Create command with insecure skip TLS
        cmd_parts = [
            "create provider",
            provider_name,
            "--type openstack",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--provider-domain-name '{creds['domain_name']}'",
            f"--provider-project-name '{creds['project_name']}'",
            "--provider-insecure-skip-tls",
        ]

        if creds.get("region_name"):
            cmd_parts.append(f"--provider-region-name '{creds['region_name']}'")

        create_cmd = " ".join(cmd_parts)

        # Create provider if it doesn't already exist
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    def test_create_plan_from_openstack(self, test_namespace, openstack_provider):
        """Test creating a migration plan from OpenStack provider."""
        # Use the first available VM as comma-separated string
        selected_vm = OPENSTACK_TEST_VMS[0]
        plan_name = f"test-plan-openstack-{int(time.time())}"

        # Create plan command
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {openstack_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
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

    def test_create_multi_vm_plan_from_openstack(
        self, test_namespace, openstack_provider
    ):
        """Test creating a migration plan with multiple VMs from OpenStack provider."""
        # Use first 3 VMs for multi-VM test as comma-separated string
        selected_vms = ",".join(OPENSTACK_TEST_VMS[:3])
        plan_name = f"test-multi-plan-openstack-{int(time.time())}"
        # Create plan command with multiple VMs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {openstack_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
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
