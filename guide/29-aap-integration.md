---
layout: default
title: "Chapter 29: AAP (Ansible Automation Platform) Integration"
parent: "V. Advanced Migration Customization and Optimization"
nav_order: 5
---

Ansible Automation Platform (AAP) integration allows migration hooks to trigger AAP job templates instead of running local container-based playbooks. This enables enterprise-grade automation workflows, centralized credential management, and reuse of existing AAP infrastructure during VM migrations.

## Overview

### What is AAP Integration?

Forklift/MTV supports two types of migration hooks:

- **Local hooks**: Run a container image with an embedded Ansible playbook (covered in [Chapter 17](17-migration-hooks.md))
- **AAP hooks**: Trigger a job template on a remote Ansible Automation Platform server

AAP hooks delegate execution to your existing AAP infrastructure, providing:

- **Centralized management**: Job templates are managed in AAP, not embedded in hook resources
- **Credential isolation**: AAP manages credentials for target systems; Forklift only needs an API token
- **Audit trail**: AAP provides built-in logging, notifications, and approval workflows
- **Reusability**: The same job template can serve multiple migration plans

### Architecture

```text
┌─────────────────┐         ┌─────────────────────┐
│  Forklift       │  REST   │  Ansible Automation  │
│  Controller     │────────▶│  Platform (AAP)      │
│                 │         │                      │
│  (triggers hook │         │  Job Template #42    │
│   at migration  │         │  → runs playbook     │
│   checkpoint)   │         │  → returns status    │
└─────────────────┘         └─────────────────────┘
```

When a migration reaches a hook checkpoint (pre-migration or post-migration), the controller calls the AAP API to launch the configured job template and polls until it completes or times out.

## Prerequisites

### 1. AAP Server Access

You need a running AAP/AWX instance with:

- At least one job template configured for your migration automation
- An API token (personal access token or application token) with permission to launch the template

### 2. Configure ForkliftController Settings

AAP connection details are configured at the controller level via settings:

```bash
# Set the AAP server URL
kubectl mtv settings set --setting aap_url --value "https://aap.example.com"

# Create a secret containing the AAP API token
kubectl create secret generic aap-token \
  --namespace openshift-mtv \
  --from-literal=token="your-aap-api-token-here"

# Point the controller to the token secret
kubectl mtv settings set --setting aap_token_secret_name --value "aap-token"
```

Optional settings:

```bash
# Set a global timeout for AAP job polling (seconds)
kubectl mtv settings set --setting aap_timeout --value "600"

# Skip TLS verification (development only)
kubectl mtv settings set --setting aap_insecure_skip_verify --value "true"

# Use a custom CA certificate for AAP server
kubectl create secret generic aap-ca \
  --namespace openshift-mtv \
  --from-file=ca.crt=/path/to/ca-certificate.pem

kubectl mtv settings set --setting aap_ca_secret_name --value "aap-ca"
```

### 3. Verify AAP Connectivity

Once configured, verify that Forklift can reach AAP by querying job templates:

```bash
kubectl mtv get inventory job-template
```

If AAP is properly configured, this displays available job templates. If not, you will see an error indicating that AAP is not configured on the ForkliftController.

## Browsing AAP Job Templates

The `get inventory job-template` command queries the AAP server through the Forklift inventory service and lists available job templates.

```bash
# List all available job templates
kubectl mtv get inventory job-template

# Filter by name pattern
kubectl mtv get inventory job-template --query "where name ~= 'migration.*'"

# Output as JSON for scripting
kubectl mtv get inventory job-template --output json
```

Example output:

```text
ID    NAME
42    pre-migration-backup
55    post-migration-validation
71    network-reconfiguration
```

Use the ID value when creating AAP hooks.

## Creating AAP Hooks

### Basic AAP Hook

Create a hook that triggers AAP job template #42:

```bash
kubectl mtv create hook --name pre-backup --aap-job-template-id 42
```

The `--aap-job-template-id` flag is mutually exclusive with `--image` and `--playbook`. You cannot combine AAP and local hook configurations.

### AAP Hook with Per-Hook Overrides

If a hook needs to target a different AAP server or use different credentials than the controller defaults:

```bash
# Use a different AAP server for this specific hook
kubectl mtv create hook --name special-hook \
  --aap-job-template-id 55 \
  --aap-url "https://aap-staging.example.com"

# Use a different token secret
kubectl mtv create hook --name secure-hook \
  --aap-job-template-id 71 \
  --aap-token-secret "team-specific-aap-token"

# Set a custom timeout for long-running jobs
kubectl mtv create hook --name long-running-hook \
  --aap-job-template-id 42 \
  --aap-timeout 1800
```

### AAP Hook Creation Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--aap-job-template-id` | AAP job template ID (required for AAP hooks) | -- |
| `--aap-url` | Per-hook AAP base URL | Controller `aap_url` setting |
| `--aap-token-secret` | Per-hook AAP token Secret name | Controller `aap_token_secret_name` |
| `--aap-timeout` | Per-hook AAP job poll timeout (seconds) | Controller `aap_timeout` |
| `--service-account` | Kubernetes ServiceAccount for the hook | Default |
| `--deadline` | Hook execution deadline in seconds | No timeout |

