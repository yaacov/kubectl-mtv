"""
Test cases for kubectl-mtv host creation error scenarios.

This test validates error handling and validation in host creation commands.
"""

import pytest

from ...utils import (
    delete_hosts_by_spec_id,
    generate_provider_name,
    get_or_create_provider,
)


# Hardcoded host IDs for error testing (from hosts.json file)
ERROR_TEST_HOSTS = ["host-8", "invalid-host-999"]

# Hardcoded network adapter names for testing (from hosts.json file)
ERROR_TEST_ADAPTERS = ["Management Network", "NonExistentAdapter"]


@pytest.mark.create
@pytest.mark.host
@pytest.mark.hosts
@pytest.mark.error_cases
class TestHostErrors:
    """Test cases for host creation error scenarios."""

    def test_create_host_missing_provider(self, test_namespace):
        """Test creating a host without specifying a provider."""
        host_id = ERROR_TEST_HOSTS[0]

        # Delete host if it exists from previous test
        delete_hosts_by_spec_id(test_namespace, host_id)

        result = test_namespace.run_mtv_command(
            f"create host {host_id} --ip-address 192.168.1.100 --username root --password secret",
            check=False,
        )
        assert result.returncode != 0
        assert 'required flag(s) "provider" not set' in result.stderr

    def test_create_host_nonexistent_provider(self, test_namespace):
        """Test creating a host with a non-existent provider."""
        host_id = ERROR_TEST_HOSTS[0]

        # Delete host if it exists from previous test
        delete_hosts_by_spec_id(test_namespace, host_id)

        result = test_namespace.run_mtv_command(
            f"create host {host_id} --provider nonexistent-provider --ip-address 192.168.1.100 --username root --password secret",
            check=False,
        )
        assert result.returncode != 0
        # Should fail when trying to validate the provider

    def test_create_host_non_vsphere_provider(
        self, test_namespace, provider_credentials
    ):
        """Test creating a host with a non-vSphere provider (should fail)."""
        ova_creds = provider_credentials["ova"]

        # Skip if OVA URL is not available
        if not ova_creds.get("url"):
            pytest.skip("OVA URL not available in environment")

        # Create an OVA provider
        provider_name = generate_provider_name("ova", ova_creds["url"], skip_tls=True)
        create_cmd = (
            f"create provider {provider_name} --type ova --url '{ova_creds['url']}'"
        )

        # Create provider if it doesn't already exist
        get_or_create_provider(test_namespace, provider_name, create_cmd)

        # Try to create a host with OVA provider (should fail)
        host_id = ERROR_TEST_HOSTS[0]
        host_result = test_namespace.run_mtv_command(
            f"create host {host_id} --provider {provider_name} --ip-address 192.168.1.100 --username root --password secret",
            check=False,
        )
        assert host_result.returncode != 0
        assert (
            "only vsphere providers support host creation" in host_result.stderr.lower()
        )

    def test_create_host_missing_ip_resolution(
        self, test_namespace, vsphere_provider, provider_credentials
    ):
        """Test creating a host without IP address or network adapter."""
        esxi_creds = provider_credentials["esxi"]
        host_id = ERROR_TEST_HOSTS[0]

        # Delete host if it exists from previous test
        delete_hosts_by_spec_id(test_namespace, host_id)

        # Try to create host without IP address or network adapter (vCenter providers need ESXi creds)
        host_result = test_namespace.run_mtv_command(
            f"create host {host_id} --provider {vsphere_provider} "
            f"--username '{esxi_creds['username']}' --password '{esxi_creds['password']}'",
            check=False,
        )
        assert host_result.returncode != 0
        assert (
            "either --ip-address OR --network-adapter must be provided"
            in host_result.stderr
        )

    def test_create_host_both_ip_and_adapter(
        self, test_namespace, vsphere_provider, provider_credentials
    ):
        """Test creating a host with both IP address and network adapter (should fail)."""
        esxi_creds = provider_credentials["esxi"]
        host_id = ERROR_TEST_HOSTS[0]

        # Delete host if it exists from previous test
        delete_hosts_by_spec_id(test_namespace, host_id)

        # Try to create host with both IP address and network adapter (vCenter providers need ESXi creds)
        host_result = test_namespace.run_mtv_command(
            f"create host {host_id} --provider {vsphere_provider} "
            f"--ip-address 192.168.1.100 --network-adapter '{ERROR_TEST_ADAPTERS[0]}' "
            f"--username '{esxi_creds['username']}' --password '{esxi_creds['password']}'",
            check=False,
        )
        assert host_result.returncode != 0
        assert (
            "cannot use both --ip-address and --network-adapter" in host_result.stderr
        )

    @pytest.fixture(scope="class")
    def vsphere_provider(self, test_namespace, provider_credentials):
        """Create a vSphere (vCenter) provider for error testing."""
        creds = provider_credentials["vsphere"]

        # Skip if vSphere credentials are not available
        if not all([creds.get("url"), creds.get("username"), creds.get("password")]):
            pytest.skip("VMware vSphere credentials not available in environment")

        # Generate provider name based on type and configuration
        provider_name = generate_provider_name("vsphere", creds["url"], skip_tls=True)

        cmd_parts = [
            "create provider",
            provider_name,
            "--type vsphere",
            f"--url '{creds['url']}'",
            f"--username '{creds['username']}'",
            f"--password '{creds['password']}'",
            "--provider-insecure-skip-tls",
            # No --sdk-endpoint esxi, so this is a vCenter provider
        ]

        create_provider_cmd = " ".join(cmd_parts)

        # Create provider if it doesn't already exist
        return get_or_create_provider(
            test_namespace, provider_name, create_provider_cmd
        )

    def test_create_host_missing_auth_for_vcenter_provider(
        self, test_namespace, vsphere_provider
    ):
        """Test creating a host with vCenter provider but missing authentication."""
        host_id = ERROR_TEST_HOSTS[0]

        # Delete host if it exists from previous test
        delete_hosts_by_spec_id(test_namespace, host_id)

        # Try to create host without authentication (should fail for non-ESXi providers)
        host_result = test_namespace.run_mtv_command(
            f"create host {host_id} --provider {vsphere_provider} --ip-address 192.168.1.100",
            check=False,
        )
        assert host_result.returncode != 0
        assert (
            "either --existing-secret OR both --username and --password must be provided"
            in host_result.stderr
        )

    def test_create_host_conflicting_auth_options(
        self, test_namespace, vsphere_provider, provider_credentials
    ):
        """Test creating a host with both existing secret and username/password."""
        esxi_creds = provider_credentials["esxi"]
        host_id = ERROR_TEST_HOSTS[0]

        # Try to create host with both existing secret and username/password
        host_result = test_namespace.run_mtv_command(
            f"create host {host_id} --provider {vsphere_provider} "
            f"--ip-address 192.168.1.100 "
            f"--existing-secret some-secret "
            f"--username '{esxi_creds['username']}' --password '{esxi_creds['password']}'",
            check=False,
        )
        assert host_result.returncode != 0
        assert (
            "cannot use both --existing-secret and --username/--password"
            in host_result.stderr
        )

    def test_create_host_invalid_host_id(
        self, test_namespace, vsphere_provider, provider_credentials
    ):
        """Test creating a host with a host ID that doesn't exist in provider inventory."""
        esxi_creds = provider_credentials["esxi"]
        host_id = ERROR_TEST_HOSTS[1]  # invalid-host-999

        # Try to create host with invalid host ID (vCenter providers need ESXi creds)
        host_result = test_namespace.run_mtv_command(
            f"create host {host_id} --provider {vsphere_provider} "
            f"--ip-address 192.168.1.100 "
            f"--username '{esxi_creds['username']}' --password '{esxi_creds['password']}'",
            check=False,
        )
        assert host_result.returncode != 0
        assert "not found in provider inventory" in host_result.stderr

    def test_create_host_nonexistent_network_adapter(
        self, test_namespace, vsphere_provider, provider_credentials
    ):
        """Test creating a host with a network adapter that doesn't exist."""
        esxi_creds = provider_credentials["esxi"]
        host_id = ERROR_TEST_HOSTS[0]
        adapter_name = ERROR_TEST_ADAPTERS[1]  # NonExistentAdapter

        # Try to create host with non-existent network adapter (vCenter providers need ESXi creds)
        host_result = test_namespace.run_mtv_command(
            f"create host {host_id} --provider {vsphere_provider} "
            f"--network-adapter '{adapter_name}' "
            f"--username '{esxi_creds['username']}' --password '{esxi_creds['password']}'",
            check=False,
        )
        assert host_result.returncode != 0
        assert "not found" in host_result.stderr.lower()

    def test_create_host_invalid_cacert_file(
        self, test_namespace, vsphere_provider, provider_credentials
    ):
        """Test creating a host with an invalid CA certificate file path."""
        esxi_creds = provider_credentials["esxi"]
        host_id = ERROR_TEST_HOSTS[0]

        # Try to create host with invalid cacert file (vCenter providers need ESXi creds)
        host_result = test_namespace.run_mtv_command(
            f"create host {host_id} --provider {vsphere_provider} "
            f"--ip-address 192.168.1.100 "
            f"--username '{esxi_creds['username']}' --password '{esxi_creds['password']}' "
            f"--cacert @/nonexistent/ca.cert",
            check=False,
        )
        assert host_result.returncode != 0
        assert "failed to read CA certificate file" in host_result.stderr
