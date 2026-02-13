"""Plans · read -- list, describe, and verify YAML of migration plans."""

import pytest
import yaml

from conftest import (
    COLD_PLAN_NAME,
    COLD_VMS,
    OCP_PROVIDER_NAME,
    TEST_NAMESPACE,
    VSPHERE_PROVIDER_NAME,
    WARM_PLAN_NAME,
    WARM_VMS,
    call_tool,
)


@pytest.mark.order(40)
async def test_get_plans(mcp_session):
    """List plans in the test namespace -- expect 2."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get plan",
        "flags": {"namespace": TEST_NAMESPACE, "output": "json"},
    })

    plans = result.get("data", [])
    # Handle the "No plans found" message dict gracefully
    if isinstance(plans, dict):
        plans = []
    assert isinstance(plans, list), (
        f"Expected list of plans, got {type(plans).__name__}: {plans}"
    )
    assert len(plans) == 2, f"Expected 2 plans, got {len(plans)}"


@pytest.mark.order(41)
async def test_describe_cold_plan(mcp_session):
    """Describe the cold plan including VMs."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "describe plan",
        "flags": {
            "name": COLD_PLAN_NAME,
            "namespace": TEST_NAMESPACE,
            "with-vms": True,
        },
    })

    output = result.get("output", "")
    assert VSPHERE_PROVIDER_NAME in output, "Source provider missing"
    for vm in COLD_VMS.split(","):
        assert vm in output, f"VM '{vm}' missing from cold plan description"


@pytest.mark.order(42)
async def test_describe_warm_plan(mcp_session):
    """Describe the warm plan including VMs and warm flag."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "describe plan",
        "flags": {
            "name": WARM_PLAN_NAME,
            "namespace": TEST_NAMESPACE,
            "with-vms": True,
        },
    })

    output = result.get("output", "")
    assert "warm" in output.lower() or "Warm" in output, "Warm indicator missing"
    for vm in WARM_VMS.split(","):
        assert vm in output, f"VM '{vm}' missing from warm plan description"


@pytest.mark.order(43)
async def test_verify_cold_plan_yaml(mcp_session):
    """Fetch the cold plan as YAML and verify structure."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get plan",
        "flags": {
            "name": COLD_PLAN_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "yaml",
        },
    })

    raw = result.get("output", "")
    assert raw, "Expected YAML output, got empty string"

    docs = list(yaml.safe_load_all(raw))
    doc = docs[0]
    if isinstance(doc, list):
        assert len(doc) >= 1, "Expected at least one resource in YAML list"
        doc = doc[0]
    spec = doc.get("spec") or doc.get("object", {}).get("spec", {})

    src_name = spec.get("provider", {}).get("source", {}).get("name", "")
    tgt_name = spec.get("provider", {}).get("destination", {}).get("name", "")
    assert src_name == VSPHERE_PROVIDER_NAME, f"Source provider mismatch: {src_name}"
    assert tgt_name == OCP_PROVIDER_NAME, f"Target provider mismatch: {tgt_name}"

    vms = spec.get("vms", [])
    expected_cold = len(COLD_VMS.split(","))
    assert len(vms) == expected_cold, f"Expected {expected_cold} VMs in cold plan, got {len(vms)}"
    assert spec.get("warm", False) is False, "Cold plan should not be warm"


@pytest.mark.order(44)
async def test_verify_warm_plan_yaml(mcp_session):
    """Fetch the warm plan as YAML and verify warm flag and VMs."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get plan",
        "flags": {
            "name": WARM_PLAN_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "yaml",
        },
    })

    raw = result.get("output", "")
    assert raw, "Expected YAML output, got empty string"

    docs = list(yaml.safe_load_all(raw))
    doc = docs[0]
    if isinstance(doc, list):
        assert len(doc) >= 1, "Expected at least one resource in YAML list"
        doc = doc[0]
    spec = doc.get("spec") or doc.get("object", {}).get("spec", {})

    assert spec.get("warm") is True, "Warm plan should have warm=true"

    vms = spec.get("vms", [])
    expected_warm = len(WARM_VMS.split(","))
    assert len(vms) == expected_warm, f"Expected {expected_warm} VMs in warm plan, got {len(vms)}"
