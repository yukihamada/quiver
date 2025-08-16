#!/bin/bash
# QUIVer Provider macOS アプリ修正スクリプト

echo "QUIVer Provider macOS アプリの修正を開始します..."

# DMGファイルを探す
DMG_FILE=$(find ~/Downloads -name "QUIVerProvider*.dmg" -type f | head -1)

if [ -z "$DMG_FILE" ]; then
    echo "エラー: QUIVerProvider DMGファイルが見つかりません"
    echo "~/Downloads フォルダにDMGファイルがあることを確認してください"
    exit 1
fi

echo "DMGファイルを見つけました: $DMG_FILE"

# DMGをマウント
echo "DMGをマウント中..."
MOUNT_OUTPUT=$(hdiutil attach "$DMG_FILE" 2>&1)
if [ $? -ne 0 ]; then
    echo "エラー: DMGのマウントに失敗しました"
    echo "$MOUNT_OUTPUT"
    exit 1
fi

# マウントポイントを取得
MOUNT_POINT=$(echo "$MOUNT_OUTPUT" | grep "/Volumes" | awk '{print $NF}')

# アプリケーションをコピー
echo "アプリケーションをインストール中..."
if [ -d "/Applications/QUIVerProvider.app" ]; then
    echo "既存のアプリケーションを削除中..."
    rm -rf "/Applications/QUIVerProvider.app"
fi

cp -R "$MOUNT_POINT/QUIVerProvider.app" /Applications/

# DMGをアンマウント
hdiutil detach "$MOUNT_POINT" -quiet

# 隔離属性を削除
echo "セキュリティ属性を修正中..."
xattr -cr /Applications/QUIVerProvider.app

# 実行権限を付与
chmod +x /Applications/QUIVerProvider.app/Contents/MacOS/*

echo "✅ 修正が完了しました！"
echo ""
echo "アプリケーションを開くには："
echo "1. Finderで アプリケーション フォルダを開く"
echo "2. QUIVerProvider を右クリック"
echo "3. 「開く」を選択"
echo ""
echo "または、ターミナルから直接実行："
echo "open /Applications/QUIVerProvider.app"