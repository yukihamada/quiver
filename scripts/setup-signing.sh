#!/bin/bash
# QUIVer Provider 署名セットアップスクリプト

echo "🔐 QUIVer Provider 署名セットアップ"
echo "=================================="
echo ""

# Apple Developer Program加入確認
echo "Apple Developer Programに加入していますか？ [y/N]"
read -r response

if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo ""
    echo "📝 署名に必要な情報を設定します..."
    
    # Developer ID確認
    echo "利用可能な証明書:"
    security find-identity -p basic -v | grep "Developer ID"
    
    echo ""
    echo "Developer ID Application証明書の名前を入力してください:"
    echo "例: Developer ID Application: Your Name (TEAMID)"
    read -r DEVELOPER_ID
    
    echo ""
    echo "Apple ID (メールアドレス):"
    read -r APPLE_ID
    
    echo ""
    echo "Team ID (10文字):"
    read -r TEAM_ID
    
    echo ""
    echo "App-specific password (xxxx-xxxx-xxxx-xxxx):"
    echo "※ https://appleid.apple.com で生成"
    read -s APP_PASSWORD
    
    # 環境変数ファイル作成
    cat > .env.signing << EOF
# Apple Developer署名設定
export DEVELOPER_ID="$DEVELOPER_ID"
export APPLE_ID="$APPLE_ID"
export APPLE_TEAM_ID="$TEAM_ID"
export APP_PASSWORD="$APP_PASSWORD"
EOF
    
    echo ""
    echo "✅ 署名設定を .env.signing に保存しました"
    echo ""
    echo "署名付きビルドを作成するには:"
    echo "source .env.signing && ./scripts/build-signed-installer.sh"
    
else
    echo ""
    echo "🆓 無料の代替方法:"
    echo ""
    echo "1. Ad-hoc署名（ローカルのみ有効）"
    echo "   ./scripts/build-adhoc-signed.sh"
    echo ""
    echo "2. 自己署名証明書の作成"
    echo "   ./scripts/create-self-signed-cert.sh"
    echo ""
    echo "⚠️  注意: これらの方法では「開発元が未確認」エラーは完全には回避できません"
    echo ""
    echo "💡 推奨: Apple Developer Program ($99/年) に加入することで、"
    echo "        ユーザーが警告なしでインストールできるようになります。"
    echo ""
    echo "詳細: https://developer.apple.com/programs/"
fi