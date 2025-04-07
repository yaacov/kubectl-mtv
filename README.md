# kubectl-mtv - Migration Toolkit for Virtualization

A kubectl plugin that helps users migrate virtualization workloads from oVirt, VMware, OpenStack, and OVA files to KubeVirt on Kubernetes.

## Overview

The Migration Toolkit for Virtualization (MTV) simplifies the process of migrating virtual machines from traditional virtualization platforms to Kubernetes using KubeVirt. It handles the complexities of different virtualization platforms and provides a consistent way to define, plan, and execute migrations.

## Installation

### Prerequisites

- Kubernetes cluster with MTV installed
- kubectl installed and configured
- Go 1.23+

### Building and Installing

```bash
# Clone the repository
git clone https://github.com/yaacov/kubectl-mtv.git
cd kubectl-mtv

# Build and install the plugin
make install
```

Make sure `$GOPATH/bin` is in your PATH to use as a kubectl plugin.

## Custom Resources

MTV uses several Custom Resource Definitions (CRDs) to manage the migration process:

### Providers

Providers define connections to virtualization platforms. This includes both source platforms (where VMs currently run) and target platforms (KubeVirt on Kubernetes).

#### Common Provider Flags

All provider creation commands support these flags:

```
--type          Required. Provider type (openshift, vsphere, ova)
--name          Required. Provider name
--namespace     Namespace for the provider (defaults to current namespace)
--secret        Optional. Name of an existing secret containing provider credentials
--url           Provider URL (required for vsphere and when providing token for openshift)
```

#### Provider Type-Specific Requirements:

1. **vSphere Provider**
   ```bash
   kubectl mtv provider create --type vsphere --name my-vsphere \
     --url "https://vcenter.example.com/sdk" \
     --username administrator@vsphere.local --password VMware123! \
     --cacert @ca.pem [--provider-insecure-skip-tls] [--vddk-init-image IMAGE_PATH]
   ```
   
   Required flags:
   - `--url`: vCenter Server URL
   - Either provide (`--username` and `--password`) OR use an existing secret with `--secret`
   - Either provide `--cacert` OR use `--provider-insecure-skip-tls`
   
   Optional flags:
   - `--vddk-init-image`: Path to Virtual Disk Development Kit container init image

2. **OVA Provider**
   ```bash
   kubectl mtv provider create --type ova --name my-ova \
     --url "https://ova-provider.example.com"
   ```
   
   Required flags:
   - `--url`: OVA Provider URL
   - `--name`: Provider name

3. **OpenShift Provider**
   ```bash
   # Creating with token (requires URL)
   kubectl mtv provider create --type openshift --name my-openshift \
     --url "https://api.openshift.example.com:6443" --token TOKEN_STRING
   
   # Creating without credentials (using service account)
   kubectl mtv provider create --type openshift --name my-openshift
   ```
   
   Required flags:
   - `--name`: Provider name
   
   Conditional requirements:
   - When using `--token`, you must also provide `--url`

#### Listing and Deleting Providers

```bash
# List providers
kubectl mtv provider list [--output json|table] [--inventory-url URL]

# Delete a provider
kubectl mtv provider delete --name PROVIDER_NAME
```

### Network and Storage Mappings

Mappings define how to translate networks and storage from the source platform to the target platform.

```bash
# Create network mapping
kubectl mtv mapping create-network --name network-map-1 --source my-ovirt --target my-kubevirt --from-file network-mapping.yaml

# Create storage mapping
kubectl mtv mapping create-storage --name storage-map-1 --source my-ovirt --target my-kubevirt --from-file storage-mapping.yaml

# List mappings
kubectl mtv mapping list [--type network|storage|all]
```

### Plans

Plans define which VMs to migrate, using which providers and mappings.

```bash
# Create a migration plan
kubectl mtv plan create --name migration-plan-1 --source my-ovirt --target my-kubevirt --network-mapping network-map-1 --storage-mapping storage-map-1 --vms vm-123,vm-456

# List plans
kubectl mtv plan list

# Start a plan
kubectl mtv plan start --name migration-plan-1

# Describe a plan
kubectl mtv plan describe --name migration-plan-1
```

## Inventory API

MTV maintains an inventory API that allows users to fetch information about the current inventory of providers.

```bash
# Common flags for inventory commands
# --provider       Required. Provider name to query
# --inventory-url  Optional. Base URL for inventory API (auto-discovered if omitted)
# --output,-o      Optional. Output format: table or json (default: table)
# --extended       Optional. Show extended information in table output
# --query          Optional. Query string to filter results

# List VMs from a provider
kubectl mtv inventory vms --provider my-ovirt

# List networks from a provider
kubectl mtv inventory networks --provider my-ovirt

# List storage from a provider
kubectl mtv inventory storage --provider my-ovirt

# List hosts from a provider
kubectl mtv inventory hosts --provider my-ovirt
```

### Query Syntax for Inventory Commands

The `--query` flag supports advanced filtering with SQL-like syntax:

```
SELECT field1, field2 AS alias, field3  # Select specific fields with optional aliases
WHERE condition                         # Filter using tree-search-language conditions
ORDER BY field1 [ASC|DESC], field2      # Sort results on multiple fields
LIMIT n                                 # Limit number of results
```

Example: `--query "SELECT name, memory WHERE memory > 1024 ORDER BY name LIMIT 10"`

## Migration Workflow

1. Create source and target providers
2. Create network and storage mappings
3. Create a migration plan
4. Start the plan
5. Monitor the migration

## Examples

### Complete Migration Example

```bash
# Create providers
kubectl mtv provider create --type vsphere --name vsphere-source \
  --url "https://vcenter.example.com/sdk" \
  --username administrator@vsphere.local --password VMware123! \
  --provider-insecure-skip-tls

kubectl mtv provider create --type openshift --name ocp-target \
  --url "https://api.ocp.example.com:6443" --token kubeadmin-token

# View available VMs in source provider
kubectl mtv inventory vms --provider vsphere-source

# Create mappings
kubectl mtv mapping create-network --name network-map --source vsphere-source --target ocp-target --from-file network-map.yaml
kubectl mtv mapping create-storage --name storage-map --source vsphere-source --target ocp-target --from-file storage-map.yaml

# Create and start migration plan
kubectl mtv plan create --name migration-plan-1 --source vsphere-source --target ocp-target --network-mapping network-map --storage-mapping storage-map --vms vm-123,vm-456
kubectl mtv plan start --name migration-plan-1

# Monitor migration
kubectl mtv migration list
kubectl mtv migration describe --name migration-plan-1-migration
```

## License

Apache-2.0 license 