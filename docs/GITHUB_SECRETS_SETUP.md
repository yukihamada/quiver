# GitHub Secrets セットアップガイド

このプロジェクトのGitHub Actionsを正常に動作させるために、以下のSecretsを設定する必要があります。

## 必要なSecrets

### 1. CLOUDFLARE_API_TOKEN
Cloudflare APIトークン。DNSレコードの管理に使用します。

**取得方法:**
1. [Cloudflare Dashboard](https://dash.cloudflare.com/) にログイン
2. My Profile → API Tokens
3. "Create Token" をクリック
4. "Custom token" を選択
5. 以下の権限を設定:
   - Zone:DNS:Edit
   - Zone:Zone:Read
6. Zone Resources: Include → Specific zone → quiver.network

### 2. CLOUDFLARE_ZONE_ID
quiver.networkドメインのZone ID。

**取得方法:**
1. Cloudflare Dashboard → quiver.network を選択
2. 右側のサイドバーの "Zone ID" をコピー

### 3. CLOUDFLARE_ACCOUNT_ID
CloudflareアカウントID（Pages用）。

**取得方法:**
1. Cloudflare Dashboard → 右上のアカウント名をクリック
2. "Account ID" をコピー

### 4. DISCORD_WEBHOOK (オプション)
デプロイ通知用のDiscord Webhook URL。

## GitHubでのSecrets設定方法

1. GitHubリポジトリページで "Settings" タブを開く
2. 左側メニューの "Secrets and variables" → "Actions" を選択
3. "New repository secret" をクリック
4. Name と Secret value を入力して "Add secret" をクリック

## 設定確認

すべてのSecretsが正しく設定されているか確認:

```bash
# GitHub CLIを使用
gh secret list
```

## トラブルシューティング

### DNS Sync が失敗する場合
- CLOUDFLARE_API_TOKEN の権限を確認
- CLOUDFLARE_ZONE_ID が正しいか確認

### Pages Deploy が失敗する場合
- CLOUDFLARE_ACCOUNT_ID が正しいか確認
- Cloudflare Pagesプロジェクトが存在するか確認

## セキュリティ注意事項

- Secretsは絶対に公開リポジトリにコミットしない
- APIトークンは最小限の権限で作成する
- 定期的にトークンをローテーションする