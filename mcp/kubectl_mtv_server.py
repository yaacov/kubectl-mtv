#!/usr/bin/env python3
"""
FastMCP Server for kubectl-mtv

This server provides tools to interact with Migration Toolkit for Virtualization (MTV)
through kubectl-mtv commands. It assumes kubectl-mtv is installed and the user is 
logged into a Kubernetes cluster with MTV deployed.

MTV Context:
- MTV helps migrate VMs from other KubeVirt clusters, vSphere, oVirt, OpenStack, and OVA files to KubeVirt
- Typical workflow: Provider -> Inventory Discovery -> Mappings -> Plans -> Migration
- These tools provide read-only discovery and monitoring capabilities
- See MTV_CONTEXT.md for detailed guidance on tool usage patterns

Tool Categories:
- Provider Management: list_providers
- Migration Planning: list_plans, list_mappings, list_hosts, list_hooks  
- Inventory Discovery: list_inventory_* tools for exploring source environments
"""

import os
import subprocess
from typing import Any

from fastmcp import FastMCP

# Get the directory containing this script and the docs directory
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
DOCS_DIR = os.path.join(os.path.dirname(SCRIPT_DIR), "docs")

# Initialize the FastMCP server
mcp = FastMCP("kubectl-mtv")


class KubectlMTVError(Exception):
    """Custom exception for kubectl-mtv command errors."""
    pass


async def run_kubectl_mtv_command(args: list[str]) -> str:
    """Run a kubectl-mtv command and return the output."""
    try:
        cmd = ["kubectl-mtv"] + args
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout
    except subprocess.CalledProcessError as e:
        error_msg = f"Command failed: {' '.join(cmd)}\nError: {e.stderr}"
        raise KubectlMTVError(error_msg) from e
    except FileNotFoundError:
        raise KubectlMTVError("kubectl-mtv not found in PATH") from None


async def _build_base_args(arguments: dict[str, Any]) -> list[str]:
    """Build base arguments for kubectl-mtv commands."""
    args = []
    
    if arguments.get("all_namespaces", False):
        args.extend(["-A"])
    elif "namespace" in arguments and arguments["namespace"]:
        args.extend(["-n", arguments["namespace"]])
    
    output_format = arguments.get("output_format", "json")
    if output_format != "table":
        args.extend(["-o", output_format])
    
    return args


async def _list_inventory(arguments: dict[str, Any], resource_type: str) -> str:
    """List inventory resources from a provider."""
    provider_name = arguments["provider_name"]
    args = ["get", "inventory", resource_type, provider_name]
    
    if "namespace" in arguments and arguments["namespace"]:
        args.extend(["-n", arguments["namespace"]])
    
    if "query" in arguments and arguments["query"]:
        args.extend(["-q", arguments["query"]])
    
    output_format = arguments.get("output_format", "json")
    if output_format != "table":
        args.extend(["-o", output_format])
    
    return await run_kubectl_mtv_command(args)


async def _list_inventory_generic(arguments: dict[str, Any]) -> str:
    """List any inventory resource type from a provider."""
    resource_type = arguments["resource_type"]
    provider_name = arguments["provider_name"]
    args = ["get", "inventory", resource_type, provider_name]
    
    if "namespace" in arguments and arguments["namespace"]:
        args.extend(["-n", arguments["namespace"]])
    
    if "query" in arguments and arguments["query"]:
        args.extend(["-q", arguments["query"]])
    
    output_format = arguments.get("output_format", "json")
    if output_format != "table":
        args.extend(["-o", output_format])
    
    return await run_kubectl_mtv_command(args)


# Resources for documentation
@mcp.resource("kubectl-mtv://usage")
async def usage_guide() -> str:
    """Comprehensive kubectl-mtv usage documentation covering all commands, providers, plans, mappings, and workflows."""
    try:
        file_path = os.path.join(DOCS_DIR, "README-usage.md")
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    except FileNotFoundError:
        return "kubectl-mtv usage documentation file not found"
    except Exception as e:
        return f"Error reading kubectl-mtv usage documentation: {str(e)}"


@mcp.resource("kubectl-mtv://inventory")
async def inventory_guide() -> str:
    """Complete guide to inventory discovery, query language, and filtering resources from providers."""
    try:
        file_path = os.path.join(DOCS_DIR, "README_inventory.md")
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    except FileNotFoundError:
        return "kubectl-mtv inventory documentation file not found"
    except Exception as e:
        return f"Error reading kubectl-mtv inventory documentation: {str(e)}"


@mcp.resource("kubectl-mtv://hooks")
async def hooks_guide() -> str:
    """Guide to migration hooks for customizing and automating migration workflows."""
    try:
        file_path = os.path.join(DOCS_DIR, "README_hooks.md")
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    except FileNotFoundError:
        return "kubectl-mtv hooks documentation file not found"
    except Exception as e:
        return f"Error reading kubectl-mtv hooks documentation: {str(e)}"


@mcp.resource("kubectl-mtv://mapping-pairs")
async def mapping_pairs_guide() -> str:
    """Guide to network and storage mapping configuration using pairs syntax."""
    try:
        file_path = os.path.join(DOCS_DIR, "README_mapping_pairs.md")
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    except FileNotFoundError:
        return "kubectl-mtv mapping pairs documentation file not found"
    except Exception as e:
        return f"Error reading kubectl-mtv mapping pairs documentation: {str(e)}"


