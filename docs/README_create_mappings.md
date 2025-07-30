# Creating Network and Storage Mappings

This guide explains how to create network and storage mappings using the `kubectl-mtv create mapping` command. Mappings define how resources from source providers are mapped to resources in target providers during VM migration.

## Overview

Mappings are reusable resources that can be referenced by multiple migration plans. They provide a consistent way to define how:
- **Network mappings**: Map source networks (VLANs, port groups, etc.) to target networks (NetworkAttachmentDefinitions, pod networks)
- **Storage mappings**: Map source storage (datastores, storage domains) to target storage classes

## Network Mappings

### Basic Syntax

```bash
kubectl-mtv create mapping network <mapping-name> \
  --source <source-provider> \
  --target <target-provider> \
  --network-pairs "<source>:<target>,<source>:<target>,..."
```

### Network Pair Format

Each network pair follows the format: `source-network:target-specification`

#### Target Specifications

1. **Multus Network**: `namespace/network-name`
   ```bash
   --network-pairs "VM Network:production/prod-net"
   ```

2. **Default Network (pod)**: `default`
   ```bash
   --network-pairs "Management Network:default"
   ```

3. **Ignored Network**: `ignored`
   ```bash
   --network-pairs "Backup Network:ignored"
   ```

4. **Default Namespace**: `network-name` (uses mapping's namespace)
   ```bash
   --network-pairs "Internal Network:internal-net"
   ```

### Network Mapping Examples

#### Simple Network Mapping
```bash
kubectl-mtv create mapping network basic-networks \
  --source vsphere-provider \
  --target openshift \
  --network-pairs "VM Network:default"
```

#### Multi-Network Mapping
```bash
kubectl-mtv create mapping network complex-networks \
  --source vsphere-provider \
  --target openshift \
  --network-pairs "Production VLAN:prod/production-net,Development VLAN:dev/development-net,Management Network:default,DMZ Network:security/dmz-net"
```

#### Network Mapping with Special Cases
```bash
kubectl-mtv create mapping network advanced-networks \
  --source vsphere-datacenter \
  --target kubernetes-cluster \
  --network-pairs "Frontend Network:frontend/web-net,Backend Network:backend/app-net,Database Network:database/db-net,Backup Network:ignored,Management:default"
```

## Storage Mappings

### Basic Syntax

```bash
kubectl-mtv create mapping storage <mapping-name> \
  --source <source-provider> \
  --target <target-provider> \
  --storage-pairs "<source>:<target>,<source>:<target>,..."
```

### Storage Pair Format

Each storage pair follows the format: `source-storage:storage-class`

- **source-storage**: The exact name of the source datastore/storage domain
- **storage-class**: The target Kubernetes StorageClass (cluster-scoped)

### Storage Mapping Examples

#### Simple Storage Mapping
```bash
kubectl-mtv create mapping storage basic-storage \
  --source vsphere-provider \
  --target openshift \
  --storage-pairs "datastore1:standard"
```

#### Performance-Tiered Storage
```bash
kubectl-mtv create mapping storage tiered-storage \
  --source vsphere-provider \
  --target openshift \
  --storage-pairs "SSD-Datastore:fast-ssd,SATA-Datastore:standard,Archive-Datastore:slow-archive"
```

#### Complex Storage Mapping
```bash
kubectl-mtv create mapping storage enterprise-storage \
  --source vsphere-datacenter \
  --target kubernetes-cluster \
  --storage-pairs "VMFS-SSD-01:ultra-fast-nvme,VMFS-SSD-02:fast-ssd,VMFS-HDD-01:standard,VMFS-HDD-02:standard,NFS-Archive:cold-storage"
```

## Provider-Specific Examples

### VMware vSphere to OpenShift

```bash
# Network mapping for vSphere
kubectl-mtv create mapping network vsphere-to-ocp-networks \
  --source vsphere-vcenter \
  --target openshift \
  --network-pairs "VM Network:openshift-sdn/vm-network,vMotion:ignored,Management Network:default,DMZ VLAN 100:dmz/dmz-network"

# Storage mapping for vSphere
kubectl-mtv create mapping storage vsphere-to-ocp-storage \
  --source vsphere-vcenter \
  --target openshift \
  --storage-pairs "vsanDatastore:ocs-storagecluster-ceph-rbd,FC-Datastore-01:fast-fc,iSCSI-Datastore:standard"
```

### oVirt/RHV to OpenShift

```bash
# Network mapping for oVirt
kubectl-mtv create mapping network ovirt-to-ocp-networks \
  --source ovirt-engine \
  --target openshift \
  --network-pairs "ovirtmgmt:default,vm_network:vms/vm-network,display:ignored"

# Storage mapping for oVirt
kubectl-mtv create mapping storage ovirt-to-ocp-storage \
  --source ovirt-engine \
  --target openshift \
  --storage-pairs "data_domain:standard,fast_domain:fast-ssd,export_domain:ignored"
```

### OpenStack to OpenShift

```bash
# Network mapping for OpenStack
kubectl-mtv create mapping network openstack-to-ocp-networks \
  --source openstack-controller \
  --target openshift \
  --network-pairs "public:openshift-sdn/public-net,private:openshift-sdn/private-net,management:default"

# Storage mapping for OpenStack
kubectl-mtv create mapping storage openstack-to-ocp-storage \
  --source openstack-controller \
  --target openshift \
  --storage-pairs "cinder-ssd:fast-ssd,cinder-standard:standard,cinder-archive:slow-archive"
```

## Advanced Usage

### Using Environment Variables

```bash
# Set inventory URL if not auto-discovered
export MTV_INVENTORY_URL="https://inventory.example.com"

kubectl-mtv create mapping network prod-networks \
  --source vsphere \
  --target openshift \
  --network-pairs "Production:prod/prod-net"
```

### Namespace-Specific Mappings

```bash
# Create mapping in specific namespace
kubectl-mtv create mapping network namespace-networks \
  --namespace migration-project \
  --source vsphere \
  --target openshift \
  --network-pairs "VM Network:migration-project/project-net"
```

## Listing and Managing Mappings

### List Mappings

```bash
# List all network mappings
kubectl-mtv get mapping network

# List all storage mappings
kubectl-mtv get mapping storage

# List mappings in specific namespace
kubectl-mtv get mapping network -n migration-project
```

### View Mapping Details

```bash
# Get detailed mapping information
kubectl-mtv get mapping network vsphere-networks -o yaml

# Get mapping in JSON format
kubectl-mtv get mapping storage vsphere-storage -o json
```

### Delete Mappings

```bash
# Delete network mapping
kubectl-mtv delete mapping network old-network-map

# Delete storage mapping
kubectl-mtv delete mapping storage obsolete-storage-map
```

## Best Practices

1. **Naming Conventions**: Use descriptive names that indicate source and target
   - Good: `vsphere-to-ocp-prod-networks`
   - Bad: `network-map-1`

2. **Document Mappings**: Keep a record of what each mapping does
   ```bash
   # Production networks for critical workloads
   kubectl-mtv create mapping network prod-critical-networks \
     --source vsphere-prod \
     --target ocp-prod \
     --network-pairs "VLAN-100-Web:web/web-tier,VLAN-200-App:app/app-tier,VLAN-300-DB:database/db-tier"
   ```

3. **Test First**: Create test mappings before production
   ```bash
   # Test mapping
   kubectl-mtv create mapping network test-networks \
     --source vsphere-dev \
     --target ocp-dev \
     --network-pairs "Test-Network:default"
   ```

4. **Group Related Resources**: Create consistent mapping sets
   ```bash
   # Matching network and storage mappings for same environment
   kubectl-mtv create mapping network prod-networks ...
   kubectl-mtv create mapping storage prod-storage ...
   ```

## Troubleshooting

### Common Issues

1. **Source Resource Not Found**
   ```bash
   Error: failed to resolve source network 'VM-Network': not found
   # Fix: Check exact network name (case-sensitive)
   kubectl-mtv get inventory networks vsphere-provider
   ```

2. **Target Namespace Missing**
   ```bash
   Error: namespace "production" not found
   # Fix: Create namespace first
   kubectl create namespace production
   ```

3. **Invalid Storage Class**
   ```bash
   Error: storage class "fast-storage" not found
   # Fix: List available storage classes
   kubectl get storageclass
   ```

### Validation Tips

Before creating mappings:

1. **Verify Source Resources**
   ```bash
   # List source networks
   kubectl-mtv get inventory networks <source-provider>
   
   # List source storage
   kubectl-mtv get inventory storage <source-provider>
   ```

2. **Verify Target Resources**
   ```bash
   # List NetworkAttachmentDefinitions
   kubectl get network-attachment-definitions -A
   
   # List StorageClasses
   kubectl get storageclass
   ```

3. **Test with Simple Mapping First**
   ```bash
   # Start with basic mapping
   kubectl-mtv create mapping network test \
     --source vsphere \
     --target openshift \
     --network-pairs "VM Network:default"
   ``` 