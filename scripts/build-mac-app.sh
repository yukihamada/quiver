#!/bin/bash
set -e

echo "🏗️ Building QUIVer Provider for macOS..."

# Check if we're on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ This script must be run on macOS"
    exit 1
fi

VERSION="1.1.0"
APP_NAME="QUIVerProvider"
BUILD_DIR="build/macos"
DMG_NAME="${APP_NAME}-${VERSION}.dmg"

# Clean and create build directory
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Build the provider binary
echo "📦 Building provider binary..."
cd provider
go build -o "../$BUILD_DIR/quiver-provider" ./cmd/provider
cd ..

# Create app bundle structure
APP_BUNDLE="$BUILD_DIR/$APP_NAME.app"
mkdir -p "$APP_BUNDLE/Contents/MacOS"
mkdir -p "$APP_BUNDLE/Contents/Resources"

# Copy binary
cp "$BUILD_DIR/quiver-provider" "$APP_BUNDLE/Contents/MacOS/"

# Create Info.plist
cat > "$APP_BUNDLE/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>quiver-provider</string>
    <key>CFBundleIdentifier</key>
    <string>network.quiver.provider</string>
    <key>CFBundleName</key>
    <string>QUIVer Provider</string>
    <key>CFBundleDisplayName</key>
    <string>QUIVer Provider</string>
    <key>CFBundleVersion</key>
    <string>$VERSION</string>
    <key>CFBundleShortVersionString</key>
    <string>$VERSION</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleSignature</key>
    <string>????</string>
    <key>LSApplicationCategoryType</key>
    <string>public.app-category.developer-tools</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSUIElement</key>
    <true/>
</dict>
</plist>
EOF

# Create launcher script
cat > "$APP_BUNDLE/Contents/MacOS/launcher.sh" << 'EOF'
#!/bin/bash

# Setup environment
export PATH="/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"

# Check if Ollama is installed
if ! command -v ollama &> /dev/null; then
    osascript -e 'display dialog "Ollamaがインストールされていません。\n\nインストール後、再度起動してください。" buttons {"インストール", "キャンセル"} default button 1'
    if [ $? -eq 0 ]; then
        open "https://ollama.ai"
    fi
    exit 1
fi

# Start Ollama service
ollama serve > /dev/null 2>&1 &

# Check for model
if ! ollama list | grep -q "llama3.2:3b"; then
    osascript -e 'display dialog "モデルをダウンロードします。\n\n初回は時間がかかります。" buttons {"OK"} default button 1'
    ollama pull llama3.2:3b
fi

# Get bootstrap peer
BOOTSTRAP_PEER="/ip4/35.221.85.1/tcp/4001/p2p/12D3KooWBBV3Sp2nCk47ygTCpCYfudkURmJUHK1Zmv33o4hCa991"

# Start provider
export PROVIDER_LISTEN_ADDR="/ip4/0.0.0.0/tcp/4002"
export PROVIDER_DHT_BOOTSTRAP_PEERS="$BOOTSTRAP_PEER"
export PROVIDER_OLLAMA_URL="http://localhost:11434"
export PROVIDER_METRICS_PORT="8091"

# Run in background
cd "$(dirname "$0")"
./quiver-provider &
PROVIDER_PID=$!

# Show status
osascript -e 'display notification "QUIVer Providerが起動しました" with title "QUIVer"'

# Keep running
wait $PROVIDER_PID
EOF

chmod +x "$APP_BUNDLE/Contents/MacOS/launcher.sh"

# Create icon (placeholder)
echo "📱 Creating app icon..."
cat > "$BUILD_DIR/icon.svg" << 'EOF'
<svg width="512" height="512" viewBox="0 0 512 512" xmlns="http://www.w3.org/2000/svg">
  <rect width="512" height="512" rx="128" fill="url(#gradient)"/>
  <text x="256" y="320" font-family="Arial, sans-serif" font-size="200" font-weight="bold" text-anchor="middle" fill="white">Q</text>
  <defs>
    <linearGradient id="gradient" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#667eea;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#764ba2;stop-opacity:1" />
    </linearGradient>
  </defs>
</svg>
EOF

# Convert SVG to ICNS (requires rsvg-convert and iconutil)
if command -v rsvg-convert &> /dev/null; then
    mkdir -p "$BUILD_DIR/icon.iconset"
    for size in 16 32 64 128 256 512; do
        rsvg-convert -w $size -h $size "$BUILD_DIR/icon.svg" -o "$BUILD_DIR/icon.iconset/icon_${size}x${size}.png"
        rsvg-convert -w $((size*2)) -h $((size*2)) "$BUILD_DIR/icon.svg" -o "$BUILD_DIR/icon.iconset/icon_${size}x${size}@2x.png" 2>/dev/null || true
    done
    iconutil -c icns "$BUILD_DIR/icon.iconset" -o "$APP_BUNDLE/Contents/Resources/AppIcon.icns" 2>/dev/null || true
    rm -rf "$BUILD_DIR/icon.iconset"
fi

# Create DMG
echo "💿 Creating DMG..."
mkdir -p "$BUILD_DIR/dmg"
cp -r "$APP_BUNDLE" "$BUILD_DIR/dmg/"

# Create Applications symlink
ln -s /Applications "$BUILD_DIR/dmg/Applications"

# Create DMG background and settings
cat > "$BUILD_DIR/dmg/.DS_Store_template" << 'EOF'
# DMG window settings would go here
EOF

# Build DMG
hdiutil create -volname "$APP_NAME" -srcfolder "$BUILD_DIR/dmg" -ov -format UDZO "$BUILD_DIR/$DMG_NAME"

# Clean up
rm -rf "$BUILD_DIR/dmg"

echo "✅ Build complete!"
echo "📦 DMG file: $BUILD_DIR/$DMG_NAME"
echo ""
echo "📤 Upload to GitHub:"
echo "gh release create v$VERSION $BUILD_DIR/$DMG_NAME --title 'QUIVer Provider v$VERSION' --notes 'macOS版QUIVer Provider'"