---
layout: page
title: "Chapter 22: Model Context Protocol (MCP) Server Integration"
---

The Model Context Protocol (MCP) server provides AI assistants with comprehensive access to Forklift migration resources, enabling intelligent automation and assistance for migration planning, execution, and troubleshooting. This chapter covers complete MCP integration from basic setup to advanced AI-assisted workflows.

## Overview: Providing AI Assistants Access to Migration Resources

### What is the MCP Server?

The kubectl-mtv MCP server is a built-in component that exposes migration toolkit functionality through the Model Context Protocol, allowing AI assistants to:

- **Read Operations**: List resources, query inventory, monitor migrations, analyze configurations
- **Write Operations**: Create, delete, and patch providers, plans, mappings, hosts, and hooks
- **Real-time Monitoring**: Watch migration progress, analyze logs, troubleshoot issues
- **Intelligent Planning**: Assist with migration strategy, resource optimization, and best practices

### Supported AI Assistants

The MCP server integrates with various AI platforms:

- **Claude Desktop**: Direct integration via configuration file or CLI
- **Cursor IDE**: Built-in MCP support for development workflows  
- **Custom Tools**: Any MCP-compatible client using the MCP SDK
- **Web Applications**: HTTP-based integrations via SSE mode

### Security Model

**Important Security Notice**: The MCP server provides both read and write access to migration resources. Use with appropriate security considerations:

- Authentication is passed through from the kubectl context
- All RBAC permissions apply to MCP operations
- Write operations can modify cluster resources
- Network access follows Kubernetes cluster policies

## Server Modes

### Stdio Mode (Default)

Stdio mode is designed for direct AI assistant integration via standard input/output streams. This is the recommended mode for most AI integrations.

#### Basic Stdio Setup

```bash
# Start MCP server in stdio mode (default)
kubectl mtv mcp-server
```

#### Stdio Mode Characteristics

- **Communication**: Uses stdin/stdout for MCP protocol messages
- **Security**: Inherits kubectl authentication and permissions
- **Performance**: Direct process communication for low latency
- **Use Cases**: AI assistant integration, local development, automated tools
- **Logging**: Server status logged to stderr

#### Stdio Mode Output Example

```
Starting kubectl-mtv MCP server in stdio mode
Server is ready and listening for MCP protocol messages on stdin/stdout
```

### SSE Mode (HTTP Server)

Server-Sent Events (SSE) mode runs an HTTP server providing MCP access over HTTP. This enables web-based integrations and remote access scenarios.

#### Basic SSE Setup

```bash
# Start MCP server in SSE mode
kubectl mtv mcp-server --sse --host 127.0.0.1 --port 8080
```

#### Advanced SSE Configuration

```bash
# SSE mode with custom host and port
kubectl mtv mcp-server --sse --host 0.0.0.0 --port 9090

# SSE mode with TLS encryption
kubectl mtv mcp-server --sse \
  --host 0.0.0.0 \
  --port 8443 \
  --cert-file /path/to/server.crt \
  --key-file /path/to/server.key
```

#### SSE Mode Characteristics

- **Communication**: HTTP/HTTPS with Server-Sent Events
- **Security**: Optional TLS encryption, bearer token authentication
- **Performance**: Network-based with HTTP overhead
- **Use Cases**: Web applications, remote access, multi-user scenarios
- **Endpoints**: `/sse` endpoint for MCP protocol communication

#### SSE Mode HTTP Headers

In SSE mode, the following HTTP headers are supported for Kubernetes authentication:

| Header | Description |
|--------|-------------|
| `Authorization: Bearer <token>` | Kubernetes authentication token (passed to kubectl via `--token` flag) |
| `X-Kubernetes-Server: <url>` | Kubernetes API server URL (passed to kubectl via `--server` flag) |

If headers are not provided, the server falls back to the default kubeconfig behavior.

## Command Line Options

### Complete Flag Reference

