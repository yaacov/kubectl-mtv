#!/usr/bin/env python3
"""E2E smoke tests for kubectl-mtv.

Runs a suite of tests against every CLI command area using an OpenShift
cluster with MTV/Forklift installed.  Build the binary first (e.g.
``make e2e`` or ``make build && python3 tests/e2e_smoke.py``).

The lifecycle tests (create, patch, delete) create a dedicated namespace,
exercise write operations, then clean up.

Usage:
    make test-e2e
"""

import json
import os
import subprocess
import sys
import time
from typing import List, Optional

BINARY = os.path.join(os.path.dirname(__file__), "..", "kubectl-mtv")

TEST_NAMESPACE = "kubectl-mtv-smoke-test"
SMOKE_PROVIDER = "smoke-ocp"

passed = 0
failed = 0
errors = []  # type: List[str]


# ---------------------------------------------------------------------------
# Color (pytest-style: green PASS, red FAIL, yellow SKIP)
# ---------------------------------------------------------------------------

def _use_color():
    if os.environ.get("NO_COLOR"):
        return False
    if os.environ.get("FORCE_COLOR") or os.environ.get("CLICOLOR_FORCE"):
        return True
    return hasattr(sys.stdout, "isatty") and sys.stdout.isatty()


_COLOR = _use_color()


def _paint(code, text):
    if not _COLOR:
        return text
    return "\033[%sm%s\033[0m" % (code, text)


def section(title):
    print(_paint("1;36", "[%s]" % title))


def skip(reason):
    print("  %s  %s" % (_paint("33", "SKIP"), reason))


def warn(msg):
    print("  %s: %s" % (_paint("33", "WARNING"), msg))


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _decode(data):
    """Coerce bytes to str; pass through if already a string."""
    if isinstance(data, bytes):
        return data.decode("utf-8", errors="replace")
    return data or ""


def run(args):
    """Run the binary with the given args, return (stdout, stderr, returncode)."""
    try:
        result = subprocess.run(
            [BINARY, *args],
            capture_output=True,
            text=True,
            timeout=60,
        )
        return result.stdout, result.stderr, result.returncode
    except subprocess.TimeoutExpired as exc:
        stdout = _decode(exc.stdout)
        stderr = _decode(exc.stderr) + "\n[TIMEOUT] command timed out after 60s"
        return stdout, stderr, 1


def kubectl(args):
    """Run kubectl with the given args, return (stdout, stderr, returncode)."""
    cmd = ["kubectl", *args]
    try:
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=60,
        )
        return result.stdout, result.stderr, result.returncode
    except subprocess.TimeoutExpired as exc:
        stdout = _decode(exc.stdout)
        stderr = _decode(exc.stderr) + f"\n[TIMEOUT] {cmd} timed out after 60s"
        return stdout, stderr, 1
    except subprocess.CalledProcessError as exc:
        return _decode(exc.stdout), _decode(exc.stderr), exc.returncode
    except Exception as exc:
        return "", f"[ERROR] {cmd}: {exc}", 1


def record(name: str, ok: bool, detail: str = ""):
    global passed, failed
    if ok:
        passed += 1
        print("  %s  %s" % (_paint("1;32", "PASS"), name))
    else:
        failed += 1
        msg = "  %s  %s" % (_paint("1;31", "FAIL"), name)
        if detail:
            msg += "  -- %s" % detail
        print(msg)
        errors.append(name)


def assert_exit_ok(name: str, rc: int, stderr: str = "") -> bool:
    ok = rc == 0
    record(name, ok, f"exit={rc} stderr={stderr[:120]}" if not ok else "")
    return ok


def assert_exit_fail(name: str, rc: int) -> bool:
    ok = rc != 0
    record(name, ok, "expected non-zero exit code" if not ok else "")
    return ok


def assert_contains(name: str, text: str, substring: str) -> bool:
    ok = substring in text
    record(name, ok, f"output missing '{substring}'" if not ok else "")
    return ok


def parse_json(text: str):
    try:
        return json.loads(text), None
    except json.JSONDecodeError as exc:
        return None, str(exc)


def assert_valid_json(name: str, text: str):
    data, err = parse_json(text)
    record(name, err is None, f"invalid JSON: {err}" if err else "")
    return data


def assert_valid_yaml(name: str, text: str):
    try:
        import yaml as _yaml
        data = _yaml.safe_load(text)
        record(name, True)
        return data
    except ImportError:
        record(name, len(text.strip()) > 0, "pyyaml not installed, checking non-empty instead")
        return None
    except _yaml.YAMLError as exc:
        record(name, False, f"invalid YAML: {exc}")
        return None


