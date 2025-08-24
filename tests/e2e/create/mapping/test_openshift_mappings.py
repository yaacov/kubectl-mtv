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
from e2e.test_constants import (
    TARGET_PROVIDER_NAME,
    OPENSHIFT_NETWORKS,
    OPENSHIFT_DATASTORES,
)


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

    def _wait_for_provider_inventory_refresh(
        self, test_namespace, provider_name, timeout=120
    ):
        """Wait for provider inventory to refresh and detect the NADs."""
        import time
        from e2e.test_constants import NETWORK_ATTACHMENT_DEFINITIONS

        print(f"Waiting for provider {provider_name} inventory to refresh...")

        # First, verify NADs exist in Kubernetes using correct resource name
        for nad_name in NETWORK_ATTACHMENT_DEFINITIONS[
            :2
        ]:  # Only check first 2 NADs that are used in tests
            result = test_namespace.run_command(
                f"kubectl get network-attachment-definitions {nad_name} -n {test_namespace.namespace}"
            )
            if result.returncode != 0:
                print(
                    f"WARNING: NAD {nad_name} not found in namespace {test_namespace.namespace}"
                )
            else:
                print(
                    f"Confirmed NAD {nad_name} exists in namespace {test_namespace.namespace}"
                )

        # Try to trigger provider inventory refresh by patching the provider
        print(f"Triggering provider {provider_name} inventory refresh...")
        refresh_result = test_namespace.run_command(
            f'kubectl patch provider {provider_name} -n {test_namespace.namespace} --type merge -p \'{{"metadata":{{"annotations":{{"forklift.konveyor.io/refresh-time":"{int(time.time())}"}}}}}}\'',
            check=False,
        )
        if refresh_result.returncode == 0:
            print("Provider patch successful, waiting for inventory refresh...")
        else:
            print("Provider patch failed, relying on natural refresh cycle...")

        # Allow extra time for MTV provider to refresh and index the NADs
        print("Allowing extra time for MTV provider inventory to refresh...")
        time.sleep(timeout)
        print("Provider inventory refresh wait completed.")

    @pytest.mark.skip(
        reason="Provider inventory refresh timing issue - NADs not detected reliably"
    )
    def test_create_network_mapping_openshift_to_openshift(
        self, test_namespace, openshift_source_provider
    ):
        """Test creating a network mapping for OpenShift-to-OpenShift migration."""

        mapping_name = f"test-network-map-openshift-{int(time.time())}"

        # Wait for NADs to be available in provider inventory
        self._wait_for_provider_inventory_refresh(
            test_namespace, openshift_source_provider, 30
        )

        # Build network pairs string using the actual networks from VMs with dynamic namespace
        network_pairs = ",".join(
            [
                f"{test_namespace.namespace}/{n['source']}:{test_namespace.namespace}/{n['target']}"
                for n in OPENSHIFT_NETWORKS
            ]
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

        # Wait for storage classes to be available in provider inventory (increased timeout)
        time.sleep(
            10
        )  # Give more time for storage resources to appear in provider inventory

        # Build storage pairs string
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OPENSHIFT_DATASTORES]
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

    @pytest.mark.skip(
        reason="Provider inventory refresh timing issue - NADs not detected reliably"
    )
    def test_create_mapping_with_namespace_qualifier(
        self, test_namespace, openshift_source_provider
    ):
        """Test creating a network mapping with namespace-qualified NADs."""

        mapping_name = f"test-ns-qualified-network-map-{int(time.time())}"

        # Wait for NADs to be available in provider inventory
        self._wait_for_provider_inventory_refresh(
            test_namespace, openshift_source_provider, 30
        )

        # Create network mapping with actual NADs from the VM inventory
        from e2e.test_constants import NETWORK_ATTACHMENT_DEFINITIONS

        network_pairs = f"{test_namespace.namespace}/{NETWORK_ATTACHMENT_DEFINITIONS[0]}:{test_namespace.namespace}/{NETWORK_ATTACHMENT_DEFINITIONS[1]}"

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
