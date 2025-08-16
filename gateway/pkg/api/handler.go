package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/quiver/gateway/pkg/p2p"
	"github.com/quiver/gateway/pkg/ratelimit"
)

// Handler handles HTTP requests and forwards them to the P2P network
type Handler struct {
	p2pClient      *p2p.Client
	limiter        *ratelimit.Limiter
	canaryRate     float64
	statsCollector *StatsCollector
}

// NewHandler creates a new API handler
func NewHandler(p2pClient *p2p.Client, limiter *ratelimit.Limiter, canaryRate float64) *Handler {
	return &Handler{
		p2pClient:      p2pClient,
		limiter:        limiter,
		canaryRate:     canaryRate,
		statsCollector: NewStatsCollector(),
	}
}

// InferenceRequest represents an inference request
type InferenceRequest struct {
	Prompt    string `json:"prompt"`
	Model     string `json:"model"`
	Token     string `json:"token"`
	MaxTokens int    `json:"max_tokens,omitempty"`
	Stream    bool   `json:"stream,omitempty"`
}

// InferenceResponse represents an inference response
type InferenceResponse struct {
	Completion string  `json:"completion"`
	Model      string  `json:"model"`
	Receipt    Receipt `json:"receipt"`
}

// Receipt represents a signed receipt from a provider
type Receipt struct {
	Receipt   InnerReceipt `json:"receipt"`
	Signature string       `json:"signature"`
}

// InnerReceipt contains the actual receipt data
type InnerReceipt struct {
	ProviderPK string `json:"provider_pk"`
	Timestamp  int64  `json:"timestamp"`
}

// Generate handles inference generation requests
func (h *Handler) Generate(c *gin.Context) {
	var req InferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate request
	if req.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt is required"})
		return
	}

	if req.Model == "" {
		req.Model = "llama3.2:3b"
	}

	// Check rate limit
	if req.Token != "" && !h.limiter.Allow(req.Token) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		return
	}

	// Find an available provider
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	providers := h.p2pClient.GetProviders(ctx)
	if len(providers) == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "No providers available"})
		return
	}

	// Try each provider until one succeeds
	for _, provider := range providers {
		result, err := h.requestInference(ctx, provider, req)
		if err != nil {
			continue // Try next provider
		}

		// Success - record stats and return response
		h.statsCollector.RecordRequest(req.Model, time.Since(startTime).Milliseconds(), 100)
		c.JSON(http.StatusOK, result)
		return
	}

	// All providers failed
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to get inference from any provider"})
}

// requestInference sends an inference request to a specific provider
func (h *Handler) requestInference(ctx context.Context, providerID peer.ID, req InferenceRequest) (*InferenceResponse, error) {
	// Create P2P inference request
	p2pReq := map[string]interface{}{
		"prompt":     req.Prompt,
		"model":      req.Model,
		"max_tokens": req.MaxTokens,
		"stream":     req.Stream,
	}

	reqData, err := json.Marshal(p2pReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to provider via P2P stream
	stream, err := h.p2pClient.SendRequest(ctx, providerID, reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to send P2P request: %w", err)
	}
	defer stream.Close()

	// Read response
	respData, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var p2pResp map[string]interface{}
	if err := json.Unmarshal(respData, &p2pResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to API response format
	resp := &InferenceResponse{
		Completion: p2pResp["completion"].(string),
		Model:      req.Model,
		Receipt: Receipt{
			Receipt: InnerReceipt{
				ProviderPK: providerID.String(),
				Timestamp:  time.Now().Unix(),
			},
			Signature: p2pResp["signature"].(string),
		},
	}

	return resp, nil
}

// Health handles health check requests
func (h *Handler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	providers := h.p2pClient.GetProviders(ctx)
	
	c.JSON(http.StatusOK, gin.H{
		"status":         "healthy",
		"timestamp":      time.Now().Unix(),
		"providers":      len(providers),
		"p2p_connected":  h.p2pClient.IsConnected(),
		"peers":          h.p2pClient.PeerCount(),
	})
}