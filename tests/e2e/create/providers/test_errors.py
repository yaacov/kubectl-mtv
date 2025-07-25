"""
Test cases for kubectl-mtv provider error conditions and edge cases.

This test validates error handling and validation for provider operations.
"""

import pytest


@pytest.mark.create
@pytest.mark.provider
@pytest.mark.providers
@pytest.mark.error_cases
class TestProviderErrors:
    """Test cases for provider error conditions."""

    def test_create_provider_invalid_type(self, test_namespace):
        """Test creating a provider with invalid type."""
        provider_name = "test-invalid-type-provider"

        # This should fail because "invalid" is not a valid provider type
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type invalid", check=False
        )

        assert result.returncode != 0
        assert "invalid" in result.stderr.lower() or "unknown" in result.stderr.lower()

    def test_create_provider_missing_type(self, test_namespace):
        """Test creating a provider without specifying type."""
        provider_name = "test-missing-type-provider"

        # This should fail because --type is required
        result = test_namespace.run_mtv_command(
            f"create provider {provider_name}", check=False
        )

        assert result.returncode != 0

    def test_create_provider_duplicate_name(self, test_namespace):
        """Test creating a provider with duplicate name."""
        provider_name = "test-duplicate-provider"

        # Create first provider (should succeed)
        result1 = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type openshift"
        )
        assert result1.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)

        # Try to create second provider with same name (should fail)
        result2 = test_namespace.run_mtv_command(
            f"create provider {provider_name} --type openshift", check=False
        )

        assert result2.returncode != 0
        assert (
            "already exists" in result2.stderr.lower()
            or "conflict" in result2.stderr.lower()
        )

        # Clean up the first provider since test passed
        test_namespace.run_mtv_command(f"delete provider {provider_name}")

    def test_get_nonexistent_provider(self, test_namespace):
        """Test getting a provider that doesn't exist."""
        nonexistent_name = "nonexistent-provider-12345"

        result = test_namespace.run_mtv_command(
            f"get provider {nonexistent_name}", check=False
        )

        assert result.returncode != 0
        assert (
            "not found" in result.stderr.lower() or "notfound" in result.stderr.lower()
        )

    def test_delete_nonexistent_provider(self, test_namespace):
        """Test deleting a provider that doesn't exist."""
        nonexistent_name = "nonexistent-provider-12345"

        result = test_namespace.run_mtv_command(
            f"delete provider {nonexistent_name}", check=False
        )

        assert result.returncode != 0
        assert (
            "not found" in result.stderr.lower() or "notfound" in result.stderr.lower()
        )
