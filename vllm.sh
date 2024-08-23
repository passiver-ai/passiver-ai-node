#!/bin/bash

# Exit on any error
set -e

# Create project directory
mkdir -p vllm-project
cd vllm-project

# Create Dockerfile
cat << EOF > Dockerfile
FROM python:3.9-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    software-properties-common \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip3 install -r requirements.txt

# Copy the application
COPY app.py .

# Expose the FastAPI port
EXPOSE 8000

# Start the FastAPI server
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]
EOF

# Create requirements.txt
cat << EOF > requirements.txt
numpy
vllm
fastapi
uvicorn
EOF

# Create app.py
cat << EOF > app.py
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from vllm import LLM, SamplingParams

app = FastAPI()
llm = LLM(model="facebook/opt-125m")

class GenerateRequest(BaseModel):
    prompt: str
    max_tokens: int = 64
    temperature: float = 0.8
    top_p: float = 0.95

@app.post("/generate")
async def generate(request: GenerateRequest):
    try:
        sampling_params = SamplingParams(
            temperature=request.temperature,
            top_p=request.top_p,
            max_tokens=request.max_tokens
        )
        outputs = llm.generate([request.prompt], sampling_params)
        generated_text = outputs[0].outputs[0].text
        return {"generated_text": generated_text}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/")
async def root():
    return {"message": "vLLM API is running"}

@app.get("/health")
async def health():
    return {"status": "healthy"}
EOF

# Build Docker image
echo "Building Docker image..."
docker build -t vllm-project .

# Run Docker container
echo "Running Docker container..."
docker run --platform linux/amd64 -p 8000:8000 vllm-project

echo "Setup complete and container is running. Access your API at http://localhost:8000"