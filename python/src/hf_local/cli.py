"""CLI for hf-local server and client operations."""

import os
import shutil
import subprocess
from pathlib import Path
from typing import Optional

import httpx
import typer
from rich.console import Console
from rich.table import Table

app = typer.Typer(help="hf-local: Lightweight local Hugging Face Hub server")
console = Console()


def find_binary() -> str:
    """Find the hf-local Go binary."""
    binary_name = "hf-local.exe" if os.name == "nt" else "hf-local"

    # Search paths - prioritize Go binary locations over Python CLI
    paths = [
        # Package root (for pip install -e .)
        Path(__file__).parent.parent.parent.parent / "server" / binary_name,
        # Parent of src (development layout)
        Path(__file__).parent.parent.parent / "server" / binary_name,
        # Same directory as package (wheel layout)
        Path(__file__).parent.parent / binary_name,
        # Current working directory
        Path.cwd() / binary_name,
        Path.cwd() / "server" / binary_name,
    ]

    for path in paths:
        if path.exists():
            return str(path.resolve())

    # Check PATH but verify it's not the Python CLI by checking help output
    path_binary = shutil.which(binary_name)
    if path_binary:
        # Verify it's the Go binary by checking for Go-specific flags
        try:
            result = subprocess.run(
                [path_binary, "-help"],
                capture_output=True,
                text=True,
                timeout=2,
            )
            # Go binary uses flags like -port, -data-dir, -log-level
            # Python CLI uses commands like serve, init, list
            help_output = result.stdout + result.stderr
            if "-port" in help_output or "-data-dir" in help_output:
                return path_binary
        except Exception:
            pass

    raise FileNotFoundError(
        "hf-local Go binary not found. Searched in:\n"
        + "\n".join(f"  - {p}" for p in paths)
        + "\nBuild with 'make server' or add Go binary to PATH.",
    )


def is_server_running(endpoint: str = "http://localhost:8080") -> bool:
    """Check if server is running."""
    try:
        response = httpx.get(f"{endpoint}/health", timeout=1.0)
        return response.status_code == 200
    except Exception:
        return False


@app.command()
def serve(
    port: int = typer.Option(8080, "--port", "-p", help="Server port"),
    data_dir: str = typer.Option(
        "./data",
        "--data-dir",
        "-d",
        help="Data storage directory",
    ),
    log_level: str = typer.Option(
        "info",
        "--log-level",
        "-l",
        help="Log level (debug, info, warn, error)",
    ),
    token: Optional[str] = typer.Option(
        None,
        "--token",
        help="Authentication token",
    ),
    auth_token: bool = typer.Option(
        False,
        "--auth-token",
        help="Enable token authentication",
    ),
    auth_hf: bool = typer.Option(
        False,
        "--auth-hf",
        help="Enable HuggingFace OAuth",
    ),
    hf_client_id: Optional[str] = typer.Option(
        None,
        "--hf-client-id",
        help="HF OAuth client ID",
    ),
    hf_client_secret: Optional[str] = typer.Option(
        None,
        "--hf-client-secret",
        help="HF OAuth client secret",
    ),
    auth_ldap: bool = typer.Option(
        False,
        "--auth-ldap",
        help="Enable LDAP authentication",
    ),
    ldap_server: Optional[str] = typer.Option(
        None,
        "--ldap-server",
        help="LDAP server address",
    ),
):
    """Start the hf-local server."""
    endpoint = f"http://localhost:{port}"

    if is_server_running(endpoint):
        console.print(f"[yellow]Server already running at {endpoint}[/yellow]")
        return

    binary = find_binary()
    console.print(f"[green]Starting hf-local server at {endpoint}[/green]")
    console.print(f"[dim]Binary: {binary}[/dim]")
    console.print(f"[dim]Data directory: {data_dir}[/dim]")

    cmd = [
        binary,
        "-port", str(port),
        "-data-dir", data_dir,
        "-log-level", log_level,
    ]

    if token:
        cmd.extend(["-token", token])
    if auth_token:
        cmd.append("-auth-token")
    if auth_hf:
        cmd.append("-auth-hf")
    if hf_client_id:
        cmd.extend(["-hf-client-id", hf_client_id])
    if hf_client_secret:
        cmd.extend(["-hf-client-secret", hf_client_secret])
    if auth_ldap:
        cmd.append("-auth-ldap")
    if ldap_server:
        cmd.extend(["-ldap-server", ldap_server])

    try:
        process = subprocess.Popen(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
        )

        for line in process.stdout:  # type: ignore
            console.print(line.strip())

    except KeyboardInterrupt:
        console.print("\n[yellow]Shutting down server...[/yellow]")
        process.terminate()
        console.print("[green]Server stopped[/green]")


