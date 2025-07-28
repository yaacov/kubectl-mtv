"""
Utility functions for kubectl-mtv e2e tests.
"""

import json
import logging
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional

import pytest

# Configure logging
logging.basicConfig(level=logging.INFO, format="%(levelname)s: %(message)s")


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


def wait_for_resource_condition(
    test_namespace,
    resource_type: str,
    resource_name: str,
    timeout: int = 120,
    poll_interval: int = 5,
) -> bool:
    """
    Generic helper method to wait for resource conditions.

    Looks for conditions with category 'Critical' or type 'Ready':
    - If Ready condition is found with status=True, exits successfully
    - If Critical condition is found, exits with error
    - Otherwise waits for timeout like kubectl wait does

    Args:
        test_namespace: Test namespace context
        resource_type: Type of resource (provider, plan, networkmap, storagemap)
        resource_name: Name of the resource
        timeout: Maximum time to wait in seconds
        poll_interval: Time between checks in seconds

    Returns:
        bool: True if resource is ready

    Raises:
        pytest.fail: If resource fails or timeout is reached
    """
    logging.info(f"Waiting for {resource_type} {resource_name} to be ready...")

    start_time = time.time()

    while time.time() - start_time < timeout:
        try:
            # Use kubectl for all resource types - they're all Kubernetes CRDs
            status_result = test_namespace.run_kubectl_command(
                f"get {resource_type} {resource_name} -o json", check=False
            )

            if status_result.returncode != 0:
                logging.warning(
                    f"Failed to get {resource_type} {resource_name} status, retrying..."
                )
                time.sleep(poll_interval)
                continue

            # kubectl always returns single objects for specific resource names
            resource_data = json.loads(status_result.stdout)

            # Get conditions from status
            status = resource_data.get("status", {})
            conditions = status.get("conditions", [])

            # Add debug logging to see actual conditions
            logging.debug(
                f"Found {len(conditions)} conditions for {resource_type} {resource_name}"
            )
            for i, condition in enumerate(conditions):
                logging.debug(
                    f"Condition {i}: type={condition.get('type')}, status={condition.get('status')}, category={condition.get('category')}"
                )

            # Look for Critical conditions first
            for condition in conditions:
                if condition.get("category") == "Critical":
                    condition_status = condition.get("status", "")
                    condition_reason = condition.get("reason", "")
                    condition_message = condition.get("message", "")
                    last_transition_time = condition.get("lastTransitionTime", "")

                    if condition_status == "True":
                        # Check if the critical condition is at least 20 seconds old
                        if last_transition_time:
                            try:
                                # Parse the Kubernetes timestamp (RFC3339 format)
                                transition_time = datetime.fromisoformat(
                                    last_transition_time.replace("Z", "+00:00")
                                )
                                current_time = datetime.now(timezone.utc)
                                time_diff = (
                                    current_time - transition_time
                                ).total_seconds()

                                if time_diff >= 20:
                                    pytest.fail(
                                        f"{resource_type.capitalize()} {resource_name} has critical condition "
                                        f"that has persisted for {time_diff:.1f} seconds. "
                                        f"Type: {condition.get('type', '')}, "
                                        f"Reason: {condition_reason}, Message: {condition_message}"
                                    )
                                else:
                                    logging.info(
                                        f"{resource_type.capitalize()} {resource_name} has critical condition "
                                        f"but it's only {time_diff:.1f} seconds old, waiting..."
                                    )
                            except (ValueError, TypeError) as e:
                                logging.warning(
                                    f"Failed to parse lastTransitionTime '{last_transition_time}': {e}"
                                )
                                # If we can't parse the time, fail immediately as before
                                pytest.fail(
                                    f"{resource_type.capitalize()} {resource_name} has critical condition. "
                                    f"Type: {condition.get('type', '')}, "
                                    f"Reason: {condition_reason}, Message: {condition_message}"
                                )
                        else:
                            # If no lastTransitionTime is available, fail immediately as before
                            pytest.fail(
                                f"{resource_type.capitalize()} {resource_name} has critical condition. "
                                f"Type: {condition.get('type', '')}, "
                                f"Reason: {condition_reason}, Message: {condition_message}"
                            )

            # Look for Ready condition
            ready_found = False
            for condition in conditions:
                if condition.get("type") == "Ready":
                    ready_found = True
                    condition_status = condition.get("status", "")
                    condition_reason = condition.get("reason", "")
                    condition_message = condition.get("message", "")

                    if condition_status == "True":
                        logging.info(
                            f"{resource_type.capitalize()} {resource_name} is ready!"
                        )
                        return True
                    elif condition_status == "False":
                        logging.info(
                            f"{resource_type.capitalize()} {resource_name} not ready yet. "
                            f"Reason: {condition_reason}, Message: {condition_message}"
                        )
                        break

            # If no Ready condition found, log for debugging
            if not ready_found:
                logging.debug(
                    f"No Ready condition found for {resource_type} {resource_name}"
                )
                logging.debug(
                    f"Available condition types: {[c.get('type') for c in conditions]}"
                )
                logging.debug(f"Full status JSON: {json.dumps(status, indent=2)}")

        except json.JSONDecodeError as e:
            logging.warning(f"Failed to parse JSON response: {e}, retrying...")
        except Exception as e:
            logging.warning(
                f"Error checking {resource_type} {resource_name} status: {e}, retrying..."
            )

        # Wait before next poll
        time.sleep(poll_interval)

    # Timeout reached
    pytest.fail(
        f"{resource_type.capitalize()} {resource_name} did not become ready within {timeout} seconds"
    )


