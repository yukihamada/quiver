# QUIVerProvider macOS クイック修正

## 最速の解決方法

### ターミナルで1行実行するだけ：

```bash
curl -fsSL https://quiver.network/fix-macos.sh | bash
```

### または手動で：

1. **右クリックで開く**
   - QUIVerProvider.appを右クリック
   - 「開く」を選択
   - 警告ダイアログで「開く」をクリック

2. **ターミナルで属性を削除**
   ```bash
   xattr -cr /Applications/QUIVerProvider.app
   open /Applications/QUIVerProvider.app
   ```

## なぜこのエラーが出るのか？

- Appleの開発者証明書で署名されていないため
- 将来的にApple Developer Programに参加して署名予定
- 現在はオープンソースプロジェクトとして無償配布

## 安全性について

- ソースコード公開: https://github.com/yukihamada/quiver
- ローカル実行のみ（外部にデータ送信しない）
- コミュニティによる監査済み