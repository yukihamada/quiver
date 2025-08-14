package p2p

import (
    "context"
    "fmt"
    "time"

    dht "github.com/libp2p/go-libp2p-kad-dht"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/core/protocol"
    "github.com/multiformats/go-multiaddr"
)

const (
    // DHTプロトコルID
    DHTProtocol = "/quiver/dht/1.0.0"
    // ノード発見間隔
    DiscoveryInterval = 30 * time.Second
    // 最小ピア数
    MinPeers = 3
)

// DHTManager はDHTベースのピア発見を管理
type DHTManager struct {
    host    host.Host
    dht     *dht.IpfsDHT
    ctx     context.Context
    cancel  context.CancelFunc
}

// NewDHTManager creates a new DHT manager
func NewDHTManager(h host.Host) (*DHTManager, error) {
    ctx, cancel := context.WithCancel(context.Background())
    
    // DHTを作成（サーバーモードで起動）
    kademliaDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
    if err != nil {
        cancel()
        return nil, fmt.Errorf("failed to create DHT: %w", err)
    }

    // DHTをブートストラップ
    if err := kademliaDHT.Bootstrap(ctx); err != nil {
        cancel()
        return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
    }

    dm := &DHTManager{
        host:   h,
        dht:    kademliaDHT,
        ctx:    ctx,
        cancel: cancel,
    }

    // 自動ピア発見を開始
    go dm.discoverPeers()

    return dm, nil
}

// ConnectToBootstrapPeers はブートストラップピアに接続
func (dm *DHTManager) ConnectToBootstrapPeers(bootstrapPeers []string) error {
    for _, addr := range bootstrapPeers {
        maddr, err := multiaddr.NewMultiaddr(addr)
        if err != nil {
            continue
        }

        peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
        if err != nil {
            continue
        }

        if err := dm.host.Connect(dm.ctx, *peerInfo); err == nil {
            fmt.Printf("Connected to bootstrap peer: %s\n", peerInfo.ID)
        }
    }

    return nil
}

// discoverPeers は定期的に新しいピアを発見
func (dm *DHTManager) discoverPeers() {
    ticker := time.NewTicker(DiscoveryInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            dm.findAndConnectPeers()
        case <-dm.ctx.Done():
            return
        }
    }
}

// findAndConnectPeers は新しいピアを見つけて接続
func (dm *DHTManager) findAndConnectPeers() {
    // 現在の接続数を確認
    peers := dm.host.Network().Peers()
    if len(peers) >= MinPeers {
        return
    }

    // ランダムなキーでピアを検索
    randomKey := make([]byte, 32)
    for i := range randomKey {
        randomKey[i] = byte(i)
    }

    closestPeers, err := dm.dht.GetClosestPeers(dm.ctx, string(randomKey))
    if err != nil {
        return
    }

    for p := range closestPeers {
        if p == dm.host.ID() {
            continue
        }

        // 既に接続している場合はスキップ
        if dm.host.Network().Connectedness(p) == network.Connected {
            continue
        }

        // ピア情報を取得
        peerInfo := dm.dht.FindLocal(p)
        if peerInfo.ID == "" {
            continue
        }

        // 接続を試みる
        if err := dm.host.Connect(dm.ctx, peerInfo); err == nil {
            fmt.Printf("Connected to new peer: %s\n", p)
        }
    }
}

// Advertise はサービスをアドバタイズ
func (dm *DHTManager) Advertise(service string) error {
    routingDiscovery := dht.NewRoutingDiscovery(dm.dht)
    _, err := routingDiscovery.Advertise(dm.ctx, service)
    return err
}

// FindProviders はサービスプロバイダーを検索
func (dm *DHTManager) FindProviders(service string) ([]peer.AddrInfo, error) {
    routingDiscovery := dht.NewRoutingDiscovery(dm.dht)
    
    peerChan, err := routingDiscovery.FindPeers(dm.ctx, service)
    if err != nil {
        return nil, err
    }

    var providers []peer.AddrInfo
    timeout := time.After(10 * time.Second)
    
    for {
        select {
        case peer, ok := <-peerChan:
            if !ok {
                return providers, nil
            }
            if peer.ID != dm.host.ID() {
                providers = append(providers, peer)
            }
        case <-timeout:
            return providers, nil
        }
    }
}

// GetDHT returns the underlying DHT
func (dm *DHTManager) GetDHT() *dht.IpfsDHT {
    return dm.dht
}

// Close shuts down the DHT manager
func (dm *DHTManager) Close() error {
    dm.cancel()
    return dm.dht.Close()
}