All MCP server flags are verified from the implementation:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--sse` | boolean | `false` | Run in SSE (Server-Sent Events) mode over HTTP |
| `--host` | string | `127.0.0.1` | Host address to bind to for SSE mode |
| `--port` | string | `8080` | Port to listen on for SSE mode |
| `--cert-file` | string | `""` | Path to TLS certificate file (enables TLS when used with --key-file) |
| `--key-file` | string | `""` | Path to TLS private key file (enables TLS when used with --cert-file) |

### Usage Examples

#### Development Mode

```bash
# Local development with default settings
kubectl mtv mcp-server

# Development with HTTP access
kubectl mtv mcp-server --sse --host 127.0.0.1 --port 8080
```

#### Production Mode

```bash
# Production with TLS encryption
kubectl mtv mcp-server --sse \
  --host 0.0.0.0 \
  --port 443 \
  --cert-file /etc/ssl/certs/mcp-server.crt \
  --key-file /etc/ssl/private/mcp-server.key

# Production with custom port and security
kubectl mtv mcp-server --sse \
  --host 10.0.0.100 \
  --port 9443 \
  --cert-file /secure/certificates/server.crt \
  --key-file /secure/certificates/server.key
```

#### Testing and Integration

```bash
# Test connectivity
curl -N http://127.0.0.1:8080/sse

# Test with authentication token
curl -N -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8080/sse

# Test with token and specific API server
curl -N \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Kubernetes-Server: https://api.example.com:6443" \
  http://127.0.0.1:8080/sse

# Validate TLS configuration
openssl s_client -connect 127.0.0.1:8443 -servername mcp-server
```

## How-To: Integrating with AI Assistants

### Claude Desktop Integration

#### Method 1: Using Claude CLI (Recommended)

The Claude CLI provides the simplest integration method:

```bash
# Install Claude CLI (if not already installed)
# Follow instructions at https://claude.ai/cli

# Add kubectl-mtv MCP server
claude mcp add kubectl-mtv kubectl mtv mcp-server

# Verify installation
claude mcp list
```

#### Method 2: Manual Configuration

For environments without the Claude CLI, configure manually:

**Step 1: Locate Claude Desktop Configuration**

```bash
# macOS
open ~/Library/Application\ Support/Claude/

# Linux
ls ~/.config/Claude/

# Windows (PowerShell)
explorer $env:APPDATA\Claude\
```

**Step 2: Edit Configuration File**

Create or edit `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "kubectl-mtv": {
      "command": "kubectl",
      "args": ["mtv", "mcp-server"],
      "env": {
        "KUBECONFIG": "/path/to/your/kubeconfig"
      }
    }
  }
}
```

**Step 3: Advanced Claude Configuration**

```json
{
  "mcpServers": {
    "kubectl-mtv-prod": {
      "command": "kubectl",
      "args": ["mtv", "mcp-server"],
      "env": {
        "KUBECONFIG": "/secure/kubeconfig/prod-cluster.yaml",
        "MTV_VDDK_INIT_IMAGE": "registry.company.com/vddk:latest"
      }
    },
    "kubectl-mtv-dev": {
      "command": "kubectl", 
      "args": ["mtv", "mcp-server"],
      "env": {
        "KUBECONFIG": "/dev/kubeconfig/dev-cluster.yaml"
      }
    }
  }
}
```

**Step 4: Restart and Verify**

```bash
# Restart Claude Desktop application

# Test configuration by asking Claude:
# "List all MTV providers in the cluster"
# "Show me the status of migration plans"
# "Help me create a migration plan for VMware VMs"
```

### Cursor IDE Integration

#### Basic Cursor Setup

**Step 1: Access MCP Settings**

1. Open Cursor IDE
2. Navigate to Settings (Cmd/Ctrl + ,)
3. Search for "MCP" or find MCP in extensions/features
4. Click "Add Server" or "Configure MCP"

**Step 2: Add kubectl-mtv Server**

Configure the MCP server:

- **Name**: `kubectl-mtv`
- **Command**: `kubectl`
- **Args**: `mtv mcp-server`

**Step 3: Advanced Cursor Configuration**

For production environments with specific requirements:

- **Name**: `kubectl-mtv-production`
- **Command**: `kubectl`  
- **Args**: `mtv mcp-server`
- **Environment Variables**:
  - `KUBECONFIG`: `/path/to/prod-kubeconfig.yaml`
  - `MTV_VDDK_INIT_IMAGE`: `registry.company.com/vddk:8.0.2`

**Step 4: Verify Integration**

Test the integration within Cursor:

```bash
# Ask in Cursor chat:
# "Show me all migration providers"
# "Create a migration plan for these VMs: web-01, web-02"
# "Check the status of plan production-migration"
```

### Custom Integration Development

#### Python MCP Client (Stdio Mode)

```python
#!/usr/bin/env python3
"""
Custom MCP client for kubectl-mtv integration
"""

