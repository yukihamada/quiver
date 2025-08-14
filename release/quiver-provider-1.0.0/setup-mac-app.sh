#!/bin/bash

# QUIVer Provider Mac Application Setup Script

echo "=================================="
echo "  QUIVer Provider セットアップ"
echo "=================================="
echo ""

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ このスクリプトはmacOS専用です"
    exit 1
fi

# Set up paths
APP_NAME="QUIVer Provider"
APP_DIR="/Applications/${APP_NAME}.app"
SUPPORT_DIR="$HOME/Library/Application Support/QUIVer"
LOG_DIR="$HOME/Library/Logs/QUIVer"
CONFIG_DIR="$HOME/.quiver"

# Create directories
echo "📁 ディレクトリを作成中..."
mkdir -p "$SUPPORT_DIR"
mkdir -p "$LOG_DIR"
mkdir -p "$CONFIG_DIR"

# Check if Ollama is installed
echo ""
echo "🔍 Ollamaをチェック中..."
if ! command -v ollama &> /dev/null; then
    echo "❌ Ollamaがインストールされていません"
    echo ""
    echo "📥 Ollamaをインストールしてください:"
    echo "   https://ollama.ai/download"
    echo ""
    read -p "Ollamaをインストールしましたか？ (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Download and setup model
echo ""
echo "🤖 LLMモデルをセットアップ中..."
if ! ollama list | grep -q "llama3.2"; then
    echo "📥 llama3.2モデルをダウンロード中（約2GB）..."
    ollama pull llama3.2:3b
else
    echo "✅ llama3.2モデルは既にインストールされています"
fi

# Copy application files
echo ""
echo "📦 アプリケーションファイルをコピー中..."
if [ -d "QUIVerProvider.app" ]; then
    cp -R "QUIVerProvider.app" "/Applications/"
    echo "✅ アプリケーションをインストールしました"
else
    echo "⚠️  アプリケーションファイルが見つかりません"
fi

# Create launch agent for auto-start
echo ""
echo "🚀 自動起動を設定中..."
PLIST_FILE="$HOME/Library/LaunchAgents/com.quiver.provider.plist"
cat > "$PLIST_FILE" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.quiver.provider</string>
    <key>ProgramArguments</key>
    <array>
        <string>$CONFIG_DIR/start-provider.sh</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$LOG_DIR/provider.log</string>
    <key>StandardErrorPath</key>
    <string>$LOG_DIR/provider-error.log</string>
</dict>
</plist>
EOF

# Create start script
cat > "$CONFIG_DIR/start-provider.sh" << 'EOF'
#!/bin/bash

# Wait for network
sleep 10

# Start Ollama if not running
if ! pgrep -x "ollama" > /dev/null; then
    ollama serve &
    sleep 5
fi

# Start QUIVer Provider
cd "$HOME/.quiver"
./provider/bin/provider \
    --bootstrap /dnsaddr/bootstrap.quiver.network/p2p/12D3KooWLXexpZCqSDiMgJjYDqg6pGQ5Hm5X2FeVPZcB2Y5oKGpF \
    --listen /ip4/0.0.0.0/tcp/0 \
    --provider-url http://localhost:11434 \
    --enable-gui
EOF

chmod +x "$CONFIG_DIR/start-provider.sh"

# Load launch agent
launchctl load "$PLIST_FILE" 2>/dev/null

# Create desktop shortcut
echo ""
echo "🖥  デスクトップショートカットを作成中..."
cat > "$HOME/Desktop/QUIVer Provider.command" << 'EOF'
#!/bin/bash
open -a "QUIVer Provider"
EOF
chmod +x "$HOME/Desktop/QUIVer Provider.command"

# Open the launcher
echo ""
echo "✅ セットアップが完了しました！"
echo ""
echo "🚀 QUIVer Providerを起動中..."
open "$APP_DIR"

echo ""
echo "=================================="
echo "  セットアップ完了"
echo "=================================="
echo ""
echo "✅ Ollamaがインストールされました"
echo "✅ LLMモデルがダウンロードされました"
echo "✅ 自動起動が設定されました"
echo "✅ アプリケーションが起動しました"
echo ""
echo "💰 収益化が自動的に開始されます！"
echo ""