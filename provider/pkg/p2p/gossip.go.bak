package p2p

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    pubsub "github.com/libp2p/go-libp2p-pubsub"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
)

const (
    // ゴシップトピック
    GossipTopic = "/quiver/gossip/1.0.0"
    // ノード情報更新間隔
    NodeUpdateInterval = 15 * time.Second
)

// NodeInfo はノードの情報
type NodeInfo struct {
    PeerID       string    `json:"peer_id"`
    Address      []string  `json:"addresses"`
    Model        string    `json:"model"`
    Available    bool      `json:"available"`
    CPUUsage     float64   `json:"cpu_usage"`
    MemoryUsage  float64   `json:"memory_usage"`
    RequestCount int64     `json:"request_count"`
    LastSeen     time.Time `json:"last_seen"`
    Reputation   float64   `json:"reputation"`
}

// GossipMessage はゴシップメッセージ
type GossipMessage struct {
    Type      string          `json:"type"`
    NodeInfo  *NodeInfo       `json:"node_info,omitempty"`
    Request   *InferRequest   `json:"request,omitempty"`
    Response  *InferResponse  `json:"response,omitempty"`
}

// InferRequest は推論リクエスト
type InferRequest struct {
    ID        string `json:"id"`
    Model     string `json:"model"`
    Prompt    string `json:"prompt"`
    MaxTokens int    `json:"max_tokens"`
}

// InferResponse は推論レスポンス
type InferResponse struct {
    RequestID string `json:"request_id"`
    Result    string `json:"result"`
    NodeID    string `json:"node_id"`
    Error     string `json:"error,omitempty"`
}

// GossipManager はゴシッププロトコルを管理
type GossipManager struct {
    host       host.Host
    pubsub     *pubsub.PubSub
    topic      *pubsub.Topic
    sub        *pubsub.Subscription
    
    nodesMu    sync.RWMutex
    nodes      map[peer.ID]*NodeInfo
    
    requestsMu sync.RWMutex
    requests   map[string]chan *InferResponse
    
    ctx        context.Context
    cancel     context.CancelFunc
}

// NewGossipManager creates a new gossip manager
func NewGossipManager(h host.Host) (*GossipManager, error) {
    ctx, cancel := context.WithCancel(context.Background())
    
    // GossipSubを作成
    ps, err := pubsub.NewGossipSub(ctx, h)
    if err != nil {
        cancel()
        return nil, fmt.Errorf("failed to create pubsub: %w", err)
    }

    // トピックに参加
    topic, err := ps.Join(GossipTopic)
    if err != nil {
        cancel()
        return nil, fmt.Errorf("failed to join topic: %w", err)
    }

    // サブスクライブ
    sub, err := topic.Subscribe()
    if err != nil {
        cancel()
        return nil, fmt.Errorf("failed to subscribe: %w", err)
    }

    gm := &GossipManager{
        host:     h,
        pubsub:   ps,
        topic:    topic,
        sub:      sub,
        nodes:    make(map[peer.ID]*NodeInfo),
        requests: make(map[string]chan *InferResponse),
        ctx:      ctx,
        cancel:   cancel,
    }

    // メッセージ受信を開始
    go gm.readLoop()
    // ノード情報を定期的にブロードキャスト
    go gm.broadcastNodeInfo()

    return gm, nil
}

// readLoop はメッセージを受信
func (gm *GossipManager) readLoop() {
    for {
        msg, err := gm.sub.Next(gm.ctx)
        if err != nil {
            return
        }

        // 自分のメッセージは無視
        if msg.ReceivedFrom == gm.host.ID() {
            continue
        }

        var gossipMsg GossipMessage
        if err := json.Unmarshal(msg.Data, &gossipMsg); err != nil {
            continue
        }

        switch gossipMsg.Type {
        case "node_info":
            gm.handleNodeInfo(msg.ReceivedFrom, gossipMsg.NodeInfo)
        case "infer_request":
            go gm.handleInferRequest(gossipMsg.Request)
        case "infer_response":
            gm.handleInferResponse(gossipMsg.Response)
        }
    }
}

// handleNodeInfo はノード情報を処理
func (gm *GossipManager) handleNodeInfo(peerID peer.ID, info *NodeInfo) {
    gm.nodesMu.Lock()
    defer gm.nodesMu.Unlock()
    
    info.LastSeen = time.Now()
    gm.nodes[peerID] = info
}

