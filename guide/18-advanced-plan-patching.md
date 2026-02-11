---
layout: page
title: "Chapter 18: Advanced Plan Patching"
---

Once migration plans are created, you often need to modify their configuration as migration requirements evolve. This chapter covers comprehensive plan modification techniques using `kubectl-mtv` patching capabilities, enabling dynamic updates without recreating plans.

## Overview: Dynamic Plan Modification

### Why Plan Patching?

Migration plans require updates throughout their lifecycle:

- **Environment Changes**: Updated network configurations, storage requirements, or target namespaces
- **Migration Strategy Evolution**: Switching between cold, warm, and live migration types
- **VM-Specific Customization**: Individual VM requirements that emerge during planning
- **Hook Integration**: Adding custom automation after initial plan creation
- **Performance Optimization**: Adjusting convertor pod scheduling for better performance

### Patching Capabilities

`kubectl-mtv` provides two primary patching mechanisms:

1. **Plan-Level Patching** (`patch plan`): Modify plan-wide settings affecting all VMs
2. **VM-Level Patching** (`patch planvm`): Customize individual VMs within the plan

Both approaches support comprehensive configuration updates verified from the command implementation.

### Patching vs. Recreation

**When to Patch:**
- Plan exists and contains complex VM selections
- Need to preserve existing plan metadata and references
- Making incremental configuration changes
- Plan is referenced by other resources or automation

**When to Recreate:**
- Fundamental changes to source/target providers
- Major VM list modifications
- Complete migration strategy overhaul

## How-To: Patching Plan Settings

### Basic Plan Configuration Updates

#### Migration Type Changes

Change migration strategies dynamically based on requirements:

```bash
# Switch from cold to warm migration
kubectl mtv patch plan production-migration \
  --migration-type warm

# Enable live migration (KubeVirt sources only)
kubectl mtv patch plan k8s-to-k8s-migration \
  --migration-type live

# Switch to cold migration for maximum reliability
kubectl mtv patch plan critical-systems \
  --migration-type cold
```

#### Network and Storage Configuration

Update network settings and target configurations:

```bash
# Change transfer network for better performance
kubectl mtv patch plan large-vm-migration \
  --transfer-network migration-network/high-bandwidth-net

# Update target namespace
kubectl mtv patch plan dev-environment \
  --target-namespace development-new

# Modify description for better documentation
kubectl mtv patch plan quarterly-migration \
  --description "Q4 2024 production workload migration to OpenShift 4.16"
```

### Advanced Plan Configuration

#### Target VM Placement Updates

Modify where target VMs will be scheduled using various placement strategies:

```bash
# Add labels to all target VMs
kubectl mtv patch plan production-apps \
  --target-labels "environment=production,migration-batch=2024-q4"

# Update node selector for hardware requirements
kubectl mtv patch plan gpu-workloads \
  --target-node-selector "accelerator=nvidia-tesla-v100,node-type=compute"

# Apply advanced affinity rules using KARL (see Chapter 28 for syntax reference)
kubectl mtv patch plan database-cluster \
  --target-affinity "REQUIRE nodes(node-role.kubernetes.io/database=true) on node"

# Set power state for all VMs after migration  
kubectl mtv patch plan maintenance-migration \
  --target-power-state off
```

#### Convertor Pod Optimization Updates

Update convertor pod scheduling for optimal migration performance:

```bash
# Move convertors to high-performance nodes
kubectl mtv patch plan data-intensive-migration \
  --convertor-node-selector "node-type=high-io,storage-class=nvme"

# Apply convertor affinity for storage proximity
kubectl mtv patch plan storage-migration \
  --convertor-affinity "REQUIRE pods(app=ceph-osd) on node"

# Add labels to convertor pods for monitoring
kubectl mtv patch plan monitored-migration \
  --convertor-labels "monitoring=enabled,migration-type=production"
```

### Template and Naming Configuration

Update naming templates for better resource organization:

```bash
# Update PVC naming template
kubectl mtv patch plan organized-migration \
  --pvc-name-template "{% raw %}{{.PlanName}}{% endraw %}-{% raw %}{{.TargetVmName}}{% endraw %}-disk-{% raw %}{{.DiskIndex}}{% endraw %}"

# Set volume naming template
kubectl mtv patch plan structured-storage \
  --volume-name-template "vol-{% raw %}{{.PVCName}}{% endraw %}-{% raw %}{{.VolumeIndex}}{% endraw %}"

# Configure network interface naming
kubectl mtv patch plan network-organized \
  --network-name-template "{% raw %}{{.TargetVmName}}{% endraw %}-{% raw %}{{.NetworkType}}{% endraw %}-{% raw %}{{.NetworkIndex}}{% endraw %}"
```

### Migration Behavior Configuration

Control various aspects of the migration process:

```bash
# Enable static IP preservation for vSphere VMs
kubectl mtv patch plan vsphere-production \
  --preserve-static-ips=true

# Configure shared disk migration
kubectl mtv patch plan cluster-workloads \
  --migrate-shared-disks=true

# Enable compatibility mode for older systems
kubectl mtv patch plan legacy-systems \
  --use-compatibility-mode=true

# Configure cleanup behavior
kubectl mtv patch plan test-migration \
  --delete-guest-conversion-pod=true \
  --delete-vm-on-fail-migration=true

# Skip guest conversion for specific use cases
kubectl mtv patch plan raw-disk-migration \
  --skip-guest-conversion=true

# Disable preflight inspection for faster warm migrations
kubectl mtv patch plan urgent-migration \
  --run-preflight-inspection=false
```

### Comprehensive Plan Update Example

```bash
# Complete plan reconfiguration for production migration
kubectl mtv patch plan enterprise-migration \
  --description "Enterprise production migration - Phase 2" \
  --migration-type warm \
  --target-namespace production-v2 \
  --transfer-network production/high-speed-migration \
  --target-labels "environment=production,phase=2,team=platform" \
  --target-node-selector "node-role.kubernetes.io/compute=true" \
  --target-affinity "PREFER nodes(topology.kubernetes.io/zone=us-east-1a) on zone" \
  --target-power-state auto \
  --convertor-labels "migration=production,performance=optimized" \
  --convertor-node-selector "node-type=high-io,network=10gbe" \
  --convertor-affinity "REQUIRE nodes(storage-tier=premium) on node" \
  --preserve-static-ips=true \
  --preserve-cluster-cpu-model=true \
  --pvc-name-template "prod-{% raw %}{{.TargetVmName}}{% endraw %}-{% raw %}{{.DiskIndex}}{% endraw %}" \
  --delete-guest-conversion-pod=true \
  --run-preflight-inspection=true
```

## How-To: Patching Individual VMs

### VM Identity and Configuration

Modify individual VM settings within a plan:

```bash
# Change target VM name
kubectl mtv patch planvm production-migration web-server-01 \
  --target-name web-prod-primary

# Set specific instance type for a VM
kubectl mtv patch planvm enterprise-migration database-main \
  --instance-type large-memory

# Specify root disk for VMs with multiple disks
kubectl mtv patch planvm complex-migration multi-disk-vm \
  --root-disk "Hard disk 1"

# Set power state for specific VM
kubectl mtv patch planvm maintenance-migration critical-db \
  --target-power-state on
```

### VM-Specific Naming Templates

Apply custom naming templates to individual VMs:

```bash
# Custom PVC naming for high-storage VM
kubectl mtv patch planvm data-migration large-database \
  --pvc-name-template "{% raw %}{{.TargetVmName}}{% endraw %}-data-{% raw %}{{.WinDriveLetter}}{% endraw %}-{% raw %}{{.DiskIndex}}{% endraw %}"

# Custom volume naming for multi-tier application
kubectl mtv patch planvm app-migration web-tier \
  --volume-name-template "{% raw %}{{.TargetVmName}}{% endraw %}-vol-{% raw %}{{.VolumeIndex}}{% endraw %}"

# Custom network naming for multi-homed VMs
kubectl mtv patch planvm network-migration firewall-vm \
  --network-name-template "{% raw %}{{.TargetVmName}}{% endraw %}-{% raw %}{{.NetworkType}}{% endraw %}-{% raw %}{{.NetworkIndex}}{% endraw %}"
```