### Dry-Run Mode

Preview the hook resource before creating it:

```bash
kubectl mtv create hook --name pre-backup \
  --aap-job-template-id 42 \
  --dry-run --output yaml
```

## Patching Hooks

### Update AAP Configuration

```bash
# Change the job template ID
kubectl mtv patch hook --name pre-backup --aap-job-template-id 55

# Update the per-hook AAP URL
kubectl mtv patch hook --name pre-backup --aap-url "https://aap-v2.example.com"

# Update the token secret
kubectl mtv patch hook --name pre-backup --aap-token-secret "new-token-secret"

# Change the timeout
kubectl mtv patch hook --name pre-backup --aap-timeout 900
```

### Switching Between Hook Types

#### Convert a local hook to an AAP hook:

```bash
kubectl mtv patch hook --name my-hook --aap-job-template-id 42 --image ""
```

#### Convert an AAP hook back to a local hook:

```bash
kubectl mtv patch hook --name my-hook \
  --clear-aap \
  --image quay.io/kubev2v/hook-runner \
  --playbook @my-playbook.yml
```

The `--clear-aap` flag removes the AAP configuration from the hook. When clearing AAP, you must provide at least `--image` or `--playbook` for the resulting local hook.

## Using AAP Hooks in Migration Plans

AAP hooks are attached to migration plans the same way as local hooks. They can serve as pre-migration or post-migration hooks.

### During Plan Creation

```bash
kubectl mtv create plan --name my-migration \
  --source vsphere-prod \
  --vms "web-server-01,web-server-02" \
  --pre-hook pre-backup \
  --post-hook post-validation
```

### Via Plan VM Patching

```bash
# Add a pre-migration AAP hook to a specific VM
kubectl mtv patch planvm --plan my-migration --vm web-server-01 \
  --pre-hook pre-backup

# Add a post-migration AAP hook
kubectl mtv patch planvm --plan my-migration --vm web-server-01 \
  --post-hook post-validation
```

## AAP Settings Reference

All AAP-related ForkliftController settings:

| Setting | Type | Description |
|---------|------|-------------|
| `aap_url` | string | Base URL for the AAP server (e.g., `https://aap.example.com`) |
| `aap_token_secret_name` | string | Name of the Secret containing the AAP API token |
| `aap_timeout` | int | Default timeout in seconds for AAP job polling |
| `aap_insecure_skip_verify` | bool | Skip TLS certificate verification for AAP connections |
| `aap_ca_secret_name` | string | Name of the Secret containing a custom CA certificate for AAP |

### Viewing AAP Settings

```bash
# View all settings (AAP settings appear under the 'aap' category)
kubectl mtv settings

# Get a specific AAP setting
kubectl mtv settings get --setting aap_url
kubectl mtv settings get --setting aap_token_secret_name
```

### Resetting AAP Settings

```bash
# Remove the AAP URL (reverts to default/unset)
kubectl mtv settings unset --setting aap_url

# Remove the token secret reference
kubectl mtv settings unset --setting aap_token_secret_name
```

## Best Practices

1. **Use controller-level defaults**: Configure `aap_url` and `aap_token_secret_name` at the controller level. Use per-hook overrides only for exceptions (e.g., a staging AAP instance).

2. **Set appropriate timeouts**: AAP jobs that involve network operations or large data transfers should have generous timeouts. Set `aap_timeout` at the controller level and override per-hook when needed.

3. **Use TLS in production**: Always configure a proper CA certificate via `aap_ca_secret_name` rather than using `aap_insecure_skip_verify`.

4. **Scope AAP tokens narrowly**: Create dedicated AAP tokens with minimal permissions -- only the ability to launch the specific job templates needed for migration hooks.

5. **Test hooks before bulk migrations**: Verify AAP hook connectivity with a single-VM test migration before running large-scale migrations.

6. **Monitor AAP job status**: Use `kubectl mtv describe plan --name <plan> --vm <vm>` to check hook execution status during migrations.

## Troubleshooting

### "AAP not configured on ForkliftController"

This error from `get inventory job-template` means the controller lacks AAP settings:

```bash
kubectl mtv settings get --setting aap_url
kubectl mtv settings get --setting aap_token_secret_name
```

Both must be set for AAP integration to function.

### Hook Timeout Errors

If AAP jobs exceed the configured timeout:

```bash
# Check current timeout
kubectl mtv settings get --setting aap_timeout

# Increase globally
kubectl mtv settings set --setting aap_timeout --value "1200"

# Or increase for a specific hook
kubectl mtv patch hook --name slow-hook --aap-timeout 1800
```

### TLS Certificate Errors

If the AAP server uses a private CA:

```bash
# Create the CA secret
kubectl create secret generic aap-ca \
  --namespace openshift-mtv \
  --from-file=ca.crt=/path/to/ca.pem

# Configure the controller to use it
kubectl mtv settings set --setting aap_ca_secret_name --value "aap-ca"
```

### Verifying Hook Type

To confirm whether a hook is configured as local or AAP:

```bash
kubectl mtv describe hook --name my-hook
```

The output will show either the container image (local) or the AAP job template ID.
