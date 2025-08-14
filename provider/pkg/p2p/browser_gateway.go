package p2p

import (
    "context"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"
    
    "github.com/quic-go/quic-go/http3"
    "github.com/quic-go/webtransport-go"
    "github.com/quiver/provider/pkg/hll"
)

// BrowserGateway handles WebTransport connections from browsers
type BrowserGateway struct {
    server      *webtransport.Server
    node        *SimpleNode
    hllSketch   *hll.HyperLogLog
    mu          sync.RWMutex
    connections map[string]*BrowserConnection
}

// BrowserConnection represents a connected browser node
type BrowserConnection struct {
    session    *webtransport.Session
    nodeID     string
    lastSeen   time.Time
    hllSketch  *hll.HyperLogLog
}

// Message types for browser communication
type BrowserMessage struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

type HLLRequest struct {
    RequestID string `json:"request_id"`
}

type HLLResponse struct {
    RequestID    string `json:"request_id"`
    HLLData      []byte `json:"hll_data"`
    NodeCount    uint64 `json:"node_count"`
    Timestamp    int64  `json:"timestamp"`
}

type NetworkStats struct {
    NodeCount    uint64   `json:"node_count"`
    OnlineNodes  []string `json:"online_nodes"`
    Timestamp    int64    `json:"timestamp"`
}

// NewBrowserGateway creates a new browser gateway
func NewBrowserGateway(node *SimpleNode, certFile, keyFile string) (*BrowserGateway, error) {
    // Load TLS certificate
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        // Use self-signed cert for development
        cert, err = generateSelfSignedCert()
        if err != nil {
            return nil, fmt.Errorf("failed to generate certificate: %w", err)
        }
    }
    
    // Create WebTransport server
    server := &webtransport.Server{
        H3: http3.Server{
            TLSConfig: &tls.Config{
                Certificates: []tls.Certificate{cert},
                NextProtos:   []string{"h3"},
            },
            Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Access-Control-Allow-Origin", "*")
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
                
                if r.Method == "OPTIONS" {
                    w.WriteHeader(http.StatusOK)
                    return
                }
                
                // Handle regular HTTP requests
                if r.URL.Path == "/stats" {
                    handleStatsRequest(w, r, node)
                    return
                }
                
                w.WriteHeader(http.StatusNotFound)
            }),
        },
    }
    
    gateway := &BrowserGateway{
        server:      server,
        node:        node,
        hllSketch:   hll.New(),
        connections: make(map[string]*BrowserConnection),
    }
    
    // Add the node itself to HLL
    gateway.hllSketch.AddNode(node.GetPeerID())
    
    return gateway, nil
}

// Start starts the browser gateway
func (bg *BrowserGateway) Start(addr string) error {
    // Set up WebTransport endpoint
    http.HandleFunc("/webtransport", bg.handleWebTransport)
    
    log.Printf("Starting browser gateway on %s", addr)
    return bg.server.ListenAndServe()
}

// handleWebTransport handles new WebTransport connections
func (bg *BrowserGateway) handleWebTransport(w http.ResponseWriter, r *http.Request) {
    session, err := bg.server.Upgrade(w, r)
    if err != nil {
        log.Printf("Failed to upgrade to WebTransport: %v", err)
        return
    }
    
    // Generate node ID for browser
    nodeID := fmt.Sprintf("browser-%d", time.Now().UnixNano())
    
    conn := &BrowserConnection{
        session:   session,
        nodeID:    nodeID,
        lastSeen:  time.Now(),
        hllSketch: hll.New(),
    }
    
    bg.mu.Lock()
    bg.connections[nodeID] = conn
    bg.hllSketch.AddNode(nodeID)
    bg.mu.Unlock()
    
    log.Printf("Browser node connected: %s", nodeID)
    
    // Handle connection
    go bg.handleConnection(conn)
}

// handleConnection handles messages from a browser connection
func (bg *BrowserGateway) handleConnection(conn *BrowserConnection) {
    defer func() {
        bg.mu.Lock()
        delete(bg.connections, conn.nodeID)
        bg.mu.Unlock()
        conn.session.CloseWithError(0, "")
        log.Printf("Browser node disconnected: %s", conn.nodeID)
    }()
    
    ctx := context.Background()
    
    for {
        stream, err := conn.session.AcceptStream(ctx)
        if err != nil {
            log.Printf("Failed to accept stream: %v", err)
            return
        }
        
        go bg.handleStream(conn, stream)
    }
}

