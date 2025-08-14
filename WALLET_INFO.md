# QUIVer ウォレット情報

⚠️ **注意**: これはテスト専用のウォレットです。本番環境では絶対に使用しないでください。

## Ethereumウォレット（スマートコントラクト用）

- **アドレス**: `0x7c45F09f52F4df656eE30c47B735BFB3b747050A`
- **用途**: スマートコントラクトのデプロイ、トークン管理

### テストトークンの取得方法

1. **Polygon Amoy Faucet**
   - https://faucet.polygon.technology/
   - 上記アドレスをコピーして貼り付け
   - 「Send Me MATIC」をクリック
   - 約0.5 MATICが送られます

2. **確認方法**
   - https://amoy.polygonscan.com/
   - アドレスを検索して残高確認

## P2Pノード情報

- **Peer ID**: `12D3KooWM4xFk5e7rnvJanUffNcvPB6E4MnXsaKW6ddrLE3bg59s`
- **キーファイル**: `node.key`

### 使用方法

```bash
# Provider起動時
./bin/provider --key-file node.key

# Gateway起動時
./bin/gateway --key-file node.key
```

## スマートコントラクトのデプロイ

テストトークンを取得したら：

```bash
cd contracts

# Hardhatをインストール
npm install

# コントラクトをコンパイル
npm run compile

# Amoyテストネットにデプロイ
npm run deploy:amoy
```

## セキュリティ

- プライベートキーは`.env`ファイルに保存済み
- `.env`は`.gitignore`に追加済み（Gitにコミットされません）
- 本番環境では新しいキーを生成してください

## 次のステップ

1. ✅ ウォレット作成完了
2. 📋 テストトークンを取得（上記Faucetから）
3. 🚀 スマートコントラクトをデプロイ
4. 🌐 P2Pネットワークを起動
5. 💰 報酬の受け取り開始！