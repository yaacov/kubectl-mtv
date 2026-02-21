---
layout: default
title: "Chapter 7: Migration Host Management (vSphere Specific)"
parent: "II. Provider, Host, and VDDK Management"
nav_order: 2
---

Migration hosts enable direct data transfer from ESXi hosts, bypassing vCenter for improved performance and control. This chapter covers the complete lifecycle management of migration hosts, which are exclusive to vSphere environments.

## Overview and Purpose of Migration Hosts

### What are Migration Hosts?

Migration hosts are specialized resources that enable Forklift to establish direct connections to ESXi hosts, providing:

- **Direct ESXi Access**: Bypass vCenter for data transfer operations
- **Performance Optimization**: Reduce network hops and potential bottlenecks
- **Network Control**: Specify which ESXi network interface to use for migration
- **Enhanced Throughput**: Direct host-to-host data transfer capabilities

### How Migration Hosts Work

By creating host resources, Forklift can utilize ESXi host interfaces directly for network transfer to OpenShift, provided the OpenShift worker nodes and ESXi host interfaces have network connectivity. This is particularly beneficial when users want to control which specific ESXi interface is used for migration, even without direct access to ESXi host credentials.

### Requirements and Limitations

- **vSphere Only**: Migration hosts are exclusively supported for vSphere providers
- **Network Connectivity**: OpenShift worker nodes must have network connectivity to ESXi host interfaces
- **Host ID Matching**: The `--host-id` flag requires **inventory host IDs** (e.g. `host-8`), NOT display names or IP addresses. Use `kubectl-mtv get inventory host --provider <name>` to list available IDs
- **Provider Dependency**: Hosts must be associated with an existing vSphere provider

## How-To: Creating Hosts

### Basic Syntax

```bash
kubectl mtv create host --host-id HOST_ID[,HOST_ID,...] --provider PROVIDER_NAME [options]
```

> **Important:** The `--host-id` flag expects **inventory host IDs** (e.g. `host-8`, `host-12`),
> which are vSphere managed object references. These are NOT the same as display names or IP
> addresses. Run `kubectl-mtv get inventory host --provider <name>` to discover available IDs.
>
> Multiple host IDs are passed as a comma-separated list (e.g., `--host-id host-8,host-12,host-15`).
> You can also use `--host-ids` as an alias.

### IP Address Resolution Methods

Migration hosts support two methods for IP address resolution:

#### Method 1: Direct IP Address (--ip-address)

Specify the exact IP address to use for the migration host:

```bash
# Single host with direct IP
kubectl mtv create host --host-id host-8 \
  --provider vsphere-provider \
  --ip-address 192.168.1.10

# Multiple hosts with same IP (load balancer scenario)
kubectl mtv create host --host-id host-8,host-12,host-15 \
  --provider vsphere-provider \
  --ip-address 192.168.1.100
```

#### Method 2: Network Adapter Lookup (--network-adapter)

Automatically resolve IP address from a network adapter name in the inventory:

```bash
# Single host using network adapter lookup
kubectl mtv create host --host-id host-8 \
  --provider vsphere-provider \
  --network-adapter "Management Network"

# Multiple hosts using network adapter lookup
kubectl mtv create host --host-id host-8,host-12,host-15 \
  --provider vsphere-provider \
  --network-adapter "vMotion Network"
```

**Note**: The `--ip-address` and `--network-adapter` flags are mutually exclusive. You must specify exactly one of them.

### Authentication Options

Migration hosts support three authentication methods:

#### Option 1: Provider Secret (Default - Recommended)

Use the provider's existing credentials automatically:

```bash
# ESXi endpoint provider with direct IP (no additional credentials needed)
kubectl mtv create host --host-id host-8 \
  --provider esxi-provider \
  --ip-address 192.168.1.10

# ESXi endpoint provider with network adapter lookup
kubectl mtv create host --host-id host-8 \
  --provider esxi-provider \
  --network-adapter "Management Network"
```

This method works best with ESXi endpoint providers where the provider credentials can be used directly for host access.

#### Option 2: Existing Secret

Use an existing Kubernetes secret for host authentication:

```bash
# Create secret first
kubectl create secret generic esxi-host-secret \
  --from-literal=user=root \
  --from-literal=password=HostSecurePassword

# Create host using existing secret
kubectl mtv create host --host-id host-8 \
  --provider vsphere-provider \
  --existing-secret esxi-host-secret \
  --ip-address 192.168.1.10

# Multiple hosts using same secret
kubectl mtv create host --host-id host-8,host-12,host-15 \
  --provider vsphere-provider \
  --existing-secret esxi-hosts-shared-secret \
  --network-adapter "Management Network"
```

