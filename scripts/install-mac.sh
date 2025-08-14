#!/bin/bash

# QUIVer Mac Installer
# P2P AI推論ネットワークのインストーラー

set -e

echo "=========================================="
echo "QUIVer P2P AI Network Installer for macOS"
echo "=========================================="
echo ""

# 色の定義
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# エラーハンドリング
trap 'echo -e "${RED}エラーが発生しました。インストールを中止します。${NC}"' ERR

# Homebrewのチェック
check_homebrew() {
    if ! command -v brew &> /dev/null; then
        echo -e "${YELLOW}Homebrewがインストールされていません。${NC}"
        echo "Homebrewをインストールしますか？ (y/n)"
        read -r response
        if [[ "$response" == "y" ]]; then
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        else
            echo -e "${RED}Homebrewが必要です。インストールを中止します。${NC}"
            exit 1
        fi
    fi
}

# Ollamaのインストール
install_ollama() {
    echo -e "${GREEN}1. Ollamaをチェック中...${NC}"
    if ! command -v ollama &> /dev/null; then
        echo "Ollamaをインストール中..."
        brew install ollama
    else
        echo "✓ Ollamaは既にインストールされています"
    fi
    
    # Ollamaサービスを起動
    echo "Ollamaサービスを起動中..."
    brew services start ollama 2>/dev/null || true
    
    # モデルのダウンロード
    echo -e "${GREEN}2. LLMモデルをダウンロード中...${NC}"
    echo "これには数分かかる場合があります。"
    ollama pull llama3.2:3b || {
        echo -e "${YELLOW}モデルのダウンロードに失敗しました。後で手動でダウンロードしてください：${NC}"
        echo "ollama pull llama3.2:3b"
    }
}

# Goのインストール
install_go() {
    echo -e "${GREEN}3. Go言語環境をチェック中...${NC}"
    if ! command -v go &> /dev/null; then
        echo "Goをインストール中..."
        brew install go
    else
        echo "✓ Goは既にインストールされています"
    fi
}

