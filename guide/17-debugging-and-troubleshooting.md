---
layout: page
title: "Chapter 17: Debugging and Troubleshooting"
---

Effective troubleshooting is essential for successful VM migrations. This chapter covers comprehensive debugging techniques, common issues and their solutions, monitoring strategies, and systematic problem resolution approaches verified from the kubectl-mtv implementation and documentation.

## Overview: Systematic Debugging Approach

### Debugging Philosophy

Effective `kubectl-mtv` troubleshooting follows a systematic approach:

1. **Enable Appropriate Logging**: Use verbosity levels for targeted information
2. **Examine Multiple Sources**: Check kubectl-mtv output, Kubernetes resources, and platform logs
3. **Validate Configuration**: Verify providers, mappings, and plan settings
4. **Monitor Resource Health**: Check pod status, events, and resource constraints
5. **Isolate Variables**: Test individual components before complex workflows

### Debugging Tools and Techniques

- **Verbosity Levels**: Progressive information disclosure from silent to trace
- **Resource Description**: Detailed status and event information
- **Event Monitoring**: Real-time cluster event tracking
- **Log Analysis**: Pattern recognition in application and system logs
- **State Validation**: Verification of expected vs. actual resource states

## Enabling Debug Output

### Verbosity Levels (Verified from Implementation)

`kubectl-mtv` supports four verbosity levels through the `--verbose/-v` flag:

```bash
# Level 0: Silent (default) - Minimal output
kubectl mtv get providers

# Level 1: Info - Basic operational information
kubectl mtv get providers -v=1

# Level 2: Debug - Detailed operation tracking  
kubectl mtv get providers -v=2

# Level 3: Trace - Maximum detail including API calls
kubectl mtv get providers -v=3
```

### Strategic Debug Usage

#### Progressive Debugging

Start with basic levels and increase detail as needed:

```bash
# Initial investigation
kubectl mtv get plan problematic-migration -v=1

# Deeper analysis
kubectl mtv create plan debug-migration \
  --source problematic-provider \
  --vms test-vm \
  -v=2

# Maximum detail for complex issues
kubectl mtv start plan complex-migration -v=3
```

#### Targeted Debugging

Focus debugging on specific operations:

```bash
# Debug provider connectivity
kubectl mtv get inventory vm problematic-provider -v=2

# Debug mapping resolution
kubectl mtv create plan mapping-test \
  --network-mapping test-map \
  --storage-mapping test-storage \
  -v=2

# Debug plan execution
kubectl mtv start plan execution-debug -v=3
```

#### Debug Output Analysis

```bash
# Redirect debug output for analysis
kubectl mtv create plan analysis-target \
  --source provider-test \
  --vms debug-vm \
  -v=2 2>&1 | tee debug-output.log

# Filter specific debug information
kubectl mtv get providers -v=2 2>&1 | grep -i "error\|warn\|fail"

# Monitor real-time debug output
kubectl mtv get plan active-migration -w -v=2
```

### Environment Variable Debugging

Enable persistent debug settings:

```bash
# Set kubectl verbosity for all operations
export KUBECTL_VERBOSE=2

# Enable klog verbosity (affects underlying Kubernetes operations)
export KLOG_V=2

# Combine with kubectl-mtv operations
kubectl mtv get providers  # Will use enhanced verbosity
```

## Troubleshooting Common Issues

### Build/Installation Failures

#### Go Version Compatibility

```bash
# Check Go version (requires 1.24+)
go version

# Error: Go version too old
# Solution: Upgrade Go
curl -OL https://golang.org/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
export PATH=/usr/local/go/bin:$PATH
```

#### Build Dependencies

```bash
# Clean build environment
make clean

# Install missing dependencies
go mod download
go mod tidy

# Rebuild with debug information
make VERBOSE=1

# Check build errors
make 2>&1 | grep -i error
```

#### Static Binary Issues

```bash
# Verify static linking (should show "not a dynamic executable")
ldd kubectl-mtv

# Rebuild static binary if needed
CGO_ENABLED=0 go build -ldflags "-s -w" -o kubectl-mtv main.go

# Verify binary works
./kubectl-mtv version
```

