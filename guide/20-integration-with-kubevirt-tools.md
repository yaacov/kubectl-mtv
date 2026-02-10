---
layout: page
title: "Chapter 20: Integration with KubeVirt Tools"
---

After successful VM migration with `kubectl-mtv`, the migrated VMs become full KubeVirt virtual machines that can be managed using the broader KubeVirt ecosystem. This chapter covers the seamless integration between `kubectl-mtv` and `virtctl`, enabling complete VM lifecycle management from migration through ongoing operations.

## Overview: Relationship between kubectl-mtv and virtctl

### Complementary Tool Ecosystem

`kubectl-mtv` and `virtctl` form a complete virtualization management solution:

| Tool | Primary Focus | Key Capabilities |
|------|---------------|------------------|
| **kubectl-mtv** | **Migration** | Provider management, inventory discovery, migration planning, execution, monitoring |
| **virtctl** | **VM Operations** | VM lifecycle, console access, resource management, advanced KubeVirt features |

### Migration to Operations Workflow

```mermaid
graph LR
    A[Source VMs] --> B[kubectl-mtv Migration]
    B --> C[KubeVirt VMs]
    C --> D[virtctl Management]
    D --> E[Ongoing Operations]
```

The integration provides:
1. **Seamless Transition**: Migrated VMs are immediately manageable with virtctl
2. **Shared Resources**: Common KubeVirt objects (DataVolumes, InstanceTypes, etc.)
3. **Consistent Experience**: Similar CLI patterns and Kubernetes-native operations
4. **Advanced Features**: Access to full KubeVirt capabilities post-migration

### Resource Continuity

Resources created during migration are compatible with virtctl operations:
- **DataVolumes**: Created by kubectl-mtv, manageable with virtctl
- **Instance Types**: Applied during migration, usable in virtctl operations
- **Network Attachments**: Configured in mapping, accessible via virtctl
- **Storage Classes**: Used in migration, available for new operations

## Tool Installation and Setup

### VirtCtl Installation

#### Option 1: Binary Installation

```bash
# Get latest release version
export VERSION=$(curl -s https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt)

# Download and install virtctl
wget https://github.com/kubevirt/kubevirt/releases/download/${VERSION}/virtctl-${VERSION}-linux-amd64
chmod +x virtctl-${VERSION}-linux-amd64
sudo mv virtctl-${VERSION}-linux-amd64 /usr/local/bin/virtctl

# Verify installation
virtctl version
```

#### Option 2: Kubectl Plugin (Krew)

```bash
# Install via Krew
kubectl krew install virt

# Use as kubectl plugin
kubectl virt version

# Create alias for convenience
echo 'alias virtctl="kubectl virt"' >> ~/.bashrc
source ~/.bashrc
```

#### Option 3: Package Manager Installation

```bash
# Ubuntu/Debian (if available in repos)
sudo apt update && sudo apt install virtctl

# RHEL/CentOS/Fedora (if available in repos)
sudo dnf install virtctl

# Verify installation
virtctl version
```

### Verification of Integration

```bash
# Check both tools are available
kubectl mtv version
virtctl version

# Verify cluster access
kubectl get vms
kubectl get virtualmachines

# Check MTV and KubeVirt resources
kubectl get providers.forklift.konveyor.io
kubectl get kubevirt -n kubevirt
```

## Post-Migration VM Management

### VM Lifecycle Operations

#### Starting and Stopping VMs

After migration, use virtctl for VM power management:

```bash
# List migrated VMs
kubectl get vms -l migration.forklift.konveyor.io/plan=production-migration

# Start a migrated VM
virtctl start web-server-01 -n production

# Stop VM gracefully (default 30 second grace period)
virtctl stop database-primary -n production

# Stop VM with custom grace period
virtctl stop web-server-02 -n production --grace-period 60

# Force stop VM immediately (emergency use)
virtctl stop problematic-vm -n production --force

# Restart VM
virtctl restart cache-server -n production
```

#### Advanced Power Management

```bash
# Pause VM (freeze execution)
virtctl pause web-server-01 -n production

# Unpause VM (resume execution)
virtctl unpause web-server-01 -n production

# Soft reboot VM guest OS
virtctl soft-reboot database-primary -n production

# Dry run operations for testing
virtctl restart test-vm --dry-run
```

### VM Status and Monitoring

#### Health and Status Checks

```bash
# Get detailed VM status
kubectl describe vm web-server-01 -n production

# Check VM conditions
kubectl get vm web-server-01 -n production -o yaml | grep -A 10 conditions

# Monitor VM events
kubectl get events -n production | grep web-server-01

# Watch VM status changes
kubectl get vms -n production -w
```

