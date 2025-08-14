package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// RelayConfig holds configuration for Circuit Relay v2
type RelayConfig struct {
	// Enable relay service (for public nodes)
	EnableRelayService bool
	// Enable using relays (for nodes behind NAT)
	EnableAutoRelay bool
	// Static relay addresses
	StaticRelays []string
	// Enable hole punching
	EnableHolePunching bool
}

// DefaultRelayConfig returns a default relay configuration
func DefaultRelayConfig() *RelayConfig {
	return &RelayConfig{
		EnableRelayService: false,
		EnableAutoRelay:    true,
		EnableHolePunching: true,
		StaticRelays: []string{
			// Public relay nodes (to be deployed)
			"/ip4/relay1.quiver.network/tcp/4001/p2p/QmRelayNode1",
			"/ip4/relay2.quiver.network/tcp/4001/p2p/QmRelayNode2",
			"/ip4/relay3.quiver.network/tcp/4001/p2p/QmRelayNode3",
		},
	}
}

// ApplyRelayOptions applies relay configuration to libp2p options
func (rc *RelayConfig) ApplyRelayOptions(opts []libp2p.Option) []libp2p.Option {
	if rc.EnableRelayService {
		opts = append(opts, libp2p.EnableRelayService())
	}

	if rc.EnableAutoRelay && len(rc.StaticRelays) > 0 {
		relayInfos := parseRelayAddrs(rc.StaticRelays)
		if len(relayInfos) > 0 {
			opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(relayInfos))
		}
	} else if rc.EnableAutoRelay {
		opts = append(opts, libp2p.EnableAutoRelay())
	}

	if rc.EnableHolePunching {
		opts = append(opts, libp2p.EnableHolePunching())
	}

	return opts
}

// parseRelayAddrs converts multiaddr strings to peer.AddrInfo
func parseRelayAddrs(addrs []string) []peer.AddrInfo {
	var relayInfos []peer.AddrInfo
	
	for _, addr := range addrs {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			fmt.Printf("Failed to parse relay address %s: %v\n", addr, err)
			continue
		}
		
		ai, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			fmt.Printf("Failed to get peer info from %s: %v\n", addr, err)
			continue
		}
		
		relayInfos = append(relayInfos, *ai)
	}
	
	return relayInfos
}

// SetupRelayNode creates a dedicated relay node
func SetupRelayNode(ctx context.Context, listenAddr string, privKeyPath string) (*Host, error) {
	priv, _, err := LoadOrGenerateKey(privKeyPath)
	if err != nil {
		return nil, err
	}

	listen, err := multiaddr.NewMultiaddr(listenAddr)
	if err != nil {
		return nil, err
	}

	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrs(listen),
		libp2p.EnableRelayService(),
		libp2p.ForceReachabilityPublic(),
		libp2p.DefaultSecurity,
			// Resource limits would be configured here in newer versions
	)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Relay node started with ID: %s\n", h.ID())
	fmt.Printf("Listening on: %s\n", h.Addrs())
	
	return &Host{
		host: h,
		ctx:  ctx,
	}, nil
}