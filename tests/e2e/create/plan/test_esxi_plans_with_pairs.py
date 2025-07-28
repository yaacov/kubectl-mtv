"""
Test cases for kubectl-mtv migration plan creation from ESXi providers using mapping pairs.

This test validates the creation of migration plans using ESXi as the source provider
with inline network and storage mapping pairs instead of pre-created mappings.
ESXi is essentially vSphere with sdk-endpoint set to esxi.
"""

import time

import pytest

from e2e.utils import wait_for_provider_ready, wait_for_plan_ready


# Hardcoded VM names from ESXi inventory data
ESXI_TEST_VMS = [
    "mtv-win2019-79-nfs-4-18",
    "mtv-win2019-79-nfs-4-17",
    "mtv-win2019-79-nfs-4-16",
    "mtv-win2019-79-ceph-rbd",
    "mtv-rhel8-warm-sanity-nfs-4-19",
]

# Hardcoded network mapping pairs from ESXi inventory data
ESXI_NETWORK_PAIRS = [
    {"source": "Mgmt Network", "target": "test-nad-1"},
    {"source": "VM Network", "target": "test-nad-2"},
]

# Hardcoded storage mapping pairs from ESXi inventory data
ESXI_STORAGE_PAIRS = [
    {
        "source": "mtv-nfs-rhos-v8",
        "target": "ocs-storagecluster-ceph-rbd-virtualization",
    },
    {"source": "nfs-us", "target": "ocs-storagecluster-ceph-rbd"},
    {"source": "mtv-nfs-us-v8", "target": "csi-manila-ceph"},
    {"source": "nfs-us-mtv-v8", "target": "ocs-storagecluster-ceph-rbd-virtualization"},
    {"source": "datastore1", "target": "ocs-storagecluster-ceph-rbd"},
]


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

        provider_name = "test-esxi-plan-pairs-skip-verify"

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

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Wait for provider to be ready
        wait_for_provider_ready(test_namespace, provider_name)

        return provider_name

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
            "--target test-openshift-target",
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
            "--target test-openshift-target",
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
        selected_vm = ESXI_TEST_VMS[1]
        plan_name = f"test-plan-esxi-pod-pairs-{int(time.time())}"

        # Use pod network for all networks
        network_pairs = ",".join([f"{n['source']}:pod" for n in ESXI_NETWORK_PAIRS])
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in ESXI_STORAGE_PAIRS]
        )

        # Create plan command with pod network mapping
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {esxi_provider}",
            "--target test-openshift-target",
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
        selected_vm = ESXI_TEST_VMS[2]
        plan_name = f"test-plan-esxi-vddk-pairs-{int(time.time())}"

        # Use simple mapping pairs
        network_pairs = "VM Network:pod"
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
            "--target test-openshift-target",
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
