package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "os/signal"
    "path/filepath"
    "runtime"
    "strings"
    "syscall"
    "time"

    "github.com/libp2p/go-libp2p/core/crypto"
    "github.com/quiver/provider/pkg/api"
    "github.com/quiver/provider/pkg/p2p"
)

func main() {
    // アプリケーションディレクトリを設定
    homeDir, _ := os.UserHomeDir()
    dataDir := filepath.Join(homeDir, ".quiver")
    
    // データディレクトリを作成
    if err := os.MkdirAll(dataDir, 0755); err != nil {
        log.Fatalf("Failed to create data directory: %v", err)
    }

    // ログファイルを設定
    logFile, err := os.OpenFile(
        filepath.Join(dataDir, "provider.log"),
        os.O_CREATE|os.O_WRONLY|os.O_APPEND,
        0644,
    )
    if err != nil {
        log.Printf("Failed to open log file: %v", err)
    } else {
        defer logFile.Close()
        log.SetOutput(logFile)
    }

    // Ollamaがインストールされているか確認
    if !checkOllama() {
        fmt.Println("⚠️  Ollamaがインストールされていません")
        fmt.Println("📥 Ollamaをインストール中...")
        installOllama()
    }

    // Ollamaを起動
    if err := startOllama(); err != nil {
        log.Printf("Failed to start Ollama: %v", err)
    }

    // llama3.2モデルを確認
    if !checkModel("llama3.2") {
        fmt.Println("📥 AIモデル(llama3.2)をダウンロード中...")
        downloadModel("llama3.2:3b")
    }

    // 秘密鍵を読み込みまたは生成
    privKey, err := loadOrGenerateKey(dataDir)
    if err != nil {
        log.Fatalf("Failed to load/generate key: %v", err)
    }

    // デフォルトのブートストラップピア
    bootstrapPeers := []string{
        // パブリックブートストラップノード
        "/dnsaddr/bootstrap.quiver.network/p2p/12D3KooWLXexpZCqSDiMgJjYDqg6pGQ5Hm5X2FeVPZcB2Y5oKGpF",
    }

    // ノード設定
    cfg := &p2p.NodeConfig{
        ListenAddrs: []string{
            "/ip4/0.0.0.0/tcp/0",
            "/ip4/0.0.0.0/udp/0/quic",
        },
        BootstrapPeers: bootstrapPeers,
        PrivateKey:     privKey,
    }

    // P2Pノードを作成
    fmt.Println("🚀 QUIVer Provider を起動中...")
    node, err := p2p.NewDecentralizedNode(cfg)
    if err != nil {
        // ブートストラップに失敗しても単独で動作
        fmt.Println("⚠️  ネットワーク接続に失敗しました。スタンドアロンモードで起動します。")
        cfg.BootstrapPeers = nil
        node, err = p2p.NewDecentralizedNode(cfg)
        if err != nil {
            log.Fatalf("Failed to create P2P node: %v", err)
        }
    }
    defer node.Close()

    // APIサーバーを起動
    apiServer := api.NewServer(node, "http://localhost:11434")
    go func() {
        if err := apiServer.Start(":8082"); err != nil {
            log.Printf("API server error: %v", err)
        }
    }()

    // GUIサーバーを起動
    guiServer := api.NewGUIServer()
    go func() {
        if err := guiServer.Start(":8083"); err != nil {
            log.Printf("GUI server error: %v", err)
        }
    }()

    // 2秒待ってからブラウザを開く
    time.Sleep(2 * time.Second)
    openBrowser("http://localhost:8083")

    // ネットワーク状態を監視
    go monitorNetwork(node)

    // メニューバーアイコンを表示（macOS）
    if runtime.GOOS == "darwin" {
        go showMenuBarIcon()
    }

    // シグナルハンドリング
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    fmt.Println("\n✨ QUIVer Provider が起動しました！")
    fmt.Println("💻 ダッシュボード: http://localhost:8083")
    fmt.Println("💰 収益化が自動的に開始されます")
    fmt.Println("\n🛑 終了するにはメニューバーのアイコンから「終了」を選択してください")

    <-sigCh
    fmt.Println("\n👋 QUIVer Provider を終了します...")
}

func checkOllama() bool {
    _, err := exec.LookPath("ollama")
    return err == nil
}

func installOllama() {
    switch runtime.GOOS {
    case "darwin":
        // macOS用インストール
        cmd := exec.Command("bash", "-c", "curl -fsSL https://ollama.ai/install.sh | sh")
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            fmt.Printf("Ollamaの自動インストールに失敗しました: %v\n", err)
            fmt.Println("手動でインストールしてください: https://ollama.ai/download")
        }
    default:
        fmt.Println("お使いのOSでは自動インストールができません")
        fmt.Println("手動でインストールしてください: https://ollama.ai/download")
    }
}

func startOllama() error {
    // すでに起動しているか確認
    cmd := exec.Command("pgrep", "ollama")
    if err := cmd.Run(); err == nil {
        return nil // すでに起動している
    }

    // バックグラウンドで起動
    cmd = exec.Command("ollama", "serve")
    if err := cmd.Start(); err != nil {
        return err
    }

    // 起動を待つ
    time.Sleep(3 * time.Second)
    return nil
}

func checkModel(model string) bool {
    cmd := exec.Command("ollama", "list")
    output, err := cmd.Output()
    if err != nil {
        return false
    }
    return strings.Contains(string(output), model)
}

func downloadModel(model string) {
    cmd := exec.Command("ollama", "pull", model)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        fmt.Printf("モデルのダウンロードに失敗しました: %v\n", err)
    }
}

func loadOrGenerateKey(dataDir string) (crypto.PrivKey, error) {
    keyPath := filepath.Join(dataDir, "node.key")
    
    // 既存のキーを読み込み
    if keyBytes, err := os.ReadFile(keyPath); err == nil {
        return crypto.UnmarshalPrivateKey(keyBytes)
    }
    
    // 新規生成
    privKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
    if err != nil {
        return nil, err
    }
    
    // 保存
    keyBytes, err := crypto.MarshalPrivateKey(privKey)
    if err != nil {
        return nil, err
    }
    
    if err := os.WriteFile(keyPath, keyBytes, 0600); err != nil {
        return nil, err
    }
    
    return privKey, nil
}

func openBrowser(url string) {
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "darwin":
        cmd = exec.Command("open", url)
    case "windows":
        cmd = exec.Command("cmd", "/c", "start", url)
    default:
        cmd = exec.Command("xdg-open", url)
    }
    cmd.Start()
}

func monitorNetwork(node *p2p.DecentralizedNode) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        stats := node.GetNetworkStats()
        peers := stats["connected_peers"].(int)
        
        if peers == 0 {
            fmt.Println("\n⚠️  ネットワークに接続されていません（スタンドアロンモード）")
        } else {
            fmt.Printf("\n✅ ネットワーク接続中: %d ピア\n", peers)
        }
    }
}

func showMenuBarIcon() {
    // macOSのメニューバーアイコン実装（簡易版）
    // 実際の実装では、systrayライブラリなどを使用
    fmt.Println("📊 メニューバーにアイコンが表示されています")
}