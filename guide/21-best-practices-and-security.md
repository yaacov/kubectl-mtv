---
layout: page
title: "Chapter 18: Best Practices and Security"
---

Implementing secure and efficient migration practices is essential for production environments. This chapter covers comprehensive best practices for plan management, provider security, query optimization, and secure operations derived from real-world deployment experience and security requirements.

## Overview: Operational Excellence Framework

### Best Practices Philosophy

Effective `kubectl-mtv` operations balance several key principles:

1. **Security by Default**: Implement least-privilege access, secure authentication, and encrypted communications
2. **Operational Efficiency**: Optimize workflows, automate repetitive tasks, and minimize manual interventions
3. **Risk Management**: Test thoroughly, implement rollback procedures, and maintain audit trails
4. **Performance Optimization**: Leverage warm migrations, optimize resource placement, and monitor performance
5. **Maintainability**: Document configurations, standardize procedures, and implement consistent naming

### Security Layers

Migration security encompasses multiple layers:
- **Authentication**: Provider credentials, service accounts, and user authentication
- **Authorization**: RBAC policies, namespace isolation, and resource access control
- **Network Security**: TLS encryption, certificate validation, and network policies
- **Data Protection**: Encryption at rest, secure key management, and data integrity
- **Audit and Compliance**: Logging, monitoring, and regulatory compliance

## Plan Management Strategies

### Testing and Validation

#### Migration Testing Hierarchy

Implement systematic testing approaches:

```bash
# 1. Single VM Test Plan
kubectl mtv create plan test-single \
  --source test-provider \
  --target test-cluster \
  --vms "small-test-vm" \
  --target-namespace test-migrations \
  --migration-type cold

# 2. Small Batch Test Plan
kubectl mtv create plan test-batch \
  --source test-provider \
  --target test-cluster \
  --vms "test-vm-01,test-vm-02,test-vm-03" \
  --target-namespace test-migrations \
  --migration-type warm

# 3. Production Pilot Plan
kubectl mtv create plan pilot-production \
  --source prod-provider \
  --target prod-cluster \
  --vms "non-critical-app-01" \
  --target-namespace production-pilot \
  --migration-type warm \
  --convertor-node-selector "migration=true"

# 4. Full Production Plan (after validation)
kubectl mtv create plan production-migration \
  --source prod-provider \
  --target prod-cluster \
  --vms @validated-production-vms.yaml \
  --target-namespace production \
  --migration-type warm \
  --network-mapping prod-network-map \
  --storage-mapping prod-storage-map \
  --pre-hook production-backup-hook \
  --post-hook production-validation-hook
```

#### Test Plan Validation

```bash
# Validate plan configuration before execution
kubectl mtv describe plan test-plan --with-vms

# Check resource availability
kubectl describe nodes | grep -A5 "Allocatable"

# Verify mapping configurations
kubectl mtv describe mapping network test-network-map
kubectl mtv describe mapping storage test-storage-map

# Test provider connectivity
kubectl mtv get inventory vm test-provider -v=2

# Validate target namespace and permissions
kubectl auth can-i create vm -n test-migrations
```

### Warm Migration Strategy

#### Optimal Warm Migration Implementation

```bash
# Production warm migration with strategic scheduling
kubectl mtv create plan production-warm \
  --source vsphere-prod \
  --target openshift-prod \
  --migration-type warm \
  --vms @production-vms.yaml \
  --target-namespace production \
  --convertor-node-selector "migration-worker=true,performance=high" \
  --convertor-affinity "REQUIRE nodes(storage=nvme) on node" \
  --target-affinity "PREFER nodes(production=true) on zone" \
  --network-mapping production-network \
  --storage-mapping production-storage

# Start with cutover scheduled for maintenance window
kubectl mtv start plan production-warm \
  --cutover "$(date -d 'next Sunday 2:00 AM' --iso-8601=seconds)"

# Monitor warm migration progress
kubectl mtv get plan production-warm -w

# Adjust cutover if needed
kubectl mtv cutover plan production-warm \
  --cutover "$(date -d '+30 minutes' --iso-8601=seconds)"
```

#### Warm Migration Benefits

- **Reduced Downtime**: Pre-copy phase minimizes service interruption
- **Validation Window**: Time to verify data transfer before cutover
- **Rollback Capability**: Source VM remains available until cutover completion
- **Performance Optimization**: Multiple attempts to optimize transfer speed

