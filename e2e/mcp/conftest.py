"""
Root conftest -- shared fixtures, helpers, and constants for MCP E2E tests.

The MCP server is started in SSE mode as a subprocess.  Tests connect
via the Python MCP SDK's SSE client.  Environment variables are loaded
from a .env file if present; otherwise plain ``os.environ`` is used.
"""

import asyncio as _asyncio
import contextlib
import json
import os
import signal
import socket
import subprocess
import time

import pytest
import pytest_asyncio
from mcp import ClientSession
from mcp.client.sse import sse_client

# ---------------------------------------------------------------------------
# Environment -- load .env file if present, otherwise use system env vars
# ---------------------------------------------------------------------------
from dotenv import load_dotenv

_env_path = os.path.join(os.path.dirname(__file__), ".env")
if os.path.exists(_env_path):
    load_dotenv(_env_path)

# ---------------------------------------------------------------------------
# Required environment variables — fail early with a clear message
# ---------------------------------------------------------------------------
_MISSING: list[str] = []


def _require(name: str) -> str:
    """Return the value of an environment variable or record it as missing."""
    val = os.environ.get(name)
    if not val:
        _MISSING.append(name)
        return ""
    return val


GOVC_URL: str = _require("GOVC_URL")
GOVC_USERNAME: str = _require("GOVC_USERNAME")
GOVC_PASSWORD: str = _require("GOVC_PASSWORD")
GOVC_INSECURE: str = os.environ.get("GOVC_INSECURE", "true")
KUBE_API_URL: str = _require("KUBE_API_URL")
KUBE_TOKEN: str = _require("KUBE_TOKEN")

# ESXi host credentials (fall back to GOVC_* values if not set explicitly)
ESXI_USERNAME: str = os.environ.get("ESXI_USERNAME") or GOVC_USERNAME
ESXI_PASSWORD: str = os.environ.get("ESXI_PASSWORD") or GOVC_PASSWORD

# Normalize vSphere URL: ensure https:// prefix and /sdk suffix
_raw_govc = GOVC_URL
if not _raw_govc.startswith(("http://", "https://")):
    _raw_govc = f"https://{_raw_govc}"
if not _raw_govc.rstrip("/").endswith("/sdk"):
    _raw_govc = _raw_govc.rstrip("/") + "/sdk"
VSPHERE_URL: str = _raw_govc

# ---------------------------------------------------------------------------
# MCP server settings
# ---------------------------------------------------------------------------
MTV_BINARY: str = os.environ.get(
    "MTV_BINARY",
    os.path.join(os.path.dirname(__file__), "..", "..", "kubectl-mtv"),
)
MCP_SSE_PORT: str = os.environ.get("MCP_SSE_PORT", "18443")
MCP_SSE_URL: str = f"http://127.0.0.1:{MCP_SSE_PORT}/sse"

# ---------------------------------------------------------------------------
# Test resources — environment-specific, REQUIRED (no defaults)
# ---------------------------------------------------------------------------
ESXI_HOST_NAME: str = _require("ESXI_HOST_NAME")
COLD_VMS: str = _require("COLD_VMS")
WARM_VMS: str = _require("WARM_VMS")
NETWORK_PAIRS: str = _require("NETWORK_PAIRS")
STORAGE_PAIRS: str = _require("STORAGE_PAIRS")

# ---------------------------------------------------------------------------
# Test naming — have sensible defaults, override only if needed
# ---------------------------------------------------------------------------
TEST_NAMESPACE: str = os.environ.get("TEST_NAMESPACE", "mcp-e2e-test")
VSPHERE_PROVIDER_NAME: str = os.environ.get("VSPHERE_PROVIDER_NAME", "mcp-e2e-vsphere")
OCP_PROVIDER_NAME: str = os.environ.get("OCP_PROVIDER_NAME", "mcp-e2e-host")
COLD_PLAN_NAME: str = os.environ.get("COLD_PLAN_NAME", "mcp-e2e-cold-plan")
WARM_PLAN_NAME: str = os.environ.get("WARM_PLAN_NAME", "mcp-e2e-warm-plan")