### Security Configuration

Configure VM-specific security settings:

```bash
# Add LUKS decryption secret for encrypted VM
kubectl mtv patch planvm secure-migration encrypted-database \
  --luks-secret db-encryption-keys

# Override plan-level deletion policy for specific VM
kubectl mtv patch planvm test-migration experimental-vm \
  --delete-vm-on-fail-migration=true
```

### Adding and Managing Hooks

Attach custom automation to specific VMs:

#### Adding Hooks

```bash
# Add pre-migration hook to database VM
kubectl mtv patch planvm production-migration database-primary \
  --add-pre-hook database-backup-hook

# Add post-migration hook to web server
kubectl mtv patch planvm production-migration web-server-01 \
  --add-post-hook health-check-hook

# Add both pre and post hooks to critical application
kubectl mtv patch planvm critical-migration app-server-main \
  --add-pre-hook app-quiesce-hook \
  --add-post-hook app-validation-hook
```

#### Managing Existing Hooks

```bash
# Remove specific hook from VM
kubectl mtv patch planvm production-migration web-server-01 \
  --remove-hook old-health-check

# Clear all hooks from VM
kubectl mtv patch planvm production-migration test-vm \
  --clear-hooks

# Replace hook by removing old and adding new
kubectl mtv patch planvm production-migration database-secondary \
  --remove-hook old-backup-hook
kubectl mtv patch planvm production-migration database-secondary \
  --add-pre-hook new-enhanced-backup-hook
```

### Comprehensive VM Update Example

```bash
# Complete VM customization within plan
kubectl mtv patch planvm enterprise-migration critical-database \
  --target-name db-prod-primary \
  --instance-type extra-large \
  --root-disk "SCSI disk 1" \
  --target-power-state on \
  --pvc-name-template "{% raw %}{{.TargetVmName}}{% endraw %}-{% raw %}{{.WinDriveLetter}}{% endraw %}-{% raw %}{{.DiskIndex}}{% endraw %}" \
  --volume-name-template "{% raw %}{{.TargetVmName}}{% endraw %}-storage-{% raw %}{{.VolumeIndex}}{% endraw %}" \
  --network-name-template "{% raw %}{{.TargetVmName}}{% endraw %}-net-{% raw %}{{.NetworkIndex}}{% endraw %}" \
  --luks-secret database-encryption-secret \
  --delete-vm-on-fail-migration=false \
  --add-pre-hook database-cluster-quiesce \
  --add-post-hook database-cluster-validate
```

## Advanced Patching Scenarios

### Scenario 1: Migration Strategy Evolution

```bash
# Initial plan with basic cold migration
kubectl mtv create plan evolving-migration \
  --source vsphere-prod --target openshift-prod \
  --vms "app-server-01,app-server-02,database-01"

# Evolve to warm migration after testing
kubectl mtv patch plan evolving-migration \
  --migration-type warm \
  --run-preflight-inspection=true

# Add convertor optimization for warm migration
kubectl mtv patch plan evolving-migration \
  --convertor-node-selector "node-type=high-io" \
  --convertor-affinity "REQUIRE nodes(network-speed=10gbe) on node"

# Configure individual database VM for special handling
kubectl mtv patch planvm evolving-migration database-01 \
  --target-name database-primary \
  --instance-type large-memory \
  --add-pre-hook database-backup-hook \
  --add-post-hook database-validation-hook
```

### Scenario 2: Performance Optimization Through Patching