// handleInferRequest は推論リクエストを処理
func (gm *GossipManager) handleInferRequest(req *InferRequest) {
    // ローカルで推論を実行（実装は省略）
    response := &InferResponse{
        RequestID: req.ID,
        Result:    fmt.Sprintf("Response to: %s", req.Prompt),
        NodeID:    gm.host.ID().String(),
    }

    // レスポンスをブロードキャスト
    gm.BroadcastMessage(&GossipMessage{
        Type:     "infer_response",
        Response: response,
    })
}

// handleInferResponse は推論レスポンスを処理
func (gm *GossipManager) handleInferResponse(resp *InferResponse) {
    gm.requestsMu.RLock()
    ch, exists := gm.requests[resp.RequestID]
    gm.requestsMu.RUnlock()
    
    if exists {
        select {
        case ch <- resp:
        default:
        }
    }
}

// broadcastNodeInfo はノード情報を定期的にブロードキャスト
func (gm *GossipManager) broadcastNodeInfo() {
    ticker := time.NewTicker(NodeUpdateInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            info := &NodeInfo{
                PeerID:       gm.host.ID().String(),
                Address:      gm.host.Addrs().Strings(),
                Model:        "llama3.2",
                Available:    true,
                CPUUsage:     30.0, // 実際の値を取得
                MemoryUsage:  2.5,  // 実際の値を取得
                RequestCount: 100,  // 実際の値を取得
                Reputation:   0.95,
            }

            gm.BroadcastMessage(&GossipMessage{
                Type:     "node_info",
                NodeInfo: info,
            })
        case <-gm.ctx.Done():
            return
        }
    }
}

// BroadcastMessage はメッセージをブロードキャスト
func (gm *GossipManager) BroadcastMessage(msg *GossipMessage) error {
    data, err := json.Marshal(msg)
    if err != nil {
        return err
    }
    
    return gm.topic.Publish(gm.ctx, data)
}

// SendInferRequest は推論リクエストを送信
func (gm *GossipManager) SendInferRequest(req *InferRequest) (*InferResponse, error) {
    // レスポンスチャネルを作成
    respCh := make(chan *InferResponse, 1)
    
    gm.requestsMu.Lock()
    gm.requests[req.ID] = respCh
    gm.requestsMu.Unlock()
    
    defer func() {
        gm.requestsMu.Lock()
        delete(gm.requests, req.ID)
        gm.requestsMu.Unlock()
    }()

    // リクエストをブロードキャスト
    if err := gm.BroadcastMessage(&GossipMessage{
        Type:    "infer_request",
        Request: req,
    }); err != nil {
        return nil, err
    }

    // レスポンスを待つ
    select {
    case resp := <-respCh:
        return resp, nil
    case <-time.After(30 * time.Second):
        return nil, fmt.Errorf("request timeout")
    case <-gm.ctx.Done():
        return nil, gm.ctx.Err()
    }
}

// GetActiveNodes はアクティブなノードのリストを返す
func (gm *GossipManager) GetActiveNodes() []*NodeInfo {
    gm.nodesMu.RLock()
    defer gm.nodesMu.RUnlock()
    
    var nodes []*NodeInfo
    cutoff := time.Now().Add(-60 * time.Second)
    
    for _, node := range gm.nodes {
        if node.LastSeen.After(cutoff) && node.Available {
            nodes = append(nodes, node)
        }
    }
    
    return nodes
}

// SelectBestNode は最適なノードを選択
func (gm *GossipManager) SelectBestNode(model string) *NodeInfo {
    nodes := gm.GetActiveNodes()
    
    var bestNode *NodeInfo
    bestScore := 0.0
    
    for _, node := range nodes {
        if node.Model != model {
            continue
        }
        
        // スコア計算（CPU使用率が低く、評判が高いノードを優先）
        score := node.Reputation * (1.0 - node.CPUUsage/100.0)
        
        if score > bestScore {
            bestScore = score
            bestNode = node
        }
    }
    
    return bestNode
}

// Close shuts down the gossip manager
func (gm *GossipManager) Close() error {
    gm.cancel()
    return gm.topic.Close()
}