---
layout: page
title: "Chapter 12: Target VM Placement (Operational Lifetime)"
---

# Chapter 12: Target VM Placement (Operational Lifetime)

Target VM placement controls where and how VMs run in the target Kubernetes cluster after migration. This chapter covers comprehensive VM scheduling, affinity rules, and operational configuration for the target environment.

## Overview of Target VM Placement

### Distinction: Target VM Configuration vs. Migration Process Optimization

kubectl-mtv provides two distinct categories of placement control:

#### Target VM Configuration (This Chapter)
- **Purpose**: Control where VMs run operationally after migration
- **Lifespan**: Permanent settings that affect the VM throughout its lifecycle
- **Scope**: Target VM pods, their scheduling, and long-term placement
- **Examples**: Production workload distribution, availability requirements

#### Migration Process Optimization (Chapter 13)
- **Purpose**: Optimize temporary migration infrastructure during the process
- **Lifespan**: Only during the migration operation
- **Scope**: Convertor pods, temporary resources, migration performance
- **Examples**: Storage access optimization, resource isolation during migration

## Target VM Configuration Flags

All target VM configuration flags are verified from the kubectl-mtv command code:

### Basic Target Configuration

```bash
# Target VM labels for identification and management
--target-labels "key1=value1,key2=value2"

# Target node selector for basic scheduling constraints
--target-node-selector "zone=east,storage=ssd"

# Target power state control after migration
--target-power-state on|off|auto

# Target affinity using KARL (Kubernetes Affinity Rule Language)
--target-affinity "REQUIRE pods(app=database) on node"
```

## Target Labels Configuration

Target labels are applied to VM pods for identification, management, and policy enforcement:

### Basic Label Usage

```bash
# Single environment label
kubectl mtv create plan labeled-migration \
  --source vsphere-prod \
  --target-labels "environment=production" \
  --vms "web-01,web-02"

# Multiple labels for comprehensive tagging
kubectl mtv create plan comprehensive-labels \
  --source vsphere-prod \
  --target-labels "environment=production,team=platform,application=web,tier=frontend" \
  --vms "web-server-01,web-server-02"
```

### Label-Based Management Examples

```bash
# Cost center and ownership tracking
kubectl mtv create plan cost-tracking \
  --source vsphere-prod \
  --target-labels "cost-center=engineering,owner=web-team,budget=ops-2024" \
  --vms "where name ~= '^web-.*'"

# Compliance and security labeling
kubectl mtv create plan compliance-labels \
  --source vsphere-prod \
  --target-labels "security-level=restricted,compliance=pci-dss,data-class=sensitive" \
  --vms "payment-processor-01,user-data-vm"

# Application lifecycle management
kubectl mtv create plan lifecycle-labels \
  --source vsphere-prod \
  --target-labels "lifecycle=production,support-level=24x7,backup=required" \
  --vms "critical-app-01,database-primary"
```

## Target Node Selector Configuration

Node selectors provide basic scheduling constraints using node labels:

### Infrastructure-Based Selection

```bash
# Zone-based placement
kubectl mtv create plan zone-east \
  --source vsphere-prod \
  --target-node-selector "zone=east" \
  --vms "east-coast-services"

# Storage-specific placement
kubectl mtv create plan ssd-placement \
  --source vsphere-prod \
  --target-node-selector "storage=ssd,performance=high" \
  --vms "database-01,cache-redis-01"

# Hardware-specific requirements
kubectl mtv create plan gpu-workload \
  --source vsphere-prod \
  --target-node-selector "accelerator=gpu,gpu-type=v100" \
  --vms "ml-training-vm,ai-inference-vm"
```

### Combined Infrastructure Constraints

```bash
# Multi-constraint node selection
kubectl mtv create plan specific-hardware \
  --source vsphere-prod \
  --target-node-selector "zone=central,storage=nvme,memory=high,network=25gb" \
  --vms "high-performance-db,analytics-engine"

# Dedicated node pools
kubectl mtv create plan dedicated-pool \
  --source vsphere-prod \
  --target-node-selector "node-pool=database,dedicated=true" \
  --vms "postgres-primary,postgres-replica-01"
```