# ---------------------------------------------------------------------------
# Fail fast if any required variables are missing
# ---------------------------------------------------------------------------
if _MISSING:
    raise EnvironmentError(
        "The following required environment variables are not set:\n"
        + "".join(f"  - {v}\n" for v in _MISSING)
        + "\nCopy e2e/mcp/env.example to .env and fill in real values, "
        "or export them in your shell."
    )

# Derived from NETWORK_PAIRS / STORAGE_PAIRS -- the source-side names the
# inventory *must* contain for the migration plans to work.
TEST_NETWORK_NAME = NETWORK_PAIRS.split(":")[0]       # e.g. "VM Network"
TEST_DATASTORE_NAME = STORAGE_PAIRS.split(":")[0]     # e.g. "nfs-us-mtv-v8"


# ---------------------------------------------------------------------------
# Custom exception
# ---------------------------------------------------------------------------
class MCPToolError(Exception):
    """Raised when an MCP tool call returns an error."""

    def __init__(self, tool: str, message: str):
        self.tool = tool
        super().__init__(f"MCP tool '{tool}' error: {message}")


# ---------------------------------------------------------------------------
# Helper: call an MCP tool and return parsed result
# ---------------------------------------------------------------------------
# Default verbosity for MCP tool calls (0=silent, 1=info, 2=debug, 3=trace)
MCP_VERBOSE: int = int(os.environ.get("MCP_VERBOSE", "1"))


def _redact_secrets(obj, secret_keys: set):
    """Return a deep copy of *obj* with values for *secret_keys* replaced."""
    if isinstance(obj, dict):
        return {
            k: ("***" if k in secret_keys else _redact_secrets(v, secret_keys))
            for k, v in obj.items()
        }
    if isinstance(obj, list):
        return [_redact_secrets(v, secret_keys) for v in obj]
    return obj


async def call_tool(
    session: ClientSession,
    tool_name: str,
    arguments: dict,
    *,
    verbose: int | None = None,
) -> dict:
    """Call an MCP tool via the SSE session and return the parsed response.

    The kubectl-mtv MCP server returns a JSON envelope::

        {"return_value": 0, "data": [...]}   # structured (json output)
        {"return_value": 0, "output": "..."}  # text (table / yaml / health)

    This helper extracts that envelope.  On ``isError`` it raises
    :class:`MCPToolError`.

    Args:
        verbose: Override the default verbosity (``MCP_VERBOSE`` env var).
            When >= 1, the tool name and arguments are printed before the
            call, and a summary of the response is printed after.
    """
    level = verbose if verbose is not None else MCP_VERBOSE

    # Inject verbose into flags when the tool supports it
    if "flags" in arguments and level > 0:
        arguments = {**arguments, "flags": {**arguments["flags"], "verbose": level}}

    _secret_keys = {"password", "token"}

    if level >= 2:
        sanitized = _redact_secrets(arguments, _secret_keys)
        print(f"\n    [call] {tool_name} {json.dumps(sanitized, indent=2)}")
    elif level >= 1:
        cmd = arguments.get("command", arguments.get("action", ""))
        flags_summary = {
            k: v for k, v in arguments.get("flags", {}).items()
            if k not in _secret_keys and k != "verbose"
        }
        print(f"\n    [call] {tool_name} {cmd}  {flags_summary}")

    result = await session.call_tool(tool_name, arguments)

    if result.isError:
        parts = []
        for content in result.content:
            if hasattr(content, "text"):
                parts.append(content.text)
        error_msg = "\n".join(parts)
        if level >= 1:
            # Strip klog noise so the real error is visible
            meaningful = "\n".join(
                ln for ln in error_msg.splitlines()
                if not ln.lstrip().startswith("I0") and ln.strip()
            )
            print(f"    [error] {tool_name}: {meaningful[:800]}")
        raise MCPToolError(tool_name, error_msg)

    # Prefer structuredContent (Go MCP SDK populates this from output schema)
    if result.structuredContent is not None:
        parsed = result.structuredContent
    else:
        # Fall back to parsing the first text content block as JSON
        parsed = {}
        for content in result.content:
            if hasattr(content, "text"):
                try:
                    parsed = json.loads(content.text)
                except json.JSONDecodeError:
                    parsed = {"output": content.text, "return_value": 0}
                break

    if level >= 2:
        preview = json.dumps(parsed, indent=2)[:800]
        print(f"    [result] {preview}")
    elif level >= 1:
        rc = parsed.get("return_value", "?")
        data_len = len(parsed.get("data", []))
        out_len = len(parsed.get("output", ""))
        print(f"    [result] return_value={rc}  data_items={data_len}  output_len={out_len}")

    return parsed


