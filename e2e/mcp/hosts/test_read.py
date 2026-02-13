"""Hosts · read -- list and describe ESXi hosts."""

import pytest

from conftest import (
    ESXI_HOST_NAME,
    TEST_NAMESPACE,
    VSPHERE_PROVIDER_NAME,
    call_tool,
)

# Module-level store so test_describe_host can reuse the discovered K8s name.
_discovered_resource_name: str | None = None


def _find_host_resource_name(hosts: list[dict], inventory_id: str) -> str | None:
    """Return the K8s resource name whose name starts with *inventory_id*.

    The ``create host`` command generates a resource name like
    ``{inventoryID}-{hash}``, so we match on prefix.
    """
    for h in hosts:
        name = h.get("name") or h.get("metadata", {}).get("name", "")
        if name.startswith(inventory_id):
            return name
    return None


@pytest.mark.order(30)
async def test_get_host(mcp_session):
    """List hosts and verify our ESXi host resource exists."""
    global _discovered_resource_name  # noqa: PLW0603

    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get host",
        "flags": {"namespace": TEST_NAMESPACE, "output": "json"},
    })

    data = result.get("data", [])
    hosts = data if isinstance(data, list) else [data]

    resource_name = _find_host_resource_name(hosts, ESXI_HOST_NAME)
    assert resource_name is not None, (
        f"No host resource starting with '{ESXI_HOST_NAME}' found; "
        f"got {[h.get('name') or h.get('metadata', {}).get('name') for h in hosts]}"
    )

    _discovered_resource_name = resource_name
    print(f"\n  ✓ Host resource '{resource_name}' found (inventory ID: {ESXI_HOST_NAME})")


@pytest.mark.order(31)
async def test_describe_host(mcp_session):
    """Describe the ESXi host and check key fields."""
    name = _discovered_resource_name
    assert name, "test_get_host must run first to discover the resource name"

    result = await call_tool(mcp_session, "mtv_read", {
        "command": "describe host",
        "flags": {"name": name, "namespace": TEST_NAMESPACE},
    })

    output = result.get("output", "")
    assert name in output, f"Resource name '{name}' missing from describe output"
    assert VSPHERE_PROVIDER_NAME in output, "Provider reference missing"
