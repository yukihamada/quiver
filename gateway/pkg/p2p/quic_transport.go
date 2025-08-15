package p2p

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	libp2pquic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
	"github.com/quic-go/quic-go/http3"
)

// QUICTransport enables QUIC transport with NAT traversal
type QUICTransport struct {
	host       host.Host
	relayNodes []peer.AddrInfo
}

// NewQUICTransport creates a new P2P host with QUIC and NAT traversal
func NewQUICTransport(ctx context.Context, listenPort int) (*QUICTransport, error) {
	// Generate identity
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Create multiaddresses for listening
	tcpAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort))
	if err != nil {
		return nil, err
	}

	quicAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", listenPort+1))
	if err != nil {
		return nil, err
	}

	// Configure libp2p host with QUIC and Circuit Relay
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrs(tcpAddr, quicAddr),
		
		// Enable QUIC transport
		libp2p.Transport(libp2pquic.NewTransport),
		libp2p.Transport(tcp.NewTCPTransport),
		
		// Enable NAT traversal
		libp2p.EnableNATService(),
		
		// Enable relay
		libp2p.EnableRelay(),
		
		// Enable hole punching
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	return &QUICTransport{
		host: h,
	}, nil
}

// ConnectToPeer establishes a connection using QUIC with fallback to relay
func (qt *QUICTransport) ConnectToPeer(ctx context.Context, peerInfo peer.AddrInfo) error {
	// Add peer addresses to peerstore
	qt.host.Peerstore().AddAddrs(peerInfo.ID, peerInfo.Addrs, peerstore.PermanentAddrTTL)

	// Try direct connection first
	if err := qt.host.Connect(ctx, peerInfo); err == nil {
		return nil
	}

	// If direct connection fails, try through relay
	for _, relay := range qt.relayNodes {
		relayAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s/p2p-circuit/p2p/%s", relay.ID, peerInfo.ID))
		if err != nil {
			continue
		}

		relayPeerInfo := peer.AddrInfo{
			ID:    peerInfo.ID,
			Addrs: []multiaddr.Multiaddr{relayAddr},
		}

		if err := qt.host.Connect(ctx, relayPeerInfo); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to connect to peer %s", peerInfo.ID)
}

// StartRelayNode starts this node as a relay
func (qt *QUICTransport) StartRelayNode() error {
	_, err := relay.New(qt.host)
	return err
}

// OpenQUICStream opens a new QUIC stream to a peer
func (qt *QUICTransport) OpenQUICStream(ctx context.Context, peerID peer.ID, protocol string) (network.Stream, error) {
	return qt.host.NewStream(ctx, peerID, "/quiver/quic/1.0.0")
}

// WebTransportBridge bridges WebTransport from browsers to P2P network
type WebTransportBridge struct {
	p2pTransport *QUICTransport
	httpServer   *http.Server
	tlsConfig    *tls.Config
}

// NewWebTransportBridge creates a bridge for browser connections
func NewWebTransportBridge(p2pTransport *QUICTransport, certFile, keyFile string) (*WebTransportBridge, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h3"},
	}

	return &WebTransportBridge{
		p2pTransport: p2pTransport,
		tlsConfig:    tlsConfig,
	}, nil
}

// Start starts the WebTransport server
func (wb *WebTransportBridge) Start(addr string) error {
	mux := http.NewServeMux()
	
	// WebTransport endpoint
	mux.HandleFunc("/webtransport", func(w http.ResponseWriter, r *http.Request) {
		// Handle WebTransport upgrade
		wb.handleWebTransport(w, r)
	})

	// CORS headers for browser access
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("QUIVer WebTransport Bridge"))
	})

	// Create HTTP/3 server
	server := &http3.Server{
		TLSConfig: wb.tlsConfig,
		Handler:   mux,
		Addr:      addr,
	}

	return server.ListenAndServe()
}

func (wb *WebTransportBridge) handleWebTransport(w http.ResponseWriter, r *http.Request) {
	// This would handle the WebTransport session
	// For now, we'll use the existing WebTransport implementation
	w.WriteHeader(http.StatusNotImplemented)
}

