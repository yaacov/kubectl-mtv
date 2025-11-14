# Chapter 21: Command Reference

This chapter provides a comprehensive reference for all `kubectl-mtv` commands, subcommands, and their flags. Commands are organized by functionality to help you quickly find the right tool for your migration tasks.

## Global Flags

These flags are available for all `kubectl-mtv` commands:

| **Flag** | **Short** | **Type** | **Default** | **Description** |
|----------|-----------|----------|-------------|-----------------|
| `--verbose` | `-v` | int | 0 | Verbose output level (0=silent, 1=info, 2=debug, 3=trace) |
| `--all-namespaces` | `-A` | bool | false | List resources across all namespaces |
| `--use-utc` | | bool | false | Format timestamps in UTC instead of local timezone |
| `--kubeconfig` | | string | | Path to the kubeconfig file |
| `--context` | | string | | The name of the kubeconfig context to use |
| `--namespace` | `-n` | string | | If present, the namespace scope for this CLI request |

## Resource Management Commands

### get - Retrieve Resources

Get various MTV resources including plans, providers, mappings, and inventory.

#### get plan [PLAN_NAME]

Retrieve migration plans.

```bash
kubectl mtv get plan [plan-name] [flags]
kubectl mtv get plans [flags]  # Alias
```

**Flags:**
- `--output, -o`: Output format (table, json, yaml)
- `--watch, -w`: Watch for changes
- `--inventory-url, -i`: Base URL for the inventory service

#### get provider [PROVIDER_NAME]

Retrieve migration providers.

```bash
kubectl mtv get provider [provider-name] [flags]
kubectl mtv get providers [flags]  # Alias
```

**Flags:**
- `--output, -o`: Output format (table, json, yaml)
- `--watch, -w`: Watch for changes

#### get mapping [MAPPING_NAME]

Retrieve network and storage mappings.

```bash
kubectl mtv get mapping [mapping-name] [flags]
kubectl mtv get mappings [flags]  # Alias
```

**Flags:**
- `--output, -o`: Output format (table, json, yaml)
- `--watch, -w`: Watch for changes

#### get host [HOST_NAME]

Retrieve migration hosts (ESXi hosts for vSphere).

```bash
kubectl mtv get host [host-name] [flags]
kubectl mtv get hosts [flags]  # Alias
```

**Flags:**
- `--output, -o`: Output format (table, json, yaml)
- `--watch, -w`: Watch for changes

#### get hook [HOOK_NAME]

Retrieve migration hooks.

```bash
kubectl mtv get hook [hook-name] [flags]
kubectl mtv get hooks [flags]  # Alias
```

**Flags:**
- `--output, -o`: Output format (table, json, yaml)
- `--watch, -w`: Watch for changes

### get inventory - Query Provider Inventory

Get inventory resources from providers using the Tree Search Language (TSL) for advanced filtering.

#### get inventory vm PROVIDER_NAME

Retrieve virtual machines from provider inventory.

```bash
kubectl mtv get inventory vm <provider-name> [flags]
kubectl mtv get inventory vms <provider-name> [flags]  # Alias
```

**Flags:**
- `--query, -q`: TSL query filter (e.g., "where powerState = 'poweredOn'")
- `--output, -o`: Output format (table, json, yaml, planvms)
- `--watch, -w`: Watch for changes
- `--inventory-url`: Inventory service URL override

**Examples:**
```bash
# List all VMs
kubectl mtv get inventory vm my-vsphere-provider

# Filter powered-on VMs with more than 4GB RAM
kubectl mtv get inventory vm my-vsphere-provider -q "where powerState = 'poweredOn' and memory.size > 4096"

# Export VMs in planvms format for migration planning
kubectl mtv get inventory vm my-vsphere-provider -o planvms > vms.yaml
```

#### get inventory network PROVIDER_NAME

Retrieve networks from provider inventory.

```bash
kubectl mtv get inventory network <provider-name> [flags]
kubectl mtv get inventory networks <provider-name> [flags]  # Alias
```

#### get inventory storage PROVIDER_NAME

Retrieve storage resources from provider inventory.

```bash
kubectl mtv get inventory storage <provider-name> [flags]  
kubectl mtv get inventory storages <provider-name> [flags]  # Alias
```

#### get inventory host PROVIDER_NAME

Retrieve hosts from provider inventory.

```bash
kubectl mtv get inventory host <provider-name> [flags]
kubectl mtv get inventory hosts <provider-name> [flags]  # Alias
```