# ---------------------------------------------------------------------------
# Namespace helpers (modelled after e2e/mcp/conftest.py)
# ---------------------------------------------------------------------------

def setup_test_namespace():
    """Delete stale namespace if present, then create a fresh one."""
    # Delete if leftover from a previous run
    stdout, _, rc = kubectl(["get", "namespace", TEST_NAMESPACE, "-o=jsonpath={.metadata.name}"])
    if rc == 0 and stdout.strip() == TEST_NAMESPACE:
        kubectl(["delete", "namespace", TEST_NAMESPACE, "--ignore-not-found", "--wait=true", "--timeout=120s"])

    _, stderr, rc = kubectl(["create", "namespace", TEST_NAMESPACE])
    if rc != 0 and "already exists" not in stderr:
        warn("failed to create namespace %s: %s" % (TEST_NAMESPACE, stderr.strip()))
        return False

    # Wait for namespace to become Active
    for _ in range(30):
        stdout, _, rc = kubectl(["get", "namespace", TEST_NAMESPACE, "-o=jsonpath={.status.phase}"])
        if rc == 0 and stdout.strip() == "Active":
            return True
        time.sleep(1)

    warn("namespace %s did not become Active" % TEST_NAMESPACE)
    return False


def teardown_test_namespace():
    """Best-effort cleanup of the test namespace."""
    kubectl(["delete", "namespace", TEST_NAMESPACE, "--ignore-not-found", "--wait=false"])


# ---------------------------------------------------------------------------
# Tests: version
# ---------------------------------------------------------------------------

def test_version():
    section("version")

    # 1. version --client (no cluster needed)
    stdout, stderr, rc = run(["version", "--client"])
    if assert_exit_ok("version --client exits 0", rc, stderr):
        assert_contains("version --client output", stdout, "kubectl-mtv")

    # 2. version --client -o json
    stdout, stderr, rc = run(["version", "--client", "-o", "json"])
    if assert_exit_ok("version --client json exits 0", rc, stderr):
        data = assert_valid_json("version --client valid json", stdout)
        if data and isinstance(data, dict):
            assert_contains("version json has clientVersion", json.dumps(data), "clientVersion")

    # 3. Full version (cluster required)
    stdout, stderr, rc = run(["version"])
    if assert_exit_ok("version exits 0", rc, stderr):
        assert_contains("version output", stdout, "kubectl-mtv")

    # 4. Full version JSON
    stdout, stderr, rc = run(["version", "-o", "json"])
    if assert_exit_ok("version json exits 0", rc, stderr):
        assert_valid_json("version valid json", stdout)


# ---------------------------------------------------------------------------
# Tests: health
# ---------------------------------------------------------------------------

def test_health():
    section("health")

    # 1. Health with --skip-logs (faster)
    stdout, stderr, rc = run(["health", "--skip-logs"])
    if assert_exit_ok("health --skip-logs exits 0", rc, stderr):
        record("health output non-empty", len(stdout.strip()) > 0,
               "empty output" if not stdout.strip() else "")

    # 2. Health JSON output
    stdout, stderr, rc = run(["health", "-o", "json", "--skip-logs"])
    if assert_exit_ok("health json exits 0", rc, stderr):
        assert_valid_json("health valid json", stdout)

    # 3. Health markdown output
    stdout, stderr, rc = run(["health", "-o", "markdown", "--skip-logs"])
    if assert_exit_ok("health markdown exits 0", rc, stderr):
        record("health markdown non-empty", len(stdout.strip()) > 0,
               "empty output" if not stdout.strip() else "")

    # 4. Health with full log analysis
    stdout, stderr, rc = run(["health"])
    assert_exit_ok("health full exits 0", rc, stderr)


# ---------------------------------------------------------------------------
# Tests: settings
# ---------------------------------------------------------------------------

