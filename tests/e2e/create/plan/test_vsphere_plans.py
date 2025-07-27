"""
Test cases for kubectl-mtv migration plan creation from vSphere providers.

This test validates the creation of migration plans using vSphere as the source provider
and verifies that plans become ready.
"""

import time

import pytest

from e2e.utils import wait_for_provider_ready, wait_for_plan_ready


# Hardcoded VM names from vSphere inventory data
VSPHERE_TEST_VMS = [
    "mtv-win2019-79-ceph-rbd-4-16",
    "mtv-func-rhel8-ameen", 
    "mtv-rhel8-warm-sanity-nfs-4-19",
    "mtv-rhel8-warm-2disks2nics-nfs-4-18",
    "mtv-rhel8-warm-sanity-nfs-4-18"
]


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
            pytest.skip("VMware vSphere credentials not available in environment")

        provider_name = "test-vsphere-plan-skip-verify"

        # Create command with insecure skip TLS
        cmd_parts = [
            "create provider",
            provider_name,
            "--type vsphere",
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

    def test_create_plan_from_vsphere(self, test_namespace, vsphere_provider):
        """Test creating a migration plan from vSphere provider."""
        # Use the first available VM as comma-separated string
        selected_vm = VSPHERE_TEST_VMS[0]
        plan_name = f"test-plan-vsphere-{int(time.time())}"
        
        # Create plan command
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {vsphere_provider}",
            "--target test-openshift-target",
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

    def test_create_multi_vm_plan_from_vsphere(self, test_namespace, vsphere_provider):
        """Test creating a migration plan with multiple VMs from vSphere provider."""
        # Use first 3 VMs for multi-VM test as comma-separated string
        selected_vms = ",".join(VSPHERE_TEST_VMS[:3])
        plan_name = f"test-multi-plan-vsphere-{int(time.time())}"
        # Create plan command with multiple VMs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {vsphere_provider}",
            "--target test-openshift-target",
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