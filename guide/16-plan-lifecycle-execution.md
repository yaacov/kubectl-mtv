---
layout: page
title: "Chapter 16: Plan Lifecycle Execution"
---

# Chapter 16: Plan Lifecycle Execution

Migration plans follow a defined lifecycle from creation through completion. This chapter covers the complete execution workflow, including starting migrations, managing warm migration cutover, canceling workloads, monitoring progress, and handling plan archival. All command examples are verified against the implementation.

## Overview: Migration Plan States and Lifecycle

### Plan Lifecycle States

Migration plans progress through distinct states during execution:

1. **Created**: Plan exists but migration hasn't started
2. **Started**: Migration is actively running
3. **Warm Running**: Warm migration in pre-copy phase (warm migrations only)
4. **Cutover Scheduled**: Warm migration cutover time set (warm migrations only)
5. **Completed**: All VMs successfully migrated
6. **Failed**: Migration encountered unrecoverable errors
7. **Canceled**: Migration was manually canceled
8. **Archived**: Plan marked for long-term retention

### Migration Types and Execution Flow

Different migration types follow distinct execution patterns:

#### Cold Migration Flow
```
Created - Started - Completed/Failed/Canceled
```

#### Warm Migration Flow
```
Created - Started - Warm Running - Cutover Scheduled - Completed/Failed/Canceled
```

#### Live Migration Flow (KubeVirt sources only)
```
Created - Started - Completed/Failed/Canceled
```

### Command Overview

Plan lifecycle management uses these verified commands:
- `kubectl mtv start plan` - Begin migration execution
- `kubectl mtv cutover plan` - Schedule warm migration cutover
- `kubectl mtv cancel plan` - Cancel specific VMs in running migration
- `kubectl mtv archive plan` - Archive completed plans
- `kubectl mtv unarchive plan` - Restore archived plans
- `kubectl mtv get plan` - Monitor migration progress
- `kubectl mtv describe plan` - Get detailed migration status

## Starting a Migration

### Basic Plan Execution

#### Single Plan Execution

```bash
# Start a specific migration plan
kubectl mtv start plan production-migration

# Start multiple plans simultaneously
kubectl mtv start plan web-tier-migration database-migration cache-migration
```

#### Bulk Plan Execution

```bash
# Start all plans in the current namespace
kubectl mtv start plan --all

# Start all plans in a specific namespace
kubectl mtv start plan --all -n production-migrations
```

### Advanced Plan Startup Options

#### Scheduled Migration Start

Start migrations with predefined cutover times for warm migrations:

```bash
# Start with cutover scheduled for specific time
kubectl mtv start plan warm-migration \
  --cutover "2024-12-31T23:59:00Z"

# Start with cutover in 1 hour (default if no cutover specified for warm migrations)
kubectl mtv start plan warm-migration

# Start with cutover using relative time
kubectl mtv start plan warm-migration \
  --cutover "$(date -d '+2 hours' --iso-8601=seconds)"

# Start multiple plans with same cutover time
kubectl mtv start plan web-migration db-migration \
  --cutover "2024-01-15T02:00:00Z"
```

#### Production Migration Examples

```bash
# Production weekend migration with scheduled cutover
kubectl mtv start plan production-phase1 \
  --cutover "$(date -d 'next Saturday 2:00 AM' --iso-8601=seconds)"

# Emergency migration (immediate start)
kubectl mtv start plan emergency-migration

# Maintenance window migration
kubectl mtv start plan maintenance-migration \
  --cutover "2024-01-20T01:00:00Z"

# Bulk weekend migration
kubectl mtv start plan --all \
  --cutover "$(date -d 'next Sunday 3:00 AM' --iso-8601=seconds)"
```

### Monitoring Migration Start

```bash
# Check plan status after starting
kubectl mtv get plan production-migration

# Watch plan status in real-time
kubectl mtv get plan production-migration -w

# Get detailed startup status
kubectl mtv describe plan production-migration

# Monitor all running plans
kubectl mtv get plans
```

## Warm Migration Cutover

### Understanding Warm Migration Cutover

