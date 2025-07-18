# kubectl-mtv Inventory Tutorial

This document provides a guide to using the `kubectl mtv get inventory` commands for exploring provider inventories.

## Overview

The `kubectl mtv get inventory` command allows you to query and explore resources from virtualization providers, including:

- Virtual Machines (VMs)
- Networks
- Storage
- Hosts (vSphere providers)
- Namespaces (Kubernetes/Openshift providers)

## General Syntax

All inventory commands follow this general pattern:

```bash
kubectl mtv get inventory <resource-type> <provider-name> [flags]
```

### Common Flags

- `-i, --inventory-url`: Base URL for the inventory service (defaults to MTV_INVENTORY_URL environment variable)
- `-o, --output`: Output format (table, json, planvms for VMs)
- `-q, --query`: Query string for filtering and sorting results
- `-e, --extended`: Show extended information in table output

## Query Language

The query language supports a SQL-like syntax with the following clauses:

- `SELECT field1, field2 AS alias, field3`: Select specific fields with optional aliases
- `WHERE condition`: Filter using tree-search-language conditions
- `ORDER BY field1 [ASC|DESC], field2`: Sort results on multiple fields
- `LIMIT n`: Limit number of results

For a more complete reference of the query language syntax and capabilities, see the [Query Language Reference](./README_quary_language.md) document.

## Examples

### Listing VMs

#### Basic VM listing

```bash
kubectl mtv get inventory vms vsphere-01
```

This displays a table with VM names, power state, CPU count, memory, and other basic information.

#### Filtering VMs by name pattern

```bash
kubectl mtv get inventory vms vsphere-01 -q "WHERE name LIKE 'web-%'"
```

Lists only VMs with names starting with "web-".

#### Sorting VMs by memory

```bash
kubectl mtv get inventory vms vsphere-01 -q "ORDER BY memoryMB DESC LIMIT 5"
```

Lists the top 5 VMs with the most memory allocation.

#### Complex queries

```bash
kubectl mtv get inventory vms vsphere-01 -q "WHERE powerState = 'poweredOn' AND memoryMB > 4096 ORDER BY cpuCount DESC"
```

Lists powered-on VMs with more than 4GB memory, sorted by CPU count in descending order.

#### JSON output for scripting

```bash
kubectl mtv get inventory vms vsphere-01 -q "LIMIT 1" -o json
```

Returns VM details in JSON format:

```json
[
  {
    "balloonedMemory": 0,
    "changeTrackingEnabled": false,
    "concerns": [
      {
        "assessment": "Changed Block Tracking (CBT) has not been enabled on this VM. This feature is a prerequisite for VM warm migration.",
        "category": "Warning",
        "label": "Changed Block Tracking (CBT) not enabled"
      },
      {
        "assessment": "Distributed resource scheduling is not currently supported by Forklift Virtualization. The VM can be migrated but it will not have this feature in the target environment.",
        "category": "Information",
        "label": "VM running in a DRS-enabled cluster"
      }
    ],
    "concernsHuman": "0/1/1",
    "connectionState": "connected",
    "coresPerSocket": 1,
    "cpuAffinity": [],
    "cpuCount": 1,
    "cpuHotAddEnabled": false,
    "cpuHotRemoveEnabled": false,
    "criticalConcerns": 0,
    "devices": [
      {
        "kind": "VirtualE1000"
      }
    ],
    "diskCapacity": "0.0 GB",
    "diskEnableUuid": false,
    "disks": [
      {
        "bus": "scsi",
        "capacity": 52428800,
        "datastore": {
          "id": "datastore-13341",
          "kind": "Datastore"
        },
        "file": "[example-iscsi-ds1] example/example.vmdk",
        "key": 2000,
        "mode": "persistent",
        "rdm": false,
        "serial": "6000C296-9fdf-7f71-8bce-3e9ad968ffca",
        "shared": false
      }
    ],
    "faultToleranceEnabled": false,
    "firmware": "bios",
    "guestId": "rhel8_64Guest",
    "guestIpStacks": [
      {
        "device": "0",
        "dns": [
          "10.47.242.10",
          "10.38.5.26"
        ],
        "gateway": "10.46.29.254"
      }
    ],
    "guestName": "Red Hat Enterprise Linux 9 (64-bit)",
    "guestNetworks": [
      {
        "device": "0",
        "dns": null,
        "ip": "10.46.29.240",
        "mac": "00:50:56:8b:ca:5c",
        "origin": "",
        "prefix": 25
      },
      {
        "device": "0",
        "dns": null,
        "ip": "fe80::2d18:5143:21ea:dc9c",
        "mac": "00:50:56:8b:ca:5c",
        "origin": "",
        "prefix": 64
      }
    ],
    "host": "host-40",
    "id": "vm-84314",
    "infoConcerns": 1,
    "ipAddress": "10.46.29.240",
    "isTemplate": false,
    "memoryGB": "4.0 GB",
    "memoryHotAddEnabled": false,
    "memoryMB": 4096,
    "name": "example",
    "networks": [
      {
        "id": "dvportgroup-17756",
        "kind": "Network"
      }
    ],
    "nics": [
      {
        "mac": "00:50:56:8b:ca:5c",
        "network": {
          "id": "dvportgroup-17756",
          "kind": "Network"
        },
        "order": 0
      }
    ],
    "numaNodeAffinity": [],
    "parent": {
      "id": "group-v4",
      "kind": "Folder"
    },
    "path": "/example-Datacenter/vm/example",
    "policyVersion": 5,
    "powerState": "poweredOn",
    "powerStateHuman": "On",
    "provider": "vmware",
    "revision": 1,
    "revisionValidated": 1,
    "secureBoot": false,
    "selfLink": "providers/vsphere/802925c1-7024-46c7-90ef-c5ffc8c4e7de/vms/vm-84314",
    "snapshot": {
      "id": "",
      "kind": ""
    },
    "storageUsed": 4431800192,
    "storageUsedGB": "4.1 GB",
    "tpmEnabled": false,
    "uuid": "420b687e-0e72-da26-a451-5dda90beed63",
    "warningConcerns": 1
  }
]
```

