#!/bin/bash

# Model installation script with progress display

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'
BOLD='\033[1m'

# Progress bar function
show_progress() {
    local current=$1
    local total=$2
    local width=50
    
    if [ $total -eq 0 ]; then
        percent=0
    else
        percent=$((current * 100 / total))
    fi
    
    completed=$((width * current / total))
    remaining=$((width - completed))
    
    # Build progress bar
    printf "\r["
    printf "%${completed}s" | tr ' ' '█'
    printf "%${remaining}s" | tr ' ' '░'
    printf "] %3d%%" $percent
}

# Spinner animation
spinner() {
    local pid=$1
    local delay=0.1
    local spinstr='⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏'
    
    while [ "$(ps a | awk '{print $1}' | grep $pid)" ]; do
        local temp=${spinstr#?}
        printf " %c  " "$spinstr"
        local spinstr=$temp${spinstr%"$temp"}
        sleep $delay
        printf "\b\b\b\b\b"
    done
    printf "    \b\b\b\b"
}

# Model selection menu
select_model() {
    echo -e "${BOLD}${BLUE}QUIVer Model Installer${NC}"
    echo "====================="
    echo ""
    echo "Available models:"
    echo ""
    echo "  ${BOLD}Recommended for most users:${NC}"
    echo "  1) llama3.2:3b      - Fast, efficient (2GB)"
    echo "  2) phi3:mini        - Very fast, compact (2.3GB)"
    echo ""
    echo "  ${BOLD}Better quality (needs more RAM):${NC}"
    echo "  3) llama3.2:7b      - Balanced performance (4.7GB)"
    echo "  4) mistral:7b       - High quality (4.1GB)"
    echo "  5) gemma2:9b        - Google's model (5.5GB)"
    echo ""
    echo "  ${BOLD}High-end (needs 16GB+ RAM):${NC}"
    echo "  6) llama3.1:70b     - Professional grade (40GB)"
    echo "  7) mixtral:8x7b     - MoE architecture (26GB)"
    echo ""
    echo "  8) Custom model"
    echo "  9) Exit"
    echo ""
    
    read -p "Select model (1-9): " choice
    
    case $choice in
        1) MODEL="llama3.2:3b" ;;
        2) MODEL="phi3:mini" ;;
        3) MODEL="llama3.2:7b" ;;
        4) MODEL="mistral:7b" ;;
        5) MODEL="gemma2:9b" ;;
        6) MODEL="llama3.1:70b" ;;
        7) MODEL="mixtral:8x7b" ;;
        8) 
            read -p "Enter model name: " MODEL
            ;;
        9) 
            echo "Exiting..."
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid choice${NC}"
            exit 1
            ;;
    esac
}

# Check system resources
check_resources() {
    echo -e "\n${YELLOW}Checking system resources...${NC}"
    
    # Get available memory
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        total_mem=$(sysctl -n hw.memsize)
        total_gb=$((total_mem / 1024 / 1024 / 1024))
        
        # Get free disk space
        free_disk=$(df -g / | awk 'NR==2 {print $4}')
    else
        # Linux
        total_mem=$(free -b | awk '/^Mem:/{print $2}')
        total_gb=$((total_mem / 1024 / 1024 / 1024))
        
        # Get free disk space
        free_disk=$(df -BG / | awk 'NR==2 {print $4}' | sed 's/G//')
    fi
    
    echo -e "  RAM: ${total_gb}GB"
    echo -e "  Free disk: ${free_disk}GB"
    
    # Warn if resources might be insufficient
    case $MODEL in
        *"70b"*|*"8x7b"*)
            if [ $total_gb -lt 16 ]; then
                echo -e "\n${YELLOW}⚠️  Warning: This model requires at least 16GB RAM${NC}"
                echo -e "Your system has ${total_gb}GB. Continue anyway? (y/n)"
                read -p "> " confirm
                if [ "$confirm" != "y" ]; then
                    exit 0
                fi
            fi
            ;;
    esac
}

# Download model with progress
download_model() {
    echo -e "\n${BLUE}Downloading $MODEL...${NC}"
    echo "This may take several minutes depending on your connection."
    echo ""
    
    # Create temp file for output
    TEMP_FILE=$(mktemp)
    
    # Start ollama pull in background and capture output
    ollama pull "$MODEL" > "$TEMP_FILE" 2>&1 &
    PULL_PID=$!
    
    # Monitor progress
    last_percent=0
    while kill -0 $PULL_PID 2>/dev/null; do
        # Parse output for progress
        if grep -q "pulling" "$TEMP_FILE"; then
            # Extract percentage if available
            percent=$(tail -n 1 "$TEMP_FILE" | grep -oE '[0-9]+%' | tr -d '%' | tail -1)
            
            if [ -n "$percent" ] && [ "$percent" != "$last_percent" ]; then
                show_progress $percent 100
                last_percent=$percent
            else
                # Show spinner if no percentage
                printf "\r${BLUE}Downloading...${NC} "
                spinner $PULL_PID &
                SPINNER_PID=$!
                wait $PULL_PID
                kill $SPINNER_PID 2>/dev/null
            fi
        fi
        
        sleep 0.5
    done
    
    # Check if successful
    wait $PULL_PID
    EXIT_CODE=$?
    
    # Clean up
    rm -f "$TEMP_FILE"
    
    if [ $EXIT_CODE -eq 0 ]; then
        echo -e "\n\n${GREEN}✓ Model $MODEL downloaded successfully!${NC}"
        return 0
    else
        echo -e "\n\n${RED}✗ Failed to download model${NC}"
        return 1
    fi
}

# Test model
test_model() {
    echo -e "\n${BLUE}Testing model...${NC}"
    
    # Simple test prompt
    response=$(ollama run "$MODEL" "Say 'Hello from QUIVer!' and nothing else" 2>&1)
    
    if [[ $response == *"Hello from QUIVer"* ]]; then
        echo -e "${GREEN}✓ Model is working correctly!${NC}"
        echo -e "\nResponse: $response"
    else
        echo -e "${YELLOW}⚠️  Model test produced unexpected output${NC}"
        echo -e "Response: $response"
    fi
}

# Main execution
main() {
    # Check if ollama is installed
    if ! command -v ollama &> /dev/null; then
        echo -e "${RED}Error: Ollama is not installed${NC}"
        echo "Please install Ollama first: https://ollama.ai"
        exit 1
    fi
    
    # Check if ollama is running
    if ! pgrep -x "ollama" > /dev/null; then
        echo -e "${YELLOW}Starting Ollama service...${NC}"
        ollama serve > /dev/null 2>&1 &
        sleep 3
    fi
    
    # Model selection
    select_model
    
    # Check resources
    check_resources
    
    # Check if model already exists
    if ollama list | grep -q "^$MODEL"; then
        echo -e "\n${YELLOW}Model $MODEL is already installed${NC}"
        echo "Do you want to re-download it? (y/n)"
        read -p "> " redownload
        if [ "$redownload" != "y" ]; then
            test_model
            exit 0
        fi
    fi
    
    # Download model
    download_model
    
    if [ $? -eq 0 ]; then
        # Test the model
        test_model
        
        # Show summary
        echo -e "\n${GREEN}Setup complete!${NC}"
        echo ""
        echo "Model: $MODEL"
        echo "Status: Ready"
        echo ""
        echo "To use with QUIVer:"
        echo "  1. Start the provider: cd provider && go run cmd/provider/main.go"
        echo "  2. The model will be automatically used for inference"
        echo ""
        echo "To test directly with Ollama:"
        echo "  ollama run $MODEL"
    fi
}

# Run main function
main