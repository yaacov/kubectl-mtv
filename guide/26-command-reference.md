---
layout: default
title: "Chapter 26: Command Reference"
parent: "VII. Reference and Appendices"
nav_order: 1
---

This chapter provides a comprehensive reference for all `kubectl-mtv` commands, subcommands, and their flags. Commands are organized by functionality to help you quickly find the right tool for your migration tasks.

## Global Flags

These flags are available for all `kubectl-mtv` commands:

| **Flag** | **Short** | **Type** | **Default** | **Description** |
|----------|-----------|----------|-------------|-----------------|
| `--verbose` | `-v` | int | 0 | Verbose output level (0=silent, 1=info, 2=debug, 3=trace) |
| `--all-namespaces` | `-A` | bool | false | List resources across all namespaces |
| `--use-utc` | | bool | false | Format timestamps in UTC instead of local timezone |
| `--inventory-url` | `-i` | string | `$MTV_INVENTORY_URL` | Base URL for the inventory service |
| `--inventory-insecure-skip-tls` | | bool | `$MTV_INVENTORY_INSECURE_SKIP_TLS` | Skip TLS verification for inventory service connections |
| `--kubeconfig` | | string | | Path to the kubeconfig file |
| `--context` | | string | | The name of the kubeconfig context to use |
| `--namespace` | `-n` | string | | If present, the namespace scope for this CLI request |
| `--no-color` | | bool | `$NO_COLOR` | Disable colored output (also respects NO_COLOR env var) |

## Resource Management Commands

### get - Retrieve Resources

Get various MTV resources including plans, providers, mappings, and inventory.

#### get plan [--name PLAN_NAME]

Retrieve migration plans.

```bash
kubectl mtv get plan [--name plan-name] [flags]
kubectl mtv get plans [flags]  # Alias
```

**Flags:**
- `--name, -M`: Plan name (optional, omit to list all)
- `--output, -o`: Output format (table, json, yaml)
- `--watch, -w`: Watch for changes
- `--vms`: Get VMs status in the migration plan (requires plan name)
- `--disk`: Get disk transfer status in the migration plan (requires plan name)
- `--vms-table`: Show all VMs across plans in a flat table with source/target inventory details
- `--query, -q`: Query filter using TSL syntax (only with `--vms-table`)
- `--inventory-url, -i`: Base URL for the inventory service

**VMs Table Examples:**

The `--vms-table` flag produces a flat table of all VMs across plans with columns: VM, SOURCE STATUS, SOURCE IP, TARGET, TARGET IP, TARGET STATUS, PLAN, PLAN STATUS, and PROGRESS.

```bash
# Show all VMs across all plans in a flat table
kubectl mtv get plans --vms-table

# Show VMs for a specific plan in a table
kubectl mtv get plan --name my-migration --vms-table

# Filter VMs table by plan status
kubectl mtv get plans --vms-table --query "where planStatus = 'Failed'"

# Filter VMs table by source power state
kubectl mtv get plans --vms-table --query "where sourceStatus = 'poweredOn'"

# Export VMs table as JSON
kubectl mtv get plans --vms-table --output json

# Watch VMs table for real-time updates
kubectl mtv get plans --vms-table --watch
```

#### get provider [--name PROVIDER_NAME]

Retrieve migration providers.

```bash
kubectl mtv get provider [--name provider-name] [flags]
kubectl mtv get providers [flags]  # Alias
```

**Flags:**
- `--name, -M`: Provider name (optional, omit to list all)
- `--output, -o`: Output format (table, json, yaml)
- `--watch, -w`: Watch for changes

#### get mapping [--name MAPPING_NAME]

Retrieve network and storage mappings.

```bash
kubectl mtv get mapping [--name mapping-name] [flags]
kubectl mtv get mappings [flags]  # Alias
```

**Flags:**
- `--name, -M`: Mapping name (optional, omit to list all)
- `--output, -o`: Output format (table, json, yaml)
- `--watch, -w`: Watch for changes

#### get host [--name HOST_NAME]

Retrieve migration hosts (ESXi hosts for vSphere).

