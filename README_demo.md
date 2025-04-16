# kubectl-mtv kubernetes demo

This document provides a step-by-step demo of the `kubectl-mtv` CLI tool for Kubernetes Forklift (upstream of Migration Toolkit for Virtualization).

## Prerequisites

- Kubernetes cluster 1.23+
- Forklift/MTV Operator installed
- Access to virtualization source platform (e.g., VMware vSphere)
- `kubectl` and `kubectl-mtv` CLI tools installed (`oc` command will also work)

## Step-by-Step Migration Process

### 1. Project Setup

Create a new Kubernetes namespace for the migration demo:

```bash
kubectl create namespace demo
```

### 2. Provider Registration

List existing providers:

```bash
kubectl mtv provider list
```

Register Kubernetes as the target provider:

```bash
kubectl mtv provider create host --type kubernetes
```

Verify that the VDDK initialization image environment variable is set:

```bash
# For example, use quay.io/private/vddk:8.0.1 as the default VDDK init image:
#    export MTV_VDDK_INIT_IMAGE=quay.io/private/vddk:8.0.1
echo $MTV_VDDK_INIT_IMAGE
```

See [VDDK Image Creation and Usage](./README_vddk.md) for instructions on building and configuring a VDDK image.

Register VMware vSphere as the source provider:

```bash
# For example, use the default VDDK image, and skip TLS verification.
kubectl mtv provider create vmware --type vsphere \
  -U https://your.vsphere.server.com/sdk \
  -u your_vsphere_username \
  -p $YOUR_PASSWORD \
  --provider-insecure-skip-tls
```

Re-fetch existing providers:

```bash
kubectl mtv provider list
```

### 3. Fetch VM Inventory

Retrieve the VM inventory from the VMware provider:

```bash
# For example, select VMs that have a name matching RegExp rule and have more than one disk:
kubectl mtv inventory vms vmware -q "where name ~= 'your_vm_name' and len disks > 1"
```

### 4. Create Migration Plan

Create a migration plan for the selected VM:

```bash
kubectl mtv plan create demo -S vmware --vms your_selected_vms
```

For more advanced options, you can use flags like:

```bash
# Create a plan with PVC naming template
kubectl mtv plan create demo-advanced -S vmware --vms your_selected_vms \
  --pvc-name-template "{{.VmName}}-disk-{{.DiskIndex}}" \
  --pvc-name-template-use-generate-name=false

# Create a warm migration plan with automatic cleanup
kubectl mtv plan create demo-warm -S vmware --vms your_selected_vms \
  --warm --delete-guest-conversion-pod
```

Optional step to edit network or storage mappings:

```bash
kubectl edit plan <plan name>
kubectl edit networkmap <networkmap-name>
kubectl edit storagemap <storagemap-name>
```

### 5. Execute Migration Plan

Review and initiate the migration:

```bash
kubectl mtv plan describe demo
kubectl mtv plan start demo
```

### 6. Monitoring Migration

Monitor migration progress and status:

```bash
kubectl mtv plan describe demo -w
kubectl mtv plan vms demo -w
kubectl mtv plan vm demo --vm your_selected_vm -w
```

Monitor logs, pods, and persistent volume claims:

```bash
# The '-w' flag is optional and keeps the command running, waiting for updates.
kubectl get pvc -l vmID=<vm-id> -w
kubectl get pod -l vmID=<vm-id> -w
# The '-f' flag is optional and keeps the command running, waiting for updates.
kubectl logs -l vmID=<vm-id> -f
```

## Resources

- [Forklift GitHub Repository](https://github.com/kubev2v/forklift)
- [kubectl-mtv GitHub Repository](https://github.com/yaacov/kubectl-mtv)
