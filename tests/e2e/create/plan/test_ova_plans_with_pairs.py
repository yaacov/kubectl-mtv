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
    "mtv-func-WIN2019",
    "SHAICTDOET005-Test_rhel9",
    "mtv-func.RHEL8_8 _TE_ST _ _1 _2"
]



# VM-specific storage mappings for individual tests
VM_STORAGE_MAPPINGS = {
    "mtv-2disks": [
        {"source": "mtv-2disks-1.vmdk", "target": "ocs-storagecluster-ceph-rbd-virtualization"},
        {"source": "mtv-2disks-2.vmdk", "target": "ocs-storagecluster-ceph-rbd"}
    ],
    "1nisim-rhel9-efi": [
        {"source": "1nisim-rhel9-efi-1.vmdk", "target": "ocs-storagecluster-ceph-rbd-virtualization"}
    ],
    "mtv-func-WIN2019": [
        {"source": "mtv-func-WIN2019-1.vmdk", "target": "ocs-storagecluster-ceph-rbd"}
    ],
    "SHAICTDOET005-Test_rhel9": [
        {"source": "SHAICTDOET005-Test_rhel9-1.vmdk", "target": "csi-manila-ceph"},
        {"source": "SHAICTDOET005-Test_rhel9-2.vmdk", "target": "csi-manila-ceph"},
        {"source": "SHAICTDOET005-Test_rhel9-3.vmdk", "target": "csi-manila-ceph"}
    ],
    "mtv-func.RHEL8_8 _TE_ST _ _1 _2": [
        {"source": "mtv-func.RHEL8_8 _TE_ST _ _1 _2-1.vmdk", "target": "ocs-storagecluster-ceph-rbd"},
        {"source": "mtv-func.RHEL8_8 _TE_ST _ _1 _2-2.vmdk", "target": "ocs-storagecluster-ceph-rbd"}
    ]
}

# VM-specific network mappings for individual tests
VM_NETWORK_MAPPINGS = {
    "mtv-2disks": [
        {"source": "VM Network", "target": "test-nad-1"},
        {"source": "Mgmt Network", "target": "test-nad-2"}
    ],
    "1nisim-rhel9-efi": [
        {"source": "Mgmt Network", "target": "test-nad-2"}
    ],
    "mtv-func-WIN2019": [
        {"source": "VM Network", "target": "test-nad-1"},
        {"source": "Mgmt Network", "target": "test-nad-2"}
    ],
    "SHAICTDOET005-Test_rhel9": [
        {"source": "VM Network", "target": "test-nad-1"},
        {"source": "Mgmt Network", "target": "test-nad-2"}
    ],
    "mtv-func.RHEL8_8 _TE_ST _ _1 _2": [
        {"source": "VM Network", "target": "test-nad-1"}
    ]
}


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
        
        # Get VM-specific network and storage mappings
        vm_networks = VM_NETWORK_MAPPINGS[selected_vm]
        vm_storage = VM_STORAGE_MAPPINGS[selected_vm]
        
        # Build network and storage pairs strings
        network_pairs = ",".join([f"{n['source']}:{n['target']}" for n in vm_networks])
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in vm_storage])
        
        # Create plan command with mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ova_provider}",
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
            for net in VM_NETWORK_MAPPINGS[vm_name]:
                all_networks.add((net['source'], net['target']))
            # Collect all storage mappings
            all_storage.extend(VM_STORAGE_MAPPINGS[vm_name])
        
        network_pairs = ",".join([f"{source}:{target}" for source, target in all_networks])
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in all_storage])
        
        # Create plan command with all VMs and mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ova_provider}",
            "--target test-openshift-target",
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
        vm_networks = VM_NETWORK_MAPPINGS[selected_vm]
        vm_storage = VM_STORAGE_MAPPINGS[selected_vm]
        
        # Use pod network for all networks
        network_pairs = ",".join([f"{n['source']}:pod" for n in vm_networks])
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in vm_storage])
        
        # Create plan command with pod network mapping
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {ova_provider}",
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
