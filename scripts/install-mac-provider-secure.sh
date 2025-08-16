#!/bin/bash
set -e

echo "🚀 QUIVer Provider Secure macOS Installer"
echo "========================================"

# Configuration
QUIVER_RELEASE_URL="https://github.com/quiver-network/quiver/releases/latest"
CHECKSUM_URL="https://quiver.network/checksums/MANIFEST.json"
INSTALL_DIR="$HOME/.quiver"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to verify checksum
verify_checksum() {
    local file=$1
    local expected_sha256=$2
    
    echo -n "Verifying checksum for $(basename "$file")... "
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        actual_sha256=$(shasum -a 256 "$file" | cut -d' ' -f1)
    else
        actual_sha256=$(sha256sum "$file" | cut -d' ' -f1)
    fi
    
    if [ "$actual_sha256" = "$expected_sha256" ]; then
        echo -e "${GREEN}✓ Valid${NC}"
        return 0
    else
        echo -e "${RED}✗ Invalid${NC}"
        echo "Expected: $expected_sha256"
        echo "Actual:   $actual_sha256"
        return 1
    fi
}

# Function to download with progress
download_with_progress() {
    local url=$1
    local output=$2
    echo "Downloading $(basename "$output")..."
    curl -L --progress-bar "$url" -o "$output"
}

# Check macOS version
MAC_VERSION=$(sw_vers -productVersion)
echo "✓ macOS $MAC_VERSION detected"

# Check architecture
if [[ $(uname -m) == "arm64" ]]; then
    ARCH="arm64"
    echo "✓ Apple Silicon detected"
else
    ARCH="amd64"
    echo "✓ Intel processor detected"
fi

# Create temporary directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Download checksums
echo ""
echo "📥 Downloading checksums..."
if ! curl -fsSL "$CHECKSUM_URL" -o "$TEMP_DIR/MANIFEST.json" 2>/dev/null; then
    echo -e "${YELLOW}⚠️  Warning: Could not download checksums. Proceeding without verification.${NC}"
    NO_VERIFY=true
else
    echo "✓ Checksums downloaded"
fi

# Install Ollama if not present
if ! command -v ollama &> /dev/null; then
    echo ""
    echo "📦 Installing Ollama..."
    curl -fsSL https://ollama.ai/install.sh | sh
else
    echo "✓ Ollama already installed"
fi

# Start Ollama service
echo ""
echo "🔧 Starting Ollama service..."
if pgrep -x "ollama" > /dev/null; then
    echo "✓ Ollama already running"
else
    ollama serve > /dev/null 2>&1 &
    OLLAMA_PID=$!
    sleep 3
    echo "✓ Ollama started (PID: $OLLAMA_PID)"
fi

# Install default model with progress
echo ""
if ! ollama list 2>/dev/null | grep -q "llama3.2:3b"; then
    echo "🤖 Installing default model (llama3.2:3b)..."
    
    # Use the progress script if available
    if [ -f "$INSTALL_DIR/scripts/install-model-with-progress.sh" ]; then
        MODEL="llama3.2:3b" "$INSTALL_DIR/scripts/install-model-with-progress.sh"
    else
        # Fallback to simple progress
        echo "Downloading... This may take several minutes."
        ollama pull llama3.2:3b || echo "Error downloading model"
    fi
else
    echo "✓ Default model already installed"
fi

# Install Go if not present
if ! command -v go &> /dev/null; then
    echo ""
    echo "📦 Installing Go..."
    GO_VERSION="1.23.0"
    GO_TAR="go$GO_VERSION.darwin-$ARCH.tar.gz"
    GO_URL="https://go.dev/dl/$GO_TAR"
    
    download_with_progress "$GO_URL" "$TEMP_DIR/$GO_TAR"
    
    # Verify Go checksum (hardcoded for security)
    GO_CHECKSUMS=(
        "go1.23.0.darwin-arm64.tar.gz:b770812aef17d7b2ea406588e2b97689e9557aac7e646fe76218b216e2c51406"
        "go1.23.0.darwin-amd64.tar.gz:ffd070acf59f054e8691b838f274d540572db0bd09654af851e4e76ab88403dc"
    )
    
    for checksum in "${GO_CHECKSUMS[@]}"; do
        file="${checksum%%:*}"
        hash="${checksum#*:}"
        if [ "$GO_TAR" = "$file" ]; then
            if ! verify_checksum "$TEMP_DIR/$GO_TAR" "$hash"; then
                echo -e "${RED}Error: Go download corrupted!${NC}"
                exit 1
            fi
            break
        fi
    done
    
    sudo tar -C /usr/local -xzf "$TEMP_DIR/$GO_TAR"
    export PATH=/usr/local/go/bin:$PATH
    
    # Add to shell profile
    if [[ -f "$HOME/.zshrc" ]]; then
        echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.zshrc
    elif [[ -f "$HOME/.bash_profile" ]]; then
        echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.bash_profile
    fi
    
    echo "✓ Go installed successfully"
else
    echo "✓ Go already installed"
fi

# Create installation directory
echo ""
echo "📂 Setting up QUIVer directory..."
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

