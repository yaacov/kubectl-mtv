"""Providers · write -- create OCP + vSphere providers, wait Ready + inventory."""

import pytest

from conftest import (
    GOVC_PASSWORD,
    GOVC_USERNAME,
    OCP_PROVIDER_NAME,
    TEST_NAMESPACE,
    VSPHERE_PROVIDER_NAME,
    VSPHERE_URL,
    _kubectl_wait,
    _mtv_base_args,
    _retry_command,
    call_tool,
)


@pytest.mark.order(10)
async def test_create_ocp_provider(mcp_session):
    """Create the OpenShift 'host' target provider (local cluster, no URL needed)."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "create provider",
        "flags": {
            "name": OCP_PROVIDER_NAME,
            "type": "openshift",
            "namespace": TEST_NAMESPACE,
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"
    print(f"\n  Created OpenShift provider '{OCP_PROVIDER_NAME}' (local cluster)")


@pytest.mark.order(11)
async def test_create_vsphere_provider(mcp_session):
    """Create the vSphere source provider with credentials from env."""
    result = await call_tool(mcp_session, "mtv_write", {
        "command": "create provider",
        "flags": {
            "name": VSPHERE_PROVIDER_NAME,
            "type": "vsphere",
            "url": VSPHERE_URL,
            "username": GOVC_USERNAME,
            "password": GOVC_PASSWORD,
            "namespace": TEST_NAMESPACE,
            "provider-insecure-skip-tls": True,
        },
    })
    assert result.get("return_value") == 0, f"Unexpected result: {result}"
    print(f"\n  Created vSphere provider '{VSPHERE_PROVIDER_NAME}' at {VSPHERE_URL}")


@pytest.mark.order(12)
async def test_wait_providers_ready(mcp_session):
    """Wait until both providers reach Ready phase using ``kubectl wait``."""
    _kubectl_wait(
        [
            f"providers.forklift.konveyor.io/{OCP_PROVIDER_NAME}",
            f"providers.forklift.konveyor.io/{VSPHERE_PROVIDER_NAME}",
        ],
        "jsonpath={.status.phase}=Ready",
        namespace=TEST_NAMESPACE,
        timeout=120,
    )
    print("\n  ✓ Both providers are Ready")


@pytest.mark.order(13)
async def test_wait_inventory_ready(mcp_session):
    """Wait until the inventory service can actually serve requests.

    ``kubectl wait --for=condition=InventoryCreated`` confirms the provider
    CR condition, then we retry a lightweight ``kubectl-mtv get inventory vm``
    until the inventory service has synced and can authenticate (no 401).
    """
    # 1. CR condition first (fast path -- usually already True)
    _kubectl_wait(
        f"providers.forklift.konveyor.io/{VSPHERE_PROVIDER_NAME}",
        "condition=InventoryCreated",
        namespace=TEST_NAMESPACE,
        timeout=180,
    )

    # 2. Verify the inventory service is actually serving VM requests
    _retry_command(
        _mtv_base_args() + [
            "get", "inventory", "vm",
            "--provider", VSPHERE_PROVIDER_NAME,
            "--namespace", TEST_NAMESPACE,
            "--output", "json",
        ],
        timeout=180,
        interval=10,
        description="vSphere VM inventory to respond",
    )
    print("\n  ✓ vSphere VM inventory is accessible")

    # 3. Verify host inventory is also available (needed by create host)
    _retry_command(
        _mtv_base_args() + [
            "get", "inventory", "host",
            "--provider", VSPHERE_PROVIDER_NAME,
            "--namespace", TEST_NAMESPACE,
            "--output", "json",
        ],
        timeout=180,
        interval=10,
        description="vSphere host inventory to respond",
    )
    print("  ✓ vSphere host inventory is accessible")
