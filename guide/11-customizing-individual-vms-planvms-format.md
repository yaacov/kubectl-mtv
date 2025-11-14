# Chapter 11: Customizing Individual VMs (PlanVMS Format)

The PlanVMS format provides granular control over individual VM migration settings, allowing customization of target names, disk configuration, networking, and migration behavior on a per-VM basis. This chapter covers the complete VM customization capabilities.

## Overview of PlanVMS Format

### What is PlanVMS Format?

PlanVMS format is a structured YAML/JSON format that defines individual VM configurations within a migration plan. It enables:

- **Per-VM Customization**: Different settings for each VM in the same plan
- **Resource Templates**: Custom naming templates for generated resources
- **Migration Behavior**: Individual migration settings and cleanup policies
- **Target Configuration**: Specific target cluster settings per VM

### When to Use PlanVMS Format

- **Complex Migrations**: When VMs require different target configurations
- **Name Normalization**: When source VM names need target-specific adjustments
- **Resource Management**: When custom resource naming is required
- **Hook Integration**: When different VMs need different automation hooks
- **Security Requirements**: When VMs have different encryption or security needs

## Detailed VM List Format

The PlanVMS format is based on the Forklift API VM specification, verified from the vendor code:

### Basic VM Structure

```yaml
# Basic VM entry
- name: source-vm-name            # Required: Source VM name
  targetName: target-vm-name      # Optional: Custom target name
  rootDisk: /dev/sda             # Optional: Boot disk selection
```

### Complete VM Structure

```yaml
- name: web-server-01
  targetName: web-prod-01
  rootDisk: /dev/sda
  instanceType: web-server
  targetPowerState: on
  deleteVmOnFailMigration: false
  nbdeClevis: false
  luks:
    name: encryption-keys
    namespace: security
  hooks:
  - step: PreHook
    hook:
      name: backup-hook
      namespace: migration-hooks
  - step: PostHook
    hook:
      name: validation-hook
      namespace: migration-hooks
  pvcNameTemplate: "{{.TargetVmName}}-{{.DiskIndex}}"
  volumeNameTemplate: "vol-{{.VolumeIndex}}"
  networkNameTemplate: "net-{{.NetworkIndex}}"
```

## Editable Fields for Customization

All fields are verified from the Forklift API VM struct definition:

### Core VM Configuration

#### Name and Identity

```yaml
# Required source VM identifier
- name: source-vm-name

# Optional custom target name
  targetName: custom-target-name
```

**Name Field Requirements:**
- `name`: Must match exactly the VM name in the source provider
- `targetName`: Must be valid Kubernetes resource name (DNS-1123 compliant)

#### Target Power State

```yaml
# Control VM power state after migration
  targetPowerState: on    # Options: on, off, auto (default)
```

**Power State Options:**
- `on`: Start VM after migration completes
- `off`: Keep VM powered off after migration
- `auto`: Match source VM power state (default behavior)

#### Instance Type Override

```yaml
# Override VM resource specifications
  instanceType: high-performance
```

Selects a predefined InstanceType resource that overrides CPU, memory, and other VM properties.

### Disk and Storage Configuration

#### Root Disk Selection

```yaml
# Specify the primary boot disk
  rootDisk: /dev/sda
```

Critical for multi-disk VMs to ensure proper boot configuration.

#### LUKS Disk Encryption

```yaml
# Reference to LUKS encryption keys
  luks:
    name: vm-encryption-keys
    namespace: security-namespace

# Enable automatic TANG/Clevis unlocking
  nbdeClevis: true
```

**LUKS Configuration:**
- `luks`: References a Kubernetes Secret containing LUKS passphrases
- `nbdeClevis`: Enables network-based automatic unlocking using TANG servers

### Migration Behavior Configuration

#### Failure Cleanup Policy

```yaml
# Control VM deletion on migration failure
  deleteVmOnFailMigration: true
```

**Cleanup Behavior:**
- `true`: Delete target VM and resources if migration fails
- `false`: Preserve target VM for troubleshooting (default)

Note: Plan-level setting overrides VM-level setting when enabled.

### Hook Integration

```yaml
# Attach migration hooks to specific VMs
  hooks:
  - step: PreHook
    hook:
      name: database-quiesce
      namespace: migration-hooks
  - step: PostHook
    hook:
      name: health-check
      namespace: migration-hooks
```

