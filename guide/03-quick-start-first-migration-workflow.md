---
layout: page
title: "Chapter 3: Quick Start - First Migration Workflow"
---

# Chapter 3: Quick Start - First Migration Workflow

This chapter provides a complete step-by-step walkthrough of your first migration using `kubectl-mtv`. We'll migrate a VM from VMware vSphere to KubeVirt, covering all essential steps from initial setup to completion monitoring.

## Prerequisites

Before starting this workflow, ensure you have:

- `kubectl-mtv` installed and working (see [Chapter 2](02-installation-and-prerequisites.md))
- Access to a Kubernetes cluster with Forklift/MTV installed
- Connection to a source virtualization platform (vSphere, oVirt, OpenStack, or OVA)
- Appropriate RBAC permissions
- Knowledge of VMs you want to migrate

## Step 1: Project Setup (Creating a Namespace)

Create a dedicated namespace for your migration project to organize resources and provide isolation.

### Create the Migration Namespace

```bash
# Create a new namespace for the migration project
kubectl create namespace migration-demo

# Set the namespace as your default for kubectl-mtv commands
kubectl config set-context --current --namespace=migration-demo

# Verify the namespace was created and is active
kubectl config view --minify | grep namespace
kubectl get namespaces | grep migration-demo
```

### Alternative: Use an Existing Namespace

If you prefer to use an existing namespace:

```bash
# List available namespaces
kubectl get namespaces

# Use an existing namespace for all kubectl-mtv commands
kubectl mtv get providers -n your-existing-namespace

# Or set it as your default context
kubectl config set-context --current --namespace=your-existing-namespace
```

### Verify Cluster Access

Before proceeding, verify you can access MTV/Forklift resources:

```bash
# Check if MTV/Forklift CRDs are installed
kubectl get crd | grep forklift

# Verify you can list existing resources
kubectl mtv get providers
kubectl mtv get plans
```

## Step 2: Registering Providers (Source and Target)

Providers represent the source and target platforms for your migration. You'll need both a source provider (your current virtualization platform) and a target provider (your Kubernetes cluster).

### Create Target Provider (OpenShift/Kubernetes)

First, create the target provider representing your Kubernetes cluster:

```bash
# Create OpenShift/Kubernetes target provider
kubectl mtv create provider k8s-target --type openshift

# Verify the target provider was created
kubectl mtv get provider k8s-target
```

### Create Source Provider

Choose the appropriate command for your source platform:

#### VMware vSphere Provider

```bash
# Basic vSphere provider
kubectl mtv create provider vsphere-source --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password YourPassword

# vSphere provider with VDDK optimization (recommended)
kubectl mtv create provider vsphere-source --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password YourPassword \
  --vddk-init-image quay.io/your-registry/vddk:8.0.1

# vSphere provider with TLS verification disabled (for testing)
kubectl mtv create provider vsphere-source --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password YourPassword \
  --provider-insecure-skip-tls
```

#### oVirt/RHV Provider

```bash
# oVirt/RHV provider
kubectl mtv create provider ovirt-source --type ovirt \
  --url https://ovirt-engine.example.com/ovirt-engine/api \
  --username admin@internal \
  --password YourPassword
```

#### OpenStack Provider

```bash
# OpenStack provider
kubectl mtv create provider openstack-source --type openstack \
  --url https://openstack.example.com:5000/v3 \
  --username admin \
  --password YourPassword \
  --provider-domain-name Default \
  --provider-project-name admin
```

#### KubeVirt/OpenShift Virtualization Provider

```bash
# Another KubeVirt cluster as source
kubectl mtv create provider kubevirt-source --type openshift \
  --url https://api.source-cluster.example.com:6443 \
  --token your-service-account-token
```

### Verify Provider Registration

```bash
# List all providers
kubectl mtv get providers

# Get detailed provider information
kubectl mtv get providers -o yaml

# Check provider status
kubectl mtv describe provider vsphere-source
kubectl mtv describe provider k8s-target
```

Wait for providers to show "Ready" status before proceeding.

## Step 3: Creating Mappings (Optional Step)

Mappings define how source networks and storage should be mapped to target resources. While optional (kubectl-mtv can use automatic defaults), explicit mappings provide better control.

### Option A: Use Automatic Default Mappings (Simplest)

Skip this step to let kubectl-mtv automatically select appropriate target networks and storage:

```bash
# kubectl-mtv will automatically:
# - Map to default network (pod network)
# - Use default storage class
# - This works well for simple scenarios
```

### Option B: Create Explicit Mappings

For production scenarios, create explicit network and storage mappings:

#### Create Network Mapping

```bash
# Create network mapping
kubectl mtv create mapping network prod-network \
  --source vsphere-source \
  --target k8s-target \
  --network-pairs "VM Network:default,Management Network:multus-network/mgmt-net"

# Verify network mapping
kubectl mtv get mapping network prod-network
```

