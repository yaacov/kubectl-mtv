"""
Utility functions for kubectl-mtv e2e tests.
"""

import os
from pathlib import Path
from typing import Optional


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
            print(f"Loaded environment variables from {env_path}")
        else:
            print(f"No .env file found at {env_path}")
    
    except ImportError:
        print("python-dotenv not installed, skipping .env file loading")


def get_env_with_fallback(primary_key: str, fallback_key: str, default: str = "") -> str:
    """Get environment variable with fallback to another key."""
    return os.getenv(primary_key) or os.getenv(fallback_key) or default


# Load .env file when module is imported
load_env_file()
