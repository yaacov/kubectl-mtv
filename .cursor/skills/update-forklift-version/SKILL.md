---
name: update-forklift-version
description: Fetch the latest kubev2v/forklift Go dependency and produce an update plan from observed diffs. Use when bumping the forklift version, checking upstream CRD or settings changes, or planning kubectl-mtv adaptations. Does not implement CLI or settings code.
---

# Updating the Forklift Dependency

Two steps: fetch latest, then plan from observed changes.
Do not implement kubectl-mtv CLI, flags, inventory, or `SupportedSettingNames`. Do not commit -- the user will commit when ready.

1. **Fetch latest** -- `go get @latest`, tidy, vendor, build.
2. **Inspect and plan** -- diff the previous module against the new tree, regenerate settings for inspection, present an update plan, and **stop**.

---

## Step 1: Fetch Latest

Record the current version from `go.mod` **before** fetching (needed for the Step 2 diff):

```bash
grep 'kubev2v/forklift' go.mod
```

Then fetch latest:

```bash
GOFLAGS=-mod=mod GOPROXY=https://proxy.golang.org,direct \
  go get github.com/kubev2v/forklift@latest
go mod tidy
go mod vendor
go build ./...
```

If the build fails, report the errors and stop. Do not wait for a version approval; do not pin a specific version.

---

## Step 2: Inspect and Plan

`vendor/` is gitignored, so `git diff -- vendor/...` is always empty. Compare the **previous module cache** against the newly vendored tree. Use the old pseudo-version recorded in Step 1 (or `git show HEAD:go.mod`):

```bash
OLD="$(go env GOMODCACHE)/github.com/kubev2v/forklift@<old-pseudo-version>"
NEW="vendor/github.com/kubev2v/forklift"

# Settings source
diff -u "$OLD/pkg/apis/forklift/v1beta1/forkliftcontroller.go" \
        "$NEW/pkg/apis/forklift/v1beta1/forkliftcontroller.go"

# CRD types
diff -rq "$OLD/pkg/apis" "$NEW/pkg/apis"
diff -ru "$OLD/pkg/apis/forklift/v1beta1" "$NEW/pkg/apis/forklift/v1beta1"

# Inventory is often not vendored; compare the module cache
# Legacy handlers: pkg/controller/provider/web
# Newer providers: pkg/provider/<name>/inventory
NEW_MOD="$(GOFLAGS=-mod=mod go list -m -f '{{.Dir}}' github.com/kubev2v/forklift)"
for rel in pkg/controller/provider/web pkg/provider; do
  OLD_DIR="$OLD/$rel"
  NEW_DIR="$NEW_MOD/$rel"
  if [ ! -d "$NEW_DIR" ]; then
    echo "Inventory directory absent: $NEW_DIR" >&2
  elif [ ! -d "$OLD_DIR" ]; then
    echo "Inventory directory absent: $OLD_DIR" >&2
  else
    diff -rq "$OLD_DIR" "$NEW_DIR"
  fi
done
```

Regenerate settings and check curated names (this may rewrite `pkg/cmd/settings/types_generated.go`; include that diff in the plan):

```bash
make verify-defaults FORKLIFT_PATH=vendor/github.com/kubev2v/forklift
```

That runs `generate-settings` then `TestSupportedSettingNames_AllExistInAllSettings`. Use `FORKLIFT_PATH=vendor/...` so generated settings match the Go dependency. The Makefile default points at a local forklift checkout which may be a different version.

Do not use `scripts/verify-defaults.sh` here -- that compares operator YAML defaults and needs a full forklift git checkout, not the vendored module.

Then present the plan and **stop**. Do not edit CLI, flags, inventory, or `SupportedSettingNames`.

Show:

- Version bump (old -> new from `go.mod`)
- Settings: generated file changed or not; test failures; proposed `SupportedSettingNames` / category edits
- CRDs: each new/changed field, enum, or type, with the kubectl-mtv files that would change
- Inventory: new endpoints or provider types, if known
- What you recommend vs what is optional / needs a decision

A new `ProviderType` is a full feature, not a silent bump. Note it in the plan as a scope decision (recognition only, create/patch, or full inventory and mappers). Do not start implementing it.

The plan should follow what the diffs actually show. Unexpected upstream changes happen; the lists below are usual places to look, not an exhaustive checklist.

---

## Planning Reference

The only codegen from upstream types is `cmd/gen-settings`: it parses `ForkliftControllerSpec` and writes `pkg/cmd/settings/types_generated.go`. There is no generator for Plan/Provider/mapping/host/hook CRDs or inventory handlers -- those CLI and inventory changes are planned by reading the diffs and naming the kubectl-mtv files that would change.

Correctness checks already in this workflow:

- `go build ./...` -- kubectl-mtv still compiles against the new Forklift types (renamed/removed fields, broken imports). It does not catch new optional fields, new enum values, or inventory JSON changes.
- `make verify-defaults` -- generated settings match `ForkliftControllerSpec`, and every `SupportedSettingNames` entry still exists.

There is no tool that checks whether create/patch/describe flags or inventory columns cover the CRD. Missing flags do not fail the build. If `go build` fails, report the errors in the plan and stop.

### Settings

Settings come from the upstream `ForkliftControllerSpec` struct:

- `pkg/cmd/settings/types_generated.go` -- generated `AllSettings` map (do not hand-edit)
- `pkg/cmd/settings/types.go` -- `SupportedSettingNames` curated list + type definitions

If `TestSupportedSettingNames_AllExistInAllSettings` fails, a setting was removed upstream -- note removing it from `SupportedSettingNames`. New settings appear in `--all` output; note any worth promoting. A new category would need `SettingCategory` / `CategoryOrder` in `types.go` and `sectionToCategory` in `cmd/gen-settings/main.go`.

### Plan CRD

```text
vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/plan.go     # PlanSpec, PlanStatus
vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/plan/vm.go  # plan.VM struct
```

New or changed PlanSpec / VM fields typically affect create, patch, and describe (`cmd/create/plan.go`, `cmd/patch/plan.go`, `pkg/cmd/patch/plan/plan.go`, `pkg/cmd/describe/plan/describe.go`). Mention MCP help if flags would change.

### Other CRDs

```text
vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/provider.go  # ProviderSpec, ProviderType
vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/mapping.go   # NetworkPair, StoragePair, DestinationNetwork
vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/host.go
vendor/github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/hook.go
```

Usual signals: new `ProviderType` constants, mapping pair fields, `StorageVendorProduct` values, `DestinationNetwork.Type` values, `MigrationType` constants, `ProviderSpec.Settings` keys.

### Inventory

Inventory is consumed via REST (`pkg/util/client/inventory.go`), not vendored Go types. JSON is decoded into `map[string]interface{}`, so upstream changes will not fail the build but can break output.

Handlers live under `pkg/controller/provider/web` (legacy per-provider sub-packages) and `pkg/provider/<name>/inventory` (newer providers such as EC2). Compare those trees to collection paths in `pkg/cmd/get/inventory/client.go`, list/column logic in `pkg/cmd/get/inventory/`, and Cobra wiring in `cmd/get/inventory_*.go`.
