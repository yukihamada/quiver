#!/bin/bash

# QUIVer Network Startup Script
# This script starts the complete QUIVer P2P network

set -e

echo "==================================="
echo "QUIVer P2P Network Startup"
echo "==================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running on macOS or Linux
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Running on macOS"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "Running on Linux"
else
    echo "Unsupported OS: $OSTYPE"
    exit 1
fi

# Function to check if port is available
check_port() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null ; then
        echo -e "${RED}Port $1 is already in use${NC}"
        return 1
    else
        echo -e "${GREEN}Port $1 is available${NC}"
        return 0
    fi
}

# Function to start a service
start_service() {
    local service_name=$1
    local service_dir=$2
    local log_file=$3
    
    echo -e "\n${YELLOW}Starting $service_name...${NC}"
    cd $service_dir
    
    # Build the service
    if [ -f "Makefile" ]; then
        make build
    fi
    
    # Start the service in background
    nohup ./bin/${service_name} > $log_file 2>&1 &
    local pid=$!
    
    # Wait a bit to check if it started successfully
    sleep 2
    
    if ps -p $pid > /dev/null; then
        echo -e "${GREEN}$service_name started with PID $pid${NC}"
        echo $pid > /tmp/quiver_${service_name}.pid
    else
        echo -e "${RED}Failed to start $service_name${NC}"
        cat $log_file
        exit 1
    fi
}

# Check required ports
echo -e "\n${YELLOW}Checking ports...${NC}"
REQUIRED_PORTS=(4001 4002 4003 8080 8081 8082 8083)
for port in "${REQUIRED_PORTS[@]}"; do
    if ! check_port $port; then
        echo "Please free up port $port before starting"
        exit 1
    fi
done

# Create logs directory
LOGS_DIR="/tmp/quiver_logs"
mkdir -p $LOGS_DIR

# Get the base directory
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
echo -e "\nBase directory: $BASE_DIR"

# Start bootstrap node
echo -e "\n${YELLOW}1. Starting Bootstrap Node${NC}"
cd $BASE_DIR
if [ ! -d "bootstrap" ]; then
    echo "Creating bootstrap node directory..."
    mkdir -p bootstrap
    cat > bootstrap/main.go << 'EOF'
package main

import (
    "bytes"
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/libp2p/go-libp2p"
    dht "github.com/libp2p/go-libp2p-kad-dht"
    "github.com/libp2p/go-libp2p/core/crypto"
    "github.com/multiformats/go-multiaddr"
)

func main() {
    ctx := context.Background()
    
    // Generate deterministic key for bootstrap node
    seed := []byte("quiver-bootstrap-node-seed-12345")
    r := bytes.NewReader(seed)
    priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, r)
    if err != nil {
        log.Fatal(err)
    }
    
    listen, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4001")
    
    h, err := libp2p.New(
        libp2p.Identity(priv),
        libp2p.ListenAddrs(listen),
        libp2p.ForceReachabilityPublic(),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Start DHT in bootstrap mode
    _, err = dht.New(ctx, h, dht.Mode(dht.ModeServer))
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Bootstrap node started\n")
    fmt.Printf("ID: %s\n", h.ID())
    fmt.Printf("Addresses:\n")
    for _, addr := range h.Addrs() {
        fmt.Printf("  %s/p2p/%s\n", addr, h.ID())
    }
    
    // Wait for interrupt
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
    <-ch
    
    h.Close()
}
EOF
    
    # Create go.mod for bootstrap
    cat > bootstrap/go.mod << 'EOF'
module github.com/quiver/bootstrap

go 1.23.0

require (
    github.com/libp2p/go-libp2p v0.33.0
    github.com/libp2p/go-libp2p-kad-dht v0.25.2
    github.com/multiformats/go-multiaddr v0.13.0
)
EOF
    
    cd bootstrap
    go mod tidy
    go build -o bootstrap main.go
    cd ..
fi

cd bootstrap
nohup ./bootstrap > $LOGS_DIR/bootstrap.log 2>&1 &
BOOTSTRAP_PID=$!
echo $BOOTSTRAP_PID > /tmp/quiver_bootstrap.pid
sleep 3

# Get bootstrap node address
BOOTSTRAP_ID=$(grep "ID:" $LOGS_DIR/bootstrap.log | awk '{print $2}')
BOOTSTRAP_ADDR="/ip4/127.0.0.1/tcp/4001/p2p/$BOOTSTRAP_ID"
echo -e "${GREEN}Bootstrap node started at: $BOOTSTRAP_ADDR${NC}"

# Export bootstrap address for other services
export QUIVER_BOOTSTRAP="$BOOTSTRAP_ADDR"

# Start Provider service
echo -e "\n${YELLOW}2. Starting Provider Service${NC}"
cd $BASE_DIR/provider
make build
QUIVER_BOOTSTRAP=$BOOTSTRAP_ADDR nohup ./bin/provider > $LOGS_DIR/provider.log 2>&1 &
PROVIDER_PID=$!
echo $PROVIDER_PID > /tmp/quiver_provider.pid
sleep 2

# Start Gateway service  
echo -e "\n${YELLOW}3. Starting Gateway Service${NC}"
cd $BASE_DIR/gateway
make build
QUIVER_BOOTSTRAP=$BOOTSTRAP_ADDR nohup ./bin/gateway > $LOGS_DIR/gateway.log 2>&1 &
GATEWAY_PID=$!
echo $GATEWAY_PID > /tmp/quiver_gateway.pid
sleep 2

# Start Aggregator service
echo -e "\n${YELLOW}4. Starting Aggregator Service${NC}"
cd $BASE_DIR/aggregator
make build
nohup ./bin/aggregator > $LOGS_DIR/aggregator.log 2>&1 &
AGGREGATOR_PID=$!
echo $AGGREGATOR_PID > /tmp/quiver_aggregator.pid
sleep 2

# Verify all services are running
echo -e "\n${YELLOW}Verifying services...${NC}"
sleep 3

SERVICES=("bootstrap:$BOOTSTRAP_PID" "provider:$PROVIDER_PID" "gateway:$GATEWAY_PID" "aggregator:$AGGREGATOR_PID")
ALL_GOOD=true

for service in "${SERVICES[@]}"; do
    IFS=':' read -r name pid <<< "$service"
    if ps -p $pid > /dev/null; then
        echo -e "${GREEN}✓ $name is running (PID: $pid)${NC}"
    else
        echo -e "${RED}✗ $name failed to start${NC}"
        ALL_GOOD=false
    fi
done

if [ "$ALL_GOOD" = true ]; then
    echo -e "\n${GREEN}==================================="
    echo "QUIVer Network Started Successfully!"
    echo "===================================${NC}"
    echo ""
    echo "Services:"
    echo "  Bootstrap: http://localhost:4001"
    echo "  Provider:  http://localhost:8082"
    echo "  Gateway:   http://localhost:8081"
    echo "  Aggregator: http://localhost:8083"
    echo ""
    echo "Logs:"
    echo "  Bootstrap:  tail -f $LOGS_DIR/bootstrap.log"
    echo "  Provider:   tail -f $LOGS_DIR/provider.log"
    echo "  Gateway:    tail -f $LOGS_DIR/gateway.log"
    echo "  Aggregator: tail -f $LOGS_DIR/aggregator.log"
    echo ""
    echo "To stop the network, run: ./scripts/stop-network.sh"
else
    echo -e "\n${RED}Some services failed to start. Check the logs for details.${NC}"
    exit 1
fi