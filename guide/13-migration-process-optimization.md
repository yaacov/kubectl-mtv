---
layout: page
title: "Chapter 13: Migration Process Optimization (Convertor Pod Scheduling)"
---

Migration process optimization focuses on optimizing the temporary infrastructure used during VM conversion and migration, particularly the virt-v2v convertor pods. This chapter covers performance tuning, resource management, and strategic placement of migration workloads.

## Overview: Optimizing Temporary virt-v2v Convertor Pods

### Understanding Convertor Pods

Convertor pods are temporary Kubernetes pods that run during migration to:

- **Convert VM Formats**: Transform source VM disk formats to target-compatible formats
- **Guest OS Conversion**: Modify guest operating systems for KubeVirt compatibility
- **Driver Installation**: Install virtio drivers and remove hardware-specific drivers
- **Data Transfer**: Handle disk data movement from source to target storage

### Distinction from Target VM Configuration

| Aspect | Migration Process Optimization | Target VM Configuration |
|--------|------------------------------|-------------------------|
| **Purpose** | Optimize temporary migration infrastructure | Configure permanent VM operational settings |
| **Lifespan** | During migration operation only | Permanent VM lifecycle |
| **Resources** | Convertor pods, migration jobs | Target VM pods, operational resources |
| **Focus** | Performance, efficiency, resource isolation | Availability, placement, operational requirements |

## Convertor Configuration Flags

All convertor flags are verified from the kubectl-mtv command code:

### Basic Convertor Configuration

```bash
# Convertor pod labels for identification and management
--convertor-labels "key1=value1,key2=value2"

# Convertor node selector for infrastructure targeting
--convertor-node-selector "key1=value1,key2=value2"

# Convertor affinity using KARL syntax
--convertor-affinity "REQUIRE pods(app=storage) on node"
```

## Why Optimize Convertor Placement?

### Performance Benefits

1. **Storage Access Optimization**: Place convertors near high-performance storage
2. **Network Proximity**: Reduce network latency for data transfers
3. **Resource Dedication**: Use nodes optimized for conversion workloads
4. **I/O Optimization**: Leverage nodes with fast local storage for temporary files

### Resource Management

1. **Isolation**: Separate migration workloads from production systems
2. **Predictability**: Ensure consistent performance during migrations
3. **Scaling**: Use dedicated resources that can be scaled for large migrations
4. **Cost Control**: Use cost-optimized nodes for temporary workloads

### Network Proximity

1. **Bandwidth Optimization**: Place convertors near network-attached storage
2. **Latency Reduction**: Minimize data transfer distances
3. **Traffic Management**: Isolate migration traffic from production networks
4. **Throughput Maximization**: Use high-bandwidth network paths

## Convertor Labels Configuration

Labels help identify, manage, and monitor convertor pods:

### Basic Convertor Labels

```bash
# Migration identification labels
kubectl mtv create plan labeled-conversion \
  --source vsphere-prod \
  --convertor-labels "migration=production,batch=phase1" \
  --vms "web-01,web-02"

# Resource tracking labels
kubectl mtv create plan resource-tracking \
  --source vsphere-prod \
  --convertor-labels "cost-center=migration,project=datacenter-exit,team=infrastructure" \
  --vms "where cluster.name = 'Legacy-Cluster'"
```

### Performance and Monitoring Labels

```bash
# Performance tier labeling
kubectl mtv create plan performance-conversion \
  --source vsphere-prod \
  --convertor-labels "performance=high,priority=urgent,sla=4hour" \
  --vms "critical-database-01,payment-processor"

# Monitoring and alerting labels
kubectl mtv create plan monitored-migration \
  --source vsphere-prod \
  --convertor-labels "monitoring=enabled,alerts=critical,dashboard=migration" \
  --vms "production-workloads"
```

## Convertor Node Selector Configuration

Node selectors target convertors to appropriate infrastructure:

### Storage-Optimized Placement

