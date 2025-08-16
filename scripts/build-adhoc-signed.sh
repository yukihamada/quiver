#!/bin/bash
# Ad-hoc署名付きインストーラービルド（無料・ローカルのみ）

set -e

echo "🔨 Ad-hoc署名付きインストーラーをビルド中..."
echo "⚠️  注意: このインストーラーはこのMacでのみ動作します"
echo ""

# 既存のビルドスクリプトを実行
./scripts/build-signed-installer.sh

# Ad-hoc署名を適用
echo "🔐 Ad-hoc署名を適用中..."
codesign --deep --force -s - dist/QUIVerProvider.pkg

# 検証
echo "✅ 署名を検証中..."
codesign --verify --verbose dist/QUIVerProvider.pkg

# ローカルインストール用スクリプト作成
cat > dist/install-local.sh << 'EOF'
#!/bin/bash
# ローカルインストールスクリプト

echo "QUIVer Provider ローカルインストール"
echo "================================="
echo ""
echo "⚠️  このインストーラーはこのMacでのみ動作します"
echo ""

# 隔離属性を削除
xattr -cr QUIVerProvider.pkg

# インストール
sudo installer -pkg QUIVerProvider.pkg -target / -allowUntrusted

echo ""
echo "✅ インストール完了！"
EOF

chmod +x dist/install-local.sh

echo ""
echo "✅ Ad-hoc署名付きインストーラーの作成が完了しました！"
echo ""
echo "📦 ファイル:"
echo "   - dist/QUIVerProvider.pkg (Ad-hoc署名済み)"
echo "   - dist/install-local.sh (インストールスクリプト)"
echo ""
echo "⚠️  重要: このパッケージは他のMacでは「開発元が未確認」エラーが出ます"