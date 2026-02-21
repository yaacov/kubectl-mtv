---
layout: default
title: "Chapter 10: Query Language Reference and Advanced Filtering"
parent: "III. Inventory and Advanced Query Language"
nav_order: 2
---

kubectl-mtv integrates the powerful [Tree Search Language (TSL)](../27-tsl-tree-search-language-reference), developed by Yaacov Zamir, which provides SQL-like filtering capabilities for inventory resources. This chapter provides a complete reference for TSL syntax and advanced query techniques.

> **Quick Reference**: For a concise, self-contained TSL syntax reference, see [Chapter 27: TSL - Tree Search Language Reference](../27-tsl-tree-search-language-reference).

## Introduction to Tree Search Language (TSL)

### What is TSL?

Tree Search Language (TSL) is a powerful search language with grammar similar to SQL's WHERE clause. TSL enables sophisticated filtering of inventory data using familiar SQL-like syntax.

### TSL Grammar Examples

Basic TSL expressions follow intuitive patterns:

```sql
-- Simple equality
name = 'web-server-01'

-- Logical operations
name = 'vm1' or name = 'vm2'

-- Pattern matching
city in ['paris', 'rome', 'milan'] and state != 'france'

-- Range queries
pages between 100 and 200 and author.name ~= 'Hilbert'
```

### TSL in kubectl-mtv Context

TSL is used throughout kubectl-mtv for inventory filtering:

```bash
# Filter VMs by power state
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState = 'poweredOn'"

# Filter by memory size
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryMB > 8192"

# Complex filtering
kubectl mtv get inventory vms --provider vsphere-prod --query "where name ~= 'prod-.*' and memoryMB >= 4096"
```

## Query Structure

TSL queries in kubectl-mtv follow this general structure:

```
kubectl mtv get inventory <resource> --provider <provider> --query "where <TSL_EXPRESSION>"
```

### Basic Query Components

1. **SELECT Clause** (optional): Select specific fields to return (e.g., `select name, memoryMB`)
2. **WHERE Clause**: The `where` keyword starts the filter expression
3. **Field References**: Access object fields using dot notation (e.g., `name`, `host.cluster.name`)
4. **Operators**: Comparison, logical, and pattern matching operators
5. **Literals**: String, numeric, boolean, and array values
6. **Functions**: Built-in functions for complex operations
7. **ORDER BY**: Sort results by a field ascending or descending
8. **LIMIT**: Restrict the number of results returned

### Selecting Fields with SELECT

Use the optional `select` clause to return only the fields you need, reducing output size:

```bash
# Return only name, memory, and CPU (compact table output)
kubectl mtv get inventory vms --provider vsphere-prod --query "select name, memoryMB, cpuCount where powerState = 'poweredOn' limit 10"

# Select with aliases
kubectl mtv get inventory vms --provider vsphere-prod --query "select name, memoryMB as mem, cpuCount as cpus where memoryMB > 4096"

# Use reducers in select (sum, len)
kubectl mtv get inventory vms --provider vsphere-prod --query "select name, sum(disks[*].capacityGB) as totalDisk where powerState = 'poweredOn' order by totalDisk desc limit 10"
```

### Sorting with ORDER BY

Sort query results by any field in ascending (default) or descending order:

```bash
# Top VMs by memory (descending)
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState = 'poweredOn' order by memoryMB desc"

# VMs alphabetically by name
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryMB > 4096 order by name"

# Ascending is the default, but can be explicit
kubectl mtv get inventory vms --provider vsphere-prod --query "where cpuCount > 2 order by cpuCount asc"
```

### Limiting Results with LIMIT

Restrict the number of results returned:

```bash
# Top 10 largest VMs by memory
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryMB > 1024 order by memoryMB desc limit 10"

# First 50 powered-on VMs alphabetically
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState = 'poweredOn' order by name limit 50"

# Just 5 VMs matching a pattern
kubectl mtv get inventory vms --provider vsphere-prod --query "where name ~= 'prod-.*' limit 5"
```

## Data Types and Literals

TSL supports various data types for matching different kinds of inventory data:

### String Literals

```sql
-- Single quotes for strings
name = 'web-server-01'
guestOS = 'rhel8'

-- Strings with spaces
cluster.name = 'Production Cluster'
```

### Numeric Literals

