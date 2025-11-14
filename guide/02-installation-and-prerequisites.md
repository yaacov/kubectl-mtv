---
layout: page
title: "Chapter 2: Installation and Prerequisites"
---

## Prerequisites

Before installing `kubectl-mtv`, ensure your environment meets the following requirements:

### System Requirements

- **Operating System**: Linux, macOS, or Windows (Linux and macOS are primarily supported)
- **Architecture**: amd64 (x86_64) and arm64 architectures supported

### Kubernetes Environment

- **Kubernetes Cluster**: Version 1.23 or higher
- **Forklift/MTV Installation**: Either upstream Forklift or downstream Migration Toolkit for Virtualization (MTV) must be installed in your cluster
- **kubectl**: Latest stable version installed and configured to access your cluster
- **Cluster Access**: Appropriate RBAC permissions to access MTV/Forklift resources

### Development Prerequisites (Method 3 Only)

If building from source, you'll need:

- **Go**: Version 1.24 or higher (current requirement based on go.mod)
- **Git**: For cloning the repository
- **Make**: For using the build system

## Installation Methods

### Method 1: Krew Plugin Manager (Recommended)

[Krew](https://sigs.k8s.io/krew) is the recommended way to install kubectl plugins, providing easy installation and updates.

#### Step 1: Install Krew (if not already installed)

```bash
# Install Krew
(
  set -x; cd "$(mktemp -d)" &&
  OS="$(uname | tr '[:upper:]' '[:lower:]')" &&
  ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" &&
  KREW="krew-${OS}_${ARCH}" &&
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/${KREW}.tar.gz" &&
  tar zxvf "${KREW}.tar.gz" &&
  ./"${KREW}" install krew
)

# Add Krew to your PATH
export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"
```

Add the PATH export to your shell profile (`.bashrc`, `.zshrc`, etc.) for persistence:

```bash
echo 'export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

#### Step 2: Install kubectl-mtv via Krew

```bash
# Install the mtv plugin
kubectl krew install mtv

# Verify installation
kubectl mtv --help
```

**Note**: Available for multiple platforms including Linux (amd64, arm64), macOS (amd64, arm64), and Windows (amd64) through Krew.

### Method 2: Downloading Release Binaries

Download pre-built binaries directly from the GitHub releases page.

#### Automated Download Script

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

#### Manual Download Process

1. Visit the [Releases page](https://github.com/yaacov/kubectl-mtv/releases)
2. Download the appropriate archive for your platform:
   - Linux amd64: `kubectl-mtv-VERSION-linux-amd64.tar.gz`
   - Linux arm64: `kubectl-mtv-VERSION-linux-arm64.tar.gz`
   - macOS amd64: `kubectl-mtv-VERSION-darwin-amd64.tar.gz`
   - macOS arm64: `kubectl-mtv-VERSION-darwin-arm64.tar.gz`
   - Windows amd64: `kubectl-mtv-VERSION-windows-amd64.zip`

3. Extract the archive:
   ```bash
   # For tar.gz files
   tar -xzf kubectl-mtv-VERSION-PLATFORM.tar.gz
   
   # For zip files (Windows)
   unzip kubectl-mtv-VERSION-windows-amd64.zip
   ```

4. Move the binary to a directory in your PATH:
   ```bash
   # Linux/macOS
   sudo mv kubectl-mtv /usr/local/bin/
   
   # Or to user bin directory
   mv kubectl-mtv ~/.local/bin/
   ```

### Method 3: Building from Source

For development, customization, or platforms without pre-built binaries.

#### Step 1: Install Prerequisites

**On Ubuntu/Debian:**
```bash
# Install Go 1.24+
wget https://go.dev/dl/go1.24.7.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.7.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install build tools
sudo apt-get update
sudo apt-get install git make
```

**On Fedora/RHEL:**
```bash
# Install Go and build tools
sudo dnf install golang git make

# Verify Go version
go version  # Should be 1.24+
```

**On macOS:**
```bash
# Using Homebrew
brew install go git

# Verify Go version
go version  # Should be 1.24+
```

#### Step 2: Build the Binary

```bash
# Clone the repository
git clone https://github.com/yaacov/kubectl-mtv.git
cd kubectl-mtv

# Build the binary
make

# Install to GOPATH/bin (ensure it's in your PATH)
cp kubectl-mtv $(go env GOPATH)/bin/

# Or install to /usr/local/bin
sudo cp kubectl-mtv /usr/local/bin/
```

#### Step 3: Build Static Binary (Optional)

The default build already produces static binaries (CGO is disabled), but you can verify:

```bash
# The default make already creates a static binary
make

# Verify it's statically linked
ldd kubectl-mtv  # Should show "not a dynamic executable"
```

#### Cross-compilation

Build for different platforms:

```bash
# Build for different platforms
make build-linux-amd64
make build-linux-arm64
make build-darwin-amd64
make build-darwin-arm64
make build-windows-amd64

# Build all platforms
make build-all

# Create distribution archives
make dist-all
```

## Verification and Configuration

### Basic Verification

After installation, verify that `kubectl-mtv` is working correctly:

```bash
# Check if kubectl recognizes the plugin
kubectl plugin list | grep mtv

# Test the plugin
kubectl mtv version

# Check help and available commands
kubectl mtv --help

# List available subcommands
kubectl mtv
```

Expected output should show the version information and available commands.

### Kubeconfig Configuration

`kubectl-mtv` uses the same kubeconfig as `kubectl`. Ensure your kubeconfig is properly configured:

```bash
# Check current context
kubectl config current-context

# List available contexts
kubectl config get-contexts

# Switch context if needed
kubectl config use-context <your-context>

# Verify cluster connectivity
kubectl cluster-info
kubectl get nodes
```

## Global Flags Reference

`kubectl-mtv` provides several global flags that can be used with any command:

### Kubernetes Connection Flags

These flags are inherited from `kubectl` and control cluster connectivity:

- `--kubeconfig string`: Path to kubeconfig file (default: `$HOME/.kube/config`)
- `--context string`: The name of the kubeconfig context to use
- `--namespace string, -n`: Namespace to use for the operation
- `--server string`: Kubernetes API server address
- `--token string`: Bearer token for authentication
- `--user string`: The name of the kubeconfig user to use

### Output and Formatting Flags

Control how command output is displayed:

- `--output string, -o`: Output format (json, yaml, table)
- `--use-utc`: Format timestamps in UTC instead of local timezone

### Operational Flags

Control command behavior and scope:

- `--verbose int, -v`: Verbose output level (0=silent, 1=info, 2=debug, 3=trace)
- `--all-namespaces, -A`: List resources across all namespaces

### Examples

```bash
# Use a specific kubeconfig file
kubectl mtv --kubeconfig=/path/to/kubeconfig get providers

# Operate in a specific namespace
kubectl mtv -n migration-ns get plans

# Enable debug logging
kubectl mtv -v=2 get inventory vms vsphere-01

# List resources across all namespaces
kubectl mtv get plans --all-namespaces

# Output in JSON format with UTC timestamps
kubectl mtv get plan migration-1 -o json --use-utc
```

## Environment Variables

Configure `kubectl-mtv` behavior using environment variables:

### Core Configuration

- **`MTV_VDDK_INIT_IMAGE`**: Default VDDK initialization image for VMware providers
  ```bash
  export MTV_VDDK_INIT_IMAGE=quay.io/your-registry/vddk:8.0.1
  ```

- **`MTV_INVENTORY_URL`**: Base URL for the inventory service (required for Kubernetes, auto-discovered on OpenShift)
  ```bash
  export MTV_INVENTORY_URL=http://inventory-service-ip:port
  ```

### Kubernetes Configuration

- **`KUBECONFIG`**: Path to kubeconfig file (if not using default location)
  ```bash
  export KUBECONFIG=/path/to/your/kubeconfig
  ```

### Setting Environment Variables Permanently

Add environment variables to your shell profile for persistence:

```bash
# Add to ~/.bashrc, ~/.zshrc, or equivalent
echo 'export MTV_VDDK_INIT_IMAGE=quay.io/your-registry/vddk:8.0.1' >> ~/.bashrc
echo 'export MTV_INVENTORY_URL=http://inventory-service-ip:port' >> ~/.bashrc
source ~/.bashrc
```

## Cluster Requirements and Setup

### Forklift/MTV Installation

`kubectl-mtv` requires either Forklift (upstream) or Migration Toolkit for Virtualization (downstream) to be installed in your cluster.

#### Option A: Install Forklift (Upstream - Any Kubernetes)

```bash
# Install Forklift operator
kubectl apply -f https://github.com/kubev2v/forklift/releases/latest/download/forklift-operator.yaml

# Wait for operator to be ready
kubectl wait --for=condition=Available deployment/forklift-operator \
  -n forklift-operator --timeout=300s

# Create Forklift controller
kubectl apply -f https://github.com/kubev2v/forklift/releases/latest/download/forklift-controller.yaml

# Verify installation
kubectl get pods -n konveyor-forklift
```

#### Option B: Install MTV (Downstream - OpenShift)

For OpenShift environments, install MTV through the Operator Hub:

**Using OpenShift Console:**
1. Navigate to Operators - OperatorHub
2. Search for "Migration Toolkit for Virtualization"
3. Install the operator
4. Create an MTV instance

**Using CLI:**
```bash
# Create subscription for MTV operator
cat <<EOF | oc apply -f -
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

# Verify installation
oc get pods -n openshift-migration
```

### RBAC Permissions

Ensure your user or service account has appropriate permissions to access MTV/Forklift resources.

#### Required Permissions

Create a ClusterRole with necessary permissions:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: mtv-user
rules:
# Forklift/MTV resources
- apiGroups: ["forklift.konveyor.io"]
  resources: ["*"]
  verbs: ["get", "list", "create", "update", "patch", "delete", "watch"]
# Core Kubernetes resources
- apiGroups: [""]
  resources: ["secrets", "configmaps", "namespaces"]
  verbs: ["get", "list", "create", "update", "patch", "delete"]
# For inventory access
- apiGroups: [""]
  resources: ["services"]
  verbs: ["get", "list"]
# For route discovery (OpenShift)
- apiGroups: ["route.openshift.io"]
  resources: ["routes"]
  verbs: ["get", "list"]
```

#### Bind Permissions to User

```bash
# Bind to a user
kubectl create clusterrolebinding mtv-user-binding \
  --clusterrole=mtv-user \
  --user=your-username

# Or bind to a service account
kubectl create clusterrolebinding mtv-serviceaccount-binding \
  --clusterrole=mtv-user \
  --serviceaccount=namespace:serviceaccount-name
```

#### Verify Permissions

```bash
# Check if you can access MTV resources
kubectl auth can-i get plans.forklift.konveyor.io
kubectl auth can-i list providers.forklift.konveyor.io
kubectl auth can-i create mappings.forklift.konveyor.io

# Check specific namespace permissions
kubectl auth can-i create secrets -n migration-namespace
```

### Service Account Setup (Optional)

For automated operations or CI/CD, create a dedicated service account:

```bash
# Create namespace and service account
kubectl create namespace migration-ops
kubectl create serviceaccount mtv-operator -n migration-ops

# Bind the ClusterRole
kubectl create clusterrolebinding mtv-operator-binding \
  --clusterrole=mtv-user \
  --serviceaccount=migration-ops:mtv-operator

# Generate a token (Kubernetes 1.24+)
kubectl create token mtv-operator -n migration-ops --duration=24h

# For long-term tokens, create a secret
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: mtv-operator-token
  namespace: migration-ops
  annotations:
    kubernetes.io/service-account.name: mtv-operator
type: kubernetes.io/service-account-token
EOF

# Retrieve the token
kubectl get secret mtv-operator-token -n migration-ops \
  -o go-template='{% raw %}{{ .data.token | base64decode }}{% endraw %}'
```

## Troubleshooting Installation

### Common Issues and Solutions

#### Issue: Plugin Not Found

**Error**: `plugin "mtv" not found`

**Solutions**:
```bash
# Ensure binary is in PATH
which kubectl-mtv
echo $PATH

# Make binary executable
chmod +x $(which kubectl-mtv)

# Verify kubectl can find plugins
kubectl plugin list
```

#### Issue: Permission Denied

**Error**: `User cannot list resources`

**Solutions**:
```bash
# Check RBAC permissions
kubectl auth can-i get plans.forklift.konveyor.io
kubectl auth can-i list providers.forklift.konveyor.io

# Verify current user
kubectl config current-context
kubectl config view --minify

# Check if Forklift/MTV is installed
kubectl get crd | grep forklift
kubectl get pods -n konveyor-forklift
```

#### Issue: Connection Issues

**Error**: `Unable to connect to cluster`

**Solutions**:
```bash
# Verify cluster connectivity
kubectl cluster-info
kubectl get nodes

# Check kubeconfig
kubectl config current-context
kubectl config get-contexts

# Test with specific kubeconfig
kubectl mtv --kubeconfig=/path/to/config get providers
```

#### Issue: MTV_INVENTORY_URL Not Set (Kubernetes)

**Error**: Commands hang or fail when querying inventory

**Solutions**:
```bash
# Find inventory service
kubectl get service -n konveyor-forklift forklift-inventory

# Set environment variable
export MTV_INVENTORY_URL=http://<service-ip>:<port>

# Or use port-forward for testing
kubectl port-forward -n konveyor-forklift svc/forklift-inventory 8080:8080 &
export MTV_INVENTORY_URL=http://localhost:8080
```

### Debug Mode

Enable verbose output for troubleshooting:

```bash
# Use debug verbosity levels
kubectl mtv -v=1 get providers  # Info level
kubectl mtv -v=2 get providers  # Debug level  
kubectl mtv -v=3 get providers  # Trace level

# Check cluster connectivity with debug
kubectl mtv -v=2 --kubeconfig=/path/to/config get providers
```

### Getting Help

- **Documentation**: Check the [complete documentation](https://github.com/yaacov/kubectl-mtv/tree/main/docs)
- **Examples**: Review [demo examples](https://github.com/yaacov/kubectl-mtv/blob/main/docs/README_demo)
- **Issues**: Open an issue on [GitHub](https://github.com/yaacov/kubectl-mtv/issues)
- **Community**: Join discussions on the Forklift community channels

## Next Steps

After successful installation and verification:

1. **Follow the Quick Start** in [Chapter 3: Quick Start - First Migration Workflow](03-quick-start-first-migration-workflow)
2. **Set up providers** for your source virtualization platforms
3. **Create your first migration plan** using the simplified workflow
4. **Explore advanced features** like VDDK optimization and migration hooks

---

*Previous: [Chapter 1: Overview of kubectl-mtv](01-overview-of-kubectl-mtv)*  
*Next: [Chapter 3: Quick Start - First Migration Workflow](03-quick-start-first-migration-workflow)*
