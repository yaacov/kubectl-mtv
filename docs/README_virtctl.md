# VirtCtl CLI Usage Guide

## Overview

VirtCtl is a command-line utility for managing KubeVirt resources. While basic VirtualMachineInstance operations can be performed with kubectl, virtctl provides advanced features such as serial and graphical console access, convenience commands for starting/stopping VirtualMachineInstances, live migration, and uploading VM disk images.

This document provides a CLI reference for using virtctl commands from the bash terminal to manage virtual machines in Kubernetes clusters running KubeVirt.

## What is KubeVirt?

KubeVirt enables running virtual machines on Kubernetes by:
- Extending Kubernetes with VM-specific resources
- Providing VM lifecycle management through standard Kubernetes APIs
- Supporting both traditional VMs and containerized workloads in the same cluster
- Offering advanced features like live migration, snapshots, and hot-plugging

## Installation

### VirtCtl Installation

VirtCtl can be installed in two ways:

**Binary Installation:**
```bash
export VERSION=$(curl https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt)
wget https://github.com/kubevirt/kubevirt/releases/download/${VERSION}/virtctl-${VERSION}-linux-amd64
chmod +x virtctl-${VERSION}-linux-amd64
sudo mv virtctl-${VERSION}-linux-amd64 /usr/local/bin/virtctl
```

**Krew Plugin Installation:**
```bash
kubectl krew install virt
# Then use as: kubectl virt <command>
```

## CLI Command Categories

### 1. VM Lifecycle Management

**Power state management for virtual machines**

VirtCtl provides comprehensive VM lifecycle operations:
- **start**: Start a stopped VM
- **stop**: Stop a running VM (with graceful/forced options)
- **restart**: Restart a VM (stop then start)
- **pause**: Pause a running VM
- **unpause**: Resume a paused VM
- **migrate**: Live migrate a VM to another node
- **soft-reboot**: Perform soft reboot of VM guest OS

**Key Features:**
- Graceful and forced operations
- Configurable grace periods
- Dry-run support for testing
- Node targeting for migrations

**CLI Commands:**

```bash
# Start a VM
virtctl start web-server -n production

# Stop a VM gracefully with 30 second grace period
virtctl stop web-server -n production --grace-period 30

# Force stop a VM immediately
virtctl stop web-server -n production --force

# Restart a VM
virtctl restart web-server -n production

# Pause a VM
virtctl pause web-server -n production

# Unpause a VM
virtctl unpause web-server -n production

# Live migrate VM to specific node
virtctl migrate database-vm -n production --node worker-2

# Soft reboot VM guest OS
virtctl soft-reboot web-server -n production

# Dry run migration
virtctl migrate test-vm --dry-run
```

### 2. VM Creation

**Comprehensive VM creation with full configuration support**

VirtCtl supports advanced VM creation with:
- Instance types and preferences with automatic inference
- Multiple volume sources (PVCs, DataVolumes, DataSources, container disks)
- Complete cloud-init integration
- Resource requirements and limits
- Access credentials for SSH and password injection
- Annotations and labels for metadata

**Volume Types:**
- **volume-import**: Clone from PVCs, DataVolumes, or DataSources
- **volume-containerdisk**: Container images with OS
- **volume-blank**: Empty volumes for data storage
- **volume-pvc**: Direct PVC mounting
- **volume-sysprep**: Windows sysprep configuration

**CLI Commands:**

