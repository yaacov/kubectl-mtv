"""Mappings · write -- create standalone network and storage mappings."""

import pytest

from conftest import (
    NETWORK_MAPPING_NAME,
    NETWORK_PAIRS,
    OCP_PROVIDER_NAME,
    STORAGE_MAPPING_NAME,
    STORAGE_PAIRS,
    TEST_NAMESPACE,
    VSPHERE_PROVIDER_NAME,
    call_tool,
)


@pytest.mark.order(17)
async def test_create_network_mapping(mcp_session):
    """Create a network mapping from NETWORK_PAIRS."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "create mapping network",
        "flags": {
            "name": NETWORK_MAPPING_NAME,
            "source": VSPHERE_PROVIDER_NAME,
            "target": OCP_PROVIDER_NAME,
            "network-pairs": NETWORK_PAIRS,
            "namespace": TEST_NAMESPACE,
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"
    print(f"\n  ✓ Created network mapping '{NETWORK_MAPPING_NAME}'")


@pytest.mark.order(18)
async def test_create_storage_mapping(mcp_session):
    """Create a storage mapping from STORAGE_PAIRS."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "create mapping storage",
        "flags": {
            "name": STORAGE_MAPPING_NAME,
            "source": VSPHERE_PROVIDER_NAME,
            "target": OCP_PROVIDER_NAME,
            "storage-pairs": STORAGE_PAIRS,
            "namespace": TEST_NAMESPACE,
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"
    print(f"\n  ✓ Created storage mapping '{STORAGE_MAPPING_NAME}'")
