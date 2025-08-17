# Cloudflare API Token 設定ガイド

## 必要な権限を持つAPI Tokenの作成

### 1. Cloudflare Dashboardにログイン
https://dash.cloudflare.com/

### 2. API Tokenを作成
1. 右上のプロフィールアイコン → "My Profile"
2. 左メニューの "API Tokens"
3. "Create Token" ボタンをクリック

### 3. カスタムトークンの設定

**Token name:** `QUIVer GitHub Actions`

**Permissions:**
以下の権限をすべて追加してください：

#### Account権限
- **Account:Cloudflare Pages:Edit** ← 重要！
- Account:Account Settings:Read

#### Zone権限
- Zone:DNS:Edit
- Zone:Zone:Read
- Zone:Zone Settings:Read
- Zone:SSL and Certificates:Read

### 4. リソースの指定

**Account Resources:**
- Include → Your Account (08519319108846c5673d8dbf1a23f6a5)

**Zone Resources:**
- Include → Specific zone → quiver.network

### 5. トークンの作成と保存

1. "Continue to summary" → "Create Token"
2. 表示されたトークンをコピー（一度しか表示されません！）
3. 安全な場所に保存

### 6. GitHub Secretsの更新

```bash
# 新しいトークンでGitHub Secretを更新
gh secret set CLOUDFLARE_API_TOKEN --body 'YOUR_NEW_TOKEN_HERE'
```

## トークンのテスト

```bash
# トークンが正しく動作するかテスト
curl -X GET "https://api.cloudflare.com/client/v4/user/tokens/verify" \
  -H "Authorization: Bearer YOUR_NEW_TOKEN_HERE" \
  -H "Content-Type: application/json"
```

成功すると以下のようなレスポンスが返ります：
```json
{
  "result": {
    "id": "...",
    "status": "active"
  },
  "success": true
}
```

## トラブルシューティング

### "Authentication error"が出る場合
- Account:Cloudflare Pages:Edit 権限があることを確認
- トークンがアクティブであることを確認
- 正しいアカウントIDを使用していることを確認

### "Not found"エラーが出る場合
- Cloudflare Pagesプロジェクトが存在することを確認
- プロジェクト名が正しいことを確認