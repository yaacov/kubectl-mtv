"""Plans · write -- create cold and warm migration plans."""

import pytest

from conftest import (
    COLD_PLAN_NAME,
    COLD_VMS,
    NETWORK_PAIRS,
    OCP_PROVIDER_NAME,
    STORAGE_PAIRS,
    TEST_NAMESPACE,
    VSPHERE_PROVIDER_NAME,
    WARM_PLAN_NAME,
    WARM_VMS,
    call_tool,
)


@pytest.mark.order(15)
async def test_create_cold_plan(mcp_session):
    """Create a cold migration plan with two VMs."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "create plan",
        "flags": {
            "name": COLD_PLAN_NAME,
            "source": VSPHERE_PROVIDER_NAME,
            "target": OCP_PROVIDER_NAME,
            "vms": COLD_VMS,
            "network-pairs": NETWORK_PAIRS,
            "storage-pairs": STORAGE_PAIRS,
            "namespace": TEST_NAMESPACE,
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"
    print(f"\n  ✓ Created cold plan '{COLD_PLAN_NAME}' with VMs: {COLD_VMS}")


@pytest.mark.order(16)
async def test_create_warm_plan(mcp_session):
    """Create a warm migration plan with two VMs."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "create plan",
        "flags": {
            "name": WARM_PLAN_NAME,
            "source": VSPHERE_PROVIDER_NAME,
            "target": OCP_PROVIDER_NAME,
            "vms": WARM_VMS,
            "migration-type": "warm",
            "network-pairs": NETWORK_PAIRS,
            "storage-pairs": STORAGE_PAIRS,
            "namespace": TEST_NAMESPACE,
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"
    print(f"\n  ✓ Created warm plan '{WARM_PLAN_NAME}' with VMs: {WARM_VMS}")