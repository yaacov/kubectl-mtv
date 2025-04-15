# VDDK Image Creation and Usage

It is strongly recommended that Migration Toolkit for Virtualization (MTV) should be used with the VMware Virtual Disk Development Kit (VDDK) SDK when transferring virtual disks from VMware vSphere.

> **Note**  
> Creating a VDDK image, although optional, is highly recommended. Using MTV without VDDK is not recommended and could result in significantly lower migration speeds.

To make use of this feature, you download the VMware Virtual Disk Development Kit (VDDK), build a VDDK image, and push the VDDK image to your image registry.

The VDDK package contains symbolic links, therefore, the procedure of creating a VDDK image must be performed on a file system that preserves symbolic links (symlinks).

> **Note**  
> Storing the VDDK image in a public registry might violate the VMware license terms.

## Prerequisites

- Red Hat OpenShift image registry.
- `podman` installed.
- You are working on a file system that preserves symbolic links (symlinks).
- If you are using an external registry, OpenShift Virtualization must be able to access it.

## Steps to Create and Use a VDDK Image

### 1. Download the VDDK SDK

Download the [VMware Virtual Disk Development Kit (VDDK) tar.gz](https://developer.vmware.com/web/sdk/8.0/vddk) file from VMware.

### 2. Build the VDDK Image

Use the `kubectl mtv vddk create` command to build and optionally push the image:

```bash
# For example:
kubectl mtv vddk create --tar ~/vmware-vix-disklib-distrib-8-0-1.tar.gz --tag quay.io/example/vddk:8
```

- `--tar`: Path to the VDDK tar.gz file (e.g., `~/vmware-vix-disklib-distrib-8-0-1.tar.gz`)
- `--tag`: Container image tag to use (e.g., `quay.io/example/vddk:8`)
- `--push`: (Optional) Push the image to the registry after building

Example with push:

```bash
kubectl mtv vddk create --tar ~/vmware-vix-disklib-distrib-8-0-1.tar.gz --tag quay.io/example/vddk:8 --push
```

### 3. Set the MTV_VDDK_INIT_IMAGE Environment Variable

Set the `MTV_VDDK_INIT_IMAGE` environment variable to the image you built and pushed:

```bash
export MTV_VDDK_INIT_IMAGE=quay.io/yourorg/vddk:8
```

You can add this line to your shell profile (e.g., `.bashrc` or `.zshrc`) for persistence.

### 4. Use the VDDK Image in MTV

When creating a vSphere provider, MTV will use the image specified in `MTV_VDDK_INIT_IMAGE` as the default VDDK init image.

If you want to override it per provider, use the `--vddk-init-image` flag when creating the provider.
