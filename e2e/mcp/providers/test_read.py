"""Providers · read -- list, get details, and verify YAML of providers."""

import pytest
import yaml

from conftest import (
    GOVC_URL,
    OCP_PROVIDER_NAME,
    TEST_NAMESPACE,
    VSPHERE_PROVIDER_NAME,
    VSPHERE_URL,
    call_tool,
)


@pytest.mark.order(20)
async def test_get_providers_list(mcp_session):
    """List providers in the test namespace -- expect at least 2."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get provider",
        "flags": {"namespace": TEST_NAMESPACE, "output": "json"},
    })

    providers = result.get("data", [])
    assert isinstance(providers, list)
    assert len(providers) >= 2, f"Expected >=2 providers, got {len(providers)}"

    names = {
        p.get("name") or p.get("metadata", {}).get("name")
        for p in providers
    }
    assert VSPHERE_PROVIDER_NAME in names
    assert OCP_PROVIDER_NAME in names


@pytest.mark.order(21)
async def test_get_vsphere_provider_details(mcp_session):
    """Verify the vSphere provider JSON has correct type and URL."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get provider",
        "flags": {
            "name": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
        },
    })

    data = result.get("data")
    provider = data[0] if isinstance(data, list) else data

    spec = provider.get("spec") or provider.get("object", {}).get("spec", {})
    assert spec.get("type") == "vsphere"
    # The stored URL should contain the hostname from GOVC_URL
    spec_url = spec.get("url", "") or ""
    assert GOVC_URL in spec_url or VSPHERE_URL in spec_url, (
        f"vSphere URL mismatch: spec.url={spec_url!r}, expected to contain {GOVC_URL!r}"
    )

    status = provider.get("status") or provider.get("object", {}).get("status", {})
    assert status.get("phase") == "Ready"


@pytest.mark.order(22)
async def test_get_ocp_provider_details(mcp_session):
    """Verify the OpenShift provider JSON."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get provider",
        "flags": {
            "name": OCP_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "json",
        },
    })

    data = result.get("data")
    provider = data[0] if isinstance(data, list) else data

    spec = provider.get("spec") or provider.get("object", {}).get("spec", {})
    assert spec.get("type") == "openshift"

    status = provider.get("status") or provider.get("object", {}).get("status", {})
    assert status.get("phase") == "Ready"


@pytest.mark.order(23)
async def test_verify_vsphere_provider_yaml(mcp_session):
    """Fetch the vSphere provider as YAML and verify key fields."""
    result = await call_tool(mcp_session, "mtv_read", {
        "command": "get provider",
        "flags": {
            "name": VSPHERE_PROVIDER_NAME,
            "namespace": TEST_NAMESPACE,
            "output": "yaml",
        },
    })

    raw = result.get("output", "")
    assert raw, "Expected YAML output, got empty string"

    docs = list(yaml.safe_load_all(raw))
    assert len(docs) >= 1, "Expected at least one YAML document"

    doc = docs[0]
    # The YAML document might be a list of resources; unwrap if needed
    if isinstance(doc, list):
        assert len(doc) >= 1, "Expected at least one resource in YAML list"
        doc = doc[0]

    spec = doc.get("spec") or doc.get("object", {}).get("spec", {})
    assert spec.get("type") == "vsphere", f"Expected type=vsphere, got {spec.get('type')}"
    spec_url = spec.get("url", "") or ""
    assert GOVC_URL in spec_url or VSPHERE_URL in spec_url, (
        f"vSphere URL not found in spec: {spec_url!r}"
    )
