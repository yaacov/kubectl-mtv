---
layout: page
title: "Chapter 8: Query Language Reference and Advanced Filtering"
---

kubectl-mtv integrates the powerful Tree Search Language (TSL), developed by Yaacov Zamir, which provides SQL-like filtering capabilities for inventory resources. This chapter provides a complete reference for TSL syntax and advanced query techniques.

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
city in ('paris', 'rome', 'milan') and state != 'france'

-- Range queries
pages between 100 and 200 and author.name ~= 'Hilbert'
```

### TSL in kubectl-mtv Context

TSL is used throughout kubectl-mtv for inventory filtering:

```bash
# Filter VMs by power state
kubectl mtv get inventory vms vsphere-prod -q "where powerState = 'poweredOn'"

# Filter by memory size
kubectl mtv get inventory vms vsphere-prod -q "where memoryMB > 8192"

# Complex filtering
kubectl mtv get inventory vms vsphere-prod -q "where name ~= 'prod-.*' and memoryMB >= 4096"
```

## Query Structure

TSL queries in kubectl-mtv follow this general structure:

```
kubectl mtv get inventory <resource> <provider> -q "where <TSL_EXPRESSION>"
```

### Basic Query Components

1. **WHERE Clause**: The `where` keyword starts the filter expression
2. **Field References**: Access object fields using dot notation (e.g., `vm.name`, `host.cluster.name`)
3. **Operators**: Comparison, logical, and pattern matching operators
4. **Literals**: String, numeric, boolean, and array values
5. **Functions**: Built-in functions for complex operations

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
-- Array values for IN operations
name in ('vm1', 'vm2', 'vm3')
powerState in ('poweredOn', 'suspended')
```

### Date and Timestamp Literals

```sql
-- Date literals (YYYY-MM-DD)
created >= '2024-01-01'

-- Timestamp literals (RFC3339 format)
lastModified >= '2024-01-01T00:00:00Z'
```

### Null Values

```sql
-- Check for null values
description is null
annotation is not null
```

## Operators Reference

TSL provides a comprehensive set of operators verified from the vendor code:

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
kubectl mtv get inventory vms vsphere-prod -q "where memoryMB >= 8192"
kubectl mtv get inventory vms vsphere-prod -q "where numCpu < 4"
kubectl mtv get inventory vms vsphere-prod -q "where powerState != 'poweredOff'"
```

### String Matching Operators

| Operator | Description | Example |
|----------|-------------|---------|
| **LIKE** | Case-sensitive pattern matching (SQL LIKE) | `name like 'web-%'` |
| **ILIKE** | Case-insensitive pattern matching | `name ilike 'WEB-%'` |
| **REQ** | Regular expression equals (`~=`) | `name ~= 'prod-.*'` |
| **RNE** | Regular expression not equals (`!~`) | `name !~ 'test-.*'` |

```bash
# String matching examples
kubectl mtv get inventory vms vsphere-prod -q "where name like 'prod-%'"
kubectl mtv get inventory vms vsphere-prod -q "where guestOS ilike '%windows%'"
kubectl mtv get inventory vms vsphere-prod -q "where name ~= '^web-[0-9]+$'"
```

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| **AND** | Logical AND | `memoryMB > 4096 and numCpu >= 2` |
| **OR** | Logical OR | `name = 'vm1' or name = 'vm2'` |
| **NOT** | Logical NOT | `not (powerState = 'poweredOff')` |

```bash
# Logical operator examples
kubectl mtv get inventory vms vsphere-prod -q "where powerState = 'poweredOn' and memoryMB > 8192"
kubectl mtv get inventory vms vsphere-prod -q "where name = 'web-01' or name = 'web-02'"
kubectl mtv get inventory vms vsphere-prod -q "where not template"
```

### Array and Range Operators

| Operator | Description | Example |
|----------|-------------|---------|
| **IN** | Value in array | `name in ('vm1', 'vm2', 'vm3')` |
| **BETWEEN** | Value between two values | `memoryMB between 4096 and 16384` |

```bash
# Array and range examples
kubectl mtv get inventory vms vsphere-prod -q "where powerState in ('poweredOn', 'suspended')"
kubectl mtv get inventory vms vsphere-prod -q "where memoryMB between 8192 and 32768"
```

### Null Checking Operators

| Operator | Description | Example |
|----------|-------------|---------|
| **IS** | Check for null/not null | `description is null` |
| **IS NOT** | Check for not null | `annotation is not null` |

```bash
# Null checking examples
kubectl mtv get inventory vms vsphere-prod -q "where description is not null"
kubectl mtv get inventory vms vsphere-prod -q "where guestIP is null"
```

## Functions Reference

TSL provides built-in functions for complex data analysis:

### LEN Function

Returns the length of arrays or strings:

```sql
-- Array length
len disks > 2
len networks = 1

