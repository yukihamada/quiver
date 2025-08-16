# インストーラー移行ガイド

## 現在の状況

- **v1.0.0**: PKGインストーラー（推奨）
- **v1.1.0**: DMGとPKGの両方が存在

## 今後の方針

### PKGインストーラーに統一
1. **理由**:
   - ワンクリックインストール
   - 「壊れているため開けません」エラーを回避
   - 自動セットアップ（LaunchAgent設定含む）

2. **移行計画**:
   - v1.2.0以降はPKGのみ提供
   - DMGは廃止

### ダウンロードURL
- 常に最新版: `https://github.com/yukihamada/quiver/releases/latest/download/QUIVerProvider.pkg`
- 特定バージョン: `https://github.com/yukihamada/quiver/releases/download/vX.X.X/QUIVerProvider.pkg`

### インストール方法
1. **ウェブ**: https://quiver.network/install
2. **コマンド**: `curl -fsSL https://quiver.network/install.sh | bash`