**Hook Configuration:**
- `step`: Hook execution phase (PreHook, PostHook)
- `hook.name`: Name of the Hook resource
- `hook.namespace`: Namespace containing the Hook resource

## Go Template Variables Reference

kubectl-mtv provides rich template variables for resource naming, verified from the API documentation:

### PVC Name Template Variables

Available in `pvcNameTemplate` field:

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `{{.VmName}}` | Original source VM name | `web-server-01` |
| `{{.TargetVmName}}` | Final target VM name | `web-prod-01` |
| `{{.PlanName}}` | Migration plan name | `production-migration` |
| `{{.DiskIndex}}` | Disk index (0-based) | `0`, `1`, `2` |
| `{{.WinDriveLetter}}` | Windows drive letter | `c`, `d`, `e` |
| `{{.RootDiskIndex}}` | Index of root/boot disk | `0` |
| `{{.Shared}}` | True if disk is shared | `true`, `false` |
| `{{.FileName}}` | VMware disk filename | `web-server-01.vmdk` |

#### PVC Template Examples

```yaml
# Basic PVC naming
  pvcNameTemplate: "{{.TargetVmName}}-disk-{{.DiskIndex}}"
  # Result: web-prod-01-disk-0, web-prod-01-disk-1

# Root vs data disk differentiation  
  pvcNameTemplate: "{{if eq .DiskIndex .RootDiskIndex}}{{.TargetVmName}}-root{{else}}{{.TargetVmName}}-data-{{.DiskIndex}}{{end}}"
  # Result: web-prod-01-root, web-prod-01-data-1

# Shared disk identification
  pvcNameTemplate: "{{if .Shared}}shared-{{end}}{{.TargetVmName}}-{{.DiskIndex}}"
  # Result: web-prod-01-0, shared-web-prod-01-1

# Windows drive letter naming
  pvcNameTemplate: "{{.TargetVmName}}-{{.WinDriveLetter}}-drive"
  # Result: windows-server-c-drive, windows-server-d-drive

# Plan-scoped naming
  pvcNameTemplate: "{{.PlanName}}-{{.TargetVmName}}-disk{{.DiskIndex}}"
  # Result: prod-migration-web-prod-01-disk0
```

### Volume Name Template Variables

Available in `volumeNameTemplate` field:

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `{{.PVCName}}` | Generated PVC name | `web-prod-01-disk-0` |
| `{{.VolumeIndex}}` | Volume interface index | `0`, `1`, `2` |

#### Volume Template Examples

```yaml
# Simple volume naming
  volumeNameTemplate: "disk-{{.VolumeIndex}}"
  # Result: disk-0, disk-1, disk-2

# PVC-based naming
  volumeNameTemplate: "vol-{{.PVCName}}"
  # Result: vol-web-prod-01-disk-0

# Combined indexing
  volumeNameTemplate: "{{.VolumeIndex}}-{{.PVCName}}"
  # Result: 0-web-prod-01-disk-0
```

### Network Name Template Variables

Available in `networkNameTemplate` field:

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `{{.NetworkName}}` | Multus network name | `production-net` |
| `{{.NetworkNamespace}}` | Network namespace | `multus-system` |
| `{{.NetworkType}}` | Network type | `Multus`, `Pod` |
| `{{.NetworkIndex}}` | Interface index | `0`, `1`, `2` |

#### Network Template Examples

```yaml
# Simple interface naming
  networkNameTemplate: "net-{{.NetworkIndex}}"
  # Result: net-0, net-1, net-2

# Type-based naming
  networkNameTemplate: "{{if eq .NetworkType \"Pod\"}}pod-net{{else}}multus-{{.NetworkIndex}}{{end}}"
  # Result: pod-net, multus-1, multus-2

# Network-specific naming
  networkNameTemplate: "{{.NetworkType}}-{{.NetworkName}}-{{.NetworkIndex}}"
  # Result: Multus-production-net-0
```

## How-To: Editing the List

### Method 1: Create PlanVMS File from Scratch

#### Basic VM List Creation

