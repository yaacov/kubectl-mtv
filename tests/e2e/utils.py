"""
Utility functions for kubectl-mtv e2e tests.
"""

import json
import logging
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


def wait_for_provider_ready(test_namespace, provider_name: str, timeout: int = 180):
    """Wait for a provider to have Ready condition = True using kubectl wait."""
    logging.info(f"Waiting for provider {provider_name} to be ready...")

    # Use kubectl wait to wait for the Ready condition
    wait_cmd = (
        f"wait --for=condition=Ready provider/{provider_name} --timeout={timeout}s"
    )

    try:
        test_namespace.run_kubectl_command(wait_cmd, check=True)
        logging.info(f"Provider {provider_name} is ready!")
        return True
    except Exception as e:
        # If kubectl wait fails, get the provider status for better error reporting
        try:
            status_result = test_namespace.run_mtv_command(
                f"get provider {provider_name} -o json", check=False
            )
            if status_result.returncode == 0:
                provider_list = json.loads(status_result.stdout)
                if len(provider_list) == 1:
                    provider_data = provider_list[0]
                    status = provider_data.get("status", {})
                    conditions = status.get("conditions", [])

                    # Find Ready condition for detailed error info
                    for condition in conditions:
                        if condition.get("type") == "Ready":
                            condition_status = condition.get("status", "")
                            condition_reason = condition.get("reason", "")
                            condition_message = condition.get("message", "")

                            if condition_status == "False":
                                pytest.fail(
                                    f"Provider {provider_name} failed to become ready. "
                                    f"Reason: {condition_reason}, Message: {condition_message}"
                                )
                            break
        except Exception:
            pass  # Fall back to original error

        # If we couldn't get detailed status, fail with the original kubectl wait error
        pytest.fail(
            f"Provider {provider_name} did not become ready within {timeout} seconds: {e}"
        )


def wait_for_plan_ready(
    test_namespace, plan_name: str, timeout: int = 120
) -> bool:
    """Wait for a migration plan to have Ready condition = True using kubectl wait."""
    logging.info(f"Waiting for plan {plan_name} to be ready...")

    # Use kubectl wait to wait for the Ready condition
    wait_cmd = (
        f"wait --for=condition=Ready plan/{plan_name} --timeout={timeout}s"
    )

    try:
        test_namespace.run_kubectl_command(wait_cmd, check=True)
        logging.info(f"Plan {plan_name} is ready!")
        return True
    except Exception as e:
        # If kubectl wait fails, get the plan status for better error reporting
        try:
            status_result = test_namespace.run_mtv_command(
                f"get plan {plan_name} -o json", check=False
            )
            if status_result.returncode == 0:
                plan_list = json.loads(status_result.stdout)
                if len(plan_list) >= 1:
                    plan_data = plan_list[0]
                    status = plan_data.get("status", {})
                    conditions = status.get("conditions", [])

                    # Find Ready condition for detailed error info
                    for condition in conditions:
                        if condition.get("type") == "Ready":
                            condition_status = condition.get("status", "")
                            condition_reason = condition.get("reason", "")
                            condition_message = condition.get("message", "")

                            if condition_status == "False":
                                pytest.fail(
                                    f"Plan {plan_name} failed to become ready. "
                                    f"Reason: {condition_reason}, Message: {condition_message}"
                                )
                            break
        except Exception:
            pass  # Fall back to original error

        # If we couldn't get detailed status, fail with the original kubectl wait error
        pytest.fail(
            f"Plan {plan_name} did not become ready within {timeout} seconds: {e}"
        )


def wait_for_network_mapping_ready(
    test_namespace, mapping_name: str, timeout: int = 120
) -> bool:
    """Wait for a network mapping to have Ready condition = True using kubectl wait."""
    logging.info(f"Waiting for network mapping {mapping_name} to be ready...")

    # Use kubectl wait to wait for the Ready condition
    wait_cmd = (
        f"wait --for=condition=Ready networkmap/{mapping_name} --timeout={timeout}s"
    )

    try:
        test_namespace.run_kubectl_command(wait_cmd, check=True)
        logging.info(f"Network mapping {mapping_name} is ready!")
        return True
    except Exception as e:
        # If kubectl wait fails, get the mapping status for better error reporting
        try:
            status_result = test_namespace.run_kubectl_command(
                f"get networkmap {mapping_name} -o json", check=False
            )
            if status_result.returncode == 0:
                mapping_data = json.loads(status_result.stdout)
                status = mapping_data.get("status", {})
                conditions = status.get("conditions", [])

                # Find Ready condition for detailed error info
                for condition in conditions:
                    if condition.get("type") == "Ready":
                        condition_status = condition.get("status", "")
                        condition_reason = condition.get("reason", "")
                        condition_message = condition.get("message", "")

                        if condition_status == "False":
                            pytest.fail(
                                f"Network mapping {mapping_name} failed to become ready. "
                                f"Reason: {condition_reason}, Message: {condition_message}"
                            )
                        break
        except Exception:
            pass  # Fall back to original error

        # If we couldn't get detailed status, fail with the original kubectl wait error
        pytest.fail(
            f"Network mapping {mapping_name} did not become ready within {timeout} seconds: {e}"
        )


def wait_for_storage_mapping_ready(
    test_namespace, mapping_name: str, timeout: int = 120
) -> bool:
    """Wait for a storage mapping to have Ready condition = True using kubectl wait."""
    logging.info(f"Waiting for storage mapping {mapping_name} to be ready...")

    # Use kubectl wait to wait for the Ready condition
    wait_cmd = (
        f"wait --for=condition=Ready storagemap/{mapping_name} --timeout={timeout}s"
    )

    try:
        test_namespace.run_kubectl_command(wait_cmd, check=True)
        logging.info(f"Storage mapping {mapping_name} is ready!")
        return True
    except Exception as e:
        # If kubectl wait fails, get the mapping status for better error reporting
        try:
            status_result = test_namespace.run_kubectl_command(
                f"get storagemap {mapping_name} -o json", check=False
            )
            if status_result.returncode == 0:
                mapping_data = json.loads(status_result.stdout)
                status = mapping_data.get("status", {})
                conditions = status.get("conditions", [])

                # Find Ready condition for detailed error info
                for condition in conditions:
                    if condition.get("type") == "Ready":
                        condition_status = condition.get("status", "")
                        condition_reason = condition.get("reason", "")
                        condition_message = condition.get("message", "")

                        if condition_status == "False":
                            pytest.fail(
                                f"Storage mapping {mapping_name} failed to become ready. "
                                f"Reason: {condition_reason}, Message: {condition_message}"
                            )
                        break
        except Exception:
            pass  # Fall back to original error

        # If we couldn't get detailed status, fail with the original kubectl wait error
        pytest.fail(
            f"Storage mapping {mapping_name} did not become ready within {timeout} seconds: {e}"
        )


# Load .env file when module is imported
load_env_file()
