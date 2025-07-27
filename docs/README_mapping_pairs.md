# Mapping Configuration in kubectl-mtv

This guide covers the different ways to configure network and storage mappings when creating migration plans and mappings in kubectl-mtv.

## Overview

Network and storage mappings define how resources from the source provider are mapped to resources in the target provider. There are multiple ways to configure these mappings:

1. **Using Existing Mappings**: Reference pre-created mapping resources
2. **Inline Mapping Pairs**: Define mappings directly in the command
3. **Default Mappings**: Let kubectl-mtv create automatic mappings

## Creating Standalone Mappings with Pairs

You can create reusable network and storage mappings that can be referenced by multiple migration plans.

### Network Mappings

```bash
# Create a network mapping with specific pairs
kubectl-mtv create mapping network production-networks \
  --source vsphere-provider \
  --target openshift-provider \
  --network-pairs "VM Network:openshift-sdn/production,Management:pod,DMZ:security/dmz-network"
```

Network pair format: `source-network:target`
- `source-network`: The source network name
- `target`: Can be:
  - `namespace/network-name`: Specific NetworkAttachmentDefinition
  - `network-name`: Uses the mapping's namespace
  - `pod`: Use pod networking
  - `ignored`: Ignore this network

### Storage Mappings

```bash
# Create a storage mapping with specific pairs
kubectl-mtv create mapping storage production-storage \
  --source vsphere-provider \
  --target openshift-provider \
  --storage-pairs "VMFS-Datastore-1:standard,VMFS-Fast-Storage:fast-ssd,Archive-Storage:slow-archive"
```

Storage pair format: `source-storage:storage-class`
- `source-storage`: The source datastore/storage domain name
- `storage-class`: The target Kubernetes StorageClass (cluster-scoped)

## Using Mappings When Creating Plans

When creating a migration plan, you have three options for configuring mappings:

### Option 1: Using Existing Mappings

Reference pre-created mapping resources by name:

```bash
# Create plan using existing mappings
kubectl-mtv create plan myplan \
  --source vsphere-provider \
  --target openshift-provider \
  --network-mapping production-networks \
  --storage-mapping production-storage \
  --vms vm1,vm2,vm3
```

**Benefits:**
- Reusable across multiple plans
- Can be managed independently
- Easier to maintain for large deployments
- Can be version controlled as separate resources

### Option 2: Using Inline Mapping Pairs

Define mappings directly in the plan creation command:

```bash
# Create plan with inline mapping pairs
kubectl-mtv create plan myplan \
  --source vsphere-provider \
  --target openshift-provider \
  --network-pairs "VM Network:pod,Management:openshift-sdn/mgmt-net" \
  --storage-pairs "datastore1:standard,datastore2:fast-ssd" \
  --vms vm1,vm2,vm3
```

**Benefits:**
- Quick one-time migrations
- No need to create separate mapping resources
- All configuration in one command
- Automatically creates mappings named `<plan-name>-network` and `<plan-name>-storage`

### Option 3: Using Default Mappings

Let kubectl-mtv automatically create mappings based on defaults:

```bash
# Create plan with default mappings
kubectl-mtv create plan myplan \
  --source vsphere-provider \
  --target openshift-provider \
  --default-target-network pod \
  --default-target-storage-class standard \
  --vms vm1,vm2,vm3
```

**Benefits:**
- Simplest approach
- Good for testing and simple environments
- Automatically maps all source resources to specified defaults

## Important Rules and Constraints

### Mutual Exclusivity

You cannot mix different mapping approaches in the same command:

```bash
# INVALID - Cannot use both existing mapping and pairs
kubectl-mtv create plan myplan \
  --network-mapping existing-map \
  --network-pairs "VM Network:pod" \
  --vms vm1

# INVALID - Cannot use both existing mapping and defaults
kubectl-mtv create plan myplan \
  --storage-mapping existing-map \
  --default-target-storage-class standard \
  --vms vm1
```

### Valid Combinations

You can mix network and storage approaches:

```bash
# VALID - Use existing network mapping with storage pairs
kubectl-mtv create plan myplan \
  --source vsphere-provider \
  --target openshift-provider \
  --network-mapping existing-network-map \
  --storage-pairs "datastore1:standard" \
  --vms vm1,vm2

# VALID - Use network pairs with existing storage mapping
kubectl-mtv create plan myplan \
  --source vsphere-provider \
  --target openshift-provider \
  --network-pairs "VM Network:pod" \
  --storage-mapping existing-storage-map \
  --vms vm1,vm2
```

## Complete Examples

### Enterprise Migration with Existing Mappings

```bash
# Step 1: Create reusable network mapping
kubectl-mtv create mapping network enterprise-networks \
  --source vsphere-datacenter1 \
  --target openshift-prod \
  --network-pairs "Production VLAN 100:prod/production-net,Development VLAN 200:dev/development-net,DMZ VLAN 300:security/dmz-net,Management VLAN 10:mgmt/management-net"

# Step 2: Create reusable storage mapping
kubectl-mtv create mapping storage enterprise-storage \
  --source vsphere-datacenter1 \
  --target openshift-prod \
  --storage-pairs "SSD-Tier1:ultra-fast-ssd,SSD-Tier2:fast-ssd,HDD-Tier3:standard,Archive:slow-archive"

# Step 3: Create multiple plans using the same mappings
kubectl-mtv create plan web-servers-migration \
  --source vsphere-datacenter1 \
  --target openshift-prod \
  --network-mapping enterprise-networks \
  --storage-mapping enterprise-storage \
  --vms web-01,web-02,web-03,web-04

kubectl-mtv create plan database-migration \
  --source vsphere-datacenter1 \
  --target openshift-prod \
  --network-mapping enterprise-networks \
  --storage-mapping enterprise-storage \
  --vms db-primary,db-replica-01,db-replica-02
```

### Quick Development Migration with Inline Pairs

```bash
# One-command migration for development environment
kubectl-mtv create plan dev-migration \
  --source vsphere-dev \
  --target openshift-dev \
  --network-pairs "Dev Network:pod" \
  --storage-pairs "dev-datastore:gp2" \
  --vms dev-app-01,dev-db-01 \
  --target-namespace development
```

### Simple Test Migration with Defaults

```bash
# Minimal configuration for testing
kubectl-mtv create plan test-migration \
  --source vsphere-test \
  --target openshift-test \
  --default-target-network pod \
  --default-target-storage-class standard \
  --vms test-vm-01
```

## Best Practices

1. **Use Existing Mappings for Production**: Create and test mappings separately, then reference them in plans
2. **Use Inline Pairs for Development**: Quick iterations and one-off migrations
3. **Use Defaults for Testing**: Simple testing scenarios where specific mappings don't matter
4. **Document Your Mappings**: Keep track of which source resources map to which target resources
5. **Validate Before Migration**: Always verify mappings are correct before starting migration

## Troubleshooting

### View Created Mappings

```bash
# List all network mappings
kubectl-mtv get mapping network

# List all storage mappings
kubectl-mtv get mapping storage

# View specific mapping details
kubectl-mtv get mapping network enterprise-networks -o yaml
```

### Common Issues

1. **Source Network/Storage Not Found**: Ensure the source names match exactly (case-sensitive)
2. **Target Resource Not Found**: Verify the target namespace and resource exist
3. **Conflicting Flags**: Remember you cannot mix existing mappings with pairs or defaults 