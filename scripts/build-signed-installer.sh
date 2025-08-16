#!/bin/bash
# QUIVer Provider 署名済みインストーラービルドスクリプト

set -e

echo "🔨 QUIVer Provider 署名済みインストーラーをビルド中..."

# 必要なディレクトリを作成
mkdir -p dist
mkdir -p build/package
mkdir -p build/scripts
mkdir -p build/Resources

# バイナリをビルド
echo "📦 バイナリをビルド中..."
cd /Users/yuki/QUIVer
make build-provider

# インストーラーパッケージの内容を準備
echo "📋 パッケージ内容を準備中..."
cp provider/bin/provider build/package/quiver-provider
cp -r docs/installer/macos/resources/* build/Resources/

# postinstallスクリプトを作成
cat > build/scripts/postinstall << 'EOF'
#!/bin/bash
# インストール後の設定

# バイナリを適切な場所に配置
mkdir -p /usr/local/bin
cp -f /tmp/quiver-provider /usr/local/bin/
chmod +x /usr/local/bin/quiver-provider

# LaunchAgentを設定
PLIST_PATH="$2/Library/LaunchAgents/network.quiver.provider.plist"
mkdir -p "$2/Library/LaunchAgents"

# ログディレクトリを作成
mkdir -p "$2/Library/Logs/QUIVerProvider"

cat > "$PLIST_PATH" << PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>network.quiver.provider</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/quiver-provider</string>
        <string>start</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$2/Library/Logs/QUIVerProvider/provider.log</string>
    <key>StandardErrorPath</key>
    <string>$2/Library/Logs/QUIVerProvider/provider.error.log</string>
</dict>
</plist>
PLIST

# サービスを開始
launchctl load "$PLIST_PATH"
launchctl start network.quiver.provider

exit 0
EOF

chmod +x build/scripts/postinstall

# PKGを作成（未署名）
echo "📦 PKGパッケージを作成中..."
pkgbuild \
    --root build/package \
    --identifier "network.quiver.provider" \
    --version "1.1.0" \
    --scripts build/scripts \
    --install-location /tmp \
    dist/QUIVerProvider-unsigned.pkg

# Developer IDで署名（証明書がある場合）
if [ ! -z "$DEVELOPER_ID" ]; then
    echo "🔏 パッケージに署名中..."
    echo "証明書: $DEVELOPER_ID"
    
    # PKGには Developer ID Installer 証明書が必要
    if echo "$DEVELOPER_ID" | grep -q "Developer ID Installer"; then
        SIGN_IDENTITY="$DEVELOPER_ID"
    elif echo "$DEVELOPER_ID" | grep -q "Developer ID Application"; then
        # Application証明書の場合、Installer証明書を探す
        INSTALLER_CERT=$(security find-identity -p basic -v | grep "Developer ID Installer" | grep "$(echo "$DEVELOPER_ID" | sed 's/.*(\(.*\))/\1/')" | head -1 | awk '{print $2}')
        if [ ! -z "$INSTALLER_CERT" ]; then
            SIGN_IDENTITY="$INSTALLER_CERT"
            echo "Installer証明書を使用: $SIGN_IDENTITY"
        else
            echo "⚠️  Developer ID Installer証明書が必要です"
            echo "https://developer.apple.com/account でInstaller証明書を作成してください"
            cp dist/QUIVerProvider-unsigned.pkg dist/QUIVerProvider.pkg
            exit 1
        fi
    else
        SIGN_IDENTITY="$DEVELOPER_ID"
    fi
    
    productsign \
        --sign "$SIGN_IDENTITY" \
        dist/QUIVerProvider-unsigned.pkg \
        dist/QUIVerProvider.pkg
    
    # 公証化（Apple IDがある場合）
    if [ ! -z "$APPLE_ID" ] && [ ! -z "$APP_PASSWORD" ] && [ ! -z "$APPLE_TEAM_ID" ]; then
        echo "🍎 Appleに公証を申請中..."
        xcrun notarytool submit dist/QUIVerProvider.pkg \
            --apple-id "$APPLE_ID" \
            --password "$APP_PASSWORD" \
            --team-id "$APPLE_TEAM_ID" \
            --wait
        
        echo "📎 公証をステープル中..."
        xcrun stapler staple dist/QUIVerProvider.pkg
        
        echo "✅ 署名と公証が完了しました！"
    else
        echo "⚠️  公証化には以下の環境変数が必要です:"
        echo "   APPLE_ID, APP_PASSWORD, APPLE_TEAM_ID"
        echo "   .env.signing.example を参考に設定してください"
    fi
else
    echo "⚠️  Developer ID証明書が設定されていません"
    echo "   以下を実行してください:"
    echo "   1. ./scripts/download-certificates.sh (証明書の作成)"
    echo "   2. cp .env.signing.example .env.signing"
    echo "   3. .env.signing を編集"
    echo "   4. source .env.signing && ./scripts/build-signed-installer.sh"
    cp dist/QUIVerProvider-unsigned.pkg dist/QUIVerProvider.pkg
fi

# クリーンアップ
rm -rf build
rm -f dist/QUIVerProvider-unsigned.pkg

echo "✅ インストーラーの作成が完了しました！"
echo "📍 場所: dist/QUIVerProvider.pkg"

# ワンクリックインストーラーのウェブページも生成
echo "🌐 ウェブインストーラーを生成中..."
cp docs/installer/web/index.html dist/install.html

echo ""
echo "🚀 配布方法:"
echo "1. ウェブサイトに dist/QUIVerProvider.pkg をアップロード"
echo "2. dist/install.html をウェブサイトに配置"
echo "3. ユーザーは https://quiver.network/install にアクセスしてワンクリックインストール"