-- String length  
len name > 10
```

```bash
# LEN function examples
kubectl mtv get inventory vms vsphere-prod -q "where len disks > 2"
kubectl mtv get inventory vms vsphere-prod -q "where len name >= 12"
```

### ANY Function

Tests if any element in an array matches a condition:

```sql
-- Check if any disk meets condition
any disks.capacityGB > 500

-- Check if any network matches
any networks.name = 'Production Network'
```

```bash
# ANY function examples
kubectl mtv get inventory vms vsphere-prod -q "where any disks.capacityGB > 100"
kubectl mtv get inventory vms vsphere-prod -q "where any networks.name ~= '.*-prod'"
```

### ALL Function

Tests if all elements in an array match a condition:

```sql
-- Check if all disks are large
all disks.capacityGB > 50

-- Check if all networks are production
all networks.name ~= '.*-prod'
```

```bash
# ALL function examples
kubectl mtv get inventory vms vsphere-prod -q "where all disks.capacityGB >= 20"
kubectl mtv get inventory vms vsphere-prod -q "where all networks.type = 'standard'"
```

### SUM Function

Calculates the sum of numeric values in an array:

```sql
-- Total disk capacity
sum disks.capacityGB > 1000

-- Total CPU count
sum hosts.numCpu > 50
```

```bash
# SUM function examples
kubectl mtv get inventory vms vsphere-prod -q "where sum disks.capacityGB > 500"
kubectl mtv get inventory clusters vsphere-prod -q "where sum hosts.memoryGB > 100000"
```

## Advanced Query Examples

### Filtering by Power State, Memory, and Name Patterns

#### Power State Filtering

```bash
# Find only powered-on VMs
kubectl mtv get inventory vms vsphere-prod -q "where powerState = 'poweredOn'"

# Find VMs that are not powered off
kubectl mtv get inventory vms vsphere-prod -q "where powerState != 'poweredOff'"

# Find suspended or powered-on VMs
kubectl mtv get inventory vms vsphere-prod -q "where powerState in ('poweredOn', 'suspended')"
```

#### Memory-Based Filtering

```bash
# High-memory VMs (>16GB)
kubectl mtv get inventory vms vsphere-prod -q "where memoryMB > 16384"

# Medium memory range (4-16GB)
kubectl mtv get inventory vms vsphere-prod -q "where memoryMB between 4096 and 16384"

# Low-memory VMs suitable for quick migration
kubectl mtv get inventory vms vsphere-prod -q "where memoryMB <= 4096 and powerState = 'poweredOn'"
```

#### Name Pattern Filtering

```bash
# Production VMs by naming convention
kubectl mtv get inventory vms vsphere-prod -q "where name ~= '^prod-.*'"

# Web servers by pattern
kubectl mtv get inventory vms vsphere-prod -q "where name ~= '^web-[0-9]+$'"

# Exclude test and development VMs
kubectl mtv get inventory vms vsphere-prod -q "where name !~ '.*(test|dev|tmp).*'"

# VMs containing specific keywords
kubectl mtv get inventory vms vsphere-prod -q "where name ilike '%database%' or name ilike '%db%'"
```

### Complex Multi-Condition Queries

#### Migration Readiness Assessment

```bash
# VMs ready for migration (powered on, sufficient resources, not templates)
kubectl mtv get inventory vms vsphere-prod -q "where powerState = 'poweredOn' and memoryMB >= 2048 and not template and len disks <= 4"

