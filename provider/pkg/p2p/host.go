package p2p

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

type Host struct {
	host host.Host
	dht  *dht.IpfsDHT
	ctx  context.Context
}

// GetHost returns the underlying libp2p host
func (h *Host) GetHost() host.Host {
	return h.host
}

func NewHost(ctx context.Context, listenAddr string, bootstrapPeers []string) (*Host, error) {
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
		// TODO: Add connection limits after resolving version compatibility
	)
	if err != nil {
		return nil, err
	}

	kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
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

	return &Host{
		host: h,
		dht:  kadDHT,
		ctx:  ctx,
	}, nil
}

func (h *Host) SetStreamHandler(proto protocol.ID, handler func(network.Stream)) {
	h.host.SetStreamHandler(proto, handler)
}

func (h *Host) Advertise(topic string) error {
	return h.dht.Bootstrap(h.ctx)
}

func (h *Host) ID() peer.ID {
	return h.host.ID()
}

func (h *Host) Addrs() []multiaddr.Multiaddr {
	return h.host.Addrs()
}

func (h *Host) Close() error {
	return h.host.Close()
}
