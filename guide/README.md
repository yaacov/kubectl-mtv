---
layout: home
title: kubectl-mtv Technical Guide
permalink: /
---

# kubectl-mtv Technical Guide

Welcome to the technical guide for `kubectl-mtv` - the powerful command-line interface that transforms how you migrate virtual machines to Kubernetes.

`kubectl-mtv` is a sophisticated CLI plugin that enables seamless migration of virtual machines from traditional virtualization platforms (VMware vSphere, Red Hat Virtualization, OpenStack, and more) directly to KubeVirt-enabled Kubernetes clusters. Built on top of the proven Forklift migration engine, it provides advanced features like SQL-like query capabilities, intelligent resource mapping, and AI-powered assistance to make complex migrations simple and reliable.

Whether you're migrating a single development VM or orchestrating enterprise-scale datacenter transformations, this guide provides everything you need to master kubectl-mtv.

---

## Table of Contents

### I. Introduction and Fundamentals

1.  **[Overview of kubectl-mtv](01-overview-of-kubectl-mtv.md)**
    *   What is `kubectl-mtv`? (A `kubectl` plugin for migrating virtualization workloads to KubeVirt using Forklift).
    *   Core Functionality and Supported Platforms (vSphere, oVirt, OpenStack, OVA).
    *   Key Features (Advanced Queries, Flexible Mapping, VDDK Support, Real-time Monitoring).
    *   Relationship with Forklift/Migration Toolkit for Virtualization (MTV).

2.  **[Installation and Prerequisites](02-installation-and-prerequisites.md)**
    *   **Prerequisites** (Kubernetes Cluster 1.23+, Forklift/MTV installed, `kubectl`, appropriate RBAC permissions).
    *   Installation Methods (Step-by-Step How-To)
        *   Method 1: **Krew Plugin Manager (Recommended)**.
        *   Method 2: Downloading Release Binaries.
        *   Method 3: Building from Source (Requires Go 1.24+, `git`, `make`).
    *   Verification and Configuration.
    *   Global Flags Reference
        *   Setting `kubeconfig` and context.
        *   Using `--namespace` (`-n`) and `--output` (`-o`).
        *   Timezone management (`--use-utc`).

3.  **[Quick Start: First Migration Workflow](03-quick-start-first-migration-workflow.md)**
    *   Step 1: Project Setup (Creating a namespace).
    *   Step 2: Registering Providers (Source and Target).
    *   Step 3: Creating Mappings (Optional step).
    *   Step 4: Creating the Migration Plan.
    *   Step 5: Executing and Monitoring the Migration (`start plan`, `get plan --watch`).

4.  **[Migration Types and Strategy Selection](03.5-migration-types-and-strategy-selection.md)**
    *   Cold Migration (Complete offline process, highest reliability).
    *   Warm Migration (Two-stage precopy/cutover, minimal downtime).
    *   Live Migration (Near-zero downtime, KubeVirt sources only).
    *   Migration Selection Decision Framework.
    *   Performance Considerations and Official Testing Data.

5.  **[Conversion Migration](03.6-conversion-migration.md)**
    *   Overview and Architecture (External storage vendor integration).
    *   Platform Requirements and Limitations (VMware vSphere only).
    *   Prerequisites and PVC Metadata Requirements.
    *   Migration Plan Configuration and Workflow.
    *   Use Cases and Enterprise Storage Integration.
    *   Best Practices and Troubleshooting.

### II. Provider, Host, and VDDK Management

6.  **[Provider Management](04-provider-management.md)**
    *   Listing, Describing, and Deleting Providers.
    *   **How-To: Creating Providers** (Syntax and Required Flags).
        *   vSphere Provider (URL, Credentials, VDDK Image).
        *   oVirt/RHV Provider.
        *   OpenStack Provider.
        *   OpenShift/Kubernetes (Target) Provider.
    *   **How-To: Patching Providers** (Updating settings securely).
        *   Updating URL, Credentials, and CA Certificates.
        *   Understanding Secret Ownership and Protection (Owned vs. Shared Secrets).

7.  **[Migration Host Management (vSphere Specific)](05-migration-host-management.md)**
    *   Overview and Purpose of Migration Hosts (Direct ESXi access, optimization).
    *   **How-To: Creating Hosts** (`kubectl mtv create host`).
        *   IP Address Resolution (Direct IP vs. Network Adapter Lookup).
        *   Authentication Options (Provider Secret, Existing Secret, New Credentials).
    *   Listing, Describing, and Deleting Hosts.
    *   Best Practices for Host Creation.