```sql
-- Integer values
memoryMB = 8192
numCpu = 4

-- Decimal values
diskGB = 100.5
utilizationPercent = 85.7
```

### Boolean Literals

```sql
-- Boolean values
template = true
powerOn = false
```

### Array Literals

```sql
-- Array values for IN operations (use square brackets)
name in ['vm1', 'vm2', 'vm3']
powerState in ['poweredOn', 'suspended']
```

### Date and Timestamp Literals

```sql
-- Date literals (YYYY-MM-DD)
created >= '2024-01-01'

-- Timestamp literals (RFC3339 format)
lastModified >= '2024-01-01T00:00:00Z'
```

### SI Unit Suffixes

TSL supports SI suffixes for expressing byte quantities concisely. The parser expands them to plain numbers at evaluation time:

| Suffix | Multiplier | Example | Expanded Value |
|--------|-----------|---------|----------------|
| `Ki` | 1,024 | `4Ki` | 4096 |
| `Mi` | 1,048,576 | `512Mi` | 536870912 |
| `Gi` | 1,073,741,824 | `4Gi` | 4294967296 |
| `Ti` | 1,099,511,627,776 | `1Ti` | 1099511627776 |
| `Pi` | 1,125,899,906,842,624 | `1Pi` | 1125899906842624 |

SI units are especially useful for oVirt, where memory values are in bytes:

```bash
# oVirt: find VMs with more than 4 GB of memory (memory field is in bytes)
kubectl mtv get inventory vms --provider ovirt-prod --query "where memory > 4Gi"

# oVirt: find VMs with memory between 2 GB and 8 GB
kubectl mtv get inventory vms --provider ovirt-prod --query "where memory between 2Gi and 8Gi"

# Disk capacity using SI units
kubectl mtv get inventory vms --provider vsphere-prod --query "where sum(disks[*].capacity) > 100Gi"
```

### Null Values

```sql
-- Check for null values
description is null
annotation is not null
```

## Operators Reference

TSL provides a comprehensive set of operators verified from the vendor code.

### Arithmetic Operators

| Operator | Description |
|----------|-------------|
| `+` | Addition |
| `-` | Subtraction |
| `*` | Multiplication |
| `/` | Division |
| `%` | Modulo |

Arithmetic operators can be used within expressions:

```bash
# VMs where CPU-to-memory ratio is significant
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryMB / cpuCount > 4096"
```

### Comparison Operators

| Operator | Symbol | Description | Example |
|----------|--------|-------------|---------|
| **EQ** | `=` | Equal to | `name = 'vm1'` |
| **NE** | `!=`, `<>` | Not equal to | `powerState != 'poweredOff'` |
| **LT** | `<` | Less than | `memoryMB < 4096` |
| **LE** | `<=` | Less than or equal | `numCpu <= 2` |
| **GT** | `>` | Greater than | `diskGB > 100` |
| **GE** | `>=` | Greater than or equal | `memoryMB >= 8192` |

```bash
# Examples of comparison operators
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryMB >= 8192"
kubectl mtv get inventory vms --provider vsphere-prod --query "where numCpu < 4"
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState != 'poweredOff'"
```

### String Matching Operators

| Operator | Description | Example |
|----------|-------------|---------|
| **LIKE** | Case-sensitive pattern matching (SQL LIKE) | `name like 'web-%'` |
| **ILIKE** | Case-insensitive pattern matching | `name ilike 'WEB-%'` |
| **REQ** | Regular expression equals (`~=`) | `name ~= 'prod-.*'` |
| **RNE** | Regular expression not equals (`~!`) | `name ~! 'test-.*'` |

```bash
# String matching examples
kubectl mtv get inventory vms --provider vsphere-prod --query "where name like 'prod-%'"
kubectl mtv get inventory vms --provider vsphere-prod --query "where guestOS ilike '%windows%'"
kubectl mtv get inventory vms --provider vsphere-prod --query "where name ~= '^web-[0-9]+$'"
```

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| **AND** | Logical AND | `memoryMB > 4096 and numCpu >= 2` |
| **OR** | Logical OR | `name = 'vm1' or name = 'vm2'` |
| **NOT** | Logical NOT | `not (powerState = 'poweredOff')` |