### Archiving and Lifecycle Management

#### Strategic Plan Archival

```bash
# Archive completed migrations for audit trail
kubectl mtv archive plan completed-q1-migration
kubectl mtv archive plan successful-pilot-test

# Bulk archive old completed plans
for plan in $(kubectl mtv get plans -o json | jq -r '.items[] | select(.status.phase == "Succeeded" and (.metadata.creationTimestamp | fromdateiso8601) < (now - 90*24*3600)) | .metadata.name'); do
  echo "Archiving old plan: $plan"
  kubectl mtv archive plan "$plan"
done

# Maintain active plans, archive obsolete ones
kubectl mtv get plans | grep -E "(Failed|Cancelled)" | awk '{print $1}' | \
  xargs -I {} kubectl mtv archive plan {}
```

#### Plan Template Management

```bash
# Create reusable plan templates
kubectl mtv create plan template-web-tier \
  --source vsphere-template \
  --migration-type warm \
  --target-namespace web-applications \
  --convertor-node-selector "workload=web" \
  --target-affinity "PREFER pods(tier=web) on zone" \
  --network-mapping web-network-map \
  --storage-mapping web-storage-map \
  --vms placeholder-vm

# Archive template for reuse
kubectl mtv archive plan template-web-tier

# Clone template for actual use
kubectl mtv unarchive plan template-web-tier
kubectl mtv patch plan template-web-tier \
  --description "Q2 2024 web tier migration" \
  --vms "web-01,web-02,web-03"
```

## Provider Security

### Credentials Management

#### Secure Provider Creation

```bash
# Use strong authentication credentials
kubectl mtv create provider secure-vsphere \
  --type vsphere \
  --url https://vcenter.secure.com/sdk \
  --username "migration_service@secure.local" \
  --password "$(openssl rand -base64 32)" \
  --cacert @/secure/certificates/vcenter-ca.crt \
  --vddk-init-image registry.secure.com/vddk:8.0.2

# Verify certificate validation is enabled
kubectl mtv describe provider secure-vsphere | grep -i "skip.*tls\|insecure"

# Rotate credentials regularly
kubectl mtv patch provider secure-vsphere \
  --password "$(openssl rand -base64 32)"
```

#### Secret Management Best Practices

```bash
# Use dedicated service accounts for providers
kubectl create serviceaccount migration-service -n migrations

# Create secrets with appropriate labels and annotations
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: vsphere-credentials
  namespace: migrations
  labels:
    app: migration
    provider-type: vsphere
    security-level: restricted
  annotations:
    kubernetes.io/description: "VMware vSphere migration credentials"
    migration.security/rotation-schedule: "monthly"
    migration.security/access-review: "quarterly"
type: Opaque
stringData:
  username: "migration_service@vsphere.local"
  password: "secure-generated-password"
EOF

# Reference secure secret in provider
kubectl mtv create provider secure-vsphere \
  --type vsphere \
  --url https://vcenter.company.com/sdk \
  --existing-secret vsphere-credentials \
  --cacert @/etc/ssl/certs/company-ca.pem
```

### TLS Verification

#### Certificate Validation

```bash
# Always validate certificates in production
kubectl mtv create provider production-vsphere \
  --type vsphere \
  --url https://vcenter.prod.company.com/sdk \
  --username "migration@company.com" \
  --password "secure-password" \
  --cacert @/etc/ssl/company/vcenter-ca.crt
  # Note: No --insecure-skip-tls flag

# For development only, use insecure connections with clear labeling
kubectl mtv create provider dev-vsphere \
  --type vsphere \
  --url https://vcenter-dev.company.com/sdk \
  --username "dev-migration@company.com" \
  --password "dev-password" \
  --provider-insecure-skip-tls
  
# Add clear labels for security auditing
kubectl label provider dev-vsphere \
  security-level=development \
  certificate-validation=disabled \
  environment=non-production
```

#### Certificate Management

```bash
# Extract and verify certificates
openssl s_client -connect vcenter.company.com:443 -showcerts < /dev/null 2>/dev/null | \
  openssl x509 -outform PEM > vcenter-cert.pem

# Verify certificate chain
openssl verify -CAfile company-ca.pem vcenter-cert.pem

# Use certificate bundle for provider
kubectl mtv create provider verified-vsphere \
  --type vsphere \
  --url https://vcenter.company.com/sdk \
  --username "migration@company.com" \
  --password "secure-password" \
  --cacert @vcenter-cert.pem
```