```bash
# High-performance storage nodes
kubectl mtv create plan storage-optimized \
  --source vsphere-prod \
  --convertor-node-selector "storage=high-performance,io=dedicated" \
  --vms "large-database-01,file-server-02"

# NVMe storage for conversion acceleration
kubectl mtv create plan nvme-conversion \
  --source vsphere-prod \
  --convertor-node-selector "storage=nvme,conversion=optimized" \
  --vms "where sum disks.capacityGB > 500"

# Local SSD for temporary conversion files
kubectl mtv create plan local-ssd \
  --source vsphere-prod \
  --convertor-node-selector "local-storage=ssd,ephemeral=fast" \
  --vms "high-iops-vm-01,database-vm"
```

### CPU and Memory Optimized Placement

```bash
# High-CPU nodes for conversion intensive workloads
kubectl mtv create plan cpu-intensive \
  --source vsphere-prod \
  --convertor-node-selector "cpu=high-performance,cores=many" \
  --vms "windows-vm-01,complex-os-vm"

# High-memory nodes for large VM conversion
kubectl mtv create plan memory-intensive \
  --source vsphere-prod \
  --convertor-node-selector "memory=high,ram=dedicated" \
  --vms "where memoryMB > 32768"

# Balanced resources for standard conversion
kubectl mtv create plan balanced-resources \
  --source vsphere-prod \
  --convertor-node-selector "instance-type=balanced,workload=conversion" \
  --vms "standard-vm-01,web-server-02"
```

### Network-Optimized Placement

```bash
# High-bandwidth network nodes
kubectl mtv create plan network-optimized \
  --source vsphere-prod \
  --convertor-node-selector "network=25gb,bandwidth=high" \
  --vms "network-intensive-app"

# Storage network proximity
kubectl mtv create plan storage-network \
  --source vsphere-prod \
  --convertor-node-selector "network-zone=storage,storage-fabric=connected" \
  --vms "storage-heavy-vm-01"
```

## Convertor Affinity with KARL Syntax

KARL affinity rules provide advanced convertor placement control using the same syntax as target VM affinity:

### Storage Co-location Patterns

#### Co-locate with Storage Controllers

```bash
# Require convertors near storage controllers
kubectl mtv create plan storage-colocation \
  --source vsphere-prod \
  --convertor-affinity "REQUIRE pods(app=storage-controller) on node" \
  --vms "storage-intensive-vm"

# Prefer convertors near Ceph OSD pods
kubectl mtv create plan ceph-proximity \
  --source vsphere-prod \
  --convertor-affinity "PREFER pods(app=ceph-osd) on node weight=90" \
  --vms "large-vm-01,database-vm-02"
```

#### Storage Array Integration

```bash
# Co-locate with storage array controllers
kubectl mtv create plan array-colocation \
  --source vsphere-prod \
  --convertor-affinity "REQUIRE pods(storage-array=flashsystem) on node" \
  --vms "performance-database-01"

# Prefer proximity to NVMe controllers
kubectl mtv create plan nvme-proximity \
  --source vsphere-prod \
  --convertor-affinity "PREFER pods(controller=nvme) on zone weight=85" \
  --vms "high-performance-workload"
```

### Network Proximity Optimization

#### Data Transfer Optimization

```bash
# Co-locate with network-attached storage
kubectl mtv create plan nas-proximity \
  --source vsphere-prod \
  --convertor-affinity "REQUIRE pods(storage=nas) on zone" \
  --vms "file-heavy-vm-01"

# Prefer proximity to backup infrastructure
kubectl mtv create plan backup-proximity \
  --source vsphere-prod \
  --convertor-affinity "PREFER pods(app=backup-controller) on node weight=80" \
  --vms "backup-target-vm"
```

### Resource Isolation Patterns

#### Avoid Production Workloads

```bash
# Avoid running convertors near production databases
kubectl mtv create plan avoid-production-db \
  --source vsphere-prod \
  --convertor-affinity "AVOID pods(tier=database,environment=production) on node" \
  --vms "test-migration-vm"

# Isolate from CPU-intensive production workloads
kubectl mtv create plan avoid-cpu-intensive \
  --source vsphere-prod \
  --convertor-affinity "AVOID pods(workload=cpu-intensive) on zone" \
  --vms "latency-sensitive-conversion"
```

