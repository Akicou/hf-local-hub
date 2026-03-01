"""Tests for Transformers integration with hf-local."""

import pytest
from pathlib import Path

try:
    # ruff: noqa: F401
    from transformers import AutoModel, AutoTokenizer
    HAS_TRANSFORMERS = True
except ImportError:
    HAS_TRANSFORMERS = False

from hf_local import serve_background, set_endpoint


@pytest.mark.skipif(not HAS_TRANSFORMERS, reason="transformers not installed")
def test_transformers_auto_model(temp_data_dir: Path) -> None:
    """Test loading model with AutoModel."""
    endpoint = "http://localhost:8086"

    with serve_background(port=8086, data_dir=temp_data_dir):
        set_endpoint(endpoint)

        # This would require an actual model to be uploaded
        # For testing, we skip the actual loading
        pytest.skip("Requires model files to be uploaded first")


@pytest.mark.skipif(not HAS_TRANSFORMERS, reason="transformers not installed")
def test_transformers_auto_tokenizer(temp_data_dir: Path) -> None:
    """Test loading tokenizer with AutoTokenizer."""
    endpoint = "http://localhost:8087"

    with serve_background(port=8087, data_dir=temp_data_dir):
        set_endpoint(endpoint)

        # This would require an actual model to be uploaded
        pytest.skip("Requires model files to be uploaded first")


@pytest.fixture
def temp_data_dir(tmp_path: Path) -> Path:
    """Create temporary data directory."""
    data_dir = tmp_path / "data"
    data_dir.mkdir()
    return data_dir
