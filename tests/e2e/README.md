# kubectl-mtv E2E Tests

This directory contains end-to-end (e2e) tests for kubectl-mtv. The tests are designed to run against a live OpenShift/Kubernetes cluster with the Migration Toolkit for Virtualization (MTV) or Forklift installed.

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

### Run All Tests

```bash
pytest -v
```

### Run Specific Test Categories

```bash
# Test only the version command
pytest test_version.py -v

# Test only provider creation
pytest test_providers.py -v
```

### Run Tests with Different Output Formats

```bash
# Generate HTML report
pytest --html=report.html --self-contained-html

# Generate JSON report
pytest --json-report --json-report-file=report.json
```

### Run Tests in Parallel

```bash
# Run tests in parallel (requires pytest-xdist)
pytest -n auto
```

## Test Structure

### Test Categories

1. **test_version.py**: Tests for the `kubectl mtv version` command
2. **test_providers.py**: Tests for provider creation, listing, and management

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

### Namespace Management

- Each test gets its own temporary namespace (`kubectl-mtv-test-<random>`)
- Namespaces are automatically created before tests and cleaned up after
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

### Test Class Structure

```python
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