#### get inventory namespace PROVIDER_NAME

Retrieve namespaces from provider inventory.

```bash
kubectl mtv get inventory namespace <provider-name> [flags]
kubectl mtv get inventory namespaces <provider-name> [flags]  # Alias
```

All inventory subcommands support the same flags as `get inventory vm`.

### describe - Detailed Resource Information

Get detailed information about specific resources.

#### describe plan PLAN_NAME

```bash
kubectl mtv describe plan <plan-name> [flags]
```

#### describe provider PROVIDER_NAME

```bash
kubectl mtv describe provider <provider-name> [flags]
```

#### describe mapping MAPPING_NAME

```bash
kubectl mtv describe mapping <mapping-name> [flags]
```

#### describe host HOST_NAME

```bash
kubectl mtv describe host <host-name> [flags]
```

#### describe hook HOOK_NAME

```bash
kubectl mtv describe hook <hook-name> [flags]
```

### delete - Remove Resources

Delete MTV resources.

#### delete plan PLAN_NAME

```bash
kubectl mtv delete plan <plan-name> [flags]
```

#### delete provider PROVIDER_NAME

```bash
kubectl mtv delete provider <provider-name> [flags]
```

#### delete mapping MAPPING_NAME

```bash
kubectl mtv delete mapping <mapping-name> [flags]
```

#### delete host HOST_NAME

```bash
kubectl mtv delete host <host-name> [flags]
```

#### delete hook HOOK_NAME

```bash
kubectl mtv delete hook <hook-name> [flags]
```

## Resource Creation Commands

### create - Create New Resources

Create various MTV resources.

#### create provider PROVIDER_NAME

Create a new migration provider.

```bash
kubectl mtv create provider <name> [flags]
```

**Common Flags:**
- `--type, -t`: Provider type (openshift, vsphere, ovirt, openstack, ova)
- `--secret`: Secret containing provider credentials
- `--url, -U`: Provider URL
- `--username, -u`: Provider credentials username
- `--password, -p`: Provider credentials password
- `--cacert`: Provider CA certificate (use @filename to load from file)
- `--provider-insecure-skip-tls`: Skip TLS verification when connecting to the provider

**OpenShift Provider Flags:**
- `--token, -T`: Provider authentication token

**vSphere Provider Flags:**
- `--vddk-init-image`: Virtual Disk Development Kit (VDDK) container init image path
- `--sdk-endpoint`: SDK endpoint type for vSphere provider (vcenter or esxi)
- `--use-vddk-aio-optimization`: Enable VDDK AIO optimization for vSphere provider
- `--vddk-buf-size-in-64k`: VDDK buffer size in 64K units

**Examples:**
```bash
# vSphere provider
kubectl mtv create provider my-vsphere --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password VMware1! \
  --vddk-init-image quay.io/kubev2v/vddk:8.0.1

# OpenShift provider
kubectl mtv create provider my-openshift --type openshift \
  --url https://api.openshift.example.com:6443 \
  --token sha256~abc123...

# oVirt provider
kubectl mtv create provider my-ovirt --type ovirt \
  --url https://ovirt-engine.example.com/ovirt-engine/api \
  --username admin@internal \
  --password oVirtPassword123
```

#### create plan PLAN_NAME

Create a new migration plan.

```bash
kubectl mtv create plan <name> [flags]
```

**Provider and Mapping Flags:**
- `--source, -S`: Source provider name (supports namespace/name pattern)
- `--target, -t`: Target provider name (supports namespace/name pattern)
- `--network-mapping`: Network mapping name
- `--storage-mapping`: Storage mapping name
- `--network-pairs`: Network mapping pairs (comma-separated)
- `--storage-pairs`: Storage mapping pairs (comma-separated with semicolon parameters)

**VM Selection Flags:**
- `--vms`: List of VM names, file path (@file.yaml), or query string ('where ...')

**Plan Configuration Flags:**
- `--description`: Plan description
- `--target-namespace`: Target namespace for migrated VMs
- `--transfer-network`: Network attachment definition for disk transfer
- `--migration-type, -m`: Migration type (cold, warm, live, conversion)
- `--warm`: Enable warm migration (legacy flag)

**Storage Enhancement Flags:**
- `--default-volume-mode`: Default volume mode (Filesystem|Block)
- `--default-access-mode`: Default access mode (ReadWriteOnce|ReadWriteMany|ReadOnlyMany)
- `--default-offload-plugin`: Default offload plugin type (vsphere)
- `--default-offload-secret`: Existing offload secret name
- `--default-offload-vendor`: Default offload plugin vendor

