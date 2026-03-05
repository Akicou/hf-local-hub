"""hf-local: Lightweight local Hugging Face Hub server and client."""

import os
import subprocess
from contextlib import contextmanager
from pathlib import Path
from typing import Optional

from .cli import app as cli
from .cli import find_binary

__version__ = "0.2.0"
__all__ = ["cli", "serve_background", "set_endpoint", "upload_folder", "login", "logout"]


def set_endpoint(endpoint: str = "http://localhost:8080") -> None:
    """Set HF_ENDPOINT environment variable for huggingface_hub integration."""
    os.environ["HF_ENDPOINT"] = endpoint


def login(token: str, endpoint: str = "http://localhost:8080") -> bool:
    """Login with authentication token.

    Args:
        token: Authentication token
        endpoint: Server endpoint URL

    Returns:
        True if login successful

    """
    import httpx

    try:
        response = httpx.post(
            f"{endpoint}/api/auth/login",
            json={"token": token},
            timeout=5.0,
        )
        response.raise_for_status()
        data = response.json()
        # Store token in environment for huggingface_hub
        os.environ["HF_TOKEN"] = data.get("token", token)
        return True
    except Exception:
        return False


def logout() -> None:
    """Logout by clearing stored credentials."""
    os.environ.pop("HF_TOKEN", None)


@contextmanager
def serve_background(
    port: int = 8080,
    data_dir: str = "./data",
    log_level: str = "info",
    timeout: int = 5,
):
    """Context manager to run server in background for testing.

    Usage:
        with serve_background():
            # Your test code here
            pass
    """
    binary = find_binary()
    endpoint = f"http://localhost:{port}"

    cmd = [
        binary,
        "-port", str(port),
        "-data-dir", str(data_dir),
        "-log-level", log_level,
    ]

    # Capture output for debugging
    process = subprocess.Popen(
        cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True
    )

    try:
        import time

        import httpx

        for _ in range(timeout * 2):
            try:
                response = httpx.get(f"{endpoint}/health", timeout=0.5)
                if response.status_code == 200:
                    break
            except:
                time.sleep(0.5)
        else:
            # Log any output from server for debugging
            output = process.stdout.read()  # type: ignore
            raise TimeoutError(
                f"Server did not start within {timeout} seconds\n"
                f"Server output: {output}"
            )

        set_endpoint(endpoint)
        yield

    finally:
        process.terminate()
        process.wait(timeout=5)


def upload_folder(
    folder_path: str,
    repo_id: str,
    endpoint: str = "http://localhost:8080",
) -> None:
    """Upload a folder to a repository.

    Args:
        folder_path: Local folder path to upload
        repo_id: Repository ID (e.g., 'user/model')
        endpoint: Server endpoint URL

    """
    from huggingface_hub import HfApi

    os.environ["HF_ENDPOINT"] = endpoint
    api = HfApi()

    api.upload_folder(
        folder_path=folder_path,
        repo_id=repo_id,
        repo_type="model",
    )


class HfLocalApi:
    """Thin wrapper around HfApi that forces local endpoint."""

    def __init__(self, endpoint: str = "http://localhost:8080"):
        """Initialize API client.

        Args:
            endpoint: Local server endpoint

        """
        from huggingface_hub import HfApi

        os.environ["HF_ENDPOINT"] = endpoint
        self._api = HfApi()

    def __getattr__(self, name):
        """Delegate all calls to HfApi."""
        return getattr(self._api, name)