# Download QUIVer Provider binary
echo ""
echo "📥 Downloading QUIVer Provider..."
PROVIDER_BINARY="quiver-provider-darwin-$ARCH"
PROVIDER_URL="$QUIVER_RELEASE_URL/download/$PROVIDER_BINARY"

if ! download_with_progress "$PROVIDER_URL" "$TEMP_DIR/$PROVIDER_BINARY" 2>/dev/null; then
    echo -e "${YELLOW}Binary not available. Building from source...${NC}"
    
    # Clone and build from source
    if [ -d "$INSTALL_DIR/source" ]; then
        cd "$INSTALL_DIR/source"
        git pull origin main
    else
        git clone https://github.com/quiver-network/quiver.git "$INSTALL_DIR/source"
        cd "$INSTALL_DIR/source"
    fi
    
    echo "Building provider..."
    cd provider
    go build -o "$INSTALL_DIR/quiver-provider" ./cmd/provider
    cd "$INSTALL_DIR"
else
    # Verify checksum if available
    if [ -z "$NO_VERIFY" ] && [ -f "$TEMP_DIR/MANIFEST.json" ]; then
        EXPECTED_SHA256=$(jq -r ".checksums.\"$PROVIDER_BINARY\".sha256" "$TEMP_DIR/MANIFEST.json" 2>/dev/null)
        if [ "$EXPECTED_SHA256" != "null" ] && [ -n "$EXPECTED_SHA256" ]; then
            if ! verify_checksum "$TEMP_DIR/$PROVIDER_BINARY" "$EXPECTED_SHA256"; then
                echo -e "${RED}Error: Binary checksum verification failed!${NC}"
                exit 1
            fi
        fi
    fi
    
    mv "$TEMP_DIR/$PROVIDER_BINARY" "$INSTALL_DIR/quiver-provider"
fi

chmod +x "$INSTALL_DIR/quiver-provider"

# Generate secure keys
echo ""
echo "🔐 Generating secure keys..."
PRIVATE_KEY="$INSTALL_DIR/provider.key"
if [ ! -f "$PRIVATE_KEY" ]; then
    openssl genpkey -algorithm ED25519 -out "$PRIVATE_KEY"
    chmod 600 "$PRIVATE_KEY"
    echo "✓ Private key generated"
else
    echo "✓ Private key already exists"
fi

# Create configuration
echo ""
echo "📝 Creating configuration..."
cat > "$INSTALL_DIR/config.json" << EOF
{
  "provider_id": "provider_$(uuidgen | tr '[:upper:]' '[:lower:]')",
  "ollama_url": "http://localhost:11434",
  "listen_addr": "/ip4/0.0.0.0/tcp/4001",
  "bootstrap_peers": [
    "/dns4/bootstrap.quiver.network/tcp/4001/p2p/12D3KooWLfChFxVatDEJocxtMgdT8yqRAePZJG26h6WHt6kuCNUW"
  ],
  "private_key_path": "$PRIVATE_KEY",
  "enable_privacy": true,
  "log_level": "info"
}
EOF

# Create LaunchAgent for auto-start
echo ""
echo "🚀 Setting up auto-start..."
PLIST_PATH="$HOME/Library/LaunchAgents/network.quiver.provider.plist"
cat > "$PLIST_PATH" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>network.quiver.provider</string>
    <key>ProgramArguments</key>
    <array>
        <string>$INSTALL_DIR/quiver-provider</string>
        <string>--config</string>
        <string>$INSTALL_DIR/config.json</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$INSTALL_DIR/provider.log</string>
    <key>StandardErrorPath</key>
    <string>$INSTALL_DIR/provider.error.log</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/go/bin</string>
    </dict>
</dict>
</plist>
EOF

# Load the service
launchctl unload "$PLIST_PATH" 2>/dev/null || true
launchctl load "$PLIST_PATH"

# Create uninstall script
cat > "$INSTALL_DIR/uninstall.sh" << 'EOF'
#!/bin/bash
echo "Uninstalling QUIVer Provider..."

# Stop service
launchctl unload ~/Library/LaunchAgents/network.quiver.provider.plist 2>/dev/null
rm -f ~/Library/LaunchAgents/network.quiver.provider.plist

# Remove files
rm -rf ~/.quiver

echo "✓ QUIVer Provider uninstalled"
echo "Note: Ollama and Go were not removed as they may be used by other applications"
EOF
chmod +x "$INSTALL_DIR/uninstall.sh"

# Final message
echo ""
echo "✅ Installation complete!"
echo ""
echo "Provider ID: $(jq -r .provider_id "$INSTALL_DIR/config.json")"
echo "Status: Running"
echo ""
echo "📊 View logs:"
echo "  tail -f $INSTALL_DIR/provider.log"
echo ""
echo "🛑 To stop:"
echo "  launchctl unload ~/Library/LaunchAgents/network.quiver.provider.plist"
echo ""
echo "🗑️  To uninstall:"
echo "  $INSTALL_DIR/uninstall.sh"
echo ""
echo "Thank you for joining the QUIVer network! 🎉"