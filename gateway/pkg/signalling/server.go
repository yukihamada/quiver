package signalling

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// SignallingServer handles WebRTC signalling for browser peers
type SignallingServer struct {
	upgrader    websocket.Upgrader
	browsers    map[string]*BrowserPeer
	providers   map[string]*ProviderPeer
	mu          sync.RWMutex
	logger      *logrus.Logger
}

// BrowserPeer represents a connected browser
type BrowserPeer struct {
	ID         string
	Conn       *websocket.Conn
	UserAgent  string
	ConnectedAt time.Time
}

// ProviderPeer represents a P2P provider node
type ProviderPeer struct {
	ID          string
	PeerID      string
	Available   bool
	ConnectedAt time.Time
}

// Message types
type SignalMessage struct {
	Type    string          `json:"type"`
	From    string          `json:"from,omitempty"`
	To      string          `json:"to,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// NewSignallingServer creates a new signalling server
func NewSignallingServer() *SignallingServer {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &SignallingServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from GitHub Pages and localhost
				origin := r.Header.Get("Origin")
				return origin == "https://yukihamada.github.io" ||
					origin == "http://localhost" ||
					origin == "https://localhost" ||
					origin == ""
			},
		},
		browsers:  make(map[string]*BrowserPeer),
		providers: make(map[string]*ProviderPeer),
		logger:    logger,
	}
}

// ServeHTTP handles WebSocket upgrade requests
func (s *SignallingServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Error("Failed to upgrade connection")
		return
	}

	// Generate peer ID
	peerID := fmt.Sprintf("peer-%d", time.Now().UnixNano())
	
	// Create browser peer
	peer := &BrowserPeer{
		ID:          peerID,
		Conn:        conn,
		UserAgent:   r.Header.Get("User-Agent"),
		ConnectedAt: time.Now(),
	}

	s.mu.Lock()
	s.browsers[peerID] = peer
	s.mu.Unlock()

	s.logger.WithField("peer_id", peerID).Info("Browser connected")

	// Send welcome message with available providers
	s.sendAvailableProviders(peer)

	// Handle messages
	go s.handleBrowserPeer(peer)
}

// handleBrowserPeer handles messages from a browser peer
func (s *SignallingServer) handleBrowserPeer(peer *BrowserPeer) {
	defer func() {
		s.mu.Lock()
		delete(s.browsers, peer.ID)
		s.mu.Unlock()
		peer.Conn.Close()
		s.logger.WithField("peer_id", peer.ID).Info("Browser disconnected")
	}()

	for {
		var msg SignalMessage
		if err := peer.Conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.WithError(err).Error("WebSocket error")
			}
			break
		}

		msg.From = peer.ID
		s.handleMessage(peer, msg)
	}
}

// handleMessage processes signalling messages
func (s *SignallingServer) handleMessage(peer *BrowserPeer, msg SignalMessage) {
	s.logger.WithFields(logrus.Fields{
		"type": msg.Type,
		"from": msg.From,
		"to":   msg.To,
	}).Debug("Handling message")

	switch msg.Type {
	case "browser_join":
		// Already handled in connection
		
	case "offer":
		// Find available provider and forward offer
		provider := s.findAvailableProvider()
		if provider == nil {
			s.sendError(peer, "No providers available")
			return
		}
		
		// In a real implementation, forward to provider via P2P network
		// For now, simulate provider response
		s.simulateProviderResponse(peer, msg)
		
	case "ice_candidate":
		// Forward ICE candidate to provider
		// In real implementation, this would go through P2P network
		
	case "get_providers":
		s.sendAvailableProviders(peer)
	}
}

// findAvailableProvider finds an available provider
func (s *SignallingServer) findAvailableProvider() *ProviderPeer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, provider := range s.providers {
		if provider.Available {
			return provider
		}
	}
	
	// If no real providers, return a simulated one
	return &ProviderPeer{
		ID:        "simulated-provider",
		PeerID:    "12D3KooWSimulated",
		Available: true,
	}
}

// sendAvailableProviders sends list of available providers to browser
func (s *SignallingServer) sendAvailableProviders(peer *BrowserPeer) {
	s.mu.RLock()
	providers := make([]string, 0, len(s.providers))
	for _, p := range s.providers {
		if p.Available {
			providers = append(providers, p.PeerID)
		}
	}
	s.mu.RUnlock()

	// Always include at least one provider for testing
	if len(providers) == 0 {
		providers = append(providers, "12D3KooWSimulated")
	}

	msg := SignalMessage{
		Type: "providers_list",
		Payload: mustMarshal(map[string]interface{}{
			"providers": providers,
			"count":     len(providers),
		}),
	}

	peer.Conn.WriteJSON(msg)
}

// simulateProviderResponse simulates a provider's WebRTC answer
func (s *SignallingServer) simulateProviderResponse(peer *BrowserPeer, offer SignalMessage) {
	// In a real implementation, this would come from the actual provider
	// For now, send a simulated answer
	time.Sleep(100 * time.Millisecond) // Simulate network delay

	answer := SignalMessage{
		Type: "answer",
		From: "provider",
		To:   peer.ID,
		Payload: mustMarshal(map[string]interface{}{
			"type": "answer",
			"sdp":  "simulated-answer-sdp",
		}),
	}

	peer.Conn.WriteJSON(answer)
}

// sendError sends an error message to a browser peer
func (s *SignallingServer) sendError(peer *BrowserPeer, errorMsg string) {
	msg := SignalMessage{
		Type: "error",
		Payload: mustMarshal(map[string]string{
			"error": errorMsg,
		}),
	}
	peer.Conn.WriteJSON(msg)
}

// RegisterProvider registers a P2P provider node
func (s *SignallingServer) RegisterProvider(peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.providers[peerID] = &ProviderPeer{
		ID:          fmt.Sprintf("provider-%d", time.Now().UnixNano()),
		PeerID:      peerID,
		Available:   true,
		ConnectedAt: time.Now(),
	}

	s.logger.WithField("peer_id", peerID).Info("Provider registered")
}

// mustMarshal marshals data to JSON or returns empty object on error
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("{}")
	}
	return data
}