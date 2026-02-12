#!/usr/bin/env python3
"""
Dump the tools/list MCP response verbatim from a running MCP server.

Supports:
  - Stdio mode: spawn the server as subprocess, communicate via stdin/stdout
  - SSE mode: connect to an already-running HTTP SSE server

Usage:
  # Stdio - spawn kubectl-mtv MCP server
  ./dump-mcp-tools.py --stdio kubectl mtv mcp-server

  # SSE - connect to running server
  ./dump-mcp-tools.py --url http://127.0.0.1:8080/sse

  # With optional auth headers (SSE)
  ./dump-mcp-tools.py --url http://127.0.0.1:8080/sse \\
    --header "Authorization: Bearer TOKEN" \\
    --header "X-Kubernetes-Server: https://..."
"""

import argparse
import json
import queue
import subprocess
import sys
import threading
import urllib.request
import urllib.error
import urllib.parse
import ssl


def parse_sse_events(stream):
    """Parse SSE events from a stream. Yields (event_name, data)."""
    buf = b""
    while True:
        while b"\n\n" not in buf:
            chunk = stream.read(256)
            if not chunk:
                if buf.strip():
                    buf += b"\n\n"
                break
            buf += chunk
        if not buf:
            break
        msg, _, buf = buf.partition(b"\n\n")
        event_name = ""
        data_lines = []
        for line in msg.split(b"\n"):
            line = line.strip()
            if line.startswith(b"event: "):
                event_name = line[7:].decode("utf-8").strip()
            elif line.startswith(b"data: "):
                data_lines.append(line[6:].decode("utf-8"))
        if data_lines:
            yield event_name, "\n".join(data_lines)


def send_stdio(proc, msg: dict) -> dict:
    """Send JSON-RPC message to stdio server and read response."""
    line = json.dumps(msg) + "\n"
    proc.stdin.write(line)
    proc.stdin.flush()
    response_line = proc.stdout.readline()
    if not response_line:
        raise RuntimeError("Server closed connection without response")
    return json.loads(response_line)