```bash
# Create simple VM with random name
virtctl create vm

# Create VM with specific name and instance type
virtctl create vm --name web-server --instancetype u1.medium

# Create VM with DataSource and automatic inference
virtctl create vm --name fedora-vm \
  --volume-import type:ds,src:fedora-42-cloud,name:rootdisk,size:20Gi \
  --infer-instancetype --infer-preference

# Create VM with cloud-init user and SSH key
virtctl create vm --name ubuntu-vm \
  --volume-import type:ds,src:ubuntu-22-04 \
  --user ubuntu \
  --ssh-key "ssh-rsa AAAA..."

# Create VM with multiple volumes
virtctl create vm --name multi-disk-vm \
  --volume-containerdisk src:quay.io/containerdisks/fedora:latest \
  --volume-blank name:data-disk,size:100Gi

# Create VM with specific resources
virtctl create vm --name resource-vm \
  --instancetype u1.large \
  --memory 8Gi \
  --cpu 4

# Create VM with access credentials from secret
virtctl create vm --name secure-vm \
  --user myuser \
  --access-cred type:ssh,src:my-keys

# Create VM with sysprep (Windows)
virtctl create vm --name windows-vm \
  --volume-import type:dv,src:windows-server-2019 \
  --volume-sysprep src:windows-config

# Pipe output to kubectl to create on cluster
virtctl create vm --name production-vm \
  --instancetype u1.medium \
  --volume-import type:ds,src:fedora-cloud | kubectl create -f -
```

### 3. Volume Management

**Hot-plug storage operations for running VMs**

VirtCtl supports dynamic volume operations:
- **addvolume**: Attach volumes to running VMs
- **removevolume**: Detach volumes from running VMs

**Supported Volume Sources:**
- PVCs (existing PersistentVolumeClaims)
- Blank volumes (dynamically created)
- Container disks from registries

**CLI Commands:**

```bash
# Add existing PVC to running VM
virtctl addvolume database-vm \
  --volume-name backup-storage \
  --volume-source pvc:backup-pvc \
  -n production

# Add blank volume with specific size and storage class
virtctl addvolume web-server \
  --volume-name temp-storage \
  --volume-source blank:temp-vol \
  --size 50Gi \
  --storage-class fast-ssd \
  -n default

# Add volume with disk configuration
virtctl addvolume database-vm \
  --volume-name data-disk \
  --volume-source pvc:data-pvc \
  --bus virtio \
  --cache writeback \
  --serial disk001 \
  -n production

# Add container disk from registry
virtctl addvolume test-vm \
  --volume-name tools \
  --volume-source registry:quay.io/tools/diagnostic:latest \
  -n testing

# Persist volume changes to VM spec
virtctl addvolume app-vm \
  --volume-name persistent-data \
  --volume-source pvc:app-data \
  --persist \
  -n apps

# Remove volume from running VM
virtctl removevolume database-vm \
  --volume-name temp-storage \
  -n production

# Remove volume with dry-run
virtctl removevolume test-vm \
  --volume-name test-volume \
  --dry-run \
  -n testing
```

### 4. Image Operations

**Advanced disk and image management**

VirtCtl provides comprehensive image management:
- **image-upload**: Upload disk images to PVCs/DataVolumes
- **vmexport**: Export VMs and create backups
- **guestfs**: Libguestfs operations for disk inspection
- **memory-dump**: Create memory dumps for debugging

**CLI Commands:**

```bash
# Upload disk image to new PVC
virtctl image-upload pvc vm-disk \
  --image-path /path/to/vm-image.qcow2 \
  --size 20Gi \
  --storage-class fast-ssd \
  -n default

# Upload with specific configuration
virtctl image-upload pvc windows-disk \
  --image-path /images/windows-server.qcow2 \
  --size 60Gi \
  --storage-class premium \
  --block-volume \
  -n production

# Upload with insecure connection (testing only)
virtctl image-upload pvc test-disk \
  --image-path /tmp/test.qcow2 \
  --size 10Gi \
  --insecure \
  -n testing

# Export VM to file
virtctl vmexport download production-vm \
  --output production-vm-backup.yaml \
  --port-forward \
  -n production

# Export with custom configuration
virtctl vmexport download web-server \
  --output /backups/web-server-$(date +%Y%m%d).yaml \
  --insecure \
  -n default

# Start libguestfs shell for disk inspection
virtctl guestfs vm-root-disk -n default

# Run libguestfs with KVM acceleration
virtctl guestfs vm-root-disk \
  --kvm \
  --pull-policy IfNotPresent \
  -n production

# Create memory dump
virtctl memory-dump get debug-vm \
  --create-claim \
  --claim-name debug-memory-dump \
  -n debugging

# Create memory dump with specific storage
virtctl memory-dump get problem-vm \
  --create-claim \
  --claim-name problem-vm-dump \
  --storage-class fast-ssd \
  -n production
```