import asyncio
import json
from mcp import StdioClient, ClientSession

async def main():
    # Create stdio client
    client = StdioClient(
        command="kubectl",
        args=["mtv", "mcp-server"]
    )
    
    async with client:
        session = await client.create_session()
        
        # List available tools
        tools = await session.list_tools()
        print("Available tools:", [tool.name for tool in tools])
        
        # Call a specific tool
        result = await session.call_tool(
            "get_providers",
            arguments={"namespace": "migrations"}
        )
        
        print("Providers:", json.dumps(result.content, indent=2))

# Run the client
if __name__ == "__main__":
    asyncio.run(main())
```

#### Python MCP Client (SSE Mode)

```python
#!/usr/bin/env python3
"""
HTTP-based MCP client for kubectl-mtv integration
"""

import asyncio
import json
import os
from mcp import SSEClient, ClientSession

async def main():
    # Get authentication credentials
    token = os.getenv("KUBERNETES_TOKEN")
    server = os.getenv("KUBERNETES_SERVER")  # Optional: API server URL
    
    # Build headers for authentication
    headers = {}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    if server:
        headers["X-Kubernetes-Server"] = server
    
    # Create SSE client
    client = SSEClient(
        url="http://127.0.0.1:8080/sse",
        headers=headers if headers else None
    )
    
    async with client:
        session = await client.create_session()
        
        # Monitor migration plans
        result = await session.call_tool(
            "get_plans",
            arguments={"watch": True}
        )
        
        print("Migration plans:", json.dumps(result.content, indent=2))

# Run the client
if __name__ == "__main__":
    asyncio.run(main())
```

#### JavaScript/TypeScript Integration

```typescript
// MCP client for web applications
import { SSEClient, ClientSession } from '@modelcontextprotocol/sdk';

interface KubeCredentials {
    token?: string;
    server?: string;
}

class MigrationDashboard {
    private client: SSEClient;
    private session: ClientSession | null = null;

    constructor(serverUrl: string, creds?: KubeCredentials) {
        // Build headers for Kubernetes authentication
        const headers: Record<string, string> = {};
        if (creds?.token) {
            headers['Authorization'] = `Bearer ${creds.token}`;
        }
        if (creds?.server) {
            headers['X-Kubernetes-Server'] = creds.server;
        }
        
        this.client = new SSEClient({
            url: serverUrl,
            headers: Object.keys(headers).length > 0 ? headers : undefined
        });
    }

    async connect(): Promise<void> {
        await this.client.connect();
        this.session = await this.client.createSession();
    }

    async getProviders(): Promise<any[]> {
        if (!this.session) throw new Error("Not connected");
        
        const result = await this.session.callTool(
            "get_providers",
            { namespace: "migrations" }
        );
        
        return result.content;
    }

    async createPlan(planConfig: any): Promise<any> {
        if (!this.session) throw new Error("Not connected");
        
        const result = await this.session.callTool(
            "create_plan",
            planConfig
        );
        
        return result.content;
    }

    async monitorMigration(planName: string): Promise<void> {
        if (!this.session) throw new Error("Not connected");
        
        // Watch migration progress
        await this.session.callTool(
            "watch_plan",
            { name: planName, namespace: "migrations" }
        );
    }
}

// Usage
const dashboard = new MigrationDashboard(
    "https://mcp-server.company.com:8443/sse",
    {
        token: process.env.KUBERNETES_TOKEN,
        server: process.env.KUBERNETES_SERVER  // Optional: for remote cluster access
    }
);

