---
layout: page
title: "Chapter 4: Provider Management"
---

# Chapter 4: Provider Management

Providers are the core resources in kubectl-mtv that represent source and target virtualization platforms. This chapter covers complete provider lifecycle management including creation, configuration, patching, and deletion.

For detailed information about provider prerequisites and requirements, see the [official Forklift provider documentation](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#prerequisites-and-software-requirements-for-all-providers).

## Overview of Providers

Providers define the connection and authentication details for virtualization platforms. Each migration requires:
- **Source Provider**: Your current virtualization platform (vSphere, oVirt, OpenStack, KubeVirt, or OVA)
- **Target Provider**: Your destination KubeVirt cluster (typically OpenShift/Kubernetes)

### Supported Provider Types

kubectl-mtv supports the following provider types:

| Provider Type | Description | Use Case |
|---------------|-------------|----------|
| `openshift` | OpenShift/Kubernetes clusters | Target provider or KubeVirt-to-KubeVirt migrations |
| `vsphere` | VMware vSphere/vCenter | Source provider for VMware environments |
| `ovirt` | oVirt/Red Hat Virtualization | Source provider for oVirt/RHV environments |
| `openstack` | OpenStack platforms | Source provider for OpenStack environments |
| `ova` | OVA/OVF files | Source provider for VM image files |

## Listing, Describing, and Deleting Providers

### List Providers

```bash
# List all providers in current namespace
kubectl mtv get providers

# List providers in specific namespace
kubectl mtv get providers -n forklift-namespace

# List providers across all namespaces
kubectl mtv get providers --all-namespaces

# List with detailed output
kubectl mtv get providers -o yaml
kubectl mtv get providers -o json

# List specific provider
kubectl mtv get provider my-vsphere-provider
```

### Describe Providers

Get detailed information about a specific provider:

```bash
# Describe a provider
kubectl mtv describe provider my-vsphere-provider

# Describe with additional inventory information
kubectl mtv get inventory provider my-vsphere-provider
```

### Delete Providers

```bash
# Delete a specific provider
kubectl mtv delete provider my-vsphere-provider

# Delete multiple providers
kubectl mtv delete provider provider1 provider2 provider3

# Delete all providers in namespace (use with caution)
kubectl mtv delete provider --all

# Alternative plural form
kubectl mtv delete providers my-vsphere-provider
```

## How-To: Creating Providers

### VMware vSphere Provider

Create providers for VMware vCenter or ESXi environments.

#### Basic vSphere Provider

```bash
# Basic vSphere provider with vCenter
kubectl mtv create provider vsphere-prod --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password YourSecurePassword

# vSphere provider with ESXi endpoint
kubectl mtv create provider esxi-host --type vsphere \
  --url https://esxi-host.example.com/sdk \
  --username root \
  --password YourSecurePassword \
  --sdk-endpoint esxi
```

#### vSphere Provider with VDDK Optimization

```bash
# vSphere provider with VDDK image (recommended for performance)
kubectl mtv create provider vsphere-prod --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password YourSecurePassword \
  --vddk-init-image quay.io/your-registry/vddk:8.0.1

# With advanced VDDK optimization
kubectl mtv create provider vsphere-prod --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password YourSecurePassword \
  --vddk-init-image quay.io/your-registry/vddk:8.0.1 \
  --use-vddk-aio-optimization \
  --vddk-buf-size-in-64k 64 \
  --vddk-buf-count 8
```

#### vSphere Provider with Custom CA Certificate

```bash
# With CA certificate from file
kubectl mtv create provider vsphere-prod --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password YourSecurePassword \
  --cacert @/path/to/ca-certificate.pem

# With inline CA certificate
kubectl mtv create provider vsphere-prod --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password YourSecurePassword \
  --cacert "-----BEGIN CERTIFICATE-----
MIIGBzCCA++gAwIBAgIJAKt..."

# Skip TLS verification (not recommended for production)
kubectl mtv create provider vsphere-test --type vsphere \
  --url https://vcenter.test.com/sdk \
  --username administrator@vsphere.local \
  --password YourSecurePassword \
  --provider-insecure-skip-tls
```

#### vSphere Provider with Existing Secret

```bash
# Use existing secret for credentials
kubectl mtv create provider vsphere-prod --type vsphere \
  --url https://vcenter.example.com/sdk \
  --secret existing-vcenter-secret \
  --vddk-init-image quay.io/your-registry/vddk:8.0.1
```

### oVirt/RHV Provider

Create providers for oVirt and Red Hat Virtualization environments.

```bash
# Basic oVirt provider
kubectl mtv create provider ovirt-prod --type ovirt \
  --url https://ovirt-engine.example.com/ovirt-engine/api \
  --username admin@internal \
  --password YourSecurePassword

# oVirt provider with CA certificate
kubectl mtv create provider ovirt-prod --type ovirt \
  --url https://ovirt-engine.example.com/ovirt-engine/api \
  --username admin@internal \
  --password YourSecurePassword \
  --cacert @/path/to/ovirt-ca.pem

# oVirt provider with existing secret
kubectl mtv create provider ovirt-prod --type ovirt \
  --url https://ovirt-engine.example.com/ovirt-engine/api \
  --secret ovirt-credentials-secret
```

### OpenStack Provider

Create providers for OpenStack environments with required project and domain information.

```bash
# Basic OpenStack provider
kubectl mtv create provider openstack-prod --type openstack \
  --url https://openstack.example.com:5000/v3 \
  --username admin \
  --password YourSecurePassword \
  --provider-domain-name Default \
  --provider-project-name admin

# OpenStack provider with region
kubectl mtv create provider openstack-west --type openstack \
  --url https://openstack-west.example.com:5000/v3 \
  --username myuser \
  --password YourSecurePassword \
  --provider-domain-name MyDomain \
  --provider-project-name MyProject \
  --provider-region-name us-west-1

# OpenStack provider with CA certificate
kubectl mtv create provider openstack-prod --type openstack \
  --url https://openstack.example.com:5000/v3 \
  --username admin \
  --password YourSecurePassword \
  --provider-domain-name Default \
  --provider-project-name admin \
  --cacert @/path/to/openstack-ca.pem
```

### OpenShift/Kubernetes (Target) Provider

Create target providers for OpenShift or Kubernetes clusters.

```bash
# Local OpenShift cluster (current context)
kubectl mtv create provider local-openshift --type openshift

# Remote OpenShift cluster with token
kubectl mtv create provider remote-openshift --type openshift \
  --url https://api.remote-cluster.example.com:6443 \
  --token your-service-account-token

# Remote OpenShift cluster with CA certificate
kubectl mtv create provider remote-openshift --type openshift \
  --url https://api.remote-cluster.example.com:6443 \
  --token your-service-account-token \
  --cacert @/path/to/cluster-ca.pem

# Skip TLS verification for testing
kubectl mtv create provider test-openshift --type openshift \
  --url https://api.test-cluster.example.com:6443 \
  --token your-service-account-token \
  --provider-insecure-skip-tls
```

### OVA Provider

Create providers for OVA/OVF file imports from NFS shares.

**Note**: OVA providers only support NFS URLs in the format `nfs_server:nfs_path` where OVA files are stored on an NFS share.

```bash
# OVA provider from NFS share
kubectl mtv create provider my-ova --type ova \
  --url nfs.example.com:/path/to/ova-files

# OVA provider with IP address
kubectl mtv create provider datacenter-ova --type ova \
  --url 192.168.1.100:/exports/vm-images
```

### Using Environment Variables

You can use the `MTV_VDDK_INIT_IMAGE` environment variable to set a default VDDK image:

```bash
# Set default VDDK image
export MTV_VDDK_INIT_IMAGE=quay.io/your-registry/vddk:8.0.1

# Create vSphere provider (will use the environment variable)
kubectl mtv create provider vsphere-prod --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password YourSecurePassword
```

## How-To: Patching Providers

Provider patching allows you to update settings of existing providers without recreating them. This is particularly useful for updating credentials, URLs, or VDDK settings.

### Understanding Secret Ownership and Protection

kubectl-mtv implements secret ownership protection to prevent accidental modification of shared secrets:

- **Owned Secrets**: Created automatically by kubectl-mtv and owned by the provider
- **Shared Secrets**: Created independently and potentially used by multiple providers

#### Owned vs. Shared Secrets

```bash
# When you create a provider with credentials, kubectl-mtv creates an owned secret
kubectl mtv create provider vsphere-prod --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username admin \
  --password secret123
# Creates: owned secret that can be patched

# When you create a provider with existing secret, it becomes shared
kubectl create secret generic shared-creds --from-literal=user=admin --from-literal=password=secret123
kubectl mtv create provider vsphere-prod --type vsphere \
  --url https://vcenter.example.com/sdk \
  --secret shared-creds
# Uses: shared secret that cannot be patched through provider commands
```

### Updating Provider Settings

#### Update URL and Connection Settings

```bash
# Update provider URL
kubectl mtv patch provider vsphere-prod \
  --url https://new-vcenter.example.com/sdk

# Update URL and disable TLS verification
kubectl mtv patch provider vsphere-test \
  --url https://test-vcenter.internal/sdk \
  --provider-insecure-skip-tls

# Update token for OpenShift provider
kubectl mtv patch provider remote-openshift \
  --token new-service-account-token
```

#### Update Credentials (Owned Secrets Only)

```bash
# Update username and password
kubectl mtv patch provider vsphere-prod \
  --username new-admin@vsphere.local \
  --password NewSecurePassword

# Update only password
kubectl mtv patch provider vsphere-prod \
  --password UpdatedPassword

# Update OpenStack domain and project
kubectl mtv patch provider openstack-prod \
  --provider-domain-name NewDomain \
  --provider-project-name NewProject
```

#### Update CA Certificates

```bash
# Update CA certificate from file
kubectl mtv patch provider vsphere-prod \
  --cacert @/path/to/new-ca-certificate.pem

# Remove CA certificate (empty string)
kubectl mtv patch provider vsphere-prod \
  --cacert ""

# Add CA certificate where none existed
kubectl mtv patch provider vsphere-prod \
  --cacert @/path/to/ca-certificate.pem
```

#### Update VDDK Settings

```bash
# Update VDDK init image
kubectl mtv patch provider vsphere-prod \
  --vddk-init-image quay.io/your-registry/vddk:8.0.2

# Enable VDDK AIO optimization
kubectl mtv patch provider vsphere-prod \
  --use-vddk-aio-optimization

# Update VDDK buffer settings
kubectl mtv patch provider vsphere-prod \
  --vddk-buf-size-in-64k 128 \
  --vddk-buf-count 16

# Disable VDDK AIO optimization
kubectl mtv patch provider vsphere-prod \
  --use-vddk-aio-optimization=false
```

### Handling Shared Secrets

When you encounter a shared secret error:

```bash
# Error message example:
# Error: cannot patch credentials for provider 'vsphere-prod': secret 'shared-creds' is not owned by this provider.
# This usually means the secret was created independently and is shared by multiple providers.
# To update credentials, either:
# 1. Update the secret directly: kubectl patch secret shared-creds -p '{...}'
# 2. Create a new secret and update the provider to use it: kubectl patch provider vsphere-prod --secret new-secret-name
```

#### Solution 1: Update the Secret Directly

```bash
# Update shared secret directly
kubectl patch secret shared-creds -p '{"data":{"user":"'$(echo -n "new-username" | base64)'","password":"'$(echo -n "new-password" | base64)'"}}'

# Or using kubectl create secret with --dry-run and --save-config
kubectl create secret generic shared-creds \
  --from-literal=user=new-username \
  --from-literal=password=new-password \
  --dry-run=client -o yaml | kubectl apply -f -
```

#### Solution 2: Create New Secret and Switch

```bash
# Create new secret
kubectl create secret generic new-vsphere-creds \
  --from-literal=user=new-username \
  --from-literal=password=new-password

# Update provider to use new secret (this requires recreating the provider or manual editing)
kubectl get provider vsphere-prod -o yaml > provider-backup.yaml
# Edit the YAML to reference the new secret, then:
kubectl apply -f provider-backup.yaml
```

### Verification and Monitoring

```bash
# Verify provider status after patching
kubectl mtv get provider vsphere-prod

# Check provider details
kubectl mtv describe provider vsphere-prod

# Test provider connectivity (through inventory)
kubectl mtv get inventory provider vsphere-prod
```

## Provider Configuration Examples

### Complete vSphere Production Setup

```bash
# 1. Create optimized vSphere provider
kubectl mtv create provider vsphere-production --type vsphere \
  --url https://vcenter.prod.company.com/sdk \
  --username svc-migration@vsphere.local \
  --password $(cat /secure/vsphere-password) \
  --cacert @/etc/ssl/certs/vcenter-ca.pem \
  --vddk-init-image quay.io/company/vddk:8.0.1 \
  --use-vddk-aio-optimization \
  --vddk-buf-size-in-64k 64 \
  --vddk-buf-count 8

# 2. Create target OpenShift provider
kubectl mtv create provider openshift-target --type openshift

# 3. Verify providers
kubectl mtv get providers
kubectl mtv describe provider vsphere-production
```

### Multi-Region OpenStack Setup

```bash
# West region provider
kubectl mtv create provider openstack-west --type openstack \
  --url https://west.openstack.company.com:5000/v3 \
  --username migration-user \
  --password SecurePassword123 \
  --provider-domain-name Production \
  --provider-project-name Migration \
  --provider-region-name us-west-2

# East region provider  
kubectl mtv create provider openstack-east --type openstack \
  --url https://east.openstack.company.com:5000/v3 \
  --username migration-user \
  --password SecurePassword123 \
  --provider-domain-name Production \
  --provider-project-name Migration \
  --provider-region-name us-east-1
```

### Development/Testing Setup

```bash
# Test vSphere provider with TLS disabled
kubectl mtv create provider vsphere-dev --type vsphere \
  --url https://vcenter-dev.internal/sdk \
  --username administrator@vsphere.local \
  --password DevPassword123 \
  --provider-insecure-skip-tls

# Local test OpenShift
kubectl mtv create provider openshift-dev --type openshift
```

## Best Practices for Provider Management

### Security Best Practices

1. **Use Strong Credentials**:
   ```bash
   # Generate secure passwords
   openssl rand -base64 32
   
   # Use service accounts with minimal required permissions
   kubectl mtv create provider vsphere-prod --type vsphere \
     --username svc-migration@vsphere.local
   ```

2. **Certificate Validation**:
   ```bash
   # Always use CA certificates in production
   kubectl mtv create provider vsphere-prod --type vsphere \
     --cacert @/path/to/ca-cert.pem
   
   # Only skip TLS for development
   kubectl mtv create provider vsphere-dev --type vsphere \
     --provider-insecure-skip-tls
   ```

3. **Secret Management**:
   ```bash
   # Use external secret management when possible
   kubectl create secret generic vsphere-creds \
     --from-file=user=<(echo -n "username") \
     --from-file=password=<(echo -n "password")
   ```

### Performance Optimization

1. **VDDK Configuration**:
   ```bash
   # Use VDDK for VMware (significant performance improvement)
   kubectl mtv create provider vsphere-prod --type vsphere \
     --vddk-init-image quay.io/your-registry/vddk:latest \
     --use-vddk-aio-optimization \
     --vddk-buf-size-in-64k 64 \
     --vddk-buf-count 8
   ```

2. **Endpoint Selection**:
   ```bash
   # Use ESXi direct connection for better performance
   kubectl mtv create provider esxi-direct --type vsphere \
     --sdk-endpoint esxi \
     --url https://esxi-host.example.com/sdk
   ```

### Naming and Organization

1. **Consistent Naming**:
   ```bash
   # Environment-specific naming
   kubectl mtv create provider vsphere-prod --type vsphere
   kubectl mtv create provider vsphere-dev --type vsphere
   kubectl mtv create provider vsphere-test --type vsphere
   
   # Location-specific naming  
   kubectl mtv create provider openstack-west --type openstack
   kubectl mtv create provider openstack-east --type openstack
   ```

2. **Namespace Organization**:
   ```bash
   # Separate namespaces for different environments
   kubectl create namespace migration-prod
   kubectl create namespace migration-dev
   
   # Create providers in appropriate namespaces
   kubectl mtv create provider vsphere-prod --type vsphere -n migration-prod
   kubectl mtv create provider vsphere-dev --type vsphere -n migration-dev
   ```

## Troubleshooting Provider Issues

### Common Provider Problems

#### Connection Issues

```bash
# Test provider connectivity
kubectl mtv get inventory provider vsphere-prod

# Check provider status
kubectl mtv describe provider vsphere-prod

# Enable debug logging
kubectl mtv get inventory provider vsphere-prod -v=2
```

#### Authentication Problems

```bash
# Verify secret contents
kubectl get secret -o yaml provider-secret-name

# Check provider credentials
kubectl mtv describe provider vsphere-prod | grep -i secret

# Test with updated credentials
kubectl mtv patch provider vsphere-prod \
  --username updated-user \
  --password updated-password
```

#### Certificate Issues

```bash
# Check certificate format
openssl x509 -in ca-cert.pem -text -noout

# Update certificate
kubectl mtv patch provider vsphere-prod \
  --cacert @/path/to/correct-ca.pem

# Temporarily disable TLS for testing
kubectl mtv patch provider vsphere-test \
  --provider-insecure-skip-tls
```

### Provider Status and Health

```bash
# Check all provider statuses
kubectl mtv get providers -o yaml | grep -A5 -B5 "conditions:"

# Monitor provider health
kubectl mtv get inventory provider vsphere-prod --watch

# Get detailed provider information
kubectl get provider vsphere-prod -o yaml
```

## Next Steps

After mastering provider management:

1. **Explore Inventory**: Learn to discover and query VMs in [Chapter 7: Inventory Management](07-inventory-management.md)
2. **Create Migration Hosts**: Set up direct ESXi connections in [Chapter 5: Migration Host Management](05-migration-host-management.md)
3. **Configure VDDK**: Optimize VMware transfers in [Chapter 6: VDDK Image Creation and Configuration](06-vddk-image-creation-and-configuration.md)
4. **Plan Migrations**: Create migration plans in [Chapter 10: Migration Plan Creation](10-migration-plan-creation.md)

---

*Previous: [Chapter 3: Quick Start - First Migration Workflow](03-quick-start-first-migration-workflow.md)*  
*Next: [Chapter 5: Migration Host Management](05-migration-host-management.md)*
