"""Dry-run · dry_run -- verify --dry-run outputs resource YAML/JSON without creating."""

import pytest

from conftest import (
    COLD_VMS,
    NETWORK_PAIRS,
    OCP_PROVIDER_NAME,
    STORAGE_PAIRS,
    TEST_NAMESPACE,
    VSPHERE_PROVIDER_NAME,
    call_tool,
)

DRY_RUN_PLAN_NAME = "dryrun-e2e-plan"
DRY_RUN_HOOK_NAME = "dryrun-e2e-hook"


@pytest.mark.order(14)
async def test_dry_run_create_provider(mcp_session):
    """dry-run create provider should output Provider YAML without creating it."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "create provider",
        "flags": {
            "name": "dryrun-ocp-provider",
            "type": "openshift",
            "namespace": TEST_NAMESPACE,
            "dry_run": True,
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"

    output = result.get("output", "")
    assert "kind" in output.lower(), f"Expected 'kind' in output: {output[:200]}"
    assert "Provider" in output, f"Expected 'Provider' in output: {output[:200]}"
    print(f"\n  dry-run create provider output length: {len(output)} chars")


@pytest.mark.order(14)
async def test_dry_run_create_hook(mcp_session):
    """dry-run create hook should output Hook YAML without creating it."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "create hook",
        "flags": {
            "name": DRY_RUN_HOOK_NAME,
            "namespace": TEST_NAMESPACE,
            "dry_run": True,
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"

    output = result.get("output", "")
    assert "Hook" in output, f"Expected 'Hook' in output: {output[:200]}"
    assert DRY_RUN_HOOK_NAME in output, f"Expected hook name in output: {output[:200]}"
    print(f"\n  dry-run create hook output length: {len(output)} chars")


@pytest.mark.order(14)
async def test_dry_run_create_plan(mcp_session):
    """dry-run create plan should output Plan YAML without creating it."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "create plan",
        "flags": {
            "name": DRY_RUN_PLAN_NAME,
            "source": VSPHERE_PROVIDER_NAME,
            "target": OCP_PROVIDER_NAME,
            "vms": COLD_VMS,
            "network-pairs": NETWORK_PAIRS,
            "storage-pairs": STORAGE_PAIRS,
            "namespace": TEST_NAMESPACE,
            "dry_run": True,
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"

    output = result.get("output", "")
    assert "Plan" in output, f"Expected 'Plan' in output: {output[:200]}"
    assert "forklift.konveyor.io" in output, (
        f"Expected API group in output: {output[:200]}"
    )
    print(f"\n  dry-run create plan output length: {len(output)} chars")


@pytest.mark.order(14)
async def test_dry_run_create_hook_json(mcp_session):
    """dry-run create hook with --output json should return valid JSON.

    Uses a single-resource command (hook) so the CLI stdout is one JSON
    object.  UnmarshalJSONResponse parses it into the 'data' field.
    """
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "create hook",
        "flags": {
            "name": "dryrun-json-hook",
            "namespace": TEST_NAMESPACE,
            "dry_run": True,
            "output": "json",
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"

    # Single-resource JSON is parsed into 'data' by UnmarshalJSONResponse
    parsed = result.get("data")
    assert parsed is not None, (
        f"Expected 'data' with parsed JSON, got keys: {list(result.keys())}"
    )
    # Normalize: if UnmarshalJSONResponse returned a single-item list, unwrap it
    if isinstance(parsed, list):
        assert len(parsed) == 1, (
            f"Expected single-element list, got {len(parsed)} items"
        )
        parsed = parsed[0]
    assert isinstance(parsed, dict), f"Expected dict, got {type(parsed).__name__}"
    assert parsed.get("kind") == "Hook", f"Expected kind=Hook, got {parsed.get('kind')}"
    assert "forklift.konveyor.io" in parsed.get("apiVersion", ""), (
        f"Expected forklift API version: {parsed.get('apiVersion')}"
    )
    print(f"\n  dry-run JSON hook: kind={parsed['kind']}, apiVersion={parsed['apiVersion']}")


@pytest.mark.order(14)
async def test_dry_run_no_resource_created(mcp_session):
    """After dry-run create plan, the plan should NOT exist in the cluster."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get plan",
        "flags": {"namespace": TEST_NAMESPACE, "output": "json"},
    })

    assert result.get("return_value") == 0, f"get plan call failed: {result}"

    data = result.get("data", [])
    plans = data if isinstance(data, list) else [data]
    plan_names = [
        p.get("name") or p.get("metadata", {}).get("name", "")
        for p in plans
    ]
    assert DRY_RUN_PLAN_NAME not in plan_names, (
        f"Plan '{DRY_RUN_PLAN_NAME}' should NOT exist after dry-run, "
        f"but found in: {plan_names}"
    )
    print(f"\n  Confirmed: '{DRY_RUN_PLAN_NAME}' was not created in the cluster")
