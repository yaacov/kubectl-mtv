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


def get_server_path() -> str:
    """Get the absolute path to the MCP server."""
    script_dir = Path(__file__).parent.absolute()
    server_path = script_dir / "kubectl_mtv_server.py"
    return str(server_path)


def generate_cursor_config() -> dict:
    """Generate Cursor MCP configuration."""
    return {
        "mcpServers": {
            "kubectl-mtv": {
                "command": "python",
                "args": [get_server_path()],
                "env": {}
            }
        }
    }


def generate_claude_desktop_config() -> dict:
    """Generate Claude Desktop MCP configuration."""
    return {
        "mcpServers": {
            "kubectl-mtv": {
                "command": "python",
                "args": [get_server_path()],
                "env": {}
            }
        }
    }


def generate_generic_config() -> dict:
    """Generate generic MCP configuration."""
    return {
        "servers": {
            "kubectl-mtv": {
                "command": "python",
                "args": [get_server_path()],
                "env": {}
            }
        }
    }


def save_config(config: dict, filename: str) -> None:
    """Save configuration to a file."""
    with open(filename, 'w') as f:
        json.dump(config, f, indent=2)
    print(f"Generated {filename}")


def main():
    """Generate MCP configuration files."""
    print("Generating MCP configuration files...")
    print(f"Server path: {get_server_path()}")
    print()
    
    # Check if server exists
    if not os.path.exists(get_server_path()):
        print(f"Error: Server file not found at {get_server_path()}")
        sys.exit(1)
    
    # Generate configurations
    configs = [
        (generate_cursor_config(), "cursor-mcp-config-generated.json", "Cursor IDE"),
        (generate_claude_desktop_config(), "claude-desktop-config-generated.json", "Claude Desktop"),
        (generate_generic_config(), "generic-mcp-config-generated.json", "Generic MCP Client"),
    ]
    
    for config, filename, description in configs:
        save_config(config, filename)
    
    print()
    print("Next steps:")
    print("1. Copy the appropriate configuration to your MCP client")
    print("2. For Cursor: Use the content of cursor-mcp-config-generated.json")
    print("3. For Claude Desktop: Copy claude-desktop-config-generated.json content to your Claude config")
    print("4. Restart your MCP client")
    print()
    print("For detailed setup instructions, see MCP_SETUP.md")


if __name__ == "__main__":
    main()