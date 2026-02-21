---
layout: default
title: "Chapter 9: Inventory Management"
parent: "III. Inventory and Advanced Query Language"
nav_order: 1
---

Inventory management is a core feature of kubectl-mtv that allows you to discover, explore, and query resources from various virtualization providers. This chapter covers comprehensive inventory operations for migration planning and resource discovery.

## Overview of Resources Available for Querying

kubectl-mtv provides access to a comprehensive inventory of resources across different provider types:

### General Resources (All Providers)

| Resource | Aliases | Description |
|----------|---------|-------------|
| `vm` | `vms` | Virtual machines and instances |
| `host` | `hosts` | Physical hosts and hypervisors |
| `network` | `networks` | Network configurations and segments |
| `storage` | `storages` | Storage systems and configurations |
| `datacenter` | `datacenters` | Data center structures |
| `cluster` | `clusters` | Compute clusters |
| `disk` | `disks` | Individual disk resources |
| `namespace` | `namespaces` | Logical resource groupings |

### Profile Resources (oVirt/RHV)

| Resource | Aliases | Description |
|----------|---------|-------------|
| `diskprofile` | `diskprofiles`, `disk-profiles` | Disk performance profiles |
| `nicprofile` | `nicprofiles`, `nic-profiles` | Network interface profiles |

### OpenStack-Specific Resources

| Resource | Aliases | Description |
|----------|---------|-------------|
| `instance` | `instances` | OpenStack compute instances |
| `image` | `images` | VM images and templates |
| `flavor` | `flavors` | Instance sizing configurations |
| `project` | `projects` | OpenStack projects/tenants |
| `volume` | `volumes` | Block storage volumes |
| `volumetype` | `volumetypes`, `volume-types` | Volume type definitions |
| `snapshot` | `snapshots` | Volume and instance snapshots |
| `subnet` | `subnets` | Network subnets |

### vSphere-Specific Resources

| Resource | Aliases | Description |
|----------|---------|-------------|
| `datastore` | `datastores` | vSphere storage datastores |
| `resourcepool` | `resourcepools`, `resource-pools` | vSphere resource pools |
| `folder` | `folders` | vSphere organizational folders |

### Kubernetes-Specific Resources

| Resource | Aliases | Description |
|----------|---------|-------------|
| `pvc` | `pvcs`, `persistentvolumeclaims` | Persistent Volume Claims |
| `datavolume` | `datavolumes`, `data-volumes` | KubeVirt DataVolumes |

### EC2-Specific Resources

| Resource | Aliases | Description |
|----------|---------|-------------|
| `ec2-instance` | `ec2-instances` | EC2 compute instances |
| `ec2-volume` | `ec2-volumes` | EBS volumes |
| `ec2-network` | `ec2-networks` | VPCs and subnets |
| `ec2-volume-type` | `ec2-volume-types` | EBS volume types |

**Note**: Generic resources (`vms`, `networks`, `storage`) also work with EC2 providers and display EC2-specific information.

### Provider Resources

| Resource | Aliases | Description |
|----------|---------|-------------|
| `provider` | `providers` | Provider inventory and status |

## General Syntax

The general syntax for inventory commands follows this pattern:

```bash
kubectl mtv get inventory <resource> --provider <provider> [flags]
```

### Basic Command Structure

- **Resource**: The type of resource to query (vm, network, storage, etc.)
- **Provider**: The name of the provider to query
- **Flags**: Optional parameters for filtering, output format, etc.

### Common Flags

- `-o, --output`: Output format (table, json, yaml, planvms for VMs)
- `-q, --query`: Query filter using [Tree Search Language (TSL)](../27-tsl-tree-search-language-reference)
- `-w, --watch`: Watch for real-time changes
- `--extended`: Show extended information (where supported)
- `--inventory-url`: Custom inventory service URL

## Common Inventory Examples

### Virtual Machines

VMs are the most commonly queried inventory resource for migration planning:

```bash
# List all VMs from a vSphere provider
kubectl mtv get inventory vms --provider vsphere-prod

# List VMs from different provider types
kubectl mtv get inventory vms --provider ovirt-prod
kubectl mtv get inventory instances --provider openstack-prod
kubectl mtv get inventory vms --provider openshift-source

# List VMs with extended information
kubectl mtv get inventory vms --provider vsphere-prod --extended

# Watch for VM changes in real-time
kubectl mtv get inventory vms --provider vsphere-prod --watch
```

