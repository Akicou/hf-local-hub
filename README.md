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

## Authentication

HF Local Hub supports three authentication methods:

### 1. Token Authentication (Simple)

The simplest method - use a shared secret token.

**Enable via CLI:**
```bash
./hf-local -token "your-secret-token" -auth-token
```

**Enable via Environment Variables:**
```bash
export HF_LOCAL_TOKEN="your-secret-token"
export HF_LOCAL_AUTH_TOKEN="true"
```

**Using with the UI:**
Click "Login" and enter your token when prompted.

**Using with CLI/Python:**
```bash
export HF_ENDPOINT=http://localhost:8080
export HF_TOKEN="your-secret-token"
huggingface-cli download user/my-model
```

**How it works:**
- Client sends token to `/api/auth/login`
- Server validates the token against the configured secret
- Server returns a JWT token (valid for 24 hours)
- Client includes JWT in `Authorization: Bearer <token>` header for subsequent requests

---

### 2. Hugging Face OAuth

Authenticates users through Hugging Face's OAuth2 flow.

**Prerequisites:**
1. Create a Hugging Face OAuth application at https://huggingface.co/settings/applications
2. Set redirect URL to: `http://localhost:8080/auth/hf/callback`
3. Note your Client ID and Client Secret

**Enable via CLI:**
```bash
./hf-local \
  -auth-hf \
  -hf-client-id "your-client-id" \
  -hf-client-secret "your-client-secret" \
  -hf-callback-url "http://localhost:8080/auth/hf/callback"
```

**Enable via Environment Variables:**
```bash
export HF_LOCAL_AUTH_HF="true"
export HF_LOCAL_HF_CLIENT_ID="your-client-id"
export HF_LOCAL_HF_CLIENT_SECRET="your-client-secret"
export HF_LOCAL_HF_CALLBACK_URL="http://localhost:8080/auth/hf/callback"
```

**Using with the UI:**
The UI's "Login" button will redirect to Hugging Face's authorization page.

**How it works:**
1. User clicks "Login" → redirect to `https://huggingface.co/oauth/authorize`
2. User approves the application on Hugging Face
3. Hugging Face redirects back with an authorization code
4. Server exchanges code for an access token
5. Server fetches user info and generates a JWT
6. Client uses JWT for authenticated requests

**Scopes requested:**
- `openid` - OpenID Connect authentication
- `profile` - User's display name and profile info
- `email` - User's email address

---

### 3. LDAP Authentication

Authenticate against your corporate LDAP/Active Directory server.

**Enable via CLI:**
```bash
./hf-local \
  -auth-ldap \
  -ldap-server "ldap.company.com" \
  -ldap-port 389 \
  -ldap-bind-dn "cn=admin,dc=company,dc=com" \
  -ldap-bind-pass "admin-password" \
  -ldap-base-dn "ou=users,dc=company,dc=com" \
  -ldap-filter "(uid=%s)"
```

**Enable via Environment Variables:**
```bash
export HF_LOCAL_AUTH_LDAP="true"
export HF_LOCAL_LDAP_SERVER="ldap.company.com"
export HF_LOCAL_LDAP_PORT="389"
export HF_LOCAL_LDAP_BIND_DN="cn=admin,dc=company,dc=com"
export HF_LOCAL_LDAP_BIND_PASS="admin-password"
export HF_LOCAL_LDAP_BASE_DN="ou=users,dc=company,dc=com"
export HF_LOCAL_LDAP_FILTER="(uid=%s)"
```

**Using with the UI:**
Currently, LDAP login requires the token endpoint. Set up a simple token auth proxy or use CLI tools.

**LDAP Configuration Options:**
- `ldap-server`: LDAP server hostname or IP
- `ldap-port`: LDAP port (389 for unencrypted, 636 for LDAPS)
- `ldap-bind-dn`: DN of the service account for binding/searching
- `ldap-bind-pass`: Password for the service account
- `ldap-base-dn`: Base DN for user searches (e.g., `ou=users,dc=company,dc=com`)
- `ldap-filter`: Search filter to find users. `%s` is replaced with username
  - For uid-based: `(uid=%s)`
  - For email-based: `(mail=%s)`
  - For sAMAccountName (AD): `(sAMAccountName=%s)`

**How it works:**
1. Server binds to LDAP using service account credentials
2. Server searches for user using the configured filter
3. Server attempts to bind using user DN and provided password
4. If successful, server generates JWT token with user attributes
5. User can retrieve email, display name, etc. from LDAP attributes

**Security Notes:**
- Use LDAPS (port 636) or StartTLS for production
- Service account should have read-only access to user directory
- Store bind credentials securely (use environment variables, not CLI args)

---

### JWT Token Details

All authentication methods issue JWT tokens with the following structure:

```json
{
  "user_id": "username-or-uid",
  "username": "Display Name",
  "provider": "token|hf|ldap",
  "exp": 1738362000,
  "iat": 1738275600
}
```

- **Expiration**: 24 hours after issue
- **Signing**: HMAC-SHA256 with configured JWT secret
- **Storage**: Client stores token (localStorage for web, env var for CLI)

---

### Protecting API Endpoints

The server supports two middleware levels:

**Optional Auth** (default):
- Users can access public repositories without authentication
- Authenticated users see their private repositories
- Used for: `GET /api/models`, `GET /api/datasets`, `GET /api/repos/:repo_id`

**Required Auth** (future):
- Must be authenticated to access
- Used for: Upload, delete, modify operations (to be added)

---

### Environment Variables Reference

You can also use a `.env` file in the project root or server directory. The server will automatically load it.

| Variable | Description | Example |
|----------|-------------|---------|
| `HF_LOCAL_TOKEN` | Shared secret for token auth | `my-secret-key` |
| `HF_LOCAL_AUTH_TOKEN` | Enable token authentication | `true` |
| `HF_LOCAL_AUTH_HF` | Enable HF OAuth | `true` |
| `HF_LOCAL_HF_CLIENT_ID` | HF OAuth client ID | `abc123` |
| `HF_LOCAL_HF_CLIENT_SECRET` | HF OAuth client secret | `xyz789` |
| `HF_LOCAL_HF_CALLBACK_URL` | HF OAuth callback URL | `http://localhost:8080/auth/hf/callback` |
| `HF_LOCAL_AUTH_LDAP` | Enable LDAP authentication | `true` |
| `HF_LOCAL_LDAP_SERVER` | LDAP server address | `ldap.company.com` |
| `HF_LOCAL_LDAP_PORT` | LDAP port | `389` |
| `HF_LOCAL_LDAP_BIND_DN` | LDAP bind DN | `cn=admin,dc=company,dc=com` |
| `HF_LOCAL_LDAP_BIND_PASS` | LDAP bind password | `admin-password` |
| `HF_LOCAL_LDAP_BASE_DN` | LDAP base DN | `ou=users,dc=company,dc=com` |
| `HF_LOCAL_LDAP_FILTER` | LDAP user search filter | `(uid=%s)` |
| `HF_LOCAL_JWT_SECRET` | JWT signing secret (defaults to TOKEN) | `change-me-in-production` |

---

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

### Install via pip

```bash
pip install hf-local-hub
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
- [x] Phase 4: Packaging & Documentation
- [x] Phase 5: CI/CD & GitHub Readiness

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
- [x] Phase 5: CI/CD & GitHub Readiness

## Current Release

**Version 0.2.0** - March 1, 2026

### Installation

```bash
# From PyPI
pip install hf-local-hub

# Or install from source
git clone https://github.com/Akicou/hf-local-hub.git
cd hf-local-hub
make server
cd python && pip install -e .
```

### Quick Start

```bash
# Start server
hf-local serve --port 8080

# Set HF_ENDPOINT
export HF_ENDPOINT=http://localhost:8080

# Use with huggingface_hub
python -c "
from huggingface_hub import snapshot_download
snapshot_download('user/my-model')
"
```

### What's Included

- ✅ Go server with REST API (single binary)
- ✅ Python CLI and library
- ✅ Full Hugging Face Hub compatibility
- ✅ Upload/download workflows
- ✅ Transformers and Diffusers integration
- ✅ Docker support
- ✅ Multiple auth methods (Token, OAuth, LDAP)
- ✅ Comprehensive tests
- ✅ Complete documentation

### Documentation

- [Usage Guide](docs/USAGE.md) - Complete usage instructions
- [API Reference](docs/API.md) - REST API and Python client
- [Examples](docs/examples/) - 4 working code examples
- [Contributing](CONTRIBUTING.md) - Development guidelines
- [Security](SECURITY.md) - Security policy

### Source & Releases

- **GitHub**: https://github.com/Akicou/hf-local-hub
- **PyPI**: https://pypi.org/project/hf-local-hub/
- **Docker Hub**: Coming soon

### Support

- **Issues**: https://github.com/Akicou/hf-local-hub/issues
- **Discussions**: https://github.com/Akicou/hf-local-hub/discussions

## Acknowledgments

Built with:
- [Gin](https://gin-gonic.com/) - Go web framework
- [GORM](https://gorm.io/) - ORM for Go
- [huggingface_hub](https://github.com/huggingface/huggingface_hub) - Python client
- [Typer](https://typer.tiangolo.com/) - CLI framework

## Roadmap

### v0.2.0
- [x] Authentication and authorization
- [ ] User management
- [ ] Repository access control
- [ ] Git operations (branches, tags)
- [ ] Model metadata search

### v0.3.0
- [x] Web UI for repository management
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
