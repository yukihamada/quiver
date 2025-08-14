# QUIVer API 使用ガイド

QUIVer ネットワークを使用して AI 推論を実行できます。

## エンドポイント

### 本番環境（GCP）
- Gateway API: `http://34.146.32.216:8081`
- WebSocket Stats: `ws://34.146.63.195:8087/ws`

## API の使い方

### 1. 推論リクエスト

AIモデルに質問を送信して回答を取得します。

```bash
# 基本的な推論リクエスト
curl -X POST http://34.146.32.216:8081/inference \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "日本の首都はどこですか？",
    "model": "llama3.2",
    "max_tokens": 100
  }'
```

### 2. 利用可能なプロバイダーを確認

現在オンラインのプロバイダーノードを確認できます。

```bash
curl http://34.146.32.216:8081/providers
```

### 3. ネットワーク統計

リアルタイムのネットワーク統計を取得します。

```bash
# HTTP API
curl http://34.146.63.195:8087/api/stats

# レスポンス例
{
  "node_count": 5,
  "online_nodes": 5,
  "node_details": {
    "12D3KooW...": {
      "peer_id": "12D3KooW...",
      "type": "provider",
      "location": "gcp"
    }
  },
  "countries": 1,
  "total_capacity": 6.0,
  "throughput": 750
}
```

### 4. WebSocket でリアルタイム更新

```javascript
// JavaScript での例
const ws = new WebSocket('ws://34.146.63.195:8087/ws');

ws.onmessage = (event) => {
  const stats = JSON.parse(event.data);
  console.log('現在のノード数:', stats.node_count);
};
```

## Python での使用例

```python
import requests
import json

# QUIVer API エンドポイント
GATEWAY_URL = "http://34.146.32.216:8081"

# 推論リクエスト
def ask_quiver(prompt, max_tokens=100):
    response = requests.post(
        f"{GATEWAY_URL}/inference",
        json={
            "prompt": prompt,
            "model": "llama3.2",
            "max_tokens": max_tokens
        }
    )
    return response.json()

# 使用例
result = ask_quiver("フィボナッチ数列を Python で実装してください")
print(result['response'])
```

## Node.js での使用例

```javascript
const axios = require('axios');

const GATEWAY_URL = 'http://34.146.32.216:8081';

async function askQuiver(prompt, maxTokens = 100) {
  const response = await axios.post(`${GATEWAY_URL}/inference`, {
    prompt: prompt,
    model: 'llama3.2',
    max_tokens: maxTokens
  });
  
  return response.data;
}

// 使用例
(async () => {
  const result = await askQuiver('JavaScriptでHello Worldを書いて');
  console.log(result.response);
})();
```

## 料金

現在はテストネットワークのため無料で使用できます。

将来的な料金体系：
- GPT-4 の 1/25 の価格
- Claude の 1/20 の価格
- 従量課金制

## 制限事項

- 最大トークン数: 2048
- リクエストサイズ: 10KB
- 同時接続数: 100

## トラブルシューティング

### 接続できない場合

1. ネットワーク状態を確認:
   ```bash
   curl http://34.146.63.195:8087/api/stats
   ```

2. プロバイダーが稼働しているか確認:
   ```bash
   curl http://34.146.32.216:8081/providers
   ```

### レスポンスが遅い場合

- ネットワーク内のノード数が少ない可能性があります
- より多くのプロバイダーがオンラインになるまでお待ちください

## サポート

- GitHub Issues: https://github.com/yukihamada/quiver/issues
- Discord: [近日公開]