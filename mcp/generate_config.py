#!/usr/bin/env python3
"""
Configuration generator for kubectl-mtv MCP server

This script generates the correct MCP configuration files with the absolute path
to the kubectl-mtv MCP server for your specific installation.
"""

import json
import os
import sys
from pathlib import Path


def get_server_path(server_type: str = "read") -> str:
    """Get the absolute path to the MCP server."""
    script_dir = Path(__file__).parent.absolute()
    if server_type == "read":
        server_path = script_dir / "kubev2v" / "kubectl_mtv_server.py"
    elif server_type == "write":
        server_path = script_dir / "kubev2v" / "kubectl_mtv_write_server.py"
    else:
        raise ValueError(f"Invalid server_type: {server_type}. Use 'read' or 'write'")
    return str(server_path)


def generate_cursor_config() -> dict:
    """Generate Cursor MCP configuration."""
    return {
        "mcpServers": {
            "kubectl-mtv": {
                "command": "python",
                "args": [get_server_path("read")],
                "env": {},
            },
            "kubectl-mtv-write": {
                "command": "python",
                "args": [get_server_path("write")],
                "env": {},
            },
        }
    }


def generate_claude_desktop_config() -> dict:
    """Generate Claude Desktop MCP configuration."""
    return {
        "mcpServers": {
            "kubectl-mtv": {
                "command": "python",
                "args": [get_server_path("read")],
                "env": {},
            },
            "kubectl-mtv-write": {
                "command": "python",
                "args": [get_server_path("write")],
                "env": {},
            },
        }
    }


def generate_generic_config() -> dict:
    """Generate generic MCP configuration."""
    return {
        "servers": {
            "kubectl-mtv": {
                "command": "python",
                "args": [get_server_path("read")],
                "env": {},
            },
            "kubectl-mtv-write": {
                "command": "python",
                "args": [get_server_path("write")],
                "env": {},
            },
        }
    }


def generate_http_server_config(host: str = "127.0.0.1", port: int = 8000) -> dict:
    """Generate HTTP server configuration for running MCP as a web service."""
    return {
        "mcpServers": {
            "kubectl-mtv": {
                "transport": "http",
                "url": f"http://{host}:{port}/mcp",
                "description": "kubectl-mtv MCP server (read-only) running as HTTP service",
            },
            "kubectl-mtv-write": {
                "transport": "http",
                "url": f"http://{host}:{port + 1}/mcp",
                "description": "kubectl-mtv MCP server (write operations) running as HTTP service",
            },
        }
    }


def save_config(config: dict, filename: str) -> None:
    """Save configuration to a file."""
    with open(filename, "w") as f:
        json.dump(config, f, indent=2)
    print(f"Generated {filename}")


def main():
    """Generate MCP configuration files."""
    print("Generating MCP configuration files...")
    print(f"Read server path: {get_server_path('read')}")
    print(f"Write server path: {get_server_path('write')}")
    print()

    # Check if both servers exist
    for server_type in ["read", "write"]:
        server_path = get_server_path(server_type)
        if not os.path.exists(server_path):
            print(
                f"Error: {server_type.capitalize()} server file not found at {server_path}"
            )
            sys.exit(1)

    # Generate configurations
    configs = [
        (generate_cursor_config(), "cursor-mcp-config-generated.json", "Cursor IDE"),
        (
            generate_claude_desktop_config(),
            "claude-desktop-config-generated.json",
            "Claude Desktop",
        ),
        (
            generate_generic_config(),
            "generic-mcp-config-generated.json",
            "Generic MCP Client",
        ),
        (
            generate_http_server_config(),
            "http-server-config-generated.json",
            "HTTP Server Client",
        ),
    ]

    for config, filename, description in configs:
        save_config(config, filename)

    print()
    print("Next steps:")
    print("1. Copy the appropriate configuration to your MCP client")
    print("2. For Cursor: Use the content of cursor-mcp-config-generated.json")
    print(
        "3. For Claude Desktop: Copy claude-desktop-config-generated.json content to your Claude config"
    )
    print(
        "4. For HTTP server: Use http-server-config-generated.json and start both servers with:"
    )
    print(
        "   python kubev2v/kubectl_mtv_server.py --transport http --host 127.0.0.1 --port 8000"
    )
    print(
        "   python kubev2v/kubectl_mtv_write_server.py --transport http --host 127.0.0.1 --port 8001"
    )
    print("5. Alternatively, if installed via pip (mtv-mcp package):")
    print("   kubectl-mtv-mcp --transport http --host 127.0.0.1 --port 8000")
    print("   kubectl-mtv-write-mcp --transport http --host 127.0.0.1 --port 8001")
    print("6. Restart your MCP client")
    print()
    print("For detailed setup instructions, see MCP_SETUP.md")


if __name__ == "__main__":
    main()
