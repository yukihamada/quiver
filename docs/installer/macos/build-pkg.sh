#!/bin/bash
set -euo pipefail

# QUIVer Provider PKG Installer Build Script
# Creates a signed macOS installer package

# Configuration
APP_NAME="QUIVer Provider"
APP_IDENTIFIER="com.quiver.provider"
VERSION="${VERSION:-1.0.0}"
BUILD_DIR="build"
DIST_DIR="dist"
RESOURCES_DIR="installer/resources"

# Binary paths
PROVIDER_BINARY="../../../provider/build/provider"
GATEWAY_BINARY="../../../gateway/build/gateway"

# Installation paths
INSTALL_PREFIX="/usr/local"
INSTALL_BIN="$INSTALL_PREFIX/bin"
INSTALL_SHARE="$INSTALL_PREFIX/share/quiver"
LAUNCH_AGENTS_DIR="/Library/LaunchAgents"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}[BUILD]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Clean previous builds
clean_build() {
    print_status "Cleaning previous builds..."
    rm -rf "$BUILD_DIR" "$DIST_DIR"
    mkdir -p "$BUILD_DIR" "$DIST_DIR"
}

# Create package structure
create_package_structure() {
    print_status "Creating package structure..."
    
    # Create root structure
    local pkg_root="$BUILD_DIR/root"
    mkdir -p "$pkg_root$INSTALL_BIN"
    mkdir -p "$pkg_root$INSTALL_SHARE"
    mkdir -p "$pkg_root$LAUNCH_AGENTS_DIR"
    
    # Copy binaries
    if [[ -f "$PROVIDER_BINARY" ]]; then
        cp "$PROVIDER_BINARY" "$pkg_root$INSTALL_BIN/quiver-provider"
        chmod 755 "$pkg_root$INSTALL_BIN/quiver-provider"
    else
        print_warning "Provider binary not found at $PROVIDER_BINARY"
    fi
    
    if [[ -f "$GATEWAY_BINARY" ]]; then
        cp "$GATEWAY_BINARY" "$pkg_root$INSTALL_BIN/quiver-gateway"
        chmod 755 "$pkg_root$INSTALL_BIN/quiver-gateway"
    else
        print_warning "Gateway binary not found at $GATEWAY_BINARY"
    fi
    
    # Create LaunchAgent plist for auto-start
    create_launch_agent "$pkg_root$LAUNCH_AGENTS_DIR/com.quiver.provider.plist"
    
    # Copy configuration templates
    mkdir -p "$pkg_root$INSTALL_SHARE/config"
    create_default_config "$pkg_root$INSTALL_SHARE/config/provider.yaml"
    create_default_config "$pkg_root$INSTALL_SHARE/config/gateway.yaml"
}

# Create LaunchAgent plist
create_launch_agent() {
    local plist_path="$1"
    cat > "$plist_path" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.quiver.provider</string>
    <key>ProgramArguments</key>
    <array>
        <string>$INSTALL_BIN/quiver-provider</string>
        <string>--config</string>
        <string>$INSTALL_SHARE/config/provider.yaml</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
    </dict>
    <key>StandardOutPath</key>
    <string>/var/log/quiver-provider.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/quiver-provider.error.log</string>
</dict>
</plist>
EOF
}

# Create default configuration
create_default_config() {
    local config_path="$1"
    local config_name=$(basename "$config_path")
    
    case "$config_name" in
        "provider.yaml")
            cat > "$config_path" << 'EOF'
# QUIVer Provider Configuration
server:
  host: "0.0.0.0"
  port: 8080

p2p:
  listen_addr: "/ip4/0.0.0.0/tcp/4001"
  bootstrap_peers: []

llm:
  provider: "ollama"
  endpoint: "http://localhost:11434"
  model: "llama2"
  temperature: 0

logging:
  level: "info"
  format: "json"
EOF
            ;;
        "gateway.yaml")
            cat > "$config_path" << 'EOF'
# QUIVer Gateway Configuration
server:
  host: "127.0.0.1"
  port: 8081

rate_limit:
  requests_per_minute: 60
  burst: 10

security:
  max_prompt_length: 4096
  allowed_models:
    - "llama2"
    - "mistral"

logging:
  level: "info"
  format: "json"
EOF
            ;;
    esac
}

# Create installer resources
create_installer_resources() {
    print_status "Creating installer resources..."
    
    mkdir -p "$BUILD_DIR/resources"
    
    # Create welcome message
    cat > "$BUILD_DIR/resources/welcome.html" << EOF
<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', sans-serif;
            line-height: 1.6;
            color: #333;
            padding: 20px;
        }
        h1 {
            color: #1a73e8;
            font-size: 24px;
            margin-bottom: 10px;
        }
        .info {
            background-color: #f0f7ff;
            border-left: 4px solid #1a73e8;
            padding: 10px 15px;
            margin: 15px 0;
        }
        ul {
            margin: 10px 0;
            padding-left: 20px;
        }
    </style>
</head>
<body>
    <h1>Welcome to QUIVer Provider</h1>
    <p>QUIVer Provider is a P2P QUIC provider system that proxies to local LLMs.</p>
    
    <div class="info">
        <strong>This installer will:</strong>
        <ul>
            <li>Install QUIVer Provider and Gateway binaries</li>
            <li>Set up automatic startup via LaunchAgent</li>
            <li>Create default configuration files</li>
            <li>Configure system permissions</li>
        </ul>
    </div>
    
    <p>After installation, QUIVer Provider will start automatically and be available at <code>http://localhost:8080</code></p>
</body>
</html>
EOF

    # Create conclusion message
    cat > "$BUILD_DIR/resources/conclusion.html" << EOF
