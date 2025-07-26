"""
Test cases for kubectl-mtv migration plan creation from oVirt providers.

This test validates the creation of migration plans using oVirt as the source provider
and verifies that plans become ready.
"""

import time

import pytest

from e2e.utils import wait_for_provider_ready, wait_for_plan_ready


# Hardcoded VM names from oVirt inventory data
OVIRT_TEST_VMS = [
    "1111ab",
    "1111-win2019", 
    "3disks",
    "arik-win",
    "auto-rhv-red-iscsi-migration-50gb-70usage-vm-1"
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

        provider_name = "test-ovirt-plan-skip-verify"

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

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Wait for provider to be ready
        wait_for_provider_ready(test_namespace, provider_name)

        return provider_name

    def test_create_plan_from_ovirt(self, test_namespace, ovirt_provider):
        """Test creating a migration plan from oVirt provider."""
        # Use the first available VM as comma-separated string
        selected_vm = OVIRT_TEST_VMS[0]
        plan_name = f"test-plan-ovirt-{int(time.time())}"
        
        # Create plan command
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ovirt_provider}",
            "--target test-openshift-target",
            f"--vms {selected_vm}",
        ]
        
        create_cmd = " ".join(cmd_parts)
        
        # Create plan
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("plan", plan_name)
        
        # Wait for plan to be ready
        wait_for_plan_ready(test_namespace, plan_name)

    def test_create_multi_vm_plan_from_ovirt(self, test_namespace, ovirt_provider):
        """Test creating a migration plan with multiple VMs from oVirt provider."""
        # Use first 3 VMs for multi-VM test as comma-separated string
        selected_vms = ",".join(OVIRT_TEST_VMS[:3])
        plan_name = f"test-multi-plan-ovirt-{int(time.time())}"
        # Create plan command with multiple VMs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ovirt_provider}",
            "--target test-openshift-target",
            f"--vms {selected_vms}",
        ]
        
        create_cmd = " ".join(cmd_parts)
        
        # Create plan
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        
        # Track for cleanup
        test_namespace.track_resource("plan", plan_name)
        
        # Wait for plan to be ready (longer timeout for multi-VM plans)
        wait_for_plan_ready(test_namespace, plan_name, timeout=900) 