# Passiver AI Node

AI Computing Platform based on Distributed GPU Network

## Overview

Passiver AI Node is a distributed network system based on user-provided GPU nodes. The system features:

- Distributed Network Architecture
- GPU Clustering Technology
- Automatic Scaling and Recovery Mechanism
- High-Performance AI Model Processing

## System Architecture

The system consists of three types of nodes:

1. **Validator Node**
   - Validates AI computation results
   - Reviews and approves transactions
   - Maintains system security and integrity

2. **Router Node**
   - Efficiently distributes data and results
   - Manages network traffic
   - Ensures Quality of Service (QoS)

3. **LLM Node**
   - Processes Large Language Models
   - Optimizes GPU resources
   - Handles high-performance computing

## Getting Started

### Requirements

- Go 1.20 or higher
- CUDA-enabled GPU
- Linux or macOS operating system

### Installation

First, install the system dependencies and LLM model:
```bash
curl -fsSL https://raw.githubusercontent.com/your-username/passiver-ai-node/main/scripts/install.sh | bash
```

Then, clone and set up the repository:
```bash
git clone https://github.com/your-username/passiver-ai-node.git
cd passiver-ai-node
go mod download
```

### Running the System

Start each node type as follows:

```bash
# Start Router Node
go run cmd/passiver/main.go -type router -id router1 -api-port 8080

# Start LLM Node
go run cmd/passiver/main.go -type llm -id llm1

# Start Validator Node
go run cmd/passiver/main.go -type validator -id validator1
```

### API Usage

Send requests to the LLM node:
```bash
curl -X POST http://localhost:8080/api/v1/llm/process \
  -H "Content-Type: application/json" \
  -d '{
    "input": "What is artificial intelligence?",
    "model_params": {
      "temperature": 0.7,
      "max_tokens": 2048,
      "top_p": 0.9,
      "top_k": 40,
      "repeat_penalty": 1.1
    }
  }'
```

Check system status:
```bash
curl http://localhost:8080/api/v1/system/status
```

## Features

### Distributed Network
- User-provided GPU nodes
- Global node distribution
- Automatic scaling and recovery

### GPU Clustering
- Optimized parallel processing
- Real-time large-scale data processing
- AI model optimization support

### Node Types
- **Validator Node**: Result validation, transaction approval, security
- **Router Node**: Data distribution, traffic management, QoS
- **LLM Node**: Language model processing, GPU optimization

### System Monitoring
- Real-time resource monitoring
- Performance metrics
- Health checks

## License

Copyright (c) 2024 Passiver AI. All rights reserved.

This software and associated documentation files (the "Software") are subject to the following conditions:

1. The use, copying, modification, and distribution of this software is prohibited without explicit written permission from Passiver AI.

2. This software may not be used for commercial purposes and is restricted to research and educational purposes only.

3. Distribution or sharing of modified versions of this software is prohibited.

4. Any derivative works developed using this software must follow the same license conditions.

5. These license terms must be included in all copies or substantial portions of the software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.