# QUIVer DNS Management

このディレクトリはQUIVerのDNSレコードをGitで管理するためのものです。

## 仕組み

1. **dns/records.yaml** - すべてのDNSレコードの定義（Source of Truth）
2. **GitHub Actions** - プッシュ時に自動的にCloudflareと同期
3. **検証** - デプロイ前にレコードの妥当性をチェック
4. **確認** - デプロイ後にDNS伝播を確認

## DNS設定の変更方法

### 1. records.yaml を編集

```yaml
frontend:
  - name: new-subdomain
    type: CNAME
    content: quiver-network.pages.dev
    proxied: true
```

### 2. ローカルで検証

```bash
python scripts/validate-dns.py
```

### 3. コミット＆プッシュ

```bash
git add dns/records.yaml
git commit -m "Add new-subdomain DNS record"
git push
```

GitHub Actionsが自動的に：
1. DNS設定を検証
2. Cloudflareに同期
3. DNS伝播を確認

## レコードタイプ

### Frontend (Cloudflare Pages)
すべて `quiver-network.pages.dev` を指すCNAMEレコード：
- quiver.network
- explorer.quiver.network
- api.quiver.network
- dashboard.quiver.network
- security.quiver.network
- quicpair.quiver.network
- playground.quiver.network

### Backend Services
実際のサービスがデプロイされたら更新：
- gateway.quiver.network → Gateway API Server
- signal.quiver.network → WebRTC Signaling Server
- registry.quiver.network → Provider Registry
- bootstrap.quiver.network → P2P Bootstrap Nodes

## 手動同期

必要に応じて手動で同期することも可能：

```bash
# 環境変数を設定
export CLOUDFLARE_API_TOKEN='your-token'
export CLOUDFLARE_ZONE_ID='your-zone-id'

# DNS同期を実行
python scripts/sync-dns.py

# DNS伝播を確認
python scripts/verify-dns.py
```

## セキュリティ

- CloudflareのAPIトークンはGitHub Secretsに保存
- 最小権限の原則（DNS編集権限のみ）
- すべての変更はGitで追跡可能

## トラブルシューティング

### DNS同期が失敗する
- APIトークンの権限を確認
- Zone IDが正しいか確認
- records.yamlの構文エラーをチェック

### DNS伝播が確認できない
- 最大48時間待つ
- 異なるDNSサーバーで確認（1.1.1.1, 8.8.8.8）
- Cloudflareダッシュボードで直接確認