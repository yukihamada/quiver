package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/libp2p/go-libp2p"
    dht "github.com/libp2p/go-libp2p-kad-dht"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/p2p/discovery/routing"
    "github.com/multiformats/go-multiaddr"
)

func main() {
    ctx := context.Background()
    
    // Create a new P2P host
    h, err := libp2p.New(
        libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/4005"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer h.Close()
    
    fmt.Printf("Node ID: %s\n", h.ID())
    
    // Create DHT
    kdht, err := dht.New(ctx, h)
    if err != nil {
        log.Fatal(err)
    }
    
    // Bootstrap DHT
    if err := kdht.Bootstrap(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Connect to bootstrap node
    bootstrapAddr := "/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWNFmgqVZJdWNkBAShVJEugVdBUwvZNexWMkiqp9ayDatb"
    addr, _ := multiaddr.NewMultiaddr(bootstrapAddr)
    peerinfo, _ := peer.AddrInfoFromP2pAddr(addr)
    
    if err := h.Connect(ctx, *peerinfo); err != nil {
        log.Printf("Failed to connect to bootstrap: %v", err)
    } else {
        fmt.Println("Connected to bootstrap node")
    }
    
    // Wait a moment for connections
    time.Sleep(2 * time.Second)
    
    // Count connected peers
    peers := h.Network().Peers()
    fmt.Printf("\nDirectly connected peers: %d\n", len(peers))
    for _, p := range peers {
        fmt.Printf("  - %s\n", p)
    }
    
    // Try to discover peers via DHT
    fmt.Println("\nDiscovering peers via DHT...")
    routingDiscovery := routing.NewRoutingDiscovery(kdht)
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
            fmt.Printf("  Discovered: %s\n", p.ID)
        case <-timeout:
            goto done
        }
    }
    
done:
    fmt.Printf("\nTotal discovered peers via DHT: %d\n", len(discoveredPeers))
    
    // Get routing table size
    rtSize := kdht.RoutingTable().Size()
    fmt.Printf("DHT Routing table size: %d\n", rtSize)
    
    // Summary
    fmt.Printf("\n=== Network Status ===\n")
    fmt.Printf("Bootstrap node: Connected\n")
    fmt.Printf("Direct connections: %d\n", len(peers))
    fmt.Printf("DHT discovered peers: %d\n", len(discoveredPeers))
    fmt.Printf("DHT routing table: %d\n", rtSize)
    fmt.Printf("Total unique nodes in network: %d\n", len(peers) + len(discoveredPeers))
}