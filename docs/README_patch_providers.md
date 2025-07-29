# Patching Providers

This guide explains how to modify existing providers using the `kubectl-mtv patch provider` command. The patch command allows you to update provider URLs, credentials, and VDDK settings while protecting immutable fields like provider type and SDK endpoint.

## Overview

The patch provider command provides secure updates for existing providers:
- **Update provider URL**: Change the connection endpoint
- **Update credentials**: Modify authentication information (with ownership validation)
- **Update VDDK settings**: Configure vSphere-specific VDDK parameters
- **Protected fields**: Provider type and SDK endpoint cannot be changed after creation

## Basic Syntax

```bash
kubectl-mtv patch provider NAME [flags]
```

## Editable Fields

### Provider Configuration
- `--url`: Provider connection URL
- `--provider-insecure-skip-tls`: Skip TLS certificate verification

### Credentials (with ownership validation)
- `--username`: Authentication username
- `--password`: Authentication password  
- `--token`: Authentication token (OpenShift)
- `--cacert`: CA certificate (use @filename to load from file)

### OpenStack-specific
- `--provider-domain-name`: OpenStack domain name
- `--provider-project-name`: OpenStack project name
- `--provider-region-name`: OpenStack region name

### vSphere VDDK Settings
- `--vddk-init-image`: VDDK container init image path
- `--use-vddk-aio-optimization`: Enable VDDK AIO optimization
- `--vddk-buf-size-in-64k`: VDDK buffer size in 64K units
- `--vddk-buf-count`: VDDK buffer count

## Usage Examples

### Update Provider URL

```bash
# Change vSphere provider URL
kubectl-mtv patch provider vsphere-provider \
  --url https://new-vcenter.example.com

# Change OpenShift provider URL
kubectl-mtv patch provider openshift-provider \
  --url https://api.new-cluster.example.com:6443
```

### Update Credentials

```bash
# Update vSphere credentials (only if secret is owned by provider)
kubectl-mtv patch provider vsphere-provider \
  --username administrator@vsphere.local \
  --password newpassword

# Update OpenShift token
kubectl-mtv patch provider openshift-provider \
  --token sha256~new-token-here

# Update credentials with CA certificate from file
kubectl-mtv patch provider vsphere-provider \
  --username newuser \
  --cacert @/path/to/ca-cert.pem
```

### Update OpenStack Credentials

```bash
# Update OpenStack provider with multiple credential fields
kubectl-mtv patch provider openstack-provider \
  --username newuser \
  --password newpass \
  --provider-domain-name new-domain \
  --provider-project-name new-project \
  --provider-region-name us-west-2
```

### Update VDDK Settings

```bash
# Update vSphere provider VDDK configuration
kubectl-mtv patch provider vsphere-provider \
  --vddk-init-image registry.example.com/vddk:v8.0.2 \
  --use-vddk-aio-optimization=true

# Update VDDK buffer settings for performance tuning
kubectl-mtv patch provider vsphere-provider \
  --vddk-buf-size-in-64k 32 \
  --vddk-buf-count 16
```

### Combined Updates

```bash
# Update multiple fields in one command
kubectl-mtv patch provider vsphere-provider \
  --url https://new-vcenter.example.com \
  --username newadmin@vsphere.local \
  --password newpassword \
  --vddk-init-image registry.example.com/vddk:latest \
  --use-vddk-aio-optimization=true
```

## Secret Ownership and Security

### Credential Update Protection

The patch command includes built-in security to prevent unauthorized credential updates:

```bash
$ kubectl-mtv patch provider shared-provider --username newuser
Error: cannot update credentials: the secret 'shared-secret' is not owned by provider 'shared-provider'. 
This usually means the secret was created independently and is shared by multiple providers. 
To update credentials, either:
1. Update the secret directly: kubectl patch secret shared-secret -p '{...}'
2. Create a new secret and update the provider to use it
```

### How Secret Ownership Works

When providers are created with `kubectl-mtv create provider`, the associated secret is automatically owned by the provider through Kubernetes owner references. This ownership allows safe credential updates via the patch command.

**Owned secrets** (safe to patch):
- Created automatically during provider creation
- Have the provider as an owner reference
- Can be safely updated via `kubectl-mtv patch provider`

**Shared secrets** (protected):
- Created independently and referenced by provider
- May be used by multiple providers
- Cannot be updated via provider patch (must be updated directly)

### Checking Secret Ownership

```bash
# Check if a secret is owned by a provider
kubectl get secret <secret-name> -o yaml | grep -A5 ownerReferences

# Example owned secret:
ownerReferences:
- apiVersion: forklift.konveyor.io/v1beta1
  kind: Provider
  name: my-provider
  uid: abc123...
```

## Provider-Specific Credential Fields

### OpenShift Providers
- **Token**: `--token` → Updates `secret.data.token`