#### Dedicated Migration Infrastructure

```bash
# Require dedicated migration nodes
kubectl mtv create plan dedicated-migration \
  --source vsphere-prod \
  --convertor-affinity "REQUIRE pods(purpose=migration) on node" \
  --convertor-node-selector "dedicated=migration" \
  --vms "batch-migration-vms"

# Prefer migration-optimized infrastructure
kubectl mtv create plan migration-optimized \
  --source vsphere-prod \
  --convertor-affinity "PREFER pods(optimized=migration) on zone weight=95" \
  --vms "large-scale-migration"
```

## Common Use Cases for Convertor Optimization

### Use Case 1: High-Performance Storage Access

#### NVMe Storage Optimization

```bash
# Maximize NVMe storage access for conversion
kubectl mtv create plan nvme-optimized \
  --source vsphere-prod \
  --convertor-node-selector "storage=nvme,performance=extreme" \
  --convertor-affinity "REQUIRE pods(storage-controller=nvme) on node" \
  --convertor-labels "performance=extreme,storage=nvme" \
  --migration-type warm \
  --vms "where sum disks.capacityGB > 1000"

# Results: Convertors use fastest available storage for temporary files
```

#### Storage Array Direct Access

```bash
# Direct access to storage arrays for conversion
kubectl mtv create plan array-direct \
  --source vsphere-prod \
  --convertor-node-selector "storage-fabric=connected,array-access=direct" \
  --convertor-affinity "REQUIRE pods(storage-array=flashsystem) on zone" \
  --convertor-labels "storage-access=direct,array=flashsystem" \
  --vms "enterprise-database-cluster"
```

### Use Case 2: Resource Isolation

#### Dedicated Conversion Infrastructure

```bash
# Isolate conversion workloads on dedicated nodes
kubectl mtv create plan isolated-conversion \
  --source vsphere-prod \
  --convertor-node-selector "workload=migration-only,isolation=dedicated" \
  --convertor-affinity "AVOID pods(environment=production) on node" \
  --convertor-labels "isolation=dedicated,workload=conversion" \
  --vms "production-migration-batch"
```

#### Development vs Production Isolation

```bash
# Separate development migrations from production
kubectl mtv create plan dev-isolation \
  --source vsphere-dev \
  --convertor-node-selector "environment=development,cost=optimized" \
  --convertor-affinity "AVOID pods(environment=production) on zone" \
  --convertor-labels "environment=dev,cost-tier=low" \
  --vms "dev-environment-vms" \
  -n development
```

### Use Case 3: Network and Bandwidth Optimization

#### High-Bandwidth Data Transfer

```bash
# Optimize for high-bandwidth data transfers
kubectl mtv create plan bandwidth-optimized \
  --source vsphere-prod \
  --convertor-node-selector "network=100gb,bandwidth=dedicated" \
  --convertor-affinity "PREFER pods(network-controller=high-bandwidth) on zone weight=90" \
  --convertor-labels "bandwidth=high,network=dedicated" \
  --vms "large-vm-migration"
```

#### Storage Network Proximity

```bash
# Place convertors near storage network infrastructure
kubectl mtv create plan storage-network \
  --source vsphere-prod \
  --convertor-node-selector "network-zone=storage,fabric=infiniband" \
  --convertor-affinity "REQUIRE pods(network=storage-fabric) on zone" \
  --convertor-labels "network=storage,fabric=infiniband" \
  --vms "storage-intensive-workloads"
```

### Use Case 4: Cost Optimization

#### Spot Instance Utilization

```bash
# Use cost-optimized nodes for non-critical migrations
kubectl mtv create plan cost-optimized \
  --source vsphere-dev \
  --convertor-node-selector "instance-type=spot,cost=optimized" \
  --convertor-labels "cost-tier=spot,priority=low" \
  --migration-type cold \
  --vms "non-critical-development"
```

