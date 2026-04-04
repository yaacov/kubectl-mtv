"""Dry-run · show_cli -- verify show_cli returns CLI commands without executing."""

import pytest

from conftest import (
    TEST_NAMESPACE,
    call_tool,
)


@pytest.mark.order(5)
async def test_show_cli_read(mcp_session):
    """show_cli on mtv_read should return the equivalent CLI command."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get plan",
        "flags": {"namespace": TEST_NAMESPACE},
        "show_cli": True,
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"

    output = result.get("output", "")
    assert "kubectl-mtv" in output, f"Expected 'kubectl-mtv' in output: {output}"
    assert "get" in output, f"Expected 'get' in output: {output}"
    assert "plan" in output, f"Expected 'plan' in output: {output}"
    print(f"\n  show_cli read output: {output}")


@pytest.mark.order(5)
async def test_show_cli_write(mcp_session):
    """show_cli on mtv_write should return the equivalent CLI command."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "create provider",
        "flags": {
            "name": "show-cli-test",
            "type": "openshift",
            "namespace": TEST_NAMESPACE,
        },
        "show_cli": True,
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"

    output = result.get("output", "")
    assert "kubectl-mtv" in output, f"Expected 'kubectl-mtv' in output: {output}"
    assert "create" in output, f"Expected 'create' in output: {output}"
    assert "provider" in output, f"Expected 'provider' in output: {output}"
    print(f"\n  show_cli write output: {output}")


@pytest.mark.order(5)
async def test_show_cli_not_executed(mcp_session):
    """show_cli should not execute the command -- no data field in response."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get plan",
        "flags": {"namespace": TEST_NAMESPACE},
        "show_cli": True,
    })
    assert result.get("return_value") == 0
    assert "data" not in result or result.get("data") is None, (
        "show_cli should not return data (command not executed)"
    )
    print("\n  Confirmed: show_cli does not execute the command")