8.  **[VDDK Image Creation and Configuration](06-vddk-image-creation-and-configuration.md)**
    *   Why VDDK is recommended for VMware disk transfers.
    *   Prerequisites for Building the Image.
    *   **How-To: Building the VDDK Image** (`kubectl mtv create vddk-image`).
    *   Setting the **`MTV_VDDK_INIT_IMAGE`** Environment Variable.
    *   Using the VDDK Image in Provider Creation.

### III. Inventory and Advanced Query Language

9.  **[Inventory Management](07-inventory-management.md)**
    *   Overview of Resources Available for Querying (VMs, Networks, Storage, Hosts, Providers).
    *   General Syntax: `kubectl mtv get inventory <resource> <provider>`.
    *   Common Inventory Examples (Listing VMs, Networks, Storage).
    *   Output Formats (Table, JSON, YAML).
    *   **How-To: Exporting VMs for Migration Planning** (`-o planvms`).

10. **[Query Language Reference and Advanced Filtering](08-query-language-reference-and-advanced-filtering.md)**
    *   Query Structure (SELECT, WHERE, ORDER BY, LIMIT clauses).
    *   **Detailed Syntax and Features**
        *   WHERE Clause (Tree Search Language - TSL).
        *   Operators (LIKE, ILIKE, Comparison, Logical: AND/OR/NOT).
        *   Functions (LEN, ANY, ALL, SUM).
    *   **Advanced Query Examples**.
        *   Filtering by Power State, Memory, and Name Patterns.
        *   Sorting Results (ASC/DESC).
        *   Finding VMs with specific migration concerns (`criticalConcerns`).
        *   Querying Provider Status and Resource Counts.
    *   Tips for Effective Inventory Queries.

### IV. Mapping and Plan Configuration

11. **[Mapping Management](09-mapping-management.md)**
    *   Overview (Defining source-to-target resource relationships).
    *   Listing, Viewing, and Deleting Mappings.
    *   **How-To: Creating Mappings** (`kubectl mtv create mapping`).
        *   **Network Mapping Pairs Format** (Multus, Pod, Ignored networks).
        *   **Storage Mapping Pairs Format** (StorageClass reference).
        *   **Enhanced Storage Options** (volumeMode, accessMode, offloadPlugin, offloadVendor).
    *   **How-To: Patching Mappings** (Adding, Updating, and Removing pairs).

12. **[Storage Array Offloading and Optimization](09.5-storage-array-offloading-and-optimization.md)**
    *   Overview and Benefits (10x faster migrations, reduced network overhead).
    *   **Supported Storage Vendors** (IBM FlashSystem, NetApp ONTAP, Pure Storage, Dell PowerMax, HPE Primera).
    *   **How-To: Configuration and Setup** (vSphere XCopy integration, credential management).
    *   **Vendor-Specific Configurations** (FlashSystem, ONTAP, Pure Storage, PowerMax optimizations).
    *   **Performance Tuning and Monitoring** (Best practices, troubleshooting, metrics analysis).

13. **[Migration Plan Creation](10-migration-plan-creation.md)**
    *   **VM Selection Methods**.
        *   Method 1: Comma-separated List of VM Names.
        *   Method 2: File Reference (`--vms @file.yaml`).
        *   Method 3: **Query String Selection** (`--vms "where ..."`).
    *   **Mapping Configuration Options in Plan Creation**.
        *   Using Existing Mappings.
        *   Using Inline Mapping Pairs.
        *   Using Default Mappings (Simplest approach).
    *   Key Plan Configuration Flags.
        *   Migration Types (`cold`, `warm`, `live`, `conversion`).
        *   Target Namespace and Transfer Network.
        *   Naming Templates (`--pvc-name-template`, `--volume-name-template`).

14. **[Customizing Individual VMs (The PlanVMS Format)](11-customizing-individual-vms-planvms-format.md)**
    *   Detailed VM List Format.
    *   Editable Fields for Customization (targetName, rootDisk, instanceType, templates, LUKS secrets).
    *   Go Template Variables Reference (PVC, Volume, Network templates).
    *   **How-To: Editing the List** (Customizing target names and attaching hooks).

### V. Advanced Migration Customization and Optimization

15. **[Target VM Placement (Operational Lifetime)](12-target-vm-placement.md)**
    *   Distinction: Target VM Configuration vs. Migration Process Optimization.
    *   Flags: `--target-labels`, `--target-node-selector`, `--target-power-state`.
    *   **Target Affinity with KARL Syntax**.
        *   KARL Rule Types (REQUIRE, PREFER, AVOID, REPEL).
        *   Topology Keys (node, zone, region).
        *   **Detailed Examples** (Co-locating with database pods, avoiding cache nodes).