def test_settings():
    section("settings")

    # 1. Default settings view
    stdout, stderr, rc = run(["settings"])
    if assert_exit_ok("settings exits 0", rc, stderr):
        record("settings output non-empty", len(stdout.strip()) > 0,
               "empty output" if not stdout.strip() else "")

    # 2. Settings JSON output
    default_count = None
    stdout, stderr, rc = run(["settings", "-o", "json"])
    if assert_exit_ok("settings json exits 0", rc, stderr):
        data = assert_valid_json("settings valid json", stdout)
        if isinstance(data, list):
            default_count = len(data)
            record("settings json is array", len(data) > 0,
                   "empty JSON array" if len(data) == 0 else "")

    # 3. Settings YAML output
    stdout, stderr, rc = run(["settings", "-o", "yaml"])
    if assert_exit_ok("settings yaml exits 0", rc, stderr):
        assert_valid_yaml("settings valid yaml", stdout)

    # 4. Settings --all (includes advanced settings)
    stdout, stderr, rc = run(["settings", "--all", "-o", "json"])
    if assert_exit_ok("settings --all json exits 0", rc, stderr):
        data = assert_valid_json("settings --all valid json", stdout)
        if isinstance(data, list):
            if default_count is not None:
                record("settings --all has more entries",
                       len(data) > default_count,
                       f"--all={len(data)} default={default_count}")
            else:
                record("settings --all has entries", len(data) > 0,
                       "empty JSON array" if len(data) == 0 else "")

    # 5. Settings get subcommand
    stdout, stderr, rc = run(["settings", "get"])
    assert_exit_ok("settings get exits 0", rc, stderr)

    # 6. Settings get --all
    stdout, stderr, rc = run(["settings", "get", "--all"])
    assert_exit_ok("settings get --all exits 0", rc, stderr)

    # 7. Settings get specific setting
    stdout, stderr, rc = run(["settings", "get", "--setting", "vddk_image"])
    assert_exit_ok("settings get --setting vddk_image exits 0", rc, stderr)


# ---------------------------------------------------------------------------
# Tests: get providers
# ---------------------------------------------------------------------------

def _extract_name(item):
    """Extract name from a JSON item (nested under metadata.name)."""
    meta = item.get("metadata") or {}
    return meta.get("name") or item.get("name") or ""


def _extract_namespace(item):
    """Extract namespace from a JSON item (nested under metadata.namespace)."""
    meta = item.get("metadata") or {}
    return meta.get("namespace") or item.get("namespace") or ""


def _first_provider_name():
    # type: () -> Optional[str]
    """Dynamically fetch the first provider name."""
    stdout, _, rc = run(["get", "providers", "-o", "json", "-A"])
    if rc != 0:
        return None
    data, _ = parse_json(stdout)
    if data and isinstance(data, list) and len(data) > 0:
        return _extract_name(data[0]) or None
    return None


def _first_provider_namespace():
    # type: () -> Optional[str]
    """Dynamically fetch the namespace of the first provider."""
    stdout, _, rc = run(["get", "providers", "-o", "json", "-A"])
    if rc != 0:
        return None
    data, _ = parse_json(stdout)
    if data and isinstance(data, list) and len(data) > 0:
        return _extract_namespace(data[0]) or None
    return None


def test_get_providers():
    section("get providers")

    # 1. List providers (all namespaces)
    stdout, stderr, rc = run(["get", "providers", "-A"])
    if assert_exit_ok("get providers -A exits 0", rc, stderr):
        record("get providers output non-empty", len(stdout.strip()) > 0,
               "empty output" if not stdout.strip() else "")

    # 2. JSON output
    stdout, stderr, rc = run(["get", "providers", "-o", "json", "-A"])
    if assert_exit_ok("get providers json exits 0", rc, stderr):
        data = assert_valid_json("get providers valid json", stdout)
        if isinstance(data, list):
            record("get providers json is array", True)

    # 3. YAML output
    stdout, stderr, rc = run(["get", "providers", "-o", "yaml", "-A"])
    if assert_exit_ok("get providers yaml exits 0", rc, stderr):
        assert_valid_yaml("get providers valid yaml", stdout)

    # 4. Markdown output
    stdout, stderr, rc = run(["get", "providers", "-o", "markdown", "-A"])
    assert_exit_ok("get providers markdown exits 0", rc, stderr)


# ---------------------------------------------------------------------------
# Tests: get plans
# ---------------------------------------------------------------------------

def _first_plan_name():
    # type: () -> Optional[str]
    """Dynamically fetch the first plan name."""
    stdout, _, rc = run(["get", "plans", "-o", "json", "-A"])
    if rc != 0:
        return None
    data, _ = parse_json(stdout)
    if data and isinstance(data, list) and len(data) > 0:
        return _extract_name(data[0]) or None
    return None


def _first_plan_namespace():
    # type: () -> Optional[str]
    """Dynamically fetch the namespace of the first plan."""
    stdout, _, rc = run(["get", "plans", "-o", "json", "-A"])
    if rc != 0:
        return None
    data, _ = parse_json(stdout)
    if data and isinstance(data, list) and len(data) > 0:
        return _extract_namespace(data[0]) or None
    return None