### 5. Service Management

**Network services and connectivity management**

VirtCtl can expose VMs as Kubernetes services:
- **expose**: Create services for VM access
- Service types: ClusterIP, NodePort, LoadBalancer

**CLI Commands:**

```bash
# Expose VM as ClusterIP service (default)
virtctl expose vm web-server \
  --name web-service \
  --port 80 \
  --target-port 8080 \
  -n default

# Expose as NodePort service
virtctl expose vm api-server \
  --name api-service \
  --port 443 \
  --target-port 8443 \
  --type NodePort \
  --node-port 30443 \
  -n production

# Expose as LoadBalancer service
virtctl expose vm web-app \
  --name web-app-lb \
  --port 80 \
  --target-port 80 \
  --type LoadBalancer \
  --protocol TCP \
  -n apps

# Expose with custom service name
virtctl expose vm database \
  --name db-service \
  --port 5432 \
  --target-port 5432 \
  --type ClusterIP \
  -n databases

# Remove service exposure (use kubectl)
kubectl delete service web-service -n default
```

### 6. VM Access and Diagnostics

**VM console access and system monitoring**

VirtCtl provides various ways to access and monitor VMs:
- **console**: Serial console access
- **ssh**: SSH connection to VMs
- **vnc**: VNC graphical console
- **port-forward**: Port forwarding for services
- **scp**: Secure file transfer
- **guestosinfo**: Guest OS information
- **fslist**: Filesystem information
- **userlist**: Active user sessions

**CLI Commands:**

```bash
# Access VM console
virtctl console web-server -n production

# Access console with timeout
virtctl console debug-vm --timeout 120s -n development

# SSH to VM
virtctl ssh ubuntu@web-server -n production

# SSH with identity file
virtctl ssh user@vm-name \
  -i ~/.ssh/id_rsa \
  -n production

# SSH with verbose output
virtctl ssh root@debug-vm -v 2 -n testing

# VNC console access
virtctl vnc desktop-vm -n workstations

# VNC proxy only (no client launch)
virtctl vnc graphics-vm --proxy-only -n design

# Port forwarding
virtctl port-forward web-app 8080:80 -n apps

# Port forwarding with specific address
virtctl port-forward database 5432:5432 \
  --address 0.0.0.0 \
  -n databases

# Multiple port forwarding
virtctl port-forward multi-service \
  8080:80 2222:22 \
  -n services

# Copy files to VM
virtctl scp /local/file.txt vm/target-vm/default:/remote/file.txt

# Copy files from VM
virtctl scp vm/source-vm/default:/remote/data.txt /local/data.txt

# Copy directories recursively
virtctl scp -r /local/directory/ vm/target-vm/default:/remote/

# Get guest OS information
virtctl guestosinfo production-vm -n default

# Get guest OS info in JSON format
virtctl guestosinfo production-vm -o json -n production

# List filesystems in VM
virtctl fslist file-server -n storage

# List active users in VM
virtctl userlist web-server -n production

# Get version information
virtctl version
```

### 7. Resource Creation

**Create instance types and preferences**

VirtCtl can create reusable VM configurations:
- **create instancetype**: CPU and memory specifications
- **create preference**: VM preferences and features
- Both cluster-scoped and namespaced variants

**CLI Commands:**

