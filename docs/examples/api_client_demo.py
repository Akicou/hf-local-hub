"""
Custom API client example.
"""

from hf_local import HfLocalApi, serve_background

# Start server in background for this example
with serve_background(port=8081, data_dir="./example-data"):
    # Create API client
    api = HfLocalApi(endpoint="http://localhost:8081")

    # Create repository
    repo = api.create_repo(
        repo_id="user/example-model",
        repo_type="model",
        private=False
    )
    print(f"Created repository: {repo.repo_id}")

    # Upload configuration
    api.upload_file(
        path_or_fileobj="./config.json",
        path_in_repo="config.json",
        repo_id="user/example-model",
        repo_type="model"
    )
    print("Uploaded config.json")

    # List all models
    models = api.list_models()
    print(f"\nTotal models: {len(models)}")
    for model in models:
        print(f"  - {model.repo_id}")

    # Get repository info
    info = api.model_info("user/example-model")
    print(f"\nRepository info:")
    print(f"  ID: {info.modelId}")
    print(f"  Created: {info.created_at}")
