# QUIVer 現在稼働中のサービス

## 🟢 稼働中のサービス

### 1. Cloudflare Pages (フロントエンド)
- **URL**: https://quiver-network.pages.dev
- **カスタムドメイン**: 
  - https://quiver.network
  - https://explorer.quiver.network
  - https://api.quiver.network
  - https://dashboard.quiver.network
  - https://security.quiver.network
  - https://quicpair.quiver.network
  - https://playground.quiver.network

### 2. Gateway Service (ローカル開発)
- **場所**: `/gateway`
- **ポート**: 8080
- **状態**: ローカルで開発中
- **機能**: 
  - LLMプロキシ (Ollama/Jan-Nano)
  - レート制限
  - 署名付きレシート発行

### 3. Signaling Server (ローカル開発)
- **場所**: `/gateway/cmd/signaling`
- **ポート**: 8081
- **状態**: ローカルで実行可能
- **機能**:
  - WebRTC シグナリング
  - P2P接続の仲介

### 4. Provider Service (ローカル開発)
- **場所**: `/provider`
- **ポート**: 4001 (libp2p)
- **状態**: 開発中
- **機能**:
  - P2P通信
  - ジョブ実行
  - レシート発行

## 🚧 開発中/未実装のサービス

### 1. Bootstrap Nodes
- **必要性**: P2Pネットワークの初期接続点
- **現状**: libp2p実装はあるが、公開ノードなし
- **今後**: 複数のリージョンに配置予定

### 2. Registry Service
- **必要性**: プロバイダーの登録・検索
- **現状**: 未実装
- **今後**: REST APIとして実装予定

### 3. Metrics/Monitoring
- **必要性**: ネットワーク状態の監視
- **現状**: 基本的なメトリクスのみ
- **今後**: Prometheus + Grafana

### 4. Settlement Service
- **必要性**: オンチェーン決済
- **現状**: モック実装のみ
- **今後**: Polygon統合

## 📝 DNS設定の推奨事項

現在、バックエンドサービスは開発中のため、以下の設定を推奨：

1. **gateway.quiver.network** → Cloudflare Pages (APIドキュメント表示)
2. **signal.quiver.network** → Cloudflare Pages (今後WebSocket対応)
3. **registry.quiver.network** → Cloudflare Pages (プレースホルダー)
4. **metrics.quiver.network** → Cloudflare Pages (ダッシュボード表示)

実際のサービスがデプロイされたら、以下に更新：
- Aレコードで実際のサーバーIPを指定
- SRVレコードでサービスディスカバリを有効化

## 🚀 次のステップ

1. **Signaling Server のデプロイ**
   - Cloudflare Durable Objects または
   - 専用VPS (WebSocket対応)

2. **Bootstrap Node の設置**
   - 最低3ノード（地理的分散）
   - 24/7稼働の保証

3. **Gateway API の公開**
   - HTTPS対応
   - 認証・レート制限
   - 負荷分散

4. **モニタリング基盤**
   - メトリクス収集
   - アラート設定
   - 可視化ダッシュボード