#!/bin/bash
# macでProviderをすぐテストするスクリプト

echo "🧪 QUIVer Provider クイックテスト"
echo "================================"

# Ollamaが起動しているか確認
if ! curl -s http://localhost:11434/api/tags > /dev/null; then
    echo "⚠️  Ollamaが起動していません。起動中..."
    ollama serve > /dev/null 2>&1 &
    sleep 3
fi

# モデルがあるか確認
if ! ollama list | grep -q "llama3.2:3b"; then
    echo "📦 テスト用モデルをダウンロード中..."
    ollama pull llama3.2:3b
fi

# Providerを一時的に起動
echo "🚀 Providerを起動中..."
cd "$(dirname "$0")/../provider"

# 環境変数設定
export PROVIDER_LISTEN_ADDR="/ip4/0.0.0.0/tcp/4002"
export PROVIDER_DHT_BOOTSTRAP_PEERS="/ip4/35.221.85.1/tcp/4001/p2p/12D3KooWBBV3Sp2nCk47ygTCpCYfudkURmJUHK1Zmv33o4hCa991"
export PROVIDER_OLLAMA_URL="http://localhost:11434"
export PROVIDER_METRICS_PORT="8091"

# ビルドして実行
if [ ! -f "./quiver-provider" ]; then
    echo "🔨 初回ビルド中..."
    go build -o quiver-provider ./cmd/provider
fi

echo ""
echo "✅ Provider起動完了!"
echo "📊 メトリクス: http://localhost:8091/metrics"
echo "🌐 P2Pネットワークに接続中..."
echo ""
echo "Ctrl+Cで終了"
echo ""

./quiver-provider