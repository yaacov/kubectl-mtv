# kubectl-mtv

A kubectl plugin for migrating virtual machines to KubeVirt using Forklift.

<p align="center">
  <img src="docs/hiking.svg" alt="kubectl-mtv logo" width="200">
</p>

## Overview

kubectl-mtv helps migrate VMs from vSphere, oVirt, OpenStack, EC2, and OVA to Kubernetes/OpenShift using KubeVirt. It's a command-line interface for the [Forklift](https://github.com/kubev2v/forklift) project.

## Installation

```bash
# Using krew
kubectl krew install mtv

# Or download from releases
# https://github.com/yaacov/kubectl-mtv/releases
```

See [Installation Guide](guide/02-installation-and-prerequisites.md) for more options.

## MCP Support

kubectl-mtv includes a built-in MCP (Model Context Protocol) server for AI agents that support MCP add‑ons, such as Cursor IDE and Claude Desktop.

See [MCP Server Guide](guide/22-model-context-protocol-mcp-server-integration.md) for detailed setup instructions and usage examples.

## Quick Start

### 1. Create Provider

```bash
# vSphere
kubectl mtv create provider --name vsphere-01 --type vsphere \
  --url https://vcenter.example.com \
  --username admin --password secret --cacert @ca.cert
```

### 2. Create Migration Plan

Network and storage mappings are created automatically with sensible defaults.
Use `--network-pairs` / `--storage-pairs` to override inline if needed.

```bash
# Using system defaults for best network and storage mapping
kubectl mtv create plan --name migration-1 \
  --source vsphere-01 \
  --vms vm1,vm2,vm3

# Overriding mappings inline
kubectl mtv create plan --name migration-1 \
  --source vsphere-01 \
  --vms vm1,vm2,vm3 \
  --network-pairs "VM Network:default" \
  --storage-pairs "datastore1:standard"
```

### 3. Start Migration

```bash
kubectl mtv start plan --name migration-1
```

### 4. Monitor Progress

```bash
# Interactive TUI with scrolling, help panel, and adjustable refresh
kubectl mtv get plans --vms --watch
```

### Advanced: Reusable Mappings

If you need to reuse the same network/storage configuration across multiple plans,
create named mappings and reference them:

```bash
# Network mapping
kubectl mtv create mapping network --name prod-net \
  --source vsphere-01 --target openshift \
  --network-pairs "VM Network:default,Management:openshift-sdn/mgmt"

# Storage mapping with enhanced features
kubectl mtv create mapping storage --name prod-storage \
  --source vsphere-01 --target openshift \
  --storage-pairs "datastore1:standard;volumeMode=Block;accessMode=ReadWriteOnce,datastore2:fast;volumeMode=Filesystem" \
  --default-offload-plugin vsphere --default-offload-vendor flashsystem

# Reference them in a plan
kubectl mtv create plan --name migration-1 \
  --source vsphere-01 \
  --network-mapping prod-net \
  --storage-mapping prod-storage \
  --vms vm1,vm2,vm3
```

For a complete walkthrough, see the [Quick Start Guide](guide/03-quick-start-first-migration-workflow.md).

## Inventory Management

Query and explore provider resources before migration:

```bash
# List VMs
kubectl mtv get inventory vms --provider vsphere-01

# Filter VMs by criteria
kubectl mtv get inventory vms --provider vsphere-01 --query "where memoryMB > 4096"

# List networks and storage
kubectl mtv get inventory networks --provider vsphere-01
kubectl mtv get inventory storages --provider vsphere-01
```

See [Inventory Management Guide](guide/09-inventory-management.md) for advanced queries and filtering.

## VDDK Support

For optimal VMware disk transfer performance, build a VDDK image from VMware's VDDK SDK:

```bash
# Build VDDK image
kubectl mtv create vddk-image \
  --tar VMware-vix-disklib-8.0.1.tar.gz \
  --tag quay.io/myorg/vddk:8.0.1

# Use it when creating a provider
kubectl mtv create provider --name vsphere-01 --type vsphere \
  --url https://vcenter.example.com \
  --vddk-init-image quay.io/myorg/vddk:8.0.1
```

See [VDDK Setup Guide](guide/08-vddk-image-creation-and-configuration.md) for detailed instructions.

## Help and Reference Topics

The built-in help system includes machine-readable output and reference topics for domain-specific query languages:

```bash
# Get help for any command
kubectl mtv help create plan

# Learn the TSL query language or KARL affinity syntax
kubectl mtv help tsl
kubectl mtv help karl

# Machine-readable command schema (JSON/YAML) for automation and AI agents
kubectl mtv help --machine
kubectl mtv help --machine --short get plan
```

See [Command Reference](guide/26-command-reference.md) for the full help command documentation.

## Features

- **Multi-Platform Support**: Migrate from vSphere, oVirt, OpenStack, EC2, and OVA
- **Auto-Mapping**: Automatic network and storage mapping for all source providers
- **Flexible Mapping**: Use existing mappings, inline pairs, or automatic defaults
- **Advanced Queries**: Filter and search inventory with powerful query language
- **VDDK Support**: Optimized VMware disk transfers
- **Real-time Monitoring**: Track migration progress live
- **Timezone-Aware Display**: View timestamps in local time or UTC with `--use-utc` flag
- **System Health Checks**: Comprehensive health diagnostics for the MTV/Forklift system with actionable recommendations
- **Settings Management**: View and configure ForkliftController settings (feature flags, performance tuning, resource limits)
- **Machine-Readable Help**: Full command schema available as JSON/YAML for automation, MCP servers, and AI agents

## Documentation

**[Complete Technical Guide](guide/)** - Comprehensive documentation covering all features and use cases

### Quick Links

- [Installation & Prerequisites](guide/02-installation-and-prerequisites.md)
- [Quick Start Tutorial](guide/03-quick-start-first-migration-workflow.md)
- [Provider Management](guide/06-provider-management.md)
- [Inventory Management](guide/09-inventory-management.md)
- [Mapping Management](guide/11-mapping-management.md)
- [Migration Plan Creation](guide/13-migration-plan-creation.md)
- [Migration Hooks](guide/17-migration-hooks.md)
- [MCP Server Integration](guide/22-model-context-protocol-mcp-server-integration.md)
- [Command Reference](guide/26-command-reference.md)

## Environment Variables

- `MTV_VDDK_INIT_IMAGE`: Default VDDK init image for VMware providers
- `MTV_INVENTORY_URL`: Base URL for inventory service
- `MTV_INVENTORY_INSECURE_SKIP_TLS`: Skip TLS verification for inventory service connections (set to "true" to enable)

## License

Apache-2.0
