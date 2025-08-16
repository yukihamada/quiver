#!/bin/bash
# Apple Developer証明書セットアップウィザード

echo "🍎 Apple Developer証明書セットアップウィザード"
echo "==========================================="
echo ""

# Team IDを取得
echo "📋 ステップ1: Team IDの確認"
echo "https://developer.apple.com/account にアクセスして"
echo "「Membership Details」でTeam IDを確認してください"
echo ""
echo "Team ID (10文字の英数字) を入力してください:"
read -r TEAM_ID

# Apple IDを取得
echo ""
echo "📧 ステップ2: Apple IDの入力"
echo "Apple Developer アカウントのメールアドレスを入力してください:"
read -r APPLE_ID

# 証明書の確認
echo ""
echo "🔐 ステップ3: 証明書の確認"
echo "現在インストールされている証明書:"
echo "-----------------------------------"
security find-identity -p basic -v | grep -E "(Developer ID|Apple Development)" || echo "証明書が見つかりません"
echo ""

# Developer ID Installer証明書の確認
if ! security find-identity -p basic -v | grep -q "Developer ID Installer"; then
    echo "⚠️  Developer ID Installer証明書が見つかりません"
    echo ""
    echo "証明書を作成しますか？ [Y/n]"
    read -r response
    
    if [[ -z "$response" ]] || [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        # CSR作成
        echo ""
        echo "📝 証明書署名要求(CSR)を作成します..."
        
        # CSRファイル名
        CSR_FILE="$HOME/Desktop/CertificateSigningRequest.certSigningRequest"
        
        # キーチェーンアクセスでCSRを作成するためのスクリプト
        cat > /tmp/create_csr.sh << 'EOF'
#!/bin/bash
osascript << 'END'
tell application "Keychain Access"
    activate
end tell

display dialog "キーチェーンアクセスで以下の手順を実行してください:

1. メニューバー → キーチェーンアクセス → 証明書アシスタント → 認証局に証明書を要求...
2. メールアドレス: Apple Developer アカウントのメール
3. 通称: あなたの名前または会社名
4. 「ディスクに保存」を選択
5. デスクトップに保存

完了したら「OK」をクリックしてください。" buttons {"OK"} default button "OK"
END
EOF
        chmod +x /tmp/create_csr.sh
        /tmp/create_csr.sh
        
        if [ -f "$CSR_FILE" ]; then
            echo "✅ CSRファイルが作成されました: $CSR_FILE"
            echo ""
            echo "次のステップ:"
            echo "1. https://developer.apple.com/account/resources/certificates/add を開く"
            echo "2. 「Developer ID Installer」を選択"
            echo "3. CSRファイルをアップロード"
            echo "4. 証明書をダウンロードしてダブルクリック"
            echo ""
            echo "証明書をインストールしたら、Enterを押してください..."
            read -r
        fi
    fi
fi

# App用パスワードの設定
echo ""
echo "🔑 ステップ4: App用パスワードの設定"
echo "公証化(Notarization)にはApp用パスワードが必要です"
echo ""
echo "1. https://appleid.apple.com/account/manage にアクセス"
echo "2. セキュリティ → App用パスワード → 「+」"
echo "3. 名前: 'QUIVer Notarization' など"
echo "4. 生成されたパスワード (xxxx-xxxx-xxxx-xxxx) をコピー"
echo ""
echo "App用パスワードを入力してください:"
read -s APP_PASSWORD
echo ""

# 証明書名の取得
echo "🎯 ステップ5: 証明書の選択"
echo "使用する証明書を選択してください:"
echo ""

# 証明書リストを配列に格納
IFS=$'\n'
CERTS=($(security find-identity -p basic -v | grep "Developer ID" | sed 's/^[[:space:]]*[0-9])//'))
unset IFS

if [ ${#CERTS[@]} -eq 0 ]; then
    echo "❌ Developer ID証明書が見つかりません"
    exit 1
fi

# 証明書を選択
for i in "${!CERTS[@]}"; do
    echo "$((i+1))) ${CERTS[$i]}"
done
echo ""
echo "番号を入力してください:"
read -r CERT_NUM

if [ $CERT_NUM -gt 0 ] && [ $CERT_NUM -le ${#CERTS[@]} ]; then
    SELECTED_CERT="${CERTS[$((CERT_NUM-1))]}"
    # 証明書名を抽出（SHA-1ハッシュを除く）
    DEVELOPER_ID=$(echo "$SELECTED_CERT" | sed 's/^[[:space:]]*[A-F0-9]* "//' | sed 's/"$//')
    echo "選択された証明書: $DEVELOPER_ID"
else
    echo "❌ 無効な選択です"
    exit 1
fi

# .env.signingファイルを作成
echo ""
echo "📝 設定ファイルを作成中..."
cat > .env.signing << EOF
# Apple Developer署名設定
# 作成日: $(date)

# Developer ID証明書
export DEVELOPER_ID="$DEVELOPER_ID"

# Apple ID (開発者アカウントのメールアドレス)
export APPLE_ID="$APPLE_ID"

# Team ID
export APPLE_TEAM_ID="$TEAM_ID"

# App用パスワード
export APP_PASSWORD="$APP_PASSWORD"
EOF

echo "✅ 設定が完了しました！"
echo ""
echo "🚀 署名付きインストーラーをビルドするには:"
echo "source .env.signing && ./scripts/build-signed-installer.sh"
echo ""
echo "📌 設定内容:"
echo "   Team ID: $TEAM_ID"
echo "   Apple ID: $APPLE_ID"
echo "   証明書: $DEVELOPER_ID"
echo "   設定ファイル: .env.signing"