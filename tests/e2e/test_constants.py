"""
Centralized constants for kubectl-mtv e2e tests.

This module contains all hardcoded resource names, VM names, and mappings
to minimize duplication and make it easier to maintain test resources.
"""

# =============================================================================
# TARGET RESOURCES (created during test setup)
# =============================================================================

# Target provider name for OpenShift
TARGET_PROVIDER_NAME = "test-openshift-skip-verify"

# Network Attachment Definitions created in test namespace
# Extended to support unique mappings per provider constraint
NETWORK_ATTACHMENT_DEFINITIONS = [
    "test-nad-1",
    "test-nad-2",
    "test-nad-3",
    "test-nad-4",
    "test-nad-5",
]

# Target storage classes available in OpenShift
TARGET_STORAGE_CLASSES = {
    "CEPH_RBD_VIRTUALIZATION": "ocs-storagecluster-ceph-rbd-virtualization",
    "CEPH_RBD": "ocs-storagecluster-ceph-rbd",
    "MANILA_CEPH": "csi-manila-ceph",
}

# Test VMs created in OpenShift namespace
OPENSHIFT_TEST_VMS = ["test-vm-1", "test-vm-2"]

# =============================================================================
# SOURCE PROVIDER VM NAMES
# =============================================================================

# VMware vSphere VM names (validated against MCP inventory)
VSPHERE_TEST_VMS = [
    "mtv-func-win2019-test",  # Windows VM (actual name from inventory)
    "mtv-rhel8-sanity",  # Linux VM (available and compliant)
]

# VMware ESXi VM names (validated against MCP inventory)
ESXI_TEST_VMS = [
    "pabreu-rhel9-vm",  # Only available VM in ESXi provider (RHEL 9 VM)
]

# Red Hat Virtualization (oVirt) VM names (actual VMs from inventory - simple VMs with 1 network, 1 disk)
OVIRT_TEST_VMS = [
    "1111ab",  # Simple Linux VM (1 NIC on ovirtmgmt, 1 disk)
    "1111-win2019",  # Simple Windows VM (1 NIC on vm network, 1 disk)
]

# OVA VM names
OVA_TEST_VMS = [
    "mtv-2disks",  # Multi-disk VM example
    "1nisim-rhel9-efi",  # Single disk RHEL VM
]

# OpenStack VM names (validated against MCP inventory) - chose simple VMs with minimal resources
OPENSTACK_TEST_VMS = [
    "infra-mtv-node-331",
    "mtv-test",
]  # infra-mtv-node-331 has only 1 network, mtv-test is a simple test VM

# =============================================================================
# HOST CONFIGURATION BY PROVIDER TYPE
# =============================================================================

# vSphere host IDs (from vSphere inventory data)
VSPHERE_TEST_HOSTS = ["host-2007"]  # host-2007 has 'green' status, host-8 has 'yellow' status

# ESXi host IDs (from ESXi inventory data)
ESXI_TEST_HOSTS = ["ha-host"]

# Network adapter names (used by both vSphere and ESXi providers)
NETWORK_ADAPTERS = ["Management Network", "Mgmt Network", "VM Network"]

# =============================================================================
# NETWORK MAPPINGS BY PROVIDER TYPE
# =============================================================================

# vSphere network mappings (validated against MCP inventory - complies with constraints)
# Maps both available networks: VM Network -> pod, Mgmt Network -> multus
VSPHERE_NETWORK_PAIRS = [
    {
        "source": "VM Network",
        "target": "default",
    },  # Pod network (network-17 from inventory)
    {
        "source": "Mgmt Network",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[0],
    },  # test-nad-1 (network-16 from inventory)
]

# ESXi network mappings (validated against MCP inventory)
# Only "VM Network" available in ESXi provider (HaNetwork-VM Network)
ESXI_NETWORK_PAIRS = [
    {
        "source": "VM Network",
        "target": "default",
    },  # Pod network (only available network)
]

