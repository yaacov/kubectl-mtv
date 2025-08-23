#!/usr/bin/env python3
"""
Setup script for kubectl-mtv MCP servers
"""

from setuptools import setup, find_namespace_packages

setup(
    name="mtv-mcp",
    version="1.0.6",
    description="MCP Servers for kubectl-mtv - Migration Toolkit for Virtualization",
    author="kubectl-mtv MCP Server",
    packages=find_namespace_packages(include=["kubev2v*"]),
    install_requires=[
        "fastmcp>=2.11.0.9",
    ],
    python_requires=">=3.8",
    entry_points={
        "console_scripts": [
            "kubectl-mtv-mcp=kubev2v.kubectl_mtv_server:mcp.run",
            "kubectl-mtv-write-mcp=kubev2v.kubectl_mtv_write_server:mcp.run",
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
