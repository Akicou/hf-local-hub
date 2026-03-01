"""Tests for hf-local CLI."""

import os

from hf_local import set_endpoint


def test_set_endpoint():
    """Test setting HF_ENDPOINT."""
    set_endpoint("http://localhost:9090")
    assert os.environ.get("HF_ENDPOINT") == "http://localhost:9090"


def test_is_server_running():
    """Test server status check."""
    from hf_local.cli import is_server_running

    # Test with non-existent server
    assert is_server_running("http://localhost:9999") is False


def test_init_command(tmp_path):
    """Test init command."""
    import subprocess

    result = subprocess.run(
        ["hf-local", "init", f"--data-dir={tmp_path}"],
        capture_output=True,
        text=True,
    )
    assert result.returncode == 0

    assert (tmp_path / "storage" / "models").exists()
    assert (tmp_path / "storage" / "datasets").exists()


def test_status_command_not_running():
    """Test status command when server not running."""
    import subprocess

    result = subprocess.run(
        ["hf-local", "status", "--endpoint=http://localhost:9999"],
        capture_output=True,
        text=True,
    )
    assert result.returncode == 1
    assert "not running" in result.stdout or "not running" in result.stderr


def test_list_command_not_running():
    """Test list command when server not running."""
    import subprocess

    result = subprocess.run(
        ["hf-local", "list-repos", "--endpoint=http://localhost:9999"],
        capture_output=True,
        text=True,
    )
    assert result.returncode == 1
    assert "Failed to connect" in result.stdout or "Failed to connect" in result.stderr