#### Option 3: New Credentials

Provide username and password directly (creates a new secret):

```bash
# Create host with new credentials and direct IP
kubectl mtv create host --host-id host-8 \
  --provider vsphere-provider \
  --username root \
  --password HostSecurePassword \
  --ip-address 192.168.1.10

# Create host with credentials using network adapter lookup
kubectl mtv create host --host-id host-8 \
  --provider vsphere-provider \
  --username administrator \
  --password HostSecurePassword \
  --network-adapter "Management Network"
```

### Advanced Host Creation Options

#### Host Creation with TLS Settings

```bash
# Skip TLS verification for host connections (testing only)
kubectl mtv create host --host-id host-8 \
  --provider vsphere-provider \
  --username root \
  --password TestPassword \
  --ip-address 192.168.100.10 \
  --host-insecure-skip-tls

# Provide CA certificate for host authentication
kubectl mtv create host --host-id host-12 \
  --provider vsphere-provider \
  --username root \
  --password ProdPassword \
  --ip-address 192.168.1.10 \
  --cacert @/path/to/esxi-ca-certificate.pem

# Inline CA certificate
kubectl mtv create host --host-id host-12 \
  --provider vsphere-provider \
  --username root \
  --password ProdPassword \
  --ip-address 192.168.1.10 \
  --cacert "-----BEGIN CERTIFICATE-----
MIIGBzCCA++gAwIBAgIJAKt..."
```

#### Bulk Host Creation

Create multiple hosts efficiently:

```bash
# Multiple hosts with same configuration
kubectl mtv create host \
  --host-id host-8,host-12,host-15,host-20 \
  --provider vsphere-cluster-provider \
  --existing-secret esxi-cluster-secret \
  --network-adapter "Management Network"

# Hosts in different clusters but same authentication
kubectl mtv create host \
  --host-id host-8,host-12,host-30,host-35 \
  --provider vsphere-multi-cluster \
  --username cluster-admin \
  --password ClusterPassword \
  --network-adapter "vMotion"
```

#### Host Creation with Custom Inventory URL

```bash
# Specify custom inventory service URL
kubectl mtv create host --host-id host-8 \
  --provider vsphere-provider \
  --ip-address 192.168.1.10 \
  --inventory-url http://custom-inventory.internal:8080

# Use environment variable for inventory URL
export MTV_INVENTORY_URL=http://inventory-service.forklift:8080
kubectl mtv create host --host-id host-8 \
  --provider vsphere-provider \
  --ip-address 192.168.1.10
```

## Listing, Describing, and Deleting Hosts

### List Migration Hosts

```bash
# List all hosts in current namespace
kubectl mtv get hosts

# List hosts in specific namespace
kubectl mtv get hosts --namespace migration-namespace

# List hosts across all namespaces
kubectl mtv get hosts --all-namespaces

# List with detailed output
kubectl mtv get hosts --output yaml
kubectl mtv get hosts --output json

# List specific host
kubectl mtv get host --name esxi-host-01
```

### Describe Migration Hosts

Get detailed information about migration hosts:

```bash
# Describe a specific host
kubectl mtv describe host --name esxi-host-01

# View host configuration and status
kubectl get host esxi-host-01 -o yaml
```

### Delete Migration Hosts

```bash
# Delete a specific host
kubectl mtv delete host --name esxi-host-01

# Delete multiple hosts
kubectl mtv delete hosts --name esxi-host-01,esxi-host-02,esxi-host-03

# Delete all hosts in namespace (use with caution)
kubectl mtv delete host --all

# Plural form for single or multiple
kubectl mtv delete hosts --name esxi-host-01
```

## Best Practices for Host Creation

### Network Planning and Design

1. **Dedicated Migration Networks**:
   ```bash
   # Use dedicated network adapters for migration traffic
   kubectl mtv create host --host-id host-8 \
     --provider vsphere-provider \
     --network-adapter "Migration Network" \
     --existing-secret migration-credentials
   ```

2. **High-Bandwidth Networks**:
   ```bash
   # Prioritize 10Gb+ network interfaces
   kubectl mtv create host --host-id host-12 \
     --provider vsphere-provider \
     --network-adapter "10Gb vMotion" \
     --username admin \
     --password SecurePassword
   ```

