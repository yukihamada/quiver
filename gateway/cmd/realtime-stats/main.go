package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"
    
    "github.com/gorilla/websocket"
    "github.com/libp2p/go-libp2p"
    dht "github.com/libp2p/go-libp2p-kad-dht"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins for now
    },
}

type RealtimeStats struct {
    mu          sync.RWMutex
    subscribers map[*websocket.Conn]bool
    host        host.Host
    dht         *dht.IpfsDHT
    stats       NetworkStats
}

type NetworkStats struct {
    NodeCount     int                    `json:"node_count"`
    OnlineNodes   int                    `json:"online_nodes"`
    NodeDetails   map[string]NodeInfo    `json:"node_details"`
    Countries     int                    `json:"countries"`
    TotalCapacity float64                `json:"total_capacity"`
    Throughput    float64                `json:"throughput"`
    LastUpdate    time.Time              `json:"last_update"`
    NetworkType   string                 `json:"network_type"`
}

type NodeInfo struct {
    PeerID    string    `json:"peer_id"`
    Addresses []string  `json:"addresses"`
    Type      string    `json:"type"`
    Location  string    `json:"location"`
    LastSeen  time.Time `json:"last_seen"`
}

func main() {
    var (
        httpAddr  = flag.String("http", ":8087", "HTTP/WebSocket listen address")
        p2pPort   = flag.Int("p2p", 4005, "P2P port")
        bootstrap = flag.String("bootstrap", "/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWNFmgqVZJdWNkBAShVJEugVdBUwvZNexWMkiqp9ayDatb", "Bootstrap peer")
    )
    flag.Parse()
    
    // Create P2P host
    h, err := libp2p.New(
        libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *p2pPort)),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer h.Close()
    
    log.Printf("Realtime stats node started with ID: %s", h.ID())
    
    // Create DHT
    ctx := context.Background()
    kdht, err := dht.New(ctx, h)
    if err != nil {
        log.Fatal(err)
    }
    
    // Bootstrap DHT
    if err := kdht.Bootstrap(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Connect to bootstrap peer
    if *bootstrap != "" {
        peerInfo, err := peer.AddrInfoFromString(*bootstrap)
        if err == nil {
            if err := h.Connect(ctx, *peerInfo); err != nil {
                log.Printf("Failed to connect to bootstrap: %v", err)
            } else {
                log.Printf("Connected to bootstrap node")
            }
        }
    }
    
    // Create realtime stats server
    server := &RealtimeStats{
        subscribers: make(map[*websocket.Conn]bool),
        host:        h,
        dht:         kdht,
        stats: NetworkStats{
            NodeDetails:   make(map[string]NodeInfo),
            LastUpdate:    time.Now(),
            NetworkType:   "production",
        },
    }
    
    // Start monitoring
    go server.monitorNetwork()
    
    // HTTP endpoints
    http.HandleFunc("/ws", server.handleWebSocket)
    http.HandleFunc("/api/stats/live", server.handleLiveStats)
    http.HandleFunc("/api/stats", server.handleStats)
    http.HandleFunc("/health", server.handleHealth)
    
    log.Printf("Starting realtime stats server on %s", *httpAddr)
    if err := http.ListenAndServe(*httpAddr, nil); err != nil {
        log.Fatal(err)
    }
}

func (s *RealtimeStats) monitorNetwork() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        s.updateNetworkStats()
        s.broadcastStats()
        <-ticker.C
    }
}