```bash
# Logical operator examples
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState = 'poweredOn' and memoryMB > 8192"
kubectl mtv get inventory vms --provider vsphere-prod --query "where name = 'web-01' or name = 'web-02'"
kubectl mtv get inventory vms --provider vsphere-prod --query "where not template"
```

### Array and Range Operators

| Operator | Description | Example |
|----------|-------------|---------|
| **IN** | Value in array | `name in ['vm1', 'vm2', 'vm3']` |
| **NOT IN** | Value not in array | `guestId not in ['rhel8_64Guest', '']` |
| **BETWEEN** | Value in inclusive range | `memoryMB between 4096 and 16384` |

Note: The `in` and `not in` operators require **square brackets** for the value list.

```bash
# Array and range examples
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState in ['poweredOn', 'suspended']"
kubectl mtv get inventory vms --provider vsphere-prod --query "where guestId not in ['rhel8_64Guest', '']"
kubectl mtv get inventory vms --provider vsphere-prod --query "where firmware in ['efi', 'bios']"
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryMB between 8192 and 32768"
```

### Null Checking Operators

| Operator | Description | Example |
|----------|-------------|---------|
| **IS** | Check for null/not null | `description is null` |
| **IS NOT** | Check for not null | `annotation is not null` |

```bash
# Null checking examples
kubectl mtv get inventory vms --provider vsphere-prod --query "where description is not null"
kubectl mtv get inventory vms --provider vsphere-prod --query "where guestIP is null"
```

## Functions Reference

TSL provides built-in functions for complex data analysis:

### LEN Function

Returns the length of arrays or strings. The comparison operator is placed **outside** the parentheses:

```sql
-- Array length
len(disks) > 2
len(networks) = 1
len(nics) >= 2
```

```bash
# LEN function examples
kubectl mtv get inventory vms --provider vsphere-prod --query "where len(disks) > 2"
kubectl mtv get inventory vms --provider vsphere-prod --query "where len(nics) >= 2"
```

### ANY Function

Tests if any element in an array matches a condition:

```sql
any(disks[*].capacity > 100Gi)
any(networks[*].name = 'Production Network')
any(concerns[*].category = 'Critical')
any(disks[*].shared = true)
```

```bash
# ANY function examples
kubectl mtv get inventory vms --provider vsphere-prod --query "where any(disks[*].capacityGB > 100)"
kubectl mtv get inventory vms --provider vsphere-prod --query "where any(networks[*].name ~= '.*-prod')"
kubectl mtv get inventory vms --provider vsphere-prod --query "where any(concerns[*].category = 'Critical')"
```

### ALL Function

Tests if all elements in an array match a condition:

```sql
all(disks[*].shared = false)
all(disks[*].capacityGB > 50)
all(networks[*].name ~= '.*-prod')
```

```bash
# ALL function examples
kubectl mtv get inventory vms --provider vsphere-prod --query "where all(disks[*].capacityGB >= 20)"
kubectl mtv get inventory vms --provider vsphere-prod --query "where all(networks[*].type = 'standard')"
```

### SUM Function

Calculates the sum of numeric values in an array:

```sql
-- Total disk capacity
sum(disks[*].capacityGB) > 1000

-- Total CPU count
sum(hosts[*].numCpu) > 50
```

```bash
# SUM function examples
kubectl mtv get inventory vms --provider vsphere-prod --query "where sum(disks[*].capacityGB) > 500"
kubectl mtv get inventory clusters --provider vsphere-prod --query "where sum(hosts[*].memoryGB) > 100000"
```

## Field Access

### Dot Notation

Access nested fields using dot notation:

```bash
# Nested object properties
kubectl mtv get inventory vms --provider vsphere-prod --query "where host.cluster.name = 'Production'"

# Deep nesting
kubectl mtv get inventory vms --provider vsphere-prod --query "where host.datacenter.name = 'DC1'"
```

### Array Index Access

Access a specific element in an array by zero-based index:

```bash
# First disk
kubectl mtv get inventory vms --provider vsphere-prod --query "where disks[0].capacityGB > 100"
```

### Array Wildcard Access

Use `[*]` to reference all elements (used inside `any`, `all`, `sum`):

```bash
kubectl mtv get inventory vms --provider vsphere-prod --query "where any(disks[*].datastore.id = 'datastore-12')"
kubectl mtv get inventory vms --provider vsphere-prod --query "where sum(disks[*].capacity) > 500Gi"
```

