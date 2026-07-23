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


# -------------------------------------------------------------------
# TSL Query Tests — Comparison Operators
# -------------------------------------------------------------------
@pytest.mark.order(61)
async def test_query_eq(mcp_session):
    """Equality operator (= and ==)."""
    cold_vm = COLD_VMS.split(",")[0]

    for op in ("=", "=="):
        result = await call_tool(mcp_session, "mtv_read", {
            "command": "get inventory vm",
            "flags": {
                "provider": VSPHERE_PROVIDER_NAME,
                "namespace": TEST_NAMESPACE,
                "output": "json",
                "query": f"where name {op} '{cold_vm}'",
            },
        })
        vms = result.get("data", [])
        assert len(vms) == 1, f"'{op}' should return exactly 1 VM, got {len(vms)}"
        assert vms[0].get("name") == cold_vm

    print(f"\n  ✓ EQ operators (= and ==) both returned '{cold_vm}'")


@pytest.mark.order(after="test_query_eq")
async def test_query_ne(mcp_session):
    """Not-equal operators (!= and <>)."""
    cold_vm = COLD_VMS.split(",")[0]

    for op in ("!=", "<>"):
        result = await call_tool(mcp_session, "mtv_read", {
            "command": "get inventory vm",
            "flags": {
                "provider": VSPHERE_PROVIDER_NAME,
                "namespace": TEST_NAMESPACE,
                "output": "json",
                "query": f"where name ~= '^mtv-rhel8' and name {op} '{cold_vm}'",
            },
        })
        vms = result.get("data", [])
        assert len(vms) >= 1, f"'{op}' should return at least 1 VM"
        vm_names = {v.get("name") for v in vms}
        assert cold_vm not in vm_names, f"'{op}' should exclude '{cold_vm}'"

    print(f"\n  ✓ NE operators (!= and <>) both excluded '{cold_vm}'")


@pytest.mark.order(after="test_query_ne")
async def test_query_lt_le(mcp_session):
    """Less-than (<) and less-than-or-equal (<=) operators."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where cpuCount < 1000",
        },
    })
    vms_lt = result.get("data", [])
    assert len(vms_lt) >= 1, "< 1000 should return VMs"
    for vm in vms_lt:
        assert vm.get("cpuCount", 0) < 1000

    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where cpuCount <= 1",
        },
    })
    vms_le = result.get("data", [])
    for vm in vms_le:
        assert vm.get("cpuCount", 0) <= 1

    print(f"\n  ✓ LT (<) returned {len(vms_lt)} VMs; LE (<=) returned {len(vms_le)} VMs")


@pytest.mark.order(after="test_query_lt_le")
async def test_query_gt_ge(mcp_session):
    """Greater-than (>) and greater-than-or-equal (>=) operators."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where cpuCount > 0",
        },
    })
    vms_gt = result.get("data", [])
    assert len(vms_gt) >= 1, "> 0 should return VMs"
    for vm in vms_gt:
        assert vm.get("cpuCount", 0) > 0

    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where cpuCount >= 1",
        },
    })
    vms_ge = result.get("data", [])
    assert len(vms_ge) >= 1, ">= 1 should return VMs"
    for vm in vms_ge:
        assert vm.get("cpuCount", 0) >= 1

    print(f"\n  ✓ GT (>) returned {len(vms_gt)} VMs; GE (>=) returned {len(vms_ge)} VMs")


# -------------------------------------------------------------------
# TSL Query Tests — Regex Operators
# -------------------------------------------------------------------
@pytest.mark.order(after="test_query_gt_ge")
async def test_query_regex_match(mcp_session):
    """Regex match operator (~=)."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where name ~= '^mtv-rhel8-.*sanity$'",
        },
    })
    vms = result.get("data", [])
    assert len(vms) >= 1, "~= should match at least one VM"
    for vm in vms:
        name = vm.get("name", "")
        assert name.startswith("mtv-rhel8-") and name.endswith("sanity")

    print(f"\n  ✓ Regex match (~=) returned {len(vms)} VMs")


@pytest.mark.order(after="test_query_regex_match")
async def test_query_regex_not_match(mcp_session):
    """Regex NOT match operator (~!)."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where name ~= '^mtv-rhel8' and name ~! 'warm'",
        },
    })
    vms = result.get("data", [])
    assert len(vms) >= 1, "~! should return at least one VM"
    for vm in vms:
        assert "warm" not in vm.get("name", "")

    print(f"\n  ✓ Regex NOT (~!) returned {len(vms)} VMs excluding 'warm'")