**Storage Array Offload Flags:**
- `--offload-vsphere-username`: vSphere username for offload secret
- `--offload-vsphere-password`: vSphere password for offload secret
- `--offload-vsphere-url`: vSphere vCenter URL for offload secret
- `--offload-storage-username`: Storage array username for offload secret
- `--offload-storage-password`: Storage array password for offload secret
- `--offload-storage-endpoint`: Storage array management endpoint URL
- `--offload-cacert`: CA certificate for offload secret
- `--offload-insecure-skip-tls`: Skip TLS verification for offload connections

**Target VM Placement Flags:**
- `--target-labels, -L`: Target labels for VMs (key1=value1,key2=value2)
- `--target-node-selector`: Target node selector for VM scheduling
- `--target-affinity`: Target affinity using KARL syntax
- `--target-power-state`: Target power state (on, off, auto)

**Convertor Pod Optimization Flags:**
- `--convertor-labels`: Labels for virt-v2v convertor pods
- `--convertor-node-selector`: Node selector for convertor pod scheduling
- `--convertor-affinity`: Convertor affinity using KARL syntax

**Template and Customization Flags:**
- `--pvc-name-template`: PVC name template for VM disks
- `--volume-name-template`: Volume interface name template
- `--network-name-template`: Network interface name template

**Advanced Flags:**
- `--preserve-cluster-cpu-model`: Preserve CPU model from oVirt cluster
- `--preserve-static-ips`: Preserve static IPs of vSphere VMs (default: true)
- `--migrate-shared-disks`: Migrate shared disks (default: true)
- `--skip-guest-conversion`: Skip guest conversion process
- `--delete-vm-on-fail-migration`: Delete target VM when migration fails
- `--install-legacy-drivers`: Install legacy Windows drivers (true/false/auto)
- `--use-compatibility-mode`: Use compatibility devices for bootability (default: true)

**Hook Flags:**
- `--pre-hook`: Pre-migration hook for all VMs
- `--post-hook`: Post-migration hook for all VMs

**Examples:**
```bash
# Simple plan with VM list
kubectl mtv create plan my-migration \
  --source my-vsphere --target my-openshift \
  --vms vm-web-01,vm-db-02,vm-app-03

# Plan with query-based VM selection
kubectl mtv create plan production-migration \
  --source vsphere-prod --target openshift-prod \
  --vms "where powerState = 'poweredOn' and name like 'prod-%'" \
  --migration-type warm

# Plan with storage offloading
kubectl mtv create plan offload-migration \
  --source vsphere-datacenter --target openshift-production \
  --storage-pairs "tier1-ds:flashsystem-gold;offloadPlugin=vsphere;offloadVendor=flashsystem" \
  --offload-vsphere-username vcenter-svc@vsphere.local \
  --offload-storage-username flashsystem-admin
```

#### create mapping MAPPING_NAME

Create network or storage mappings.

```bash
kubectl mtv create mapping network <name> [flags]
kubectl mtv create mapping storage <name> [flags]
```

**Network Mapping Flags:**
- `--source, -S`: Source provider name
- `--target, -T`: Target provider name
- `--network-pairs`: Network mapping pairs

**Storage Mapping Flags:**
- `--source, -S`: Source provider name
- `--target, -T`: Target provider name
- `--storage-pairs`: Storage mapping pairs with enhanced options
- `--default-volume-mode`: Default volume mode
- `--default-access-mode`: Default access mode
- `--default-offload-plugin`: Default offload plugin type
- `--default-offload-secret`: Default offload secret name
- `--default-offload-vendor`: Default offload plugin vendor

**Storage Array Offload Flags:** (Same as create plan)

**Examples:**
```bash
# Network mapping
kubectl mtv create mapping network prod-networks \
  --source vsphere-prod --target openshift-prod \
  --network-pairs "Production VLAN:prod-network,Management:mgmt-network"

# Storage mapping with offloading
kubectl mtv create mapping storage enterprise-storage \
  --source vsphere-prod --target openshift-prod \
  --storage-pairs "premium-ds:flashsystem-tier1;offloadPlugin=vsphere;offloadVendor=flashsystem" \
  --offload-vsphere-username admin@vsphere.local
```

#### create host HOST_NAME

Create migration hosts for vSphere environments.

