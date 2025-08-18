# GitHub Actions自動デプロイ設定

## 概要

GitHubにpushするたびに、自動的にCloudflare Pagesにデプロイされるように設定しました。

## セットアップ手順

### 1. GitHub Secretsの設定

以下のSecretsを設定する必要があります：

1. GitHubリポジトリページを開く
2. Settings → Secrets and variables → Actions
3. 「New repository secret」をクリック
4. 以下の2つのSecretを追加：

#### CLOUDFLARE_API_TOKEN
- Name: `CLOUDFLARE_API_TOKEN`
- Value: `8sFAg2aVWcYm5rLZ7NHPJwtx_KswmzH9U3GOpC4n`

#### CLOUDFLARE_ACCOUNT_ID
- Name: `CLOUDFLARE_ACCOUNT_ID`
- Value: `08519319108846c5673d8dbf1a23f6a5`

### 2. 自動デプロイの動作

#### トリガー条件
- `main`ブランチへのpush
- `docs/`ディレクトリ内のファイル変更
- 手動実行（Actions → Run workflow）

#### デプロイ先
- プロジェクト: `quiver-network-v2`
- ドメイン: `quiver.network`および全サブドメイン

### 3. デプロイの確認

#### GitHub Actionsで確認
1. Actions タブを開く
2. 「Deploy to Cloudflare Pages」ワークフローを選択
3. 実行状況を確認

#### 実際のサイトで確認
```bash
# メインドメイン
curl -I https://quiver.network/

# サブドメイン
curl -I https://api.quiver.network/
curl -I https://docs.quiver.network/
```

### 4. トラブルシューティング

#### デプロイが失敗する場合
1. GitHub Secretsが正しく設定されているか確認
2. Cloudflare API トークンの権限を確認
3. プロジェクト名が`quiver-network-v2`であることを確認

#### サイトが404を返す場合
1. デプロイが完了するまで5-10分待つ
2. DNSが正しく設定されているか確認
3. Cloudflareダッシュボードでカスタムドメインの状態を確認

## ワークフローファイル

### メインワークフロー
- `.github/workflows/deploy.yml` - 全体的なCI/CDパイプライン
- `.github/workflows/pages-deploy.yml` - Cloudflare Pages専用デプロイ

### 実行タイミング
1. コード変更をコミット
2. `git push origin main`
3. 自動的にGitHub Actionsが起動
4. Cloudflare Pagesにデプロイ
5. 全ドメインで公開

## 次のステップ

1. このドキュメントの手順に従ってGitHub Secretsを設定
2. `docs/`内のファイルを変更してコミット
3. pushして自動デプロイを確認
4. 数分後にサイトにアクセスして確認

## 参考リンク

- [Cloudflare Pages Documentation](https://developers.cloudflare.com/pages/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Wrangler CLI Documentation](https://developers.cloudflare.com/workers/wrangler/)