---
layout: page
title: "Chapter 1: Overview of kubectl-mtv"
---

## What is kubectl-mtv?

`kubectl-mtv` is a powerful command-line interface (CLI) plugin for Kubernetes that enables seamless migration of virtual machines from various virtualization platforms to KubeVirt-enabled Kubernetes clusters. MTV stands for "Migration Toolkit for Virtualization" and serves as the primary user interface for the [Forklift](https://github.com/kubev2v/forklift) project.

As a `kubectl` plugin, `kubectl-mtv` integrates directly with your existing Kubernetes workflow, providing a familiar command-line experience for managing complex VM migrations. It acts as a bridge between traditional virtualization infrastructure and modern container orchestration platforms, enabling organizations to modernize their workloads by migrating to Kubernetes-native virtualization.

For complete information about the underlying migration technology, see the [official Forklift documentation](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html).

## Core Functionality and Supported Platforms

### Supported Source Platforms

`kubectl-mtv` supports migration from the following virtualization platforms:

- **VMware vSphere** - Full support for vCenter-managed VMware environments with advanced features like VDDK optimization (see [VMware Prerequisites](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#vmware-prerequisites))
- **Red Hat Virtualization (oVirt/RHV)** - Complete support for oVirt and Red Hat Virtualization platforms (see [oVirt Prerequisites](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#ovirt-prerequisites))
- **Red Hat OpenStack Platform** - Support for OpenStack-based virtualization environments (see [OpenStack Prerequisites](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#openstack-prerequisites))
- **KubeVirt/OpenShift Virtualization** - Support for migrating VMs between KubeVirt clusters or within the same cluster (see [KubeVirt Prerequisites](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#kubevirt-prerequisites))
- **OVA Files** - Direct import and conversion of OVA (Open Virtualization Appliance) files (see [OVA Prerequisites](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#open-virtual-appliance-ova-prerequisites))

### Target Platform

- **KubeVirt on Kubernetes/OpenShift** - All migrations target KubeVirt-enabled Kubernetes clusters, supporting both upstream Kubernetes and Red Hat OpenShift platforms

### Migration Types

The tool supports multiple migration strategies to accommodate different use cases:

- **Cold Migration** - Traditional offline migration where the source VM is powered down during the process (see [Forklift Cold Migration](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#cold-migration))
- **Warm Migration** - Pre-copy migration that minimizes downtime by transferring data while the VM is running, followed by a brief cutover period (see [Forklift Warm Migration](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#warm-migration))
- **Live Migration** - Advanced migration with minimal downtime using KubeVirt's live migration capabilities (available only for KubeVirt/OpenShift Virtualization sources) (see [Forklift Live Migration](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#about-live-migration))
- **Conversion Migration** - Perform only guest conversion and VM creation when storage vendors provide pre-populated PVCs (VMware sources only) (see [Chapter 3.6: Conversion Migration](03.6-conversion-migration))

For detailed information about migration types, see:
- [Chapter 3.5: Migration Types and Strategy Selection](03.5-migration-types-and-strategy-selection) - Cold, warm, and live migration strategies
- [Chapter 3.6: Conversion Migration](03.6-conversion-migration) - External storage vendor integration workflows

## Key Features

### 1. Advanced Query Language Integration

`kubectl-mtv` incorporates the powerful **Tree Search Language (TSL)**, developed by Yaacov Zamir, which enables sophisticated filtering and searching of inventory resources. This is a kubectl-mtv-specific enhancement not available in the base Forklift web interface:

- **SQL-like syntax** for intuitive resource queries
- **Complex filtering** with support for logical operators (AND, OR, NOT)
- **Pattern matching** with LIKE and ILIKE operators
- **Comparison operators** for numeric and date-based filtering
- **Built-in functions** such as LEN, ANY, ALL, and SUM

Example query syntax:
```bash
kubectl mtv get inventory vms vsphere-01 -q "where memoryMB > 4096 and powerState = 'poweredOn'"
```

### 2. Flexible Mapping System

The tool provides multiple approaches to resource mapping:

- **Explicit Mappings** - Pre-defined network and storage mappings for consistent, reusable configurations
- **Inline Mapping Pairs** - Direct specification of source-to-target mappings during plan creation
- **Automatic Default Mappings** - Intelligent mapping based on available target resources
- **Enhanced Storage Options** - Support for volume modes, access modes, and advanced storage array offloading for up to 10x faster migrations (see [Chapter 9.5: Storage Array Offloading](09.5-storage-array-offloading-and-optimization))

### 3. VDDK (Virtual Disk Development Kit) Support

For VMware environments, `kubectl-mtv` provides optimized disk transfer capabilities:

- **Custom VDDK Image Creation** - Build optimized disk transfer containers
- **Performance Optimization** - Significantly faster disk transfers compared to standard methods
- **Advanced Configuration** - Support for buffer sizing and AIO optimization
- **Seamless Integration** - Automatic VDDK usage when configured

### 4. Real-time Monitoring and Management

Comprehensive lifecycle management for migrations:

- **Live Progress Tracking** - Real-time monitoring with `--watch` flag
- **Detailed Status Reporting** - Complete visibility into migration phases and steps
- **Migration Control** - Start, pause, resume, cancel, and cutover operations
- **Archive Management** - Archive completed migrations for historical tracking

### 5. Kubernetes-native Resource Management (KARL)

Integration with the **Kubernetes Affinity Rule Language (KARL)** for advanced pod placement. This is a kubectl-mtv-specific enhancement that simplifies complex Kubernetes affinity expressions:

- **Declarative Affinity Rules** - Express complex placement requirements using natural language syntax
- **Topology-aware Placement** - Support for node, zone, region, and rack-level affinity
- **Multiple Rule Types** - REQUIRE (hard), PREFER (soft), AVOID (hard anti-affinity), REPEL (soft anti-affinity)
- **Target VM and Convertor Pod Optimization** - Separate affinity rules for runtime VMs and migration workloads

Example KARL syntax:
```bash
--target-affinity "REQUIRE pods with app=database on node"
--convertor-affinity "PREFER pods with storage=high-performance on zone"
```

### 6. Migration Hooks and Automation

Support for custom automation through migration hooks:

- **Pre-migration Hooks** - Execute custom logic before migration begins
- **Post-migration Hooks** - Perform cleanup or configuration after migration completes
- **Ansible Playbook Integration** - Run Ansible playbooks as part of the migration process
- **Shell Script Support** - Execute custom shell scripts with access to migration context
- **Context Access** - Hooks receive detailed migration context including plan and workload information

### 7. Model Context Protocol (MCP) Integration

Advanced AI assistant integration for enhanced user experience. This is a kubectl-mtv-specific feature not available in standard Forklift:

- **AI Assistant Access** - Provide Claude, Cursor IDE, and other MCP-compatible tools access to migration resources
- **Multiple Server Modes** - Support for both stdio and SSE (Server-Sent Events) communication
- **Natural Language Queries** - Enable AI assistants to interpret and execute migration commands
- **Intelligent Assistance** - AI-powered recommendations and troubleshooting support

## Relationship with Forklift/Migration Toolkit for Virtualization (MTV)

`kubectl-mtv` serves as the primary command-line interface for the Forklift project, which is Red Hat's upstream migration toolkit. The relationship between these components is as follows:

### Forklift Controller
- **Backend Engine** - Handles the actual migration orchestration and execution (see [Forklift Architecture](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#architecture))
- **Custom Resource Management** - Manages Kubernetes custom resources for providers, plans, and mappings (see [Forklift Custom Resources](https://kubev2v.github.io/forklift-documentation/documentation/doc-Migration_Toolkit_for_Virtualization/master/index.html#forklift-custom-resources))
- **Integration Layer** - Interfaces with source virtualization platforms and target KubeVirt infrastructure

### kubectl-mtv CLI
- **User Interface** - Provides intuitive command-line access to Forklift functionality
- **Resource Management** - Simplifies creation, modification, and monitoring of migration resources
- **Workflow Orchestration** - Streamlines complex migration workflows into simple command sequences
- **Advanced Features** - Adds sophisticated query capabilities (TSL), affinity management (KARL), hooks, and AI integration

### Migration Toolkit for Virtualization (MTV)
- **Product Integration** - MTV is Red Hat's supported product offering that includes Forklift
- **Enterprise Features** - Additional enterprise-grade features, support, and integration
- **OpenShift Integration** - Deep integration with Red Hat OpenShift Container Platform

### Architecture Flow

1. **Resource Definition** - Users define providers, mappings, and plans using `kubectl-mtv`
2. **API Translation** - Commands are translated to Kubernetes API calls creating Forklift custom resources
3. **Controller Processing** - Forklift controllers process these resources and orchestrate migrations
4. **Status Reporting** - Migration status flows back through the API to `kubectl-mtv` for user visibility

This architecture ensures that `kubectl-mtv` provides a powerful, user-friendly interface while leveraging the robust, battle-tested migration engine provided by Forklift. The separation of concerns allows for specialized optimization in both the user experience and the migration execution layers.

## Getting Started

With this foundational understanding of `kubectl-mtv`, you're ready to proceed to the next chapter covering installation and prerequisites. The tool's comprehensive feature set enables everything from simple single-VM migrations to complex enterprise-scale migration projects with hundreds of virtual machines.

The following chapters will guide you through:
- Detailed installation procedures
- Provider setup and configuration
- Advanced querying and filtering techniques
- Mapping creation and management
- Migration plan development and execution
- Optimization strategies for performance and placement
- Troubleshooting and best practices

---

*Next: [Chapter 2: Installation and Prerequisites](02-installation-and-prerequisites)*