```bash
kubectl mtv create host <name> [flags]
```

**Flags:**
- `--provider`: vSphere provider name (required)
- `--ip-address`: IP address for disk transfer (mutually exclusive with --network-adapter)
- `--network-adapter`: Network adapter name to get IP from inventory (mutually exclusive with --ip-address)
- `--secret`: Existing secret with host credentials
- `--username`: Host username (creates new secret if no --secret provided)
- `--password`: Host password (creates new secret if no --secret provided)
- `--host-insecure-skip-tls`: Skip TLS verification for host connections
- `--cacert`: CA certificate for host authentication

**Examples:**
```bash
# Host with direct IP
kubectl mtv create host esxi-host-01 \
  --provider my-vsphere-provider \
  --ip-address 192.168.1.10

# Host with network adapter lookup
kubectl mtv create host esxi-host-02 \
  --provider my-vsphere-provider \
  --network-adapter vmk1 \
  --username root --password ESXiPassword123
```

#### create hook HOOK_NAME

Create migration hooks for custom automation.

```bash
kubectl mtv create hook <name> [flags]
```

**Flags:**
- `--image`: Container image to run (default: quay.io/kubev2v/hook-runner)
- `--playbook`: Ansible playbook content (use @filename to load from file)
- `--service-account`: Service account for hook execution
- `--deadline`: Hook execution deadline in seconds

**Examples:**
```bash
# Hook with inline playbook
kubectl mtv create hook backup-hook \
  --playbook "- hosts: localhost\n  tasks:\n  - debug: msg='Backup complete'"

# Hook with playbook file
kubectl mtv create hook database-backup \
  --playbook @backup-playbook.yml \
  --service-account hook-runner-sa \
  --deadline 300
```

#### create vddk-image

Build VDDK container images for VMware environments.

```bash
kubectl mtv create vddk-image [flags]
```

**Flags:**
- `--tar`: Path to VMware VDDK tar.gz file (required)
- `--tag`: Container image tag (required)
- `--build-dir`: Build directory (optional, uses temp dir if not set)
- `--runtime`: Container runtime (auto, podman, docker)
- `--platform`: Target platform (amd64, arm64)
- `--dockerfile`: Path to custom Dockerfile
- `--push`: Push image after build

**Examples:**
```bash
# Build VDDK image
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-8.0.1.tar.gz \
  --tag quay.io/myorg/vddk:8.0.1

# Build and push
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-8.0.1.tar.gz \
  --tag quay.io/myorg/vddk:8.0.1 --push
```

## Plan Lifecycle Commands

### start - Begin Migration

Start migration plans.

#### start plan PLAN_NAME

```bash
kubectl mtv start plan <plan-name> [plan-name2 ...] [flags]
```

Start one or more migration plans.

### cancel - Stop Migration

Cancel running migration plans.

#### cancel plan PLAN_NAME

```bash
kubectl mtv cancel plan <plan-name> [plan-name2 ...] [flags]
```

Cancel one or more running migration plans.

### cutover - Complete Warm Migration

Perform cutover for warm migrations.

#### cutover plan PLAN_NAME

```bash
kubectl mtv cutover plan <plan-name> [plan-name2 ...] [flags]
```

Complete the cutover phase for warm migration plans.

### archive - Archive Plans

Archive completed migration plans.

#### archive plan PLAN_NAME

```bash
kubectl mtv archive plan <plan-name> [plan-name2 ...] [flags]
```

Archive one or more migration plans to preserve history while hiding from active lists.

### unarchive - Restore Plans

Restore archived migration plans.

#### unarchive plan PLAN_NAME

```bash
kubectl mtv unarchive plan <plan-name> [plan-name2 ...] [flags]
```

Restore one or more archived migration plans to active status.

## Resource Modification Commands

### patch - Modify Existing Resources

Modify existing MTV resources.

#### patch plan PLAN_NAME

Update migration plan settings.

```bash
kubectl mtv patch plan <plan-name> [flags]
```

**Plan-Level Flags:**
- `--migration-type`: Update migration type
- `--transfer-network`: Update transfer network
- `--target-labels`: Update target VM labels
- `--target-node-selector`: Update target node selector
- `--target-affinity`: Update target affinity rules
- `--convertor-labels`: Update convertor pod labels
- `--convertor-node-selector`: Update convertor node selector
- `--convertor-affinity`: Update convertor affinity rules

