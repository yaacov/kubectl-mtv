"""
Test cases for kubectl-mtv network and storage mapping creation from OVA providers.

This test validates the creation of network and storage mappings using OVA as the source provider
and OpenShift as the target provider.
"""

import time

import pytest

from e2e.utils import (
    wait_for_provider_ready,
    wait_for_network_mapping_ready,
    wait_for_storage_mapping_ready,
)


# Hardcoded network names from OVA inventory data
OVA_NETWORKS = [
    {"source": "VM Network", "target": "test-nad-1"},
    {"source": "Mgmt Network", "target": "test-nad-2"},
]

# Hardcoded storage names from OVA inventory data - using unique VMDK names
OVA_STORAGE = [
    {
        "source": "1nisim-rhel9-efi-1.vmdk",
        "target": "ocs-storagecluster-ceph-rbd-virtualization",
    },
    {"source": "mtv-func-WIN2019-1.vmdk", "target": "ocs-storagecluster-ceph-rbd"},
    {"source": "SHAICTDOET005-Test_rhel9-1.vmdk", "target": "csi-manila-ceph"},
]


@pytest.mark.create
@pytest.mark.mapping
@pytest.mark.ova
@pytest.mark.requires_credentials
class TestOVAMappingCreation:
    """Test cases for network and storage mapping creation from OVA providers."""

    @pytest.fixture(scope="class")
    def ova_provider(self, test_namespace, provider_credentials):
        """Create an OVA provider for mapping testing."""
        creds = provider_credentials.get("ova", {})

        # Skip if OVA URL is not available
        if not creds.get("url"):
            pytest.skip("OVA URL not available in environment")

        provider_name = "test-ova-map-skip-verify"

        # Create command for OVA provider with URL
        create_cmd = (
            f"create provider {provider_name} --type ova --url '{creds['url']}'"
        )

        # Create provider
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Wait for provider to be ready
        wait_for_provider_ready(test_namespace, provider_name)

        return provider_name

    def test_create_network_mapping_from_ova(self, test_namespace, ova_provider):
        """Test creating a network mapping from OVA provider."""
        mapping_name = f"test-network-map-ova-{int(time.time())}"

        # Build network pairs string
        network_pairs = ",".join([f"{n['source']}:{n['target']}" for n in OVA_NETWORKS])

        # Create network mapping command
        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {ova_provider}",
            "--target test-openshift-target",
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

    def test_create_minimal_network_mapping_from_ova(
        self, test_namespace, ova_provider
    ):
        """Test creating a minimal network mapping from OVA provider with a single network."""
        mapping_name = f"test-minimal-network-map-ova-{int(time.time())}"

        # Create network mapping command with single network
        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {ova_provider}",
            "--target test-openshift-target",
            "--network-pairs 'VM Network:default'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create network mapping
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("networkmap", mapping_name)

        # Wait for network mapping to be ready
        wait_for_network_mapping_ready(test_namespace, mapping_name)

        # Verify VM Network is in the mapping
        result = test_namespace.run_kubectl_command(
            f"get networkmap {mapping_name} -o yaml"
        )
        assert result.returncode == 0
        assert "VM Network" in result.stdout

    def test_create_storage_mapping_from_ova(self, test_namespace, ova_provider):
        """Test creating a storage mapping from OVA provider."""
        mapping_name = f"test-storage-map-ova-{int(time.time())}"

        # Build storage pairs string
        storage_pairs = ",".join([f"{s['source']}:{s['target']}" for s in OVA_STORAGE])

        # Create storage mapping command
        cmd_parts = [
            "create mapping storage",
            mapping_name,
            f"--source {ova_provider}",
            "--target test-openshift-target",
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