#### Creating migration plans from VM inventory

```bash
kubectl mtv get inventory vms vsphere-01 -q "WHERE powerState = 'poweredOn'" -o planvms > vms-to-migrate.yaml
kubectl mtv create plan my-migration-plan --source vsphere-01 --target kubernetes-target --vms @vms-to-migrate.yaml
```

This creates a list of powered-on VMs in a format suitable for migration planning, then creates a plan using that list.

See [Editing the VMs List for Migration Plans (planvms)](./README_planvms.md) for details on the planvms file format and how to edit it before migration.

#### Selecting specific fields

```bash
kubectl mtv get inventory vms vsphere-01 -q "SELECT name, powerStateHuman AS state, memoryGB, cpuCount, ipAddress WHERE memoryMB > 2048"
```

Shows only selected fields for VMs with more than 2GB of memory.

### Listing Networks

#### Basic network listing

```bash
kubectl mtv get inventory networks vsphere-01
```

#### Filtering networks

```bash
kubectl mtv get inventory networks vsphere-01 -q "WHERE name LIKE '%prod%'"
```

Lists only networks with "prod" in their name.

#### Network JSON output example

```bash
kubectl mtv get inventory networks vsphere-01 -q "LIMIT 1" -o json
```

Returns network details in JSON format:

```json
[
  {
    "host": [
      {
        "Host": {
          "id": "host-36",
          "kind": "Host"
        },
        "PNIC": [
          "vmnic0",
          "vmnic1"
        ]
      }
    ],
    "hostCount": 7,
    "id": "dvs-17579",
    "name": "vDSwitch0",
    "parent": {
      "id": "group-n7",
      "kind": "Folder"
    },
    "path": "/example-Datacenter/network/vDSwitch0",
    "provider": "vsphere-01",
    "revision": 1,
    "selfLink": "providers/vsphere/4462af50-0c57-4dcd-a6f0-16c54ca97eb6/networks/dvs-17579",
    "variant": "DvSwitch",
    "vlanId": ""
  }
]
```

#### Advanced network queries

```bash
kubectl mtv get inventory networks vsphere-01 -q "WHERE variant = 'DvSwitch' AND hostCount > 5"
```

Lists distributed virtual switches connected to more than 5 hosts.

### Listing Storage

#### Basic storage listing

```bash
kubectl mtv get inventory storage vsphere-01
```

#### Finding available storage

```bash
kubectl mtv get inventory storage vsphere-01 -q "WHERE free > 500Gi ORDER BY free DESC"
```

Lists storage with more than 500GB free space, sorted by available space.

#### Storage JSON output example

```bash
kubectl mtv get inventory storage vsphere-01 -q "LIMIT 1" -o json
```

Returns storage details in JSON format:

```json
[
  {
    "capacity": 822486237184,
    "capacityHuman": "766.0 GB",
    "free": 348943024128,
    "freeHuman": "325.0 GB",
    "id": "datastore-45",
    "maintenance": "normal",
    "name": "exampleesxi02_ds1",
    "parent": {
      "id": "group-s6",
      "kind": "Folder"
    },
    "path": "/example-Datacenter/datastore/exampleesxi02_ds1",
    "provider": "vsphere-01",
    "revision": 1,
    "selfLink": "providers/vsphere/4462af50-0c57-4dcd-a6f0-16c54ca97eb6/datastores/datastore-45",
    "type": "VMFS"
  }
]
```

#### Storage maintenance status check

```bash
kubectl mtv get inventory storage vsphere-01 -q "WHERE maintenance != 'normal'"
```

Lists storage datastores that are in maintenance mode.

#### Storage utilization analysis

