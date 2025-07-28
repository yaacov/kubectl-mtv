# Creating Migration Hosts

This guide explains how to create migration hosts using the `kubectl-mtv create host` command. Migration hosts represent ESXi servers that can be used as migration sources or intermediaries in vSphere environments.

## Overview

Migration hosts are Kubernetes resources that represent individual ESXi servers in your vSphere infrastructure. They enable:
- **Direct ESXi Access**: Connect directly to ESXi hosts for migration operations
- **Performance Optimization**: Reduce vCenter load by distributing connections
- **Fine-grained Control**: Select specific hosts for migration workloads
- **Authentication Management**: Flexible credential management per host

**Note**: Host creation is only supported for vSphere providers. Other provider types (oVirt, OpenStack, OVA) do not support host resources.

## Basic Syntax

```bash
kubectl-mtv create host <host-id-1> [host-id-2] [...] \
  --provider <provider-name> \
  (--ip-address <ip> | --network-adapter <adapter-name>) \
  [authentication-options] \
  [tls-options]
```

**Required**: Either `--ip-address` or `--network-adapter` must be specified (but not both).

## Authentication Options

### Option 1: Use Provider Secret (ESXi Endpoints)

For vSphere providers configured with ESXi endpoints, hosts can automatically reuse the provider's authentication credentials:

```bash
# Using direct IP address
kubectl-mtv create host esxi-host-01 esxi-host-02 \
  --provider vsphere-esxi-provider \
  --ip-address 192.168.1.100

# Using network adapter lookup
kubectl-mtv create host esxi-host-01 esxi-host-02 \
  --provider vsphere-esxi-provider \
  --network-adapter "Management Network"
```

This is the most efficient option when your provider is configured with `sdkEndpoint: esxi`.

### Option 2: Use Existing Secret

Reference an existing Kubernetes secret containing host credentials:

```bash
# Using direct IP address
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-provider \
  --existing-secret esxi-credentials \
  --ip-address 192.168.1.100

# Using network adapter lookup
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-provider \
  --existing-secret esxi-credentials \
  --network-adapter "Management Network"
```

The secret should contain:
- `user`: ESXi username
- `password`: ESXi password
- `insecureSkipVerify`: "true" for insecure connections (optional)
- `cacert`: CA certificate for secure connections (optional) - can be certificate content or use `@filename` to read from file

### Option 3: Create New Credentials

Provide new credentials that will be stored in a generated secret:

```bash
# Using direct IP address
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-provider \
  --username root \
  --password secret123 \
  --ip-address 192.168.1.100

# Using network adapter lookup
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-provider \
  --username root \
  --password secret123 \
  --network-adapter "Management Network"
```

With TLS options:
```bash
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-provider \
  --username root \
  --password secret123 \
  --ip-address 192.168.1.100 \
  --host-insecure-skip-tls \
  --cacert @/path/to/ca.cert
```

## IP Address Resolution (Required)

**One of these options must be specified for each host:**

### Option 1: Direct IP Address

Specify the IP address directly:

```bash
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-provider \
  --ip-address 192.168.1.100 \
  --username root \
  --password secret123
```

### Option 2: Network Adapter Lookup

Let the tool resolve the IP from a named network adapter:

```bash
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-provider \
  --network-adapter "Management Network" \
  --username root \
  --password secret123
```

This queries the provider's inventory to find the specified network adapter and extract its IP address.

**Note**: You cannot specify both `--ip-address` and `--network-adapter` in the same command.

## Complete Examples

### Single Host with Direct IP

```bash
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-datacenter \
  --ip-address 192.168.1.100 \
  --username root \
  --password supersecret \
  --host-insecure-skip-tls
```

### Multiple Hosts with Network Adapter Resolution

```bash
kubectl-mtv create host esxi-host-01 esxi-host-02 esxi-host-03 \
  --provider vsphere-datacenter \
  --network-adapter "Management Network" \
  --username root \
  --password supersecret
```

### Host with Existing Secret

```bash
# First create the secret
kubectl create secret generic esxi-creds \
  --from-literal=user=root \
  --from-literal=password=supersecret \
  --from-literal=insecureSkipVerify=true

# Then create the host
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-datacenter \
  --ip-address 192.168.1.100 \
  --existing-secret esxi-creds
```

### Host with CA Certificate

```bash
# Using CA certificate from file
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-datacenter \
  --ip-address 192.168.1.100 \
  --username root \
  --password supersecret \
  --cacert @/path/to/esxi-ca.cert

# Or provide certificate content directly
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-datacenter \
  --ip-address 192.168.1.100 \
  --username root \
  --password supersecret \
  --cacert "-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJANcT3I7d4I6BA...
-----END CERTIFICATE-----"
```

### ESXi Provider with Automatic Credentials

```bash
# Create ESXi provider first
kubectl-mtv create provider vsphere-esxi \
  --type vsphere \
  --url https://esxi-host.example.com \
  --sdk-endpoint esxi \
  --username root \
  --password secret

# Create hosts using provider credentials
kubectl-mtv create host esxi-host-01 esxi-host-02 \
  --provider vsphere-esxi
```

## Advanced Usage

### Multiple Hosts with Different IPs

When creating multiple hosts, all hosts in a single command use the same IP resolution method:

```bash
# All hosts use network adapter resolution
kubectl-mtv create host esxi-host-01 esxi-host-02 esxi-host-03 \
  --provider vsphere-datacenter \
  --network-adapter "Management Network" \
  --username root --password secret

# For different IPs per host, create them separately:
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-datacenter \
  --ip-address 192.168.1.100 \
  --username root --password secret

kubectl-mtv create host esxi-host-02 \
  --provider vsphere-datacenter \
  --ip-address 192.168.1.101 \
  --username root --password secret
```

### Namespace-Specific Host Creation

```bash
kubectl-mtv create host esxi-host-01 \
  --namespace migration-project \
  --provider vsphere-datacenter \
  --ip-address 192.168.1.100 \
  --username root --password secret
```

### Using Environment Variables

```bash
# Set inventory URL if not auto-discovered
export MTV_INVENTORY_URL="https://inventory.example.com"

kubectl-mtv create host esxi-host-01 \
  --provider vsphere-datacenter \
  --network-adapter "Management Network" \
  --username root --password secret
```

## Host Resource Management

### List Hosts

```bash
# List all hosts
kubectl get hosts

# List hosts with more details
kubectl-mtv get host

# List hosts in specific namespace
kubectl-mtv get host -n migration-project
```

### View Host Details

```bash
# Get host details
kubectl get host esxi-host-01-a1b2 -o yaml

# Describe host status
kubectl describe host esxi-host-01-a1b2
```

### Delete Hosts

```bash
# Delete host using native kubectl (this will also clean up associated secrets if no other hosts use them)
kubectl delete host esxi-host-01-a1b2

# Delete multiple hosts using native kubectl
kubectl delete host esxi-host-01-a1b2 esxi-host-02-c3d4

# Delete host using MTV command
kubectl mtv delete host esxi-host-01-a1b2

# Delete multiple hosts using MTV command  
kubectl mtv delete host esxi-host-01-a1b2 esxi-host-02-c3d4
```

## Host Naming and Ownership

### Resource Naming

Hosts are created with names in the format: `<host-id>-<hash>`
- `host-id`: The ESXi host identifier from provider inventory
- `hash`: A 4-character hash to prevent collisions

Example: `esxi-host-01-a1b2c3d4`

### Ownership Relationships

1. **Provider Ownership**: Hosts are owned by their provider for lifecycle management
2. **Secret Ownership**: When creating new secrets, hosts become owners for garbage collection
3. **Shared Secrets**: Multiple hosts can share the same secret, each becoming an owner

## Best Practices

### 1. Use Descriptive Host IDs

Choose host IDs that match your ESXi hostnames for clarity:
```bash
# Good: matches ESXi hostname
kubectl-mtv create host esxi-prod-01.example.com \
  --provider vsphere-prod

# Avoid: generic names
kubectl-mtv create host host1 --provider vsphere-prod
```

### 2. Leverage Provider Secrets for ESXi Endpoints

For ESXi providers, reuse provider credentials when possible:
```bash
# Efficient: reuses provider secret
kubectl-mtv create host esxi-host-01 esxi-host-02 \
  --provider vsphere-esxi-provider
```

### 3. Group Hosts by Environment

Create hosts in environment-specific namespaces:
```bash
# Production hosts
kubectl-mtv create host prod-esxi-01 prod-esxi-02 \
  --namespace production \
  --provider vsphere-prod

# Development hosts  
kubectl-mtv create host dev-esxi-01 \
  --namespace development \
  --provider vsphere-dev
```

### 4. Use Network Adapter Resolution When Possible

This automatically handles IP changes and is more maintainable:
```bash
# Preferred: automatic IP resolution
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-datacenter \
  --network-adapter "Management Network"

# Manual: requires updates if IP changes
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-datacenter \
  --ip-address 192.168.1.100
```

### 5. Secure Credential Management

- Use existing secrets for shared credentials
- Enable TLS verification when possible
- Rotate credentials regularly

```bash
# Secure approach with CA certificate from file
kubectl-mtv create host esxi-host-01 \
  --provider vsphere-datacenter \
  --existing-secret secure-esxi-creds \
  --cacert @ca.cert
```

## Troubleshooting

### Common Issues

#### 1. Host ID Not Found
```bash
Error: the following host IDs were not found in provider inventory: invalid-host-01
```
**Solution**: Verify the host exists in provider inventory:
```bash
kubectl-mtv get inventory hosts vsphere-provider
```

#### 2. Provider Not vSphere
```bash
Error: only vSphere providers support host creation, got provider type: ovirt
```
**Solution**: Host creation only works with vSphere providers. Use provider inventory directly for other types.

#### 3. Missing IP Resolution
```bash
Error: either --ip-address OR --network-adapter must be provided
```
**Solution**: Specify either IP address or network adapter:
```bash
# Using direct IP
kubectl-mtv create host esxi-host-01 --provider vsphere-provider --ip-address 192.168.1.100 --username root --password secret

# Using network adapter
kubectl-mtv create host esxi-host-01 --provider vsphere-provider --network-adapter "Management Network" --username root --password secret
```

#### 4. Network Adapter Not Found
```bash
Error: network adapter 'Management' not found for host 'esxi-host-01' or no IP address available
```
**Solution**: Check available network adapters:
```bash
kubectl-mtv get inventory hosts vsphere-datacenter -o json | jq '.[] | select(.id=="esxi-host-01") | .networkAdapters'
```