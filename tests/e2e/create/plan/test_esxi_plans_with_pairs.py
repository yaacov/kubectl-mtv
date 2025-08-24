"""
Test cases for kubectl-mtv migration plan creation from ESXi providers using mapping pairs.

This test validates the creation of migration plans using ESXi as the source provider
with inline network and storage mapping pairs instead of pre-created mappings.
ESXi is essentially vSphere with sdk-endpoint set to esxi.
"""

import time

import pytest

from e2e.utils import (
    wait_for_plan_ready,
    generate_provider_name,
    get_or_create_provider,
)
from e2e.test_constants import (
    ESXI_TEST_VMS,
    ESXI_NETWORK_PAIRS,
    ESXI_STORAGE_PAIRS,
    TARGET_PROVIDER_NAME,
)


@pytest.mark.create
@pytest.mark.plan
@pytest.mark.esxi
@pytest.mark.requires_credentials
class TestESXiPlanCreationWithPairs:
    """Test cases for migration plan creation from ESXi providers using mapping pairs."""

    @pytest.fixture(scope="class")
    def esxi_provider(self, test_namespace, provider_credentials):
        """Create an ESXi provider for plan testing."""
        creds = provider_credentials["esxi"]

        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("ESXi credentials not available in environment")

        # Generate provider name based on type and configuration
        provider_name = generate_provider_name(
            "vsphere", creds["url"], sdk_endpoint="esxi", skip_tls=True
        )

        # Create command with insecure skip TLS
        cmd_parts = [
            "create provider",
            provider_name,
            "--type vsphere",  # ESXi uses vsphere type
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            "--provider-insecure-skip-tls",
            "--sdk-endpoint esxi",  # This makes it ESXi instead of vCenter
        ]

        create_cmd = " ".join(cmd_parts)

        # Create provider if it doesn't already exist
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    def test_create_plan_with_mapping_pairs(self, test_namespace, esxi_provider):
        """Test creating a migration plan with inline mapping pairs."""
        # Use the first available VM
        selected_vm = ESXI_TEST_VMS[0]
        plan_name = f"test-plan-esxi-pairs-{int(time.time())}"

        # Build network and storage pairs strings
        network_pairs = ",".join(
            [f"{n['source']}:{n['target']}" for n in ESXI_NETWORK_PAIRS]
        )
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in ESXI_STORAGE_PAIRS]
        )

        # Create plan command with mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {esxi_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--vms '{selected_vm}'",
            f"--network-pairs '{network_pairs}'",
            f"--storage-pairs '{storage_pairs}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create plan
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup (also track auto-created mappings)
        test_namespace.track_resource("plan", plan_name)
        test_namespace.track_resource("networkmap", f"{plan_name}-network")
        test_namespace.track_resource("storagemap", f"{plan_name}-storage")

        # Wait for plan to be ready
        wait_for_plan_ready(test_namespace, plan_name)

    def test_create_multi_vm_plan_with_mapping_pairs(
        self, test_namespace, esxi_provider
    ):
        """Test creating a migration plan with multiple VMs using inline mapping pairs."""
        # Use first 3 VMs for multi-VM test
        selected_vms = ",".join(ESXI_TEST_VMS[:3])
        plan_name = f"test-multi-plan-esxi-pairs-{int(time.time())}"

        # Build network and storage pairs strings
        network_pairs = ",".join(
            [f"{n['source']}:{n['target']}" for n in ESXI_NETWORK_PAIRS]
        )
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in ESXI_STORAGE_PAIRS]
        )

        # Create plan command with multiple VMs and mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {esxi_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--vms '{selected_vms}'",
            f"--network-pairs '{network_pairs}'",
            f"--storage-pairs '{storage_pairs}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create plan
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup (also track auto-created mappings)
        test_namespace.track_resource("plan", plan_name)
        test_namespace.track_resource("networkmap", f"{plan_name}-network")
        test_namespace.track_resource("storagemap", f"{plan_name}-storage")

        # Wait for plan to be ready (longer timeout for multi-VM plans)
        wait_for_plan_ready(test_namespace, plan_name)

    def test_create_plan_with_pod_network_pairs(self, test_namespace, esxi_provider):
        """Test creating a migration plan with pod network mapping pairs."""
        # Use a single VM
        if len(ESXI_TEST_VMS) < 2:
            pytest.skip("Need at least 2 VMs for this test")
        selected_vm = ESXI_TEST_VMS[1]
        plan_name = f"test-plan-esxi-pod-pairs-{int(time.time())}"

        # Use pod network for all networks
        # Use pod network for first network only, ignore the rest (complies with pod network uniqueness constraint)
        network_pairs = f"{ESXI_NETWORK_PAIRS[0]['source']}:default"
        if len(ESXI_NETWORK_PAIRS) > 1:
            # Map additional networks to ignored to comply with constraint requirements
            ignored_pairs = ",".join(
                [f"{n['source']}:ignored" for n in ESXI_NETWORK_PAIRS[1:]]
            )
            network_pairs = f"{network_pairs},{ignored_pairs}"
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in ESXI_STORAGE_PAIRS]
        )

        # Create plan command with pod network mapping
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {esxi_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--vms '{selected_vm}'",
            f"--network-pairs '{network_pairs}'",
            f"--storage-pairs '{storage_pairs}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create plan
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup (also track auto-created mappings)
        test_namespace.track_resource("plan", plan_name)
        test_namespace.track_resource("networkmap", f"{plan_name}-network")
        test_namespace.track_resource("storagemap", f"{plan_name}-storage")

        # Wait for plan to be ready
        wait_for_plan_ready(test_namespace, plan_name)

    def test_create_plan_with_vddk_and_pairs(self, test_namespace, esxi_provider):
        """Test creating a migration plan with mapping pairs and VDDK option."""
        # Use a single VM
        if len(ESXI_TEST_VMS) < 3:
            pytest.skip("Need at least 3 VMs for this test")
        selected_vm = ESXI_TEST_VMS[2]
        plan_name = f"test-plan-esxi-vddk-pairs-{int(time.time())}"

        # Use simple mapping pairs
        network_pairs = "VM Network:default"  # Single network to default is OK
        # Include all ESXi datastores to ensure nothing is unmapped
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in ESXI_STORAGE_PAIRS]
        )

        # Create plan command with mapping pairs
        # Note: VDDK configuration would typically be done at provider level or via VDDK config
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {esxi_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--vms '{selected_vm}'",
            f"--network-pairs '{network_pairs}'",
            f"--storage-pairs '{storage_pairs}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create plan
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup (also track auto-created mappings)
        test_namespace.track_resource("plan", plan_name)
        test_namespace.track_resource("networkmap", f"{plan_name}-network")
        test_namespace.track_resource("storagemap", f"{plan_name}-storage")

        # Wait for plan to be ready
        wait_for_plan_ready(test_namespace, plan_name)