### Implicit Traversal

Dot notation on an array field implicitly traverses all elements (equivalent to `[*]`):

```sql
-- These are equivalent:
disks.capacity
disks[*].capacity
```

## VM Fields by Provider

Field names depend on the provider type and are derived from the inventory JSON. To discover all available fields for your environment, run:

```bash
kubectl mtv get inventory vms --provider <provider> --output json
```

Any field visible in the JSON output can be used in queries with dot notation.

### vSphere Fields

| Category | Fields |
|----------|--------|
| Identity | `name`, `id`, `uuid`, `path`, `parent.id`, `parent.kind` |
| State | `powerState`, `connectionState` |
| Compute | `cpuCount`, `coresPerSocket`, `memoryMB` |
| Guest | `guestId`, `guestName`, `firmware`, `isTemplate` |
| Network | `ipAddress`, `hostName`, `host` |
| Storage | `storageUsed` |
| Security | `secureBoot`, `tpmEnabled`, `changeTrackingEnabled` |
| Disks | `disks[*].capacity`, `disks[*].datastore.id`, `disks[*].datastore.name`, `disks[*].file`, `disks[*].shared` |
| NICs | `nics[*].mac`, `nics[*].network.id` |
| Networks | `networks[*].id`, `networks[*].kind` |
| Concerns | `concerns[*].category`, `concerns[*].assessment`, `concerns[*].label` |

### oVirt / RHV Fields

| Category | Fields |
|----------|--------|
| Identity | `name`, `id`, `path`, `cluster`, `host` |
| State | `status` (`up`, `down`, ...) |
| Compute | `cpuSockets`, `cpuCores`, `cpuThreads`, `memory` (bytes -- use SI units like `4Gi`) |
| Guest | `osType`, `guestName`, `guest.distribution`, `guest.fullVersion` |
| Config | `haEnabled`, `stateless`, `placementPolicyAffinity`, `display` |
| Disks | `diskAttachments[*].disk`, `diskAttachments[*].interface` |
| NICs | `nics[*].name`, `nics[*].mac`, `nics[*].interface`, `nics[*].ipAddress`, `nics[*].profile` |
| Concerns | `concerns[*].category`, `concerns[*].assessment`, `concerns[*].label` |

### OpenStack Fields

| Category | Fields |
|----------|--------|
| Identity | `name`, `id`, `status` |
| Resources | `flavor.name`, `image.name`, `project.name` |
| Volumes | `attachedVolumes[*].ID` |

### EC2 Fields (PascalCase)

| Category | Fields |
|----------|--------|
| Identity | `name`, `InstanceType`, `State.Name`, `PlatformDetails` |
| Placement | `Placement.AvailabilityZone` |
| Network | `PublicIpAddress`, `PrivateIpAddress`, `VpcId`, `SubnetId` |

### Computed Fields (All Providers)

kubectl-mtv adds the following computed fields to every VM record. These are calculated from the provider data and are available for all provider types:

| Field | Description |
|-------|-------------|
| `criticalConcerns` | Count of critical migration concerns |
| `warningConcerns` | Count of warning migration concerns |
| `infoConcerns` | Count of informational migration concerns |
| `concernsHuman` | Human-readable concern summary |
| `memoryGB` | Memory converted to GB (from MB or bytes) |
| `storageUsedGB` | Storage used converted to GB |
| `diskCapacity` | Total disk capacity |
| `powerStateHuman` | Human-readable power state |
| `provider` | Provider name |

```bash
# Use computed fields for easy filtering
kubectl mtv get inventory vms --provider vsphere-prod --query "where criticalConcerns > 0"
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryGB > 16"
kubectl mtv get inventory vms --provider vsphere-prod --query "where warningConcerns = 0 and criticalConcerns = 0"
```

## Advanced Query Examples

### Filtering by Power State, Memory, and Name Patterns

#### Power State Filtering

```bash
# Find only powered-on VMs
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState = 'poweredOn'"

# Find VMs that are not powered off
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState != 'poweredOff'"

# Find suspended or powered-on VMs
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState in ['poweredOn', 'suspended']"
```

#### Memory-Based Filtering

