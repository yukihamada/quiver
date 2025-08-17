# Cloudflare Pages 手動セットアップガイド

## 手動でCloudflare Pagesプロジェクトを作成する方法

### 1. Cloudflare Dashboardにログイン
https://dash.cloudflare.com/

### 2. Workers & Pagesに移動
左メニューから「Workers & Pages」を選択

### 3. Pagesプロジェクトを作成
1. 「Create application」ボタンをクリック
2. 「Pages」タブを選択
3. 「Connect to Git」をクリック
4. GitHubアカウントを接続（未接続の場合）
5. リポジトリ「yukihamada/quiver」を選択

### 4. プロジェクト設定
- **Project name**: `quiver-network`
- **Production branch**: `main`
- **Build command**: （空欄のまま）
- **Build output directory**: `docs`

### 5. 環境変数（不要）
静的サイトなので環境変数は不要です。「Save and Deploy」をクリック。

### 6. カスタムドメインの追加
デプロイ完了後、プロジェクトの設定画面で：

1. 「Custom domains」タブを開く
2. 「Set up a custom domain」をクリック
3. 以下のドメインを1つずつ追加：
   - quiver.network
   - www.quiver.network
   - api.quiver.network
   - blog.quiver.network
   - community.quiver.network
   - dashboard.quiver.network
   - docs.quiver.network
   - explorer.quiver.network
   - playground.quiver.network
   - quicpair.quiver.network
   - security.quiver.network
   - status.quiver.network

### 7. SSL証明書の確認
各ドメインを追加後、自動的にSSL証明書が発行されます（数分かかります）。

### 8. 動作確認
すべてのドメインでグリーンのチェックマークが表示されたら、各URLにアクセスして確認：
- https://quiver.network
- https://api.quiver.network
- など

## トラブルシューティング

### 403エラーが表示される場合
- カスタムドメインが正しく追加されているか確認
- SSL証明書が発行完了しているか確認（最大15分かかる場合があります）

### ページが表示されない場合
- DNSレコードが正しく設定されているか確認
- DNS伝播を待つ（最大48時間、通常は数分）

## 自動デプロイの設定
GitHubと連携済みなので、`main`ブランチへのプッシュで自動的にデプロイされます。