## Target Power State Control

Control VM power state after migration completion:

### Power State Options

| State | Description | Use Case |
|-------|-------------|----------|
| `on` | Start VM after migration | Production services |
| `off` | Keep VM powered off | Backup/archive VMs |
| `auto` | Match source power state | Default behavior |

### Power State Examples

```bash
# Ensure production VMs start immediately
kubectl mtv create plan production-online \
  --source vsphere-prod \
  --target-power-state on \
  --vms "web-service-01,api-gateway-01,database-01"

# Keep backup VMs offline initially
kubectl mtv create plan backup-offline \
  --source vsphere-prod \
  --target-power-state off \
  --vms "backup-vm-01,archive-vm-02"

# Preserve original power states
kubectl mtv create plan preserve-state \
  --source vsphere-prod \
  --target-power-state auto \
  --vms "where powerState in ('poweredOn', 'poweredOff')"
```

## Target Affinity with KARL Syntax

KARL (Kubernetes Affinity Rule Language) provides expressive affinity rules for advanced VM placement.

### KARL Rule Types

Verified from the KARL interpreter vendor code:

| Rule Type | Constraint | Description | Use Case |
|-----------|------------|-------------|----------|
| `REQUIRE` | Hard affinity | Must be scheduled with target | Critical dependencies |
| `PREFER` | Soft affinity | Prefer scheduling with target | Performance optimization |
| `AVOID` | Hard anti-affinity | Must not be scheduled with target | Conflict avoidance |
| `REPEL` | Soft anti-affinity | Prefer not scheduling with target | Load distribution |

### Topology Keys

Available topology domains from the KARL code:

| Topology | Description | Use Case |
|----------|-------------|----------|
| `node` | Same physical node | Co-location for performance |
| `zone` | Same availability zone | Regional co-location |
| `region` | Same geographical region | Compliance/latency requirements |
| `rack` | Same physical rack | Hardware co-location |

### KARL Syntax Structure

```
RULE_TYPE pods(selector) on TOPOLOGY_KEY [weight=N]
```

**Components:**
- `RULE_TYPE`: REQUIRE, PREFER, AVOID, REPEL
- `pods(selector)`: Label selector for target pods
- `TOPOLOGY_KEY`: node, zone, region, rack
- `weight=N`: Optional weight for soft constraints (1-100)

## Detailed KARL Examples

### Co-location Patterns

#### Database Co-location

```bash
# Require application VMs to run on same node as database
kubectl mtv create plan app-with-database \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(app=database) on node" \
  --vms "app-server-01,app-server-02"

# Prefer application VMs near database for performance
kubectl mtv create plan app-near-database \
  --source vsphere-prod \
  --target-affinity "PREFER pods(app=database,tier=primary) on zone weight=80" \
  --vms "web-app-01,api-service-01"
```

#### Storage Co-location

```bash
# Require VMs to run with storage controller pods
kubectl mtv create plan storage-colocation \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(app=storage-controller) on node" \
  --vms "high-iops-vm-01,database-vm-01"

# Prefer VMs near cache pods
kubectl mtv create plan cache-colocation \
  --source vsphere-prod \
  --target-affinity "PREFER pods(app=redis,role=cache) on node weight=90" \
  --vms "cache-client-01,session-service"
```

### Anti-affinity Patterns

#### High Availability Distribution

```bash
# Avoid running web servers on same node
kubectl mtv create plan ha-web-servers \
  --source vsphere-prod \
  --target-affinity "AVOID pods(app=web-server) on node" \
  --vms "web-01,web-02,web-03"

# Soft anti-affinity for load distribution
kubectl mtv create plan distributed-workers \
  --source vsphere-prod \
  --target-affinity "REPEL pods(app=worker) on node weight=70" \
  --vms "worker-01,worker-02,worker-03,worker-04"
```

#### Resource Isolation