await dashboard.connect();
const providers = await dashboard.getProviders();
console.log("Available providers:", providers);
```

## AI-Assisted Migration Workflows

### Claude-Assisted Migration Planning

#### Interactive Migration Discovery

Ask Claude to help with migration planning:

```
Claude, help me plan a migration from VMware to OpenShift. Here's what I need:

1. List all VMware providers in my cluster
2. Show me VMs that are powered on and have more than 8GB memory
3. Create a migration plan for production web servers
4. Optimize the plan for performance
```

Claude can then:
- Query your inventory automatically
- Analyze VM configurations
- Suggest optimal migration strategies
- Generate migration plans with best practices

#### Migration Troubleshooting

Use Claude for intelligent troubleshooting:

```
Claude, my migration plan "prod-migration" is failing. Can you:

1. Check the plan status
2. Look at recent events
3. Analyze any error messages
4. Suggest solutions
```

### Cursor IDE Development Integration

#### Code-Aware Migration Scripts

Cursor can help write migration automation scripts:

```python
# Ask Cursor: "Write a script to automate batch migration creation"

import subprocess
import json
from typing import List, Dict

def create_batch_migration(
    provider: str,
    vm_filter: str,
    target_namespace: str,
    batch_size: int = 5
) -> List[str]:
    """
    Create multiple migration plans in batches
    Generated with Cursor + kubectl-mtv MCP integration
    """
    
    # Get VMs matching filter
    result = subprocess.run([
        "kubectl", "mtv", "get", "inventory", "vms", "--provider", provider,
        "--query", vm_filter,
        "-o", "json"
    ], capture_output=True, text=True)
    
    vms = json.loads(result.stdout)
    
    # Create batches
    batches = [vms[i:i+batch_size] for i in range(0, len(vms), batch_size)]
    plan_names = []
    
    for i, batch in enumerate(batches):
        plan_name = f"batch-migration-{i+1}"
        vm_names = [vm['name'] for vm in batch]
        
        # Create migration plan
        subprocess.run([
            "kubectl", "mtv", "create", "plan", "--name", plan_name,
            "--source", provider,
            "--target-namespace", target_namespace,
            "--vms", ",".join(vm_names)
        ])
        
        plan_names.append(plan_name)
    
    return plan_names
```

### Advanced AI Workflows

#### Migration Optimization Assistant

```
Claude, analyze my migration environment and suggest optimizations:

1. Review all my providers and their configurations
2. Check current migration plans and their efficiency
3. Analyze resource utilization during migrations
4. Suggest improvements for:
   - Convertor pod placement
   - Network and storage mappings
   - Migration timing and scheduling
   - VDDK configuration
```

#### Compliance and Audit Assistant

```
Claude, help me ensure migration compliance:

1. Review all provider credentials and security settings
2. Check RBAC permissions for migration operations
3. Verify TLS configuration and certificate validation
4. Generate a compliance report for our SOC2 audit
5. Suggest security improvements
```

## Security and Authentication

### Authentication Flow

The MCP server uses kubectl's existing authentication:

```bash
# Authentication is inherited from kubectl context
kubectl config current-context
kubectl auth whoami

# MCP server uses the same authentication
kubectl mtv mcp-server
```

### Token-Based Authentication (SSE Mode)

```bash
# Generate service account token
kubectl create serviceaccount mcp-user -n migrations
kubectl create token mcp-user -n migrations --duration=24h

# Use token with HTTP client
curl -N -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:8080/sse

# Use token with specific API server (for remote cluster access)
curl -N \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Kubernetes-Server: https://api.example.com:6443" \
  http://127.0.0.1:8080/sse
```

#### Header Authentication Details

The SSE mode supports two HTTP headers for Kubernetes authentication:

**Authorization Header**
```bash
# Extract token from service account
TOKEN=$(kubectl create token mcp-user -n migrations)

# Use with curl
curl -N -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8080/sse

# Use with service account token from pod
curl -N \
  -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
  http://127.0.0.1:8080/sse
