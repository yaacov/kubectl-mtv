"""
Test cases for kubectl-mtv vSphere host creation.

This test validates the creation of migration hosts for vSphere providers.
Hosts represent ESXi servers in the vSphere infrastructure.
"""

import pytest

from ...utils import (
    wait_for_host_ready,
    delete_hosts_by_spec_id,
)


# Use centralized constants from test_constants.py
from ...test_constants import VSPHERE_TEST_HOSTS, NETWORK_ADAPTERS


@pytest.mark.create
@pytest.mark.host
@pytest.mark.hosts
@pytest.mark.vsphere
@pytest.mark.requires_credentials
class TestVSphereHosts:
    """Test cases for vSphere migration host creation."""

    # Provider fixtures are now session-scoped in conftest.py

    def test_create_host_with_ip_address(
        self, test_namespace, vsphere_provider, provider_credentials
    ):
        """Test creating a host with direct IP address specification."""
        esxi_creds = provider_credentials["esxi"]
        host_id = VSPHERE_TEST_HOSTS[0]  # Use the single available host
        host_ip = "10.6.46.29"  # Actual IP from inventory

        # Delete host if it exists from previous test
        delete_hosts_by_spec_id(test_namespace, host_id)

        # Create host with direct IP address and ESXi credentials (vCenter providers need explicit auth)
        host_result = test_namespace.run_mtv_command(
            f"create host {host_id} --provider {vsphere_provider} "
            f"--ip-address {host_ip} "
            f"--username '{esxi_creds['username']}' "
            f"--password '{esxi_creds['password']}' "
            f"--host-insecure-skip-tls"
        )
        assert host_result.returncode == 0

        # Parse the created host name from output (format: "host/actual-name created")
        host_name = None
        for line in host_result.stdout.split("\n"):
            if line.startswith("host/") and line.endswith(" created"):
                host_name = line.split("host/")[1].split(" created")[0]
                break

        assert host_name is not None, "Could not parse host name from output"

        # Track for cleanup
        test_namespace.track_resource("host", host_name)

        # Wait for host to be ready
        wait_for_host_ready(test_namespace, host_name)

    def test_create_host_with_network_adapter(
        self, test_namespace, vsphere_provider, provider_credentials
    ):
        """Test creating a host using network adapter resolution."""
        esxi_creds = provider_credentials["esxi"]

        # Use the single available host with network adapter resolution
        host_id = VSPHERE_TEST_HOSTS[0]  # host-8
        adapter_name = NETWORK_ADAPTERS[0]  # Management Network

        # Delete host if it exists from previous test
        delete_hosts_by_spec_id(test_namespace, host_id)

        # Create host with vCenter provider using network adapter (needs explicit ESXi credentials)
        host_result = test_namespace.run_mtv_command(
            f"create host {host_id} --provider {vsphere_provider} "
            f"--network-adapter '{adapter_name}' "
            f"--username '{esxi_creds['username']}' "
            f"--password '{esxi_creds['password']}' "
            f"--host-insecure-skip-tls"
        )
        assert host_result.returncode == 0

        # Parse the created host name from output (format: "host/actual-name created")
        host_name = None
        for line in host_result.stdout.split("\n"):
            if line.startswith("host/") and line.endswith(" created"):
                host_name = line.split("host/")[1].split(" created")[0]
                break

        assert host_name is not None, "Could not parse host name from output"

        # Track for cleanup
        test_namespace.track_resource("host", host_name)

        # Wait for host to be ready
        wait_for_host_ready(test_namespace, host_name)