#### Krew Installation Problems

```bash
# Check Krew installation
kubectl krew version

# Verify plugin installation
kubectl plugin list | grep mtv

# Reinstall if corrupted
kubectl krew uninstall mtv
kubectl krew install mtv

# Manual installation fallback
cp kubectl-mtv ~/.krew/bin/
kubectl mtv version
```

### Permission and Connection Issues

#### Kubernetes Cluster Access

```bash
# Verify cluster connectivity
kubectl cluster-info
kubectl auth whoami

# Test basic operations
kubectl get nodes
kubectl get namespaces

# Debug connection with verbose output
kubectl get pods -v=8
```

#### RBAC Permission Issues

```bash
# Check current permissions
kubectl auth can-i "*" "*"
kubectl auth can-i get providers.forklift.konveyor.io

# Test specific MTV operations
kubectl auth can-i list providers
kubectl auth can-i create plans
kubectl auth can-i get inventory

# Debug permission failures
kubectl mtv get providers -v=2  # Look for permission denied errors
```

#### Service Account Configuration

```bash
# Check service account
kubectl get serviceaccount default -o yaml

# Verify role bindings
kubectl get rolebindings,clusterrolebindings -o yaml | grep -A5 -B5 mtv

# Create admin access for troubleshooting
kubectl create clusterrolebinding mtv-admin \
  --clusterrole=cluster-admin \
  --serviceaccount=default:default

# Test with enhanced permissions
kubectl mtv get providers
```

#### Namespace Access Issues

```bash
# Check current namespace
kubectl config view --minify -o jsonpath='{..namespace}'

# List accessible namespaces
kubectl get namespaces

# Test namespace-specific operations
kubectl mtv get providers -n openshift-mtv
kubectl mtv get providers -A  # All namespaces

# Debug namespace resolution
kubectl mtv get providers -v=2  # Check namespace selection logic
```

### Convertor Pods Stuck in Pending State

#### Node Selector Constraint Issues

```bash
# Check convertor pod status
kubectl get pods -l forklift.app/plan=stuck-plan

# Describe pending convertor pods
kubectl describe pod convertor-pod-name | grep -A10 Events

# Verify node selector targets exist
kubectl get nodes -l node-type=migration-worker
kubectl get nodes --show-labels | grep migration

# Check node availability for scheduling
kubectl describe nodes | grep -A5 -B5 "Taints\|Unschedulable"
```

#### Resource Availability Problems

```bash
# Check node resource utilization
kubectl top nodes
kubectl describe nodes | grep -A5 "Allocated resources"

# Monitor resource requests vs capacity
kubectl get pods -o yaml | grep -A5 "requests\|limits"

# Check for resource quotas
kubectl get resourcequota --all-namespaces
kubectl describe resourcequota -n migration-namespace

# Validate storage requirements
kubectl get pv,pvc | grep convertor
kubectl describe pvc convertor-pvc-name
```

#### Affinity Rule Conflicts

```bash
# Debug affinity rule interpretation
kubectl describe pod convertor-pod | grep -A20 "Node-Selectors\|Affinity"

# Check KARL rule validation
kubectl mtv create plan test-affinity \
  --convertor-affinity "REQUIRE nodes(invalid-selector=true) on node" \
  -v=2

# Verify target pods for affinity rules exist
kubectl get pods -l app=storage  # For "REQUIRE pods(app=storage) on node"

# Check zone/region availability
kubectl get nodes --show-labels | grep topology.kubernetes.io
```

#### Convertor Resource Configuration

```bash
# Monitor convertor pod resource usage
kubectl top pod convertor-pod-name --containers

# Check resource limits and requests
kubectl describe pod convertor-pod-name | grep -A10 "Limits\|Requests"

# Validate storage class availability
kubectl get storageclass
kubectl describe storageclass fast-ssd  # Used in convertor optimization

# Check for competing resource consumers
kubectl get pods --sort-by='.spec.containers[0].resources.requests.memory'
```

### Mapping Issues (Source/Target Not Found)

#### Network Mapping Problems

