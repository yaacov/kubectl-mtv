---
layout: page
title: "Chapter 7: Inventory Management"
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
kubectl mtv get inventory <resource> <provider> [flags]
```

### Basic Command Structure

- **Resource**: The type of resource to query (vm, network, storage, etc.)
- **Provider**: The name of the provider to query
- **Flags**: Optional parameters for filtering, output format, etc.

### Common Flags

- `-o, --output`: Output format (table, json, yaml, planvms for VMs)
- `-q, --query`: Query filter using Tree Search Language (TSL)
- `-w, --watch`: Watch for real-time changes
- `--extended`: Show extended information (where supported)
- `--inventory-url`: Custom inventory service URL

## Common Inventory Examples

### Virtual Machines

VMs are the most commonly queried inventory resource for migration planning:

```bash
# List all VMs from a vSphere provider
kubectl mtv get inventory vms vsphere-prod

# List VMs from different provider types
kubectl mtv get inventory vms ovirt-prod
kubectl mtv get inventory instances openstack-prod
kubectl mtv get inventory vms openshift-source

# List VMs with extended information
kubectl mtv get inventory vms vsphere-prod --extended

# Watch for VM changes in real-time
kubectl mtv get inventory vms vsphere-prod --watch
```

### Networks

Network inventory helps plan network mappings:

```bash
# List all networks from vSphere
kubectl mtv get inventory networks vsphere-prod

# List networks from oVirt
kubectl mtv get inventory networks ovirt-prod

# List subnets from OpenStack
kubectl mtv get inventory subnets openstack-prod

# View network details in YAML format
kubectl mtv get inventory networks vsphere-prod -o yaml
```

### Storage

Storage inventory assists with storage mapping configuration:

```bash
# List storage from vSphere (datastores)
kubectl mtv get inventory datastores vsphere-prod

# List storage from oVirt
kubectl mtv get inventory storages ovirt-prod

# List volume types from OpenStack
kubectl mtv get inventory volumetypes openstack-prod

# View storage details in JSON format
kubectl mtv get inventory datastores vsphere-prod -o json
```

### Hosts and Infrastructure

Discover infrastructure layout for planning:

```bash
# List ESXi hosts in vSphere
kubectl mtv get inventory hosts vsphere-prod

# List oVirt hosts
kubectl mtv get inventory hosts ovirt-prod

# List datacenters
kubectl mtv get inventory datacenters vsphere-prod

# List resource pools (vSphere)
kubectl mtv get inventory resourcepools vsphere-prod

# List clusters
kubectl mtv get inventory clusters vsphere-prod
```

### Provider Status

Check provider health and connectivity:

```bash
# List inventory from all providers
kubectl mtv get inventory providers

# Get detailed inventory from specific provider
kubectl mtv get inventory provider vsphere-prod

# Monitor provider status
kubectl mtv get inventory provider vsphere-prod --watch
```

## Output Formats

### Table Format (Default)

The default table format provides a concise overview:

```bash
# Default table output
kubectl mtv get inventory vms vsphere-prod

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
kubectl mtv get inventory vms vsphere-prod -o json

# Extract specific fields using jq
kubectl mtv get inventory vms vsphere-prod -o json | jq '.items[].name'

# Complex data extraction
kubectl mtv get inventory vms vsphere-prod -o json | \
  jq '.items[] | select(.powerState == "poweredOn") | .name'
```

### YAML Format

YAML format for human-readable structured data:

```bash
# YAML output
kubectl mtv get inventory vms vsphere-prod -o yaml

# Save to file for analysis
kubectl mtv get inventory vms vsphere-prod -o yaml > vms-inventory.yaml
```

### Extended Output

Extended output provides additional details where supported:

```bash
# Extended information for VMs
kubectl mtv get inventory vms vsphere-prod --extended

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
kubectl mtv get inventory vms vsphere-prod -o planvms

# Save to file for migration plan creation
kubectl mtv get inventory vms vsphere-prod -o planvms > migration-vms.yaml
```

### Using PlanVMs with Migration Plans

```bash
# Create migration plan using exported VMs
kubectl mtv create plan production-migration \
  --source vsphere-prod \
  --vms @migration-vms.yaml

# Or use the planvms format directly in plan creation
kubectl mtv get inventory vms vsphere-prod \
  -q "where name ~= 'prod-.*' and powerState = 'poweredOn'" \
  -o planvms > prod-vms.yaml

kubectl mtv create plan prod-migration \
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

Use the Tree Search Language (TSL) for sophisticated filtering:

```bash
# Find powered-on VMs with high memory
kubectl mtv get inventory vms vsphere-prod \
  -q "where powerState = 'poweredOn' and memoryMB > 8192"

# Find VMs matching name patterns
kubectl mtv get inventory vms vsphere-prod \
  -q "where name ~= 'web-.*' or name ~= 'app-.*'"

# Find VMs with multiple disks
kubectl mtv get inventory vms vsphere-prod \
  -q "where len disks > 1"

# Complex queries with multiple conditions
kubectl mtv get inventory vms vsphere-prod \
  -q "where powerState = 'poweredOn' and memoryMB >= 4096 and name ~= 'prod-.*'"
```