```bash
# Start with basic plan
kubectl mtv create plan performance-migration \
  --source vsphere-datacenter --target openshift-cluster \
  --vms "where tags.category='performance-critical'"

# Add performance optimizations
kubectl mtv patch plan performance-migration \
  --transfer-network performance/dedicated-migration-net \
  --convertor-node-selector "node-type=high-performance,storage=nvme" \
  --convertor-labels "priority=high,performance=optimized" \
  --convertor-affinity "REQUIRE nodes(cpu-type=intel-skylake) on node"

# Optimize target placement
kubectl mtv patch plan performance-migration \
  --target-node-selector "node-role.kubernetes.io/compute=true,performance-tier=premium" \
  --target-affinity "PREFER nodes(topology.kubernetes.io/zone=performance-zone) on zone" \
  --target-labels "performance=critical,monitoring=enhanced"
```

### Scenario 3: Incremental Hook Integration

```bash
# Basic plan without hooks
kubectl mtv create plan progressive-automation \
  --source vmware-test --target k8s-test \
  --vms database-test,web-test,cache-test

# Add database-specific automation
kubectl mtv patch planvm progressive-automation database-test \
  --add-pre-hook database-quiesce \
  --add-post-hook database-health-check

# Add web server automation
kubectl mtv patch planvm progressive-automation web-test \
  --add-pre-hook lb-drain-connections \
  --add-post-hook web-health-validation

# Add cache server automation
kubectl mtv patch planvm progressive-automation cache-test \
  --add-post-hook cache-warmup-hook

# Add plan-level notification
# Note: Plan-level hooks require adding to all VMs individually
for vm in database-test web-test cache-test; do
  kubectl mtv patch planvm progressive-automation $vm \
    --add-post-hook migration-notification
done
```

### Scenario 4: Environment-Specific Customization

```bash
# Development environment configuration
kubectl mtv patch plan dev-migration \
  --target-namespace development \
  --target-labels "environment=dev,auto-shutdown=true" \
  --target-power-state off \
  --delete-vm-on-fail-migration=true \
  --delete-guest-conversion-pod=true

# Production environment configuration  
kubectl mtv patch plan prod-migration \
  --target-namespace production \
  --target-labels "environment=prod,backup=required,monitoring=critical" \
  --target-node-selector "node-role.kubernetes.io/production=true" \
  --target-affinity "REQUIRE nodes(reliability-tier=high) on node" \
  --preserve-static-ips=true \
  --preserve-cluster-cpu-model=true \
  --delete-vm-on-fail-migration=false
```

## Patching with Provider Updates

While less common, provider configurations can also be updated:

```bash
# Update provider URL (maintenance window)
kubectl mtv patch provider my-vsphere-provider \
  --url https://new-vcenter.example.com/sdk

# Update provider credentials
kubectl mtv patch provider my-openstack-provider \
  --username new_service_account \
  --password new_secure_password

# Add VDDK image to existing provider
kubectl mtv patch provider vsphere-prod \
  --vddk-init-image registry.company.com/vddk:8.0.2
```

## Best Practices: Plan-Level vs. VM-Level Changes

### Use Plan-Level Patching When:

1. **Uniform Changes**: All VMs need the same configuration updates
2. **Migration Strategy**: Changing migration type, network, or target namespace
3. **Performance Optimization**: Convertor pod scheduling affects all VMs
4. **Environment Settings**: Target cluster configuration applies globally
5. **Template Updates**: Naming templates should be consistent across VMs

### Use VM-Level Patching When:

1. **Individual Customization**: Specific VMs need unique configurations
2. **Selective Hooks**: Only certain VMs require automation
3. **Security Requirements**: LUKS secrets or power states vary per VM
4. **Instance Types**: VMs have different resource requirements
5. **Target Names**: Custom naming for specific VMs

### Strategic Patching Approach

#### 1. Plan First, Then Customize

```bash
# Start with plan-wide optimizations
kubectl mtv patch plan enterprise-migration \
  --migration-type warm \
  --target-namespace production \
  --convertor-node-selector "node-type=high-io"

# Then customize specific VMs
kubectl mtv patch planvm enterprise-migration database-cluster \
  --instance-type extra-large \
  --add-pre-hook cluster-backup
```

#### 2. Group Similar VMs

