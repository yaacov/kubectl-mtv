"""
Test cases for kubectl-mtv migration plan creation from OpenStack providers.

This test validates the creation of migration plans using OpenStack as the source provider
and verifies that plans become ready.
"""

import time

import pytest

from e2e.utils import (
    wait_for_plan_ready,
    generate_provider_name,
    get_or_create_provider,
)
from e2e.test_constants import TARGET_PROVIDER_NAME


# Import VM names from centralized constants
from ...test_constants import OPENSTACK_TEST_VMS


@pytest.mark.create
@pytest.mark.plan
@pytest.mark.openstack
@pytest.mark.requires_credentials
class TestOpenStackPlanCreation:
    """Test cases for migration plan creation from OpenStack providers."""

    @pytest.fixture(scope="class")
    def openstack_provider(self, test_namespace, provider_credentials):
        """Create an OpenStack provider for plan testing."""
        creds = provider_credentials["openstack"]

        # Skip if credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("OpenStack credentials not available in environment")

        # Generate provider name based on type and configuration
        provider_name = generate_provider_name("openstack", creds["url"], skip_tls=True)

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

        # Create provider if it doesn't already exist
        return get_or_create_provider(test_namespace, provider_name, create_cmd)

    def test_create_plan_from_openstack(self, test_namespace, openstack_provider):
        """Test creating a migration plan from OpenStack provider."""
        # Use the first available VM as comma-separated string
        selected_vm = OPENSTACK_TEST_VMS[0]
        plan_name = f"test-plan-openstack-{int(time.time())}"

        # Import storage pairs and create storage mapping first to avoid unmapped storage
        from e2e.test_constants import OPENSTACK_STORAGE_PAIRS

        # Create storage mapping to ensure all VM storage is mapped
        storage_mapping_name = f"test-storage-mapping-{plan_name}"
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OPENSTACK_STORAGE_PAIRS]
        )

        storage_cmd_parts = [
            "create mapping storage",
            storage_mapping_name,
            f"--source {openstack_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--storage-pairs '{storage_pairs}'",
        ]

        storage_cmd = " ".join(storage_cmd_parts)

        # Create storage mapping
        result = test_namespace.run_mtv_command(storage_cmd)
        assert result.returncode == 0

        # Track storage mapping for cleanup
        test_namespace.track_resource("storagemap", storage_mapping_name)

        # Wait for storage mapping to be ready
        from e2e.utils import wait_for_storage_mapping_ready

        wait_for_storage_mapping_ready(test_namespace, storage_mapping_name)

        # Create plan command with storage mapping
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {openstack_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--vms '{selected_vm}'",
            f"--storage-mapping {storage_mapping_name}",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create plan
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("plan", plan_name)

        # Wait for plan to be ready
        wait_for_plan_ready(test_namespace, plan_name)

    def test_create_multi_vm_plan_from_openstack(
        self, test_namespace, openstack_provider
    ):
        """Test creating a migration plan with multiple VMs from OpenStack provider."""
        # Use first 3 VMs for multi-VM test as comma-separated string
        selected_vms = ",".join(OPENSTACK_TEST_VMS[:3])
        plan_name = f"test-multi-plan-openstack-{int(time.time())}"

        # Import storage pairs and create storage mapping first
        from e2e.test_constants import OPENSTACK_STORAGE_PAIRS

        # Create storage mapping to ensure all VM storage is mapped
        storage_mapping_name = f"test-storage-mapping-{plan_name}"
        storage_pairs = ",".join(
            [f"{s['source']}:{s['target']}" for s in OPENSTACK_STORAGE_PAIRS]
        )

        storage_cmd_parts = [
            "create mapping storage",
            storage_mapping_name,
            f"--source {openstack_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--storage-pairs '{storage_pairs}'",
        ]

        storage_cmd = " ".join(storage_cmd_parts)

        # Create storage mapping
        result = test_namespace.run_mtv_command(storage_cmd)
        assert result.returncode == 0

        # Track storage mapping for cleanup
        test_namespace.track_resource("storagemap", storage_mapping_name)

        # Wait for storage mapping to be ready
        from e2e.utils import wait_for_storage_mapping_ready

        wait_for_storage_mapping_ready(test_namespace, storage_mapping_name)

        # Create plan command with multiple VMs and the storage mapping
        cmd_parts = [
            "create plan",
            plan_name,
            f"--source {openstack_provider}",
            f"--target {TARGET_PROVIDER_NAME}",
            f"--vms '{selected_vms}'",
            f"--storage-mapping {storage_mapping_name}",
        ]

        create_cmd = " ".join(cmd_parts)

        # Create plan
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("plan", plan_name)

        # Wait for plan to be ready (longer timeout for multi-VM plans)
        wait_for_plan_ready(test_namespace, plan_name)
