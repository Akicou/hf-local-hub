# hf-local-hub

Lightweight local Hugging Face Hub server and client - run HF Hub entirely on your machine.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Python](https://img.shields.io/badge/Python-3.11+-3776AB?style=flat&logo=python)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)
![Docker](https://img.shields.io/badge/docker-ready-blue.svg)

## Features

- **Single Binary**: Static Go binary, no dependencies
- **Full API Compatibility**: Emulates essential Hugging Face Hub API
- **Zero Configuration**: Works out of the box
- **Python Integration**: Seamless with `huggingface_hub` via `HF_ENDPOINT`
- **Local Storage**: All models and datasets stored on your filesystem

## Quick Start

### Using Go Binary

```bash
# Build
make server

# Run server
./hf-local

# Use with huggingface_hub
HF_ENDPOINT=http://localhost:8080 huggingface-cli download user/my-model
```

### Using Python CLI

```bash
# Install
pip install -e .

# Start server
hf-local serve --port 8080

# Upload model
hf-local upload ./my-model user/my-model

# List repositories
hf-local list
```

### Using Transformers

```python
from transformers import AutoModel

# Point to local server
import os
os.environ["HF_ENDPOINT"] = "http://localhost:8080"

model = AutoModel.from_pretrained("user/my-model")
```

## Architecture

```
┌─────────────────┐
│   Client Apps   │  (Transformers, Diffusers, etc.)
└────────┬────────┘
         │ HF_ENDPOINT=http://localhost:8080
         ▼
┌─────────────────┐
│  Go Server      │  (Gin + SQLite)
│  - API Layer    │
│  - File Serving │
│  - Auth (opt)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Local FS       │  (data/storage/models/...)
└─────────────────┘
```

## Development

```bash
# Install dependencies
make dev

# Run tests
make test

# Lint
make lint

# Docker
make docker-build
make docker-run
```

## Project Status

- [x] Phase 0: Project Initialization
- [x] Phase 1: Go Server Core
- [ ] Phase 2: Python Package & CLI
- [ ] Phase 3: Full HF Compatibility
- [ ] Phase 4: Packaging & Documentation
- [ ] Phase 5: CI/CD & GitHub Readiness

## License

MIT License - see [LICENSE](LICENSE) file

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Acknowledgments

Built with:
- [Gin](https://gin-gonic.com/) - Go web framework
- [GORM](https://gorm.io/) - ORM for Go
- [huggingface_hub](https://github.com/huggingface/huggingface_hub) - Python client
- [Typer](https://typer.tiangolo.com/) - CLI framework