```yaml
# Save as vm-customizations.yaml
- name: web-server-01
  targetName: web-prod-01
  targetPowerState: on
  pvcNameTemplate: "{{.TargetVmName}}-{{.DiskIndex}}"

- name: database-01
  targetName: db-prod-01
  rootDisk: /dev/sda
  targetPowerState: on
  instanceType: database-server
  deleteVmOnFailMigration: false

- name: app-server-01
  targetName: app-prod-01
  hooks:
  - step: PreHook
    hook:
      name: app-shutdown
      namespace: migration-hooks
  - step: PostHook
    hook:
      name: app-startup
      namespace: migration-hooks
```

#### Use Custom VM List in Plan

```bash
kubectl mtv create plan custom-vm-migration \
  --source vsphere-prod \
  --vms @vm-customizations.yaml \
  --network-mapping prod-network-map \
  --storage-mapping prod-storage-map
```

### Method 2: Export and Modify Existing Inventory

#### Export VMs in PlanVMS Format

```bash
# Export all VMs from provider
kubectl mtv get inventory vms vsphere-prod -o planvms > all-vms.yaml

# Export filtered VMs
kubectl mtv get inventory vms vsphere-prod \
  -q "where powerState = 'poweredOn' and memoryMB >= 4096" \
  -o planvms > production-vms.yaml
```

#### Modify Exported VMs

```yaml
# Original exported format
- name: web-server-01
  targetName: ""
  rootDisk: ""

# Modified with customizations
- name: web-server-01
  targetName: web-prod-01
  rootDisk: /dev/sda
  targetPowerState: on
  pvcNameTemplate: "prod-{{.TargetVmName}}-disk-{{.DiskIndex}}"
  hooks:
  - step: PostHook
    hook:
      name: web-validation
      namespace: migration-hooks
```

### Method 3: Template-Based Mass Customization

#### Create Template for Similar VMs

```yaml
# Template for web servers
- name: web-server-01
  targetName: web-prod-01
  targetPowerState: on
  pvcNameTemplate: "web-{{.TargetVmName}}-{{.DiskIndex}}"
  volumeNameTemplate: "vol-{{.VolumeIndex}}"
  hooks:
  - step: PostHook
    hook:
      name: web-health-check
      namespace: migration-hooks

- name: web-server-02
  targetName: web-prod-02
  targetPowerState: on
  pvcNameTemplate: "web-{{.TargetVmName}}-{{.DiskIndex}}"
  volumeNameTemplate: "vol-{{.VolumeIndex}}"
  hooks:
  - step: PostHook
    hook:
      name: web-health-check
      namespace: migration-hooks
```

## Advanced Customization Scenarios

### Scenario 1: Database Cluster Migration

```yaml
# Database cluster with shared storage
- name: db-primary-01
  targetName: postgres-primary
  rootDisk: /dev/sda
  instanceType: database-primary
  targetPowerState: on
  pvcNameTemplate: "{{.TargetVmName}}-{{if .Shared}}shared-{{end}}{{.DiskIndex}}"
  luks:
    name: db-encryption-keys
    namespace: database-security
  hooks:
  - step: PreHook
    hook:
      name: database-backup
      namespace: db-hooks
  - step: PostHook
    hook:
      name: database-validate
      namespace: db-hooks

- name: db-replica-01
  targetName: postgres-replica-01
  rootDisk: /dev/sda
  instanceType: database-replica
  targetPowerState: on
  pvcNameTemplate: "{{.TargetVmName}}-{{if .Shared}}shared-{{end}}{{.DiskIndex}}"
  luks:
    name: db-encryption-keys
    namespace: database-security

- name: db-replica-02
  targetName: postgres-replica-02
  rootDisk: /dev/sda
  instanceType: database-replica
  targetPowerState: on
  pvcNameTemplate: "{{.TargetVmName}}-{{if .Shared}}shared-{{end}}{{.DiskIndex}}"
  luks:
    name: db-encryption-keys
    namespace: database-security
```

### Scenario 2: Windows Domain Migration

