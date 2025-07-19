# Installation Guide

This guide provides detailed instructions for installing kubectl-mtv, a kubectl plugin for migrating VMs from various virtualization platforms to KubeVirt on Kubernetes.

## Prerequisites

Before installing kubectl-mtv, ensure you have:

- **Kubernetes cluster** (version 1.23+) with Forklift or Migration Toolkit for Virtualization (MTV) installed
- **kubectl** installed and configured to access your cluster
- **Appropriate permissions** to access MTV/Forklift resources in your cluster

## Installation Methods

### Method 1: Krew Plugin Manager (Recommended)

[Krew](https://sigs.k8s.io/krew) is the recommended way to install kubectl plugins.

#### Install Krew (if not already installed)

```bash
# Install krew
(
  set -x; cd "$(mktemp -d)" &&
  OS="$(uname | tr '[:upper:]' '[:lower:]')" &&
  ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" &&
  KREW="krew-${OS}_${ARCH}" &&
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/${KREW}.tar.gz" &&
  tar zxvf "${KREW}.tar.gz" &&
  ./"${KREW}" install krew
)

# Add krew to your PATH
export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"
```

#### Install kubectl-mtv via Krew

```bash
# Install the mtv plugin
kubectl krew install mtv

# Verify installation
kubectl mtv --help
```

**Note**: Currently available for linux-amd64 architecture.

### Method 2: Download Release Binaries

Download pre-built binaries from the releases page.

#### Automated Download

```bash
# Set variables
REPO=yaacov/kubectl-mtv
ASSET=kubectl-mtv.tar.gz

# Get latest version
LATEST_VER=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep -m1 '"tag_name"' | cut -d'"' -f4)

# Download and extract
curl -L -o $ASSET https://github.com/$REPO/releases/download/$LATEST_VER/$ASSET
tar -xzf $ASSET

# Make executable and move to PATH
chmod +x kubectl-mtv
sudo mv kubectl-mtv /usr/local/bin/

# Verify installation
kubectl mtv --help
```

#### Manual Download

1. Go to the [Releases page](https://github.com/yaacov/kubectl-mtv/releases)
2. Download the appropriate archive for your platform
3. Extract the archive:

   ```bash
   tar -xzf kubectl-mtv.tar.gz
   ```

4. Move the binary to a directory in your PATH:

   ```bash
   sudo mv kubectl-mtv /usr/local/bin/
   ```

### Method 3: Package Manager (Fedora)

For Fedora 41, 42, and compatible amd64 systems:

```bash
# Enable the COPR repository
sudo dnf copr enable yaacov/kubesql

# Install kubectl-mtv
sudo dnf install kubectl-mtv

# Verify installation
kubectl mtv --help
```

### Method 4: Build from Source

#### Prerequisites for Building

- Go 1.23 or higher
- git
- make
- CGO enabled (default)
- musl-gcc (for static builds)

#### Build Steps

```bash
# Clone the repository
git clone https://github.com/yaacov/kubectl-mtv.git
cd kubectl-mtv

# Build the binary
make

# Install to GOPATH/bin (ensure it's in your PATH)
cp kubectl-mtv $GOPATH/bin/

# Or install to /usr/local/bin
sudo cp kubectl-mtv /usr/local/bin/
```

#### Build Static Binary

For containerized environments or systems without dynamic libraries:

```bash
# Install musl-gcc (if not already installed)
# On Ubuntu/Debian:
sudo apt-get install musl-tools

# On Fedora/RHEL:
sudo dnf install musl-gcc

# Build static binary
make kubectl-mtv-static
```

## Verification

After installation, verify that kubectl-mtv is working correctly:

```bash
# Check if kubectl recognizes the plugin
kubectl plugin list | grep mtv

# Test the plugin
kubectl mtv version

# Check help
kubectl mtv --help
```

## Configuration

### Kubeconfig

kubectl-mtv uses the same kubeconfig as kubectl. Ensure your kubeconfig is properly configured:

```bash
# Check current context
kubectl config current-context

# List available contexts
kubectl config get-contexts

# Switch context if needed
kubectl config use-context <your-context>
```

### Environment Variables

Set these environment variables for enhanced functionality:

```bash
# VDDK initialization image (for VMware migrations)
export MTV_VDDK_INIT_IMAGE=quay.io/your-registry/vddk:8.0.1

# Custom kubeconfig location (if not using default)
export KUBECONFIG=/path/to/your/kubeconfig
```

Add these to your shell profile (`.bashrc`, `.zshrc`, etc.) for persistence.

## Cluster Requirements

### Forklift/MTV Installation

kubectl-mtv requires Forklift (upstream) or Migration Toolkit for Virtualization (downstream) to be installed in your cluster.

#### Install Forklift (Upstream - Kubernetes)

```bash
# Install Forklift operator
kubectl apply -f https://github.com/kubev2v/forklift/releases/latest/download/forklift-operator.yaml

# Wait for operator to be ready
kubectl wait --for=condition=Available deployment/forklift-operator -n forklift-operator --timeout=300s

# Create Forklift controller
kubectl apply -f https://github.com/kubev2v/forklift/releases/latest/download/forklift-controller.yaml
```

#### Install MTV (Downstream - OpenShift)

For OpenShift environments, install MTV through the Operator Hub or CLI:

```bash
# Using OpenShift CLI
oc apply -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: mtv-operator
  namespace: openshift-migration
spec:
  channel: release-v2.6
  name: mtv-operator
  source: redhat-operators
  sourceNamespace: openshift-marketplace
EOF
```

### RBAC Permissions

Ensure your user has appropriate permissions to access MTV/Forklift resources:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: mtv-user
rules:
- apiGroups: ["forklift.konveyor.io"]
  resources: ["*"]
  verbs: ["get", "list", "create", "update", "patch", "delete", "watch"]
- apiGroups: [""]
  resources: ["secrets", "configmaps", "namespaces"]
  verbs: ["get", "list", "create", "update", "patch", "delete"]
```

## Troubleshooting

### Common Issues

#### Plugin Not Found

```bash
# Error: plugin "mtv" not found
# Solution: Ensure the binary is in your PATH and executable
which kubectl-mtv
chmod +x $(which kubectl-mtv)
```

#### Permission Denied

```bash
# Error: User cannot list resources
# Solution: Check RBAC permissions
kubectl auth can-i get plans.forklift.konveyor.io
kubectl auth can-i list providers.forklift.konveyor.io
```

#### Connection Issues

```bash
# Error: Unable to connect to cluster
# Solution: Verify kubeconfig and cluster connectivity
kubectl cluster-info
kubectl get nodes
```

### Debug Mode

Enable verbose output for troubleshooting:

```bash
# Use kubectl debug flags
kubectl mtv get providers -v=8

# Check cluster connectivity
kubectl mtv get providers --kubeconfig=/path/to/config
```

### Getting Help

- Check the [documentation](https://github.com/yaacov/kubectl-mtv/tree/main/docs)
- Review [demo examples](README_demo.md)
- Open an issue on [GitHub](https://github.com/yaacov/kubectl-mtv/issues)

## Next Steps

After successful installation:

1. **Read the [Usage Guide](README-usage.md)** for detailed command reference
2. **Follow the [Demo](README_demo.md)** for a complete migration example
3. **Review [VDDK setup](README_vddk.md)** for VMware migrations
4. **Check [Development Guide](README-development.md)** if you want to contribute

## Uninstallation

### Remove via Krew

```bash
kubectl krew uninstall mtv
```

### Remove Binary Installation

```bash
# Remove binary
sudo rm /usr/local/bin/kubectl-mtv
# or
rm $GOPATH/bin/kubectl-mtv
```

### Remove via Package Manager

```bash
# Fedora/RHEL
sudo dnf remove kubectl-mtv
```
