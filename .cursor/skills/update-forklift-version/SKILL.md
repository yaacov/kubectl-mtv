---
name: update-forklift-version
description: Step-by-step guide for updating the kubev2v/forklift Go dependency in kubectl-mtv. Use when bumping the forklift version, syncing settings or CRD types, or checking for upstream changes.
---

# Updating the Forklift Dependency

## 1. Fetch the Latest Version

```bash
GOFLAGS=-mod=mod GOPROXY=https://proxy.golang.org,direct \
  go get github.com/kubev2v/forklift@latest
go mod tidy
go mod vendor
go build ./...
```

Verify the build succeeds before proceeding.

## 2. Check ForkliftController Settings

Settings are defined in `pkg/cmd/settings/types.go` in two maps:
- `SupportedSettings` -- commonly configured settings
- `ExtendedSettings` -- advanced/less-common settings

### Upstream Source of Truth

Fetch the Ansible defaults file (canonical defaults for all settings):

```text
https://raw.githubusercontent.com/kubev2v/forklift/main/operator/roles/forkliftcontroller/defaults/main.yml
```

### What to Compare

For every setting in our maps, compare against the upstream file:

| Check | Action |
|-------|--------|
| **Default value mismatch** | Update `Default` field in our definition |
| **Setting removed upstream** | Search vendored code (`vendor/github.com/kubev2v/forklift/`) to confirm removal, then delete from our map |
| **New setting upstream** | Add to `SupportedSettings` (user-facing) or `ExtendedSettings` (advanced) with correct type, default, description, and category |
| **New category needed** | Add to `SettingCategory` constants and `CategoryOrder` slice in `types.go` |

### Verification

Run `go test ./pkg/cmd/settings/...` -- the test suite checks:
- Name/key consistency
- No overlap between Supported and Extended
- All categories in CategoryOrder are covered
- Sorting within categories

## 3. Check Plan CRD Changes

The Plan CRD types live in the vendor at:

```text
vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/plan.go     # PlanSpec, PlanStatus
vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/plan/vm.go  # plan.VM struct
```

### Files to Update for New PlanSpec Fields

| File | What to update |
|------|---------------|
| `cmd/create/plan.go` | Add CLI flag, wire into `planSpec` |
| `cmd/patch/plan.go` | Add CLI flag, wire into `PatchPlanOptions` |
| `pkg/cmd/patch/plan/plan.go` | Handle new field in patch builder |
| `pkg/cmd/describe/plan/describe.go` | Display new field in `buildSpecSection` |
| MCP help generator | If the flag should be visible to AI assistants |

### Files to Update for New VM Struct Fields

| File | What to update |
|------|---------------|
| `cmd/patch/plan.go` (`NewPlanVMCmd`) | Add CLI flag for VM-level field |
| `pkg/cmd/patch/plan/plan.go` (`PatchPlanVM`) | Handle new VM field in patch |
| `pkg/cmd/describe/plan/describe.go` (`buildVMsSection`) | Display new VM field |

### Boolean Fields with kubebuilder Defaults

When a PlanSpec bool has `+kubebuilder:default:=true`, the API server auto-sets it.
If the CLI needs to explicitly set it to `false`, a post-create patch is required
(see existing pattern in `pkg/cmd/create/plan/plan.go` for `MigrateSharedDisks`,
`PreserveStaticIPs`, `UseCompatibilityMode`, etc.).

### Pointer Bool Fields (*bool)

Fields like `InstallLegacyDrivers` and `EnableNestedVirtualization` use `*bool`
so they can be nil (auto-detect), true, or false. The CLI flag should be a
`StringVar` accepting `"true"`, `"false"`, or `"auto"` (= nil/auto-detect).
Use `"auto"` as the default in `create` commands. In `patch` commands use `""`
as the default (meaning "don't change") and add a `Changed` flag so that
`--flag auto` can explicitly clear the field back to nil.