```bash
# List available source networks
kubectl mtv get inventory networks source-provider

# Verify network mapping configuration
kubectl mtv describe mapping network network-map-name

# Test network mapping resolution
kubectl mtv create plan network-test \
  --source source-provider \
  --network-mapping problematic-network-map \
  --vms test-vm \
  -v=2

# Check target network availability
kubectl get network-attachment-definitions --all-namespaces
kubectl get networks.k8s.cni.cncf.io --all-namespaces
```

#### Storage Mapping Problems

```bash
# List available source storage
kubectl mtv get inventory storage source-provider

# Verify storage mapping configuration
kubectl mtv describe mapping storage storage-map-name

# Check storage class availability
kubectl get storageclass
kubectl describe storageclass target-storage-class

# Test storage mapping resolution
kubectl mtv create plan storage-test \
  --source source-provider \
  --storage-mapping problematic-storage-map \
  --vms test-vm \
  -v=2
```

#### Provider Inventory Issues

```bash
# Debug provider connectivity
kubectl mtv describe provider source-provider

# Test inventory access
kubectl mtv get inventory vm source-provider -v=2

# Check provider authentication
kubectl get secrets | grep provider
kubectl describe secret provider-secret-name

# Verify provider URL accessibility
kubectl run debug-pod --rm -it --image=curlimages/curl -- \
  curl -k https://provider-url/sdk

# Test inventory service availability
kubectl mtv get inventory hosts source-provider
```

### Provider-Specific Issues

#### VMware vSphere Problems

```bash
# Test vSphere connectivity
kubectl mtv describe provider vsphere-provider

# Debug VDDK configuration
kubectl mtv create vddk-image --tar /path/to/vddk.tar.gz --tag test:latest -v=2

# Check ESXi host connectivity
kubectl mtv describe host esxi-host-01

# Validate migration host configuration
kubectl get migrationhosts --all-namespaces
kubectl describe migrationhost host-name

# Test vCenter API access
kubectl run vsphere-debug --rm -it --image=vmware/govc -- \
  govc about -u 'user:pass@vcenter-url'
```

#### OpenStack Provider Issues

```bash
# Debug OpenStack authentication
kubectl mtv describe provider openstack-provider

# Test OpenStack API connectivity
kubectl run openstack-debug --rm -it --image=openstacktools/python-openstackclient -- \
  openstack --os-auth-url https://keystone-url server list

# Check domain/project configuration
kubectl get secrets openstack-provider-secret -o yaml | base64 -d
```

#### oVirt/RHV Provider Issues

```bash
# Test oVirt Engine connectivity
kubectl mtv describe provider ovirt-provider

# Debug oVirt API access
kubectl run ovirt-debug --rm -it --image=ovirt/python-sdk -- \
  python3 -c "from ovirtsdk4 import Connection; print('oVirt SDK available')"

# Check certificate issues
kubectl get secret ovirt-provider-secret -o yaml | grep ca.crt | base64 -d
```

## Monitoring Techniques

### Describing Resources

#### Comprehensive Resource Description

```bash
# Detailed plan analysis
kubectl mtv describe plan problem-plan --with-vms

# Provider status investigation
kubectl mtv describe provider failing-provider

# VM-specific migration status
kubectl mtv describe plan migration-plan --vm stuck-vm

# Mapping configuration verification
kubectl mtv describe mapping network network-mapping
kubectl mtv describe mapping storage storage-mapping
```

#### Resource Status Monitoring

```bash
# Monitor plan execution progress
kubectl mtv get plan active-migration -w

# Check VM migration status
kubectl mtv get plan active-migration --vms

# Monitor provider health
kubectl mtv get providers -o yaml | grep -A5 -B5 status

# Resource dependency tracking
kubectl get providers,plans,mappings,hosts --all-namespaces
```

### Checking Kubernetes Events

#### Event-Based Troubleshooting

```bash
# Get recent events sorted by time
kubectl get events --sort-by='.metadata.creationTimestamp' -A

# Filter MTV-related events
kubectl get events -A | grep -i "forklift\|mtv\|migration"

# Monitor events in real-time
kubectl get events -w --all-namespaces

# Specific resource events
kubectl describe plan problem-plan | grep -A20 Events
kubectl describe pod convertor-pod | grep -A20 Events
```

