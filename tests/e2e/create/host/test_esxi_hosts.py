"""
Test cases for kubectl-mtv ESXi host creation.

This test validates the creation of migration hosts specifically for ESXi providers
which can reuse provider credentials for host authentication.
"""

import pytest

from ...utils import (
    wait_for_host_ready,
    delete_hosts_by_spec_id,
)


# Use centralized constants from test_constants.py
from ...test_constants import ESXI_TEST_HOSTS


@pytest.mark.create
@pytest.mark.host
@pytest.mark.hosts
@pytest.mark.esxi
@pytest.mark.requires_credentials
class TestESXiHosts:
    """Test cases for ESXi migration host creation."""

    # Provider fixtures are now session-scoped in conftest.py

    def test_create_esxi_host_with_auto_credentials(
        self, test_namespace, esxi_provider
    ):
        """Test creating an ESXi host that automatically uses provider credentials."""
        host_id = ESXI_TEST_HOSTS[0]
        # Use actual IP address from inventory (Management Network)
        host_ip = "10.6.46.30"

        # Delete host if it exists from previous test
        delete_hosts_by_spec_id(test_namespace, host_id)

        # Create host with manual IP (ESXi providers automatically use provider credentials)
        host_result = test_namespace.run_mtv_command(
            f"create host {host_id} --provider {esxi_provider} --ip-address {host_ip}"
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
