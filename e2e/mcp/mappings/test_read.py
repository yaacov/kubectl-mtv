"""Mappings · read -- verify auto-generated and standalone network/storage mappings."""

import pytest

from conftest import NETWORK_MAPPING_NAME, STORAGE_MAPPING_NAME, TEST_NAMESPACE, call_tool


def _mapping_names(result):
    data = result.get("data", [])
    mappings = data if isinstance(data, list) else [data]
    return {
        m.get("name") or m.get("metadata", {}).get("name", "")
        for m in mappings if m
    }


@pytest.mark.order(50)
async def test_get_network_mappings(mcp_session):
    """Verify network mappings exist, including the standalone create."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get mapping network",
        "flags": {"namespace": TEST_NAMESPACE, "output": "json"},
    })

    names = _mapping_names(result)
    assert len(names) >= 1, "Expected at least 1 network mapping"
    assert NETWORK_MAPPING_NAME in names, (
        f"Standalone mapping '{NETWORK_MAPPING_NAME}' not in {names}"
    )


@pytest.mark.order(51)
async def test_get_storage_mappings(mcp_session):
    """Verify storage mappings exist, including the standalone create."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get mapping storage",
        "flags": {"namespace": TEST_NAMESPACE, "output": "json"},
    })

    names = _mapping_names(result)
    assert len(names) >= 1, "Expected at least 1 storage mapping"
    assert STORAGE_MAPPING_NAME in names, (
        f"Standalone mapping '{STORAGE_MAPPING_NAME}' not in {names}"
    )


@pytest.mark.order(52)
async def test_describe_network_mapping(mcp_session):
    """Describe the standalone network mapping."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "describe mapping network",
        "flags": {
            "name": NETWORK_MAPPING_NAME,
            "namespace": TEST_NAMESPACE,
        },
    })
    output = result.get("output", "")
    assert NETWORK_MAPPING_NAME in output, (
        f"Mapping name missing from describe: {output[:200]}"
    )


@pytest.mark.order(53)
async def test_describe_storage_mapping(mcp_session):
    """Describe the standalone storage mapping."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "describe mapping storage",
        "flags": {
            "name": STORAGE_MAPPING_NAME,
            "namespace": TEST_NAMESPACE,
        },
    })
    output = result.get("output", "")
    assert STORAGE_MAPPING_NAME in output, (
        f"Mapping name missing from describe: {output[:200]}"
    )

