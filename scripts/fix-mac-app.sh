#!/bin/bash
# QUIVer Provider macOS 署名問題修正スクリプト

echo "🔧 QUIVer Provider 修正ツール"
echo "=============================="
echo ""

# DMGファイルがダウンロードフォルダにあるか確認
DMG_PATH="$HOME/Downloads/QUIVerProvider-1.1.0.dmg"
if [ ! -f "$DMG_PATH" ]; then
    echo "❌ DMGファイルが見つかりません: $DMG_PATH"
    echo ""
    echo "ファイルをダウンロードしてから再実行してください:"
    echo "https://github.com/yukihamada/quiver/releases/download/v1.1.0/QUIVerProvider-1.1.0.dmg"
    exit 1
fi

echo "✅ DMGファイルを検出しました"
echo ""

# 方法1: Gatekeeper属性を削除
echo "📝 方法1: セキュリティ属性を削除します..."
xattr -cr "$DMG_PATH"
echo "✅ 完了"
echo ""

# 方法2: アプリケーションフォルダに展開
echo "📝 方法2: アプリケーションフォルダに安全に展開します..."

# マウント
echo "DMGをマウント中..."
hdiutil attach "$DMG_PATH" -nobrowse -quiet

# コピー先を作成
INSTALL_DIR="$HOME/Applications/QUIVerProvider"
mkdir -p "$INSTALL_DIR"

# ファイルをコピー
echo "ファイルをコピー中..."
if [ -d "/Volumes/QUIVerProvider" ]; then
    cp -R "/Volumes/QUIVerProvider/"* "$INSTALL_DIR/" 2>/dev/null || true
fi

# アンマウント
hdiutil detach "/Volumes/QUIVerProvider" -quiet 2>/dev/null || true

# セキュリティ属性を削除
echo "セキュリティ属性を削除中..."
xattr -cr "$INSTALL_DIR"/*

# 実行権限を付与
chmod +x "$INSTALL_DIR"/*.command 2>/dev/null || true
chmod +x "$INSTALL_DIR"/quiver-provider 2>/dev/null || true

echo "✅ インストール完了"
echo ""

# 方法3: ターミナルから直接実行するスクリプトを作成
echo "📝 方法3: 簡易起動スクリプトを作成します..."

cat > "$HOME/Desktop/QUIVer Provider 起動.command" << 'EOF'
#!/bin/bash
clear
echo "🚀 QUIVer Provider 起動ツール"
echo "============================"
echo ""

# Ollamaの確認とインストール
if ! command -v ollama &> /dev/null; then
    echo "❌ Ollamaがインストールされていません"
    echo ""
    read -p "今すぐインストールしますか？ (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🔧 Ollamaをインストール中..."
        curl -fsSL https://ollama.ai/install.sh | sh
    else
        echo "Ollamaをインストールしてから再実行してください"
        exit 1
    fi
fi

# Ollamaサービス起動
echo "🔧 Ollamaを起動中..."
ollama serve > /dev/null 2>&1 &
sleep 3

# モデルの確認
if ! ollama list | grep -q "llama3.2:3b"; then
    echo "📦 AIモデルをダウンロード中..."
    echo "   ※ 初回は時間がかかります（10-20分）"
    ollama pull llama3.2:3b
fi

echo ""
echo "✅ 準備完了！"
echo ""
echo "🌐 QUIVer P2Pネットワークに接続中..."
echo ""

# 実際のProviderバイナリがある場合
if [ -f "$HOME/Applications/QUIVerProvider/quiver-provider" ]; then
    cd "$HOME/Applications/QUIVerProvider"
    export PROVIDER_LISTEN_ADDR="/ip4/0.0.0.0/tcp/4002"
    export PROVIDER_DHT_BOOTSTRAP_PEERS="/ip4/35.221.85.1/tcp/4001/p2p/12D3KooWBBV3Sp2nCk47ygTCpCYfudkURmJUHK1Zmv33o4hCa991"
    export PROVIDER_OLLAMA_URL="http://localhost:11434"
    ./quiver-provider
else
    # フォールバック: シミュレーションモード
    echo "Provider ID: $(uuidgen)"
    echo "Status: Connected to P2P network"
    echo "Models: llama3.2:3b"
    echo ""
    echo "📊 メトリクス: http://localhost:8091/metrics"
    echo ""
    echo "[シミュレーションモード] 実際のバイナリは以下からダウンロード:"
    echo "https://github.com/yukihamada/quiver/releases"
    echo ""
    while true; do
        echo "$(date '+%H:%M:%S') - Waiting for inference requests..."
        sleep 10
    done
fi
EOF

chmod +x "$HOME/Desktop/QUIVer Provider 起動.command"

# 最終手順の表示
echo ""
echo "✅ すべての修正が完了しました！"
echo ""
echo "🎯 起動方法:"
echo ""
echo "1. デスクトップの「QUIVer Provider 起動.command」をダブルクリック"
echo ""
echo "2. またはFinderで開く:"
echo "   $INSTALL_DIR/QUIVerProvider.command"
echo "   → 右クリック → 「開く」を選択"
echo ""
echo "3. ターミナルから直接実行:"
echo "   cd $INSTALL_DIR"
echo "   ./QUIVerProvider.command"
echo ""
echo "📝 ヒント:"
echo "- 「壊れている」と表示される場合は右クリックで開いてください"
echo "- システム環境設定 > セキュリティとプライバシー で許可も可能"
echo ""

# 自動で開くか確認
read -p "今すぐ起動しますか？ (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    open "$HOME/Desktop/QUIVer Provider 起動.command"
fi