```bash
# Create cluster-scoped instance type
virtctl create cluster-instancetype \
  --name high-performance \
  --cpu 8 \
  --memory 16Gi

# Create namespaced instance type
virtctl create instancetype \
  --name medium-workload \
  --cpu 4 \
  --memory 8Gi \
  -n production

# Create instance type with GPU
virtctl create cluster-instancetype \
  --name gpu-workload \
  --cpu 8 \
  --memory 32Gi \
  --gpu nvidia.com/gpu

# Create instance type with IO threads
virtctl create cluster-instancetype \
  --name io-intensive \
  --cpu 16 \
  --memory 64Gi \
  --iothread-policy auto

# Create cluster-scoped preference
virtctl create cluster-preference \
  --name linux-server \
  --machine-type q35

# Create namespaced preference
virtctl create preference \
  --name windows-desktop \
  --machine-type q35 \
  --cpu-sockets 2 \
  --cpu-cores 2 \
  --cpu-threads 2 \
  -n workstations

# Create preference with CPU features
virtctl create cluster-preference \
  --name secure-vm \
  --machine-type q35 \
  --cpu-feature nx \
  --cpu-feature smx

# Output as YAML without creating
virtctl create cluster-instancetype \
  --name test-config \
  --cpu 2 \
  --memory 4Gi \
  -o yaml
```

### 8. DataSource Management

**Managing VM boot images and templates**

DataSources provide ready-to-use OS images. While virtctl doesn't directly create DataSources, you can manage them using kubectl and CDI tools.

**Common Operations:**

```bash
# List available DataSources
kubectl get datasources --all-namespaces

# List DataSources with details
kubectl get datasources -o wide -n default

# Get specific DataSource details
kubectl describe datasource fedora-42-cloud -n default

# View DataSource YAML
kubectl get datasource ubuntu-22-04 -o yaml -n default

# Create DataSource from HTTP URL (using kubectl)
cat <<EOF | kubectl apply -f -
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataSource
metadata:
  name: fedora-42-cloud
  namespace: default
  annotations:
    instancetype.kubevirt.io/default-instancetype: u1.small
    instancetype.kubevirt.io/default-preference: fedora
spec:
  source:
    http:
      url: "https://download.fedoraproject.org/pub/fedora/linux/releases/42/Cloud/x86_64/images/Fedora-Cloud-Base-42-1.6.x86_64.qcow2"
  storage:
    resources:
      requests:
        storage: 5Gi
    storageClassName: fast-ssd
EOF

# Create DataSource from container registry
cat <<EOF | kubectl apply -f -
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataSource
metadata:
  name: ubuntu-22-04
  namespace: default
spec:
  source:
    registry:
      url: "docker://quay.io/containerdisks/ubuntu:22.04"
  storage:
    resources:
      requests:
        storage: 8Gi
EOF

# Clone existing DataSource
kubectl get datasource fedora-42-cloud -o yaml | \
  sed 's/name: fedora-42-cloud/name: fedora-42-custom/' | \
  kubectl apply -f -

# Delete DataSource
kubectl delete datasource old-image -n default
```

### 9. Resource Discovery

**Discover available cluster resources for VM configuration**

Use kubectl commands to discover resources available for VM creation:

```bash
# List cluster instance types
kubectl get virtualmachineclusterinstancetypes

# List namespaced instance types
kubectl get virtualmachineinstancetypes --all-namespaces

# Get instance type details
kubectl describe virtualmachineclusterinstancetype u1.medium

# List cluster preferences
kubectl get virtualmachineclusterpreferences

# List namespaced preferences
kubectl get virtualmachinepreferences -n production

# Get preference details
kubectl describe virtualmachineclusterpreference fedora

# List available DataSources
kubectl get datasources --all-namespaces

# Show DataSource details
kubectl get datasources -o wide -n default

# List DataSources with labels
kubectl get datasources --show-labels -n default

# Find DataSources by label selector
kubectl get datasources -l os=linux --all-namespaces

# List storage classes suitable for VMs
kubectl get storageclasses

# Show storage class details
kubectl describe storageclass fast-ssd

# Find storage classes with virtualization annotation
kubectl get storageclasses \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.annotations.storageclass\.kubevirt\.io/is-default-virt-class}{"\n"}{end}'

# List all KubeVirt resources
kubectl api-resources --api-group=kubevirt.io

# List all CDI resources (for DataSources)
kubectl api-resources --api-group=cdi.kubevirt.io
```

