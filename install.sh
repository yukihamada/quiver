#!/bin/bash

# QUIVer Provider Quick Installer
# This script downloads and installs QUIVer Provider on macOS

set -e

echo "======================================"
echo "   QUIVer Provider インストーラー"
echo "======================================"
echo ""

# Check OS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ このインストーラーはmacOS専用です"
    exit 1
fi

# Check if already installed
if [ -d "$HOME/.quiver" ]; then
    echo "⚠️  QUIVer Providerは既にインストールされています"
    echo ""
    echo "再インストールしますか？ (y/n)"
    read -r response
    if [[ "$response" != "y" ]]; then
        exit 0
    fi
    rm -rf "$HOME/.quiver"
fi

echo "📥 QUIVer Providerをダウンロード中..."

# Create temp directory
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

# Download latest release
DOWNLOAD_URL="https://github.com/yukihamada/quiver/releases/latest/download/quiver-provider-macos.tar.gz"
curl -L -o quiver.tar.gz "$DOWNLOAD_URL" || {
    echo "❌ ダウンロードに失敗しました"
    echo "手動でダウンロードしてください: https://github.com/yukihamada/quiver/releases"
    exit 1
}

echo "📦 インストール中..."

# Extract
tar -xzf quiver.tar.gz

# Install to home directory
mkdir -p "$HOME/.quiver"
cp -r quiver/* "$HOME/.quiver/"

# Create desktop shortcut
cat > "$HOME/Desktop/QUIVer Provider.command" << 'EOF'
#!/bin/bash
cd "$HOME/.quiver"
open gui/index.html

# Start provider if not running
if ! pgrep -f "quiver-provider" > /dev/null; then
    ./scripts/start-network.sh
fi
EOF

chmod +x "$HOME/Desktop/QUIVer Provider.command"

# Create Applications link
if [ ! -e "/Applications/QUIVer Provider.app" ]; then
    ln -s "$HOME/.quiver/QUIVerProvider.app" "/Applications/QUIVer Provider.app" 2>/dev/null || true
fi

# Check Ollama
if ! command -v ollama &> /dev/null; then
    echo ""
    echo "⚠️  AI実行環境（Ollama）が必要です"
    echo ""
    echo "今すぐインストールしますか？ (y/n)"
    read -r response
    if [[ "$response" == "y" ]]; then
        echo "Ollamaをインストール中..."
        curl -fsSL https://ollama.ai/install.sh | sh
        
        # Install recommended model
        echo "推奨モデルをダウンロード中..."
        ollama pull llama3.2:3b
    fi
fi

# Cleanup
cd "$HOME"
rm -rf "$TEMP_DIR"

echo ""
echo "✅ インストール完了！"
echo ""
echo "起動方法:"
echo "  1. デスクトップの 'QUIVer Provider' をダブルクリック"
echo "  2. または Applications から 'QUIVer Provider' を起動"
echo ""
echo "予想収益:"
echo "  • Mac mini: 月 ¥60,000〜90,000"
echo "  • MacBook Pro: 月 ¥100,000〜150,000"
echo "  • Mac Studio: 月 ¥200,000〜300,000"
echo ""
echo "📊 ダッシュボード: http://localhost:8082"
echo "📱 スマホで確認: https://app.quiver.network"
echo ""

# Auto start
echo "今すぐ起動しますか？ (y/n)"
read -r response
if [[ "$response" == "y" ]]; then
    open "/Applications/QUIVer Provider.app" || open "$HOME/Desktop/QUIVer Provider.command"
fi