#### Resource Utilization

```bash
# Get VM resource usage
kubectl top vm web-server-01 -n production

# Monitor all VMs in namespace
kubectl top vms -n production

# Get detailed resource information
virtctl vmexport get web-server-01 -n production --include-secret
```

## Console and Remote Access

### Serial Console Access

```bash
# Access VM serial console (text-based)
virtctl console web-server-01 -n production

# Exit console: Ctrl+]

# Console with specific timeout
virtctl console database-primary -n production --timeout 300
```

### VNC Console Access

```bash
# Access VM graphical console via VNC
virtctl vnc web-server-01 -n production

# VNC with specific local port
virtctl vnc web-server-01 -n production --port 5901

# VNC with proxy configuration
virtctl vnc web-server-01 -n production --proxy-only
```

### SSH Access

#### Direct SSH to VMs

```bash
# SSH to VM (requires guest agent and SSH setup)
virtctl ssh user@web-server-01 -n production

# SSH with specific identity file
virtctl ssh -i ~/.ssh/vm_key user@web-server-01 -n production

# SSH with port forwarding
virtctl ssh user@web-server-01 -n production -L 8080:localhost:80

# SSH with command execution
virtctl ssh user@web-server-01 -n production "systemctl status nginx"
```

#### SSH Key Management

```bash
# Inject SSH key during VM creation (post-migration)
virtctl create vm new-vm \
  --image registry.company.com/ubuntu:20.04 \
  --ssh-key "$(cat ~/.ssh/id_rsa.pub)" \
  --user admin

# Add SSH authorized keys to running VM
virtctl addauthorizedkey web-server-01 -n production \
  --user admin \
  --key "$(cat ~/.ssh/new_key.pub)"

# Remove SSH authorized keys
virtctl removeauthorizedkey web-server-01 -n production \
  --user admin \
  --key "$(cat ~/.ssh/old_key.pub)"
```

## Advanced VM Operations

### Live Migration

```bash
# Migrate VM to another node
virtctl migrate web-server-01 -n production

# Migrate to specific node
virtctl migrate database-primary -n production --node worker-node-3

# Dry run migration
virtctl migrate test-vm -n test --dry-run

# Cancel ongoing migration
virtctl migrate-cancel web-server-01 -n production
```

### VM Cloning and Expansion

```bash
# Create VM from existing DataVolume (created during migration)
virtctl create vm cloned-web-server \
  --source-pvc web-server-01-disk-0 \
  --instancetype medium \
  --namespace production

# Expand VM disk (if supported by storage class)
virtctl expand web-server-01 -n production --size 100Gi

# Create VM snapshot
virtctl snapshot create web-server-01-snapshot \
  --vm web-server-01 \
  --namespace production
```

### Resource Management

#### Memory and CPU Management

```bash
# Hot-plug CPU (if supported)
virtctl cpu hot-plug web-server-01 -n production --cores 4

# Hot-plug memory (if supported)
virtctl memory hot-plug web-server-01 -n production --size 8Gi

# Add additional disk
virtctl add-volume web-server-01 \
  --volume-name additional-storage \
  --pvc-name additional-pvc \
  --namespace production

# Remove volume
virtctl remove-volume web-server-01 \
  --volume-name additional-storage \
  --namespace production
```

## Integration Workflows

### Complete Migration to Operations Workflow

