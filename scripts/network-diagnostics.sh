#!/bin/bash

echo "🔍 QUIVer Network Diagnostics"
echo "============================"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Function to check if a port is open
check_port() {
    local host=$1
    local port=$2
    local service=$3
    
    echo -n "Checking $service ($host:$port)... "
    
    if nc -z -w 2 $host $port 2>/dev/null; then
        echo -e "${GREEN}✓ Open${NC}"
        return 0
    else
        echo -e "${RED}✗ Closed${NC}"
        return 1
    fi
}

# Function to check service health
check_health() {
    local url=$1
    local service=$2
    
    echo -n "Checking $service health... "
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null)
    
    if [ "$response" = "200" ]; then
        echo -e "${GREEN}✓ Healthy${NC}"
        
        # Get detailed health info
        health_data=$(curl -s "$url" 2>/dev/null)
        if [ -n "$health_data" ]; then
            echo "  Details: $health_data" | jq -C '.' 2>/dev/null || echo "  $health_data"
        fi
        return 0
    else
        echo -e "${RED}✗ Unhealthy (HTTP $response)${NC}"
        return 1
    fi
}

# Function to check P2P connectivity
check_p2p() {
    local bootstrap=$1
    
    echo -n "Checking P2P bootstrap... "
    
    # Extract host and port from multiaddr
    if [[ $bootstrap =~ /ip4/([0-9.]+)/tcp/([0-9]+) ]]; then
        host="${BASH_REMATCH[1]}"
        port="${BASH_REMATCH[2]}"
        
        if nc -z -w 2 $host $port 2>/dev/null; then
            echo -e "${GREEN}✓ Reachable${NC}"
            return 0
        else
            echo -e "${RED}✗ Unreachable${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠️  Invalid address format${NC}"
        return 1
    fi
}

# 1. Check local services
echo -e "\n${YELLOW}1. Local Services${NC}"
echo "----------------"

check_port localhost 8080 "Gateway API"
check_port localhost 4001 "Provider P2P"
check_port localhost 8090 "Provider Metrics"
check_port localhost 11434 "Ollama"

# 2. Check service health
echo -e "\n${YELLOW}2. Service Health${NC}"
echo "----------------"

check_health "http://localhost:8080/health" "Gateway"
check_health "http://localhost:8090/health" "Provider"

# 3. Check Ollama
echo -e "\n${YELLOW}3. Ollama Status${NC}"
echo "---------------"

echo -n "Checking Ollama models... "
models=$(ollama list 2>/dev/null | grep -v "NAME" | wc -l)
if [ "$models" -gt 0 ]; then
    echo -e "${GREEN}✓ $models models available${NC}"
    ollama list 2>/dev/null | head -5
else
    echo -e "${RED}✗ No models found${NC}"
fi

# 4. Check P2P network
echo -e "\n${YELLOW}4. P2P Network${NC}"
echo "-------------"

# Default bootstrap nodes
BOOTSTRAP_NODES=(
    "/ip4/34.85.7.88/tcp/4001/p2p/12D3KooWDVNd5wPRYZb3JYdszYW3QaHZnqJDNZ9w2oqST8TGYnJF"
    "/ip4/35.243.115.136/tcp/4001/p2p/12D3KooWLjvJznPvHRuH2KNhgF7z2v2RRoZLvT7bUbYXCdXPmiBF"
    "/dns4/bootstrap.quiver.network/tcp/4001/p2p/12D3KooWLfChFxVatDEJocxtMgdT8yqRAePZJG26h6WHt6kuCNUW"
)

for node in "${BOOTSTRAP_NODES[@]}"; do
    check_p2p "$node"
done

# 5. Check DNS resolution
echo -e "\n${YELLOW}5. DNS Resolution${NC}"
echo "----------------"

echo -n "Resolving quiver.network... "
if host quiver.network > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Resolved${NC}"
    host quiver.network | grep "has address" | head -2
else
    echo -e "${RED}✗ Failed${NC}"
fi

echo -n "Resolving bootstrap.quiver.network... "
if host bootstrap.quiver.network > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Resolved${NC}"
    host bootstrap.quiver.network | grep "has address" | head -2
else
    echo -e "${RED}✗ Failed${NC}"
fi

# 6. Check firewall (macOS specific)
echo -e "\n${YELLOW}6. Firewall Status${NC}"
echo "-----------------"

if [[ "$OSTYPE" == "darwin"* ]]; then
    fw_status=$(/usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate 2>/dev/null | grep "enabled")
    if [ -n "$fw_status" ]; then
        echo -e "${YELLOW}⚠️  Firewall is enabled${NC}"
        echo "  You may need to allow incoming connections for QUIVer"
    else
        echo -e "${GREEN}✓ Firewall is disabled${NC}"
    fi
fi

# 7. Network interfaces
echo -e "\n${YELLOW}7. Network Interfaces${NC}"
echo "--------------------"

if command -v ip > /dev/null 2>&1; then
    ip -4 addr show | grep -E "inet|^[0-9]+:" | grep -v "127.0.0.1"
else
    ifconfig | grep -E "inet |^[a-z]" | grep -v "127.0.0.1" | head -10
fi

# 8. Test inference
echo -e "\n${YELLOW}8. Test Inference${NC}"
echo "----------------"

if check_port localhost 8080 "Gateway" > /dev/null 2>&1; then
    echo "Sending test request to gateway..."
    response=$(curl -s -X POST http://localhost:8080/generate \
        -H "Content-Type: application/json" \
        -d '{"prompt": "Hello", "model": "llama3.2:3b", "max_tokens": 10}' 2>&1)
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Request successful${NC}"
        echo "$response" | jq -C '.' 2>/dev/null || echo "$response"
    else
        echo -e "${RED}✗ Request failed${NC}"
        echo "$response"
    fi
else
    echo -e "${YELLOW}⚠️  Gateway not running, skipping test${NC}"
fi

# Summary
echo -e "\n${YELLOW}Summary${NC}"
echo "======="

issues=0

# Check critical services
if ! check_port localhost 8080 "Gateway" > /dev/null 2>&1; then
    echo -e "${RED}• Gateway is not running${NC}"
    ((issues++))
fi

if ! check_port localhost 4001 "Provider" > /dev/null 2>&1; then
    echo -e "${RED}• Provider is not running${NC}"
    ((issues++))
fi

if ! check_port localhost 11434 "Ollama" > /dev/null 2>&1; then
    echo -e "${RED}• Ollama is not running${NC}"
    ((issues++))
fi

if [ $issues -eq 0 ]; then
    echo -e "${GREEN}✓ All systems operational${NC}"
else
    echo -e "${RED}✗ $issues issue(s) found${NC}"
    echo ""
    echo "Troubleshooting:"
    echo "1. Start missing services:"
    echo "   - Gateway: cd gateway && go run cmd/gateway/main.go"
    echo "   - Provider: cd provider && go run cmd/provider/main.go"
    echo "   - Ollama: ollama serve"
    echo "2. Check logs for errors"
    echo "3. Ensure ports are not blocked by firewall"
fi