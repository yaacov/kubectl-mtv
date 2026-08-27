"""Plans · write -- patch plan and planvm without starting a migration."""

import pytest

from conftest import COLD_PLAN_NAME, COLD_VMS, TEST_NAMESPACE, call_tool


def _plan_spec(result):
    """Return Plan spec from get plan JSON (flattened view wraps the CR in object)."""
    data = result.get("data")
    plan = data[0] if isinstance(data, list) else data
    if not isinstance(plan, dict):
        return {}
    obj = plan.get("object") if isinstance(plan.get("object"), dict) else plan
    return obj.get("spec") or plan.get("spec") or {}


@pytest.mark.order(45)
async def test_patch_plan_preserve_static_ips(mcp_session):
    """Set --preserve-static-ips true then false on the cold plan."""
    for value in (True, False):
        result = await call_tool(mcp_session, "mtv_write", {
            "command": "patch plan",
            "flags": {
                "plan-name": COLD_PLAN_NAME,
                "namespace": TEST_NAMESPACE,
                "preserve-static-ips": value,
            },
        })
        assert result.get("return_value") == 0, f"patch preserve-static-ips={value}: {result}"

        got = await call_tool(mcp_session, "mtv_read", {
            "command": "get plan",
            "flags": {
                "name": COLD_PLAN_NAME,
                "namespace": TEST_NAMESPACE,
                "output": "json",
            },
        })
        spec = _plan_spec(got)
        actual = spec.get("preserveStaticIPs")
        assert actual is value, (
            f"preserveStaticIPs expected {value}, got {actual}"
        )

    print("\n  ✓ Patched cold plan preserveStaticIPs true then false")


@pytest.mark.order(46)
async def test_patch_planvm_target_name(mcp_session):
    """Set a custom --target-name on the first cold-plan VM."""
    vm_name = COLD_VMS.split(",")[0].strip()
    target_name = f"{vm_name}-e2e-target"

    result = await call_tool(mcp_session, "mtv_write", {
        "command": "patch planvm",
        "flags": {
            "plan-name": COLD_PLAN_NAME,
            "vm-name": vm_name,
            "namespace": TEST_NAMESPACE,
            "target-name": target_name,
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"

    got = await call_tool(mcp_session, "mtv_read", {
        "command": "get plan",
        "flags": {
            "name": COLD_PLAN_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
        },
    })
    vms = _plan_spec(got).get("vms") or []
    match = next(
        (vm for vm in vms if vm.get("name") == vm_name or vm.get("id") == vm_name),
        None,
    )
    assert match is not None, f"VM '{vm_name}' not found in plan vms={vms}"
    assert match.get("targetName") == target_name, (
        f"targetName expected {target_name}, got {match.get('targetName')}"
    )
    print(f"\n  ✓ Patched planvm '{vm_name}' targetName={target_name}")