```bash
#!/bin/bash
# complete-migration-workflow.sh - End-to-end migration and operations setup

PLAN_NAME="production-web-tier"
NAMESPACE="production"

echo "=== Phase 1: Migration with kubectl-mtv ==="

# 1. Create and execute migration
kubectl mtv create plan "$PLAN_NAME" \
  --source vsphere-prod \
  --target openshift-prod \
  --vms "web-01,web-02,web-03" \
  --target-namespace "$NAMESPACE" \
  --migration-type warm

kubectl mtv start plan "$PLAN_NAME" \
  --cutover "$(date -d '+2 hours' --iso-8601=seconds)"

# 2. Monitor migration completion
echo "Waiting for migration completion..."
while true; do
  STATUS=$(kubectl mtv get plan "$PLAN_NAME" -o jsonpath='{.status.phase}')
  if [ "$STATUS" = "Succeeded" ]; then
    echo "Migration completed successfully"
    break
  elif [ "$STATUS" = "Failed" ]; then
    echo "Migration failed"
    exit 1
  fi
  sleep 30
done

echo "=== Phase 2: Post-Migration Operations with virtctl ==="

# 3. Verify migrated VMs
kubectl get vms -n "$NAMESPACE" -l "forklift.konveyor.io/plan=$PLAN_NAME"

# 4. Start migrated VMs
for vm in web-01 web-02 web-03; do
  echo "Starting VM: $vm"
  virtctl start "$vm" -n "$NAMESPACE"
done

# 5. Wait for VMs to be ready
echo "Waiting for VMs to start..."
for vm in web-01 web-02 web-03; do
  while true; do
    if kubectl get vm "$vm" -n "$NAMESPACE" -o jsonpath='{.status.ready}' | grep -q true; then
      echo "VM $vm is ready"
      break
    fi
    sleep 10
  done
done

# 6. Validate VM connectivity
for vm in web-01 web-02 web-03; do
  echo "Testing connectivity to $vm..."
  virtctl ssh admin@"$vm" -n "$NAMESPACE" "echo 'VM $vm is accessible via SSH'" || echo "SSH not ready for $vm"
done

echo "=== Migration and operations setup complete ==="
```

### Monitoring and Health Checks

```bash
#!/bin/bash
# vm-health-monitor.sh - Comprehensive VM health monitoring

NAMESPACE="production"

echo "=== VM Health Dashboard ==="

# VM Status Overview
echo "1. VM Status Overview:"
kubectl get vms -n "$NAMESPACE" -o custom-columns="NAME:.metadata.name,STATUS:.status.ready,AGE:.metadata.creationTimestamp"
echo

# Resource Utilization
echo "2. Resource Utilization:"
kubectl top vms -n "$NAMESPACE" 2>/dev/null || echo "Metrics not available"
echo

# Network Connectivity
echo "3. Network Connectivity Tests:"
for vm in $(kubectl get vms -n "$NAMESPACE" -o jsonpath='{.items[*].metadata.name}'); do
  echo "Testing VM: $vm"
  
  # Check VM is running
  if kubectl get vm "$vm" -n "$NAMESPACE" -o jsonpath='{.status.ready}' | grep -q true; then
    echo "  [OK] VM is running"
    
    # Test SSH connectivity
    if virtctl ssh admin@"$vm" -n "$NAMESPACE" "echo 'SSH OK'" &>/dev/null; then
      echo "  [OK] SSH accessible"
    else
      echo "  [FAIL] SSH not accessible"
    fi
    
    # Test application ports (example)
    if virtctl ssh admin@"$vm" -n "$NAMESPACE" "nc -z localhost 80" &>/dev/null; then
      echo "  [OK] HTTP service responding"
    else
      echo "  [FAIL] HTTP service not responding"
    fi
  else
    echo "  [FAIL] VM not running"
  fi
  echo
done

# Recent Events
echo "4. Recent VM Events:"
kubectl get events -n "$NAMESPACE" --sort-by='.metadata.creationTimestamp' | tail -10
```

### Backup and Disaster Recovery

```bash
#!/bin/bash
# vm-backup-restore.sh - VM backup and restore operations

NAMESPACE="production"
BACKUP_NAMESPACE="backups"

backup_vm() {
  local vm_name="$1"
  local backup_name="${vm_name}-backup-$(date +%Y%m%d-%H%M%S)"
  
  echo "Creating backup for VM: $vm_name"
  
  # Create VM snapshot
  virtctl snapshot create "$backup_name" \
    --vm "$vm_name" \
    --namespace "$NAMESPACE"
  
  # Export VM configuration
  kubectl get vm "$vm_name" -n "$NAMESPACE" -o yaml > "${backup_name}-config.yaml"
  
  echo "Backup created: $backup_name"
}

restore_vm() {
  local backup_name="$1"
  local new_vm_name="$2"
  
  echo "Restoring VM from backup: $backup_name"
  
  # Restore from snapshot
  virtctl snapshot restore "$backup_name" \
    --vm "$new_vm_name" \
    --namespace "$NAMESPACE"
  
  echo "VM restored as: $new_vm_name"
}

# Backup critical VMs
for vm in web-01 database-primary cache-server; do
  backup_vm "$vm"
done

# Restore example (when needed)
# restore_vm "web-01-backup-20240115-140000" "web-01-restored"
```

## Best Practices for Tool Integration

### Operational Workflows

#### Daily Operations