## Common Workflow Patterns

### 1. Complete VM Deployment

```bash
# 1. Discover available DataSources
kubectl get datasources --all-namespaces

# 2. Create VM with DataSource and inference
virtctl create vm --name web-server \
  --volume-import type:ds,src:ubuntu-22-04-cloud,name:rootdisk \
  --infer-instancetype --infer-preference \
  --user ubuntu \
  --ssh-key "ssh-rsa AAAA..." \
  -n production | kubectl create -f -

# 3. Start the VM
virtctl start web-server -n production

# 4. Expose as service
virtctl expose vm web-server \
  --name web-service \
  --port 80 \
  --type LoadBalancer \
  -n production

# 5. Access the VM
virtctl console web-server -n production
# or
virtctl ssh ubuntu@web-server -n production
```

### 2. DataSource Management Workflow

```bash
# 1. Create DataSource from official image
cat <<EOF | kubectl apply -f -
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataSource
metadata:
  name: ubuntu-22-04-base
  namespace: default
spec:
  source:
    registry:
      url: "docker://quay.io/containerdisks/ubuntu:22.04"
EOF

# 2. Clone for customization
kubectl get datasource ubuntu-22-04-base -o yaml | \
  sed 's/name: ubuntu-22-04-base/name: ubuntu-22-04-custom/' | \
  kubectl apply -f -

# 3. Use in VM creation
virtctl create vm --name app-server \
  --volume-import type:ds,src:ubuntu-22-04-custom,name:rootdisk \
  -n applications | kubectl create -f -
```

### 3. VM Maintenance Workflow

```bash
# 1. Create memory dump for debugging
virtctl memory-dump get production-vm \
  --create-claim \
  --claim-name production-vm-dump \
  -n production

# 2. Live migrate VM to different node
virtctl migrate production-vm --node worker-3 -n production

# 3. Add temporary storage for maintenance
virtctl addvolume production-vm \
  --volume-name maintenance-disk \
  --volume-source blank:temp-storage \
  --size 50Gi \
  -n production

# 4. Monitor system health
virtctl guestosinfo production-vm -n production
virtctl fslist production-vm -n production

# 5. Remove temporary storage after maintenance
virtctl removevolume production-vm \
  --volume-name maintenance-disk \
  -n production
```

### 4. Development Environment Setup

```bash
# 1. Create development VM with tools
virtctl create vm --name dev-workstation \
  --instancetype u1.large \
  --volume-import type:ds,src:fedora-42-cloud \
  --volume-blank name:workspace,size:100Gi \
  --user developer \
  --ssh-key "$(cat ~/.ssh/id_rsa.pub)" | kubectl create -f -

# 2. Start and wait for VM
virtctl start dev-workstation -n development

# 3. Port forward for development services
virtctl port-forward dev-workstation 8080:8080 3000:3000 -n development &

# 4. SSH access for development
virtctl ssh developer@dev-workstation -n development

# 5. Copy files to/from VM
virtctl scp -r ./project/ vm/dev-workstation/development:/home/developer/
```

## Best Practices

### Resource Management
- Use instance types and preferences for consistent VM sizing
- Leverage DataSource inference for automatic resource selection
- Monitor resource usage through diagnostic commands
- Name resources consistently across environments

```bash
# Good: Use consistent naming
virtctl create vm --name web-prod-01 --instancetype u1.medium
virtctl create vm --name web-prod-02 --instancetype u1.medium

# Good: Use inference for consistency
virtctl create vm --name app-vm \
  --volume-import type:ds,src:standard-ubuntu \
  --infer-instancetype --infer-preference
```

