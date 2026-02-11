---
layout: page
title: "Chapter 24: System Health Checks"
---

The `health` command provides comprehensive diagnostics for the MTV/Forklift migration system. It checks every layer of the stack -- from the operator and controller to pods, providers, and migration plans -- and produces a report with actionable recommendations. All command examples are verified against the implementation.

## Overview

Running `kubectl mtv health` performs the following checks in order:

1. **Operator** -- Verifies that the MTV operator is installed, detects its version and namespace.
2. **ForkliftController** -- Checks for the ForkliftController CR, inspects feature flags, VDDK image, custom images, log level, and status conditions.
3. **Pods** -- Lists all Forklift operator pods and checks their status, readiness, restart counts, and termination reasons (e.g., OOMKilled).
4. **Log Analysis** -- Optionally scans recent pod logs for errors and warnings.
5. **Providers** -- Inspects provider connectivity, readiness, validation, and inventory status.
6. **Plans** -- Checks migration plan readiness, status, VM counts, and failure counts.

The report concludes with a list of detected issues (with severity levels), a summary of counts, and concrete recommendations.

## Running a Health Check

### Basic Usage

```bash
kubectl mtv health
```

This runs all checks and displays a colored table report.

### Output Formats

```bash
# Default table output (colored, human-readable)
kubectl mtv health

# JSON output (for automation and scripting)
kubectl mtv health -o json

# YAML output
kubectl mtv health -o yaml
```

### Flags

| **Flag** | **Short** | **Default** | **Description** |
|----------|-----------|-------------|-----------------|
| `--output` | `-o` | `table` | Output format: `table`, `json`, `yaml` |
| `--skip-logs` | | `false` | Skip pod log analysis (faster execution) |
| `--log-lines` | | `100` | Number of log lines per pod to analyze |
| `--namespace` | `-n` | | Scope providers and plans to a namespace |
| `--all-namespaces` | `-A` | | Scan providers and plans across all namespaces |

## Understanding the Report

The table output is organized into clearly labeled sections:

### Operator Status

Shows whether the MTV operator is installed, its version, and the detected namespace (typically `openshift-mtv`).

### ForkliftController

Displays the controller name, namespace, status conditions, feature flags, VDDK image (if configured), custom image overrides, and the log verbosity level.

### Forklift Pods

Lists each Forklift pod with its status, readiness, restart count, and any detected issues such as crash loops or OOM terminations.

### Pod Log Analysis

When log analysis is enabled (the default), this section shows the number of errors and warnings found in recent log lines for each pod, along with sample error lines.

### Providers

Lists each provider with its type, phase, readiness, connectivity, validation, and inventory status. Issues are flagged if a provider is not ready or not connected.

### Plans

Lists each migration plan with its readiness, status, VM count, and failure/success counts.

### Summary

Provides aggregate counts: total and healthy pods, providers, and plans, plus the number of critical and warning issues.

## Overall Status Values

The report assigns one of four overall status values:

| **Status** | **Meaning** |
|------------|-------------|
| **Healthy** | No critical or warning issues detected |
| **Warning** | One or more non-critical issues need attention |
| **Critical** | One or more critical issues require immediate action |
| **Unknown** | Health could not be fully determined |

## Namespace Scoping

The `health` command handles namespaces differently for operator components and user resources:

- **Operator components** (controller, pods, log analysis) always use the auto-detected operator namespace regardless of the `-n` flag.
- **User resources** (providers, plans) respect the `-n` and `-A` flags:

```bash
# Check health scoped to a specific namespace
kubectl mtv health -n my-migration-project

# Check health across all namespaces
kubectl mtv health -A
```

When neither `-n` nor `-A` is specified, providers and plans are scoped to the operator namespace.

## Example Workflows

### Quick Cluster Health Check

```bash
kubectl mtv health
```

### Fast Check Without Log Analysis

```bash
kubectl mtv health --skip-logs
```

### Detailed Check With More Log Lines

```bash
kubectl mtv health --log-lines 500
```

### Cluster-Wide Health Report

```bash
kubectl mtv health -A
```

### JSON Output for CI/CD Pipelines

```bash
# Get health report as JSON for automated processing
kubectl mtv health -o json | jq '.overallStatus'

# Check if the system is healthy
if [ "$(kubectl mtv health -o json | jq -r '.overallStatus')" = "Healthy" ]; then
  echo "System is healthy, proceeding with migration"
fi
```

### Namespace-Scoped Check

```bash
kubectl mtv health -n production-migrations
```

## Next Steps

After verifying system health:

1. **Manage ForkliftController Settings**: Tune performance and enable features in [Chapter 25: Settings Management](/kubectl-mtv/25-settings-management)
2. **Review Command Reference**: See all available commands in [Chapter 26: Command Reference](/kubectl-mtv/26-command-reference)

---

*Previous: [Chapter 23: Integration with KubeVirt Tools](/kubectl-mtv/23-integration-with-kubevirt-tools)*
*Next: [Chapter 25: Settings Management](/kubectl-mtv/25-settings-management)*
