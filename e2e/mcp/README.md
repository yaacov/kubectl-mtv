# MCP E2E Test Suite

End-to-end tests for the kubectl-mtv MCP (Model Context Protocol) server.

## Overview

These tests verify the complete MCP server functionality including:
- Provider management (vSphere, OpenShift)
- Inventory queries with TSL (Tree Search Language)
- Host configuration for direct ESXi transfers
- Migration plan creation and management
- Network and storage mappings
- Bearer token authentication

## Prerequisites

1. **Python 3.9+** with pip
2. **kubectl-mtv binary** or container image
3. **Access to a Kubernetes/OpenShift cluster** with Forklift installed
4. **Access to a vSphere environment** with VMs to test

## Quick Start

### 1. Configure Environment

Copy the example environment file and fill in your credentials:

```bash
cd e2e/mcp
cp env.example .env
```

Edit `.env` with your actual values:
- vCenter/vSphere credentials
- Kubernetes/OpenShift API and token
- VM names and resource mappings

### 2. Run Tests (3-Step Workflow)

#### Option A: Automatic (Recommended)

```bash
# Binary mode - start server, run tests, stop server
make test-full

# Container mode - start server, run tests, stop server
make test-full-image MCP_IMAGE=quay.io/yaacov/kubectl-mtv-mcp-server:latest
```

#### Option B: Manual Control

```bash
# Step 1: Start server
make server-start                    # Binary mode
# OR
make server-start-image MCP_IMAGE=quay.io/yaacov/kubectl-mtv-mcp-server:latest

# Step 2: Run tests
make test

# Step 3: Stop server
make server-stop
```

#### Option C: External Server

If you have a server running elsewhere (e.g., in production, different machine):

```bash
# Run tests against remote server
MCP_SSE_URL=http://remote-host:8080/sse make test
```

## Server Management

The Makefile uses helper scripts in `scripts/` for server management. You can call these directly or use the make targets.

### Start Server

**Binary Mode** (local kubectl-mtv):
```bash
make server-start
# or directly: ./scripts/server-start.sh
```

**Container Mode**:
```bash
MCP_IMAGE=quay.io/yaacov/kubectl-mtv-mcp-server:latest make server-start-image
# or directly:
MCP_IMAGE=quay.io/yaacov/kubectl-mtv-mcp-server:latest ./scripts/server-start-image.sh
```

The server will run in the background. Logs are written to `.server.log`.

### Check Server Status

```bash
make server-status
# or directly: ./scripts/server-status.sh
```

Shows whether server is running, PID/container name, and verifies connectivity.

### Stop Server

Each mode has its own stop script:

**Binary Mode**:
```bash
make server-stop
# or directly: ./scripts/server-stop.sh
```

**Container Mode**:
```bash
make server-stop-image
# or directly: ./scripts/server-stop-image.sh
```

**Stop All** (convenience - stops both modes):
```bash
make server-stop-all
```

Each script is mode-specific and silently succeeds if that mode isn't running.

## Running Tests

### Full Test Suite

```bash
make test                    # All tests (requires running server)
```

### By Concern

```bash
make test-setup              # Server connectivity + namespace setup
make test-providers          # Provider create/read
make test-hosts              # ESXi host configuration
make test-plans              # Migration plans
make test-mappings           # Network/storage mappings
make test-inventory          # vSphere inventory queries
make test-health             # MTV health checks
make test-auth               # Authentication/authorization
```

### By Operation Type

```bash
make test-create             # All write operations
make test-read               # All read operations
```

## Configuration

### Environment Variables

#### Required

```bash
# vSphere/vCenter
GOVC_URL=vcenter.example.com
GOVC_USERNAME=administrator@vsphere.local
GOVC_PASSWORD=changeme

# Kubernetes/OpenShift
KUBE_API_URL=https://api.cluster.example.com:6443
KUBE_TOKEN=sha256~xxxx...

# Test Resources
ESXI_HOST_NAME=host-00
COLD_VMS=vm1,vm2
WARM_VMS=vm3,vm4
NETWORK_PAIRS=VM Network:default
STORAGE_PAIRS=datastore-1:storageclass-1
```

#### Optional

```bash
# Server Connection
MCP_SSE_HOST=127.0.0.1               # Server bind address
MCP_SSE_PORT=18443                   # Server port
MCP_SSE_URL=http://...               # Full URL (overrides host/port)

# Server Management
MTV_BINARY=../../kubectl-mtv         # Binary location
MCP_IMAGE=quay.io/...                # Container image
CONTAINER_ENGINE=podman              # docker or podman

# Test Configuration
TEST_NAMESPACE=mcp-e2e-test
MCP_VERBOSE=1                        # 0=silent, 1=info, 2=debug, 3=trace
```

## Server Modes

### Binary Mode (Default)

Runs the local `kubectl-mtv` binary:

```bash
make server-start
# Server: ../../kubectl-mtv mcp-server --sse --port 18443
```

**Use when:**
- Developing locally
- Testing unreleased changes
- Debugging with local builds

### Container Mode

Runs a container image (docker/podman):

```bash
make server-start-image MCP_IMAGE=quay.io/yaacov/kubectl-mtv-mcp-server:latest
```

**Use when:**
- Testing released images
- CI/CD pipelines
- Reproducing production issues

### External/Remote Mode

Connects to an already-running server:

```bash
MCP_SSE_URL=http://remote-host:8080/sse make test
```

**Use when:**
- Testing production deployments
- Multi-environment testing
- Server runs on different machine/cluster

## Troubleshooting

### Server Won't Start

```bash
# Check what's using the port
lsof -i :18443
netstat -tulpn | grep 18443

# Try a different port
export MCP_SSE_PORT=19443
make server-start
```

### Connection Refused

```bash
# Verify server is listening
make server-status

# Check server logs
tail -f .server.log

# For container mode
docker logs mcp-e2e-18443
# or
podman logs mcp-e2e-18443
```

### Authentication Errors

```bash
# Verify token is valid
kubectl --server $KUBE_API_URL --token $KUBE_TOKEN --insecure-skip-tls-verify get namespaces

# Check token hasn't expired
echo $KUBE_TOKEN | cut -d. -f2 | base64 -d 2>/dev/null | jq .exp
```

### Test Failures

```bash
# Run with more verbose output
MCP_VERBOSE=2 make test

# Run specific test
make test PYTEST_ARGS="-v -s -k test_create_vsphere_provider"

# Stop and clean everything
make clean
```

## Cleanup

Remove all test artifacts:

```bash
make clean                   # Stops server, removes venv, logs, caches
```