3. **Network Segmentation**:
   ```bash
   # Separate production and development hosts
   kubectl mtv create host --host-id host-8 \
     --provider vsphere-prod \
     --network-adapter "Prod Migration Net" \
     --namespace production
   
   kubectl mtv create host --host-id host-20 \
     --provider vsphere-dev \
     --network-adapter "Dev Migration Net" \
     --namespace development
   ```

### Security Best Practices

1. **Credential Management**:
   ```bash
   # Use dedicated service accounts for migration
   kubectl create secret generic esxi-migration-creds \
     --from-literal=user=svc-migration \
     --from-literal=password=$(openssl rand -base64 32)
   
   kubectl mtv create host --host-id host-8 \
     --provider vsphere-provider \
     --existing-secret esxi-migration-creds \
     --ip-address 192.168.1.10
   ```

2. **Certificate Validation**:
   ```bash
   # Always use CA certificates in production
   kubectl mtv create host --host-id host-12 \
     --provider vsphere-provider \
     --username svc-migration \
     --password SecurePassword \
     --ip-address 192.168.1.10 \
     --cacert @/secure/esxi-ca.pem
   ```

3. **Minimal Privileges**:
   ```bash
   # Create ESXi users with minimal required permissions
   # Required privileges: Host.Config.Connection, Host.Config.NetService
   kubectl mtv create host --host-id host-8 \
     --provider vsphere-provider \
     --username migration-user \
     --password LimitedPrivPassword \
     --ip-address 192.168.1.10
   ```

### Performance Optimization

1. **Network Interface Selection**:
   ```bash
   # Choose high-performance network adapters
   kubectl mtv create host --host-id host-8 \
     --provider vsphere-provider \
     --network-adapter "25Gb Ethernet" \
     --existing-secret high-perf-creds
   ```

2. **Load Distribution**:
   ```bash
   # Distribute hosts across different network segments
   kubectl mtv create host --host-id host-8 \
     --provider vsphere-provider \
     --network-adapter "Rack1 Migration" \
     --existing-secret rack1-creds
   
   kubectl mtv create host --host-id host-20 \
     --provider vsphere-provider \
     --network-adapter "Rack2 Migration" \
     --existing-secret rack2-creds
   ```

3. **Provider Endpoint Optimization**:
   ```bash
   # Use ESXi endpoint providers for direct host access
   kubectl mtv create provider --name esxi-direct --type vsphere \
     --url https://esxi-host.example.com/sdk \
     --sdk-endpoint esxi \
     --username root \
     --password DirectPassword
   
   # Create host using ESXi endpoint provider
   kubectl mtv create host --host-id host-8 \
     --provider esxi-direct \
     --ip-address 192.168.1.10
   ```

### Monitoring and Validation

1. **Host Connectivity Testing**:
   ```bash
   # Verify host creation and status
   kubectl mtv describe host --name esxi-host-01
   
   # Check underlying Kubernetes resource
   kubectl get host esxi-host-01 -o yaml
   
   # Monitor host events
   kubectl get events --field-selector involvedObject.name=esxi-host-01
   ```

2. **Network Connectivity Validation**:
   ```bash
   # Test connectivity from OpenShift worker nodes
   # This should be done from the cluster nodes
   for node in $(kubectl get nodes -o name); do
     kubectl debug $node -it --image=nicolaka/netshoot -- \
       ping -c 3 192.168.1.10
   done
   ```

## Complete Migration Host Workflow Examples

### Example 1: Production ESXi Cluster Setup

```bash
# Step 1: Create vSphere provider for cluster
kubectl mtv create provider --name vsphere-cluster --type vsphere \
  --url https://vcenter.prod.company.com/sdk \
  --username svc-migration@vsphere.local \
  --password $(cat /secure/vcenter-password) \
  --vddk-init-image quay.io/company/vddk:8.0.1

# Step 2: Create shared credentials for ESXi hosts
kubectl create secret generic esxi-cluster-creds \
  --from-literal=user=migration-svc \
  --from-literal=password=$(cat /secure/esxi-password)

# Step 3: Create migration hosts for each ESXi server (use IDs from 'get inventory host')
kubectl mtv create host \
  --host-id host-8,host-12,host-15,host-20 \
  --provider vsphere-cluster \
  --existing-secret esxi-cluster-creds \
  --network-adapter "Migration Network" \
  --cacert @/secure/esxi-ca.pem

# Step 4: Verify host creation
kubectl mtv get hosts
kubectl mtv describe host --name esxi-prod-01
```