```bash
kubectl mtv get hosts [flags]                    # List all hosts
kubectl mtv get host --name <host-name> [flags]  # Get specific host
```

**Flags:**
- `--name, -M`: Host name (optional, omit to list all)
- `--output, -o`: Output format (table, json, yaml)
- `--watch, -w`: Watch for changes

#### get hook [--name HOOK_NAME]

Retrieve migration hooks.

```bash
kubectl mtv get hooks [flags]                    # List all hooks
kubectl mtv get hook --name <hook-name> [flags]  # Get specific hook
```

**Flags:**
- `--name, -M`: Hook name (optional, omit to list all)
- `--output, -o`: Output format (table, json, yaml)
- `--watch, -w`: Watch for changes

### get inventory - Query Provider Inventory

Get inventory resources from providers using the Tree Search Language (TSL) for advanced filtering.

#### get inventory vms --provider PROVIDER_NAME

Retrieve virtual machines from provider inventory.

```bash
kubectl mtv get inventory vms --provider <provider-name> [flags]
```

**Flags:**
- `--provider, -p`: Provider name (required)
- `--query, -q`: [TSL](../27-tsl-tree-search-language-reference) query filter (e.g., "where powerState = 'poweredOn'")
- `--output, -o`: Output format (table, json, yaml, planvms)
- `--watch, -w`: Watch for changes
- `--inventory-url`: Inventory service URL override

**Examples:**
```bash
# List all VMs
kubectl mtv get inventory vms --provider my-vsphere-provider

# Filter powered-on VMs with more than 4GB RAM
kubectl mtv get inventory vms --provider my-vsphere-provider --query "where powerState = 'poweredOn' and memory.size > 4096"

# Export VMs in planvms format for migration planning
kubectl mtv get inventory vms --provider my-vsphere-provider --output planvms > vms.yaml
```

#### get inventory networks --provider PROVIDER_NAME

Retrieve networks from provider inventory.

```bash
kubectl mtv get inventory networks --provider <provider-name> [flags]
```

#### get inventory storages --provider PROVIDER_NAME

Retrieve storage resources from provider inventory.

```bash
kubectl mtv get inventory storages --provider <provider-name> [flags]
```

#### get inventory hosts --provider PROVIDER_NAME

Retrieve hosts from provider inventory.

```bash
kubectl mtv get inventory hosts --provider <provider-name> [flags]
```

#### get inventory namespaces --provider PROVIDER_NAME

Retrieve namespaces from provider inventory.

```bash
kubectl mtv get inventory namespaces --provider <provider-name> [flags]
```

All inventory subcommands support the same flags as `get inventory vms`.

### describe - Detailed Resource Information

Get detailed information about specific resources.

#### describe plan --name PLAN_NAME

```bash
kubectl mtv describe plan --name <plan-name> [flags]
```

#### describe provider --name PROVIDER_NAME

```bash
kubectl mtv describe provider --name <provider-name> [flags]
```

#### describe mapping network --name NAME / describe mapping storage --name NAME

```bash
kubectl mtv describe mapping network --name <mapping-name> [flags]
kubectl mtv describe mapping storage --name <mapping-name> [flags]
```

#### describe host --name HOST_NAME

```bash
kubectl mtv describe host --name <host-name> [flags]
```

#### describe hook --name HOOK_NAME

```bash
kubectl mtv describe hook --name <hook-name> [flags]
```

### delete - Remove Resources

Delete MTV resources. To delete multiple resources at once, pass a comma-separated
list to `--name, -M` (e.g., `--name plan1,plan2,plan3`). You can also use `--names` as an
alias for `--name`.

#### delete plan --name PLAN_NAME

```bash
kubectl mtv delete plan --name <plan-name> [flags]
```

#### delete provider --name PROVIDER_NAME

```bash
kubectl mtv delete provider --name <provider-name> [flags]
```

#### delete mapping network --name NAME / delete mapping storage --name NAME

```bash
kubectl mtv delete mapping network --name <mapping-name> [flags]
kubectl mtv delete mapping storage --name <mapping-name> [flags]
```

