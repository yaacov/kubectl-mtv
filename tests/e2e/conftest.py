"""
Pytest configuration and fixtures for kubectl-mtv e2e tests.

This module provides common test fixtures and setup for e2e testing
of kubectl-mtv against an OpenShift cluster.
"""

import os
import subprocess
import tempfile
import uuid
from pathlib import Path
from typing import Generator, Dict, Any

import pytest

def pytest_addoption(parser):
    """Add custom command-line options."""
    parser.addoption(
        "--no-cleanup",
        action="store_true",
        default=False,
        help="Skip cleanup of test namespace and resources for debugging"
    )
    parser.addoption(
        "--shared-namespace",
        action="store_true",
        default=False,
        help="Use shared namespace for all tests in a collection to speed up test execution"
    )

# Import utils to load .env file
try:
    from . import utils
    # Explicitly ensure .env file is loaded
    utils.load_env_file()
except ImportError:
    # Fallback: try to load .env file manually if utils import fails
    try:
        from pathlib import Path
        from dotenv import load_dotenv
        env_path = Path(__file__).parent / ".env"
        if env_path.exists():
            load_dotenv(env_path)
            print(f"Loaded environment variables from {env_path}")
    except ImportError:
        print("Warning: Could not load .env file - python-dotenv not available")


class KubectlMTVError(Exception):
    """Exception raised when kubectl-mtv commands fail."""
    pass


class TestContext:
    """Test context with namespace and cleanup management."""
    
    def __init__(self, namespace: str, binary_path: str, no_cleanup: bool = False, shared_namespace: bool = False):
        self.namespace = namespace
        self.binary_path = binary_path
        self.no_cleanup = no_cleanup
        self.shared_namespace = shared_namespace
        self._created_resources = []
        self._session_resources = getattr(TestContext, '_session_resources', [])
        if not hasattr(TestContext, '_session_resources'):
            TestContext._session_resources = []
    
    def run_mtv_command(self, command: str, check: bool = True, capture_output: bool = True) -> subprocess.CompletedProcess:
        """Run kubectl-mtv command with test namespace."""
        full_command = f"{self.binary_path} {command} -n {self.namespace}"
        result = subprocess.run(
            full_command,
            shell=True,
            capture_output=capture_output,
            text=True
        )
        
        if check and result.returncode != 0:
            raise KubectlMTVError(
                f"Command failed: {full_command}\n"
                f"Return code: {result.returncode}\n"
                f"STDOUT: {result.stdout}\n"
                f"STDERR: {result.stderr}"
            )
        
        return result
    
    def run_kubectl_command(self, command: str, check: bool = True, capture_output: bool = True) -> subprocess.CompletedProcess:
        """Run kubectl command with test namespace."""
        full_command = f"kubectl {command} -n {self.namespace}"
        result = subprocess.run(
            full_command,
            shell=True,
            capture_output=capture_output,
            text=True
        )
        
        if check and result.returncode != 0:
            raise KubectlMTVError(
                f"Command failed: {full_command}\n"
                f"Return code: {result.returncode}\n"
                f"STDOUT: {result.stdout}\n"
                f"STDERR: {result.stderr}"
            )
        
        return result
    
    def track_resource(self, resource_type: str, resource_name: str):
        """Track a resource for cleanup."""
        resource_tuple = (resource_type, resource_name)
        self._created_resources.append(resource_tuple)
        
        # If using shared namespace, also track in session resources
        if self.shared_namespace:
            TestContext._session_resources.append(resource_tuple)
    
    def cleanup_resources(self, session_cleanup: bool = False):
        """Clean up tracked resources."""
        if self.no_cleanup:
            print(f"Skipping resource cleanup (no-cleanup mode enabled)")
            return
        
        # For shared namespace, only clean up on session cleanup or if not shared
        if self.shared_namespace and not session_cleanup:
            print(f"Skipping resource cleanup (shared namespace mode - resources will be cleaned up at session end)")
            return
            
        # Choose which resources to clean up
        resources_to_clean = self._created_resources
        if session_cleanup and self.shared_namespace:
            resources_to_clean = TestContext._session_resources
            
        for resource_type, resource_name in reversed(resources_to_clean):
            try:
                self.run_kubectl_command(f"delete {resource_type} {resource_name}", check=False)
                print(f"Cleaned up {resource_type}/{resource_name}")
            except Exception as e:
                print(f"Warning: Failed to cleanup {resource_type}/{resource_name}: {e}")
        
        # Clear the appropriate resource list after cleanup
        if session_cleanup and self.shared_namespace:
            TestContext._session_resources.clear()
        else:
            self._created_resources.clear()


