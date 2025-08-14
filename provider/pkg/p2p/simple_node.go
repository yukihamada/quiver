package p2p

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/crypto"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
)

// NodeInfo represents node information
type NodeInfo struct {
    PeerID     string    `json:"peer_id"`
    Addresses  []string  `json:"addresses"`
    CPU        float64   `json:"cpu"`
    Memory     float64   `json:"memory"`
    Models     []string  `json:"models"`
    Reputation float64   `json:"reputation"`
    Timestamp  time.Time `json:"timestamp"`
}

// SimpleNode represents a basic P2P node without pubsub
type SimpleNode struct {
    host       host.Host
    ctx        context.Context
    cancel     context.CancelFunc
    mu         sync.RWMutex
    reputation *ReputationManager
    nodeInfo   *NodeInfo
}

// NodeConfig configuration for the node
type NodeConfig struct {
    ListenAddrs    []string
    BootstrapPeers []string
    PrivateKey     crypto.PrivKey
}

// NewDecentralizedNode creates a new simple P2P node
func NewDecentralizedNode(cfg *NodeConfig) (*SimpleNode, error) {
    ctx, cancel := context.WithCancel(context.Background())
    
    // Create libp2p host
    var opts []libp2p.Option
    if cfg.PrivateKey != nil {
        opts = append(opts, libp2p.Identity(cfg.PrivateKey))
    }
    
    for _, addr := range cfg.ListenAddrs {
        opts = append(opts, libp2p.ListenAddrStrings(addr))
    }
    
    h, err := libp2p.New(opts...)
    if err != nil {
        cancel()
        return nil, fmt.Errorf("failed to create host: %w", err)
    }
    
    node := &SimpleNode{
        host:       h,
        ctx:        ctx,
        cancel:     cancel,
        reputation: NewReputationManager(),
        nodeInfo: &NodeInfo{
            PeerID:     h.ID().String(),
            Addresses:  []string{},
            CPU:        75.0,
            Memory:     50.0,
            Models:     []string{"llama3.2"},
            Reputation: 1.0,
            Timestamp:  time.Now(),
        },
    }
    
    // Connect to bootstrap peers
    for _, peerAddr := range cfg.BootstrapPeers {
        go func(addr string) {
            if addr == "" {
                return
            }
            peerInfo, err := peer.AddrInfoFromString(addr)
            if err != nil {
                fmt.Printf("Failed to parse bootstrap peer %s: %v\n", addr, err)
                return
            }
            if err := h.Connect(ctx, *peerInfo); err != nil {
                fmt.Printf("Failed to connect to bootstrap peer %s: %v\n", addr, err)
            }
        }(peerAddr)
    }
    
    return node, nil
}

// Close shuts down the node
func (n *SimpleNode) Close() error {
    n.cancel()
    return n.host.Close()
}

// GetNetworkStats returns network statistics
func (n *SimpleNode) GetNetworkStats() map[string]interface{} {
    n.mu.RLock()
    defer n.mu.RUnlock()
    
    peers := n.host.Network().Peers()
    
    return map[string]interface{}{
        "total_nodes":     len(peers),
        "connected_peers": len(peers),
        "total_capacity":  float64(len(peers)) * 75.0,
        "avg_reputation":  1.0,
        "models_available": []string{"llama3.2"},
        "dht_routing_table_size": 0,
    }
}

// SendInferenceRequest sends an inference request (simplified version)
func (n *SimpleNode) SendInferenceRequest(model string, prompt string, maxTokens int) (string, error) {
    // In standalone mode, just return a mock response
    return fmt.Sprintf("Mock response for: %s", prompt), nil
}

// GetPeerID returns the peer ID
func (n *SimpleNode) GetPeerID() string {
    return n.host.ID().String()
}

// StartAsBootstrapNode marks this as a bootstrap node
func (n *SimpleNode) StartAsBootstrapNode() {
    fmt.Println("Starting as bootstrap node...")
}

// Alias for SimpleNode
type DecentralizedNode = SimpleNode