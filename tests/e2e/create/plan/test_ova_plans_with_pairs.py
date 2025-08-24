"""
Test cases for kubectl-mtv migration plan creation from OVA providers using mapping pairs.

This test validates the creation of migration plans using OVA as the source provider
with inline network and storage mapping pairs instead of pre-created mappings.
"""

import time

import pytest

from e2e.utils import (
    wait_for_plan_ready,
    generate_provider_name,
    get_or_create_provider,
)
from e2e.test_constants import (
    OVA_TEST_VMS,
    OVA_VM_STORAGE_MAPPINGS,
    OVA_VM_NETWORK_MAPPINGS,
    TARGET_PROVIDER_NAME,
)


@pytest.mark.create
@pytest.mark.plan
@pytest.mark.ova
@pytest.mark.requires_credentials
class TestOVAPlanCreationWithPairs:
    """Test cases for migration plan creation from OVA providers using mapping pairs."""

    @pytest.fixture(scope="class")
    def ova_provider(self, test_namespace, provider_credentials):
        """Create an OVA provider for plan testing."""
        creds = provider_credentials.get("ova", {})

        # Skip if OVA URL is not available
        if not creds.get("url"):
            pytest.skip("OVA URL not available in environment")

        # Generate provider name based on type and configuration
        provider_name = generate_provider_name("ova", creds["url"], skip_tls=True)

        # Create command for OVA provider with URL
        create_cmd = (
            f"create provider {provider_name} --type ova --url '{creds['url']}'"
        )

        # Create provider if it doesn't already exist
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    def test_create_plan_with_mapping_pairs(self, test_namespace, ova_provider):
        """Test creating a migration plan with inline mapping pairs."""
        # Use the first available VM (mtv-2disks which has 2 disks)
        selected_vm = OVA_TEST_VMS[0]
        plan_name = f"test-plan-ova-pairs-{int(time.time())}"

        # Get VM-specific network and storage mappings
        vm_networks = OVA_VM_NETWORK_MAPPINGS[selected_vm]
        vm_storage = OVA_VM_STORAGE_MAPPINGS[selected_vm]

        # Build network and storage pairs strings
        network_pairs = ",".join([f"{n['source']}:{n['target']}" for n in vm_networks])
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in vm_storage])

        # Create plan command with mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ova_provider}",
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

    def test_create_multi_vm_plan_with_pairs(self, test_namespace, ova_provider):
        """Test creating a migration plan with multiple VMs using complete mapping pairs."""
        # Use all available VMs
        selected_vms = ",".join(OVA_TEST_VMS)
        plan_name = f"test-multi-plan-ova-pairs-{int(time.time())}"

        # Build comprehensive network and storage pairs covering all VMs
        all_networks = set()
        all_storage = []

        for vm_name in OVA_TEST_VMS:
            # Collect unique networks
            for net in OVA_VM_NETWORK_MAPPINGS[vm_name]:
                all_networks.add((net["source"], net["target"]))
            # Collect all storage mappings
            all_storage.extend(OVA_VM_STORAGE_MAPPINGS[vm_name])

        network_pairs = ",".join(
            [f"{source}:{target}" for source, target in all_networks]
        )
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in all_storage])

        # Create plan command with all VMs and mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ova_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--vms '{selected_vms}'",
            f"--network-pairs '{network_pairs}'",
            f"--storage-pairs '{storage_pairs}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create plan
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0, "Failed to create multi-VM plan with pairs"

        # Track for cleanup
        test_namespace.track_resource("plan", plan_name)
        test_namespace.track_resource("networkmap", f"{plan_name}-network")
        test_namespace.track_resource("storagemap", f"{plan_name}-storage")

        # Wait for plan to be ready (longer timeout for multi-VM plans)
        wait_for_plan_ready(test_namespace, plan_name)

    def test_create_plan_with_pod_network_pairs(self, test_namespace, ova_provider):
        """Test creating a migration plan with pod network mapping pairs."""
        # Use the first available VM (mtv-2disks which has 2 disks)
        selected_vm = OVA_TEST_VMS[0]
        plan_name = f"test-plan-ova-pod-pairs-{int(time.time())}"

        # Get VM-specific mappings
        vm_networks = OVA_VM_NETWORK_MAPPINGS[selected_vm]
        vm_storage = OVA_VM_STORAGE_MAPPINGS[selected_vm]

        # Use pod network for all networks
        # Use pod network for first network only, ignore the rest (complies with pod network uniqueness constraint)
        if vm_networks:
            network_pairs = f"{vm_networks[0]['source']}:default"
            if len(vm_networks) > 1:
                # Map additional networks to ignored to comply with constraint requirements
                ignored_pairs = ",".join(
                    [f"{n['source']}:ignored" for n in vm_networks[1:]]
                )
                network_pairs = f"{network_pairs},{ignored_pairs}"
        else:
            network_pairs = ""
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in vm_storage])

        # Create plan command with pod network mapping
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ova_provider}",
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
