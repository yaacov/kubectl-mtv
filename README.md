# kubectl-mtv

A kubectl plugin for migrating virtual machines to KubeVirt using Forklift.

<p align="center">
  <img src="docs/hiking.svg" alt="kubectl-mtv logo" width="200">
</p>

## Overview

kubectl-mtv helps migrate VMs from vSphere, oVirt, OpenStack, and OVA to Kubernetes/OpenShift using KubeVirt. It's a command-line interface for the [Forklift](https://github.com/kubev2v/forklift) project.

## Installation

```bash
# Using krew
kubectl krew install mtv

# Or download from releases
# https://github.com/yaacov/kubectl-mtv/releases
```

See [Installation Guide](docs/README-install.md) for more options.

## Quick Start

### 1. Create Provider

```bash
kubectl mtv create provider vsphere-01 --type vsphere \
  --url https://vcenter.example.com \
  -u admin --password secret --cacert @ca.cert
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
  --default-offload-plugin vsphere --default-offload-vendor vantara
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
kubectl mtv get plan --watch
```

For a complete walkthrough, see the [Migration Demo Tutorial](docs/README_demo.md).

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

See [Inventory Commands Tutorial](docs/README_inventory.md) for advanced queries and filtering.

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

See [VDDK Setup Guide](docs/README_vddk.md) for detailed instructions.

## Features

- **Multi-Platform Support**: Migrate from vSphere, oVirt, OpenStack, and OVA
- **Flexible Mapping**: Use existing mappings, inline pairs, or automatic defaults
- **Advanced Queries**: Filter and search inventory with powerful query language
- **VDDK Support**: Optimized VMware disk transfers
- **Real-time Monitoring**: Track migration progress live
- **Timezone-Aware Display**: View timestamps in local time or UTC with `--use-utc` flag

## Documentation

- [Installation Guide](docs/README-install.md)
- [Usage Guide](docs/README-usage.md)
- [Migration Demo](docs/README_demo.md)
- [Migration Hooks](docs/README_hooks.md)
- [Inventory Queries](docs/README_inventory.md)
- [Mapping Configuration](docs/README_mapping_pairs.md)
- [Creating Mappings](docs/README_create_mappings.md)
- [Patching Mappings](docs/README_patch_mappings.md)
- [Patching Providers](docs/README_patch_providers.md)
- [Patching Plans](docs/README_patch_plans.md)
- [Creating Migration Hosts](docs/README_host_creation.md)
- [Development Guide](docs/README-development.md)

## Environment Variables

- `MTV_VDDK_INIT_IMAGE`: Default VDDK init image for VMware providers
- `MTV_INVENTORY_URL`: Base URL for inventory service

## License

Apache-2.0