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

## Table of Contents

- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Krew plugin manager](#krew-plugin-manager)
  - [Download release binaries](#download-release-binaries)
  - [Installing from DNF (Fedora)](#installing-from-dnf-fedora)
  - [Building and Installing](#building-and-installing)
- [Usage](#usage)
  - [Global Flags](#global-flags)
  - [Provider Management](#provider-management)
  - [Mapping Management](#mapping-management)
  - [Inventory Management](#inventory-management)
  - [Migration Plan Management](#migration-plan-management)
  - [VDDK Image Management](#vddk-image-management)
- [Environment Variables](#environment-variables)
- [Tutorials](#tutorials)
- [planvms VM List Editing](./README_planvms.md)
- [Logo Attribution](#logo-attribution)
- [License](#license)

## Installation

### Prerequisites

- Kubernetes cluster with Forklift or MTV installed
- kubectl installed and configured
- Go 1.23+

### Krew plugin manager

Using [krew](https://sigs.k8s.io/krew) plugin manager to install:

``` bash
# Available for linux-amd64
kubectl krew install mtv
kubectl mtv --help
```

### Download release binaries

Go to the [Releases page](https://github.com/yaacov/kubectl-mtv/releases) and download the appropriate archive for your platform.

Or, use the following commands to download and extract the latest release:

```bash
REPO=yaacov/kubectl-mtv
ASSET=kubectl-mtv.tar.gz
LATEST_VER=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep -m1 '"tag_name"' | cut -d'"' -f4)
curl -L -o $ASSET https://github.com/$REPO/releases/download/$LATEST_VER/$ASSET
tar -xzf $ASSET
```

### Installing from DNF (Fedora)

On Fedora 41, 42, and other compatible amd64 systems, you can install kubectl-mtv directly from the COPR repository:

```bash
# Enable the COPR repository
dnf copr enable yaacov/kubesql

# Install kubectl-mtv
dnf install kubectl-mtv
```

### Building and Installing

```bash
# Clone the repository
git clone https://github.com/yaacov/kubectl-mtv.git
cd kubectl-mtv

# Build the plugin
make
```

Make sure `$GOPATH/bin` is in your PATH to use as a kubectl plugin.

## Usage

See the demo documentation for a migration [demo flow using kubectl-mtv](./README_demo.md).

### Global Flags

These flags are available for all commands:

```bash
--kubeconfig string      Path to the kubeconfig file
--context string         The name of the kubeconfig context to use
--namespace string       Namespace (defaults to active namespace from kubeconfig)
```

### Provider Management

#### Create Provider

Create a provider connection to a virtualization platform.

```bash
kubectl mtv create provider NAME --type TYPE [flags]
```

**Required Flags:**

- `--type`: Provider type (openshift, vsphere, ovirt, openstack, ova)

**Optional Flags:**

- `--secret`: Secret containing provider credentials
- `-U, --url`: Provider URL
- `-u, --username`: Provider credentials username
- `-p, --password`: Provider credentials password
- `-T, --token`: Provider authentication token (used for openshift provider)
- `--cacert`: Provider CA certificate (use @filename to load from file)
- `--provider-insecure-skip-tls`: Skip TLS verification when connecting to the provider
- `--vddk-init-image`: Virtual Disk Development Kit (VDDK) container init image path

**Examples:**

```bash
# Create a VMware provider
kubectl mtv create provider vsphere-01 --type vsphere --url https://vcenter.example.com \
  -u admin --password secret --cacert @ca.cert

# Create an OpenShift provider
kubectl mtv create provider openshift-target --type openshift \
  --url https://api.cluster.example.com:6443 --token eyJhbGc...
```

#### List Providers

List all providers in a namespace.

```bash
kubectl mtv get provider [flags]
```

**Optional Flags:**

- `-i, --inventory-url`: Base URL for the inventory service
- `-o, --output`: Output format. One of: table, json (default "table")

#### Delete Provider

Delete a provider.

```bash
kubectl mtv delete provider NAME [flags]
```

### Mapping Management

Mappings define how resources from source providers are mapped to target providers.

> **Note**  
> While `kubectl-mtv` provide manual mapping management, mapping is handled automatically by
> the `kubectl-mtv` tool. When using `kubectl-mtv` users will not create mappings manually,
> during the creation of a migration plan, `kubectl-mtv` will check the network and storage
> resources used by the virtual machines in the source cluster, and will create a "best parctice"
> network and storage mappings, users can review and edit the mappings while customizing
> the migration plan, before executing it.

#### Create Network Mapping

Create a network mapping between source and target providers.

```bash
kubectl mtv create mapping NAME --type network [flags]
```

**Required Flags:**

- `--type`: Mapping type (network, storage)

**Optional Flags:**

- `-S, --source`: Source provider name
- `-T, --target`: Target provider name
- `-f, --from-file`: Create mapping from YAML/JSON file

#### Create Storage Mapping

Create a storage mapping between source and target providers.

```bash
kubectl mtv create mapping NAME --type storage [flags]
```

**Required Flags:**

- `--type`: Mapping type (network, storage)

**Optional Flags:**

- `-S, --source`: Source provider name
- `-T, --target`: Target provider name
- `-f, --from-file`: Create mapping from YAML/JSON file

#### List Mappings

List all mappings in a namespace.

```bash
kubectl mtv get mapping [flags]
```

**Optional Flags:**

- `--type`: Mapping type (network, storage, all) (default "all")
- `-o, --output`: Output format. One of: table, json (default "table")

#### Delete Mapping

Delete a mapping by name.

```bash
kubectl mtv delete mapping NAME [flags]
```

**Optional Flags:**

- `--type`: Mapping type (network, storage)

### Inventory Management

Query and explore the inventory of providers.

#### List VMs

List VMs from a provider.

```bash
kubectl mtv get inventory vms PROVIDER [flags]
```

**Optional Flags:**

- `-i, --inventory-url`: Base URL for the inventory service
- `-o, --output`: Output format. One of: table, json, planvms (default "table")
- `-e, --extended`: Show extended information in table output
- `-q, --query`: Query string with 'where', 'order by', and 'limit' clauses

**Query Syntax:**

- `SELECT field1, field2 AS alias, field3`: Select specific fields with optional aliases
- `WHERE condition`: Filter using tree-search-language conditions
- `ORDER BY field1 [ASC|DESC], field2`: Sort results on multiple fields
- `LIMIT n`: Limit number of results

**Examples:**

```bash
# List all VMs
kubectl mtv get inventory vms vsphere-01

# List VMs with a specific query
kubectl mtv get inventory vms vsphere-01 -q "WHERE name LIKE 'db-%' ORDER BY memory DESC LIMIT 10"

# Output VM list in a format suitable for migration plans
kubectl mtv get inventory vms vsphere-01 -o planvms > vms.yaml
```

#### List Networks

List networks from a provider.

```bash
kubectl mtv get inventory networks PROVIDER [flags]
```

**Optional Flags:**

- `-i, --inventory-url`: Base URL for the inventory service
- `-o, --output`: Output format. One of: table, json (default "table")
- `-e, --extended`: Show extended information in table output
- `-q, --query`: Query string with 'where', 'order by', and 'limit' clauses

#### List Storage

List storage from a provider.

```bash
kubectl mtv get inventory storage PROVIDER [flags]
```

**Optional Flags:**

- `-i, --inventory-url`: Base URL for the inventory service
- `-o, --output`: Output format. One of: table, json (default "table")
- `-e, --extended`: Show extended information in table output
- `-q, --query`: Query string with 'where', 'order by', and 'limit' clauses

#### List Hosts

List hosts from a provider.

```bash
kubectl mtv get inventory hosts PROVIDER [flags]
```

**Optional Flags:**

- `-i, --inventory-url`: Base URL for the inventory service
- `-o, --output`: Output format. One of: table, json (default "table")
- `-e, --extended`: Show extended information in table output
- `-q, --query`: Query string with 'where', 'order by', and 'limit' clauses

#### List Namespaces

List namespaces from a provider.

```bash
kubectl mtv get inventory namespaces PROVIDER [flags]
```

**Optional Flags:**

- `-i, --inventory-url`: Base URL for the inventory service
- `-o, --output`: Output format. One of: table, json (default "table")
- `-q, --query`: Query string with 'where', 'order by', and 'limit' clauses

### Migration Plan Management

Create and manage migration plans.

#### Create Migration Plan

Create a migration plan to move VMs from a source provider to a target provider.

```bash
kubectl mtv create plan NAME [flags]
```

**Optional Flags:**

- `-S, --source`: Source provider name
- `-t, --target`: Target provider name
- `--network-mapping`: Network mapping name
- `--storage-mapping`: Storage mapping name
- `--vms`: Comma separated list of VM names (comma-separated) or path to YAML/JSON file containing a list of VM structs (prefix with @)
- `--description`: Plan description
- `--target-namespace`: Target namespace
- `--warm`: Whether this is a warm migration
- `--transfer-network`: Network attachment definition for disk transfer
- `--preserve-cluster-cpu-model`: Preserve the CPU model from the source cluster
- `--preserve-static-ips`: Preserve static IPs of VMs in vSphere
- `--pvc-name-template`: Template for generating PVC names for VM disks
- `--pvc-name-template-use-generate-name`: Use generateName instead of name for PVC name template (default true)
- `--volume-name-template`: Template for generating volume interface names
- `--network-name-template`: Template for generating network interface names
- `--migrate-shared-disks`: Determines if the plan should migrate shared disks (default true)
- `--archived`: Whether this plan should be archived
- `--disk-bus`: Disk bus type (deprecated: will be deprecated in 2.8)
- `--delete-guest-conversion-pod`: Delete guest conversion pod after successful migration
- `-i, --inventory-url`: Base URL for the inventory service

**Examples:**

```bash
# Create a plan with specific VMs
kubectl mtv create plan my-plan --source vsphere-01 --target openshift-target \
  --vms "web-vm-1,db-vm-2,app-vm-3"

# Create a plan with VMs defined in a file
kubectl mtv create plan my-plan --source vsphere-01 --target openshift-target \
  --vms @vms.yaml

# Create a warm migration plan with options for PVC naming
kubectl mtv create plan warm-plan --source vsphere-01 --target openshift-target \
  --vms "web-vm-1" --warm --pvc-name-template "{{.VmName}}-disk-{{.DiskIndex}}" \
  --pvc-name-template-use-generate-name=false
```

See [Editing the VMs List for Migration Plans (planvms)](./README_planvms.md) for details on the planvms file format and how to edit it before migration.

#### List Migration Plans

List migration plans in a namespace.

```bash
kubectl mtv get plan [flags]
```

**Optional Flags:**

- `--watch`: Watch migration plans with live updates
- `-o, --output`: Output format. One of: table, json (default "table")

#### Start Migration Plan

Start a migration plan execution.

```bash
kubectl mtv start plan NAME [flags]
```

Examples:

```bash
# Cutover in 10m
kubectl mtv start plan demo --cutover $(date -d '+10 minutes' --rfc-3339=seconds)
```

```bash
# Cutover on the next round hour
kubectl mtv start plan demo --cutover $(date -d "$(date +'%Y-%m-%d %H:00:00') +1 hour" --rfc-3339=seconds)
```

**Optional Flags:**

- `--cutover`: Cutover time in ISO8601 format (e.g., 2023-04-01T14:30:00Z) for warm migrations, if missing cutover is set to 1h from now.

#### Describe Migration Plan

Show detailed information about a migration plan.

```bash
kubectl mtv describe plan NAME
```

#### Describe VM in Migration Plan

Show detailed information about a specific VM in a migration plan.

```bash
kubectl mtv describe plan-vm NAME --vm VM_NAME [flags]
```

**Required Flags:**

- `--vm`: VM name to describe

**Optional Flags:**

- `-w, --watch`: Watch VM status with live updates

#### List VMs in Migration Plan

List all VMs in a migration plan with their migration status.

```bash
kubectl mtv plan vms NAME [flags]
```

**Optional Flags:**

- `-w, --watch`: Watch VM status with live updates

#### Cancel VMs in Migration Plan

Cancel specific VMs in a running migration.

```bash
kubectl mtv plan cancel-vms NAME --vms VMLIST [flags]
```

**Required Flags:**

- `--vms`: Comma separated list of VM names to cancel (comma-separated) or path to file containing VM names (prefix with @)

#### Set Cutover Time

Set the cutover time for a warm migration.

```bash
kubectl mtv plan cutover NAME [flags]
```

**Optional Flags:**

- `--cutover`: Cutover time in ISO8601 format. If not specified, current time will be used.

#### Delete Migration Plan

Delete a migration plan.

```bash
kubectl mtv plan delete NAME
```

#### Archive Migration Plan

Archive or unarchive a migration plan for long-term storage.

```bash
kubectl mtv plan archive NAME [flags]
```

**Optional Flags:**

- `--unarchive`: Unarchive the plan instead of archiving it

### VDDK Image Management

It is strongly recommended that Forklift (Migration Toolkit for Virtualization/MTV) should be used with the VMware Virtual Disk Development Kit (VDDK) SDK when transferring virtual disks from VMware vSphere.

> **Note**  
> Storing the VDDK image in a public registry might violate the VMware license terms.

#### VDDK Image Management prerequisites

- `podman` installed.
- You are working on a file system that preserves symbolic links (symlinks).
- If you are using an external registry, KubeVirt must be able to access it.

#### Creating and Using a VDDK Image

See [VDDK Image Creation and Usage](./README_vddk.md) for a step-by-step guide.

#### Command Example

```bash
kubectl mtv vddk create --tar ~/vmware-vix-disklib-distrib-8-0-1.tar.gz --tag quay.io/example/vddk:8
```

## Environment Variables

The following environment variables are used by `kubectl-mtv`:

- `MTV_VDDK_INIT_IMAGE`: Specifies the default Virtual Disk Development Kit (VDDK) container init image path. This value is used as the default for the `--vddk-init-image` flag when creating a provider.
- `MTV_INVENTORY_URL`: Specifies the base URL for the inventory service. This value is used as the default for the `--inventory-url` flag in various commands, such as listing providers, VMs, networks, and storage.

## Tutorials

- [Migration Demo Tutorial](./README_demo.md): Step-by-step guide to performing a VM migration
- [Inventory Commands Tutorial](./README_inventory.md): Comprehensive guide to using inventory commands and queries

## Logo Attribution

The gopher logo is from [github.com/egonelbre/gophers](https://github.com/egonelbre/gophers) by Renee French.

## License

Apache-2.0 license
