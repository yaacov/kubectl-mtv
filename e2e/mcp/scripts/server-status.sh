#!/bin/bash
# Check MCP server status

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
SERVER_PID_FILE="$MCP_DIR/.server.pid"
CONTAINER_NAME="mcp-e2e-${MCP_SSE_PORT}"

RUNNING=0

# Check binary server
if [ -f "$SERVER_PID_FILE" ]; then
    PID=$(cat "$SERVER_PID_FILE")
    if is_process_running "$PID"; then
        success "Binary server running"
        info "PID:  $PID"
        info "Log:  $MCP_DIR/.server.log"
        info "URL:  http://$MCP_SSE_HOST:$MCP_SSE_PORT/sse"
        RUNNING=1
    else
        error "Binary server not running (stale PID file)"
        rm -f "$SERVER_PID_FILE"
    fi
fi

# Check container
if ENGINE=$(detect_container_engine); then
    if $ENGINE ps --format '{{.Names}}' 2>/dev/null | grep -q "^${CONTAINER_NAME}$"; then
        STATUS=$($ENGINE inspect --format='{{.State.Status}}' "$CONTAINER_NAME" 2>/dev/null || echo "unknown")
        success "Container server running"
        info "Name:   $CONTAINER_NAME"
        info "Status: $STATUS"
        info "URL:    http://$MCP_SSE_HOST:$MCP_SSE_PORT/sse"
        info "Logs:   $ENGINE logs -f $CONTAINER_NAME"
        RUNNING=1
    fi
fi

if [ $RUNNING -eq 0 ]; then
    error "No server running"
    echo ""
    echo "Start a server with:"
    info "make server-start          # Binary mode"
    info "make server-start-image    # Container mode"
    exit 1
fi

# Check if server is actually listening
echo ""
echo "Connectivity check..."
if check_port_listening "$MCP_SSE_HOST" "$MCP_SSE_PORT"; then
    success "Server is listening on $MCP_SSE_HOST:$MCP_SSE_PORT"
else
    error "Server is not listening on $MCP_SSE_HOST:$MCP_SSE_PORT"
    exit 1
fi

exit 0
