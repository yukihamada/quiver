# 🚀 QUIVer Provider ガイド

## プロバイダーになって収益を得る方法

### 📱 かんたん3ステップ

1. **インストール** (5分)
   ```bash
   curl -fsSL https://quiver.network/install.sh | sh
   ```

2. **起動**
   - デスクトップの「QUIVer Provider」をダブルクリック
   - または: `~/QUIVer/scripts/start-network.sh`

3. **収益確認**
   - ダッシュボード: http://localhost:8082
   - リアルタイムで収益を確認

## 💰 収益の仕組み

### 報酬レート
- **基本報酬**: 0.01 QUIV/トークン（約$0.001）
- **平均リクエスト**: 500トークン = 5 QUIV（約$0.50）
- **1日100リクエスト**: 500 QUIV（約$50）

### ボーナス報酬
| ボーナスタイプ | 追加報酬 | 条件 |
|--------------|----------|------|
| 高速応答 | +20% | 5秒以内 |
| 高可用性 | +10% | 99%以上稼働 |
| ステーキング | +5~30% | QUIV保有量 |

### 収益例

#### Mac Mini（趣味レベル）
- 1日8時間稼働
- 約400リクエスト処理
- **月収**: $600-1,200

#### MacBook Pro（副業レベル）
- 1日16時間稼働
- 約1,000リクエスト処理
- **月収**: $1,500-3,000

#### 専用マシン（本格運用）
- 24時間稼働
- GPU搭載
- **月収**: $5,000-15,000

## 🖥️ 必要なスペック

### 最小要件
- macOS 10.15+ または Ubuntu 20.04+
- 8GB RAM
- 20GB 空き容量
- インターネット接続

### 推奨スペック
- Apple Silicon Mac または GPU搭載PC
- 16GB+ RAM
- 100GB+ SSD
- 安定した高速回線

## 🛠️ セットアップ

### Mac用（推奨）
```bash
# 1. Ollamaインストール
curl -fsSL https://ollama.ai/install.sh | sh

# 2. モデルダウンロード
ollama pull llama3.2:3b
ollama pull qwen2.5:3b

# 3. QUIVerインストール
git clone https://github.com/quiver/quiver
cd quiver
./install-provider.sh
```

### Ubuntu/Debian
```bash
# 1. 依存関係
sudo apt update
sudo apt install curl git build-essential

# 2. Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# 3. QUIVer
git clone https://github.com/quiver/quiver
cd quiver
make build-all
./scripts/start-network.sh
```

## 📊 ダッシュボード機能

### リアルタイム監視
- 現在の収益
- 処理リクエスト数
- 平均応答時間
- ネットワーク接続状況

### 収益管理
- 自動請求（毎日）
- 手動請求
- 収益履歴
- ステーキング管理

### 設定
- モデル選択
- GPU使用ON/OFF
- 自動起動設定
- 優先度設定

## 🎯 収益最大化のコツ

### 1. 高速モデルを選ぶ
```
推奨モデル:
- llama3.2:3b (最速)
- phi3:mini
- qwen2.5:3b
```

### 2. 稼働時間を増やす
- 自動起動を有効化
- スリープ無効化
- UPS使用推奨

### 3. ステーキング
- Silver (10,000 QUIV) で+10%
- Gold (100,000 QUIV) で+20%

### 4. 複数マシン運用
- 異なるIPで複数ノード
- ロードバランシング
- 冗長性確保

## ❓ よくある質問

### Q: 電気代は？
A: Mac Mini M2で月額約$10-20。収益の2-3%程度。

### Q: 初期投資は？
A: 既存のMacがあれば$0。ステーキング用に1,000 QUIV（約$100）推奨。

### Q: リスクは？
A: 
- ハードウェア負荷は低い
- データは暗号化
- いつでも停止可能

### Q: 税金は？
A: 収益は課税対象。各国の仮想通貨税制に従ってください。

## 🆘 サポート

### コミュニティ
- Discord: https://discord.gg/quiver
- Twitter: @quivernetwork
- Reddit: r/quiver

### トラブルシューティング
```bash
# ログ確認
tail -f /tmp/quiver_logs/provider.log

# 再起動
./scripts/stop-network.sh
./scripts/start-network.sh

# 設定リセット
rm -rf ~/.quiver
./install-provider.sh
```

## 🚀 今すぐ始める

1. **収益計算機で確認**: [calculator.html](gui/calculator.html)
2. **インストール実行**: `./install-provider.sh`
3. **ダッシュボード確認**: http://localhost:8082

---

*あなたの余っているコンピューティングパワーを収益に変えましょう！*