16. **[Migration Process Optimization (Convertor Pod Scheduling)](13-migration-process-optimization.md)**
    *   Overview: Optimizing temporary virt-v2v convertor pods.
    *   Flags: `--convertor-labels`, `--convertor-node-selector`, `--convertor-affinity`.
    *   Why Optimize? (Performance, Resource Management, Network Proximity).
    *   **How-To: Convertor Affinity using KARL** (Same syntax as Target Affinity).
    *   Common Use Cases (High-Performance Storage Access, Resource Isolation).
    *   Resource Sizing Considerations (CPU, Memory, I/O).

17. **[Migration Hooks](14-migration-hooks.md)**
    *   Overview: Enabling custom automation (pre-migration and post-migration).
    *   Accessing Migration Context (`plan.yml`, `workload.yml`).
    *   Parameters (`--image`, `--playbook`, `--service-account`, `--deadline`).
    *   **Detailed Examples and How-Tos**.
        *   Database Backup Hook.
        *   Using Shell Script Hooks.
        *   Adding Hooks via Plan Creation Flags.
        *   Managing Hooks via `patch planvm`.

18. **[Advanced Plan Patching](15-advanced-plan-patching.md)**
    *   **How-To: Patching Plan Settings** (`kubectl mtv patch plan`).
        *   Updating Migration Type, Transfer Network, and Placement settings.
        *   Updating Convertor Pod Optimization settings.
    *   **How-To: Patching Individual VMs** (`kubectl mtv patch planvm`).
        *   Modifying custom names, instance types, and naming templates per VM.
        *   Adding/Removing/Clearing Hooks.
    *   Best Practices: Plan-Level vs. VM-Level Changes.

### VI. Operational Excellence, Debugging, and AI Integration

19. **[Plan Lifecycle Execution](16-plan-lifecycle-execution.md)**
    *   Starting a Migration (`kubectl mtv start plan`).
    *   Warm Migration Cutover (`kubectl mtv cutover plan`).
    *   Canceling Workloads (`kubectl mtv cancel plan`).
    *   Archiving and Unarchiving Plans.

20. **[Debugging and Troubleshooting](17-debugging-and-troubleshooting.md)**
    *   Enabling Debug Output (`--v=N`).
    *   Troubleshooting Common Issues.
        *   Build/Installation Failures.
        *   Permission and Connection Issues.
        *   Convertor Pods Stuck in Pending State.
        *   Mapping Issues (Source/Target Not Found).
    *   Monitoring Techniques (Describing resources, checking Kubernetes events).

21. **[Best Practices and Security](18-best-practices-and-security.md)**
    *   Plan Management Strategies (Testing, Warm Migrations, Archiving).
    *   Provider Security (Credentials, TLS verification, RBAC).
    *   Query Optimization Tips.
    *   Secure Service Account Setup for Admin Access.

22. **[Model Context Protocol (MCP) Server Integration](19-model-context-protocol-mcp-server-integration.md)**
    *   Overview: Providing AI assistants (Claude, Cursor IDE) access to migration resources.
    *   Server Modes (Stdio Mode vs. SSE Mode).
    *   Command Line Options (`--sse`, `--host`, `--port`).
    *   **How-To: Integrating with AI Assistants** (Claude Desktop, Cursor IDE).

23. **[Integration with KubeVirt Tools](20-integration-with-kubevirt-tools.md)**
    *   Relationship between `kubectl-mtv` and `virtctl`.
    *   Using `virtctl` for post-migration VM lifecycle management (start, stop, console, ssh).

### VII. Reference and Appendices

24. **[Command Reference](21-command-reference.md)**
    *   **Global Flags** (verbose, all-namespaces, kubeconfig, context, namespace).
    *   **Resource Management Commands** (get, describe, delete with all subcommands).
    *   **Inventory Commands** (get inventory vm/network/storage/host/namespace with TSL query syntax).
    *   **Creation Commands** (create provider/plan/mapping/host/hook/vddk-image with all flags).
    *   **Plan Lifecycle Commands** (start, cancel, cutover, archive, unarchive).
    *   **Modification Commands** (patch plan/planvm/mapping/provider).
    *   **AI Integration Commands** (mcp-server with stdio and SSE modes).
    *   **Query Language Reference** (TSL operators, functions, examples).
    *   **KARL Syntax Reference** (affinity rules, topology keys, examples).
    *   **Common Command Patterns** (complete migration workflows, troubleshooting).
