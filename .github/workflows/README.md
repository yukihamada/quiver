# GitHub Actions Workflows

QUIVerプロジェクトのCI/CDパイプライン。

## ワークフロー一覧

### 1. test.yml - テストとビルド
**トリガー**: プッシュ、プルリクエスト

- **test-go**: 各コンポーネント（gateway, provider, aggregator）のGoテスト
- **test-frontend**: HTMLバリデーション、リンクチェック
- **lint**: golangci-lint、go mod tidyチェック
- **security**: Trivy、gosecによるセキュリティスキャン
- **validate-dns**: DNS設定ファイルの検証
- **build**: 全コンポーネントのビルド
- **integration**: 統合テスト

### 2. deploy.yml - デプロイメント
**トリガー**: mainブランチへのプッシュ

- **deploy-pages**: Cloudflare Pagesへのフロントエンドデプロイ
- **deploy-workers**: Cloudflare Workersへのバックエンドデプロイ
- **release**: タグ付きリリースの作成（[release]コミットメッセージ時）
- **notify**: Discord通知

### 3. dns-sync.yml - DNS同期
**トリガー**: dns/ディレクトリの変更、手動実行

- **validate-dns**: DNS設定の妥当性チェック
- **sync-dns**: CloudflareへのDNSレコード同期
- **verify-dns**: DNS伝播の確認

## 必要なSecrets

GitHub Settings > Secrets and variables > Actions で設定：

```
CLOUDFLARE_API_TOKEN    # Cloudflare APIトークン
CLOUDFLARE_ACCOUNT_ID   # CloudflareアカウントID
CLOUDFLARE_ZONE_ID      # CloudflareゾーンID（quiver.network）
DISCORD_WEBHOOK         # Discord通知用Webhook URL（オプション）
```

## 使い方

### 通常の開発フロー

1. ブランチを作成して開発
2. プルリクエストを作成 → 自動でテスト実行
3. マージ → mainブランチに自動デプロイ

### DNS変更

1. `dns/records.yaml`を編集
2. プッシュ → 自動でCloudflareに同期

### リリース作成

コミットメッセージに`[release]`を含める：
```bash
git commit -m "[release] v1.2.0 - New features"
git push
```

### 手動デプロイ

GitHub Actions > workflow > Run workflow から手動実行可能。

## ローカルでのテスト

GitHub Actionsで実行されるテストをローカルで確認：

```bash
# Goテスト
cd gateway && go test -v -race ./...

# Lint
golangci-lint run

# DNS検証
python scripts/validate-dns.py

# HTMLチェック
pip install html5validator
html5validator --root docs/
```

## トラブルシューティング

### デプロイが失敗する
- Secretsが正しく設定されているか確認
- Cloudflare APIトークンの権限を確認
- wranglerのバージョンを確認

### テストが失敗する
- Go 1.23がインストールされているか確認
- 依存関係が最新か確認（go mod tidy）
- テスト環境の環境変数を確認

### DNS同期が失敗する
- dns/records.yamlの構文エラーをチェック
- CloudflareのAPI制限に達していないか確認