def _first_named(*resource_path):
    # type: (*str) -> tuple
    """Return (name, namespace) of the first item from ``get <resource> -A -o json``."""
    stdout, _, rc = run(["get", *resource_path, "-o", "json", "-A"])
    if rc != 0:
        return None, None
    data, _ = parse_json(stdout)
    if data and isinstance(data, list) and len(data) > 0:
        return _extract_name(data[0]) or None, _extract_namespace(data[0]) or None
    return None, None


WRITE_VERBS = (
    "create", "delete", "patch", "start", "cancel", "cutover", "archive", "unarchive",
)
EXPECTED_MACHINE_VERBS = ("get", "create", "patch", "start", "cutover", "archive")


def _command_paths(schema):
    # type: (dict) -> List[str]
    """Extract path_string values from help --machine schema."""
    cmds = schema.get("commands") or []
    paths = []
    for cmd in cmds:
        if not isinstance(cmd, dict):
            continue
        path = cmd.get("path_string") or " ".join(cmd.get("path") or [])
        if path:
            paths.append(path)
    return paths


def _first_verb(path):
    # type: (str) -> str
    return path.split()[0] if path else ""


def test_get_plans():
    section("get plans")

    # 1. List plans (all namespaces)
    stdout, stderr, rc = run(["get", "plans", "-A"])
    if assert_exit_ok("get plans -A exits 0", rc, stderr):
        if stdout.strip():
            record("get plans output non-empty", True)
        else:
            skip("get plans output (no plans found)")

    # 2. JSON output
    stdout, stderr, rc = run(["get", "plans", "-o", "json", "-A"])
    if assert_exit_ok("get plans json exits 0", rc, stderr):
        assert_valid_json("get plans valid json", stdout)

    # 3. YAML output
    stdout, stderr, rc = run(["get", "plans", "-o", "yaml", "-A"])
    if assert_exit_ok("get plans yaml exits 0", rc, stderr):
        assert_valid_yaml("get plans valid yaml", stdout)

    # 4. If a plan exists, get it by name
    plan_name = _first_plan_name()
    plan_ns = _first_plan_namespace()
    if plan_name and plan_ns:
        stdout, stderr, rc = run([
            "get", "plan", "--name", plan_name,
            "--namespace", plan_ns, "-o", "json",
        ])
        if assert_exit_ok("get plan by name json exits 0", rc, stderr):
            assert_valid_json("get plan by name valid json", stdout)

        # 5. get plan --vms (requires plan name)
        stdout, stderr, rc = run([
            "get", "plan", "--name", plan_name,
            "--namespace", plan_ns, "--vms",
        ])
        assert_exit_ok("get plan --vms exits 0", rc, stderr)
    else:
        skip("get plan by name (no plans found)")
        skip("get plan --vms (no plans found)")


# ---------------------------------------------------------------------------
# Tests: get mappings
# ---------------------------------------------------------------------------

def test_get_mappings():
    section("get mappings")

    # 1. List all mappings (network + storage)
    stdout, stderr, rc = run(["get", "mappings", "-A"])
    if assert_exit_ok("get mappings -A exits 0", rc, stderr):
        if stdout.strip():
            record("get mappings output non-empty", True)
        else:
            skip("get mappings output (no mappings found)")

    # 2. JSON output
    stdout, stderr, rc = run(["get", "mappings", "-o", "json", "-A"])
    if assert_exit_ok("get mappings json exits 0", rc, stderr):
        assert_valid_json("get mappings valid json", stdout)

    # 3. Network mappings
    stdout, stderr, rc = run(["get", "mapping", "network", "-o", "json", "-A"])
    if assert_exit_ok("get mapping network json exits 0", rc, stderr):
        assert_valid_json("get mapping network valid json", stdout)

    # 4. Storage mappings
    stdout, stderr, rc = run(["get", "mapping", "storage", "-o", "json", "-A"])
    if assert_exit_ok("get mapping storage json exits 0", rc, stderr):
        assert_valid_json("get mapping storage valid json", stdout)


# ---------------------------------------------------------------------------
# Tests: describe
# ---------------------------------------------------------------------------