# ---------------------------------------------------------------------------
# Helper: kubectl wait
# ---------------------------------------------------------------------------
def _kubectl_wait(
    resource: str | list[str],
    condition: str,
    *,
    namespace: str | None = None,
    timeout: int = 120,
) -> subprocess.CompletedProcess:
    """Run ``kubectl wait --for=<condition>`` and return the result.

    Args:
        resource: One or more resources, e.g. ``"namespace/foo"`` or
            ``["providers.forklift.konveyor.io/a", "...b"]``.
        condition: The ``--for`` value, e.g. ``"delete"`` or
            ``"jsonpath={.status.phase}=Ready"``.
        namespace: Kubernetes namespace (omit for cluster-scoped resources).
        timeout: Maximum seconds to wait.

    Raises:
        RuntimeError: If ``kubectl wait`` exits with a non-zero code.
    """
    if isinstance(resource, str):
        resource = [resource]

    cmd = _kubectl_base_args() + [
        "wait", *resource,
        f"--for={condition}",
        f"--timeout={timeout}s",
    ]
    if namespace:
        cmd += ["-n", namespace]

    r = subprocess.run(cmd, capture_output=True, text=True, timeout=timeout + 30)
    if r.returncode != 0:
        raise RuntimeError(
            f"kubectl wait failed (rc={r.returncode}): {r.stderr.strip()}"
        )
    return r


# ---------------------------------------------------------------------------
# Helper: retry a CLI command until it succeeds
# ---------------------------------------------------------------------------
def _retry_command(
    cmd: list[str],
    *,
    timeout: int = 180,
    interval: int = 10,
    description: str = "command",
) -> subprocess.CompletedProcess:
    """Retry *cmd* until it exits 0 or *timeout* seconds elapse.

    Use this for application-level readiness checks where no Kubernetes
    resource condition exists (e.g. waiting for the inventory service to
    start serving requests after a provider becomes Ready).

    Raises:
        RuntimeError: If the command never succeeds within *timeout*.
    """
    deadline = time.monotonic() + timeout
    last_result = None
    while time.monotonic() < deadline:
        last_result = subprocess.run(
            cmd, capture_output=True, text=True, timeout=30,
        )
        if last_result.returncode == 0:
            return last_result
        time.sleep(interval)

    stderr = last_result.stderr.strip() if last_result else "(no output)"
    raise RuntimeError(
        f"Timed out after {timeout}s waiting for {description}: {stderr}"
    )


# ---------------------------------------------------------------------------
# Helper: kubectl namespace management (direct subprocess, not via MCP)
# ---------------------------------------------------------------------------
def _kubectl_base_args() -> list[str]:
    return [
        "kubectl",
        "--server", KUBE_API_URL,
        "--token", KUBE_TOKEN,
        "--insecure-skip-tls-verify",
    ]


def _mtv_base_args() -> list[str]:
    return [
        MTV_BINARY,
        "--server", KUBE_API_URL,
        "--token", KUBE_TOKEN,
        "--insecure-skip-tls-verify",
    ]


def _create_namespace(name: str) -> None:
    cmd = _kubectl_base_args() + ["create", "namespace", name]
    r = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
    if r.returncode != 0 and "already exists" not in r.stderr:
        raise RuntimeError(f"Failed to create namespace {name}: {r.stderr}")


def _delete_namespace(name: str) -> None:
    cmd = _kubectl_base_args() + ["delete", "namespace", name, "--ignore-not-found"]
    subprocess.run(cmd, capture_output=True, text=True, timeout=120)