```bash
# Morning health check routine
#!/bin/bash
echo "=== Daily VM Health Check ==="

# 1. Check all VM status
kubectl get vms --all-namespaces | grep -v Running | head -20

# 2. Check resource utilization
kubectl top vms --all-namespaces | head -10

# 3. Check for failed migrations
kubectl mtv get plans --all-namespaces | grep Failed

# 4. Test critical VM connectivity
for vm in critical-db web-primary cache-main; do
  virtctl ssh admin@"$vm" -n production "uptime" || echo "ALERT: $vm not accessible"
done
```

#### Maintenance Procedures

```bash
# Planned maintenance workflow
maintenance_vm() {
  local vm_name="$1"
  local namespace="$2"
  
  echo "Starting maintenance for $vm_name"
  
  # 1. Create pre-maintenance snapshot
  virtctl snapshot create "${vm_name}-pre-maintenance" \
    --vm "$vm_name" --namespace "$namespace"
  
  # 2. Migrate to maintenance node (if needed)
  virtctl migrate "$vm_name" -n "$namespace" --node maintenance-node
  
  # 3. Perform maintenance operations
  virtctl ssh admin@"$vm_name" -n "$namespace" "sudo apt update && sudo apt upgrade -y"
  
  # 4. Restart VM
  virtctl restart "$vm_name" -n "$namespace"
  
  # 5. Verify health
  sleep 60
  virtctl ssh admin@"$vm_name" -n "$namespace" "systemctl status nginx" || echo "Service check failed"
  
  echo "Maintenance completed for $vm_name"
}
```

### Resource Optimization

#### Performance Tuning

```bash
# VM performance optimization
optimize_vm() {
  local vm_name="$1"
  local namespace="$2"
  
  echo "Optimizing performance for $vm_name"
  
  # Check current resource usage
  kubectl top vm "$vm_name" -n "$namespace"
  
  # Optimize based on usage patterns
  # (These operations require VM restart)
  
  # Hot-plug additional CPU if needed
  if [ $(kubectl top vm "$vm_name" -n "$namespace" | awk 'NR==2 {print $2}' | sed 's/%//') -gt 80 ]; then
    echo "High CPU usage detected, consider scaling up"
    # virtctl cpu hot-plug "$vm_name" -n "$namespace" --cores 4
  fi
  
  # Hot-plug additional memory if needed
  if [ $(kubectl top vm "$vm_name" -n "$namespace" | awk 'NR==2 {print $3}' | sed 's/%//') -gt 80 ]; then
    echo "High memory usage detected, consider scaling up"
    # virtctl memory hot-plug "$vm_name" -n "$namespace" --size 8Gi
  fi
}
```

#### Storage Management

```bash
# VM storage operations
manage_vm_storage() {
  local vm_name="$1"
  local namespace="$2"
  
  echo "Managing storage for $vm_name"
  
  # Check disk usage via SSH
  virtctl ssh admin@"$vm_name" -n "$namespace" "df -h" | grep -E "(Filesystem|/dev/)"
  
  # Add additional storage if needed
  # virtctl add-volume "$vm_name" \
  #   --volume-name extra-storage \
  #   --pvc-name "${vm_name}-extra-pvc" \
  #   --namespace "$namespace"
  
  # Create new PVC for expansion
  # kubectl apply -f - <<EOF
  # apiVersion: v1
  # kind: PersistentVolumeClaim
  # metadata:
  #   name: ${vm_name}-extra-pvc
  #   namespace: $namespace
  # spec:
  #   accessModes: [ReadWriteOnce]
  #   resources:
  #     requests:
  #       storage: 50Gi
  # EOF
}
```

## Security and Access Control

### VM Security Management

```bash
# VM security hardening
secure_vm() {
  local vm_name="$1"
  local namespace="$2"
  
  echo "Applying security hardening to $vm_name"
  
  # Update SSH configuration
  virtctl ssh admin@"$vm_name" -n "$namespace" "
    sudo sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
    sudo systemctl reload sshd
  "
  
  # Apply security updates
  virtctl ssh admin@"$vm_name" -n "$namespace" "
    sudo apt update
    sudo apt upgrade -y
    sudo apt autoremove -y
  "
  
  # Configure firewall
  virtctl ssh admin@"$vm_name" -n "$namespace" "
    sudo ufw --force enable
    sudo ufw default deny incoming
    sudo ufw allow ssh
    sudo ufw allow http
    sudo ufw allow https
  "
  
  echo "Security hardening completed for $vm_name"
}
```

### User and Access Management