# Large VMs requiring special handling
kubectl mtv get inventory vms vsphere-prod -q "where memoryMB > 32768 or sum disks.capacityGB > 2000"

# VMs with migration concerns
kubectl mtv get inventory vms vsphere-prod -q "where len disks > 8 or memoryMB > 65536 or any disks.capacityGB > 2000"
```

#### Environment-Specific Filtering

```bash
# Production VMs in specific clusters
kubectl mtv get inventory vms vsphere-prod -q "where cluster.name ilike '%production%' and powerState = 'poweredOn'"

# VMs on specific hosts
kubectl mtv get inventory vms vsphere-prod -q "where host.name in ('esxi-01.prod.com', 'esxi-02.prod.com')"

# VMs in specific datacenters
kubectl mtv get inventory vms vsphere-prod -q "where datacenter.name = 'DC-East' and not template"
```

### Storage and Network Analysis

#### Storage-Based Queries

```bash
# VMs with large storage requirements
kubectl mtv get inventory vms vsphere-prod -q "where sum disks.capacityGB > 500"

# VMs with multiple disks
kubectl mtv get inventory vms vsphere-prod -q "where len disks > 2"

# VMs on specific datastores
kubectl mtv get inventory vms vsphere-prod -q "where any disks.datastore.name = 'SSD-Datastore-01'"

# VMs with thin-provisioned disks
kubectl mtv get inventory vms vsphere-prod -q "where any disks.thinProvisioned = true"
```

#### Network-Based Queries

```bash
# VMs with multiple network adapters
kubectl mtv get inventory vms vsphere-prod -q "where len networks > 1"

# VMs on specific networks
kubectl mtv get inventory vms vsphere-prod -q "where any networks.name = 'Production Network'"

# VMs with static IP assignments
kubectl mtv get inventory vms vsphere-prod -q "where any networks.ipAddress is not null"
```

### Querying Provider Status and Resource Counts

#### Provider Health Monitoring

```bash
# Get provider inventory status
kubectl mtv get inventory provider vsphere-prod -q "where status = 'Ready'"

# Check provider resource counts
kubectl mtv get inventory provider vsphere-prod -q "where vmCount > 100"

# Monitor provider connectivity
kubectl mtv get inventory providers -q "where lastHeartbeat > '2024-01-01T00:00:00Z'"
```

#### Infrastructure Resource Queries

```bash
# Hosts with high VM density
kubectl mtv get inventory hosts vsphere-prod -q "where vmCount > 20"

# Datastores with low free space
kubectl mtv get inventory datastores vsphere-prod -q "where freeSpaceGB < 100"

# Resource pools with high utilization
kubectl mtv get inventory resourcepools vsphere-prod -q "where memoryUsageGB > 50000"
```

## Provider-Specific Query Examples

### vSphere Specific Queries

```bash
# VMs with VMware Tools installed
kubectl mtv get inventory vms vsphere-prod -q "where vmwareToolsStatus = 'toolsOk'"

# VMs requiring VMware Tools updates
kubectl mtv get inventory vms vsphere-prod -q "where vmwareToolsStatus in ('toolsOld', 'toolsNotInstalled')"

# VMs with snapshots
kubectl mtv get inventory vms vsphere-prod -q "where snapshotCount > 0"

# VMs in specific folders
kubectl mtv get inventory vms vsphere-prod -q "where folder.name ~= '.*Production.*'"
```

### oVirt Specific Queries

```bash
# VMs with high availability enabled
kubectl mtv get inventory vms ovirt-prod -q "where highAvailability = true"

# VMs using specific disk profiles
kubectl mtv get inventory vms ovirt-prod -q "where any disks.profile.name = 'high-performance'"

# VMs with ballooning enabled
kubectl mtv get inventory vms ovirt-prod -q "where memoryBalloon = true"
```

### OpenStack Specific Queries

```bash
# Instances by flavor
kubectl mtv get inventory instances openstack-prod -q "where flavor.name = 'm1.large'"

# Instances with floating IPs
kubectl mtv get inventory instances openstack-prod -q "where floatingIP is not null"

