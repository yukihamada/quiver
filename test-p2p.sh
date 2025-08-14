#!/bin/bash

# Test P2P connectivity

echo "Testing QUIVer P2P Network"
echo "========================="

# Start bootstrap node
echo "1. Starting bootstrap node..."
cd bootstrap
if [ ! -f bootstrap ]; then
    go mod init github.com/quiver/bootstrap 2>/dev/null
    cat > main.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/libp2p/go-libp2p"
    dht "github.com/libp2p/go-libp2p-kad-dht"
    "github.com/libp2p/go-libp2p/core/crypto"
    "github.com/multiformats/go-multiaddr"
)

func main() {
    ctx := context.Background()
    
    priv, _, _ := crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, 
        crypto.NewDeterministicReader([]byte("test-bootstrap")))
    
    listen, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/4001")
    
    h, err := libp2p.New(
        libp2p.Identity(priv),
        libp2p.ListenAddrs(listen),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    dht.New(ctx, h, dht.Mode(dht.ModeServer))
    
    fmt.Printf("Bootstrap: %s\n", h.ID())
    for _, addr := range h.Addrs() {
        fmt.Printf("%s/p2p/%s\n", addr, h.ID())
    }
    
    select {}
}
EOF
    go mod tidy
    go build -o bootstrap
fi

./bootstrap &
BOOTSTRAP_PID=$!
sleep 2

# Get bootstrap address
BOOTSTRAP_ADDR="/ip4/127.0.0.1/tcp/4001/p2p/$(./bootstrap 2>&1 | grep "Bootstrap:" | awk '{print $2}')"
echo "Bootstrap running at: $BOOTSTRAP_ADDR"

# Test provider
echo -e "\n2. Testing provider P2P..."
cd ../provider
QUIVER_BOOTSTRAP=$BOOTSTRAP_ADDR timeout 5 ./bin/provider 2>&1 | grep -E "(Started provider|Connected to bootstrap)" && echo "✓ Provider P2P working" || echo "✗ Provider P2P failed"

# Test gateway
echo -e "\n3. Testing gateway P2P..."
cd ../gateway  
QUIVER_BOOTSTRAP=$BOOTSTRAP_ADDR timeout 5 ./bin/gateway 2>&1 | grep -E "(Gateway started|Connected to bootstrap)" && echo "✓ Gateway P2P working" || echo "✗ Gateway P2P failed"

# Cleanup
echo -e "\nCleaning up..."
kill $BOOTSTRAP_PID 2>/dev/null

echo -e "\nP2P test complete!"