```bash
# Apply similar configurations to VM groups
for vm in web-01 web-02 web-03; do
  kubectl mtv patch planvm web-migration $vm \
    --target-labels "tier=web,load-balancer=true" \
    --add-post-hook web-health-check
done

for vm in db-primary db-secondary; do
  kubectl mtv patch planvm db-migration $vm \
    --instance-type large-memory \
    --add-pre-hook database-backup \
    --add-post-hook database-validation
done
```

#### 3. Iterative Refinement

```bash
# Initial basic configuration
kubectl mtv patch plan iterative-migration \
  --migration-type cold \
  --target-namespace staging

# Test and refine
kubectl mtv patch plan iterative-migration \
  --migration-type warm \
  --convertor-node-selector "storage=ssd"

# Final production configuration
kubectl mtv patch plan iterative-migration \
  --target-namespace production \
  --target-affinity "REQUIRE nodes(reliability=high) on node"
```

## Validation and Verification

### Verify Plan Changes

```bash
# Check plan configuration after patching
kubectl get plan production-migration -o yaml

# Verify specific fields
kubectl get plan production-migration -o jsonpath='{.spec.targetNamespace}'

# Check VM configurations
kubectl get plan production-migration -o jsonpath='{.spec.vms[*].name}'
```

### Monitor Patch Results

```bash
# Watch plan status after patching
kubectl get plan production-migration -w

# Check plan conditions
kubectl describe plan production-migration | grep -A 5 Conditions

# Verify VM-specific changes
kubectl get plan production-migration -o jsonpath='{.spec.vms[?(@.name=="database-01")].hooks}'
```

## Common Patching Scenarios and Solutions

### Problem: Migration Too Slow

**Solution: Optimize with Patching**

```bash
# Add high-performance convertor scheduling
kubectl mtv patch plan slow-migration \
  --convertor-node-selector "node-type=high-io,network=10gbe" \
  --convertor-affinity "REQUIRE nodes(storage-tier=premium) on node"

# Switch to warm migration for reduced downtime
kubectl mtv patch plan slow-migration \
  --migration-type warm
```

### Problem: VM Naming Conflicts

**Solution: Update Naming Templates**

```bash
# Fix naming conflicts with better templates
kubectl mtv patch plan naming-conflict \
  --pvc-name-template "{% raw %}{{.PlanName}}{% endraw %}-{% raw %}{{.TargetVmName}}{% endraw %}-{% raw %}{{.DiskIndex}}{% endraw %}"

# Individual VM name fixes
kubectl mtv patch planvm naming-conflict conflicting-vm \
  --target-name unique-vm-name-prod
```

### Problem: Missing Automation

**Solution: Add Hooks Incrementally**

```bash
# Add hooks to specific VMs needing automation
kubectl mtv patch planvm manual-migration database-server \
  --add-pre-hook backup-automation \
  --add-post-hook validation-automation
```

### Problem: Security Compliance

**Solution: Add Security Configurations**

```bash
# Add LUKS support for encrypted VMs
kubectl mtv patch planvm secure-migration encrypted-vm \
  --luks-secret encryption-keys-secret

# Configure secure target placement
kubectl mtv patch plan secure-migration \
  --target-node-selector "security-level=high,compliance=pci-dss"
```

## Next Steps

After mastering plan patching techniques:

1. **Execute Migrations**: Learn to start and manage migration execution in [Chapter 19: Plan Lifecycle Execution](/kubectl-mtv/19-plan-lifecycle-execution)
2. **Handle Problems**: Master troubleshooting in [Chapter 20: Debugging and Troubleshooting](/kubectl-mtv/20-debugging-and-troubleshooting)
3. **Optimize Operations**: Learn best practices in [Chapter 21: Best Practices and Security](/kubectl-mtv/21-best-practices-and-security)
4. **AI Integration**: Explore advanced automation in [Chapter 22: Model Context Protocol (MCP) Server Integration](/kubectl-mtv/22-model-context-protocol-mcp-server-integration)

---

*Previous: [Chapter 17: Migration Hooks](/kubectl-mtv/17-migration-hooks)*  
*Next: [Chapter 19: Plan Lifecycle Execution](/kubectl-mtv/19-plan-lifecycle-execution)*