#### delete host --name HOST_NAME

```bash
kubectl mtv delete host --name <host-name> [flags]
```

#### delete hook --name HOOK_NAME

```bash
kubectl mtv delete hook --name <hook-name> [flags]
```

## Resource Creation Commands

### create - Create New Resources

Create various MTV resources.

#### create provider --name PROVIDER_NAME

Create a new migration provider.

```bash
kubectl mtv create provider --name <name> [flags]
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
- `--provider-token, -T`: Provider authentication token

**vSphere Provider Flags:**
- `--vddk-init-image`: Virtual Disk Development Kit (VDDK) container init image path
- `--sdk-endpoint`: SDK endpoint type for vSphere provider (vcenter or esxi)
- `--use-vddk-aio-optimization`: Enable VDDK AIO optimization for vSphere provider
- `--vddk-buf-size-in-64k`: VDDK buffer size in 64K units

**Examples:**
```bash
# vSphere provider
kubectl mtv create provider --name my-vsphere --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password VMware1! \
  --vddk-init-image quay.io/kubev2v/vddk:8.0.1

# OpenShift provider
kubectl mtv create provider --name my-openshift --type openshift \
  --url https://api.openshift.example.com:6443 \
  --provider-token sha256~abc123...

# oVirt provider
kubectl mtv create provider --name my-ovirt --type ovirt \
  --url https://ovirt-engine.example.com/ovirt-engine/api \
  --username admin@internal \
  --password oVirtPassword123
```

#### create plan --name PLAN_NAME

Create a new migration plan.

```bash
kubectl mtv create plan --name <name> [flags]
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
- `--target-affinity`: Target affinity using [KARL](../28-karl-kubernetes-affinity-rule-language-reference) syntax
- `--target-power-state`: Target power state (on, off, auto)

**Convertor Pod Optimization Flags:**
- `--convertor-labels`: Labels for virt-v2v convertor pods
- `--convertor-node-selector`: Node selector for convertor pod scheduling
- `--convertor-affinity`: Convertor affinity using [KARL](../28-karl-kubernetes-affinity-rule-language-reference) syntax

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
kubectl mtv create plan --name my-migration \
  --source my-vsphere --target my-openshift \
  --vms vm-web-01,vm-db-02,vm-app-03

# Plan with query-based VM selection
kubectl mtv create plan --name production-migration \
  --source vsphere-prod --target openshift-prod \
  --vms "where powerState = 'poweredOn' and name like 'prod-%'" \
  --migration-type warm

# Plan with storage offloading
kubectl mtv create plan --name offload-migration \
  --source vsphere-datacenter --target openshift-production \
  --storage-pairs "tier1-ds:flashsystem-gold;offloadPlugin=vsphere;offloadVendor=flashsystem" \
  --offload-vsphere-username vcenter-svc@vsphere.local \
  --offload-storage-username flashsystem-admin
```

#### create mapping network --name NAME / create mapping storage --name NAME

Create network or storage mappings.

```bash
kubectl mtv create mapping network --name <name> [flags]
kubectl mtv create mapping storage --name <name> [flags]
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
kubectl mtv create mapping network --name prod-networks \
  --source vsphere-prod --target openshift-prod \
  --network-pairs "Production VLAN:prod-network,Management:mgmt-network"

# Storage mapping with offloading
kubectl mtv create mapping storage --name enterprise-storage \
  --source vsphere-prod --target openshift-prod \
  --storage-pairs "premium-ds:flashsystem-tier1;offloadPlugin=vsphere;offloadVendor=flashsystem" \
  --offload-vsphere-username admin@vsphere.local
