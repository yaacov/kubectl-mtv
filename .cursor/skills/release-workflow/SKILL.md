---
name: release-workflow
description: Reference for building, releasing, publishing images, and deploying kubectl-mtv. Use when creating releases, building binaries or container images, deploying to OpenShift, or updating the Krew plugin index.
---

# Release Workflow

## Build

| Command | Purpose |
|---------|---------|
| `make kubectl-mtv` | Build for current platform |
| `make build-all` | Build all platforms (linux/darwin amd64+arm64, windows/amd64) |
| `make dist-all` | Build all + create `.tar.gz`/`.zip` archives + `.sha256sum` checksums |
| `make dist` | Build + tarball for current platform only |

Version is set via `VERSION` env var or defaults to `git describe`:

```bash
make dist-all VERSION=v0.5.0
```

## Container Images

| Command | Purpose |
|---------|---------|
| `make image-build-all` | Build container images (amd64 + arm64) |
| `make image-push-all` | Push images + create multi-arch manifest |

Default image: `quay.io/yaacov/kubectl-mtv-mcp-server`.

The `Containerfile` is a multi-stage build that compiles the Go binary and produces a minimal image.

## GitHub Release

Creating a GitHub release triggers `.github/workflows/release.yml`:

1. Checks if assets already exist (idempotent).
2. Runs `make dist-all` with the release tag as `VERSION`.
3. Uploads archives and checksums via `softprops/action-gh-release`.
4. Verifies all assets are downloadable.
5. Runs `krew-release-bot` to update the Krew plugin index.

## Krew Plugin

`.krew.yaml` is a template processed by `krew-release-bot`. It defines:

- Plugin name: `mtv`
- Platforms: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- Download URLs point to GitHub release assets

The bot automatically submits a PR to the `krew-index` repo when a release is created.

## OpenShift Deployment

Deploy manifests live in `deploy/`:

| File | Purpose |
|------|---------|
| `mcp-server.yaml` | ServiceAccount + Deployment + Service for MCP server |
| `mcp-route.yaml` | OpenShift Route with TLS edge termination |
| `olsconfig-patch.yaml` | Registers MCP server with OpenShift Lightspeed |

### Deploy Targets

| Command | Purpose |
|---------|---------|
| `make deploy` | Deploy MCP server (SA + Deployment + Service) |
| `make undeploy` | Remove MCP server |
| `make deploy-route` | Create the Route |
| `make undeploy-route` | Remove the Route |
| `make deploy-olsconfig` | Register with Lightspeed |
| `make undeploy-olsconfig` | Unregister from Lightspeed |
| `make deploy-all` | Deploy everything |
| `make undeploy-all` | Remove everything |

The MCP server deploys to the `openshift-mtv` namespace. Key environment variables in the Deployment:

- `MCP_HOST`, `MCP_PORT` -- listen address
- `MCP_KUBE_SERVER` -- K8s API server URL
- `MCP_READ_ONLY` -- disable write operations
- `MCP_OUTPUT_FORMAT` -- default output format (json/text)
- `MCP_MAX_RESPONSE_CHARS` -- truncation limit

## GitHub Pages (Guide)

Push to `main` with changes in `guide/`, `_config.yml`, `_includes/`, or `Gemfile` triggers `.github/workflows/pages.yml`:

1. Build Jekyll site.
2. Deploy to GitHub Pages.

## Release Checklist

1. Ensure `make lint` and `make test` pass.
2. Create a GitHub release with a tag (e.g. `v0.5.0`).
3. The release workflow builds, uploads, and publishes automatically.
4. Verify the Krew plugin update PR is submitted.
5. Optionally update the container image: `make image-build-all && make image-push-all`.
6. Update OpenShift deployment if needed: `make deploy`.
