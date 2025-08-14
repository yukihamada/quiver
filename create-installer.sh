#!/bin/bash

echo "================================"
echo "QUIVer Provider DMG作成スクリプト"
echo "================================"

# 一時ディレクトリ作成
TEMP_DIR="/tmp/QUIVerProvider"
DMG_NAME="QUIVerProvider-1.0.0.dmg"
APP_NAME="QUIVer Provider.app"

# クリーンアップ
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"

# アプリケーションバンドル作成
echo "アプリケーションバンドル作成中..."
APP_DIR="$TEMP_DIR/$APP_NAME"
mkdir -p "$APP_DIR/Contents/MacOS"
mkdir -p "$APP_DIR/Contents/Resources"

# Info.plist
cat > "$APP_DIR/Contents/Info.plist" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>QUIVerProvider</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>com.quiver.provider</string>
    <key>CFBundleName</key>
    <string>QUIVer Provider</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSAppleEventsUsageDescription</key>
    <string>QUIVer Provider needs to control Terminal to run background services.</string>
</dict>
</plist>
EOF

# 実行ファイル
cat > "$APP_DIR/Contents/MacOS/QUIVerProvider" << 'EOF'
#!/bin/bash

# QUIVer Provider Launcher
APP_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
INSTALL_DIR="$HOME/.quiver"

# 初回起動チェック
if [ ! -d "$INSTALL_DIR" ]; then
    # インストールダイアログ
    osascript -e 'display dialog "QUIVer Providerを初めて起動します。\n\n必要なコンポーネントをインストールしますか？" buttons {"キャンセル", "インストール"} default button "インストール" with icon note with title "QUIVer Provider"'
    
    if [ $? -eq 0 ]; then
        # プログレスバー表示
        osascript -e 'display notification "インストール中..." with title "QUIVer Provider"'
        
        # インストール実行
        mkdir -p "$INSTALL_DIR"
        cp -r "$APP_DIR/Contents/Resources/quiver/"* "$INSTALL_DIR/"
        
        # Ollamaチェック
        if ! command -v ollama &> /dev/null; then
            osascript -e 'display dialog "AIモデル実行環境(Ollama)が必要です。\n\nインストールページを開きますか？" buttons {"後で", "開く"} default button "開く"'
            if [ $? -eq 0 ]; then
                open "https://ollama.ai"
            fi
        fi
        
        osascript -e 'display notification "インストール完了！" with title "QUIVer Provider"'
    else
        exit 0
    fi
fi

# メインGUIを開く
open "$INSTALL_DIR/gui/index.html"

# バックグラウンドでプロバイダー起動
if ! pgrep -f "quiver-provider" > /dev/null; then
    cd "$INSTALL_DIR"
    nohup ./provider/bin/provider > "$HOME/.quiver.log" 2>&1 &
fi
EOF

chmod +x "$APP_DIR/Contents/MacOS/QUIVerProvider"

# リソースをコピー
echo "リソースをコピー中..."
mkdir -p "$APP_DIR/Contents/Resources/quiver"
cp -r provider "$APP_DIR/Contents/Resources/quiver/"
cp -r gateway "$APP_DIR/Contents/Resources/quiver/"
cp -r aggregator "$APP_DIR/Contents/Resources/quiver/"
cp -r gui "$APP_DIR/Contents/Resources/quiver/"
cp -r scripts "$APP_DIR/Contents/Resources/quiver/"
cp -r contracts "$APP_DIR/Contents/Resources/quiver/"

# アイコン作成（簡易版）
cat > "$APP_DIR/Contents/Resources/icon_generator.py" << 'EOF'
import os
from PIL import Image, ImageDraw, ImageFont
import subprocess

# 512x512のアイコン作成
size = 512
img = Image.new('RGBA', (size, size), (102, 126, 234, 255))
draw = ImageDraw.Draw(img)

# Qロゴを描画
circle_size = 300
x = (size - circle_size) // 2
y = (size - circle_size) // 2
draw.ellipse([x, y, x + circle_size, y + circle_size], fill=(255, 255, 255, 255))
draw.ellipse([x + 50, y + 50, x + circle_size - 50, y + circle_size - 50], fill=(102, 126, 234, 255))

# テキスト追加
try:
    font = ImageFont.truetype("/System/Library/Fonts/Helvetica.ttc", 180)
    draw.text((size//2, size//2), "Q", fill=(255, 255, 255, 255), font=font, anchor="mm")
except:
    pass

# 保存
img.save("/tmp/quiver_icon.png")

# icnsに変換
subprocess.run(["sips", "-s", "format", "icns", "/tmp/quiver_icon.png", "--out", "$APP_DIR/Contents/Resources/AppIcon.icns"])
EOF

# Pythonが利用可能な場合はアイコンを生成
if command -v python3 &> /dev/null && python3 -c "import PIL" 2>/dev/null; then
    python3 "$APP_DIR/Contents/Resources/icon_generator.py"
fi

# インストーラー用の背景画像とレイアウト
echo "DMGレイアウト作成中..."
mkdir -p "$TEMP_DIR/.background"

# READMEファイル
cat > "$TEMP_DIR/はじめにお読みください.txt" << 'EOF'
QUIVer Provider インストール方法
================================

1. 「QUIVer Provider」アプリをApplicationsフォルダにドラッグしてください
2. Applicationsフォルダから「QUIVer Provider」をダブルクリックして起動
3. 初回起動時に必要なコンポーネントが自動インストールされます

収益の目安
----------
• Mac mini: 月6〜10万円
• MacBook Pro: 月10〜15万円  
• Mac Studio: 月20万円以上

サポート
--------
Web: https://quiver.network
Discord: https://discord.gg/quiver
Email: support@quiver.network

EOF

# Applications へのシンボリックリンク
ln -s /Applications "$TEMP_DIR/Applications"

# DMG作成
echo "DMG作成中..."
hdiutil create -volname "QUIVer Provider" -srcfolder "$TEMP_DIR" -ov -format UDZO "$DMG_NAME"

# クリーンアップ
rm -rf "$TEMP_DIR"

echo ""
echo "✅ 作成完了: $DMG_NAME"
echo ""
echo "このファイルをWebサイトにアップロードしてください。"