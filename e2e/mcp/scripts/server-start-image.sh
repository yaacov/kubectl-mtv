#!/bin/bash
# Start MCP server in container mode (docker/podman)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MCP_DIR="$(dirname "$SCRIPT_DIR")"

# Load shared library
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib.sh"

# Load environment
load_env "$MCP_DIR"

# Configuration
MCP_SSE_HOST="${MCP_SSE_HOST:-127.0.0.1}"
MCP_SSE_PORT="${MCP_SSE_PORT:-18443}"
MCP_IMAGE="${MCP_IMAGE:-}"
CONTAINER_NAME="mcp-e2e-${MCP_SSE_PORT}"

# Validate required variables
require_env KUBE_API_URL MCP_IMAGE

# Detect container engine
if ! ENGINE=$(detect_container_engine); then
    error "No container engine found (docker or podman)"
    info "Install one or set CONTAINER_ENGINE environment variable"
    exit 1
fi

# Check if container already exists
if $ENGINE ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^${CONTAINER_NAME}$"; then
    # Check if it's running
    if $ENGINE ps --format '{{.Names}}' 2>/dev/null | grep -q "^${CONTAINER_NAME}$"; then
        error "Container already running: $CONTAINER_NAME"
        exit 1
    else
        echo "Removing stopped container: $CONTAINER_NAME"
        $ENGINE rm "$CONTAINER_NAME" >/dev/null
    fi
fi

# Start container
echo "Starting MCP server (container mode)..."
info "Image:    $MCP_IMAGE"
info "Engine:   $ENGINE"
info "Host:     $MCP_SSE_HOST"
info "Port:     $MCP_SSE_PORT"
info "API:      $KUBE_API_URL"

$ENGINE run -d \
    --name "$CONTAINER_NAME" \
    -p "${MCP_SSE_HOST}:${MCP_SSE_PORT}:8080" \
    -e "MCP_KUBE_SERVER=${KUBE_API_URL}" \
    -e "MCP_KUBE_INSECURE=true" \
    -e "MCP_PORT=8080" \
    -e "MCP_HOST=0.0.0.0" \
    "$MCP_IMAGE" >/dev/null

# Verify container is still running
sleep 1
if ! $ENGINE ps --format '{{.Names}}' 2>/dev/null | grep -q "^${CONTAINER_NAME}$"; then
    error "Container stopped unexpectedly"
    info "Check logs: $ENGINE logs $CONTAINER_NAME"
    $ENGINE rm "$CONTAINER_NAME" 2>/dev/null || true
    exit 1
fi

# Wait for container to start listening
if ! wait_for_server "$MCP_SSE_HOST" "$MCP_SSE_PORT" 30 "Container"; then
    error "Container is running but not accepting connections"
    info "Check logs: $ENGINE logs $CONTAINER_NAME"
    exit 1
fi

success "Container started successfully: $CONTAINER_NAME"
info "URL:  http://$MCP_SSE_HOST:$MCP_SSE_PORT/sse"
info "Logs: $ENGINE logs -f $CONTAINER_NAME"
exit 0
