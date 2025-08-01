"""
Test cases for kubectl-mtv network and storage mapping creation from OpenStack providers.

This test validates the creation of network and storage mappings using OpenStack as the source provider
and OpenShift as the target provider.
"""

import time

import pytest

from e2e.utils import (
    wait_for_provider_ready,
    wait_for_network_mapping_ready,
    wait_for_storage_mapping_ready,
)


# Hardcoded network names from OpenStack inventory data
OPENSTACK_NETWORKS = [
    {"source": "provider_net_cci_13", "target": "test-nad-1"},
    {"source": "provider_net_shared_2", "target": "test-nad-2"},
    {"source": "provider_net_ipv6_only", "target": "test-nad-1"},
]

# Hardcoded storage names from OpenStack inventory data
OPENSTACK_VOLUME_TYPES = [
    {"source": "__DEFAULT__", "target": "ocs-storagecluster-ceph-rbd-virtualization"},
    {"source": "tripleo", "target": "ocs-storagecluster-ceph-rbd"},
    {"source": "ceph", "target": "csi-manila-ceph"},
]


@pytest.mark.create
@pytest.mark.mapping
@pytest.mark.openstack
@pytest.mark.requires_credentials
class TestOpenStackMappingCreation:
    """Test cases for network and storage mapping creation from OpenStack providers."""

    @pytest.fixture(scope="class")
    def openstack_provider(self, test_namespace, provider_credentials):
        """Create an OpenStack provider for mapping testing."""
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

        provider_name = "test-openstack-map-skip-verify"

        # Create command with insecure skip TLS
        cmd_parts = [
            "create provider",
            provider_name,
            "--type openstack",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            f"--provider-project-name '{creds['project_name']}'",
            f"--provider-domain-name '{creds['domain_name']}'",
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

    def test_create_network_mapping_from_openstack(
        self, test_namespace, openstack_provider
    ):
        """Test creating a network mapping from OpenStack provider."""
        mapping_name = f"test-network-map-openstack-{int(time.time())}"

        # Build network pairs string
        network_pairs = ",".join(
            [f"{n['source']}:{n['target']}" for n in OPENSTACK_NETWORKS]
        )

        # Create network mapping command
        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {openstack_provider}",
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

    def test_create_storage_mapping_from_openstack(
        self, test_namespace, openstack_provider
    ):
        """Test creating a storage mapping from OpenStack provider."""
        mapping_name = f"test-storage-map-openstack-{int(time.time())}"

        # Build storage pairs string
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OPENSTACK_VOLUME_TYPES]
        )

        # Create storage mapping command
        cmd_parts = [
            "create mapping storage",
            mapping_name,
            f"--source {openstack_provider}",
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

    def test_create_mapping_with_pod_network(self, test_namespace, openstack_provider):
        """Test creating a network mapping with pod network destination."""
        mapping_name = f"test-pod-network-map-openstack-{int(time.time())}"

        # Create network mapping with pod network
        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {openstack_provider}",
            "--target test-openshift-target",
            "--network-pairs 'provider_net_shared:pod'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create network mapping
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("networkmap", mapping_name)

        # Wait for network mapping to be ready
        wait_for_network_mapping_ready(test_namespace, mapping_name)

        # Verify pod network is in the mapping
        result = test_namespace.run_kubectl_command(
            f"get networkmap {mapping_name} -o yaml"
        )
        assert result.returncode == 0
        assert "pod" in result.stdout
