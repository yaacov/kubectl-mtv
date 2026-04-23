#!/usr/bin/env python3
"""
Dump the tools/list MCP response verbatim from a running MCP server.

Supports:
  - Stdio mode: spawn the server as subprocess, communicate via stdin/stdout
  - HTTP mode: connect to an already-running Streamable HTTP server

Usage:
  # Stdio - spawn kubectl-mtv MCP server
  ./dump-mcp-tools.py --stdio kubectl mtv mcp-server

  # HTTP - connect to running server
  ./dump-mcp-tools.py --url http://127.0.0.1:8080/mcp

  # With optional auth headers (HTTP)
  ./dump-mcp-tools.py --url http://127.0.0.1:8080/mcp \\
    --header "Authorization: Bearer TOKEN" \\
    --header "X-Kubernetes-Server: https://..."
"""

import argparse
import json
import subprocess
import sys
import urllib.parse
import urllib.request
import urllib.error
import ssl


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
        help="HTTP mode: URL of running MCP server (e.g. http://127.0.0.1:8080/mcp)",
    )
    parser.add_argument(
        "--header",
        action="append",
        metavar="HEADER",
        default=[],
        help="HTTP header (e.g. 'Authorization: Bearer TOKEN'). Can repeat.",
    )
    parser.add_argument(
        "--insecure",
        action="store_true",
        help="Skip TLS verification for HTTPS URLs",
    )
    parser.add_argument(
        "--raw",
        action="store_true",
        help="Output raw JSON without pretty-printing (single line)",
    )
    args = parser.parse_args()

    # MCP protocol version
    protocol_version = "2025-03-26"

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
        # Streamable HTTP mode: POST JSON-RPC to the /mcp endpoint
        url = args.url.rstrip("/")
        if not url.endswith("/mcp"):
            url = url + "/mcp"

        parsed_url = urllib.parse.urlparse(url)
        if parsed_url.scheme not in ("http", "https"):
            print(f"Invalid URL scheme {parsed_url.scheme!r}: must be http or https", file=sys.stderr)
            sys.exit(1)

        # Build headers -- Accept both JSON and SSE per the Streamable HTTP spec
        headers = {
            "Content-Type": "application/json",
            "Accept": "application/json, text/event-stream",
        }
        for h in args.header:
            if ":" in h:
                k, v = h.split(":", 1)
                headers[k.strip()] = v.strip()

        ctx = ssl.create_default_context()
        if args.insecure:
            ctx.check_hostname = False
            ctx.verify_mode = ssl.CERT_NONE

        session_id = None

        def post_message(msg: dict) -> dict:
            """POST a JSON-RPC message and return the parsed JSON response."""
            nonlocal session_id
            req_headers = dict(headers)
            if session_id:
                req_headers["Mcp-Session-Id"] = session_id

            data = json.dumps(msg).encode()
            req = urllib.request.Request(url, data=data, headers=req_headers, method="POST")
            try:
                with urllib.request.urlopen(req, context=ctx, timeout=30) as r:
                    sid = r.headers.get("Mcp-Session-Id")
                    if sid:
                        session_id = sid

                    content_type = r.headers.get("Content-Type", "")
                    body = r.read().decode("utf-8")

                    if content_type.startswith("application/json"):
                        return json.loads(body) if body.strip() else {}
                    if content_type.startswith("text/event-stream"):
                        for line in body.splitlines():
                            if line.startswith("data:"):
                                payload = line[len("data:"):].strip()
                                if payload:
                                    return json.loads(payload)
                        return {}
                    # Unknown content type -- attempt JSON parse as fallback
                    return json.loads(body) if body.strip() else {}
            except urllib.error.HTTPError as e:
                body = e.read().decode("utf-8", errors="replace")
                print(f"HTTP error {e.code}: {e.reason}\n{body}", file=sys.stderr)
                sys.exit(1)

        # Initialize
        init_resp = post_message(init_req)
        if "error" in init_resp:
            print("Initialize error:", json.dumps(init_resp, indent=2), file=sys.stderr)
            sys.exit(1)

        # Initialized notification (may return empty)
        post_message(initialized_notif)

        # Tools/list
        tools_resp = post_message(tools_list_req)

    # Output tools/list response verbatim
    if args.raw:
        print(json.dumps(tools_resp))
    else:
        print(json.dumps(tools_resp, indent=2))


if __name__ == "__main__":
    main()
