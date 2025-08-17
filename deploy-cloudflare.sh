#!/bin/bash

# Cloudflare Pages & DNS 自動設定スクリプト
# 事前準備: 
# 1. npm install -g wrangler
# 2. wrangler login
# 3. Cloudflare API Token を環境変数に設定

set -e

# 設定
DOMAIN="quiver.network"
PROJECT_NAME="quiver-network"
DOCS_DIR="docs"

# 色付きログ
log() { echo -e "\033[1;34m[INFO]\033[0m $1"; }
success() { echo -e "\033[1;32m[SUCCESS]\033[0m $1"; }
error() { echo -e "\033[1;31m[ERROR]\033[0m $1"; }

# Cloudflare API設定
if [ -z "$CLOUDFLARE_API_TOKEN" ]; then
    error "CLOUDFLARE_API_TOKEN が設定されていません"
    echo "以下を実行してください:"
    echo "export CLOUDFLARE_API_TOKEN='your-api-token'"
    exit 1
fi

if [ -z "$CLOUDFLARE_ZONE_ID" ]; then
    error "CLOUDFLARE_ZONE_ID が設定されていません"
    echo "以下を実行してください:"
    echo "export CLOUDFLARE_ZONE_ID='your-zone-id'"
    exit 1
fi

# Step 1: Cloudflare Pagesにデプロイ
log "Cloudflare Pagesにデプロイ中..."
cd /Users/yuki/QUIVer
wrangler pages deploy $DOCS_DIR --project-name=$PROJECT_NAME --branch=main

# デプロイURLを取得
PAGES_URL="${PROJECT_NAME}.pages.dev"
success "デプロイ完了: https://$PAGES_URL"

# Step 2: カスタムドメインを追加
log "カスタムドメインを設定中..."
wrangler pages project domains add $PROJECT_NAME $DOMAIN

# Step 3: DNS レコードを設定
log "DNSレコードを設定中..."

# API経由でCNAMEレコードを作成する関数
create_cname() {
    local subdomain=$1
    local target=$2
    
    log "CNAMEレコード作成: $subdomain.$DOMAIN → $target"
    
    curl -X POST "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records" \
         -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
         -H "Content-Type: application/json" \
         --data "{
           \"type\": \"CNAME\",
           \"name\": \"$subdomain\",
           \"content\": \"$target\",
           \"ttl\": 1,
           \"proxied\": true
         }" | jq .
}

# ルートドメインのCNAME（Cloudflareは自動的にCNAME Flatteningを行う）
create_cname "@" "$PAGES_URL"

# サブドメインの設定
SUBDOMAINS=(
    "explorer"
    "api"
    "dashboard"
    "security"
    "quicpair"
    "playground"
    "www"
)

for subdomain in "${SUBDOMAINS[@]}"; do
    create_cname "$subdomain" "$PAGES_URL"
done

# Step 4: Page Rules を設定（サブドメインのパス転送）
log "Page Rulesを設定中..."

create_page_rule() {
    local pattern=$1
    local forward_path=$2
    
    log "Page Rule作成: $pattern → $forward_path"
    
    curl -X POST "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/pagerules" \
         -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
         -H "Content-Type: application/json" \
         --data "{
           \"targets\": [{
             \"target\": \"url\",
             \"constraint\": {
               \"operator\": \"matches\",
               \"value\": \"$pattern\"
             }
           }],
           \"actions\": [{
             \"id\": \"forwarding_url\",
             \"value\": {
               \"url\": \"$forward_path\",
               \"status_code\": 302
             }
           }],
           \"priority\": 1,
           \"status\": \"active\"
         }" | jq .
}

# 各サブドメインのPage Rules
create_page_rule "explorer.$DOMAIN/*" "https://$DOMAIN/explorer/\$1"
create_page_rule "api.$DOMAIN/*" "https://$DOMAIN/api/\$1"
create_page_rule "dashboard.$DOMAIN/*" "https://$DOMAIN/dashboard/\$1"
create_page_rule "security.$DOMAIN/*" "https://$DOMAIN/security/\$1"
create_page_rule "quicpair.$DOMAIN/*" "https://$DOMAIN/quicpair/\$1"
create_page_rule "playground.$DOMAIN/*" "https://$DOMAIN/playground-stream.html"

# Step 5: SSL/TLS設定を確認
log "SSL/TLS設定を確認中..."
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/settings/ssl" \
     -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
     -H "Content-Type: application/json" \
     --data '{"value":"full"}' | jq .

# Step 6: キャッシュ設定
log "キャッシュ設定を最適化中..."
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/settings/browser_cache_ttl" \
     -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
     -H "Content-Type: application/json" \
     --data '{"value":14400}' | jq .

success "すべての設定が完了しました！"
echo ""
echo "確認URL:"
echo "- メインサイト: https://$DOMAIN"
echo "- Explorer: https://explorer.$DOMAIN"
echo "- API: https://api.$DOMAIN"
echo "- Dashboard: https://dashboard.$DOMAIN"
echo "- Security: https://security.$DOMAIN"
echo "- QuicPair: https://quicpair.$DOMAIN"
echo "- Playground: https://playground.$DOMAIN"
echo ""
echo "DNS伝播には最大48時間かかる場合があります。"