**Examples:**
```bash
# Change migration type
kubectl mtv patch plan my-migration --migration-type warm

# Update placement settings
kubectl mtv patch plan my-migration \
  --target-affinity "REQUIRE pods(app=database) on node" \
  --convertor-affinity "PREFER nodes(storage=true) on node"
```

#### patch planvm PLAN_NAME VM_NAME

Update settings for individual VMs within a plan.

```bash
kubectl mtv patch planvm <plan-name> <vm-name> [flags]
```

**VM-Specific Flags:**
- `--target-name`: Custom target VM name
- `--instance-type`: KubeVirt instance type
- `--target-power-state`: VM power state after migration (on, off, auto)
- `--add-hook`: Add a hook (format: step:hook-namespace/hook-name)
- `--remove-hook`: Remove a hook (format: step:hook-namespace/hook-name)
- `--clear-hooks`: Clear all hooks for the VM
- `--pvc-name-template`: Custom PVC naming template for this VM
- `--volume-name-template`: Custom volume naming template for this VM
- `--network-name-template`: Custom network naming template for this VM

**Examples:**
```bash
# Customize VM name and instance type
kubectl mtv patch planvm my-migration vm-web-01 \
  --target-name web-server-prod \
  --instance-type large

# Add hooks to a VM
kubectl mtv patch planvm my-migration vm-db-01 \
  --add-hook PreMigration:default/backup-hook \
  --add-hook PostMigration:default/validation-hook
```

#### patch mapping MAPPING_NAME

Update network or storage mappings.

```bash
kubectl mtv patch mapping network <mapping-name> [flags]
kubectl mtv patch mapping storage <mapping-name> [flags]
```

**Mapping Update Flags:**
- `--add-pairs`: Add new mapping pairs
- `--update-pairs`: Update existing mapping pairs
- `--remove-pairs`: Remove mapping pairs (specify source names)

**Examples:**
```bash
# Add network mapping pairs
kubectl mtv patch mapping network prod-networks \
  --add-pairs "DMZ Network:security/dmz-net,Backup Network:ignored"

# Update storage mapping pairs
kubectl mtv patch mapping storage enterprise-storage \
  --update-pairs "premium-ds:ultra-fast-ssd;volumeMode=Block"
```

#### patch provider PROVIDER_NAME

Update provider settings.

```bash
kubectl mtv patch provider <provider-name> [flags]
```

**Provider Update Flags:**
- `--url`: Update provider URL
- `--username`: Update username
- `--password`: Update password
- `--token`: Update authentication token
- `--cacert`: Update CA certificate
- `--insecure-skip-tls`: Update TLS verification setting

## AI Integration Commands

### mcp-server - Model Context Protocol Server

Start the MCP server for AI assistant integration.

```bash
kubectl mtv mcp-server [flags]
```

**Flags:**
- `--sse`: Run in SSE (Server-Sent Events) mode over HTTP
- `--port`: Port to listen on for SSE mode (default: 8080)
- `--host`: Host address to bind to for SSE mode (default: 127.0.0.1)
- `--cert-file`: Path to TLS certificate file
- `--key-file`: Path to TLS private key file

**Modes:**
- **Default (Stdio)**: For direct AI assistant integration
- **SSE Mode**: HTTP server mode with optional TLS

**Examples:**
```bash
# Stdio mode (default) - for AI assistant integration
kubectl mtv mcp-server

# HTTP server mode
kubectl mtv mcp-server --sse --port 8080

# HTTPS server mode with TLS
kubectl mtv mcp-server --sse --port 8443 \
  --cert-file /path/to/cert.pem \
  --key-file /path/to/key.pem
```

## Utility Commands

### version - Version Information

Display version information.

```bash
kubectl mtv version [flags]
```

Shows the `kubectl-mtv` version, build information, and runtime details.

## Query Language (TSL) Syntax

The Tree Search Language (TSL) is used with inventory commands for advanced filtering:

### Basic Syntax

```
kubectl mtv get inventory vm <provider> --query "where <condition>"
```

### Operators

| **Operator** | **Description** | **Example** |
|--------------|-----------------|-------------|
| `=` | Equal | `name = 'vm-01'` |
| `!=` | Not equal | `powerState != 'poweredOff'` |
| `>`, `>=` | Greater than | `memory.size > 4096` |
| `<`, `<=` | Less than | `disks.count <= 2` |
| `like` | Pattern match (case-sensitive) | `name like 'web-%'` |
| `ilike` | Pattern match (case-insensitive) | `name ilike 'WEB-%'` |
| `in` | In list | `powerState in ('poweredOn', 'suspended')` |
| `and` | Logical AND | `powerState = 'poweredOn' and memory.size > 2048` |
| `or` | Logical OR | `name like 'web-%' or name like 'app-%'` |
| `not` | Logical NOT | `not powerState = 'poweredOff'` |

