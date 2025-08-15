# DNS設定ガイド

QUIVer NetworkのDNS設定手順を説明します。

## 概要

QUIVer Networkでは、以下のドメインを使用しています：
- `quiver.network` - メインドメイン（GitHub Pages）
- `*.quiver.network` - 各種サービス用サブドメイン

## Cloudflare DNS設定

### 1. 基本設定

CloudflareでDNSを管理する場合の設定：

#### Aレコード
```
# GitHub Pages用
@           A       185.199.108.153
@           A       185.199.109.153
@           A       185.199.110.153
@           A       185.199.111.153

# Bootstrap nodes
bootstrap1  A       34.85.77.119
bootstrap2  A       34.85.5.83
bootstrap3  A       34.146.202.132

# Signaling server
signal      A       34.146.216.182
```

#### CNAMEレコード
```
www         CNAME   yukihamada.github.io
docs        CNAME   yukihamada.github.io
api         CNAME   quiver-global-lb.quiver.network
playground  CNAME   yukihamada.github.io
```

### 2. リージョン別エンドポイント

地理的に分散したエンドポイント：

```
# アジアリージョン
quiver-asia         A       34.117.120.147

# USリージョン  
quiver-us           A       <US_IP_ADDRESS>

# EUリージョン
quiver-eu           A       <EU_IP_ADDRESS>

# グローバルロードバランサー
quiver-global-lb    A       34.117.120.147
```

### 3. WebSocketエンドポイント

WebSocket接続用：

```
signal-asia    CNAME   signal.quiver.network
wt             A       34.146.216.182  # WebTransport
ws             A       34.146.97.202   # WebSocket Stats
```

## GitHub Pages設定

### 1. カスタムドメイン設定

1. GitHubリポジトリの Settings > Pages に移動
2. Custom domain に `quiver.network` を入力
3. Enforce HTTPS を有効化

### 2. CNAMEファイル

`/docs/CNAME` ファイルに以下を記載：
```
quiver.network
```

## SSL/TLS設定

### Cloudflare SSL設定

1. SSL/TLS > Overview で "Full (strict)" を選択
2. Edge Certificates で以下を有効化：
   - Always Use HTTPS
   - Automatic HTTPS Rewrites
   - Opportunistic Encryption

### GCP Load Balancer SSL

GCPロードバランサー用の管理SSL証明書：
```hcl
resource "google_compute_managed_ssl_certificate" "quiver_cert" {
  name = "quiver-cert"
  managed {
    domains = ["quiver.network", "*.quiver.network"]
  }
}
```

## 動作確認

### DNSレコードの確認
```bash
# Aレコード確認
dig quiver.network
dig bootstrap1.quiver.network

# CNAMEレコード確認  
dig www.quiver.network
dig api.quiver.network

# 全レコード確認
dig ANY quiver.network
```

### 接続テスト
```bash
# HTTP/HTTPS接続確認
curl -I https://quiver.network
curl -I https://api.quiver.network/health

# WebSocket接続確認
wscat -c wss://ws.quiver.network/ws
```

## トラブルシューティング

### DNS伝播の確認

DNSの変更は伝播に時間がかかります（最大48時間）：

```bash
# 複数のDNSサーバーで確認
dig @8.8.8.8 quiver.network
dig @1.1.1.1 quiver.network
dig @208.67.222.222 quiver.network
```

### SSL証明書エラー

- Cloudflareで "Full (strict)" 使用時は、オリジンサーバーに有効な証明書が必要
- Let's Encryptまたは自己署名証明書を設定

### CORSエラー

APIエンドポイントでCORSエラーが発生する場合：
- Cloudflare > Rules > Transform Rules でCORSヘッダーを追加
- またはオリジンサーバーでCORS設定

## 自動化スクリプト

`deploy/gcp/setup-cloudflare-dns.sh` を使用して自動設定が可能：

```bash
export CLOUDFLARE_API_TOKEN="your-api-token"
export CLOUDFLARE_ZONE_ID="your-zone-id"
./deploy/gcp/setup-cloudflare-dns.sh
```

## 関連ドキュメント

- [Cloudflare API Documentation](https://developers.cloudflare.com/api/)
- [GitHub Pages Custom Domain](https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site)
- [GCP Load Balancer SSL](https://cloud.google.com/load-balancing/docs/ssl-certificates)