# kubectl-mtv

A kubectl plugin that helps users migrate virtualization workloads from oVirt, VMware, OpenStack, and OVA files to KubeVirt on Kubernetes.

<p align="center">
  <img src="docs/hiking.svg" alt="kubectl-mtv logo" width="200">
</p>

## Overview

The Migration Toolkit for Virtualization (MTV) simplifies the process of migrating virtual machines from traditional virtualization platforms to Kubernetes using KubeVirt. It handles the complexities of different virtualization platforms and provides a consistent way to define, plan, and execute migrations.

[Migration Toolkit for Virtualization (MTV)](https://github.com/kubev2v/forklift)

## Table of Contents

- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Installing from DNF (Fedora)](#installing-from-dnf-fedora)
  - [Installing from Release Binaries](#installing-from-release-binaries)
  - [Building and Installing](#building-and-installing)
- [Usage](#usage)
  - [Global Flags](#global-flags)
  - [Provider Management](#provider-management)
  - [Mapping Management](#mapping-management)
  - [Inventory Management](#inventory-management)
  - [Migration Plan Management](#migration-plan-management)
- [Environment Variables](#environment-variables)
- [Logo Attribution](#logo-attribution)
- [License](#license)

## Installation

### Prerequisites

- Kubernetes cluster with MTV installed
- kubectl installed and configured
- Go 1.23+

### Installing from DNF (Fedora)

On Fedora 41, 42, and other compatible amd64 systems, you can install kubectl-mtv directly from the COPR repository:

```bash
# Enable the COPR repository
dnf copr enable yaacov/kubesql

# Install kubectl-mtv
dnf install kubectl-mtv
```

### Installing from Release Binaries

You can download and install pre-built binaries for linux/amd64:

```bash
# Set the version to download
VERSION=v0.1.7  # Replace with the desired version

# Download the release binary and its checksum
curl -LO https://github.com/yaacov/kubectl-mtv/releases/download/${VERSION}/kubectl-mtv.tar.gz
curl -LO https://github.com/yaacov/kubectl-mtv/releases/download/${VERSION}/kubectl-mtv.tar.gz.sha256sum

# Verify checksum
sha256sum kubectl-mtv.tar.gz
cat kubectl-mtv.tar.gz.sha256sum
# Compare the output of both commands to ensure they match

# Extract the binary
tar -xzf kubectl-mtv.tar.gz

# Install the binary (requires root privileges)
sudo mv kubectl-mtv /usr/local/bin/
sudo chmod +x /usr/local/bin/kubectl-mtv
```

### Building and Installing

```bash
# Clone the repository
git clone https://github.com/yaacov/kubectl-mtv.git
cd kubectl-mtv

# Build and install the plugin
make install
```

Make sure `$GOPATH/bin` is in your PATH to use as a kubectl plugin.

## Usage

### Global Flags

These flags are available for all commands:

```
--kubeconfig string      Path to the kubeconfig file
--context string         The name of the kubeconfig context to use
--namespace string       Namespace (defaults to active namespace from kubeconfig)
```

### Provider Management

#### Create Provider

Create a provider connection to a virtualization platform.

```bash
kubectl mtv provider create NAME --type TYPE [flags]
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
kubectl mtv provider create vsphere-01 --type vsphere --url https://vcenter.example.com \
  -u admin --password secret --cacert @ca.cert

# Create an OpenShift provider
kubectl mtv provider create openshift-target --type openshift \
  --url https://api.cluster.example.com:6443 --token eyJhbGc...
```

#### List Providers

List all providers in a namespace.

```bash
kubectl mtv provider list [flags]
```

**Optional Flags:**
- `-i, --inventory-url`: Base URL for the inventory service
- `-o, --output`: Output format. One of: table, json (default "table")

#### Delete Provider

Delete a provider.

```bash
kubectl mtv provider delete NAME [flags]
```

### Mapping Management

Mappings define how resources from source providers are mapped to target providers.

#### Create Network Mapping

Create a network mapping between source and target providers.

```bash
kubectl mtv mapping create-network NAME [flags]
```

**Optional Flags:**
- `-S, --source`: Source provider name
- `-t, --target`: Target provider name
- `--from-file`: Create from YAML file

#### Create Storage Mapping

Create a storage mapping between source and target providers.

```bash
kubectl mtv mapping create-storage NAME [flags]
```

**Optional Flags:**
- `-S, --source`: Source provider name
- `-t, --target`: Target provider name
- `--from-file`: Create from YAML file

#### List Mappings

List all mappings in a namespace.

```bash
kubectl mtv mapping list [flags]
```

**Optional Flags:**
- `--type`: Mapping type (network, storage, all) (default "all")
- `-o, --output`: Output format. One of: table, json (default "table")

### Inventory Management

Query and explore the inventory of providers.

#### List VMs

List VMs from a provider.

```bash
kubectl mtv inventory vms PROVIDER [flags]
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
kubectl mtv inventory vms vsphere-01

# List VMs with a specific query
kubectl mtv inventory vms vsphere-01 -q "WHERE name LIKE 'db-%' ORDER BY memory DESC LIMIT 10"

# Output VM list in a format suitable for migration plans
kubectl mtv inventory vms vsphere-01 -o planvms > vms.yaml
```

#### List Networks

List networks from a provider.

```bash
kubectl mtv inventory networks PROVIDER [flags]
```

**Optional Flags:**
- `-i, --inventory-url`: Base URL for the inventory service
- `-o, --output`: Output format. One of: table, json (default "table")
- `-e, --extended`: Show extended information in table output
- `-q, --query`: Query string with 'where', 'order by', and 'limit' clauses

#### List Storage

List storage from a provider.

```bash
kubectl mtv inventory storage PROVIDER [flags]
```

**Optional Flags:**
- `-i, --inventory-url`: Base URL for the inventory service
- `-o, --output`: Output format. One of: table, json (default "table")
- `-e, --extended`: Show extended information in table output
- `-q, --query`: Query string with 'where', 'order by', and 'limit' clauses

#### List Hosts

List hosts from a provider.

```bash
kubectl mtv inventory hosts PROVIDER [flags]
```

**Optional Flags:**
- `-i, --inventory-url`: Base URL for the inventory service
- `-o, --output`: Output format. One of: table, json (default "table")
- `-e, --extended`: Show extended information in table output
- `-q, --query`: Query string with 'where', 'order by', and 'limit' clauses

#### List Namespaces

List namespaces from a provider.

```bash
kubectl mtv inventory namespaces PROVIDER [flags]
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
kubectl mtv plan create NAME [flags]
```

**Optional Flags:**
- `-S, --source`: Source provider name
- `-t, --target`: Target provider name
- `--network-mapping`: Network mapping name
- `--storage-mapping`: Storage mapping name
- `--vms`: List of VM names (comma-separated) or path to YAML/JSON file containing a list of VM structs (prefix with @)
- `--description`: Plan description
- `--target-namespace`: Target namespace
- `--warm`: Whether this is a warm migration
- `--transfer-network`: Network attachment definition for disk transfer
- `--preserve-cluster-cpu-model`: Preserve the CPU model from the source cluster
- `--preserve-static-ips`: Preserve static IPs of VMs in vSphere
- `--pvc-name-template`: Template for generating PVC names for VM disks
- `--volume-name-template`: Template for generating volume interface names
- `--network-name-template`: Template for generating network interface names
- `--migrate-shared-disks`: Determines if the plan should migrate shared disks (default true)
- `-i, --inventory-url`: Base URL for the inventory service

**Examples:**

```bash
# Create a plan with specific VMs
kubectl mtv plan create my-plan --source vsphere-01 --target openshift-target \
  --vms "web-vm-1,db-vm-2,app-vm-3"

# Create a plan with VMs defined in a file
kubectl mtv plan create my-plan --source vsphere-01 --target openshift-target \
  --vms @vms.yaml
```

#### List Migration Plans

List migration plans in a namespace.

```bash
kubectl mtv plan list [flags]
```

**Optional Flags:**
- `--watch`: Watch migration plans with live updates
- `-o, --output`: Output format. One of: table, json (default "table")

#### Start Migration Plan

Start a migration plan execution.

```bash
kubectl mtv plan start NAME [flags]
```

**Optional Flags:**
- `--cutover`: Cutover time in RFC3339 format (e.g., 2023-04-01T14:30:00Z) for warm migrations, if missing cutover is set to 1h from now.

#### Describe Migration Plan

Show detailed information about a migration plan.

```bash
kubectl mtv plan describe NAME
```

#### Describe VM in Migration Plan

Show detailed information about a specific VM in a migration plan.

```bash
kubectl mtv plan vm NAME --vm VM_NAME [flags]
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
- `--vms`: List of VM names to cancel (comma-separated) or path to file containing VM names (prefix with @)

#### Set Cutover Time

Set the cutover time for a warm migration.

```bash
kubectl mtv plan cutover NAME [flags]
```

**Optional Flags:**
- `--cutover`: Cutover time in RFC3339 format. If not specified, current time will be used.

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

## Environment Variables

The following environment variables are used by `kubectl-mtv`:

- `MTV_VDDK_INIT_IMAGE`: Specifies the default Virtual Disk Development Kit (VDDK) container init image path. This value is used as the default for the `--vddk-init-image` flag when creating a provider.
- `MTV_INVENTORY_URL`: Specifies the base URL for the inventory service. This value is used as the default for the `--inventory-url` flag in various commands, such as listing providers, VMs, networks, and storage.

## Logo Attribution

The gopher logo is from [github.com/egonelbre/gophers](https://github.com/egonelbre/gophers) by Renee French.

## License

Apache-2.0 license
