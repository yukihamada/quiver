# QUIVer Provider - Earn $1,000+/Month with Your Mac

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/Platform-macOS-lightgrey.svg)](https://www.apple.com/macos/)

[日本語版 →](README_JP.md)

Turn your idle Mac into a passive income machine. QUIVer Provider lets you earn cryptocurrency by sharing your Mac's computing power for AI tasks.

## Features

- **QUIC Transport**: Fast, secure P2P communication with built-in encryption
- **NAT Traversal**: Circuit Relay v2 support for nodes behind NAT
- **Blockchain Settlement**: Polygon-based payment channels for micropayments
- **Token Economics**: QUIV token with staking and governance
- **Cryptographic Receipts**: Ed25519-signed receipts with Merkle tree aggregation

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Client    │────▶│   Gateway    │────▶│  Provider   │
│  (iPhone)   │     │  (P2P Node)  │     │ (GPU Node)  │
└─────────────┘     └──────────────┘     └─────────────┘
                            │                     │
                            ▼                     ▼
                    ┌──────────────┐     ┌─────────────┐
                    │  Aggregator  │     │ Blockchain  │
                    │   (Merkle)   │     │ (Polygon)   │
                    └──────────────┘     └─────────────┘
```

## Quick Start

### Prerequisites

- Go 1.23+
- Node.js 18+ (for smart contracts)
- Make
- jq (for demo script)

### Installation

```bash
# Clone the repository
git clone https://github.com/quiver/quiver
cd quiver

# Build all services
make build-all

# Install contract dependencies
cd contracts
npm install
```

### Running the Network

1. **Start the complete network:**
   ```bash
   ./scripts/start-network.sh
   ```

2. **Run the demo:**
   ```bash
   ./scripts/demo.sh
   ```

3. **Stop the network:**
   ```bash
   ./scripts/stop-network.sh
   ```

## Service Endpoints

- **Bootstrap Node**: `localhost:4001` (P2P)
- **Gateway API**: `http://localhost:8081`
- **Provider API**: `http://localhost:8082`
- **Aggregator API**: `http://localhost:8083`

## API Usage

### Make an Inference Request

```bash
curl -X POST http://localhost:8081/inference \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Explain quantum computing",
    "model": "llama2",
    "max_tokens": 100
  }'
```

### Check Available Providers

```bash
curl http://localhost:8081/providers
```

### View Aggregated Receipts

```bash
curl http://localhost:8083/receipts
```

## Smart Contracts

Deploy contracts to testnet:

```bash
cd contracts

# Create .env file with your private key
cp .env.example .env
# Edit .env and add your PRIVATE_KEY

# Deploy to Polygon Amoy testnet
npm run deploy:amoy

# Deploy to Sepolia testnet
npm run deploy:sepolia
```

## Token Economics

- **Total Supply**: 1 billion QUIV tokens
- **Distribution**:
  - 40% Network rewards
  - 20% Team & advisors (4-year vesting)
  - 20% Ecosystem fund
  - 10% Public sale
  - 10% Private sale

### Staking Tiers

| Tier     | Required QUIV | Benefits                    |
|----------|---------------|-----------------------------|  
| Bronze   | 1,000         | 5% fee discount            |
| Silver   | 10,000        | 10% discount, priority     |
| Gold     | 100,000       | 20% discount, governance   |
| Platinum | 1,000,000     | 30% discount, all benefits |

## Configuration

### Environment Variables

- `QUIVER_BOOTSTRAP`: Bootstrap node address
- `QUIVER_LISTEN`: Listen address (default: `/ip4/0.0.0.0/tcp/0`)
- `QUIVER_PROVIDER_URL`: LLM provider URL
- `QUIVER_PRIVATE_KEY`: Node private key path

### P2P Network

The network uses libp2p with QUIC transport. To run a public relay node:

```bash
cd provider
./bin/provider --relay --public-ip YOUR_PUBLIC_IP
```

## Development

### Running Tests

```bash
# Run all tests
make test-all

# Run specific service tests
cd provider && go test ./...
```

### Code Structure

```
.
├── provider/       # AI inference provider service
├── gateway/        # Client-facing API gateway
├── aggregator/     # Receipt aggregation service
├── contracts/      # Smart contracts
├── scripts/        # Deployment and utility scripts
└── docs/          # Documentation
```

## Troubleshooting

### Port Already in Use

If you see "Port already in use" errors:
```bash
# Stop all services
./scripts/stop-network.sh

# Check for remaining processes
lsof -i :4001 -i :8081 -i :8082 -i :8083

# Kill any remaining processes
kill -9 <PID>
```

### Provider Not Found

If no providers are found:
1. Check provider logs: `tail -f /tmp/quiver_logs/provider.log`
2. Ensure bootstrap node is running
3. Check firewall settings

### Contract Deployment Issues

1. Ensure you have testnet tokens (MATIC for Polygon, ETH for Sepolia)
2. Check your RPC endpoint is correct
3. Verify your private key has sufficient balance

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- Documentation: https://docs.quiver.network
- Discord: https://discord.gg/quiver
- Twitter: @quivernetwork

## Acknowledgments

- libp2p team for the excellent P2P framework
- QUIC working group for the protocol specification
- Polygon team for the scalable blockchain infrastructure