#### Event Analysis Patterns

```bash
# Look for scheduling failures
kubectl get events -A | grep -i "FailedScheduling\|Pending"

# Identify resource constraints
kubectl get events -A | grep -i "Insufficient\|OutOf"

# Find permission issues
kubectl get events -A | grep -i "Forbidden\|Unauthorized"

# Monitor hook execution
kubectl get events -A | grep -i "hook\|job"
```

### Log Analysis

#### Application Log Analysis

```bash
# MTV Forklift controller logs
kubectl logs -n openshift-mtv deployment/forklift-controller

# Inventory service logs
kubectl logs -n openshift-mtv deployment/forklift-inventory

# Validation service logs
kubectl logs -n openshift-mtv deployment/forklift-validation

# Follow logs in real-time
kubectl logs -f -n openshift-mtv deployment/forklift-controller
```

#### Migration-Specific Logs

```bash
# Convertor pod logs
kubectl logs convertor-pod-name -c virt-v2v

# Hook execution logs
kubectl logs hook-job-pod -c hook-runner

# VM import logs
kubectl logs vm-import-pod -c vm-import

# CDI pod logs (storage)
kubectl logs -l app=containerized-data-importer
```

### Systematic Problem Resolution

#### Problem Isolation Workflow

```bash
#!/bin/bash
# troubleshoot-migration.sh - Systematic migration troubleshooting

PLAN_NAME="$1"
NAMESPACE="${2:-default}"

if [ -z "$PLAN_NAME" ]; then
  echo "Usage: $0 <plan-name> [namespace]"
  exit 1
fi

echo "=== Troubleshooting Plan: $PLAN_NAME ==="
echo

# 1. Check plan status
echo "1. Plan Status:"
kubectl mtv get plan "$PLAN_NAME" -n "$NAMESPACE"
echo

# 2. Describe plan for detailed information
echo "2. Plan Details:"
kubectl mtv describe plan "$PLAN_NAME" -n "$NAMESPACE" | head -50
echo

# 3. Check provider status
echo "3. Provider Status:"
PROVIDERS=$(kubectl get plan "$PLAN_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.provider.source.name},{.spec.provider.destination.name}')
IFS=',' read -r SOURCE_PROVIDER DEST_PROVIDER <<< "$PROVIDERS"

if [ -n "$SOURCE_PROVIDER" ]; then
  echo "Source Provider: $SOURCE_PROVIDER"
  kubectl mtv get provider "$SOURCE_PROVIDER" -n "$NAMESPACE"
fi

if [ -n "$DEST_PROVIDER" ]; then
  echo "Destination Provider: $DEST_PROVIDER"
  kubectl mtv get provider "$DEST_PROVIDER" -n "$NAMESPACE"
fi
echo

# 4. Check for related pods
echo "4. Related Pods:"
kubectl get pods -n "$NAMESPACE" -l forklift.app/plan="$PLAN_NAME"
echo

# 5. Recent events
echo "5. Recent Events:"
kubectl get events -n "$NAMESPACE" --sort-by='.metadata.creationTimestamp' | tail -20
echo

# 6. Resource utilization
echo "6. Node Resources:"
kubectl top nodes | head -5
```

#### Debug Configuration Validation

```bash
#!/bin/bash
# validate-mtv-config.sh - Validate MTV configuration

echo "=== MTV Configuration Validation ==="
echo

# Check MTV installation
echo "1. MTV Installation Status:"
kubectl get crds | grep forklift.konveyor.io || echo "ERROR: MTV CRDs not found"
kubectl get deployments -A | grep forklift || echo "ERROR: Forklift deployments not found"
echo

# Check RBAC
echo "2. RBAC Validation:"
kubectl auth can-i list providers.forklift.konveyor.io || echo "ERROR: No provider access"
kubectl auth can-i create plans.forklift.konveyor.io || echo "ERROR: No plan creation access"
echo

# Check providers
echo "3. Provider Status:"
kubectl mtv get providers --all-namespaces || echo "ERROR: Cannot list providers"
echo

# Check inventory access
echo "4. Inventory Service:"
kubectl get routes -A | grep inventory || kubectl get services -A | grep inventory || echo "WARNING: Inventory service not found"
echo

# Check storage classes
echo "5. Available Storage Classes:"
kubectl get storageclass || echo "ERROR: No storage classes available"
echo

# Check network configurations
echo "6. Network Resources:"
kubectl get network-attachment-definitions -A | head -5 || echo "INFO: No Multus networks found"
kubectl get networks.k8s.cni.cncf.io -A | head -5 || echo "INFO: No CNI networks found"
```

