"""Mappings · read -- verify auto-generated network and storage mappings."""

import pytest

from conftest import TEST_NAMESPACE, call_tool


@pytest.mark.order(50)
async def test_get_network_mappings(mcp_session):
    """Verify auto-generated network mappings exist."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get mapping network",
        "flags": {"namespace": TEST_NAMESPACE, "output": "json"},
    })

    data = result.get("data", [])
    mappings = data if isinstance(data, list) else [data]
    assert len(mappings) >= 1, "Expected at least 1 network mapping"


@pytest.mark.order(51)
async def test_get_storage_mappings(mcp_session):
    """Verify auto-generated storage mappings exist."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get mapping storage",
        "flags": {"namespace": TEST_NAMESPACE, "output": "json"},
    })

    data = result.get("data", [])
    mappings = data if isinstance(data, list) else [data]
    assert len(mappings) >= 1, "Expected at least 1 storage mapping"
