---
layout: page
title: "Chapter 27: TSL - Tree Search Language Reference"
---

TSL (Tree Search Language) is a powerful, open-source query language developed by Yaacov Zamir that provides SQL-like filtering capabilities for structured data. This chapter is a **pure language reference** covering TSL syntax, operators, functions, and field access rules.

For practical kubectl-mtv examples, provider-specific field tables, migration workflows, and troubleshooting, see [Chapter 10: Query Language Reference and Advanced Filtering](/kubectl-mtv/10-query-language-reference-and-advanced-filtering).

## What is TSL?

TSL is a query language with grammar similar to SQL's WHERE clause. It parses human-readable filter expressions into an abstract syntax tree that can be evaluated against structured data such as JSON objects.

Key characteristics:

- **SQL-like syntax** -- familiar to anyone who has written SQL WHERE clauses
- **Type-aware** -- supports strings, numbers, booleans, dates, arrays, and null
- **Array-capable** -- first-class functions for working with array fields (`len`, `any`, `all`, `sum`)
- **SI unit literals** -- use `Ki`, `Mi`, `Gi`, `Ti`, `Pi` suffixes for byte quantities
- **Regex support** -- `~=` and `~!` operators for regular expression matching

## Query Structure

A full TSL query follows this structure:

```
[SELECT fields] WHERE condition [ORDER BY field [ASC|DESC]] [LIMIT n]
```

The `SELECT` clause is optional. The most common form is:

```
where <condition> [order by <field> [asc|desc]] [limit <n>]
```

### Clauses

| Clause | Required | Description |
|--------|----------|-------------|
| `SELECT` | No | Select specific fields (rarely used in kubectl-mtv) |
| `WHERE` | Yes | Filter condition |
| `ORDER BY` | No | Sort by field, optionally `ASC` (default) or `DESC` |
| `LIMIT` | No | Maximum number of results to return |

### Examples

```sql
-- Filter only
where name = 'web-01'

-- Filter with sorting
where powerState = 'poweredOn' order by memoryMB desc

-- Filter with sorting and limit
where memoryMB > 1024 order by memoryMB desc limit 10

-- Select specific fields (reduces output size)
select name, memoryMB, cpuCount where powerState = 'poweredOn' limit 10

-- Full query: select, where, order by, limit
select name, memoryMB as mem where memoryMB > 4096 order by memoryMB desc limit 5
```

## Data Types and Literals

### Strings

Strings are enclosed in single quotes:

```sql
name = 'web-server-01'
cluster.name = 'Production Cluster'
```

### Numbers

Integer and decimal literals:

```sql
memoryMB = 8192
cpuCount = 4
diskGB = 100.5
```

### SI Unit Suffixes

Use SI suffixes to express byte quantities concisely. The parser expands them to plain numbers:

| Suffix | Multiplier | Example | Expanded |
|--------|-----------|---------|----------|
| `Ki` | 1,024 | `4Ki` | 4096 |
| `Mi` | 1,048,576 | `512Mi` | 536870912 |
| `Gi` | 1,073,741,824 | `4Gi` | 4294967296 |
| `Ti` | 1,099,511,627,776 | `1Ti` | 1099511627776 |
| `Pi` | 1,125,899,906,842,624 | `1Pi` | 1125899906842624 |

```sql
-- Use SI units for byte-valued fields
memory > 4Gi
sum(disks[*].capacity) > 100Gi
```

### Booleans

```sql
isTemplate = false
haEnabled = true
```

### Arrays (for `IN` operator)

Arrays use square brackets with single-quoted elements:

```sql
name in ['vm1', 'vm2', 'vm3']
powerState in ['poweredOn', 'suspended']
```

### Null

```sql
description is null
guestIP is not null
```

### Dates and Timestamps

Date and timestamp literals are expressed as strings:

```sql
-- Date (YYYY-MM-DD)
created >= '2024-01-01'

-- Timestamp (RFC3339)
lastModified >= '2024-01-01T00:00:00Z'

-- Timestamp range
lastModified between '2024-01-01T00:00:00Z' and '2024-12-31T23:59:59Z'
```

## Operators

### Comparison Operators