```bash
# High-memory VMs (>16GB)
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryMB > 16384"

# Medium memory range (4-16GB)
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryMB between 4096 and 16384"

# Low-memory VMs suitable for quick migration
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryMB <= 4096 and powerState = 'poweredOn'"
```

#### Name Pattern Filtering

```bash
# Production VMs by naming convention
kubectl mtv get inventory vms --provider vsphere-prod --query "where name ~= '^prod-.*'"

# Web servers by pattern
kubectl mtv get inventory vms --provider vsphere-prod --query "where name ~= '^web-[0-9]+$'"

# Exclude test and development VMs
kubectl mtv get inventory vms --provider vsphere-prod --query "where name ~! '.*(test|dev|tmp).*'"

# VMs containing specific keywords
kubectl mtv get inventory vms --provider vsphere-prod --query "where name ilike '%database%' or name ilike '%db%'"
```

### Sorting and Limiting Results

```bash
# Top 10 largest VMs by memory
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryMB > 1024 order by memoryMB desc limit 10"

# First 50 powered-on VMs alphabetically
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState = 'poweredOn' order by name limit 50"

# Smallest VMs first (candidates for quick migration)
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState = 'poweredOn' order by memoryMB asc limit 20"
```

### Migration Concerns Analysis

Use the computed fields `criticalConcerns`, `warningConcerns`, and `infoConcerns` to assess migration readiness:

```bash
# VMs with critical migration concerns
kubectl mtv get inventory vms --provider vsphere-prod --query "where criticalConcerns > 0"

# VMs with any warnings
kubectl mtv get inventory vms --provider vsphere-prod --query "where warningConcerns > 0"

# Clean VMs with no concerns at all (safest to migrate first)
kubectl mtv get inventory vms --provider vsphere-prod --query "where len(concerns) = 0"
kubectl mtv get inventory vms --provider vsphere-prod --query "where criticalConcerns = 0 and warningConcerns = 0"

# VMs with specific concern categories
kubectl mtv get inventory vms --provider vsphere-prod --query "where any(concerns[*].category = 'Critical')"
kubectl mtv get inventory vms --provider vsphere-prod --query "where any(concerns[*].category = 'Warning')"

# Use memoryGB computed field for easier comparison
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryGB > 16 and criticalConcerns = 0"
```

### Complex Multi-Condition Queries

#### Migration Readiness Assessment

```bash
# VMs ready for migration (powered on, sufficient resources, not templates)
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState = 'poweredOn' and memoryMB >= 2048 and isTemplate = false and len(disks) <= 4"

# Large VMs requiring special handling
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryMB > 32768 or sum(disks[*].capacityGB) > 2000"

# VMs with migration concerns
kubectl mtv get inventory vms --provider vsphere-prod --query "where len(disks) > 8 or memoryMB > 65536 or any(disks[*].capacityGB > 2000)"
```

#### Environment-Specific Filtering

```bash
# Production VMs in specific clusters
kubectl mtv get inventory vms --provider vsphere-prod --query "where cluster.name ilike '%production%' and powerState = 'poweredOn'"

# VMs on specific hosts
kubectl mtv get inventory vms --provider vsphere-prod --query "where host.name in ['esxi-01.prod.com', 'esxi-02.prod.com']"

# VMs in specific datacenters
kubectl mtv get inventory vms --provider vsphere-prod --query "where datacenter.name = 'DC-East' and not template"
```

### Storage and Network Analysis

#### Storage-Based Queries

```bash
# VMs with large storage requirements
kubectl mtv get inventory vms --provider vsphere-prod --query "where sum(disks[*].capacityGB) > 500"

# VMs with multiple disks
kubectl mtv get inventory vms --provider vsphere-prod --query "where len(disks) > 2"

# VMs on specific datastores
kubectl mtv get inventory vms --provider vsphere-prod --query "where any(disks[*].datastore.name = 'SSD-Datastore-01')"

# VMs with thin-provisioned disks
kubectl mtv get inventory vms --provider vsphere-prod --query "where any(disks[*].thinProvisioned = true)"
```

#### Network-Based Queries

```bash
# VMs with multiple network adapters
kubectl mtv get inventory vms --provider vsphere-prod --query "where len(networks) > 1"

# VMs on specific networks
kubectl mtv get inventory vms --provider vsphere-prod --query "where any(networks[*].name = 'Production Network')"

# VMs with static IP assignments
kubectl mtv get inventory vms --provider vsphere-prod --query "where any(networks[*].ipAddress is not null)"
```

