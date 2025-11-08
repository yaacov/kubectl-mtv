# Storage Offload Configuration

This guide explains how to configure storage offload functionality for optimized migration performance using kubectl-mtv.

## Overview

Storage offload enables **direct storage array operations** during migration, bypassing traditional host-based copying for significantly improved performance. Instead of copying data through compute hosts, offload leverages storage array capabilities like XCOPY, cloning, or snapshot operations.

### Benefits

- **Performance**: Direct storage operations are much faster than host-based copying
- **Reduced Load**: Minimizes CPU, memory, and network usage on migration hosts
- **Efficiency**: Leverages hardware acceleration and optimized storage protocols
- **Vendor Support**: Works with enterprise storage arrays from major vendors

### How It Works

```
Traditional Migration:
Source Storage → Host → Network → Target Host → Target Storage
     ↑                                              ↓
  [Slow, resource intensive, network bottleneck]

Offloaded Migration:  
Source Storage ←→ Target Storage (Direct Array Communication)
     ↑                    ↓
  [Fast, efficient, hardware optimized]
```

## Supported Configuration

### Offload Plugins
- **`vsphere`**: Currently the only supported plugin (uses vSphere XCOPY operations)

### Supported Storage Vendors
- `flashsystem` - IBM FlashSystem
- `vantara` - Hitachi Vantara  
- `ontap` - NetApp ONTAP
- `primera3par` - HPE Primera/3PAR
- `pureFlashArray` - Pure Storage FlashArray
- `powerflex` - Dell PowerFlex
- `powermax` - Dell PowerMax  
- `powerstore` - Dell PowerStore
- `infinibox` - Infinidat InfiniBox

## Secret Requirements

Offload operations require authentication to **both** the vSphere environment and the storage array:

### Required Secret Fields

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-offload-secret
type: Opaque
data:
  # vSphere Credentials (required)
  user: <base64-encoded-vsphere-username>      # e.g. administrator@vsphere.local
  password: <base64-encoded-vsphere-password>  # vSphere password
  url: <base64-encoded-vcenter-url>            # e.g. https://vcenter.company.com

  # Storage Array Credentials (required)  
  storageUser: <base64-encoded-storage-username>     # Storage management username
  storagePassword: <base64-encoded-storage-password> # Storage management password
  storageEndpoint: <base64-encoded-storage-api-url>  # Storage management API URL

  # Optional TLS Configuration
  insecureSkipVerify: <base64-encoded-true-or-false> # Skip TLS verification
  cacert: <base64-encoded-ca-certificate>            # CA certificate for TLS
```

## Usage Methods

kubectl-mtv provides **two ways** to configure offload secrets:

### Method 1: Use Existing Secret

Reference a pre-created Kubernetes secret:

```bash
# Create secret manually first
kubectl create secret generic my-offload-secret \
  --from-literal=user=administrator@vsphere.local \
  --from-literal=password=vsphere-pass \
  --from-literal=url=https://vcenter.company.com \
  --from-literal=storageUser=storage-admin \
  --from-literal=storagePassword=storage-pass \
  --from-literal=storageEndpoint=https://powerstore.company.com

# Reference it in storage mapping
kubectl mtv create mapping storage my-storage \
  --source vsphere-provider \
  --target openshift-provider \
  --storage-pairs "datastore1:premium;offloadPlugin=vsphere;offloadVendor=powerstore" \
  --default-offload-secret my-offload-secret
```

### Method 2: Inline Secret Creation (New!)

Automatically create secrets during mapping/plan creation:

```bash
kubectl mtv create mapping storage my-storage \
  --source vsphere-provider \
  --target openshift-provider \
  --storage-pairs "datastore1:premium;offloadPlugin=vsphere;offloadVendor=powerstore" \
  --default-offload-plugin vsphere \
  --default-offload-vendor powerstore \
  --offload-vsphere-username administrator@vsphere.local \
  --offload-vsphere-password vsphere-pass \
  --offload-vsphere-url https://vcenter.company.com \
  --offload-storage-username storage-admin \
  --offload-storage-password storage-pass \
  --offload-storage-endpoint https://powerstore.company.com
```

## Inline Secret Creation Flags

When using inline secret creation, these flags are available:

### vSphere Authentication Flags
| Flag | Description | Example |
|------|-------------|---------|
| `--offload-vsphere-username` | vSphere username | `administrator@vsphere.local` |
| `--offload-vsphere-password` | vSphere password | `secret123` |
| `--offload-vsphere-url` | vCenter URL | `https://vcenter.company.com` |

### Storage Array Authentication Flags
| Flag | Description | Example |
|------|-------------|---------|
| `--offload-storage-username` | Storage management username | `storage-admin` |
| `--offload-storage-password` | Storage management password | `storage-pass` |
| `--offload-storage-endpoint` | Storage management API URL | `https://powerstore.company.com` |

### Optional TLS Flags
| Flag | Description | Example |
|------|-------------|---------|
| `--offload-cacert` | CA certificate file | `@/path/to/ca.cert` |
| `--offload-insecure-skip-tls` | Skip TLS verification | (boolean flag) |

