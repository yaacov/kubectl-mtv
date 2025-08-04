"""
Test cases for kubectl-mtv network and storage mapping creation for OpenShift-to-OpenShift migrations.

This test validates the creation of network and storage mappings using OpenShift as both source and target provider.
"""

import time

import pytest

from e2e.utils import (
    wait_for_network_mapping_ready,
    wait_for_storage_mapping_ready,
    generate_provider_name,
    get_or_create_provider,
)
from e2e.test_constants import TARGET_PROVIDER_NAME


# Hardcoded network names for OpenShift to OpenShift mappings
OPENSHIFT_NETWORKS = [
    {"source": "test-nad-1", "target": "test-nad-2"},
    {"source": "test-nad-2", "target": "test-nad-1"},
]

# Hardcoded storage names for OpenShift to OpenShift mappings
# Using storage classes that should exist in the cluster
OPENSHIFT_STORAGE_CLASSES = [
    {
        "source": "ocs-storagecluster-ceph-rbd-virtualization",
        "target": "ocs-storagecluster-ceph-rbd-virtualization",
    },
    {"source": "ocs-storagecluster-ceph-rbd", "target": "ocs-storagecluster-ceph-rbd"},
]


@pytest.mark.create
@pytest.mark.mapping
@pytest.mark.openshift
class TestOpenShiftMappingCreation:
    """Test cases for network and storage mapping creation for OpenShift-to-OpenShift migrations."""

    @pytest.fixture(scope="class")
    def openshift_source_provider(self, test_namespace):
        """Create an OpenShift source provider for mapping testing."""
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

    def test_create_network_mapping_openshift_to_openshift(
        self, test_namespace, openshift_source_provider
    ):
        """Test creating a network mapping for OpenShift-to-OpenShift migration."""
        mapping_name = f"test-network-map-openshift-{int(time.time())}"

        # Build network pairs string
        network_pairs = ",".join(
            [f"{n['source']}:{n['target']}" for n in OPENSHIFT_NETWORKS]
        )

        # Create network mapping command
        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {openshift_source_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--network-pairs '{network_pairs}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create network mapping
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("networkmap", mapping_name)

        # Wait for network mapping to be ready
        wait_for_network_mapping_ready(test_namespace, mapping_name)

    def test_create_storage_mapping_openshift_to_openshift(
        self, test_namespace, openshift_source_provider
    ):
        """Test creating a storage mapping for OpenShift-to-OpenShift migration."""
        mapping_name = f"test-storage-map-openshift-{int(time.time())}"

        # Build storage pairs string
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OPENSHIFT_STORAGE_CLASSES]
        )

        # Create storage mapping command
        cmd_parts = [
            "create mapping storage",
            mapping_name,
            f"--source {openshift_source_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--storage-pairs '{storage_pairs}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create storage mapping
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("storagemap", mapping_name)

        # Wait for storage mapping to be ready
        wait_for_storage_mapping_ready(test_namespace, mapping_name)

    def test_create_mapping_with_namespace_qualifier(
        self, test_namespace, openshift_source_provider
    ):
        """Test creating a network mapping with namespace-qualified NADs."""
        mapping_name = f"test-ns-qualified-network-map-{int(time.time())}"

        # Create network mapping with namespace-qualified target NAD only
        # Source NADs in OpenShift don't use namespace qualification in inventory
        network_pairs = f"test-nad-1:{test_namespace.namespace}/test-nad-2"

        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {openshift_source_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--network-pairs '{network_pairs}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create network mapping
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("networkmap", mapping_name)

        # Wait for network mapping to be ready
        wait_for_network_mapping_ready(test_namespace, mapping_name)

        # Verify namespace is in the mapping
        result = test_namespace.run_kubectl_command(
            f"get networkmap {mapping_name} -o yaml"
        )
        assert result.returncode == 0
        assert test_namespace.namespace in result.stdout