#### Resource Right-Sizing

```bash
# Right-size convertor resources based on workload
kubectl mtv create plan rightsized-conversion \
  --source vsphere-prod \
  --convertor-node-selector "sizing=optimal,efficiency=high" \
  --convertor-labels "efficiency=optimized,sizing=matched" \
  --vms "where memoryMB between 4096 and 16384"
```

## Resource Sizing Considerations

### CPU Requirements

#### CPU-Intensive Conversion Scenarios

```bash
# Windows OS conversion (driver removal/installation)
kubectl mtv create plan windows-conversion \
  --source vsphere-prod \
  --convertor-node-selector "cpu=high,cores=8plus" \
  --convertor-labels "os=windows,cpu-usage=high" \
  --vms "where guestOS ~= '.*windows.*'"

# Multi-disk VM conversion
kubectl mtv create plan multi-disk-conversion \
  --source vsphere-prod \
  --convertor-node-selector "cpu=performance,parallel=supported" \
  --convertor-labels "disk-count=multiple,cpu=parallel" \
  --vms "where len disks > 4"
```

### Memory Requirements

#### Memory-Intensive Scenarios

```bash
# Large VM conversion requiring substantial memory
kubectl mtv create plan large-vm-conversion \
  --source vsphere-prod \
  --convertor-node-selector "memory=32gb-plus,swap=disabled" \
  --convertor-labels "memory-usage=high,vm-size=large" \
  --vms "where memoryMB > 32768"

# Memory-optimized conversion for database VMs
kubectl mtv create plan database-conversion \
  --source vsphere-prod \
  --convertor-node-selector "memory=high,performance=database" \
  --convertor-labels "workload=database,memory=optimized" \
  --vms "where name ~= '.*database.*' or name ~= '.*db.*'"
```

### I/O Performance Requirements

#### High-I/O Conversion Workloads

```bash
# High-IOPS conversion for database and storage VMs
kubectl mtv create plan high-iops-conversion \
  --source vsphere-prod \
  --convertor-node-selector "iops=high,storage=dedicated" \
  --convertor-affinity "REQUIRE pods(storage=high-iops) on node" \
  --convertor-labels "iops=high,storage=performance" \
  --vms "where any disks.iops > 10000"

# Sequential I/O optimization for large files
kubectl mtv create plan sequential-io \
  --source vsphere-prod \
  --convertor-node-selector "io-pattern=sequential,throughput=high" \
  --convertor-labels "io=sequential,throughput=optimized" \
  --vms "where any disks.capacityGB > 500"
```

## Advanced Convertor Optimization Scenarios

### Scenario 1: Large-Scale Enterprise Migration

```bash
# Phase 1: Infrastructure preparation with dedicated conversion nodes
kubectl mtv create plan enterprise-phase1 \
  --source vsphere-production \
  --convertor-node-selector "migration=dedicated,performance=extreme,storage=nvme" \
  --convertor-affinity "REQUIRE pods(infrastructure=migration) on zone" \
  --convertor-labels "migration=enterprise,phase=1,performance=critical" \
  --migration-type warm \
  --vms "where cluster.name = 'Critical-Production' and powerState = 'poweredOn'"

# Phase 2: Batch processing with resource isolation
kubectl mtv create plan enterprise-phase2 \
  --source vsphere-production \
  --convertor-node-selector "batch-processing=enabled,isolation=complete" \
  --convertor-affinity "AVOID pods(migration=enterprise,phase=1) on node" \
  --convertor-labels "migration=enterprise,phase=2,workload=batch" \
  --migration-type cold \
  --vms "where cluster.name = 'Standard-Production'"
```

### Scenario 2: Multi-Cloud Storage Integration