```bash
# Avoid running near resource-intensive applications
kubectl mtv create plan avoid-intensive \
  --source vsphere-prod \
  --target-affinity "AVOID pods(resource=intensive) on node" \
  --vms "latency-sensitive-01,real-time-service"

# Avoid CPU-heavy workloads
kubectl mtv create plan avoid-cpu-heavy \
  --source vsphere-prod \
  --target-affinity "AVOID pods(workload=cpu-intensive) on zone" \
  --vms "interactive-app-01,user-facing-service"
```

### Zone and Regional Affinity

#### Availability Zone Management

```bash
# Require VMs in specific zone for compliance
kubectl mtv create plan zone-compliance \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(zone=east) on zone" \
  --vms "compliance-app-01,audit-service"

# Distribute across zones for availability
kubectl mtv create plan multi-zone-ha \
  --source vsphere-prod \
  --target-affinity "AVOID pods(app=frontend) on zone" \
  --vms "frontend-east,frontend-west,frontend-central"
```

#### Regional Data Locality

```bash
# Keep data processing near data sources
kubectl mtv create plan data-locality \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(data-region=us-east) on region" \
  --vms "data-processor-01,analytics-engine"

# Prefer regional processing for performance
kubectl mtv create plan regional-processing \
  --source vsphere-prod \
  --target-affinity "PREFER pods(region=us-west) on region weight=95" \
  --vms "west-coast-analytics,regional-cache"
```

### Complex Multi-Rule Scenarios

#### Database Cluster Placement

```bash
# Primary database: require dedicated nodes
kubectl mtv create plan db-primary \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(node-type=database) on node" \
  --target-node-selector "dedicated=database" \
  --vms "postgres-primary"

# Replica databases: avoid primary, prefer database nodes
kubectl mtv create plan db-replicas \
  --source vsphere-prod \
  --target-affinity "AVOID pods(role=primary) on node" \
  --target-node-selector "node-type=database" \
  --vms "postgres-replica-01,postgres-replica-02"
```

#### Multi-Tier Application Deployment

```bash
# Web tier: distribute across zones, avoid other web servers
kubectl mtv create plan web-tier \
  --source vsphere-prod \
  --target-affinity "AVOID pods(tier=web) on node" \
  --target-labels "tier=web,layer=frontend" \
  --vms "web-server-01,web-server-02,web-server-03"

# App tier: prefer near web tier, avoid intensive workloads
kubectl mtv create plan app-tier \
  --source vsphere-prod \
  --target-affinity "PREFER pods(tier=web) on zone weight=80" \
  --target-labels "tier=app,layer=business" \
  --vms "app-server-01,app-server-02"

# Data tier: require dedicated storage nodes
kubectl mtv create plan data-tier \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(storage=dedicated) on node" \
  --target-node-selector "storage=high-performance" \
  --target-labels "tier=data,layer=persistence" \
  --vms "database-01,cache-01"
```

## Advanced Target Configuration Scenarios

### Scenario 1: Multi-Region Deployment

```bash
# East region deployment
kubectl mtv create plan east-region \
  --source vsphere-east \
  --target-affinity "REQUIRE pods(region=east) on region" \
  --target-node-selector "zone=us-east-1" \
  --target-labels "region=east,availability-zone=us-east-1" \
  --target-power-state on \
  --vms "east-web-01,east-api-01,east-cache-01"

# West region deployment with cross-region anti-affinity
kubectl mtv create plan west-region \
  --source vsphere-west \
  --target-affinity "AVOID pods(region=east) on region" \
  --target-node-selector "zone=us-west-1" \
  --target-labels "region=west,availability-zone=us-west-1" \
  --target-power-state on \
  --vms "west-web-01,west-api-01,west-cache-01"
```

### Scenario 2: Performance-Critical Application

```bash
# High-performance database with strict placement
kubectl mtv create plan performance-db \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(storage=nvme) on node" \
  --target-node-selector "performance=extreme,storage=nvme,cpu=high" \
  --target-labels "performance=critical,sla=tier1,monitoring=intensive" \
  --target-power-state on \
  --vms "trading-db-primary"

# Application servers near performance database
kubectl mtv create plan performance-apps \
  --source vsphere-prod \
  --target-affinity "PREFER pods(app=trading-db) on zone weight=95" \
  --target-node-selector "performance=high,latency=low" \
  --target-labels "performance=high,sla=tier1" \
  --target-power-state on \
  --vms "trading-app-01,trading-app-02"
```

