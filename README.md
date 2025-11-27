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

See [MCP Server Guide](guide/19-model-context-protocol-mcp-server-integration.md) for detailed setup instructions and usage examples.

## Quick Start

### 1. Create Provider

```bash
# vSphere
kubectl mtv create provider vsphere-01 --type vsphere \
  --url https://vcenter.example.com \
  -u admin --password secret --cacert @ca.cert

# EC2 (URL is optional, auto-generated from region)
kubectl mtv create provider ec2-01 --type ec2 \
  --region us-east-1 \
  --access-key-id AKIA... --secret-access-key secret
```

### 2. Create Mappings (Optional)

```bash
# Network mapping
kubectl mtv create mapping network prod-net \
  --source vsphere-01 --target openshift \
  --network-pairs "VM Network:default,Management:openshift-sdn/mgmt"

# Storage mapping with enhanced features
kubectl mtv create mapping storage prod-storage \
  --source vsphere-01 --target openshift \
  --storage-pairs "datastore1:standard;volumeMode=Block;accessMode=ReadWriteOnce,datastore2:fast;volumeMode=Filesystem" \
  --default-offload-plugin vsphere --default-offload-vendor flashsystem
```

### 3. Create Migration Plan

```bash
# Using system defaults for best network and storage mapping
kubectl mtv create plan migration-1 \
  --source vsphere-01 \
  --vms vm1,vm2,vm3

# Using existing mappings
kubectl mtv create plan migration-1 \
  --source vsphere-01 \
  --network-mapping prod-net \
  --storage-mapping prod-storage \
  --vms vm1,vm2,vm3
```

### 4. Start Migration

```bash
kubectl mtv start plan migration-1
```

### 5. Monitor Progress

```bash
# Interactive TUI with scrolling, help panel, and adjustable refresh
kubectl mtv get plan --watch
```

**NEW**: Watch mode now features an interactive Terminal UI with:
- Smooth screen updates without flickering
- Scrollable output (arrow keys, pgup/pgdn)
- Interactive help panel (press ?)
- Adjustable refresh interval (+/- keys)
- Manual refresh (press r)

For a complete walkthrough, see the [Quick Start Guide](guide/03-quick-start-first-migration-workflow.md).

## Inventory Management

Query and explore provider resources before migration:

```bash
# List VMs
kubectl mtv get inventory vms vsphere-01

# Filter VMs by criteria
kubectl mtv get inventory vms vsphere-01 -q "where memoryMB > 4096"

# List networks and storage
kubectl mtv get inventory networks vsphere-01
kubectl mtv get inventory storage vsphere-01
```

See [Inventory Management Guide](guide/07-inventory-management.md) for advanced queries and filtering.

## VDDK Support

For optimal VMware disk transfer performance, build a VDDK image from VMware's VDDK SDK:

```bash
# Build VDDK image
kubectl mtv create vddk-image \
  --tar VMware-vix-disklib-8.0.1.tar.gz \
  --tag quay.io/myorg/vddk:8.0.1

# Use it when creating a provider
kubectl mtv create provider vsphere-01 --type vsphere \
  --url https://vcenter.example.com \
  --vddk-init-image quay.io/myorg/vddk:8.0.1
```

See [VDDK Setup Guide](guide/06-vddk-image-creation-and-configuration.md) for detailed instructions.

## Features

- **Multi-Platform Support**: Migrate from vSphere, oVirt, OpenStack, EC2, and OVA
- **Auto-Mapping**: Automatic network and storage mapping for all source providers
- **Flexible Mapping**: Use existing mappings, inline pairs, or automatic defaults
- **Advanced Queries**: Filter and search inventory with powerful query language
- **VDDK Support**: Optimized VMware disk transfers
- **Real-time Monitoring**: Track migration progress live
- **Timezone-Aware Display**: View timestamps in local time or UTC with `--use-utc` flag

## Documentation

**[Complete Technical Guide](guide/)** - Comprehensive documentation covering all features and use cases

### Quick Links

- [Installation & Prerequisites](guide/02-installation-and-prerequisites.md)
- [Quick Start Tutorial](guide/03-quick-start-first-migration-workflow.md)
- [Provider Management](guide/04-provider-management.md)
- [Inventory Management](guide/07-inventory-management.md)
- [Mapping Management](guide/09-mapping-management.md)
- [Migration Plan Creation](guide/10-migration-plan-creation.md)
- [Migration Hooks](guide/14-migration-hooks.md)
- [MCP Server Integration](guide/19-model-context-protocol-mcp-server-integration.md)
- [Command Reference](guide/21-command-reference.md)

## Environment Variables

- `MTV_VDDK_INIT_IMAGE`: Default VDDK init image for VMware providers
- `MTV_INVENTORY_URL`: Base URL for inventory service
- `MTV_INVENTORY_INSECURE_SKIP_TLS`: Skip TLS verification for inventory service connections (set to "true" to enable)

## License

Apache-2.0
