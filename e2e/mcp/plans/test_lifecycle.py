"""Plans · write -- start --dry-run and archive/unarchive (no real migration)."""

import json
import subprocess

import pytest

from conftest import (
    COLD_PLAN_NAME,
    TEST_NAMESPACE,
    _kubectl_base_args,
    _kubectl_wait,
    call_tool,
)


@pytest.mark.order(47)
async def test_start_plan_dry_run(mcp_session):
    """start plan --dry-run should emit a Migration CR without creating one."""
    _kubectl_wait(
        f"plans.forklift.konveyor.io/{COLD_PLAN_NAME}",
        "condition=Ready",
        namespace=TEST_NAMESPACE,
        timeout=300,
    )

    result = await call_tool(mcp_session, "mtv_write", {
        "command": "start plan",
        "flags": {
            "name": COLD_PLAN_NAME,
            "namespace": TEST_NAMESPACE,
            "dry_run": True,
            "output": "json",
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"

    parsed = result.get("data")
    output = result.get("output", "")
    blob = json.dumps(parsed) if parsed is not None else output
    assert "Migration" in blob, f"Expected Migration CR in dry-run output: {blob[:300]}"
    assert "forklift.konveyor.io" in blob, (
        f"Expected API group in dry-run output: {blob[:300]}"
    )

    listed = subprocess.run(
        _kubectl_base_args() + [
            "get", "migrations.forklift.konveyor.io",
            "-n", TEST_NAMESPACE,
            "-o", "json",
        ],
        capture_output=True, text=True, timeout=30,
    )
    assert listed.returncode == 0, listed.stderr
    items = json.loads(listed.stdout).get("items") or []
    assert items == [], (
        f"dry-run must not create Migration CRs, found {[i.get('metadata', {}).get('name') for i in items]}"
    )
    print("\n  ✓ start plan --dry-run emitted Migration YAML/JSON and created none")


@pytest.mark.order(48)
async def test_archive_and_unarchive_plan(mcp_session):
    """Archive then unarchive a plan that was never started."""
    archived = await call_tool(mcp_session, "mtv_write", {
        "command": "archive plan",
        "flags": {
            "name": COLD_PLAN_NAME,
            "namespace": TEST_NAMESPACE,
        },
    })
    assert archived.get("return_value") == 0, f"archive failed: {archived}"

    got = await call_tool(mcp_session, "mtv_read", {
        "command": "get plan",
        "flags": {
            "name": COLD_PLAN_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
        },
    })
    data = got.get("data")
    plan = data[0] if isinstance(data, list) else data
    obj = plan.get("object") if isinstance(plan, dict) else {}
    spec = (obj.get("spec") if isinstance(obj, dict) else None) or (
        plan.get("spec") if isinstance(plan, dict) else {}
    ) or {}
    assert spec.get("archived") is True, f"Expected spec.archived=true, got {spec.get('archived')}"

    restored = await call_tool(mcp_session, "mtv_write", {
        "command": "unarchive plan",
        "flags": {
            "name": COLD_PLAN_NAME,
            "namespace": TEST_NAMESPACE,
        },
    })
    assert restored.get("return_value") == 0, f"unarchive failed: {restored}"

    got = await call_tool(mcp_session, "mtv_read", {
        "command": "get plan",
        "flags": {
            "name": COLD_PLAN_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
        },
    })
    data = got.get("data")
    plan = data[0] if isinstance(data, list) else data
    obj = plan.get("object") if isinstance(plan, dict) else {}
    spec = (obj.get("spec") if isinstance(obj, dict) else None) or (
        plan.get("spec") if isinstance(plan, dict) else {}
    ) or {}
    assert not spec.get("archived"), f"Expected spec.archived false/absent, got {spec.get('archived')}"
    print("\n  ✓ Archived and unarchived cold plan")
