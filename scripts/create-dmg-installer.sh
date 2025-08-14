#!/bin/bash

# Create DMG with auto-installer

echo "Creating auto-install DMG..."

VERSION="1.0.0"
APP_NAME="QUIVer Provider"
DMG_NAME="QUIVerProvider-$VERSION.dmg"
VOLUME_NAME="QUIVer Provider Installer"
DMG_DIR="/tmp/quiver-dmg"
INSTALLER_DIR="$DMG_DIR"

# Clean up old files
rm -rf "$DMG_DIR"
rm -f "$DMG_NAME"

# Create DMG directory structure
mkdir -p "$DMG_DIR"

# Copy application
echo "Copying application..."
cp -R "QUIVerProvider.app" "$DMG_DIR/$APP_NAME.app"

# Copy auto-installer
echo "Adding auto-installer..."
cp installer/auto-install.command "$DMG_DIR/インストール.command"

# Create background image
echo "Creating background..."
mkdir -p "$DMG_DIR/.background"
cat > "$DMG_DIR/.background/background.html" << 'EOF'
<!DOCTYPE html>
<html>
<head>
<style>
body {
    font-family: -apple-system, BlinkMacSystemFont, sans-serif;
    background: linear-gradient(135deg, #0071e3 0%, #5e5ce6 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
    margin: 0;
    text-align: center;
}
.content {
    padding: 40px;
}
h1 {
    font-size: 48px;
    margin-bottom: 20px;
}
p {
    font-size: 24px;
    opacity: 0.9;
}
.arrow {
    font-size: 64px;
    margin: 40px 0;
    animation: bounce 2s infinite;
}
@keyframes bounce {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-20px); }
}
</style>
</head>
<body>
<div class="content">
    <h1>QUIVer Provider</h1>
    <p>インストールを開始するには</p>
    <div class="arrow">↓</div>
    <p>「インストール」をダブルクリック</p>
</div>
</body>
</html>
EOF

# Create README
cat > "$DMG_DIR/はじめにお読みください.txt" << 'EOF'
QUIVer Provider インストールガイド
================================

1. 「インストール.command」をダブルクリックしてください
2. 自動的にインストールが開始されます
3. パスワードを求められた場合は、Macのログインパスワードを入力してください

インストール内容：
- QUIVer Provider.app
- Ollama（AI推論エンジン）
- llama3.2モデル（約2GB）

インストール後：
- 自動的にアプリが起動します
- 収益化が自動的に開始されます
- メニューバーからいつでも停止できます

サポート：
https://github.com/yukihamada/quiver
EOF

# Create DS_Store for window settings
echo "Setting window properties..."
cat > "$DMG_DIR/.DS_Store_template" << 'EOF'
# DMG window settings
# This would normally be a binary file
# In production, use create-dmg tool for proper window settings
EOF

# Create DMG
echo "Building DMG..."
hdiutil create -volname "$VOLUME_NAME" \
    -srcfolder "$DMG_DIR" \
    -ov -format UDZO \
    -fs HFS+ \
    "release/$DMG_NAME"

# Add auto-open settings
echo "Adding auto-open settings..."
# This requires additional tools in production
# For now, we'll add instructions

# Clean up
rm -rf "$DMG_DIR"

echo ""
echo "✅ Auto-install DMG created: release/$DMG_NAME"
echo ""
echo "When users open this DMG:"
echo "1. They will see the installer"
echo "2. Double-clicking 'インストール.command' starts auto-installation"
echo "3. Everything is set up automatically"
echo ""