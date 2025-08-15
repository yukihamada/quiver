# QUIVer ウェブサイトデプロイ手順

## 🚀 GitHub Pages設定

### 1. GitHub リポジトリ設定
1. https://github.com/yukihamada/quiver/settings/pages にアクセス
2. Source: "GitHub Actions" を選択
3. Save

### 2. ワークフロー実行
```bash
# 手動でワークフローをトリガー
gh workflow run deploy-website.yml
```

または、webフォルダに変更を加えてpush:
```bash
echo " " >> web/index.html
git add web/index.html
git commit -m "Trigger website deployment"
git push
```

### 3. デプロイ確認
- https://yukihamada.github.io/quiver/ でアクセス可能
- カスタムドメイン設定後: https://quiver.network/

## 🌐 カスタムドメイン設定（quiver.network）

### DNSレコード設定
```
Type: A
Name: @
Value: 185.199.108.153
       185.199.109.153
       185.199.110.153
       185.199.111.153

Type: CNAME
Name: www
Value: yukihamada.github.io
```

### SSL証明書
GitHub Pagesが自動的にLet's Encrypt証明書を発行

## 📝 ローカルテスト

```bash
# Python HTTP server
cd web
python3 -m http.server 8000

# ブラウザで確認
open http://localhost:8000
```

## 🔄 更新方法

1. webフォルダ内のファイルを編集
2. git add, commit, push
3. GitHub Actionsが自動的にデプロイ

## 📊 アクセス解析

GitHub Pages Insightsで確認:
https://github.com/yukihamada/quiver/pulse

## トラブルシューティング

### デプロイが反映されない
- GitHub Actions の実行状況を確認
- ブラウザキャッシュをクリア（Cmd+Shift+R）

### 404エラー
- .github/workflows/deploy-website.yml が正しく設定されているか確認
- GitHub Pages が有効になっているか確認