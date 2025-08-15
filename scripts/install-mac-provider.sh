#!/bin/bash
set -e

echo "🚀 QUIVer Provider macOSインストーラー"
echo "=================================="

# Check macOS version
MAC_VERSION=$(sw_vers -productVersion)
echo "✓ macOS $MAC_VERSION 検出"

# Install Ollama if not present
if ! command -v ollama &> /dev/null; then
    echo "📦 Ollamaをインストール中..."
    curl -fsSL https://ollama.ai/install.sh | sh
else
    echo "✓ Ollama インストール済み"
fi

# Start Ollama service
echo "🔧 Ollamaサービスを起動中..."
ollama serve > /dev/null 2>&1 &
OLLAMA_PID=$!
sleep 3

# Pull default model
echo "🤖 LLMモデル(llama3.2:3b)をダウンロード中..."
ollama pull llama3.2:3b || echo "モデルダウンロード済み"

# Install Go if not present
if ! command -v go &> /dev/null; then
    echo "📦 Goをインストール中..."
    if [[ $(uname -m) == "arm64" ]]; then
        GO_ARCH="darwin-arm64"
    else
        GO_ARCH="darwin-amd64"
    fi
    curl -L "https://go.dev/dl/go1.23.0.$GO_ARCH.tar.gz" -o /tmp/go.tar.gz
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    export PATH=/usr/local/go/bin:$PATH
    echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.zshrc
else
    echo "✓ Go インストール済み"
fi

# Clone or update QUIVer
QUIVER_DIR="$HOME/.quiver"
if [ -d "$QUIVER_DIR" ]; then
    echo "📂 QUIVerを更新中..."
    cd "$QUIVER_DIR"
    git pull
else
    echo "📂 QUIVerをクローン中..."
    git clone https://github.com/yukihamada/quiver.git "$QUIVER_DIR"
    cd "$QUIVER_DIR"
fi

# Build provider
echo "🔨 Providerをビルド中..."
cd provider
go mod download
go build -o quiver-provider ./cmd/provider

# Create launch script
echo "📝 起動スクリプトを作成中..."
cat > "$HOME/.quiver/start-provider.sh" << 'EOF'
#!/bin/bash
cd "$HOME/.quiver/provider"

# Get bootstrap node info
BOOTSTRAP_PEER="/ip4/35.221.85.1/tcp/4001/p2p/12D3KooWBBV3Sp2nCk47ygTCpCYfudkURmJUHK1Zmv33o4hCa991"

# Start provider with proper configuration
export PROVIDER_LISTEN_ADDR="/ip4/0.0.0.0/tcp/4002"
export PROVIDER_DHT_BOOTSTRAP_PEERS="$BOOTSTRAP_PEER"
export PROVIDER_OLLAMA_URL="http://localhost:11434"
export PROVIDER_METRICS_PORT="8091"

echo "🚀 QUIVer Provider起動中..."
echo "PeerID生成中..."
./quiver-provider
EOF
chmod +x "$HOME/.quiver/start-provider.sh"

# Create LaunchAgent for auto-start
echo "⚙️  自動起動を設定中..."
mkdir -p ~/Library/LaunchAgents
cat > ~/Library/LaunchAgents/com.quiver.provider.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.quiver.provider</string>
    <key>ProgramArguments</key>
    <array>
        <string>$HOME/.quiver/start-provider.sh</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$HOME/.quiver/provider.log</string>
    <key>StandardErrorPath</key>
    <string>$HOME/.quiver/provider.error.log</string>
</dict>
</plist>
EOF

echo ""
echo "✅ インストール完了!"
echo ""
echo "🎯 次のステップ:"
echo "1. Providerを今すぐ起動:"
echo "   $HOME/.quiver/start-provider.sh"
echo ""
echo "2. 自動起動を有効化:"
echo "   launchctl load ~/Library/LaunchAgents/com.quiver.provider.plist"
echo ""
echo "3. ログを確認:"
echo "   tail -f $HOME/.quiver/provider.log"
echo ""
echo "4. メトリクスを確認:"
echo "   open http://localhost:8091/metrics"
echo ""
echo "📊 あなたのProviderがP2Pネットワークに接続されます!"
echo "💰 推論リクエストを処理して報酬を獲得しましょう!"