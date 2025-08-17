# QUIVer DNS設定ガイド

## 現在の構成

すべてのサブドメインは `/Users/yuki/QUIVer/docs/` 配下の静的ファイルを配信します。

## DNS設定方法

### 1. Vercel を使う場合（推奨）

```bash
# Vercelにデプロイ
cd /Users/yuki/QUIVer/docs
vercel --prod

# カスタムドメインを追加
vercel domains add quiver.network
```

Vercelダッシュボードで以下を設定：

| サブドメイン | ディレクトリ | 説明 |
|------------|------------|------|
| quiver.network | / | メインサイト |
| explorer.quiver.network | /explorer | ネットワーク可視化 |
| api.quiver.network | /api | API仕様・ポータル |
| dashboard.quiver.network | /dashboard | Provider管理 |
| security.quiver.network | /security | セキュリティ情報 |
| quicpair.quiver.network | /quicpair | QuicPair統合 |
| playground.quiver.network | /playground-stream.html | WebRTCデモ |

### 2. Cloudflare Pages を使う場合

```bash
# Cloudflare Pagesにデプロイ
cd /Users/yuki/QUIVer
wrangler pages deploy docs --project-name=quiver-network
```

Cloudflareダッシュボードで：
1. カスタムドメインを追加
2. 各サブドメインのCNAMEレコードを設定

### 3. Netlify を使う場合

```toml
# netlify.toml
[build]
  publish = "docs"

[[redirects]]
  from = "https://explorer.quiver.network/*"
  to = "/explorer/:splat"
  status = 200

[[redirects]]
  from = "https://api.quiver.network/*"
  to = "/api/:splat"
  status = 200

[[redirects]]
  from = "https://dashboard.quiver.network/*"
  to = "/dashboard/:splat"
  status = 200

[[redirects]]
  from = "https://security.quiver.network/*"
  to = "/security/:splat"
  status = 200

[[redirects]]
  from = "https://quicpair.quiver.network/*"
  to = "/quicpair/:splat"
  status = 200
```

### 4. GitHub Pages + Cloudflare を使う場合

1. GitHubリポジトリの Settings > Pages で docs フォルダを公開
2. CloudflareでDNSを管理：

```
# Aレコード（GitHubのIP）
quiver.network          A     185.199.108.153
quiver.network          A     185.199.109.153
quiver.network          A     185.199.110.153
quiver.network          A     185.199.111.153

# CNAMEレコード（サブドメイン）
explorer                CNAME  yukihamada.github.io
api                     CNAME  yukihamada.github.io
dashboard               CNAME  yukihamada.github.io
security                CNAME  yukihamada.github.io
quicpair                CNAME  yukihamada.github.io
```

3. Page Rulesでパス転送：
```
explorer.quiver.network/* → quiver.network/explorer/$1
api.quiver.network/* → quiver.network/api/$1
dashboard.quiver.network/* → quiver.network/dashboard/$1
security.quiver.network/* → quiver.network/security/$1
quicpair.quiver.network/* → quiver.network/quicpair/$1
```

### 5. 独自サーバー（Nginx）の場合

```nginx
# /etc/nginx/sites-available/quiver.network

# メインサイト
server {
    listen 443 ssl http2;
    server_name quiver.network;
    root /var/www/quiver/docs;
    
    ssl_certificate /etc/letsencrypt/live/quiver.network/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/quiver.network/privkey.pem;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
}

# Explorer
server {
    listen 443 ssl http2;
    server_name explorer.quiver.network;
    root /var/www/quiver/docs/explorer;
    
    ssl_certificate /etc/letsencrypt/live/quiver.network/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/quiver.network/privkey.pem;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
}

# API
server {
    listen 443 ssl http2;
    server_name api.quiver.network;
    root /var/www/quiver/docs/api;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    # OpenAPI仕様へのアクセス
    location /openapi.yaml {
        add_header Access-Control-Allow-Origin *;
    }
}

# Dashboard
server {
    listen 443 ssl http2;
    server_name dashboard.quiver.network;
    root /var/www/quiver/docs/dashboard;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
}

# Security
server {
    listen 443 ssl http2;
    server_name security.quiver.network;
    root /var/www/quiver/docs/security;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    # .well-knownの設定
    location /.well-known/security.txt {
        alias /var/www/quiver/docs/.well-known/security.txt;
    }
}

# QuicPair
server {
    listen 443 ssl http2;
    server_name quicpair.quiver.network;
    root /var/www/quiver/docs/quicpair;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
}

# Playground (特殊ケース)
server {
    listen 443 ssl http2;
    server_name playground.quiver.network;
    root /var/www/quiver/docs;
    
    location / {
        rewrite ^/$ /playground-stream.html break;
        try_files $uri $uri/ =404;
    }
}
```

## SSL証明書の取得

```bash
# Let's Encryptを使用
sudo certbot certonly --nginx -d quiver.network -d "*.quiver.network"
```

## 動作確認

デプロイ後、以下のURLにアクセスして確認：

- https://quiver.network/ - メインサイト
- https://explorer.quiver.network/ - ネットワーク可視化
- https://api.quiver.network/ - API仕様
- https://dashboard.quiver.network/ - Provider管理
- https://security.quiver.network/ - セキュリティ情報
- https://quicpair.quiver.network/ - QuicPair統合
- https://playground.quiver.network/ - WebRTCデモ

## 推奨事項

1. **Vercel** が最も簡単で、自動的にサブドメインのルーティングを処理
2. **Cloudflare Pages** は無料枠が大きく、エッジでのキャッシュが優秀
3. **GitHub Pages + Cloudflare** は完全無料だが、設定がやや複雑
4. **独自サーバー** は最も柔軟だが、運用コストがかかる

## トラブルシューティング

### サブドメインが表示されない
- DNSの伝播待ち（最大48時間）
- ブラウザキャッシュをクリア
- `dig explorer.quiver.network` でDNS解決を確認

### HTTPSエラー
- SSL証明書がワイルドカード対応か確認
- 証明書の有効期限を確認

### 404エラー
- ディレクトリ構造が正しいか確認
- index.htmlが存在するか確認
- リダイレクト設定を確認