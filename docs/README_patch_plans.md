# Patching Migration Plans

This guide explains how to modify existing migration plans using the `kubectl-mtv patch plan` and `kubectl-mtv patch planvm` commands. These commands allow you to update plan configurations and individual VM settings without recreating the entire migration plan.

## Overview

The patch plan functionality provides two distinct commands:

- **`patch plan`**: Updates plan-level settings like migration type, target configuration, and transfer networks
- **`patch planvm`**: Modifies individual VM configurations within a plan's VM list

**Protected Fields**: Source/target providers, network/storage mappings, and the VM list itself cannot be changed through `patch plan`. Use `patch planvm` for VM-specific modifications.

## Patch Plan Command

### Basic Syntax

```bash
kubectl-mtv patch plan PLAN_NAME [flags]
```

### Editable Plan Fields

| Flag | Description | Example Values |
|------|-------------|----------------|
| `--transfer-network` | Network for VM data transfer | `network-name` or `namespace/network-name` |
| `--install-legacy-drivers` | Install legacy drivers | `true`, `false` |
| `--migration-type` | Type of migration | `cold`, `warm`, `live`, `conversion` |
| `--target-labels` | Labels for target VMs | `env=prod,team=platform` |
| `--target-node-selector` | Node selector for target VMs | `node-type=compute,zone=us-east` |
| `--use-compatibility-mode` | Enable compatibility mode | `true`, `false` |
| `--target-affinity` | Target VM affinity (KARL syntax) | `REQUIRE pods(app=database) on node` |
| `--target-namespace` | Target namespace for VMs | `production` |
| `--description` | Plan description | `Production migration batch 1` |
| `--preserve-cluster-cpu-model` | Preserve CPU model from oVirt cluster | `true`, `false` |
| `--preserve-static-ips` | Preserve static IPs in vSphere | `true`, `false` |
| `--pvc-name-template` | Template for PVC names | `{{.VmName}}-disk-{{.DiskIndex}}` |
| `--volume-name-template` | Template for volume interface names | `{{.PVCName}}-vol{{.VolumeIndex}}` |
| `--network-name-template` | Template for network interface names | `{{.VmName}}-net{{.NetworkIndex}}` |
| `--migrate-shared-disks` | Migrate shared disks | `true`, `false` |
| `--archived` | Archive the plan | `true`, `false` |
| `--pvc-name-template-use-generate-name` | Use generateName for PVCs | `true`, `false` |
| `--delete-guest-conversion-pod` | Delete conversion pod after migration | `true`, `false` |
| `--skip-guest-conversion` | Skip guest conversion process | `true`, `false` |
| `--warm` | Enable warm migration (legacy flag) | `true`, `false` |
| `--target-power-state` | Target power state for VMs after migration | `on`, `off`, `auto` |
| `--delete-vm-on-fail-migration` | Delete target VM when migration fails | `true`, `false` |

### Usage Examples

#### Update Migration Type and Target Settings

```bash
# Change from cold to warm migration
kubectl-mtv patch plan my-migration-plan \
  --migration-type warm \
  --target-namespace production

# Enable compatibility mode
kubectl-mtv patch plan legacy-systems \
  --use-compatibility-mode=true \
  --install-legacy-drivers=true
```

#### Configure Transfer Network

```bash
# Use network in current namespace
kubectl-mtv patch plan my-plan \
  --transfer-network fast-network

# Use network in specific namespace
kubectl-mtv patch plan my-plan \
  --transfer-network openshift-sriov-network/high-speed-net
```

#### Set Target VM Configuration

```bash
# Apply labels and node selector to all VMs
kubectl-mtv patch plan production-migration \
  --target-labels "env=production,tier=web" \
  --target-node-selector "node-type=compute,disk=ssd"

# Multiple labels and selectors
kubectl-mtv patch plan my-plan \
  --target-labels "app=web,version=v2" \
  --target-labels "team=platform" \
  --target-node-selector "zone=us-east-1a,instance-type=m5.large"
```

#### Configure Advanced Affinity

```bash
# Set pod affinity using KARL syntax
kubectl-mtv patch plan my-plan \
  --target-affinity 'REQUIRE pods(app=database) on node'

# Zone-based pod affinity (soft constraint)
kubectl-mtv patch plan production-plan \
  --target-affinity 'PREFER pods(tier=web) on zone'

# Pod anti-affinity rule
kubectl-mtv patch plan distributed-app \
  --target-affinity 'AVOID pods(app=web) on node'
```

