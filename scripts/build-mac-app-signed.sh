#!/bin/bash
set -e

echo "🏗️ Building QUIVer Provider for macOS (Unsigned Version)..."

VERSION="1.1.1"
APP_NAME="QUIVerProvider"
BUILD_DIR="build/macos"
DMG_NAME="${APP_NAME}-${VERSION}-unsigned.dmg"

# Clean and create build directory
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Build the provider binary with CGO disabled for compatibility
echo "📦 Building provider binary..."
cd provider
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o "../$BUILD_DIR/quiver-provider-arm64" ./cmd/provider
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o "../$BUILD_DIR/quiver-provider-amd64" ./cmd/provider
cd ..

# Create universal binary
echo "🔨 Creating universal binary..."
lipo -create "$BUILD_DIR/quiver-provider-arm64" "$BUILD_DIR/quiver-provider-amd64" -output "$BUILD_DIR/quiver-provider"
rm "$BUILD_DIR/quiver-provider-arm64" "$BUILD_DIR/quiver-provider-amd64"

# Create simple launcher script that doesn't require app bundle
cat > "$BUILD_DIR/QUIVerProvider.command" << 'EOF'
#!/bin/bash

# QUIVer Provider Launcher
echo "🚀 QUIVer Provider Launcher"
echo "=========================="
echo ""

# Get the directory of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$DIR"

# Check if Ollama is installed
if ! command -v ollama &> /dev/null; then
    echo "❌ Ollamaがインストールされていません"
    echo ""
    echo "以下のコマンドでインストールしてください:"
    echo "curl -fsSL https://ollama.ai/install.sh | sh"
    echo ""
    echo "インストール後、再度このスクリプトを実行してください。"
    read -p "Press Enter to exit..."
    exit 1
fi

# Start Ollama service
echo "🔧 Ollamaサービスを起動中..."
ollama serve > /dev/null 2>&1 &
OLLAMA_PID=$!
sleep 3

# Check for model
if ! ollama list | grep -q "llama3.2:3b"; then
    echo "📦 初回セットアップ: モデルをダウンロードします"
    echo "   ※ 初回は10-20分かかる場合があります"
    ollama pull llama3.2:3b || {
        echo "❌ モデルのダウンロードに失敗しました"
        exit 1
    }
fi

# Bootstrap configuration
BOOTSTRAP_PEER="/ip4/35.221.85.1/tcp/4001/p2p/12D3KooWBBV3Sp2nCk47ygTCpCYfudkURmJUHK1Zmv33o4hCa991"

# Start provider
echo ""
echo "🌐 P2Pネットワークに接続中..."
echo ""

export PROVIDER_LISTEN_ADDR="/ip4/0.0.0.0/tcp/4002"
export PROVIDER_DHT_BOOTSTRAP_PEERS="$BOOTSTRAP_PEER"
export PROVIDER_OLLAMA_URL="http://localhost:11434"
export PROVIDER_METRICS_PORT="8091"

# Show instructions
echo "✅ QUIVer Providerが起動しました!"
echo ""
echo "📊 メトリクス: http://localhost:8091/metrics"
echo "💰 推論リクエストを処理して報酬を獲得できます"
echo ""
echo "終了するには Ctrl+C を押してください"
echo ""

# Run provider
./quiver-provider

# Cleanup
kill $OLLAMA_PID 2>/dev/null || true
EOF

chmod +x "$BUILD_DIR/QUIVerProvider.command"

# Create README
cat > "$BUILD_DIR/README.txt" << 'EOF'
QUIVer Provider for macOS
=========================

セキュリティ警告の回避方法:

1. QUIVerProvider.command を右クリック
2. "開く" を選択
3. 警告ダイアログで "開く" をクリック

または、ターミナルで以下を実行:
  xattr -cr QUIVerProvider.command
  xattr -cr quiver-provider

詳細: https://quiver.network
EOF

# Create Info.plist for metadata
cat > "$BUILD_DIR/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>
    <string>QUIVer Provider</string>
    <key>CFBundleVersion</key>
    <string>$VERSION</string>
    <key>CFBundleIdentifier</key>
    <string>network.quiver.provider</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
</dict>
</plist>
EOF

# Create DMG
echo "💿 Creating DMG..."
mkdir -p "$BUILD_DIR/dmg"
cp "$BUILD_DIR/quiver-provider" "$BUILD_DIR/dmg/"
cp "$BUILD_DIR/QUIVerProvider.command" "$BUILD_DIR/dmg/"
cp "$BUILD_DIR/README.txt" "$BUILD_DIR/dmg/"
cp "$BUILD_DIR/Info.plist" "$BUILD_DIR/dmg/"

# Create background image with instructions
cat > "$BUILD_DIR/dmg/使い方.txt" << 'EOF'
QUIVer Provider の使い方
=======================

1. QUIVerProvider.command をダブルクリック
2. セキュリティ警告が出たら:
   - 右クリック → "開く" → "開く"
3. ターミナルウィンドウが開きます
4. 初回はモデルのダウンロードで時間がかかります
5. 起動後は自動的にP2Pネットワークに接続されます

トラブルシューティング:
- "壊れている"と表示される場合は、右クリックで開いてください
- それでも開けない場合は、ターミナルで実行してください
EOF

# Build DMG without code signing
hdiutil create -volname "$APP_NAME" -srcfolder "$BUILD_DIR/dmg" -ov -format UDZO "$BUILD_DIR/$DMG_NAME"

# Clean up
rm -rf "$BUILD_DIR/dmg"

echo ""
echo "✅ ビルド完了!"
echo "📦 DMGファイル: $BUILD_DIR/$DMG_NAME"
echo ""
echo "⚠️  注意: このビルドは署名されていません"
echo "   ユーザーは右クリック→開くで実行する必要があります"
echo ""
echo "📤 GitHubにアップロード:"
echo "gh release create v$VERSION $BUILD_DIR/$DMG_NAME --title 'QUIVer Provider v$VERSION' --notes '# macOS版 QUIVer Provider

## ⚠️ セキュリティ警告について

macOSのセキュリティ機能により警告が表示されます。以下の方法で回避してください：

### 方法1: 右クリックで開く（推奨）
1. ダウンロードしたDMGを開く
2. **QUIVerProvider.command を右クリック**
3. **「開く」を選択**
4. 警告ダイアログで**「開く」をクリック**

### 方法2: ターミナルから実行
\`\`\`bash
# ダウンロードフォルダで実行
xattr -cr ~/Downloads/QUIVerProvider.command
./QUIVerProvider.command
\`\`\`

## 🚀 機能
- 15種類以上のAIモデル対応
- ワンクリックでP2Pネットワーク接続
- 自動モデルダウンロード
- リアルタイムメトリクス表示

## 💻 システム要件
- macOS 10.15以降
- 8GB以上のRAM（モデルによる）

詳細: https://quiver.network'"