```bash
# Cloud storage optimization for hybrid migrations
kubectl mtv create plan cloud-storage-optimized \
  --source vsphere-prod \
  --convertor-node-selector "cloud-connectivity=optimized,bandwidth=unlimited" \
  --convertor-affinity "PREFER pods(storage=cloud-gateway) on zone weight=95" \
  --convertor-labels "storage=cloud,connectivity=hybrid,transfer=optimized" \
  --vms "cloud-migration-candidates"

# Edge location optimization
kubectl mtv create plan edge-optimized \
  --source vsphere-edge \
  --convertor-node-selector "location=edge,connectivity=cellular" \
  --convertor-affinity "REQUIRE pods(edge-gateway=true) on zone" \
  --convertor-labels "location=edge,connectivity=limited" \
  --migration-type cold \
  --vms "edge-workloads"
```

### Scenario 3: Security and Compliance

```bash
# Secure conversion with air-gapped networks
kubectl mtv create plan secure-conversion \
  --source vsphere-secure \
  --convertor-node-selector "security=classified,network=airgapped" \
  --convertor-affinity "REQUIRE pods(security=classified) on node" \
  --convertor-labels "security=classified,compliance=required,network=isolated" \
  --vms "classified-workloads"

# FIPS compliance conversion
kubectl mtv create plan fips-compliant \
  --source vsphere-gov \
  --convertor-node-selector "fips=enabled,compliance=gov" \
  --convertor-affinity "REQUIRE pods(fips=true) on zone" \
  --convertor-labels "compliance=fips,government=true,security=maximum" \
  --vms "government-systems"
```

### Scenario 4: Performance Benchmarking

```bash
# Baseline performance measurement
kubectl mtv create plan performance-baseline \
  --source vsphere-test \
  --convertor-node-selector "monitoring=enabled,benchmarking=true" \
  --convertor-labels "benchmark=baseline,monitoring=detailed,metrics=enabled" \
  --vms "benchmark-vm-set"

# Optimized performance comparison
kubectl mtv create plan performance-optimized \
  --source vsphere-test \
  --convertor-node-selector "performance=maximum,optimization=enabled" \
  --convertor-affinity "REQUIRE pods(storage=fastest) on node" \
  --convertor-labels "benchmark=optimized,performance=maximum,comparison=enabled" \
  --vms "benchmark-vm-set"
```

## Monitoring and Performance Validation

### Convertor Performance Monitoring

```bash
# Monitor convertor pod resource usage
kubectl top pods -l migration=enterprise --containers

# Check convertor pod placement
kubectl get pods -o wide -l workload=conversion

# Monitor storage I/O performance
kubectl exec -it convertor-pod -- iostat -x 1

# Network bandwidth monitoring
kubectl exec -it convertor-pod -- iftop -i eth0
```

### Performance Metrics Collection

```bash
# Collect conversion performance metrics
kubectl logs convertor-pod | grep "conversion.*complete\|performance\|throughput"

# Monitor node resource utilization during conversion
kubectl top nodes | grep conversion-node

# Check storage performance metrics
kubectl get pvc -l conversion=active -o yaml | grep -A5 "capacity\|usage"
```

### Optimization Validation

```bash
# Verify convertor placement meets affinity rules
kubectl describe pod convertor-pod | grep -A10 "Node-Selectors\|Tolerations"

# Check co-location with storage controllers
kubectl get pods -o wide | grep -E "(convertor|storage-controller)" | sort -k7

# Validate network proximity
kubectl get pods -o wide | grep -E "(convertor|network.*controller)" | sort -k7
```

## Troubleshooting Convertor Optimization

### Common Convertor Issues

#### Resource Constraints

```bash
# Check for resource-constrained nodes
kubectl describe nodes | grep -A5 -B5 "OutOf\|Pressure"

# Monitor convertor pod resource requests vs limits
kubectl describe pod convertor-pod | grep -A10 "Requests\|Limits"

# Check for pending convertor pods
kubectl get pods | grep Pending | grep convertor
```

#### Affinity Rule Conflicts

```bash
# Check for conflicting affinity rules
kubectl describe pod convertor-pod | grep -A20 "Events.*FailedScheduling"

# Validate KARL rule syntax in convertor affinity
kubectl logs -n konveyor-forklift deployment/forklift-controller | grep "karl\|affinity"

# Check for unavailable target pods in affinity rules
kubectl get pods -l "app=storage-controller" # Verify target pods exist
```