#### Configure Plan Templates and Settings

```bash
# Set naming templates for all VMs in the plan
kubectl-mtv patch plan production-migration \
  --pvc-name-template "prod-{{.VmName}}-disk-{{.DiskIndex}}" \
  --volume-name-template "vol-{{.VolumeIndex}}-{{.PVCName}}" \
  --network-name-template "{{.VmName}}-net{{.NetworkIndex}}"

# Configure vSphere-specific settings
kubectl-mtv patch plan vsphere-migration \
  --preserve-static-ips=true \
  --preserve-cluster-cpu-model=true

# Configure oVirt-specific settings  
kubectl-mtv patch plan ovirt-migration \
  --preserve-cluster-cpu-model=true \
  --migrate-shared-disks=false

# Set plan metadata and behavior
kubectl-mtv patch plan legacy-app-migration \
  --description "Legacy application migration - batch 2" \
  --skip-guest-conversion=true \
  --use-compatibility-mode=true \
  --delete-guest-conversion-pod=true
```

#### Configure PVC and Storage Settings

```bash
# Configure PVC generation behavior
kubectl-mtv patch plan storage-migration \
  --pvc-name-template "{{.VmName}}-storage-{{.DiskIndex}}" \
  --pvc-name-template-use-generate-name=false

# Archive completed plans
kubectl-mtv patch plan completed-migration \
  --archived=true \
  --description "Completed migration - archived for records"

# Configure power management and failure handling
kubectl-mtv patch plan production-migration \
  --target-power-state on \
  --delete-vm-on-fail-migration=true \
  --description "Production migration with automatic cleanup on failure"
```

## Patch Plan VMs Command

### Basic Syntax

```bash
kubectl-mtv patch planvm PLAN_NAME VM_NAME [flags]
```

### Editable VM Fields

| Flag | Description | Example Values |
|------|-------------|----------------|
| `--target-name` | Custom VM name in target cluster | `production-web-server-01` |
| `--root-disk` | Primary boot disk identifier | `disk-1`, `hard-disk-1` |
| `--instance-type` | VM instance type override | `m5.large`, `Standard_D4s_v3` |
| `--pvc-name-template` | PVC naming template | `{{.VmName}}-disk-{{.DiskIndex}}` |
| `--volume-name-template` | Volume interface naming template | `{{.PVCName}}-vol{{.VolumeIndex}}` |
| `--network-name-template` | Network interface naming template | `{{.VmName}}-nic{{.NetworkIndex}}` |
| `--luks-secret` | Kubernetes Secret name containing LUKS disk decryption keys | `vm-encryption-keys` |
| `--add-pre-hook` | Add a pre-migration hook | `data-backup-hook` |
| `--add-post-hook` | Add a post-migration hook | `cleanup-hook` |
| `--remove-hook` | Remove a hook by name | `old-hook-name` |
| `--clear-hooks` | Remove all hooks from VM | `true`, `false` |
| `--target-power-state` | Target power state for this VM after migration | `on`, `off`, `auto` |
| `--delete-vm-on-fail-migration` | Delete target VM when migration fails (overrides plan-level) | `true`, `false` |

### Template Variables

The template fields support Go template syntax with different variables for each template type:

**PVC Name Template Variables:**
- `{{.VmName}}` - VM name
- `{{.PlanName}}` - Migration plan name
- `{{.DiskIndex}}` - Initial volume index of the disk
- `{{.WinDriveLetter}}` - Windows drive letter (lowercase, requires guest agent)
- `{{.RootDiskIndex}}` - Index of the root disk
- `{{.Shared}}` - True if volume is shared by multiple VMs
- `{{.FileName}}` - Source file name (vSphere only, requires guest agent)

**Volume Name Template Variables:**
- `{{.PVCName}}` - Name of the PVC mounted to the VM
- `{{.VolumeIndex}}` - Sequential index of volume interface (0-based)

**Network Name Template Variables:**
- `{{.NetworkName}}` - Multus network attachment definition name (if applicable)
- `{{.NetworkNamespace}}` - Namespace of network attachment definition (if applicable)
- `{{.NetworkType}}` - Network type ("Multus" or "Pod")
- `{{.NetworkIndex}}` - Sequential index of network interface (0-based)

### Usage Examples

#### Update VM Target Configuration