def test_describe():
    section("describe")

    # Describe provider (dynamic)
    provider_name = _first_provider_name()
    provider_ns = _first_provider_namespace()
    if provider_name and provider_ns:
        stdout, stderr, rc = run([
            "describe", "provider", "--name", provider_name,
            "--namespace", provider_ns,
        ])
        if assert_exit_ok("describe provider exits 0", rc, stderr):
            assert_contains("describe provider has name", stdout, provider_name)

        # JSON output
        stdout, stderr, rc = run([
            "describe", "provider", "--name", provider_name,
            "--namespace", provider_ns, "-o", "json",
        ])
        if assert_exit_ok("describe provider json exits 0", rc, stderr):
            assert_valid_json("describe provider valid json", stdout)
    else:
        skip("describe provider (no providers found)")

    # Describe plan (dynamic)
    plan_name = _first_plan_name()
    plan_ns = _first_plan_namespace()
    if plan_name and plan_ns:
        stdout, stderr, rc = run([
            "describe", "plan", "--name", plan_name,
            "--namespace", plan_ns,
        ])
        if assert_exit_ok("describe plan exits 0", rc, stderr):
            assert_contains("describe plan has name", stdout, plan_name)

        # JSON output
        stdout, stderr, rc = run([
            "describe", "plan", "--name", plan_name,
            "--namespace", plan_ns, "-o", "json",
        ])
        if assert_exit_ok("describe plan json exits 0", rc, stderr):
            assert_valid_json("describe plan valid json", stdout)
    else:
        skip("describe plan (no plans found)")

    # Describe mapping (dynamic)
    mapping_name, mapping_ns = _first_named("mapping", "network")
    if mapping_name and mapping_ns:
        stdout, stderr, rc = run([
            "describe", "mapping", "network", "--name", mapping_name,
            "--namespace", mapping_ns,
        ])
        if assert_exit_ok("describe mapping network exits 0", rc, stderr):
            assert_contains("describe mapping network has name", stdout, mapping_name)
    else:
        skip("describe mapping network (no network mappings found)")


# ---------------------------------------------------------------------------
# Tests: get inventory / host / hook / conversion (skip if absent)
# ---------------------------------------------------------------------------

def test_get_inventory():
    section("get inventory")

    provider_name = _first_provider_name()
    provider_ns = _first_provider_namespace()
    if not provider_name or not provider_ns:
        skip("get inventory (no providers found)")
        return

    for kind in ("vm", "network", "storage"):
        stdout, stderr, rc = run([
            "get", "inventory", kind,
            "--provider", provider_name,
            "--namespace", provider_ns,
            "-o", "json",
        ])
        if rc != 0:
            skip("get inventory %s (%s)" % (kind, stderr.strip()[:80] or "command failed"))
            continue
        record(f"get inventory {kind} exits 0", True)
        assert_valid_json(f"get inventory {kind} valid json", stdout)


def test_get_optional_resources():
    section("get host / hook / conversion")

    for resource in ("hosts", "hooks", "conversions"):
        stdout, stderr, rc = run(["get", resource, "-o", "json", "-A"])
        if rc != 0:
            skip("get %s (%s)" % (resource, stderr.strip()[:80] or "command failed"))
            continue
        record(f"get {resource} json exits 0", True)
        data = assert_valid_json(f"get {resource} valid json", stdout)
        items = data if isinstance(data, list) else []
        if not items:
            skip("get %s (none found)" % resource)


# ---------------------------------------------------------------------------
# Tests: dry-run create provider
# ---------------------------------------------------------------------------

def test_dry_run_create_provider():
    section("dry-run create provider")

    stdout, stderr, rc = run([
        "create", "provider", "smoke-dryrun-ocp",
        "--type", "openshift",
        "--dry-run",
        "-o", "json",
    ])
    if assert_exit_ok("create provider --dry-run json exits 0", rc, stderr):
        data = assert_valid_json("create provider --dry-run valid json", stdout)
        if data is not None:
            text = json.dumps(data) if not isinstance(data, str) else data
            record("create provider --dry-run has Provider",
                   "Provider" in text or "provider" in text.lower(),
                   f"output keys/preview={text[:160]}")


