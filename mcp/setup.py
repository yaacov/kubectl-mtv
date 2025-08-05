#!/usr/bin/env python3
"""
Setup script for kubectl-mtv MCP server
"""

from setuptools import setup, find_packages

setup(
    name="kubectl-mtv-mcp",
    version="1.0.0",
    description="MCP Server for kubectl-mtv - Migration Toolkit for Virtualization",
    author="kubectl-mtv MCP Server",
    packages=find_packages(),
    install_requires=[
        "mcp>=1.0.0",
    ],
    python_requires=">=3.8",
    entry_points={
        "console_scripts": [
            "kubectl-mtv-mcp=kubectl_mtv_server:main",
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