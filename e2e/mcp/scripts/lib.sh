#!/bin/bash
# Shared library functions for MCP server management scripts

# Load environment from .env if present
load_env() {
    local mcp_dir="$1"
    if [ -f "$mcp_dir/.env" ]; then
        set -a
        # shellcheck disable=SC1091
        source "$mcp_dir/.env"
        set +a
    fi
}

# Check if a port is listening on a host
# Usage: check_port_listening HOST PORT
# Returns: 0 if listening, 1 if not
check_port_listening() {
    local host="$1"
    local port="$2"
    
    if command -v nc >/dev/null 2>&1; then
        nc -z "$host" "$port" 2>/dev/null
        return $?
    elif command -v timeout >/dev/null 2>&1; then
        timeout 1 bash -c "cat < /dev/null > /dev/tcp/$host/$port" 2>/dev/null
        return $?
    else
        # Fallback: try bash TCP redirection without timeout
        (echo >/dev/tcp/"$host"/"$port") 2>/dev/null
        return $?
    fi
}

# Wait for server to start listening on a port
# Usage: wait_for_server HOST PORT TIMEOUT_SECONDS DESCRIPTION
# Returns: 0 if server started, 1 if timeout
wait_for_server() {
    local host="$1"
    local port="$2"
    local timeout="$3"
    local description="$4"
    
    echo "Waiting for server to start listening..."
    
    for i in $(seq 1 "$timeout"); do
        if check_port_listening "$host" "$port"; then
            return 0
        fi
        sleep 1
    done
    
    echo "✗ $description failed to start listening within $timeout seconds" >&2
    return 1
}

# Validate required environment variables
# Usage: require_env VAR_NAME [VAR_NAME...]
# Exits with error if any variable is not set
require_env() {
    local missing=()
    for var in "$@"; do
        if [ -z "${!var:-}" ]; then
            missing+=("$var")
        fi
    done
    
    if [ ${#missing[@]} -gt 0 ]; then
        echo "Error: Required environment variables not set:" >&2
        for var in "${missing[@]}"; do
            echo "  - $var" >&2
        done
        exit 1
    fi
}

# Detect container engine (docker or podman)
# Usage: ENGINE=$(detect_container_engine)
detect_container_engine() {
    if [ -n "${CONTAINER_ENGINE:-}" ]; then
        echo "$CONTAINER_ENGINE"
    elif command -v docker >/dev/null 2>&1; then
        echo "docker"
    elif command -v podman >/dev/null 2>&1; then
        echo "podman"
    else
        echo "" >&2
        return 1
    fi
}

# Check if a process is running by PID
# Usage: is_process_running PID
is_process_running() {
    local pid="$1"
    kill -0 "$pid" 2>/dev/null
}

# Print success message with checkmark
success() {
    echo "✓ $*"
}

# Print error message with X
error() {
    echo "✗ $*" >&2
}

# Print info message with bullet
info() {
    echo "  $*"
}