def test_positional_args():
    """Test that 'get' commands accept names as positional args.

    MarkRequiredForMCP deliberately avoids Cobra's built-in flag enforcement
    so positional args can supply the value instead.  Positional args would
    be blocked only if a command used Cobra's cmd.MarkFlagRequired()
    (e.g. describe, create, patch use MarkRequiredForMCP, not that).
    """
    section("positional args")

    # Discover a provider dynamically for positional arg tests
    provider_name = _first_provider_name()
    provider_ns = _first_provider_namespace()
    if provider_name and provider_ns:
        # get provider <name> (positional)
        stdout, stderr, rc = run([
            "get", "provider", provider_name,
            "--namespace", provider_ns, "-o", "json",
        ])
        if assert_exit_ok("get provider positional exits 0", rc, stderr):
            assert_valid_json("get provider positional valid json", stdout)

        # get provider <name> table (positional)
        stdout, stderr, rc = run([
            "get", "provider", provider_name,
            "--namespace", provider_ns,
        ])
        if assert_exit_ok("get provider positional table exits 0", rc, stderr):
            assert_contains("get provider positional has name", stdout, provider_name)
    else:
        skip("get provider positional (no providers found)")

    # Discover a plan dynamically for positional arg tests
    plan_name = _first_plan_name()
    plan_ns = _first_plan_namespace()
    if plan_name and plan_ns:
        # get plan <name> (positional)
        stdout, stderr, rc = run([
            "get", "plan", plan_name,
            "--namespace", plan_ns, "-o", "json",
        ])
        if assert_exit_ok("get plan positional exits 0", rc, stderr):
            assert_valid_json("get plan positional valid json", stdout)

        # get plan <name> table (positional)
        stdout, stderr, rc = run([
            "get", "plan", plan_name,
            "--namespace", plan_ns,
        ])
        if assert_exit_ok("get plan positional table exits 0", rc, stderr):
            assert_contains("get plan positional has name", stdout, plan_name)
    else:
        skip("get plan positional (no plans found)")


# ---------------------------------------------------------------------------
# Tests: create / patch / delete lifecycle
# ---------------------------------------------------------------------------

def test_create_patch_delete():
    """Full lifecycle: create namespace, create/get/describe/patch/delete provider, cleanup."""
    section("create / patch / delete lifecycle")

    # --- Setup: create test namespace ---
    if not setup_test_namespace():
        record("setup namespace", False, f"failed to create {TEST_NAMESPACE}")
        return

    record("setup namespace", True)

    try:
        _lifecycle_create_provider()
        _lifecycle_get_provider()
        _lifecycle_describe_provider()
        _lifecycle_patch_provider()
        _lifecycle_delete_provider()
    finally:
        # --- Teardown: always clean up ---
        teardown_test_namespace()
        record("teardown namespace", True)


def _lifecycle_create_provider():
    """Create an OpenShift provider using positional arg for name."""
    stdout, stderr, rc = run([
        "create", "provider", SMOKE_PROVIDER,
        "--type", "openshift",
        "--namespace", TEST_NAMESPACE,
    ])
    assert_exit_ok("create provider (positional) exits 0", rc, stderr)

    # Verify it exists
    stdout, stderr, rc = run([
        "get", "providers", "-o", "json",
        "--namespace", TEST_NAMESPACE,
    ])
    if assert_exit_ok("get providers after create exits 0", rc, stderr):
        data = assert_valid_json("get providers after create json", stdout)
        if isinstance(data, list):
            names = [_extract_name(p) for p in data]
            record("created provider appears in list",
                   SMOKE_PROVIDER in names,
                   f"provider '{SMOKE_PROVIDER}' not found in {names}")


def _lifecycle_get_provider():
    """Get the created provider using --name flag and positional arg."""
    # Using --name flag (JSON)
    stdout, stderr, rc = run([
        "get", "provider", "--name", SMOKE_PROVIDER,
        "--namespace", TEST_NAMESPACE, "-o", "json",
    ])
    if assert_exit_ok("get provider --name flag exits 0", rc, stderr):
        assert_valid_json("get provider --name flag json", stdout)

    # Using positional arg (get supports positional args, no MarkRequiredForMCP)
    stdout, stderr, rc = run([
        "get", "provider", SMOKE_PROVIDER,
        "--namespace", TEST_NAMESPACE, "-o", "json",
    ])
    if assert_exit_ok("get provider positional arg exits 0", rc, stderr):
        assert_valid_json("get provider positional arg json", stdout)

    # Table output
    stdout, stderr, rc = run([
        "get", "provider", SMOKE_PROVIDER,
        "--namespace", TEST_NAMESPACE,
    ])
    if assert_exit_ok("get provider table exits 0", rc, stderr):
        assert_contains("get provider table has name", stdout, SMOKE_PROVIDER)


def _lifecycle_describe_provider():
    """Describe the created provider using positional arg."""
    # Table output (positional arg)
    stdout, stderr, rc = run([
        "describe", "provider", SMOKE_PROVIDER,
        "--namespace", TEST_NAMESPACE,
    ])
    if assert_exit_ok("describe provider (positional) exits 0", rc, stderr):
        assert_contains("describe provider has name", stdout, SMOKE_PROVIDER)

    # JSON output (positional arg)
    stdout, stderr, rc = run([
        "describe", "provider", SMOKE_PROVIDER,
        "--namespace", TEST_NAMESPACE, "-o", "json",
    ])
    if assert_exit_ok("describe provider json exits 0", rc, stderr):
        data = assert_valid_json("describe provider valid json", stdout)
        if data and isinstance(data, dict):
            assert_contains("describe provider json has name",
                            json.dumps(data), SMOKE_PROVIDER)


