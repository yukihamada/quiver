# Cloudflare Pages セットアップガイド

## 事前準備

### 1. Wranglerのインストール
```bash
npm install -g wrangler
```

### 2. Cloudflareにログイン
```bash
wrangler login
```

### 3. CloudflareのAPI情報を取得

1. [Cloudflareダッシュボード](https://dash.cloudflare.com) にログイン
2. 右上のプロフィール → "My Profile" → "API Tokens"
3. "Create Token" → "Edit zone DNS" テンプレートを選択
4. 必要な権限:
   - Zone:DNS:Edit
   - Zone:Page Rules:Edit
   - Zone:SSL and Certificates:Edit
   - Account:Cloudflare Pages:Edit

### 4. Zone IDを取得

1. Cloudflareダッシュボードで quiver.network を選択
2. 右側のサイドバーから "Zone ID" をコピー

## 自動デプロイ

### 方法1: スクリプトを使用（推奨）

```bash
# 環境変数を設定
export CLOUDFLARE_API_TOKEN='your-api-token-here'
export CLOUDFLARE_ZONE_ID='your-zone-id-here'

# デプロイスクリプトを実行
cd /Users/yuki/QUIVer
./deploy-cloudflare.sh
```

### 方法2: 手動でデプロイ

```bash
# 1. Cloudflare Pagesにデプロイ
cd /Users/yuki/QUIVer
wrangler pages deploy docs --project-name=quiver-network

# 2. カスタムドメインを追加
wrangler pages project domains add quiver-network quiver.network

# 3. 各サブドメインを追加
wrangler pages project domains add quiver-network explorer.quiver.network
wrangler pages project domains add quiver-network api.quiver.network
wrangler pages project domains add quiver-network dashboard.quiver.network
wrangler pages project domains add quiver-network security.quiver.network
wrangler pages project domains add quiver-network quicpair.quiver.network
wrangler pages project domains add quiver-network playground.quiver.network
```

## Cloudflareダッシュボードでの設定

### DNS設定（自動スクリプトが失敗した場合）

1. DNS → Records に移動
2. 以下のレコードを追加:

```
Type    Name        Content                     Proxy
CNAME   @          quiver-network.pages.dev     ✓
CNAME   explorer   quiver-network.pages.dev     ✓
CNAME   api        quiver-network.pages.dev     ✓
CNAME   dashboard  quiver-network.pages.dev     ✓
CNAME   security   quiver-network.pages.dev     ✓
CNAME   quicpair   quiver-network.pages.dev     ✓
CNAME   playground quiver-network.pages.dev     ✓
CNAME   www        quiver-network.pages.dev     ✓
```

### Page Rules設定

Rules → Page Rules で以下を追加:

1. `explorer.quiver.network/*` → Forward URL (302) → `https://quiver.network/explorer/$1`
2. `api.quiver.network/*` → Forward URL (302) → `https://quiver.network/api/$1`
3. `dashboard.quiver.network/*` → Forward URL (302) → `https://quiver.network/dashboard/$1`
4. `security.quiver.network/*` → Forward URL (302) → `https://quiver.network/security/$1`
5. `quicpair.quiver.network/*` → Forward URL (302) → `https://quiver.network/quicpair/$1`
6. `playground.quiver.network/*` → Forward URL (302) → `https://quiver.network/playground-stream.html`

### SSL/TLS設定

1. SSL/TLS → Overview
2. "SSL/TLS encryption mode" を "Full" に設定
3. Edge Certificates → Always Use HTTPS を有効化

## 動作確認

```bash
# DNSの確認
dig explorer.quiver.network
dig api.quiver.network

# HTTPSアクセス確認
curl -I https://quiver.network
curl -I https://explorer.quiver.network
curl -I https://api.quiver.network
```

## トラブルシューティング

### "Project not found" エラー
```bash
# プロジェクトを作成
wrangler pages project create quiver-network
```

### DNS伝播の確認
```bash
# macOS/Linux
dig @1.1.1.1 explorer.quiver.network

# 複数のDNSサーバーで確認
dig @8.8.8.8 explorer.quiver.network
dig @1.1.1.1 explorer.quiver.network
```

### キャッシュのパージ
Cloudflareダッシュボード → Caching → Configuration → "Purge Everything"

## 継続的デプロイ

GitHubと連携する場合:
1. Cloudflare Pages → Create a project → Connect to Git
2. リポジトリを選択
3. Build settings:
   - Production branch: main
   - Build command: (空欄)
   - Build output directory: docs