```bash
# Set custom target name and instance type
kubectl-mtv patch planvm production-plan web-server-vm \
  --target-name production-web-01 \
  --instance-type m5.xlarge

# Configure boot disk
kubectl-mtv patch planvm my-plan database-vm \
  --root-disk "hard-disk-1" \
  --instance-type memory-optimized
```

#### Configure Naming Templates

```bash
# Set PVC naming template
kubectl-mtv patch planvm my-plan app-server \
  --pvc-name-template "{{.VmName}}-storage-{{.DiskIndex}}"

# Configure all naming templates
kubectl-mtv patch planvm production-plan web-vm \
  --pvc-name-template "prod-{{.VmName}}-disk-{{.DiskIndex}}" \
  --volume-name-template "prod-vol{{.VolumeIndex}}-{{.PVCName}}" \
  --network-name-template "prod-{{.VmName}}-net{{.NetworkIndex}}"
```

#### Set Encryption Configuration

```bash
# Configure LUKS disk encryption
kubectl-mtv patch planvm secure-plan encrypted-vm \
  --luks-secret disk-encryption-keys \
  --target-name secure-production-vm
```

#### Combined VM Updates

```bash
# Update multiple VM settings at once
kubectl-mtv patch planvm enterprise-migration critical-app-vm \
  --target-name prod-critical-app-01 \
  --instance-type c5.2xlarge \
  --root-disk "disk-0" \
  --pvc-name-template "{{.VmName}}-disk{{.DiskIndex}}-storage" \
  --luks-secret app-encryption-keys

# Configure power management and failure handling for specific VM
kubectl-mtv patch planvm production-migration critical-database \
  --target-name prod-db-primary \
  --target-power-state on \
  --delete-vm-on-fail-migration=false
```

#### Manage VM Hooks

```bash
# Add pre-migration hook for data backup
kubectl-mtv patch planvm production-plan database-vm \
  --add-pre-hook backup-database-hook

# Add post-migration hook for cleanup
kubectl-mtv patch planvm production-plan web-server-vm \
  --add-post-hook cleanup-temp-files-hook

# Add both pre and post hooks
kubectl-mtv patch planvm critical-migration app-server \
  --add-pre-hook stop-services-hook \
  --add-post-hook start-services-hook

# Remove a specific hook
kubectl-mtv patch planvm production-plan legacy-vm \
  --remove-hook old-migration-hook

# Clear all hooks from a VM
kubectl-mtv patch planvm test-migration test-vm \
  --clear-hooks

# Combined hook and configuration update
kubectl-mtv patch planvm enterprise-migration database-vm \
  --target-name prod-db-primary \
  --instance-type memory-optimized \
  --add-pre-hook database-snapshot-hook \
  --add-post-hook database-verify-hook \
  --luks-secret database-encryption-keys
```

## Advanced Usage Patterns

### Conditional Updates

```bash
# Only update if specific conditions are met
if kubectl-mtv get plan my-plan -o yaml | grep -q "type: cold"; then
  kubectl-mtv patch plan my-plan --migration-type warm
  echo "Upgraded to warm migration"
fi
```

### Batch VM Updates

```bash
#!/bin/bash
# Update multiple VMs in a plan with consistent naming
PLAN_NAME="production-migration"
VMS=("web-01" "web-02" "api-server" "database")

for vm in "${VMS[@]}"; do
  kubectl-mtv patch planvm "$PLAN_NAME" "$vm" \
    --target-name "prod-${vm}" \
    --pvc-name-template "prod-{{.VmName}}-disk{{.DiskIndex}}" \
    --instance-type "m5.large"
done
```

### Migration Type Progression

```bash
# Upgrade migration strategy progressively
kubectl-mtv patch plan test-migration --migration-type warm
echo "Testing warm migration..."

# After validation, update production settings
kubectl-mtv patch plan production-migration \
  --migration-type warm \
  --target-namespace production \
  --target-labels "env=production,validated=true"
```

## Best Practices

### 1. Plan-Level vs VM-Level Changes

**Use `patch plan` for:**
- Migration strategy changes (cold/warm)
- Target environment configuration
- Network and infrastructure settings
- Labels and selectors affecting all VMs

**Use `patch planvm` for:**
- Individual VM customization
- VM-specific resource requirements
- Custom naming schemes
- Encryption configuration

### 2. Template Design

