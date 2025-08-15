# macOS Provider設定ガイド

## 🚀 クイックスタート（1分で完了）

```bash
# インストールスクリプトを実行
curl -fsSL https://raw.githubusercontent.com/yukihamada/quiver/main/scripts/install-mac-provider.sh | bash
```

## 📱 macOSアプリ版（GUI版）

既にビルド済みのアプリがあります：
1. `QUIVerProvider.dmg`をダウンロード
2. アプリケーションフォルダにドラッグ
3. 起動するだけ！

## 🔧 手動セットアップ

### 1. 必要なソフトウェア
```bash
# Ollamaをインストール
curl -fsSL https://ollama.ai/install.sh | sh

# モデルをダウンロード
ollama pull llama3.2:3b
```

### 2. Providerをビルド
```bash
# リポジトリをクローン
git clone https://github.com/yukihamada/quiver.git
cd quiver/provider

# ビルド
go build -o quiver-provider ./cmd/provider
```

### 3. 実行
```bash
# 環境変数を設定
export PROVIDER_LISTEN_ADDR="/ip4/0.0.0.0/tcp/4002"
export PROVIDER_DHT_BOOTSTRAP_PEERS="/ip4/35.221.85.1/tcp/4001/p2p/12D3KooWBBV3Sp2nCk47ygTCpCYfudkURmJUHK1Zmv33o4hCa991"
export PROVIDER_OLLAMA_URL="http://localhost:11434"

# 起動
./quiver-provider
```

## 📊 収益化の仕組み

1. **自動接続** - P2Pネットワークに自動参加
2. **推論処理** - LLMリクエストを処理
3. **レシート発行** - 処理ごとに署名付きレシート
4. **報酬獲得** - レシートを集約して報酬請求

## 🛡️ セキュリティ

- ローカルでLLMを実行（データ流出なし）
- Ed25519署名による認証
- レート制限でリソース保護

## 📈 パフォーマンス最適化

### M1/M2 Macの場合
```bash
# Metal GPUアクセラレーション有効
ollama run llama3.2:3b --gpu-layers 35
```

### メモリ設定
```bash
# Ollamaのメモリ制限を設定
export OLLAMA_MAX_LOADED_MODELS=1
export OLLAMA_NUM_PARALLEL=2
```

## 🔍 トラブルシューティング

### Provider起動を確認
```bash
# ログを確認
tail -f ~/.quiver/provider.log

# プロセスを確認
ps aux | grep quiver-provider

# P2P接続を確認
curl http://localhost:8091/health
```

### ファイアウォール設定
```bash
# macOSファイアウォールで許可
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --add /path/to/quiver-provider
```

## 🌐 ネットワーク参加状況

現在のネットワーク：
- Bootstrap: 35.221.85.1:4001
- Provider数: 増加中
- Gateway数: 3台稼働中

あなたのProviderも参加しましょう！