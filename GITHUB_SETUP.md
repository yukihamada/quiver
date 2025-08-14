# GitHub リポジトリ作成手順

## 1. GitHubでリポジトリ作成

1. https://github.com/new にアクセス
2. Repository name: `quiver`
3. Description: `QUIVer Provider - Earn $1,000+/Month with Your Mac | あなたのMacで月10万円稼ぐ`
4. Public を選択
5. **Initialize this repository with:** のチェックは全て外す
6. `Create repository` をクリック

## 2. ローカルからプッシュ

```bash
# リモートリポジトリを追加
git remote add origin https://github.com/yukihamada/quiver.git

# mainブランチにプッシュ
git push -u origin main
```

## 3. GitHub Pages を有効化

1. リポジトリの Settings → Pages
2. Source: Deploy from a branch
3. Branch: main → /website
4. Save

## 4. リリースを作成

```bash
# タグを作成
git tag -a v1.0.0 -m "Initial release"

# タグをプッシュ（自動的にリリースが作成される）
git push origin v1.0.0
```

## 5. 追加設定

### Topics を追加
- `macos`
- `cryptocurrency`
- `passive-income`
- `ai`
- `p2p`
- `quic`
- `golang`

### Social Preview
website/social-preview.png をアップロード（1280x640px推奨）

## 6. README バッジ更新

プッシュ後、以下のバッジが自動的に有効になります：
- License
- Go Version
- Platform

## 完了！

これで https://github.com/yukihamada/quiver からダウンロード可能になります。