# OVA network mappings (complies with pod network uniqueness constraint)
# First network -> pod, others -> multus NADs
OVA_NETWORK_PAIRS = [
    {
        "source": "VM Network",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[1],
    },  # test-nad-2 (avoid reusing test-nad-1)
    {
        "source": "Mgmt Network",
        "target": "default",
    },  # Pod network (different from vSphere to avoid conflicts)
]

# oVirt network mappings (complies with constraints - uses unique targets)
OVIRT_NETWORK_PAIRS = [
    {
        "source": "ovirtmgmt",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[2],
    },  # test-nad-3 (unique target)
    {
        "source": "vm",
        "target": "default",
    },  # Pod network (reused but in different provider context)
]

# OpenStack network mappings (reordered to match VM order - first pair must match first VM's network)
OPENSTACK_NETWORK_PAIRS = [
    {
        "source": "provider_net_shared",
        "target": "default",
    },  # Network from infra-mtv-node-331 (OPENSTACK_TEST_VMS[0]) -> pod network
    {
        "source": "provider_net_shared_2",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[3],
    },  # Network for mtv-test (OPENSTACK_TEST_VMS[1]) -> test-nad-4
]

# OpenShift network mappings (based on actual VM networks - maps NADs that exist on the test VMs)
# Note: Namespace will be prepended dynamically by tests using test_namespace.namespace
OPENSHIFT_NETWORK_PAIRS = [
    {
        "source": NETWORK_ATTACHMENT_DEFINITIONS[0],
        "target": "default",
    },  # test-nad-1 -> pod network
    {
        "source": NETWORK_ATTACHMENT_DEFINITIONS[1],
        "target": NETWORK_ATTACHMENT_DEFINITIONS[4],
    },  # test-nad-2 -> test-nad-5
]

# =============================================================================
# STORAGE MAPPINGS BY PROVIDER TYPE
# =============================================================================

