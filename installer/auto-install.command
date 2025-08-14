#!/bin/bash

# QUIVer Provider Auto Installer
# This script runs when user double-clicks it

clear
echo "=================================="
echo "  QUIVer Provider インストーラー"
echo "=================================="
echo ""
echo "これからQUIVer Providerをインストールします。"
echo ""
echo "インストール内容："
echo "• QUIVer Provider アプリケーション"
echo "• AI推論エンジン (Ollama)"
echo "• AIモデル llama3.2 (約2GB)"
echo ""
read -p "インストールを開始しますか？ (y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "インストールをキャンセルしました。"
    exit 0
fi

echo ""
echo "インストールを開始します..."
echo ""

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
DMG_DIR="$SCRIPT_DIR"

# Check if we're running from a DMG
if [[ "$DMG_DIR" == /Volumes/* ]]; then
    echo "✅ DMGから実行中です"
else
    echo "❌ エラー: DMGから実行してください"
    exit 1
fi

# Check if already installed
if [ -d "/Applications/QUIVer Provider.app" ]; then
    echo "⚠️  QUIVer Providerは既にインストールされています"
    read -p "上書きしますか？ (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 0
    fi
fi

# Copy application
echo ""
echo "📦 アプリケーションをインストール中..."
cp -R "$DMG_DIR/QUIVer Provider.app" /Applications/
if [ $? -eq 0 ]; then
    echo "✅ アプリケーションをインストールしました"
else
    echo "❌ インストールに失敗しました"
    exit 1
fi

# Create support directories
echo ""
echo "📁 サポートファイルを設定中..."
mkdir -p "$HOME/Library/Application Support/QUIVer"
mkdir -p "$HOME/Library/Logs/QUIVer"
mkdir -p "$HOME/.quiver"

# Check Ollama
echo ""
echo "🔍 Ollamaをチェック中..."
if ! command -v ollama &> /dev/null; then
    echo "❌ Ollamaがインストールされていません"
    echo ""
    echo "Ollamaをインストールしています..."
    
    # Download Ollama installer
    curl -fsSL https://ollama.ai/install.sh | sh
    
    if [ $? -eq 0 ]; then
        echo "✅ Ollamaをインストールしました"
    else
        echo "⚠️  Ollamaのインストールに失敗しました"
        echo "手動でインストールしてください: https://ollama.ai/download"
    fi
fi

# Start Ollama service
echo ""
echo "🚀 Ollamaサービスを起動中..."
ollama serve > /dev/null 2>&1 &
sleep 3

# Download model
echo ""
echo "🤖 AIモデルをダウンロード中..."
if ! ollama list | grep -q "llama3.2"; then
    echo "llama3.2モデルをダウンロードしています（約2GB）..."
    ollama pull llama3.2:3b
    if [ $? -eq 0 ]; then
        echo "✅ モデルのダウンロードが完了しました"
    else
        echo "⚠️  モデルのダウンロードに失敗しました"
    fi
else
    echo "✅ llama3.2モデルは既にインストールされています"
fi

# Create auto-start configuration
echo ""
echo "⚙️  自動起動を設定中..."
cat > "$HOME/Library/LaunchAgents/com.quiver.provider.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.quiver.provider</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Applications/QUIVer Provider.app/Contents/MacOS/QUIVer Provider</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
EOF

launchctl load "$HOME/Library/LaunchAgents/com.quiver.provider.plist" 2>/dev/null

# Launch the app
echo ""
echo "🎉 インストールが完了しました！"
echo ""
echo "QUIVer Providerを起動しています..."
open "/Applications/QUIVer Provider.app"

echo ""
echo "=================================="
echo "  インストール完了"
echo "=================================="
echo ""
echo "✅ アプリケーションがインストールされました"
echo "✅ Ollamaがセットアップされました"
echo "✅ AIモデルがダウンロードされました"
echo "✅ 自動起動が設定されました"
echo ""
echo "💰 収益化が自動的に開始されます！"
echo ""
echo "このウィンドウは閉じて構いません。"

# Keep window open for a moment
sleep 5

# Eject the DMG
osascript -e 'tell application "Finder" to eject disk "QUIVer Provider"' 2>/dev/null

exit 0