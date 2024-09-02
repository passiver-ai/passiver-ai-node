#!/bin/sh

# Tailscale auth key
TAILSCALE_AUTH_KEY="tskey-client-kDWz8gXGM921CNTRL-hcPKruBtPEU3sjk8tDiwDUewmyXmUCZB"
SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImVmdWNrZm1yam1idHhmaXFicmZuIiwicm9sZSI6ImFub24iLCJpYXQiOjE3MjQyMTY1MzgsImV4cCI6MjAzOTc5MjUzOH0.bd4KzujtfqKDZBGhWuxnYxuRi82cgQU3cjZkwChAaSU"

# Install python if not installed                               
if [ -z "$(command -v python3)" ]; then
    echo "Python is not installed. Installing..."
    sudo apt-get update
    sudo apt-get install -y python3
fi

# Install pip if not installed
if [ -z "$(command -v pip3)" ]; then
    echo "pip is not installed. Installing..."
    sudo apt-get install -y python3-pip
fi

# Install ollama if not installed
if [ -z "$(command -v ollama)" ]; then
    echo "Ollama is not installed. Installing..."
    curl -fsSL https://ollama.com/install.sh | sh
fi

# Install tailscale if not installed
if [ -z "$(command -v tailscale)" ]; then
    echo "Tailscale is not installed. Installing..."
    curl -fsSL https://tailscale.com/install.sh | sh
fi

# Setup tailscale tunnel if already installed
if [ -n "$(command -v tailscale)" ]; then
    echo "Tailscale is installed. Setting up tunnel..."
    sudo tailscale up --operator=$USER --auth-key=$TAILSCALE_AUTH_KEY --advertise-tags=tag:passiver
fi

# Get current tailscale dns
TAILSCALE_DNS=$(tailscale status --json | jq -r '.Self.DNSName' | sed 's/\.$//')

echo "Tailscale DNS: $TAILSCALE_DNS"

if [ -n "$TAILSCALE_DNS" ]; then
    # Enable model
    echo "Adding model to our nodes infrastructure"
    curl -s -L -X POST 'https://efuckfmrjmbtxfiqbrfn.supabase.co/functions/v1/llm/model/add' \
  -H 'Authorization: Bearer '$SUPABASE_ANON_KEY'' \
  -H 'Content-Type: application/json' \
  -d '{
    "model_name": "llama3.1",
    "llm_provider": "ollama",
        "base_url": "http://'$TAILSCALE_DNS':11434"
    }' > /dev/null
    echo "Node connected to the network."
fi

# Ollama install model
ollama run llama3.1