```

**X-Kubernetes-Server Header**
```bash
# Connect to a specific Kubernetes API server
curl -N \
  -H "X-Kubernetes-Server: https://kubernetes.default.svc" \
  http://127.0.0.1:8080/sse

# Combine with authentication for remote cluster access
curl -N \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Kubernetes-Server: https://remote-cluster.example.com:6443" \
  http://127.0.0.1:8080/sse
```

**Fallback Behavior**

When headers are not provided:
- If `Authorization` header is missing: Uses credentials from the server's kubeconfig
- If `X-Kubernetes-Server` header is missing: Uses the API server from the server's kubeconfig
- If both headers are missing: Falls back entirely to the default kubeconfig

### Secure Production Setup

```yaml
# Service account for MCP server
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mcp-server
  namespace: migration-tools
---
# RBAC permissions
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: mcp-server-role
rules:
- apiGroups: ["forklift.konveyor.io"]
  resources: ["*"]
  verbs: ["get", "list", "create", "update", "patch", "delete", "watch"]
- apiGroups: [""]
  resources: ["secrets", "configmaps"]
  verbs: ["get", "list", "create", "update", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: mcp-server-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: mcp-server-role
subjects:
- kind: ServiceAccount
  name: mcp-server
  namespace: migration-tools
```

## Troubleshooting MCP Integration

### Common Issues and Solutions

#### MCP Server Not Starting

```bash
# Check kubectl connectivity
kubectl cluster-info
kubectl auth whoami

# Verify kubectl-mtv installation
kubectl mtv version

# Check MCP server logs
kubectl mtv mcp-server -v=2
```

#### AI Assistant Cannot Connect

```bash
# Verify server is running
ps aux | grep "kubectl.*mtv.*mcp-server"

# Test stdio communication
echo '{"method": "ping"}' | kubectl mtv mcp-server

# Test SSE endpoint
curl -v http://127.0.0.1:8080/sse
```

#### Authentication Issues

```bash
# Check current kubectl context
kubectl config current-context

# Verify permissions
kubectl auth can-i get providers.forklift.konveyor.io
kubectl auth can-i create plans.forklift.konveyor.io

# Test with explicit kubeconfig
KUBECONFIG=/path/to/config kubectl mtv mcp-server
```

#### Performance Issues

```bash
# Monitor resource usage
top -p $(pgrep -f "kubectl.*mtv.*mcp-server")

# Increase verbosity for debugging
kubectl mtv mcp-server -v=3

# Check for network issues (SSE mode)
netstat -tlnp | grep 8080
```

## Best Practices for MCP Integration

### Security Best Practices

1. **Use Least Privilege**: Configure RBAC with minimal required permissions
2. **Enable TLS**: Use certificate-based encryption for SSE mode
3. **Monitor Access**: Log and audit MCP server usage
4. **Rotate Tokens**: Regularly rotate service account tokens
5. **Network Security**: Restrict MCP server network access

### Performance Optimization

1. **Use Stdio Mode**: For direct AI integrations, stdio mode is more efficient
2. **Limit Concurrency**: Avoid simultaneous large operations
3. **Cache Results**: Cache inventory queries in AI applications
4. **Monitor Resources**: Watch CPU and memory usage during operations
5. **Optimize Queries**: Use specific filters in inventory queries

### Operational Excellence

1. **Document Integration**: Maintain clear AI integration documentation
2. **Version Control**: Track MCP configuration changes
3. **Test Regularly**: Validate AI assistant functionality
4. **Monitor Health**: Implement MCP server health checks
5. **Backup Configuration**: Maintain backup of AI integration configs

## Next Steps

After mastering MCP integration:

1. **Complete Migration Toolkit**: Explore complementary tools in [Chapter 23: Integration with KubeVirt Tools](/kubectl-mtv/23-integration-with-kubevirt-tools)

---

*Previous: [Chapter 21: Best Practices and Security](/kubectl-mtv/21-best-practices-and-security)*  
*Next: [Chapter 23: Integration with KubeVirt Tools](/kubectl-mtv/23-integration-with-kubevirt-tools)*
