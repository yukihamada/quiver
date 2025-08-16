# 🍎 Apple Developer証明書セットアップガイド

## 現在の状況
✅ Apple Developer Program加入済み  
❌ Developer ID Installer証明書が必要

## ステップ1: 証明書の作成

### A. CSR（証明書署名要求）の作成

1. **キーチェーンアクセス**を開く
2. メニューバー → **キーチェーンアクセス** → **証明書アシスタント** → **認証局に証明書を要求...**
3. 情報入力:
   - **メールアドレス**: Apple Developer アカウントのメール
   - **通称**: Yuki Hamada (または会社名)
   - **ディスクに保存**を選択
4. デスクトップに保存

### B. Apple Developerで証明書作成

1. https://developer.apple.com/account/resources/certificates/add を開く
2. **「Developer ID Installer」**を選択
3. 作成したCSRファイルをアップロード
4. 証明書をダウンロード
5. ダウンロードした証明書をダブルクリックしてインストール

## ステップ2: 必要な情報の収集

### A. Team ID
1. https://developer.apple.com/account を開く
2. **「Membership Details」**をクリック
3. **Team ID**をメモ（10文字の英数字）

### B. App用パスワード
1. https://appleid.apple.com/account/manage を開く
2. **セキュリティ** → **App用パスワード** → **「+」**
3. 名前: `QUIVer Notarization`
4. 生成されたパスワード（xxxx-xxxx-xxxx-xxxx）をメモ

## ステップ3: 設定ファイルの作成

```bash
# 設定ファイルをコピー
cp .env.signing.example .env.signing

# .env.signing を編集
nano .env.signing
```

以下の形式で入力:
```bash
export DEVELOPER_ID="Developer ID Installer: Yuki Hamada (TEAMID)"
export APPLE_ID="your@email.com"
export APPLE_TEAM_ID="TEAMID"
export APP_PASSWORD="xxxx-xxxx-xxxx-xxxx"
```

## ステップ4: 署名付きビルド

```bash
# 設定を読み込んでビルド
source .env.signing && ./scripts/build-signed-installer.sh
```

## 完了後の効果

✅ 「開発元が未確認」エラーが完全に消える  
✅ ユーザーは警告なしでインストール可能  
✅ プロフェッショナルな配布が実現

---

**準備ができたら以下のコマンドで確認:**
```bash
security find-identity -p basic -v | grep "Developer ID"
```