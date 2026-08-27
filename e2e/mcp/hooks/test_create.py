"""Hooks · write -- create, get, describe, and delete a local hook."""

import pytest

from conftest import HOOK_NAME, TEST_NAMESPACE, call_tool


@pytest.mark.order(19)
async def test_create_hook(mcp_session):
    """Create a local hook using the default hook-runner image."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "create hook",
        "flags": {
            "name": HOOK_NAME,
            "namespace": TEST_NAMESPACE,
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"
    print(f"\n  ✓ Created hook '{HOOK_NAME}'")


@pytest.mark.order(55)
async def test_get_hook(mcp_session):
    """List hooks and verify the test hook exists."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get hook",
        "flags": {"namespace": TEST_NAMESPACE, "output": "json"},
    })
    data = result.get("data", [])
    hooks = data if isinstance(data, list) else [data]
    names = {
        h.get("name") or h.get("metadata", {}).get("name", "")
        for h in hooks
    }
    assert HOOK_NAME in names, f"Hook '{HOOK_NAME}' not in {names}"
    print(f"\n  ✓ Hook '{HOOK_NAME}' listed")


@pytest.mark.order(56)
async def test_describe_hook(mcp_session):
    """Describe the test hook."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "describe hook",
        "flags": {"name": HOOK_NAME, "namespace": TEST_NAMESPACE},
    })
    output = result.get("output", "")
    assert HOOK_NAME in output, f"Hook name missing from describe: {output[:200]}"
    print(f"\n  ✓ Described hook '{HOOK_NAME}'")


@pytest.mark.order(57)
async def test_delete_hook(mcp_session):
    """Delete the test hook and confirm it is gone."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "delete hook",
        "flags": {"name": HOOK_NAME, "namespace": TEST_NAMESPACE},
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"

    got = await call_tool(mcp_session, "mtv_read", {
        "command": "get hook",
        "flags": {"namespace": TEST_NAMESPACE, "output": "json"},
    })
    data = got.get("data", [])
    hooks = data if isinstance(data, list) else [data]
    names = {
        h.get("name") or h.get("metadata", {}).get("name", "")
        for h in hooks if h
    }
    assert HOOK_NAME not in names, f"Hook '{HOOK_NAME}' still present: {names}"
    print(f"\n  ✓ Deleted hook '{HOOK_NAME}'")