### Automatic Secret Creation Logic

The system creates a secret automatically when:
1. **No existing secret specified** (`--default-offload-secret` not provided)
2. **Credentials provided** (any `--offload-*` credential flags used)

**Note**: Offload plugins can be configured either globally (`--default-offload-plugin`) or per-pair (`offloadPlugin=vsphere` in storage pairs). Secret creation works with both approaches.

## Storage Mapping Examples

### Basic Offload Configuration

```bash
kubectl mtv create mapping storage basic-offload \
  --source vsphere-provider \
  --target openshift-provider \
  --storage-pairs "production-datastore:premium-ssd;offloadPlugin=vsphere;offloadVendor=powerstore" \
  --default-offload-secret existing-powerstore-secret
```

### Per-Pair Offload with Inline Secrets

```bash
# Offload configured per-pair (no global --default-offload-plugin needed)
kubectl mtv create mapping storage per-pair-offload \
  --source vsphere-provider \
  --target openshift-provider \
  --storage-pairs "ds1:premium;offloadPlugin=vsphere;offloadVendor=powerstore,ds2:standard;offloadPlugin=vsphere;offloadVendor=infinibox" \
  --offload-vsphere-username admin@vsphere.local \
  --offload-vsphere-password vsphere-secret \
  --offload-vsphere-url https://vcenter.example.com \
  --offload-storage-username unified-storage-admin \
  --offload-storage-password unified-storage-pass \
  --offload-storage-endpoint https://storage-mgmt.example.com
```

### Multiple Vendors with Global Plugin

```bash
kubectl mtv create mapping storage multi-vendor \
  --source vsphere-provider \
  --target openshift-provider \
  --storage-pairs "ds1:premium;offloadVendor=powerstore,ds2:standard;offloadVendor=infinibox" \
  --default-offload-plugin vsphere \
  --offload-vsphere-username admin@vsphere.local \
  --offload-vsphere-password vsphere-secret \
  --offload-vsphere-url https://vcenter.example.com \
  --offload-storage-username unified-storage-admin \
  --offload-storage-password unified-storage-pass \
  --offload-storage-endpoint https://storage-mgmt.example.com
```

### Per-Pair Secret Configuration

```bash
kubectl mtv create mapping storage per-pair-secrets \
  --source vsphere-provider \
  --target openshift-provider \
  --storage-pairs "fast-ds:premium;offloadPlugin=vsphere;offloadSecret=powerstore-secret;offloadVendor=powerstore,archive-ds:standard;offloadPlugin=vsphere;offloadSecret=infinibox-secret;offloadVendor=infinibox"
```

## Plan Creation Examples

### Plan with Inline Offload Secret

```bash
kubectl mtv create plan production-migration \
  --source vsphere-prod \
  --target openshift-prod \
  --vms @production-vms.yaml \
  --storage-pairs "datastore1:premium;offloadPlugin=vsphere;offloadVendor=ontap" \
  --default-offload-plugin vsphere \
  --default-offload-vendor ontap \
  --offload-vsphere-username prod-admin@vsphere.local \
  --offload-vsphere-password prod-secret \
  --offload-vsphere-url https://vcenter-prod.company.com \
  --offload-storage-username ontap-admin \
  --offload-storage-password ontap-secret \
  --offload-storage-endpoint https://ontap-cluster.company.com
```

### Plan with Existing Secret

```bash
kubectl mtv create plan development-migration \
  --source vsphere-dev \
  --target openshift-dev \
  --vms vm-dev1,vm-dev2,vm-dev3 \
  --storage-pairs "dev-datastore:standard;offloadPlugin=vsphere;offloadVendor=flashsystem" \
  --default-offload-secret dev-flashsystem-secret
```

## Advanced Configuration

### TLS Configuration

```bash
kubectl mtv create mapping storage secure-offload \
  --source vsphere-provider \
  --target openshift-provider \
  --storage-pairs "secure-ds:premium;offloadPlugin=vsphere;offloadVendor=primera3par" \
  --default-offload-plugin vsphere \
  --default-offload-vendor primera3par \
  --offload-vsphere-username admin@vsphere.local \
  --offload-vsphere-password secure-pass \
  --offload-vsphere-url https://vcenter.company.com \
  --offload-storage-username primera-admin \
  --offload-storage-password primera-pass \
  --offload-storage-endpoint https://primera.company.com \
  --offload-cacert @/path/to/primera-ca.crt
```

### Mixed Offload and Traditional Migration

```bash
kubectl mtv create mapping storage mixed-approach \
  --source vsphere-provider \
  --target openshift-provider \
  --storage-pairs "fast-ds:premium;offloadPlugin=vsphere;offloadVendor=pureFlashArray,slow-ds:standard" \
  --default-offload-secret pure-storage-secret
```

## See Also

- [Storage Mapping Creation](README_create_mappings.md) - General storage mapping documentation
- [Plan Creation](README-usage.md#create-plans) - Migration plan creation guide  
- [Patch Mappings](README_patch_mappings.md) - Updating existing storage mappings
- [VDDK Configuration](README_vddk.md) - vSphere Virtual Disk Development Kit setup
