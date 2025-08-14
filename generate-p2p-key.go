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
}