"""
Test cases for kubectl-mtv migration plan creation from oVirt providers.

This test validates the creation of migration plans using oVirt as the source provider
and verifies that plans become ready.
"""

import time

import pytest

from e2e.utils import (
    wait_for_plan_ready,
    generate_provider_name,
    get_or_create_provider,
)


# Hardcoded VM names from oVirt inventory data
OVIRT_TEST_VMS = [
    "1111ab",
    "1111-win2019",
    "3disks",
    "arik-win",
    "auto-rhv-red-iscsi-migration-50gb-70usage-vm-1",
]


@pytest.mark.create
@pytest.mark.plan
@pytest.mark.ovirt
@pytest.mark.requires_credentials
class TestOvirtPlanCreation:
    """Test cases for migration plan creation from oVirt providers."""

    @pytest.fixture(scope="class")
    def ovirt_provider(self, test_namespace, provider_credentials):
        """Create an oVirt provider for plan testing."""
        creds = provider_credentials["ovirt"]

        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("oVirt credentials not available in environment")

        # Generate provider name based on type and configuration
        provider_name = generate_provider_name("ovirt", creds["url"], skip_tls=True)

        # Create command with insecure skip TLS
        cmd_parts = [
            "create provider",
            provider_name,
            "--type ovirt",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            "--provider-insecure-skip-tls",
        ]

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

    def test_create_plan_from_ovirt(
        self, test_namespace, ovirt_provider, target_provider
    ):
        """Test creating a migration plan from oVirt provider."""
        # Use the first available VM as comma-separated string
        selected_vm = OVIRT_TEST_VMS[0]
        plan_name = f"test-plan-ovirt-{int(time.time())}"

        # Create plan command
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ovirt_provider}",
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

    def test_create_multi_vm_plan_from_ovirt(
        self, test_namespace, ovirt_provider, target_provider
    ):
        """Test creating a migration plan with multiple VMs from oVirt provider."""
        # Use multiple VMs from the inventory dump data
        if len(OVIRT_TEST_VMS) < 2:
            pytest.skip("Need at least 2 VMs for multi-VM test")

        # Use all available VMs for comprehensive testing
        selected_vms = ",".join(OVIRT_TEST_VMS)
        plan_name = f"test-multi-plan-ovirt-{int(time.time())}"

        # Create plan command with multiple VMs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ovirt_provider}",
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
