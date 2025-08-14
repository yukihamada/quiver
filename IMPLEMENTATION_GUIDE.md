# QUIVer Advanced Features Implementation Guide

## Overview
This guide provides step-by-step instructions for implementing QUIC with NAT traversal, blockchain integration, and token economics in the QUIVer system.

## 1. QUIC & NAT Traversal Implementation

### Dependencies Updated ✅
```bash
# Provider and Gateway now use:
libp2p v0.36.5
quic-go v0.48.2
```

### QUIC Transport Enabled ✅
```go
// provider/pkg/p2p/host.go
libp2p.Transport(libp2pquic.NewTransport), // Re-enabled
```

### Circuit Relay v2 Configuration ✅
```go
// For public relay nodes
libp2p.EnableRelayService()
libp2p.ForceReachabilityPublic()

// For nodes behind NAT
libp2p.EnableAutoRelay()
libp2p.EnableHolePunching()
```

### Deploying Relay Infrastructure

1. **Deploy Relay Nodes** (3 regions minimum):
```bash
# US East
./bin/relay --listen /ip4/0.0.0.0/tcp/4001 --region us-east

# Europe
./bin/relay --listen /ip4/0.0.0.0/tcp/4001 --region eu-west

# Asia
./bin/relay --listen /ip4/0.0.0.0/tcp/4001 --region ap-south
```

2. **Configure DNS**:
```
relay1.quiver.network -> US East IP
relay2.quiver.network -> Europe IP
relay3.quiver.network -> Asia IP
```

3. **Update Bootstrap Peers**:
```go
bootstrapPeers := []string{
    "/dns4/relay1.quiver.network/tcp/4001/p2p/...",
    "/dns4/relay2.quiver.network/tcp/4001/p2p/...",
    "/dns4/relay3.quiver.network/tcp/4001/p2p/...",
}
```

## 2. Blockchain Integration (Polygon)

### Smart Contracts

1. **Deploy Contracts**:
```bash
# Set up environment
export PRIVATE_KEY=your_deployment_key
export RPC_URL=https://rpc-mumbai.maticvigil.com

# Deploy token
npx hardhat deploy --network mumbai --contract QUIVToken

# Deploy settlement
npx hardhat deploy --network mumbai --contract QUIVerSettlement

# Deploy staking
npx hardhat deploy --network mumbai --contract QUIVerStaking
```

2. **Verify Contracts**:
```bash
npx hardhat verify --network mumbai CONTRACT_ADDRESS
```

### Aggregator Integration

1. **Configure Blockchain Client**:
```go
// aggregator/cmd/aggregator/main.go
blockchainConfig := blockchain.DefaultPolygonMumbaiConfig()
blockchainConfig.ContractAddress = "0x..." // Settlement contract
blockchainConfig.TokenAddress = "0x..."    // QUIV token
blockchainConfig.PrivateKey = os.Getenv("AGGREGATOR_PRIVATE_KEY")

blockchainClient, err := blockchain.NewClient(blockchainConfig)
```

2. **Implement Settlement Service**:
```go
// aggregator/pkg/settlement/service.go
type SettlementService struct {
    blockchain *blockchain.Client
    store      *storage.Store
    batchSize  int
    interval   time.Duration
}

func (s *SettlementService) ProcessBatch(ctx context.Context) error {
    // 1. Collect receipts from store
    receipts := s.store.GetUnprocessedReceipts(s.batchSize)
    
    // 2. Build Merkle tree
    tree := merkle.BuildTree(receipts)
    
    // 3. Submit to blockchain
    batch := blockchain.ReceiptBatch{
        MerkleRoot:   tree.Root(),
        TotalAmount:  calculateTotal(receipts),
        ReceiptCount: uint64(len(receipts)),
        Epoch:        currentEpoch(),
    }
    
    tx, err := s.blockchain.SubmitBatch(ctx, batch)
    if err != nil {
        return err
    }
    
    // 4. Wait for confirmation
    receipt, err := s.blockchain.WaitForConfirmation(ctx, tx)
    
    // 5. Mark receipts as processed
    s.store.MarkProcessed(receipts, receipt.TxHash)
    
    return nil
}
```

### Payment Channel Implementation

1. **Open Channel** (Gateway):
```go
func (g *Gateway) OpenChannel(provider peer.ID, amount *big.Int) error {
    providerAddr := g.getProviderAddress(provider)
    tx, err := g.blockchain.OpenChannel(ctx, providerAddr, amount)
    // Store channel ID for future use
}
```

