---
layout: page
title: "Chapter 6: VDDK Image Creation and Configuration"
---

VMware Virtual Disk Development Kit (VDDK) provides optimized disk transfer capabilities for VMware vSphere migrations. This chapter covers creating VDDK container images and configuring them for maximum performance.

For background information on VDDK and its integration with Forklift, see the [official VDDK documentation](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#creating-a-vddk-image).

## Why VDDK is Recommended for VMware Disk Transfers

### Performance Benefits

VDDK provides significant performance improvements over standard disk transfer methods:

- **Optimized Data Transfer**: Direct access to VMware's optimized disk I/O APIs
- **Reduced Network Overhead**: Efficient data streaming and compression
- **Better Throughput**: Can achieve 2-5x faster transfer speeds compared to standard methods
- **Resource Efficiency**: Lower CPU and memory usage during transfers

### Technical Advantages

- **Native VMware Integration**: Uses VMware's official SDK for optimal compatibility
- **Advanced Features**: Support for changed block tracking (CBT) and incremental transfers
- **Error Handling**: Better error detection and recovery mechanisms
- **Storage Array Integration**: Support for storage array offloading when available

### When to Use VDDK

- **Production Migrations**: Always recommended for production VMware environments
- **Large VMs**: Essential for VMs with large disk sizes (>100GB)
- **Performance-Critical**: When migration time is a critical factor
- **Storage Array Offloading**: When using compatible storage arrays with offloading capabilities

## Prerequisites for Building the Image

### System Requirements

Before building VDDK images, ensure you have:

- **Container Runtime**: Podman or Docker installed and working
- **Kubernetes Registry Access**: Access to a container registry (internal or external)
- **File System**: A file system that preserves symbolic links (symlinks)
- **Network Access**: If using external registries, ensure KubeVirt can access them

### VMware License Compliance

> **Important License Notice**  
> Storing VDDK images in public registries might violate VMware license terms. Always use private registries and ensure compliance with VMware licensing requirements.

### Download VDDK SDK

1. **Visit Broadcom Developer Portal**: https://developer.broadcom.com/sdks/vmware-virtual-disk-development-kit-vddk/latest
2. **Download VDDK**: Get the VMware Virtual Disk Development Kit (VDDK) tar.gz file
3. **Supported Versions**: VDDK 8.0.x is recommended for optimal performance
4. **File Verification**: Verify the downloaded file integrity if checksums are provided

### Runtime Prerequisites

#### Podman (Recommended)

```bash
# Install podman (RHEL/Fedora)
sudo dnf install podman

# Install podman (Ubuntu/Debian)  
sudo apt-get install podman

# Verify podman installation
podman --version
```

#### Docker (Alternative)

```bash
# Install Docker (Ubuntu/Debian)
sudo apt-get update
sudo apt-get install docker.io

# Install Docker (RHEL/Fedora)
sudo dnf install docker

# Start and enable Docker
sudo systemctl start docker
sudo systemctl enable docker

# Verify Docker installation
docker --version
```

## How-To: Building the VDDK Image

### Basic VDDK Image Creation

The `kubectl mtv create vddk-image` command automates the entire VDDK image build process:

```bash
# Basic VDDK image creation
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag quay.io/your-registry/vddk:8.0.1

# With automatic push to registry
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag quay.io/your-registry/vddk:8.0.1 \
  --push
```

### Command Syntax and Options

```bash
kubectl mtv create vddk-image [OPTIONS]
```

#### Required Flags

- `--tar PATH`: Path to VMware VDDK tar.gz file (required)
- `--tag IMAGE:TAG`: Container image tag (required)

#### Optional Flags

- `--build-dir PATH`: Custom build directory (uses temp directory if not specified)
- `--runtime RUNTIME`: Container runtime (auto, podman, docker) - default: auto
- `--platform ARCH`: Target platform (amd64, arm64) - default: amd64
- `--dockerfile PATH`: Path to custom Dockerfile (uses default if not specified)
- `--push`: Push image to registry after successful build

### Detailed Build Examples

#### Standard Production Build

```bash
# Production VDDK image with push
kubectl mtv create vddk-image \
  --tar ~/downloads/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag quay.io/company/vddk:8.0.1 \
  --runtime podman \
  --platform amd64 \
  --push
```

#### Custom Build Directory

```bash
# Use specific build directory for large environments
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag harbor.company.com/migration/vddk:latest \
  --build-dir /tmp/vddk-build \
  --runtime podman \
  --push
```

#### Multi-Architecture Build

```bash
# Build for ARM64 architecture
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag quay.io/company/vddk:8.0.1-arm64 \
  --platform arm64 \
  --runtime podman \
  --push

# Build for AMD64 architecture (default)
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag quay.io/company/vddk:8.0.1-amd64 \
  --platform amd64 \
  --runtime podman \
  --push
```

#### Custom Dockerfile Build

```bash
# Create custom Dockerfile with additional tools
cat > custom-vddk.dockerfile << 'EOF'
FROM registry.redhat.io/ubi8/ubi:latest

# Install additional debugging tools
RUN dnf install -y tcpdump netstat-ng && dnf clean all

# Copy VDDK libraries (will be handled by kubectl-mtv)
# Additional customizations can be added here

EOF

# Build with custom Dockerfile
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag quay.io/company/vddk:8.0.1-custom \
  --dockerfile custom-vddk.dockerfile \
  --push
```

#### Different Container Runtimes

```bash
# Force use of Docker
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag localhost:5000/vddk:8.0.1 \
  --runtime docker \
  --push

# Force use of Podman
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag quay.io/company/vddk:8.0.1 \
  --runtime podman \
  --push

# Auto-detect runtime (default)
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag quay.io/company/vddk:8.0.1 \
  --runtime auto \
  --push
```

### Build Process Verification

```bash
# Verify image was built successfully
podman images | grep vddk
# or
docker images | grep vddk

# Test image functionality
podman run --rm quay.io/company/vddk:8.0.1 /usr/bin/vmware-vdiskmanager -h

# Verify image layers and size
podman inspect quay.io/company/vddk:8.0.1
```

## Setting the MTV_VDDK_INIT_IMAGE Environment Variable

### Setting the Default VDDK Image

The `MTV_VDDK_INIT_IMAGE` environment variable provides a global default for all vSphere provider creation:

```bash
# Set the default VDDK image
export MTV_VDDK_INIT_IMAGE=quay.io/your-registry/vddk:8.0.1

# Verify the environment variable
echo $MTV_VDDK_INIT_IMAGE
```

### Persistent Configuration

#### Shell Profile Configuration

```bash
# Add to ~/.bashrc for bash users
echo 'export MTV_VDDK_INIT_IMAGE=quay.io/your-registry/vddk:8.0.1' >> ~/.bashrc
source ~/.bashrc

# Add to ~/.zshrc for zsh users
echo 'export MTV_VDDK_INIT_IMAGE=quay.io/your-registry/vddk:8.0.1' >> ~/.zshrc
source ~/.zshrc
```

#### System-wide Configuration

```bash
# Create system-wide environment file
sudo tee /etc/environment.d/mtv-vddk.conf << EOF
MTV_VDDK_INIT_IMAGE=quay.io/your-registry/vddk:8.0.1
EOF

# Or add to /etc/profile.d/
sudo tee /etc/profile.d/mtv-vddk.sh << EOF
export MTV_VDDK_INIT_IMAGE=quay.io/your-registry/vddk:8.0.1
EOF
```

#### Container/Pod Environment

```bash
# For containerized kubectl-mtv usage
docker run -e MTV_VDDK_INIT_IMAGE=quay.io/your-registry/vddk:8.0.1 \
  kubectl-mtv-image create provider vsphere-prod --type vsphere

# In Kubernetes pods
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: kubectl-mtv
    env:
    - name: MTV_VDDK_INIT_IMAGE
      value: "quay.io/your-registry/vddk:8.0.1"
```

### Environment Variable Validation

```bash
# Verify environment variable is set
if [ -z "$MTV_VDDK_INIT_IMAGE" ]; then
  echo "MTV_VDDK_INIT_IMAGE is not set"
else
  echo "MTV_VDDK_INIT_IMAGE is set to: $MTV_VDDK_INIT_IMAGE"
fi

# Test with provider creation (should use default image)
kubectl mtv create provider vsphere-test --type vsphere \
  --url https://vcenter.test.com/sdk \
  --username admin \
  --password password123 \
  --dry-run
```

## Using the VDDK Image in Provider Creation

### Automatic VDDK Image Usage

When `MTV_VDDK_INIT_IMAGE` is set, providers automatically use the VDDK image:

```bash
# This will automatically use the VDDK image from MTV_VDDK_INIT_IMAGE
kubectl mtv create provider vsphere-auto --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password YourPassword
```

### Explicit VDDK Image Specification

Override the default VDDK image per provider:

```bash
# Use specific VDDK image for this provider
kubectl mtv create provider vsphere-custom --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password YourPassword \
  --vddk-init-image quay.io/company/vddk:8.0.2
```

### VDDK Performance Optimization

Enable advanced VDDK optimization features:

```bash
# Provider with VDDK AIO optimization
kubectl mtv create provider vsphere-optimized --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password YourPassword \
  --vddk-init-image quay.io/company/vddk:8.0.1 \
  --use-vddk-aio-optimization

# Provider with custom VDDK buffer settings
kubectl mtv create provider vsphere-tuned --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username administrator@vsphere.local \
  --password YourPassword \
  --vddk-init-image quay.io/company/vddk:8.0.1 \
  --use-vddk-aio-optimization \
  --vddk-buf-size-in-64k 128 \
  --vddk-buf-count 16
```

### VDDK Configuration Parameters

#### Buffer Size Optimization

The `--vddk-buf-size-in-64k` parameter controls the buffer size in 64KB units:

```bash
# Small VMs (default - automatic sizing)
--vddk-buf-size-in-64k 0

# Medium VMs (8MB buffer)
--vddk-buf-size-in-64k 128

# Large VMs (16MB buffer)
--vddk-buf-size-in-64k 256

# Very large VMs (32MB buffer)
--vddk-buf-size-in-64k 512
```

#### Buffer Count Tuning

The `--vddk-buf-count` parameter controls the number of parallel buffers:

```bash
# Low concurrency (default)
--vddk-buf-count 0

# Medium concurrency
--vddk-buf-count 8

# High concurrency
--vddk-buf-count 16

# Maximum concurrency (use with caution)
--vddk-buf-count 32
```

## Complete VDDK Workflow Examples

### Example 1: Enterprise Production Setup

```bash
# Step 1: Download VDDK from VMware
# (Manual step - download VMware-vix-disklib-distrib-8.0.1.tar.gz)

# Step 2: Build and push VDDK image
kubectl mtv create vddk-image \
  --tar ~/downloads/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag harbor.company.com/migration/vddk:8.0.1 \
  --runtime podman \
  --platform amd64 \
  --push

# Step 3: Set global environment variable
export MTV_VDDK_INIT_IMAGE=harbor.company.com/migration/vddk:8.0.1
echo 'export MTV_VDDK_INIT_IMAGE=harbor.company.com/migration/vddk:8.0.1' >> ~/.bashrc

# Step 4: Create optimized vSphere provider
kubectl mtv create provider vsphere-production --type vsphere \
  --url https://vcenter.prod.company.com/sdk \
  --username svc-migration@vsphere.local \
  --password $(cat /secure/vsphere-password) \
  --use-vddk-aio-optimization \
  --vddk-buf-size-in-64k 256 \
  --vddk-buf-count 16

# Step 5: Verify VDDK integration
kubectl mtv describe provider vsphere-production | grep -i vddk
```

### Example 2: Multi-Architecture Deployment

```bash
# Build for both AMD64 and ARM64
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag quay.io/company/vddk:8.0.1-amd64 \
  --platform amd64 \
  --push

kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag quay.io/company/vddk:8.0.1-arm64 \
  --platform arm64 \
  --push

# Create manifest list for multi-arch support
podman manifest create quay.io/company/vddk:8.0.1
podman manifest add quay.io/company/vddk:8.0.1 quay.io/company/vddk:8.0.1-amd64
podman manifest add quay.io/company/vddk:8.0.1 quay.io/company/vddk:8.0.1-arm64
podman manifest push quay.io/company/vddk:8.0.1

# Use multi-arch image
export MTV_VDDK_INIT_IMAGE=quay.io/company/vddk:8.0.1
```

### Example 3: Development and Testing

```bash
# Build test VDDK image with debugging tools
cat > debug-vddk.dockerfile << 'EOF'
FROM registry.redhat.io/ubi8/ubi:latest

# Install debugging and monitoring tools
RUN dnf install -y \
    tcpdump \
    netstat-ng \
    iotop \
    htop \
    strace \
    && dnf clean all

# Install additional utilities
RUN dnf install -y \
    curl \
    wget \
    telnet \
    && dnf clean all
EOF

# Build with debugging tools
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag localhost:5000/vddk:debug \
  --dockerfile debug-vddk.dockerfile \
  --runtime podman \
  --push

# Create test provider with debug image
kubectl mtv create provider vsphere-debug --type vsphere \
  --url https://vcenter-test.internal/sdk \
  --username administrator@vsphere.local \
  --password TestPassword \
  --vddk-init-image localhost:5000/vddk:debug \
  --provider-insecure-skip-tls \
  -n testing
```

## Advanced VDDK Configuration

### Registry Authentication

For private registries, ensure proper authentication:

```bash
# Login to registry
podman login quay.io
# or
docker login quay.io

# Create registry secret in Kubernetes
kubectl create secret docker-registry vddk-registry-secret \
  --docker-server=quay.io \
  --docker-username=your-username \
  --docker-password=your-password \
  --docker-email=your-email@company.com

# Use secret in migration namespace
kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "vddk-registry-secret"}]}'
```

### Performance Monitoring and Tuning

#### VDDK Performance Metrics

```bash
# Monitor VDDK pod performance during migration
kubectl top pods -l app=vddk

# Check VDDK container logs
kubectl logs -l app=vddk -f

# Monitor network usage
kubectl exec -it vddk-pod -- netstat -i
```

#### Buffer Tuning Guidelines

| VM Size | Buffer Size (64K units) | Buffer Count | Total Memory |
|---------|-------------------------|--------------|--------------|
| Small (< 50GB) | 64 | 4 | 16MB |
| Medium (50-200GB) | 128 | 8 | 64MB |
| Large (200-500GB) | 256 | 16 | 256MB |
| Very Large (> 500GB) | 512 | 32 | 1GB |

```bash
# Apply tuning based on VM size
kubectl mtv create provider vsphere-large-vms --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username admin \
  --password password \
  --vddk-init-image quay.io/company/vddk:8.0.1 \
  --use-vddk-aio-optimization \
  --vddk-buf-size-in-64k 512 \
  --vddk-buf-count 32
```

### Storage Array Offloading

For compatible storage arrays, enable offloading:

```bash
# Provider with storage array offloading
kubectl mtv create provider vsphere-offload --type vsphere \
  --url https://vcenter.example.com/sdk \
  --username admin \
  --password password \
  --vddk-init-image quay.io/company/vddk:8.0.1 \
  --use-vddk-aio-optimization

# Use with storage mapping that supports offloading
kubectl mtv create mapping storage offload-mapping \
  --source vsphere-offload \
  --target openshift \
  --storage-pairs "datastore1:fast-ssd;offloadPlugin=vsphere;offloadVendor=flashsystem"
```

## Troubleshooting VDDK Issues

### Common Build Problems

#### VDDK Tar File Issues

```bash
# Verify VDDK tar file integrity
file ~/VMware-vix-disklib-distrib-8.0.1.tar.gz
tar -tzf ~/VMware-vix-disklib-distrib-8.0.1.tar.gz | head -10

# Check file permissions
ls -la ~/VMware-vix-disklib-distrib-8.0.1.tar.gz
```

#### Container Runtime Issues

```bash
# Test container runtime
podman --version
podman run hello-world

# Check registry connectivity
podman login quay.io
podman pull registry.redhat.io/ubi8/ubi:latest
```

#### Build Directory Problems

```bash
# Check available disk space
df -h /tmp

# Use custom build directory with more space
mkdir -p /data/vddk-build
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag quay.io/company/vddk:8.0.1 \
  --build-dir /data/vddk-build
```

### Runtime Issues

#### Image Pull Problems

```bash
# Test image accessibility
podman pull quay.io/company/vddk:8.0.1

# Check image manifest
podman inspect quay.io/company/vddk:8.0.1

# Verify registry credentials
kubectl get secret vddk-registry-secret -o yaml
```

#### Performance Issues

```bash
# Monitor VDDK container resource usage
kubectl top pods -l app=vddk

# Check for resource constraints
kubectl describe pod vddk-pod-name

# Verify VDDK buffer settings
kubectl get provider vsphere-prod -o yaml | grep -A5 vddk
```

### Debug and Logging

```bash
# Enable verbose logging for VDDK builds
kubectl mtv create vddk-image \
  --tar ~/VMware-vix-disklib-distrib-8.0.1.tar.gz \
  --tag quay.io/company/vddk:debug \
  -v=2

# Test VDDK functionality
podman run --rm quay.io/company/vddk:8.0.1 \
  /usr/bin/vmware-vdiskmanager -h

# Check VDDK library versions
podman run --rm quay.io/company/vddk:8.0.1 \
  find /opt -name "*.so" -exec ls -la {} \;
```

## Best Practices Summary

### Security
- Always use private registries for VDDK images
- Implement proper registry authentication
- Follow VMware licensing requirements
- Use least-privilege access for registry credentials

### Performance
- Use VDDK 8.0.x for optimal performance
- Enable AIO optimization for production
- Tune buffer sizes based on VM characteristics
- Monitor and adjust based on actual performance

### Operations
- Automate VDDK image builds in CI/CD pipelines
- Version VDDK images with semantic versioning
- Test images thoroughly before production deployment
- Maintain separate images for different environments

## Next Steps

After configuring VDDK:

1. **Explore Inventory**: Discover VMs and resources in [Chapter 7: Inventory Management](07-inventory-management)
2. **Create Mappings**: Configure resource mappings in [Chapter 9: Mapping Management](09-mapping-management)
3. **Optimize Performance**: Learn advanced techniques in [Chapter 13: Migration Process Optimization](13-migration-process-optimization)
4. **Plan Migrations**: Create optimized plans in [Chapter 10: Migration Plan Creation](10-migration-plan-creation)

---

*Previous: [Chapter 5: Migration Host Management](/guide/05-migration-host-management)*  
*Next: [Chapter 7: Inventory Management](/guide/07-inventory-management)*
