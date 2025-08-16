# QUIVer Provider macOS インストールガイド

## 「壊れているため開けません」エラーの解決方法

macOSのセキュリティ機能により、未署名のアプリケーションがブロックされることがあります。以下の方法で解決できます。

### 方法1: 右クリックで開く（推奨）

1. ダウンロードしたDMGファイルをダブルクリックしてマウント
2. QUIVerProvider.appを**右クリック**（またはControlキーを押しながらクリック）
3. 「開く」を選択
4. 警告ダイアログで「開く」をクリック

### 方法2: セキュリティとプライバシー設定

1. アプリケーションを通常通りダブルクリック（エラーが出る）
2. システム設定 > プライバシーとセキュリティ を開く
3. 「セキュリティ」セクションで「"QUIVerProvider"は開発元を確認できないため...」の横にある「このまま開く」をクリック
4. パスワードを入力して許可

### 方法3: ターミナルから実行

```bash
# DMGをマウント
hdiutil attach ~/Downloads/QUIVerProvider-*.dmg

# アプリケーションフォルダにコピー
cp -R /Volumes/QUIVerProvider/QUIVerProvider.app /Applications/

# 隔離属性を削除
xattr -cr /Applications/QUIVerProvider.app

# アプリを開く
open /Applications/QUIVerProvider.app
```

### 方法4: Homebrewを使用（開発者向け）

```bash
# Homebrewでインストール
brew tap yukihamada/quiver
brew install quiver-provider

# 実行
quiver-provider start
```

## セキュリティについて

QUIVerProviderは以下の理由で安全です：

1. **オープンソース**: ソースコードは[GitHub](https://github.com/yukihamada/quiver)で公開
2. **ローカル実行**: すべての処理はローカルで実行され、機密データは送信されません
3. **暗号化通信**: P2P通信はQUICプロトコルで暗号化されています

## トラブルシューティング

### それでも開けない場合

1. macOSのバージョンを確認（10.15以降が必要）
2. Rosetta 2がインストールされているか確認（M1/M2/M3 Macの場合）
   ```bash
   softwareupdate --install-rosetta
   ```

3. ファイアウォール設定を確認
   - システム設定 > ネットワーク > ファイアウォール
   - 「着信接続をブロック」がオフになっているか確認

### ログを確認

```bash
# アプリケーションログを確認
tail -f ~/Library/Logs/QUIVerProvider/provider.log

# システムログを確認
log show --predicate 'process == "QUIVerProvider"' --last 1h
```

## お問い合わせ

問題が解決しない場合は、以下のチャンネルでサポートを受けられます：

- [GitHub Issues](https://github.com/yukihamada/quiver/issues)
- [Discord](https://discord.gg/quiver)
- メール: support@quiver.network