### vSphere/oVirt/OVA Providers  
- **Username**: `--username` → Updates `secret.data.user`
- **Password**: `--password` → Updates `secret.data.password`

### OpenStack Providers
- **Username**: `--username` → Updates `secret.data.username`
- **Password**: `--password` → Updates `secret.data.password`
- **Domain**: `--provider-domain-name` → Updates `secret.data.domainName`
- **Project**: `--provider-project-name` → Updates `secret.data.projectName`
- **Region**: `--provider-region-name` → Updates `secret.data.regionName`

### All Provider Types
- **CA Certificate**: `--cacert` → Updates `secret.data.cacert`

## Immutable Fields

These fields cannot be changed after provider creation:

### Provider Type
```bash
$ kubectl-mtv patch provider my-provider --type openshift
Error: unknown flag: --type
```

**Reason**: Provider type determines the fundamental behavior and cannot be changed. Create a new provider if you need a different type.

### SDK Endpoint
```bash
$ kubectl-mtv patch provider my-provider --sdk-endpoint esxi
Error: unknown flag: --sdk-endpoint  
```

**Reason**: SDK endpoint affects how the provider connects to the source system and cannot be changed safely.

## Environment Variables

The patch command respects the same environment variables as create:

```bash
# Set default VDDK init image
export MTV_VDDK_INIT_IMAGE=registry.example.com/vddk:v8.0.2
kubectl-mtv patch provider vsphere-provider --use-vddk-aio-optimization=true
```

## Best Practices

### 1. Verify Current Settings

```bash
# Check current provider configuration
kubectl-mtv describe provider my-provider

# Check current secret contents (if owned)
kubectl get secret <secret-name> -o yaml
```

### 2. Test Connectivity After Updates

```bash
# After updating provider settings, verify connectivity
kubectl-mtv get inventory networks --provider my-provider
```

### 3. Use File-based CA Certificates

```bash
# Safer to load CA certificates from files
kubectl-mtv patch provider my-provider --cacert @/path/to/ca-cert.pem

# Instead of pasting certificate content directly
```

### 4. Update Related Resources

```bash
# If changing provider URL, you may need to update related mappings
kubectl-mtv get mapping --output json | grep my-provider

# Consider if host resources need updates
kubectl-mtv get host --output json | grep my-provider
```

### 5. Backup Before Major Changes

```bash
# Export provider configuration before major updates
kubectl get provider my-provider -o yaml > provider-backup.yaml
kubectl get secret <secret-name> -o yaml > secret-backup.yaml
```

## Troubleshooting

### Common Issues

#### Secret Not Found
```
Error: failed to get secret 'my-secret': secrets "my-secret" not found
```
**Solution**: The provider references a secret that doesn't exist. Create the secret or update the provider to reference an existing one.

#### Secret Not Owned
```
Error: cannot update credentials: the secret 'shared-secret' is not owned by provider 'my-provider'
```
**Solution**: Either update the secret directly or create a new secret for this provider.

#### Provider Not Found
```
Error: failed to get provider 'nonexistent': providers.forklift.konveyor.io "nonexistent" not found
```
**Solution**: Verify the provider name and namespace:
```bash
kubectl-mtv get provider
```

#### Permission Denied
```
Error: failed to update provider: providers.forklift.konveyor.io "my-provider" is forbidden
```
**Solution**: Ensure you have update permissions for provider resources in the target namespace.

### Debugging

Enable verbose logging for detailed information:

```bash
# Basic logging
kubectl-mtv patch provider my-provider --url https://new-url.com -v 2

# Detailed logging
kubectl-mtv patch provider my-provider --username newuser -v 3
```

### Validation

After patching, verify the changes took effect:

```bash
# Check provider status
kubectl-mtv describe provider my-provider

# Test provider connectivity
kubectl-mtv get inventory --provider my-provider

# Check secret contents (if credentials were updated)
kubectl get secret <secret-name> -o jsonpath='{.data}' | base64 -d
```

## Integration with Migration Workflows

Updated providers are immediately available for use:

```bash
# Patch provider credentials
kubectl-mtv patch provider vsphere-provider --password newpassword

# Use updated provider in new migration plan
kubectl-mtv create plan migration-with-updated-provider \
  --source-provider vsphere-provider \
  --target-provider openshift-provider \
  --vm "my-vm"
```

## Related Commands

- `kubectl-mtv create provider` - Create new providers
- `kubectl-mtv get provider` - List existing providers
- `kubectl-mtv describe provider` - View provider details  
- `kubectl-mtv delete provider` - Delete providers

## See Also

- [Creating Providers](README_create_providers.md) - How to create initial providers
- [Provider Types](README_provider_types.md) - Understanding different provider types
- [Migration Plans](README_planvms.md) - Using providers in migration plans
- [Inventory Management](README_inventory.md) - Working with provider inventory 