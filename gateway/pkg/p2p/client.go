package p2p

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

const protocolID = "/jan-nano/1.0.0"
const dhtTopic = "ai.providers.jan-nano/1.0.0"

type Client struct {
	host host.Host
	dht  *dht.IpfsDHT
	ctx  context.Context
}

type StreamRequest struct {
	Prompt    string `json:"prompt"`
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
}

type StreamResponse struct {
	Completion string      `json:"completion"`
	Receipt    interface{} `json:"receipt"`
	Error      string      `json:"error,omitempty"`
}

func NewClient(ctx context.Context, listenAddr string, bootstrapPeers []string) (*Client, error) {
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, rand.Reader)
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
		// Use default transports (TCP + QUIC) for better compatibility
		libp2p.DefaultTransports,
		libp2p.DefaultSecurity,
		// NAT traversal will be configured separately
	)
	if err != nil {
		return nil, err
	}

	kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeClient))
	if err != nil {
		return nil, err
	}

	if err = kadDHT.Bootstrap(ctx); err != nil {
		return nil, err
	}

	for _, peerAddr := range bootstrapPeers {
		addr, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			continue
		}

		peerinfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			continue
		}

		if err := h.Connect(ctx, *peerinfo); err == nil {
			fmt.Printf("Connected to bootstrap peer: %s\n", peerinfo.ID)
		}
	}

	return &Client{
		host: h,
		dht:  kadDHT,
		ctx:  ctx,
	}, nil
}

func (c *Client) FindProviders() ([]peer.AddrInfo, error) {
	peers := c.dht.RoutingTable().ListPeers()

	var providers []peer.AddrInfo
	for _, p := range peers {
		if p != c.host.ID() {
			addrs := c.host.Peerstore().Addrs(p)
			if len(addrs) > 0 {
				providers = append(providers, peer.AddrInfo{
					ID:    p,
					Addrs: addrs,
				})
			}
		}
	}

	return providers, nil
}

func (c *Client) CallProvider(ctx context.Context, providerID peer.ID, req *StreamRequest) (*StreamResponse, error) {
	stream, err := c.host.NewStream(ctx, providerID, protocol.ID(protocolID))
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	encoder := json.NewEncoder(stream)
	if err := encoder.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var resp StreamResponse
	decoder := json.NewDecoder(stream)
	if err := decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("provider error: %s", resp.Error)
	}

	return &resp, nil
}

func (c *Client) Close() error {
	return c.host.Close()
}
