# Convertor Pod Scheduling and Optimization

This guide explains how to optimize the placement and resource utilization of **virt-v2v convertor pods** during the migration process. These are temporary pods that run during migration to handle disk conversion, driver installation, and data transfer from source VMs to target storage.

The convertor pod scheduling flags (`--convertor-affinity`, `--convertor-labels`, `--convertor-node-selector`) allow you to optimize migration performance by controlling where the conversion work happens and how it utilizes cluster resources.

## Overview

**Convertor pods are temporary migration infrastructure** that:
- Run virt-v2v to convert VMware disks and install KVM guest agents/drivers
- Handle data migration for cold migrations from VMware to KubeVirt
- Consume significant CPU, memory, and I/O resources during disk conversion
- Are automatically cleaned up after migration completion

**Why optimize convertor pod placement?**
- **Performance**: Co-locate with fast storage for optimal disk I/O
- **Resource Management**: Avoid interference with production workloads
- **Network Proximity**: Place near source infrastructure for data transfer
- **Cost Optimization**: Use appropriate node types for intensive conversion work

## Basic Syntax

### Convertor Labels

Apply labels to convertor pods for identification and resource management:

```bash
kubectl mtv create plan <plan-name> \
  --source-provider <source> \
  --vms <vm-list> \
  --convertor-labels "performance=high,workload=migration,batch=1"
```

### Convertor Node Selector

Constrain convertor pods to specific nodes:

```bash
kubectl mtv create plan <plan-name> \
  --source-provider <source> \
  --vms <vm-list> \
  --convertor-node-selector "node-type=storage,disk=nvme,zone=us-east"
```

### Convertor Affinity

Control advanced convertor pod placement using KARL syntax:

```bash
kubectl mtv create plan <plan-name> \
  --source-provider <source> \
  --vms <vm-list> \
  --convertor-affinity "REQUIRE pods(app=ceph-osd) on node"
```

## KARL Syntax for Convertor Affinity

Convertor affinity uses the same KARL (Kubernetes Affinity Rule Language) syntax as target affinity:

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

## Common Use Cases

### High-Performance Storage Access

Co-locate convertor pods with storage infrastructure for optimal I/O:

```bash
# Co-locate with Ceph storage nodes
kubectl mtv create plan high-io-migration \
  --source-provider vmware \
  --vms database1,database2,storage-vm \
  --convertor-affinity 'REQUIRE pods(app=ceph-osd) on node' \
  --convertor-labels "workload=migration,io-intensive=true"

# Prefer nodes with fast local storage
kubectl mtv create plan nvme-migration \
  --source-provider vmware \
  --vms vm1,vm2,vm3 \
  --convertor-node-selector "disk=nvme,storage-tier=fast" \
  --convertor-affinity 'PREFER pods(app=storage) on zone'
```

### Resource Isolation

Avoid interfering with production workloads:

```bash
# Avoid nodes with CPU-intensive workloads
kubectl mtv create plan isolated-migration \
  --source-provider vmware \
  --vms test-vm1,test-vm2 \
  --convertor-affinity 'AVOID pods(cpu-intensive=true) on node' \
  --convertor-labels "migration-type=non-prod"

# Isolate convertor pods from critical applications
kubectl mtv create plan production-migration \
  --source-provider vmware \
  --vms prod-app1,prod-app2 \
  --convertor-affinity 'AVOID pods(tier=critical) on node' \
  --convertor-node-selector "workload-type=migration"
```

### Load Distribution

Distribute conversion load across multiple nodes/zones:

```bash
# Spread convertor pods across availability zones
kubectl mtv create plan distributed-conversion \
  --source-provider vmware \
  --vms vm1,vm2,vm3,vm4,vm5 \
  --convertor-affinity 'REPEL pods(app=virt-v2v) on zone' \
  --convertor-labels "distribution=spread,batch=large"

# Avoid overloading single nodes with multiple conversions
kubectl mtv create plan balanced-migration \
  --source-provider vmware \
  --vms app1,app2,app3 \
  --convertor-affinity 'REPEL pods(workload=migration) on node'
```

### Network Proximity

Optimize network access to source infrastructure:

```bash
# Place convertor pods close to VMware infrastructure
kubectl mtv create plan vmware-optimized \
  --source-provider vmware \
  --vms vmware-vm1,vmware-vm2 \
  --convertor-node-selector "network-zone=vmware,bandwidth=high" \
  --convertor-affinity 'PREFER pods(app=vmware-proxy) on zone'
```

## Advanced Configuration Examples

### Multi-Tier Optimization

```bash
# Comprehensive convertor optimization for large migration
kubectl mtv create plan enterprise-migration \
  --source-provider vmware \
  --vms @large-vm-list.yaml \
  --convertor-labels "migration=enterprise,priority=high,batch=1" \
  --convertor-node-selector "node-type=migration,cpu=high,memory=high" \
  --convertor-affinity 'REQUIRE pods(app=ceph) on zone' \
  --migration-type warm
```

### Cost-Optimized Migration

```bash
# Use spot/preemptible nodes for cost savings
kubectl mtv create plan cost-optimized \
  --source-provider vmware \
  --vms non-critical-vm1,non-critical-vm2 \
  --convertor-node-selector "node-lifecycle=spot,cost-tier=low" \
  --convertor-labels "cost-optimization=true,priority=low" \
  --convertor-affinity 'AVOID pods(tier=critical) on node'
```