### Networks

Network inventory helps plan network mappings:

```bash
# List all networks from vSphere
kubectl mtv get inventory networks --provider vsphere-prod

# List networks from oVirt
kubectl mtv get inventory networks --provider ovirt-prod

# List subnets from OpenStack
kubectl mtv get inventory subnets --provider openstack-prod

# View network details in YAML format
kubectl mtv get inventory networks --provider vsphere-prod --output yaml
```

### Storage

Storage inventory assists with storage mapping configuration:

```bash
# List storage from vSphere (datastores)
kubectl mtv get inventory datastores --provider vsphere-prod

# List storage from oVirt
kubectl mtv get inventory storages --provider ovirt-prod

# List volume types from OpenStack
kubectl mtv get inventory volumetypes --provider openstack-prod

# View storage details in JSON format
kubectl mtv get inventory datastores --provider vsphere-prod --output json
```

### Hosts and Infrastructure

Discover infrastructure layout for planning:

```bash
# List ESXi hosts in vSphere
kubectl mtv get inventory hosts --provider vsphere-prod

# List oVirt hosts
kubectl mtv get inventory hosts --provider ovirt-prod

# List datacenters
kubectl mtv get inventory datacenters --provider vsphere-prod

# List resource pools (vSphere)
kubectl mtv get inventory resource-pools --provider vsphere-prod

# List clusters
kubectl mtv get inventory clusters --provider vsphere-prod
```

### Provider Status

Check provider health and connectivity:

```bash
# List inventory from all providers
kubectl mtv get inventory providers

# Get detailed inventory from specific provider
kubectl mtv get inventory providers --name vsphere-prod

# Monitor provider status
kubectl mtv get inventory providers --name vsphere-prod --watch
```

## Output Formats

### Table Format (Default)

The default table format provides a concise overview:

```bash
# Default table output
kubectl mtv get inventory vms --provider vsphere-prod

# Example output:
# NAME          POWER STATE    MEMORY(MB)    CPU    DISKS
# web-server-01 poweredOn      4096          2      2
# db-server-01  poweredOn      8192          4      3
# test-vm-01    poweredOff     2048          1      1
```

### JSON Format

JSON output provides complete data for automation:

```bash
# JSON output for scripting
kubectl mtv get inventory vms --provider vsphere-prod --output json

# Extract specific fields using jq
kubectl mtv get inventory vms --provider vsphere-prod --output json | jq '.items[].name'

# Complex data extraction
kubectl mtv get inventory vms --provider vsphere-prod --output json | \
  jq '.items[] | select(.powerState == "poweredOn") | .name'
```

### YAML Format

YAML format for human-readable structured data:

```bash
# YAML output
kubectl mtv get inventory vms --provider vsphere-prod --output yaml

# Save to file for analysis
kubectl mtv get inventory vms --provider vsphere-prod --output yaml > vms-inventory.yaml
```

### Extended Output

Extended output provides additional details where supported:

```bash
# Extended information for VMs
kubectl mtv get inventory vms --provider vsphere-prod --extended

# Extended output shows additional fields like:
# - Guest OS information
# - IP addresses
# - Detailed disk information
# - Network adapter details
```

## How-To: Exporting VMs for Migration Planning

The special `planvms` output format is designed specifically for migration planning:

### PlanVMs Output Format

```bash
# Export VMs in planvms format
kubectl mtv get inventory vms --provider vsphere-prod --output planvms

# Save to file for migration plan creation
kubectl mtv get inventory vms --provider vsphere-prod --output planvms > migration-vms.yaml
```

### Using PlanVMs with Migration Plans

```bash
# Create migration plan using exported VMs
kubectl mtv create plan --name production-migration \
  --source vsphere-prod \
  --vms @migration-vms.yaml

# Or use the planvms format directly in plan creation
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where name ~= 'prod-.*' and powerState = 'poweredOn'" \
  --output planvms > prod-vms.yaml

kubectl mtv create plan --name prod-migration \
  --source vsphere-prod \
  --vms @prod-vms.yaml
```

### PlanVMs Format Structure

The planvms format provides a structured list suitable for plan creation:

```yaml
# Example planvms output
apiVersion: v1
kind: List
items:
- name: web-server-01
  targetName: ""  # Can be customized
  rootDisk: ""    # Can be specified
- name: db-server-01
  targetName: ""
  rootDisk: ""
# Additional VMs...
```

## Advanced Inventory Querying

### Query-Based VM Discovery

Use the [Tree Search Language (TSL)](../27-tsl-tree-search-language-reference) for sophisticated filtering:

```bash
# Find powered-on VMs with high memory
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where powerState = 'poweredOn' and memoryMB > 8192"

# Find VMs matching name patterns
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where name ~= 'web-.*' or name ~= 'app-.*'"

# Find VMs with multiple disks
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where len(disks) > 1"

# Complex queries with multiple conditions
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where powerState = 'poweredOn' and memoryMB >= 4096 and name ~= 'prod-.*'"
```

### Sorting and Limiting Results

Use `order by` and `limit` to control how many results are returned and in what order:

```bash
# Top 10 largest VMs by memory (sorted descending, limited to 10)
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where powerState = 'poweredOn' order by memoryMB desc limit 10"

# First 50 powered-on VMs alphabetically
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where powerState = 'poweredOn' order by name limit 50"

# Smallest VMs first (candidates for quick migration)
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where powerState = 'poweredOn' order by memoryMB asc limit 20"

# Datastores with >100 GB free (use free in bytes with SI units)
kubectl mtv get inventory datastores --provider vsphere-prod \
  --query "where free > 100Gi order by free asc limit 5"
```

### Selecting Specific Fields

Use the `select` clause to return only the fields you need, reducing output size:

```bash
# Return only name, memory, and CPU (compact output)
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "select name, memoryMB, cpuCount where powerState = 'poweredOn' limit 10"

# Select with aliases and combined with order/limit
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "select name, memoryMB as mem, cpuCount where memoryMB > 4096 order by memoryMB desc limit 5"

# Sum of disk capacity with select
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "select name, sum(disks[*].capacityGB) as totalDisk where powerState = 'poweredOn' order by totalDisk desc limit 10"
```

### Provider-Specific Resource Queries

#### vSphere Resource Discovery

```bash
# Find datastores with available space
kubectl mtv get inventory datastores --provider vsphere-prod \
  --query "where freeSpaceGB > 100"

# List resource pools by availability
kubectl mtv get inventory resource-pools --provider vsphere-prod

# Find hosts in specific clusters
kubectl mtv get inventory hosts --provider vsphere-prod \
  --query "where cluster.name = 'Production-Cluster'"
```

#### oVirt Resource Discovery

```bash
# Find VMs with specific OS types
kubectl mtv get inventory vms --provider ovirt-prod \
  --query "where guestOS ~= 'rhel.*'"

# List disk profiles
kubectl mtv get inventory disk-profiles --provider ovirt-prod

# Find storage domains
kubectl mtv get inventory storages --provider ovirt-prod \
  --query "where type = 'data'"
```

#### OpenStack Resource Discovery

```bash
# Find instances by flavor
kubectl mtv get inventory instances --provider openstack-prod \
  --query "where flavor.name = 'm1.large'"

# List available images
kubectl mtv get inventory images --provider openstack-prod \
  --query "where status = 'active'"

# Find volumes by size
kubectl mtv get inventory volumes --provider openstack-prod \
  --query "where size >= 100"
```

### Inventory Automation and Scripting

#### Automated Resource Discovery

```bash
#!/bin/bash
# Script to discover migration candidates

PROVIDER="vsphere-prod"
OUTPUT_DIR="inventory-$(date +%Y%m%d)"

mkdir -p "$OUTPUT_DIR"

# Export all VMs
kubectl mtv get inventory vms --provider "$PROVIDER" --output json > "$OUTPUT_DIR/all-vms.json"

# Export powered-on VMs only
kubectl mtv get inventory vms --provider "$PROVIDER" \
  --query "where powerState = 'poweredOn'" \
  --output planvms > "$OUTPUT_DIR/active-vms.yaml"

# Export large VMs (>50GB total disk)
kubectl mtv get inventory vms --provider "$PROVIDER" \
  --query "where sum(disks[*].capacityGB) > 50" \
  --output json > "$OUTPUT_DIR/large-vms.json"

# Export network information
kubectl mtv get inventory networks --provider "$PROVIDER" --output yaml > "$OUTPUT_DIR/networks.yaml"

# Export storage information
kubectl mtv get inventory datastores --provider "$PROVIDER" --output yaml > "$OUTPUT_DIR/datastores.yaml"

echo "Inventory exported to $OUTPUT_DIR/"
```

