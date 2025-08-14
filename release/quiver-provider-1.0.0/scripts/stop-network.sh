#!/bin/bash

# QUIVer Network Stop Script

echo "==================================="
echo "Stopping QUIVer P2P Network"
echo "==================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to stop a service
stop_service() {
    local service_name=$1
    local pid_file="/tmp/quiver_${service_name}.pid"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat $pid_file)
        if ps -p $pid > /dev/null 2>&1; then
            echo -e "${YELLOW}Stopping $service_name (PID: $pid)...${NC}"
            kill $pid
            sleep 1
            
            # Force kill if still running
            if ps -p $pid > /dev/null 2>&1; then
                kill -9 $pid
            fi
            
            echo -e "${GREEN}$service_name stopped${NC}"
        else
            echo -e "${YELLOW}$service_name not running${NC}"
        fi
        rm -f $pid_file
    else
        echo -e "${YELLOW}No PID file found for $service_name${NC}"
    fi
}

# Stop all services
SERVICES=("aggregator" "gateway" "provider" "bootstrap")

for service in "${SERVICES[@]}"; do
    stop_service $service
done

echo -e "\n${GREEN}All services stopped${NC}"