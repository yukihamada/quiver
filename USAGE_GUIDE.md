# QUIVer 使い方ガイド

## 🎯 QUIVerとは
P2P QUIC Provider - 分散型のLLM推論システムで、暗号学的証明付きのレシートを生成します。

## 🏗️ システム構成
```
┌─────────────┐     ┌─────────────┐     ┌──────────────┐
│   Client    │────▶│   Gateway   │────▶│   Provider   │
│ (iPhone/Mac)│     │  (API入口)  │     │ (LLM実行)   │
└─────────────┘     └─────────────┘     └──────────────┘
                            │                    │
                            ▼                    ▼
                    ┌──────────────┐    ┌──────────────┐
                    │  Aggregator  │    │    Ollama    │
                    │ (Receipt集約)│    │  (LLMモデル) │
                    └──────────────┘    └──────────────┘
```

## 📦 必要なもの
- macOS (M1/M2/Intel)
- Go 1.21以上
- Ollama (https://ollama.ai)
- 10GB以上の空き容量

## 🚀 クイックスタート（簡易版）

### 1. セットアップ
```bash
# リポジトリに移動
cd /Users/yuki/QUIVer

# セットアップスクリプトを実行
./setup_demo.sh
```

### 2. サービス起動
```bash
# 簡易ゲートウェイを起動（iPhone/Mac両対応）
./bin/gateway_simple
```

### 3. 使用方法

**Macから:**
```bash
curl -X POST http://localhost:8080/generate \
  -H "Content-Type: application/json" \
  -d '{"prompt": "富士山の高さは？"}'
```

**iPhoneから:**
- 同じWiFiネットワークに接続
- Safariで: http://192.168.0.194:8080/generate
- またはショートカットアプリで設定（iOS_Quick_Setup.md参照）

## 🔧 本格的な起動方法（P2P版）

### 1. 全サービスのビルド
```bash
# 依存関係の問題を回避するため、個別にビルド
cd provider && go build -o ../bin/provider cmd/provider/main.go && cd ..
cd gateway && go build -o ../bin/gateway cmd/gateway/main.go && cd ..
cd aggregator && go build -o ../bin/aggregator cmd/aggregator/main.go && cd ..
```

### 2. 各サービスを順番に起動

**ターミナル1: Aggregator（レシート集約）**
```bash
./bin/aggregator
# ポート8082で起動
```

**ターミナル2: Provider（LLM実行）**
```bash
# Ollamaが起動していることを確認
ollama serve

# 別ターミナルで
./bin/provider
# P2PポートとヘルスチェックAPIポート8090で起動
```

**ターミナル3: Gateway（APIゲートウェイ）**
```bash
./bin/gateway
# ポート8081で起動
```

### 3. 動作確認
```bash
# ヘルスチェック
curl http://localhost:8081/health
curl http://localhost:8090/health
curl http://localhost:8082/health

# 推論実行
curl -X POST http://localhost:8081/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "量子コンピュータを簡単に説明して",
    "model": "llama3.2:3b",
    "max_tokens": 100
  }'
```

## 📊 レシート機能の使用

### 1. レシートの確認
推論実行すると、暗号署名付きレシートが返されます：
```json
{
  "completion": "量子コンピュータは...",
  "receipt": {
    "receipt": {
      "receipt_id": "...",
      "provider_key": "...",
      "prompt_hash": "...",
      "output_hash": "...",
      "tokens_in": 15,
      "tokens_out": 85,
      "timestamp_start": "...",
      "timestamp_end": "..."
    },
    "signature": "..."
  }
}
```

### 2. レシートの集約とコミット
```bash
# Aggregatorにレシートを送信
curl -X POST http://localhost:8082/commit \
  -H "Content-Type: application/json" \
  -d '{
    "epoch": 19953,
    "receipts": [/* レシートの配列 */]
  }'
```

### 3. レシートのクレーム（報酬請求）
```bash
curl -X POST http://localhost:8082/claim \
  -H "Content-Type: application/json" \
  -d '{
    "receipt_id": "...",
    "epoch": 19953,
    "merkle_proof": [...]
  }'
```

## 🔍 モニタリング

### Prometheusメトリクス
各サービスで利用可能：
- Provider: http://localhost:8090/metrics
- Gateway: http://localhost:8081/metrics  
- Aggregator: http://localhost:8082/metrics

### ログ確認
すべてのサービスはJSON形式でログ出力します。

## 🛠️ トラブルシューティング

### よくある問題

1. **「Ollama not found」エラー**
   ```bash
   # Ollamaをインストール
   curl -fsSL https://ollama.ai/install.sh | sh
   ```

2. **「port already in use」エラー**
   ```bash
   # 使用中のポートを確認
   lsof -i :8080
   # プロセスを終了
   kill -9 <PID>
   ```

3. **iPhoneから接続できない**
   - 同じWiFiネットワークか確認
   - Macのファイアウォール設定を確認
   - IPアドレスを再確認: `ifconfig en0 | grep inet`

4. **ビルドエラー**
   ```bash
   # 依存関係をクリーンアップ
   go clean -modcache
   # 再度依存関係を取得
   go mod download
   ```

## 📱 モバイルアプリ開発者向け

### REST API仕様
```
POST /generate
Content-Type: application/json

{
  "prompt": "string",
  "model": "string (optional, default: llama3.2:3b)",
  "max_tokens": "number (optional, default: 100)"
}

Response:
{
  "completion": "string",
  "receipt": {
    // 暗号署名付きレシート
  }
}
```

### SwiftUIサンプル
```swift
struct QUIVerClient {
    static func generate(prompt: String) async throws -> String {
        let url = URL(string: "http://192.168.0.194:8080/generate")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        let body = ["prompt": prompt]
        request.httpBody = try JSONEncoder().encode(body)
        
        let (data, _) = try await URLSession.shared.data(for: request)
        let response = try JSONDecoder().decode(GenerateResponse.self, from: data)
        return response.completion
    }
}
```

## 🔐 セキュリティ注意事項

1. **本番環境では必ずHTTPS化**
2. **APIキー認証を実装**
3. **レート制限を適切に設定**
4. **プライベートキーは安全に管理**

## 🚪 サービスの停止

```bash
# すべてのサービスを停止
pkill -f "gateway|provider|aggregator"

# Ollamaも停止
pkill ollama
```

---

詳細な技術仕様は README.md を参照してください。