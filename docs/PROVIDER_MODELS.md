# QUIVer Provider モデルガイド

## 🚀 クイックスタート

環境変数でモデルを指定してインストール:

```bash
# 軽量モデル (2GB RAM)
export QUIVER_MODEL="qwen3:0.6b"
curl -fsSL https://quiver.network/install.sh | bash

# 標準モデル (8GB RAM)
export QUIVER_MODEL="qwen3:3b"
curl -fsSL https://quiver.network/install.sh | bash

# 高性能モデル (16GB RAM)
export QUIVER_MODEL="qwen3:7b"
curl -fsSL https://quiver.network/install.sh | bash
```

## 📊 モデル一覧

### 汎用モデル

| モデル | RAM要件 | 特徴 | インストールコマンド |
|--------|---------|------|---------------------|
| Qwen3 0.6B | 2GB | 超軽量・高速レスポンス | `export QUIVER_MODEL="qwen3:0.6b"` |
| Qwen3 3B | 8GB | バランス型・日常使用 | `export QUIVER_MODEL="qwen3:3b"` |
| Qwen3 7B | 16GB | 高性能・複雑なタスク | `export QUIVER_MODEL="qwen3:7b"` |
| Qwen3 14B | 32GB | プロフェッショナル | `export QUIVER_MODEL="qwen3:14b"` |
| Qwen3 32B | 64GB | エンタープライズ | `export QUIVER_MODEL="qwen3:32b"` |
| Llama 3.2 3B | 8GB | Meta公式・安定 | `export QUIVER_MODEL="llama3.2:3b"` |
| Mistral 7B | 16GB | 欧州製・効率的 | `export QUIVER_MODEL="mistral:7b"` |
| Phi-3 Mini | 4GB | Microsoft製・コンパクト | `export QUIVER_MODEL="phi3:mini"` |
| Gemma 2 9B | 20GB | Google製・高品質 | `export QUIVER_MODEL="gemma2:9b"` |

### コーディング特化

| モデル | RAM要件 | 特徴 | インストールコマンド |
|--------|---------|------|---------------------|
| Qwen3-Coder 30B | 64GB | コード生成特化 | `export QUIVER_MODEL="qwen3-coder:30b"` |

### 長文対応

| モデル | RAM要件 | コンテキスト | インストールコマンド |
|--------|---------|------------|---------------------|
| Jan-Nano 32K | 16GB | 32,768 tokens | `export QUIVER_MODEL="jan-nano:32k"` |
| Jan-Nano 128K | 32GB | 131,072 tokens | `export QUIVER_MODEL="jan-nano:128k"` |

### フラッグシップ

| モデル | RAM要件 | 特徴 | インストールコマンド |
|--------|---------|------|---------------------|
| GPT-OSS 20B | 48GB | ローカルGPT | `export QUIVER_MODEL="gpt-oss:20b"` |
| GPT-OSS 120B | 256GB | 最高性能 | `export QUIVER_MODEL="gpt-oss:120b"` |

## 💻 スペック別推奨設定

### エントリーレベル (8GB RAM)
```bash
# 最も軽量
export QUIVER_MODEL="qwen3:0.6b"

# バランス重視
export QUIVER_MODEL="llama3.2:3b"
```

### ミドルレンジ (16-32GB RAM)
```bash
# 汎用高性能
export QUIVER_MODEL="qwen3:7b"

# 長文対応
export QUIVER_MODEL="jan-nano:32k"
```

### ハイエンド (64GB+ RAM)
```bash
# コーディング特化
export QUIVER_MODEL="qwen3-coder:30b"

# 最高性能
export QUIVER_MODEL="qwen3:32b"
```

### サーバーグレード (256GB+ RAM)
```bash
# フラッグシップ
export QUIVER_MODEL="gpt-oss:120b"
```

## 🔧 モデル切り替え

インストール後もモデルを切り替え可能:

```bash
# 現在のモデルを確認
ollama list

# 新しいモデルをダウンロード
ollama pull qwen3:7b

# 設定ファイルを編集
nano ~/.quiver/config/provider.yaml
# model: "qwen3:7b" に変更

# Providerを再起動
systemctl restart quiver-provider  # Linux
launchctl restart com.quiver.provider  # macOS
```

## 📈 パフォーマンス最適化

### GPU利用 (NVIDIA)
```bash
# CUDA対応モデルを自動検出
export OLLAMA_CUDA_VISIBLE_DEVICES=0
```

### Apple Silicon最適化
```bash
# Metal GPU自動利用
# 特別な設定不要
```

### メモリ最適化
```bash
# 同時ロードモデル数を制限
export OLLAMA_MAX_LOADED_MODELS=1

# 並列処理数を調整
export OLLAMA_NUM_PARALLEL=2
```

## 🎯 用途別選択ガイド

### チャットボット・カスタマーサポート
- **推奨**: Qwen3 3B, Llama 3.2 3B
- **理由**: バランスの良い応答速度と品質

### コード生成・技術文書
- **推奨**: Qwen3-Coder 30B
- **理由**: プログラミング言語に特化

### 長文要約・分析
- **推奨**: Jan-Nano 32K/128K
- **理由**: 大量のコンテキストを処理可能

### 研究・高度な推論
- **推奨**: GPT-OSS 120B, Qwen3 32B
- **理由**: 最高レベルの推論能力

## 🚨 トラブルシューティング

### モデルダウンロードが遅い
```bash
# プロキシ設定
export HTTPS_PROXY=http://proxy.example.com:8080

# ダウンロード再開
ollama pull <model-name>
```

### メモリ不足エラー
```bash
# より小さいモデルに切り替え
export QUIVER_MODEL="qwen3:0.6b"
curl -fsSL https://quiver.network/install.sh | bash
```

### モデルが見つからない
```bash
# 利用可能なモデルを確認
ollama list

# モデル名を正確に指定
export QUIVER_MODEL="qwen3:3b"  # 正しい
# export QUIVER_MODEL="qwen3-3b"  # 間違い
```

## 📞 サポート

- Discord: https://discord.gg/quiver
- GitHub Issues: https://github.com/yukihamada/quiver/issues
- Email: support@quiver.network