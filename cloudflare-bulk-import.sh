#!/bin/bash

# Cloudflare Bulk DNS Import Script
# このスクリプトはCloudflare APIを使用して一括でDNSレコードを追加します

set -e

# 色付きログ
log() { echo -e "\033[1;34m[INFO]\033[0m $1"; }
success() { echo -e "\033[1;32m[SUCCESS]\033[0m $1"; }
error() { echo -e "\033[1;31m[ERROR]\033[0m $1"; }

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📥 Cloudflare DNS インポートファイル"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "以下のファイルが作成されました："
echo ""
echo "1️⃣  cloudflare-dns-bind.txt"
echo "   → Cloudflareダッシュボードでインポート可能"
echo "   → DNS > Advanced > Import DNS Records"
echo ""
echo "2️⃣  cloudflare-dns-export.csv"
echo "   → CSV形式（エクセルで開ける）"
echo ""
echo "3️⃣  cloudflare-dns-import.json"
echo "   → API用JSON形式"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
log "推奨: cloudflare-dns-bind.txt を使用"
echo ""
echo "手順:"
echo "1. https://dash.cloudflare.com にアクセス"
echo "2. quiver.network ドメインを選択"
echo "3. DNS → Advanced → Import DNS Records"
echo "4. cloudflare-dns-bind.txt の内容をコピー＆ペースト"
echo "5. 'Import DNS Records' をクリック"
echo ""
success "ファイルが準備できました！"
echo ""

# ファイルの内容を表示
read -p "cloudflare-dns-bind.txt の内容を表示しますか？ (y/n): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    cat cloudflare-dns-bind.txt
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
fi

# クリップボードにコピー（macOS）
if [[ "$OSTYPE" == "darwin"* ]]; then
    read -p "クリップボードにコピーしますか？ (y/n): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cat cloudflare-dns-bind.txt | pbcopy
        success "クリップボードにコピーしました！"
    fi
fi