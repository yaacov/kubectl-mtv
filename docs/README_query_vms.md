# Using Query Strings to Select VMs for Plan Creation

The `create plan` command supports three methods for specifying which VMs to include in a migration plan:

1. **Comma-separated list**: Explicit VM names
2. **File**: YAML/JSON file containing VM definitions (prefix with `@`)
3. **Query string**: Dynamic VM selection using query filters (prefix with `where `)

## Query String Method

When you prefix the `--vms` flag value with `where `, the command will:

1. Fetch the VM inventory from the specified source provider
2. Apply the query filter to select matching VMs
3. Use the filtered VMs to create the migration plan

### Syntax

```bash
kubectl mtv create plan PLAN_NAME \
  --source SOURCE_PROVIDER \
  --vms "where QUERY_EXPRESSION"
```

### Query Language

The query language is the same as used by the `get inventory vms` command. It supports:

- **Filtering**: `where name like 'test%'`
- **Logical operators**: `where name like 'test%' and powerState = 'On'`
- **Comparison operators**: `=`, `!=`, `<`, `>`, `<=`, `>=`, `like`, `not like`
- **Sorting**: `where name like 'prod%' order by name`
- **Limiting**: `where powerState = 'Off' limit 5`

### Examples

#### Example 1: Select VMs by name pattern

Select all VMs whose names start with "test":

```bash
kubectl mtv create plan test-migration \
  --source vsphere-provider \
  --vms "where name like 'test%'"
```

#### Example 2: Select powered-off VMs

Select only VMs that are powered off:

```bash
kubectl mtv create plan offline-migration \
  --source ovirt-provider \
  --vms "where powerState = 'Off'"
```

#### Example 3: Complex query with multiple conditions

Select VMs with specific criteria and sort them:

```bash
kubectl mtv create plan production-migration \
  --source vsphere-provider \
  --vms "where name like 'prod%' and cpuCount >= 4 and memoryMB > 8192 order by name"
```

#### Example 4: Limited selection

Select only the first 10 VMs matching criteria:

```bash
kubectl mtv create plan batch-migration \
  --source ovirt-provider \
  --vms "where name like 'app%' limit 10"
```

#### Example 5: VMs with concerns

Select VMs that have critical migration concerns:

```bash
kubectl mtv create plan critical-vms \
  --source vsphere-provider \
  --vms "where criticalConcerns > 0"
```

### Testing Your Query

Before creating a plan, you can test your query using the `get inventory vms` command to see which VMs would be selected:

```bash
# Test the query first
kubectl mtv get inventory vms SOURCE_PROVIDER -q "where name like 'test%'"

# Once satisfied, use the same query in create plan
kubectl mtv create plan test-migration \
  --source SOURCE_PROVIDER \
  --vms "where name like 'test%'"
```

### Available VM Fields for Queries

Common fields you can use in queries:

- `name`: VM name
- `id`: VM identifier
- `powerState`: Power state (On, Off, etc.)
- `cpuCount`: Number of CPUs
- `memoryMB`: Memory in megabytes
- `memoryGB`: Memory in gigabytes (calculated)
- `guestId`: Guest OS identifier
- `criticalConcerns`: Number of critical concerns
- `warningConcerns`: Number of warning concerns
- `infoConcerns`: Number of informational concerns
- `storageUsed`: Storage used in bytes
- `diskCapacity`: Total disk capacity

### Notes

- The query must start with `where ` (case-sensitive)
- Query syntax is validated **before** fetching inventory (fast failure for syntax errors)
- If no VMs match the query, the plan creation will fail with an error
- The command will print the number of VMs found matching the query before creating the plan
- VM IDs are automatically populated from the inventory
- The source provider must be specified with the `--source` flag when using queries

### Comparison with Other Methods

#### Method 1: Comma-separated list
```bash
kubectl mtv create plan my-plan --source provider --vms "vm1,vm2,vm3"
```

#### Method 2: File
```bash
kubectl mtv create plan my-plan --source provider --vms "@vms.yaml"
```

#### Method 3: Query (NEW)
```bash
kubectl mtv create plan my-plan --source provider --vms "where name like 'app%'"
```

