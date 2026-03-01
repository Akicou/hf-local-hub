# Usage Guide

Complete guide for using hf-local-hub with Hugging Face libraries.

## Table of Contents

- [Quick Start](#quick-start)
- [Server Management](#server-management)
- [Python CLI](#python-cli)
- [Hugging Face Hub Integration](#hugging-face-hub-integration)
- [Transformers Integration](#transformers-integration)
- [Diffusers Integration](#diffusers-integration)
- [Advanced Usage](#advanced-usage)
- [Troubleshooting](#troubleshooting)

## Quick Start

### 1. Start the Server

**Using Go Binary:**
```bash
# Build the server
make server

# Run on default port (8080)
./hf-local

# Run with custom settings
./hf-local -port 9000 -data-dir ./my-models
```

**Using Python CLI:**
```bash
# Install Python package
pip install -e ./python

# Start server
hf-local serve --port 8080

# With custom data directory
hf-local serve --port 8080 --data-dir ./storage
```

**Using Docker:**
```bash
# Build
docker build -t hf-local .

# Run
docker run -p 8080:8080 -v $(pwd)/data:/app/data hf-local
```

### 2. Set HF_ENDPOINT

```bash
# Bash/Zsh
export HF_ENDPOINT=http://localhost:8080

# PowerShell
$env:HF_ENDPOINT = "http://localhost:8080"
```

### 3. Use with Hugging Face Libraries

```python
import os
os.environ["HF_ENDPOINT"] = "http://localhost:8080"

from transformers import AutoModel
model = AutoModel.from_pretrained("user/my-model")
```

## Server Management

### Check Server Status

```bash
hf-local status

# With custom endpoint
hf-local status --endpoint http://localhost:9000
```

### Initialize Data Directory

```bash
hf-local init --data-dir ./my-storage

# Creates:
# my-storage/
#   ├── storage/
#   │   ├── models/
#   │   └── datasets/
```

### Configuration Options

| Option | Default | Description |
|---------|----------|-------------|
| `--port`, `-p` | 8080 | Server port |
| `--data-dir`, `-d` | ./data | Storage directory |
| `--log-level`, `-l` | info | Logging level (debug, info, warn, error) |

## Python CLI

### Upload Files

```bash
# Upload single file
hf-local upload ./config.json user/my-model

# Upload folder
hf-local upload ./my-model-folder user/my-model

# With custom endpoint
hf-local upload ./model.bin user/my-model --endpoint http://localhost:9000
```

### List Repositories

```bash
# List all models
hf-local list-repos

# With custom endpoint
hf-local list-repos --endpoint http://localhost:9000
```

## Hugging Face Hub Integration

### Create Repository

```python
from huggingface_hub import HfApi
import os

os.environ["HF_ENDPOINT"] = "http://localhost:8080"

api = HfApi()

# Create new repository
repo = api.create_repo(
    repo_id="user/my-new-model",
    repo_type="model",
    private=False
)
print(f"Created: {repo.repo_id}")
```

### Upload Files

```python
from huggingface_hub import HfApi

api = HfApi()

# Upload single file
api.upload_file(
    path_or_fileobj="./config.json",
    path_in_repo="config.json",
    repo_id="user/my-model",
    repo_type="model"
)

# Upload folder
api.upload_folder(
    folder_path="./my-model",
    repo_id="user/my-model",
    repo_type="model"
)
```

### Download Files

```python
from huggingface_hub import hf_hub_download, snapshot_download

# Download single file
file_path = hf_hub_download(
    repo_id="user/my-model",
    filename="config.json"
)
print(f"Downloaded: {file_path}")

# Download entire repository
snapshot_dir = snapshot_download(
    repo_id="user/my-model"
)
print(f"Snapshot: {snapshot_dir}")
```

### List Models

```python
from huggingface_hub import list_models

models = list_models()

for model in models:
    print(f"{model.repo_id} - {model.modelId}")
```

## Transformers Integration

### Load Model

```python
import os
from transformers import AutoModel, AutoTokenizer

os.environ["HF_ENDPOINT"] = "http://localhost:8080"

# Load model
model = AutoModel.from_pretrained("user/my-model")

# Load tokenizer
tokenizer = AutoTokenizer.from_pretrained("user/my-model")

# Use
inputs = tokenizer("Hello world!", return_tensors="pt")
outputs = model(**inputs)
```

### Save Model

```python
from transformers import AutoModel

# Load and modify
model = AutoModel.from_pretrained("bert-base-uncased")
# ... modify model ...

# Save to local server
model.save_pretrained("user/my-fine-tuned-model")
```

## Diffusers Integration

```python
import os
from diffusers import StableDiffusionPipeline

os.environ["HF_ENDPOINT"] = "http://localhost:8080"

# Load pipeline
pipe = StableDiffusionPipeline.from_pretrained("user/my-sd-model")

# Generate
image = pipe("A beautiful landscape").images[0]
image.save("output.png")
```

## Advanced Usage

### Multiple Servers

```bash
# Terminal 1: Server on port 8080
hf-local serve --port 8080 --data-dir ./data1

# Terminal 2: Server on port 8081
hf-local serve --port 8081 --data-dir ./data2

# Use different endpoints
export HF_ENDPOINT=http://localhost:8080
# or
export HF_ENDPOINT=http://localhost:8081
```

### Context Manager (Python)

```python
from hf_local import serve_background

# Start server in background for testing
with serve_background(port=8081, data_dir="./test-data"):
    # Your test code here
    from transformers import AutoModel
    model = AutoModel.from_pretrained("test/model")
# Server automatically stopped
```

### Custom API Client

```python
from hf_local import HfLocalApi

# Create client pointing to local server
api = HfLocalApi(endpoint="http://localhost:8080")

# Use just like HfApi
repo = api.create_repo("user/model")
api.upload_file("config.json", repo_id="user/model")
```

## Troubleshooting

### Server Not Running

```bash
# Check status
hf-local status

# If not running, start server
hf-local serve

# Check logs
# Server logs are printed to stdout
```

### Port Already in Use

```bash
# Use different port
hf-local serve --port 9000

# Update HF_ENDPOINT
export HF_ENDPOINT=http://localhost:9000
```

### File Not Found

```python
# Check if file exists
from huggingface_hub import file_exists
exists = file_exists("user/model", "config.json")
print(f"File exists: {exists}")

# List all files
from huggingface_hub import list_repo_tree
files = list_repo_tree("user/model")
for file in files:
    print(file.path)
```

### Connection Errors

```python
import httpx

# Check server health
response = httpx.get("http://localhost:8080/health")
print(response.status_code)  # Should be 200
```

### Permission Errors

```bash
# Ensure data directory is writable
chmod -R 755 ./data

# Or use custom directory with proper permissions
hf-local serve --data-dir /path/to/writable/dir
```

## Best Practices

1. **Set HF_ENDPOINT early**: Always set before importing HF libraries
2. **Use specific versions**: Pin versions in requirements.txt
3. **Test locally first**: Verify models work before production use
4. **Monitor disk space**: Large models can consume significant storage
5. **Backup regularly**: Use version control for important models
6. **Use data directories**: Organize models with separate data dirs

## Examples

See the `examples/` directory for complete examples:
- `basic_upload.py` - Upload model to local server
- `transformers_demo.py` - Use with Transformers
- `diffusers_demo.py` - Use with Diffusers
- `api_client_demo.py` - Custom API client usage
