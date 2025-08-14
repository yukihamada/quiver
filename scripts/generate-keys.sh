#!/bin/bash

# Private Key Generation Script for QUIVer

echo "================================"
echo "QUIVer Private Key Generator"
echo "================================"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "\n${YELLOW}Choose key generation method:${NC}"
echo "1. Generate Ethereum wallet (for smart contracts)"
echo "2. Generate P2P node key (for libp2p)"
echo "3. Generate both"
echo -n "Select (1-3): "
read choice

case $choice in
    1|3)
        echo -e "\n${GREEN}Generating Ethereum Wallet...${NC}"
        
        # Method 1: Using Node.js
        if command -v node &> /dev/null; then
            cat > generate-eth-key.js << 'EOF'
const crypto = require('crypto');

// Generate private key
const privateKey = crypto.randomBytes(32);
const privateKeyHex = privateKey.toString('hex');

// Generate public key (simplified - for demo)
console.log('\n🔑 Ethereum Private Key (KEEP SECRET!):');
console.log('0x' + privateKeyHex);
console.log('\n⚠️  IMPORTANT: Save this key securely and NEVER share it!');
console.log('\n📝 Add to contracts/.env:');
console.log(`PRIVATE_KEY=${privateKeyHex}`);
EOF
            node generate-eth-key.js
            rm generate-eth-key.js
        else
            echo "Node.js not found. Install Node.js to generate Ethereum keys."
        fi
        
        if [ "$choice" != "3" ]; then
            exit 0
        fi
        ;;
esac

case $choice in
    2|3)
        echo -e "\n${GREEN}Generating P2P Node Key...${NC}"
        
        # Create key generator if not exists
        if [ ! -f keygen/main.go ]; then
            mkdir -p keygen
            cat > keygen/main.go << 'EOF'
package main

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "os"
    
    "github.com/libp2p/go-libp2p/core/crypto"
    "github.com/libp2p/go-libp2p/core/peer"
)

func main() {
    // Generate Ed25519 key pair
    priv, pub, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, rand.Reader)
    if err != nil {
        panic(err)
    }
    
    // Get peer ID
    id, err := peer.IDFromPublicKey(pub)
    if err != nil {
        panic(err)
    }
    
    // Marshal private key
    privBytes, err := crypto.MarshalPrivateKey(priv)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("\n🔑 P2P Node Key Generated:")
    fmt.Printf("Peer ID: %s\n", id.String())
    fmt.Printf("Private Key (base64): %s\n", base64.StdEncoding.EncodeToString(privBytes))
    
    // Save to file
    keyFile := "node.key"
    err = os.WriteFile(keyFile, privBytes, 0600)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("\n✅ Key saved to: %s\n", keyFile)
    fmt.Println("\n📝 To use this key:")
    fmt.Printf("export QUIVER_PRIVATE_KEY=%s\n", keyFile)
    fmt.Println("or")
    fmt.Printf("./bin/provider --key-file %s\n", keyFile)
}
EOF
            cd keygen
            go mod init keygen
            go get github.com/libp2p/go-libp2p/core/crypto
            go get github.com/libp2p/go-libp2p/core/peer
            go build
            cd ..
        fi
        
        cd keygen
        ./keygen
        mv node.key ../
        cd ..
        
        echo -e "\n${GREEN}P2P key generated and saved to: node.key${NC}"
        ;;
        
    *)
        echo "Invalid choice"
        exit 1
        ;;
esac

echo -e "\n${GREEN}Key generation complete!${NC}"