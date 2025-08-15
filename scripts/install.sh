#!/bin/bash
set -e

# QUIVer Provider Universal Installer
# Supports custom model selection via QUIVER_MODEL environment variable

echo ""
echo "🚀 QUIVer Provider インストーラー"
echo "=================================="
echo ""

# Detect OS
OS="unknown"
ARCH="unknown"

if [[ "$OSTYPE" == "darwin"* ]]; then
    OS="darwin"
    if [[ $(uname -m) == "arm64" ]]; then
        ARCH="arm64"
    else
        ARCH="amd64"
    fi
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
    ARCH=$(uname -m)
    if [[ "$ARCH" == "x86_64" ]]; then
        ARCH="amd64"
    fi
else
    echo "❌ サポートされていないOS: $OSTYPE"
    exit 1
fi

echo "✓ 検出: $OS ($ARCH)"

# Default model or use environment variable
MODEL="${QUIVER_MODEL:-llama3.2:3b}"
echo "📦 選択モデル: $MODEL"

# Check system requirements based on model
check_requirements() {
    local model=$1
    local total_ram=$(free -g 2>/dev/null | awk '/^Mem:/{print $2}' || sysctl -n hw.memsize 2>/dev/null | awk '{print int($1/1024/1024/1024)}')
    
    case $model in
        "qwen3:0.6b") MIN_RAM=2 ;;
        "qwen3:3b"|"llama3.2:3b") MIN_RAM=8 ;;
        "qwen3:7b"|"mistral:7b") MIN_RAM=16 ;;
        "qwen3:14b") MIN_RAM=32 ;;
        "qwen3:32b"|"qwen3-coder:30b"|"jan-nano:128k") MIN_RAM=64 ;;
        "gpt-oss:20b") MIN_RAM=48 ;;
        "gpt-oss:120b") MIN_RAM=256 ;;
        *) MIN_RAM=8 ;;
    esac
    
    if [[ $total_ram -lt $MIN_RAM ]]; then
        echo "⚠️  警告: $MODEL は ${MIN_RAM}GB 以上のRAMを推奨 (現在: ${total_ram}GB)"
        echo "より軽量なモデルをお勧めします:"
        echo "  - qwen3:0.6b (2GB)"
        echo "  - qwen3:3b (8GB)"
        read -p "続行しますか？ (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

check_requirements "$MODEL"

# Install Ollama if not present
if ! command -v ollama &> /dev/null; then
    echo "📦 Ollamaをインストール中..."
    if [[ "$OS" == "darwin" ]]; then
        curl -fsSL https://ollama.ai/install.sh | sh
    else
        curl -fsSL https://ollama.ai/install.sh | sudo sh
    fi
else
    echo "✓ Ollama インストール済み"
fi

# Start Ollama
echo "🔧 Ollamaサービスを起動中..."
if [[ "$OS" == "darwin" ]]; then
    ollama serve > /dev/null 2>&1 &
    OLLAMA_PID=$!
else
    sudo systemctl enable ollama 2>/dev/null || true
    sudo systemctl start ollama 2>/dev/null || true
fi
sleep 3

# Pull selected model
echo "🤖 モデル ($MODEL) をダウンロード中..."
echo "   ※ 初回は時間がかかります"
ollama pull "$MODEL" || {
    echo "❌ モデルのダウンロードに失敗しました"
    echo "   利用可能なモデル: ollama list"
    exit 1
}

# Install Go if needed
install_go() {
    echo "📦 Goをインストール中..."
    GO_VERSION="1.23.0"
    
    if [[ "$OS" == "darwin" ]]; then
        GO_PACKAGE="go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
    else
        GO_PACKAGE="go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
    fi
    
    wget -q "https://go.dev/dl/$GO_PACKAGE" -O /tmp/go.tar.gz
    
    if [[ "$OS" == "darwin" ]]; then
        sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    else
        sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    fi
    
    rm /tmp/go.tar.gz
    export PATH=/usr/local/go/bin:$PATH
    
    # Add to shell profile
    if [[ -f ~/.zshrc ]]; then
        echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.zshrc
    elif [[ -f ~/.bashrc ]]; then
        echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.bashrc
    fi
}

