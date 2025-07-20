"""
Test cases for kubectl-mtv version command.

This test validates the version subcommand functionality.
"""

import json
import re
import subprocess

import pytest
import yaml


@pytest.mark.version
class TestVersion:
    """Test cases for the version command."""
    
    def test_version_command_basic(self, test_namespace):
        """Test basic version command execution."""
        result = test_namespace.run_mtv_command("version")
        
        assert result.returncode == 0
        assert result.stdout is not None
        assert len(result.stdout.strip()) > 0
        
        # Should contain client version info
        assert "Operator version" in result.stdout
    
    def test_version_command_json_output(self, test_namespace):
        """Test version command with JSON output format."""
        result = test_namespace.run_mtv_command("version -o json")
        
        assert result.returncode == 0
        
        # Parse JSON output
        try:
            version_data = json.loads(result.stdout)
            assert isinstance(version_data, dict)
            
            # Should have required fields
            required_fields = ["clientVersion", "operatorVersion", "operatorStatus", "inventoryURL", "inventoryStatus"]
            for field in required_fields:
                assert field in version_data, f"Missing required field: {field}"
            
            # Client version should not be empty
            assert version_data["clientVersion"], "Client version should not be empty"
            
        except json.JSONDecodeError:
            pytest.fail(f"Version command did not return valid JSON: {result.stdout}")
    
    def test_version_command_yaml_output(self, test_namespace):
        """Test version command with YAML output format."""
        result = test_namespace.run_mtv_command("version -o yaml")
        
        assert result.returncode == 0
        
        # Parse YAML output
        try:
            version_data = yaml.safe_load(result.stdout)
            assert isinstance(version_data, dict)
            
            # Should have required fields
            required_fields = ["clientVersion", "operatorVersion", "operatorStatus", "inventoryURL", "inventoryStatus"]
            for field in required_fields:
                assert field in version_data, f"Missing required field: {field}"
            
            # Client version should not be empty
            assert version_data["clientVersion"], "Client version should not be empty"
            
        except yaml.YAMLError:
            pytest.fail(f"Version command did not return valid YAML: {result.stdout}")
    
    def test_version_command_table_output(self, test_namespace):
        """Test version command with table output format."""
        result = test_namespace.run_mtv_command("version -o table")
        
        assert result.returncode == 0
        assert result.stdout is not None
        
        # Table output should contain version information
        assert "Operator version" in result.stdout
    
    def test_version_shows_server_info_when_available(self, test_namespace):
        """Test that version command shows server info when MTV is installed."""
        result = test_namespace.run_mtv_command("version")
        
        assert result.returncode == 0
        
        # The output should contain information about MTV/Forklift installation
        # This might show "not available" if MTV is not installed, which is also valid
        output_lower = result.stdout.lower()
        assert any(keyword in output_lower for keyword in [
            "server", "inventory", "mtv", "forklift", "available", "not available"
        ])
    
    def test_version_help_flag(self, test_namespace):
        """Test version command help flag."""
        result = test_namespace.run_mtv_command("version --help")
        
        assert result.returncode == 0
        assert "version" in result.stdout.lower()
        assert "usage" in result.stdout.lower() or "help" in result.stdout.lower()
    
    def test_version_matches_build_info(self, kubectl_mtv_binary):
        """Test that version output matches build information."""
        # Run version command directly with binary
        result = subprocess.run(
            [kubectl_mtv_binary, "version"],
            capture_output=True,
            text=True
        )
        
        assert result.returncode == 0
        
        # Version should not be "unknown" in a proper build
        # This test might be skipped in development builds
        if "unknown" not in result.stdout.lower():
            # Should contain version number pattern (e.g., v1.2.3 or similar)
            version_pattern = r'v?\d+\.\d+\.\d+'
            assert re.search(version_pattern, result.stdout), \
                f"Version output should contain version number: {result.stdout}"
