"""Setup · banner -- print environment, versions, and configuration summary.

This is the very first test to run (order 0).  It prints a block of
diagnostic info so the operator can see *exactly* which cluster,
credentials, and tools are in use.  It also verifies that ``kubectl-mtv``
is executable and will fail the entire suite early if it is not.
"""

import subprocess
import sys

import pytest

from conftest import (
    GOVC_URL,
    GOVC_USERNAME,
    KUBE_API_URL,
    MCP_IMAGE,
    MCP_SSE_PORT,
    MCP_SSE_URL,
    MCP_VERBOSE,
    MTV_BINARY,
    TEST_NAMESPACE,
    COLD_PLAN_NAME,
    WARM_PLAN_NAME,
    COLD_VMS,
    WARM_VMS,
    NETWORK_PAIRS,
    STORAGE_PAIRS,
    VSPHERE_PROVIDER_NAME,
    VSPHERE_URL,
    OCP_PROVIDER_NAME,
    ESXI_HOST_NAME,
    _mtv_base_args,
)


def _cli_version() -> tuple[str, bool]:
    """Run ``kubectl-mtv version --server ... --token ...`` and return (output, ok).

    Returns a tuple of (version_string, success).  When the binary is missing
    or the command fails, *success* is ``False``.
    """
    try:
        r = subprocess.run(
            _mtv_base_args() + ["version"],
            capture_output=True,
            text=True,
            timeout=10,
        )
        raw = (r.stdout.strip() or r.stderr.strip() or "(no output)")
        # Indent continuation lines so they align under the header
        lines = raw.splitlines()
        formatted = ("\n" + " " * 22).join(lines)
        return formatted, r.returncode == 0
    except FileNotFoundError:
        return "(kubectl-mtv binary not found)", False
    except Exception as exc:
        return f"(error: {exc})", False


def _section(title: str) -> str:
    return f"\n  --- {title} ---"


@pytest.mark.order(0)
async def test_print_banner(mcp_session):
    """Print versions, credentials, and test configuration.

    Fails the entire suite if the ``kubectl-mtv`` binary cannot be executed.
    """
    cli_ver, cli_ok = _cli_version()

    server_mode = f"container ({MCP_IMAGE})" if MCP_IMAGE else f"local binary ({MTV_BINARY})"

    banner = "\n".join([
        "",
        "=" * 60,
        "  MCP E2E TEST SUITE",
        "=" * 60,
        _section("Versions"),
        f"  Python:           {sys.version.split()[0]}",
        f"  pytest:           {pytest.__version__}",
        f"  kubectl-mtv:      {cli_ver}",
        _section("Cluster"),
        f"  API URL:          {KUBE_API_URL}",
        f"  MCP server mode:  {server_mode}",
        f"  MCP SSE port:     {MCP_SSE_PORT}",
        f"  MCP SSE URL:      {MCP_SSE_URL}",
        _section("vSphere source"),
        f"  GOVC_URL:         {GOVC_URL}",
        f"  Provider URL:     {VSPHERE_URL}",
        f"  Username:         {GOVC_USERNAME}",
        f"  Provider name:    {VSPHERE_PROVIDER_NAME}",
        _section("OpenShift target"),
        f"  API URL:          {KUBE_API_URL}",
        f"  Provider name:    {OCP_PROVIDER_NAME}",
        f"  ESXi host:        {ESXI_HOST_NAME}",
        _section("Migration plans"),
        f"  Namespace:        {TEST_NAMESPACE}",
        f"  Cold plan:        {COLD_PLAN_NAME}",
        f"  Warm plan:        {WARM_PLAN_NAME}",
        f"  Cold VMs:         {COLD_VMS}",
        f"  Warm VMs:         {WARM_VMS}",
        f"  Network pairs:    {NETWORK_PAIRS}",
        f"  Storage pairs:    {STORAGE_PAIRS}",
        _section("Diagnostics"),
        f"  MCP_VERBOSE:      {MCP_VERBOSE}",
        "",
        "=" * 60,
        "",
    ])
    print(banner)

    assert cli_ok, (
        f"kubectl-mtv version failed — cannot continue.\n  {cli_ver}"
    )
