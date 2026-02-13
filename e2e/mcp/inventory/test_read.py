"""Inventory · read -- query vSphere inventory for VMs, datastores, networks, hosts."""

import pytest

from conftest import (
    COLD_VMS,
    ESXI_HOST_NAME,
    TEST_DATASTORE_NAME,
    TEST_NAMESPACE,
    TEST_NETWORK_NAME,
    VSPHERE_PROVIDER_NAME,
    WARM_VMS,
    call_tool,
)

# All test VMs that must exist in the inventory
_TEST_VM_NAMES = set((COLD_VMS + "," + WARM_VMS).split(","))


# ===================================================================
# VMs
# ===================================================================
@pytest.mark.order(60)
async def test_inventory_vms(mcp_session):
    """List VMs and verify that every test VM exists in the inventory."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
        },
    })

    vms = result.get("data", [])
    assert isinstance(vms, list)
    assert len(vms) >= len(_TEST_VM_NAMES), (
        f"Expected at least {len(_TEST_VM_NAMES)} VMs, got {len(vms)}"
    )

    vm_names = {v.get("name") for v in vms}
    for vm in _TEST_VM_NAMES:
        assert vm in vm_names, f"Test VM '{vm}' not found in inventory"

    print(f"\n  ✓ All {len(_TEST_VM_NAMES)} test VMs found (total VMs: {len(vms)})")


@pytest.mark.order(61)
async def test_inventory_vms_query(mcp_session):
    """Filter VMs using a TSL query."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where name ~= 'mtv-rhel8-.*'",
        },
    })

    vms = result.get("data", [])
    assert isinstance(vms, list)
    assert len(vms) >= 1, "TSL query should return at least one VM"

    for vm in vms:
        name = vm.get("name", "")
        assert name.startswith("mtv-rhel8-"), (
            f"VM '{name}' does not match query pattern 'mtv-rhel8-.*'"
        )


# ===================================================================
# Datastores
# ===================================================================
@pytest.mark.order(62)
async def test_inventory_datastores(mcp_session):
    """List datastores and verify the test datastore exists."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory datastore",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
        },
    })

    datastores = result.get("data", [])
    assert isinstance(datastores, list)
    assert len(datastores) >= 1, "Expected at least 1 datastore"

    ds_names = {d.get("name") for d in datastores}
    assert TEST_DATASTORE_NAME in ds_names, (
        f"Test datastore '{TEST_DATASTORE_NAME}' not found; got {ds_names}"
    )

    print(f"\n  ✓ Test datastore '{TEST_DATASTORE_NAME}' found (total: {len(datastores)})")


# ===================================================================
# Networks
# ===================================================================
@pytest.mark.order(63)
async def test_inventory_networks(mcp_session):
    """List networks and verify the test network exists."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory network",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
        },
    })

    networks = result.get("data", [])
    assert isinstance(networks, list)
    assert len(networks) >= 1, "Expected at least 1 network"

    net_names = {n.get("name") for n in networks}
    assert TEST_NETWORK_NAME in net_names, (
        f"Test network '{TEST_NETWORK_NAME}' not found; got {net_names}"
    )

    print(f"\n  ✓ Test network '{TEST_NETWORK_NAME}' found (total: {len(networks)})")


# ===================================================================
# Hosts
# ===================================================================
@pytest.mark.order(64)
async def test_inventory_hosts(mcp_session):
    """List inventory hosts and verify the ESXi host used for testing."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory host",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
        },
    })

    hosts = result.get("data", [])
    assert isinstance(hosts, list)
    assert len(hosts) >= 1, "Expected at least 1 inventory host"

    # ESXI_HOST_NAME may be an inventory ID (e.g. "host-8") or a display
    # name (e.g. "10.6.46.29").  Check both fields.
    host_ids = {h.get("id") for h in hosts}
    host_names = {h.get("name") for h in hosts}
    assert ESXI_HOST_NAME in host_ids | host_names, (
        f"ESXi host '{ESXI_HOST_NAME}' not found in inventory; "
        f"ids={host_ids}, names={host_names}"
    )

    print(f"\n  ✓ ESXi host '{ESXI_HOST_NAME}' found (total: {len(hosts)})")
