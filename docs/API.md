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

## Python Library

### set_endpoint

Set HF_ENDPOINT environment variable.

```python
from hf_local import set_endpoint

set_endpoint("http://localhost:8080")
```

**Parameters:**
- `endpoint` (str): Server endpoint URL

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

Authentication is not currently implemented. All requests are public.

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
