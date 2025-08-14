package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/libp2p/go-libp2p/core/crypto"
    "github.com/quiver/provider/pkg/p2p"
)

func main() {
    var (
        listenAddr = flag.String("listen", ":4433", "WebTransport listen address")
        certFile   = flag.String("cert", "", "TLS certificate file")
        keyFile    = flag.String("key", "", "TLS key file")
        p2pPort    = flag.Int("p2p-port", 0, "P2P port (0 for random)")
        bootstrap  = flag.String("bootstrap", "", "Bootstrap peers (comma-separated)")
    )
    flag.Parse()
    
    // Create P2P node
    privKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
    if err != nil {
        log.Fatalf("Failed to generate key: %v", err)
    }
    
    nodeConfig := &p2p.NodeConfig{
        ListenAddrs: []string{
            fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *p2pPort),
        },
        PrivateKey: privKey,
    }
    
    if *bootstrap != "" {
        nodeConfig.BootstrapPeers = []string{*bootstrap}
    }
    
    node, err := p2p.NewDecentralizedNode(nodeConfig)
    if err != nil {
        log.Fatalf("Failed to create P2P node: %v", err)
    }
    defer node.Close()
    
    log.Printf("P2P node started with ID: %s", node.GetPeerID())
    
    // Create browser gateway
    gateway, err := p2p.NewBrowserGateway(node, *certFile, *keyFile)
    if err != nil {
        log.Fatalf("Failed to create browser gateway: %v", err)
    }
    
    // Start periodic cleanup
    go func() {
        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()
        
        for range ticker.C {
            gateway.Cleanup()
        }
    }()
    
    // Add HTTP endpoints for stats
    http.HandleFunc("/api/network/stats", func(w http.ResponseWriter, r *http.Request) {
        stats := node.GetNetworkStats()
        hllSketch := gateway.GetHLLSketch()
        
        stats["estimated_nodes"] = hllSketch.Estimate()
        stats["gateway_url"] = fmt.Sprintf("https://%s", *listenAddr)
        
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Access-Control-Allow-Origin", "*")
        json.NewEncoder(w).Encode(stats)
    })
    
    // Start servers in goroutines
    errChan := make(chan error, 2)
    
    // WebTransport server
    go func() {
        log.Printf("Starting WebTransport gateway on %s", *listenAddr)
        if err := gateway.Start(*listenAddr); err != nil {
            errChan <- fmt.Errorf("WebTransport server error: %w", err)
        }
    }()
    
    // HTTP server for API
    go func() {
        httpAddr := ":8085"
        log.Printf("Starting HTTP API on %s", httpAddr)
        if err := http.ListenAndServe(httpAddr, nil); err != nil {
            errChan <- fmt.Errorf("HTTP server error: %w", err)
        }
    }()
    
    // Handle shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    select {
    case sig := <-sigChan:
        log.Printf("Received signal %v, shutting down...", sig)
    case err := <-errChan:
        log.Printf("Server error: %v", err)
    }
}