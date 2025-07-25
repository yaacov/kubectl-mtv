#!/bin/bash
set -euo pipefail


# Load OVA NFS credentials from OVA_URL env variable
OVA_URL_ENV="${OVA_URL:-}"  # e.g. 127.0.0.1:/home/nfs-share
MOUNT_POINT="${OVA_NFS_MOUNT:-/mnt/ova_nfs}"
OVA_SUBDIR="ova_files"
OVA_DOWNLOAD_URL="https://github.com/kubev2v/forkliftci/releases/download/v9.0/vm.ova"
OVA_FILENAME="vm.ova"

if [[ -z "$OVA_URL_ENV" ]]; then
  echo "Error: OVA_URL must be set in the environment."
  exit 1
fi

# Parse OVA_URL into server and path
NFS_SERVER="${OVA_URL_ENV%%:*}"
NFS_PATH="${OVA_URL_ENV#*:}"

if [[ -z "$NFS_SERVER" || -z "$NFS_PATH" ]]; then
  echo "Error: OVA_URL must be in the format <ip>:<path> (e.g. 127.0.0.1:/home/nfs-share)"
  exit 1
fi

# Create mount point if it doesn't exist
mkdir -p "$MOUNT_POINT"

# Mount the NFS directory
sudo mount -t nfs "$NFS_SERVER:$NFS_PATH" "$MOUNT_POINT"
echo "Mounted NFS: $NFS_SERVER:$NFS_PATH at $MOUNT_POINT"

# Create subdirectory for OVA files
mkdir -p "$MOUNT_POINT/$OVA_SUBDIR"

# Download the OVA file into the subdirectory
curl -L "$OVA_DOWNLOAD_URL" -o "$MOUNT_POINT/$OVA_SUBDIR/$OVA_FILENAME"
echo "Downloaded OVA to $MOUNT_POINT/$OVA_SUBDIR/$OVA_FILENAME"

# Unmount the NFS directory
sudo umount "$MOUNT_POINT"
echo "Unmounted $MOUNT_POINT"