```bash
# VM user management
manage_vm_users() {
  local vm_name="$1"
  local namespace="$2"
  
  echo "Managing users for $vm_name"
  
  # List current users
  virtctl ssh admin@"$vm_name" -n "$namespace" "cat /etc/passwd | grep -E '/bin/(bash|sh)$'"
  
  # Add new SSH key for user
  virtctl addauthorizedkey "$vm_name" -n "$namespace" \
    --user admin \
    --key "$(cat ~/.ssh/new_team_member.pub)"
  
  # Remove old SSH key
  virtctl removeauthorizedkey "$vm_name" -n "$namespace" \
    --user admin \
    --key "$(cat ~/.ssh/departed_member.pub)"
  
  echo "User management completed for $vm_name"
}
```

## Troubleshooting Integration Issues

### Common Issues and Solutions

#### VM Not Starting After Migration

```bash
# Troubleshoot VM startup issues
debug_vm_startup() {
  local vm_name="$1"
  local namespace="$2"
  
  echo "Debugging startup issues for $vm_name"
  
  # Check VM status
  kubectl describe vm "$vm_name" -n "$namespace" | grep -A 20 "Status:"
  
  # Check VMI events
  kubectl get events -n "$namespace" | grep "$vm_name"
  
  # Check DataVolume status
  kubectl get datavolumes -n "$namespace" | grep "$vm_name"
  
  # Check pod logs (if VM is running)
  if kubectl get pod -n "$namespace" -l "kubevirt.io/vm=$vm_name" &>/dev/null; then
    kubectl logs -n "$namespace" -l "kubevirt.io/vm=$vm_name"
  fi
  
  # Attempt to start VM
  virtctl start "$vm_name" -n "$namespace"
}
```

#### SSH Connectivity Issues

```bash
# Troubleshoot SSH connectivity
debug_ssh_connectivity() {
  local vm_name="$1"
  local namespace="$2"
  
  echo "Debugging SSH connectivity for $vm_name"
  
  # Check VM IP address
  VM_IP=$(kubectl get vm "$vm_name" -n "$namespace" -o jsonpath='{.status.interfaces[0].ipAddress}')
  echo "VM IP: $VM_IP"
  
  # Test network connectivity
  kubectl run debug-pod --rm -i --tty --image=nicolaka/netshoot -- \
    ping -c 3 "$VM_IP"
  
  # Check if SSH service is running via console
  echo "Checking SSH service via console..."
  virtctl console "$vm_name" -n "$namespace" --timeout 30 <<EOF
systemctl status sshd
exit
EOF
  
  # Test SSH port specifically
  kubectl run debug-pod --rm -i --tty --image=nicolaka/netshoot -- \
    nc -zv "$VM_IP" 22
}
```

#### Performance Issues

```bash
# Debug VM performance issues
debug_vm_performance() {
  local vm_name="$1"
  local namespace="$2"
  
  echo "Analyzing performance for $vm_name"
  
  # Check resource allocation
  kubectl describe vm "$vm_name" -n "$namespace" | grep -A 10 "Resources:"
  
  # Check current usage
  kubectl top vm "$vm_name" -n "$namespace"
  
  # Check node resources
  NODE=$(kubectl get vm "$vm_name" -n "$namespace" -o jsonpath='{.status.nodeName}')
  kubectl describe node "$NODE" | grep -A 10 "Allocated resources:"
  
  # Check VM internal performance
  virtctl ssh admin@"$vm_name" -n "$namespace" "
    top -bn1 | head -20
    free -h
    df -h
    iostat -x 1 3
  "
}
```

## Summary: Complete Virtualization Management

The integration of `kubectl-mtv` and `virtctl` provides comprehensive virtualization management:

### Migration Phase (kubectl-mtv)
- Discover and assess source environments
- Plan and execute migrations
- Monitor migration progress
- Handle migration-specific configurations

### Operations Phase (virtctl)
- Manage VM lifecycle (start/stop/restart)
- Provide console and SSH access
- Handle resource management and optimization
- Implement backup and disaster recovery
- Maintain security and compliance

### Continuous Integration
Both tools work with the same Kubernetes resources, enabling:
- Seamless handoff from migration to operations
- Consistent resource management
- Unified monitoring and alerting
- Integrated backup and disaster recovery

This complete toolchain enables organizations to migrate from traditional virtualization platforms to Kubernetes while maintaining operational excellence throughout the VM lifecycle.

---

*Previous: [Chapter 19: Model Context Protocol (MCP) Server Integration](/kubectl-mtv/19-model-context-protocol-mcp-server-integration)*  
*Next: [Chapter 21: System Health Checks](/kubectl-mtv/21-system-health-checks)*