### RBAC Configuration

#### Least-Privilege Provider Access

```yaml
# Minimal RBAC for migration operators
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: migration-operator
rules:
# MTV/Forklift resources
- apiGroups: ["forklift.konveyor.io"]
  resources: ["providers", "plans", "mappings", "hosts"]
  verbs: ["get", "list", "create", "update", "patch", "delete", "watch"]
# Secret access for credentials
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "list", "create", "update", "patch"]
  resourceNames: ["vsphere-*", "ovirt-*", "openstack-*"]
# Inventory access
- apiGroups: [""]
  resources: ["services"]
  verbs: ["get", "list"]
# Route discovery (OpenShift)
- apiGroups: ["route.openshift.io"]
  resources: ["routes"]
  verbs: ["get", "list"]
---
# Bind to service account
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: migration-operator-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: migration-operator
subjects:
- kind: ServiceAccount
  name: migration-operator
  namespace: migrations
```

#### Namespace-Scoped RBAC

```yaml
# Namespace-specific migration permissions
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: production-migrations
  name: migration-user
rules:
- apiGroups: ["forklift.konveyor.io"]
  resources: ["plans"]
  verbs: ["get", "list", "create", "update", "patch", "watch"]
- apiGroups: [""]
  resources: ["secrets", "configmaps"]
  verbs: ["get", "list", "create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: production-migration-users
  namespace: production-migrations
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: migration-user
subjects:
- kind: User
  name: alice@company.com
  apiGroup: rbac.authorization.k8s.io
- kind: User
  name: bob@company.com
  apiGroup: rbac.authorization.k8s.io
```

## Query Optimization Tips

### Efficient Inventory Queries

#### Performance-Optimized Queries

```bash
# Use specific filters instead of broad queries
# Good: Specific criteria
kubectl mtv get inventory vm vsphere-prod \
  --query "where powerState = 'poweredOn' and memory.size > 8192 and name like 'prod-%'"

# Avoid: Broad unfiltered queries
# kubectl mtv get inventory vm vsphere-prod  # Returns everything

# Good: Targeted network queries
kubectl mtv get inventory network vsphere-prod \
  --query "where name ~= '.*production.*' and type != 'dvPortGroup'"

# Good: Storage queries with size filters
kubectl mtv get inventory storage vsphere-prod \
  --query "where capacity > 1073741824 and type = 'VMFS'"  # > 1GB
```

#### Index-Friendly Query Patterns

```bash
# Use exact matches when possible (more efficient)
kubectl mtv get inventory vm vsphere-prod \
  --query "where name = 'specific-vm-name'"

# Use LIKE with anchored patterns
kubectl mtv get inventory vm vsphere-prod \
  --query "where name like 'prod-web-%'"  # Anchored prefix

# Combine filters efficiently
kubectl mtv get inventory vm vsphere-prod \
  --query "where powerState = 'poweredOn' and guestOS like 'linux%' and memory.size between 4096 and 16384"
```

### Query Result Management

#### Efficient Result Processing

```bash
# Export large queries to files for processing
kubectl mtv get inventory vm large-provider \
  --query "where tags.category = 'production'" \
  -o planvms > production-vms.yaml

# Use pagination for massive datasets
kubectl mtv get inventory vm huge-provider \
  --query "where name ~= '^[a-m].*'"  # First half alphabetically

kubectl mtv get inventory vm huge-provider \
  --query "where name ~= '^[n-z].*'"  # Second half

# Process results in manageable chunks
for prefix in {a..z}; do
  kubectl mtv get inventory vm large-provider \
    --query "where name like '${prefix}%'" \
    -o planvms > "vms-${prefix}.yaml"
done
```

## Secure Service Account Setup for Admin Access

### Administrative Service Account

#### Complete Admin Service Account Setup