```

#### create host --host-id HOST_ID

Create migration hosts for vSphere environments.

> **Important:** The `--host-id` flag expects **inventory host IDs** (e.g. `host-8`), not display
> names or IP addresses. Use `kubectl-mtv get inventory host --provider <name>` to list available IDs.

```bash
kubectl mtv create host --host-id <id> [flags]
```

**Flags:**
- `--host-id`: Inventory host ID(s) to create, comma-separated (required); use `get inventory host` to list IDs
- `--provider, -p`: vSphere provider name (required)
- `--ip-address`: IP address for disk transfer (mutually exclusive with --network-adapter)
- `--network-adapter`: Network adapter name to get IP from inventory (mutually exclusive with --ip-address)
- `--existing-secret`: Existing secret with host credentials
- `--username`: Host username (creates new secret if no --existing-secret provided)
- `--password`: Host password (creates new secret if no --existing-secret provided)
- `--host-insecure-skip-tls`: Skip TLS verification for host connections
- `--cacert`: CA certificate for host authentication

**Examples:**
```bash
# List available host IDs first
kubectl mtv get inventory host --provider my-vsphere-provider

# Host with direct IP
kubectl mtv create host --host-id host-8 \
  --provider my-vsphere-provider \
  --ip-address 192.168.1.10

# Host with network adapter lookup
kubectl mtv create host --host-id host-12 \
  --provider my-vsphere-provider \
  --network-adapter "Management Network" \
  --username root --password ESXiPassword123
```

#### create hook --name HOOK_NAME

Create migration hooks for custom automation.

```bash
kubectl mtv create hook --name <name> [flags]
```

**Flags:**
- `--image`: Container image to run (default: quay.io/kubev2v/hook-runner)
- `--playbook`: Ansible playbook content (use @filename to load from file)
- `--service-account`: Service account for hook execution
- `--deadline`: Hook execution deadline in seconds

**Examples:**
```bash
# Hook with inline playbook
kubectl mtv create hook --name backup-hook \
  --playbook "- hosts: localhost\n  tasks:\n  - debug: msg='Backup complete'"

# Hook with playbook file
kubectl mtv create hook --name database-backup \
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

Commands that accept multiple resource names use comma-separated values in `--name, -M`
(e.g., `--name plan1,plan2,plan3`). You can also use `--names` as an alias for `--name`.

### start - Begin Migration

Start migration plans.

#### start plan / start plans

```bash
kubectl mtv start plan --name <plan-name> [flags]       # Start one plan
kubectl mtv start plans --name plan1,plan2,plan3 [flags]  # Start multiple plans
```

Start one or more migration plans.

### cancel - Stop Migration

Cancel running migration plans.

#### cancel plan --name PLAN_NAME

```bash
kubectl mtv cancel plan --name <plan-name> [flags]
```

Cancel specific VMs in one or more running migration plans.

### cutover - Complete Warm Migration

Perform cutover for warm migrations.

#### cutover plan / cutover plans

```bash
kubectl mtv cutover plan --name <plan-name> [flags]      # Cutover one plan
kubectl mtv cutover plans --name plan1,plan2 [flags]    # Cutover multiple plans
```

Complete the cutover phase for warm migration plans.

### archive - Archive Plans

Archive completed migration plans.

#### archive plan / archive plans

```bash
kubectl mtv archive plan --name <plan-name> [flags]     # Archive one plan
kubectl mtv archive plans --name plan1,plan2 [flags]    # Archive multiple plans
```

Archive one or more migration plans.

### unarchive - Restore Plans

Restore archived migration plans.

#### unarchive plan / unarchive plans

```bash
kubectl mtv unarchive plan --name <plan-name> [flags]   # Unarchive one plan
kubectl mtv unarchive plans --name plan1,plan2 [flags]  # Unarchive multiple plans
```

Restore one or more archived migration plans.

## Resource Modification Commands

### patch - Modify Existing Resources

Modify existing MTV resources.

#### patch plan --plan-name PLAN_NAME

Update migration plan settings.

```bash
kubectl mtv patch plan --plan-name <plan-name> [flags]
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
kubectl mtv patch plan --plan-name my-migration --migration-type warm

# Update placement settings
kubectl mtv patch plan --plan-name my-migration \
  --target-affinity "REQUIRE pods(app=database) on node" \
  --convertor-affinity "PREFER nodes(storage=true) on node"
```

#### patch planvm --plan-name PLAN_NAME --vm-name VM_NAME

Update settings for individual VMs within a plan.