```yaml
# Windows domain controller
- name: dc01
  targetName: domain-controller-01
  rootDisk: /dev/sda
  targetPowerState: on
  instanceType: windows-server
  pvcNameTemplate: "{{.TargetVmName}}-{{.WinDriveLetter}}"
  volumeNameTemplate: "{{.WinDriveLetter}}-drive"
  hooks:
  - step: PreHook
    hook:
      name: ad-replication-pause
      namespace: windows-hooks
  - step: PostHook
    hook:
      name: ad-health-check
      namespace: windows-hooks

# Windows file server
- name: fileserver01
  targetName: file-server-01
  rootDisk: /dev/sda
  targetPowerState: on
  instanceType: file-server
  pvcNameTemplate: "{{.TargetVmName}}-{{.WinDriveLetter}}-{{if .Shared}}shared{{else}}local{{end}}"
  deleteVmOnFailMigration: false
```

### Scenario 3: Multi-Tier Application

```yaml
# Web tier
- name: web-lb-01
  targetName: web-loadbalancer
  targetPowerState: on
  instanceType: load-balancer
  pvcNameTemplate: "web-{{.TargetVmName}}-{{.DiskIndex}}"
  hooks:
  - step: PreHook
    hook:
      name: drain-connections
      namespace: web-hooks

# Application tier
- name: app-server-01
  targetName: app-primary
  targetPowerState: on
  instanceType: application-server
  pvcNameTemplate: "app-{{.TargetVmName}}-{{.DiskIndex}}"
  hooks:
  - step: PreHook
    hook:
      name: app-graceful-shutdown
      namespace: app-hooks
  - step: PostHook
    hook:
      name: app-health-check
      namespace: app-hooks

# Data tier
- name: cache-redis-01
  targetName: redis-cache
  targetPowerState: on
  instanceType: cache-server
  pvcNameTemplate: "cache-{{.TargetVmName}}-{{.DiskIndex}}"
  luks:
    name: cache-encryption
    namespace: security
```

### Scenario 4: Development Environment Normalization

```yaml
# Normalize development VM names
- name: "Dev Web Server 01"  # Source name with spaces
  targetName: dev-web-01      # Kubernetes-compliant name
  targetPowerState: on
  pvcNameTemplate: "dev-{{.TargetVmName | lower}}-disk{{.DiskIndex}}"

- name: "Test Database (MySQL)"
  targetName: test-mysql-db
  rootDisk: /dev/sda
  targetPowerState: on
  pvcNameTemplate: "test-{{.TargetVmName}}-{{if eq .DiskIndex .RootDiskIndex}}os{{else}}data{{end}}"

- name: "QA_Environment_App"
  targetName: qa-app-server
  targetPowerState: off  # Keep powered off initially
  deleteVmOnFailMigration: true  # Clean up failures in test env
```

## Template Functions and Advanced Usage

### Built-in Template Functions

kubectl-mtv supports Go template functions for advanced string manipulation:

#### String Functions

```yaml
# Lowercase conversion
  pvcNameTemplate: "{{.TargetVmName | lower}}-{{.DiskIndex}}"

# Replace characters
  pvcNameTemplate: "{{.VmName | replace \" \" \"-\" | lower}}-disk{{.DiskIndex}}"

# Conditional logic
  pvcNameTemplate: "{{if .Shared}}shared-{{else}}local-{{end}}{{.TargetVmName}}"
```

#### Complex Conditional Templates

```yaml
# Multi-condition PVC naming
  pvcNameTemplate: "{{if eq .DiskIndex .RootDiskIndex}}root{{else if .Shared}}shared-data{{else}}data{{end}}-{{.DiskIndex}}"

# Windows vs Linux differentiation
  volumeNameTemplate: "{{if .WinDriveLetter}}{{.WinDriveLetter}}-drive{{else}}disk-{{.VolumeIndex}}{{end}}"

# Network type-based naming
  networkNameTemplate: "{{if eq .NetworkType \"Pod\"}}pod{{else}}{{.NetworkName | lower}}{{end}}-{{.NetworkIndex}}"
```

## Validation and Testing

### PlanVMS Format Validation

#### Syntax Validation

```bash
# Validate YAML syntax
yamllint vm-customizations.yaml

# Test with kubectl dry-run
kubectl mtv create plan test-validation \
  --source vsphere-prod \
  --vms @vm-customizations.yaml \
  --dry-run=client
```

#### Template Testing

```bash
# Test template rendering (requires actual plan creation)
kubectl mtv create plan template-test \
  --source vsphere-prod \
  --vms @template-test.yaml \
  -n test-namespace

# Check generated resource names
kubectl get pvc -n test-namespace
kubectl describe vm template-test-vm -n test-namespace
```

