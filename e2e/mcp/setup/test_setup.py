"""Setup -- verify MCP server, clean old namespace, then create a fresh one."""

import subprocess

import pytest

from conftest import (
    KUBE_API_URL,
    MCP_SSE_PORT,
    MCP_SSE_URL,
    TEST_NAMESPACE,
    _create_namespace,
    _delete_namespace,
    _kubectl_base_args,
    _kubectl_wait,
    call_tool,
)


@pytest.mark.order(1)
async def test_mcp_server_running(mcp_session):
    """Verify the MCP SSE server is up and the client session is connected."""
    result = await call_tool(mcp_session, "mtv_help", {"command": "get plan"})
    assert result, "MCP server returned empty response to mtv_help"

    print(f"\n  MCP SSE server responding on port {MCP_SSE_PORT}")
    print(f"  Client connected to {MCP_SSE_URL}")


@pytest.mark.order(2)
async def test_clean_namespace(mcp_session):
    """Delete the test namespace (if it exists) and wait until it is gone."""
    # Check whether the namespace exists
    probe = subprocess.run(
        _kubectl_base_args() + [
            "get", "namespace", TEST_NAMESPACE,
            "-o=jsonpath={.metadata.name}",
        ],
        capture_output=True, text=True, timeout=30,
    )

    if probe.returncode == 0 and probe.stdout.strip() == TEST_NAMESPACE:
        print(f"\n  Namespace '{TEST_NAMESPACE}' exists -- deleting ...")
        _delete_namespace(TEST_NAMESPACE)

        # kubectl wait --for=delete exits 0 when the resource is gone
        _kubectl_wait(f"namespace/{TEST_NAMESPACE}", "delete", timeout=120)

        print(f"  Namespace '{TEST_NAMESPACE}' deleted and gone")
    else:
        print(f"\n  Namespace '{TEST_NAMESPACE}' does not exist -- nothing to clean")


@pytest.mark.order(3)
async def test_create_namespace(mcp_session):
    """Create a fresh test namespace and wait until it is fully initialised."""
    _create_namespace(TEST_NAMESPACE)

    # Wait for namespace to become Active
    _kubectl_wait(
        f"namespace/{TEST_NAMESPACE}",
        "jsonpath={.status.phase}=Active",
        timeout=60,
    )

    # Wait for the default ServiceAccount -- this signals the namespace
    # controller has finished setting up RBAC and the namespace is ready
    # for workloads.
    _kubectl_wait(
        "serviceaccount/default",
        "jsonpath={.metadata.name}=default",
        namespace=TEST_NAMESPACE,
        timeout=30,
    )

    print(f"\n  Namespace '{TEST_NAMESPACE}' created and fully initialised")
    print(f"  OCP API: {KUBE_API_URL}")