### Querying Provider Status and Resource Counts

#### Provider Health Monitoring

```bash
# Get provider inventory status
kubectl mtv get inventory providers --name vsphere-prod --query "where status = 'Ready'"

# Check provider resource counts
kubectl mtv get inventory providers --name vsphere-prod --query "where vmCount > 100"

# Monitor provider connectivity
kubectl mtv get inventory provider --query "where lastHeartbeat > '2024-01-01T00:00:00Z'"
```

#### Infrastructure Resource Queries

```bash
# Hosts with high VM density
kubectl mtv get inventory hosts --provider vsphere-prod --query "where vmCount > 20"

# Datastores with low free space
kubectl mtv get inventory datastores --provider vsphere-prod --query "where freeSpaceGB < 100"

# Resource pools with high utilization
kubectl mtv get inventory resource-pools --provider vsphere-prod --query "where memoryUsageGB > 50000"
```

## Provider-Specific Query Examples

### vSphere Specific Queries

```bash
# VMs with VMware Tools installed
kubectl mtv get inventory vms --provider vsphere-prod --query "where vmwareToolsStatus = 'toolsOk'"

# VMs requiring VMware Tools updates
kubectl mtv get inventory vms --provider vsphere-prod --query "where vmwareToolsStatus in ['toolsOld', 'toolsNotInstalled']"

# VMs with snapshots
kubectl mtv get inventory vms --provider vsphere-prod --query "where snapshotCount > 0"

# VMs in specific folders
kubectl mtv get inventory vms --provider vsphere-prod --query "where folder.name ~= '.*Production.*'"
```

### oVirt Specific Queries

```bash
# VMs with high availability enabled
kubectl mtv get inventory vms --provider ovirt-prod --query "where highAvailability = true"

# VMs using specific disk profiles
kubectl mtv get inventory vms --provider ovirt-prod --query "where any(disks[*].profile.name = 'high-performance')"

# VMs with ballooning enabled
kubectl mtv get inventory vms --provider ovirt-prod --query "where memoryBalloon = true"
```

### OpenStack Specific Queries

```bash
# Instances by flavor
kubectl mtv get inventory instances --provider openstack-prod --query "where flavor.name = 'm1.large'"

# Instances with floating IPs
kubectl mtv get inventory instances --provider openstack-prod --query "where floatingIP is not null"

# Instances in specific availability zones
kubectl mtv get inventory instances --provider openstack-prod --query "where availabilityZone = 'nova'"

# Active instances only
kubectl mtv get inventory instances --provider openstack-prod --query "where status = 'ACTIVE'"
```

### EC2 Specific Queries

EC2 fields use PascalCase naming conventions:

```bash
# Running instances of a specific type
kubectl mtv get inventory vms --provider ec2-prod --query "where State.Name = 'running' and InstanceType = 'm5.xlarge'"

# Instances in a specific availability zone
kubectl mtv get inventory vms --provider ec2-prod --query "where Placement.AvailabilityZone = 'us-east-1a'"

# Instances by platform
kubectl mtv get inventory vms --provider ec2-prod --query "where PlatformDetails ~= '.*Linux.*'"

# Instances in a specific VPC
kubectl mtv get inventory vms --provider ec2-prod --query "where VpcId = 'vpc-0123456789abcdef0'"

# Instances with public IP addresses
kubectl mtv get inventory vms --provider ec2-prod --query "where PublicIpAddress is not null"
```

## Query Optimization and Performance Tips

### Effective Query Design

#### Use Specific Filters Early

```bash
# Good: Filter by specific conditions first
kubectl mtv get inventory vms --provider vsphere-prod --query "where cluster.name = 'Prod-Cluster' and powerState = 'poweredOn'"

# Less efficient: Broad queries with complex conditions
kubectl mtv get inventory vms --provider vsphere-prod --query "where len(name) > 5 and memoryMB > 1024"
```

#### Leverage Indexes

```bash
# Indexed fields (typically perform better)
kubectl mtv get inventory vms --provider vsphere-prod --query "where name = 'specific-vm'"
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState = 'poweredOn'"

# Use exact matches when possible
kubectl mtv get inventory vms --provider vsphere-prod --query "where host.name = 'esxi-01.example.com'"
```