Warm migrations consist of two phases:
1. **Pre-copy Phase**: Data transfer while source VM remains running
2. **Cutover Phase**: Final synchronization with brief downtime

### Scheduling Cutover

#### Setting Cutover Time

```bash
# Schedule cutover for specific time
kubectl mtv cutover plan warm-production \
  --cutover "2024-01-15T03:00:00Z"

# Schedule immediate cutover (current time)
kubectl mtv cutover plan warm-production

# Schedule cutover using relative time
kubectl mtv cutover plan warm-production \
  --cutover "$(date -d '+30 minutes' --iso-8601=seconds)"

# Bulk cutover scheduling for multiple plans
kubectl mtv cutover plan web-warm db-warm cache-warm \
  --cutover "2024-01-15T02:30:00Z"
```

#### Cutover Management Scenarios

```bash
# Emergency cutover (immediate)
kubectl mtv cutover plan emergency-warm

# Delayed cutover (reschedule)
kubectl mtv cutover plan delayed-warm \
  --cutover "$(date -d '+1 day 2:00 AM' --iso-8601=seconds)"

# Coordinated multi-application cutover
kubectl mtv cutover plan app-tier-warm \
  --cutover "2024-01-20T02:00:00Z"
kubectl mtv cutover plan database-warm \
  --cutover "2024-01-20T02:05:00Z"  # 5 minutes later

# Bulk cutover for all warm migrations
kubectl mtv cutover plan --all \
  --cutover "$(date -d 'next Sunday 3:00 AM' --iso-8601=seconds)"
```

### Monitoring Cutover Progress

```bash
# Monitor cutover status
kubectl mtv get plan warm-migration -w

# Check cutover timing
kubectl mtv describe plan warm-migration | grep -i cutover

# Verify cutover completion
kubectl mtv get plan warm-migration -o jsonpath='{.status.phase}'
```

## Canceling Workloads

### VM-Specific Cancellation

The cancel command allows canceling specific VMs within a running migration plan:

#### Single VM Cancellation

```bash
# Cancel a specific VM in running migration
kubectl mtv cancel plan production-migration \
  --vms web-server-01

# Cancel multiple VMs by name
kubectl mtv cancel plan production-migration \
  --vms "web-server-01,web-server-02,cache-server-01"
```

#### File-Based VM Cancellation

```bash
# Create file with VMs to cancel
cat > vms-to-cancel.yaml << 'EOF'
- web-server-problematic
- database-secondary
- cache-server-02
EOF

# Cancel VMs listed in file
kubectl mtv cancel plan production-migration \
  --vms @vms-to-cancel.yaml

# JSON format file example
cat > vms-to-cancel.json << 'EOF'
["web-server-01", "problematic-vm", "test-server"]
EOF

kubectl mtv cancel plan test-migration \
  --vms @vms-to-cancel.json
```

### Cancellation Scenarios

#### Problem VM Cancellation

```bash
# Cancel VMs encountering issues during migration
kubectl mtv cancel plan large-migration \
  --vms "stuck-vm-01,error-vm-02,timeout-vm-03"

# Cancel non-critical VMs to focus on critical ones
kubectl mtv cancel plan mixed-priority \
  --vms "test-vm-01,dev-vm-02,sandbox-vm-03"
```

#### Strategic Cancellation for Performance

```bash
# Cancel resource-intensive VMs during peak hours
kubectl mtv cancel plan performance-sensitive \
  --vms "large-database-vm,memory-intensive-app"

# Cancel and reschedule for off-peak hours
kubectl mtv cancel plan peak-hour-migration \
  --vms @resource-intensive-vms.yaml
# Later create new plan for these VMs
```

### Monitoring Cancellation Effects

```bash
# Check which VMs were canceled
kubectl mtv describe plan production-migration | grep -i cancel

# Verify plan continues with remaining VMs
kubectl mtv get plan production-migration

# Monitor overall plan progress after cancellation
kubectl mtv get plan production-migration -w
```

## Migration Progress Monitoring

### Real-Time Monitoring

#### Basic Status Monitoring

