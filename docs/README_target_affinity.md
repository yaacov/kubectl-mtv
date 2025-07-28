# Using Target Affinity in Migration Plans

The `targetAffinity` flag allows you to control how migrated virtual machines are scheduled onto target nodes in your Kubernetes cluster. This feature uses the KARL (Kubernetes Affinity Rule Language) interpreter to translate a human-readable syntax into complex Kubernetes affinity and anti-affinity rules.

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

### Prefer Nodes in a Specific Zone

This rule tells the scheduler to prefer nodes with the label `topology.kubernetes.io/zone=us-east-1a`, but allows scheduling on other nodes if none are available.

**Use Case:** Placing VMs in a preferred availability zone for resilience or cost optimization, while still allowing for fallback.

```bash
--target-affinity 'PREFER nodes(topology.kubernetes.io/zone=us-east-1a)'
```

**Generated YAML:**
```yaml
spec:
  targetAffinity:
    nodeAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 1
        preference:
          matchExpressions:
          - key: topology.kubernetes.io/zone
            operator: In
            values:
            - us-east-1a
```

### Avoid Nodes with a Specific Taint

This rule prevents the migrated VM from being scheduled on nodes that have the taint `special-workload=true`.

**Use Case:** Ensuring that general-purpose VMs do not land on nodes reserved for specialized tasks (e.g., nodes with GPUs).

```bash
--target-affinity 'AVOID nodes(special-workload=true)'
```

**Generated YAML:**
```yaml
spec:
  targetAffinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: special-workload
            operator: NotIn
            values:
            - "true"
```

## How It Works

When you provide a `targetAffinity` rule, `kubectl-mtv` uses the KARL interpreter to convert the string into a standard Kubernetes `Affinity` object. This object is then embedded into the `Plan` custom resource. The Forklift controller reads this affinity rule and applies it to the `VirtualMachine` resource it creates, which in turn influences the Kubernetes scheduler's decision-making process.

For more information on the underlying Kubernetes affinity and anti-affinity concepts, please refer to the official Kubernetes documentation on [Assigning Pods to Nodes](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node-affinity/). 