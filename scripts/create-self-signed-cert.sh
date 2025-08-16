#!/bin/bash
# 自己署名証明書作成スクリプト（無料）

set -e

echo "🔐 自己署名証明書の作成"
echo "====================="
echo ""

CERT_NAME="QUIVer Developer Certificate"

# 既存の証明書を確認
if security find-certificate -c "$CERT_NAME" >/dev/null 2>&1; then
    echo "⚠️  既に証明書が存在します: $CERT_NAME"
    echo "削除して新規作成しますか？ [y/N]"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        security delete-certificate -c "$CERT_NAME"
    else
        exit 0
    fi
fi

echo "📝 証明書情報を入力してください:"
echo ""
echo "組織名 (例: QUIVer Network):"
read -r ORG_NAME

echo "メールアドレス:"
read -r EMAIL

# 証明書作成
echo ""
echo "🔨 証明書を作成中..."

# OpenSSLで秘密鍵と証明書を生成
openssl req -x509 -newkey rsa:2048 -keyout /tmp/quiver-key.pem -out /tmp/quiver-cert.pem -days 365 -nodes \
    -subj "/CN=$CERT_NAME/O=$ORG_NAME/emailAddress=$EMAIL"

# PKCS12形式に変換
openssl pkcs12 -export -out /tmp/quiver-cert.p12 -inkey /tmp/quiver-key.pem -in /tmp/quiver-cert.pem \
    -name "$CERT_NAME" -passout pass:

# キーチェーンにインポート
security import /tmp/quiver-cert.p12 -k ~/Library/Keychains/login.keychain-db -P "" -T /usr/bin/codesign

# 一時ファイルを削除
rm -f /tmp/quiver-key.pem /tmp/quiver-cert.pem /tmp/quiver-cert.p12

echo "✅ 証明書が作成されました！"
echo ""
echo "🔐 証明書名: $CERT_NAME"
echo ""

# 信頼設定
echo "証明書を信頼済みに設定しますか？ [Y/n]"
read -r response
if [[ -z "$response" ]] || [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo "管理者パスワードが必要です..."
    security add-trusted-cert -d -r trustRoot -k ~/Library/Keychains/login.keychain-db "$CERT_NAME" || true
fi

echo ""
echo "📦 署名付きインストーラーをビルドするには:"
echo "./scripts/build-signed-installer.sh"
echo ""
echo "署名時に使用する証明書名:"
echo "export DEVELOPER_ID=\"$CERT_NAME\""
echo ""
echo "⚠️  注意: 自己署名証明書では、他のMacで「開発元が未確認」エラーが出ます"