# API Reference

Complete API reference for hf-local-hub server and Python client.

## Table of Contents

- [Server API](#server-api)
- [Authentication](#authentication)
- [API Token Management](#api-token-management)
- [Python CLI](#python-cli)
- [Python Library](#python-library)

## Server API

Base URL: `http://localhost:8080`

### Health Check

Check if server is running. **Public endpoint.**

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

Get enabled authentication methods. **Public endpoint.**

```http
GET /auth/config
```

**Response:**
```json
{
  "hf": true,
  "ldap": false
}
```

---

## Authentication

All API endpoints (except `/health` and `/auth/config`) require authentication. Use either JWT tokens (from OAuth/LDAP login) or API tokens.

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

### Get Current User

Get information about the currently authenticated user. **Requires authentication.**

```http
GET /api/user
Authorization: Bearer <jwt-or-api-token>
```

**Response:**
```json
{
  "id": "user-123",
  "username": "John Doe",
  "email": "john@example.com",
  "is_admin": false,
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

## API Token Management

Users can create API tokens with specific permissions for programmatic access. These tokens work like Hugging Face API tokens.

### List API Tokens

List all API tokens for the authenticated user. **Requires JWT authentication.**

```http
GET /api/tokens/
Authorization: Bearer <jwt-token>
```

**Response:**
```json
{
  "tokens": [
    {
      "id": 1,
      "name": "My API Token",
      "token": "hf_***xxxxxxxx",
      "expires_at": "2026-04-05T00:00:00Z",
      "last_used_at": "2026-03-05T10:00:00Z",
      "created_at": "2026-03-05T00:00:00Z"
    }
  ]
}
```

### Create API Token

Create a new API token with specified permissions. **Requires JWT authentication.**

```http
POST /api/tokens/
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "My API Token",
  "read": true,
  "write": true,
  "delete": false,
  "admin": false,
  "expires_in": 720
}
```

**Parameters:**
- `name` (string, required): Token name for identification
- `read` (boolean): Can read repositories and files (default: false)
- `write` (boolean): Can create/update repositories and files (default: false)
- `delete` (boolean): Can delete repositories and files (default: false)
- `admin` (boolean): Can manage users and tokens (default: false)
- `expires_in` (int): Expiration time in hours (0 = no expiration)

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
  "expires_at": "2026-04-05T00:00:00Z",
  "created_at": "2026-03-05T00:00:00Z"
}
```

**Note:** The full token is only shown once when created. Store it securely!

### Delete API Token

Delete an API token. **Requires JWT authentication.**

```http
DELETE /api/tokens/:id
Authorization: Bearer <jwt-token>
```

**Response:**
```json
{
  "message": "Token deleted"
}
```

---

## Repository Endpoints

All repository endpoints require authentication.

### Create Repository

Create a new repository. **Requires `write` permission.**

```http
POST /api/repos/create
Authorization: Bearer <token>
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
  "updated_at": "2024-01-01T00:00:00Z",
  "url": "http://localhost:8080/api/models/user/model-name"
}
```

### List Repositories

List all repositories. **Requires authentication.**

```http
GET /api/repos/
Authorization: Bearer <token>
```

**Query Parameters:**
- `type` (string): Filter by type (`model` or `dataset`)

### Get Repository

Get details of a specific repository. **Requires authentication.**

```http
GET /api/repos/:repo_id
Authorization: Bearer <token>
```

### Delete Repository

Delete a repository. **Requires `delete` permission.**

```http
DELETE /api/repos/:repo_id
Authorization: Bearer <token>
```

### List Models

List all model repositories. **Requires authentication.**

```http
GET /api/models/
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": "user/model-1",
    "modelId": "model-1",
    "namespace": "user",
    "type": "model",
    "private": false,
    "createdAt": "2024-01-01T00:00:00.000000Z"
  }
]
```

### Get Model

Get details of a specific model. **Requires authentication.**

```http
GET /api/models/:repo_id
Authorization: Bearer <token>
```

### List Files

List files in a repository. **Requires authentication.**

```http
GET /api/models/:repo_id/files
Authorization: Bearer <token>
```

**Response:**
```json
{
  "files": [
    {
      "path": "config.json",
      "size": 1024,
      "is_dir": false,
      "mod_time": 1704067200
    }
  ],
  "count": 1
}
```

### Upload File

Upload a file to a repository. **Requires `write` permission.**

```http
POST /api/repos/:repo_id/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**Form Data:**
- `file`: File to upload
- `path`: Target path in repository (optional, defaults to filename)

**Query Parameters:**
- `revision`: Git revision (default: "main")

**Response:**
```json
{
  "path": "config.json",
  "size": 1024,
  "sha256": "abc123..."
}
```

### Preupload

Prepare for upload (huggingface_hub compatibility). **Requires `write` permission.**

```http
POST /api/models/:repo_id/preupload
Authorization: Bearer <token>
```

### Commit

Commit uploaded files (huggingface_hub compatibility). **Requires `write` permission.**

```http
POST /api/models/:repo_id/commit
Authorization: Bearer <token>
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
      "lfs": false,
      "sha": "abc123"
    }
  ]
}
```

### Resolve File

Download a specific file. **Requires authentication.**

```http
GET /api/models/:repo_id/resolve/:revision/*path
Authorization: Bearer <token>
```

**Parameters:**
- `repo_id` (path) - Repository ID
- `revision` (path) - Git revision (branch, tag, or commit hash)
- `path` (wildcard) - File path in repository

**Response:**
- Binary file content

### Get Raw File

Same as resolve but returns raw file. **Requires authentication.**

```http
GET /api/models/:repo_id/raw/:revision/*path
Authorization: Bearer <token>
```

### LFS Batch

Handle LFS batch operations.

```http
POST /api/repos/:repo_id/lfs/info/lfs/batch
Content-Type: application/json
```

**Request Body:**
```json
{
  "operation": "download",
  "objects": [
    {"oid": "abc123", "size": 1024}
  ]
}
```

### LFS Upload Object

Upload an LFS object. **Requires `write` permission.**

```http
PUT /api/repos/:repo_id/lfs/objects/:oid
Content-Type: application/octet-stream
```

### LFS Download Object

Download an LFS object.

```http
GET /api/repos/:repo_id/lfs/objects/:oid
```

### LFS Info

Check LFS status.

```http
GET /api/models/:repo_id/info/lfs?oid=abc123
```

**Response:**
```json
{
  "lfs": true,
  "size": 1024,
  "oid": "abc123"
}
```

---

## Datasets Endpoints

Similar to models endpoints but for datasets.

### List Datasets

```http
GET /api/datasets/
Authorization: Bearer <token>
```

### Get Dataset

```http
GET /api/datasets/:repo_id
Authorization: Bearer <token>
```

### Dataset Files

```http
GET /api/datasets/:repo_id/files
Authorization: Bearer <token>
```

---

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

**Example:**
```bash
hf-local serve --port 9000 --data-dir ./models --log-level debug
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
hf-local upload ./model-folder user/my-model
```

### list-repos

List all repositories.

```bash
hf-local list-repos [OPTIONS]
```

**Options:**
- `--endpoint`, `-e`: Server endpoint (default: http://localhost:8080)

### init

Initialize a new hf-local instance.

```bash
hf-local init [OPTIONS]
```

**Options:**
- `--data-dir`, `-d`: Data directory path (default: ./data)

### status

Check server status.

```bash
hf-local status [OPTIONS]
```

**Options:**
- `--endpoint`, `-e`: Server endpoint (default: http://localhost:8080)

---

## Python Library

### set_endpoint

Set HF_ENDPOINT environment variable.

```python
from hf_local import set_endpoint

set_endpoint("http://localhost:8080")
```

### login

Login with authentication token.

```python
from hf_local import login

success = login(token="your-jwt-or-api-token", endpoint="http://localhost:8080")
if success:
    print("Logged in")
```

### logout

Logout and clear stored credentials.

```python
from hf_local import logout

logout()
```

### serve_background

Context manager to run server in background.

```python
from hf_local import serve_background

with serve_background(port=8081, data_dir="./test-data"):
    # Your code here
    pass
```

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

### HfLocalApi

Thin wrapper around HfApi that forces local endpoint.

```python
from hf_local import HfLocalApi

api = HfLocalApi(endpoint="http://localhost:8080")

# Use like HfApi
repo = api.create_repo("user/model")
api.upload_file("config.json", repo_id="user/model")
```

---

## Error Handling

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized (authentication required) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Not Found |
| 500 | Internal Server Error |

### Error Response Format

```json
{
  "error": "Error message description"
}
```

---

## CORS

Cross-Origin Resource Sharing is enabled for all origins.

**Response Headers:**
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

---

## Data Model

### Repository

```typescript
interface Repo {
  id: string;           // "namespace/name"
  modelId: string;      // "name"
  namespace: string;
  type: "model" | "dataset";
  private: boolean;
  createdAt: string;    // ISO 8601
}
```

### User

```typescript
interface User {
  id: number;
  user_id: string;
  username: string;
  email: string;
  provider: "local" | "hf" | "ldap";
  is_active: boolean;
  is_admin: boolean;
  created_at: string;
  updated_at: string;
}
```

### API Token

```typescript
interface APIToken {
  id: number;
  token: string;        // "hf_xxxx..."
  name: string;
  user_id: string;
  permissions: string;  // JSON: {"read":true,"write":true,...}
  expires_at: string | null;
  last_used_at: string | null;
  created_at: string;
}
```

### Token Permissions

```typescript
interface TokenPermissions {
  read: boolean;    // Can read repos/files
  write: boolean;   // Can create/update repos/files
  delete: boolean;  // Can delete repos/files
  admin: boolean;   // Can manage users and tokens
}
```

---

## Compatibility

### Hugging Face Hub Compatibility

| Feature | Status |
|---------|---------|
| Create repo | ✓ |
| Upload file | ✓ |
| Upload folder | ✓ |
| Download file | ✓ |
| Download snapshot | ✓ |
| List models | ✓ |
| Get model info | ✓ |
| LFS support | ✓ |
| HF OAuth | ✓ |
| LDAP | ✓ |
| API Tokens | ✓ |
| PostgreSQL | ✓ |
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