```bash
kubectl mtv patch planvm --plan-name <plan-name> --vm-name <vm-name> [flags]
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
kubectl mtv patch planvm --plan-name my-migration --vm-name vm-web-01 \
  --target-name web-server-prod \
  --instance-type large

# Add hooks to a VM
kubectl mtv patch planvm --plan-name my-migration --vm-name vm-db-01 \
  --add-hook PreMigration:default/backup-hook \
  --add-hook PostMigration:default/validation-hook
```

#### patch mapping --name MAPPING_NAME

Update network or storage mappings.

```bash
kubectl mtv patch mapping network --name <mapping-name> [flags]
kubectl mtv patch mapping storage --name <mapping-name> [flags]
```

**Mapping Update Flags:**
- `--add-pairs`: Add new mapping pairs
- `--update-pairs`: Update existing mapping pairs
- `--remove-pairs`: Remove mapping pairs (specify source names)

**Examples:**
```bash
# Add network mapping pairs
kubectl mtv patch mapping network --name prod-networks \
  --add-pairs "DMZ Network:security/dmz-net,Backup Network:ignored"

# Update storage mapping pairs
kubectl mtv patch mapping storage --name enterprise-storage \
  --update-pairs "premium-ds:ultra-fast-ssd;volumeMode=Block"
```

#### patch provider --name PROVIDER_NAME

Update provider settings.

```bash
kubectl mtv patch provider --name <provider-name> [flags]
```

**Provider Update Flags:**
- `--url`: Update provider URL
- `--username`: Update username
- `--password`: Update password
- `--provider-token`: Update authentication token
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

## Health and Settings Commands

### health - System Health Check

Run comprehensive diagnostics on the MTV/Forklift system.

```bash
kubectl mtv health [flags]
```

**Flags:**
- `--output, -o`: Output format (table, json, yaml)
- `--skip-logs`: Skip pod log analysis for faster execution
- `--log-lines`: Number of log lines per pod to analyze (default: 100)
- `--namespace, -n`: Scope providers and plans to a namespace
- `--all-namespaces, -A`: Scan providers and plans across all namespaces

**Checks performed:**
1. Operator installation and version
2. ForkliftController status and feature flags
3. Forklift pod health (status, restarts, OOM)
4. Pod log analysis (errors and warnings)
5. Provider connectivity and readiness
6. Migration plan status

**Examples:**
```bash
# Quick health check
kubectl mtv health

# JSON output for automation
kubectl mtv health --output json

# Fast check without log analysis
kubectl mtv health --skip-logs

# Cluster-wide check
kubectl mtv health --all-namespaces

# Namespace-scoped check
kubectl mtv health --namespace production
```

### settings - ForkliftController Settings Management

View and configure ForkliftController settings.

#### settings get [--setting SETTING]

Get current setting values.

```bash
kubectl mtv settings [flags]
kubectl mtv settings get [--setting <setting-name>] [flags]
```

**Flags:**
- `--output, -o`: Output format (table, json, yaml)
- `--all`: Include all settings (supported + extended)

**Examples:**
```bash
# View all supported settings
kubectl mtv settings

# View all settings including extended ones
kubectl mtv settings --all

# Get a specific setting
kubectl mtv settings get --setting controller_max_vm_inflight

# Output as JSON
kubectl mtv settings --output json
```

#### settings set --setting SETTING --value VALUE

Set a ForkliftController setting value.

```bash
kubectl mtv settings set --setting <setting-name> --value <value>
```

**Examples:**
```bash
# Increase concurrent VM migrations
kubectl mtv settings set --setting controller_max_vm_inflight --value 40

# Enable a feature flag
kubectl mtv settings set --setting feature_ocp_live_migration --value true

# Change log level
kubectl mtv settings set --setting controller_log_level --value 5

# Set virt-v2v memory limit
kubectl mtv settings set --setting virt_v2v_container_limits_memory --value 16Gi
```

#### settings unset --setting SETTING

Remove a setting to revert it to the default value.

```bash
kubectl mtv settings unset --setting <setting-name>
```

**Examples:**
```bash
# Revert max concurrent VMs to default
kubectl mtv settings unset --setting controller_max_vm_inflight

# Revert log level to default
kubectl mtv settings unset --setting controller_log_level
```

