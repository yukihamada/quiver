# QUIVer Provider - あなたのMacで月10万円稼ぐ

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/Platform-macOS-lightgrey.svg)](https://www.apple.com/macos/)

<p align="center">
  <img src="https://github.com/yukihamada/quiver/assets/123456789/demo.gif" alt="QUIVer Demo" width="600">
</p>

## 🚀 3分で始める

### 方法1: DMGインストーラー（推奨）

1. **ダウンロード**
   ```
   https://github.com/yukihamada/quiver/releases/latest
   ```
   から `QUIVerProvider-1.0.0.dmg` をダウンロード

2. **インストール**
   - DMGを開く
   - QUIVer Provider を Applications にドラッグ

3. **起動**
   - Applications から QUIVer Provider をダブルクリック
   - 初回起動時に自動セットアップ

### 方法2: コマンドライン

```bash
# クイックインストール
curl -fsSL https://raw.githubusercontent.com/yukihamada/quiver/main/install.sh | bash

# または手動インストール
git clone https://github.com/yukihamada/quiver.git
cd quiver
./install-provider.sh
```

## 💰 収益シミュレーター

| デバイス | 1日の収益 | 月収 | 年収 |
|---------|----------|------|------|
| Mac mini M2 | ¥3,000 | ¥90,000 | ¥1,080,000 |
| MacBook Pro M3 | ¥4,500 | ¥135,000 | ¥1,620,000 |
| Mac Studio M2 Ultra | ¥8,000 | ¥240,000 | ¥2,880,000 |

[収益計算機を試す →](https://yukihamada.github.io/quiver/calculator.html)

## 📱 スマホで収益確認

<p align="center">
  <img src="https://github.com/yukihamada/quiver/assets/123456789/mobile-dashboard.png" alt="Mobile Dashboard" width="300">
</p>

iPhoneやAndroidから収益をリアルタイムで確認：
- https://app.quiver.network
- またはQRコードをスキャン

## ⚡ 特徴

- **完全自動**: 一度設定すれば24時間365日稼働
- **省電力**: Mac miniなら月1,000円の電気代
- **安全**: あなたのデータには一切アクセスしません
- **透明性**: 収益はリアルタイムで確認可能
- **簡単停止**: いつでもワンクリックで停止

## 🛠 必要環境

- macOS 10.15 (Catalina) 以降
- メモリ 8GB 以上
- ストレージ 20GB 以上の空き
- インターネット接続

## 📊 ダッシュボード

<p align="center">
  <img src="https://github.com/yukihamada/quiver/assets/123456789/dashboard.png" alt="Dashboard" width="800">
</p>

## 🔧 高度な設定

### 環境変数

```bash
# カスタムポート
export QUIVER_PORT=8082

# GPUを使用
export QUIVER_GPU=true

# 最大CPU使用率を設定
export QUIVER_MAX_CPU=80
```

### 複数Mac運用

```bash
# マスターノード
./scripts/start-network.sh --master

# ワーカーノード
./scripts/start-network.sh --worker --master-ip 192.168.1.100
```

## 💎 ステーキング報酬

| ティア | 必要QUIV | ボーナス | 特典 |
|--------|----------|----------|------|
| Bronze | 1,000 | +5% | 優先処理 |
| Silver | 10,000 | +10% | API アクセス |
| Gold | 100,000 | +20% | ガバナンス投票 |
| Platinum | 1,000,000 | +30% | VIPサポート |

## 🆘 トラブルシューティング

### プロバイダーが起動しない

```bash
# ログを確認
tail -f ~/.quiver/logs/provider.log

# リセット
rm -rf ~/.quiver
./install-provider.sh
```

### 収益が入らない

1. ネットワーク接続を確認
2. ファイアウォール設定を確認
3. Ollamaが起動しているか確認

## 📖 ドキュメント

- [インストールガイド](docs/installation.md)
- [収益の仕組み](docs/rewards.md)
- [セキュリティ](docs/security.md)
- [FAQ](docs/faq.md)

## 🤝 コミュニティ

- [Discord](https://discord.gg/quiver)
- [Twitter](https://twitter.com/quivernetwork)
- [Reddit](https://reddit.com/r/quiver)

## 📄 ライセンス

このプロジェクトは MIT ライセンスの下で公開されています。

---

<p align="center">
  Made with ❤️ by QUIVer Team
</p>