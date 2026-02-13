"""Hosts · write -- create an ESXi host resource."""

import pytest

from conftest import (
    ESXI_HOST_NAME,
    ESXI_USERNAME,
    ESXI_PASSWORD,
    TEST_NAMESPACE,
    VSPHERE_PROVIDER_NAME,
    call_tool,
)


@pytest.mark.order(14)
async def test_create_host(mcp_session):
    """Create an ESXi host resource using network-adapter lookup."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "create host",
        "flags": {
            "host-id": ESXI_HOST_NAME,
            "provider": VSPHERE_PROVIDER_NAME,
            "username": ESXI_USERNAME,
            "password": ESXI_PASSWORD,
            "network-adapter": "Management Network",
            "host-insecure-skip-tls": True,
            "namespace": TEST_NAMESPACE,
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"
    print(f"\n  Created ESXi host '{ESXI_HOST_NAME}'")