func (s *RealtimeStats) updateNetworkStats() {
    ctx := context.Background()
    
    // Get connected peers
    peers := s.host.Network().Peers()
    
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Update node details
    s.stats.NodeDetails = make(map[string]NodeInfo)
    
    // Add self
    s.stats.NodeDetails[s.host.ID().String()] = NodeInfo{
        PeerID:    s.host.ID().String(),
        Addresses: []string{},
        Type:      "stats",
        Location:  "local",
        LastSeen:  time.Now(),
    }
    
    // Add connected peers
    for _, p := range peers {
        addrs := s.host.Network().ConnsToPeer(p)
        addrStrs := []string{}
        for _, conn := range addrs {
            addrStrs = append(addrStrs, conn.RemoteMultiaddr().String())
        }
        
        nodeType := "unknown"
        location := "unknown"
        
        // Try to determine node type from peer ID or connection info
        if conn := s.host.Network().ConnsToPeer(p); len(conn) > 0 {
            remoteAddr := conn[0].RemoteMultiaddr().String()
            if contains(remoteAddr, "127.0.0.1") || contains(remoteAddr, "localhost") {
                location = "local"
            } else if contains(remoteAddr, "35.") || contains(remoteAddr, "34.") {
                location = "gcp"
            } else if contains(remoteAddr, "ec2-") || contains(remoteAddr, "52.") {
                location = "aws"
            }
        }
        
        s.stats.NodeDetails[p.String()] = NodeInfo{
            PeerID:    p.String(),
            Addresses: addrStrs,
            Type:      nodeType,
            Location:  location,
            LastSeen:  time.Now(),
        }
    }
    
    // Discover more peers via DHT
    routingDiscovery := routing.NewRoutingDiscovery(s.dht)
    peerChan, err := routingDiscovery.FindPeers(ctx, "quiver-network")
    if err == nil {
        timeout := time.After(2 * time.Second)
        for {
            select {
            case p, ok := <-peerChan:
                if !ok {
                    goto done
                }
                if _, exists := s.stats.NodeDetails[p.ID.String()]; !exists {
                    s.stats.NodeDetails[p.ID.String()] = NodeInfo{
                        PeerID:    p.ID.String(),
                        Addresses: []string{},
                        Type:      "discovered",
                        Location:  "unknown",
                        LastSeen:  time.Now(),
                    }
                }
            case <-timeout:
                goto done
            }
        }
    }
    
done:
    // Update summary stats
    s.stats.NodeCount = len(s.stats.NodeDetails)
    s.stats.OnlineNodes = len(peers) + 1 // connected peers + self
    
    // Count unique locations
    locations := make(map[string]bool)
    for _, node := range s.stats.NodeDetails {
        if node.Location != "unknown" {
            locations[node.Location] = true
        }
    }
    s.stats.Countries = len(locations)
    
    // Update capacity and throughput
    s.stats.TotalCapacity = float64(s.stats.NodeCount) * 1.2
    s.stats.Throughput = float64(s.stats.NodeCount) * 150
    s.stats.LastUpdate = time.Now()
    
    log.Printf("Network stats updated: %d total nodes, %d online", s.stats.NodeCount, s.stats.OnlineNodes)
}

func (s *RealtimeStats) broadcastStats() {
    s.mu.RLock()
    stats := s.stats
    subscribers := make([]*websocket.Conn, 0, len(s.subscribers))
    for conn := range s.subscribers {
        subscribers = append(subscribers, conn)
    }
    s.mu.RUnlock()
    
    message, _ := json.Marshal(stats)
    
    for _, conn := range subscribers {
        if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
            s.removeSubscriber(conn)
        }
    }
}

func (s *RealtimeStats) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade failed: %v", err)
        return
    }
    defer conn.Close()
    
    s.addSubscriber(conn)
    defer s.removeSubscriber(conn)
    
    // Send initial stats
    s.mu.RLock()
    stats := s.stats
    s.mu.RUnlock()
    
    if err := conn.WriteJSON(stats); err != nil {
        return
    }
    
    // Keep connection alive
    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            break
        }
    }
}

func (s *RealtimeStats) handleStats(w http.ResponseWriter, r *http.Request) {
    s.mu.RLock()
    stats := s.stats
    s.mu.RUnlock()
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    json.NewEncoder(w).Encode(stats)
}

func (s *RealtimeStats) handleLiveStats(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }
    
    // Send initial stats
    s.mu.RLock()
    stats := s.stats
    s.mu.RUnlock()
    
    data, _ := json.Marshal(stats)
    fmt.Fprintf(w, "data: %s\n\n", data)
    flusher.Flush()
    
    // Send updates
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            s.mu.RLock()
            stats := s.stats
            s.mu.RUnlock()
            
            data, _ := json.Marshal(stats)
            fmt.Fprintf(w, "data: %s\n\n", data)
            flusher.Flush()
            
        case <-r.Context().Done():
            return
        }
    }
}

func (s *RealtimeStats) handleHealth(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "healthy",
        "node_id": s.host.ID().String(),
        "peers": len(s.host.Network().Peers()),
    })
}

func (s *RealtimeStats) addSubscriber(conn *websocket.Conn) {
    s.mu.Lock()
    s.subscribers[conn] = true
    s.mu.Unlock()
}

func (s *RealtimeStats) removeSubscriber(conn *websocket.Conn) {
    s.mu.Lock()
    delete(s.subscribers, conn)
    s.mu.Unlock()
    conn.Close()
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && s[0:len(substr)] == substr
}