#### Combine Related Conditions

```bash
# Good: Combine related VM criteria
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState = 'poweredOn' and memoryMB >= 4096 and len(disks) <= 3"

# Better structure for complex queries
kubectl mtv get inventory vms --provider vsphere-prod --query "where (name ~= '^prod-.*' or name ~= '^web-.*') and powerState != 'poweredOff'"
```

### Query Performance Best Practices

#### Use Appropriate Operators

```bash
# Prefer equality over patterns when possible
kubectl mtv get inventory vms --provider vsphere-prod --query "where name = 'exact-vm-name'"

# Use case-insensitive matching sparingly
kubectl mtv get inventory vms --provider vsphere-prod --query "where name ilike '%database%'"

# Optimize regex patterns
kubectl mtv get inventory vms --provider vsphere-prod --query "where name ~= '^prod-web-[0-9]{2}$'"
```

#### Limit Result Sets

```bash
# Use specific conditions to reduce results
kubectl mtv get inventory vms --provider vsphere-prod --query "where cluster.name = 'Production' and powerState = 'poweredOn'"

# Filter by infrastructure components
kubectl mtv get inventory vms --provider vsphere-prod --query "where datacenter.name = 'DC1' and host.name ~= 'esxi-0[1-5].*'"
```

## Query Debugging and Troubleshooting

### Common Query Errors

#### Syntax Errors

```bash
# Incorrect: Missing quotes
kubectl mtv get inventory vms --provider vsphere-prod --query "where name = web-01"
# Error: Unexpected token

# Correct: Proper quoting
kubectl mtv get inventory vms --provider vsphere-prod --query "where name = 'web-01'"
```

#### Field Reference Errors

```bash
# Incorrect: Invalid field reference
kubectl mtv get inventory vms --provider vsphere-prod --query "where vm.name = 'test'"
# Error: Unknown field

# Correct: Valid field reference  
kubectl mtv get inventory vms --provider vsphere-prod --query "where name = 'test'"
```

#### Type Mismatch Errors

```bash
# Incorrect: String comparison with number
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryMB = '8192'"
# May cause type conversion issues

# Correct: Numeric comparison
kubectl mtv get inventory vms --provider vsphere-prod --query "where memoryMB = 8192"
```

#### Missing `where` Keyword

```bash
# Incorrect: Forgetting the where keyword
kubectl mtv get inventory vms --provider vsphere-prod --query "name = 'vm1'"

# Correct: Include where
kubectl mtv get inventory vms --provider vsphere-prod --query "where name = 'vm1'"
```

#### Using Parentheses Instead of Brackets for `IN`

```bash
# Incorrect: Parentheses
kubectl mtv get inventory vms --provider vsphere-prod --query "where name in ('vm1', 'vm2')"

# Correct: Square brackets
kubectl mtv get inventory vms --provider vsphere-prod --query "where name in ['vm1', 'vm2']"
```

### Query Testing and Validation

#### Start Simple

```bash
# Test basic connectivity
kubectl mtv get inventory vms --provider vsphere-prod --query "where name is not null"

# Test specific field access
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState = 'poweredOn'"

# Gradually add complexity
kubectl mtv get inventory vms --provider vsphere-prod --query "where powerState = 'poweredOn' and memoryMB > 4096"
```

#### Use JSON Output for Field Discovery

```bash
# Discover available fields
kubectl mtv get inventory vms --provider vsphere-prod --output json | jq '.items[0]' | head -20

# Understand field structure
kubectl mtv get inventory vms --provider vsphere-prod --output json | jq '.items[0].disks[0]'

# Test field access
kubectl mtv get inventory vms --provider vsphere-prod --query "where disks[0].capacityGB > 50"
```

#### Debug with Verbosity

```bash
# Enable debug logging
kubectl mtv get inventory vms --provider vsphere-prod --query "where name = 'test-vm'" -v=2

# Check inventory service logs
kubectl logs -n konveyor-forklift deployment/forklift-inventory -f
```

## Integration with Migration Planning

### Query-Driven Plan Creation

```bash
# Create migration plans based on queries
kubectl mtv get inventory vms --provider vsphere-prod \
  --query "where powerState = 'poweredOn' and memoryMB <= 8192 and len(disks) <= 2" \
  --output planvms > small-vms.yaml

kubectl mtv create plan --name small-vm-migration \
  --source vsphere-prod \
  --vms @small-vms.yaml
```