```yaml
# Good: Descriptive and consistent
pvc-name-template: "{{.VmName}}-storage-{{.DiskIndex}}"
volume-name-template: "{{.PVCName}}-vol{{.VolumeIndex}}"

# Avoid: Generic or conflicting names
pvc-name-template: "pvc-{{.DiskIndex}}"  # Too generic
```

### 3. Namespace Management

```bash
# Explicit namespace specification
kubectl-mtv patch plan my-plan \
  --target-namespace production \
  --transfer-network "openshift-sriov/fast-net"

# Verify target namespace exists
kubectl get namespace production || kubectl create namespace production
```

### 4. Validation Workflow

```bash
# 1. Check current plan configuration
kubectl-mtv describe plan my-plan

# 2. Apply changes
kubectl-mtv patch plan my-plan --migration-type warm

# 3. Verify changes
kubectl-mtv get plan my-plan -o yaml | grep -A5 "spec:"
```

## Error Handling

### Common Errors and Solutions

#### VM Not Found in Plan
```bash
Error: VM 'missing-vm' not found in plan 'my-plan'
```
**Solution**: List VMs in the plan first:
```bash
kubectl-mtv get plan my-plan -o yaml | grep -A10 "vms:"
```

#### Invalid Migration Type
```bash
Error: invalid migration type 'hot' (must be 'cold', 'warm', 'live', or 'conversion')
```
**Solution**: Use valid migration types with tab completion:
```bash
kubectl-mtv patch plan my-plan --migration-type <TAB>
```

#### Invalid Target Affinity KARL Rule
```bash
Error: failed to parse target affinity KARL rule: syntax error at position 15
```
**Solution**: Validate KARL syntax before applying:
```bash
# Test KARL rule syntax
kubectl-mtv patch plan my-plan \
  --target-affinity 'REQUIRE pods(app=database) on node' \
  --dry-run

# Common KARL syntax patterns
kubectl-mtv patch plan my-plan \
  --target-affinity 'PREFER pods(tier=frontend) on zone'
```

#### Transfer Network Not Found
```bash
Error: NetworkAttachmentDefinition "network-name" not found
```
**Solution**: Verify network exists in correct namespace:
```bash
kubectl get network-attachment-definitions -n openshift-sriov-network
```

## Integration with Migration Workflow

### Pre-Migration Updates

```bash
# Prepare plan for production migration
kubectl-mtv patch plan staging-to-prod \
  --migration-type warm \
  --target-namespace production \
  --target-labels "env=production,migration-batch=1"

# Configure critical VMs
kubectl-mtv patch planvm staging-to-prod database-server \
  --target-name prod-db-primary \
  --instance-type memory-optimized \
  --luks-secret database-encryption
```

### During Migration Monitoring

```bash
# Switch to cold migration if warm migration issues occur
kubectl-mtv patch plan active-migration --migration-type cold

# Update individual VM if needed
kubectl-mtv patch planvm active-migration problematic-vm \
  --instance-type smaller-instance
```

### Post-Migration Cleanup

```bash
# Update plan for next batch after successful migration
kubectl-mtv patch plan next-batch \
  --target-labels "env=production,migration-batch=2,validated=true" \
  --use-compatibility-mode=false
```

## Troubleshooting

### Debug Plan Changes

```bash
# Enable verbose logging
kubectl-mtv patch plan my-plan --migration-type warm -v=3

# Check plan status after changes
kubectl-mtv describe plan my-plan
```

### Validate Template Syntax

```bash
# Test template rendering (conceptual)
kubectl-mtv patch planvm my-plan test-vm \
  --pvc-name-template "test-{{.VmName}}-disk{{.DiskIndex}}" \
  --dry-run
```

### Rollback Changes

```bash
# Plans don't have automatic rollback, revert manually
kubectl-mtv patch plan my-plan --migration-type cold  # Revert to cold
kubectl-mtv patch plan my-plan --use-compatibility-mode=false  # Revert compatibility mode
```

## Related Commands

- `kubectl-mtv get plan` - List and view migration plans
- `kubectl-mtv describe plan` - Get detailed plan information  
- `kubectl-mtv create plan` - Create new migration plans
- `kubectl-mtv start plan` - Begin plan execution
- `kubectl-mtv cancel plan` - Cancel running plans

---

For more information about migration plans and VM configuration, see:
- [Plan VMs Guide](README_planvms.md)
- [Main Usage Guide](README-usage.md)
- [Migration Workflow Documentation](README_demo.md) 