#!/usr/bin/env python3
"""
Setup script for kubectl-mtv MCP servers
"""

from setuptools import setup, find_namespace_packages

def main_kubectl_mtv_server():
    """Entry point for kubectl-mtv MCP read server"""
    from kubev2v.kubectl_mtv_server import mcp
    mcp.run()

def main_kubectl_mtv_write_server():
    """Entry point for kubectl-mtv MCP write server"""
    from kubev2v.kubectl_mtv_write_server import mcp
    mcp.run()

setup(
    name="mtv-mcp",
    version="1.0.0",
    description="MCP Servers for kubectl-mtv - Migration Toolkit for Virtualization",
    author="kubectl-mtv MCP Server",
    packages=find_namespace_packages(include=['kubev2v*']),
    install_requires=[
        "fastmcp>=2.11.0.9",
    ],
    python_requires=">=3.8",
    entry_points={
        "console_scripts": [
            "kubectl-mtv-mcp=setup:main_kubectl_mtv_server",
            "kubectl-mtv-write-mcp=setup:main_kubectl_mtv_write_server",
        ],
    },
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers", 
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
    ],
)