### Automated Migration Workflows

```bash
#!/bin/bash
# Automated query-based migration planning

PROVIDER="vsphere-prod"
DATE=$(date +%Y%m%d)

# Phase 1: Small VMs
kubectl mtv get inventory vms --provider "$PROVIDER" \
  --query "where powerState = 'poweredOn' and memoryMB <= 4096" \
  --output planvms > "phase1-${DATE}.yaml"

# Phase 2: Medium VMs  
kubectl mtv get inventory vms --provider "$PROVIDER" \
  --query "where powerState = 'poweredOn' and memoryMB between 4097 and 16384" \
  --output planvms > "phase2-${DATE}.yaml"

# Phase 3: Large VMs
kubectl mtv get inventory vms --provider "$PROVIDER" \
  --query "where powerState = 'poweredOn' and memoryMB > 16384" \
  --output planvms > "phase3-${DATE}.yaml"

echo "Migration phases planned for $DATE"
```

## Advanced Pattern Matching

### Regular Expression Patterns

The `~=` (regex match) and `~!` (regex not match) operators support full regular expressions:

```bash
# Anchored patterns
kubectl mtv get inventory vms --provider vsphere-prod --query "where name ~= '^web-[0-9]+$'"

# Case-insensitive patterns  
kubectl mtv get inventory vms --provider vsphere-prod --query "where guestOS ~= '(?i)windows.*'"

# Complex patterns with alternation
kubectl mtv get inventory vms --provider vsphere-prod --query "where name ~= '^(web|app|db)-[a-z]+-[0-9]{2}$'"

# Exclude patterns
kubectl mtv get inventory vms --provider vsphere-prod --query "where name ~! '.*(test|dev|tmp).*'"
```

### Date and Time Queries

Date and timestamp literals are expressed as strings and compared using standard operators:

```bash
# Date comparisons (YYYY-MM-DD)
kubectl mtv get inventory vms --provider vsphere-prod --query "where created >= '2024-01-01'"

# Timestamp ranges (RFC3339)
kubectl mtv get inventory vms --provider vsphere-prod --query "where lastModified between '2024-01-01T00:00:00Z' and '2024-12-31T23:59:59Z'"

# Recent changes
kubectl mtv get inventory vms --provider vsphere-prod --query "where lastModified >= '2024-11-01T00:00:00Z'"
```

## Query Language Summary

TSL provides a powerful, SQL-like query interface for kubectl-mtv inventory operations:

- **Operators**: Comparison, arithmetic, logical, string matching, set, range, and null operators
- **Functions**: `len`, `any`, `all`, `sum` for complex data analysis
- **Data Types**: String, numeric, boolean, date, timestamp, array, null, and SI unit suffixes
- **Field Access**: Dot notation, array indexing, wildcards, and implicit traversal
- **Pattern Matching**: `like`, `ilike`, `~=` (regex), and `~!` (regex not match)
- **Sorting and Limiting**: `order by` and `limit` for result control
- **Provider Fields**: vSphere, oVirt, OpenStack, EC2 fields plus computed fields
- **Performance**: Optimized for large-scale inventory queries

> **See also**: [Chapter 27: TSL - Tree Search Language Reference](../27-tsl-tree-search-language-reference) for a concise, printable quick-reference card.

## Next Steps

After mastering TSL queries:

1. **TSL Quick Reference**: See the self-contained syntax reference in [Chapter 27: TSL - Tree Search Language Reference](../27-tsl-tree-search-language-reference)
2. **Apply to Mappings**: Use queries for mapping creation in [Chapter 11: Mapping Management](../11-mapping-management)
3. **Plan Creation**: Leverage queries for plan development in [Chapter 13: Migration Plan Creation](../13-migration-plan-creation)
4. **VM Customization**: Use query results for VM customization in [Chapter 14: Customizing Individual VMs](../14-customizing-individual-vms-planvms-format)
5. **Optimization**: Apply query insights in [Chapter 16: Migration Process Optimization](../16-migration-process-optimization)

---

*Previous: [Chapter 9: Inventory Management](../09-inventory-management)*  
*Next: [Chapter 11: Mapping Management](../11-mapping-management)*