# -------------------------------------------------------------------
# TSL Query Tests — String Operators (like, ilike)
# -------------------------------------------------------------------
@pytest.mark.order(after="test_query_regex_not_match")
async def test_query_like(mcp_session):
    """LIKE operator with SQL wildcards (% and _)."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where name like 'mtv-rhel8-%sanity'",
        },
    })
    vms = result.get("data", [])
    assert len(vms) >= 1, "LIKE should match at least one VM"
    for vm in vms:
        name = vm.get("name", "")
        assert name.startswith("mtv-rhel8-") and name.endswith("sanity")

    print(f"\n  ✓ LIKE returned {len(vms)} VMs matching 'mtv-rhel8-%sanity'")


@pytest.mark.order(after="test_query_like")
async def test_query_ilike(mcp_session):
    """ILIKE operator (case-insensitive LIKE)."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where name ilike 'MTV-RHEL8-%SANITY'",
        },
    })
    vms = result.get("data", [])
    assert len(vms) >= 1, "ILIKE should match case-insensitively"
    for vm in vms:
        name = vm.get("name", "").lower()
        assert name.startswith("mtv-rhel8-") and name.endswith("sanity")

    print(f"\n  ✓ ILIKE returned {len(vms)} VMs (case-insensitive match)")


# -------------------------------------------------------------------
# TSL Query Tests — Set/Range Operators (in, between)
# -------------------------------------------------------------------
@pytest.mark.order(after="test_query_ilike")
async def test_query_in(mcp_session):
    """IN operator with value list."""
    cold_vm = COLD_VMS.split(",")[0]
    warm_vm = WARM_VMS.split(",")[0]

    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": f"where name in ['{cold_vm}', '{warm_vm}']",
        },
    })
    vms = result.get("data", [])
    assert len(vms) == 2, f"IN should return exactly 2 VMs, got {len(vms)}"
    vm_names = {v.get("name") for v in vms}
    assert cold_vm in vm_names and warm_vm in vm_names

    print(f"\n  ✓ IN operator returned exactly 2 VMs: {sorted(vm_names)}")


@pytest.mark.order(after="test_query_in")
async def test_query_between(mcp_session):
    """BETWEEN operator for numeric range."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where cpuCount between 1 and 64",
        },
    })
    vms = result.get("data", [])
    assert len(vms) >= 1, "BETWEEN 1 and 64 should return VMs"
    for vm in vms:
        cpu = vm.get("cpuCount", 0)
        assert 1 <= cpu <= 64, f"cpuCount={cpu} not in [1, 64]"

    print(f"\n  ✓ BETWEEN returned {len(vms)} VMs with cpuCount in [1, 64]")


# -------------------------------------------------------------------
# TSL Query Tests — Logical Operators (and, or, not) and brackets
# -------------------------------------------------------------------
@pytest.mark.order(after="test_query_between")
async def test_query_and(mcp_session):
    """AND logical operator."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where name ~= '^mtv-rhel8' and cpuCount >= 1",
        },
    })
    vms = result.get("data", [])
    assert len(vms) >= 1, "AND should return VMs"
    for vm in vms:
        assert vm.get("name", "").startswith("mtv-rhel8")
        assert vm.get("cpuCount", 0) >= 1

    print(f"\n  ✓ AND returned {len(vms)} VMs satisfying both conditions")


@pytest.mark.order(after="test_query_and")
async def test_query_or(mcp_session):
    """OR logical operator."""
    cold_vm = COLD_VMS.split(",")[0]
    warm_vm = WARM_VMS.split(",")[0]

    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": f"where name = '{cold_vm}' or name = '{warm_vm}'",
        },
    })
    vms = result.get("data", [])
    assert len(vms) == 2, f"OR should return 2 VMs, got {len(vms)}"
    vm_names = {v.get("name") for v in vms}
    assert cold_vm in vm_names and warm_vm in vm_names

    print(f"\n  ✓ OR returned 2 VMs: {sorted(vm_names)}")