@mcp.resource("kubectl-mtv://demo")
async def demo_guide() -> str:
    """Step-by-step tutorial demonstrating a complete migration workflow."""
    try:
        file_path = os.path.join(DOCS_DIR, "README_demo.md")
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    except FileNotFoundError:
        return "kubectl-mtv demo documentation file not found"
    except Exception as e:
        return f"Error reading kubectl-mtv demo documentation: {str(e)}"


# Tool functions using FastMCP
@mcp.tool()
async def list_providers(
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json"
) -> str:
    """List all MTV providers in the cluster.
    
    Providers connect MTV to source virtualization platforms (vSphere, oVirt, OpenStack, OVA, other KubeVirt clusters). 
    This is typically the first step in understanding your MTV setup.
    
    Args:
        namespace: Kubernetes namespace to query (optional, defaults to current namespace)
        all_namespaces: List providers across all namespaces
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted provider information or table output
    """
    args = ["get", "provider"] + await _build_base_args({
        "namespace": namespace, 
        "all_namespaces": all_namespaces, 
        "output_format": output_format
    })
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def list_plans(
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json",
    watch: bool = False
) -> str:
    """List all MTV migration plans in the cluster.
    
    Plans define which VMs to migrate and track migration progress. 
    Use this to monitor active migrations or see migration history.
    
    Args:
        namespace: Kubernetes namespace to query (optional)
        all_namespaces: List plans across all namespaces
        output_format: Output format (json, yaml, or table)
        watch: Watch for changes (not recommended for MCP)
        
    Returns:
        JSON/YAML formatted plan information or table output
    """
    args = ["get", "plan"] + await _build_base_args({
        "namespace": namespace,
        "all_namespaces": all_namespaces,
        "output_format": output_format
    })
    
    if watch:
        args.append("--watch")
    
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def list_mappings(
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json"
) -> str:
    """List all MTV mappings (network and storage) in the cluster.
    
    Args:
        namespace: Kubernetes namespace to query (optional)
        all_namespaces: List mappings across all namespaces
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted mapping information or table output
    """
    args = ["get", "mapping"] + await _build_base_args({
        "namespace": namespace,
        "all_namespaces": all_namespaces,
        "output_format": output_format
    })
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def list_hosts(
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json"
) -> str:
    """List all MTV migration hosts in the cluster.
    
    Args:
        namespace: Kubernetes namespace to query (optional)
        all_namespaces: List hosts across all namespaces
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted host information or table output
    """
    args = ["get", "host"] + await _build_base_args({
        "namespace": namespace,
        "all_namespaces": all_namespaces,
        "output_format": output_format
    })
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def list_hooks(
    namespace: str = "",
    all_namespaces: bool = False,
    output_format: str = "json"
) -> str:
    """List all MTV migration hooks in the cluster.
    
    Args:
        namespace: Kubernetes namespace to query (optional)
        all_namespaces: List hooks across all namespaces  
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted hook information or table output
    """
    args = ["get", "hook"] + await _build_base_args({
        "namespace": namespace,
        "all_namespaces": all_namespaces,
        "output_format": output_format
    })
    return await run_kubectl_mtv_command(args)


@mcp.tool()
async def list_inventory_vms(
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List VMs from a specific provider's inventory.
    
    This shows all VMs available for migration from a source platform. 
    Use queries to filter by memory, CPU, power state, or name. Essential for migration planning.
    
    Args:
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query (e.g., 'where memoryMB > 4096')
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted VM inventory or table output
    """
    return await _list_inventory({
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    }, "vm")


@mcp.tool()
async def list_inventory_networks(
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List networks from a specific provider's inventory.
    
    Args:
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted network inventory or table output
    """
    return await _list_inventory({
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    }, "network")


@mcp.tool()
async def list_inventory_storage(
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List storage from a specific provider's inventory.
    
    Args:
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted storage inventory or table output
    """
    return await _list_inventory({
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    }, "storage")


@mcp.tool()
async def list_inventory_hosts(
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List hosts from a specific provider's inventory.
    
    Args:
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted host inventory or table output
    """
    return await _list_inventory({
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    }, "host")


@mcp.tool()
async def list_inventory_clusters(
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List clusters from a specific provider's inventory (oVirt, vSphere).
    
    Args:
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted cluster inventory or table output
    """
    return await _list_inventory({
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    }, "cluster")


@mcp.tool()
async def list_inventory_datacenters(
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List datacenters from a specific provider's inventory (oVirt, vSphere).
    
    Args:
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted datacenter inventory or table output
    """
    return await _list_inventory({
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    }, "datacenter")


@mcp.tool()
async def list_inventory_generic(
    resource_type: str,
    provider_name: str,
    namespace: str = "",
    query: str = "",
    output_format: str = "json"
) -> str:
    """List any inventory resource type from a provider (advanced users).
    
    Args:
        resource_type: Type of inventory resource to list (vm, network, storage, host, cluster, 
                      datacenter, datastore, disk, disk-profile, flavor, folder, image, 
                      instance, namespace, nic-profile, project, pvc, resource-pool, data-volume)
        provider_name: Name of the provider to query
        namespace: Kubernetes namespace containing the provider (optional)
        query: Optional filter query
        output_format: Output format (json, yaml, or table)
        
    Returns:
        JSON/YAML formatted inventory or table output
    """
    return await _list_inventory_generic({
        "resource_type": resource_type,
        "provider_name": provider_name,
        "namespace": namespace,
        "query": query,
        "output_format": output_format
    })


if __name__ == "__main__":
    mcp.run()