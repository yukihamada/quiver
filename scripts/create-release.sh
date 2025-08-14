#!/bin/bash

# Create release package for QUIVer Provider

echo "Creating QUIVer Provider release package..."

VERSION="1.0.0"
RELEASE_DIR="release/quiver-provider-$VERSION"
ARCHIVE_NAME="quiver-provider-macos.tar.gz"

# Clean and create release directory
rm -rf release
mkdir -p "$RELEASE_DIR"

# Build all binaries
echo "Building binaries..."
make build-all

# Copy necessary files
echo "Copying files..."
cp -r provider/bin "$RELEASE_DIR/provider/"
cp -r gateway/bin "$RELEASE_DIR/gateway/"
cp -r aggregator/bin "$RELEASE_DIR/aggregator/"
cp -r gui "$RELEASE_DIR/"
cp -r scripts "$RELEASE_DIR/"
cp -r website "$RELEASE_DIR/"

# Copy documentation
cp README.md "$RELEASE_DIR/"
cp README_JP.md "$RELEASE_DIR/"
cp PROVIDER_GUIDE.md "$RELEASE_DIR/"
cp REWARDS.md "$RELEASE_DIR/"
cp USAGE.md "$RELEASE_DIR/"

# Create minimal bootstrap
mkdir -p "$RELEASE_DIR/bootstrap"
cat > "$RELEASE_DIR/bootstrap/start.sh" << 'EOF'
#!/bin/bash
cd "$(dirname "$0")/.."
./scripts/start-network.sh
EOF
chmod +x "$RELEASE_DIR/bootstrap/start.sh"

# Create app bundle
APP_DIR="$RELEASE_DIR/QUIVerProvider.app"
cp -r QUIVerProvider.app "$APP_DIR"

# Copy GUI launcher
cp gui/QUIVerLauncher.html "$APP_DIR/Contents/Resources/"

# Copy setup script
cp scripts/setup-mac-app.sh "$RELEASE_DIR/"

# Remove unnecessary files
find "$RELEASE_DIR" -name "*.go" -delete
find "$RELEASE_DIR" -name "*_test.go" -delete
find "$RELEASE_DIR" -name "go.mod" -delete
find "$RELEASE_DIR" -name "go.sum" -delete

# Create archive
echo "Creating archive..."
cd release
tar -czf "$ARCHIVE_NAME" "quiver-provider-$VERSION"

# Create DMG with installer
echo "Creating DMG with auto-installer..."

# Create DMG directory
DMG_DIR="dmg-contents"
rm -rf "$DMG_DIR"
mkdir -p "$DMG_DIR"

# Copy app and installer
cp -R "QUIVerProvider.app" "$DMG_DIR/"
cp -R "../installer/InstallQUIVer.app" "$DMG_DIR/" 2>/dev/null || echo "Installer app not found"
cp "../installer/auto-install.command" "$DMG_DIR/インストール.command"

# Create background and instructions
cat > "$DMG_DIR/インストール方法.txt" << 'INSTRUCTIONS'
QUIVer Provider インストール方法
================================

方法1（推奨）: 自動インストール
1. 「InstallQUIVer」アイコンをダブルクリック
2. 「インストール」ボタンをクリック
3. パスワードを入力（必要な場合）
4. 自動的にセットアップが完了します

方法2: 手動インストール
1. QUIVer Provider.appをApplicationsフォルダにドラッグ
2. Applicationsフォルダから起動

インストール後：
- 自動的に収益化が開始されます
- ダッシュボードで収益を確認できます
- いつでも停止可能です

サポート: https://github.com/yukihamada/quiver
INSTRUCTIONS

# Create DMG
hdiutil create -volname "QUIVer Provider" \
    -srcfolder "$DMG_DIR" \
    -ov -format UDZO \
    "QUIVerProvider-$VERSION-autoinstall.dmg"

# Clean up
rm -rf "$DMG_DIR"

echo ""
echo "✅ Release package created:"
echo "  - release/$ARCHIVE_NAME"
echo "  - release/QUIVerProvider-$VERSION.dmg"
echo ""
echo "Upload these files to GitHub releases"