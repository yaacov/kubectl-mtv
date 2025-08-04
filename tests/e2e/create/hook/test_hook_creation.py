"""
Test cases for kubectl-mtv hook creation.

This test validates the creation of hook resources with various configurations,
including playbook content handling and base64 encoding verification.
"""

import time
import tempfile
import os

import pytest

from e2e.utils import verify_hook_playbook


# Sample Ansible playbook content for testing
SAMPLE_PLAYBOOK = """---
- name: Test Migration Hook
  hosts: localhost
  gather_facts: false
  tasks:
    - name: Print migration info
      debug:
        msg: "Migration hook executed successfully"

    - name: Check system readiness
      command: echo "System is ready for migration"
      register: readiness_check

    - name: Display readiness result
      debug:
        var: readiness_check.stdout
"""

MINIMAL_PLAYBOOK = """---
- name: Minimal Hook
  hosts: localhost
  tasks:
    - debug:
        msg: "Hook executed"
"""


@pytest.mark.create
@pytest.mark.hook
class TestHookCreation:
    """Test cases for hook creation with various configurations."""

    def test_create_hook_minimal(self, test_namespace):
        """Test creating a minimal hook with only required parameters."""
        hook_name = f"test-hook-minimal-{int(time.time())}"

        # Create hook with only required image parameter
        create_cmd = f"create hook {hook_name} --image nginx:latest"

        # Create hook
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        assert f"hook/{hook_name} created" in result.stdout

        # Track for cleanup
        test_namespace.track_resource("hook", hook_name)

        # Verify hook exists and has expected properties
        result = test_namespace.run_kubectl_command(f"get hook {hook_name} -o json")
        assert result.returncode == 0

        import json

        hook_data = json.loads(result.stdout)
        spec = hook_data.get("spec", {})

        # Verify required fields
        assert spec.get("image") == "nginx:latest"
        assert spec.get("serviceAccount", "") == ""  # Should be empty or not present
        assert spec.get("playbook", "") == ""  # Should be empty or not present
        assert spec.get("deadline", 0) == 0  # Should be 0 or not present

        # Verify playbook is empty as expected
        verify_hook_playbook(test_namespace, hook_name, expected_playbook_content=None)

    def test_create_hook_with_service_account(self, test_namespace):
        """Test creating a hook with a service account."""
        hook_name = f"test-hook-sa-{int(time.time())}"

        # Create hook with service account
        create_cmd = (
            f"create hook {hook_name} --image nginx:latest --service-account default"
        )

        # Create hook
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        assert f"hook/{hook_name} created" in result.stdout

        # Track for cleanup
        test_namespace.track_resource("hook", hook_name)

        # Verify hook has service account set
        result = test_namespace.run_kubectl_command(f"get hook {hook_name} -o json")
        assert result.returncode == 0

        import json

        hook_data = json.loads(result.stdout)
        spec = hook_data.get("spec", {})

        assert spec.get("serviceAccount", "") == "default"

    def test_create_hook_with_deadline(self, test_namespace):
        """Test creating a hook with a deadline."""
        hook_name = f"test-hook-deadline-{int(time.time())}"

        # Create hook with deadline
        create_cmd = f"create hook {hook_name} --image nginx:latest --deadline 300"

        # Create hook
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        assert f"hook/{hook_name} created" in result.stdout

        # Track for cleanup
        test_namespace.track_resource("hook", hook_name)

        # Verify hook has deadline set
        result = test_namespace.run_kubectl_command(f"get hook {hook_name} -o json")
        assert result.returncode == 0

        import json

        hook_data = json.loads(result.stdout)
        spec = hook_data.get("spec", {})

        assert spec.get("deadline", 0) == 300

    def test_create_hook_with_inline_playbook(self, test_namespace):
        """Test creating a hook with inline playbook content."""
        hook_name = f"test-hook-inline-{int(time.time())}"

        # Create hook with inline playbook
        create_cmd = f"create hook {hook_name} --image nginx:latest --playbook '{MINIMAL_PLAYBOOK}'"

        # Create hook
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0
        assert f"hook/{hook_name} created" in result.stdout

        # Track for cleanup
        test_namespace.track_resource("hook", hook_name)

        # Verify playbook content is properly base64 encoded
        verify_hook_playbook(
            test_namespace, hook_name, expected_playbook_content=MINIMAL_PLAYBOOK
        )

    def test_create_hook_with_playbook_file(self, test_namespace):
        """Test creating a hook with playbook content from a file using @ convention."""
        hook_name = f"test-hook-file-{int(time.time())}"

        # Create temporary playbook file
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            f.write(SAMPLE_PLAYBOOK)
            playbook_file = f.name

        try:
            # Create hook with playbook from file
            create_cmd = f"create hook {hook_name} --image nginx:latest --playbook @{playbook_file}"

            # Create hook
            result = test_namespace.run_mtv_command(create_cmd)
            assert result.returncode == 0
            assert f"hook/{hook_name} created" in result.stdout

            # Track for cleanup
            test_namespace.track_resource("hook", hook_name)

            # Verify playbook content is properly base64 encoded
            verify_hook_playbook(
                test_namespace, hook_name, expected_playbook_content=SAMPLE_PLAYBOOK
            )

        finally:
            # Clean up temporary file
            os.unlink(playbook_file)

    def test_create_hook_full_configuration(self, test_namespace):
        """Test creating a hook with all parameters specified."""
        hook_name = f"test-hook-full-{int(time.time())}"

        # Create temporary playbook file
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            f.write(SAMPLE_PLAYBOOK)
            playbook_file = f.name

        try:
            # Create hook with all parameters
            create_cmd = (
                f"create hook {hook_name} "
                f"--image registry.redhat.io/ubi8/ubi:latest "
                f"--service-account migration-sa "
                f"--deadline 600 "
                f"--playbook @{playbook_file}"
            )

            # Create hook
            result = test_namespace.run_mtv_command(create_cmd)
            assert result.returncode == 0
            assert f"hook/{hook_name} created" in result.stdout

            # Track for cleanup
            test_namespace.track_resource("hook", hook_name)

            # Verify all fields are set correctly
            result = test_namespace.run_kubectl_command(f"get hook {hook_name} -o json")
            assert result.returncode == 0

            import json

            hook_data = json.loads(result.stdout)
            spec = hook_data.get("spec", {})

            assert spec.get("image") == "registry.redhat.io/ubi8/ubi:latest"
            assert spec.get("serviceAccount", "") == "migration-sa"
            assert spec.get("deadline", 0) == 600

            # Verify playbook content is properly base64 encoded
            verify_hook_playbook(
                test_namespace, hook_name, expected_playbook_content=SAMPLE_PLAYBOOK
            )

        finally:
            # Clean up temporary file
            os.unlink(playbook_file)

    def test_create_hook_duplicate_name(self, test_namespace):
        """Test that creating a hook with an existing name fails."""
        hook_name = f"test-hook-duplicate-{int(time.time())}"

        # Create first hook
        create_cmd = f"create hook {hook_name} --image nginx:latest"
        result = test_namespace.run_mtv_command(create_cmd)
        assert result.returncode == 0

        # Track for cleanup
        test_namespace.track_resource("hook", hook_name)

        # Try to create hook with same name
        result = test_namespace.run_mtv_command(create_cmd, check=False)
        assert result.returncode != 0
        assert "already exists" in result.stderr

    def test_create_hook_missing_image(self, test_namespace):
        """Test that creating a hook without required image parameter fails."""
        hook_name = f"test-hook-no-image-{int(time.time())}"

        # Try to create hook without image
        create_cmd = f"create hook {hook_name} --service-account default"
        result = test_namespace.run_mtv_command(create_cmd, check=False)
        assert result.returncode != 0
        assert "image is required" in result.stderr or "required flag" in result.stderr

    def test_create_hook_invalid_deadline(self, test_namespace):
        """Test that creating a hook with negative deadline fails."""
        hook_name = f"test-hook-bad-deadline-{int(time.time())}"

        # Try to create hook with negative deadline
        create_cmd = f"create hook {hook_name} --image nginx:latest --deadline -100"
        result = test_namespace.run_mtv_command(create_cmd, check=False)
        assert result.returncode != 0

    def test_create_hook_nonexistent_playbook_file(self, test_namespace):
        """Test that creating a hook with non-existent playbook file fails."""
        hook_name = f"test-hook-bad-file-{int(time.time())}"

        # Try to create hook with non-existent file
        create_cmd = f"create hook {hook_name} --image nginx:latest --playbook @/nonexistent/file.yaml"
        result = test_namespace.run_mtv_command(create_cmd, check=False)
        assert result.returncode != 0
        assert (
            "failed to read playbook file" in result.stderr
            or "no such file" in result.stderr.lower()
        )