### Performance Monitoring

```bash
# Label convertor pods for monitoring and alerting
kubectl mtv create plan monitored-migration \
  --source-provider vmware \
  --vms critical-db,critical-app \
  --convertor-labels "monitoring=enabled,alert-level=high,migration-id=prod-001" \
  --convertor-node-selector "monitoring=enabled" \
  --convertor-affinity 'PREFER pods(app=prometheus) on zone'
```

## Patching Convertor Configuration

Update existing plans with convertor optimization:

```bash
# Add convertor optimization to existing plan
kubectl mtv patch plan existing-migration \
  --convertor-affinity 'REQUIRE pods(app=storage) on node' \
  --convertor-labels "performance=high,workload=conversion" \
  --convertor-node-selector "disk=ssd"

# Optimize for high-performance storage access
kubectl mtv patch plan slow-migration \
  --convertor-node-selector "storage-tier=premium,disk=nvme" \
  --convertor-affinity 'REQUIRE pods(app=ceph-osd) on node'

# Isolate from production workloads
kubectl mtv patch plan production-migration \
  --convertor-affinity 'AVOID pods(tier=production) on node' \
  --convertor-labels "isolation=required"
```

## Resource Sizing Considerations

### CPU and Memory Requirements

Convertor pods are resource-intensive and benefit from:
- **High CPU**: Disk conversion and compression are CPU-intensive
- **High Memory**: Large disk images require significant memory for buffering
- **Fast Storage**: Local NVMe/SSD storage improves conversion performance

```bash
# Target high-performance nodes for resource-intensive conversions
kubectl mtv create plan resource-intensive \
  --source-provider vmware \
  --vms large-database,large-application \
  --convertor-node-selector "cpu-cores=16+,memory=64Gi+,storage=nvme" \
  --convertor-labels "resource-class=high,workload=intensive"
```

### I/O Optimization

```bash
# Co-locate with distributed storage for optimal I/O patterns
kubectl mtv create plan io-optimized \
  --source-provider vmware \
  --vms storage-heavy-vm1,storage-heavy-vm2 \
  --convertor-affinity 'REQUIRE pods(app=rook-ceph) on node' \
  --convertor-node-selector "storage-bandwidth=high"
```

## Best Practices

### 1. Separation of Concerns

**Different goals require different strategies:**

```bash
# Target affinity: Long-term VM placement
--target-affinity 'REQUIRE pods(app=database) on node'         # Production DB co-location
--target-labels "env=production,tier=database"                 # Production labels
--target-node-selector "node-type=database,ssd=true"          # Production nodes

# Convertor affinity: Temporary migration optimization  
--convertor-affinity 'REQUIRE pods(app=ceph-osd) on node'     # Storage I/O optimization
--convertor-labels "workload=migration,batch=1"               # Migration tracking
--convertor-node-selector "migration-worker=true"             # Migration nodes
```

### 2. Resource Planning

```bash
# Plan for conversion resource requirements
kubectl mtv create plan planned-migration \
  --source-provider vmware \
  --vms vm1,vm2,vm3 \
  --convertor-node-selector "cpu=high,memory=high,temporary-workload=true" \
  --convertor-labels "resource-profile=high,duration=temporary"
```

### 3. Network Topology Awareness

```bash
# Consider network proximity to source infrastructure
kubectl mtv create plan network-aware \
  --source-provider vmware \
  --vms vmware-vm1,vmware-vm2 \
  --convertor-node-selector "network-zone=datacenter-a,vmware-access=direct" \
  --convertor-affinity 'PREFER pods(app=vmware-tools) on zone'
```

### 4. Monitoring and Observability

```bash
# Enable proper labeling for monitoring
kubectl mtv create plan observable-migration \
  --source-provider vmware \
  --vms prod-vm1,prod-vm2 \
  --convertor-labels "migration-id=M-2024-001,environment=production,monitoring=required" \
  --convertor-node-selector "observability=enabled"
```

## Troubleshooting

### Common Issues

#### Convertor Pods Stuck in Pending State
```bash
# Check node selector constraints
kubectl describe pods -l forklift.app/plan=my-plan

# Verify node availability
kubectl get nodes -l migration-worker=true

# Check resource availability
kubectl describe nodes -l storage-tier=fast
```

#### Poor Migration Performance
```bash
# Check if convertor pods are co-located with storage
kubectl get pods -l forklift.app/plan=my-plan -o wide

# Verify storage node placement
kubectl get pods -l app=ceph-osd -o wide

# Update plan for better I/O access
kubectl mtv patch plan slow-migration \
  --convertor-affinity 'REQUIRE pods(app=ceph-osd) on node'
```

#### Resource Contention
```bash
# Move convertor pods away from production workloads
kubectl mtv patch plan resource-conflict \
  --convertor-affinity 'AVOID pods(tier=production) on node' \
  --convertor-node-selector "workload-type=migration"
```

## Related Documentation

- [Target VM Affinity Guide](README_target_affinity.md) - Long-term VM placement
- [Main Usage Guide](README-usage.md) - Complete command reference
- [Plan Patching Guide](README_patch_plans.md) - Updating existing plans
- [Performance Optimization](README_vddk.md) - VDDK and storage optimization

---

**Remember**: Convertor pod settings are for **temporary migration optimization**, while target settings are for **long-term VM operational placement**. Plan accordingly for both phases of your migration lifecycle.
