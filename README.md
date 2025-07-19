# kubectl-mtv

A kubectl plugin that helps users of Forklift migrate virtualization workloads from oVirt, VMware, OpenStack, and OVA files to KubeVirt on Kubernetes.

<p align="center">
  <img src="docs/hiking.svg" alt="kubectl-mtv logo" width="200">
</p>

## Overview

The Forklift project (upstream of Migration Toolkit for Virtualization) simplifies the process of migrating virtual machines from traditional virtualization platforms to Kubernetes using KubeVirt. It handles the complexities of different virtualization platforms and provides a consistent way to define, plan, and execute migrations.

[Forklift GitHub Repository](https://github.com/kubev2v/forklift)

> **Note**  
> The Migration Toolkit for Virtualization (MTV) is the downstream (OpenShift) distribution of the upstream Kubernetes Forklift project.
> Similarly, the `oc` CLI is the downstream (OpenShift) version of the upstream `kubectl` CLI.
> This project (`kubectl-mtv`) is compatible with both upstream (Kubernetes/Forklift/kubectl) and downstream (OpenShift/MTV/oc) environments, but documentation here uses upstream naming unless otherwise noted.

## Installation

### Krew plugin manager

Using [krew](https://sigs.k8s.io/krew) plugin manager to install:

``` bash
# Available for linux-amd64
kubectl krew install mtv
kubectl mtv --help
```

### Download release binaries

Go to the [Releases page](https://github.com/yaacov/kubectl-mtv/releases) and download the appropriate archive for your platform.

For additional installation methods and detailed setup instructions, see the [Installation Guide](docs/README-install.md).

## Usage

For a complete migration demo, see the [Migration Demo Tutorial](docs/README_demo.md).

### Quick Start Examples

#### Create a VMware provider

```bash
kubectl mtv create provider vsphere-01 --type vsphere --url https://vcenter.example.com \
  -u admin --password secret --cacert @ca.cert
```

#### List VMs from a provider

```bash
kubectl mtv get inventory vms vsphere-01
```

#### Create a migration plan

```bash
kubectl mtv create plan my-plan --source vsphere-01 --target openshift-target \
  --vms "web-vm-1,db-vm-2,app-vm-3"
```

#### Start migration

```bash
kubectl mtv start plan my-plan
```

#### Monitor migration progress

```bash
kubectl mtv get plan --watch
```

For comprehensive usage instructions and detailed command reference, see the [Usage Guide](docs/README-usage.md).

## Key Features

- **Provider Management**: Connect to VMware vSphere, OpenShift, oVirt, OpenStack, and OVA providers
- **Inventory Exploration**: Query and filter VMs, networks, storage, and hosts with advanced search capabilities
- **Migration Planning**: Create and manage migration plans with automatic network and storage mapping
- **VDDK Integration**: Support for VMware Virtual Disk Development Kit for optimal disk transfer performance
- **Live Monitoring**: Watch migration progress with real-time status updates

For detailed feature documentation, see:

- [Inventory Commands Tutorial](docs/README_inventory.md): Advanced querying and filtering
- [planvms VM List Editing](docs/README_planvms.md): Customizing VM migration configurations
- [VDDK Image Creation and Usage](docs/README_vddk.md): Setting up VMware VDDK support

## Environment Variables

The following environment variables are used by `kubectl-mtv`:

- `MTV_VDDK_INIT_IMAGE`: Specifies the default Virtual Disk Development Kit (VDDK) container init image path. This value is used as the default for the `--vddk-init-image` flag when creating a provider.
- `MTV_INVENTORY_URL`: Specifies the base URL for the inventory service. This value is used as the default for the `--inventory-url` flag in various commands, such as listing providers, VMs, networks, and storage.

## Tutorials

- [Migration Demo Tutorial](docs/README_demo.md): Step-by-step guide to performing a VM migration
- [Inventory Commands Tutorial](docs/README_inventory.md): Comprehensive guide to using inventory commands and queries

## Documentation

- [Installation Guide](docs/README-install.md)
- [Usage Guide](docs/README-usage.md)
- [Development Guide](docs/README-development.md)

## Logo Attribution

The gopher logo is from [github.com/egonelbre/gophers](https://github.com/egonelbre/gophers) by Renee French.

## License

Apache-2.0 license