### Provider-Specific Resource Queries

#### vSphere Resource Discovery

```bash
# Find datastores with available space
kubectl mtv get inventory datastores vsphere-prod \
  -q "where freeSpaceGB > 100"

# List resource pools by availability
kubectl mtv get inventory resourcepools vsphere-prod

# Find hosts in specific clusters
kubectl mtv get inventory hosts vsphere-prod \
  -q "where cluster.name = 'Production-Cluster'"
```

#### oVirt Resource Discovery

```bash
# Find VMs with specific OS types
kubectl mtv get inventory vms ovirt-prod \
  -q "where guestOS ~= 'rhel.*'"

# List disk profiles
kubectl mtv get inventory diskprofiles ovirt-prod

# Find storage domains
kubectl mtv get inventory storages ovirt-prod \
  -q "where type = 'data'"
```

#### OpenStack Resource Discovery

```bash
# Find instances by flavor
kubectl mtv get inventory instances openstack-prod \
  -q "where flavor.name = 'm1.large'"

# List available images
kubectl mtv get inventory images openstack-prod \
  -q "where status = 'active'"

# Find volumes by size
kubectl mtv get inventory volumes openstack-prod \
  -q "where size >= 100"
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
kubectl mtv get inventory vms "$PROVIDER" -o json > "$OUTPUT_DIR/all-vms.json"

# Export powered-on VMs only
kubectl mtv get inventory vms "$PROVIDER" \
  -q "where powerState = 'poweredOn'" \
  -o planvms > "$OUTPUT_DIR/active-vms.yaml"

# Export large VMs (>50GB total disk)
kubectl mtv get inventory vms "$PROVIDER" \
  -q "where sum disks.capacityGB > 50" \
  -o json > "$OUTPUT_DIR/large-vms.json"

# Export network information
kubectl mtv get inventory networks "$PROVIDER" -o yaml > "$OUTPUT_DIR/networks.yaml"

# Export storage information
kubectl mtv get inventory datastores "$PROVIDER" -o yaml > "$OUTPUT_DIR/datastores.yaml"

echo "Inventory exported to $OUTPUT_DIR/"
```

#### Migration Planning Automation

```bash
#!/bin/bash
# Automated migration plan generation

PROVIDER="vsphere-prod"
PLAN_NAME="auto-migration-$(date +%Y%m%d)"

# Discover migration candidates
kubectl mtv get inventory vms "$PROVIDER" \
  -q "where powerState = 'poweredOn' and memoryMB <= 8192 and len disks <= 2" \
  -o planvms > "small-vms.yaml"

# Create migration plan
kubectl mtv create plan "$PLAN_NAME" \
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
kubectl mtv get inventory datacenters vsphere-prod
kubectl mtv get inventory clusters vsphere-prod
kubectl mtv get inventory hosts vsphere-prod
kubectl mtv get inventory resourcepools vsphere-prod
kubectl mtv get inventory folders vsphere-prod
kubectl mtv get inventory datastores vsphere-prod
kubectl mtv get inventory networks vsphere-prod
kubectl mtv get inventory vms vsphere-prod
```

### oVirt Provider Inventory

oVirt provides enterprise virtualization resources:

```bash
# oVirt-specific resources
kubectl mtv get inventory vms ovirt-prod
kubectl mtv get inventory hosts ovirt-prod
kubectl mtv get inventory storages ovirt-prod
kubectl mtv get inventory networks ovirt-prod
kubectl mtv get inventory diskprofiles ovirt-prod
kubectl mtv get inventory nicprofiles ovirt-prod
```

### OpenStack Provider Inventory

OpenStack provides cloud-native resource discovery:

```bash
# OpenStack-specific resources
kubectl mtv get inventory instances openstack-prod
kubectl mtv get inventory images openstack-prod
kubectl mtv get inventory flavors openstack-prod
kubectl mtv get inventory projects openstack-prod
kubectl mtv get inventory volumes openstack-prod
kubectl mtv get inventory volumetypes openstack-prod
kubectl mtv get inventory snapshots openstack-prod
kubectl mtv get inventory subnets openstack-prod
```

### OpenShift/KubeVirt Provider Inventory

For KubeVirt-to-KubeVirt migrations:

```bash
# KubeVirt-specific resources
kubectl mtv get inventory vms openshift-source
kubectl mtv get inventory pvcs openshift-source
kubectl mtv get inventory datavolumes openshift-source
kubectl mtv get inventory namespaces openshift-source
```

## Real-Time Monitoring and Watching

### Watch Mode

Monitor inventory changes in real-time:

```bash
# Watch VM state changes
kubectl mtv get inventory vms vsphere-prod --watch

# Watch provider status
kubectl mtv get inventory provider vsphere-prod --watch

# Monitor network changes
kubectl mtv get inventory networks vsphere-prod --watch
```

