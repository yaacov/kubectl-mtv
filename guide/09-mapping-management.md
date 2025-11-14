---
layout: page
title: "Chapter 9: Mapping Management"
---

# Chapter 9: Mapping Management

Mappings define the critical relationships between source and target resources, ensuring VMs are migrated to appropriate networks and storage systems. This chapter covers comprehensive mapping management for both network and storage resources.

For foundational concepts about network and storage mappings, see the [official Forklift mapping documentation](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#about-network-maps-in-migration-plans) and [storage mapping documentation](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#about-storage-maps-in-migration-plans).

## Overview

### What are Mappings?

Mappings define source-to-target resource relationships for migration:

- **Network Mappings**: Define how source VM networks map to target Kubernetes/OpenShift networks
- **Storage Mappings**: Define how source storage systems map to target StorageClasses

### Why Mappings are Essential

1. **Resource Translation**: Convert source virtualization concepts to Kubernetes resources
2. **Policy Enforcement**: Ensure VMs land on appropriate networks and storage classes
3. **Performance Optimization**: Route workloads to optimal target resources
4. **Security Compliance**: Maintain network isolation and storage encryption requirements

### Mapping Types Supported

| Mapping Type | Source Resources | Target Resources |
|--------------|------------------|------------------|
| **Network** | VM Networks, Port Groups, VLANs | Pod networking, Multus networks, Default networking |
| **Storage** | Datastores, Storage Domains, Volumes | StorageClasses with volume modes and access patterns |

## Listing, Viewing, and Deleting Mappings

### List Mappings

```bash
# List all network mappings
kubectl mtv get mappings network

# List all storage mappings
kubectl mtv get mappings storage

# List mappings in specific namespace
kubectl mtv get mappings network -n migration-namespace

# List mappings across all namespaces
kubectl mtv get mappings storage --all-namespaces
```

### View Mapping Details

```bash
# Describe network mapping
kubectl mtv describe mapping network my-network-mapping

# Describe storage mapping
kubectl mtv describe mapping storage my-storage-mapping

# View mapping in YAML format
kubectl get networkmapping my-network-mapping -o yaml
kubectl get storagemapping my-storage-mapping -o yaml
```

### Delete Mappings

```bash
# Delete specific network mapping
kubectl mtv delete mapping network my-network-mapping

# Delete specific storage mapping
kubectl mtv delete mapping storage my-storage-mapping

# Delete multiple mappings
kubectl mtv delete mapping network mapping1 mapping2 mapping3

# Delete all network mappings in namespace (use with caution)
kubectl mtv delete mapping network --all

# Delete all storage mappings in namespace (use with caution)
kubectl mtv delete mapping storage --all
```

## How-To: Creating Mappings

### Network Mapping Creation

Network mappings translate source VM networks to target Kubernetes networking:

#### Basic Network Mapping Syntax

```bash
kubectl mtv create mapping network NAME \
  --source SOURCE_PROVIDER \
  --target TARGET_PROVIDER \
  --network-pairs "source1:target1,source2:target2,..."
```

#### Network Mapping Pairs Format

kubectl-mtv supports four network mapping formats verified from the command code:

| Format | Description | Example |
|--------|-------------|---------|
| `source:target-namespace/target-network` | Multus network with namespace | `'VM Network:multus-system/bridge-network'` |
| `source:target-network` | Multus network in same namespace | `'Management Network:mgmt-network'` |
| `source:default` | Default pod networking | `'VM Network:default'` |
| `source:ignored` | Ignore/skip network | `'Backup Network:ignored'` |

#### Network Mapping Examples

##### Basic Network Mapping

```bash
# Simple network mapping for vSphere to OpenShift
kubectl mtv create mapping network prod-network-mapping \
  --source vsphere-prod \
  --target openshift-target \
  --network-pairs "VM Network:default,Management Network:multus-system/mgmt-net"
```

##### Advanced Network Mapping

```bash
# Comprehensive network mapping with multiple network types
kubectl mtv create mapping network comprehensive-network \
  --source vsphere-prod \
  --target openshift-target \
  --network-pairs "Production VLAN:prod-network,Management Network:multus-system/management,DMZ Network:dmz-net,Backup Network:ignored"
```

##### Development Environment Mapping

```bash
# Development network mapping with simplified targeting
kubectl mtv create mapping network dev-network-mapping \
  --source vsphere-dev \
  --target openshift-dev \
  --network-pairs "Dev Network:default,Test Network:default,Isolated Network:ignored"
```

##### Multi-Namespace Network Mapping

```bash
# Cross-namespace network mapping
kubectl mtv create mapping network cross-ns-mapping \
  --source ovirt-prod \
  --target openshift-target \
  --network-pairs "ovirtmgmt:openshift-sdn/default,production:prod-namespace/prod-net,dmz:security-namespace/dmz-network"
```

### Storage Mapping Creation

Storage mappings define how source storage systems map to target StorageClasses with enhanced options:

#### Basic Storage Mapping Syntax

```bash
kubectl mtv create mapping storage NAME \
  --source SOURCE_PROVIDER \
  --target TARGET_PROVIDER \
  --storage-pairs "source1:storageclass1,source2:storageclass2,..."
```

#### Storage Mapping Pairs Format

Storage pairs support enhanced options verified from the command code:

```
'source:storage-class[;volumeMode=Block|Filesystem][;accessMode=ReadWriteOnce|ReadWriteMany|ReadOnlyMany][;offloadPlugin=vsphere][;offloadSecret=secret-name][;offloadVendor=vendor-name]'
```

#### Enhanced Storage Options

##### Volume Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `Filesystem` | Traditional file system access | General application storage |
| `Block` | Raw block device access | Databases, high-performance applications |

##### Access Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `ReadWriteOnce` | Read-write by single node | Single-instance applications |
| `ReadWriteMany` | Read-write by multiple nodes | Shared file systems |
| `ReadOnlyMany` | Read-only by multiple nodes | Shared configuration, content |

##### Offload Vendors (Storage Array Integration)

Verified vendor list from command code:
- `flashsystem` - IBM FlashSystem
- `vantara` - Hitachi Vantara
- `ontap` - NetApp ONTAP
- `primera3par` - HPE Primera/3PAR
- `pureFlashArray` - Pure Storage FlashArray
- `powerflex` - Dell PowerFlex
- `powermax` - Dell PowerMax
- `powerstore` - Dell PowerStore
- `infinibox` - Infinidat InfiniBox

#### Storage Mapping Examples

##### Basic Storage Mapping

```bash
# Simple storage mapping for vSphere to OpenShift
kubectl mtv create mapping storage basic-storage-mapping \
  --source vsphere-prod \
  --target openshift-target \
  --storage-pairs "datastore1:fast-ssd,datastore2:standard-storage"
```

##### Advanced Storage Mapping with Enhanced Options

```bash
# Storage mapping with volume modes and access patterns
kubectl mtv create mapping storage advanced-storage \
  --source vsphere-prod \
  --target openshift-target \
  --storage-pairs "fast-datastore:premium-ssd;volumeMode=Block;accessMode=ReadWriteOnce,shared-datastore:shared-storage;volumeMode=Filesystem;accessMode=ReadWriteMany,archive-datastore:backup-storage;accessMode=ReadOnlyMany"
```

##### Storage Array Offloading Configuration

```bash
# Storage mapping with FlashSystem offloading
kubectl mtv create mapping storage offload-flashsystem \
  --source vsphere-prod \
  --target openshift-target \
  --storage-pairs "production-ds:flashsystem-fast;offloadPlugin=vsphere;offloadVendor=flashsystem;offloadSecret=flashsystem-creds" \
  --default-volume-mode Block \
  --default-access-mode ReadWriteOnce
```

##### NetApp ONTAP Offloading with Secret Creation

```bash
# Create storage mapping with automatic offload secret creation
kubectl mtv create mapping storage ontap-offload \
  --source vsphere-prod \
  --target openshift-target \
  --storage-pairs "netapp-datastore:ontap-gold;offloadPlugin=vsphere;offloadVendor=ontap" \
  --offload-vsphere-username vcenter-service@vsphere.local \
  --offload-vsphere-password VCenterPassword123 \
  --offload-vsphere-url https://vcenter.company.com \
  --offload-storage-username ontap-admin \
  --offload-storage-password NetAppPassword123 \
  --offload-storage-endpoint https://netapp-cluster.company.com
```

##### Default Options for Multiple Storage Classes

```bash
# Storage mapping with global defaults
kubectl mtv create mapping storage default-options \
  --source ovirt-prod \
  --target openshift-target \
  --storage-pairs "ovirt-data:standard-rwo,ovirt-shared:shared-rwx,ovirt-fast:premium-block" \
  --default-volume-mode Filesystem \
  --default-access-mode ReadWriteOnce \
  --default-offload-plugin vsphere \
  --default-offload-vendor vantara
```

#### Complete Storage Array Offloading Examples

##### IBM FlashSystem Configuration

```bash
# FlashSystem storage array offloading
kubectl mtv create mapping storage flashsystem-prod \
  --source vsphere-prod \
  --target openshift-target \
  --storage-pairs "ssd-datastore:flashsystem-tier1;offloadPlugin=vsphere;offloadVendor=flashsystem;volumeMode=Block;accessMode=ReadWriteOnce" \
  --offload-vsphere-username svc-flashsystem@vsphere.local \
  --offload-vsphere-password FlashSystemVCPassword \
  --offload-vsphere-url https://vcenter.prod.company.com \
  --offload-storage-username flashsystem-admin \
  --offload-storage-password FlashSystemPassword \
  --offload-storage-endpoint https://flashsystem.company.com:7443
```

##### Pure Storage FlashArray Configuration

```bash
# Pure Storage FlashArray offloading
kubectl mtv create mapping storage pure-flasharray \
  --source vsphere-prod \
  --target openshift-target \
  --storage-pairs "pure-datastore:pure-block-gold;offloadPlugin=vsphere;offloadVendor=pureFlashArray;volumeMode=Block" \
  --offload-storage-username pureuser \
  --offload-storage-password PureStorageAPIKey \
  --offload-storage-endpoint https://pure-array.company.com \
  --offload-cacert @/certs/pure-ca.pem
```

##### Dell PowerMax Configuration

```bash
# Dell PowerMax storage array configuration
kubectl mtv create mapping storage powermax-enterprise \
  --source vsphere-prod \
  --target openshift-target \
  --storage-pairs "powermax-gold:powermax-tier1;offloadPlugin=vsphere;offloadVendor=powermax;volumeMode=Block;accessMode=ReadWriteOnce" \
  --offload-storage-username powermax-svc \
  --offload-storage-password PowerMaxPassword \
  --offload-storage-endpoint https://powermax.company.com:8443 \
  --default-volume-mode Block
```

## How-To: Patching Mappings

Mapping patching allows you to add, update, or remove pairs without recreating the entire mapping:

### Network Mapping Patching

#### Add Network Pairs

```bash
# Add new network pairs to existing mapping
kubectl mtv patch mapping network prod-network-mapping \
  --add-pairs "New VLAN:new-network,Guest Network:guest-net"
```

#### Update Network Pairs

```bash
# Update existing network mappings
kubectl mtv patch mapping network prod-network-mapping \
  --update-pairs "Management Network:multus-system/new-mgmt-net,Production VLAN:updated-prod-network"
```

#### Remove Network Pairs

```bash
# Remove networks from mapping
kubectl mtv patch mapping network prod-network-mapping \
  --remove-pairs "Old Network,Deprecated VLAN"
```

#### Combined Network Operations

```bash
# Perform multiple operations in single command
kubectl mtv patch mapping network prod-network-mapping \
  --add-pairs "New Production:prod-new" \
  --update-pairs "Management Network:mgmt-updated" \
  --remove-pairs "Legacy Network"
```

### Storage Mapping Patching

#### Add Storage Pairs

```bash
# Add new storage pairs with enhanced options
kubectl mtv patch mapping storage prod-storage-mapping \
  --add-pairs "new-datastore:premium-nvme;volumeMode=Block;accessMode=ReadWriteOnce,backup-datastore:backup-storage;volumeMode=Filesystem;accessMode=ReadOnlyMany"
```

#### Update Storage Pairs

```bash
# Update existing storage mappings with new options
kubectl mtv patch mapping storage prod-storage-mapping \
  --update-pairs "fast-datastore:ultra-fast-ssd;volumeMode=Block;offloadPlugin=vsphere;offloadVendor=flashsystem"
```

#### Remove Storage Pairs

```bash
# Remove storage pairs from mapping
kubectl mtv patch mapping storage prod-storage-mapping \
  --remove-pairs "old-datastore,deprecated-storage"
```

#### Advanced Storage Patching with Offloading

```bash
# Add storage with complete offloading configuration
kubectl mtv patch mapping storage enterprise-mapping \
  --add-pairs "tier1-datastore:flashsystem-gold;volumeMode=Block;accessMode=ReadWriteOnce;offloadPlugin=vsphere;offloadVendor=flashsystem;offloadSecret=flashsystem-secret" \
  --default-volume-mode Block \
  --default-offload-plugin vsphere
```

## Complete Mapping Workflow Examples

### Example 1: Enterprise Production Setup

```bash
# Step 1: Create comprehensive network mapping
kubectl mtv create mapping network enterprise-network \
  --source vsphere-production \
  --target openshift-production \
  --network-pairs "Production VLAN:prod-network,Management Network:multus-system/mgmt-network,DMZ VLAN:security-namespace/dmz-net,Backup Network:ignored"

# Step 2: Create enterprise storage mapping with offloading
kubectl mtv create mapping storage enterprise-storage \
  --source vsphere-production \
  --target openshift-production \
  --storage-pairs "Tier1-SSD:flashsystem-premium;volumeMode=Block;accessMode=ReadWriteOnce;offloadPlugin=vsphere;offloadVendor=flashsystem,Tier2-SATA:standard-storage;volumeMode=Filesystem;accessMode=ReadWriteOnce,Shared-NFS:shared-nfs;volumeMode=Filesystem;accessMode=ReadWriteMany" \
  --offload-vsphere-username svc-migration@vsphere.local \
  --offload-vsphere-password $(cat /secure/vcenter-password) \
  --offload-vsphere-url https://vcenter.prod.company.com \
  --offload-storage-username flashsystem-admin \
  --offload-storage-password $(cat /secure/flashsystem-password) \
  --offload-storage-endpoint https://flashsystem.company.com:7443

# Step 3: Verify mappings
kubectl mtv describe mapping network enterprise-network
kubectl mtv describe mapping storage enterprise-storage
```

### Example 2: Development Environment

```bash
# Development network mapping (simple)
kubectl mtv create mapping network dev-simple-network \
  --source vsphere-dev \
  --target openshift-dev \
  --network-pairs "Dev VLAN:default,Test Network:default,Management:ignored" \
  -n development

# Development storage mapping (basic)
kubectl mtv create mapping storage dev-simple-storage \
  --source vsphere-dev \
  --target openshift-dev \
  --storage-pairs "dev-datastore:standard-ssd,test-datastore:test-storage" \
  --default-volume-mode Filesystem \
  --default-access-mode ReadWriteOnce \
  -n development
```

### Example 3: Multi-Provider Setup

```bash
# oVirt to OpenShift mapping
kubectl mtv create mapping network ovirt-to-ocp \
  --source ovirt-production \
  --target openshift-target \
  --network-pairs "ovirtmgmt:default,production:prod-net,dmz:dmz-network"

kubectl mtv create mapping storage ovirt-storage \
  --source ovirt-production \
  --target openshift-target \
  --storage-pairs "data:standard-rwo,shared:shared-rwx,fast:premium-ssd" \
  --default-volume-mode Filesystem

# OpenStack to OpenShift mapping
kubectl mtv create mapping network openstack-to-ocp \
  --source openstack-prod \
  --target openshift-target \
  --network-pairs "internal:default,external:multus-system/external-net,provider:provider-network"

kubectl mtv create mapping storage openstack-storage \
  --source openstack-prod \
  --target openshift-target \
  --storage-pairs "__DEFAULT__:standard-rwo,ssd:premium-ssd,ceph:ceph-rbd" \
  --default-volume-mode Block
```

## Advanced Mapping Scenarios

### Network Mapping Patterns

#### Multus Network Integration

```bash
# Complex Multus network mapping with multiple namespaces
kubectl mtv create mapping network multus-complex \
  --source vsphere-prod \
  --target openshift-target \
  --network-pairs "Frontend-VLAN:web-namespace/frontend-net,Database-VLAN:db-namespace/database-net,Management:kube-system/management,Monitoring:monitoring/prometheus-net,Storage-Network:rook-ceph/cluster-network"
```

#### Security Zone Mapping

```bash
# Security zone-based network mapping
kubectl mtv create mapping network security-zones \
  --source vsphere-security \
  --target openshift-security \
  --network-pairs "Public-DMZ:dmz-namespace/public-dmz,Internal-DMZ:dmz-namespace/internal-dmz,Secure-Zone:secure-namespace/secure-net,Management-Zone:mgmt-namespace/mgmt-secure,Guest-Network:ignored"
```

#### VLAN-to-Namespace Mapping

```bash
# Map VLANs to specific namespaces
kubectl mtv create mapping network vlan-namespace \
  --source vsphere-prod \
  --target openshift-target \
  --network-pairs "VLAN-100:production/prod-net,VLAN-200:development/dev-net,VLAN-300:testing/test-net,VLAN-400:staging/stage-net"
```

### Storage Mapping Patterns

#### Performance Tier Mapping

```bash
# Map storage tiers to appropriate StorageClasses
kubectl mtv create mapping storage performance-tiers \
  --source vsphere-prod \
  --target openshift-target \
  --storage-pairs "NVMe-Tier1:ultra-fast-nvme;volumeMode=Block;accessMode=ReadWriteOnce,SSD-Tier2:fast-ssd;volumeMode=Block;accessMode=ReadWriteOnce,SATA-Tier3:standard-storage;volumeMode=Filesystem;accessMode=ReadWriteOnce,NFS-Shared:shared-nfs;volumeMode=Filesystem;accessMode=ReadWriteMany" \
  --default-volume-mode Block
```

#### Application-Specific Storage

```bash
# Application-specific storage class mapping
kubectl mtv create mapping storage app-specific \
  --source vsphere-prod \
  --target openshift-target \
  --storage-pairs "Database-Storage:database-optimized;volumeMode=Block;accessMode=ReadWriteOnce,Web-Storage:web-optimized;volumeMode=Filesystem;accessMode=ReadWriteOnce,Log-Storage:log-aggregation;volumeMode=Filesystem;accessMode=ReadWriteMany,Backup-Storage:backup-tier;volumeMode=Filesystem;accessMode=ReadOnlyMany"
```

#### Multi-Vendor Storage Array Integration

```bash
# Multiple storage vendor integration
kubectl mtv create mapping storage multi-vendor \
  --source vsphere-enterprise \
  --target openshift-target \
  --storage-pairs "FlashSystem-Gold:flashsystem-tier1;offloadPlugin=vsphere;offloadVendor=flashsystem;volumeMode=Block,NetApp-Silver:ontap-tier2;offloadPlugin=vsphere;offloadVendor=ontap;volumeMode=Filesystem,Pure-Platinum:pure-tier0;offloadPlugin=vsphere;offloadVendor=pureFlashArray;volumeMode=Block,Dell-Bronze:powermax-tier3;offloadPlugin=vsphere;offloadVendor=powermax;volumeMode=Block" \
  --default-volume-mode Block \
  --default-access-mode ReadWriteOnce
```

## Mapping Validation and Testing

### Validate Network Mappings

```bash
# Check network availability in target
kubectl get networks -A | grep -E "(prod-net|mgmt-net|dmz-network)"

# Verify Multus network definitions
kubectl get network-attachment-definitions -A

# Test network connectivity
kubectl mtv get inventory networks openshift-target
```

### Validate Storage Mappings

```bash
# Check StorageClass availability
kubectl get storageclass

# Verify storage array integration
kubectl get csidriver

# Test storage provisioning
kubectl apply -f - << EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: premium-ssd
EOF
```

### Mapping Pre-Migration Validation

```bash
#!/bin/bash
# Pre-migration mapping validation script

NETWORK_MAPPING="enterprise-network"
STORAGE_MAPPING="enterprise-storage"
NAMESPACE="konveyor-forklift"

echo "Validating network mapping..."
kubectl describe networkmapping "$NETWORK_MAPPING" -n "$NAMESPACE"

echo "Validating storage mapping..."
kubectl describe storagemapping "$STORAGE_MAPPING" -n "$NAMESPACE"

echo "Checking target resources..."
kubectl get storageclass
kubectl get network-attachment-definitions -A

echo "Validation complete"
```

## Troubleshooting Mapping Issues

### Common Mapping Problems

#### Network Mapping Issues

```bash
# Check if target networks exist
kubectl get network-attachment-definitions -A | grep target-network-name

# Verify Multus installation
kubectl get pods -n kube-system | grep multus

# Check network mapping status
kubectl describe networkmapping mapping-name -n konveyor-forklift
```

#### Storage Mapping Issues

```bash
# Check StorageClass availability
kubectl get storageclass storage-class-name

# Verify CSI driver installation
kubectl get csidriver

# Check storage mapping status
kubectl describe storagemapping mapping-name -n konveyor-forklift

# Test storage provisioning
kubectl get pv,pvc -A | grep storage-class-name
```

### Mapping Syntax Validation

```bash
# Test network pairs syntax
echo "VM Network:default,Management:mgmt-net" | tr ',' '\n'

# Test storage pairs syntax
echo "ds1:sc1;volumeMode=Block,ds2:sc2;accessMode=ReadWriteMany" | tr ',' '\n'
```

### Debug Mapping Creation

```bash
# Enable verbose logging
kubectl mtv create mapping network test-mapping \
  --source vsphere-prod \
  --target openshift-target \
  --network-pairs "VM Network:default" \
  -v=2

# Check mapping resource creation
kubectl get networkmapping,storagemapping -A
```

## Integration with Migration Plans

Mappings integrate seamlessly with migration plan creation:

```bash
# Create migration plan using mappings
kubectl mtv create plan enterprise-migration \
  --source vsphere-production \
  --target openshift-production \
  --network-mapping enterprise-network \
  --storage-mapping enterprise-storage \
  --vms "web-server-01,db-server-01,app-server-01"

# Alternative: Use inline mapping pairs
kubectl mtv create plan inline-migration \
  --source vsphere-prod \
  --network-pairs "VM Network:default,Management:mgmt-net" \
  --storage-pairs "datastore1:standard-ssd,datastore2:premium-nvme" \
  --vms "test-vm-01,test-vm-02"
```

## Best Practices for Mapping Management

### Network Mapping Best Practices

1. **Use Descriptive Names**: Create mappings with clear, descriptive names
2. **Plan Network Isolation**: Map security zones to appropriate target networks
3. **Consider Performance**: Route high-bandwidth workloads to appropriate networks
4. **Document Dependencies**: Maintain mapping documentation for operations teams

### Storage Mapping Best Practices

1. **Match Performance Requirements**: Map high-IOPS workloads to fast storage classes
2. **Consider Access Patterns**: Use appropriate volume modes and access modes
3. **Leverage Storage Array Features**: Use offloading for compatible storage systems (see [Chapter 9.5: Storage Array Offloading](09.5-storage-array-offloading-and-optimization.md) for detailed information)
4. **Plan for Scale**: Consider storage capacity and performance implications

### Operational Best Practices

1. **Test Mappings**: Validate mappings before production migrations
2. **Version Control**: Maintain mapping configurations in version control
3. **Monitor Resources**: Ensure target resources have sufficient capacity
4. **Update Gradually**: Use patching for incremental mapping updates

## Next Steps

After mastering mapping management:

1. **Plan Creation**: Use mappings in [Chapter 10: Migration Plan Creation](10-migration-plan-creation.md)
2. **VM Customization**: Apply mappings to specific VMs in [Chapter 11: Customizing Individual VMs](11-customizing-individual-vms-planvms-format.md)
3. **Optimization**: Leverage mapping insights in [Chapter 13: Migration Process Optimization](13-migration-process-optimization.md)
4. **Advanced Patching**: Learn plan patching in [Chapter 15: Advanced Plan Patching](15-advanced-plan-patching.md)

---

*Previous: [Chapter 8: Query Language Reference and Advanced Filtering](08-query-language-reference-and-advanced-filtering.md)*
*Next: [Chapter 9.5: Storage Array Offloading and Optimization](09.5-storage-array-offloading-and-optimization.md)*