# vSphere storage mappings
VSPHERE_STORAGE_PAIRS = [
    {"source": "datastore1", "target": TARGET_STORAGE_CLASSES["CEPH_RBD"]},
    {
        "source": "mtv-nfs-us-v8",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {
        "source": "mtv-nfs-rhos-v8",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
]

# ESXi storage mappings (validated against MCP inventory)
# Only "datastore1" available in ESXi provider inventory
ESXI_STORAGE_PAIRS = [
    {
        "source": "datastore1",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD"],
    },  # Only available datastore in ESXi
]

# oVirt storage mappings (using actual storage domains from inventory)
OVIRT_STORAGE_PAIRS = [
    {
        "source": "L0_Group_4_LUN1",  # Actual data storage domain from inventory
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {
        "source": "L0_Group_4_LUN2",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD"],
    },  # Second data storage domain
]

# OpenStack storage mappings
OPENSTACK_STORAGE_PAIRS = [
    {
        "source": "__DEFAULT__",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {"source": "tripleo", "target": TARGET_STORAGE_CLASSES["CEPH_RBD"]},
]

# OpenShift storage mappings (for OpenShift-to-OpenShift migration)
OPENSHIFT_STORAGE_PAIRS = [
    {
        "source": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {
        "source": TARGET_STORAGE_CLASSES["CEPH_RBD"],
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD"],
    },
]

# =============================================================================
# OVA-SPECIFIC VM MAPPINGS (for VMs with multiple disks/networks)
# =============================================================================

# VM-specific storage mappings for OVA plans with pairs
OVA_VM_STORAGE_MAPPINGS = {
    "mtv-2disks": [
        {
            "source": "mtv-2disks-1.vmdk",
            "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
        },
        {"source": "mtv-2disks-2.vmdk", "target": TARGET_STORAGE_CLASSES["CEPH_RBD"]},
    ],
    "1nisim-rhel9-efi": [
        {
            "source": "1nisim-rhel9-efi-1.vmdk",
            "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
        },
    ],
}

# VM-specific network mappings for OVA plans with pairs
OVA_VM_NETWORK_MAPPINGS = {
    "mtv-2disks": [
        {
            "source": "VM Network",
            "target": NETWORK_ATTACHMENT_DEFINITIONS[0],
        },  # test-nad-1
        {
            "source": "Mgmt Network",
            "target": NETWORK_ATTACHMENT_DEFINITIONS[1],
        },  # test-nad-2
    ],
    "1nisim-rhel9-efi": [
        {
            "source": "Mgmt Network",
            "target": NETWORK_ATTACHMENT_DEFINITIONS[1],
        },  # test-nad-2
    ],
}

# =============================================================================
# MAPPING TEST CONSTANTS
# =============================================================================

# Network mappings for mapping creation tests (simplified versions of above)
VSPHERE_NETWORKS = [
    {
        "source": "Mgmt Network",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[0],
    },  # test-nad-1
    {"source": "VM Network", "target": NETWORK_ATTACHMENT_DEFINITIONS[1]},  # test-nad-2
]

ESXI_NETWORKS = VSPHERE_NETWORKS.copy()

OVA_NETWORKS = [
    {"source": "VM Network", "target": NETWORK_ATTACHMENT_DEFINITIONS[0]},  # test-nad-1
    {
        "source": "Mgmt Network",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[1],
    },  # test-nad-2
]

OVIRT_NETWORKS = [
    {"source": "ovirtmgmt", "target": NETWORK_ATTACHMENT_DEFINITIONS[0]},  # test-nad-1
    {"source": "vm", "target": NETWORK_ATTACHMENT_DEFINITIONS[1]},  # test-nad-2
]

OPENSTACK_NETWORKS = [
    {
        "source": "provider_net_shared",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[0],
    },  # Network from infra-mtv-node-331 (matches VM order) -> test-nad-1
    {
        "source": "provider_net_shared_2",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[1],
    },  # Network for mtv-test -> test-nad-2
]

OPENSHIFT_NETWORKS = [
    {
        "source": NETWORK_ATTACHMENT_DEFINITIONS[0],
        "target": NETWORK_ATTACHMENT_DEFINITIONS[1],
    },  # test-nad-1 -> test-nad-2 (networks that exist on test VMs)
    {
        "source": NETWORK_ATTACHMENT_DEFINITIONS[1],
        "target": NETWORK_ATTACHMENT_DEFINITIONS[0],
    },  # test-nad-2 -> test-nad-1 (networks that exist on test VMs)
]

# Storage mappings for mapping creation tests
VSPHERE_DATASTORES = [
    {"source": "datastore1", "target": TARGET_STORAGE_CLASSES["CEPH_RBD"]},
    {
        "source": "mtv-nfs-us-v8",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {
        "source": "mtv-nfs-rhos-v8",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
]

ESXI_DATASTORES = [
    {
        "source": "datastore1",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD"],
    },  # Only available datastore in ESXi (validated against MCP inventory)
]

OVA_STORAGE = [
    {
        "source": "1nisim-rhel9-efi-1.vmdk",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {"source": "mtv-2disks-1.vmdk", "target": TARGET_STORAGE_CLASSES["CEPH_RBD"]},
]

OVIRT_DATASTORES = [
    {
        "source": "L0_Group_4_LUN1",  # Actual data storage domain from inventory
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {
        "source": "L0_Group_4_LUN2",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD"],
    },  # Second data storage domain
]

OPENSTACK_DATASTORES = [
    {
        "source": "__DEFAULT__",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {"source": "tripleo", "target": TARGET_STORAGE_CLASSES["CEPH_RBD"]},
]

OPENSHIFT_DATASTORES = [
    {
        "source": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {
        "source": TARGET_STORAGE_CLASSES["CEPH_RBD"],
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD"],
    },
]
