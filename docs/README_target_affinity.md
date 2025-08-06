# Using Target Affinity in Migration Plans

The `targetAffinity` flag allows you to control how migrated virtual machines are scheduled onto target nodes in your Kubernetes cluster. This feature uses the KARL (Kubernetes Affinity Rule Language) interpreter to translate a human-readable syntax into Kubernetes pod affinity and anti-affinity rules.

This guide explains how to use `targetAffinity` to achieve common scheduling scenarios.

## Basic Syntax

The basic syntax for the `targetAffinity` flag is:

```bash
kubectl mtv create plan <plan-name> \
  --source-provider <source> \
  --vms <vm-list> \
  --target-affinity "<RULE>"
```

The `<RULE>` is a KARL expression that defines your scheduling constraints.

## KARL Syntax Reference

KARL supports pod affinity and anti-affinity rules with the following syntax:

**Format:** `[RULE_TYPE] pods([SELECTORS]) on [TOPOLOGY]`

**Rule Types:**
- `REQUIRE` - Hard pod affinity (must be satisfied)
- `PREFER` - Soft pod affinity (preferred but not required) 
- `AVOID` - Hard pod anti-affinity (must be avoided)
- `REPEL` - Soft pod anti-affinity (avoided if possible)

**Topology Keys:**
- `node` - Same kubernetes node (kubernetes.io/hostname)
- `zone` - Same availability zone (topology.kubernetes.io/zone)  
- `region` - Same region (topology.kubernetes.io/region)
- `rack` - Same rack (topology.kubernetes.io/rack)

**Label Selectors:**
- `app=database` - Simple equality
- `tier in [web,api]` - Match multiple values
- `env not in [dev,test]` - Exclude values
- `has monitoring` - Key exists
- `not has legacy` - Key does not exist

**Limitations:**
- Only `pods()` targets supported (no node affinity)
- No AND/OR compound logic (single rule only)
- Generates pod affinity/anti-affinity (not node affinity)

## KARL Rule Examples

Here are some examples of KARL rules you can use with `targetAffinity`:

### Require Pods with a Specific Label on the Same Node

This rule ensures that the migrated VM will be scheduled on a node that is already running pods with the label `app=database`.

**Use Case:** Co-locating a migrated application server with its database for low-latency communication.

```bash
--target-affinity 'REQUIRE pods(app=database) on node'
```

**Generated YAML:**
```yaml
spec:
  targetAffinity:
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

### Prefer Same Zone as Web Pods

This rule tells the scheduler to prefer nodes that are running pods with the label `tier=web` in the same zone, but allows scheduling on other nodes if none are available.

**Use Case:** Co-locating migrated VMs with web tier pods for improved performance and reduced network latency.

```bash
--target-affinity 'PREFER pods(tier=web) on zone'
```

**Generated YAML:**
```yaml
spec:
  targetAffinity:
    podAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchExpressions:
            - key: tier
              operator: In
              values:
              - web
          topologyKey: topology.kubernetes.io/zone
```

### Avoid Same Node as Cache Pods

This rule prevents the migrated VM from being scheduled on nodes that are already running pods with the label `app=cache`.

**Use Case:** Ensuring that migrated VMs do not compete for resources with cache pods on the same node, improving performance isolation.

```bash
--target-affinity 'AVOID pods(app=cache) on node'
```

**Generated YAML:**
```yaml
spec:
  targetAffinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
          - key: app
            operator: In
            values:
            - cache
        topologyKey: kubernetes.io/hostname
```

### Soft Avoidance of Heavy Workload Pods

This rule uses soft anti-affinity to avoid scheduling on nodes with heavy workload pods, but allows it if necessary.

**Use Case:** Preferentially avoiding resource contention while maintaining scheduling flexibility.

```bash
--target-affinity 'REPEL pods(workload=heavy) on node'
```

**Generated YAML:**
```yaml
spec:
  targetAffinity:
    podAntiAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchExpressions:
            - key: workload
              operator: In
              values:
              - heavy
          topologyKey: kubernetes.io/hostname
```

## How It Works

When you provide a `targetAffinity` rule, `kubectl-mtv` uses the KARL interpreter to convert the string into a standard Kubernetes `Affinity` object. This object is then embedded into the `Plan` custom resource. The Forklift controller reads this affinity rule and applies it to the `VirtualMachine` resource it creates, which in turn influences the Kubernetes scheduler's decision-making process.

For more information on the underlying Kubernetes affinity and anti-affinity concepts, please refer to the official Kubernetes documentation on [Assigning Pods to Nodes](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node-affinity/). 