```bash
# Create dedicated namespace for migration administration
kubectl create namespace migration-admin

# Create administrative service account
kubectl create serviceaccount migration-admin -n migration-admin

# Create comprehensive admin cluster role
cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: migration-admin
rules:
# Full MTV/Forklift access
- apiGroups: ["forklift.konveyor.io"]
  resources: ["*"]
  verbs: ["*"]
# Core Kubernetes resources for migration
- apiGroups: [""]
  resources: ["secrets", "configmaps", "namespaces", "services", "pods", "persistentvolumes", "persistentvolumeclaims"]
  verbs: ["get", "list", "create", "update", "patch", "delete", "watch"]
# Storage management
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses", "volumeattachments"]
  verbs: ["get", "list"]
# Network management
- apiGroups: ["k8s.cni.cncf.io"]
  resources: ["network-attachment-definitions"]
  verbs: ["get", "list", "create", "update", "patch", "delete"]
# OpenShift routes (if applicable)
- apiGroups: ["route.openshift.io"]
  resources: ["routes"]
  verbs: ["get", "list"]
# Node information
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get", "list", "watch"]
# Events for troubleshooting
- apiGroups: [""]
  resources: ["events"]
  verbs: ["get", "list", "watch"]
EOF

# Bind admin role to service account
kubectl create clusterrolebinding migration-admin-binding \
  --clusterrole=migration-admin \
  --serviceaccount=migration-admin:migration-admin
```

#### Token Management for Automation

```bash
# Create long-lived token secret (Kubernetes 1.24+)
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: migration-admin-token
  namespace: migration-admin
  annotations:
    kubernetes.io/service-account.name: migration-admin
type: kubernetes.io/service-account-token
EOF

# Extract token for use in automation
ADMIN_TOKEN=$(kubectl get secret migration-admin-token -n migration-admin -o jsonpath='{.data.token}' | base64 -d)

# Create kubeconfig for service account
kubectl config set-cluster migration-cluster \
  --server=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}') \
  --certificate-authority-data=$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')

kubectl config set-credentials migration-admin \
  --token="$ADMIN_TOKEN"

kubectl config set-context migration-admin-context \
  --cluster=migration-cluster \
  --user=migration-admin

# Test service account access
kubectl --context=migration-admin-context auth whoami
kubectl --context=migration-admin-context get providers --all-namespaces
```

### Security Audit and Monitoring

#### Access Monitoring

```bash
# Monitor service account usage
kubectl get events --all-namespaces | grep migration-admin

# Audit RBAC permissions
kubectl auth can-i --list --as=system:serviceaccount:migration-admin:migration-admin

# Check recent authentication events
kubectl get events --all-namespaces | grep -i "authentication\|authorization"

# Review service account tokens
kubectl get secrets --all-namespaces | grep service-account-token | grep migration
```

#### Security Compliance

```bash
# Regular permission reviews
echo "=== Migration Admin Permissions Review ==="
kubectl describe clusterrolebinding migration-admin-binding
kubectl describe clusterrole migration-admin

# Token rotation schedule
kubectl delete secret migration-admin-token -n migration-admin
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: migration-admin-token
  namespace: migration-admin
  annotations:
    kubernetes.io/service-account.name: migration-admin
    rotation-date: "$(date --iso-8601)"
type: kubernetes.io/service-account-token
EOF

# Access pattern analysis
kubectl get events --all-namespaces --sort-by='.metadata.creationTimestamp' | \
  grep migration-admin | tail -50
```

## Operational Security Framework

### Network Security

#### Network Policies for Migration

```yaml
# Restrict migration namespace network access
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: migration-network-policy
  namespace: migrations
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: migration-admin
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: openshift-mtv
  - to: []
    ports:
    - protocol: TCP
      port: 443  # HTTPS for provider APIs
    - protocol: TCP
      port: 6443 # Kubernetes API
```

#### Secure Communication

```bash
# Ensure all provider communications use TLS
kubectl mtv get providers -o yaml | grep -B5 -A5 "insecureSkipTLS: true" || echo "All providers use TLS"

# Verify certificate validation
for provider in $(kubectl mtv get providers -o jsonpath='{.items[*].metadata.name}'); do
  echo "Provider: $provider"
  kubectl mtv describe provider "$provider" | grep -i "certificate\|tls\|insecure"
done
```

### Data Protection

#### Encryption and Key Management

```bash
# Use encrypted secrets for provider credentials
kubectl create secret generic secure-provider-creds \
  --from-literal=username=encrypted_user \
  --from-literal=password="$(gpg --symmetric --armor --cipher-algo AES256 <<< 'actual-password')" \
  -n migrations

# Label secrets for encryption compliance
kubectl label secret secure-provider-creds \
  encryption=required \
  compliance=sox \
  data-classification=confidential

# Verify secret encryption at rest
kubectl get secrets -o yaml | grep -A5 -B5 "encryption"
```

#### Audit Trail Management

