"""
Test cases for kubectl-mtv migration plan creation from OpenShift providers using mapping pairs.

This test validates the creation of migration plans using OpenShift as the source provider
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
    OPENSHIFT_TEST_VMS,
    OPENSHIFT_NETWORK_PAIRS,
    OPENSHIFT_STORAGE_PAIRS,
    TARGET_PROVIDER_NAME,
)


@pytest.mark.create
@pytest.mark.plan
@pytest.mark.openshift
@pytest.mark.skip(reason="Skipping OpenShift to OpenShift migration plan tests")
class TestOpenShiftPlanCreationWithPairs:
    """Test cases for migration plan creation from OpenShift providers using mapping pairs."""

    @pytest.fixture(scope="class")
    def openshift_provider(self, test_namespace):
        """Create an OpenShift provider for plan testing."""
        # Generate provider name based on type and configuration
        provider_name = generate_provider_name("openshift", "localhost", skip_tls=True)

        # Create command for OpenShift provider
        cmd_parts = [
            "create provider",
            provider_name,
            "--type openshift",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create provider if it doesn't already exist
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    def test_create_plan_with_mapping_pairs(self, test_namespace, openshift_provider):
        """Test creating a migration plan with inline mapping pairs."""
        # Use the first available VM
        selected_vm = OPENSHIFT_TEST_VMS[0]
        plan_name = f"test-plan-openshift-pairs-{int(time.time())}"

        # Build network and storage pairs strings
        network_pairs = ",".join(
            [f"{n['source']}:{n['target']}" for n in OPENSHIFT_NETWORK_PAIRS]
        )
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OPENSHIFT_STORAGE_PAIRS]
        )

        # Create plan command with mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {openshift_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--vms '{selected_vm}'",
            f"--network-pairs '{network_pairs}'",
            f"--storage-pairs '{storage_pairs}'",
            "--target-namespace default",
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
        self, test_namespace, openshift_provider
    ):
        """Test creating a migration plan with multiple VMs using inline mapping pairs."""
        # Use all available VMs
        selected_vms = ",".join(OPENSHIFT_TEST_VMS)
        plan_name = f"test-multi-plan-openshift-pairs-{int(time.time())}"

        # Build network and storage pairs strings
        network_pairs = ",".join(
            [f"{n['source']}:{n['target']}" for n in OPENSHIFT_NETWORK_PAIRS]
        )
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OPENSHIFT_STORAGE_PAIRS]
        )

        # Create plan command with multiple VMs and mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {openshift_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--vms '{selected_vms}'",
            f"--network-pairs '{network_pairs}'",
            f"--storage-pairs '{storage_pairs}'",
            "--target-namespace default",
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

    def test_create_plan_with_pod_network_pairs(
        self, test_namespace, openshift_provider
    ):
        """Test creating a migration plan with pod network mapping pairs."""
        # Use a single VM
        selected_vm = OPENSHIFT_TEST_VMS[0]
        plan_name = f"test-plan-openshift-pod-pairs-{int(time.time())}"

        # Use pod network for all networks
        network_pairs = ",".join(
            [f"{n['source']}:default" for n in OPENSHIFT_NETWORK_PAIRS]
        )
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OPENSHIFT_STORAGE_PAIRS]
        )

        # Create plan command with pod network mapping
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {openshift_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--vms '{selected_vm}'",
            f"--network-pairs '{network_pairs}'",
            f"--storage-pairs '{storage_pairs}'",
            "--target-namespace default",
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

    def test_create_plan_with_namespace_qualified_pairs(
        self, test_namespace, openshift_provider
    ):
        """Test creating a migration plan with namespace-qualified network mapping pairs."""
        # Use a single VM
        selected_vm = OPENSHIFT_TEST_VMS[1]
        plan_name = f"test-plan-openshift-ns-pairs-{int(time.time())}"

        # Use namespace-qualified network targets
        network_pairs = f"test-nad-1:{test_namespace.namespace}/test-nad-2,test-nad-2:{test_namespace.namespace}/test-nad-1"
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OPENSHIFT_STORAGE_PAIRS]
        )

        # Create plan command with namespace-qualified network mapping
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {openshift_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--vms '{selected_vm}'",
            f"--network-pairs '{network_pairs}'",
            f"--storage-pairs '{storage_pairs}'",
            "--target-namespace default",
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
