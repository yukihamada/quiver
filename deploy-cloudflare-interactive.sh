#!/bin/bash

# Cloudflare Pages インタラクティブデプロイスクリプト
# APIキー不要 - wrangler loginで認証

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

# Step 1: Wranglerがインストールされているか確認
if ! command -v wrangler &> /dev/null; then
    error "wranglerがインストールされていません"
    echo "以下のコマンドでインストールしてください:"
    echo "npm install -g wrangler"
    exit 1
fi

# Step 2: ログイン確認
log "Cloudflareアカウントにログインします..."
wrangler whoami &> /dev/null || {
    warning "ログインが必要です。ブラウザが開きます..."
    wrangler login
}

success "ログイン完了"

# Step 3: プロジェクトの作成または更新
log "Cloudflare Pagesプロジェクトを設定中..."

# プロジェクトが存在するか確認
if wrangler pages project list | grep -q "$PROJECT_NAME"; then
    log "既存のプロジェクトを更新します: $PROJECT_NAME"
else
    log "新しいプロジェクトを作成します: $PROJECT_NAME"
    wrangler pages project create "$PROJECT_NAME" --production-branch main
fi

# Step 4: デプロイ
log "ファイルをデプロイ中..."
cd /Users/yuki/QUIVer

# _redirectsファイルをdocsにコピー
cp _redirects docs/_redirects 2>/dev/null || true

# デプロイ実行
DEPLOY_OUTPUT=$(wrangler pages deploy "$DOCS_DIR" --project-name="$PROJECT_NAME" --branch=main)
echo "$DEPLOY_OUTPUT"

# デプロイURLを抽出
PAGES_URL=$(echo "$DEPLOY_OUTPUT" | grep -o 'https://[a-z0-9-]*.pages.dev' | head -1)
success "デプロイ完了: $PAGES_URL"

# Step 5: カスタムドメインを設定
log "カスタムドメインを設定中..."

# メインドメイン
log "メインドメインを追加: $DOMAIN"
wrangler pages project domains add "$PROJECT_NAME" "$DOMAIN" || warning "$DOMAIN は既に設定されています"

# サブドメイン
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
    log "サブドメインを追加: $subdomain.$DOMAIN"
    wrangler pages project domains add "$PROJECT_NAME" "$subdomain.$DOMAIN" || warning "$subdomain.$DOMAIN は既に設定されています"
done

# Step 6: 設定ファイルを生成
log "設定ファイルを生成中..."

cat > /Users/yuki/QUIVer/docs/_headers << EOF
/*
  X-Frame-Options: DENY
  X-Content-Type-Options: nosniff
  Referrer-Policy: strict-origin-when-cross-origin

/api/*
  Access-Control-Allow-Origin: *
  Access-Control-Allow-Methods: GET, POST, OPTIONS
  Access-Control-Allow-Headers: Content-Type, Authorization

/.well-known/*
  Access-Control-Allow-Origin: *
EOF

# Step 7: 最終デプロイ（_headersと_redirectsを含む）
log "設定ファイルを含めて再デプロイ中..."
wrangler pages deploy "$DOCS_DIR" --project-name="$PROJECT_NAME" --branch=main

# Step 8: DNS設定の案内
echo ""
success "デプロイが完了しました！"
echo ""
warning "次のステップ: Cloudflareダッシュボードで以下のDNS設定を行ってください"
echo ""
echo "1. https://dash.cloudflare.com にアクセス"
echo "2. $DOMAIN を選択"
echo "3. DNS → Records で以下を追加:"
echo ""
echo "   Type    Name        Content                          Proxy"
echo "   -----------------------------------------------------------"
echo "   CNAME   @          $PROJECT_NAME.pages.dev           ✓"
echo "   CNAME   explorer   $PROJECT_NAME.pages.dev           ✓"
echo "   CNAME   api        $PROJECT_NAME.pages.dev           ✓"
echo "   CNAME   dashboard  $PROJECT_NAME.pages.dev           ✓"
echo "   CNAME   security   $PROJECT_NAME.pages.dev           ✓"
echo "   CNAME   quicpair   $PROJECT_NAME.pages.dev           ✓"
echo "   CNAME   playground $PROJECT_NAME.pages.dev           ✓"
echo "   CNAME   www        $PROJECT_NAME.pages.dev           ✓"
echo ""
echo "4. SSL/TLS → Overview で 'Full' モードを選択"
echo ""
echo "確認URL:"
echo "- Cloudflare Pages: $PAGES_URL"
echo "- メインサイト: https://$DOMAIN (DNS設定後)"
echo "- Explorer: https://explorer.$DOMAIN"
echo "- API: https://api.$DOMAIN"
echo "- Dashboard: https://dashboard.$DOMAIN"
echo "- Security: https://security.$DOMAIN"
echo "- QuicPair: https://quicpair.$DOMAIN"
echo "- Playground: https://playground.$DOMAIN"
echo ""
echo "DNS設定後、伝播には最大48時間かかる場合があります。"

# オプション: DNS設定を自動化するPythonスクリプトを生成
cat > /Users/yuki/QUIVer/setup-dns.py << 'EOF'
#!/usr/bin/env python3
"""
Cloudflare DNS設定用スクリプト
使い方: python3 setup-dns.py

このスクリプトはブラウザを開いてDNS設定ページに移動し、
必要な設定をクリップボードにコピーします。
"""

import webbrowser
import subprocess
import time

def copy_to_clipboard(text):
    """テキストをクリップボードにコピー"""
    process = subprocess.Popen(['pbcopy'], stdin=subprocess.PIPE)
    process.communicate(text.encode())

def main():
    print("🔧 Cloudflare DNS設定ヘルパー")
    print("-" * 50)
    
    # ドメイン名
    domain = "quiver.network"
    project = "quiver-network"
    
    # DNS設定
    dns_records = f"""# 以下をCloudflare DNSに追加してください:

@ CNAME {project}.pages.dev (Proxied)
explorer CNAME {project}.pages.dev (Proxied)
api CNAME {project}.pages.dev (Proxied)
dashboard CNAME {project}.pages.dev (Proxied)
security CNAME {project}.pages.dev (Proxied)
quicpair CNAME {project}.pages.dev (Proxied)
playground CNAME {project}.pages.dev (Proxied)
www CNAME {project}.pages.dev (Proxied)"""
    
    print(dns_records)
    print("\n📋 DNS設定をクリップボードにコピーしました")
    copy_to_clipboard(dns_records)
    
    print("\n🌐 Cloudflareダッシュボードを開きます...")
    time.sleep(2)
    webbrowser.open(f"https://dash.cloudflare.com/?to=/:account/domains/dns/{domain}")
    
    print("\n✅ 完了！DNSページで設定を追加してください。")

if __name__ == "__main__":
    main()
EOF

chmod +x /Users/yuki/QUIVer/setup-dns.py
log "DNS設定ヘルパー: ./setup-dns.py"