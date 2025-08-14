# プライベートキーの作成方法

## 1. Ethereumウォレット用プライベートキー（スマートコントラクト用）

### 方法1: MetaMaskを使う（推奨）

1. **MetaMaskをインストール**
   - Chrome/Firefox/Brave: [metamask.io](https://metamask.io)
   - iOS/Android: App Storeで「MetaMask」検索

2. **新しいウォレットを作成**
   - 「Create a new wallet」を選択
   - パスワード設定
   - シードフレーズを安全に保管

3. **プライベートキーをエクスポート**
   - MetaMaskで ⋮ → Account details → Export private key
   - パスワード入力してキーを表示
   - `0x`を除いた部分をコピー

### 方法2: コマンドラインで生成

```bash
# スクリプトを使う
./scripts/generate-keys.sh
# オプション1を選択

# または、OpenSSLを使う
openssl rand -hex 32
```

### 方法3: オンラインツール（テストネット専用）

**本番環境では絶対に使わないこと！**

- [vanity-eth.tk](https://vanity-eth.tk) - テスト用のみ
- [iancoleman.io/bip39](https://iancoleman.io/bip39/) - オフラインで使用

## 2. P2Pノード用プライベートキー

```bash
# スクリプトを使う
./scripts/generate-keys.sh
# オプション2を選択

# 生成されたキーを使う
export QUIVER_PRIVATE_KEY=node.key
./bin/provider --key-file node.key
```

## 3. テストネットトークンの取得

### Polygon Amoyテストネット

1. **ウォレットアドレスを取得**
   ```bash
   # MetaMaskまたは生成したキーからアドレスを確認
   ```

2. **Faucetでテストトークンを取得**
   - [Polygon Faucet](https://faucet.polygon.technology/)
   - [Alchemy Faucet](https://www.alchemy.com/faucets/polygon-amoy)
   - 1日0.5 MATIC取得可能

### Sepoliaテストネット

1. **Sepolia ETHを取得**
   - [Sepolia Faucet](https://sepoliafaucet.com/)
   - [Infura Faucet](https://www.infura.io/faucet/sepolia)

## セキュリティ注意事項

### ⚠️ 重要

1. **本番用のプライベートキー**
   - 絶対にGitHubにコミットしない
   - `.env`ファイルは`.gitignore`に追加
   - 安全な場所（ハードウェアウォレット等）に保管

2. **テスト用のプライベートキー**
   - テストネット専用のキーを使う
   - 本番環境のキーとは必ず分ける

3. **キーの保管方法**
   ```bash
   # 環境変数で設定（推奨）
   export PRIVATE_KEY=your_key_here
   
   # .envファイルで設定
   echo "PRIVATE_KEY=your_key_here" > contracts/.env
   
   # 絶対にやってはいけないこと
   git add .env  # ❌
   ```

## キーの使い方

### スマートコントラクトのデプロイ

```bash
cd contracts
cp .env.example .env

# .envファイルを編集
# PRIVATE_KEY=生成したキー（0xなし）

# デプロイ
npm run deploy:amoy
```

### P2Pノードの起動

```bash
# キーファイルを指定
./bin/provider --key-file node.key

# または環境変数
export QUIVER_PRIVATE_KEY=node.key
./bin/provider
```

## トラブルシューティング

### "insufficient funds"エラー

1. Faucetからテストトークンを取得
2. ウォレットアドレスが正しいか確認
3. ネットワーク設定を確認

### "invalid private key"エラー

1. キーが64文字（32バイト）か確認
2. `0x`プレフィックスを除去
3. 改行や空白が含まれていないか確認

### MetaMaskでAmoyネットワークが表示されない

```javascript
// カスタムネットワークを追加
Network Name: Polygon Amoy
RPC URL: https://rpc-amoy.polygon.technology
Chain ID: 80002
Currency Symbol: MATIC
Block Explorer: https://amoy.polygonscan.com
```