### Scenario 3: Security-Sensitive Workloads

```bash
# Secure workloads on dedicated nodes
kubectl mtv create plan secure-workloads \
  --source vsphere-security \
  --target-affinity "REQUIRE pods(security=dedicated) on node" \
  --target-node-selector "security=restricted,isolation=physical" \
  --target-labels "security-level=restricted,compliance=hipaa,isolation=required" \
  --target-power-state on \
  --vms "patient-data-vm,financial-processor"

# Avoid co-location with non-secure workloads
kubectl mtv create plan security-isolation \
  --source vsphere-security \
  --target-affinity "AVOID pods(security!=restricted) on node" \
  --target-node-selector "security=restricted" \
  --target-labels "security-level=restricted,data-class=sensitive" \
  --vms "encryption-service,key-manager"
```

### Scenario 4: Development Environment Organization

```bash
# Development VMs with relaxed constraints
kubectl mtv create plan dev-environment \
  --source vsphere-dev \
  --target-affinity "PREFER pods(environment=dev) on zone weight=50" \
  --target-node-selector "environment=dev" \
  --target-labels "environment=dev,lifecycle=temporary,cost-optimize=true" \
  --target-power-state off \
  --vms "dev-web-01,dev-api-01,dev-db-01" \
  -n development

# Test VMs isolated from production
kubectl mtv create plan test-isolation \
  --source vsphere-test \
  --target-affinity "AVOID pods(environment=production) on zone" \
  --target-node-selector "environment=test" \
  --target-labels "environment=test,temporary=true" \
  --target-power-state off \
  --vms "test-suite-01,load-test-vm" \
  -n testing
```

## Integration with PlanVMS Format

Target placement settings can be combined with individual VM customization:

### Per-VM Target Configuration

```yaml
# vm-with-placement.yaml
- name: database-primary
  targetName: postgres-primary
  targetPowerState: on
  # VM inherits plan-level affinity and node selector

- name: database-replica
  targetName: postgres-replica-01
  targetPowerState: on
  # Different placement can be achieved through separate plans
```

### Plan-Level vs VM-Level Settings

```bash
# Plan-level settings apply to all VMs
kubectl mtv create plan database-cluster \
  --source vsphere-prod \
  --target-affinity "REQUIRE pods(storage=database) on node" \
  --target-node-selector "dedicated=database" \
  --target-labels "app=postgres,environment=production" \
  --vms @database-vms.yaml

# Individual VMs inherit plan-level target settings
# PlanVMS format currently doesn't override plan-level target affinity
```

## Target Placement Validation

### Pre-Migration Validation

```bash
# Verify target nodes exist with required labels
kubectl get nodes --show-labels | grep "storage=ssd"

# Check node availability and capacity
kubectl describe node node-with-ssd-storage

# Validate existing pod distribution
kubectl get pods -o wide --show-labels | grep "app=database"
```

### Affinity Rule Testing

```bash
# Test KARL rule parsing (requires migration plan creation)
kubectl mtv create plan affinity-test \
  --source vsphere-test \
  --target-affinity "REQUIRE pods(app=test) on node" \
  --vms "test-vm-01" \
  --dry-run=client

# Verify generated affinity rules
kubectl get vm test-vm-01 -o yaml | grep -A10 affinity
```

### Post-Migration Verification

```bash
# Check VM pod placement
kubectl get pods -o wide --show-labels | grep migrated-vm

# Verify affinity constraints are met
kubectl describe pod migrated-vm-pod | grep -A20 "Node-Selectors\|Tolerations"

# Monitor placement compliance
kubectl get pods --field-selector spec.nodeName=target-node
```

## Troubleshooting Target Placement

### Common Placement Issues