## Utility Commands

### version - Version Information

Display version information.

```bash
kubectl mtv version [flags]
```

Shows the `kubectl-mtv` version, build information, and runtime details.

### help - Help and Reference

Get help for any command, browse help topics, or output machine-readable command schemas.

```bash
kubectl mtv help [command]
kubectl mtv help [topic]
kubectl mtv help --machine [command] [flags]
```

#### Human-Readable Help

```bash
# Show general help
kubectl mtv help

# Get help for a specific command
kubectl mtv help get plan
kubectl mtv help create provider
```

#### Help Topics

Built-in reference topics are available for domain-specific languages:

| **Topic** | **Description** |
|-----------|-----------------|
| `tsl` | Tree Search Language (TSL) query syntax reference |
| `karl` | Kubernetes Affinity Rule Language (KARL) syntax reference |

```bash
# Learn about the TSL query language
kubectl mtv help tsl

# Learn about the KARL affinity syntax
kubectl mtv help karl
```

#### Machine-Readable Output

Use `--machine` to output command schemas in JSON or YAML for integration with
MCP servers, automation tools, and AI assistants.

| **Flag** | **Type** | **Default** | **Description** |
|----------|----------|-------------|-----------------|
| `--machine` | bool | false | Enable machine-readable output |
| `--short` | bool | false | Omit long descriptions and examples (with `--machine`) |
| `-o`, `--output` | string | `json` | Output format: `json` or `yaml` |
| `--read-only` | bool | false | Include only read-only commands |
| `--write` | bool | false | Include only write commands |
| `--include-global-flags` | bool | true | Include global flags in output |

```bash
# Output complete command schema as JSON
kubectl mtv help --machine

# Output schema for a single command
kubectl mtv help --machine get plan

# Output schema for all "get" commands
kubectl mtv help --machine get

# Condensed schema without long descriptions or examples
kubectl mtv help --machine --short

# Output in YAML format
kubectl mtv help --machine --output yaml

# Only read-only commands
kubectl mtv help --machine --read-only

# Get TSL reference in machine-readable format
kubectl mtv help --machine tsl
```

## Query Language (TSL) Syntax

> **Full Reference**: See [Chapter 27: TSL - Tree Search Language Reference](../27-tsl-tree-search-language-reference) for the complete TSL reference.

The [Tree Search Language (TSL)](../27-tsl-tree-search-language-reference) is used with inventory commands for advanced filtering:

### Basic Syntax

```
kubectl mtv get inventory vms --provider <provider> --query "where <condition>"
```

### Operators

| **Operator** | **Description** | **Example** |
|--------------|-----------------|-------------|
| `=` | Equal | `name = 'vm-01'` |
| `!=`, `<>` | Not equal | `powerState != 'poweredOff'` |
| `>`, `>=` | Greater than (or equal) | `memoryMB > 4096` |
| `<`, `<=` | Less than (or equal) | `cpuCount <= 2` |
| `+`, `-`, `*`, `/`, `%` | Arithmetic | `memoryMB / 1024 > 4` |
| `like` | Pattern match (case-sensitive, `%` wildcard) | `name like 'web-%'` |
| `ilike` | Pattern match (case-insensitive) | `name ilike 'WEB-%'` |
| `~=` | Regex match | `name ~= 'prod-.*'` |
| `~!` | Regex not match | `name ~! 'test-.*'` |
| `in` | In list | `powerState in ['poweredOn', 'suspended']` |
| `not in` | Not in list | `powerState not in ['poweredOff']` |
| `between` | Range | `memoryMB between 4096 and 16384` |
| `is null` | Null check | `description is null` |
| `is not null` | Not null check | `guestIP is not null` |
| `and` | Logical AND | `powerState = 'poweredOn' and memoryMB > 2048` |
| `or` | Logical OR | `name like 'web-%' or name like 'app-%'` |
| `not` | Logical NOT | `not powerState = 'poweredOff'` |

### Functions

