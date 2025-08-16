# QUIVer API Reference

## Table of Contents

- [Authentication](#authentication)
- [Gateway API](#gateway-api)
- [Provider API](#provider-api)
- [Aggregator API](#aggregator-api)
- [WebSocket API](#websocket-api)
- [Error Handling](#error-handling)
- [Rate Limits](#rate-limits)

## Authentication

QUIVer supports two authentication methods:

### API Key Authentication

```bash
curl -X POST https://api.quiver.network/generate \
  -H "Authorization: qvr_test_1234567890abcdef" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello", "model": "llama3.2:3b"}'
```

### JWT Authentication

```bash
curl -X POST https://api.quiver.network/generate \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello", "model": "llama3.2:3b"}'
```

## Gateway API

### Generate Completion

Generate text completion using specified model.

**Endpoint:** `POST /generate`

**Request Body:**
```json
{
  "prompt": "Explain quantum computing",
  "model": "llama3.2:3b",
  "max_tokens": 200,
  "temperature": 0.7,
  "stream": false,
  "seed": 42
}
```

**Parameters:**
- `prompt` (string, required): The input text prompt
- `model` (string, required): Model identifier (e.g., "llama3.2:3b", "phi3:mini")
- `max_tokens` (integer, optional): Maximum tokens to generate (default: 100)
- `temperature` (float, optional): Sampling temperature 0-1 (default: 0)
- `stream` (boolean, optional): Enable streaming response (default: false)
- `seed` (integer, optional): Random seed for deterministic output

**Response:**
```json
{
  "completion": "Quantum computing is a revolutionary approach to computation...",
  "model": "llama3.2:3b",
  "usage": {
    "prompt_tokens": 4,
    "completion_tokens": 50,
    "total_tokens": 54
  },
  "receipt": {
    "id": "rcpt_1234567890abcdef",
    "timestamp": 1701234567,
    "signature": "Ed25519:abcdef...",
    "provider_id": "12D3KooW..."
  }
}
```

### Streaming Response

When `stream: true`, responses are sent as Server-Sent Events:

```
data: {"token": "Quantum", "index": 0}
data: {"token": " computing", "index": 1}
data: {"token": " is", "index": 2}
data: {"done": true, "receipt": {...}}
```

### Health Check

Check gateway health and connectivity.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "service": "gateway",
  "version": "1.0.0",
  "uptime_seconds": 3600,
  "connections": {
    "providers": 5,
    "active_requests": 2
  }
}
```

## Provider API

### Provider Health

Check provider status and capabilities.

**Endpoint:** `GET /health` (Port 8090)

**Response:**
```json
{
  "status": "healthy",
  "service": "provider",
  "peer_id": "12D3KooWLjvJznPvHRuH2KNhgF7z2v2RRoZLvT7bUbYXCdXPmiBF",
  "uptime_seconds": 7200,
  "connections": {
    "connected_peers": 3,
    "total_peers": 5,
    "connections": 3,
    "peers": [
      {
        "id": "12D3KooW...",
        "last_seen": "2024-01-15T10:30:00Z",
        "connected": true
      }
    ]
  },
  "models": [
    {
      "name": "llama3.2:3b",
      "size": 2147483648,
      "loaded": true
    }
  ]
}
```

### Provider Metrics

Prometheus-compatible metrics endpoint.

**Endpoint:** `GET /metrics` (Port 8090)

**Response:** (Prometheus format)
```
# HELP quiver_provider_requests_total Total number of inference requests
# TYPE quiver_provider_requests_total counter
quiver_provider_requests_total 12345

# HELP quiver_provider_tokens_processed Total tokens processed
# TYPE quiver_provider_tokens_processed counter
quiver_provider_tokens_processed 567890

# HELP quiver_provider_request_duration_seconds Request processing duration
# TYPE quiver_provider_request_duration_seconds histogram
quiver_provider_request_duration_seconds_bucket{le="0.1"} 100
```

## Aggregator API

### Submit Receipts

Submit inference receipts for aggregation.

**Endpoint:** `POST /commit`

**Request Body:**
```json
{
  "receipts": [
    {
      "id": "rcpt_1234567890abcdef",
      "timestamp": 1701234567,
      "provider_id": "12D3KooW...",
      "usage": {
        "prompt_tokens": 10,
        "completion_tokens": 50,
        "model": "llama3.2:3b"
      },
      "signature": "Ed25519:abcdef..."
    }
  ],
  "epoch": 1701234000
}
```

**Response:**
```json
{
  "merkle_root": "0x1234567890abcdef...",
  "epoch": 1701234000,
  "receipt_count": 1,
  "total_tokens": 60
}
```

### Claim Rewards

Submit Merkle proof to claim rewards.

**Endpoint:** `POST /claim`

**Request Body:**
```json
{
  "receipt_id": "rcpt_1234567890abcdef",
  "merkle_proof": [
    "0xabcdef...",
    "0x123456..."
  ],
  "epoch": 1701234000
}
```

**Response:**
```json
{
  "valid": true,
  "amount": 0.001,
  "currency": "ETH",
  "transaction_id": "0xtx123..."
}
```

## WebSocket API

### Real-time Inference Stream

Connect via WebSocket for real-time streaming.

**Endpoint:** `wss://api.quiver.network/ws`

**Connection:**
```javascript
const ws = new WebSocket('wss://api.quiver.network/ws');

ws.send(JSON.stringify({
  type: 'auth',
  token: 'your-api-key'
}));

ws.send(JSON.stringify({
  type: 'generate',
  prompt: 'Hello',
  model: 'llama3.2:3b',
  stream: true
}));
```

**Messages:**
```json
// Token stream
{
  "type": "token",
  "token": "Hello",
  "index": 0
}

// Completion
{
  "type": "done",
  "receipt": {...}
}

// Error
{
  "type": "error",
  "error": "Model not available"
}
```

## Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Please retry after 60 seconds.",
    "details": {
      "limit": 10,
      "window": "1m",
      "retry_after": 60
    }
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|------------|-------------|
| `INVALID_REQUEST` | 400 | Invalid request parameters |
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `MODEL_NOT_AVAILABLE` | 503 | Requested model not available |
| `INTERNAL_ERROR` | 500 | Internal server error |

## Rate Limits

Rate limits are enforced per API key/user:

| Plan | Requests/sec | Requests/month | Max Tokens/request |
|------|-------------|----------------|-------------------|
| Free | 1 | 10,000 | 1,000 |
| Starter | 10 | 100,000 | 2,000 |
| Pro | 50 | 1,000,000 | 4,000 |
| Enterprise | 500 | Unlimited | 8,000 |

### Rate Limit Headers

```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 8
X-RateLimit-Reset: 1701234567
```

## Code Examples

### Python
```python
import requests

response = requests.post(
    "https://api.quiver.network/generate",
    headers={
        "Authorization": "qvr_your_api_key",
        "Content-Type": "application/json"
    },
    json={
        "prompt": "Explain quantum computing",
        "model": "llama3.2:3b",
        "max_tokens": 200
    }
)

data = response.json()
print(data["completion"])
```

### JavaScript/Node.js
```javascript
const response = await fetch('https://api.quiver.network/generate', {
  method: 'POST',
  headers: {
    'Authorization': 'qvr_your_api_key',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    prompt: 'Explain quantum computing',
    model: 'llama3.2:3b',
    max_tokens: 200
  })
});

const data = await response.json();
console.log(data.completion);
```

### Go
```go
client := &http.Client{}
reqBody := map[string]interface{}{
    "prompt": "Explain quantum computing",
    "model": "llama3.2:3b",
    "max_tokens": 200,
}

jsonBody, _ := json.Marshal(reqBody)
req, _ := http.NewRequest("POST", "https://api.quiver.network/generate", bytes.NewBuffer(jsonBody))
req.Header.Set("Authorization", "qvr_your_api_key")
req.Header.Set("Content-Type", "application/json")

resp, _ := client.Do(req)
defer resp.Body.Close()

var result map[string]interface{}
json.NewDecoder(resp.Body).Decode(&result)
fmt.Println(result["completion"])
```

### cURL
```bash
curl -X POST https://api.quiver.network/generate \
  -H "Authorization: qvr_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Explain quantum computing",
    "model": "llama3.2:3b",
    "max_tokens": 200
  }'
```

## SDK Libraries

Official SDKs are available for:

- **JavaScript/TypeScript**: `npm install @quiver/sdk`
- **Python**: `pip install quiver-sdk`
- **Go**: `go get github.com/quiver-network/quiver-go`
- **Rust**: `cargo add quiver-sdk`

See individual SDK documentation for language-specific usage.