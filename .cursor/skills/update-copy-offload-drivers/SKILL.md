---
name: update-copy-offload-drivers
description: Step-by-step guide for syncing copy-offload (XCOPY) storage vendor drivers and their credential/secret properties between upstream kubev2v/forklift and kubectl-mtv. Use when adding a new offload vendor, updating vendor-specific secret fields, or auditing for upstream changes.
---

# Updating Copy-Offload Storage Vendor Drivers

This skill covers how to keep kubectl-mtv's offload vendor list and secret builder
in sync with the upstream `kubev2v/forklift` copy-offload populator.

---

## Architecture Overview

### How Copy Offload Works

1. A **StorageMap** entry can have an `OffloadPlugin` with either:
   - `VSphereXcopyPluginConfig` — uses the vSphere XCOPY volume populator pod
   - `CsiVolumeImport` — uses CSI driver native import (currently primera3par only)

2. The populator pod reads credentials from a **Kubernetes Secret** containing:
   - vSphere credentials (injected automatically from the provider)
   - Storage array credentials (from the user-provided offload secret)

3. kubectl-mtv creates the offload secret via `pkg/cmd/create/mapping/offload/offload.go`

### Upstream Source of Truth

```text
kubev2v/forklift/pkg/apis/forklift/v1beta1/mapping.go
  - StorageVendorProduct enum (canonical list of vendors)
  - StorageVendorProducts() helper function
  - OffloadPlugin, VSphereXcopyPluginConfig, CsiVolumeImport structs

kubev2v/forklift/cmd/vsphere-copy-offload-populator/
  - vsphere-copy-offload-populator.go  (main switch, env vars consumed)
  - internal/<vendor>/                  (per-vendor cloner implementations)
```

---

## Current Vendor List and Credentials

### StorageVendorProduct Enum

Source: `vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/mapping.go`

| Constant | Value | Plugin Path |
|----------|-------|-------------|
| StorageVendorProductFlashSystem | `flashsystem` | xcopy |
| StorageVendorProductVantara | `vantara` | xcopy |
| StorageVendorProductOntap | `ontap` | xcopy |
| StorageVendorProductPrimera3Par | `primera3par` | xcopy + csiVolumeImport |
| StorageVendorProductPureFlashArray | `pureFlashArray` | xcopy |
| StorageVendorProductPowerFlex | `powerflex` | xcopy |
| StorageVendorProductPowerMax | `powermax` | xcopy |
| StorageVendorProductPowerStore | `powerstore` | xcopy |
| StorageVendorProductInfinibox | `infinibox` | xcopy |

### Secret Data Keys (Environment Variables)

These are the keys the populator pod reads from the mounted secret:

| Key | Required By | Description |
|-----|-------------|-------------|
| `STORAGE_HOSTNAME` | All vendors | Storage array management API endpoint |
| `STORAGE_USERNAME` | All (unless token) | Storage array username |
| `STORAGE_PASSWORD` | All (unless token) | Storage array password |
| `STORAGE_TOKEN` | pureFlashArray (alt) | API token (alternative to user/pass) |
| `STORAGE_SKIP_SSL_VERIFICATION` | Optional for all | `"true"` to skip TLS verification |
| `STORAGE_HTTP_TIMEOUT_SECONDS` | pureFlashArray (opt) | HTTP timeout for slow arrays |
| `PURE_CLUSTER_PREFIX` | pureFlashArray | `px_<first-8-chars-of-cluster-uid>` |
| `POWERFLEX_SYSTEM_ID` | powerflex | System ID from vxflexos-config secret |
| `POWERMAX_SYMMETRIX_ID` | powermax | Symmetrix ID of the storage array |
| `POWERMAX_PORT_GROUP_NAME` | powermax | Port group for masking view creation |
| `ONTAP_SVM` | ontap | SVM name (from TridentBackend.config.ontap_config.svm) |

### Per-Vendor Constructor Signatures

These show what parameters each vendor's cloner needs:

