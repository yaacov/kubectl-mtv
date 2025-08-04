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
NETWORK_ATTACHMENT_DEFINITIONS = ["test-nad-1", "test-nad-2"]

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

# VMware vSphere VM names
VSPHERE_TEST_VMS = [
    "mtv-win2019-79-ceph-rbd-4-16",  # Windows VM
    "mtv-rhel8-ameen",  # Linux VM (found in inventory - different from mtv-func-rhel8-ameen)
]

# VMware ESXi VM names (same as vSphere)
ESXI_TEST_VMS = [
    "mtv-win2019-79-ceph-rbd-4-16",  # Windows VM
    "mtv-rhel8-ameen",  # Linux VM (found in inventory - different from mtv-func-rhel8-ameen)
]

# Red Hat Virtualization (oVirt) VM names
OVIRT_TEST_VMS = [
    "vCenter8-02",  # Linux VM (found in inventory)
    "mtv-win2019-79-ceph-rbd-4-16",  # Windows VM (reusing available one)
]

# OVA VM names
OVA_TEST_VMS = [
    "mtv-2disks",  # Multi-disk VM example
    "1nisim-rhel9-efi",  # Single disk RHEL VM
]

# OpenStack VM names
OPENSTACK_TEST_VMS = ["infra-mtv-node-225", "qemtv-05-mlfp6-worker-0-vfww8"]

# =============================================================================
# NETWORK MAPPINGS BY PROVIDER TYPE
# =============================================================================

# Common network mappings (most providers use these)
COMMON_NETWORK_PAIRS = [
    {"source": "VM Network", "target": NETWORK_ATTACHMENT_DEFINITIONS[0]},  # test-nad-1
    {
        "source": "Mgmt Network",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[1],
    },  # test-nad-2
]

# vSphere network mappings
VSPHERE_NETWORK_PAIRS = COMMON_NETWORK_PAIRS.copy()

# ESXi network mappings (same as vSphere)
ESXI_NETWORK_PAIRS = COMMON_NETWORK_PAIRS.copy()

# OVA network mappings (same as vSphere but reversed order)
OVA_NETWORK_PAIRS = [
    {"source": "VM Network", "target": NETWORK_ATTACHMENT_DEFINITIONS[0]},  # test-nad-1
    {
        "source": "Mgmt Network",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[1],
    },  # test-nad-2
]

# oVirt network mappings
OVIRT_NETWORK_PAIRS = [
    {"source": "ovirtmgmt", "target": NETWORK_ATTACHMENT_DEFINITIONS[0]},  # test-nad-1
    {"source": "vm", "target": NETWORK_ATTACHMENT_DEFINITIONS[1]},  # test-nad-2
]

# OpenStack network mappings
OPENSTACK_NETWORK_PAIRS = [
    {
        "source": "provider_net_cci_13",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[0],
    },  # test-nad-1
    {
        "source": "provider_net_shared_2",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[1],
    },  # test-nad-2
]

# OpenShift network mappings (for OpenShift-to-OpenShift migration)
OPENSHIFT_NETWORK_PAIRS = [
    {
        "source": NETWORK_ATTACHMENT_DEFINITIONS[0],
        "target": NETWORK_ATTACHMENT_DEFINITIONS[1],
    },  # test-nad-1 -> test-nad-2
    {
        "source": NETWORK_ATTACHMENT_DEFINITIONS[1],
        "target": NETWORK_ATTACHMENT_DEFINITIONS[0],
    },  # test-nad-2 -> test-nad-1
]

# =============================================================================
# STORAGE MAPPINGS BY PROVIDER TYPE
# =============================================================================

# vSphere storage mappings
VSPHERE_STORAGE_PAIRS = [
    {
        "source": "nfs-us-mtv-v8",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {"source": "datastore1", "target": TARGET_STORAGE_CLASSES["CEPH_RBD"]},
    {"source": "mtv-nfs-us-v8", "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"]},
    {"source": "mtv-nfs-rhos-v8", "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"]},
]

# ESXi storage mappings
ESXI_STORAGE_PAIRS = [
    {
        "source": "nfs-us-mtv-v8",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {"source": "datastore1", "target": TARGET_STORAGE_CLASSES["CEPH_RBD"]},
    {"source": "mtv-nfs-rhos-v8", "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"]},
    {"source": "mtv-nfs-us-v8", "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"]},
]

# oVirt storage mappings
OVIRT_STORAGE_PAIRS = [
    {
        "source": "L1_Group_4_Storage",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {"source": "L0_Group_4_LUN1", "target": TARGET_STORAGE_CLASSES["CEPH_RBD"]},
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
        "source": "provider_net_cci_13",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[0],
    },  # test-nad-1
    {
        "source": "provider_net_shared_2",
        "target": NETWORK_ATTACHMENT_DEFINITIONS[1],
    },  # test-nad-2
]

OPENSHIFT_NETWORKS = [
    {
        "source": NETWORK_ATTACHMENT_DEFINITIONS[0],
        "target": NETWORK_ATTACHMENT_DEFINITIONS[1],
    },  # test-nad-1 -> test-nad-2
    {
        "source": NETWORK_ATTACHMENT_DEFINITIONS[1],
        "target": NETWORK_ATTACHMENT_DEFINITIONS[0],
    },  # test-nad-2 -> test-nad-1
]

# Storage mappings for mapping creation tests
VSPHERE_DATASTORES = [
    {
        "source": "nfs-us-mtv-v8",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
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
        "source": "nfs-us-mtv-v8",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {"source": "datastore1", "target": TARGET_STORAGE_CLASSES["CEPH_RBD"]},
    {
        "source": "mtv-nfs-rhos-v8",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {"source": "nfs-us", "target": TARGET_STORAGE_CLASSES["CEPH_RBD"]},
    {
        "source": "mtv-nfs-us-v8",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
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
        "source": "L1_Group_4_Storage",
        "target": TARGET_STORAGE_CLASSES["CEPH_RBD_VIRTUALIZATION"],
    },
    {"source": "L0_Group_4_LUN1", "target": TARGET_STORAGE_CLASSES["CEPH_RBD"]},
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