### Example 2: Development Environment Setup

```bash
# Step 1: Create development vSphere provider
kubectl mtv create provider --name vsphere-dev --type vsphere \
  --url https://vcenter-dev.internal/sdk \
  --username administrator@vsphere.local \
  --password DevPassword123 \
  --provider-insecure-skip-tls

# Step 2: Create development hosts with relaxed security (use IDs from 'get inventory host')
kubectl mtv create host --host-id host-8,host-12 \
  --provider vsphere-dev \
  --username root \
  --password DevHostPassword \
  --network-adapter "VM Network" \
  --host-insecure-skip-tls \
  --namespace development

# Step 3: Verify setup
kubectl mtv get hosts --namespace development
```

### Example 3: Multi-Site Migration Setup

```bash
# Site A hosts (use IDs from 'get inventory host --provider vsphere-site-a')
kubectl mtv create host \
  --host-id host-8,host-12 \
  --provider vsphere-site-a \
  --ip-address 10.1.1.10 \
  --existing-secret site-a-creds

# Site B hosts (use IDs from 'get inventory host --provider vsphere-site-b')
kubectl mtv create host \
  --host-id host-30,host-35 \
  --provider vsphere-site-b \
  --ip-address 10.2.1.10 \
  --existing-secret site-b-creds

# Verify multi-site setup
kubectl mtv get hosts --all-namespaces
```

## Troubleshooting Migration Hosts

### Common Host Issues

#### Host Creation Failures

```bash
# Check provider status
kubectl mtv describe provider --name vsphere-provider

# List available host IDs from inventory
kubectl mtv get inventory host --provider vsphere-provider

# Check network connectivity
kubectl debug node-name -it --image=nicolaka/netshoot -- \
  ping -c 3 192.168.1.10
```

#### Authentication Problems

```bash
# Verify secret contents
kubectl get secret esxi-host-secret -o yaml

# Test ESXi connectivity manually
curl -k -u root:password https://192.168.1.10/sdk

# Check host events for authentication errors
kubectl get events --field-selector involvedObject.name=esxi-host-01
```

#### Network Adapter Issues

```bash
# List available network adapters
kubectl mtv get inventory networks --provider vsphere-provider

# Verify network adapter name matches inventory
kubectl mtv get inventory hosts --provider vsphere-provider --output yaml | grep -A5 networks
```

### Host Status Monitoring

```bash
# Monitor host status
kubectl mtv get hosts --watch

# Check detailed host information
kubectl get host esxi-host-01 -o yaml

# View host conditions and events
kubectl describe host esxi-host-01
```

### Performance Troubleshooting

```bash
# Check network latency from cluster nodes
kubectl run network-test --image=nicolaka/netshoot -it --rm -- \
  ping -c 10 192.168.1.10

# Monitor bandwidth during migration
kubectl run bandwidth-test --image=nicolaka/netshoot -it --rm -- \
  iperf3 -c 192.168.1.10

# Verify ESXi host performance
kubectl debug node-name -it --image=nicolaka/netshoot -- \
  curl -k https://192.168.1.10/ui/
```

## Integration with Migration Plans

Migration hosts automatically integrate with migration plans when available:

```bash
# Create migration plan - will automatically use available hosts
kubectl mtv create plan --name migration-with-hosts \
  --source vsphere-provider \
  --vms "vm1,vm2,vm3"

# The plan will automatically leverage created migration hosts
# for improved performance during disk transfer operations
```

## Next Steps

After setting up migration hosts:

1. **Optimize VDDK**: Configure VDDK images in [Chapter 8: VDDK Image Creation and Configuration](../08-vddk-image-creation-and-configuration)
2. **Explore Inventory**: Discover available resources in [Chapter 9: Inventory Management](../09-inventory-management)
3. **Create Mappings**: Define resource mappings in [Chapter 11: Mapping Management](../11-mapping-management)
4. **Plan Migrations**: Create optimized migration plans in [Chapter 13: Migration Plan Creation](../13-migration-plan-creation)

---

*Previous: [Chapter 6: Provider Management](../06-provider-management)*  
*Next: [Chapter 8: VDDK Image Creation and Configuration](../08-vddk-image-creation-and-configuration)*
