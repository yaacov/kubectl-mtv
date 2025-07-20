# kubectl-mtv E2E Tests

This directory contains end-to-end (e2e) tests for kubectl-mtv. The tests are designed to run against a live OpenShift/Kubernetes cluster with the Migration Toolkit for Virtualization (MTV) or Forklift installed.

## Test Architecture

**Shared Namespace Design**: All tests run in a shared namespace that is created once per test session and preserved for debugging. This approach provides:

- Faster test execution (no namespace creation/deletion overhead)
- Easier debugging (namespace remains for inspection)
- Automatic resource cleanup (resources are removed, namespace preserved)
- Unique resource naming to prevent conflicts

## Prerequisites

### Cluster Requirements

1. **OpenShift/Kubernetes Cluster**: You must be logged into an OpenShift or Kubernetes cluster with admin privileges
2. **MTV/Forklift Installed**: The cluster should have MTV (Migration Toolkit for Virtualization) or Forklift installed
3. **Admin Access**: Tests require cluster admin privileges to create namespaces and resources

### Local Requirements

1. **Python 3.8+**: Tests are written in Python using pytest
2. **kubectl-mtv Binary**: The kubectl-mtv binary must be available (built or in PATH)
3. **kubectl/oc CLI**: Must be logged into the target cluster

## Setup

### 1. Install Python Dependencies

```bash
cd tests/e2e
pip install -r requirements.txt
```

### 2. Build kubectl-mtv Binary

From the project root:

```bash
make
```

### 3. Configure Test Environment

Copy the environment template and customize it:

```bash
cp .env.template .env
# Edit .env with your provider credentials
```

### 4. Verify Cluster Access

Ensure you're logged into your OpenShift/Kubernetes cluster:

```bash
# For OpenShift
oc login https://api.your-cluster.com:6443

# For Kubernetes
kubectl config current-context

# Verify admin access
kubectl auth can-i '*' '*' --all-namespaces
```

### 5. Configure Inventory URL (Kubernetes only)

For Kubernetes clusters (non-OpenShift), you need to manually set the inventory service URL since Kubernetes doesn't support route discovery:

```bash
# Find the inventory service URL
kubectl get service -n konveyor-forklift forklift-inventory

# Set the environment variable
export MTV_INVENTORY_URL=http://<inventory-service-ip>:<port>

# Or add it to your .env file
echo "MTV_INVENTORY_URL=http://<inventory-service-ip>:<port>" >> .env
```

**Note**: OpenShift clusters can auto-discover the inventory URL via routes, so this step is typically not needed for OpenShift.

## Running Tests

### Using Makefile (Recommended)

The easiest way to run tests is using the provided Makefile:

```bash
# Run all tests
make test

# Run version command tests only
make test-version

# Run all provider tests
make test-providers

# Run specific provider type tests
make test-openshift      # OpenShift provider tests
make test-vsphere        # VMware vSphere provider tests  
make test-ovirt          # oVirt provider tests
make test-openstack      # OpenStack provider tests
make test-ova            # OVA provider tests

# Run error/edge case tests
make test-errors

# Run tests that don't require credentials
make test-no-creds

# Run tests in parallel
make test-fast

# Generate HTML and JSON reports
make test-report

# Run with coverage
make test-coverage
```

### Namespace Management

Tests use a shared namespace that is preserved for debugging:

```bash
# List current test namespaces
make list-ns

# Clean up test namespaces when done debugging
make cleanup

# Show test environment info
make info
```

### Using pytest directly

You can also run tests directly with pytest:

```bash
# Run all tests
pytest -v

# Run specific test files
pytest test_version.py -v
pytest test_provider_openshift.py -v
pytest test_provider_vsphere.py -v

# Run by markers
pytest -v -m version                    # Version tests
pytest -v -m provider                   # All provider tests
pytest -v -m "openshift"               # OpenShift provider tests
pytest -v -m "requires_credentials"     # Tests needing credentials
pytest -v -m "not requires_credentials" # Tests not needing credentials

# Debug failed tests by preserving namespace and resources
pytest --no-cleanup test_provider_openshift.py -v
```

### Run Tests with Different Output Formats

```bash
# Generate HTML report
pytest --html=report.html --self-contained-html

# Generate JSON report
pytest --json-report --json-report-file=report.json
```

### Test kubectl-mtv Output Formats

The version command supports multiple output formats:

```bash
# Test default output
./kubectl-mtv version

# Test JSON output  
./kubectl-mtv version -o json

# Test YAML output
./kubectl-mtv version -o yaml
```

### Run Tests in Parallel

```bash
# Run tests in parallel (requires pytest-xdist)
pytest -n auto

# Or using Makefile
make test-fast
```

### Environment-Specific Testing

```bash
# Run only tests that work without provider credentials
make test-no-creds

# Run tests for specific provider types (requires credentials)
make test-vsphere      # Requires VSPHERE_* env vars
make test-ovirt        # Requires OVIRT_* env vars  
make test-openstack    # Requires OPENSTACK_* env vars
make test-ova          # Requires OVA_URL env var

# OpenShift provider tests (usually work with current cluster)
make test-openshift
```

## Test Structure

### Test Files

The tests are organized into separate files for better maintainability and targeted testing:

1. **test_version.py**: Tests for the `kubectl mtv version` command
2. **test_provider_openshift.py**: OpenShift target provider tests
3. **test_provider_vsphere.py**: VMware vSphere provider tests  
4. **test_provider_ovirt.py**: oVirt provider tests
5. **test_provider_openstack.py**: OpenStack provider tests
6. **test_provider_ova.py**: OVA provider tests
7. **test_provider_errors.py**: Error conditions and edge case tests

### Test Markers

Tests are organized with pytest markers for easy filtering:

- `@pytest.mark.version` - Version command tests
- `@pytest.mark.provider` - All provider-related tests
- `@pytest.mark.openshift` - OpenShift provider tests
- `@pytest.mark.vsphere` - VMware vSphere provider tests
- `@pytest.mark.ovirt` - oVirt provider tests
- `@pytest.mark.openstack` - OpenStack provider tests  
- `@pytest.mark.ova` - OVA provider tests
- `@pytest.mark.requires_credentials` - Tests requiring provider credentials
- `@pytest.mark.error_cases` - Error condition and validation tests

### Test Fixtures

- **test_namespace**: Creates a temporary namespace for each test and cleans up afterwards
- **cluster_check**: Verifies cluster connectivity and admin privileges
- **provider_credentials**: Loads provider credentials from environment variables
- **kubectl_mtv_binary**: Locates the kubectl-mtv binary

## Environment Variables

The tests use environment variables to configure provider credentials. See `.env.template` for a complete list.

### Required for Each Provider Type

#### VMware vSphere

```bash
VSPHERE_URL=https://vcenter.example.com
VSPHERE_USERNAME=administrator@vsphere.local  
VSPHERE_PASSWORD=your-password
```

#### oVirt

```bash
OVIRT_URL=https://ovirt-engine.example.com/ovirt-engine/api
OVIRT_USERNAME=admin@internal
OVIRT_PASSWORD=your-password
```

#### OpenStack

```bash
OPENSTACK_URL=https://openstack.example.com:5000/v3
OPENSTACK_USERNAME=admin
OPENSTACK_PASSWORD=your-password
OPENSTACK_DOMAIN_NAME=Default
OPENSTACK_PROJECT_NAME=admin
```

#### OVA

```bash
OVA_URL=https://example.com/path/to/vm.ova
```

### Optional Variables

- Provider credentials not available will cause related tests to be skipped
- SSL certificates can be provided via `*_CACERT` variables
- Use `*_INSECURE_SKIP_TLS=true` for testing with self-signed certificates

#### MTV/Forklift Service Configuration

```bash
# Required for Kubernetes clusters (auto-discovered on OpenShift)
MTV_INVENTORY_URL=http://inventory-service-ip:port

# Optional - VDDK image for VMware providers
MTV_VDDK_INIT_IMAGE=registry.example.com/vddk-init:latest
```

**Important for Kubernetes**: Unlike OpenShift which can auto-discover services via routes, Kubernetes clusters require manual configuration of the `MTV_INVENTORY_URL` environment variable.

## Test Behavior

### Namespaces

- Each test gets its own temporary namespace (`kubectl-mtv-test-<random>`)
- Namespaces are automatically created before tests and cleaned up after
- For debugging failed tests, use `--no-cleanup` to preserve the test namespace and resources

### Debugging Failed Tests

When tests fail, you can preserve the test environment for debugging:

```bash
# Using pytest directly
pytest --no-cleanup test_provider_openshift.py -v

# Using make targets (easier)
make debug-test-openshift

# The test will print the namespace name for manual inspection
# Example output:
# === DEBUG MODE ===
# Test namespace: kubectl-mtv-test-a1b2c3d4
# Cleanup disabled - namespace will be preserved for debugging
# To manually cleanup later, run: kubectl delete namespace kubectl-mtv-test-a1b2c3d4
# ==================

# You can then inspect the namespace manually:
kubectl get all -n kubectl-mtv-test-a1b2c3d4
kubectl describe provider test-provider -n kubectl-mtv-test-a1b2c3d4

# When done debugging, cleanup manually:
kubectl delete namespace kubectl-mtv-test-a1b2c3d4
```

- All test resources are created within these temporary namespaces

### Resource Cleanup

- Tests track created resources and clean them up automatically
- Even if tests fail, cleanup is attempted in the teardown phase
- Namespaces are always deleted at the end of each test

### Credential Handling

- Tests check for required credentials and skip if not available
- No credentials are stored in the test code - only environment variables
- Supports both explicit credentials and current cluster context for OpenShift providers

## Writing New Tests

### Test File Organization

When adding new tests, follow the existing organization pattern:

- **Version/CLI tests**: Add to `test_version.py`
- **Provider-specific tests**: Add to appropriate `test_provider_<type>.py` file
- **Error/validation tests**: Add to `test_provider_errors.py`
- **New functionality**: Create new test files with descriptive names

### Test Class Structure

```python
@pytest.mark.your_marker
class TestNewFeature:
    """Test cases for new feature."""
    
    def test_basic_functionality(self, test_namespace):
        """Test basic feature functionality."""
        # Use test_namespace.run_mtv_command() to run kubectl-mtv commands
        result = test_namespace.run_mtv_command("your-command")
        assert result.returncode == 0
        
        # Track resources for cleanup
        test_namespace.track_resource("resource-type", "resource-name")
```

### Adding New Markers

When adding new test categories, remember to:

1. Add the marker to `pytest.ini`
2. Add corresponding Makefile targets if needed
3. Update this README with the new marker documentation

### Best Practices

1. **Use Descriptive Test Names**: Test names should clearly describe what is being tested
2. **Check Prerequisites**: Skip tests if required credentials/resources are not available
3. **Track Resources**: Always track created resources for cleanup
4. **Test Error Cases**: Include negative test cases for invalid inputs
5. **Use Assertions**: Make meaningful assertions about command output and behavior
6. **Wait for Async Operations**: Some operations may take time; add appropriate waits

## Troubleshooting

### Common Issues

1. **"Not logged into cluster"**: Ensure you're logged into the cluster with admin privileges
2. **"kubectl-mtv binary not found"**: Build the binary with `make` or ensure it's in PATH  
3. **"Provider credentials not available"**: Check your `.env` file configuration
4. **"Permission denied"**: Ensure you have cluster admin privileges
5. **"Inventory service not found" (Kubernetes)**: Set `MTV_INVENTORY_URL` environment variable manually since Kubernetes doesn't support route auto-discovery

### Debug Mode

Run tests with verbose output to see detailed command execution:

```bash
pytest -v -s
```

### Manual Testing

You can run individual kubectl-mtv commands manually to debug issues:

```bash
# Create a test namespace
kubectl create namespace test-debug

# Run kubectl-mtv commands manually
./kubectl-mtv version -n test-debug
./kubectl-mtv create provider test-provider --type openshift -n test-debug

# Cleanup
kubectl delete namespace test-debug
```

## Quick Start Examples

### Run Basic Tests (No Credentials Required)

```bash
# Setup environment
make setup

# Check prerequisites  
make check-cluster
make check-binary

# Run basic tests
make test-version
make test-openshift
make test-errors
```

### Run Provider Tests with Credentials

```bash
# Setup credentials in .env file
cp .env.template .env
# Edit .env with your provider credentials

# Run specific provider tests
make test-vsphere     # Tests VMware vSphere provider
make test-ovirt       # Tests oVirt provider  
make test-openstack   # Tests OpenStack provider
make test-ova         # Tests OVA provider
```

### Development Workflow

```bash
# Install dev dependencies
make dev-setup

# Run linting
make lint

# Run fast tests during development
make test-no-creds

# Run full test suite
make test-report

# View test results
open reports/report.html

# Debug failed tests (preserve environment)
make debug-test-openshift
# Then manually inspect: kubectl get all -n kubectl-mtv-test-XXXXXXXX
```