# ---------------------------------------------------------------------------
# Fixture: start the MCP server subprocess in SSE mode
# ---------------------------------------------------------------------------
@pytest.fixture(scope="session")
def mcp_server_process():
    """Start kubectl-mtv mcp-server --sse and yield the subprocess."""
    cmd = [
        MTV_BINARY, "mcp-server",
        "--sse",
        "--port", MCP_SSE_PORT,
        "--server", KUBE_API_URL,
        "--token", KUBE_TOKEN,
    ]
    proc = subprocess.Popen(
        cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    # Give the server a moment to start listening
    deadline = time.monotonic() + 15
    started = False
    while time.monotonic() < deadline:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            if s.connect_ex(("127.0.0.1", int(MCP_SSE_PORT))) == 0:
                started = True
                break
        time.sleep(0.3)

    if not started:
        proc.kill()
        stderr = proc.stderr.read().decode() if proc.stderr else ""
        raise RuntimeError(
            f"MCP SSE server failed to start on port {MCP_SSE_PORT}: {stderr}"
        )

    yield proc

    # Teardown: graceful shutdown
    proc.send_signal(signal.SIGTERM)
    try:
        proc.wait(timeout=10)
    except subprocess.TimeoutExpired:
        proc.kill()
        proc.wait(timeout=5)


# ---------------------------------------------------------------------------
# Fixture: MCP client session over SSE (session-scoped)
# ---------------------------------------------------------------------------
@contextlib.asynccontextmanager
async def _safe_sse_client(*args, **kwargs):
    """Wrap ``sse_client`` and suppress the harmless anyio cancel-scope error.

    During pytest-asyncio session-scoped fixture teardown the event loop
    may finalize the ``sse_client`` context manager in a different task
    than the one that created it, causing::

        RuntimeError: Attempted to exit cancel scope in a different task …

    This wrapper catches that specific error so the test run exits cleanly.
    """
    try:
        async with sse_client(*args, **kwargs) as streams:
            yield streams
    except RuntimeError as exc:
        if "cancel scope" in str(exc):
            pass  # harmless teardown race
        else:
            raise
    except BaseException:
        raise


@pytest_asyncio.fixture(loop_scope="session", scope="session")
async def mcp_session(mcp_server_process):
    """Connect to the running MCP SSE server and yield a ClientSession."""
    async with _safe_sse_client(MCP_SSE_URL, timeout=30, sse_read_timeout=120) as (
        read_stream,
        write_stream,
    ):
        async with ClientSession(read_stream, write_stream) as session:
            await session.initialize()
            yield session


# ---------------------------------------------------------------------------
# Fixture: create / destroy the test namespace (session-scoped, autouse)
# ---------------------------------------------------------------------------
@pytest_asyncio.fixture(loop_scope="session", scope="session", autouse=True)
async def cleanup_test_resources(mcp_session):
    """Clean up all test resources after the session completes."""
    yield

    # --- TEARDOWN (best-effort, reverse order) ---

    # Delete plans
    try:
        await call_tool(mcp_session, "mtv_write", {
            "command": "delete plan",
            "flags": {
                "name": f"{COLD_PLAN_NAME},{WARM_PLAN_NAME}",
                "namespace": TEST_NAMESPACE,
                "skip-archive": True,
            },
        })
    except Exception:
        pass

    # Delete host -- discover the K8s resource name(s) first, since
    # create host generates names like "{inventoryID}-{hash}".
    try:
        result = await call_tool(mcp_session, "mtv_read", {
            "command": "get host",
            "flags": {"namespace": TEST_NAMESPACE, "output": "json"},
        })
        data = result.get("data", [])
        hosts = data if isinstance(data, list) else [data]
        host_names = [
            h.get("name") or h.get("metadata", {}).get("name", "")
            for h in hosts
            if (h.get("name") or h.get("metadata", {}).get("name", "")).startswith(ESXI_HOST_NAME)
        ]
        if host_names:
            await call_tool(mcp_session, "mtv_write", {
                "command": "delete host",
                "flags": {
                    "name": ",".join(host_names),
                    "namespace": TEST_NAMESPACE,
                },
            })
    except Exception:
        pass

    # Delete providers
    try:
        await call_tool(mcp_session, "mtv_write", {
            "command": "delete provider",
            "flags": {
                "name": f"{VSPHERE_PROVIDER_NAME},{OCP_PROVIDER_NAME}",
                "namespace": TEST_NAMESPACE,
            },
        })
    except Exception:
        pass

    # Delete namespace last (subprocess -- no async context needed)
    try:
        _delete_namespace(TEST_NAMESPACE)
    except Exception:
        pass

    # Allow the event loop a moment to drain before session fixtures tear down
    await _asyncio.sleep(0.5)
