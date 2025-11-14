---
layout: page
title: "Chapter 10: Migration Plan Creation"
---

Migration plans are the core orchestration resources that define which VMs to migrate, how to map their resources, and what settings to apply during migration. This chapter covers comprehensive plan creation with all supported configuration options.

## Overview of Migration Plans

### What is a Migration Plan?

A migration plan is a Kubernetes custom resource that defines:

- **VM Selection**: Which VMs to migrate from the source provider
- **Resource Mapping**: How to translate networks and storage to target resources
- **Migration Configuration**: Type, timing, and behavioral settings
- **Target Customization**: Where and how VMs should run on the target platform

### Plan Components

Every migration plan includes:

1. **Source and Target Providers**: Define the migration endpoints
2. **VM List**: Specific VMs to include in the migration
3. **Mappings**: Network and storage resource translations
4. **Migration Settings**: Type, hooks, templates, and optimization options

## VM Selection Methods

kubectl-mtv supports three flexible methods for VM selection, verified from the command code:

### Method 1: Comma-separated List of VM Names

The simplest method specifies VM names directly:

```bash
# Basic VM list
kubectl mtv create plan simple-migration \
  --source vsphere-prod \
  --vms "web-server-01,db-server-01,app-server-01"

# Single VM migration
kubectl mtv create plan single-vm \
  --source vsphere-prod \
  --vms "critical-database-01"
```

### Method 2: File Reference (`--vms @file.yaml`)

Use file references for complex VM configurations or large VM sets:

#### Create VM List File

```yaml
# Save as vm-list.yaml
- name: web-server-01
  targetName: web-prod-01
- name: web-server-02
  targetName: web-prod-02
- name: database-01
  targetName: db-prod-01
  rootDisk: /dev/sda
```

#### Use File in Plan Creation

```bash
# Reference VM file
kubectl mtv create plan file-based-migration \
  --source vsphere-prod \
  --vms @vm-list.yaml

# Alternative: JSON format
kubectl mtv create plan json-migration \
  --source vsphere-prod \
  --vms @vm-list.json
```

### Method 3: Query String Selection (`--vms "where ..."`)

Use Tree Search Language (TSL) queries for dynamic VM selection:

#### Basic Query Selection

```bash
# Migrate all powered-on VMs
kubectl mtv create plan powered-on-vms \
  --source vsphere-prod \
  --vms "where powerState = 'poweredOn'"

# Migrate VMs by name pattern
kubectl mtv create plan production-vms \
  --source vsphere-prod \
  --vms "where name ~= '^prod-.*'"
```

#### Advanced Query Selection

```bash
# Select VMs by resource criteria
kubectl mtv create plan small-vms \
  --source vsphere-prod \
  --vms "where powerState = 'poweredOn' and memoryMB <= 8192 and len disks <= 3"

# Select VMs in specific infrastructure
kubectl mtv create plan cluster-migration \
  --source vsphere-prod \
  --vms "where cluster.name = 'Production-Cluster' and not template"

# Complex multi-condition queries
kubectl mtv create plan filtered-migration \
  --source vsphere-prod \
  --vms "where (name ~= '^web-.*' or name ~= '^app-.*') and powerState = 'poweredOn' and memoryMB >= 4096"
```

## Mapping Configuration Options in Plan Creation

Migration plans support three approaches for resource mapping:

### Using Existing Mappings

Reference pre-created mapping resources:

```bash
# Use existing network and storage mappings
kubectl mtv create plan mapped-migration \
  --source vsphere-prod \
  --network-mapping production-network-map \
  --storage-mapping production-storage-map \
  --vms "web-01,web-02,db-01"

# Use mappings from different namespaces
kubectl mtv create plan cross-ns-mappings \
  --source vsphere-prod \
  --network-mapping shared-mappings/network-map \
  --storage-mapping shared-mappings/storage-map \
  --vms @vm-list.yaml
```

### Using Inline Mapping Pairs

Define mappings directly in the plan creation command:

