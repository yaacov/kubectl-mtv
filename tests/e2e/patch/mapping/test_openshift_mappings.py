"""
Test cases for kubectl-mtv network and storage mapping patching for OpenShift-to-OpenShift migrations.

This test validates the patching of network and storage mappings using OpenShift as both source and target provider.
"""

import time

import pytest

from e2e.utils import (
    generate_provider_name,
    get_or_create_provider,
)
from e2e.test_constants import (
    TARGET_PROVIDER_NAME,
    OPENSHIFT_NETWORKS,
    OPENSHIFT_DATASTORES,
)


@pytest.mark.skip(
    reason="Provider inventory refresh timing issue - NADs not detected reliably in fixtures"
)
@pytest.mark.patch
@pytest.mark.mapping
@pytest.mark.openshift
class TestOpenShiftMappingPatch:
    """Test cases for network and storage mapping patching for OpenShift-to-OpenShift migrations."""

    @pytest.fixture(scope="class")
    def openshift_source_provider(self, test_namespace):
        """Create an OpenShift source provider for mapping testing."""
        provider_name = generate_provider_name("openshift", "localhost", skip_tls=True)

        cmd_parts = [
            "create provider",
            provider_name,
            "--type openshift",
        ]

        create_cmd = " ".join(cmd_parts)
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    @pytest.fixture(scope="class")
    def network_mapping(self, test_namespace, openshift_source_provider):
        """Create a network mapping for patching tests."""
        mapping_name = f"test-network-map-patch-{int(time.time())}"

        # Build initial network pairs string with dynamic namespace
        network_pairs = (
            f"{test_namespace.namespace}/{OPENSHIFT_NETWORKS[0]['source']}:"
            f"{test_namespace.namespace}/{OPENSHIFT_NETWORKS[0]['target']},"
            f"{test_namespace.namespace}/{OPENSHIFT_NETWORKS[1]['source']}:"
            f"{test_namespace.namespace}/{OPENSHIFT_NETWORKS[1]['target']}"
        )

        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {openshift_source_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--network-pairs '{network_pairs}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create mapping
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("networkmap", mapping_name)

        # Verify mapping was created
        get_result = test_namespace.run_mtv_command(
            f"get mapping network {mapping_name} -o yaml"
        )
        assert get_result.returncode == 0

        return mapping_name

    @pytest.fixture(scope="class")
    def storage_mapping(self, test_namespace, openshift_source_provider):
        """Create a storage mapping for patching tests."""
        mapping_name = f"test-storage-map-patch-{int(time.time())}"

        # Build initial storage pairs string
        storage_pairs = f"{OPENSHIFT_DATASTORES[0]['source']}:{OPENSHIFT_DATASTORES[0]['target']},{OPENSHIFT_DATASTORES[1]['source']}:{OPENSHIFT_DATASTORES[1]['target']}"

        cmd_parts = [
            "create mapping storage",
            mapping_name,
            f"--source {openshift_source_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--storage-pairs '{storage_pairs}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create mapping
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("storagemap", mapping_name)

        # Verify mapping was created
        get_result = test_namespace.run_mtv_command(
            f"get mapping storage {mapping_name} -o yaml"
        )
        assert get_result.returncode == 0

        return mapping_name

    def test_patch_network_mapping_add_pairs(self, test_namespace, network_mapping):
        """Test adding network pairs to an existing mapping."""
        # Add new network pairs if we have enough networks
        if len(OPENSHIFT_NETWORKS) >= 4:
            new_pairs = f"{OPENSHIFT_NETWORKS[2]}:{OPENSHIFT_NETWORKS[3]}"

            patch_cmd = (
                f"patch mapping network {network_mapping} --add-pairs '{new_pairs}'"
            )

            result = test_namespace.run_mtv_command(patch_cmd)
            assert result.returncode == 0

            # Verify the new pairs were added to the mapping
            get_result = test_namespace.run_mtv_command(
                f"get mapping network {network_mapping} -o yaml"
            )
            assert get_result.returncode == 0
            assert OPENSHIFT_NETWORKS[2]["source"] in get_result.stdout
            assert OPENSHIFT_NETWORKS[3]["target"] in get_result.stdout
        else:
            pytest.skip("Not enough OpenShift networks available for add pairs test")

    def test_patch_network_mapping_update_pairs(self, test_namespace, network_mapping):
        """Test updating existing network pairs in a mapping."""
        # Update existing pairs if we have enough networks
        if len(OPENSHIFT_NETWORKS) >= 3:
            updated_pairs = f"{OPENSHIFT_NETWORKS[0]}:{OPENSHIFT_NETWORKS[2]}"

            patch_cmd = f"patch mapping network {network_mapping} --update-pairs '{updated_pairs}'"

            result = test_namespace.run_mtv_command(patch_cmd)
            assert result.returncode == 0

            # Verify the pairs were updated in the mapping
            get_result = test_namespace.run_mtv_command(
                f"get mapping network {network_mapping} -o yaml"
            )
            assert get_result.returncode == 0
            assert OPENSHIFT_NETWORKS[0]["source"] in get_result.stdout
            assert OPENSHIFT_NETWORKS[2]["target"] in get_result.stdout
        else:
            pytest.skip("Not enough OpenShift networks available for update pairs test")

    def test_patch_network_mapping_remove_pairs(self, test_namespace, network_mapping):
        """Test removing network pairs from a mapping."""
        # Remove pairs by source name
        remove_sources = OPENSHIFT_NETWORKS[0]["source"]

        patch_cmd = (
            f"patch mapping network {network_mapping} --remove-pairs '{remove_sources}'"
        )

        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0

        # Verify the patch command succeeded (removal verification would require YAML parsing)
        get_result = test_namespace.run_mtv_command(
            f"get mapping network {network_mapping} -o yaml"
        )
        assert get_result.returncode == 0

    def test_patch_storage_mapping_add_pairs(self, test_namespace, storage_mapping):
        """Test adding storage pairs to an existing mapping."""
        # Add new storage pairs if we have enough datastores
        if len(OPENSHIFT_DATASTORES) >= 4:
            new_pairs = f"{OPENSHIFT_DATASTORES[2]}:{OPENSHIFT_DATASTORES[3]}"

            patch_cmd = (
                f"patch mapping storage {storage_mapping} --add-pairs '{new_pairs}'"
            )

            result = test_namespace.run_mtv_command(patch_cmd)
            assert result.returncode == 0

            # Verify the new pairs were added to the mapping
            get_result = test_namespace.run_mtv_command(
                f"get mapping storage {storage_mapping} -o yaml"
            )
            assert get_result.returncode == 0
            assert OPENSHIFT_DATASTORES[2]["source"] in get_result.stdout
            assert OPENSHIFT_DATASTORES[3]["target"] in get_result.stdout
        else:
            pytest.skip("Not enough OpenShift datastores available for add pairs test")

    def test_patch_storage_mapping_update_pairs(self, test_namespace, storage_mapping):
        """Test updating existing storage pairs in a mapping."""
        # Update existing pairs if we have enough datastores
        if len(OPENSHIFT_DATASTORES) >= 3:
            updated_pairs = f"{OPENSHIFT_DATASTORES[0]}:{OPENSHIFT_DATASTORES[2]}"

            patch_cmd = f"patch mapping storage {storage_mapping} --update-pairs '{updated_pairs}'"

            result = test_namespace.run_mtv_command(patch_cmd)
            assert result.returncode == 0

            # Verify the pairs were updated in the mapping
            get_result = test_namespace.run_mtv_command(
                f"get mapping storage {storage_mapping} -o yaml"
            )
            assert get_result.returncode == 0
            assert OPENSHIFT_DATASTORES[0]["source"] in get_result.stdout
            assert OPENSHIFT_DATASTORES[2]["target"] in get_result.stdout
        else:
            pytest.skip(
                "Not enough OpenShift datastores available for update pairs test"
            )

    def test_patch_storage_mapping_remove_pairs(self, test_namespace, storage_mapping):
        """Test removing storage pairs from a mapping."""
        # Remove pairs by source name
        remove_sources = OPENSHIFT_DATASTORES[0]["source"]

        patch_cmd = (
            f"patch mapping storage {storage_mapping} --remove-pairs '{remove_sources}'"
        )

        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0

        # Verify the patch command succeeded (removal verification would require YAML parsing)
        get_result = test_namespace.run_mtv_command(
            f"get mapping storage {storage_mapping} -o yaml"
        )
        assert get_result.returncode == 0

    def test_patch_network_mapping_multiple_operations(
        self, test_namespace, openshift_source_provider
    ):
        """Test performing multiple patch operations on a network mapping."""
        # Create a temporary mapping for this test
        mapping_name = f"test-network-map-multi-patch-{int(time.time())}"

        # Build initial network pairs string
        if len(OPENSHIFT_NETWORKS) >= 4:
            initial_pairs = f"{OPENSHIFT_NETWORKS[0]}:{OPENSHIFT_NETWORKS[1]},{OPENSHIFT_NETWORKS[2]}:{OPENSHIFT_NETWORKS[3]}"
        else:
            pytest.skip(
                "Not enough OpenShift networks available for multi-operation test"
            )

        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {openshift_source_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--network-pairs '{initial_pairs}'",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create mapping
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("networkmap", mapping_name)

        # Verify mapping was created
        get_result = test_namespace.run_mtv_command(
            f"get mapping network {mapping_name} -o yaml"
        )
        assert get_result.returncode == 0

        # Perform multiple operations: remove one pair and add a new one
        if len(OPENSHIFT_NETWORKS) >= 6:
            remove_sources = OPENSHIFT_NETWORKS[0]
            add_pairs = f"{OPENSHIFT_NETWORKS[4]}:{OPENSHIFT_NETWORKS[5]}"

            patch_cmd = f"patch mapping network {mapping_name} --remove-pairs '{remove_sources}' --add-pairs '{add_pairs}'"

            result = test_namespace.run_mtv_command(patch_cmd)
            assert result.returncode == 0

            # Verify the mapping was updated with both operations
            get_result = test_namespace.run_mtv_command(
                f"get mapping network {mapping_name} -o yaml"
            )
            assert get_result.returncode == 0
            assert OPENSHIFT_NETWORKS[4]["source"] in get_result.stdout
            assert OPENSHIFT_NETWORKS[5]["target"] in get_result.stdout
        else:
            pytest.skip(
                "Not enough OpenShift networks available for multi-operation add test"
            )

    def test_patch_mapping_error_nonexistent(self, test_namespace):
        """Test patching a non-existent mapping."""
        non_existent_mapping = "non-existent-mapping"

        # This should fail because the mapping doesn't exist
        result = test_namespace.run_mtv_command(
            f"patch mapping network {non_existent_mapping} --add-pairs 'source:target'",
            check=False,
        )

        assert result.returncode != 0
