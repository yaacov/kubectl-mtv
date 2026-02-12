---
layout: page
title: "Chapter 25: Settings Management"
---

The `settings` command lets you view and configure ForkliftController settings directly from the command line. You can inspect current values, enable feature flags, tune performance parameters, adjust container resource limits, and revert any setting to its default. All command examples are verified against the implementation.

## Overview

ForkliftController configuration is stored in the ForkliftController custom resource. The `settings` command provides a convenient interface to read and write individual settings without manually editing YAML.

Settings are organized into categories:

| **Category** | **Description** |
|--------------|-----------------|
| image | Container images (VDDK, virt-v2v, custom FQINs) |
| feature | Feature flags (warm migration, copy offload, live migration) |
| performance | Performance tuning (max concurrent VMs, precopy interval, timeouts) |
| debug | Debugging (controller log level) |
| virt-v2v | virt-v2v container resource limits and extra arguments |
| populator | Volume populator container resources |
| hook | Hook container resources |
| ova | OVA provider server container resources |
| hyperv | HyperV provider server container resources |

By default, only the curated **supported settings** are shown. Use `--all` to include the full set of **extended settings** (controller, inventory, API, UI plugin, validation, CLI download, OVA proxy, ConfigMaps, and advanced settings).

## Subcommands

### View Settings

```bash
# View all supported settings (default)
kubectl mtv settings

# Equivalent explicit form
kubectl mtv settings get
```

### Get a Specific Setting

```bash
kubectl mtv settings get --setting controller_max_vm_inflight
```

### Set a Setting Value

```bash
kubectl mtv settings set --setting controller_max_vm_inflight --value 30
```

### Unset a Setting (Revert to Default)

```bash
kubectl mtv settings unset --setting controller_max_vm_inflight
```

## Flags

| **Flag** | **Short** | **Default** | **Description** |
|----------|-----------|-------------|-----------------|
| `--output` | `-o` | `table` | Output format: `table`, `json`, `yaml` |
| `--all` | | `false` | Include all settings (supported + extended) |

## Commonly Used Settings

### Feature Flags

| **Setting** | **Type** | **Default** | **Description** |
|-------------|----------|-------------|-----------------|
| `controller_vsphere_incremental_backup` | bool | true | Enable CBT-based warm migration for vSphere |
| `controller_ovirt_warm_migration` | bool | true | Enable warm migration from oVirt |
| `feature_copy_offload` | bool | true | Enable storage array offload (XCOPY) |
| `feature_ocp_live_migration` | bool | false | Enable cross-cluster OpenShift live migration |
| `feature_vmware_system_serial_number` | bool | true | Use VMware system serial number for migrated VMs |
| `feature_ova_appliance_management` | bool | false | Enable appliance management for OVF-based providers |

### Performance Tuning

| **Setting** | **Type** | **Default** | **Description** |
|-------------|----------|-------------|-----------------|
| `controller_max_vm_inflight` | int | 20 | Maximum concurrent VM migrations |
| `controller_precopy_interval` | int | 60 | Minutes between warm migration precopies |
| `controller_max_concurrent_reconciles` | int | 10 | Maximum concurrent controller reconciles |
| `controller_snapshot_removal_timeout_minuts` | int | 120 | Timeout for snapshot removal (minutes) |
| `controller_filesystem_overhead` | int | 10 | Filesystem overhead percentage |

### Container Images

| **Setting** | **Type** | **Default** | **Description** |
|-------------|----------|-------------|-----------------|
| `vddk_image` | string | (empty) | VDDK container image for vSphere migrations |
| `virt_v2v_image_fqin` | string | (empty) | Custom virt-v2v container image |

### Container Resources (virt-v2v)

| **Setting** | **Type** | **Default** | **Description** |
|-------------|----------|-------------|-----------------|
| `virt_v2v_container_limits_cpu` | string | 4000m | virt-v2v container CPU limit |
| `virt_v2v_container_limits_memory` | string | 8Gi | virt-v2v container memory limit |
| `virt_v2v_container_requests_cpu` | string | 1000m | virt-v2v container CPU request |
| `virt_v2v_container_requests_memory` | string | 1Gi | virt-v2v container memory request |

### Debugging

| **Setting** | **Type** | **Default** | **Description** |
|-------------|----------|-------------|-----------------|
| `controller_log_level` | int | 3 | Controller log verbosity (0-9) |

## Supported vs Extended Settings

The `settings` command distinguishes between two tiers:

- **Supported settings** (default): A curated subset of commonly configured settings. These are shown by `kubectl mtv settings` and `kubectl mtv settings get`.
- **Extended settings** (`--all`): The full set of all known ForkliftController spec fields, including controller, inventory, API, UI plugin, validation, CLI download, and OVA proxy container resources, ConfigMap names, and advanced options.

```bash
# Show only supported settings
kubectl mtv settings

# Show all settings including extended ones
kubectl mtv settings --all

# Get a specific extended setting
kubectl mtv settings get --setting inventory_container_limits_memory --all
```

Both `settings set` and `settings unset` work with any valid setting name (supported or extended).

## Example Workflows

### View All Current Settings

```bash
kubectl mtv settings
```

### Check a Specific Value

```bash
kubectl mtv settings get --setting controller_max_vm_inflight
```

### Enable a Feature Flag

```bash
kubectl mtv settings set --setting feature_ocp_live_migration --value true
```

### Increase Concurrent VM Migrations

```bash
kubectl mtv settings set --setting controller_max_vm_inflight --value 40
```

### Increase virt-v2v Memory for Large VMs

```bash
kubectl mtv settings set --setting virt_v2v_container_limits_memory --value 16Gi
kubectl mtv settings set --setting virt_v2v_container_requests_memory --value 4Gi
```

### Raise Controller Log Level for Debugging

```bash
kubectl mtv settings set --setting controller_log_level --value 5
```

### Revert a Setting to Its Default

```bash
kubectl mtv settings unset --setting controller_max_vm_inflight
```

### Export Settings as JSON

```bash
kubectl mtv settings --output json
```

## Next Steps

After configuring settings:

1. **Verify System Health**: Confirm the system is operating correctly in [Chapter 24: System Health Checks](/kubectl-mtv/24-system-health-checks)
2. **Review Command Reference**: See all available commands in [Chapter 26: Command Reference](/kubectl-mtv/26-command-reference)

---

*Previous: [Chapter 24: System Health Checks](/kubectl-mtv/24-system-health-checks)*
*Next: [Chapter 26: Command Reference](/kubectl-mtv/26-command-reference)*
