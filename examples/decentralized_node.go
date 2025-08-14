package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/libp2p/go-libp2p/core/crypto"
    "github.com/quiver/quiver/provider/pkg/p2p"
)

func main() {
    // フラグを定義
    var (
        port           = flag.Int("port", 0, "Listen port (0 for random)")
        bootstrapPeers = flag.String("bootstrap", "", "Comma-separated list of bootstrap peers")
        isBootstrap    = flag.Bool("bootstrap-node", false, "Start as bootstrap node")
        keyFile        = flag.String("key", "", "Private key file")
    )
    flag.Parse()

    // 秘密鍵を読み込みまたは生成
    var privKey crypto.PrivKey
    var err error
    if *keyFile != "" {
        // ファイルから読み込み
        keyBytes, err := os.ReadFile(*keyFile)
        if err != nil {
            log.Fatalf("Failed to read key file: %v", err)
        }
        privKey, err = crypto.UnmarshalPrivateKey(keyBytes)
        if err != nil {
            log.Fatalf("Failed to unmarshal private key: %v", err)
        }
    } else {
        // 新規生成
        privKey, _, err = crypto.GenerateKeyPair(crypto.Ed25519, -1)
        if err != nil {
            log.Fatalf("Failed to generate key pair: %v", err)
        }
    }

    // リッスンアドレスを設定
    listenAddrs := []string{
        fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *port),
        fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", *port),
    }

    // ブートストラップピアを解析
    var bootstrapList []string
    if *bootstrapPeers != "" {
        bootstrapList = splitAndTrim(*bootstrapPeers, ",")
    }

    // ノード設定
    cfg := &p2p.NodeConfig{
        ListenAddrs:    listenAddrs,
        BootstrapPeers: bootstrapList,
        PrivateKey:     privKey,
    }

    // ノードを作成
    node, err := p2p.NewDecentralizedNode(cfg)
    if err != nil {
        log.Fatalf("Failed to create node: %v", err)
    }
    defer node.Close()

    // ブートストラップノードとして起動
    if *isBootstrap {
        node.StartAsBootstrapNode()
    }

    // ネットワーク統計を定期的に表示
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()

        for range ticker.C {
            stats := node.GetNetworkStats()
            fmt.Println("\n=== Network Statistics ===")
            fmt.Printf("Total nodes: %v\n", stats["total_nodes"])
            fmt.Printf("Connected peers: %v\n", stats["connected_peers"])
            fmt.Printf("Total capacity: %.2f%%\n", stats["total_capacity"])
            fmt.Printf("Average reputation: %.2f\n", stats["avg_reputation"])
            fmt.Printf("Models available: %v\n", stats["models_available"])
            fmt.Printf("DHT routing table size: %v\n", stats["dht_routing_table_size"])
            fmt.Println("========================")
        }
    }()

    // 推論リクエストの例（ブートストラップノードでない場合）
    if !*isBootstrap {
        go func() {
            // ネットワークが安定するまで待つ
            time.Sleep(10 * time.Second)

            // 定期的に推論リクエストを送信
            ticker := time.NewTicker(60 * time.Second)
            defer ticker.Stop()

            for range ticker.C {
                fmt.Println("\nSending inference request...")
                result, err := node.SendInferenceRequest(
                    "llama3.2",
                    "What is the meaning of life?",
                    100,
                )
                if err != nil {
                    fmt.Printf("Inference failed: %v\n", err)
                } else {
                    fmt.Printf("Inference result: %s\n", result)
                }
            }
        }()
    }

    // シグナルハンドリング
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    fmt.Println("\nNode is running. Press Ctrl+C to stop.")
    
    select {
    case <-sigCh:
        fmt.Println("\nShutting down...")
    case <-ctx.Done():
    }
}

func splitAndTrim(s string, sep string) []string {
    if s == "" {
        return nil
    }
    
    parts := strings.Split(s, sep)
    result := make([]string, 0, len(parts))
    
    for _, part := range parts {
        trimmed := strings.TrimSpace(part)
        if trimmed != "" {
            result = append(result, trimmed)
        }
    }
    
    return result
}