if ! command -v go &> /dev/null; then
    install_go
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
fi

# Build provider
echo "🔨 Providerをビルド中..."
cd "$QUIVER_DIR/provider"
go mod download
go build -o quiver-provider ./cmd/provider

# Create config with selected model
echo "📝 設定ファイルを作成中..."
mkdir -p "$HOME/.quiver/config"
cat > "$HOME/.quiver/config/provider.yaml" << EOF
# QUIVer Provider Configuration
model: "$MODEL"
ollama_url: "http://localhost:11434"
listen_addr: "/ip4/0.0.0.0/tcp/4002"
bootstrap_peers:
  - "/ip4/35.221.85.1/tcp/4001/p2p/12D3KooWBBV3Sp2nCk47ygTCpCYfudkURmJUHK1Zmv33o4hCa991"
metrics_port: 8091
max_prompt_bytes: 32768
tokens_per_second: 100
EOF

# Create start script
cat > "$HOME/.quiver/start-provider.sh" << 'EOF'
#!/bin/bash
cd "$HOME/.quiver/provider"

# Load configuration
export PROVIDER_CONFIG="$HOME/.quiver/config/provider.yaml"
export PROVIDER_LISTEN_ADDR="/ip4/0.0.0.0/tcp/4002"
export PROVIDER_DHT_BOOTSTRAP_PEERS="/ip4/35.221.85.1/tcp/4001/p2p/12D3KooWBBV3Sp2nCk47ygTCpCYfudkURmJUHK1Zmv33o4hCa991"
export PROVIDER_OLLAMA_URL="http://localhost:11434"
export PROVIDER_METRICS_PORT="8091"

echo "🚀 QUIVer Provider 起動中..."
echo "📊 メトリクス: http://localhost:8091/metrics"
echo "🌐 P2Pネットワークに接続中..."
echo ""

./quiver-provider
EOF
chmod +x "$HOME/.quiver/start-provider.sh"

# Create systemd service (Linux) or launchd plist (macOS)
if [[ "$OS" == "linux" ]]; then
    echo "⚙️  systemdサービスを作成中..."
    sudo tee /etc/systemd/system/quiver-provider.service > /dev/null << EOF
[Unit]
Description=QUIVer Provider
After=network-online.target ollama.service
Wants=network-online.target
Requires=ollama.service

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME/.quiver/provider
ExecStart=$HOME/.quiver/start-provider.sh
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
    
    sudo systemctl daemon-reload
    echo "✓ systemdサービス作成完了"
    
elif [[ "$OS" == "darwin" ]]; then
    echo "⚙️  launchdサービスを作成中..."
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
    echo "✓ launchdサービス作成完了"
fi

# Model-specific optimization
echo "🔧 モデル最適化設定中..."
case $MODEL in
    "qwen3:0.6b"|"qwen3:3b")
        echo "   軽量モデル用に最適化"
        ;;
    "qwen3-coder:30b")
        echo "   コーディング特化設定を適用"
        ;;
    "jan-nano:32k"|"jan-nano:128k")
        echo "   長文コンテキスト用に最適化"
        ;;
    *)
        echo "   標準設定を適用"
        ;;
esac

echo ""
echo "✅ インストール完了!"
echo ""
echo "🎯 次のステップ:"
echo ""
echo "1. 今すぐProviderを起動:"
echo "   $HOME/.quiver/start-provider.sh"
echo ""
echo "2. 自動起動を有効化:"
if [[ "$OS" == "linux" ]]; then
    echo "   sudo systemctl enable quiver-provider"
    echo "   sudo systemctl start quiver-provider"
else
    echo "   launchctl load ~/Library/LaunchAgents/com.quiver.provider.plist"
fi
echo ""
echo "3. ログを確認:"
if [[ "$OS" == "linux" ]]; then
    echo "   journalctl -u quiver-provider -f"
else
    echo "   tail -f $HOME/.quiver/provider.log"
fi
echo ""
echo "4. メトリクスを確認:"
echo "   open http://localhost:8091/metrics"
echo ""
echo "📊 選択したモデル: $MODEL"
echo "💰 推論リクエストを処理して報酬を獲得しましょう!"
echo ""
echo "詳細: https://quiver.network"