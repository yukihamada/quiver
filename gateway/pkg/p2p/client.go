package p2p

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	libp2pquic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
)

const protocolID = "/quiver/inference/1.0.0"
const dhtTopic = "quiver.providers"

type Client struct {
	host host.Host
	dht  *dht.IpfsDHT
	ctx  context.Context
}

type StreamRequest struct {
	Prompt    string `json:"prompt"`
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	Stream    bool   `json:"stream,omitempty"`
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

	// Extract port from multiaddr
	portStr := "4002"
	parts := strings.Split(listenAddr, "/")
	for i, part := range parts {
		if part == "tcp" && i+1 < len(parts) {
			portStr = parts[i+1]
			break
		}
	}
	
	// Create additional QUIC listen address
	quicAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/udp/%s/quic-v1", portStr))
	
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrs(listen, quicAddr),
		
		// Enable both TCP and QUIC transports
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(libp2pquic.NewTransport),
		
		// Enable NAT traversal
		libp2p.EnableNATService(),
		
		// Enable hole punching for NAT traversal
		libp2p.EnableHolePunching(),
		
		// Enable relay for nodes behind NAT
		libp2p.EnableRelay(),
		
		// Default security
		libp2p.DefaultSecurity,
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
	// First try DHT routing table
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
	
	// If no providers found, try to find via DHT topic
	if len(providers) == 0 {
		// Try finding peers that support our protocol
		ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
		defer cancel()
		
		// Get all connected peers
		for _, p := range c.host.Network().Peers() {
			if p != c.host.ID() {
				// Check if peer supports our protocol
				protos, err := c.host.Peerstore().GetProtocols(p)
				if err == nil {
					for _, proto := range protos {
						if proto == protocolID {
							addrs := c.host.Peerstore().Addrs(p)
							if len(addrs) > 0 {
								providers = append(providers, peer.AddrInfo{
									ID:    p,
									Addrs: addrs,
								})
								break
							}
						}
					}
				}
			}
		}
		
		// As last resort, check if any peer responds to our protocol
		if len(providers) == 0 {
			for _, p := range c.host.Network().Peers() {
				if p != c.host.ID() {
					stream, err := c.host.NewStream(ctx, p, protocol.ID(protocolID))
					if err == nil {
						stream.Close()
						addrs := c.host.Peerstore().Addrs(p)
						if len(addrs) > 0 {
							providers = append(providers, peer.AddrInfo{
								ID:    p,
								Addrs: addrs,
							})
						}
					}
				}
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

func (c *Client) CreateStream(ctx context.Context, providerID peer.ID, req *StreamRequest) (network.Stream, error) {
	stream, err := c.host.NewStream(ctx, providerID, protocol.ID(protocolID))
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}
	
	encoder := json.NewEncoder(stream)
	if err := encoder.Encode(req); err != nil {
		stream.Close()
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	
	return stream, nil
}

// IsConnected checks if the client is connected to the P2P network
func (c *Client) IsConnected() bool {
	return len(c.host.Network().Peers()) > 0
}

// PeerCount returns the number of connected peers
func (c *Client) PeerCount() int {
	return len(c.host.Network().Peers())
}

// GetProviders returns a list of available inference providers
func (c *Client) GetProviders(ctx context.Context) []peer.ID {
	providers, _ := c.FindProviders()
	peerIDs := make([]peer.ID, len(providers))
	for i, p := range providers {
		peerIDs[i] = p.ID
	}
	return peerIDs
}

// SendRequest sends an inference request to a provider
func (c *Client) SendRequest(ctx context.Context, providerID peer.ID, data []byte) (network.Stream, error) {
	var req StreamRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}
	return c.CreateStream(ctx, providerID, &req)
}

func (c *Client) Close() error {
	return c.host.Close()
}