### Functions

| **Function** | **Description** | **Example** |
|--------------|-----------------|-------------|
| `len(field)` | Length of array | `len(disks) > 1` |
| `sum(field)` | Sum of numeric values | `sum(disks.capacityInBytes) > 107374182400` |

### Query Examples

```bash
# Powered-on VMs with more than 4GB RAM
kubectl mtv get inventory vm vsphere-prod -q "where powerState = 'poweredOn' and memory.size > 4096"

# VMs with names starting with "prod-"
kubectl mtv get inventory vm vsphere-prod -q "where name like 'prod-%'"

# VMs with multiple disks and large total capacity
kubectl mtv get inventory vm vsphere-prod -q "where len(disks) > 1 and sum(disks.capacityInBytes) > 53687091200"

# VMs in specific power states
kubectl mtv get inventory vm vsphere-prod -q "where powerState in ('poweredOn', 'suspended')"

# Complex query with multiple conditions
kubectl mtv get inventory vm vsphere-prod -q "where (name like 'web-%' or name like 'app-%') and powerState = 'poweredOn' and memory.size between 2048 and 8192"
```

## KARL Syntax (Kubernetes Affinity Rule Language)

KARL is used for advanced scheduling with `--target-affinity` and `--convertor-affinity` flags:

### Rule Types

| **Rule Type** | **Description** |
|---------------|-----------------|
| `REQUIRE` | Hard affinity (must be satisfied) |
| `PREFER` | Soft affinity (preferred but not required) |
| `AVOID` | Hard anti-affinity (must be avoided) |
| `REPEL` | Soft anti-affinity (avoided when possible) |

### Topology Keys

| **Topology** | **Description** |
|--------------|-----------------|
| `node` | Specific node placement |
| `zone` | Availability zone placement |
| `region` | Region-based placement |

### KARL Examples

```bash
# Require co-location with database pods
--target-affinity "REQUIRE pods(app=database) on node"

# Prefer nodes with SSD storage
--target-affinity "PREFER nodes(storage=ssd) on node"

# Avoid nodes running cache workloads
--target-affinity "AVOID nodes(workload=cache) on node"

# Distribute across zones
--target-affinity "REPEL pods(app=myapp) on zone"
```

## Common Command Patterns

### Migration Workflow

```bash
# 1. Create providers
kubectl mtv create provider vsphere-source --type vsphere --url https://vcenter.company.com --username admin --password pass123
kubectl mtv create provider openshift-target --type openshift --url https://api.openshift.company.com:6443 --token sha256~token

# 2. Create mappings (optional)
kubectl mtv create mapping network net-mapping --source vsphere-source --target openshift-target --network-pairs "VLAN100:prod-network"
kubectl mtv create mapping storage stor-mapping --source vsphere-source --target openshift-target --storage-pairs "datastore1:premium-ssd"

# 3. Create migration plan
kubectl mtv create plan my-migration --source vsphere-source --target openshift-target --vms vm1,vm2,vm3 --migration-type warm

# 4. Start migration
kubectl mtv start plan my-migration

# 5. Monitor progress
kubectl mtv get plan my-migration --watch

# 6. Complete warm migration (if applicable)
kubectl mtv cutover plan my-migration
```

### Inventory Exploration

```bash
# List all VMs
kubectl mtv get inventory vm vsphere-source

# Filter and explore
kubectl mtv get inventory vm vsphere-source -q "where powerState = 'poweredOn'"
kubectl mtv get inventory vm vsphere-source -q "where memory.size > 4096" -o json

# Export for planning
kubectl mtv get inventory vm vsphere-source -o planvms > migration-candidates.yaml
```

### Troubleshooting

```bash
# Verbose output for debugging
kubectl mtv get plan my-migration -v=2

# Detailed resource information
kubectl mtv describe plan my-migration
kubectl mtv describe provider vsphere-source

# Check inventory connectivity
kubectl mtv get inventory vm vsphere-source -v=3
```

---

*Previous: [Chapter 20: Integration with KubeVirt Tools](20-integration-with-kubevirt-tools.md)*