```bash
# Get current status of all plans
kubectl mtv get plans

# Monitor specific plan progress
kubectl mtv get plan production-migration

# Watch plan status with updates
kubectl mtv get plan production-migration -w

# Get detailed plan information
kubectl mtv describe plan production-migration
```

#### Advanced Status Queries

```bash
# Get plan status in JSON for processing
kubectl mtv get plan production-migration -o json

# Extract specific status information
kubectl mtv get plan production-migration \
  -o jsonpath='{.status.phase}'

# Get VM-level status
kubectl mtv get plan production-migration \
  -o jsonpath='{.status.vms[*].name}'

# Check for any error conditions
kubectl mtv describe plan production-migration | grep -i error
```

### Monitoring Multiple Plans

```bash
# Monitor all plans across namespaces
kubectl mtv get plans --all-namespaces

# Filter plans by status using kubectl
kubectl mtv get plans -o json | jq '.items[] | select(.status.phase == "Running")'

# Monitor plans with custom output
kubectl mtv get plans -o custom-columns="NAME:.metadata.name,PHASE:.status.phase,VMS:.spec.vms | length"
```

### Comprehensive Monitoring Dashboard

```bash
# Create monitoring script for multiple plans
#!/bin/bash

echo "=== Migration Plan Status Dashboard ==="
echo

echo "Running Plans:"
kubectl mtv get plans -o json | jq -r '.items[] | select(.status.phase == "Running") | .metadata.name'
echo

echo "Completed Plans:"
kubectl mtv get plans -o json | jq -r '.items[] | select(.status.phase == "Succeeded") | .metadata.name'
echo

echo "Failed Plans:"
kubectl mtv get plans -o json | jq -r '.items[] | select(.status.phase == "Failed") | .metadata.name'
echo

echo "Detailed Status for Running Plans:"
for plan in $(kubectl mtv get plans -o json | jq -r '.items[] | select(.status.phase == "Running") | .metadata.name'); do
  echo "Plan: $plan"
  kubectl mtv describe plan "$plan" | grep -A 10 "Status:"
  echo "---"
done
```

## Archiving and Unarchiving Plans

### Plan Archival

Archiving helps manage completed or obsolete plans for long-term retention:

#### Single Plan Archival

```bash
# Archive a completed migration plan
kubectl mtv archive plan completed-migration

# Archive multiple plans
kubectl mtv archive plan old-migration-1 old-migration-2 completed-test
```

#### Bulk Plan Archival

```bash
# Archive all plans in namespace
kubectl mtv archive plan --all

# Archive all plans in specific namespace
kubectl mtv archive plan --all -n old-migrations
```

### Plan Restoration

Restore archived plans when needed for reference or reuse:

#### Single Plan Restoration

```bash
# Unarchive a specific plan
kubectl mtv unarchive plan archived-migration

# Unarchive multiple plans
kubectl mtv unarchive plan reference-migration template-migration
```

#### Bulk Plan Restoration

```bash
# Unarchive all plans in namespace
kubectl mtv unarchive plan --all

# Unarchive all plans in specific namespace
kubectl mtv unarchive plan --all -n restored-migrations
```

### Archive Management Scenarios

#### Periodic Cleanup

```bash
# Archive completed migrations monthly
kubectl mtv get plans -o json | \
  jq -r '.items[] | select(.status.phase == "Succeeded" and (.metadata.creationTimestamp | fromdateiso8601) < (now - 30*24*3600)) | .metadata.name' | \
  xargs -I {} kubectl mtv archive plan {}

# Archive failed migrations after analysis
kubectl mtv get plans -o json | \
  jq -r '.items[] | select(.status.phase == "Failed") | .metadata.name' | \
  xargs -I {} kubectl mtv archive plan {}
```

#### Template Management

```bash
# Archive template plans for reuse
kubectl mtv archive plan template-web-migration template-db-migration

# Restore templates when needed
kubectl mtv unarchive plan template-web-migration
# Modify and use as basis for new migration
```

## Complete Migration Workflow Examples

### Production Migration Workflow

