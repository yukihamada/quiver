package p2p

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/crypto"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/multiformats/go-multiaddr"
)

// DecentralizedNode は完全分散型ノード
type DecentralizedNode struct {
    host    host.Host
    dht     *DHTManager
    gossip  *GossipManager
    ctx     context.Context
    cancel  context.CancelFunc
}

// NodeConfig はノードの設定
type NodeConfig struct {
    ListenAddrs    []string
    BootstrapPeers []string
    PrivateKey     crypto.PrivKey
}

// NewDecentralizedNode creates a new decentralized node
func NewDecentralizedNode(cfg *NodeConfig) (*DecentralizedNode, error) {
    ctx, cancel := context.WithCancel(context.Background())

    // libp2pホストを作成
    var opts []libp2p.Option
    
    if cfg.PrivateKey != nil {
        opts = append(opts, libp2p.Identity(cfg.PrivateKey))
    }
    
    // リッスンアドレスを設定
    listenAddrs := cfg.ListenAddrs
    if len(listenAddrs) == 0 {
        listenAddrs = []string{
            "/ip4/0.0.0.0/tcp/0",
            "/ip4/0.0.0.0/udp/0/quic",
        }
    }
    
    for _, addr := range listenAddrs {
        maddr, err := multiaddr.NewMultiaddr(addr)
        if err != nil {
            continue
        }
        opts = append(opts, libp2p.ListenAddrs(maddr))
    }
    
    // 追加オプション
    opts = append(opts,
        libp2p.DefaultTransports,
        libp2p.DefaultMuxers,
        libp2p.DefaultSecurity,
        libp2p.NATPortMap(),
        libp2p.EnableAutoRelay(),
        libp2p.EnableNATService(),
    )

    // ホストを作成
    h, err := libp2p.New(opts...)
    if err != nil {
        cancel()
        return nil, fmt.Errorf("failed to create host: %w", err)
    }

    fmt.Printf("Node started with ID: %s\n", h.ID())
    for _, addr := range h.Addrs() {
        fmt.Printf("Listening on: %s/p2p/%s\n", addr, h.ID())
    }

    // DHTマネージャーを作成
    dhtMgr, err := NewDHTManager(h)
    if err != nil {
        h.Close()
        cancel()
        return nil, fmt.Errorf("failed to create DHT manager: %w", err)
    }

    // ゴシップマネージャーを作成
    gossipMgr, err := NewGossipManager(h)
    if err != nil {
        dhtMgr.Close()
        h.Close()
        cancel()
        return nil, fmt.Errorf("failed to create gossip manager: %w", err)
    }

    node := &DecentralizedNode{
        host:   h,
        dht:    dhtMgr,
        gossip: gossipMgr,
        ctx:    ctx,
        cancel: cancel,
    }

    // ブートストラップピアに接続
    if len(cfg.BootstrapPeers) > 0 {
        if err := dhtMgr.ConnectToBootstrapPeers(cfg.BootstrapPeers); err != nil {
            fmt.Printf("Warning: failed to connect to some bootstrap peers: %v\n", err)
        }
    }

    // サービスをアドバタイズ
    go func() {
        time.Sleep(5 * time.Second) // DHTが安定するまで待つ
        if err := dhtMgr.Advertise("quiver-inference"); err != nil {
            fmt.Printf("Failed to advertise service: %v\n", err)
        }
    }()

    return node, nil
}

// SendInferenceRequest は推論リクエストを送信
func (n *DecentralizedNode) SendInferenceRequest(model, prompt string, maxTokens int) (string, error) {
    req := &InferRequest{
        ID:        uuid.New().String(),
        Model:     model,
        Prompt:    prompt,
        MaxTokens: maxTokens,
    }

    // 最適なノードを選択
    bestNode := n.gossip.SelectBestNode(model)
    if bestNode != nil {
        fmt.Printf("Selected node %s for inference (CPU: %.1f%%, Reputation: %.2f)\n",
            bestNode.PeerID, bestNode.CPUUsage, bestNode.Reputation)
    }

    // リクエストをブロードキャスト
    resp, err := n.gossip.SendInferRequest(req)
    if err != nil {
        return "", err
    }

    return resp.Result, nil
}

// GetNetworkStats はネットワーク統計を取得
func (n *DecentralizedNode) GetNetworkStats() map[string]interface{} {
    activeNodes := n.gossip.GetActiveNodes()
    
    totalNodes := len(activeNodes)
    totalCapacity := 0.0
    avgReputation := 0.0
    
    modelCounts := make(map[string]int)
    
    for _, node := range activeNodes {
        totalCapacity += (100.0 - node.CPUUsage)
        avgReputation += node.Reputation
        modelCounts[node.Model]++
    }
    
    if totalNodes > 0 {
        avgReputation /= float64(totalNodes)
    }
    
    return map[string]interface{}{
        "total_nodes":      totalNodes,
        "connected_peers":  len(n.host.Network().Peers()),
        "total_capacity":   totalCapacity,
        "avg_reputation":   avgReputation,
        "models_available": modelCounts,
        "dht_routing_table_size": n.dht.GetDHT().RoutingTable().Size(),
    }
}

// JoinNetwork は既存のネットワークに参加
func (n *DecentralizedNode) JoinNetwork(peerAddr string) error {
    maddr, err := multiaddr.NewMultiaddr(peerAddr)
    if err != nil {
        return fmt.Errorf("invalid multiaddr: %w", err)
    }

    peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
    if err != nil {
        return fmt.Errorf("failed to parse peer info: %w", err)
    }

    if err := n.host.Connect(n.ctx, *peerInfo); err != nil {
        return fmt.Errorf("failed to connect to peer: %w", err)
    }

    fmt.Printf("Successfully joined network via peer: %s\n", peerInfo.ID)
    return nil
}

// StartAsBootstrapNode はブートストラップノードとして起動
func (n *DecentralizedNode) StartAsBootstrapNode() {
    fmt.Println("Starting as bootstrap node...")
    fmt.Println("Other nodes can connect using:")
    for _, addr := range n.host.Addrs() {
        fmt.Printf("  %s/p2p/%s\n", addr, n.host.ID())
    }
}

// GetPeerID returns the node's peer ID
func (n *DecentralizedNode) GetPeerID() peer.ID {
    return n.host.ID()
}

// GetListenAddresses returns the node's listen addresses
func (n *DecentralizedNode) GetListenAddresses() []multiaddr.Multiaddr {
    return n.host.Addrs()
}

// Close shuts down the node
func (n *DecentralizedNode) Close() error {
    n.cancel()
    
    if err := n.gossip.Close(); err != nil {
        fmt.Printf("Error closing gossip: %v\n", err)
    }
    
    if err := n.dht.Close(); err != nil {
        fmt.Printf("Error closing DHT: %v\n", err)
    }
    
    return n.host.Close()
}