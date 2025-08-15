#!/bin/bash

echo "QUIVer Gateway を外部からアクセス可能にするスクリプト"
echo ""

# Check if ngrok is installed
if ! command -v ngrok &> /dev/null; then
    echo "ngrok がインストールされていません。"
    echo "インストール方法:"
    echo "  brew install ngrok"
    echo ""
    echo "または https://ngrok.com/download からダウンロード"
    exit 1
fi

# Check if gateway is running
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "エラー: Gateway が起動していません。"
    echo "先に 'quiver start' を実行してください。"
    exit 1
fi

echo "Gateway を公開中..."
echo ""
echo "ngrok を起動します。表示される HTTPS URL を使用してください。"
echo "例: https://xxxx-xx-xx-xx-xx.ngrok.io"
echo ""
echo "GitHub Pages の playground-stream.html で使用する場合:"
echo "1. 表示された HTTPS URL をコピー"
echo "2. ブラウザのコンソールで以下を実行:"
echo "   localStorage.setItem('customEndpoint', 'https://YOUR-NGROK-URL.ngrok.io/generate')"
echo ""
echo "Ctrl+C で終了"
echo ""

# Start ngrok
ngrok http 8080