@app.command()
def upload(
    local_path: str = typer.Argument(..., help="Local file or directory path"),
    repo_id: str = typer.Argument(..., help="Repository ID (e.g., 'user/model')"),
    endpoint: str = typer.Option(
        "http://localhost:8080",
        "--endpoint",
        "-e",
        help="Server endpoint",
    ),
):
    """Upload files to a local repository."""
    from huggingface_hub import HfApi

    os.environ["HF_ENDPOINT"] = endpoint
    api = HfApi()

    path = Path(local_path)
    if not path.exists():
        console.print(f"[red]Error: Path {local_path} does not exist[/red]")
        raise typer.Exit(1)

    console.print(f"[green]Uploading {local_path} to {repo_id}[/green]")

    if path.is_dir():
        try:
            api.upload_folder(
                folder_path=str(path),
                repo_id=repo_id,
                repo_type="model",
            )
        except Exception as e:
            console.print(f"[red]Upload failed: {e}[/red]")
            raise typer.Exit(1)
    else:
        try:
            api.upload_file(
                path_or_fileobj=str(path),
                path_in_repo=path.name,
                repo_id=repo_id,
                repo_type="model",
            )
        except Exception as e:
            console.print(f"[red]Upload failed: {e}[/red]")
            raise typer.Exit(1)

    console.print(f"[green]Successfully uploaded to {repo_id}[/green]")


@app.command()
def list_repos(
    endpoint: str = typer.Option(
        "http://localhost:8080",
        "--endpoint",
        "-e",
        help="Server endpoint",
    ),
):
    """List repositories."""
    try:
        response = httpx.get(f"{endpoint}/api/models", timeout=5.0)
        response.raise_for_status()
        repos = response.json()

        if not repos:
            console.print("[dim]No repositories found[/dim]")
            return

        table = Table(title="Repositories")
        table.add_column("ID", style="cyan")
        table.add_column("Namespace", style="green")
        table.add_column("Name", style="blue")
        table.add_column("Type", style="yellow")
        table.add_column("Private", justify="center")

        for repo in repos:
            table.add_row(
                repo.get("repo_id", "-"),
                repo.get("namespace", "-"),
                repo.get("name", "-"),
                repo.get("type", "-"),
                "Yes" if repo.get("private") else "No",
            )

        console.print(table)

    except httpx.RequestError as e:
        console.print(f"[red]Failed to connect to server: {e}[/red]")
        raise typer.Exit(1)


@app.command()
def init(data_dir: str = typer.Option("./data", "--data-dir", "-d", help="Data directory path")):
    """Initialize a new hf-local instance."""
    path = Path(data_dir)
    console.print(f"[green]Initializing hf-local in {path.absolute()}[/green]")

    dirs = [
        path / "storage" / "models",
        path / "storage" / "datasets",
    ]

    for dir_path in dirs:
        dir_path.mkdir(parents=True, exist_ok=True)
        console.print(f"[dim]Created: {dir_path}[/dim]")

    console.print("[green][OK] hf-local initialized successfully[/green]")
    console.print(f"[dim]Start server with: hf-local serve --data-dir {data_dir}[/dim]")


@app.command()
def status(
    endpoint: str = typer.Option(
        "http://localhost:8080",
        "--endpoint",
        "-e",
        help="Server endpoint",
    ),
):
    """Check server status."""
    console.print(f"Checking server at {endpoint}...")

    if is_server_running(endpoint):
        try:
            response = httpx.get(f"{endpoint}/api/models", timeout=5.0)
            response.raise_for_status()
            repos = response.json()
            console.print("[green]✓ Server is running[/green]")
            console.print(f"[dim]Repositories: {len(repos)}[/dim]")
        except Exception as e:
            console.print("[green]✓ Server is running[/green]")
            console.print(f"[yellow]Could not fetch repository list: {e}[/yellow]")
    else:
        console.print("[red]✗ Server is not running[/red]")
        console.print("[dim]Start with: hf-local serve[/dim]")
        raise typer.Exit(1)


@app.command()
def login(
    token: str = typer.Option(..., "--token", "-t", help="Authentication token"),
    endpoint: str = typer.Option(
        "http://localhost:8080",
        "--endpoint",
        "-e",
        help="Server endpoint",
    ),
):
    """Login with authentication token."""
    import hf_local

    if hf_local.login(token, endpoint):
        console.print("[green]✓ Logged in successfully[/green]")
        console.print("[dim]Token set in HF_TOKEN environment variable[/dim]")
    else:
        console.print("[red]✗ Login failed[/red]")
        raise typer.Exit(1)


@app.command()
def logout():
    """Logout and clear stored credentials."""
    import hf_local

    hf_local.logout()
    console.print("[green]✓ Logged out[/green]")


if __name__ == "__main__":
    app()
