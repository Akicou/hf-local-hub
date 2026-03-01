# API Reference

Complete API reference for hf-local-hub server and Python client.

## Table of Contents

- [Server API](#server-api)
- [Python CLI](#python-cli)
- [Python Library](#python-library)

## Server API

Base URL: `http://localhost:8080`

### Health Check

Check if server is running.

```http
GET /health
```

**Response:**
```json
{
  "status": "ok"
}
```

### Auth Config

Get enabled authentication methods.

```http
GET /auth/config
```

**Response:**
```json
{
  "token": true,
  "hf": false,
  "ldap": false
}
```

### Token Login

Login with authentication token.

```http
POST /api/auth/login
Content-Type: application/json
```

**Request Body:**
```json
{
  "token": "your-secret-token"
}
```

**Response:**
```json
{
  "token": "<jwt-token>",
  "user": {
    "id": "token-user",
    "name": "Token User"
  }
}
```

### HF OAuth Login

Redirect to Hugging Face OAuth authorization.

```http
GET /api/auth/hf/login
```

**Response:**
Redirects to Hugging Face OAuth consent screen.

### HF OAuth Callback

Handle OAuth callback from Hugging Face.

```http
GET /api/auth/hf/callback?code=...&state=...
```

**Response:**
```json
{
  "token": "<jwt-token>",
  "user": {
    "sub": "hf-user",
    "name": "HF User",
    "email": "user@huggingface.co"
  }
}
```

### LDAP Login

Login with LDAP credentials.

```http
POST /api/auth/ldap/login
Content-Type: application/json
```

**Request Body:**
```json
{
  "username": "john.doe",
  "password": "secure-password"
}
```

**Response:**
```json
{
  "token": "<jwt-token>",
  "user": {
    "id": "john.doe",
    "name": "john.doe"
  }
}
```

### Create Repository

Create a new repository.

```http
POST /api/repos/create
Content-Type: application/json
```

**Request Body:**
```json
{
  "repo_id": "user/model-name",
  "namespace": "user",
  "name": "model-name",
  "type": "model",
  "private": false
}
```

**Response:**
```json
{
  "id": 1,
  "repo_id": "user/model-name",
  "namespace": "user",
  "name": "model-name",
  "type": "model",
  "private": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### List Models

List all model repositories.

```http
GET /api/models
```

**Response:**
```json
[
  {
    "id": 1,
    "repo_id": "user/model-1",
    "namespace": "user",
    "name": "model-1",
    "type": "model",
    "private": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### Get Repository

Get details of a specific repository.

```http
GET /api/models/:repo_id
```

**Parameters:**
- `repo_id` (path) - Repository ID (e.g., "user/model-name")

**Response:**
```json
{
  "id": 1,
  "repo_id": "user/model-name",
  "namespace": "user",
  "name": "model-name",
  "type": "model",
  "private": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Preupload

Prepare for upload (huggingface_hub compatibility).

```http
POST /api/models/:repo_id/preupload
```

**Parameters:**
- `repo_id` (path) - Repository ID

**Response:**
```json
{
  "repo_id": "user/model-name",
  "status": "ready"
}
```

### Commit

Commit uploaded files (huggingface_hub compatibility).

```http
POST /api/models/:repo_id/commit
Content-Type: application/json
```

**Request Body:**
```json
{
  "commit_id": "abc123def456",
  "message": "Add model files",
  "files": [
    {
      "path": "config.json",
      "size": 1024,
      "lfs": false
    },
    {
      "path": "model.safetensors",
      "size": 500000000,
      "lfs": true
    }
  ]
}
```

**Response:**
```json
{
  "id": 1,
  "repo_id": "user/model-name",
  "commit_id": "abc123def456",
  "message": "Add model files",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Resolve File

Download a specific file.

```http
GET /:repo_id/resolve/:revision/*path
```

**Parameters:**
- `repo_id` (path) - Repository ID
- `revision` (path) - Git revision (branch, tag, or commit hash)
- `path` (wildcard) - File path in repository

**Response:**
- Binary file content

### Resolve File (API)

```http
GET /api/models/:repo_id/resolve/:revision/*path
```

Same as above but under API path.

### Get Raw File

```http
GET /api/models/:repo_id/raw/:revision/*path
```

Same as resolve but returns raw file.

### LFS Info

Check LFS status (stub - always returns regular files).

```http
GET /api/models/:repo_id/info/lfs
```

**Parameters:**
- `repo_id` (path) - Repository ID

**Response:**
```json
{
  "lfs": false,
  "size": 0
}
```

## Python CLI

### serve

Start the hf-local server.

```bash
hf-local serve [OPTIONS]
```

**Options:**
- `--port`, `-p`: Server port (default: 8080)
- `--data-dir`, `-d`: Data storage directory (default: ./data)
- `--log-level`, `-l`: Log level (default: info)
- `--token`: Authentication token
- `--auth-token`: Enable token authentication
- `--auth-hf`: Enable HuggingFace OAuth
- `--hf-client-id`: HF OAuth client ID
- `--hf-client-secret`: HF OAuth client secret
- `--auth-ldap`: Enable LDAP authentication
- `--ldap-server`: LDAP server address

**Example:**
```bash
# Basic server
hf-local serve --port 9000 --data-dir ./models --log-level debug

# With token auth
hf-local serve --token "my-secret" --auth-token

# With OAuth
hf-local serve --auth-hf --hf-client-id "abc123" --hf-client-secret "xyz789"

# With LDAP
hf-local serve --auth-ldap --ldap-server "ldap.company.com"
```

### upload

Upload files or folders to a repository.

```bash
hf-local upload LOCAL_PATH REPO_ID [OPTIONS]
```

**Arguments:**
- `LOCAL_PATH`: Path to file or folder
- `REPO_ID`: Repository ID (e.g., "user/model-name")

**Options:**
- `--endpoint`, `-e`: Server endpoint (default: http://localhost:8080)

**Example:**
```bash
# Upload single file
hf-local upload ./config.json user/my-model

# Upload folder
hf-local upload ./model-folder user/my-model

# With custom endpoint
hf-local upload ./model.bin user/my-model --endpoint http://localhost:9000
```

### list-repos

List all repositories.

```bash
hf-local list-repos [OPTIONS]
```

**Options:**
- `--endpoint`, `-e`: Server endpoint (default: http://localhost:8080)

**Example:**
```bash
hf-local list-repos
```

### init

Initialize a new hf-local instance.

```bash
hf-local init [OPTIONS]
```

**Options:**
- `--data-dir`, `-d`: Data directory path (default: ./data)

**Example:**
```bash
hf-local init --data-dir ./my-storage
```

### status

Check server status.

```bash
hf-local status [OPTIONS]
```

**Options:**
- `--endpoint`, `-e`: Server endpoint (default: http://localhost:8080)

**Example:**
```bash
hf-local status
```

### login

Login with authentication token.

```bash
hf-local login [OPTIONS]
```

**Options:**
- `--token`, `-t`: Authentication token (required)
- `--endpoint`, `-e`: Server endpoint (default: http://localhost:8080)

**Example:**
```bash
hf-local login --token "my-secret-token"
```

### logout

Logout and clear stored credentials.

```bash
hf-local logout
```

**Example:**
```bash
hf-local logout
```

## Python Library

### set_endpoint

Set HF_ENDPOINT environment variable.

```python
from hf_local import set_endpoint

set_endpoint("http://localhost:8080")
```

**Parameters:**
- `endpoint` (str): Server endpoint URL

### login

Login with authentication token.

```python
from hf_local import login

success = login(token="my-secret-token", endpoint="http://localhost:8080")
if success:
    print("Logged in")
```

**Parameters:**
- `token` (str): Authentication token
- `endpoint` (str, default: "http://localhost:8080"): Server endpoint URL

**Returns:** bool - True if login successful

### logout

Logout and clear stored credentials.

```python
from hf_local import logout

logout()
```

**Note:** Clears HF_TOKEN environment variable.

### serve_background

Context manager to run server in background.

```python
from hf_local import serve_background

with serve_background(port=8081, data_dir="./test-data"):
    # Your code here
    pass
```

**Parameters:**
- `port` (int, default: 8080): Server port
- `data_dir` (str, default: "./data"): Data directory
- `log_level` (str, default: "info"): Log level
- `timeout` (int, default: 5): Startup timeout in seconds

### upload_folder

Upload a folder to a repository.

```python
from hf_local import upload_folder

upload_folder(
    folder_path="./my-model",
    repo_id="user/my-model",
    endpoint="http://localhost:8080"
)
```

**Parameters:**
- `folder_path` (str): Local folder path
- `repo_id` (str): Repository ID
- `endpoint` (str, default: "http://localhost:8080"): Server endpoint

### HfLocalApi

Thin wrapper around HfApi that forces local endpoint.

```python
from hf_local import HfLocalApi

api = HfLocalApi(endpoint="http://localhost:8080")

# Use like HfApi
repo = api.create_repo("user/model")
api.upload_file("config.json", repo_id="user/model")
```

**Parameters:**
- `endpoint` (str, default: "http://localhost:8080"): Server endpoint

**Methods:**
All HfApi methods are available:
- `create_repo()`
- `upload_file()`
- `upload_folder()`
- `list_models()`
- `model_info()`
- And more...

## Error Handling

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 404 | Not Found |
| 500 | Internal Server Error |

### Error Response Format

```json
{
  "error": "Error message description"
}
```

## Rate Limiting

Currently, no rate limiting is implemented for local server usage.

## CORS

Cross-Origin Resource Sharing is enabled for all origins.

**Response Headers:**
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

## Authentication

hf-local-hub supports multiple authentication methods. All authenticated endpoints require a JWT token in the `Authorization` header:

```http
Authorization: Bearer <jwt-token>
```

### Authentication Methods

| Method | Enable Flag | Environment Variable |
|--------|-------------|---------------------|
| Token | `-auth-token` | `HF_LOCAL_AUTH_TOKEN=true` |
| HF OAuth | `-auth-hf` | `HF_LOCAL_AUTH_HF=true` |
| LDAP | `-auth-ldap` | `HF_LOCAL_AUTH_LDAP=true` |

### Token Authentication

Simple shared secret authentication:

```bash
# Server
hf-local serve --token "my-secret" --auth-token

# Client login
hf-local login --token "my-secret"
```

### Hugging Face OAuth

OAuth2 flow with Hugging Face:

```bash
# Server with OAuth
hf-local serve \
  --auth-hf \
  --hf-client-id "your-client-id" \
  --hf-client-secret "your-client-secret" \
  --hf-callback-url "http://localhost:8080/auth/hf/callback"
```

### LDAP

Corporate directory authentication:

```bash
# Server with LDAP
hf-local serve \
  --auth-ldap \
  --ldap-server "ldap.company.com" \
  --ldap-port 389 \
  --ldap-bind-dn "cn=admin,dc=company,dc=com" \
  --ldap-bind-pass "password" \
  --ldap-base-dn "ou=users,dc=company,dc=com" \
  --ldap-filter "(uid=%s)"
```

### JWT Token

All authentication methods return a JWT token with this structure:

```json
{
  "user_id": "username",
  "username": "Display Name",
  "provider": "token|hf|ldap",
  "exp": 1738362000,
  "iat": 1738275600
}
```

Token expiration: 24 hours

## Data Model

### Repository

```typescript
interface Repo {
  id: number;
  repo_id: string;
  namespace: string;
  name: string;
  type: "model" | "dataset";
  private: boolean;
  created_at: string;
  updated_at: string;
}
```

### Commit

```typescript
interface Commit {
  id: number;
  repo_id: string;
  commit_id: string;
  message: string;
  created_at: string;
}
```

### File Index

```typescript
interface FileIndex {
  id: number;
  repo_id: string;
  commit_id: string;
  path: string;
  size: number;
  lfs: boolean;
  sha256: string;
  created_at: string;
}
```

## Compatibility

### Hugging Face Hub Compatibility

hf-local-hub implements a subset of the Hugging Face Hub API:

| Feature | Status |
|---------|---------|
| Create repo | ✓ |
| Upload file | ✓ |
| Upload folder | ✓ |
| Download file | ✓ |
| Download snapshot | ✓ |
| List models | ✓ |
| Get model info | ✓ |
| LFS support | ✓ (stub) |
| Token auth | ✓ |
| OAuth | ✓ |
| LDAP | ✓ |
| Git operations | ✗ |
| Webhooks | ✗ |
| Discussions | ✗ |
| PR/Issues | ✗ |

### Library Compatibility

| Library | Version | Status |
|----------|----------|---------|
| huggingface_hub | ≥0.25.0 | ✓ |
| transformers | ≥4.30.0 | ✓ |
| diffusers | ≥0.20.0 | ✓ |
| accelerate | ≥0.20.0 | ✓ |
| datasets | ≥2.10.0 | ✓ |