| Vendor | Constructor | Parameters |
|--------|-------------|------------|
| vantara | `NewVantaraClonner(hostname, username, password)` | basic auth |
| ontap | `NewNetappClonner(hostname, username, password)` | basic auth |
| flashsystem | `NewFlashSystemClonner(hostname, username, password, sslSkipVerify)` | basic + TLS |
| primera3par | `NewPrimera3ParClonner(hostname, username, password, sslSkipVerify)` | basic + TLS |
| pureFlashArray | `NewFlashArrayClonner(hostname, username, password, apiToken, sslSkipVerify, clusterPrefix, httpTimeout)` | token + extras |
| powerflex | `NewPowerflexClonner(hostname, username, password, sslSkipVerify, systemId)` | basic + system ID |
| powermax | `NewPowermaxClonner(hostname, username, password, sslSkipVerify)` | basic + TLS |
| powerstore | `NewPowerstoreClonner(hostname, username, password, sslSkipVerify)` | basic + TLS |
| infinibox | `NewInfiniboxClonner(hostname, username, password, insecure)` | basic + TLS |

---

## Step 1: Check Upstream for Changes

After bumping the Forklift dependency, verify the vendored `StorageVendorProducts()`
matches what we expose in validation, help text, and shell completions:

```bash
# Upstream vendor list (source of truth — this function drives our OffloadVendors)
grep -A 15 'func StorageVendorProducts' \
  vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/mapping.go

# Our derived list (pkg/util/flags/output_format_type.go calls StorageVendorProducts())
go run . create mapping storage --help | grep "offload-vendor"
```

Since `pkg/util/flags/output_format_type.go` calls `v1beta1.StorageVendorProducts()`
directly, the vendor list updates automatically on dependency bump. Verify that
the help text, completion, and validation all reflect any new entries.

Check if the upstream populator has new env vars or vendors:

```bash
# Env vars consumed by the populator
grep 'os.Getenv\|flag.StringVar' \
  vendor/github.com/kubev2v/forklift/cmd/vsphere-copy-offload-populator/vsphere-copy-offload-populator.go 2>/dev/null \
  || echo "Check ../../kubev2v/forklift/cmd/vsphere-copy-offload-populator/vsphere-copy-offload-populator.go"

# New internal vendor packages
ls vendor/github.com/kubev2v/forklift/cmd/vsphere-copy-offload-populator/internal/ 2>/dev/null \
  || ls ../../kubev2v/forklift/cmd/vsphere-copy-offload-populator/internal/
```

---

## Step 2: Update the Vendor List (if new vendor found)

All locations where the vendor list is hardcoded:

| # | File | Location | Purpose |
|---|------|----------|---------|
| 1 | `pkg/cmd/create/mapping/storage.go` | `validateOffloadVendor()` switch | **Validation (source of truth in kubectl-mtv)** |
| 2 | `pkg/cmd/create/mapping/storage.go` | Error message string | User-facing error |
| 3 | `cmd/create/plan.go` | `--default-offload-vendor` flag help | CLI help |
| 4 | `cmd/create/plan.go` | `RegisterFlagCompletionFunc` | Tab completion |
| 5 | `cmd/create/mapping.go` | `--default-offload-vendor` flag help | CLI help |
| 6 | `cmd/create/mapping.go` | `RegisterFlagCompletionFunc` | Tab completion |
| 7 | `cmd/patch/mapping.go` | `--default-offload-vendor` flag help | CLI help |
| 8 | `cmd/patch/mapping.go` | `RegisterFlagCompletionFunc` | Tab completion |
| 9 | `pkg/cmd/create/mapping/mapping_test.go` | `TestValidateOffloadVendor_Valid` | Test |

**To add a new vendor:**

