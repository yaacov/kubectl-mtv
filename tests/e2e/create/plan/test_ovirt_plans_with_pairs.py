"""
Test cases for kubectl-mtv migration plan creation from oVirt providers using mapping pairs.

This test validates the creation of migration plans using oVirt as the source provider
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
    OVIRT_TEST_VMS,
    OVIRT_NETWORK_PAIRS,
    OVIRT_STORAGE_PAIRS,
    TARGET_PROVIDER_NAME,
)


@pytest.mark.create
@pytest.mark.plan
@pytest.mark.ovirt
@pytest.mark.requires_credentials
class TestOvirtPlanCreationWithPairs:
    """Test cases for migration plan creation from oVirt providers using mapping pairs."""

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

    def test_create_plan_with_mapping_pairs(self, test_namespace, ovirt_provider):
        """Test creating a migration plan with inline mapping pairs."""
        # Use the first available VM
        selected_vm = OVIRT_TEST_VMS[0]
        plan_name = f"test-plan-ovirt-pairs-{int(time.time())}"

        # Build network and storage pairs strings (use first 2 networks and 3 storage domains)
        network_pairs = ",".join(
            [f"{n['source']}:{n['target']}" for n in OVIRT_NETWORK_PAIRS[:2]]
        )
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OVIRT_STORAGE_PAIRS[:3]]
        )

        # Create plan command with mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ovirt_provider}",
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
        self, test_namespace, ovirt_provider
    ):
        """Test creating a migration plan with multiple VMs using inline mapping pairs."""
        # Use first 3 VMs for multi-VM test
        selected_vms = ",".join(OVIRT_TEST_VMS[:3])
        plan_name = f"test-multi-plan-ovirt-pairs-{int(time.time())}"

        # Build network and storage pairs strings (use all mappings)
        network_pairs = ",".join(
            [f"{n['source']}:{n['target']}" for n in OVIRT_NETWORK_PAIRS]
        )
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OVIRT_STORAGE_PAIRS]
        )

        # Create plan command with multiple VMs and mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ovirt_provider}",
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

    def test_create_plan_with_pod_network_pairs(self, test_namespace, ovirt_provider):
        """Test creating a migration plan with pod network mapping pairs."""
        # Use a single VM
        selected_vm = OVIRT_TEST_VMS[1]
        plan_name = f"test-plan-ovirt-pod-pairs-{int(time.time())}"

        # Use pod network for all networks
        # Use pod network for first network only, ignore the rest (complies with pod network uniqueness constraint)
        network_pairs = f"{OVIRT_NETWORK_PAIRS[0]['source']}:default"
        if len(OVIRT_NETWORK_PAIRS) > 1:
            # Map additional networks to ignored to comply with constraint requirements
            ignored_pairs = ",".join(
                [f"{n['source']}:ignored" for n in OVIRT_NETWORK_PAIRS[1:]]
            )
            network_pairs = f"{network_pairs},{ignored_pairs}"
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OVIRT_STORAGE_PAIRS[:3]]
        )

        # Create plan command with pod network mapping
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ovirt_provider}",
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
