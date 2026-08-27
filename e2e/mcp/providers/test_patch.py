"""Providers · write -- patch a provider setting without breaking inventory."""

import pytest

from conftest import TEST_NAMESPACE, VSPHERE_PROVIDER_NAME, call_tool


@pytest.mark.order(24)
async def test_patch_vsphere_vddk_aio(mcp_session):
    """Toggle --use-vddk-aio-optimization on the vSphere provider."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "patch provider",
        "flags": {
            "name": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "use-vddk-aio-optimization": True,
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"

    got = await call_tool(mcp_session, "mtv_read", {
        "command": "get provider",
        "flags": {
            "name": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
        },
    })
    data = got.get("data")
    provider = data[0] if isinstance(data, list) else data
    spec = provider.get("spec") or provider.get("object", {}).get("spec", {})
    settings = spec.get("settings") or {}
    aio = str(settings.get("useVddkAioOptimization", "")).lower()
    assert aio in ("true", "1"), (
        f"Expected useVddkAioOptimization true, got settings={settings}"
    )
    print("\n  ✓ Patched vSphere provider useVddkAioOptimization=true")
