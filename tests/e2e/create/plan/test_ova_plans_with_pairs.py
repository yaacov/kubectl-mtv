"""
Test cases for kubectl-mtv migration plan creation from OVA providers using mapping pairs.

This test validates the creation of migration plans using OVA as the source provider
with inline network and storage mapping pairs instead of pre-created mappings.
"""

import time

import pytest

from e2e.utils import wait_for_provider_ready, wait_for_plan_ready


# Hardcoded VM names from OVA inventory data
OVA_TEST_VMS = [
    "mtv-2disks",
    "1nisim-rhel9-efi",
    "mtv-func-WIN2019"
]

# Hardcoded network mapping pairs from OVA inventory data
OVA_NETWORK_PAIRS = [
    {"source": "VM Network", "target": "test-nad-1"},
    {"source": "Mgmt Network", "target": "test-nad-2"}
]

# Hardcoded storage mapping pairs from OVA inventory data - covering all VM disks
OVA_STORAGE_PAIRS = [
    # mtv-2disks VMs (both instances have disks with same names but different IDs)
    {"source": "mtv-2disks-1.vmdk", "target": "ocs-storagecluster-ceph-rbd-virtualization"},
    {"source": "mtv-2disks-2.vmdk", "target": "ocs-storagecluster-ceph-rbd"},
    # 1nisim-rhel9-efi VM (1 disk)
    {"source": "1nisim-rhel9-efi-1.vmdk", "target": "ocs-storagecluster-ceph-rbd-virtualization"},
    # mtv-func-WIN2019 VM (1 disk)
    {"source": "mtv-func-WIN2019-1.vmdk", "target": "ocs-storagecluster-ceph-rbd"},
    # SHAICTDOET005-Test_rhel9 VM (3 disks)
    {"source": "SHAICTDOET005-Test_rhel9-1.vmdk", "target": "csi-manila-ceph"},
    {"source": "SHAICTDOET005-Test_rhel9-2.vmdk", "target": "csi-manila-ceph"},
    {"source": "SHAICTDOET005-Test_rhel9-3.vmdk", "target": "csi-manila-ceph"},
    # mtv-func.RHEL8_8 _TE_ST _ _1 _2 VM (2 disks)
    {"source": "mtv-func.RHEL8_8 _TE_ST _ _1 _2-1.vmdk", "target": "ocs-storagecluster-ceph-rbd"},
    {"source": "mtv-func.RHEL8_8 _TE_ST _ _1 _2-2.vmdk", "target": "ocs-storagecluster-ceph-rbd"}
]


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

        provider_name = "test-ova-plan-pairs-skip-verify"

        # Create command for OVA provider with URL
        create_cmd = f"create provider {provider_name} --type ova --url '{creds['url']}'"

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Wait for provider to be ready
        wait_for_provider_ready(test_namespace, provider_name)

        return provider_name

    def test_create_plan_with_mapping_pairs(self, test_namespace, ova_provider):
        """Test creating a migration plan with inline mapping pairs."""
        # Use the first available VM (mtv-2disks which has 2 disks)
        selected_vm = OVA_TEST_VMS[0]
        plan_name = f"test-plan-ova-pairs-{int(time.time())}"
        
        # Build network and storage pairs strings
        network_pairs = ",".join([f"{n['source']}:{n['target']}" for n in OVA_NETWORK_PAIRS])
        
        # Get storage pairs for mtv-2disks VM (first 2 pairs in the list)
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in OVA_STORAGE_PAIRS[:2]])
        
        # Create plan command with mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ova_provider}",
            "--target test-openshift-target",
            f"--vms {selected_vm}",
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

    def test_create_plan_with_pod_network_pairs(self, test_namespace, ova_provider):
        """Test creating a migration plan with pod network mapping pairs."""
        # Use the first available VM (mtv-2disks which has 2 disks)
        selected_vm = OVA_TEST_VMS[0]
        plan_name = f"test-plan-ova-pod-pairs-{int(time.time())}"
        
        # Use pod network for all networks
        network_pairs = ",".join([f"{n['source']}:pod" for n in OVA_NETWORK_PAIRS])
        
        # Get storage pairs for mtv-2disks VM (first 2 pairs in the list)
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in OVA_STORAGE_PAIRS[:2]])
        
        # Create plan command with pod network mapping
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ova_provider}",
            "--target test-openshift-target",
            f"--vms {selected_vm}",
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

    def test_create_plan_with_minimal_pairs(self, test_namespace, ova_provider):
        """Test creating a migration plan with minimal mapping pairs."""
        # Use the first available VM (mtv-2disks which has 2 disks)
        selected_vm = OVA_TEST_VMS[0]
        plan_name = f"test-plan-ova-minimal-pairs-{int(time.time())}"
        
        # Use minimal network pairs but all required storage pairs for the VM
        network_pairs = "VM Network:pod"
        
        # Get storage pairs for mtv-2disks VM (first 2 pairs in the list)
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in OVA_STORAGE_PAIRS[:2]])
        
        # Create plan command with minimal mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ova_provider}",
            "--target test-openshift-target",
            f"--vms {selected_vm}",
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