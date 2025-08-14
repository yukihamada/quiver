# QUIVer Provider インストールのトラブルシューティング

## 「開発元を検証できないため開けません」エラーの解決方法

### 方法1: 右クリックで開く（推奨）
1. 「インストール.command」を**右クリック**
2. 「開く」を選択
3. 警告ダイアログで「開く」をクリック

### 方法2: Controlキーを押しながら開く
1. **Controlキー**を押しながら「インストール.command」をクリック
2. 「開く」を選択
3. 警告ダイアログで「開く」をクリック

### 方法3: システム環境設定から許可
1. システム環境設定 → セキュリティとプライバシー
2. 「一般」タブ
3. 「このまま開く」をクリック

## よくある質問

### Q: なぜこの警告が表示されるのですか？
A: これはmacOSのセキュリティ機能（Gatekeeper）によるものです。App Store以外からダウンロードしたアプリに対する標準的な警告です。

### Q: 安全ですか？
A: はい、QUIVer Providerは安全です。ソースコードは[GitHub](https://github.com/yukihamada/quiver)で公開されており、誰でも確認できます。

### Q: インストール後の削除方法は？
A: 以下の手順で完全に削除できます：
1. `/Applications/QUIVer Provider.app` を削除
2. `~/Library/Application Support/QUIVer` を削除
3. `~/Library/LaunchAgents/com.quiver.provider.plist` を削除

### Q: Ollamaのインストールに失敗する
A: 手動でインストールしてください：
```bash
curl -fsSL https://ollama.ai/install.sh | sh
```

### Q: モデルのダウンロードが遅い
A: llama3.2モデルは約2GBあります。回線速度により5-20分かかる場合があります。

## サポート

問題が解決しない場合は、以下にお問い合わせください：
- GitHub Issues: https://github.com/yukihamada/quiver/issues
- Email: support@quiver.network