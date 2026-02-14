#!/bin/bash
# Start MCP server in binary mode (background process)

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
MTV_BINARY="${MTV_BINARY:-$MCP_DIR/../../kubectl-mtv}"
SERVER_PID_FILE="$MCP_DIR/.server.pid"
SERVER_LOG_FILE="$MCP_DIR/.server.log"

# Validate required variables
require_env KUBE_API_URL

# Check if server already running
if [ -f "$SERVER_PID_FILE" ]; then
    PID=$(cat "$SERVER_PID_FILE")
    if is_process_running "$PID"; then
        error "Server already running (PID $PID)"
        exit 1
    else
        echo "Removing stale PID file"
        rm -f "$SERVER_PID_FILE"
    fi
fi

# Check binary exists
if [ ! -x "$MTV_BINARY" ]; then
    error "kubectl-mtv binary not found or not executable: $MTV_BINARY"
    info "Build it with: make build"
    exit 1
fi

# Start server
echo "Starting MCP server (binary mode)..."
info "Binary: $MTV_BINARY"
info "Host:   $MCP_SSE_HOST"
info "Port:   $MCP_SSE_PORT"
info "API:    $KUBE_API_URL"

# Start server in background with no ambient credentials
KUBECONFIG=/dev/null "$MTV_BINARY" mcp-server \
    --sse \
    --port "$MCP_SSE_PORT" \
    --host "$MCP_SSE_HOST" \
    --server "$KUBE_API_URL" \
    --insecure-skip-tls-verify \
    > "$SERVER_LOG_FILE" 2>&1 &

SERVER_PID=$!
echo "$SERVER_PID" > "$SERVER_PID_FILE"

# Verify process started and is still running
sleep 1
if ! is_process_running "$SERVER_PID"; then
    error "Server process died"
    info "Check log: $SERVER_LOG_FILE"
    rm -f "$SERVER_PID_FILE"
    exit 1
fi

# Wait for server to start listening
if ! wait_for_server "$MCP_SSE_HOST" "$MCP_SSE_PORT" 30 "Server"; then
    error "Server is running but not accepting connections"
    info "Check log: $SERVER_LOG_FILE"
    kill "$SERVER_PID" 2>/dev/null || true
    rm -f "$SERVER_PID_FILE"
    exit 1
fi

success "Server started successfully (PID $SERVER_PID)"
info "Log: $SERVER_LOG_FILE"
info "URL: http://$MCP_SSE_HOST:$MCP_SSE_PORT/sse"
exit 0
