"""
Transformers integration example.
"""

import os
from transformers import AutoModel, AutoTokenizer

from hf_local import set_endpoint

# Point to local server
set_endpoint("http://localhost:8080")

# Load model and tokenizer from local server
model = AutoModel.from_pretrained("user/my-model")
tokenizer = AutoTokenizer.from_pretrained("user/my-model")

# Use the model
text = "Hello, world!"
inputs = tokenizer(text, return_tensors="pt")
outputs = model(**inputs)

print(f"Model loaded: {model.__class__.__name__}")
print(f"Output shape: {outputs.last_hidden_state.shape}")

# Save fine-tuned model back to local server
model.save_pretrained("user/my-fine-tuned-model")
print("Fine-tuned model saved!")