```bash
kubectl mtv get inventory storage vsphere-01 -q "SELECT name, capacityHuman, freeHuman, type WHERE (capacity - free) / capacity > 0.7"
```

Lists datastores that are over 70% utilized.

### Listing Hosts

#### Basic host listing

```bash
kubectl mtv get inventory hosts vsphere-01
```

#### Filtering hosts

```bash
kubectl mtv get inventory hosts vsphere-01 -q "WHERE inMaintenance = false"
```

#### Host JSON output example

```bash
kubectl mtv get inventory hosts vsphere-01 -q "LIMIT 1" -o json
```

Returns host details in JSON format:

```json
[
  {
    "cluster": "domain-c34",
    "cpuCores": 32,
    "cpuSockets": 2,
    "datastores": [
      {
        "id": "datastore-3036",
        "kind": "Datastore"
      }
    ],
    "id": "host-3152",
    "inMaintenance": false,
    "managementServerIp": "10.46.29.211",
    "name": "10.46.29.133",
    "networkAdapters": [
      {
        "ipAddress": "10.46.52.4",
        "linkSpeed": 10000,
        "mtu": 9000,
        "name": "vDSwitch0",
        "subnetMask": "255.255.255.128"
      }
    ],
    "networking": {
      "pNICs": [
        {
          "key": "key-vim.host.PhysicalNic-vmnic0",
          "linkSpeed": 10000
        }
      ],
      "portGroups": [],
      "switches": [],
      "vNICs": []
    },
    "networks": [
      {
        "id": "dvportgroup-17763",
        "kind": "Network"
      }
    ],
    "parent": {
      "id": "domain-c34",
      "kind": "Cluster"
    },
    "path": "/example-Datacenter/host/example-Cluster/10.46.29.133",
    "productName": "VMware ESXi",
    "productVersion": "8.0.3",
    "provider": "vsphere-01",
    "status": "green",
    "thumbprint": "D7:CC:22:02:41:8F:BF:7D:6D:95:A8:01:E1:43:F8:E6:90:E6:DF:75",
    "timezone": "UTC",
    "vms": null
  }
]
```

#### Advanced host queries

```bash
kubectl mtv get inventory hosts vsphere-01 -q "WHERE productVersion LIKE '8.%' AND cpuCores > 16"
```

Lists ESXi 8.x hosts with more than 16 CPU cores.

```bash
kubectl mtv get inventory hosts vsphere-01 -q "SELECT name, productVersion, cpuCores, cpuSockets, status WHERE inMaintenance = false"
```

Lists active hosts not in maintenance mode with selected fields.

```bash
kubectl mtv get inventory hosts vsphere-01 -q "WHERE status != 'green'"
```

Lists hosts that have warning or error status.

### Listing Namespaces

```bash
kubectl mtv get inventory namespaces kubernetes-target
```

## Common Query Examples

### Finding VMs with specific guest OS

```bash
kubectl mtv get inventory vms vsphere-01 -q "WHERE guestName LIKE '%Linux%'"
```

### Finding VMs with migration concerns

```bash
kubectl mtv get inventory vms vsphere-01 -q "WHERE warningConcerns > 0 OR criticalConcerns > 0"
```

### Finding VMs with specific network connections

```bash
kubectl mtv inventory vms vsphere-01 -q "WHERE ipAddress LIKE '10.10.%'"
```

### Finding VMs with shared disks

```bash
kubectl mtv inventory vms vsphere-01 -q "SELECT name, disks[*].shared WHERE any (disks[*].shared = true)"
```

## Tips for Effective Inventory Queries

1. **Start Simple**: Begin with basic queries and refine as needed
2. **Use JSON for Exploration**: Use `-o json` to see all available fields for querying
3. **Limit Results**: Use `LIMIT` to avoid overwhelming output when exploring large inventories
4. **Export for Migration**: Use `-o planvms` to generate VM lists ready for migration plans

## Common Fields for Querying

### VM Fields

- `name`: VM name
- `powerState`: Current power status (poweredOn, poweredOff)
- `cpuCount`: Number of vCPUs
- `memoryMB`: Memory in MB
- `memoryGB`: Memory in GB (formatted string)
- `ipAddress`: Primary IP address
- `guestName`: Guest OS name
- `warningConcerns`: Count of migration warnings
- `criticalConcerns`: Count of critical migration concerns

### Network Fields

- `name`: Network name
- `id`: Network ID
- `variant`: Network type (DvSwitch, etc.)
- `vlanId`: VLAN identifier
- `hostCount`: Number of connected hosts
- `path`: Network path in the inventory

### Storage Fields

- `name`: Storage name
- `type`: Storage type (VMFS, NFS, etc.)
- `capacity`: Total capacity in bytes
- `capacityHuman`: Formatted total capacity
- `free`: Available space in bytes
- `freeHuman`: Formatted available space
- `maintenance`: Maintenance status
- `path`: Storage path in the inventory
