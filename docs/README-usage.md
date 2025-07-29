# Usage Guide

This comprehensive guide covers all kubectl-mtv commands and usage patterns for managing VM migrations with Forklift/MTV.

## Table of Contents

- [Global Flags](#global-flags)
- [Provider Management](#provider-management)
- [Host Management](#host-management)
- [Inventory Management](#inventory-management)
- [Mapping Management](#mapping-management)
- [Migration Plan Management](#migration-plan-management)
- [Plan Lifecycle Commands](#plan-lifecycle-commands)
- [VDDK Image Management](#vddk-image-management)
- [Query Language](#query-language)
- [Output Formats](#output-formats)
- [Common Workflows](#common-workflows)

## Global Flags

kubectl-mtv supports standard kubectl flags plus additional MTV-specific options:

```bash
# Standard kubectl flags
kubectl mtv <command> --namespace <namespace>
kubectl mtv <command> --kubeconfig <config-file>
kubectl mtv <command> -v <verbosity-level>

# Output format options
kubectl mtv <command> -o table|json|yaml
kubectl mtv <command> --output table|json|yaml
```

### Common Global Flags

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--namespace` | `-n` | Kubernetes namespace | `-n migration-ns` |
| `--output` | `-o` | Output format | `-o json` |
| `--kubeconfig` | | Path to kubeconfig file | `--kubeconfig ~/.kube/config` |
| `--context` | | Kubernetes context to use | `--context dev-cluster` |

## Provider Management

Providers represent source and target virtualization platforms.

### List Providers

```bash
# List all providers
kubectl mtv get providers

# List providers in specific namespace
kubectl mtv get providers -n forklift-operator

# Get detailed provider information
kubectl mtv get providers -o yaml
```

### Create Providers

#### OpenShift/Kubernetes (Target) Provider

```bash
# Create OpenShift target provider
kubectl mtv create provider host --type openshift

# Create with custom name
kubectl mtv create provider my-openshift --type openshift
```

#### VMware vSphere Provider

```bash
# Basic VMware provider
kubectl mtv create provider vmware --type vsphere \
  -U https://vcenter.example.com/sdk \
  -u administrator@vsphere.local \
  -p "your-password"

# VMware with VDDK and TLS options
kubectl mtv create provider vmware --type vsphere \
  -U https://vcenter.example.com/sdk \
  -u administrator@vsphere.local \
  -p "your-password" \
  --vddk-init-image quay.io/your-org/vddk:8.0.1 \
  --provider-insecure-skip-tls

# VMware with CA certificate from file
kubectl mtv create provider vmware --type vsphere \
  -U https://vcenter.example.com/sdk \
  -u administrator@vsphere.local \
  -p "your-password" \
  --cacert @/path/to/vcenter-ca.crt
```

#### oVirt/RHV Provider

```bash
# Create oVirt provider
kubectl mtv create provider ovirt --type ovirt \
  -U https://ovirt-engine.example.com/ovirt-engine/api \
  -u admin@internal \
  -p "your-password" \
  --cacert @/path/to/ca.crt

# Or provide certificate content directly
kubectl mtv create provider ovirt --type ovirt \
  -U https://ovirt-engine.example.com/ovirt-engine/api \
  -u admin@internal \
  -p "your-password" \
  --cacert "-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJANcT3I7d4I6BA...
-----END CERTIFICATE-----"
```

#### OpenStack Provider

```bash
# Create OpenStack provider
kubectl mtv create provider openstack --type openstack \
  -U https://openstack.example.com:5000/v3 \
  -u admin \
  -p "your-password" \
  --provider-domain-name default \
  --provider-project-name admin \
  --provider-region-name regionOne
```

#### OVA Provider

```bash
# Create OVA provider
kubectl mtv create provider ova --type ova \
  -U nfs://nfs.example.com/ova-files

# OVA provider with CA certificate (for secure NFS/HTTPS)
kubectl mtv create provider ova --type ova \
  -U https://fileserver.example.com/ova-files \
  --cacert @/path/to/server-ca.crt
```

### Provider Configuration Options

| Provider Type | Required Flags | Optional Flags |
|---------------|----------------|----------------|
| `openshift` | `--type` | `--name` |
| `vsphere` | `--type`, `-U`, `-u`, `-p` | `--vddk-init-image`, `--provider-insecure-skip-tls`, `--cacert` |
| `ovirt` | `--type`, `-U`, `-u`, `-p` | `--cacert`, `--provider-insecure-skip-tls` |
| `openstack` | `--type`, `-U`, `-u`, `-p` | `--provider-domain-name`, `--provider-project-name`, `--provider-region-name` |
| `ova` | `--type`, `-U` | `--cacert` |

**Note**: The `--cacert` flag accepts either certificate content directly or use `@filename` to load from file.

### Patch Providers

Modify existing providers by updating URL, credentials, or VDDK settings.

```bash
# Update provider URL
kubectl mtv patch provider vsphere-provider \
  --url https://new-vcenter.example.com

# Update credentials (only if secret is owned by provider)
kubectl mtv patch provider vsphere-provider \
  --username administrator@vsphere.local \
  --password newpassword

# Update OpenShift token
kubectl mtv patch provider openshift-provider \
  --token sha256~new-token-here

# Update VDDK settings for vSphere
kubectl mtv patch provider vsphere-provider \
  --vddk-init-image registry.example.com/vddk:v8.0.2 \
  --use-vddk-aio-optimization=true

# Update OpenStack credentials
kubectl mtv patch provider openstack-provider \
  --username newuser \
  --provider-domain-name new-domain \
  --provider-project-name new-project

# Combined updates
kubectl mtv patch provider vsphere-provider \
  --url https://new-vcenter.example.com \
  --username newadmin@vsphere.local \
  --vddk-init-image registry.example.com/vddk:latest
```

**Important**: Provider type and SDK endpoint cannot be changed after creation. Credential updates are only allowed if the provider owns its secret (created automatically with the provider).

**Note**: See [Patch Providers Guide](README_patch_providers.md) for detailed usage and security considerations.

### Patch Plans

Modify existing migration plans by updating plan-level settings or individual VM configurations.

#### Patch Plan Settings

Update migration plan configuration without changing providers, mappings, or VMs.

```bash
# Update migration type and target settings
kubectl mtv patch plan my-migration-plan \
  --migration-type warm \
  --target-namespace production \
  --use-compatibility-mode=true

# Configure transfer network
kubectl mtv patch plan my-plan \
  --transfer-network openshift-sriov-network/fast-network

# Set target VM labels and node selector
kubectl mtv patch plan production-migration \
  --target-labels "env=production,tier=web" \
  --target-node-selector "node-type=compute,disk=ssd"

# Configure advanced affinity with KARL syntax
kubectl mtv patch plan my-plan \
  --target-affinity 'REQUIRE pods(app=database) on node'

# Set plan templates and description
kubectl mtv patch plan production-plan \
  --description "Production migration batch 1" \
  --pvc-name-template "prod-{{.Name}}-{{.DiskName}}" \
  --preserve-static-ips=true
```

#### Patch Individual VMs

Modify specific VM configurations within a plan's VM list.

```bash
# Update VM target name and instance type
kubectl mtv patch plan-vms production-plan web-server-vm \
  --target-name production-web-01 \
  --instance-type m5.xlarge

# Configure naming templates
kubectl mtv patch plan-vms my-plan app-server \
  --pvc-name-template "{{.Name}}-storage-{{.DiskName}}" \
  --volume-name-template "{{.Name}}-vol-{{.Index}}"

# Set encryption configuration
kubectl mtv patch plan-vms secure-plan encrypted-vm \
  --luks-secret disk-encryption-keys \
  --root-disk "hard-disk-1"

# Combined VM updates
kubectl mtv patch plan-vms enterprise-migration critical-app \
  --target-name prod-critical-app-01 \
  --instance-type c5.2xlarge \
  --pvc-name-template "prod-{{.Name}}-{{.DiskName}}" \
  --luks-secret app-encryption-keys

# Manage VM hooks
kubectl mtv patch plan-vms production-plan database-vm \
  --add-pre-hook backup-database-hook \
  --add-post-hook verify-migration-hook

# Remove hooks from VM
kubectl mtv patch plan-vms test-plan test-vm \
  --remove-hook old-hook \
  --clear-hooks
```

**Protected Fields**: Source/target providers, network/storage mappings, and the VM list itself cannot be changed through patch plan. Individual VM modifications use patch plan-vms.

**Note**: See [Patch Plans Guide](README_patch_plans.md) for comprehensive usage examples, template variables, and best practices.

### Delete Providers

```bash
# Delete a specific provider
kubectl mtv delete provider <provider-name>

# Delete with confirmation
kubectl mtv delete provider vmware-prod
```

## Host Management

Manage host resources for migration providers.

### List Hosts

```bash
# List all host resources
kubectl mtv get host

# Get details for a specific host
kubectl mtv get host my-esxi-host

# Get host details in yaml format
kubectl mtv get host my-esxi-host -o yaml
```

### Create Hosts

```bash
# Create a single host
kubectl mtv create host esxi-host-01 \
  --provider vsphere-datacenter \
  --ip-address 192.168.1.100 \
  --username root --password secret

# Create multiple hosts
kubectl mtv create host esxi-host-01 esxi-host-02 \
  --provider vsphere-datacenter \
  --network-adapter "Management Network" \
  --username root --password secret
```

### Describe Hosts

```bash
# Get detailed host information
kubectl mtv describe host esxi-host-01
```

### Delete Hosts

```bash
# Delete a specific host
kubectl mtv delete host <host-name>

# Delete multiple hosts  
kubectl mtv delete host host1 host2 host3

# Examples
kubectl mtv delete host esxi-host-01
kubectl mtv delete host esxi-host-01 esxi-host-02 esxi-host-03
```

## Inventory Management

Browse and query resources from source providers.

### List Virtual Machines

```bash
# List all VMs from a provider
kubectl mtv get inventory vms <provider-name>

# List VMs with query filter
kubectl mtv get inventory vms vmware -q "where name ~= 'web-.*'"

# Get VMs with more than 2 disks
kubectl mtv get inventory vms vmware -q "where len disks > 2"

# List VMs with specific power state
kubectl mtv get inventory vms vmware -q "where powerState == 'poweredOn'"
```

### List Other Inventory Resources

```bash
# List hosts
kubectl mtv get inventory hosts <provider-name>

# List networks
kubectl mtv get inventory networks <provider-name>

# List storage domains/datastores
kubectl mtv get inventory storage <provider-name>

# List namespaces (for OpenShift/Kubernetes providers)
kubectl mtv get inventory namespaces <provider-name>
```

### Inventory Query Examples

```bash
# VMs with specific OS
kubectl mtv get inventory vms vmware -q "where guestOS ~= '.*linux.*'"

# VMs with memory > 4GB
kubectl mtv get inventory vms vmware -q "where memoryMB > 4096"

# VMs in specific folder
kubectl mtv get inventory vms vmware -q "where folder == 'Production VMs'"

# Complex query with multiple conditions
kubectl mtv get inventory vms vmware -q "where name ~= 'prod-.*' and powerState == 'poweredOn' and len disks > 1"
```

## Mapping Management

Mappings define how source resources map to target resources.

### List Mappings

```bash
# List specific mapping types
kubectl mtv get mapping network
kubectl mtv get mapping storage
```

### Create Network Mappings

```bash
# Basic network mapping (will use default mappings)
kubectl mtv create mapping network net-map \
  --source-provider vmware \
  --target-provider host

# Network mapping with specific network pairs
kubectl mtv create mapping network production-nets \
  --source-provider vmware \
  --target-provider host \
  --network-pairs "VM Network:default/vm-network,DMZ:default/dmz-network"

# Complex network mapping with multiple network pairs
kubectl mtv create mapping network enterprise-networks \
  --source-provider vmware \
  --target-provider host \
  --network-pairs "Production VLAN:prod-namespace/production-net,Development VLAN:dev-namespace/dev-net,Management Network:mgmt-namespace/management-net"

# Network mapping for multi-site deployment
kubectl mtv create mapping network multi-site-networks \
  --source-provider vcenter-site1 \
  --target-provider openshift-site2 \
  --network-pairs "Site1-Production:site2-prod-namespace/production-network,Site1-DMZ:site2-dmz-namespace/dmz-network"
```

### Create Storage Mappings

```bash
# Basic storage mapping (will use default mappings)
kubectl mtv create mapping storage storage-map \
  --source-provider vmware \
  --target-provider host

# Storage mapping with specific storage pairs
kubectl mtv create mapping storage production-storage \
  --source-provider vmware \
  --target-provider host \
  --storage-pairs "datastore1:default/fast-ssd,datastore2:default/standard-hdd"

# Performance-tiered storage mapping
kubectl mtv create mapping storage tiered-storage \
  --source-provider vmware \
  --target-provider host \
  --storage-pairs "SSD-Datastore:production/premium-ssd,SATA-Datastore:production/standard-hdd,NVMe-Datastore:production/ultra-ssd"

# Multi-cluster storage mapping with specific storage classes
kubectl mtv create mapping storage distributed-storage \
  --source-provider ovirt-cluster \
  --target-provider kubernetes-cluster \
  --storage-pairs "ovirt-data:default/ceph-rbd,ovirt-fast:default/local-nvme,ovirt-backup:backup-ns/slow-storage"

# Storage mapping for different workload types
kubectl mtv create mapping storage workload-storage \
  --source-provider vmware \
  --target-provider host \
  --storage-pairs "Database-Storage:db-namespace/high-iops-ssd,Application-Storage:app-namespace/balanced-ssd,Archive-Storage:archive-namespace/cold-storage"
```

#### Network Mapping Pairs Format

Network pairs use the format: `"source-network:target-namespace/target-network"`

Examples:

- `"VM Network:default/pod-network"` - Maps VMware "VM Network" to "pod-network" in default namespace
- `"Production VLAN:prod/production-net"` - Maps to specific namespace and network
- `"DMZ-100:security/dmz-network"` - Maps VLAN to security namespace

#### Storage Mapping Pairs Format

Storage pairs use the format: `"source-storage:target-namespace/target-storage-class"`

Examples:

- `"datastore1:default/fast-ssd"` - Maps VMware datastore to Kubernetes storage class
- `"tier1-storage:production/premium-ssd"` - Maps to specific namespace and storage class
- `"shared-nfs:shared/nfs-storage"` - Maps shared storage to NFS storage class

### Patch Mappings

Modify existing mappings by adding, updating, or removing pairs.

#### Patch Network Mappings

```bash
# Add new network pairs to existing mapping
kubectl mtv patch mapping network production-networks \
  --add-pairs "VM Network:openshift-sdn/production,Management:pod"

# Update existing network pairs
kubectl mtv patch mapping network production-networks \
  --update-pairs "VM Network:openshift-sdn/staging,DMZ:security/dmz-network"

# Remove network pairs by source name
kubectl mtv patch mapping network production-networks \
  --remove-pairs "Old-Network,Unused-VLAN"

# Combined operations
kubectl mtv patch mapping network production-networks \
  --add-pairs "New-Network:production/new-net" \
  --update-pairs "VM Network:production/updated-net" \
  --remove-pairs "Deprecated-Network"
```

#### Patch Storage Mappings

```bash
# Add new storage pairs to existing mapping
kubectl mtv patch mapping storage production-storage \
  --add-pairs "VMFS-Datastore-1:standard,VMFS-Fast-Storage:fast-ssd"

# Update existing storage pairs
kubectl mtv patch mapping storage production-storage \
  --update-pairs "VMFS-Datastore-1:premium-ssd,Archive-Storage:slow-archive"

# Remove storage pairs by source name
kubectl mtv patch mapping storage production-storage \
  --remove-pairs "Old-Datastore,Decommissioned-Storage"

# Combined operations
kubectl mtv patch mapping storage production-storage \
  --add-pairs "New-Datastore:standard" \
  --update-pairs "VMFS-Fast-Storage:premium-nvme" \
  --remove-pairs "Legacy-Storage"
```

**Note**: The patch command automatically prevents duplicate sources when adding pairs and provides warnings for any conflicts. See [Patch Mappings Guide](README_patch_mappings.md) for detailed usage.

### Delete Mappings

```bash
# Delete network mapping
kubectl mtv delete mapping network <mapping-name>

# Delete storage mapping
kubectl mtv delete mapping storage <mapping-name>
```

## Migration Plan Management

Migration plans define which VMs to migrate and how.

### List Plans

```bash
# List all migration plans
kubectl mtv get plans

# Get detailed plan information
kubectl mtv get plan <plan-name> -o yaml

# List VMs in a plan
kubectl mtv get plan-vms <plan-name>
```

### Create Migration Plans

#### Basic Plan Creation

```bash
# Create a simple migration plan
kubectl mtv create plan my-migration \
  --source-provider vmware \
  --vms vm1,vm2,vm3

# Create plan with target namespace
kubectl mtv create plan web-servers \
  --source-provider vmware \
  --target-namespace production \
  --vms web-01,web-02,web-03
```

#### Advanced Plan Options

```bash
# Create warm migration plan
kubectl mtv create plan warm-migration \
  --source-provider vmware \
  --vms critical-vm1,critical-vm2 \
  --warm

# Plan with custom PVC naming
kubectl mtv create plan custom-storage \
  --source-provider vmware \
  --vms database-vm \
  --pvc-name-template "{{.VmName}}-disk-{{.DiskIndex}}" \
  --pvc-name-template-use-generate-name=false

# Plan with cleanup options
kubectl mtv create plan ephemeral-migration \
  --source-provider vmware \
  --vms test-vm1,test-vm2 \
  --delete-guest-conversion-pod
```

#### Plan with Mappings

```bash
# Create plan with specific mappings
kubectl mtv create plan mapped-migration \
  --source-provider vmware \
  --vms app-server1,app-server2 \
  --network-map production-nets \
  --storage-map production-storage
```

### Plan Configuration Options

| Flag | Description | Example |
|------|-------------|---------|
| `--source`, `-S` | Source provider name | `--source vmware` |
| `--target`, `-t` | Target provider name | `--target openshift` |
| `--vms` | Comma-separated list of VMs or path to a file | `--vms vm1,vm2` or `--vms @vms.json` |
| `--network-mapping` | Network mapping name | `--network-mapping net-map` |
| `--storage-mapping` | Storage mapping name | `--storage-mapping storage-map` |
| `--network-pairs` | Inline network mapping pairs | `--network-pairs 'src-net:tgt-net'` |
| `--storage-pairs` | Inline storage mapping pairs | `--storage-pairs 'src-ds:tgt-sc'` |
| `--description` | Plan description | `--description "Migrate web servers"` |
| `--target-namespace` | Target Kubernetes namespace | `--target-namespace production` |
| `--transfer-network` | Network to use for disk transfer | `--transfer-network my-nad` |
| `--preserve-cluster-cpu-model` | Preserve oVirt cluster CPU model | `--preserve-cluster-cpu-model` |
| `--preserve-static-ips` | Preserve static IPs from vSphere | `--preserve-static-ips` |
| `--pvc-name-template` | Template for PVC names | `--pvc-name-template "{{.VmName}}-{{.DiskIndex}}"` |
| `--volume-name-template` | Template for volume names | `--volume-name-template "{{.VmName}}-vol"` |
| `--network-name-template` | Template for network interface names | `--network-name-template "{{.VmName}}-nic"` |
| `--migrate-shared-disks` | Migrate shared disks | `--migrate-shared-disks=false` |
| `--inventory-url`, `-i` | Inventory service URL | `-i http://inventory.svc` |
| `--archived` | Create the plan in an archived state | `--archived` |
| `--pvc-name-template-use-generate-name` | Use `generateName` for PVCs | `--pvc-name-template-use-generate-name=false` |
| `--delete-guest-conversion-pod` | Clean up conversion pods | `--delete-guest-conversion-pod` |
| `--skip-guest-conversion` | Skip guest OS conversion | `--skip-guest-conversion` |
| `--install-legacy-drivers` | Install legacy Windows drivers | `--install-legacy-drivers=true` |
| `--migration-type`, `-m` | Migration type: `cold`, `warm`, or `live` | `--migration-type warm` |
| `--warm` | Enable warm migration (deprecated) | `--warm` |
| `--default-target-network`, `-N` | Default target network for mapping | `-N pod` |
| `--default-target-storage-class` | Default target storage class | `--default-target-storage-class thin` |
| `--use-compatibility-mode` | Use compatibility devices for bootability | `--use-compatibility-mode=false` |
| `--target-labels`, `-L` | Labels to add to the migrated VM | `-L app=web,tier=frontend` |
| `--target-node-selector` | Node selector for the migrated VM | `--target-node-selector 'disktype=ssd'` |
| `--target-affinity` | Constrain VM scheduling using KARL syntax | `--target-affinity 'REQUIRE pods(app=db)'` |

For advanced scheduling, see the [Target Affinity Guide](README_target_affinity.md).

### Describe Plans

```bash
# Get detailed plan status
kubectl mtv describe plan <plan-name>

# Describe specific VM in plan
kubectl mtv describe plan-vm <plan-name> --vm <vm-name>
```

### Delete Plans

```bash
# Delete a migration plan
kubectl mtv delete plan <plan-name>
```

## Plan Lifecycle Commands

Control the execution of migration plans.

### Start Migration

```bash
# Start a migration plan
kubectl mtv start plan <plan-name>

# Start with confirmation
kubectl mtv start plan production-migration
```

### Cancel Migration workload

```bash
# Cancel specific VMs in a plan
kubectl mtv cancel plan <plan-name> --vms vm1,vm2
```

### Cutover Migration

For warm migrations, perform final cutover:

```bash
# Cutover entire plan immediately (current time)
kubectl mtv cutover plan <plan-name>

# Schedule cutover for a specific time (ISO8601 format)
kubectl mtv cutover plan <plan-name> --cutover "2024-01-15T20:30:00Z"

# Schedule cutover for a specific time using date command
kubectl mtv cutover plan <plan-name> --cutover "$(date --iso-8601=sec)"

# Schedule cutover for 2 hours from now
kubectl mtv cutover plan <plan-name> --cutover "$(date -d '+2 hours' --iso-8601=sec)"

# Example: Schedule cutover during maintenance window
kubectl mtv cutover plan critical-migration --cutover "2024-01-15T02:00:00Z"
```

### Archive Plans

```bash
# Archive completed plan
kubectl mtv archive plan <plan-name>

# Archive multiple plans
kubectl mtv archive plan plan1 plan2 plan3
```

### Unarchive Plans

```bash
# Unarchive a plan
kubectl mtv unarchive plan <plan-name>
```

## VDDK Image Management

For VMware migrations, manage VDDK (Virtual Disk Development Kit) images.

### Create VDDK Image

```bash
# Create VDDK image from URL
kubectl mtv create vddk my-vddk \
  --tag registry.example.com/vddk:8.0.1 \
  --tar vddk-801-file.tar.gz
```

## Query Language

kubectl-mtv includes a powerful query language for filtering inventory.

### Basic Query Syntax

```bash
# Field equality
kubectl mtv get inventory vms vmware -q "where name == 'web-server'"

# Pattern matching (regex)
kubectl mtv get inventory vms vmware -q "where name ~= 'web-.*'"

# Numeric comparisons
kubectl mtv get inventory vms vmware -q "where memoryMB > 4096"
kubectl mtv get inventory vms vmware -q "where len disks >= 2"

# Logical operators
kubectl mtv get inventory vms vmware -q "where powerState == 'poweredOn' and memoryMB > 2048"
kubectl mtv get inventory vms vmware -q "where name ~= 'test-.*' or name ~= 'dev-.*'"
```

### Advanced Query Features

```bash
# Array length functions
kubectl mtv get inventory vms vmware -q "where len disks > 1"
kubectl mtv get inventory vms vmware -q "where len nics == 2"

# String functions
kubectl mtv get inventory vms vmware -q "where tolower name ~= '.*production.*'"

# Nested field access
kubectl mtv get inventory vms vmware -q "where host.name == 'esx-01.example.com'"

# Complex conditions
kubectl mtv get inventory vms vmware -q "where (memoryMB > 8192 and len disks > 2) or name ~= 'critical-.*'"
```

### Common Query Patterns

```bash
# Production VMs with high resources
kubectl mtv get inventory vms vmware -q "where name ~= 'prod-.*' and memoryMB > 4096 and len disks > 1"

# Powered-on Linux VMs
kubectl mtv get inventory vms vmware -q "where powerState == 'poweredOn' and guestOS ~= '.*[Ll]inux.*'"

# VMs in specific cluster
kubectl mtv get inventory vms vmware -q "where cluster == 'Production-Cluster'"

# VMs with multiple network interfaces
kubectl mtv get inventory vms vmware -q "where len nics > 1"

# Large VMs (memory > 16GB and storage > 100GB)
kubectl mtv get inventory vms vmware -q "where memoryMB > 16384 and sum disks.*.capacityInKB > 104857600"
```

## Output Formats

kubectl-mtv supports multiple output formats for different use cases.

### Table Format (Default)

```bash
# Default table output
kubectl mtv get providers

# Explicit table format
kubectl mtv get providers -o table
```

### JSON Format

```bash
# JSON output for automation
kubectl mtv get providers -o json

# Pretty-printed JSON
kubectl mtv get providers -o json | jq .
```

### YAML Format

```bash
# YAML output for configuration
kubectl mtv get providers -o yaml

# Save configuration to file
kubectl mtv get plan my-plan -o yaml > my-plan.yaml
```

## Common Workflows

### Complete Migration Workflow

```bash
# 1. Create target provider
kubectl mtv create provider host --type openshift

# 2. Create source provider
kubectl mtv create provider vmware --type vsphere \
  -U https://vcenter.example.com/sdk \
  -u admin \
  -p "password" \
  --vddk-init-image quay.io/org/vddk:8.0.1

# 3. Browse available VMs
kubectl mtv get inventory vms vmware

# 4. Select VMs with query
kubectl mtv get inventory vms vmware -q "where name ~= 'prod-web-.*'"

# 5. Create detailed mappings with specific pairs
# Network mapping with production and DMZ networks
kubectl mtv create mapping network production-networks \
  --source-provider vmware \
  --target-provider host \
  --network-pairs "Production VLAN:production/prod-network,DMZ Network:security/dmz-network"

# Storage mapping with performance tiers
kubectl mtv create mapping storage production-storage \
  --source-provider vmware \
  --target-provider host \
  --storage-pairs "SSD-Datastore:production/fast-ssd,Standard-Datastore:production/standard-hdd"

# 6. Create migration plan with custom mappings
kubectl mtv create plan web-migration \
  --source-provider vmware \
  --vms prod-web-01,prod-web-02 \
  --network-map production-networks \
  --storage-map production-storage

# 7. Start migration
kubectl mtv start plan web-migration

# 8. Monitor progress
kubectl mtv describe plan web-migration
kubectl mtv get plan-vms web-migration

# 9. Archive completed plan
kubectl mtv archive plan web-migration
```

### Warm Migration Workflow

```bash
# 1. Create warm migration plan
kubectl mtv create plan critical-migration \
  --source-provider vmware \
  --vms critical-db,critical-app \
  --warm

# 2. Start initial sync
kubectl mtv start plan critical-migration

# 3. Monitor sync progress
kubectl mtv describe plan critical-migration

# 4. Wait for initial sync to complete and verify readiness
kubectl mtv get plan-vms critical-migration

# 5. During maintenance window, perform final cutover
# Option A: Cutover immediately
kubectl mtv cutover plan critical-migration

# Option B: Schedule cutover for specific maintenance window
kubectl mtv cutover plan critical-migration --cutover "2024-01-15T02:00:00Z"

# Option C: Schedule cutover for 30 minutes from now
kubectl mtv cutover plan critical-migration --cutover "$(date -d '+30 minutes' --iso-8601=sec)"

# 6. Verify migration completion
kubectl mtv get plan-vms critical-migration
```

### Batch Operations

```bash
# Create multiple providers
kubectl mtv create provider site1-vmware --type vsphere -U https://site1.example.com/sdk -u admin -p pass1
kubectl mtv create provider site2-vmware --type vsphere -U https://site2.example.com/sdk -u admin -p pass2

# Get inventory from multiple providers
kubectl mtv get inventory vms site1-vmware -o json > site1-vms.json
kubectl mtv get inventory vms site2-vmware -o json > site2-vms.json

# Create multiple plans
kubectl mtv create plan site1-migration --source-provider site1-vmware --vms vm1,vm2,vm3
kubectl mtv create plan site2-migration --source-provider site2-vmware --vms vm4,vm5,vm6

# Archive multiple completed plans
kubectl mtv archive plan site1-migration site2-migration
```

### Troubleshooting Workflows

```bash
# Check provider status
kubectl mtv get providers -o yaml

# Verify connectivity
kubectl mtv get inventory hosts <provider-name>

# Check plan details
kubectl mtv describe plan <plan-name>

# Monitor VM migration status
kubectl mtv describe plan-vm <plan-name> <vm-name>

# Get logs (using kubectl directly)
kubectl logs -n forklift-operator deployment/forklift-controller

# Check migration progress with verbose output
kubectl mtv get plan-vms <plan-name> -v=5
```

## Tips and Best Practices

### Query Optimization

- Use specific field comparisons when possible instead of broad pattern matching
- Combine conditions to reduce result sets
- Use parentheses for complex logical expressions

### Plan Management

- Start with small test migrations before large production migrations
- Use warm migrations for critical systems to minimize downtime
- Archive completed plans to keep the environment clean
- Test network and storage mappings with non-critical VMs first

### Provider Security

- Use secrets for sensitive credentials instead of command-line passwords
- Enable TLS verification in production environments
- Regularly rotate provider credentials
- Use least-privilege access for provider accounts

### Monitoring

- Use `describe` commands for detailed status information
- Monitor Kubernetes events for migration progress
- Set up alerts for failed migrations
- Keep provider inventories updated by refreshing regularly

For more detailed examples and advanced use cases, see:

- [Demo Walkthrough](README_demo.md)
- [VDDK Setup Guide](README_vddk.md)
- [Query Language Reference](README_quary_language.md)
