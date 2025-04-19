# Query Language Reference

This document provides a comprehensive guide to the query language used in `kubectl-mtv inventory` commands.

## Query Structure

A complete query can include the following clauses:

```sql
SELECT field1, field2 AS alias, function(field3) AS name
WHERE condition
ORDER BY field1 [ASC|DESC], field2
LIMIT n
```

All clauses are optional, allowing for flexible queries ranging from simple to complex.

## SELECT Clause

The SELECT clause specifies which fields to include in the output and how to name them.

### Syntax (SELECT Clause)

```sql
SELECT field1, field2 AS alias, function(field3) AS name
```

### Features

- **Field Selection**: Specify fields to include (default is all fields)
- **Aliasing**: Rename fields using `AS alias`
- **Functions**: Apply functions to fields (with optional parentheses)
  - `sum field` or `sum(field)`: Sum of numeric values in array
  - `len field` or `len(field)`: Length/count of array elements
  - `any field` or `any(field)`: True if any array elements are truthy
  - `all field` or `all(field)`: True if all array elements are truthy

### Examples

```sql
SELECT name, powerState AS status, memoryMB
SELECT sum(disks) AS totalDisks, len(networks) AS networkCount
```

## WHERE Clause

The WHERE clause filters items using Tree Search Language (TSL) expressions.

### Syntax (WHERE Clause)

```sql
WHERE condition
```

### Identifiers

- Start with a letter or underscore, may include letters, digits, `_`, `.`, `/`, `-`
- Array access: `[index]`, `[*]` (wildcard), `[name]` (map key)
- Examples: `name`, `user.age`, `pods[0].status`, `services[my.service].ip`

### Literals

- String: `'text'`, `"text"`, `` `text` ``
- Numeric: integer, decimal, scientific, with optional SI suffix (`Ki`, `M`, etc.)
- Date/Time: `YYYY-MM-DD` or RFC3339 `YYYY-MM-DDThh:mm:ssZ`
- Boolean: `true`, `false`
- Null: `null`
- Arrays: `[expr, expr, ...]`

### Operators

1. Logical
   - `AND`, `OR`, `NOT`
2. Comparison
   - `=`, `!=`, `<`, `<=`, `>`, `>=`
3. Pattern
   - `LIKE`, `ILIKE` (case‑insensitive), `~=` (regex match), `~!` (regex not match)
4. Membership
   - `IN`, `NOT IN`, `BETWEEN … AND …`
5. Arithmetic
   - `+`, `-`, `*`, `/`, `%`
6. Array functions
   - `LEN x`, `ANY x`, `ALL x`, `SUM x`

### Precedence (high→low)

1. Unary: `NOT`, `LEN`, `ANY`, `ALL`, `SUM`, unary `-`
2. `*`, `/`, `%`
3. `+`, `-`
4. `IN`, `BETWEEN`, `LIKE`, `ILIKE`, `IS`, etc.
5. Comparisons: `=`, `!=`, `<`, `<=`, `>`, `>=`, `~=`, `~!`
6. `AND`
7. `OR`

### Examples (WHERE Clause)

```sql
WHERE name LIKE '%web%' AND memoryMB > 4096
WHERE tags IN ['production', 'db'] OR createdAt > '2023-01-01'
WHERE ANY (disks[*].shared = true)
```

## ORDER BY Clause

The ORDER BY clause sorts results by one or more fields.

### Syntax

```sql
ORDER BY field1 [ASC|DESC], field2 [ASC|DESC]
```

### Features (ORDER BY Clause)

- **Multiple Fields**: Sort on multiple criteria with comma separation
- **Direction**: Sort ascending (ASC, default) or descending (DESC)
- **Field References**: Can use fields from SELECT clause aliases
- **Function Results**: Can sort by results of functions (sum, len, any, all)

### Examples (ORDER BY Clause)

```sql
ORDER BY memoryMB DESC
ORDER BY powerState ASC, cpuCount DESC
ORDER BY totalDisks DESC  # where totalDisks is an alias from SELECT
```

## LIMIT Clause

The LIMIT clause restricts the number of results returned.

### Syntax (LIMIT Clause)

```sql
LIMIT n
```

Where `n` is a positive integer.

### Examples (LIMIT Clause)

```sql
LIMIT 10
LIMIT 100
```

## Complete Examples

### Basic Filtering

```sql
WHERE powerState = 'poweredOn'
```

### Filtering with Sorting and Limit

```sql
WHERE memoryMB > 4096 ORDER BY cpuCount DESC LIMIT 5
```

### Field Selection with Filtering and Sorting

```sql
SELECT name, memoryGB AS memory, cpuCount WHERE powerState = 'poweredOn' ORDER BY memoryGB DESC
```

### Complex Query with Functions

```sql
SELECT name, 
       sum(disks[*].capacity) AS totalStorage, 
       len(networks) AS networkCount 
WHERE powerState = 'poweredOn' AND ANY (disks[*].capacity > 10Gi) 
ORDER BY totalStorage DESC 
LIMIT 10
```

### Using Aliases in ORDER BY

```sql
SELECT name, memoryMB / 1024 AS memoryGB 
WHERE powerState = 'poweredOn' 
ORDER BY memoryGB DESC
```

### Date Filtering

```sql
WHERE createdAt BETWEEN '2023-01-01' AND '2023-12-31'
ORDER BY createdAt DESC
```

## Working with Special Fields

### JSON Paths

You can access nested fields using dot notation and array indices:

```sql
SELECT metadata.name, spec.containers[0].image
WHERE metadata.labels['app'] = 'web'
```

### Arrays and Functions

For arrays, you can use special functions:

```sql
SELECT name, sum(disks[*].capacity) AS totalSize
WHERE len(networks) > 1 AND any(disks[*].shared = true)
```