### Field Validation

#### Required Field Check

```yaml
# Minimal valid VM entry
- name: source-vm-name  # Required

# Invalid: missing name
- targetName: target-only  # Error: name is required
```

#### Target Name Validation

```yaml
# Valid target names (DNS-1123 compliant)
- name: source-vm
  targetName: valid-vm-name-01

# Invalid target names
- name: source-vm
  targetName: "Invalid Name With Spaces"  # Error: invalid format
```

## Integration with Plan Creation

### Using PlanVMS in Migration Plans

```bash
# Create plan with custom VM configurations
kubectl mtv create plan customized-migration \
  --source vsphere-prod \
  --target openshift-prod \
  --network-mapping prod-network-map \
  --storage-mapping prod-storage-map \
  --migration-type warm \
  --vms @customized-vms.yaml

# Combine with plan-level settings
kubectl mtv create plan comprehensive-migration \
  --source vsphere-prod \
  --target-namespace production \
  --migration-type warm \
  --preserve-static-ips \
  --vms @enterprise-vms.yaml
```

### Template Override Behavior

Plan-level templates are overridden by VM-level templates:

```bash
# Plan-level template
kubectl mtv create plan plan-template \
  --pvc-name-template "{{.PlanName}}-{{.VmName}}-{{.DiskIndex}}" \
  --vms @vms-with-templates.yaml

# VM-level template overrides plan-level
# VMs with pvcNameTemplate: use VM template
# VMs without pvcNameTemplate: use plan template
```

## Troubleshooting PlanVMS Issues

### Common PlanVMS Errors

#### YAML Format Issues

```bash
# Check YAML syntax
python -c "import yaml; yaml.safe_load(open('vm-list.yaml'))"

# Validate with yq
yq eval . vm-list.yaml
```

#### Template Rendering Errors

```bash
# Check template variables
kubectl logs -n konveyor-forklift deployment/forklift-controller | grep template

# Validate generated names
kubectl get pvc,vm -n target-namespace --show-labels
```

#### Name Conflicts

```bash
# Check for duplicate target names
grep -n "targetName:" vm-list.yaml | sort -k2

# Verify uniqueness in target namespace
kubectl get vm -n target-namespace -o name
```

### Debug PlanVMS Processing

```bash
# Monitor plan creation with verbosity
kubectl mtv create plan debug-planvms \
  --vms @debug-vms.yaml \
  -v=2

# Check plan status
kubectl describe plan debug-planvms

# Monitor VM processing
kubectl get vmstatus -n migration-namespace --watch
```

## Best Practices for PlanVMS Usage

### Design Principles

1. **Consistency**: Use consistent naming patterns across similar VMs
2. **Clarity**: Make target names self-documenting
3. **Scalability**: Design templates that work for large VM sets
4. **Security**: Properly configure LUKS and encryption settings

### Operational Guidelines

1. **Version Control**: Store PlanVMS files in version control systems
2. **Documentation**: Document custom template logic and naming conventions  
3. **Testing**: Validate PlanVMS configurations in test environments first
4. **Monitoring**: Track resource usage of generated PVCs and volumes

### Template Design Best Practices

1. **Readable Names**: Generate human-readable resource names
2. **Unique Identifiers**: Ensure generated names are unique
3. **Length Limits**: Keep names under Kubernetes limits (63 characters)
4. **Special Characters**: Avoid special characters in generated names

## Next Steps

After mastering PlanVMS customization:

1. **Advanced Placement**: Learn VM placement strategies in [Chapter 12: Target VM Placement](12-target-vm-placement.md)
2. **Performance Optimization**: Apply customization insights in [Chapter 13: Migration Process Optimization](13-migration-process-optimization.md)
3. **Hook Development**: Create custom hooks in [Chapter 14: Migration Hooks](14-migration-hooks.md)
4. **Plan Patching**: Modify plans dynamically in [Chapter 15: Advanced Plan Patching](15-advanced-plan-patching.md)

---

*Previous: [Chapter 10: Migration Plan Creation](10-migration-plan-creation.md)*  
*Next: [Chapter 12: Target VM Placement](12-target-vm-placement.md)*
