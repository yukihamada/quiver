package webrtc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pion/webrtc/v4"
	"github.com/quiver/gateway/pkg/p2p"
	"github.com/sirupsen/logrus"
)

// WebRTCTransport enables browser-to-P2P connections via WebRTC
type WebRTCTransport struct {
	p2pClient    *p2p.Client
	peerConnections map[string]*webrtc.PeerConnection
	mu           sync.RWMutex
	logger       *logrus.Logger
	config       webrtc.Configuration
}

// NewWebRTCTransport creates a new WebRTC transport
func NewWebRTCTransport(p2pClient *p2p.Client) *WebRTCTransport {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Configure STUN/TURN servers
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{
					"stun:stun.l.google.com:19302",
					"stun:stun1.l.google.com:19302",
				},
			},
			// Add public TURN servers if needed
			{
				URLs: []string{
					"turn:openrelay.metered.ca:80",
					"turn:openrelay.metered.ca:443",
				},
				Username:   "openrelayproject",
				Credential: "openrelayproject",
			},
		},
	}

	return &WebRTCTransport{
		p2pClient:       p2pClient,
		peerConnections: make(map[string]*webrtc.PeerConnection),
		logger:          logger,
		config:          config,
	}
}

// CreateOffer creates a WebRTC offer for a browser peer
func (wt *WebRTCTransport) CreateOffer(peerID string) (*webrtc.SessionDescription, error) {
	// Create a new peer connection
	pc, err := webrtc.NewPeerConnection(wt.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	// Create a data channel for P2P communication
	dc, err := pc.CreateDataChannel("quiver", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create data channel: %w", err)
	}

	// Set up data channel handlers
	dc.OnOpen(func() {
		wt.logger.WithField("peer_id", peerID).Info("Data channel opened")
	})

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		wt.handleDataChannelMessage(peerID, msg)
	})

	// Create offer
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	// Set local description
	if err := pc.SetLocalDescription(offer); err != nil {
		return nil, fmt.Errorf("failed to set local description: %w", err)
	}

	// Store peer connection
	wt.mu.Lock()
	wt.peerConnections[peerID] = pc
	wt.mu.Unlock()

	return &offer, nil
}

// HandleAnswer handles the WebRTC answer from a browser peer
func (wt *WebRTCTransport) HandleAnswer(peerID string, answer webrtc.SessionDescription) error {
	wt.mu.RLock()
	pc, exists := wt.peerConnections[peerID]
	wt.mu.RUnlock()

	if !exists {
		return fmt.Errorf("peer connection not found for %s", peerID)
	}

	// Set remote description
	if err := pc.SetRemoteDescription(answer); err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	return nil
}

// HandleICECandidate handles ICE candidates from browser peers
func (wt *WebRTCTransport) HandleICECandidate(peerID string, candidate webrtc.ICECandidateInit) error {
	wt.mu.RLock()
	pc, exists := wt.peerConnections[peerID]
	wt.mu.RUnlock()

	if !exists {
		return fmt.Errorf("peer connection not found for %s", peerID)
	}

	return pc.AddICECandidate(candidate)
}

// handleDataChannelMessage processes messages from browser peers
func (wt *WebRTCTransport) handleDataChannelMessage(peerID string, msg webrtc.DataChannelMessage) {
	var request struct {
		Type    string          `json:"type"`
		ID      string          `json:"id"`
		Payload json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		wt.logger.WithError(err).Error("Failed to unmarshal message")
		return
	}

	switch request.Type {
	case "generate":
		wt.handleGenerateRequest(peerID, request.ID, request.Payload)
	case "ping":
		wt.sendResponse(peerID, request.ID, "pong", nil)
	}
}

// handleGenerateRequest processes AI generation requests from browser
func (wt *WebRTCTransport) handleGenerateRequest(peerID, requestID string, payload json.RawMessage) {
	var req struct {
		Prompt string `json:"prompt"`
		Model  string `json:"model"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		wt.sendError(peerID, requestID, "Invalid request")
		return
	}

	// Find providers
	providers, err := wt.p2pClient.FindProviders()
	if err != nil || len(providers) == 0 {
		wt.sendError(peerID, requestID, "No providers available")
		return
	}

	// Call provider through P2P network
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	streamReq := &p2p.StreamRequest{
		Prompt:    req.Prompt,
		Model:     req.Model,
		MaxTokens: 256,
	}

	// Try each provider
	for _, provider := range providers {
		resp, err := wt.p2pClient.CallProvider(ctx, provider.ID, streamReq)
		if err != nil {
			continue
		}

		// Send response back to browser
		response := map[string]interface{}{
			"completion": resp.Completion,
			"receipt":    resp.Receipt,
		}

		wt.sendResponse(peerID, requestID, "generate_response", response)
		return
	}

	wt.sendError(peerID, requestID, "All providers failed")
}

// sendResponse sends a response to a browser peer
func (wt *WebRTCTransport) sendResponse(peerID, requestID, msgType string, payload interface{}) {
	wt.mu.RLock()
	pc, exists := wt.peerConnections[peerID]
	wt.mu.RUnlock()

	if !exists {
		return
	}

	// Find data channel
	var dc *webrtc.DataChannel
	// This is simplified - in production, store data channel reference

	if dc == nil || dc.ReadyState() != webrtc.DataChannelStateOpen {
		return
	}

	response := map[string]interface{}{
		"type": msgType,
		"id":   requestID,
	}

	if payload != nil {
		response["payload"] = payload
	}

	data, _ := json.Marshal(response)
	dc.Send(data)
}

// sendError sends an error response to a browser peer
func (wt *WebRTCTransport) sendError(peerID, requestID, errorMsg string) {
	wt.sendResponse(peerID, requestID, "error", map[string]string{
		"error": errorMsg,
	})
}

// GetICECandidates returns the ICE candidates for a peer
func (wt *WebRTCTransport) GetICECandidates(peerID string) []webrtc.ICECandidate {
	wt.mu.RLock()
	pc, exists := wt.peerConnections[peerID]
	wt.mu.RUnlock()

	if !exists {
		return nil
	}

	var candidates []webrtc.ICECandidate
	// Collect ICE candidates
	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			candidates = append(candidates, *candidate)
		}
	})

	return candidates
}

// Close closes all peer connections
func (wt *WebRTCTransport) Close() error {
	wt.mu.Lock()
	defer wt.mu.Unlock()

	for peerID, pc := range wt.peerConnections {
		if err := pc.Close(); err != nil {
			wt.logger.WithError(err).WithField("peer_id", peerID).Error("Failed to close peer connection")
		}
	}

	wt.peerConnections = make(map[string]*webrtc.PeerConnection)
	return nil
}