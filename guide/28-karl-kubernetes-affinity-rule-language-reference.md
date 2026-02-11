---
layout: page
title: "Chapter 28: KARL - Kubernetes Affinity Rule Language Reference"
---

KARL (Kubernetes Affinity Rule Language) is a concise, human-readable syntax for expressing Kubernetes pod affinity and anti-affinity rules. In kubectl-mtv, KARL is used with the `--target-affinity` and `--convertor-affinity` flags to control where migrated VMs and convertor pods are scheduled.

This chapter is a self-contained reference for KARL syntax, rule types, topology keys, label selectors, and usage patterns.

## What is KARL?

Kubernetes affinity rules are normally expressed as deeply nested YAML structures. KARL compresses that complexity into a single line:

```
RULE_TYPE pods(selector[,selector...]) on TOPOLOGY [weight=N]
```

The KARL interpreter in kubectl-mtv parses this line and generates the corresponding Kubernetes `podAffinity` or `podAntiAffinity` specification.

### YAML Equivalent Example

The KARL rule:

```
REQUIRE pods(app=database) on node
```

is equivalent to this Kubernetes YAML:

```yaml
affinity:
  podAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
            - key: app
              operator: In
              values:
                - database
        topologyKey: kubernetes.io/hostname
```

## Where KARL is Used in kubectl-mtv

KARL appears in two flags, available on `create plan` and `patch plan`:

### `--target-affinity`

Controls where the **migrated VM** runs in the target cluster for its operational lifetime:

```bash
kubectl mtv create plan my-plan \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(app=database) on node" \
  --vms "app-server-01"
```

### `--convertor-affinity`

Controls where the **temporary convertor pod** (virt-v2v) runs during the migration process:

```bash
kubectl mtv create plan my-plan \
  --source vsphere-prod \
  --convertor-affinity "PREFER pods(app=storage-controller) on node weight=80" \
  --vms "large-vm-01"
```

## KARL Syntax

```
RULE_TYPE pods(selector[,selector...]) on TOPOLOGY [weight=N]
```

### Components

| Component | Required | Description |
|-----------|----------|-------------|
| `RULE_TYPE` | Yes | One of `REQUIRE`, `PREFER`, `AVOID`, `REPEL` |
| `pods(...)` | Yes | Label selector targeting existing pods |
| `on TOPOLOGY` | Yes | Topology domain: `node`, `zone`, `region`, `rack` |
| `weight=N` | Only for `PREFER`/`REPEL` | Scheduling weight from 1 to 100 |

## Rule Types

KARL defines four rule types that map to Kubernetes affinity and anti-affinity concepts:

### `REQUIRE` -- Hard Affinity

The pod **must** be scheduled on a topology domain where pods matching the selector are already running. If no node satisfies the rule, the pod stays in `Pending`.

```
REQUIRE pods(app=database) on node
```

**Use when:** The workload has a strict dependency -- for example, an application that must be co-located with its database.

### `PREFER` -- Soft Affinity

The scheduler **tries** to place the pod near matching pods, but will place it elsewhere if necessary. Higher weights are given more priority.

```
PREFER pods(app=cache) on zone weight=80
```

**Use when:** Co-location is beneficial for performance but not mandatory.

### `AVOID` -- Hard Anti-Affinity

The pod **must not** be scheduled on a topology domain where matching pods are running. If no node satisfies the rule, the pod stays in `Pending`.

```
AVOID pods(app=web-server) on node
```

**Use when:** Workloads must be separated -- for example, spreading replicas for high availability.

### `REPEL` -- Soft Anti-Affinity

The scheduler **tries** to keep the pod away from matching pods, but will co-locate them if necessary. Higher weights increase the preference.

```
REPEL pods(tier in [batch,worker]) on zone weight=50
```

**Use when:** Separation is preferred but not critical -- for example, distributing load across zones.

### Rule Type Summary

| Rule | Kubernetes Mapping | Hard/Soft | Placement Effect |
|------|--------------------|-----------|------------------|
| `REQUIRE` | `podAffinity.requiredDuringSchedulingIgnoredDuringExecution` | Hard | Must co-locate |
| `PREFER` | `podAffinity.preferredDuringSchedulingIgnoredDuringExecution` | Soft | Prefer co-locate |
| `AVOID` | `podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution` | Hard | Must separate |
| `REPEL` | `podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution` | Soft | Prefer separate |

## Topology Keys

The topology key determines the scheduling domain:

| KARL Keyword | Kubernetes topologyKey | Meaning |
|--------------|----------------------|---------|
| `node` | `kubernetes.io/hostname` | Same physical/virtual node |
| `zone` | `topology.kubernetes.io/zone` | Same availability zone |
| `region` | `topology.kubernetes.io/region` | Same cloud region |
| `rack` | `topology.kubernetes.io/rack` | Same physical rack |

