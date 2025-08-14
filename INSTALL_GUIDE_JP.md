# 🚀 QUIVer Provider インストールガイド

## 📥 ダウンロード

1. **DMGファイルをダウンロード**
   - [QUIVerProvider-1.0.0.dmg をダウンロード](https://github.com/yukihamada/quiver/releases/download/v1.0.0/QUIVerProvider-1.0.0.dmg)
   - ファイルサイズ: 約45MB

## 🔧 インストール手順

### 1. DMGファイルを開く
- ダウンロードしたDMGファイルをダブルクリック
- QUIVer Providerアイコンが表示されます

### 2. アプリケーションフォルダにドラッグ
- QUIVer ProviderアイコンをApplicationsフォルダにドラッグ&ドロップ
- コピーが完了するまで待ちます

### 3. 初回起動
- Applicationsフォルダから「QUIVer Provider」をダブルクリック
- 初回起動時は「開発元が未確認」の警告が出る場合があります
  - システム環境設定 → セキュリティとプライバシー → 「このまま開く」をクリック

## 🎯 自動セットアップ

アプリを起動すると、自動的にセットアップウィザードが開始されます：

### ステップ1: Ollamaのインストール
- AI推論エンジン「Ollama」が自動的にチェックされます
- 未インストールの場合は、ダウンロードページが開きます
- インストーラーの指示に従ってインストールしてください

### ステップ2: LLMモデルのダウンロード
- 推論に使用する「llama3.2」モデルが自動的にダウンロードされます
- ファイルサイズ: 約2GB
- ダウンロード時間: 5-10分程度（回線速度による）

### ステップ3: 収益化の開始
- セットアップが完了すると、自動的に収益化が開始されます
- ダッシュボードでリアルタイムの収益を確認できます

## 💰 収益確認

### ダッシュボード機能
- **本日の収益**: リアルタイムで更新される収益額
- **処理リクエスト数**: 処理したAI推論の回数
- **稼働時間**: プロバイダーが動作している時間
- **システム状態**: CPU/メモリ使用率、ネットワーク状態

### 収益の目安
- **Mac mini (M2)**: 月 5,000〜60,000円
- **MacBook Pro (M3)**: 月 10,000〜120,000円  
- **Mac Studio (M2 Ultra)**: 月 20,000〜200,000円

## ⚙️ 設定とカスタマイズ

### 自動起動の設定
インストール時に自動起動が設定されます。無効にしたい場合：
```bash
launchctl unload ~/Library/LaunchAgents/com.quiver.provider.plist
```

### CPU使用率の調整
デフォルトでは最大70%に制限されています。
設定ファイル: `~/.quiver/config.json`

### ログの確認
```bash
tail -f ~/Library/Logs/QUIVer/provider.log
```

## 🆘 トラブルシューティング

### Ollamaが起動しない
```bash
# 手動でOllamaを起動
ollama serve
```

### モデルがダウンロードできない
```bash
# 手動でモデルをダウンロード
ollama pull llama3.2:3b
```

### ネットワークに接続できない
- ファイアウォールの設定を確認
- ポート4001, 8082が開いているか確認

## 📞 サポート

- GitHub Issues: https://github.com/yukihamada/quiver/issues
- Discord: https://discord.gg/quiver
- Email: support@quiver.network

## 🎉 収益化を始めよう！

インストールが完了したら、あとは自動的に収益が発生します。
Macを使っていない時間も有効活用して、毎月の収入を増やしましょう！