#### Unschedulable VMs

```bash
# Check for scheduling issues
kubectl get pods | grep Pending

# Describe pod for scheduling errors
kubectl describe pod unschedulable-vm-pod

# Check node resource availability
kubectl top nodes

# Verify node labels match selectors
kubectl get nodes --show-labels | grep required-label
```

#### Affinity Rule Conflicts

```bash
# Check for conflicting affinity rules
kubectl describe pod conflicted-vm-pod | grep -A10 -B10 "FailedScheduling"

# Validate KARL rule syntax
echo "REQUIRE pods(app=database) on node" | # Test in plan creation

# Check existing pod distribution
kubectl get pods -o wide --selector app=database
```

#### Node Selector Mismatches

```bash
# List available node labels
kubectl get nodes -o yaml | grep -A5 labels:

# Check if required labels exist
kubectl get nodes -l "storage=ssd,zone=east"

# Update node labels if needed
kubectl label node worker-node-01 storage=ssd zone=east
```

### Debug Target Configuration

```bash
# Enable verbose logging for plan creation
kubectl mtv create plan debug-target \
  --target-affinity "REQUIRE pods(app=debug) on node" \
  --target-node-selector "debug=true" \
  -v=2

# Monitor VM creation and placement
kubectl get events --sort-by='.metadata.creationTimestamp' | grep scheduling

# Check generated VM specifications
kubectl get vm debug-vm -o yaml | grep -A20 spec
```

## Performance Impact of Target Placement

### Placement Strategy Performance Considerations

#### Network Locality
```bash
# Co-locate network-intensive applications
kubectl mtv create plan network-locality \
  --target-affinity "REQUIRE pods(app=network-storage) on node" \
  --vms "high-bandwidth-app"
```

#### Storage Locality
```bash
# Place VMs near storage controllers
kubectl mtv create plan storage-locality \
  --target-affinity "PREFER pods(app=ceph-osd) on node weight=90" \
  --vms "storage-intensive-vm"
```

#### CPU/Memory Optimization
```bash
# Distribute CPU-intensive workloads
kubectl mtv create plan cpu-distribution \
  --target-affinity "AVOID pods(cpu-usage=high) on node" \
  --vms "cpu-intensive-01,cpu-intensive-02"
```

## Best Practices for Target VM Placement

### Design Principles

1. **Availability**: Use anti-affinity for high availability
2. **Performance**: Co-locate related services when beneficial
3. **Resource Utilization**: Distribute workloads for optimal resource usage
4. **Compliance**: Ensure placement meets security and regulatory requirements

### Operational Guidelines

1. **Start Simple**: Begin with basic node selectors, add complexity gradually
2. **Test Thoroughly**: Validate placement rules in test environments
3. **Monitor Impact**: Track performance and availability after placement changes
4. **Document Rules**: Maintain clear documentation of placement rationale

### KARL Rule Design Best Practices

1. **Readable Rules**: Use descriptive label selectors in KARL rules
2. **Appropriate Constraints**: Choose REQUIRE/PREFER/AVOID/REPEL based on criticality
3. **Weight Tuning**: Use appropriate weights for soft constraints (50-100)
4. **Topology Alignment**: Match topology keys to infrastructure design

## Next Steps

After mastering target VM placement:

1. **Optimize Migration Process**: Learn convertor optimization in [Chapter 13: Migration Process Optimization](13-migration-process-optimization.md)
2. **Create Hooks**: Develop migration automation in [Chapter 14: Migration Hooks](14-migration-hooks.md)
3. **Advanced Plan Modification**: Learn plan patching in [Chapter 15: Advanced Plan Patching](15-advanced-plan-patching.md)
4. **Execute Migrations**: Manage plan lifecycle in [Chapter 16: Plan Lifecycle Execution](16-plan-lifecycle-execution.md)

---

*Previous: [Chapter 11: Customizing Individual VMs (PlanVMS Format)](11-customizing-individual-vms-planvms-format.md)*  
*Next: [Chapter 13: Migration Process Optimization](13-migration-process-optimization.md)*