// handleStream handles a single message stream
func (bg *BrowserGateway) handleStream(conn *BrowserConnection, stream webtransport.Stream) {
    defer stream.Close()
    
    decoder := json.NewDecoder(stream)
    var msg BrowserMessage
    
    if err := decoder.Decode(&msg); err != nil {
        log.Printf("Failed to decode message: %v", err)
        return
    }
    
    switch msg.Type {
    case "hll_request":
        bg.handleHLLRequest(conn, stream, msg.Payload)
    case "ping":
        bg.handlePing(conn, stream)
    case "get_stats":
        bg.handleGetStats(conn, stream)
    default:
        log.Printf("Unknown message type: %s", msg.Type)
    }
}

// handleHLLRequest handles HLL data requests
func (bg *BrowserGateway) handleHLLRequest(conn *BrowserConnection, stream webtransport.Stream, payload json.RawMessage) {
    var req HLLRequest
    if err := json.Unmarshal(payload, &req); err != nil {
        log.Printf("Failed to unmarshal HLL request: %v", err)
        return
    }
    
    bg.mu.RLock()
    // Merge HLL sketches from all connections
    merged := hll.New()
    merged.Merge(bg.hllSketch)
    
    for _, c := range bg.connections {
        merged.Merge(c.hllSketch)
    }
    bg.mu.RUnlock()
    
    // Get network stats from the P2P node
    p2pStats := bg.node.GetNetworkStats()
    if p2pNodes, ok := p2pStats["total_nodes"].(int); ok && p2pNodes > 0 {
        // Add P2P nodes to HLL
        for i := 0; i < p2pNodes; i++ {
            merged.AddNode(fmt.Sprintf("p2p-node-%d", i))
        }
    }
    
    response := HLLResponse{
        RequestID: req.RequestID,
        HLLData:   merged.Export(),
        NodeCount: merged.Estimate(),
        Timestamp: time.Now().Unix(),
    }
    
    encoder := json.NewEncoder(stream)
    if err := encoder.Encode(response); err != nil {
        log.Printf("Failed to encode HLL response: %v", err)
    }
}

// handlePing handles ping messages
func (bg *BrowserGateway) handlePing(conn *BrowserConnection, stream webtransport.Stream) {
    conn.lastSeen = time.Now()
    
    response := map[string]interface{}{
        "type":      "pong",
        "timestamp": time.Now().Unix(),
    }
    
    encoder := json.NewEncoder(stream)
    encoder.Encode(response)
}

// handleGetStats returns network statistics
func (bg *BrowserGateway) handleGetStats(conn *BrowserConnection, stream webtransport.Stream) {
    bg.mu.RLock()
    browserCount := len(bg.connections)
    bg.mu.RUnlock()
    
    p2pStats := bg.node.GetNetworkStats()
    
    stats := map[string]interface{}{
        "browser_nodes": browserCount,
        "p2p_nodes":     p2pStats["total_nodes"],
        "total_nodes":   browserCount + p2pStats["total_nodes"].(int),
        "timestamp":     time.Now().Unix(),
    }
    
    encoder := json.NewEncoder(stream)
    encoder.Encode(stats)
}

// GetHLLSketch returns the current HLL sketch
func (bg *BrowserGateway) GetHLLSketch() *hll.HyperLogLog {
    bg.mu.RLock()
    defer bg.mu.RUnlock()
    
    // Create a merged sketch
    merged := hll.New()
    merged.Merge(bg.hllSketch)
    
    for _, conn := range bg.connections {
        merged.Merge(conn.hllSketch)
    }
    
    return merged
}

// Cleanup removes stale connections
func (bg *BrowserGateway) Cleanup() {
    bg.mu.Lock()
    defer bg.mu.Unlock()
    
    now := time.Now()
    for nodeID, conn := range bg.connections {
        if now.Sub(conn.lastSeen) > 5*time.Minute {
            delete(bg.connections, nodeID)
            conn.session.CloseWithError(0, "timeout")
            log.Printf("Cleaned up stale browser connection: %s", nodeID)
        }
    }
}

// handleStatsRequest handles HTTP stats requests
func handleStatsRequest(w http.ResponseWriter, r *http.Request, node *SimpleNode) {
    stats := node.GetNetworkStats()
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    json.NewEncoder(w).Encode(stats)
}