### Storage Operations
- Test volume operations with dry-run before execution
- Use appropriate storage classes for different workload types
- Backup important VMs before disk modifications
- Monitor disk usage regularly

```bash
# Good: Test before executing
virtctl addvolume production-vm --volume-name test --dry-run

# Good: Use appropriate storage classes
virtctl addvolume db-vm \
  --volume-name db-storage \
  --volume-source blank:db-data \
  --size 500Gi \
  --storage-class fast-ssd

# Good: Monitor filesystem usage
virtctl fslist production-vm -n production
```

### Network Configuration
- Use appropriate service types for access patterns
- Test connectivity after exposing services
- Use port forwarding for development/debugging
- Secure access with proper authentication

```bash
# Good: Internal services use ClusterIP
virtctl expose vm internal-api --type ClusterIP

# Good: External services use LoadBalancer
virtctl expose vm public-web --type LoadBalancer

# Good: Development port forwarding
virtctl port-forward dev-vm 8080:80 --address 127.0.0.1
```

### VM Access and Security
- Use SSH keys instead of passwords
- Configure cloud-init for automated setup
- Regular security updates
- Monitor VM access and sessions

```bash
# Good: SSH key authentication
virtctl create vm --name secure-vm \
  --user admin \
  --ssh-key "$(cat ~/.ssh/id_rsa.pub)" \
  --ga-manage-ssh

# Good: Monitor active sessions
virtctl userlist production-vm -n production

# Good: Regular OS information checks
virtctl guestosinfo production-vm -n production
```

### Operational Excellence
- Use descriptive VM and resource names
- Implement regular backup procedures
- Monitor VM health proactively
- Document VM configurations and purposes
- Use automation for repetitive tasks

```bash
# Good: Descriptive naming
virtctl create vm --name wordpress-prod-db-primary

# Good: Regular health checks
for vm in $(kubectl get vm -o name); do
  echo "=== $vm ==="
  virtctl guestosinfo ${vm#*/} -n production
done

# Good: Backup procedures
virtctl memory-dump get critical-vm \
  --create-claim \
  --claim-name "critical-vm-backup-$(date +%Y%m%d)"
```

## Integration with kubectl-mtv

VirtCtl complements kubectl-mtv for complete VM lifecycle management:

### Shared Resources
- VMs migrated via MTV can be managed with virtctl commands
- DataSources created for MTV migrations are reusable
- Instance types and preferences work across both tools

```bash
# MTV creates DataSources during migration
kubectl get datasources -l forklift.io/name

# Use MTV DataSources with virtctl
virtctl create vm --name new-vm \
  --volume-import type:ds,src:migrated-vm-datasource \
  --infer-instancetype
```

### Operational Workflows
- Use virtctl for post-migration VM management
- Leverage MTV inventory data for planning
- Apply virtctl diagnostics to MTV-migrated VMs

```bash
# Manage MTV-migrated VM with virtctl
virtctl start migrated-web-server -n migration-target
virtctl addvolume migrated-web-server --volume-name logs --size 20Gi
virtctl expose vm migrated-web-server --port 80 --type LoadBalancer

# Diagnose migrated VMs
virtctl guestosinfo migrated-database -n migration-target
virtctl fslist migrated-database -n migration-target
```

### Resource Discovery
- Discover MTV-created resources using kubectl
- Identify migration artifacts and configurations
- Plan new VMs based on migration patterns

```bash
# Find migration-created resources
kubectl get datasources -l migration=mtv
kubectl get instancetypes -l created-by=forklift

# Discover available resources for VM creation
kubectl get storageclasses
kubectl get datasources --all-namespaces
kubectl get virtualmachineclusterinstancetypes
```

This comprehensive CLI guide provides complete coverage of virtctl usage patterns for managing KubeVirt virtual machines in Kubernetes environments, with practical examples and best practices for real-world deployments.