<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', sans-serif;
            line-height: 1.6;
            color: #333;
            padding: 20px;
        }
        h1 {
            color: #4caf50;
            font-size: 24px;
            margin-bottom: 10px;
        }
        .success {
            background-color: #f0f9ff;
            border-left: 4px solid #4caf50;
            padding: 10px 15px;
            margin: 15px 0;
        }
        code {
            background-color: #f4f4f4;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: 'SF Mono', Monaco, monospace;
        }
    </style>
</head>
<body>
    <h1>Installation Complete!</h1>
    
    <div class="success">
        <strong>QUIVer Provider has been successfully installed.</strong>
    </div>
    
    <p><strong>Getting Started:</strong></p>
    <ul>
        <li>QUIVer Provider is now running in the background</li>
        <li>Access the gateway at: <code>http://localhost:8081</code></li>
        <li>Configuration files are located at: <code>/usr/local/share/quiver/config/</code></li>
    </ul>
    
    <p><strong>Useful Commands:</strong></p>
    <ul>
        <li>Check status: <code>launchctl list | grep quiver</code></li>
        <li>View logs: <code>tail -f /var/log/quiver-provider.log</code></li>
        <li>Stop service: <code>launchctl unload /Library/LaunchAgents/com.quiver.provider.plist</code></li>
        <li>Start service: <code>launchctl load /Library/LaunchAgents/com.quiver.provider.plist</code></li>
    </ul>
</body>
</html>
EOF

    # Create license file
    cat > "$BUILD_DIR/resources/license.txt" << EOF
MIT License

Copyright (c) 2024 QUIVer

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF
}

# Create scripts
create_scripts() {
    print_status "Creating installation scripts..."
    
    mkdir -p "$BUILD_DIR/scripts"
    
    # Preinstall script
    cat > "$BUILD_DIR/scripts/preinstall" << 'EOF'
#!/bin/bash
# Pre-installation script

# Stop existing service if running
if launchctl list | grep -q "com.quiver.provider"; then
    echo "Stopping existing QUIVer Provider service..."
    launchctl unload /Library/LaunchAgents/com.quiver.provider.plist 2>/dev/null || true
fi

# Create log directory
mkdir -p /var/log
touch /var/log/quiver-provider.log
touch /var/log/quiver-provider.error.log

exit 0
EOF
    chmod 755 "$BUILD_DIR/scripts/preinstall"
    
    # Postinstall script
    cat > "$BUILD_DIR/scripts/postinstall" << 'EOF'
#!/bin/bash
# Post-installation script

# Set proper permissions
chmod 644 /Library/LaunchAgents/com.quiver.provider.plist

# Load the LaunchAgent
echo "Starting QUIVer Provider service..."
launchctl load /Library/LaunchAgents/com.quiver.provider.plist

# Wait a moment for service to start
sleep 2

# Check if service is running
if launchctl list | grep -q "com.quiver.provider"; then
    echo "QUIVer Provider service started successfully"
else
    echo "Warning: QUIVer Provider service may not have started correctly"
fi

exit 0
EOF
    chmod 755 "$BUILD_DIR/scripts/postinstall"
}

# Build the package
build_package() {
    print_status "Building installer package..."
    
    local pkg_path="$DIST_DIR/QUIVerProvider-$VERSION.pkg"
    
    # Build component package
    pkgbuild \
        --root "$BUILD_DIR/root" \
        --identifier "$APP_IDENTIFIER" \
        --version "$VERSION" \
        --scripts "$BUILD_DIR/scripts" \
        --install-location "/" \
        "$BUILD_DIR/QUIVerProvider-component.pkg"
    
    # Create distribution XML
    cat > "$BUILD_DIR/distribution.xml" << EOF
<?xml version="1.0" encoding="utf-8"?>
<installer-gui-script minSpecVersion="2.0">
    <title>QUIVer Provider</title>
    <organization>com.quiver</organization>
    <domains enable_localSystem="true"/>
    <options customize="never" require-scripts="true" rootVolumeOnly="true"/>
    
    <welcome file="welcome.html"/>
    <license file="license.txt"/>
    <conclusion file="conclusion.html"/>
    
    <pkg-ref id="$APP_IDENTIFIER">
        <bundle-version/>
    </pkg-ref>
    
    <choices-outline>
        <line choice="default">
            <line choice="$APP_IDENTIFIER"/>
        </line>
    </choices-outline>
    
    <choice id="default"/>
    <choice id="$APP_IDENTIFIER" visible="false">
        <pkg-ref id="$APP_IDENTIFIER"/>
    </choice>
    <pkg-ref id="$APP_IDENTIFIER" version="$VERSION" onConclusion="none">QUIVerProvider-component.pkg</pkg-ref>
</installer-gui-script>
EOF
    
    # Build final product package
    productbuild \
        --distribution "$BUILD_DIR/distribution.xml" \
        --resources "$BUILD_DIR/resources" \
        --package-path "$BUILD_DIR" \
        "$pkg_path"
    
    print_status "Package built: $pkg_path"
    
    # Sign and notarize if credentials are available
    if [[ -n "${APPLE_DEVELOPER_ID:-}" ]]; then
        print_status "Signing and notarizing package..."
        ./notarize.sh "$pkg_path"
    else
        print_warning "APPLE_DEVELOPER_ID not set, skipping signing and notarization"
    fi
}

# Main execution
main() {
    print_status "Building QUIVer Provider installer v$VERSION"
    
    clean_build
    create_package_structure
    create_installer_resources
    create_scripts
    build_package
    
    print_status "Build complete!"
}

main "$@"