#### Create Storage Mapping

```bash
# Basic storage mapping
kubectl mtv create mapping storage prod-storage \
  --source vsphere-source \
  --target k8s-target \
  --storage-pairs "datastore1:fast-ssd,datastore2:standard"

# Advanced storage mapping with volume options
kubectl mtv create mapping storage prod-storage-advanced \
  --source vsphere-source \
  --target k8s-target \
  --storage-pairs "datastore1:fast-ssd;volumeMode=Block;accessMode=ReadWriteOnce,datastore2:standard;volumeMode=Filesystem"
```

#### Verify Mappings

```bash
# List all mappings
kubectl mtv get mappings

# Describe specific mappings
kubectl mtv describe mapping network prod-network
kubectl mtv describe mapping storage prod-storage
```

## Step 4: Creating the Migration Plan

The migration plan defines which VMs to migrate and how to migrate them.

### Discover Available VMs

First, explore the available VMs on your source provider:

```bash
# List all VMs
kubectl mtv get inventory vms vsphere-source

# Filter VMs with queries
kubectl mtv get inventory vms vsphere-source -q "where powerState = 'poweredOn'"
kubectl mtv get inventory vms vsphere-source -q "where memoryMB > 4096"
kubectl mtv get inventory vms vsphere-source -q "where name ~= 'web-.*'"

# Get detailed VM information
kubectl mtv get inventory vms vsphere-source -o yaml
```

### Create Migration Plan

Choose one of these approaches to create your migration plan:

#### Option A: Simple Plan with Automatic Mappings

```bash
# Create a simple migration plan (uses automatic mappings)
kubectl mtv create plan demo-migration \
  --source vsphere-source \
  --vms "web-server-01,database-01,app-server-01"

# Cold migration (default)
kubectl mtv create plan demo-migration \
  --source vsphere-source \
  --vms "web-server-01,database-01" \
  --migration-type cold
```

#### Option B: Plan with Explicit Mappings

```bash
# Create plan using explicit mappings
kubectl mtv create plan prod-migration \
  --source vsphere-source \
  --network-mapping prod-network \
  --storage-mapping prod-storage \
  --vms "web-server-01,database-01,app-server-01"
```

#### Option C: Warm Migration Plan

```bash
# Create warm migration plan (minimizes downtime)
kubectl mtv create plan warm-migration \
  --source vsphere-source \
  --vms "critical-app-01,critical-db-01" \
  --migration-type warm
```

#### Option D: VM Selection by Query

```bash
# Select VMs using a query
kubectl mtv create plan query-migration \
  --source vsphere-source \
  --vms "where name ~= 'prod-.*' and powerState = 'poweredOn'"
```

### Verify Migration Plan

```bash
# List all plans
kubectl mtv get plans

# Get detailed plan information
kubectl mtv describe plan demo-migration

# View plan with VM details
kubectl mtv describe plan demo-migration --with-vms

# Check plan status
kubectl mtv get plan demo-migration -o yaml
```

Wait for the plan to show "Ready" status before starting the migration.

## Step 5: Executing and Monitoring the Migration

Now execute the migration and monitor its progress.

### Start the Migration

```bash
# Start the migration plan
kubectl mtv start plan demo-migration

# For warm migrations, you can schedule a cutover time
kubectl mtv start plan warm-migration --cutover "2024-01-15T10:00:00Z"

# Start multiple plans
kubectl mtv start plan demo-migration prod-migration
```

### Monitor Migration Progress

#### Real-time Monitoring

```bash
# Watch all plans (live updates)
kubectl mtv get plans --watch

# Watch specific plan
kubectl mtv get plan demo-migration --watch

# Watch VM-level progress
kubectl mtv get plan demo-migration --vms --watch
```

#### Check Migration Status

```bash
# Get current plan status
kubectl mtv get plan demo-migration

# Detailed plan status
kubectl mtv describe plan demo-migration

# View VM migration details
kubectl mtv describe plan demo-migration --vm web-server-01

# JSON output for scripting
kubectl mtv get plan demo-migration -o json
```

#### Monitor with Different Output Formats

```bash
# Table format (default)
kubectl mtv get plans

# YAML format for detailed information
kubectl mtv get plan demo-migration -o yaml

# JSON format for automation
kubectl mtv get plan demo-migration -o json | jq '.status'
```

### Migration Lifecycle Commands

#### For Warm Migrations

```bash
# After warm migration completes initial sync, perform cutover
kubectl mtv cutover plan warm-migration

# Or schedule cutover for later
kubectl mtv cutover plan warm-migration --cutover "2024-01-15T02:00:00Z"
```

#### Emergency Operations

```bash
# Cancel a running migration if needed
kubectl mtv cancel plan demo-migration

# Cancel specific VMs within a plan
kubectl mtv cancel plan demo-migration --vm web-server-01
```

#### Post-Migration Operations