```bash
# Inline network and storage pairs
kubectl mtv create plan inline-mappings \
  --source vsphere-prod \
  --network-pairs "VM Network:default,Management Network:multus-system/mgmt-net" \
  --storage-pairs "datastore1:fast-ssd,datastore2:standard-storage" \
  --vms "web-server-01,app-server-01"

# Complex inline mappings with enhanced storage options
kubectl mtv create plan advanced-inline \
  --source vsphere-prod \
  --network-pairs "Production VLAN:prod-net,DMZ Network:security/dmz-net,Backup Network:ignored" \
  --storage-pairs "fast-ds:premium-ssd;volumeMode=Block;accessMode=ReadWriteOnce,shared-ds:shared-nfs;volumeMode=Filesystem;accessMode=ReadWriteMany" \
  --vms "where cluster.name = 'Prod-Cluster'"
```

### Using Default Mappings (Simplest Approach)

Let kubectl-mtv create simple default mappings:

```bash
# Use default network and storage class
kubectl mtv create plan default-migration \
  --source vsphere-prod \
  --default-target-network default \
  --default-target-storage-class standard-ssd \
  --vms "test-vm-01,test-vm-02"

# Default pod networking and specific storage
kubectl mtv create plan pod-network-migration \
  --source vsphere-prod \
  --default-target-network default \
  --default-target-storage-class premium-nvme \
  --vms "where name ~= '^test-.*'"
```

## Key Plan Configuration Flags

### Migration Types