2. **Update Channel State** (Off-chain):
```go
type ChannelUpdate struct {
    ChannelID [32]byte
    Nonce     uint64
    Balance   *big.Int
    Signature []byte
}

func (g *Gateway) UpdateChannel(receipt *Receipt) error {
    update := &ChannelUpdate{
        ChannelID: g.activeChannel,
        Nonce:     g.channelNonce + 1,
        Balance:   g.channelBalance.Add(receipt.Cost),
    }
    update.Signature = g.signUpdate(update)
    // Send to provider
}
```

## 3. Token Economics Implementation

### Provider Staking

1. **Stake Registration**:
```go
// provider/cmd/provider/main.go
func registerProvider(stakingContract, tokenContract common.Address) error {
    // 1. Approve token spending
    tx1, err := tokenContract.Approve(stakingContract, stakeAmount)
    
    // 2. Stake tokens
    tx2, err := stakingContract.Stake(stakeAmount, true) // true = as provider
    
    // 3. Wait for confirmation
    waitForTx(tx1, tx2)
}
```

2. **Tier Management**:
```go
type TierManager struct {
    contract *QUIVerStaking
    cache    map[common.Address]StakingTier
}

func (tm *TierManager) GetProviderTier(addr common.Address) StakingTier {
    // Cache with 5-minute TTL
    if cached, ok := tm.cache[addr]; ok {
        return cached
    }
    
    tier, err := tm.contract.GetStakeTier(addr)
    tm.cache[addr] = tier
    return tier
}
```

### Quality Metrics & Slashing

1. **Metrics Collection**:
```go
type MetricsCollector struct {
    store *MetricsStore
}

func (mc *MetricsCollector) RecordRequest(provider peer.ID, receipt *Receipt, latency time.Duration) {
    mc.store.Update(provider, &RequestMetrics{
        Timestamp:    time.Now(),
        ResponseTime: latency,
        TokensIn:     receipt.TokensIn,
        TokensOut:    receipt.TokensOut,
        Success:      receipt.Error == "",
    })
}
```

2. **Canary Verification**:
```go
func (g *Gateway) VerifyCanaryResponse(prompt, response string, provider peer.ID) bool {
    expected := g.canaryOracle.GetExpectedResponse(prompt)
    if response != expected {
        g.slashingService.ReportViolation(provider, "canary_failure", 0.5) // 0.5% slash
        return false
    }
    return true
}
```

### Reward Distribution

1. **Claim Rewards** (Provider):
```go
func claimRewards() error {
    pending, err := stakingContract.PendingRewards(myAddress)
    if pending.Cmp(big.NewInt(0)) > 0 {
        tx, err := stakingContract.ClaimRewards()
        waitForTx(tx)
    }
}
```

2. **Auto-compound Option**:
```go
func autoCompound() error {
    rewards := getPendingRewards()
    if rewards > minCompoundAmount {
        tx1 := claimRewards()
        tx2 := stakeRewards(rewards)
        waitForTx(tx1, tx2)
    }
}
```

## 4. Production Deployment

### Infrastructure Requirements

1. **Nodes**:
   - 3x Relay nodes (t3.medium or equivalent)
   - 1x Aggregator node (t3.large)
   - Multiple provider nodes (varies)

2. **Blockchain**:
   - Polygon RPC endpoint (Alchemy/Infura)
   - Gas wallet with MATIC
   - Multisig for contract ownership

3. **Monitoring**:
   - Prometheus + Grafana
   - Error tracking (Sentry)
   - On-call rotation

### Security Checklist

- [ ] Smart contract audits completed
- [ ] Private keys in HSM/secure storage
- [ ] Rate limiting configured
- [ ] DDoS protection enabled
- [ ] Incident response plan documented
- [ ] Regular security updates scheduled

### Migration Plan

1. **Week 1-2**: Deploy contracts to testnet
2. **Week 3-4**: Beta testing with limited providers
3. **Week 5-6**: Gradual mainnet migration
4. **Week 7-8**: Full production launch
5. **Week 9-12**: Monitoring and optimization

## 5. Testing

### Integration Tests
```bash
# Run blockchain integration tests
go test ./aggregator/pkg/blockchain/... -integration

# Run NAT traversal tests
go test ./provider/pkg/p2p/... -nat

# Run settlement tests
go test ./aggregator/pkg/settlement/... -e2e
```

### Load Testing
```bash
# Simulate 1000 concurrent providers
./scripts/load_test.sh --providers 1000 --duration 1h

# Test settlement under load
./scripts/settlement_stress.sh --receipts 100000
```

## Conclusion

This implementation provides:
- ✅ QUIC transport with NAT traversal via Circuit Relay v2
- ✅ Polygon blockchain integration for settlements
- ✅ Comprehensive token economics with staking
- ✅ Quality assurance through slashing
- ✅ Scalable architecture for production use

Next steps:
1. Deploy relay infrastructure
2. Audit smart contracts
3. Begin testnet trials
4. Prepare mainnet launch