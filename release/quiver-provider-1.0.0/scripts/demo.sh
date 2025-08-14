#!/bin/bash

# QUIVer Demo Script
# Demonstrates inference request through P2P network

set -e

echo "==================================="
echo "QUIVer P2P Inference Demo"
echo "==================================="

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if network is running
if ! pgrep -f "gateway" > /dev/null; then
    echo -e "${YELLOW}QUIVer network not running. Starting network...${NC}"
    ./scripts/start-network.sh
    sleep 5
fi

echo -e "\n${BLUE}1. Checking Gateway Health${NC}"
curl -s http://localhost:8080/health | jq .

echo -e "\n${BLUE}2. Making Inference Request${NC}"
echo "Prompt: 'Explain how QUIC protocol improves on TCP'"

RESPONSE=$(curl -s -X POST http://localhost:8080/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Explain how QUIC protocol improves on TCP",
    "model": "llama3.2:3b",
    "token": "test-token-123"
  }')

echo -e "\n${GREEN}Response:${NC}"
echo "$RESPONSE" | jq .

# Extract receipt for verification
RECEIPT=$(echo "$RESPONSE" | jq -r '.receipt')
if [ "$RECEIPT" != "null" ]; then
    echo -e "\n${BLUE}4. Verifying Receipt${NC}"
    echo "Receipt signature: $(echo $RECEIPT | jq -r '.signature' | cut -c1-32)..."
    echo "Provider ID: $(echo $RECEIPT | jq -r '.provider_id' | cut -c1-32)..."
    echo "Timestamp: $(echo $RECEIPT | jq -r '.timestamp')"
fi

echo -e "\n${GREEN}Demo completed successfully!${NC}"