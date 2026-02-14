#!/bin/bash
# Stop MCP server (container mode)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MCP_DIR="$(dirname "$SCRIPT_DIR")"

# Load shared library
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib.sh"

# Load environment
load_env "$MCP_DIR"

# Configuration
MCP_SSE_PORT="${MCP_SSE_PORT:-18443}"
CONTAINER_NAME="mcp-e2e-${MCP_SSE_PORT}"

# Detect container engine
if ! ENGINE=$(detect_container_engine); then
    echo "No container engine found (docker or podman)"
    exit 0
fi

# Check if container exists
if ! $ENGINE ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^${CONTAINER_NAME}$"; then
    echo "No container server running (container not found)"
    exit 0
fi

# Stop and remove the container
echo "Stopping container: $CONTAINER_NAME..."
$ENGINE stop -t 5 "$CONTAINER_NAME" >/dev/null 2>&1 || true
$ENGINE rm "$CONTAINER_NAME" >/dev/null 2>&1 || true
success "Container server stopped"
exit 0
