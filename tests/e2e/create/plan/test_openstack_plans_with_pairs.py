"""
Test cases for kubectl-mtv migration plan creation from OpenStack providers using mapping pairs.

This test validates the creation of migration plans using OpenStack as the source provider
with inline network and storage mapping pairs instead of pre-created mappings.
"""

import time

import pytest

from e2e.utils import wait_for_provider_ready, wait_for_plan_ready


OPENSTACK_TEST_VMS = ["infra-mtv-node-207", "infra-mtv-node-18"]

OPENSTACK_NETWORK_PAIRS = [
    {"source": "provider_net_shared", "target": "test-nad-1"},
    {"source": "provider_net_shared_3", "target": "test-nad-2"},
]

OPENSTACK_STORAGE_PAIRS = [
    {"source": "__DEFAULT__", "target": "ocs-storagecluster-ceph-rbd-virtualization"},
    {"source": "tripleo", "target": "ocs-storagecluster-ceph-rbd"},
    {"source": "ceph", "target": "csi-manila-ceph"},
]


@pytest.mark.create
@pytest.mark.plan
@pytest.mark.openstack
@pytest.mark.requires_credentials
class TestOpenStackPlanCreationWithPairs:
    """Test cases for migration plan creation from OpenStack providers using mapping pairs."""

    @pytest.fixture(scope="class")
    def openstack_provider(self, test_namespace, provider_credentials):
        """Create an OpenStack provider for plan testing."""
        creds = provider_credentials["openstack"]

        # Skip if credentials are not available
        if not all(
            [
                creds.get("url"),
                creds.get("username"),
                creds.get("password"),
                creds.get("project_name"),
                creds.get("domain_name"),
            ]
        ):
            pytest.skip("OpenStack credentials not available in environment")

        provider_name = "test-openstack-plan-pairs-skip-verify"

        # Create command with insecure skip TLS
        cmd_parts = [
            "create provider",
            provider_name,
            "--type openstack",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--provider-domain-name '{creds['domain_name']}'",
            f"--provider-project-name '{creds['project_name']}'",
            "--provider-insecure-skip-tls",
        ]

        if creds.get("region_name"):
            cmd_parts.append(f"--provider-region-name '{creds['region_name']}'")

        create_cmd = " ".join(cmd_parts)

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Wait for provider to be ready
        wait_for_provider_ready(test_namespace, provider_name)

        return provider_name

    def test_create_plan_with_mapping_pairs(self, test_namespace, openstack_provider):
        """Test creating a migration plan with inline mapping pairs."""
        # Use the first available VM
        selected_vm = OPENSTACK_TEST_VMS[0]
        plan_name = f"test-plan-openstack-pairs-{int(time.time())}"

        # Build network and storage pairs strings
        network_pairs = ",".join(
            [f"{n['source']}:{n['target']}" for n in OPENSTACK_NETWORK_PAIRS]
        )
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OPENSTACK_STORAGE_PAIRS]
        )

        # Create plan command with mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {openstack_provider}",
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
        self, test_namespace, openstack_provider
    ):
        """Test creating a migration plan with multiple VMs using inline mapping pairs."""
        # Use both available VMs for multi-VM test
        selected_vms = ",".join(OPENSTACK_TEST_VMS)
        plan_name = f"test-multi-plan-openstack-pairs-{int(time.time())}"

        # Build network and storage pairs strings
        network_pairs = ",".join(
            [f"{n['source']}:{n['target']}" for n in OPENSTACK_NETWORK_PAIRS]
        )
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OPENSTACK_STORAGE_PAIRS]
        )

        # Create plan command with multiple VMs and mapping pairs
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {openstack_provider}",
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

    def test_create_plan_with_pod_network_pairs(
        self, test_namespace, openstack_provider
    ):
        """Test creating a migration plan with pod network mapping pairs."""
        # Use a single VM
        selected_vm = OPENSTACK_TEST_VMS[1]
        plan_name = f"test-plan-openstack-pod-pairs-{int(time.time())}"

        # Use pod network for all networks
        network_pairs = ",".join(
            [f"{n['source']}:default" for n in OPENSTACK_NETWORK_PAIRS]
        )
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OPENSTACK_STORAGE_PAIRS]
        )

        # Create plan command with pod network mapping
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {openstack_provider}",
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
