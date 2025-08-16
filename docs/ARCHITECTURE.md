# QUIVer Architecture Overview

## System Design

QUIVer is built as a modular, microservices-based architecture that enables decentralized AI inference at scale. The system consists of four main components that work together to create a robust P2P network.

```
┌─────────────────────────────────────────────────────────────────────┐
│                           Client Applications                        │
│                    (Web, Mobile, SDK, API Clients)                  │
└───────────────────────┬─────────────────────────────────────────────┘
                        │ HTTPS/WSS
                        ▼
┌─────────────────────────────────────────────────────────────────────┐
│                            Gateway Layer                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌───────────┐ │
│  │  Gateway 1  │  │  Gateway 2  │  │  Gateway 3  │  │    ...    │ │
│  │  (Region A) │  │  (Region B) │  │  (Region C) │  │           │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └─────┬─────┘ │
└─────────┼────────────────┼────────────────┼───────────────┼───────┘
          │                │                │               │
          └────────────────┴────────────────┴───────────────┘
                                    │ P2P (QUIC)
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         P2P Network Layer                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌───────────┐ │
│  │ Provider 1  │  │ Provider 2  │  │ Provider 3  │  │    ...    │ │
│  │ (GPU Node)  │◄─┤ (GPU Node)  │◄─┤ (CPU Node)  │◄─┤           │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └───────────┘ │
│         ▲                ▲                ▲               ▲         │
│         └────────────────┴────────────────┴───────────────┘         │
│                          Kademlia DHT                               │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Settlement Layer                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                │
│  │ Aggregator  │  │  Receipts   │  │   Smart     │                │
│  │  Service    │──┤   Storage   │──┤  Contracts  │                │
│  └─────────────┘  └─────────────┘  └─────────────┘                │
└─────────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Gateway Service

The Gateway acts as the entry point for client applications, providing a familiar REST API while handling P2P complexity.

**Responsibilities:**
- HTTP/WebSocket API endpoints
- Authentication and authorization
- Rate limiting and quota management
- Request routing to optimal providers
- Response aggregation and caching
- Monitoring and metrics collection

**Key Features:**
- Horizontal scaling with load balancing
- Geographic distribution for low latency
- Automatic failover and retry logic
- Request/response validation
- API versioning support

### 2. Provider Service

Providers are the workhorses of the network, running AI models and processing inference requests.

**Responsibilities:**
- Local LLM hosting (via Ollama)
- P2P network participation
- Stream protocol handling
- Receipt generation and signing
- Resource management
- Health monitoring

**Key Features:**
- Multi-model support
- GPU/CPU optimization
- Deterministic inference (seed support)
- Privacy-preserving execution
- Automatic model management
- Performance benchmarking

### 3. P2P Network Layer

Built on libp2p, this layer handles all peer-to-peer communication and discovery.

**Components:**
- **Transport**: QUIC protocol for fast, secure connections
- **Discovery**: Kademlia DHT for peer discovery
- **Routing**: Content-based routing for optimal provider selection
- **Security**: TLS 1.3 encryption, Ed25519 peer identity

**Protocol Stack:**
```
Application Layer:    /quiver/inference/1.0.0
Stream Multiplexing:  yamux
Security:            TLS 1.3
Transport:           QUIC
Network:             TCP/UDP
```

### 4. Aggregator Service

Handles receipt collection, verification, and settlement preparation.

**Responsibilities:**
- Receipt validation and storage
- Merkle tree construction
- Epoch management
- Settlement batch preparation
- Fraud detection
- Provider reputation tracking

**Key Features:**
- Efficient receipt batching
- Merkle proof generation
- On-chain settlement interface
- Provider analytics
- Dispute resolution support

## Data Flow

### Inference Request Flow

1. **Client Request**
   ```
   Client → Gateway (HTTPS)
   ```
   - Client sends inference request with authentication
   - Gateway validates request and checks quota

2. **Provider Selection**
   ```
   Gateway → DHT Lookup → Provider List
   ```
   - Gateway queries DHT for available providers
   - Filters by model, capacity, latency, reputation

3. **P2P Streaming**
   ```
   Gateway ←→ Provider (QUIC Stream)
   ```
   - Establishes encrypted QUIC stream
   - Sends inference request
   - Receives streaming response

4. **Receipt Generation**
   ```
   Provider → Ed25519 Sign → Receipt
   ```
   - Provider signs usage data
   - Includes timestamp, tokens, model info
   - Returns with response

5. **Response Delivery**
   ```
   Gateway → Client (HTTPS/WSS)
   ```
   - Gateway forwards response
   - Includes signed receipt
   - Updates metrics

### Settlement Flow

1. **Receipt Collection**
   ```
   Providers → Aggregator (Batch Submit)
   ```
   - Providers submit receipts periodically
   - Aggregator validates signatures

2. **Merkle Tree Construction**
   ```
   Receipts → Merkle Tree → Root Hash
   ```
   - Groups receipts by epoch
   - Builds Merkle tree
   - Computes root hash

3. **On-chain Settlement**
   ```
   Aggregator → Smart Contract → Settlement
   ```
   - Submits Merkle root on-chain
   - Providers claim with proofs
   - Automatic payment distribution

## Security Model

### Cryptographic Primitives

- **Identity**: Ed25519 key pairs for all nodes
- **Transport**: TLS 1.3 with QUIC
- **Signatures**: Ed25519 for receipts
- **Hashing**: SHA-256 for Merkle trees

### Trust Assumptions

1. **Providers** are untrusted but economically incentivized
2. **Gateways** are semi-trusted (operated by QUIVer or partners)
3. **Aggregators** use cryptographic proofs (trustless)
4. **Smart contracts** provide final settlement guarantees

### Security Features

- End-to-end encryption for all communications
- Receipt signatures prevent forgery
- Merkle proofs enable trustless verification
- Rate limiting prevents abuse
- Canary responses detect misbehavior

## Performance Characteristics

### Latency Breakdown

```
Total Latency: 50-200ms

