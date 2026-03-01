"""hf-local: Lightweight local Hugging Face Hub server and client."""

from .cli import app as cli

__version__ = "0.1.0"
__all__ = ["cli"]


def set_endpoint(endpoint: str = "http://localhost:8080") -> None:
    """Set HF_ENDPOINT environment variable for huggingface_hub integration."""
    import os

    os.environ["HF_ENDPOINT"] = endpoint
