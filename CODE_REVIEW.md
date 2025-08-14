# CODE_REVIEW.md

## Correctness Issues

• **Unchecked errors**: Multiple instances of ignored error returns
  - `encoder.Encode()` in provider/pkg/stream/handler.go:125,131
  - `w.Write()` in provider/pkg/llm/client_test.go:15
  - `json.Unmarshal()` in gateway/pkg/api/handlers_test.go:267
  - `store.Store()` in aggregator tests
  - `tree.Build()` in aggregator/pkg/merkle tests
  - `epochManager.FinalizeEpoch()` in aggregator/pkg/api/handlers.go:77

• **Nil pointer dereference**: gateway/pkg/api/handlers_test.go:117 accesses handler fields after nil check suggests it could be nil

• **Unused code**: 
  - Entire mockStream struct and methods in provider/pkg/stream/handler_test.go
  - const dhtTopic in gateway/pkg/p2p/client.go:20

## Concurrency Issues

• **Missing synchronization**: No mutex protection found for shared state in:
  - Provider sequence counter (handler.seq) incremented without locks
  - Gateway canary rate could be accessed concurrently
  - Aggregator epoch state transitions lack explicit locking

• **Rate limiter cleanup**: gateway/pkg/ratelimit has no goroutine to clean expired entries

## Error Handling

• **Silent failures**: Error returns ignored in critical paths:
  - Receipt storage operations could fail silently
  - Merkle tree building errors not propagated in tests
  - Network write operations don't check errors

• **Panic potential**: No recovery mechanisms for:
  - JSON marshaling failures in handlers
  - Network stream operations
  - Cryptographic operations

## Boundary Validation

• **Input validation gaps**:
  - No max receipt size validation in aggregator
  - Missing prompt content validation (only size checked)
  - Epoch numbers not validated for reasonable ranges
  - No validation of Merkle proof depth

• **Resource limits missing**:
  - No connection limits in P2P hosts
  - Unbounded receipt storage in aggregator
  - No timeout on LLM generation calls

## Security Vulnerabilities

• **Known CVEs detected**:
  - GO-2024-3302: QUIC-go ICMP injection (v0.41.0 → v0.48.2)
  - GO-2025-3595: golang.org/x/net vulnerability (v0.21.0 → v0.38.0)
  - GO-2024-3218: libp2p DHT censorship vulnerability (no fix available)
  - GO-2024-2687: golang.org/x/net issue (v0.21.0 → v0.23.0)
  - GO-2024-2682: QUIC-go vulnerability (v0.41.0 → v0.42.0)

• **No secret scanning**: Project lacks .secretlintrc configuration

## TODOs and Missing Implementation

• **License**: No LICENSE file or copyright headers found
• **Integration tests**: Docker-based tests not implemented
• **E2E tests**: Python test scripts missing
• **Benchmarks**: K6 scripts referenced but not created
• **Error recovery**: No retry logic for failed operations
• **Monitoring**: No metrics or health check implementations beyond basic endpoints
• **Configuration**: Hardcoded values throughout (ports, timeouts, limits)