### Automated Monitoring

```bash
#!/bin/bash
# Monitor for migration readiness

PROVIDER="vsphere-prod"

echo "Monitoring VMs for migration readiness..."

kubectl mtv get inventory vms "$PROVIDER" \
  -q "where powerState = 'poweredOn'" \
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
kubectl mtv get inventory vms vsphere-prod \
  -q "where datacenter.name = 'DC-East'" \
  -o json

# Focus on specific clusters
kubectl mtv get inventory vms vsphere-prod \
  -q "where cluster.name = 'Production-Cluster'"

# Limit results with targeted queries
kubectl mtv get inventory vms vsphere-prod \
  -q "where memoryMB >= 4096 and powerState = 'poweredOn'" \
  --extended
```

### Inventory Caching

```bash
# Cache frequently accessed inventory
kubectl mtv get inventory vms vsphere-prod -o json > vms-cache.json
kubectl mtv get inventory networks vsphere-prod -o yaml > networks-cache.yaml
kubectl mtv get inventory datastores vsphere-prod -o yaml > storage-cache.yaml

# Use cached data for planning
jq '.items[] | select(.powerState == "poweredOn")' vms-cache.json
```

## Troubleshooting Inventory Issues

### Common Inventory Problems

#### Provider Connectivity Issues

```bash
# Check provider status
kubectl mtv describe provider vsphere-prod

# Test inventory service connectivity
kubectl mtv get inventory provider vsphere-prod

# Check inventory service URL
echo $MTV_INVENTORY_URL
```

#### Query Syntax Errors

```bash
# Test queries with simple examples first
kubectl mtv get inventory vms vsphere-prod -q "where name = 'test-vm'"

# Use JSON output to understand available fields
kubectl mtv get inventory vms vsphere-prod -o json | jq '.items[0]' | head -20

# Check query syntax documentation
kubectl mtv get inventory vms vsphere-prod --help
```

#### Performance Issues

```bash
# Use specific queries instead of listing all resources
kubectl mtv get inventory vms vsphere-prod \
  -q "where cluster.name = 'specific-cluster'"

# Enable debug logging
kubectl mtv get inventory vms vsphere-prod -v=2

# Check inventory service performance
kubectl get pods -n konveyor-forklift -l app=forklift-inventory
```

### Debug and Logging

```bash
# Enable verbose logging
kubectl mtv get inventory vms vsphere-prod -v=2

# Check inventory service logs
kubectl logs -n konveyor-forklift deployment/forklift-inventory -f

# Monitor network traffic
kubectl exec -it forklift-inventory-pod -- netstat -an
```

## Integration with Migration Workflows

### Inventory-Driven Migration Planning

```bash
# 1. Discover migration candidates
kubectl mtv get inventory vms vsphere-prod \
  -q "where powerState = 'poweredOn' and memoryMB <= 16384" \
  -o planvms > candidates.yaml

# 2. Analyze resource requirements
kubectl mtv get inventory vms vsphere-prod \
  -q "where name in ('vm1', 'vm2', 'vm3')" \
  -o json | jq '.items[] | {name, memory: .memoryMB, disks: [.disks[].capacityGB]}'

# 3. Check network and storage availability
kubectl mtv get inventory networks vsphere-prod -o yaml
kubectl mtv get inventory datastores vsphere-prod \
  -q "where freeSpaceGB > 100" -o yaml

# 4. Create migration plan
kubectl mtv create plan inventory-driven-migration \
  --source vsphere-prod \
  --vms @candidates.yaml
```

### Validation and Verification

```bash
# Verify VMs exist before migration
VM_NAMES=$(kubectl get plan migration-plan -o json | jq -r '.spec.vms[].name')
for vm in $VM_NAMES; do
  kubectl mtv get inventory vms vsphere-prod -q "where name = '$vm'" -o json | \
    jq -r '.items[0].name // "NOT FOUND: '$vm'"'
done

# Check resource availability
kubectl mtv get inventory datastores vsphere-prod \
  -q "where freeSpaceGB > 500" -o table
```

## Next Steps

After mastering inventory management:

1. **Learn Query Language**: Master advanced filtering in [Chapter 8: Query Language Reference and Advanced Filtering](/kubectl-mtv/08-query-language-reference-and-advanced-filtering)
2. **Create Mappings**: Define resource mappings in [Chapter 9: Mapping Management](/kubectl-mtv/09-mapping-management)
3. **Plan Migrations**: Use inventory data for planning in [Chapter 10: Migration Plan Creation](/kubectl-mtv/10-migration-plan-creation)
4. **Optimize Performance**: Apply inventory insights in [Chapter 13: Migration Process Optimization](/kubectl-mtv/13-migration-process-optimization)

---

*Previous: [Chapter 6: VDDK Image Creation and Configuration](/kubectl-mtv/06-vddk-image-creation-and-configuration)*  
*Next: [Chapter 8: Query Language Reference and Advanced Filtering](/kubectl-mtv/08-query-language-reference-and-advanced-filtering)*