#### Performance Monitoring

```bash
# Monitor migration performance
kubectl top pods -l forklift.app/plan=performance-test --containers

# Check I/O performance
kubectl exec convertor-pod -- iostat -x 1 5

# Monitor network throughput
kubectl exec convertor-pod -- ss -tuln | grep :443

# Check resource constraints
kubectl describe nodes | grep -A5 -B5 "Pressure\|OutOf"
```

## Advanced Debugging Techniques

### Debug Plan Creation

```bash
# Create minimal test plan
kubectl mtv create plan debug-minimal \
  --source test-provider \
  --vms simple-vm \
  -v=3 \
  --dry-run  # If supported

# Validate individual components
kubectl mtv get inventory vm test-provider simple-vm -v=2
kubectl mtv describe provider test-provider
kubectl mtv get mappings --all-namespaces
```

### Debug Migration Execution

```bash
# Monitor migration start
kubectl mtv start plan debug-migration -v=2

# Watch resource creation
kubectl get pods,jobs,pvcs -l forklift.app/plan=debug-migration -w

# Check controller logs during execution
kubectl logs -f -n openshift-mtv deployment/forklift-controller | grep debug-migration
```

### Debug Hook Execution

```bash
# Monitor hook job creation
kubectl get jobs -n openshift-mtv -w | grep hook

# Check hook pod logs
kubectl logs hook-job-pod -c hook-runner

# Validate hook configuration
kubectl describe hook hook-name -n openshift-mtv

# Debug hook context files
kubectl exec hook-pod -- cat /tmp/hook/plan.yml
kubectl exec hook-pod -- cat /tmp/hook/workload.yml
```

## Common Error Patterns and Solutions

### Error: "Provider not found"

```bash
# Diagnosis:
kubectl mtv get providers --all-namespaces | grep provider-name

# Solution:
# 1. Verify provider exists in correct namespace
# 2. Check provider name spelling
# 3. Ensure RBAC access to provider namespace
```

### Error: "Insufficient resources"

```bash
# Diagnosis:
kubectl describe nodes | grep -A10 "Allocated resources"
kubectl get events | grep -i insufficient

# Solution:
# 1. Add more worker nodes
# 2. Adjust convertor resource requirements  
# 3. Use node selector to target larger nodes
```

### Error: "Image pull backoff"

```bash
# Diagnosis:
kubectl describe pod failing-pod | grep -A5 "Failed to pull image"

# Solution:
# 1. Check image registry accessibility
# 2. Verify image tag exists
# 3. Check registry authentication
# 4. Use image pull secrets if needed
```

### Error: "Storage class not found"

```bash
# Diagnosis:
kubectl get storageclass
kubectl mtv describe mapping storage mapping-name

# Solution:
# 1. Create missing storage class
# 2. Update storage mapping
# 3. Use default storage class
```

## Next Steps

After mastering debugging and troubleshooting:

1. **Best Practices**: Learn operational excellence in [Chapter 18: Best Practices and Security](18-best-practices-and-security)
2. **AI Integration**: Explore advanced automation in [Chapter 19: Model Context Protocol (MCP) Server Integration](19-model-context-protocol-mcp-server-integration)
3. **Tool Integration**: Learn KubeVirt ecosystem integration in [Chapter 20: Integration with KubeVirt Tools](20-integration-with-kubevirt-tools)

---

*Previous: [Chapter 16: Plan Lifecycle Execution](/guide/16-plan-lifecycle-execution)*  
*Next: [Chapter 18: Best Practices and Security](/guide/18-best-practices-and-security)*