```bash
# Phase 1: Create and validate plan
kubectl mtv create plan production-q4 \
  --source vsphere-prod --target openshift-prod \
  --vms @production-vms.yaml \
  --migration-type warm \
  --network-mapping prod-net-map \
  --storage-mapping prod-storage-map

# Phase 2: Start migration with scheduled cutover
kubectl mtv start plan production-q4 \
  --cutover "$(date -d 'next Saturday 2:00 AM' --iso-8601=seconds)"

# Phase 3: Monitor progress
kubectl mtv get plan production-q4 -w

# Phase 4: Handle issues (if needed)
kubectl mtv cancel plan production-q4 \
  --vms problematic-vm-01

# Phase 5: Execute cutover
# (Automatic based on scheduled time, or manual adjustment)
kubectl mtv cutover plan production-q4 \
  --cutover "$(date --iso-8601=seconds)"

# Phase 6: Verify completion and archive
kubectl mtv describe plan production-q4
kubectl mtv archive plan production-q4
```

### Emergency Migration Workflow

```bash
# Emergency migration with immediate execution
kubectl mtv create plan emergency-recovery \
  --source vsphere-dr --target openshift-prod \
  --vms critical-app-01,critical-db-01 \
  --migration-type live

# Start immediately
kubectl mtv start plan emergency-recovery

# Monitor closely
kubectl mtv get plan emergency-recovery -w

# Handle any issues quickly
if [ "$(kubectl mtv get plan emergency-recovery -o jsonpath='{.status.phase}')" == "Failed" ]; then
  kubectl mtv describe plan emergency-recovery
  # Take corrective action
fi
```

### Development Environment Migration

```bash
# Development migration with testing
kubectl mtv create plan dev-environment \
  --source vsphere-dev --target openshift-dev \
  --vms @dev-vms.yaml \
  --migration-type cold \
  --target-namespace development

# Start development migration
kubectl mtv start plan dev-environment

# Test cancellation (if needed)
kubectl mtv cancel plan dev-environment \
  --vms test-vm-01,experimental-vm

# Complete and clean up
kubectl mtv get plan dev-environment
# Keep for reference, don't archive immediately
```

### Multi-Phase Migration Campaign

```bash
# Phase 1: Web tier migration
kubectl mtv start plan web-tier-migration \
  --cutover "$(date -d '+1 hour' --iso-8601=seconds)"

# Phase 2: Application tier (after web tier completion)
kubectl mtv start plan app-tier-migration \
  --cutover "$(date -d '+2 hours' --iso-8601=seconds)"

# Phase 3: Database tier (final phase)  
kubectl mtv start plan database-migration \
  --cutover "$(date -d '+3 hours' --iso-8601=seconds)"

# Monitor all phases
for plan in web-tier-migration app-tier-migration database-migration; do
  echo "=== $plan ==="
  kubectl mtv get plan "$plan"
done

# Archive completed phases
kubectl mtv archive plan web-tier-migration app-tier-migration
```

## Advanced Lifecycle Management

### Lifecycle Automation Scripts

#### Migration Status Monitor

```bash
#!/bin/bash
# migration-monitor.sh - Monitor migration plan status

PLAN_NAME="$1"
CHECK_INTERVAL="${2:-30}"

if [ -z "$PLAN_NAME" ]; then
  echo "Usage: $0 <plan-name> [check-interval-seconds]"
  exit 1
fi

echo "Monitoring plan: $PLAN_NAME"
echo "Check interval: ${CHECK_INTERVAL}s"
echo "Press Ctrl+C to stop"
echo

while true; do
  STATUS=$(kubectl mtv get plan "$PLAN_NAME" -o jsonpath='{.status.phase}' 2>/dev/null)
  
  if [ $? -eq 0 ]; then
    TIMESTAMP=$(date --iso-8601=seconds)
    echo "[$TIMESTAMP] Plan: $PLAN_NAME, Status: $STATUS"
    
    if [ "$STATUS" = "Succeeded" ] || [ "$STATUS" = "Failed" ]; then
      echo "Migration completed with status: $STATUS"
      kubectl mtv describe plan "$PLAN_NAME" | grep -A 5 "Conditions:"
      break
    fi
  else
    echo "Error: Could not get status for plan $PLAN_NAME"
  fi
  
  sleep "$CHECK_INTERVAL"
done
```

