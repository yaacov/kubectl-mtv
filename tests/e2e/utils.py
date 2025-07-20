"""
Utility functions for kubectl-mtv e2e tests.
"""

import json
import logging
import os
import time
import uuid
from pathlib import Path
from typing import Optional

import pytest

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(levelname)s: %(message)s')


def load_env_file(env_file: Optional[str] = None) -> None:
    """Load environment variables from .env file if it exists."""
    try:
        from dotenv import load_dotenv
        
        if env_file:
            env_path = Path(env_file)
        else:
            # Look for .env file in the same directory as this script
            env_path = Path(__file__).parent / ".env"
        
        if env_path.exists():
            load_dotenv(env_path)
            logging.info(f"Loaded environment variables from {env_path}")
        else:
            logging.info(f"No .env file found at {env_path}")
    
    except ImportError:
        logging.warning("python-dotenv not installed, skipping .env file loading")


def get_env_with_fallback(primary_key: str, fallback_key: str, default: str = "") -> str:
    """Get environment variable with fallback to another key."""
    return os.getenv(primary_key) or os.getenv(fallback_key) or default


def generate_unique_resource_name(base_name: str) -> str:
    """Generate a unique resource name with UUID suffix to avoid conflicts in shared namespace."""
    return f"{base_name}-{uuid.uuid4().hex[:8]}"


def verify_provider_created(test_namespace, provider_name: str, provider_type: str):
    """Verify that a provider was created successfully."""
    # Wait a moment for provider to be created
    time.sleep(2)
    
    # Check if provider exists
    result = test_namespace.run_mtv_command(f"get provider {provider_name} -o json", check=False)
    
    # If command failed, provider doesn't exist
    if result.returncode != 0:
        pytest.fail(f"Provider {provider_name} not found. Command output: {result.stderr}")
    
    # Parse provider data
    try:
        provider_list = json.loads(result.stdout)
    except json.JSONDecodeError as e:
        pytest.fail(f"Failed to parse provider JSON output: {e}. Output: {result.stdout}")
    
    if len(provider_list) != 1:
        pytest.fail(f"Expected 1 provider, got {len(provider_list)}")
    
    provider_data = provider_list[0]
    
    # Verify basic provider properties
    metadata = provider_data.get("metadata", {})
    spec = provider_data.get("spec", {})
    
    if metadata.get("name") != provider_name:
        pytest.fail(f"Provider name mismatch: expected {provider_name}, got {metadata.get('name')}")
    
    if spec.get("type") != provider_type:
        pytest.fail(f"Provider type mismatch: expected {provider_type}, got {spec.get('type')}")
    
    # Check provider status for errors
    status = provider_data.get("status", {})
    logging.info(f"Provider {provider_name} status: {status}")
    
    # Check for error conditions
    conditions = status.get("conditions", [])
    for condition in conditions:
        condition_type = condition.get("type", "")
        condition_status = condition.get("status", "")
        condition_reason = condition.get("reason", "")
        condition_message = condition.get("message", "")
        
        # Fail if there are explicit error conditions
        if condition_type in ["Ready", "ConnectionTestSucceeded"] and condition_status == "False":
            if condition_reason in ["Error", "Failed", "ValidationFailed", "ConnectionFailed"]:
                pytest.fail(
                    f"Provider {provider_name} is in error state. "
                    f"Condition: {condition_type}={condition_status}, "
                    f"Reason: {condition_reason}, Message: {condition_message}"
                )
    
    # Additional validation: check if provider has been marked as invalid
    if status.get("phase") == "Failed":
        pytest.fail(f"Provider {provider_name} is in Failed phase: {status}")
    
    logging.info(f"Provider {provider_name} verified successfully")


# Load .env file when module is imported
load_env_file()
