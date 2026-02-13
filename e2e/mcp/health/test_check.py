"""Health · read -- verify the MTV health report."""

import pytest

from conftest import (
    OCP_PROVIDER_NAME,
    TEST_NAMESPACE,
    VSPHERE_PROVIDER_NAME,
    call_tool,
)


@pytest.mark.order(70)
async def test_health_report(mcp_session):
    """Run the health command scoped to the test namespace."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "health",
        "flags": {"namespace": TEST_NAMESPACE, "skip-logs": True},
    })

    output = result.get("output", "")
    assert result.get("return_value") == 0, f"Health check failed: {result}"
    assert "HEALTH REPORT" in output, f"Expected 'HEALTH REPORT' in output: {output[:200]}"
    assert VSPHERE_PROVIDER_NAME in output or OCP_PROVIDER_NAME in output, (
        "Health report should mention test providers"
    )
