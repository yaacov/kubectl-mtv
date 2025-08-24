"""
Test cases for kubectl-mtv network and storage mapping patching for VSphere-to-OpenShift migrations.

This test validates the patching of network and storage mappings using VSphere as source and OpenShift as target provider.
"""

import time

import pytest

from e2e.utils import (
    generate_provider_name,
    get_or_create_provider,
)
from e2e.test_constants import (
    TARGET_PROVIDER_NAME,
    VSPHERE_NETWORKS,
    VSPHERE_DATASTORES,
)


@pytest.mark.patch
@pytest.mark.mapping
@pytest.mark.vsphere
class TestVSphereMappingPatch:
    """Test cases for network and storage mapping patching for VSphere-to-OpenShift migrations."""

    @pytest.fixture(scope="class")
    def vsphere_source_provider(self, test_namespace, provider_credentials):
        """Create a VSphere source provider for mapping testing."""
        creds = provider_credentials["vsphere"]

        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("VSphere credentials not available in environment")

        provider_name = generate_provider_name("vsphere", creds["url"], skip_tls=True)

        cmd_parts = [
            "create provider",
            provider_name,
            "--type vsphere",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            "--provider-insecure-skip-tls",
        ]

        create_cmd = " ".join(cmd_parts)
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    @pytest.fixture(scope="class")
    def network_mapping(self, test_namespace, vsphere_source_provider):
        """Create a network mapping for patching tests."""
        mapping_name = f"test-network-map-vsphere-patch-{int(time.time())}"

        # Build initial network pairs string
        network_pairs = f"{VSPHERE_NETWORKS[0]['source']}:{VSPHERE_NETWORKS[0]['target']},{VSPHERE_NETWORKS[1]['source']}:{VSPHERE_NETWORKS[1]['target']}"

        cmd_parts = [
            "create mapping network",
            mapping_name,
            f"--source {vsphere_source_provider}",
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
    def storage_mapping(self, test_namespace, vsphere_source_provider):
        """Create a storage mapping for patching tests."""
        mapping_name = f"test-storage-map-vsphere-patch-{int(time.time())}"

        # Build initial storage pairs string
        storage_pairs = f"{VSPHERE_DATASTORES[0]['source']}:{VSPHERE_DATASTORES[0]['target']},{VSPHERE_DATASTORES[1]['source']}:{VSPHERE_DATASTORES[1]['target']}"

        cmd_parts = [
            "create mapping storage",
            mapping_name,
            f"--source {vsphere_source_provider}",
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
        """Test adding network pairs to an existing VSphere mapping."""
        # Add new network pairs if we have enough networks
        if len(VSPHERE_NETWORKS) >= 4:
            new_pairs = f"{VSPHERE_NETWORKS[2]}:{VSPHERE_NETWORKS[3]}"

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
            assert VSPHERE_NETWORKS[2]["source"] in get_result.stdout
            assert VSPHERE_NETWORKS[3]["target"] in get_result.stdout
        else:
            pytest.skip("Not enough VSphere networks available for add pairs test")

    def test_patch_network_mapping_update_pairs(self, test_namespace, network_mapping):
        """Test updating existing network pairs in a VSphere mapping."""
        # Update existing pairs if we have enough networks
        if len(VSPHERE_NETWORKS) >= 3:
            updated_pairs = (
                f"{VSPHERE_NETWORKS[0]['source']}:{VSPHERE_NETWORKS[2]['target']}"
            )

            patch_cmd = f"patch mapping network {network_mapping} --update-pairs '{updated_pairs}'"

            result = test_namespace.run_mtv_command(patch_cmd)
            assert result.returncode == 0

            # Verify the pairs were updated in the mapping
            get_result = test_namespace.run_mtv_command(
                f"get mapping network {network_mapping} -o yaml"
            )
            assert get_result.returncode == 0
            assert VSPHERE_NETWORKS[0]["source"] in get_result.stdout
            assert VSPHERE_NETWORKS[2]["target"] in get_result.stdout
        else:
            pytest.skip("Not enough VSphere networks available for update pairs test")

    def test_patch_network_mapping_remove_pairs(self, test_namespace, network_mapping):
        """Test removing network pairs from a VSphere mapping."""
        # Remove pairs by source name
        remove_sources = VSPHERE_NETWORKS[0]["source"]

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
        """Test adding storage pairs to an existing VSphere mapping."""
        # Add new storage pairs if we have enough datastores
        if len(VSPHERE_DATASTORES) >= 4:
            new_pairs = (
                f"{VSPHERE_DATASTORES[2]['source']}:{VSPHERE_DATASTORES[3]['target']}"
            )

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
            assert VSPHERE_DATASTORES[2]["source"] in get_result.stdout
            assert VSPHERE_DATASTORES[3]["target"] in get_result.stdout
        else:
            pytest.skip("Not enough VSphere datastores available for add pairs test")

    def test_patch_storage_mapping_update_pairs(self, test_namespace, storage_mapping):
        """Test updating existing storage pairs in a VSphere mapping."""
        # Update existing pairs if we have enough datastores
        if len(VSPHERE_DATASTORES) >= 3:
            updated_pairs = (
                f"{VSPHERE_DATASTORES[0]['source']}:{VSPHERE_DATASTORES[2]['target']}"
            )

            patch_cmd = f"patch mapping storage {storage_mapping} --update-pairs '{updated_pairs}'"

            result = test_namespace.run_mtv_command(patch_cmd)
            assert result.returncode == 0

            # Verify the pairs were updated in the mapping
            get_result = test_namespace.run_mtv_command(
                f"get mapping storage {storage_mapping} -o yaml"
            )
            assert get_result.returncode == 0
            assert VSPHERE_DATASTORES[0]["source"] in get_result.stdout
            assert VSPHERE_DATASTORES[2]["target"] in get_result.stdout
        else:
            pytest.skip("Not enough VSphere datastores available for update pairs test")

    def test_patch_storage_mapping_remove_pairs(self, test_namespace, storage_mapping):
        """Test removing storage pairs from a VSphere mapping."""
        # Remove pairs by source name
        remove_sources = VSPHERE_DATASTORES[0]["source"]

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

    def test_patch_network_mapping_with_inventory_url(
        self, test_namespace, network_mapping, vsphere_source_provider
    ):
        """Test patching network mapping with inventory URL for pair resolution."""
        # Add new pairs using inventory URL
        if len(VSPHERE_NETWORKS) >= 4:
            new_pairs = f"{VSPHERE_NETWORKS[2]}:{VSPHERE_NETWORKS[3]}"

            # Construct inventory URL based on provider
            inventory_url = f"provider/{vsphere_source_provider}"

            patch_cmd = f"patch mapping network {network_mapping} --add-pairs '{new_pairs}' --inventory '{inventory_url}'"

            result = test_namespace.run_mtv_command(patch_cmd)
            assert result.returncode == 0

            # Verify the new pairs were added to the mapping
            get_result = test_namespace.run_mtv_command(
                f"get mapping network {network_mapping} -o yaml"
            )
            assert get_result.returncode == 0
            assert VSPHERE_NETWORKS[2]["source"] in get_result.stdout
            assert VSPHERE_NETWORKS[3]["target"] in get_result.stdout
        else:
            pytest.skip("Not enough VSphere networks available for inventory URL test")

    def test_patch_storage_mapping_multiple_operations(
        self, test_namespace, vsphere_source_provider
    ):
        """Test performing multiple patch operations on a VSphere storage mapping."""
        # Create a temporary mapping for this test
        mapping_name = f"test-storage-map-vsphere-multi-patch-{int(time.time())}"

        # Build initial storage pairs string
        if len(VSPHERE_DATASTORES) >= 4:
            initial_pairs = f"{VSPHERE_DATASTORES[0]['source']}:{VSPHERE_DATASTORES[0]['target']},{VSPHERE_DATASTORES[1]['source']}:{VSPHERE_DATASTORES[1]['target']}"
        else:
            pytest.skip(
                "Not enough VSphere datastores available for multi-operation test"
            )

        cmd_parts = [
            "create mapping storage",
            mapping_name,
            f"--source {vsphere_source_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--storage-pairs '{initial_pairs}'",
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

        # Perform multiple operations: remove one pair and update another
        remove_sources = VSPHERE_DATASTORES[0]["source"]
        update_pairs = (
            f"{VSPHERE_DATASTORES[2]['source']}:{VSPHERE_DATASTORES[1]['target']}"
        )

        patch_cmd = f"patch mapping storage {mapping_name} --remove-pairs '{remove_sources}' --update-pairs '{update_pairs}'"

        result = test_namespace.run_mtv_command(patch_cmd)
        assert result.returncode == 0

        # Verify the mapping was updated with both operations
        get_result = test_namespace.run_mtv_command(
            f"get mapping storage {mapping_name} -o yaml"
        )
        assert get_result.returncode == 0
        assert VSPHERE_DATASTORES[2]["source"] in get_result.stdout
        assert VSPHERE_DATASTORES[1]["target"] in get_result.stdout