def _lifecycle_patch_provider():
    """Patch the created provider using positional arg — update URL."""
    stdout, stderr, rc = run([
        "patch", "provider", SMOKE_PROVIDER,
        "--namespace", TEST_NAMESPACE,
        "--url", "https://api.smoke-test.example.com:6443",
    ])
    assert_exit_ok("patch provider (positional) exits 0", rc, stderr)

    # Verify the patch took effect via describe JSON
    stdout, stderr, rc = run([
        "describe", "provider", SMOKE_PROVIDER,
        "--namespace", TEST_NAMESPACE, "-o", "json",
    ])
    if assert_exit_ok("describe after patch exits 0", rc, stderr):
        data = assert_valid_json("describe after patch valid json", stdout)
        if data and isinstance(data, dict):
            assert_contains("patch updated URL",
                            json.dumps(data), "smoke-test.example.com")


def _lifecycle_delete_provider():
    """Delete the created provider and verify it's gone."""
    stdout, stderr, rc = run([
        "delete", "provider", "--name", SMOKE_PROVIDER,
        "--namespace", TEST_NAMESPACE,
    ])
    assert_exit_ok("delete provider exits 0", rc, stderr)

    # Verify it's gone
    stdout, stderr, rc = run([
        "get", "providers", "-o", "json",
        "--namespace", TEST_NAMESPACE,
    ])
    if assert_exit_ok("get providers after delete exits 0", rc, stderr):
        data, _ = parse_json(stdout)
        if isinstance(data, list):
            names = [_extract_name(p) for p in data]
            record("deleted provider is gone",
                   SMOKE_PROVIDER not in names,
                   f"provider '{SMOKE_PROVIDER}' still in {names}")


# ---------------------------------------------------------------------------
# Tests: help
# ---------------------------------------------------------------------------

def test_help():
    section("help")

    # 1. Default help
    stdout, stderr, rc = run(["help"])
    if assert_exit_ok("help exits 0", rc, stderr):
        assert_contains("help mentions get", stdout, "get")
        assert_contains("help mentions create", stdout, "create")

    # 2. Help for a specific command
    stdout, stderr, rc = run(["help", "get", "plan"])
    if assert_exit_ok("help get plan exits 0", rc, stderr):
        assert_contains("help get plan mentions migration", stdout, "migration")

    # 3. TSL topic
    stdout, stderr, rc = run(["help", "tsl"])
    if assert_exit_ok("help tsl exits 0", rc, stderr):
        record("help tsl non-empty", len(stdout.strip()) > 0,
               "empty output" if not stdout.strip() else "")

    # 4. KARL topic
    stdout, stderr, rc = run(["help", "karl"])
    if assert_exit_ok("help karl exits 0", rc, stderr):
        record("help karl non-empty", len(stdout.strip()) > 0,
               "empty output" if not stdout.strip() else "")

    # 5. Offload topic
    stdout, stderr, rc = run(["help", "offload"])
    if assert_exit_ok("help offload exits 0", rc, stderr):
        record("help offload non-empty", len(stdout.strip()) > 0,
               "empty output" if not stdout.strip() else "")

    # 6. Machine-readable JSON schema
    stdout, stderr, rc = run(["help", "--machine"])
    if assert_exit_ok("help --machine exits 0", rc, stderr):
        data = assert_valid_json("help --machine valid json", stdout)
        if data and isinstance(data, dict):
            assert_contains("help --machine has commands", json.dumps(data), "commands")
            verbs = {_first_verb(p) for p in _command_paths(data)}
            for verb in EXPECTED_MACHINE_VERBS:
                record(f"help --machine has {verb}",
                       verb in verbs,
                       f"verbs={sorted(verbs)}")

    # 7. Machine-readable YAML schema
    stdout, stderr, rc = run(["help", "--machine", "-o", "yaml"])
    if assert_exit_ok("help --machine yaml exits 0", rc, stderr):
        assert_valid_yaml("help --machine valid yaml", stdout)

    # 8. Machine schema for a specific command
    stdout, stderr, rc = run(["help", "--machine", "get", "plan"])
    if assert_exit_ok("help --machine get plan exits 0", rc, stderr):
        data = assert_valid_json("help --machine get plan json", stdout)
        if data and isinstance(data, dict):
            assert_contains("help --machine get plan has commands",
                            json.dumps(data), "plan")

    # 9. Machine --short (condensed)
    stdout, stderr, rc = run(["help", "--machine", "--short"])
    if assert_exit_ok("help --machine --short exits 0", rc, stderr):
        assert_valid_json("help --machine --short valid json", stdout)

    # 10. Machine --read-only
    stdout, stderr, rc = run(["help", "--machine", "--read-only"])
    if assert_exit_ok("help --machine --read-only exits 0", rc, stderr):
        data = assert_valid_json("help --machine --read-only json", stdout)
        if data and isinstance(data, dict):
            write_found = [
                p for p in _command_paths(data) if _first_verb(p) in WRITE_VERBS
            ]
            record("help --machine --read-only excludes write verbs",
                   len(write_found) == 0,
                   f"write commands: {write_found[:8]}")

    # 11. Machine --write
    stdout, stderr, rc = run(["help", "--machine", "--write"])
    if assert_exit_ok("help --machine --write exits 0", rc, stderr):
        assert_valid_json("help --machine --write json", stdout)

    # 12. Help for TSL topic in machine format
    stdout, stderr, rc = run(["help", "--machine", "tsl"])
    if assert_exit_ok("help --machine tsl exits 0", rc, stderr):
        assert_valid_json("help --machine tsl valid json", stdout)


