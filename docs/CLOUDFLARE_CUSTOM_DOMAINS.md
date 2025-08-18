# Cloudflare Custom Domains Setup Guide

## 現在の状況

すべてのサブドメインがHTTP 403エラー（Cloudflare Error 1014）を返しています。これは、ドメインが別のCloudflareアカウントで管理されているためです。

## 解決方法

### 1. Cloudflare Pagesプロジェクトにカスタムドメインを追加

1. [Cloudflare Dashboard](https://dash.cloudflare.com)にログイン
2. アカウントID: `08519319108846c5673d8dbf1a23f6a5`を選択
3. 左メニューから「Workers & Pages」を選択
4. `quiver-network-dab`プロジェクトをクリック
5. 「Settings」タブ → 「Custom domains」セクション
6. 「Add custom domain」ボタンをクリック

### 2. 以下のドメインを一つずつ追加

```
quiver.network
www.quiver.network
docs.quiver.network
explorer.quiver.network
api.quiver.network
dashboard.quiver.network
security.quiver.network
quicpair.quiver.network
playground.quiver.network
status.quiver.network
blog.quiver.network
community.quiver.network
cdn.quiver.network
```

### 3. 各ドメイン追加時の手順

1. ドメイン名を入力（例: `docs.quiver.network`）
2. 「Add domain」をクリック
3. DNS設定の確認画面が表示される
4. 「Activate domain」をクリック
5. SSL証明書の発行を待つ（最大15分）

### 4. 確認事項

- ✅ DNSレコードは既に正しく設定済み（CNAME → quiver-network.pages.dev）
- ✅ Cloudflare Pagesプロジェクトは稼働中
- ❌ カスタムドメインの追加が必要

### 5. トラブルシューティング

もし「Domain is not allowed」エラーが出る場合：
1. ドメインを管理しているCloudflareアカウントでログイン
2. DNS設定でCNAMEレコードが正しく設定されているか確認
3. 必要に応じて、ドメインの所有権を確認

### 6. 完了後の確認

すべてのドメインを追加した後、以下のコマンドで確認：

```bash
# 各サブドメインの確認
curl -I https://docs.quiver.network
curl -I https://api.quiver.network
curl -I https://explorer.quiver.network
# ... 他のサブドメインも同様に確認
```

## 自動化の制限

Cloudflare APIは`.network`ドメインのカスタムドメイン追加を「invalid TLD」として拒否するため、手動での設定が必要です。これはCloudflare側の制限であり、APIでの回避はできません。

## 参考リンク

- [Cloudflare Pages Custom Domains Documentation](https://developers.cloudflare.com/pages/platform/custom-domains/)
- [Cloudflare Dashboard](https://dash.cloudflare.com)