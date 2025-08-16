# QUIVer Provider 署名ガイド

## 「開発元が未確認」エラーを完全に回避する方法

### 🏆 方法1: Apple Developer Program（推奨）
**コスト**: $99/年  
**効果**: ✅ 完全にエラーを回避

1. [Apple Developer Program](https://developer.apple.com/programs/)に加入
2. Developer ID証明書を取得
3. アプリに署名 + 公証化（Notarization）

```bash
# セットアップ
./scripts/setup-signing.sh

# 署名付きビルド
source .env.signing && ./scripts/build-signed-installer.sh
```

**メリット**:
- ユーザーは警告なしでインストール可能
- プロフェッショナルな配布
- 自動アップデート機能も実装可能

### 🆓 方法2: 無料の代替案

#### A. Ad-hoc署名（ローカルのみ）
```bash
./scripts/build-adhoc-signed.sh
```
- ✅ 無料
- ❌ ビルドしたMacでのみ動作
- ❌ 他のMacでは警告が出る

#### B. 自己署名証明書
```bash
./scripts/create-self-signed-cert.sh
```
- ✅ 無料
- ✅ 証明書名が表示される
- ❌ 「信頼されていない開発元」と表示
- ❌ ユーザーが手動で信頼設定が必要

### 📊 比較表

| 方法 | コスト | エラー回避 | 配布の容易さ | プロ向け |
|------|--------|------------|--------------|----------|
| Apple Developer | $99/年 | ✅ 完全 | ✅ 簡単 | ✅ |
| Ad-hoc署名 | 無料 | ❌ | ❌ 困難 | ❌ |
| 自己署名 | 無料 | ⚠️ 部分的 | ⚠️ 中程度 | ❌ |
| 未署名 | 無料 | ❌ | ⚠️ 要説明 | ❌ |

### 🚀 今すぐ始める

1. **本格的な配布を考えている場合**:
   ```bash
   # Apple Developer Programに加入後
   ./scripts/setup-signing.sh
   ```

2. **とりあえず試したい場合**:
   ```bash
   # 自己署名証明書を作成
   ./scripts/create-self-signed-cert.sh
   ```

### 💡 ヒント

- オープンソースプロジェクトでも、ユーザー体験を考えるとApple Developer Programは価値があります
- 年間$99は月額約$8.25 - コーヒー2杯分です
- 一度設定すれば、GitHub Actionsで自動署名も可能

### 🔗 参考リンク

- [Apple Developer Program](https://developer.apple.com/programs/)
- [Code Signing Guide](https://developer.apple.com/documentation/security/code_signing_guide)
- [Notarizing macOS Software](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)