## 4. Check Other CRD Changes

Also review for changes in:

```text
vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/provider.go  # ProviderSpec, ProviderType
vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/mapping.go   # NetworkPair, StoragePair, DestinationNetwork
vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/host.go
vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/hook.go
```

Key things to watch:
- New `ProviderType` constants (new source provider types)
- New `StorageVendorProduct` enum values (update completion lists)
- New `DestinationNetwork.Type` enum values (update network pair parsing)
- New `MigrationType` constants
- Changes to `ProviderSpec.Settings` keys

## 5. Check Inventory API Changes

The inventory service is consumed via REST (`pkg/util/client/inventory.go`), not via
vendored Go types.  JSON responses are decoded into `map[string]interface{}`, so
upstream changes will **not** cause compile errors but can silently break output.

### Upstream Source of Truth

Inventory REST endpoints are defined in the upstream Forklift repo under:

```text
https://github.com/kubev2v/forklift/tree/main/pkg/controller/provider/web
```

Each provider type has a sub-package (e.g. `vsphere/`, `ovirt/`, `openstack/`,
`ec2/`) that registers collection handlers.

### What to Compare

| Check | Action |
|-------|--------|
| **New collection endpoint** | Add `Get<Resource>` / `Get<Resource>ByID` methods to `pkg/cmd/get/inventory/client.go`, create a new list file in `pkg/cmd/get/inventory/`, wire the subcommand in `cmd/get/inventory_<provider>.go` |
| **New provider type needing inventory** | Add a new `cmd/get/inventory_<provider>.go`, extend provider-type switch lists in `pkg/cmd/get/inventory/vms.go` (`FetchVMsByQueryWithInsecure` and `listVMsOnce`) |
| **Renamed/removed collection** | Update the string path in the corresponding `Get<Resource>` method in `client.go` |
| **Changed JSON field names** | Update `augmentVMInfo`, `vmColumns`, and any code that extracts fields from `map[string]interface{}` in `pkg/cmd/get/inventory/` |
| **New JSON fields worth displaying** | Add columns to `vmColumns()` in `vms.go` or equivalent column definitions in other list files |
| **Changed detail levels or query params** | Update `GetResourceCollection` detail param in `client.go` or fetch helpers in `pkg/util/client/inventory.go` |
| **New global (non-provider) endpoint** | Add a fetch function in `pkg/util/client/inventory.go` (see `FetchAAPJobTemplatesWithInsecure` as a pattern) |

### How to Discover Changes

Compare the upstream handler tree against our `ProviderClient` methods:

```bash
# List all collection path strings in our client
grep -oP 'GetResourceCollection\(ctx, "\K[^"]+' pkg/cmd/get/inventory/client.go | sort

# Then compare with the upstream web handler directories/files for each provider
```

### Files at a Glance

| File | Role |
|------|------|
| `pkg/util/client/inventory.go` | Low-level HTTP client, global endpoints |
| `pkg/cmd/get/inventory/client.go` | `ProviderClient` with typed `Get<Resource>` methods |
| `pkg/cmd/get/inventory/vms.go` | VM list logic, `augmentVMInfo`, `vmColumns`, provider-type switches |
| `pkg/cmd/get/inventory/*.go` | Per-resource list functions (hosts, networks, storage, etc.) |
| `cmd/get/inventory_*.go` | Cobra subcommand wiring per provider |
| `pkg/cmd/get/inventory/ec2_helpers.go` | EC2 envelope normalization |

## 6. Run Full Test Suite

```bash
go test ./... -count=1
go build ./...
```

## Quick Diff Checklist

After vendoring the new version, a fast way to spot changes:

```bash
# See what files changed in the vendored forklift package
git diff --name-only -- vendor/github.com/kubev2v/forklift/

# Detailed diff of CRD types
git diff -- vendor/github.com/kubev2v/forklift/pkg/apis/
```