# QUIVerのインストール
install_quiver() {
    echo -e "${GREEN}4. QUIVerをインストール中...${NC}"
    
    # インストールディレクトリ
    INSTALL_DIR="$HOME/.quiver"
    mkdir -p "$INSTALL_DIR"
    
    # GitHubから最新版をダウンロード
    echo "最新版をダウンロード中..."
    cd /tmp
    rm -rf quiver-temp
    git clone https://github.com/yukihamada/quiver.git quiver-temp
    cd quiver-temp
    
    # ビルド
    echo "QUIVerをビルド中..."
    make build
    
    # ブートストラップもビルド
    go build -o bin/bootstrap bootstrap/main.go
    
    # バイナリをコピー
    if [ -d "bin" ]; then
        cp bin/* "$INSTALL_DIR/" 2>/dev/null || true
    else
        echo "エラー: ビルドに失敗しました"
        exit 1
    fi
    
    # 設定ファイルを作成
    mkdir -p "$HOME/.config/quiver"
    cat > "$HOME/.config/quiver/config.json" << EOF
{
    "network": "mainnet",
    "bootstrap_nodes": [
        "/ip4/34.146.32.216/tcp/4001/p2p/QmBootstrap1",
        "/ip4/34.146.63.195/tcp/4001/p2p/QmBootstrap2"
    ],
    "api_token": "demo-token",
    "ollama_url": "http://localhost:11434"
}
EOF
    
    # クリーンアップ
    cd /
    rm -rf /tmp/quiver-temp
}

# 起動スクリプトの作成
create_launch_scripts() {
    echo -e "${GREEN}5. 起動スクリプトを作成中...${NC}"
    
    # QUIVerコマンドの作成
    cat > "$HOME/.quiver/quiver" << 'EOF'
#!/bin/bash

QUIVER_HOME="$HOME/.quiver"
CONFIG_FILE="$HOME/.config/quiver/config.json"

case "$1" in
    start)
        echo "QUIVer P2Pネットワークを起動中..."
        
        # Ollamaが起動しているか確認
        if ! pgrep -x "ollama" > /dev/null; then
            echo "Ollamaを起動中..."
            ollama serve > /dev/null 2>&1 &
            sleep 3
        fi
        
        # ブートストラップノードを起動
        echo "ブートストラップノードを起動中..."
        "$QUIVER_HOME/bootstrap" --port 4001 > "$HOME/.quiver/bootstrap.log" 2>&1 &
        echo $! > "$HOME/.quiver/bootstrap.pid"
        sleep 2
        
        # ブートストラップPeerIDを取得
        BOOTSTRAP_PEERID=$(curl -s http://localhost:8090/health | grep -o '"peer_id":"[^"]*' | cut -d'"' -f4)
        export QUIVER_BOOTSTRAP="/ip4/127.0.0.1/tcp/4001/p2p/$BOOTSTRAP_PEERID"
        
        # プロバイダーノードを起動
        echo "プロバイダーノードを起動中..."
        "$QUIVER_HOME/provider" > "$HOME/.quiver/provider.log" 2>&1 &
        echo $! > "$HOME/.quiver/provider.pid"
        
        # ゲートウェイを起動
        echo "ゲートウェイを起動中..."
        "$QUIVER_HOME/gateway" > "$HOME/.quiver/gateway.log" 2>&1 &
        echo $! > "$HOME/.quiver/gateway.pid"
        
        echo ""
        echo "✅ QUIVerが起動しました！"
        echo ""
        echo "APIエンドポイント: http://localhost:8080"
        echo ""
        echo "テストコマンド:"
        echo 'curl -X POST http://localhost:8080/generate \'
        echo '  -H "Content-Type: application/json" \'
        echo '  -d '"'"'{"prompt": "こんにちは", "model": "llama3.2:3b", "token": "test"}'"'"
        ;;
        
    stop)
        echo "QUIVerを停止中..."
        
        # PIDファイルから停止
        for service in gateway provider bootstrap; do
            if [ -f "$HOME/.quiver/$service.pid" ]; then
                PID=$(cat "$HOME/.quiver/$service.pid")
                if kill -0 $PID 2>/dev/null; then
                    kill $PID
                    rm "$HOME/.quiver/$service.pid"
                    echo "✓ ${service}を停止しました"
                fi
            fi
        done
        ;;
        
    status)
        echo "QUIVer ステータス:"
        echo ""
        
        for service in bootstrap provider gateway; do
            if [ -f "$HOME/.quiver/$service.pid" ]; then
                PID=$(cat "$HOME/.quiver/$service.pid")
                if kill -0 $PID 2>/dev/null; then
                    echo "✅ $service: 稼働中 (PID: $PID)"
                else
                    echo "❌ $service: 停止"
                fi
            else
                echo "❌ $service: 停止"
            fi
        done
        
        echo ""
        echo "ログファイル:"
        echo "  $HOME/.quiver/*.log"
        ;;
        
    logs)
        tail -f "$HOME/.quiver"/*.log
        ;;
        
    test)
        echo "P2P推論をテスト中..."
        curl -X POST http://localhost:8080/generate \
          -H "Content-Type: application/json" \
          -d '{"prompt": "What is 2+2?", "model": "llama3.2:3b", "token": "test"}'
        echo ""
        ;;
        
    *)
        echo "使い方: quiver {start|stop|status|logs|test}"
        echo ""
        echo "  start  - QUIVer P2Pネットワークを起動"
        echo "  stop   - QUIVerを停止"
        echo "  status - 稼働状況を確認"
        echo "  logs   - ログを表示"
        echo "  test   - P2P推論をテスト"
        ;;
esac
EOF
    
    chmod +x "$HOME/.quiver/quiver"
    
    # PATHに追加
    if ! grep -q "export PATH=\"\$HOME/.quiver:\$PATH\"" "$HOME/.zshrc" 2>/dev/null; then
        echo 'export PATH="$HOME/.quiver:$PATH"' >> "$HOME/.zshrc"
    fi
    
    # エイリアスも作成
    ln -sf "$HOME/.quiver/quiver" /usr/local/bin/quiver 2>/dev/null || true
}

# メインインストール処理
main() {
    echo "QUIVer P2P AIネットワークをインストールします。"
    echo ""
    
    # 依存関係のチェックとインストール
    check_homebrew
    install_ollama
    install_go
    install_quiver
    create_launch_scripts
    
    echo ""
    echo -e "${GREEN}=========================================="
    echo "✅ インストールが完了しました！"
    echo "=========================================="
    echo ""
    echo "次のコマンドでQUIVerを起動できます："
    echo ""
    echo "  quiver start    # P2Pネットワークを起動"
    echo "  quiver test     # 推論をテスト"
    echo "  quiver status   # ステータス確認"
    echo "  quiver stop     # 停止"
    echo ""
    echo "新しいターミナルを開くか、以下を実行してください："
    echo "  source ~/.zshrc"
    echo -e "${NC}"
}

# 実行
main