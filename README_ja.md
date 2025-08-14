# QUIVer - P2P AI推論ネットワーク

世界中の余剰コンピューティングパワーを活用した分散型AI推論ネットワーク

## 特徴

- 🌐 **完全分散型**: 中央サーバー不要のP2P通信
- 💰 **超低価格**: GPT-4の1/25、Claudeの1/20の価格
- 🚀 **簡単導入**: ワンクリックでMacにインストール
- 🔒 **セキュア**: Ed25519暗号署名による検証

## クイックスタート（Mac）

### ワンラインインストール

```bash
curl -fsSL https://raw.githubusercontent.com/yukihamada/quiver/main/scripts/install-mac.sh | bash
```

### 起動

```bash
# P2Pネットワークを起動
quiver start

# 推論をテスト
quiver test

# ステータス確認
quiver status
```

## APIの使い方

### 基本的な推論リクエスト

```bash
curl -X POST http://localhost:8080/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "こんにちは、元気ですか？",
    "model": "llama3.2:3b",
    "token": "test-token"
  }'
```

### Pythonでの使用例

```python
import requests

def ask_quiver(prompt):
    response = requests.post(
        "http://localhost:8080/generate",
        json={
            "prompt": prompt,
            "model": "llama3.2:3b",
            "token": "test-token"
        }
    )
    return response.json()

# 使用例
result = ask_quiver("フィボナッチ数列を説明してください")
print(result['completion'])
```

## アーキテクチャ

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│   Gateway   │────▶│  Provider   │
│  (あなた)    │     │  (API接続)   │     │ (AI実行)    │
└─────────────┘     └─────────────┘     └─────────────┘
                           │                    │
                           ▼                    ▼
                    ┌─────────────┐     ┌─────────────┐
                    │  Bootstrap  │     │   Ollama    │
                    │ (P2P接続点) │     │ (LLMモデル) │
                    └─────────────┘     └─────────────┘
```

## 動作確認済みモデル

- llama3.2:3b (推奨)
- llama3.2:1b
- qwen2.5:3b
- gemma3:4b

## トラブルシューティング

### "No providers available"エラー

```bash
# ノードの状態を確認
quiver status

# ログを確認
quiver logs

# 再起動
quiver stop && quiver start
```

### Ollamaが起動しない

```bash
# Ollamaを手動で起動
ollama serve

# モデルをダウンロード
ollama pull llama3.2:3b
```

## 開発者向け

### ソースからビルド

```bash
git clone https://github.com/yukihamada/quiver.git
cd quiver
make build
```

### コンポーネント

- **Bootstrap Node**: P2P接続の初期接続点
- **Provider Node**: AI推論を実行するノード
- **Gateway Node**: HTTPからP2Pへの変換
- **HyperLogLog**: 効率的なノード数カウント

## コントリビューション

プルリクエストを歓迎します！

1. このリポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. コミット (`git commit -m 'Add amazing feature'`)
4. プッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

## ライセンス

MIT License - 詳細は[LICENSE](LICENSE)を参照

## リンク

- [英語版README](README.md)
- [APIドキュメント](docs/api-usage.md)
- [アーキテクチャ詳細](docs/architecture.md)
- [GitHub](https://github.com/yukihamada/quiver)