kubectl-mtv supports four migration types (see [Forklift Migration Types](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#about-cold-and-warm-migration)):

| Type | Description | Use Case |
|------|-------------|----------|
| `cold` | Offline migration | Production VMs where downtime is acceptable |
| `warm` | Pre-copy with minimal downtime | Large VMs where downtime must be minimized |
| `live` | Live migration (KubeVirt sources only) | Zero-downtime migration between KubeVirt clusters |
| `conversion` | Guest conversion only (VMware only) | When storage vendors provide pre-populated PVCs |

For detailed information about conversion migration, including prerequisites, workflow, and integration requirements, see [Chapter 3.6: Conversion Migration](03.6-conversion-migration).

#### Migration Type Examples

```bash
# Cold migration (default)
kubectl mtv create plan cold-migration \
  --source vsphere-prod \
  --migration-type cold \
  --vms "batch-processor-01,backup-server-01"

# Warm migration for large VMs
kubectl mtv create plan warm-migration \
  --source vsphere-prod \
  --migration-type warm \
  --vms "where memoryMB > 16384"

# Live migration (KubeVirt to KubeVirt)
kubectl mtv create plan live-migration \
  --source kubevirt-cluster1 \
  --migration-type live \
  --vms "production-workload-01"

# Conversion-only migration (VMware only)
# Prerequisites: PVCs must be pre-created with proper labels and annotations
kubectl mtv create plan conversion-only \
  --source vsphere-prod \
  --migration-type conversion \
  --vms "vm-with-precreated-pvcs"
```

### Target Namespace and Transfer Network

#### Target Namespace Configuration

```bash
# Specify target namespace
kubectl mtv create plan namespaced-migration \
  --source vsphere-prod \
  --target-namespace production-workloads \
  --vms "prod-web-01,prod-api-01"

# Use current namespace (default)
kubectl mtv create plan current-ns-migration \
  --source vsphere-prod \
  --vms "dev-app-01,dev-db-01" \
  -n development
```

#### Transfer Network Configuration

```bash
# Use specific transfer network for disk operations
kubectl mtv create plan transfer-net-migration \
  --source vsphere-prod \
  --transfer-network migration-net \
  --vms "large-vm-01,large-vm-02"

# Cross-namespace transfer network
kubectl mtv create plan cross-ns-transfer \
  --source vsphere-prod \
  --transfer-network network-system/high-bandwidth \
  --vms "where sum disks.capacityGB > 500"
```

### Naming Templates

kubectl-mtv provides Go template variables for customizing resource names:

#### PVC Name Template

Available variables: `{% raw %}{{.VmName}}{% endraw %}`, `{% raw %}{{.PlanName}}{% endraw %}`, `{% raw %}{{.DiskIndex}}{% endraw %}`, `{% raw %}{{.WinDriveLetter}}{% endraw %}`, `{% raw %}{{.RootDiskIndex}}{% endraw %}`, `{% raw %}{{.Shared}}{% endraw %}`, `{% raw %}{{.FileName}}{% endraw %}`

```bash
# Custom PVC naming with plan and VM name
kubectl mtv create plan custom-pvc-names \
  --source vsphere-prod \
  --pvc-name-template "{% raw %}{{.PlanName}}{% endraw %}-{% raw %}{{.VmName}}{% endraw %}-disk-{% raw %}{{.DiskIndex}}{% endraw %}" \
  --vms "web-01,web-02"

# Windows-specific PVC naming
kubectl mtv create plan windows-migration \
  --source vsphere-prod \
  --pvc-name-template "{% raw %}{{.VmName}}{% endraw %}-{% raw %}{{.WinDriveLetter}}{% endraw %}-disk" \
  --vms "where guestOS ~= '.*windows.*'"

# Shared disk identification
kubectl mtv create plan shared-disk-migration \
  --source vsphere-prod \
  --pvc-name-template "{% raw %}{{.VmName}}{% endraw %}-{% raw %}{{if .Shared}}{% endraw %}shared{% raw %}{{else}}{% endraw %}local{% raw %}{{end}}{% endraw %}-{% raw %}{{.DiskIndex}}{% endraw %}" \
  --vms "cluster-node-01,cluster-node-02"
```

#### Volume Name Template

Available variables: `{% raw %}{{.PVCName}}{% endraw %}`, `{% raw %}{{.VolumeIndex}}{% endraw %}`

```bash
# Custom volume interface names
kubectl mtv create plan custom-volumes \
  --source vsphere-prod \
  --volume-name-template "vol-{% raw %}{{.VolumeIndex}}{% endraw %}-{% raw %}{{.PVCName}}{% endraw %}" \
  --vms "multi-disk-vm-01"
```

#### Network Name Template

Available variables: `{% raw %}{{.NetworkName}}{% endraw %}`, `{% raw %}{{.NetworkNamespace}}{% endraw %}`, `{% raw %}{{.NetworkType}}{% endraw %}`, `{% raw %}{{.NetworkIndex}}{% endraw %}`

```bash
# Custom network interface names
kubectl mtv create plan custom-networks \
  --source vsphere-prod \
  --network-name-template "{% raw %}{{.NetworkType}}{% endraw %}-{% raw %}{{.NetworkIndex}}{% endraw %}" \
  --vms "multi-nic-vm-01"
```

## Advanced Plan Configuration

### VM-Level Customization

#### Target VM Labels and Node Scheduling

```bash
# Add labels to target VMs
kubectl mtv create plan labeled-migration \
  --source vsphere-prod \
  --target-labels "environment=production,team=platform,tier=web" \
  --vms "web-server-01,web-server-02"

# Node selector for VM placement
kubectl mtv create plan node-constrained \
  --source vsphere-prod \
  --target-node-selector "zone=east,storage=ssd" \
  --vms "performance-app-01"

# Combined labels and node selector
kubectl mtv create plan full-scheduling \
  --source vsphere-prod \
  --target-labels "app=database,performance=high" \
  --target-node-selector "zone=central,memory=high" \
  --vms "database-cluster-01"
```

#### VM Affinity with KARL (Kubernetes Affinity Rule Language)

```bash
# Require co-location with specific pods
kubectl mtv create plan affinity-colocation \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(app=database) on node" \
  --vms "app-server-01"

# Anti-affinity for high availability
kubectl mtv create plan ha-placement \
  --source vsphere-prod \
  --target-affinity "AVOID pods(app=web) on node" \
  --vms "web-server-03,web-server-04"

# Complex affinity rules
kubectl mtv create plan complex-affinity \
  --source vsphere-prod \
  --target-affinity "REQUIRE zone(east) AND AVOID pods(resource=intensive)" \
  --vms "latency-sensitive-01"
```

#### Target Power State Control

```bash
# Keep VMs powered off after migration
kubectl mtv create plan offline-migration \
  --source vsphere-prod \
  --target-power-state off \
  --vms "backup-vm-01,archive-vm-01"

# Ensure VMs start after migration
kubectl mtv create plan online-migration \
  --source vsphere-prod \
  --target-power-state on \
  --vms "web-service-01,api-service-01"

# Auto power state (match source)
kubectl mtv create plan auto-power \
  --source vsphere-prod \
  --target-power-state auto \
  --vms "where powerState in ('poweredOn', 'suspended')"
```

### Guest OS and Compatibility Settings

#### Guest Conversion Configuration

```bash
# Skip guest conversion for Linux VMs
kubectl mtv create plan linux-migration \
  --source vsphere-prod \
  --skip-guest-conversion \
  --use-compatibility-mode \
  --vms "where guestOS ~= '.*linux.*'"

# Enable guest conversion with cleanup
kubectl mtv create plan windows-conversion \
  --source vsphere-prod \
  --delete-guest-conversion-pod \
  --vms "where guestOS ~= '.*windows.*'"

# Legacy driver installation for Windows
kubectl mtv create plan legacy-windows \
  --source vsphere-prod \
  --install-legacy-drivers true \
  --vms "windows-2012-server"
```

#### Static IP Preservation

```bash
# Preserve static IPs (vSphere only)
kubectl mtv create plan preserve-ips \
  --source vsphere-prod \
  --preserve-static-ips \
  --vms "database-01,web-lb-01"

# Disable IP preservation
kubectl mtv create plan new-ips \
  --source vsphere-prod \
  --preserve-static-ips=false \
  --vms "test-vm-01,dev-vm-01"
```

### Convertor Pod Configuration

Configure the virt-v2v conversion pods:

```bash
# Convertor pod labels and scheduling
kubectl mtv create plan convertor-config \
  --source vsphere-prod \
  --convertor-labels "team=migration,priority=high" \
  --convertor-node-selector "conversion=true,cpu=high" \
  --vms "large-workload-01"

# Convertor affinity with KARL
kubectl mtv create plan convertor-affinity \
  --source vsphere-prod \
  --convertor-affinity "REQUIRE nodes(conversion=dedicated)" \
  --vms "complex-vm-01"
```

### Disk and Storage Configuration

#### Shared Disk Migration

```bash
# Include shared disks in migration
kubectl mtv create plan shared-disks \
  --source vsphere-prod \
  --migrate-shared-disks \
  --vms "cluster-vm-01,cluster-vm-02"

# Exclude shared disks
kubectl mtv create plan no-shared-disks \
  --source vsphere-prod \
  --migrate-shared-disks=false \
  --vms "standalone-vm-01"
```

#### Preflight Inspection (Warm Migrations)

```bash
# Enable preflight disk inspection for warm migrations
kubectl mtv create plan warm-with-preflight \
  --source vsphere-prod \
  --migration-type warm \
  --run-preflight-inspection \
  --vms "critical-database-01"

# Disable preflight for faster warm start
kubectl mtv create plan warm-no-preflight \
  --source vsphere-prod \
  --migration-type warm \
  --run-preflight-inspection=false \
  --vms "test-warm-vm-01"
```

## Migration Hooks Integration

Add hooks to run custom automation during migrations:

### Pre-migration and Post-migration Hooks

```bash
# Add pre and post hooks to all VMs
kubectl mtv create plan hooked-migration \
  --source vsphere-prod \
  --pre-hook backup-hook \
  --post-hook cleanup-hook \
  --vms "web-01,db-01"

# Pre-hook only for preparation tasks
kubectl mtv create plan prep-migration \
  --source vsphere-prod \
  --pre-hook application-quiesce \
  --vms "database-cluster-01"

# Post-hook only for validation
kubectl mtv create plan validated-migration \
  --source vsphere-prod \
  --post-hook health-check \
  --vms "web-service-01"
```

## Complete Plan Creation Examples

### Example 1: Enterprise Production Migration

```bash
# Comprehensive enterprise migration plan
kubectl mtv create plan enterprise-production \
  --source vsphere-production \
  --target openshift-production \
  --network-mapping enterprise-network-map \
  --storage-mapping enterprise-storage-map \
  --migration-type warm \
  --target-namespace production-workloads \
  --transfer-network migration-backbone \
  --preserve-static-ips \
  --target-labels "environment=production,migration=phase1" \
  --target-node-selector "zone=east,performance=high" \
  --pvc-name-template "{% raw %}{{.PlanName}}{% endraw %}-{% raw %}{{.VmName}}{% endraw %}-{% raw %}{{.DiskIndex}}{% endraw %}" \
  --pre-hook production-backup \
  --post-hook production-validation \
  --run-preflight-inspection \
  --delete-guest-conversion-pod \
  --vms "where cluster.name = 'Production-East' and powerState = 'poweredOn' and not template"
```

### Example 2: Development Environment Migration

```bash
# Simple development migration
kubectl mtv create plan dev-migration \
  --source vsphere-dev \
  --migration-type cold \
  --default-target-network default \
  --default-target-storage-class standard-ssd \
  --target-namespace development \
  --skip-guest-conversion \
  --use-compatibility-mode \
  --target-power-state on \
  --vms "dev-web-01,dev-api-01,dev-db-01" \
  -n development
```

### Example 3: Query-Based Batch Migration

```bash
# Large-scale query-driven migration
kubectl mtv create plan batch-small-vms \
  --source vsphere-prod \
  --migration-type cold \
  --network-pairs "VM Network:default,Management Network:ignored" \
  --storage-pairs "datastore1:standard-ssd,datastore2:premium-nvme;volumeMode=Block" \
  --target-labels "batch=phase1,size=small" \
  --convertor-node-selector "batch-processing=true" \
  --pvc-name-template "batch-{% raw %}{{.PlanName}}{% endraw %}-{% raw %}{{.VmName}}{% endraw %}-disk{% raw %}{{.DiskIndex}}{% endraw %}" \
  --delete-vm-on-fail-migration \
  --vms "where powerState = 'poweredOn' and memoryMB <= 4096 and len disks <= 2 and not template"
```

### Example 4: Multi-Provider Migration

```bash
# oVirt to OpenShift migration
kubectl mtv create plan ovirt-migration \
  --source ovirt-production \
  --target openshift-target \
  --migration-type warm \
  --network-pairs "ovirtmgmt:default,production:prod-net" \
  --storage-pairs "data:standard-rwo;volumeMode=Filesystem,fast:premium-ssd;volumeMode=Block" \
  --preserve-cluster-cpu-model \
  --target-namespace migrated-workloads \
  --vms @ovirt-vm-list.yaml

# OpenStack to OpenShift migration
kubectl mtv create plan openstack-migration \
  --source openstack-prod \
  --target openshift-target \
  --migration-type cold \
  --network-pairs "internal:default,external:multus-system/external" \
  --storage-pairs "__DEFAULT__:standard-rwo,ssd:premium-ssd" \
  --target-labels "source=openstack,migration=batch2" \
  --vms "where flavor.name = 'm1.medium' and status = 'ACTIVE'"
```

### Example 5: Storage Array Offloading

```bash
# Plan with storage array offloading
kubectl mtv create plan offload-migration \
  --source vsphere-prod \
  --storage-pairs "tier1-ds:flashsystem-gold;offloadPlugin=vsphere;offloadVendor=flashsystem" \
  --default-offload-plugin vsphere \
  --offload-vsphere-username vcenter-svc@vsphere.local \
  --offload-vsphere-password VCenterPassword \
  --offload-vsphere-url https://vcenter.company.com \
  --offload-storage-username flashsystem-admin \
  --offload-storage-password FlashSystemPassword \
  --offload-storage-endpoint https://flashsystem.company.com \
  --migration-type warm \
  --vms "where sum disks.capacityGB > 500"
```

## Plan Validation and Testing

### Pre-Creation Validation

```bash
# Test VM query before creating plan
kubectl mtv get inventory vms vsphere-prod \
  -q "where powerState = 'poweredOn' and memoryMB <= 8192" \
  -o table

# Validate mapping resources exist
kubectl mtv describe mapping network enterprise-network-map
kubectl mtv describe mapping storage enterprise-storage-map

# Check target namespace and resources
kubectl get storageclass premium-ssd
kubectl get network-attachment-definitions -n production
```

### Dry-Run Planning

```bash
# Use --dry-run to validate without creating
kubectl mtv create plan test-validation \
  --source vsphere-prod \
  --vms "test-vm-01" \
  --dry-run=client

# Validate query results
kubectl mtv get inventory vms vsphere-prod \
  -q "where cluster.name = 'Test-Cluster'" \
  -o planvms > test-query-results.yaml
```

## Plan Lifecycle Management

### Plan Status Monitoring

```bash
# Check plan creation status
kubectl mtv get plans

# Detailed plan information
kubectl mtv describe plan enterprise-production

# Monitor plan progress
kubectl mtv get plan enterprise-production --watch
```

### Plan Modification

```bash
# Plans are immutable after creation, but you can:

# 1. Create modified copy
kubectl get plan original-plan -o yaml > modified-plan.yaml
# Edit modified-plan.yaml
kubectl apply -f modified-plan.yaml

# 2. Use plan patching (covered in Chapter 15)
kubectl mtv patch plan enterprise-production --archived=true
```

## Troubleshooting Plan Creation

### Common Plan Creation Issues

#### VM Selection Errors

```bash
# Debug query results
kubectl mtv get inventory vms vsphere-prod \
  -q "where powerState = 'poweredOn'" \
  -v=2

# Check if VMs exist
for vm in vm1 vm2 vm3; do
  kubectl mtv get inventory vms vsphere-prod -q "where name = '$vm'"
done
```

#### Mapping Resolution Errors

```bash
# Verify mappings exist
kubectl get networkmapping,storagemapping -A

# Check mapping content
kubectl describe networkmapping enterprise-network-map

# Test inline pairs syntax
echo "source1:target1,source2:target2" | tr ',' '\n'
```

#### Resource Availability Issues

```bash
# Check target resources
kubectl get storageclass
kubectl get network-attachment-definitions -A

# Verify namespace access
kubectl auth can-i create virtualmachines -n target-namespace
```

### Debug Plan Creation

```bash
# Enable verbose logging
kubectl mtv create plan debug-plan \
  --source vsphere-prod \
  --vms "debug-vm-01" \
  -v=2

# Check plan resource creation
kubectl get plan debug-plan -o yaml

# Monitor plan events
kubectl get events --sort-by='.metadata.creationTimestamp' | grep plan
```

## Integration with Other Components

### Integration with Inventory

```bash
# Use inventory queries in plans
kubectl mtv create plan inventory-driven \
  --source vsphere-prod \
  --vms "where cluster.name = 'Production' and host.name ~= 'esxi-0[1-3].*'"

# Export VMs for plan creation
kubectl mtv get inventory vms vsphere-prod \
  -q "where powerState = 'poweredOn'" \
  -o planvms > selected-vms.yaml

kubectl mtv create plan exported-vms \
  --source vsphere-prod \
  --vms @selected-vms.yaml
```

### Integration with Mappings

```bash
# Reference existing mappings
kubectl mtv create plan mapped-migration \
  --network-mapping existing-network-map \
  --storage-mapping existing-storage-map

# Mix mappings and inline pairs (not allowed)
# This will error:
kubectl mtv create plan mixed-mappings \
  --network-mapping existing-map \
  --network-pairs "VM Network:default"  # Error: conflicting options
```

## Best Practices for Plan Creation

### Planning Strategy

1. **Start Small**: Begin with test/development VMs
2. **Batch by Similarity**: Group VMs with similar requirements
3. **Use Queries Effectively**: Leverage TSL for dynamic VM selection
4. **Plan Resource Capacity**: Ensure target resources can handle the workload

### Configuration Best Practices

1. **Use Descriptive Names**: Name plans clearly for operational clarity
2. **Leverage Templates**: Use naming templates for consistent resource naming
3. **Document Dependencies**: Maintain clear documentation of plan relationships
4. **Test Mappings**: Validate mappings before large-scale migrations

### Operational Best Practices

1. **Monitor Plan Progress**: Use watching and logging for plan execution
2. **Plan for Rollback**: Understand rollback procedures before starting
3. **Coordinate Teams**: Ensure all stakeholders understand migration timing
4. **Validate Results**: Have testing procedures ready for post-migration validation

## Next Steps

After mastering plan creation:

1. **Customize VMs**: Learn detailed VM customization in [Chapter 11: Customizing Individual VMs (PlanVMS Format)](11-customizing-individual-vms-planvms-format)
2. **Optimize Placement**: Configure advanced placement in [Chapter 12: Target VM Placement](12-target-vm-placement)
3. **Execute Plans**: Manage plan lifecycle in [Chapter 16: Plan Lifecycle Execution](16-plan-lifecycle-execution)
4. **Advanced Patching**: Modify existing plans in [Chapter 15: Advanced Plan Patching](15-advanced-plan-patching)

---

*Previous: [Chapter 9.5: Storage Array Offloading and Optimization](09.5-storage-array-offloading-and-optimization)*  
*Next: [Chapter 11: Customizing Individual VMs (PlanVMS Format)](11-customizing-individual-vms-planvms-format)*
