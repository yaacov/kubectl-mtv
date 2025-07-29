# Patching Network and Storage Mappings

This guide explains how to modify existing network and storage mappings using the `kubectl-mtv patch mapping` command. The patch command allows you to add, update, or remove mapping pairs without recreating the entire mapping resource.

## Overview

The patch command provides flexible operations for modifying existing mappings:
- **Add pairs**: Append new source-to-target mappings
- **Update pairs**: Modify existing mappings by source name/ID
- **Remove pairs**: Delete specific mappings by source name/ID

All operations can be combined in a single command and support the same friendly pair formats used in the create command.

## Network Mappings

### Basic Syntax

```bash
kubectl-mtv patch mapping network <mapping-name> \
  [--add-pairs "<source>:<target>,<source>:<target>,..."] \
  [--update-pairs "<source>:<target>,<source>:<target>,..."] \
  [--remove-pairs "<source>,<source>,..."] \
  [--inventory-url <url>]
```

### Network Pair Format

Network pairs use the same format as the create command: `source-network:target-specification`

#### Target Specifications

1. **Multus Network**: `namespace/network-name`
2. **Pod Network**: `pod`
3. **Ignored Network**: `ignored`
4. **Default Namespace**: `network-name` (uses mapping's namespace)

### Network Patch Examples

#### Adding Network Pairs

```bash
# Add new network mappings
kubectl-mtv patch mapping network production-networks \
  --add-pairs "VM Network:openshift-sdn/production,Management:pod"
```

#### Updating Network Pairs

```bash
# Update existing network mappings
kubectl-mtv patch mapping network production-networks \
  --update-pairs "VM Network:openshift-sdn/staging,DMZ:security/dmz-network"
```

#### Removing Network Pairs

```bash
# Remove network mappings by source name
kubectl-mtv patch mapping network production-networks \
  --remove-pairs "Old-Network,Unused-VLAN"
```

#### Combined Operations

```bash
# Perform multiple operations in one command
kubectl-mtv patch mapping network production-networks \
  --add-pairs "New-Network:production/new-net" \
  --update-pairs "VM Network:production/updated-net" \
  --remove-pairs "Deprecated-Network"
```

## Storage Mappings

### Basic Syntax

```bash
kubectl-mtv patch mapping storage <mapping-name> \
  [--add-pairs "<source>:<storage-class>,<source>:<storage-class>,..."] \
  [--update-pairs "<source>:<storage-class>,<source>:<storage-class>,..."] \
  [--remove-pairs "<source>,<source>,..."] \
  [--inventory-url <url>]
```

### Storage Pair Format

Storage pairs follow the format: `source-storage:storage-class`

- **source-storage**: The source datastore/storage domain name
- **storage-class**: The target Kubernetes StorageClass (cluster-scoped)

### Storage Patch Examples

#### Adding Storage Pairs

```bash
# Add new storage mappings
kubectl-mtv patch mapping storage production-storage \
  --add-pairs "VMFS-Datastore-1:standard,VMFS-Fast-Storage:fast-ssd"
```

#### Updating Storage Pairs

```bash
# Update existing storage mappings
kubectl-mtv patch mapping storage production-storage \
  --update-pairs "VMFS-Datastore-1:premium-ssd,Archive-Storage:slow-archive"
```

#### Removing Storage Pairs

```bash
# Remove storage mappings by source name
kubectl-mtv patch mapping storage production-storage \
  --remove-pairs "Old-Datastore,Decommissioned-Storage"
```

#### Combined Operations

```bash
# Perform multiple operations in one command
kubectl-mtv patch mapping storage production-storage \
  --add-pairs "New-Datastore:standard" \
  --update-pairs "VMFS-Fast-Storage:premium-nvme" \
  --remove-pairs "Legacy-Storage"
```

## Advanced Usage

### Working with Provider Types

The patch command automatically detects the source provider type and uses the appropriate inventory resolution:

- **VMware vSphere**: Resolves datastore names to IDs
- **Red Hat Virtualization (oVirt)**: Resolves storage domain names to IDs  
- **OpenStack**: Supports volume types and `__DEFAULT__` storage
- **OpenShift**: Uses storage class names directly
- **OVA**: Resolves storage references from OVA files

### Duplicate Source Handling

When adding pairs, the system automatically prevents duplicates:

```bash
$ kubectl-mtv patch mapping network my-mapping --add-pairs "VM Network:production"
Warning: Skipping duplicate sources: VM Network
networkmap/my-mapping patched
```

The command will:
- Detect existing sources by name or ID
- Display warnings for duplicates
- Skip duplicate pairs and continue with unique ones
- Provide feedback on actual changes made

### Multiple Source Resolution

Some sources may resolve to multiple resources (e.g., networks with the same name in different datacenters). The patch command handles this automatically:

```bash
# If "VM Network" exists in multiple datacenters, pairs will be created for each
kubectl-mtv patch mapping network multi-dc-mapping \
  --add-pairs "VM Network:production/shared-net"
```

## Best Practices

### 1. Use Descriptive Names

```bash
# Good: descriptive source names
--add-pairs "Production-VLAN-100:production/app-network"

# Avoid: generic names that might conflict
--add-pairs "Network:production/app-network"
```

### 2. Verify Before Patching

```bash
# Check current mapping state
kubectl-mtv describe mapping network production-networks

# Then apply patches
kubectl-mtv patch mapping network production-networks --add-pairs "..."
```

### 3. Group Related Operations

```bash
# Efficient: combine related changes
kubectl-mtv patch mapping network production-networks \
  --remove-pairs "Old-Network" \
  --add-pairs "New-Network:production/new-net"

# Less efficient: separate commands
kubectl-mtv patch mapping network production-networks --remove-pairs "Old-Network"
kubectl-mtv patch mapping network production-networks --add-pairs "New-Network:production/new-net"
```

### 4. Handle Large Changes Incrementally

For mappings with many pairs, consider incremental updates:

```bash
# Update in batches
kubectl-mtv patch mapping storage large-mapping \
  --update-pairs "DS1:premium,DS2:premium,DS3:premium"

kubectl-mtv patch mapping storage large-mapping \
  --update-pairs "DS4:standard,DS5:standard,DS6:standard"
```

## Troubleshooting

### Common Issues

#### Source Not Found
```
Error: failed to resolve source network 'NonExistent-Network'
```
**Solution**: Verify the source name exists in the provider inventory:
```bash
kubectl-mtv get inventory networks --provider source-provider
```

#### Invalid Target Format
```
Error: invalid target format 'invalid-target'
```
**Solution**: Use correct target format:
- Networks: `namespace/name`, `name`, `pod`, or `ignored`
- Storage: `storage-class-name`

#### Permission Issues
```
Error: failed to update network mapping: access denied
```
**Solution**: Ensure you have update permissions for the mapping resource in the target namespace.

### Debugging

Enable verbose logging to troubleshoot issues:

```bash
# Basic logging
kubectl-mtv patch mapping network my-mapping --add-pairs "..." -v 2

# Detailed logging  
kubectl-mtv patch mapping network my-mapping --add-pairs "..." -v 3
```

## Integration with Migration Plans

Patched mappings are immediately available for use in migration plans:

```bash
# Patch the mapping
kubectl-mtv patch mapping network production-networks --add-pairs "New-Network:prod/new-net"

# Use in a new migration plan
kubectl-mtv create plan my-plan \
  --source-provider vsphere-provider \
  --target-provider openshift-provider \
  --network-mapping production-networks \
  --vm "my-vm"
```

## Related Commands

- `kubectl-mtv create mapping` - Create new mappings
- `kubectl-mtv get mapping` - List existing mappings  
- `kubectl-mtv describe mapping` - View mapping details
- `kubectl-mtv delete mapping` - Delete mappings

## See Also

- [Creating Mappings](README_create_mappings.md) - How to create initial mappings
- [Mapping Pairs](README_mapping_pairs.md) - Understanding mapping pair formats
- [Migration Plans](README_planvms.md) - Using mappings in migration plans 