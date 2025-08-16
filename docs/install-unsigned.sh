#!/bin/bash
# QUIVer Provider 未署名パッケージインストーラー

set -e

echo ""
echo "🚀 QUIVer Provider インストーラー"
echo "================================"
echo ""

# macOSチェック
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ このインストーラーはmacOS専用です"
    exit 1
fi

echo "📦 最新版をダウンロード中..."

# 一時ディレクトリ作成
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

# 最新のPKGをダウンロード
if curl -L -o QUIVerProvider.pkg "https://github.com/yukihamada/quiver/releases/latest/download/QUIVerProvider.pkg" 2>/dev/null; then
    echo "✅ ダウンロード完了"
else
    echo "❌ ダウンロードに失敗しました"
    exit 1
fi

echo ""
echo "🔓 セキュリティ設定を一時的に変更します..."
echo "   ※ パスワードの入力が必要です"
echo ""

# Gatekeeperを一時的に無効化
sudo spctl --master-disable

echo "📦 インストール中..."

# PKGをインストール（署名チェックをスキップ）
if sudo installer -pkg QUIVerProvider.pkg -target / -allowUntrusted; then
    echo "✅ インストール完了！"
else
    echo "❌ インストールに失敗しました"
    # Gatekeeperを再度有効化
    sudo spctl --master-enable
    exit 1
fi

# Gatekeeperを再度有効化
echo "🔒 セキュリティ設定を元に戻します..."
sudo spctl --master-enable

# 隔離属性を削除
echo "🧹 アプリケーションの隔離属性を削除中..."
if [ -d "/Applications/QUIVerProvider.app" ]; then
    sudo xattr -cr /Applications/QUIVerProvider.app
fi
sudo xattr -cr /usr/local/bin/quiver-provider 2>/dev/null || true

# クリーンアップ
cd /
rm -rf "$TEMP_DIR"

echo ""
echo "🎉 QUIVer Providerのインストールが完了しました！"
echo ""
echo "📊 ダッシュボード: http://localhost:8090"
echo "📚 ドキュメント: https://quiver.network/docs"
echo ""

# サービスを開始
echo "🚀 サービスを開始します..."
launchctl load ~/Library/LaunchAgents/network.quiver.provider.plist 2>/dev/null || true
launchctl start network.quiver.provider 2>/dev/null || true

# ダッシュボードを開く
if command -v open >/dev/null 2>&1; then
    echo ""
    echo "ダッシュボードを開きますか？ [Y/n] "
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]] || [[ -z "$response" ]]; then
        sleep 2  # サービス起動を待つ
        open http://localhost:8090
    fi
fi