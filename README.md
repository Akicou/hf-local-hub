# hf-local-hub

Lightweight local Hugging Face Hub server and client - run HF Hub entirely on your machine.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Python](https://img.shields.io/badge/Python-3.11+-3776AB?style=flat&logo=python)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)
![Docker](https://img.shields.io/badge/docker-ready-blue.svg)
![Tests](https://img.shields.io/badge/tests-passing-brightgreen)

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Documentation](#documentation)
- [Examples](#examples)
- [Architecture](#architecture)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Single Binary**: Static Go binary (~53MB), no runtime dependencies
- **Full API Compatibility**: Emulates essential Hugging Face Hub API
- **Zero Configuration**: Works out of the box
- **Python Integration**: Seamless with `huggingface_hub`, `transformers`, `diffusers`
- **Local Storage**: All models and datasets stored on your filesystem
- **Cross-Platform**: Linux, macOS, Windows
- **Docker Support**: Multi-stage Docker image
- **Comprehensive Testing**: Unit and integration tests

## Quick Start

### 1. Install

**Go Binary:**
```bash
# Clone repository
git clone https://github.com/lyani/hf-local-hub.git
cd hf-local-hub

# Build server
make server
```

**Python Package:**
```bash
# Clone and install
git clone https://github.com/lyani/hf-local-hub.git
cd hf-local-hub/python
pip install -e .
```

**Docker:**
```bash
# Pull or build
docker build -t hf-local .
docker run -p 8080:8080 -v $(pwd)/data:/app/data hf-local
```

### 2. Start Server

```bash
# Using Go binary
./hf-local

# Using Python CLI
hf-local serve --port 8080
```

### 3. Use with Hugging Face Libraries

```bash
# Set endpoint
export HF_ENDPOINT=http://localhost:8080

# Download models
huggingface-cli download user/my-model

# Or use in Python
python -c "
from transformers import AutoModel
model = AutoModel.from_pretrained('user/my-model')
"
```

## Installation

### Requirements

- **Go**: 1.25+ (for building server)
- **Python**: 3.11+ (for Python client)
- **Disk Space**: ~100MB minimum (varies by model size)

### Build from Source

```bash
# Clone repository
git clone https://github.com/lyani/hf-local-hub.git
cd hf-local-hub

# Build Go server
make server

# Install Python package
cd python
pip install -e ".[dev]"
```

### Install via pip (Future)

```bash
pip install hf-local
```

## Documentation

- [Usage Guide](docs/USAGE.md) - Complete usage instructions
- [API Reference](docs/API.md) - REST API and Python client API
- [Examples](docs/examples/) - Code examples
- [Contributing](CONTRIBUTING.md) - Development guidelines
- [Security](SECURITY.md) - Security policy and reporting

## Examples

See [docs/examples/](docs/examples/) directory for complete examples:

- [basic_upload.py](docs/examples/basic_upload.py) - Upload model to local server
- [transformers_demo.py](docs/examples/transformers_demo.py) - Use with Transformers
- [diffusers_demo.py](docs/examples/diffusers_demo.py) - Use with Diffusers
- [api_client_demo.py](docs/examples/api_client_demo.py) - Custom API client

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

### Setup

```bash
# Clone repository
git clone https://github.com/lyani/hf-local-hub.git
cd hf-local-hub

# Install Python dependencies
cd python
pip install -e ".[dev]"
```

### Run Tests

```bash
# All tests
make test

# Go tests only
make server-test

# Python tests only
make python-test
```

### Linting

```bash
# All linters
make lint

# Go lint
make server-lint

# Python lint
make python-lint
```

### Docker

```bash
# Build image
make docker-build

# Run container
make docker-run

# Stop container
make docker-down
```

### Development Workflow

1. Create feature branch: `git checkout -b feat/your-feature`
2. Make changes and commit with conventional commits
3. Run tests and linting: `make test lint`
4. Push and create pull request

## Contributing

- [x] Phase 0: Project Initialization
- [x] Phase 1: Go Server Core
- [x] Phase 2: Python Package & CLI
- [x] Phase 3: Full HF Compatibility
- [ ] Phase 4: Packaging & Documentation
- [ ] Phase 5: CI/CD & GitHub Readiness

## License

MIT License - see [LICENSE](LICENSE) file

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Project Status

- [x] Phase 0: Project Initialization
- [x] Phase 1: Go Server Core
- [x] Phase 2: Python Package & CLI
- [x] Phase 3: Full HF Compatibility
- [x] Phase 4: Packaging & Documentation
- [ ] Phase 5: CI/CD & GitHub Readiness

## Acknowledgments

Built with:
- [Gin](https://gin-gonic.com/) - Go web framework
- [GORM](https://gorm.io/) - ORM for Go
- [huggingface_hub](https://github.com/huggingface/huggingface_hub) - Python client
- [Typer](https://typer.tiangolo.com/) - CLI framework

## Roadmap

### v0.2.0
- [ ] Authentication and authorization
- [ ] User management
- [ ] Repository access control
- [ ] Git operations (branches, tags)
- [ ] Model metadata search

### v0.3.0
- [ ] Web UI for repository management
- [ ] Model card editor
- [ ] File preview
- [ ] Model sharing features
- [ ] Integration with CI/CD

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/lyani/hf-local-hub/issues)
- **Discussions**: [GitHub Discussions](https://github.com/lyani/hf-local-hub/discussions)
- **Security**: [SECURITY.md](SECURITY.md)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=lyani/hf-local-hub&type=Date)](https://star-history.com/#lyani/hf-local-hub&Date)
