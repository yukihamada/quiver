# GitHub Secrets 設定手順

## 新しいAPIトークンで設定を更新

### 1. GitHub Secretsページにアクセス
https://github.com/yukihamada/quiver/settings/secrets/actions

### 2. 以下のSecretsを設定/更新

#### CLOUDFLARE_API_TOKEN
- Name: `CLOUDFLARE_API_TOKEN`
- Secret: `C5SQVzb2wALhmn6_bU4WjjX9o048hXTgUPMxuKto`

#### CLOUDFLARE_ACCOUNT_ID
- Name: `CLOUDFLARE_ACCOUNT_ID`
- Secret: `08519319108846c5673d8dbf1a23f6a5`

### 3. 設定方法
1. 既存のSecretがある場合：「Update」をクリック
2. 新規の場合：「New repository secret」をクリック
3. NameとSecretを入力して「Add secret」または「Update secret」

### 4. 設定後の確認
- GitHub Actions: https://github.com/yukihamada/quiver/actions
- 最新のワークフローを再実行またはpushで新規実行

## 確認された情報
- Zone ID: `a56354ca4082aa4640456f87304fde80`
- Account ID: `08519319108846c5673d8dbf1a23f6a5`
- DNS設定: フル（SSL）
- プラン: Free プラン