@pytest.mark.order(after="test_query_or")
async def test_query_not(mcp_session):
    """NOT logical operator."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where not (name ~= 'warm')",
        },
    })
    vms = result.get("data", [])
    assert len(vms) >= 1, "NOT should return VMs"
    for vm in vms:
        assert "warm" not in vm.get("name", "")

    print(f"\n  ✓ NOT returned {len(vms)} VMs excluding 'warm'")


@pytest.mark.order(after="test_query_not")
async def test_query_brackets_precedence(mcp_session):
    """Brackets to control operator precedence: (A or B) and C."""
    cold_vm = COLD_VMS.split(",")[0]
    warm_vm = WARM_VMS.split(",")[0]

    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": (
                f"where (name = '{cold_vm}' or name = '{warm_vm}') "
                f"and cpuCount >= 1"
            ),
        },
    })
    vms = result.get("data", [])
    assert len(vms) >= 1, "Bracketed OR + AND should return VMs"
    vm_names = {v.get("name") for v in vms}
    assert vm_names <= {cold_vm, warm_vm}, (
        f"Results should only contain the two target VMs, got {vm_names}"
    )

    print(f"\n  ✓ Brackets precedence returned {len(vms)} VMs: {sorted(vm_names)}")


# -------------------------------------------------------------------
# TSL Query Tests — Unary Functions (len, not)
# -------------------------------------------------------------------
@pytest.mark.order(after="test_query_brackets_precedence")
async def test_query_len_disks(mcp_session):
    """LEN function on array field (disks)."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where len(disks) >= 1",
        },
    })
    vms = result.get("data", [])
    assert len(vms) >= 1, "len(disks) >= 1 should return VMs"
    for vm in vms:
        disks = vm.get("disks", [])
        assert len(disks) >= 1, (
            f"VM '{vm.get('name')}' has {len(disks)} disks, expected >= 1"
        )

    print(f"\n  ✓ len(disks) >= 1 returned {len(vms)} VMs")


@pytest.mark.order(after="test_query_len_disks")
async def test_query_len_concerns(mcp_session):
    """LEN function on concerns array."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where len(concerns) >= 0",
        },
    })
    vms = result.get("data", [])
    assert len(vms) >= 1, "len(concerns) >= 0 should return all VMs"

    print(f"\n  ✓ len(concerns) >= 0 returned {len(vms)} VMs")


# -------------------------------------------------------------------
# TSL Query Tests — Deep Field Access (dot notation, wildcards)
# -------------------------------------------------------------------
@pytest.mark.order(after="test_query_len_concerns")
async def test_query_deep_field_dot(mcp_session):
    """Deep field access using dot notation (disks.capacity)."""
    # Use the 2-disks VM for a guaranteed multi-disk target
    warm_2disks = [v for v in WARM_VMS.split(",") if "2disks" in v]
    vm_name = warm_2disks[0] if warm_2disks else COLD_VMS.split(",")[0]

    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": f"where name = '{vm_name}'",
        },
    })
    vms = result.get("data", [])
    assert len(vms) == 1, f"Expected 1 VM for '{vm_name}'"
    disks = vms[0].get("disks", [])
    assert len(disks) >= 1, "VM should have at least 1 disk"

    # Now query using deep field: first disk capacity > 0
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where disks[0].capacity > 0",
        },
    })
    vms_deep = result.get("data", [])
    assert len(vms_deep) >= 1, "disks[0].capacity > 0 should match VMs"

    print(f"\n  ✓ Deep field disks[0].capacity > 0 returned {len(vms_deep)} VMs")


@pytest.mark.order(after="test_query_deep_field_dot")
async def test_query_deep_field_wildcard(mcp_session):
    """Deep field with wildcard [*] to access all array elements."""
    # Query VMs where any disk has capacity > 0 (using implicit wildcard)
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where any(disks[*].capacity > 0)",
        },
    })
    vms = result.get("data", [])
    assert len(vms) >= 1, "any(disks[*].capacity > 0) should match VMs"

    print(f"\n  ✓ Wildcard disks[*].capacity with any() returned {len(vms)} VMs")


# -------------------------------------------------------------------
# TSL Query Tests — ORDER BY and LIMIT
# -------------------------------------------------------------------
@pytest.mark.order(after="test_query_deep_field_wildcard")
async def test_query_order_by_limit(mcp_session):
    """ORDER BY + LIMIT."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where name ~= '^mtv-' order by name limit 3",
        },
    })
    vms = result.get("data", [])
    assert 1 <= len(vms) <= 3, f"LIMIT 3 should return 1-3 VMs, got {len(vms)}"
    names = [v.get("name", "") for v in vms]
    assert names == sorted(names), f"Should be sorted ascending: {names}"

    print(f"\n  ✓ ORDER BY name LIMIT 3 returned: {names}")


@pytest.mark.order(after="test_query_order_by_limit")
async def test_query_order_by_desc(mcp_session):
    """ORDER BY DESC."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get inventory vm",
        "flags": {
            "provider": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
            "query": "where name ~= '^mtv-rhel8' order by name desc",
        },
    })
    vms = result.get("data", [])
    assert len(vms) >= 2, "Should return multiple VMs for sorting"
    names = [v.get("name", "") for v in vms]
    assert names == sorted(names, reverse=True), (
        f"Should be sorted descending: {names}"
    )

    print(f"\n  ✓ ORDER BY name DESC returned: {names}")


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
