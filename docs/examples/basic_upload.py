"""
Basic example: Upload model to local hf-local server.
"""

from hf_local import set_endpoint, upload_folder

# Point to local server
set_endpoint("http://localhost:8080")

# Upload a folder containing model files
upload_folder(
    folder_path="./my-model",
    repo_id="user/my-awesome-model",
    endpoint="http://localhost:8080"
)

print("Model uploaded successfully!")