# Instances in specific availability zones
kubectl mtv get inventory instances openstack-prod -q "where availabilityZone = 'nova'"

# Active instances only
kubectl mtv get inventory instances openstack-prod -q "where status = 'ACTIVE'"
```

## Query Optimization and Performance Tips

### Effective Query Design

#### Use Specific Filters Early

```bash
# Good: Filter by specific conditions first
kubectl mtv get inventory vms vsphere-prod -q "where cluster.name = 'Prod-Cluster' and powerState = 'poweredOn'"

# Less efficient: Broad queries with complex conditions
kubectl mtv get inventory vms vsphere-prod -q "where len name > 5 and memoryMB > 1024"
```

#### Leverage Indexes

```bash
# Indexed fields (typically perform better)
kubectl mtv get inventory vms vsphere-prod -q "where name = 'specific-vm'"
kubectl mtv get inventory vms vsphere-prod -q "where powerState = 'poweredOn'"

# Use exact matches when possible
kubectl mtv get inventory vms vsphere-prod -q "where host.name = 'esxi-01.example.com'"
```

#### Combine Related Conditions

```bash
# Good: Combine related VM criteria
kubectl mtv get inventory vms vsphere-prod -q "where powerState = 'poweredOn' and memoryMB >= 4096 and len disks <= 3"

# Better structure for complex queries
kubectl mtv get inventory vms vsphere-prod -q "where (name ~= '^prod-.*' or name ~= '^web-.*') and powerState != 'poweredOff'"
```

### Query Performance Best Practices

#### Use Appropriate Operators

```bash
# Prefer equality over patterns when possible
kubectl mtv get inventory vms vsphere-prod -q "where name = 'exact-vm-name'"

# Use case-insensitive matching sparingly
kubectl mtv get inventory vms vsphere-prod -q "where name ilike '%database%'"

# Optimize regex patterns
kubectl mtv get inventory vms vsphere-prod -q "where name ~= '^prod-web-[0-9]{2}$'"
```

#### Limit Result Sets

```bash
# Use specific conditions to reduce results
kubectl mtv get inventory vms vsphere-prod -q "where cluster.name = 'Production' and powerState = 'poweredOn'"

# Filter by infrastructure components
kubectl mtv get inventory vms vsphere-prod -q "where datacenter.name = 'DC1' and host.name ~= 'esxi-0[1-5].*'"
```

## Query Debugging and Troubleshooting

### Common Query Errors

#### Syntax Errors

```bash
# Incorrect: Missing quotes
kubectl mtv get inventory vms vsphere-prod -q "where name = web-01"
# Error: Unexpected token

# Correct: Proper quoting
kubectl mtv get inventory vms vsphere-prod -q "where name = 'web-01'"
```

#### Field Reference Errors

```bash
# Incorrect: Invalid field reference
kubectl mtv get inventory vms vsphere-prod -q "where vm.name = 'test'"
# Error: Unknown field

# Correct: Valid field reference  
kubectl mtv get inventory vms vsphere-prod -q "where name = 'test'"
```

#### Type Mismatch Errors

```bash
# Incorrect: String comparison with number
kubectl mtv get inventory vms vsphere-prod -q "where memoryMB = '8192'"
# May cause type conversion issues

# Correct: Numeric comparison
kubectl mtv get inventory vms vsphere-prod -q "where memoryMB = 8192"
```

### Query Testing and Validation

#### Start Simple

```bash
# Test basic connectivity
kubectl mtv get inventory vms vsphere-prod -q "where name is not null"

# Test specific field access
kubectl mtv get inventory vms vsphere-prod -q "where powerState = 'poweredOn'"

# Gradually add complexity
kubectl mtv get inventory vms vsphere-prod -q "where powerState = 'poweredOn' and memoryMB > 4096"
```

#### Use JSON Output for Field Discovery

```bash
# Discover available fields
kubectl mtv get inventory vms vsphere-prod -o json | jq '.items[0]' | head -20

# Understand field structure
kubectl mtv get inventory vms vsphere-prod -o json | jq '.items[0].disks[0]'