Choose the topology key based on what "near" or "away from" means for your workload:

- **`node`** -- strictest: same machine. Use for latency-critical co-location or HA spreading.
- **`zone`** -- same failure domain. Use for availability-zone-aware placement.
- **`region`** -- broadest standard scope. Use for data sovereignty or compliance.
- **`rack`** -- physical rack proximity. Use when rack-level failure isolation matters.

## Label Selectors

Inside `pods(...)`, you define which existing pods the rule references. Multiple selectors are separated by commas and are **AND-ed** together.

### Selector Types

| Syntax | Meaning | Example |
|--------|---------|---------|
| `key=value` | Label equals value | `app=database` |
| `key in [v1,v2,v3]` | Label is one of the values | `tier in [web,api]` |
| `key not in [v1,v2]` | Label is not any of the values | `env not in [staging,dev]` |
| `has key` | Label exists (any value) | `has monitoring` |
| `not has key` | Label does not exist | `not has temporary` |

### Combining Selectors

All selectors within `pods(...)` are AND-ed:

```
REQUIRE pods(app=web, tier=frontend, has monitoring) on node
```

This matches pods that have **all three** conditions: `app=web` AND `tier=frontend` AND the label `monitoring` exists.

## Weight (Soft Rules Only)

For `PREFER` and `REPEL` rules, the optional `weight=N` parameter controls scheduling priority:

- Range: **1 to 100**
- Default (if omitted): **100**
- Higher values mean stronger preference

```
PREFER pods(app=cache) on zone weight=90
REPEL pods(app=batch) on node weight=30
```

When multiple soft rules compete, the scheduler adds up the weights of all satisfied rules per node and picks the node with the highest total.

## Examples

### Co-location Patterns

```bash
# Application must run on same node as database
kubectl mtv create plan app-db-colocation \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(app=database) on node" \
  --vms "app-server-01,app-server-02"

# Prefer placing near cache for performance
kubectl mtv create plan near-cache \
  --source vsphere-prod \
  --target-affinity "PREFER pods(app=redis,role=cache) on node weight=90" \
  --vms "web-app-01"

# Co-locate within the same zone as the API tier
kubectl mtv create plan api-zone \
  --source vsphere-prod \
  --target-affinity "PREFER pods(tier=api) on zone weight=80" \
  --vms "api-client-01,api-client-02"
```

### Anti-Affinity Patterns

```bash
# Spread web servers across nodes for HA
kubectl mtv create plan ha-web \
  --source vsphere-prod \
  --target-affinity "AVOID pods(app=web-server) on node" \
  --vms "web-01,web-02,web-03"

# Soft spread of workers across zones
kubectl mtv create plan spread-workers \
  --source vsphere-prod \
  --target-affinity "REPEL pods(app=worker) on zone weight=60" \
  --vms "worker-01,worker-02,worker-03"

# Keep away from batch workloads
kubectl mtv create plan avoid-batch \
  --source vsphere-prod \
  --target-affinity "AVOID pods(tier in [batch,worker]) on node" \
  --vms "latency-sensitive-app"
```

### Zone and Region Placement

```bash
# Require same zone for compliance
kubectl mtv create plan zone-compliance \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(compliance-zone=east) on zone" \
  --vms "regulated-app-01"

# Spread across zones for disaster recovery
kubectl mtv create plan multi-zone \
  --source vsphere-prod \
  --target-affinity "AVOID pods(app=frontend) on zone" \
  --vms "frontend-01,frontend-02,frontend-03"

# Prefer same region for data locality
kubectl mtv create plan regional-data \
  --source vsphere-prod \
  --target-affinity "PREFER pods(data-region=us-west) on region weight=95" \
  --vms "data-processor-01"
```

### Advanced Label Selectors

```bash
# Multiple AND-ed selectors
kubectl mtv create plan multi-selector \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(app=web,tier=frontend,has monitoring) on node" \
  --vms "monitored-web-01"

# Using set-based selectors
kubectl mtv create plan set-selector \
  --source vsphere-prod \
  --target-affinity "AVOID pods(env in [staging,dev]) on node" \
  --vms "production-app-01"

# Excluding labels
kubectl mtv create plan exclude-ephemeral \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(storage not in [ephemeral]) on node" \
  --vms "persistent-workload-01"
```

### Convertor Pod Optimization

```bash
# Place convertor near storage controller for faster disk transfer
kubectl mtv create plan fast-transfer \
  --source vsphere-prod \
  --convertor-affinity "PREFER pods(app=storage-controller) on node weight=80" \
  --vms "large-disk-vm-01"

# Isolate convertor from production workloads
kubectl mtv create plan isolated-conversion \
  --source vsphere-prod \
  --convertor-affinity "AVOID pods(environment=production) on node" \
  --vms "risky-vm-01"

# Convertor on high-performance nodes
kubectl mtv create plan perf-convertor \
  --source vsphere-prod \
  --convertor-affinity "REQUIRE pods(node-role=migration) on node" \
  --convertor-node-selector "performance=high" \
  --vms "big-vm-01"
```