# ---------------------------------------------------------------------------
# Tests: error handling
# ---------------------------------------------------------------------------

def test_errors():
    section("error handling")

    # 1. get plan --vms without plan name
    _, _, rc = run(["get", "plan", "--vms"])
    assert_exit_fail("get plan --vms missing name", rc)

    # 2. get plan --disk without plan name
    _, _, rc = run(["get", "plan", "--disk"])
    assert_exit_fail("get plan --disk missing name", rc)

    # 3. describe provider without --name
    _, _, rc = run(["describe", "provider"])
    assert_exit_fail("describe provider missing --name", rc)

    # 4. describe plan without --name
    _, _, rc = run(["describe", "plan"])
    assert_exit_fail("describe plan missing --name", rc)

    # 5. create provider without --type
    _, _, rc = run(["create", "provider", "--name", "test-smoke"])
    assert_exit_fail("create provider missing --type", rc)

    # 6. create plan without required flags
    _, _, rc = run(["create", "plan"])
    assert_exit_fail("create plan missing required flags", rc)

    # 7. settings set without --setting
    _, _, rc = run(["settings", "set"])
    assert_exit_fail("settings set missing --setting", rc)

    # 8. settings unset without --setting
    _, _, rc = run(["settings", "unset"])
    assert_exit_fail("settings unset missing --setting", rc)

    # 9. delete plan without --name or --all
    _, _, rc = run(["delete", "plan"])
    assert_exit_fail("delete plan missing --name", rc)

    # 10. delete provider without --name or --all
    _, _, rc = run(["delete", "provider"])
    assert_exit_fail("delete provider missing --name", rc)

    # 11. Unknown subcommand
    _, _, rc = run(["nonexistent-command"])
    assert_exit_fail("unknown subcommand", rc)

    # 12. help --machine conflicting flags
    _, _, rc = run(["help", "--machine", "--read-only", "--write"])
    assert_exit_fail("help --machine --read-only --write conflict", rc)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    if not os.path.isfile(BINARY):
        print(_paint("1;31", "Binary not found: %s" % BINARY))
        print("Run 'make build' first, or use 'make test-e2e'.")
        sys.exit(2)

    bar = _paint("1;36", "=" * 60)
    print(bar)
    print(_paint("1;36", "E2E Smoke Tests"))
    print(bar)

    test_version()
    test_health()
    test_settings()
    test_get_providers()
    test_get_plans()
    test_get_mappings()
    test_get_inventory()
    test_get_optional_resources()
    test_describe()
    test_positional_args()
    test_dry_run_create_provider()
    test_create_patch_delete()
    test_help()
    test_errors()

    print(bar)
    summary = "Results: %d passed, %d failed" % (passed, failed)
    print(_paint("1;31" if failed else "1;32", summary))
    if errors:
        print(_paint("1;31", "Failed tests:"))
        for name in errors:
            print(_paint("31", "  - %s" % name))
    print(bar)

    sys.exit(1 if failed > 0 else 0)


if __name__ == "__main__":
    main()