```bash
# Enable audit logging for migration operations
cat <<EOF | sudo tee /etc/kubernetes/audit-policy.yaml
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: Detailed
  resources:
  - group: "forklift.konveyor.io"
    resources: ["*"]
- level: Request
  verbs: ["create", "update", "patch", "delete"]
  resources:
  - group: ""
    resources: ["secrets"]
  namespaces: ["migrations", "migration-admin"]
EOF

# Monitor audit logs for migration activities
sudo journalctl -u kubelet | grep -i "forklift\|migration" | tail -20
```

### Performance and Resource Security

#### Resource Limits and Quotas

```yaml
# Resource quota for migration namespace
apiVersion: v1
kind: ResourceQuota
metadata:
  name: migration-quota
  namespace: migrations
spec:
  hard:
    requests.cpu: "20"
    requests.memory: 64Gi
    limits.cpu: "40"
    limits.memory: 128Gi
    persistentvolumeclaims: "50"
    secrets: "20"
---
# Limit range for migration pods
apiVersion: v1
kind: LimitRange
metadata:
  name: migration-limits
  namespace: migrations
spec:
  limits:
  - type: Pod
    max:
      cpu: "8"
      memory: 16Gi
    default:
      cpu: "2"
      memory: 4Gi
    defaultRequest:
      cpu: "1"
      memory: 2Gi
```

#### Secure Convertor Pod Configuration

```bash
# Create convertor pods with security context
kubectl mtv create plan secure-migration \
  --source vsphere-secure \
  --convertor-node-selector "security=restricted,taint=dedicated" \
  --convertor-labels "security-context=restricted,compliance=required" \
  --vms @secure-vms.yaml

# Verify convertor security configuration
kubectl describe pod convertor-pod | grep -A10 "Security Context"
```

## Compliance and Governance

### Migration Documentation

#### Required Documentation

```bash
# Document all migration plans
cat <<EOF > migration-documentation.md
# Migration Plan Documentation

## Plan: production-q4-migration
- **Purpose**: Quarterly production workload migration
- **Source**: VMware vSphere 7.0 (vcenter.company.com)  
- **Target**: OpenShift 4.12 (prod-cluster.company.com)
- **Migration Window**: 2024-01-15 02:00-06:00 UTC
- **Rollback Plan**: Documented in rollback-procedures.md
- **Approvals**: IT Manager (John Smith), Security (Jane Doe)
- **Risk Assessment**: Medium (tested in staging)

## VMs Included:
$(kubectl mtv get plan production-q4-migration -o jsonpath='{.spec.vms[*].name}' | tr ' ' '\n' | sort)

## Security Considerations:
- TLS verification enabled for all providers
- RBAC follows least-privilege principle  
- Audit logging enabled
- Network policies restrict access
EOF
```

### Change Management Integration

```bash
# Migration change request workflow
#!/bin/bash
# migration-change-request.sh

PLAN_NAME="$1"
CHANGE_TICKET="$2"

if [ -z "$PLAN_NAME" ] || [ -z "$CHANGE_TICKET" ]; then
  echo "Usage: $0 <plan-name> <change-ticket-id>"
  exit 1
fi

# Label plan with change management metadata
kubectl label plan "$PLAN_NAME" \
  change-ticket="$CHANGE_TICKET" \
  approval-status=approved \
  risk-level=medium \
  change-window="$(date --iso-8601=date)"

# Add change management annotation
kubectl annotate plan "$PLAN_NAME" \
  change-management/approver="migration-board@company.com" \
  change-management/window="Maintenance window 02:00-06:00 UTC" \
  change-management/rollback="Available until cutover completion"

echo "Plan $PLAN_NAME tagged with change request $CHANGE_TICKET"
```

## Next Steps

After implementing comprehensive best practices and security:

1. **AI Integration**: Explore advanced automation in [Chapter 22: Model Context Protocol (MCP) Server Integration](/kubectl-mtv/22-model-context-protocol-mcp-server-integration)
2. **Tool Integration**: Learn KubeVirt ecosystem integration in [Chapter 23: Integration with KubeVirt Tools](/kubectl-mtv/23-integration-with-kubevirt-tools)

---

*Previous: [Chapter 20: Debugging and Troubleshooting](/kubectl-mtv/20-debugging-and-troubleshooting)*  
*Next: [Chapter 22: Model Context Protocol (MCP) Server Integration](/kubectl-mtv/22-model-context-protocol-mcp-server-integration)*