| Operator | Symbol(s) | Description | Example |
|----------|-----------|-------------|---------|
| Equal | `=` | Equal to | `name = 'vm1'` |
| Not equal | `!=`, `<>` | Not equal to | `status != 'down'` |
| Less than | `<` | Less than | `memoryMB < 4096` |
| Less or equal | `<=` | Less than or equal to | `cpuCount <= 2` |
| Greater than | `>` | Greater than | `diskGB > 100` |
| Greater or equal | `>=` | Greater than or equal to | `memoryMB >= 8192` |

### Arithmetic Operators

| Operator | Description |
|----------|-------------|
| `+` | Addition |
| `-` | Subtraction |
| `*` | Multiplication |
| `/` | Division |
| `%` | Modulo |

```sql
memoryMB / cpuCount > 4096
diskGB * 1024 > storageUsed
```

### String Matching Operators

| Operator | Keyword | Description | Example |
|----------|---------|-------------|---------|
| SQL LIKE | `like` | Case-sensitive pattern (`%` = any chars, `_` = one char) | `name like 'web-%'` |
| SQL ILIKE | `ilike` | Case-insensitive LIKE | `name ilike 'WEB-%'` |
| Regex match | `~=` | Regular expression match | `name ~= 'prod-.*'` |
| Regex not match | `~!` | Regular expression does not match | `name ~! 'test-.*'` |

#### Regular Expression Notes

The `~=` and `~!` operators accept full regular expressions:

```sql
-- Anchored patterns
name ~= '^web-[0-9]+$'

-- Alternation
name ~= '^(web|app|db)-.*'

-- Case-insensitive regex
guestOS ~= '(?i)windows.*'
```

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `and` | Logical AND | `memoryMB > 4096 and cpuCount >= 2` |
| `or` | Logical OR | `name = 'vm1' or name = 'vm2'` |
| `not` | Logical NOT | `not (status = 'down')` |

Parentheses control precedence:

```sql
(name ~= '^prod-.*' or name ~= '^web-.*') and status != 'down'
not (isTemplate = true or status = 'down')
```

### Set and Range Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `in` | Value in array | `name in ['vm1', 'vm2', 'vm3']` |
| `not in` | Value not in array | `guestId not in ['rhel8_64Guest', '']` |
| `between ... and` | Value in inclusive range | `memoryMB between 4096 and 16384` |

**Important**: The `in` and `not in` operators require **square brackets** `[...]` for the value list, not parentheses.

### Null Checking Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `is null` | Field is null | `description is null` |
| `is not null` | Field is not null | `guestIP is not null` |

## SELECT, ORDER BY, and LIMIT

### SELECT Clause

The optional `SELECT` clause limits which fields are returned, reducing output size:

```sql
-- Select specific fields
select name, memoryMB, cpuCount where powerState = 'poweredOn'

-- Field with alias (as)
select memoryMB as mem, cpuCount as cpus where memoryMB > 4096

-- Reducers: sum, len, any, all
select name, sum(disks[*].capacityGB) as totalDisk where powerState = 'poweredOn'
select name, len(disks) as diskCount where len(disks) > 1
```

### ORDER BY Clause

Sort results by field in ascending (default) or descending order:

```sql
-- Ascending (default)
where powerState = 'poweredOn' order by name

-- Descending
where memoryMB > 1024 order by memoryMB desc

-- Multiple sort keys
where powerState = 'poweredOn' order by memoryMB desc, name asc

-- Order by alias
select name, memoryMB as mem where memoryMB > 0 order by mem desc
```

### LIMIT Clause

Restrict the number of results returned:

```sql
-- Top 10
where powerState = 'poweredOn' order by memoryMB desc limit 10

-- First 50 alphabetically
where powerState = 'poweredOn' order by name limit 50

-- Limit without order (first N matching)
where name ~= 'prod-.*' limit 5
```

### Combined Examples (kubectl-mtv)

```bash
# Top 10 largest VMs by memory
kubectl mtv get inventory vm vsphere-prod -q "where powerState = 'poweredOn' order by memoryMB desc limit 10"

# Compact output: only name, memory, CPU
kubectl mtv get inventory vm vsphere-prod -q "select name, memoryMB, cpuCount where powerState = 'poweredOn' limit 10"

# Full query: select + where + order + limit
kubectl mtv get inventory vm vsphere-prod -q "select name, memoryMB as mem where memoryMB > 4096 order by mem desc limit 5"
```

## Functions

### `len(field)`

Returns the length of an array or string. The comparison operator goes **outside** the parentheses:

```sql
len(disks) > 2
len(nics) >= 2
len(networks) = 1
len(name) > 10
```

### `any(condition)`

Returns true if **any** element in the array matches the condition. Use `[*]` wildcard inside:

```sql
any(disks[*].shared = true)
any(concerns[*].category = 'Critical')
any(networks[*].name ~= '.*-prod')
any(disks[*].capacity > 100Gi)
```

### `all(condition)`

Returns true if **all** elements in the array match the condition:

```sql
all(disks[*].shared = false)
all(disks[*].capacityGB >= 20)
all(nics[*].connected = true)
```

### `sum(field)`

Returns the sum of numeric values across array elements:

```sql
sum(disks[*].capacityGB) > 500
sum(disks[*].capacity) > 100Gi
sum(hosts[*].numCpu) > 50
```

## Field Access

### Dot Notation

Access nested fields with dots:

```sql
host.cluster.name = 'Production'
guest.distribution ~= 'Red Hat.*'
parent.kind = 'Folder'
```

### Array Index Access

Access a specific array element by zero-based index:

```sql
disks[0].capacityGB > 100
nics[0].mac = '00:50:56:xx:xx:xx'
```

### Array Wildcard Access

Use `[*]` to reference all elements (used inside `any`, `all`, `sum`):

```sql
any(disks[*].datastore.id = 'datastore-12')
sum(disks[*].capacity) > 500Gi
all(nics[*].connected = true)
```

### Implicit Traversal

Dot notation on an array field implicitly traverses all elements (equivalent to `[*]`):

```sql
-- These are equivalent:
disks.capacity
disks[*].capacity
```

## Common Syntax Pitfalls

| Mistake | Correct Form |
|---------|-------------|
| Missing quotes: `name = web-01` | `name = 'web-01'` |
| Parentheses for IN: `name in ('a','b')` | `name in ['a','b']` |
| Number as string: `memoryMB = '8192'` | `memoryMB = 8192` |
| Missing `where`: `"name = 'vm1'"` | `"where name = 'vm1'"` |

## Quick Reference Card

```
QUERY STRUCTURE
  [SELECT fields] WHERE condition [ORDER BY field [ASC|DESC]] [LIMIT n]

EXAMPLES
  where name = 'vm1'
  where memoryMB > 4096 order by memoryMB desc limit 10
  select name, memoryMB where powerState = 'poweredOn' order by name limit 5

DATA TYPES
  Strings       'single quoted'
  Numbers       42, 3.14
  SI Units      Ki  Mi  Gi  Ti  Pi
  Booleans      true, false
  Arrays        ['a', 'b', 'c']
  Null          null
  Dates         '2024-01-01', '2024-01-01T00:00:00Z'

COMPARISONS     =  !=  <>  <  <=  >  >=
ARITHMETIC      +  -  *  /  %
STRINGS         like  ilike  ~= (regex)  ~! (regex not)
LOGIC           and  or  not  ( )
SETS            in [...]  not in [...]  between X and Y
NULLS           is null  is not null

FUNCTIONS
  len(field)          array/string length
  any(cond)           true if any element matches
  all(cond)           true if all elements match
  sum(field)          sum of numeric array values

FIELD ACCESS
  field.sub           dot notation
  field[0]            index access (zero-based)
  field[*].sub        wildcard (all elements)
  field.sub           implicit traversal (same as field[*].sub)

SORTING         order by field [asc|desc]
LIMITING        limit N
```

## Built-in Help

Access the TSL reference directly from the command line:

```bash
kubectl mtv help tsl
```

## Further Reading

- [Chapter 10: Query Language Reference and Advanced Filtering](/kubectl-mtv/10-query-language-reference-and-advanced-filtering) -- provider field tables, practical examples, migration workflows, and optimization tips
- [Chapter 9: Inventory Management](/kubectl-mtv/09-inventory-management) -- inventory commands that accept TSL queries
- [Chapter 13: Migration Plan Creation](/kubectl-mtv/13-migration-plan-creation) -- using TSL for query-driven VM selection

---

*Previous: [Chapter 26: Command Reference](/kubectl-mtv/26-command-reference)*
*Next: [Chapter 28: KARL - Kubernetes Affinity Rule Language Reference](/kubectl-mtv/28-karl-kubernetes-affinity-rule-language-reference)*
