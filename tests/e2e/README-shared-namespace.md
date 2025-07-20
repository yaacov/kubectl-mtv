# Shared Namespace Testing

This document describes the enhanced testing capabilities that allow tests to run in a shared namespace for improved test execution speed.

## Overview

By default, each test runs in its own isolated namespace, which is created and cleaned up for each test. While this provides excellent isolation, it can be slow when running many tests. The shared namespace feature allows all tests in a test collection to run in the same namespace, significantly speeding up test execution.

## Usage

### Default Behavior (Individual Namespaces)

```bash
# Each test gets its own namespace
pytest tests/e2e/
```

### Shared Namespace Mode

```bash
# All tests share the same namespace
pytest tests/e2e/ --shared-namespace
```

### Debugging with Shared Namespace

```bash
# Shared namespace with no cleanup for debugging
pytest tests/e2e/ --shared-namespace --no-cleanup
```

## How It Works

### Individual Namespace Mode (Default)
- Each test function gets its own unique namespace (e.g., `kubectl-mtv-test-abc12345`)
- Resources are tracked and cleaned up after each test
- Namespace is deleted after test completion
- Provides maximum isolation but slower execution

### Shared Namespace Mode
- All tests in the collection share a single namespace (e.g., `kubectl-mtv-shared-abc12345`)
- Resources are tracked but only cleaned up at the end of the test session
- Namespace is deleted after all tests complete
- Faster execution but requires careful resource naming to avoid conflicts

## Resource Naming in Shared Namespace

When using shared namespaces, it's important to ensure resource names don't conflict. The `utils.py` module provides a helper function:

```python
from utils import generate_unique_resource_name

# Generate unique names for shared namespace scenarios
def test_create_provider(test_namespace):
    provider_name = generate_unique_resource_name("test-provider", test_namespace.shared_namespace)
    # provider_name will be "test-provider" in individual mode
    # provider_name will be "test-provider-abc12345" in shared mode
```

## Benefits

### Shared Namespace Benefits
- **Faster execution**: No namespace creation/deletion overhead per test
- **Reduced API calls**: Less load on the Kubernetes API server
- **Efficient for large test suites**: Particularly beneficial when running many tests

### Individual Namespace Benefits
- **Complete isolation**: Tests cannot interfere with each other
- **Easier debugging**: Each test has its own clean environment
- **No naming conflicts**: Resource names can be simple and predictable

## Best Practices

1. **Use shared namespace for regression testing**: When running all tests for CI/CD
2. **Use individual namespaces for development**: When debugging specific tests
3. **Always use unique resource names**: When tests might run in shared namespace
4. **Clean up resources properly**: Use `test_namespace.track_resource()` for proper cleanup

## Examples

### Running specific test categories with shared namespace

```bash
# Run all provider tests in shared namespace
pytest tests/e2e/ -m provider --shared-namespace

# Run VSphere tests only in shared namespace
pytest tests/e2e/ -m vsphere --shared-namespace

# Run with debugging enabled
pytest tests/e2e/ -m provider --shared-namespace --no-cleanup -v
```

### Test code example

```python
@pytest.mark.provider
class TestProviderSharedNamespace:
    def test_create_vsphere_provider(self, test_namespace, provider_credentials):
        # Generate unique provider name for shared namespace
        provider_name = generate_unique_resource_name("test-vsphere-provider", test_namespace.shared_namespace)
        
        # Create provider...
        test_namespace.run_mtv_command(f"create provider {provider_name} ...")
        
        # Track for cleanup
        test_namespace.track_resource("provider", provider_name)
        
        # Verify provider...
        verify_provider_created(test_namespace, provider_name, "vsphere")
```

## Troubleshooting

### Common Issues

1. **Resource conflicts**: If tests fail due to resource naming conflicts, ensure you're using unique names
2. **State pollution**: Previous test state affecting current test - consider if shared namespace is appropriate
3. **Cleanup issues**: Resources not being cleaned up properly - check resource tracking

### Debug Commands

```bash
# List resources in shared namespace
kubectl get all -n kubectl-mtv-shared-abc12345

# Check specific resource types
kubectl get providers -n kubectl-mtv-shared-abc12345
kubectl get plans -n kubectl-mtv-shared-abc12345

# Manual cleanup if needed
kubectl delete namespace kubectl-mtv-shared-abc12345
```

## Implementation Details

The shared namespace feature is implemented through:

1. **Command line option**: `--shared-namespace` flag
2. **Enhanced TestContext**: Tracks both individual and session-level resources
3. **Smart fixture**: `test_namespace` fixture detects mode and behaves accordingly
4. **Session cleanup**: Resources are cleaned up at session end in shared mode
5. **Backwards compatibility**: Existing tests work without modification