| **Function** | **Description** | **Example** |
|--------------|-----------------|-------------|
| `len(field)` | Length of array | `len(disks) > 1` |
| `sum(field[*].sub)` | Sum of numeric values | `sum(disks[*].capacityInBytes) > 107374182400` |
| `any(field[*].sub)` | Any element matches | `any(concerns[*].category = 'Critical')` |
| `all(field[*].sub)` | All elements match | `all(disks[*].capacityGB >= 20)` |

### Array Access and SI Units

| **Feature** | **Syntax** | **Example** |
|-------------|------------|-------------|
| Index access | `field[N]` | `disks[0].capacityGB > 100` |
| Wildcard | `field[*].sub` | `disks[*].datastore.id` |
| SI units | `Ki, Mi, Gi, Ti, Pi` | `memory > 4Gi` (= 4294967296) |

### Query Examples

```bash
# Powered-on VMs with more than 4GB RAM
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState = 'poweredOn' and memoryMB > 4096"

# VMs with names starting with "prod-"
kubectl mtv get inventory vms --provider vsphere-prod --query "where name like 'prod-%'"

# VMs with multiple disks and large total capacity
kubectl mtv get inventory vms --provider vsphere-prod --query "where len(disks) > 1 and sum(disks[*].capacityInBytes) > 53687091200"

# VMs in specific power states
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState in ['poweredOn', 'suspended']"

# Complex query with multiple conditions
kubectl mtv get inventory vms --provider vsphere-prod --query "where (name like 'web-%' or name like 'app-%') and powerState = 'poweredOn' and memoryMB between 2048 and 8192"

# Regex match and not-match
kubectl mtv get inventory vms --provider vsphere-prod --query "where name ~= 'prod-.*' and name ~! 'test-.*'"
```

## KARL Syntax (Kubernetes Affinity Rule Language)

> **Full Reference**: See [Chapter 28: KARL - Kubernetes Affinity Rule Language Reference](../28-karl-kubernetes-affinity-rule-language-reference) for the complete KARL reference.

[KARL](../28-karl-kubernetes-affinity-rule-language-reference) is used for advanced scheduling with `--target-affinity` and `--convertor-affinity` flags:

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
kubectl mtv create provider --name vsphere-source --type vsphere --url https://vcenter.company.com --username admin --password pass123
kubectl mtv create provider --name openshift-target --type openshift --url https://api.openshift.company.com:6443 --provider-token sha256~token

# 2. Create mappings (optional)
kubectl mtv create mapping network --name net-mapping --source vsphere-source --target openshift-target --network-pairs "VLAN100:prod-network"
kubectl mtv create mapping storage --name stor-mapping --source vsphere-source --target openshift-target --storage-pairs "datastore1:premium-ssd"

# 3. Create migration plan
kubectl mtv create plan --name my-migration --source vsphere-source --target openshift-target --vms vm1,vm2,vm3 --migration-type warm

# 4. Start migration
kubectl mtv start plan --name my-migration

# 5. Monitor progress
kubectl mtv get plan --name my-migration --watch

# 6. Complete warm migration (if applicable)
kubectl mtv cutover plan --name my-migration
```

### Inventory Exploration

```bash
# List all VMs
kubectl mtv get inventory vms --provider vsphere-source

# Filter and explore
kubectl mtv get inventory vms --provider vsphere-source --query "where powerState = 'poweredOn'"
kubectl mtv get inventory vms --provider vsphere-source --query "where memory.size > 4096" --output json

# Export for planning
kubectl mtv get inventory vms --provider vsphere-source --output planvms > migration-candidates.yaml
```

### Troubleshooting

```bash
# Verbose output for debugging
kubectl mtv get plan --name my-migration -v=2

# Detailed resource information
kubectl mtv describe plan --name my-migration
kubectl mtv describe provider --name vsphere-source

# Check inventory connectivity
kubectl mtv get inventory vms --provider vsphere-source -v=3
```

---

*Previous: [Chapter 25: Settings Management](../25-settings-management)*  
*Next: [Chapter 27: TSL - Tree Search Language Reference](../27-tsl-tree-search-language-reference)*