#### Migration Planning Automation

```bash
#!/bin/bash
# Automated migration plan generation

PROVIDER="vsphere-prod"
PLAN_NAME="auto-migration-$(date +%Y%m%d)"

# Discover migration candidates
kubectl mtv get inventory vms --provider "$PROVIDER" \
  --query "where powerState = 'poweredOn' and memoryMB <= 8192 and len(disks) <= 2" \
  --output planvms > "small-vms.yaml"

# Create migration plan
kubectl mtv create plan --name "$PLAN_NAME" \
  --source "$PROVIDER" \
  --vms @small-vms.yaml \
  --migration-type cold

echo "Created migration plan: $PLAN_NAME"
```

## Provider-Specific Inventory Features

### vSphere Provider Inventory

vSphere provides the richest inventory with hierarchical structure:

```bash
# Complete vSphere infrastructure discovery
kubectl mtv get inventory datacenters --provider vsphere-prod
kubectl mtv get inventory clusters --provider vsphere-prod
kubectl mtv get inventory hosts --provider vsphere-prod
kubectl mtv get inventory resource-pools --provider vsphere-prod
kubectl mtv get inventory folders --provider vsphere-prod
kubectl mtv get inventory datastores --provider vsphere-prod
kubectl mtv get inventory networks --provider vsphere-prod
kubectl mtv get inventory vms --provider vsphere-prod
```

### oVirt Provider Inventory

oVirt provides enterprise virtualization resources:

```bash
# oVirt-specific resources
kubectl mtv get inventory vms --provider ovirt-prod
kubectl mtv get inventory hosts --provider ovirt-prod
kubectl mtv get inventory storages --provider ovirt-prod
kubectl mtv get inventory networks --provider ovirt-prod
kubectl mtv get inventory disk-profiles --provider ovirt-prod
kubectl mtv get inventory nic-profiles --provider ovirt-prod
```

### OpenStack Provider Inventory

OpenStack provides cloud-native resource discovery:

```bash
# OpenStack-specific resources
kubectl mtv get inventory instances --provider openstack-prod
kubectl mtv get inventory images --provider openstack-prod
kubectl mtv get inventory flavors --provider openstack-prod
kubectl mtv get inventory projects --provider openstack-prod
kubectl mtv get inventory volumes --provider openstack-prod
kubectl mtv get inventory volumetypes --provider openstack-prod
kubectl mtv get inventory snapshots --provider openstack-prod
kubectl mtv get inventory subnets --provider openstack-prod
```

### OpenShift/KubeVirt Provider Inventory

For KubeVirt-to-KubeVirt migrations:

```bash
# KubeVirt-specific resources
kubectl mtv get inventory vms --provider openshift-source
kubectl mtv get inventory pvcs --provider openshift-source
kubectl mtv get inventory data-volumes --provider openshift-source
kubectl mtv get inventory namespaces --provider openshift-source
```

## Real-Time Monitoring and Watching

### Watch Mode

Monitor inventory changes in real-time:

```bash
# Watch VM state changes
kubectl mtv get inventory vms --provider vsphere-prod --watch

# Watch provider status
kubectl mtv get inventory providers --name vsphere-prod --watch

# Monitor network changes
kubectl mtv get inventory networks --provider vsphere-prod --watch
```

### Automated Monitoring

```bash
#!/bin/bash
# Monitor for migration readiness

PROVIDER="vsphere-prod"

echo "Monitoring VMs for migration readiness..."

kubectl mtv get inventory vms --provider "$PROVIDER" \
  --query "where powerState = 'poweredOn'" \
  --watch | while read -r line; do
    echo "$(date): $line"
    # Add notification logic here
done
```

## Inventory Performance and Optimization

### Large Environment Optimization

For large virtualization environments:

```bash
# Use specific queries to reduce data transfer
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where datacenter.name = 'DC-East'" \
  --output json

# Focus on specific clusters
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where cluster.name = 'Production-Cluster'"

# Limit results with targeted queries (order by + limit)
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where memoryMB >= 4096 and powerState = 'poweredOn' order by memoryMB desc limit 25" \
  --extended
```

### Inventory Caching

```bash
# Cache frequently accessed inventory
kubectl mtv get inventory vms --provider vsphere-prod --output json > vms-cache.json
kubectl mtv get inventory networks --provider vsphere-prod --output yaml > networks-cache.yaml
kubectl mtv get inventory datastores --provider vsphere-prod --output yaml > storage-cache.yaml

# Use cached data for planning
jq '.items[] | select(.powerState == "poweredOn")' vms-cache.json
```

## Troubleshooting Inventory Issues

### Common Inventory Problems

#### Provider Connectivity Issues

```bash
# Check provider status
kubectl mtv describe provider --name vsphere-prod

# Test inventory service connectivity
kubectl mtv get inventory providers --name vsphere-prod

# Check inventory service URL
echo $MTV_INVENTORY_URL
```

#### Query Syntax Errors

```bash
# Test queries with simple examples first
kubectl mtv get inventory vms --provider vsphere-prod --query "where name = 'test-vm'"

# Use JSON output to understand available fields
kubectl mtv get inventory vms --provider vsphere-prod --output json | jq '.items[0]' | head -20

# Check query syntax documentation
kubectl mtv get inventory vms --provider vsphere-prod --help
```

#### Performance Issues

```bash
# Use specific queries instead of listing all resources
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where cluster.name = 'specific-cluster'"

# Enable debug logging
kubectl mtv get inventory vms --provider vsphere-prod -v=2

# Check inventory service performance
kubectl get pods -n konveyor-forklift -l app=forklift-inventory
```

### Debug and Logging

```bash
# Enable verbose logging
kubectl mtv get inventory vms --provider vsphere-prod -v=2

# Check inventory service logs
kubectl logs -n konveyor-forklift deployment/forklift-inventory -f

# Monitor network traffic
kubectl exec -it forklift-inventory-pod -- netstat -an
```

## Integration with Migration Workflows

### Inventory-Driven Migration Planning

```bash
# 1. Discover migration candidates
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where powerState = 'poweredOn' and memoryMB <= 16384" \
  --output planvms > candidates.yaml

# 2. Analyze resource requirements
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where name in ['vm1', 'vm2', 'vm3']" \
  --output json | jq '.items[] | {name, memory: .memoryMB, disks: [.disks[].capacityGB]}'

# 3. Check network and storage availability
kubectl mtv get inventory networks --provider vsphere-prod --output yaml
kubectl mtv get inventory datastores --provider vsphere-prod \
  --query "where freeSpaceGB > 100" --output yaml

# 4. Create migration plan
kubectl mtv create plan --name inventory-driven-migration \
  --source vsphere-prod \
  --vms @candidates.yaml
```

### Validation and Verification

```bash
# Verify VMs exist before migration
VM_NAMES=$(kubectl get plan migration-plan -o json | jq -r '.spec.vms[].name')
for vm in $VM_NAMES; do
  kubectl mtv get inventory vms --provider vsphere-prod --query "where name = '$vm'" --output json | \
    jq -r '.items[0].name // "NOT FOUND: '$vm'"'
done

# Check resource availability
kubectl mtv get inventory datastores --provider vsphere-prod \
  --query "where freeSpaceGB > 500" --output table
```

## Next Steps

After mastering inventory management:

1. **Learn Query Language**: Master advanced filtering in [Chapter 10: Query Language Reference and Advanced Filtering](../10-query-language-reference-and-advanced-filtering)
2. **Create Mappings**: Define resource mappings in [Chapter 11: Mapping Management](../11-mapping-management)
3. **Plan Migrations**: Use inventory data for planning in [Chapter 13: Migration Plan Creation](../13-migration-plan-creation)
4. **Optimize Performance**: Apply inventory insights in [Chapter 16: Migration Process Optimization](../16-migration-process-optimization)

---

*Previous: [Chapter 8: VDDK Image Creation and Configuration](../08-vddk-image-creation-and-configuration)*  
*Next: [Chapter 10: Query Language Reference and Advanced Filtering](../10-query-language-reference-and-advanced-filtering)*