def wait_for_provider_ready(test_namespace, provider_name: str, timeout: int = 360):
    """Wait for a provider to have Ready condition = True."""
    return wait_for_resource_condition(
        test_namespace, "provider", provider_name, timeout
    )


def wait_for_plan_ready(test_namespace, plan_name: str, timeout: int = 180) -> bool:
    """Wait for a migration plan to have Ready condition = True."""
    return wait_for_resource_condition(test_namespace, "plan", plan_name, timeout)


def wait_for_network_mapping_ready(
    test_namespace, mapping_name: str, timeout: int = 180
) -> bool:
    """Wait for a network mapping to have Ready condition = True."""
    return wait_for_resource_condition(
        test_namespace, "networkmap", mapping_name, timeout
    )


def wait_for_storage_mapping_ready(
    test_namespace, mapping_name: str, timeout: int = 180
) -> bool:
    """Wait for a storage mapping to have Ready condition = True."""
    return wait_for_resource_condition(
        test_namespace, "storagemap", mapping_name, timeout
    )


def wait_for_host_ready(test_namespace, host_name: str, timeout: int = 360) -> bool:
    """Wait for a migration host to have Ready condition = True."""
    return wait_for_resource_condition(test_namespace, "host", host_name, timeout)


def verify_hook_playbook(test_namespace, hook_name: str, expected_playbook_content: str = None) -> bool:
    """
    Verify that a hook's spec.playbook contains the expected base64 encoded content.
    
    Args:
        test_namespace: Test namespace context
        hook_name: Name of the hook resource
        expected_playbook_content: Expected plaintext playbook content (will be base64 encoded for comparison).
                                 If None, verifies that spec.playbook is empty.
    
    Returns:
        bool: True if the playbook matches expectations
        
    Raises:
        pytest.fail: If hook doesn't exist or playbook doesn't match
    """
    import base64
    
    logging.info(f"Verifying hook {hook_name} playbook content...")
    
    try:
        # Get the hook resource
        result = test_namespace.run_kubectl_command(
            f"get hook {hook_name} -o json"
        )
        
        if result.returncode != 0:
            pytest.fail(f"Failed to get hook {hook_name}: {result.stderr}")
        
        hook_data = json.loads(result.stdout)
        spec = hook_data.get("spec", {})
        actual_playbook = spec.get("playbook", "")
        
        if expected_playbook_content is None:
            # Expect empty playbook
            if actual_playbook == "":
                logging.info(f"Hook {hook_name} has empty playbook as expected")
                return True
            else:
                pytest.fail(f"Hook {hook_name} expected empty playbook but got: {actual_playbook}")
        else:
            # Expect base64 encoded content
            expected_encoded = base64.b64encode(expected_playbook_content.encode()).decode()
            
            if actual_playbook == expected_encoded:
                logging.info(f"Hook {hook_name} playbook content matches expected base64 encoding")
                return True
            else:
                # For debugging, try to decode the actual content
                try:
                    actual_decoded = base64.b64decode(actual_playbook).decode()
                    pytest.fail(
                        f"Hook {hook_name} playbook mismatch.\n"
                        f"Expected (decoded): {expected_playbook_content}\n"
                        f"Actual (decoded):   {actual_decoded}\n"
                        f"Expected (base64):  {expected_encoded}\n"
                        f"Actual (base64):    {actual_playbook}"
                    )
                except Exception:
                    pytest.fail(
                        f"Hook {hook_name} playbook mismatch.\n"
                        f"Expected (base64): {expected_encoded}\n"
                        f"Actual (base64):   {actual_playbook}"
                    )
        
    except json.JSONDecodeError as e:
        pytest.fail(f"Failed to parse hook {hook_name} JSON: {e}")
    except Exception as e:
        pytest.fail(f"Error verifying hook {hook_name} playbook: {e}")


def delete_hosts_by_spec_id(test_namespace, host_id: str) -> None:
    """Delete all hosts in the namespace where spec.id matches the given host_id."""
    try:
        # Get all hosts in the namespace
        list_result = test_namespace.run_kubectl_command("get hosts -o json")

        if list_result.returncode != 0:
            print(f"DEBUG: Failed to list hosts: {list_result.stderr}")
            return

        import json

        hosts_data = json.loads(list_result.stdout)

        if not hosts_data.get("items"):
            print("DEBUG: No hosts found in namespace")
            return

        deleted_count = 0
        for host in hosts_data["items"]:
            host_name = host["metadata"]["name"]
            spec_id = host.get("spec", {}).get("id")

            if spec_id == host_id:
                print(
                    f"DEBUG: Found host '{host_name}' with spec.id '{spec_id}' - deleting"
                )
                delete_result = test_namespace.run_kubectl_command(
                    f"delete host {host_name} --ignore-not-found=true"
                )
                if delete_result.returncode == 0:
                    print(f"Deleted host: {host_name}")
                    deleted_count += 1
                else:
                    print(
                        f"DEBUG: Failed to delete host '{host_name}': {delete_result.stderr}"
                    )
            else:
                print(f"DEBUG: Host '{host_name}' has spec.id '{spec_id}' - skipping")

        if deleted_count == 0:
            print(f"DEBUG: No hosts found with spec.id '{host_id}'")
        else:
            print(f"DEBUG: Deleted {deleted_count} host(s) with spec.id '{host_id}'")

    except json.JSONDecodeError as e:
        print(f"DEBUG: Failed to parse hosts JSON: {e}")
    except Exception as e:
        print(f"DEBUG: Host deletion failed with exception: {e}")
        import traceback

        print(f"Traceback: {traceback.format_exc()}")


# Load .env file when module is imported
load_env_file()