def check_cluster_login() -> bool:
    """Check if we are logged into an OpenShift/Kubernetes cluster as admin."""
    try:
        # Check basic connectivity
        result = subprocess.run(
            "kubectl auth can-i '*' '*' --all-namespaces",
            shell=True,
            capture_output=True,
            text=True
        )
        
        if result.returncode != 0:
            return False
        
        # Check if we can create namespaces (admin privilege)
        result = subprocess.run(
            "kubectl auth can-i create namespaces",
            shell=True,
            capture_output=True,
            text=True
        )
        
        return result.returncode == 0 and "yes" in result.stdout.lower()
        
    except Exception:
        return False


def find_kubectl_mtv_binary() -> str:
    """Find the kubectl-mtv binary."""
    # Get the project root directory (tests/e2e -> project root is two levels up)
    current_dir = Path(__file__).parent  # tests/e2e directory
    project_root = current_dir.parent.parent  # project root directory
    
    # First try in the project root (most common location after 'make')
    binary_path = project_root / "kubectl-mtv"
    if binary_path.exists() and binary_path.is_file():
        return str(binary_path)
    
    # Try in PATH
    result = subprocess.run(
        "which kubectl-mtv",
        shell=True,
        capture_output=True,
        text=True
    )
    
    if result.returncode == 0:
        return result.stdout.strip()
        
    raise FileNotFoundError(
        "kubectl-mtv binary not found. Please build it first with 'make' or ensure it's in PATH."
    )


@pytest.fixture(scope="session")
def cluster_check():
    """Ensure we are logged into a cluster with admin privileges."""
    if not check_cluster_login():
        pytest.skip("Not logged into OpenShift/Kubernetes cluster with admin privileges")


@pytest.fixture(scope="session")
def kubectl_mtv_binary():
    """Find and return the kubectl-mtv binary path."""
    return find_kubectl_mtv_binary()


@pytest.fixture
def test_namespace(cluster_check, kubectl_mtv_binary, request) -> Generator[TestContext, None, None]:
    """Create a temporary namespace for testing and provide test context.
    
    If --shared-namespace is used, this will return the shared namespace context.
    Otherwise, it creates a new namespace for each test.
    """
    # Check if shared namespace mode is enabled
    shared_namespace_mode = request.config.getoption("--shared-namespace")
    
    if shared_namespace_mode:
        # Use or create the shared namespace
        if not hasattr(request.session, '_shared_test_context'):
            # Create shared namespace for the first time in this session
            namespace_name = f"kubectl-mtv-shared-{uuid.uuid4().hex[:8]}"
            no_cleanup = request.config.getoption("--no-cleanup")
            
            # Create namespace
            subprocess.run(
                f"kubectl create namespace {namespace_name}",
                shell=True,
                check=True
            )
            
            print(f"\n=== SHARED NAMESPACE MODE ===")
            print(f"Shared test namespace: {namespace_name}")
            print(f"All tests will run in this namespace")
            if no_cleanup:
                print(f"Cleanup disabled - namespace will be preserved for debugging")
            print(f"==============================\n")
            
            # Create and store shared context
            context = TestContext(namespace_name, kubectl_mtv_binary, no_cleanup, shared_namespace=True)
            request.session._shared_test_context = context
            request.session._shared_namespace_name = namespace_name
            
            # Register session cleanup
            def cleanup_shared_namespace():
                if not no_cleanup:
                    print(f"\n=== SESSION CLEANUP ===")
                    context.cleanup_resources(session_cleanup=True)
                    print(f"=======================\n")
                    
                    # Cleanup namespace
                    subprocess.run(
                        f"kubectl delete namespace {namespace_name} --ignore-not-found=true",
                        shell=True,
                        check=False
                    )
                else:
                    print(f"\n=== DEBUG INFO ===")
                    print(f"Skipping session resource cleanup in namespace: {namespace_name}")
                    if hasattr(TestContext, '_session_resources') and TestContext._session_resources:
                        print(f"Created resources:")
                        for resource_type, resource_name in TestContext._session_resources:
                            print(f"  - {resource_type}/{resource_name}")
                    print(f"Shared namespace {namespace_name} preserved for debugging")
                    print(f"==================\n")
            
            request.addfinalizer(cleanup_shared_namespace)
        
        # Return the shared context
        yield request.session._shared_test_context
        return
    
    # Create individual namespace for this test (original behavior)
    namespace_name = f"kubectl-mtv-test-{uuid.uuid4().hex[:8]}"
    
    # Check if cleanup should be skipped
    no_cleanup = request.config.getoption("--no-cleanup")
    
    # Create namespace
    subprocess.run(
        f"kubectl create namespace {namespace_name}",
        shell=True,
        check=True
    )
    
    # Print namespace info if no cleanup is requested
    if no_cleanup:
        print(f"\n=== DEBUG MODE ===")
        print(f"Test namespace: {namespace_name}")
        print(f"Cleanup disabled - namespace will be preserved for debugging")
        print(f"To manually cleanup later, run: kubectl delete namespace {namespace_name}")
        print(f"==================\n")
    
    try:
        # Create test context
        context = TestContext(namespace_name, kubectl_mtv_binary, no_cleanup, shared_namespace=False)
        yield context
        
        # Cleanup tracked resources (unless no-cleanup is specified)
        if not no_cleanup:
            context.cleanup_resources()
        else:
            print(f"\n=== DEBUG INFO ===")
            print(f"Skipping resource cleanup in namespace: {namespace_name}")
            if context._created_resources:
                print(f"Created resources:")
                for resource_type, resource_name in context._created_resources:
                    print(f"  - {resource_type}/{resource_name}")
            print(f"==================\n")
        
    finally:
        # Cleanup namespace (unless no-cleanup is specified)
        if not no_cleanup:
            subprocess.run(
                f"kubectl delete namespace {namespace_name} --ignore-not-found=true",
                shell=True,
                check=False
            )
        else:
            print(f"Namespace {namespace_name} preserved for debugging")