#### Performance Issues

```bash
# Monitor convertor I/O performance
kubectl exec convertor-pod -- iostat -x 1 10

# Check network throughput
kubectl exec convertor-pod -- iperf3 -c storage-server -t 30

# Monitor CPU and memory usage
kubectl exec convertor-pod -- top -n 1
```

### Debug Convertor Configuration

```bash
# Enable verbose logging for convertor scheduling
kubectl mtv create plan debug-convertor \
  --convertor-node-selector "debug=enabled" \
  --convertor-affinity "REQUIRE pods(debug=true) on node" \
  -v=2

# Check generated convertor pod specifications
kubectl get pod convertor-pod -o yaml | grep -A30 spec

# Monitor convertor events
kubectl get events --sort-by='.metadata.creationTimestamp' | grep convertor
```

## Integration with Migration Planning

### Combined Target and Convertor Optimization

```bash
# Optimize both target VMs and conversion process
kubectl mtv create plan comprehensive-optimization \
  --source vsphere-prod \
  --target-node-selector "production=true,performance=high" \
  --target-affinity "AVOID pods(app=web) on node" \
  --target-labels "environment=production,tier=database" \
  --convertor-node-selector "conversion=optimized,storage=fast" \
  --convertor-affinity "REQUIRE pods(storage=high-performance) on node" \
  --convertor-labels "conversion=optimized,storage=fast" \
  --vms "critical-database-01"
```

### Migration Type Considerations

```bash
# Warm migration with conversion optimization
kubectl mtv create plan warm-optimized \
  --source vsphere-prod \
  --migration-type warm \
  --convertor-node-selector "storage=nvme,memory=high" \
  --convertor-affinity "PREFER pods(storage=nvme) on zone weight=90" \
  --vms "large-vm-warm-migration"

# Cold migration with cost optimization
kubectl mtv create plan cold-cost-optimized \
  --source vsphere-prod \
  --migration-type cold \
  --convertor-node-selector "cost=optimized,instance-type=spot" \
  --convertor-labels "cost=optimized,priority=low" \
  --vms "non-critical-batch"
```

## Best Practices for Convertor Optimization

### Performance Best Practices

1. **Storage Co-location**: Place convertors near high-performance storage
2. **Resource Right-Sizing**: Match convertor resources to workload requirements  
3. **Network Optimization**: Ensure high-bandwidth connectivity for data transfers
4. **I/O Optimization**: Use local fast storage for temporary conversion files

### Resource Management Best Practices

1. **Isolation Strategy**: Separate conversion workloads from production systems
2. **Capacity Planning**: Ensure adequate resources for peak conversion loads
3. **Monitoring**: Implement comprehensive monitoring of conversion performance
4. **Cost Control**: Use cost-optimized nodes for non-critical conversions

### Operational Best Practices

1. **Staging**: Test convertor optimization in non-production environments
2. **Gradual Rollout**: Start with small batches when implementing optimization
3. **Performance Baselining**: Establish performance baselines before optimization
4. **Documentation**: Document optimization patterns and their performance impact

## Next Steps

After mastering migration process optimization:

1. **Implement Hooks**: Add custom automation in [Chapter 14: Migration Hooks](14-migration-hooks)
2. **Advanced Plan Management**: Learn plan patching in [Chapter 15: Advanced Plan Patching](15-advanced-plan-patching)
3. **Execute Migrations**: Manage plan lifecycle in [Chapter 16: Plan Lifecycle Execution](16-plan-lifecycle-execution)
4. **Troubleshooting**: Master debugging in [Chapter 17: Debugging and Troubleshooting](17-debugging-and-troubleshooting)

---

*Previous: [Chapter 12: Target VM Placement](/guide/12-target-vm-placement)*  
*Next: [Chapter 14: Migration Hooks](/guide/14-migration-hooks)*