### Combining Target and Convertor Affinity

```bash
# Different rules for migration process vs. operational lifetime
kubectl mtv create plan full-placement \
  --source vsphere-prod \
  --convertor-affinity "PREFER pods(app=ceph-osd) on node weight=90" \
  --target-affinity "AVOID pods(app=database) on node" \
  --target-node-selector "tier=web" \
  --target-labels "app=web,environment=production" \
  --vms "web-server-01"
```

## Multi-Tier Application Example

A complete example placing a three-tier application with appropriate affinity rules:

```bash
# Web tier: spread across nodes, label as frontend
kubectl mtv create plan web-tier \
  --source vsphere-prod \
  --target-affinity "AVOID pods(tier=web) on node" \
  --target-labels "tier=web,layer=frontend" \
  --vms "web-01,web-02,web-03"

# App tier: prefer near web tier within the same zone
kubectl mtv create plan app-tier \
  --source vsphere-prod \
  --target-affinity "PREFER pods(tier=web) on zone weight=80" \
  --target-labels "tier=app,layer=business" \
  --vms "app-01,app-02"

# Data tier: require dedicated database nodes
kubectl mtv create plan data-tier \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(node-role=database) on node" \
  --target-node-selector "dedicated=database" \
  --target-labels "tier=data,layer=persistence" \
  --vms "db-primary,db-replica"
```

## Patching Affinity Rules

Affinity rules can be updated on existing plans using `patch plan`:

```bash
# Update target affinity
kubectl mtv patch plan my-plan \
  --target-affinity "PREFER pods(app=cache) on zone weight=70"

# Update convertor affinity
kubectl mtv patch plan my-plan \
  --convertor-affinity "REQUIRE pods(storage=fast) on node"
```

## Troubleshooting

### Pod Stuck in Pending

If a migrated VM or convertor pod is stuck in `Pending`, the affinity rule may be unsatisfiable:

```bash
# Check scheduling events
kubectl describe pod <pod-name> | grep -A5 Events

# Look for FailedScheduling messages
kubectl get events --field-selector reason=FailedScheduling
```

Common causes:
- **No matching pods exist** -- the label selector does not match any running pod
- **No node satisfies topology** -- all candidate nodes violate the hard rule
- **Conflicting rules** -- multiple REQUIRE or AVOID rules cannot all be satisfied simultaneously

### Relaxing Rules

If a hard rule is too restrictive, switch to the soft equivalent:

| From | To |
|------|----|
| `REQUIRE` | `PREFER ... weight=100` |
| `AVOID` | `REPEL ... weight=100` |

```bash
# Hard rule causing Pending:
--target-affinity "REQUIRE pods(app=database) on node"

# Relaxed soft rule:
--target-affinity "PREFER pods(app=database) on node weight=100"
```

### Verifying Generated Affinity

After creating a plan, inspect the generated Kubernetes affinity spec:

```bash
# Check the VM spec
kubectl get vm <vm-name> -o yaml | grep -A 30 affinity
```

## Quick Reference Card

```
RULE TYPES
  REQUIRE   hard affinity       must co-locate
  PREFER    soft affinity       prefer co-locate    weight=1..100
  AVOID     hard anti-affinity  must separate
  REPEL     soft anti-affinity  prefer separate     weight=1..100

TOPOLOGY
  node      kubernetes.io/hostname
  zone      topology.kubernetes.io/zone
  region    topology.kubernetes.io/region
  rack      topology.kubernetes.io/rack

SELECTORS (comma-separated, AND-ed)
  key=value             equality
  key in [v1,v2]        set membership
  key not in [v1,v2]    set exclusion
  has key               label exists
  not has key           label absent

SYNTAX
  RULE_TYPE pods(sel1,sel2,...) on TOPOLOGY [weight=N]

FLAGS
  --target-affinity     operational VM placement
  --convertor-affinity  migration-time convertor pod placement
```

## Built-in Help

Access the KARL reference directly from the command line:

```bash
kubectl mtv help karl
```

## Further Reading

- [Chapter 15: Target VM Placement](/kubectl-mtv/15-target-vm-placement) -- detailed target affinity scenarios and examples
- [Chapter 16: Migration Process Optimization](/kubectl-mtv/16-migration-process-optimization) -- convertor affinity for migration performance
- [Chapter 18: Advanced Plan Patching](/kubectl-mtv/18-advanced-plan-patching) -- patching affinity rules on existing plans

---

*Previous: [Chapter 27: TSL - Tree Search Language Reference](/kubectl-mtv/27-tsl-tree-search-language-reference)*
*Next: Return to [Table of Contents](/kubectl-mtv/)*
