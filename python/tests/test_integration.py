"""Integration tests for hf-local with huggingface_hub."""

import os
import shutil
from pathlib import Path

import pytest
from huggingface_hub import HfApi, snapshot_download, hf_hub_download

from hf_local import serve_background, set_endpoint, upload_folder


@pytest.fixture
def temp_data_dir(tmp_path):
    """Create temporary data directory."""
    data_dir = tmp_path / "data"
    data_dir.mkdir()
    return data_dir


@pytest.fixture
def hf_api():
    """Create HfApi instance."""
    api = HfApi()
    return api


def test_create_repo(temp_data_dir, hf_api):
    """Test creating a repository."""
    endpoint = "http://localhost:8081"

    with serve_background(port=8081, data_dir=temp_data_dir):
        set_endpoint(endpoint)

        repo_id = "test-user/test-model"

        # Create repo via API
        repo = hf_api.create_repo(
            repo_id=repo_id,
            repo_type="model",
            private=False,
        )

        assert repo.repo_id == repo_id
        assert repo.private is False


def test_upload_file(temp_data_dir, hf_api):
    """Test uploading a single file."""
    endpoint = "http://localhost:8082"

    with serve_background(port=8082, data_dir=temp_data_dir):
        set_endpoint(endpoint)

        repo_id = "test-user/file-upload"

        # Create repo
        hf_api.create_repo(repo_id=repo_id, repo_type="model")

        # Create test file
        test_file = Path("test_upload_file.txt")
        test_file.write_text("Test content")

        try:
            # Upload file
            hf_api.upload_file(
                path_or_fileobj=str(test_file),
                path_in_repo="config.json",
                repo_id=repo_id,
                repo_type="model",
            )

            # Verify file exists
            downloaded = hf_hub_download(
                repo_id=repo_id,
                filename="config.json",
            )

            assert Path(downloaded).exists()
            assert Path(downloaded).read_text() == "Test content"

        finally:
            test_file.unlink(missing_ok=True)


def test_upload_folder(temp_data_dir, hf_api):
    """Test uploading a folder."""
    endpoint = "http://localhost:8083"

    with serve_background(port=8083, data_dir=temp_data_dir):
        set_endpoint(endpoint)

        repo_id = "test-user/folder-upload"

        # Create repo
        hf_api.create_repo(repo_id=repo_id, repo_type="model")

        # Create test folder structure
        test_folder = Path("test_model_folder")
        test_folder.mkdir(exist_ok=True)

        (test_folder / "config.json").write_text('{"model_type": "test"}')
        (test_folder / "model.safetensors").write_bytes(b"fake model weights")
        (test_folder / "README.md").write_text("# Test Model")

        try:
            # Upload folder
            upload_folder(
                folder_path=str(test_folder),
                repo_id=repo_id,
                endpoint=endpoint,
            )

            # Snapshot download
            snapshot_dir = snapshot_download(
                repo_id=repo_id,
            )

            snapshot_path = Path(snapshot_dir)
            assert (snapshot_path / "config.json").exists()
            assert (snapshot_path / "model.safetensors").exists()
            assert (snapshot_path / "README.md").exists()

            assert (snapshot_path / "config.json").read_text() == '{"model_type": "test"}'

        finally:
            shutil.rmtree(test_folder, ignore_errors=True)


def test_snapshot_download(temp_data_dir, hf_api):
    """Test snapshot_download."""
    endpoint = "http://localhost:8084"

    with serve_background(port=8084, data_dir=temp_data_dir):
        set_endpoint(endpoint)

        repo_id = "test-user/snapshot-test"

        # Create repo and upload files
        hf_api.create_repo(repo_id=repo_id, repo_type="model")

        test_file = Path("test_snapshot.json")
        test_file.write_text('{"snapshot": "test"}')

        try:
            hf_api.upload_file(
                path_or_fileobj=str(test_file),
                path_in_repo="config.json",
                repo_id=repo_id,
            )

            # Snapshot download
            snapshot_dir = snapshot_download(repo_id=repo_id)
            snapshot_path = Path(snapshot_dir)

            assert snapshot_path.exists()
            assert (snapshot_path / "config.json").exists()
            assert (snapshot_path / "config.json").read_text() == '{"snapshot": "test"}'

        finally:
            test_file.unlink(missing_ok=True)


def test_list_repositories(temp_data_dir, hf_api):
    """Test listing repositories."""
    endpoint = "http://localhost:8085"

    with serve_background(port=8085, data_dir=temp_data_dir):
        set_endpoint(endpoint)

        # Create multiple repos
        for i in range(3):
            hf_api.create_repo(
                repo_id=f"test-user/repo-{i}",
                repo_type="model",
            )

        # List repos
        repos = hf_api.list_models()
        repo_ids = [r.repo_id for r in repos]

        assert "test-user/repo-0" in repo_ids
        assert "test-user/repo-1" in repo_ids
        assert "test-user/repo-2" in repo_ids