@pytest.fixture(scope="session")
def provider_credentials() -> Dict[str, Any]:
    """Load provider credentials from environment variables."""
    return {
        # VMware vSphere credentials
        "vsphere": {
            "url": os.getenv("VSPHERE_URL"),
            "username": os.getenv("VSPHERE_USERNAME"),
            "password": os.getenv("VSPHERE_PASSWORD"),
            "cacert": os.getenv("VSPHERE_CACERT"),
            "insecure": os.getenv("VSPHERE_INSECURE_SKIP_TLS", "false").lower() == "true",
            "vddk_init_image": os.getenv("VSPHERE_VDDK_INIT_IMAGE"),
        },
        
        # oVirt credentials
        "ovirt": {
            "url": os.getenv("OVIRT_URL"),
            "username": os.getenv("OVIRT_USERNAME"),
            "password": os.getenv("OVIRT_PASSWORD"),
            "cacert": os.getenv("OVIRT_CACERT"),
            "insecure": os.getenv("OVIRT_INSECURE_SKIP_TLS", "false").lower() == "true",
        },
        
        # OpenStack credentials
        "openstack": {
            "url": os.getenv("OPENSTACK_URL"),
            "username": os.getenv("OPENSTACK_USERNAME"),
            "password": os.getenv("OPENSTACK_PASSWORD"),
            "domain_name": os.getenv("OPENSTACK_DOMAIN_NAME"),
            "project_name": os.getenv("OPENSTACK_PROJECT_NAME"),
            "region_name": os.getenv("OPENSTACK_REGION_NAME"),
            "cacert": os.getenv("OPENSTACK_CACERT"),
            "insecure": os.getenv("OPENSTACK_INSECURE_SKIP_TLS", "false").lower() == "true",
        },
        
        # OVA file path
        "ova": {
            "url": os.getenv("OVA_URL"),  # URL or file path to OVA
        },
        
        # OpenShift target cluster (usually current cluster)
        "openshift": {
            "url": os.getenv("OPENSHIFT_TARGET_URL"),
            "token": os.getenv("OPENSHIFT_TARGET_TOKEN"),
            "cacert": os.getenv("OPENSHIFT_CACERT"),
            "insecure": os.getenv("OPENSHIFT_INSECURE_SKIP_TLS", "false").lower() == "true",
        },
    }


@pytest.fixture
def temp_file() -> Generator[str, None, None]:
    """Create a temporary file and clean it up after test."""
    with tempfile.NamedTemporaryFile(mode='w', delete=False) as f:
        temp_path = f.name
    
    try:
        yield temp_path
    finally:
        try:
            os.unlink(temp_path)
        except OSError:
            pass
