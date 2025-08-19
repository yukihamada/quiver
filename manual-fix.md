# Cloudflare Pages 手動設定手順

## 現在の問題
- APIでのデプロイは成功するが、ファイルが正しくサーブされない
- GitHub Actionsは動作しているが、サイトにアクセスできない

## 解決方法

### 方法1: Cloudflareダッシュボードで直接設定

1. **Cloudflareダッシュボードにログイン**
   https://dash.cloudflare.com

2. **Workers & Pages → Overview**

3. **「Create」→「Pages」→「Connect to Git」**

4. **GitHubアカウントを接続**
   - GitHubで認証
   - リポジトリ「yukihamada/quiver」を選択

5. **ビルド設定**
   - Production branch: `main`
   - Build command: (空白)
   - Build output directory: `docs`

6. **環境変数** (必要な場合)
   - 特に必要なし

7. **「Save and Deploy」**

### 方法2: 既存プロジェクトの修正

1. **Workers & Pages → quiver-network-v2**

2. **Settings → Builds & deployments**
   - Framework preset: None
   - Build command: (空白)
   - Build output directory: `/docs` または `docs`

3. **Settings → Custom domains**
   - すべてのドメインが「Active」になっているか確認

### 方法3: wrangler CLIでローカルから直接デプロイ

```bash
# 1. wranglerをインストール
npm install -g wrangler

# 2. ログイン
wrangler login

# 3. デプロイ
cd /path/to/quiver
wrangler pages deploy docs --project-name=quiver-network-v2
```

## 確認事項

1. **DNS設定**
   - すべてのCNAMEが `quiver-network-v2.pages.dev` を指している
   - プロキシが有効（オレンジ色の雲）

2. **SSL設定**
   - SSL/TLS → Overview → Full

3. **カスタムドメイン**
   - すべてのドメインが「Active」状態

## トラブルシューティング

もし上記でも解決しない場合：

1. **新しいプロジェクトを作成**
   - プロジェクト名: `quiver-prod`
   - GitHubと直接接続

2. **Cloudflareサポートに連絡**
   - .network TLDの特殊な扱いについて確認

3. **代替案**
   - Vercel、Netlify、GitHub Pagesなどの他のホスティングサービスを検討