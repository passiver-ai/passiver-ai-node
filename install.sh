#!/bin/sh

# Tailscale auth key
TAILSCALE_AUTH_KEY="tskey-client-kDWz8gXGM921CNTRL-hcPKruBtPEU3sjk8tDiwDUewmyXmUCZB"
SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImVmdWNrZm1yam1idHhmaXFicmZuIiwicm9sZSI6ImFub24iLCJpYXQiOjE3MjQyMTY1MzgsImV4cCI6MjAzOTc5MjUzOH0.bd4KzujtfqKDZBGhWuxnYxuRi82cgQU3cjZkwChAaSU"

# Install python if not installed                               
if ! command -v python3 &> /dev/null; then
    echo "Python is not installed. Installing..."
    sudo apt-get update
    sudo apt-get install -y python3
fi

# Install pip if not installed
if ! command -v pip3 &> /dev/null; then
    echo "pip is not installed. Installing..."
    sudo apt-get install -y python3-pip
fi

# Install ollama if not installed
if ! command -v ollama &> /dev/null; then
    echo "Ollama is not installed. Installing..."
    curl -fsSL https://ollama.com/install.sh | sh
fi

# Install tailscale if not installed
if ! command -v tailscale &> /dev/null; then
    echo "Tailscale is not installed. Installing..."
    curl -fsSL https://tailscale.com/install.sh | sh --auth-key=$TAILSCALE_AUTH_KEY
fi

# Setup tailscale tunnel if already installed
if command -v tailscale &> /dev/null; then
    echo "Tailscale is installed. Setting up tunnel..."
    tailscale up --auth-key=$TAILSCALE_AUTH_KEY
fi

# Get current tailscale dns
TAILSCALE_DNS=$(tailscale status --json | jq -r '.Self.DNSName' | sed 's/\.$//')

# Enable model
curl -L -X POST 'https://efuckfmrjmbtxfiqbrfn.supabase.co/functions/v1/llm/model/add' \
  -H 'Authorization: Bearer '$SUPABASE_TOKEN'' \
  -H 'Content-Type: application/json' \
  -d '{
    "model_name": "llama3.1",
    "llm_provider": "ollama",
    "base_url": "https://'$TAILSCALE_DNS':11434"
  }'

# Ollama install model
ollama run llama3.1