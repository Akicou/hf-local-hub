# hf-local Python Package

Python client and CLI for the hf-local server.

## Installation

```bash
pip install -e .
```

## Usage

```bash
hf-local serve --port 8080
hf-local upload ./my-model user/my-model
hf-local list
```

## Development

```bash
pip install -e ".[dev]"
pytest
ruff check src/
mypy src/
```
