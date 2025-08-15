# QUIVer iOS App

高速P2P AIチャットアプリ - 様々なAIモデルに瞬時にアクセス

## 🚀 特徴

### 超高速P2P接続
- WebSocketベースの低遅延通信
- 最適なProviderを自動選択
- リアルタイムレイテンシ表示

### 豊富なモデル選択
- **Qwen3シリーズ**: 0.6B〜32B
- **GPT-OSS**: 20B/120B
- **Jan-Nano**: 長文対応（32K/128K）
- **Llama, Mistral等**: 人気モデル多数

### スマート機能
- 需要の高いモデルを自動的にサーバーに適応
- モデルの人気度をリアルタイム表示
- ネットワーク状況に応じた最適化

## 📱 画面構成

### 1. チャット画面
- モデル選択メニュー
- リアルタイムレスポンス
- レイテンシ表示
- 履歴管理

### 2. モデル一覧
- 人気モデルセクション
- カテゴリ別表示（汎用/コーディング/長文）
- 利用可能状況の表示
- 検索機能

### 3. ネットワーク状況
- P2P接続状態
- 利用可能なProvider一覧
- レイテンシモニタリング
- 地域別Provider表示

### 4. 設定
- パフォーマンス調整
- プライバシー設定
- キャッシュ管理

## 🏗️ アーキテクチャ

### P2P通信
```swift
// WebSocketで高速通信
WebSocket → Gateway → P2P Network → Providers
```

### モデル管理
```swift
// 人気度に基づいた自動配置
Popular Models → Priority Providers → Fast Response
```

## 🔧 技術仕様

- **言語**: Swift 5.9+
- **最小iOS**: 15.0
- **通信**: WebSocket (wss://)
- **UI**: SwiftUI
- **P2P**: libp2p互換プロトコル

## 📦 ビルド方法

```bash
# Xcodeで開く
open ios/QUIVerApp.xcodeproj

# または Swift Package Manager
cd ios
swift build
```

## 🚀 今後の機能

- [ ] オフラインモード
- [ ] 音声入力対応
- [ ] マルチモーダル（画像）対応
- [ ] Apple Siliconローカル実行

## 📄 ライセンス

MIT License