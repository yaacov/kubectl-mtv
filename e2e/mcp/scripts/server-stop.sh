#!/bin/bash
# Stop MCP server (binary mode)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MCP_DIR="$(dirname "$SCRIPT_DIR")"

# Load shared library
# shellcheck disable=SC1091
source "$SCRIPT_DIR/lib.sh"

# Configuration
SERVER_PID_FILE="$MCP_DIR/.server.pid"

# Check if PID file exists
if [ ! -f "$SERVER_PID_FILE" ]; then
    echo "No binary server running (PID file not found)"
    exit 0
fi

PID=$(cat "$SERVER_PID_FILE")

# Check if process is actually running
if ! is_process_running "$PID"; then
    echo "Binary server not running (removing stale PID file)"
    rm -f "$SERVER_PID_FILE"
    exit 0
fi

# Stop the server
echo "Stopping binary server (PID $PID)..."
kill "$PID" 2>/dev/null || true

# Wait up to 10 seconds for graceful shutdown
for i in {1..10}; do
    if ! is_process_running "$PID"; then
        success "Binary server stopped"
        rm -f "$SERVER_PID_FILE"
        exit 0
    fi
    sleep 1
done

# Force kill if still running
echo "Force killing server..."
kill -9 "$PID" 2>/dev/null || true
rm -f "$SERVER_PID_FILE"
success "Binary server stopped (forced)"
exit 0
