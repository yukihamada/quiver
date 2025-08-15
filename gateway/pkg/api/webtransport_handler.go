package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
	"github.com/quiver/gateway/pkg/p2p"
	"github.com/sirupsen/logrus"
)

// WebTransportHandler handles direct P2P connections from browsers
type WebTransportHandler struct {
	p2pClient *p2p.Client
	server    *webtransport.Server
	sessions  map[string]*BrowserSession
	mu        sync.RWMutex
	logger    *logrus.Logger
}

// BrowserSession represents a connected browser
type BrowserSession struct {
	session  *webtransport.Session
	id       string
	lastSeen time.Time
}

// Message types for browser-gateway communication
type WSMessage struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

// Use types from models.go instead of redefining

// NewWebTransportHandler creates a new WebTransport handler
func NewWebTransportHandler(p2pClient *p2p.Client) *WebTransportHandler {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &WebTransportHandler{
		p2pClient: p2pClient,
		sessions:  make(map[string]*BrowserSession),
		logger:    logger,
	}
}

// SetupWebTransport configures WebTransport endpoints
func (h *Handler) SetupWebTransport(router *gin.Engine, certFile, keyFile string) error {
	// Create WebTransport server
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		// Use self-signed certificate for development
		cert, err = generateSelfSignedCert()
		if err != nil {
			return fmt.Errorf("failed to generate certificate: %w", err)
		}
	}

	wtHandler := NewWebTransportHandler(h.p2pClient)

	// HTTP/3 server for WebTransport
	server := &http3.Server{
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h3"},
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/webtransport" {
				wtHandler.HandleWebTransport(w, r)
				return
			}
			
			// Serve regular HTTP/3 requests
			router.ServeHTTP(w, r)
		}),
		EnableDatagrams: true,
	}

	// Start HTTP/3 server on separate port
	go func() {
		h.logger.Info("Starting WebTransport server on :8443")
		server.Addr = ":8443"
		if err := server.ListenAndServe(); err != nil {
			h.logger.WithError(err).Error("WebTransport server failed")
		}
	}()

	// Add WebTransport info endpoint
	router.GET("/webtransport/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"available": true,
			"url":       "https://localhost:8443/webtransport",
			"protocol":  "h3",
		})
	})

	return nil
}

// HandleWebTransport handles WebTransport upgrade requests
func (wth *WebTransportHandler) HandleWebTransport(w http.ResponseWriter, r *http.Request) {
	session, err := wth.server.Upgrade(w, r)
	if err != nil {
		wth.logger.WithError(err).Error("Failed to upgrade to WebTransport")
		http.Error(w, "WebTransport upgrade failed", http.StatusBadRequest)
		return
	}

	// Generate session ID
	sessionID := fmt.Sprintf("browser-%d", time.Now().UnixNano())
	
	wth.mu.Lock()
	wth.sessions[sessionID] = &BrowserSession{
		session:  session,
		id:       sessionID,
		lastSeen: time.Now(),
	}
	wth.mu.Unlock()

	wth.logger.WithField("session_id", sessionID).Info("New WebTransport session")

	// Handle session
	go wth.handleSession(sessionID, session)
}

func (wth *WebTransportHandler) handleSession(sessionID string, session *webtransport.Session) {
	defer func() {
		wth.mu.Lock()
		delete(wth.sessions, sessionID)
		wth.mu.Unlock()
		session.CloseWithError(0, "")
	}()

	ctx := session.Context()

	for {
		stream, err := session.AcceptStream(ctx)
		if err != nil {
			return
		}

		go wth.handleStream(sessionID, stream)
	}
}

func (wth *WebTransportHandler) handleStream(sessionID string, stream webtransport.Stream) {
	defer stream.Close()

	decoder := json.NewDecoder(stream)
	encoder := json.NewEncoder(stream)

	for {
		var msg WSMessage
		if err := decoder.Decode(&msg); err != nil {
			if err != io.EOF {
				wth.logger.WithError(err).Error("Failed to decode message")
			}
			return
		}

		switch msg.Type {
		case "generate":
			wth.handleGenerate(msg, encoder)
		case "ping":
			encoder.Encode(WSMessage{Type: "pong", ID: msg.ID})
		default:
			encoder.Encode(WSMessage{
				Type: "error",
				ID:   msg.ID,
				Payload: json.RawMessage(`{"error": "unknown message type"}`),
			})
		}
	}
}

func (wth *WebTransportHandler) handleGenerate(msg WSMessage, encoder *json.Encoder) {
	var req GenerateRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		wth.sendError(encoder, msg.ID, "Invalid request")
		return
	}

	// Find providers
	providers, err := wth.p2pClient.FindProviders()
	if err != nil || len(providers) == 0 {
		wth.sendError(encoder, msg.ID, "No providers available")
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
		resp, err := wth.p2pClient.CallProvider(ctx, provider.ID, streamReq)
		if err != nil {
			continue
		}

		// Send response
		response := GenerateResponse{
			Completion: resp.Completion,
			Receipt:    resp.Receipt,
		}

		payload, _ := json.Marshal(response)
		encoder.Encode(WSMessage{
			Type:    "generate_response",
			ID:      msg.ID,
			Payload: payload,
		})
		return
	}

	wth.sendError(encoder, msg.ID, "All providers failed")
}

func (wth *WebTransportHandler) sendError(encoder *json.Encoder, msgID, errorMsg string) {
	payload, _ := json.Marshal(ErrorResponse{Error: errorMsg})
	encoder.Encode(WSMessage{
		Type:    "error",
		ID:      msgID,
		Payload: payload,
	})
}

// generateSelfSignedCert generates a self-signed certificate for development
func generateSelfSignedCert() (tls.Certificate, error) {
	// This is a placeholder - in production, use proper certificates
	return tls.Certificate{}, fmt.Errorf("self-signed cert generation not implemented")
}