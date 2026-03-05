# hf-local-hub

Lightweight local Hugging Face Hub server and client - run HF Hub entirely on your machine.

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![Python](https://img.shields.io/badge/Python-3.11+-3776AB?style=flat&logo=python)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)
![Docker](https://img.shields.io/badge/docker-ready-blue.svg)
![Tests](https://img.shields.io/badge/tests-passing-brightgreen)

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Database Configuration](#database-configuration)
- [Authentication](#authentication)
- [API Tokens](#api-tokens)
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
- **PostgreSQL Support**: Use SQLite (default) or PostgreSQL for production
- **API Token Management**: Create and manage API tokens with fine-grained permissions
- **Comprehensive Testing**: Unit and integration tests

## Quick Start

### 1. Install

**Go Binary:**
```bash
# Clone repository
git clone https://github.com/Akicou/hf-local-hub.git
cd hf-local-hub

# Build server
make server
```

**Python Package:**
```bash
# Clone and install
git clone https://github.com/Akicou/hf-local-hub.git
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

## Database Configuration

HF Local Hub supports both SQLite (default) and PostgreSQL.

### SQLite (Default)

No configuration needed - SQLite database is automatically created at `{data_dir}/hf-local.db`.

### PostgreSQL

For production deployments, PostgreSQL is recommended:

**Via CLI:**
```bash
./hf-local \
  -db-type postgres \
  -db-host localhost \
  -db-port 5432 \
  -db-user postgres \
  -db-password your-password \
  -db-name hf_local_hub
```

**Via Environment Variables:**
```bash
export HF_LOCAL_DB_TYPE=postgres
export HF_LOCAL_DB_HOST=localhost
export HF_LOCAL_DB_PORT=5432
export HF_LOCAL_DB_USER=postgres
export HF_LOCAL_DB_PASSWORD=your-password
export HF_LOCAL_DB_NAME=hf_local_hub
```

**Via Docker Compose:**
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: hf_local_hub
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: your-password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  hf-local:
    build: .
    ports:
      - "8080:8080"
    environment:
      HF_LOCAL_DB_TYPE: postgres
      HF_LOCAL_DB_HOST: postgres
      HF_LOCAL_DB_PORT: 5432
      HF_LOCAL_DB_USER: postgres
      HF_LOCAL_DB_PASSWORD: your-password
      HF_LOCAL_DB_NAME: hf_local_hub
    depends_on:
      - postgres

volumes:
  postgres_data:
```

## Authentication

HF Local Hub requires authentication for all API operations. Users must log in before accessing the API.

### 1. Hugging Face OAuth

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

### 2. LDAP Authentication

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

---

## API Tokens

Users can create API tokens with specific permissions, similar to Hugging Face's token system. These tokens can be used for programmatic access to the API.

### Creating API Tokens

**Via API:**
```bash
# First, authenticate with JWT token
curl -X POST http://localhost:8080/api/tokens/ \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My API Token",
    "read": true,
    "write": true,
    "delete": false,
    "admin": false,
    "expires_in": 720
  }'
```

**Response:**
```json
{
  "id": 1,
  "token": "hf_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "name": "My API Token",
  "permissions": {
    "read": true,
    "write": true,
    "delete": false,
    "admin": false
  },
  "expires_at": "2026-03-05T12:00:00Z",
  "created_at": "2026-03-05T00:00:00Z"
}
```

### Token Permissions

| Permission | Description |
|------------|-------------|
| `read` | Can read repositories and files |
| `write` | Can create/update repositories and files |
| `delete` | Can delete repositories and files |
| `admin` | Can manage users and tokens |

### Using API Tokens

```bash
# Use token like Hugging Face token
export HF_TOKEN=hf_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
export HF_ENDPOINT=http://localhost:8080

# Use with huggingface_hub
huggingface-cli download user/my-model
```

### Managing Tokens

**List tokens:**
```bash
curl -X GET http://localhost:8080/api/tokens/ \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Delete token:**
```bash
curl -X DELETE http://localhost:8080/api/tokens/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

### JWT Token Details

All authentication methods issue JWT tokens with the following structure:

```json
{
  "user_id": "username-or-uid",
  "username": "Display Name",
  "provider": "hf|ldap",
  "exp": 1738362000,
  "iat": 1738275600
}
```

- **Expiration**: 24 hours after issue
- **Signing**: HMAC-SHA256 with configured JWT secret
- **Storage**: Client stores token (localStorage for web, env var for CLI)

---

### Environment Variables Reference

You can also use a `.env` file in the project root or server directory. The server will automatically load it.

| Variable | Description | Example |
|----------|-------------|---------|
| `HF_LOCAL_JWT_SECRET` | JWT signing secret (auto-generated if not set) | `change-me-in-production` |
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
| `HF_LOCAL_DB_TYPE` | Database type (sqlite or postgres) | `postgres` |
| `HF_LOCAL_DB_HOST` | PostgreSQL host | `localhost` |
| `HF_LOCAL_DB_PORT` | PostgreSQL port | `5432` |
| `HF_LOCAL_DB_USER` | PostgreSQL user | `postgres` |
| `HF_LOCAL_DB_PASSWORD` | PostgreSQL password | `your-password` |
| `HF_LOCAL_DB_NAME` | PostgreSQL database name | `hf_local_hub` |
| `HF_LOCAL_DB_SSLMODE` | PostgreSQL SSL mode | `disable` |

---

## Installation

### Requirements

- **Go**: 1.25+ (for building server)
- **Python**: 3.11+ (for Python client)
- **Disk Space**: ~100MB minimum (varies by model size)

### Build from Source

```bash
# Clone repository
git clone https://github.com/Akicou/hf-local-hub.git
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
│  Go Server      │  (Gin + GORM)
│  - API Layer    │
│  - File Serving │
│  - Auth (JWT)   │
│  - API Tokens   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Database       │  (SQLite or PostgreSQL)
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
git clone https://github.com/Akicou/hf-local-hub.git
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
- [x] Phase 6: PostgreSQL Support & API Tokens

## License

MIT License - see [LICENSE](LICENSE) file

## Current Release

**Version 0.2.0** - March 2026

### What's Included

- ✅ Go server with REST API (single binary)
- ✅ Python CLI and library
- ✅ Full Hugging Face Hub compatibility
- ✅ Upload/download workflows
- ✅ Transformers and Diffusers integration
- ✅ Docker support
- ✅ OAuth (HF) and LDAP authentication
- ✅ API token management with permissions
- ✅ PostgreSQL support
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

### v0.3.0
- [ ] User management UI
- [ ] Repository access control
- [ ] Git operations (branches, tags)
- [ ] Model metadata search

### v0.4.0
- [ ] Model card editor
- [ ] File preview
- [ ] Model sharing features
- [ ] Integration with CI/CD

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/Akicou/hf-local-hub/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Akicou/hf-local-hub/discussions)
- **Security**: [SECURITY.md](SECURITY.md)