```bash
# Archive completed migration
kubectl mtv archive plan demo-migration

# List archived plans
kubectl mtv get plans --all

# Unarchive if needed
kubectl mtv unarchive plan demo-migration
```

### Understanding Migration Phases

During monitoring, you'll see VMs progress through these phases:

1. **Started** - Migration begins
2. **PreHook** - Pre-migration hooks execute (if configured)
3. **CreateSnapshot** - Initial snapshot created (warm migrations)
4. **CreateDataVolumes** - Target storage provisioned
5. **CopyDisks** - Disk data transfer
6. **ConvertGuest** - Guest OS conversion and drivers
7. **CreateVM** - Target VM creation
8. **PostHook** - Post-migration hooks execute (if configured)
9. **Completed** - Migration finished successfully

### Troubleshooting During Migration

If issues occur during migration:

```bash
# Check detailed VM status
kubectl mtv describe plan demo-migration --vm problematic-vm

# Enable debug logging
kubectl mtv get plan demo-migration --watch -v=2

# Check Kubernetes events
kubectl get events --sort-by='.lastTimestamp' | grep demo-migration

# Check underlying Forklift resources
kubectl get migration -o yaml
kubectl get virtualmachine -o yaml
```

## Migration Success Verification

After migration completes, verify your VMs:

```bash
# Check final plan status
kubectl mtv get plan demo-migration

# List created VMs
kubectl get vms

# Check VM status
kubectl get vms -o wide

# Access VMs (using virtctl)
virtctl console web-server-01
virtctl ssh user@web-server-01
```

## Clean Up

After successful migration:

```bash
# Archive the completed plan
kubectl mtv archive plan demo-migration

# Optionally delete the plan if no longer needed
kubectl mtv delete plan demo-migration

# Clean up providers if no longer needed
kubectl mtv delete provider vsphere-source

# Keep the namespace for future migrations or delete it
kubectl delete namespace migration-demo
```

## Common Migration Patterns

### Pattern 1: Test Migration

```bash
# Create a small test plan first
kubectl mtv create plan test-migration \
  --source vsphere-source \
  --vms "test-vm-01" \
  --migration-type cold

kubectl mtv start plan test-migration
kubectl mtv get plan test-migration --watch
```

### Pattern 2: Phased Production Migration

```bash
# Phase 1: Non-critical systems
kubectl mtv create plan phase1-migration \
  --source vsphere-source \
  --vms "dev-server-01,test-server-01" \
  --migration-type cold

# Phase 2: Critical systems with warm migration
kubectl mtv create plan phase2-migration \
  --source vsphere-source \
  --vms "prod-app-01,prod-db-01" \
  --migration-type warm
```

### Pattern 3: Query-Based Migration

```bash
# Migrate all development VMs
kubectl mtv create plan dev-migration \
  --source vsphere-source \
  --vms "where name ~= 'dev-.*' and powerState = 'poweredOn'"

# Migrate VMs with specific characteristics
kubectl mtv create plan memory-migration \
  --source vsphere-source \
  --vms "where memoryMB > 8192 and len disks <= 2"
```

## Next Steps

After completing your first migration:

1. **Explore Advanced Features**: Learn about [Provider Management](04-provider-management.md) and [VDDK Optimization](06-vddk-image-creation-and-configuration.md)
2. **Master Query Language**: Dive into [Advanced Filtering](08-query-language-reference-and-advanced-filtering.md)
3. **Optimize Performance**: Study [Migration Process Optimization](13-migration-process-optimization.md)
4. **Add Automation**: Implement [Migration Hooks](14-migration-hooks.md)
5. **Scale Up**: Plan larger migrations with [Best Practices](18-best-practices-and-security.md)

## Troubleshooting Quick Reference

### Common Issues

| Issue | Command | Solution |
|-------|---------|----------|
| Provider not ready | `kubectl mtv describe provider NAME` | Check credentials and network connectivity |
| Plan stuck in "Not Ready" | `kubectl mtv describe plan NAME` | Verify provider status and VM names |
| Migration stuck | `kubectl mtv describe plan NAME --vm VM` | Check VM-specific errors and resource constraints |
| Disk copy slow | `kubectl mtv get plan NAME -v=2` | Consider VDDK optimization for VMware |
| Target VM won't start | `kubectl get vm NAME -o yaml` | Check resource requirements and node capacity |

### Useful Debug Commands

```bash
# Enable detailed logging
kubectl mtv get plans --watch -v=2

# Check cluster resources
kubectl top nodes
kubectl get pv,pvc

# Monitor network policies
kubectl get networkpolicies

# Check MTV operator logs
kubectl logs -n konveyor-forklift deployment/forklift-controller
```

---

*Previous: [Chapter 2: Installation and Prerequisites](02-installation-and-prerequisites.md)*  
*Next: [Chapter 4: Provider Management](04-provider-management.md)*
