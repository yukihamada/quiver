package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

// WebSocketHandler handles WebSocket connections for browser peers
type WebSocketHandler struct {
	host       host.Host
	upgrader   websocket.Upgrader
	peers      map[string]*WSPeer
	mu         sync.RWMutex
	logger     *logrus.Logger
}

// WSPeer represents a WebSocket connected peer
type WSPeer struct {
	ID       string
	Conn     *websocket.Conn
	PeerInfo peer.AddrInfo
}

// Message types
type WSMessage struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(h host.Host) *WebSocketHandler {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &WebSocketHandler{
		host: h,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from browsers
				return true
			},
		},
		peers:  make(map[string]*WSPeer),
		logger: logger,
	}
}

// ServeHTTP handles WebSocket upgrade requests
func (wsh *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wsh.logger.WithError(err).Error("Failed to upgrade WebSocket")
		return
	}

	// Create peer ID for this WebSocket connection
	peerID := fmt.Sprintf("ws-peer-%d", time.Now().UnixNano())
	
	peer := &WSPeer{
		ID:   peerID,
		Conn: conn,
	}

	wsh.mu.Lock()
	wsh.peers[peerID] = peer
	wsh.mu.Unlock()

	wsh.logger.WithField("peer_id", peerID).Info("WebSocket peer connected")

	// Send bootstrap info
	wsh.sendBootstrapInfo(peer)

	// Handle messages
	go wsh.handlePeer(peer)
}

// handlePeer handles messages from a WebSocket peer
func (wsh *WebSocketHandler) handlePeer(peer *WSPeer) {
	defer func() {
		wsh.mu.Lock()
		delete(wsh.peers, peer.ID)
		wsh.mu.Unlock()
		peer.Conn.Close()
		wsh.logger.WithField("peer_id", peer.ID).Info("WebSocket peer disconnected")
	}()

	for {
		var msg WSMessage
		if err := peer.Conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				wsh.logger.WithError(err).Error("WebSocket read error")
			}
			break
		}

		wsh.handleMessage(peer, msg)
	}
}

// handleMessage processes WebSocket messages
func (wsh *WebSocketHandler) handleMessage(peer *WSPeer, msg WSMessage) {
	switch msg.Type {
	case "get_peers":
		wsh.sendPeerList(peer)
		
	case "find_providers":
		wsh.sendProviders(peer)
		
	case "connect_peer":
		var payload struct {
			PeerID string `json:"peer_id"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			wsh.sendError(peer, msg.ID, "Invalid payload")
			return
		}
		
		// Connect to peer via libp2p
		if err := wsh.connectToPeer(payload.PeerID); err != nil {
			wsh.sendError(peer, msg.ID, err.Error())
			return
		}
		
		wsh.sendSuccess(peer, msg.ID, "Connected")
	}
}

// sendBootstrapInfo sends bootstrap node information
func (wsh *WebSocketHandler) sendBootstrapInfo(peer *WSPeer) {
	info := map[string]interface{}{
		"peer_id":   wsh.host.ID().String(),
		"addresses": []string{},
		"protocols": wsh.host.Mux().Protocols(),
	}

	// Get multiaddresses
	for _, addr := range wsh.host.Addrs() {
		info["addresses"] = append(info["addresses"].([]string), addr.String())
	}

	msg := WSMessage{
		Type:    "bootstrap_info",
		Payload: mustMarshal(info),
	}

	peer.Conn.WriteJSON(msg)
}

// sendPeerList sends the list of known peers
func (wsh *WebSocketHandler) sendPeerList(peer *WSPeer) {
	peers := wsh.host.Network().Peers()
	peerList := make([]map[string]interface{}, 0, len(peers))

	for _, p := range peers {
		peerInfo := map[string]interface{}{
			"id":        p.String(),
			"protocols": wsh.host.Peerstore().GetProtocols(p),
		}
		
		// Get addresses
		addrs := wsh.host.Peerstore().Addrs(p)
		addrStrings := make([]string, 0, len(addrs))
		for _, addr := range addrs {
			addrStrings = append(addrStrings, addr.String())
		}
		peerInfo["addresses"] = addrStrings
		
		peerList = append(peerList, peerInfo)
	}

	msg := WSMessage{
		Type:    "peer_list",
		Payload: mustMarshal(map[string]interface{}{"peers": peerList}),
	}

	peer.Conn.WriteJSON(msg)
}

// sendProviders sends the list of AI providers
func (wsh *WebSocketHandler) sendProviders(peer *WSPeer) {
	// In a real implementation, this would query the DHT
	// For now, return connected peers that support the provider protocol
	providers := []map[string]interface{}{}
	
	for _, p := range wsh.host.Network().Peers() {
		protocols := wsh.host.Peerstore().GetProtocols(p)
		for _, proto := range protocols {
			if proto == "/quiver/provider/1.0.0" {
				providers = append(providers, map[string]interface{}{
					"id":       p.String(),
					"latency":  wsh.host.Peerstore().LatencyEWMA(p).Milliseconds(),
				})
				break
			}
		}
	}

	msg := WSMessage{
		Type:    "providers",
		Payload: mustMarshal(map[string]interface{}{"providers": providers}),
	}

	peer.Conn.WriteJSON(msg)
}

// connectToPeer connects to a peer by ID
func (wsh *WebSocketHandler) connectToPeer(peerID string) error {
	pid, err := peer.Decode(peerID)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	// Check if already connected
	if wsh.host.Network().Connectedness(pid) == network.Connected {
		return nil
	}

	// Get peer info from peerstore or DHT
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return wsh.host.Connect(ctx, peer.AddrInfo{ID: pid})
}

// sendError sends an error message
func (wsh *WebSocketHandler) sendError(peer *WSPeer, msgID, error string) {
	msg := WSMessage{
		Type: "error",
		ID:   msgID,
		Payload: mustMarshal(map[string]string{"error": error}),
	}
	peer.Conn.WriteJSON(msg)
}

// sendSuccess sends a success message
func (wsh *WebSocketHandler) sendSuccess(peer *WSPeer, msgID string, data interface{}) {
	msg := WSMessage{
		Type: "success",
		ID:   msgID,
		Payload: mustMarshal(data),
	}
	peer.Conn.WriteJSON(msg)
}

// mustMarshal marshals data to JSON
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("{}")
	}
	return data
}