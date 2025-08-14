#!/bin/bash

echo "======================================"
echo "QUIVer Provider インストーラー"
echo "======================================"
echo ""

# Check if Ollama is installed
if ! command -v ollama &> /dev/null; then
    echo "⚠️  Ollamaがインストールされていません"
    echo ""
    echo "Ollamaをインストールしますか？ (y/n)"
    read -r response
    if [[ "$response" == "y" ]]; then
        echo "Ollamaをダウンロード中..."
        curl -fsSL https://ollama.ai/install.sh | sh
    else
        echo "OllamaなしではQUIVer Providerは動作しません"
        exit 1
    fi
fi

echo "✅ Ollama検出済み"

# Check for models
echo ""
echo "利用可能なモデルを確認中..."
MODELS=$(ollama list 2>/dev/null | tail -n +2 | awk '{print $1}')

if [ -z "$MODELS" ]; then
    echo "⚠️  モデルがインストールされていません"
    echo ""
    echo "推奨モデルをインストールしますか？ (y/n)"
    read -r response
    if [[ "$response" == "y" ]]; then
        echo "llama3.2:3b をインストール中..."
        ollama pull llama3.2:3b
    fi
else
    echo "✅ インストール済みモデル:"
    echo "$MODELS" | while read -r model; do
        echo "   - $model"
    done
fi

# Install QUIVer
echo ""
echo "QUIVer Providerをインストール中..."

# Create directories
mkdir -p ~/QUIVer
cp -r . ~/QUIVer/

# Create desktop shortcut
cat > ~/Desktop/QUIVer\ Provider.command << 'EOF'
#!/bin/bash
cd ~/QUIVer
open gui/index.html

# Start provider if not running
if ! pgrep -f "provider" > /dev/null; then
    ./scripts/start-network.sh
fi
EOF

chmod +x ~/Desktop/QUIVer\ Provider.command

echo ""
echo "✅ インストール完了！"
echo ""
echo "デスクトップの 'QUIVer Provider' をダブルクリックして開始してください"
echo ""
echo "予想収益:"
echo "  • カジュアル利用: 月額 $150-300"
echo "  • 24時間稼働: 月額 $1,000-3,000"
echo "  • GPU利用: 月額 $5,000+"
echo ""
echo "詳細: https://quiver.network"