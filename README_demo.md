# kubectl-mtv openshift demo

This document provides step-by-step demo of the `kubectl-mtv` CLI tool for OpenShift Virtualization.

## Prerequisites

- OpenShift Container Platform 4.7+
- Migration Toolkit for Virtualization (MTV) Operator installed
- Access to virtualization source platform (e.g., VMware vSphere)
- `oc` and `kubectl-mtv` CLI tools installed

## Step-by-Step Migration Process

### 1. Project Setup

Create a new OpenShift project for the migration demo:

```bash
oc new-project demo
```

### 2. Provider Registration

List existing providers:

```bash
oc mtv provider list
```

Register OpenShift as the target provider:

```bash
oc mtv provider create host --type openshift
```

Verify that the VDDK initialization image environment variable is set:

```bash
# For example, use quay.io/private/vddk:8.0.1 as the default VDDK init image:
echo $MTV_VDDK_INIT_IMAGE
```

Register VMware vSphere as the source provider:

```bash
# For example, use the default VDDK image, and cont verify the tls connection.
oc mtv provider create vmware --type vsphere \
  -U https://your.vsphere.server.com/sdk \
  -u your_vsphere_username \
  -p $YOUR_PASSWORD \
  --provider-insecure-skip-tls
```

Re fetch existing providers:

```bash
oc mtv provider list
```

### 3. Fetch VM Inventory

Retrieve the VM inventory from the VMware provider:

```bash
# For example, select VMs that have a name matching RegExp rule and have more then one disk:
oc mtv inventory vms vmware -q "where name ~= 'your_vm_name' and len disks > 1"
```

### 4. Create Migration Plan

Create a migration plan for the selected VM:

```bash
oc mtv plan create demo -S vmware --vms your_selected_vms
```

For more advanced options, you can use flags like:

```bash
# Create a plan with PVC naming template
oc mtv plan create demo-advanced -S vmware --vms your_selected_vms \
  --pvc-name-template "{{.VmName}}-disk-{{.DiskIndex}}" \
  --pvc-name-template-use-generate-name=false

# Create a warm migration plan with automatic cleanup
oc mtv plan create demo-warm -S vmware --vms your_selected_vms \
  --warm --delete-guest-conversion-pod
```

Optional step to edit network or storage mappings:

```bash
oc edit plan <plan name>
oc edit networkmap <networkmap-name>
oc edit storagemap <storagemap-name>
```

### 5. Execute Migration Plan

Review and initiate the migration:

```bash
oc mtv plan describe demo
oc mtv plan start demo
```

### 6. Monitoring Migration

Monitor migration progress and status:

```bash
oc mtv plan describe demo -w
oc mtv plan vms demo -w
oc mtv plan vm demo --vm your_selected_vm -w
```

Monitor logs, pods, and persistent volume claims:

```bash
# The '-w' flag is optional flag that keep the command running, waiting for updates.
oc get pvc -l vmID=<vm-id> -w
oc get pod -l vmID=<vm-id> -w
# The '-f' flag is optional flag that keep the command running, waiting for updates.
oc logs -l vmID=<vm-id> -f
```

## Resources

- [forklift GitHub Repository](https://github.com/kubev2v/forklift)
- [kubectl-mtv GitHub Repository](https://github.com/yaacov/kubectl-mtv)
