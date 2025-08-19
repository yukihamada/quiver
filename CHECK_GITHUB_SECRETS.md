# GitHub Secrets確認方法

## 1. GitHub Secretsの確認

1. https://github.com/yukihamada/quiver/settings/secrets/actions にアクセス
2. 以下の2つのSecretsが存在することを確認：
   - `CLOUDFLARE_API_TOKEN`
   - `CLOUDFLARE_ACCOUNT_ID`

## 2. もし設定されていない場合

### CLOUDFLARE_API_TOKEN
1. 「New repository secret」をクリック
2. Name: `CLOUDFLARE_API_TOKEN`
3. Secret: `新しいAPIトークンを生成してください`（古いトークンは削除済み）

### CLOUDFLARE_ACCOUNT_ID  
1. 「New repository secret」をクリック
2. Name: `CLOUDFLARE_ACCOUNT_ID`
3. Secret: `08519319108846c5673d8dbf1a23f6a5`

## 3. 新しいAPIトークンの生成方法

1. https://dash.cloudflare.com/profile/api-tokens にアクセス
2. 「Create Token」をクリック
3. 「Custom token」を選択
4. 以下の権限を付与：
   - Account: Cloudflare Pages:Edit
   - Zone: Zone:Read, DNS:Edit
5. 「Continue to summary」→「Create Token」
6. 生成されたトークンをGitHub Secretsに設定

## 4. 動作確認

設定後、以下のコマンドでテスト：

```bash
# ローカルでテスト（GitHub Secretsを環境変数に設定後）
export CLOUDFLARE_API_TOKEN="your-new-token"
export CLOUDFLARE_ACCOUNT_ID="08519319108846c5673d8dbf1a23f6a5"
npx wrangler pages deploy docs --project-name=quiver-network-v2
```

## 5. GitHub Actionsの再実行

1. https://github.com/yukihamada/quiver/actions
2. 失敗したワークフローを選択
3. 「Re-run jobs」→「Re-run failed jobs」

## 重要な注意事項

- APIトークンは絶対にコードにハードコードしない
- GitHub Secretsのみで管理する
- トークンが漏洩した場合は即座に無効化して新規作成