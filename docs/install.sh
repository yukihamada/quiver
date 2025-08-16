#!/bin/bash
# QUIVer Provider macOS ワンクリックインストーラー

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

# 管理者権限チェック
if [ "$EUID" -eq 0 ]; then 
   echo "❌ rootユーザーでは実行しないでください"
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
    echo "手動でダウンロードしてください: https://github.com/yukihamada/quiver/releases"
    exit 1
fi

echo ""
echo "🔐 インストーラーを起動します..."
echo "   ※ パスワードの入力が必要です"
echo ""

# PKGをインストール
if sudo installer -pkg QUIVerProvider.pkg -target /; then
    echo "✅ インストール完了！"
else
    echo "❌ インストールに失敗しました"
    exit 1
fi

# クリーンアップ
cd /
rm -rf "$TEMP_DIR"

echo ""
echo "🎉 QUIVer Providerのインストールが完了しました！"
echo ""
echo "📊 ダッシュボード: http://localhost:8090"
echo "📚 ドキュメント: https://quiver.network/docs"
echo ""
echo "💡 ヒント: メニューバーのQUIVerアイコンから設定できます"
echo ""

# ダッシュボードを開く
if command -v open >/dev/null 2>&1; then
    echo "ダッシュボードを開きますか？ [Y/n] "
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]] || [[ -z "$response" ]]; then
        open http://localhost:8090
    fi
fi