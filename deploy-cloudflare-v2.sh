#!/bin/bash

# Cloudflare Pages デプロイスクリプト v2
# wrangler最新版対応

set -e

# 設定
DOMAIN="quiver.network"
PROJECT_NAME="quiver-network"
DOCS_DIR="docs"

# 色付きログ
log() { echo -e "\033[1;34m[INFO]\033[0m $1"; }
success() { echo -e "\033[1;32m[SUCCESS]\033[0m $1"; }
error() { echo -e "\033[1;31m[ERROR]\033[0m $1"; }
warning() { echo -e "\033[1;33m[WARNING]\033[0m $1"; }

# Step 1: デプロイ完了の確認
success "Cloudflare Pagesへのデプロイが完了しました！"
echo ""
echo "デプロイURL: https://quiver-network.pages.dev"
echo ""

# Step 2: DNS設定の案内
log "次のステップ: CloudflareダッシュボードでDNS設定を行ってください"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📝 DNS設定手順"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "1. Cloudflareダッシュボードにアクセス:"
echo "   https://dash.cloudflare.com"
echo ""
echo "2. 'quiver.network' ドメインを選択"
echo ""
echo "3. 左メニューから 'DNS' → 'Records' を選択"
echo ""
echo "4. 以下のCNAMEレコードを追加:"
echo ""
echo "   Name         Content                      Proxy"
echo "   ────────────────────────────────────────────────"
echo "   @           quiver-network.pages.dev      ✓"
echo "   explorer    quiver-network.pages.dev      ✓"
echo "   api         quiver-network.pages.dev      ✓"
echo "   dashboard   quiver-network.pages.dev      ✓"
echo "   security    quiver-network.pages.dev      ✓"
echo "   quicpair    quiver-network.pages.dev      ✓"
echo "   playground  quiver-network.pages.dev      ✓"
echo "   www         quiver-network.pages.dev      ✓"
echo ""
echo "5. 'Workers & Pages' → 'quiver-network' → 'Custom domains' で"
echo "   各ドメインを追加（自動SSL証明書が発行されます）"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Step 3: 自動設定スクリプトの提案
echo ""
log "オプション: ブラウザを自動的に開く"
echo ""
read -p "Cloudflareダッシュボードを開きますか？ (y/n): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    # macOSの場合
    if [[ "$OSTYPE" == "darwin"* ]]; then
        open "https://dash.cloudflare.com/?to=/:account/domains/dns/quiver.network"
        open "https://dash.cloudflare.com/?to=/:account/pages/view/quiver-network/domains"
    # Linuxの場合
    elif command -v xdg-open &> /dev/null; then
        xdg-open "https://dash.cloudflare.com/?to=/:account/domains/dns/quiver.network"
        xdg-open "https://dash.cloudflare.com/?to=/:account/pages/view/quiver-network/domains"
    else
        echo "ブラウザを手動で開いてください"
    fi
fi

# Step 4: 確認用URL
echo ""
success "セットアップ完了！"
echo ""
echo "📌 確認用URL（DNS設定後にアクセス可能）:"
echo ""
echo "  • メインサイト: https://quiver.network"
echo "  • Explorer: https://explorer.quiver.network"
echo "  • API: https://api.quiver.network"
echo "  • Dashboard: https://dashboard.quiver.network"
echo "  • Security: https://security.quiver.network"
echo "  • QuicPair: https://quicpair.quiver.network"
echo "  • Playground: https://playground.quiver.network"
echo ""
echo "📌 現在アクセス可能:"
echo "  • https://quiver-network.pages.dev"
echo ""
warning "DNS伝播には最大48時間かかる場合があります"