# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

QUIVer is a P2P QUIC provider system that proxies to local LLMs (Ollama/Jan-Nano), issues signed usage receipts, and handles on-chain settlement. The system operates in deterministic mode with strict security constraints.

## Architecture

### Core Components

1. **Provider** (`/provider`) - Go service with libp2p QUIC host
   - Advertises on DHT
   - Streams protocol: `/jan-nano/1.0.0`
   - Issues Ed25519 signed receipts for usage

2. **Gateway** (`/gateway`) - API service (Go or Python FastAPI)
   - Exposes `POST /generate` endpoint
   - Proxies requests to LLM providers
   - Enforces rate limiting and input constraints

3. **Aggregator** (`/aggregator`) - Receipt aggregation service
   - Batches receipts into Merkle trees per epoch
   - Exposes `/commit` and `/claim` endpoints
   - Produces Merkle roots for on-chain settlement

4. **Contracts** (`/contracts`) - Settlement layer
   - TypeScript/Solidity for EVM or Anchor for Solana
   - In-memory mock for local testing
   - Verifies Merkle proofs and handles claims

### Design Principles

- **Clean Architecture**: Strict separation of concerns with interface boundaries
- **Small Cohesive Packages**: Each package has a single, well-defined responsibility
- **Structured Logging**: All components use structured logging (no PII in logs)
- **Error Handling**: Comprehensive error handling at all boundaries
- **Deterministic Operation**: Fixed temperature=0, consistent seeds

## Development Commands

### Build System
```bash
# Build all components
make build

# Build specific component
make build-provider
make build-gateway
make build-aggregator

# Run all tests
make test

# Run linting
make lint

# Run benchmarks
make bench

# Clean build artifacts
make clean
```

### Component-Specific Commands

#### Provider
```bash
# Run provider locally
cd provider && go run cmd/provider/main.go

# Run provider tests
cd provider && go test ./...

# Run specific test
cd provider && go test -run TestName ./pkg/...
```

#### Gateway
```bash
# Run gateway (Go version)
cd gateway && go run cmd/gateway/main.go

# Run gateway (Python version)
cd gateway && python -m uvicorn main:app --reload

# Run gateway tests
cd gateway && go test ./... # or pytest
```

#### Aggregator
```bash
# Run aggregator
cd aggregator && go run cmd/aggregator/main.go

# Run aggregator tests
cd aggregator && go test ./...
```

#### Contracts
```bash
# Compile contracts (EVM)
cd contracts && npx hardhat compile

# Run contract tests
cd contracts && npx hardhat test

# Deploy to local network
cd contracts && npx hardhat run scripts/deploy.js --network localhost
```

## Testing Strategy

### Integration Tests
```bash
# Run full integration test suite
make test-integration

# Run P2P connectivity tests
cd tests && pytest test_p2p_connectivity.py

# Run end-to-end flow tests
cd tests && pytest test_e2e_flow.py
```

### Unit Tests
- Each package should have comprehensive unit tests
- Mock external dependencies (LLM, blockchain)
- Test deterministic behavior explicitly

## Security Constraints

1. **No PII in Logs**: Use structured logging with sanitization
2. **Rate Limiting**: Implement at gateway level
3. **Input Validation**: Strict limits on prompt size and content
4. **Deterministic Mode**: Always use temperature=0 or fixed seed
5. **Reproducible Builds**: Use fixed dependency versions

## API Specifications

### Gateway API
```yaml
POST /generate
Content-Type: application/json

Request:
{
  "prompt": "string",
  "model": "string",
  "max_tokens": number,
  "seed": number
}

Response:
{
  "completion": "string",
  "receipt": {
    "id": "string",
    "timestamp": number,
    "signature": "string",
    "usage": {
      "prompt_tokens": number,
      "completion_tokens": number
    }
  }
}
```

### Aggregator API
```yaml
POST /commit
Content-Type: application/json

Request:
{
  "receipts": [Receipt],
  "epoch": number
}

Response:
{
  "merkle_root": "string",
  "epoch": number,
  "receipt_count": number
}

POST /claim
Content-Type: application/json

Request:
{
  "receipt_id": "string",
  "merkle_proof": [string]
}

Response:
{
  "valid": boolean,
  "amount": number
}
```

## Implementation Notes

1. **P2P Communication**: Use libp2p with QUIC transport for all peer communication
2. **Receipt Format**: Ed25519 signatures over canonical JSON representation
3. **Merkle Tree**: Use standard binary Merkle tree construction
4. **Epoch Management**: Fixed-duration epochs (e.g., 1 hour) for batching
5. **Mock Chain**: In-memory implementation for local testing before EVM/Solana

## Output Policies

- Design specs should be machine-readable (JSON/YAML/Schema)
- Generate buildable code incrementally
- Minimize non-code documentation
- Focus on implementation over explanation