# Test field access
kubectl mtv get inventory vms vsphere-prod -q "where disks[0].capacityGB > 50"
```

#### Debug with Verbosity

```bash
# Enable debug logging
kubectl mtv get inventory vms vsphere-prod -q "where name = 'test-vm'" -v=2

# Check inventory service logs
kubectl logs -n konveyor-forklift deployment/forklift-inventory -f
```

## Integration with Migration Planning

### Query-Driven Plan Creation

```bash
# Create migration plans based on queries
kubectl mtv get inventory vms vsphere-prod \
  -q "where powerState = 'poweredOn' and memoryMB <= 8192 and len disks <= 2" \
  -o planvms > small-vms.yaml

kubectl mtv create plan small-vm-migration \
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
kubectl mtv get inventory vms "$PROVIDER" \
  -q "where powerState = 'poweredOn' and memoryMB <= 4096" \
  -o planvms > "phase1-${DATE}.yaml"

# Phase 2: Medium VMs  
kubectl mtv get inventory vms "$PROVIDER" \
  -q "where powerState = 'poweredOn' and memoryMB between 4097 and 16384" \
  -o planvms > "phase2-${DATE}.yaml"

# Phase 3: Large VMs
kubectl mtv get inventory vms "$PROVIDER" \
  -q "where powerState = 'poweredOn' and memoryMB > 16384" \
  -o planvms > "phase3-${DATE}.yaml"

echo "Migration phases planned for $DATE"
```

## Advanced TSL Features

### Nested Field Access

```bash
# Access nested object properties
kubectl mtv get inventory vms vsphere-prod -q "where host.cluster.name = 'Production'"

# Array element access
kubectl mtv get inventory vms vsphere-prod -q "where disks[0].capacityGB > 100"

# Complex nested queries
kubectl mtv get inventory vms vsphere-prod -q "where host.datacenter.name = 'DC1' and any networks.vlan > 100"
```

### Regular Expression Patterns

```bash
# Anchored patterns
kubectl mtv get inventory vms vsphere-prod -q "where name ~= '^web-[0-9]+$'"

# Case-insensitive patterns  
kubectl mtv get inventory vms vsphere-prod -q "where guestOS ~= '(?i)windows.*'"

# Complex patterns
kubectl mtv get inventory vms vsphere-prod -q "where name ~= '^(web|app|db)-[a-z]+-[0-9]{2}$'"
```

### Date and Time Queries

```bash
# Date comparisons
kubectl mtv get inventory vms vsphere-prod -q "where created >= '2024-01-01'"

# Timestamp ranges
kubectl mtv get inventory vms vsphere-prod -q "where lastModified between '2024-01-01T00:00:00Z' and '2024-12-31T23:59:59Z'"

# Recent changes
kubectl mtv get inventory vms vsphere-prod -q "where lastModified >= '2024-11-01T00:00:00Z'"
```

## Query Language Summary

TSL provides a powerful, SQL-like query interface for kubectl-mtv inventory operations:

- **Operators**: Complete set of comparison, logical, string, and array operators
- **Functions**: LEN, ANY, ALL, SUM for complex data analysis  
- **Data Types**: String, numeric, boolean, date, timestamp, array, and null support
- **Field Access**: Dot notation for nested object properties
- **Pattern Matching**: LIKE, ILIKE, and regular expression support
- **Performance**: Optimized for large-scale inventory queries

## Next Steps

After mastering TSL queries:

1. **Apply to Mappings**: Use queries for mapping creation in [Chapter 9: Mapping Management](09-mapping-management)
2. **Plan Creation**: Leverage queries for plan development in [Chapter 10: Migration Plan Creation](10-migration-plan-creation)
3. **VM Customization**: Use query results for VM customization in [Chapter 11: Customizing Individual VMs](11-customizing-individual-vms-planvms-format)
4. **Optimization**: Apply query insights in [Chapter 13: Migration Process Optimization](13-migration-process-optimization)

---

*Previous: [Chapter 7: Inventory Management](07-inventory-management)*  
*Next: [Chapter 9: Mapping Management](09-mapping-management)*
