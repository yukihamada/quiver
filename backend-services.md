# QUIVer バックエンドサービス DNS設定ガイド

## 必要なバックエンドサービス

### 1. Gateway API (`gateway.quiver.network`)
- **ポート**: 8080 (HTTP), 443 (HTTPS)
- **役割**: LLMプロバイダーへのプロキシ、レート制限、認証
- **必要なレコード**:
  ```
  gateway    A    <Gateway Server IP>
  gw         CNAME gateway.quiver.network
  api-gw     CNAME gateway.quiver.network
  ```

### 2. WebRTC Signaling Server (`signal.quiver.network`)
- **ポート**: 8081 (WebSocket), 443 (WSS)
- **役割**: WebRTC接続の確立、シグナリング
- **必要なレコード**:
  ```
  signal     A    <Signal Server IP>
  signaling  CNAME signal.quiver.network
  ws         CNAME signal.quiver.network
  wss        CNAME signal.quiver.network
  ```

### 3. Bootstrap Nodes (`bootstrap.quiver.network`)
- **ポート**: 4001 (libp2p)
- **役割**: P2Pネットワークの初期接続点
- **必要なレコード**:
  ```
  bootstrap   A    <Bootstrap Node IP>
  bootstrap1  A    <Bootstrap Node 1 IP>
  bootstrap2  A    <Bootstrap Node 2 IP>
  bootstrap3  A    <Bootstrap Node 3 IP>
  ```

### 4. Provider Registry (`registry.quiver.network`)
- **ポート**: 8082
- **役割**: プロバイダー登録、検索、メタデータ管理
- **必要なレコード**:
  ```
  registry   A    <Registry Server IP>
  ```

## デプロイオプション

### オプション1: 単一サーバー（開発/テスト用）
すべてのサービスを1つのサーバーで実行：
```
gateway    A    123.45.67.89
signal     A    123.45.67.89
bootstrap  A    123.45.67.89
registry   A    123.45.67.89
```

### オプション2: 分散構成（本番用）
```
gateway    A    123.45.67.89  # API Gateway
signal     A    123.45.67.90  # Signaling Server
bootstrap  A    123.45.67.91  # Bootstrap Node 1
bootstrap1 A    123.45.67.92  # Bootstrap Node 2
bootstrap2 A    123.45.67.93  # Bootstrap Node 3
registry   A    123.45.67.94  # Registry Server
```

### オプション3: クラウドサービス利用
```
# Cloudflare Workers/Durable Objects
gateway    CNAME gateway.quiver.workers.dev

# Cloudflare Calls (WebRTC)
signal     CNAME quiver.calls.cloudflare.com

# IPFS/Filecoin Bootstrap
bootstrap  CNAME bootstrap.ipfs.io
```

## SSL証明書の設定

### Cloudflare Proxy有効（推奨）
- 自動SSL証明書
- DDoS保護
- CDN最適化

### Let's Encrypt（直接接続の場合）
```bash
certbot certonly --nginx -d gateway.quiver.network
certbot certonly --nginx -d signal.quiver.network
```

## 現在の実装状況

### ✅ 実装済み
- Gateway基本機能 (`/gateway`)
- WebRTCトランスポート (`/gateway/pkg/webrtc`)
- Provider基本機能 (`/provider`)

### 🚧 追加実装が必要
1. **Signaling Server**
   - WebSocketサーバー
   - STUN/TURNコーディネーション
   - 接続状態管理

2. **Bootstrap Service**
   - libp2p DHT設定
   - ピア検出とアナウンス
   - NAT traversal

3. **Registry Service**
   - Provider登録API
   - 能力クエリAPI
   - ヘルスチェック

## 簡易実装例

### Signaling Server (Node.js)
```javascript
// signal-server.js
const WebSocket = require('ws');
const wss = new WebSocket.Server({ port: 8081 });

const rooms = new Map();

wss.on('connection', (ws) => {
  ws.on('message', (message) => {
    const data = JSON.parse(message);
    
    switch(data.type) {
      case 'join':
        // Room管理
        break;
      case 'offer':
      case 'answer':
      case 'ice-candidate':
        // シグナリング転送
        break;
    }
  });
});
```

### Bootstrap Node (Go)
```go
// bootstrap/main.go
package main

import (
    "github.com/libp2p/go-libp2p"
    dht "github.com/libp2p/go-libp2p-kad-dht"
)

func main() {
    host, _ := libp2p.New(
        libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/4001"),
    )
    
    kademlia, _ := dht.New(ctx, host)
    kademlia.Bootstrap(ctx)
}
```

## 次のステップ

1. バックエンドサービスのIPアドレスを決定
2. `cloudflare-dns-complete.txt` のIP部分を更新
3. CloudflareでDNSレコードをインポート
4. 各サービスをデプロイ
5. SSL証明書を設定