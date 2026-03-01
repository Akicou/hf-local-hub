"""
Diffusers integration example.
"""

import os
from diffusers import StableDiffusionPipeline

from hf_local import set_endpoint

# Point to local server
set_endpoint("http://localhost:8080")

# Load stable diffusion pipeline from local server
pipe = StableDiffusionPipeline.from_pretrained(
    "user/my-sd-model",
    torch_dtype="float16"
)
pipe = pipe.to("cuda")

# Generate image
prompt = "A beautiful landscape with mountains and a lake"
image = pipe(prompt).images[0]

# Save image
image.save("output.png")
print(f"Image saved: output.png")
print(f"Prompt: {prompt}")