#### Automated Cleanup Script

```bash
#!/bin/bash
# cleanup-completed-migrations.sh - Archive completed migrations

NAMESPACE="${1:-default}"
DRY_RUN="${2:-false}"

echo "Cleaning up completed migrations in namespace: $NAMESPACE"

# Get completed plans
COMPLETED_PLANS=$(kubectl mtv get plans -n "$NAMESPACE" -o json | \
  jq -r '.items[] | select(.status.phase == "Succeeded") | .metadata.name')

if [ -z "$COMPLETED_PLANS" ]; then
  echo "No completed plans found in namespace $NAMESPACE"
  exit 0
fi

echo "Found completed plans:"
echo "$COMPLETED_PLANS"

if [ "$DRY_RUN" = "true" ]; then
  echo "DRY RUN: Would archive these plans"
else
  echo "Archiving plans..."
  echo "$COMPLETED_PLANS" | while read -r plan; do
    echo "Archiving: $plan"
    kubectl mtv archive plan "$plan" -n "$NAMESPACE"
  done
fi
```

### Integration with Monitoring Systems

#### Prometheus Metrics Collection

```bash
# Export plan status for monitoring
kubectl mtv get plans -o json | \
  jq -r '.items[] | "migration_plan_status{name=\"\(.metadata.name)\",namespace=\"\(.metadata.namespace)\",phase=\"\(.status.phase)\"} 1"' \
  > /var/lib/node_exporter/migration_plans.prom
```

#### Alerting Integration

```bash
# Check for failed migrations and alert
FAILED_PLANS=$(kubectl mtv get plans --all-namespaces -o json | \
  jq -r '.items[] | select(.status.phase == "Failed") | "\(.metadata.namespace)/\(.metadata.name)"')

if [ -n "$FAILED_PLANS" ]; then
  echo "ALERT: Failed migration plans detected:"
  echo "$FAILED_PLANS"
  # Send to alerting system (Slack, PagerDuty, etc.)
fi
```

## Troubleshooting Lifecycle Operations

### Common Issues and Solutions

#### Plan Won't Start

```bash
# Check plan validation
kubectl mtv describe plan stuck-plan | grep -A 10 "Conditions"

# Verify provider connectivity
kubectl mtv get providers

# Check resource availability
kubectl get nodes
kubectl describe plan stuck-plan
```

#### Cutover Not Executing

```bash
# Check cutover time setting
kubectl mtv describe plan warm-plan | grep -i cutover

# Verify warm migration is in correct state
kubectl mtv get plan warm-plan -o jsonpath='{.status.phase}'

# Reset cutover time if needed
kubectl mtv cutover plan warm-plan \
  --cutover "$(date --iso-8601=seconds)"
```

#### Cancellation Not Working

```bash
# Verify VM names are correct
kubectl mtv get plan problem-plan -o jsonpath='{.spec.vms[*].name}'

# Check if VMs are in cancelable state
kubectl mtv describe plan problem-plan

# Force cancellation through kubectl if needed
kubectl patch plan problem-plan --type='merge' -p='{"spec":{"vms":[]}}'
```

## Next Steps

After mastering plan lifecycle execution:

1. **Troubleshooting**: Learn to debug migration issues in [Chapter 17: Debugging and Troubleshooting](17-debugging-and-troubleshooting.md)
2. **Best Practices**: Master operational excellence in [Chapter 18: Best Practices and Security](18-best-practices-and-security.md)
3. **AI Integration**: Explore advanced automation in [Chapter 19: Model Context Protocol (MCP) Server Integration](19-model-context-protocol-mcp-server-integration.md)
4. **Tool Integration**: Learn KubeVirt ecosystem integration in [Chapter 20: Integration with KubeVirt Tools](20-integration-with-kubevirt-tools.md)

---

*Previous: [Chapter 15: Advanced Plan Patching](15-advanced-plan-patching.md)*  
*Next: [Chapter 17: Debugging and Troubleshooting](17-debugging-and-troubleshooting.md)*
