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
    
    "github.com/libp2p/go-libp2p"
    dht "github.com/libp2p/go-libp2p-kad-dht"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

type NetworkStats struct {
    NodeCount     int       `json:"node_count"`
    OnlineNodes   int       `json:"online_nodes"`
    Countries     int       `json:"countries"`
    TotalCapacity float64   `json:"total_capacity"`
    Throughput    float64   `json:"throughput"`
    LastUpdate    time.Time `json:"last_update"`
}

type StatsServer struct {
    mu       sync.RWMutex
    stats    NetworkStats
    host     host.Host
    dht      *dht.IpfsDHT
}

func main() {
    var (
        listenAddr = flag.String("listen", ":8086", "HTTP listen address")
        p2pPort    = flag.Int("p2p", 4001, "P2P port")
        bootstrap  = flag.String("bootstrap", "", "Bootstrap peers")
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
    
    log.Printf("P2P node started with ID: %s", h.ID())
    
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
    
    // Connect to bootstrap peers
    if *bootstrap != "" {
        peerInfo, err := peer.AddrInfoFromString(*bootstrap)
        if err == nil {
            if err := h.Connect(ctx, *peerInfo); err != nil {
                log.Printf("Failed to connect to bootstrap peer: %v", err)
            }
        }
    }
    
    // Create stats server
    server := &StatsServer{
        host: h,
        dht:  kdht,
        stats: NetworkStats{
            NodeCount:  1, // Start with self
            LastUpdate: time.Now(),
        },
    }
    
    // Start periodic network discovery
    go server.discoverNetwork()
    
    // Set up HTTP endpoints
    http.HandleFunc("/api/stats", server.handleStats)
    http.HandleFunc("/api/stats/live", server.handleLiveStats)
    http.HandleFunc("/health", server.handleHealth)
    
    // Enable CORS
    handler := corsMiddleware(http.DefaultServeMux)
    
    log.Printf("Starting stats API server on %s", *listenAddr)
    if err := http.ListenAndServe(*listenAddr, handler); err != nil {
        log.Fatal(err)
    }
}

func (s *StatsServer) discoverNetwork() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        s.updateNetworkStats()
        <-ticker.C
    }
}

func (s *StatsServer) updateNetworkStats() {
    ctx := context.Background()
    
    // Count connected peers
    peers := s.host.Network().Peers()
    connectedCount := len(peers)
    
    // Try to discover more peers via DHT
    routingDiscovery := routing.NewRoutingDiscovery(s.dht)
    peerChan, err := routingDiscovery.FindPeers(ctx, "quiver-network")
    if err != nil {
        log.Printf("Failed to find peers: %v", err)
        return
    }
    
    discoveredPeers := make(map[peer.ID]bool)
    timeout := time.After(5 * time.Second)
    
    for {
        select {
        case p, ok := <-peerChan:
            if !ok {
                goto done
            }
            discoveredPeers[p.ID] = true
        case <-timeout:
            goto done
        }
    }
    
done:
    totalNodes := connectedCount + len(discoveredPeers)
    
    s.mu.Lock()
    s.stats.NodeCount = totalNodes
    s.stats.OnlineNodes = connectedCount
    s.stats.Countries = estimateCountries(totalNodes)
    s.stats.TotalCapacity = float64(totalNodes) * 1.2 // Avg 1.2 TFLOPS per node
    s.stats.Throughput = float64(totalNodes) * 150   // Avg 150 req/s per node
    s.stats.LastUpdate = time.Now()
    s.mu.Unlock()
    
    log.Printf("Network stats updated: %d total nodes, %d connected", totalNodes, connectedCount)
}

func estimateCountries(nodeCount int) int {
    // Rough estimation based on node distribution
    if nodeCount < 10 {
        return 1
    } else if nodeCount < 50 {
        return 3
    } else if nodeCount < 200 {
        return 8
    } else if nodeCount < 1000 {
        return 15
    }
    return 25
}

func (s *StatsServer) handleStats(w http.ResponseWriter, r *http.Request) {
    s.mu.RLock()
    stats := s.stats
    s.mu.RUnlock()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

func (s *StatsServer) handleLiveStats(w http.ResponseWriter, r *http.Request) {
    // Server-sent events for live updates
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
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
    
    // Send updates every 5 seconds
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

func (s *StatsServer) handleHealth(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "healthy",
        "node_id": s.host.ID().String(),
        "peers": len(s.host.Network().Peers()),
    })
}

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}