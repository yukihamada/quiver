# QUIVer ローカルセットアップガイド

QUIVerをローカルで動かしてP2P推論を実行する方法を説明します。

## 必要なもの

- Go 1.22以降
- Ollama（LLMモデル実行用）
- macOS または Linux

## セットアップ手順

### 1. Ollamaをインストール

```bash
# macOSの場合
brew install ollama

# Ollamaを起動
ollama serve
```

### 2. モデルをダウンロード

```bash
# llama3.2モデルをダウンロード
ollama pull llama3.2:3b
```

### 3. QUIVerをビルド

```bash
# リポジトリをクローン
git clone https://github.com/yukihamada/quiver.git
cd quiver

# ビルド
make build
```

### 4. ノードを起動

各ノードを別々のターミナルで起動します。

#### ターミナル1: ブートストラップノード

```bash
./bin/bootstrap --port 4001 --metrics-port 8090
```

#### ターミナル2: プロバイダーノード

```bash
# 環境変数でブートストラップノードを指定
export QUIVER_BOOTSTRAP=/ip4/127.0.0.1/tcp/4001/p2p/[BOOTSTRAP_PEER_ID]
export PROVIDER_METRICS_PORT=8091

./bin/provider
```

注: `[BOOTSTRAP_PEER_ID]`はブートストラップノード起動時に表示されるPeerIDに置き換えてください。

#### ターミナル3: ゲートウェイノード

```bash
# 環境変数でブートストラップノードを指定
export QUIVER_BOOTSTRAP=/ip4/127.0.0.1/tcp/4001/p2p/[BOOTSTRAP_PEER_ID]
export QUIVER_GATEWAY_PORT=8080

./bin/gateway
```

## API使用方法

### 推論リクエスト

```bash
curl -X POST http://localhost:8080/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "こんにちは、元気ですか？",
    "model": "llama3.2:3b",
    "token": "test-token"
  }'
```

### レスポンス例

```json
{
  "completion": "元気です！今日はどのようなお手伝いができますか？",
  "receipt": {
    "provider_id": "12D3KooW...",
    "timestamp": "2025-01-14T12:00:00Z",
    "signature": "..."
  }
}
```

## トラブルシューティング

### "No providers available"エラー

1. すべてのノードが起動しているか確認
2. ブートストラップノードのPeerIDが正しいか確認
3. Ollamaが起動しているか確認（`ollama list`でモデル確認）

### 接続できない場合

1. ファイアウォールの設定を確認
2. ポートが他のプロセスで使用されていないか確認

```bash
# ポート使用状況を確認
lsof -i :4001
lsof -i :8080
```

## プロダクション使用

実際のプロダクション環境では、GCPやAWSにデプロイされたゲートウェイを使用できます：

```bash
# 本番環境のゲートウェイ（例）
curl -X POST https://api.quiver.network/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -d '{
    "prompt": "Explain quantum computing",
    "model": "llama3.2:3b"
  }'
```

## 次のステップ

- [API詳細ドキュメント](./api-usage.md)
- [アーキテクチャ概要](./architecture.md)
- [デプロイメントガイド](../deploy/README.md)