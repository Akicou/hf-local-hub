"""CLI for hf-local server and client operations."""

import typer

app = typer.Typer(help="hf-local: Lightweight local Hugging Face Hub server")


@app.command()
def serve(
    port: int = typer.Option(8080, "--port", "-p", help="Server port"),
    data_dir: str = typer.Option("./data", "--data-dir", "-d", help="Data storage directory"),
):
    """Start the hf-local server."""
    typer.echo(f"Starting hf-local server on port {port}...")
    typer.echo(f"Data directory: {data_dir}")
    # TODO: Launch Go binary
    raise NotImplementedError("Server launch not yet implemented")


@app.command()
def upload(
    local_path: str = typer.Argument(..., help="Local file or directory path"),
    repo_id: str = typer.Argument(..., help="Repository ID (e.g., 'user/model')"),
):
    """Upload files to a local repository."""
    typer.echo(f"Uploading {local_path} to {repo_id}...")
    # TODO: Implement upload logic
    raise NotImplementedError("Upload not yet implemented")


@app.command()
def list_repos(repo_id: str = typer.Argument(None, help="Filter by repository ID")):
    """List repositories."""
    typer.echo("Listing repositories...")
    # TODO: Implement list logic
    raise NotImplementedError("List not yet implemented")


@app.command()
def init(data_dir: str = typer.Option("./data", "--data-dir", "-d", help="Data directory path")):
    """Initialize a new hf-local instance."""
    typer.echo(f"Initializing hf-local in {data_dir}...")
    # TODO: Create directory structure
    raise NotImplementedError("Init not yet implemented")


@app.command()
def status():
    """Check server status."""
    typer.echo("Checking server status...")
    # TODO: Check if server is running
    raise NotImplementedError("Status check not yet implemented")


if __name__ == "__main__":
    app()