Client → Gateway:        5-20ms   (depends on region)
Gateway → Provider:     10-50ms   (P2P routing)
Inference Processing:   20-100ms  (model dependent)
Provider → Gateway:     10-30ms   (streaming)
Gateway → Client:        5-20ms   (response)
```

### Throughput

- **Gateway**: 10,000+ requests/second per instance
- **Provider**: 10-100 requests/second (GPU dependent)
- **Network**: 1M+ requests/second (aggregate)

### Scalability

The system scales horizontally at every layer:

1. **Gateways**: Add instances in new regions
2. **Providers**: Unlimited P2P nodes can join
3. **Aggregators**: Shard by epoch or provider
4. **Settlement**: Multi-chain deployment

## Deployment Architecture

### Development Environment

```yaml
docker-compose:
  - gateway (1 instance)
  - provider (2 instances)
  - aggregator (1 instance)
  - ollama-mock (1 instance)
  - postgres (receipts DB)
```

### Production Environment

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Load      │────▶│  Gateway    │────▶│  Gateway    │
│  Balancer   │     │  Cluster    │     │  Cluster    │
│  (Global)   │     │  (Region 1) │     │  (Region 2) │
└─────────────┘     └─────────────┘     └─────────────┘
                            │                    │
                            ▼                    ▼
                    ┌─────────────┐     ┌─────────────┐
                    │  P2P Swarm  │◄───▶│  P2P Swarm  │
                    │  (Region 1) │     │  (Region 2) │
                    └─────────────┘     └─────────────┘
```

### Infrastructure Requirements

**Gateway Nodes:**
- 4 vCPU, 8GB RAM
- 100GB SSD
- 1Gbps network

**Provider Nodes:**
- 8+ vCPU or GPU
- 16GB+ RAM  
- 500GB+ SSD
- 100Mbps+ network

**Aggregator Nodes:**
- 8 vCPU, 16GB RAM
- 1TB SSD (receipt storage)
- 1Gbps network

## Monitoring and Observability

### Metrics (Prometheus)

- Request rate and latency
- Provider availability
- Model usage statistics
- Network topology metrics
- Receipt processing rate
- Settlement success rate

### Logging (Structured)

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "service": "gateway",
  "trace_id": "abc123",
  "method": "POST",
  "path": "/generate",
  "latency_ms": 87,
  "status": 200
}
```

### Tracing (OpenTelemetry)

- End-to-end request tracing
- P2P communication visualization
- Bottleneck identification
- Error tracking

## Future Architecture Considerations

### Phase 2: Enhanced Privacy
- Trusted Execution Environments (TEE)
- Homomorphic encryption support
- Private model inference

### Phase 3: Advanced Routing
- ML-based provider selection
- Predictive load balancing
- Quality-of-service guarantees

### Phase 4: Decentralized Governance
- DAO for protocol upgrades
- Provider reputation system
- Community-driven model curation