1. Add the vendor string to the `switch` in `validateOffloadVendor()` (file #1)
2. Update the error message (file #2)
3. Add to all flag help strings (files #3, #5, #7)
4. Add to all completion funcs (files #4, #6, #8)
5. Add to the test list (file #9)
6. Regenerate MCP help testdata: `go run . help --machine > pkg/mcp/discovery/testdata/help_machine_output.json`

---

## Step 3: Update Secret Builder (if new secret keys needed)

The offload secret builder is in: `pkg/cmd/create/mapping/offload/offload.go`

### Current Secret Keys Created

```go
secretData["STORAGE_USERNAME"]             // from --offload-storage-username
secretData["STORAGE_PASSWORD"]             // from --offload-storage-password
secretData["STORAGE_HOSTNAME"]             // from --offload-storage-endpoint
secretData["STORAGE_SKIP_SSL_VERIFICATION"] // from insecure flag
secretData["ca.crt"]                       // from --offload-cacert
```

### Known Missing Keys (vendor-specific)

| Key | Flag to Add | Needed By |
|-----|-------------|-----------|
| `STORAGE_TOKEN` | `--offload-storage-token` | pureFlashArray (alternative to user/pass) |
| `PURE_CLUSTER_PREFIX` | `--offload-pure-cluster-prefix` | pureFlashArray |
| `POWERFLEX_SYSTEM_ID` | `--offload-powerflex-system-id` | powerflex |
| `STORAGE_HTTP_TIMEOUT_SECONDS` | `--offload-storage-timeout` | pureFlashArray (optional) |

### Files to Update for New Secret Keys

| File | What to update |
|------|---------------|
| `pkg/cmd/create/mapping/offload/offload.go` | Add field to `SecretOptions`, add to `BuildSecret()`, update `NeedsSecret()` and `ValidateSecretFields()` |
| `pkg/cmd/create/mapping/offload/offload_test.go` | Add test for new key |
| `cmd/create/plan.go` | Add CLI flag, wire to `SecretOptions` |
| `cmd/create/mapping.go` | Add CLI flag, wire to `SecretOptions` |
| `cmd/patch/mapping.go` | Add CLI flag if applicable |

### Validation Rules per Vendor

When adding vendor-specific fields, validation should be context-aware:

- **pureFlashArray**: Requires either `STORAGE_TOKEN` OR (`STORAGE_USERNAME` + `STORAGE_PASSWORD`). Also requires `PURE_CLUSTER_PREFIX`.
- **powerflex**: Requires `POWERFLEX_SYSTEM_ID` in addition to basic auth.
- **All others**: Require `STORAGE_USERNAME` + `STORAGE_PASSWORD` + `STORAGE_HOSTNAME`.

---

## Step 4: Update CsiVolumeImport (if new CSI vendor)

The CSI import path (alternative to xcopy populator pod) is in:

```text
vendor/github.com/kubev2v/forklift/pkg/controller/plan/adapter/vsphere/csi_import.go
```

Currently only `primera3par` is supported for CSI import. If a new vendor is added:

1. Check if upstream added a new case in `newCsiImportPlugin()`
2. The kubectl-mtv StorageMap just sets the `CsiVolumeImport` struct — no populator-pod secret is needed for this path (the CSI driver handles it natively)

---

## Step 5: Verify and Test

```bash
# Build
go build ./...

# Run unit tests
go test ./pkg/cmd/create/mapping/... -v
go test ./cmd/... -v

# Regenerate MCP testdata
go run . help --machine > pkg/mcp/discovery/testdata/help_machine_output.json

# Run MCP tests
go test ./pkg/mcp/... -v

# Verify completion lists work
go run . create plan --default-offload-vendor <TAB>
go run . create mapping storage --default-offload-vendor <TAB>
```

---

## Quick Audit Command

Run this to quickly check if kubectl-mtv is out of sync with upstream:

```bash
echo "=== Upstream vendors ==="
grep 'StorageVendorProduct\w*\s*=' \
  vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/mapping.go 2>/dev/null \
  || grep 'StorageVendorProduct\w*\s*=' \
    ../../kubev2v/forklift/pkg/apis/forklift/v1beta1/mapping.go

echo ""
echo "=== kubectl-mtv vendors ==="
grep -A1 'func validateOffloadVendor' pkg/cmd/create/mapping/storage.go

echo ""
echo "=== Upstream populator env vars ==="
grep 'os.Getenv' \
  vendor/github.com/kubev2v/forklift/cmd/vsphere-copy-offload-populator/vsphere-copy-offload-populator.go 2>/dev/null \
  || grep 'os.Getenv' \
    ../../kubev2v/forklift/cmd/vsphere-copy-offload-populator/vsphere-copy-offload-populator.go

echo ""
echo "=== kubectl-mtv secret keys ==="
grep 'secretData\[' pkg/cmd/create/mapping/offload/offload.go
```
