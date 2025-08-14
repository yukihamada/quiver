# QUIVer Usage Guide

## Quick Start

### 1. Start the Network

```bash
# Start all services
make run
```

This will:
- Build all services
- Start a bootstrap node
- Start provider, gateway, and aggregator services
- Display connection information

### 2. Make an Inference Request

From your iPhone or Mac:

```bash
curl -X POST http://localhost:8081/inference \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "What is the capital of France?",
    "model": "llama2",
    "max_tokens": 50
  }'
```

Response:
```json
{
  "completion": "The capital of France is Paris.",
  "receipt": {
    "id": "receipt_abc123",
    "provider_id": "12D3KooW...",
    "signature": "0x1234...",
    "timestamp": "2024-01-15T10:30:00Z"
  },
  "tokens_used": 8
}
```

### 3. Check Network Status

```bash
# View available providers
curl http://localhost:8081/providers

# Check aggregator status
curl http://localhost:8083/status

# View health endpoints
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
```

### 4. Stop the Network

```bash
make stop
```

## Advanced Usage

### Running as a Provider

To earn QUIV tokens by providing inference:

1. **Install Ollama**:
   ```bash
   # macOS
   brew install ollama
   
   # Start Ollama
   ollama serve
   
   # Pull a model
   ollama pull llama2
   ```

2. **Start Provider Node**:
   ```bash
   cd provider
   ./bin/provider --listen /ip4/0.0.0.0/tcp/4002
   ```

3. **Configure for Public Access** (optional):
   ```bash
   ./bin/provider --relay --public-ip YOUR_PUBLIC_IP
   ```

### Running a Gateway

To run your own gateway:

```bash
cd gateway
./bin/gateway --bootstrap /ip4/127.0.0.1/tcp/4001/p2p/BOOTSTRAP_ID
```

### NAT Traversal

For nodes behind NAT:

1. **Use a relay node**:
   ```bash
   # Connect through relay
   ./bin/provider --relay-addr /ip4/relay.quiver.network/tcp/4001/p2p/RELAY_ID
   ```

2. **Enable UPnP** (if supported by router):
   ```bash
   ./bin/provider --enable-upnp
   ```

## Token Operations

### Check QUIV Balance

```bash
# Using aggregator API
curl http://localhost:8083/balance/YOUR_ADDRESS
```

### Stake QUIV Tokens

```bash
curl -X POST http://localhost:8083/stake \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "10000",
    "duration": 30
  }'
```

### Claim Rewards

```bash
curl -X POST http://localhost:8083/claim \
  -H "Content-Type: application/json" \
  -d '{
    "epoch": 100,
    "receipts": ["receipt_id1", "receipt_id2"]
  }'
```

## Monitoring

### View Logs

```bash
# Provider logs
tail -f /tmp/quiver_logs/provider.log

# Gateway logs
tail -f /tmp/quiver_logs/gateway.log

# Aggregator logs
tail -f /tmp/quiver_logs/aggregator.log
```

### Metrics

All services expose Prometheus metrics:

- Provider: http://localhost:8082/metrics
- Gateway: http://localhost:8081/metrics
- Aggregator: http://localhost:8083/metrics

## Troubleshooting

### "No providers found"

1. Check provider is running: `ps aux | grep provider`
2. Check bootstrap connectivity
3. Wait 30 seconds for DHT discovery

### "Connection refused"

1. Check service ports: `lsof -i :8081`
2. Ensure Ollama is running: `ollama list`
3. Check firewall settings

### "Signature verification failed"

1. Ensure time sync between nodes
2. Check provider key consistency
3. Verify receipt format

## Environment Variables

```bash
# Bootstrap node address
export QUIVER_BOOTSTRAP=/ip4/127.0.0.1/tcp/4001/p2p/BOOTSTRAP_ID

# Provider configuration
export QUIVER_PROVIDER_URL=http://localhost:11434
export QUIVER_MODEL=llama2

# Blockchain configuration
export QUIVER_CHAIN_RPC=https://rpc-amoy.polygon.technology
export QUIVER_PRIVATE_KEY=your_private_key_here
```

## Performance Tuning

### Provider Optimization

```bash
# Increase connection limits
ulimit -n 4096

# Set performance mode
./bin/provider --max-streams 100 --cache-size 1GB
```

### Gateway Configuration

```bash
# Enable request batching
./bin/gateway --batch-size 10 --batch-timeout 100ms
```

## Security

### Using Custom Keys

```bash
# Generate new node key
./bin/provider --gen-key > node.key

# Use existing key
./bin/provider --key-file node.key
```

### TLS Configuration

```bash
# Enable TLS
./bin/gateway --tls-cert cert.pem --tls-key key.pem
```

## Next Steps

- [Deploy Smart Contracts](contracts/README.md)
- [Run Production Node](docs/production.md)
- [Join Testnet](docs/testnet.md)
- [Develop Applications](docs/sdk.md)