def main():
    parser = argparse.ArgumentParser(
        description="Dump tools/list MCP response verbatim from an MCP server",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument(
        "--stdio",
        nargs="+",
        metavar="CMD",
        help="Stdio mode: spawn server (e.g. kubectl mtv mcp-server)",
    )
    group.add_argument(
        "--url",
        metavar="URL",
        help="SSE mode: URL of running MCP server (e.g. http://127.0.0.1:8080/sse)",
    )
    parser.add_argument(
        "--header",
        action="append",
        metavar="HEADER",
        default=[],
        help="HTTP header for SSE (e.g. 'Authorization: Bearer TOKEN'). Can repeat.",
    )
    parser.add_argument(
        "--insecure",
        action="store_true",
        help="Skip TLS verification for HTTPS SSE URLs",
    )
    parser.add_argument(
        "--raw",
        action="store_true",
        help="Output raw JSON without pretty-printing (single line)",
    )
    args = parser.parse_args()

    # MCP protocol version
    protocol_version = "2024-11-05"

    # Initialize request
    init_req = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "initialize",
        "params": {
            "protocolVersion": protocol_version,
            "capabilities": {"roots": {"listChanged": True}, "sampling": {}},
            "clientInfo": {"name": "dump-mcp-tools", "version": "1.0.0"},
        },
    }

    # Initialized notification (no id)
    initialized_notif = {
        "jsonrpc": "2.0",
        "method": "notifications/initialized",
    }

    # Tools/list request
    tools_list_req = {
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/list",
        "params": {},
    }

    if args.stdio:
        # Stdio mode: spawn process
        proc = subprocess.Popen(
            args.stdio,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        try:
            # Initialize
            init_resp = send_stdio(proc, init_req)
            if "error" in init_resp:
                print("Initialize error:", json.dumps(init_resp, indent=2), file=sys.stderr)
                sys.exit(1)

            # Send initialized notification
            proc.stdin.write(json.dumps(initialized_notif) + "\n")
            proc.stdin.flush()

            # Tools/list
            tools_resp = send_stdio(proc, tools_list_req)
        finally:
            proc.terminate()
            try:
                proc.wait(timeout=2)
            except subprocess.TimeoutExpired:
                proc.kill()
                proc.wait(timeout=10)

    else:
        # SSE mode: MCP Go SDK uses GET for SSE stream, POST returns 202 - responses come via SSE
        base_url = args.url.rstrip("/")
        if not base_url.endswith("/sse"):
            sse_url = base_url + "/sse"
        else:
            sse_url = base_url

        # Build headers
        headers = {"Accept": "text/event-stream"}
        for h in args.header:
            if ":" in h:
                k, v = h.split(":", 1)
                headers[k.strip()] = v.strip()

        ctx = ssl.create_default_context()
        if args.insecure:
            ctx.check_hostname = False
            ctx.verify_mode = ssl.CERT_NONE

        # Step 1: GET to establish SSE stream, read endpoint event
        get_req = urllib.request.Request(sse_url, headers=headers, method="GET")
        try:
            resp = urllib.request.urlopen(get_req, context=ctx, timeout=30)
        except urllib.error.HTTPError as e:
            print(f"HTTP error {e.code}: {e.reason}", file=sys.stderr)
            sys.exit(1)
        except urllib.error.URLError as e:
            print(f"URL error: {e.reason}", file=sys.stderr)
            sys.exit(1)

        # Parse first SSE event: endpoint (data is raw path like /?sessionid=xxx)
        message_endpoint = None
        response_queue = queue.Queue()

        def read_sse():
            for evt_name, evt_data in parse_sse_events(resp):
                response_queue.put((evt_name, evt_data))

        reader = threading.Thread(target=read_sse, daemon=True)
        reader.start()

        # Get endpoint from first event (data is path, not JSON)
        try:
            evt_name, evt_data = response_queue.get(timeout=10)
            if evt_name != "endpoint":
                print(f"Expected endpoint event, got {evt_name!r}", file=sys.stderr)
                sys.exit(1)
            # Resolve path against base URL (evt_data is e.g. /sse?sessionid=xxx)
            message_endpoint = urllib.parse.urljoin(sse_url, evt_data)
        except queue.Empty:
            print("Timeout waiting for endpoint event", file=sys.stderr)
            sys.exit(1)

        def post_message(msg: dict) -> None:
            """POST message; response will arrive via SSE stream."""
            data = json.dumps(msg).encode()
            req = urllib.request.Request(
                message_endpoint,
                data=data,
                headers={**headers, "Content-Type": "application/json"},
                method="POST",
            )
            with urllib.request.urlopen(req, context=ctx, timeout=30) as r:
                status = getattr(r, "status", r.getcode())
                if status not in (200, 202):
                    raise RuntimeError(f"POST failed: {status}")

        def wait_response(expected_id):
            """Wait for JSON-RPC response with matching id."""
            while True:
                try:
                    evt_name, evt_data = response_queue.get(timeout=30)
                except queue.Empty:
                    raise RuntimeError("Timeout waiting for response") from None
                if evt_name == "message":
                    msg = json.loads(evt_data)
                    if msg.get("id") == expected_id:
                        return msg
                    # Logging/other events; continue waiting

        # Initialize
        post_message(init_req)
        init_resp = wait_response(1)
        if "error" in init_resp:
            print("Initialize error:", json.dumps(init_resp, indent=2), file=sys.stderr)
            sys.exit(1)

        # Initialized notification (no response expected)
        post_message(initialized_notif)

        # Tools/list
        post_message(tools_list_req)
        tools_resp = wait_response(2)

    # Output tools/list response verbatim
    if args.raw:
        print(json.dumps(tools_resp))
    else:
        print(json.dumps(tools_resp, indent=2))


if __name__ == "__main__":
    main()
