"""
Test cases for kubectl-mtv migration plan creation from oVirt providers using mapping pairs.

This test validates the creation of migration plans using oVirt as the source provider
with inline network and storage mapping pairs instead of pre-created mappings.
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

# Hardcoded network mapping pairs from oVirt inventory data
OVIRT_NETWORK_PAIRS = [
    {"source": "ovirtmgmt", "target": "test-nad-1"},
    {"source": "vm", "target": "test-nad-2"},
    {"source": "internal", "target": "test-nad-1"},
    {"source": "vlan10", "target": "test-nad-2"}
]

# Hardcoded storage mapping pairs from oVirt inventory data  
OVIRT_STORAGE_PAIRS = [
    {"source": "hosted_storage", "target": "ocs-storagecluster-ceph-rbd-virtualization"},
    {"source": "L0_Group_4_LUN1", "target": "ocs-storagecluster-ceph-rbd"},
    {"source": "L0_Group_4_LUN2", "target": "csi-manila-ceph"},
    {"source": "L0_Group_4_LUN3", "target": "csi-manila-netapp"},
    {"source": "export2", "target": "ocs-storagecluster-ceph-rbd"}
]


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

        provider_name = "test-ovirt-plan-pairs-skip-verify"

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

    def test_create_plan_with_mapping_pairs(self, test_namespace, ovirt_provider):
        """Test creating a migration plan with inline mapping pairs."""
        # Use the first available VM
        selected_vm = OVIRT_TEST_VMS[0]
        plan_name = f"test-plan-ovirt-pairs-{int(time.time())}"
        
        # Build network and storage pairs strings (use first 2 networks and 3 storage domains)
        network_pairs = ",".join([f"{n['source']}:{n['target']}" for n in OVIRT_NETWORK_PAIRS[:2]])
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in OVIRT_STORAGE_PAIRS[:3]])
        
        # Create plan command with mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ovirt_provider}",
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

    def test_create_multi_vm_plan_with_mapping_pairs(self, test_namespace, ovirt_provider):
        """Test creating a migration plan with multiple VMs using inline mapping pairs."""
        # Use first 3 VMs for multi-VM test
        selected_vms = ",".join(OVIRT_TEST_VMS[:3])
        plan_name = f"test-multi-plan-ovirt-pairs-{int(time.time())}"
        
        # Build network and storage pairs strings (use all mappings)
        network_pairs = ",".join([f"{n['source']}:{n['target']}" for n in OVIRT_NETWORK_PAIRS])
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in OVIRT_STORAGE_PAIRS])
        
        # Create plan command with multiple VMs and mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ovirt_provider}",
            "--target test-openshift-target",
            f"--vms {selected_vms}",
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
        network_pairs = ",".join([f"{n['source']}:pod" for n in OVIRT_NETWORK_PAIRS[:2]])
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in OVIRT_STORAGE_PAIRS[:3]])
        
        # Create plan command with pod network mapping
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ovirt_provider}",
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